package service

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.expect.digital/translate/pkg/fuzzy"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo/factory"
	"go.expect.digital/translate/pkg/server"
	"go.expect.digital/translate/pkg/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
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
	RunE:  RootCmdRunE,
}

func RootCmdRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	addr := viper.GetString("service.host") + ":" + viper.GetString("service.port")
	// Gracefully shutdown on Ctrl+C and Termination signal
	terminationChan := make(chan os.Signal, 1)
	signal.Notify(terminationChan, syscall.SIGTERM, syscall.SIGINT)

	tp, err := tracer.TracerProvider(ctx)
	if err != nil {
		return fmt.Errorf("set tracer provider: %w", err)
	}

	defer func() {
		if tpShutdownErr := tp.Shutdown(ctx); tpShutdownErr != nil {
			log.Panicf("gracefully shutdown tracer: %v", tpShutdownErr)
		}
	}()

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	// Gracefully stops gRPC server.
	defer grpcServer.GracefulStop()

	mux := runtime.NewServeMux()

	repo, err := factory.NewRepo(ctx, viper.GetString("service.db"))
	if err != nil {
		return fmt.Errorf("create new repo: %w", err)
	}

	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			log.Printf("close repo: %v", closeErr)
		}
	}()

	var (
		translator    fuzzy.Translator
		translatorStr string
	)

	switch translatorStr = viper.GetString("service.translator"); translatorStr {
	case "":
		translator = &fuzzy.NoopTranslate{}
	case "AWSTranslate":
		translator, err = fuzzy.NewAWSTranslate(ctx, fuzzy.WithDefaultAWSClient(ctx))
	case "GoogleTranslate":
		var closeTranslate func() error
		translator, closeTranslate, err = fuzzy.NewGoogleTranslate(
			ctx, fuzzy.WithDefaultGoogleClient(ctx))

		defer func() {
			if closeErr := closeTranslate(); closeErr != nil {
				log.Printf("close GoogleTranslate client: %v\n", closeErr)
			}
		}()
	default:
		return fmt.Errorf("unsupported translator: %s", translatorStr)
	}

	if err != nil {
		return fmt.Errorf("create new %s client: %w", translatorStr, err)
	}

	translatev1.RegisterTranslateServiceServer(grpcServer, server.NewTranslateServiceServer(repo, translator))

	// gRPC Server Reflection provides information about publicly-accessible gRPC services on a server,
	// and assists clients at runtime to construct RPC requests and responses without precompiled service information.
	// https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md
	reflection.Register(grpcServer)

	err = translatev1.RegisterTranslateServiceHandlerFromEndpoint(
		ctx,
		mux,
		addr,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		})
	if err != nil {
		return fmt.Errorf("register translate service: %w", err)
	}

	httpServer := http.Server{
		Addr:              addr,
		Handler:           grpcHandlerFunc(grpcServer, otelhttp.NewHandler(mux, "grpc-gateway")),
		ReadHeaderTimeout: time.Second * 5, //nolint:mnd
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panicf("server serve: %v", err)
		}
	}()

	// Block until termination signal is received.
	<-terminationChan

	return nil
}

func Serve() {
	// Execute adds all child commands to the root command and sets flags appropriately.
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "translate.yaml", "config file")
	rootCmd.PersistentFlags().Uint("port", 8080, "port to run service on") //nolint:mnd
	rootCmd.PersistentFlags().String("host", "0.0.0.0", "host to run service on")
	rootCmd.PersistentFlags().String("db", "badgerdb", factory.Usage())
	rootCmd.PersistentFlags().String("translator", "", fuzzy.Usage())
}

var mutex = &sync.Mutex{}

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

	// Prevent concurrent writes to Viper which happens in tests.
	mutex.Lock()

	err := viper.BindPFlag("service.port", rootCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		log.Panicf("bind port flag: %v", err)
	}

	err = viper.BindPFlag("service.host", rootCmd.PersistentFlags().Lookup("host"))
	if err != nil {
		log.Panicf("bind host flag: %v", err)
	}

	err = viper.BindPFlag("service.db", rootCmd.PersistentFlags().Lookup("db"))
	if err != nil {
		log.Panicf("bind db flag: %v", err)
	}

	err = viper.BindPFlag("service.translator", rootCmd.PersistentFlags().Lookup("translator"))
	if err != nil {
		log.Panicf("bind translator flag: %v", err)
	}

	mutex.Unlock()
}
