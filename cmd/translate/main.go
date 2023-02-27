package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"go.expect.digital/translate/pkg/tracer"
	"go.expect.digital/translate/pkg/translate"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "translate",
	Short: "Enables translation for Cloud-native systems",
	Long:  `Enables translation for Cloud-native systems`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := ":" + viper.GetString("service.port")
		// Gracefully shutdown on Ctrl+C and Termination signal
		terminationChan := make(chan os.Signal, 1)
		signal.Notify(terminationChan, syscall.SIGTERM, syscall.SIGINT)

		tp, err := tracer.TracerProvider()
		if err != nil {
			log.Panic(err)
		}

		defer func() {
			if tpShutdownErr := tp.Shutdown(context.Background()); tpShutdownErr != nil {
				log.Panic(tpShutdownErr)
			}
		}()

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		)
		// Gracefully stops GRPC server and closes listener (multiplexer).
		defer grpcServer.GracefulStop()

		mux := runtime.NewServeMux()
		pb.RegisterTranslateServiceServer(grpcServer, &translate.TranslateServiceServer{})

		err = pb.RegisterTranslateServiceHandlerFromEndpoint(
			context.Background(),
			mux,
			addr,
			[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
		if err != nil {
			log.Panic(err)
		}

		httpServer := http.Server{
			Handler:           mux,
			ReadHeaderTimeout: time.Second * 5, //nolint:gomnd
		}

		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Panic(err)
		}

		multiplexer := cmux.New(l)
		// a different listener for HTTP1
		httpL := multiplexer.Match(cmux.HTTP1Fast())
		// a different listener for HTTP2 since gRPC uses HTTP2
		grpcL := multiplexer.Match(cmux.HTTP2())

		go func() {
			grpcErr := grpcServer.Serve(grpcL)
			// After grpcServer.GracefulStop(), Serve() returns nil.
			if grpcErr != nil {
				log.Panicf("gRPC serve: %v", grpcErr)
			}
		}()

		go func() {
			httpErr := httpServer.Serve(httpL)
			// After grpcServer.GracefulStop(), Serve() returns cmux.ErrServerClosed as the multiplexer is already closed.
			if httpErr != nil && !errors.Is(httpErr, cmux.ErrServerClosed) {
				log.Panicf("http serve: %v", httpErr)
			}
		}()

		go func() {
			muxErr := multiplexer.Serve()
			// After grpcServer.GracefulStop(), Serve() returns net.ErrClosed as the multiplexer is already closed.
			if muxErr != nil && !errors.Is(muxErr, net.ErrClosed) {
				log.Panicf("multiplexer serve :%v", muxErr)
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./translate.yaml)")
	rootCmd.PersistentFlags().Uint("port", 8080, "port to run on") //nolint:gomnd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Use default config file if not set.
	if cfgFile == "" {
		cfgFile = "translate.yaml"
	}

	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("translate")
	// Replace underscores with dots in environment variable names.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to read config.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// For now manually bind CLI argument to viper.
	err := viper.BindPFlag("service.port", rootCmd.Flags().Lookup("port"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
