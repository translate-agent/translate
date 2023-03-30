package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/repo/mysql"
	"go.expect.digital/translate/pkg/tracer"
	"go.expect.digital/translate/pkg/translate"
)

var cfgFile string

// grpcHandlerFunc returns an http.Handler that routes gRPC and non-gRPC requests to the appropriate handler.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "translate",
	Short: "Enables translation for Cloud-native systems",
	Long:  `Enables translation for Cloud-native systems`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		addr := viper.GetString("service.host") + ":" + viper.GetString("service.port")
		// Gracefully shutdown on Ctrl+C and Termination signal
		terminationChan := make(chan os.Signal, 1)
		signal.Notify(terminationChan, syscall.SIGTERM, syscall.SIGINT)

		tp, err := tracer.TracerProvider()
		if err != nil {
			log.Panicf("set tracer provider: %v", err)
		}

		defer func() {
			if tpShutdownErr := tp.Shutdown(ctx); tpShutdownErr != nil {
				log.Panicf("gracefully shutdown tracer: %v", tpShutdownErr)
			}
		}()

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		)
		// Gracefully stops GRPC server and closes listener (multiplexer).
		defer grpcServer.GracefulStop()

		mux := runtime.NewServeMux()

		var (
			repository repo.Repo
			errRepo    error
		)

		switch v := strings.TrimSpace(strings.ToLower(viper.GetString("service.db"))); v {
		case "mysql":
			repository, errRepo = mysql.NewRepo(mysql.WithDefaultDB(ctx))
		default:
			log.Panicf("unsupported db: '%s'", v)
		}

		if errRepo != nil {
			log.Panicf("create new repo: %v", errRepo)
		}

		translatev1.RegisterTranslateServiceServer(grpcServer, translate.NewTranslateServiceServer(repository))

		err = translatev1.RegisterTranslateServiceHandlerFromEndpoint(
			ctx,
			mux,
			addr,
			[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
		if err != nil {
			log.Panicf("register translate service: %v", err)
		}

		httpServer := http.Server{
			Addr:              addr,
			Handler:           grpcHandlerFunc(grpcServer, mux),
			ReadHeaderTimeout: time.Second * 5, //nolint:gomnd
		}

		go func() {
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Panicf("server serve: %v", err)
			}
		}()

		// Block until termination signal is received.
		<-terminationChan
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "translate.yaml", "config file")
	rootCmd.PersistentFlags().Uint("port", 8080, "port to run service on") //nolint:gomnd
	rootCmd.PersistentFlags().String("host", "localhost", "host to run service on")
	rootCmd.PersistentFlags().String("db", "mysql", "database to use with service")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("translate")
	// Replace underscores with dots in environment variable names.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to read config.
	if err := viper.ReadInConfig(); err != nil && cfgFile != "translate.yaml" {
		log.Panicf("read config: %v", err)
	}

	err := viper.BindPFlag("service.port", rootCmd.Flags().Lookup("port"))
	if err != nil {
		log.Panicf("bind port flag: %v", err)
	}

	err = viper.BindPFlag("service.host", rootCmd.Flags().Lookup("host"))
	if err != nil {
		log.Panicf("bind host flag: %v", err)
	}

	err = viper.BindPFlag("service.db", rootCmd.Flags().Lookup("db"))
	if err != nil {
		log.Panicf("bind db flag: %v", err)
	}
}
