package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"go.expect.digital/translate/pkg/tracer"
	"go.expect.digital/translate/pkg/translate"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var cfgFile string

// package level termination channel for integration tests.
var terminationChan = make(chan os.Signal, 1)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "translate",
	Short: "Enables translation for Cloud-native systems",
	Long:  `Enables translation for Cloud-native systems`,
	Run: func(cmd *cobra.Command, args []string) {
		tpShutdown, err := tracer.TracerProvider("http://localhost:14268/api/traces", "translate")
		if err != nil {
			log.Panic(err)
		}

		defer func() {
			fmt.Printf("\n\n\nShutting down tp\n\n\n")
			if tpErr := tpShutdown(context.Background()); err != nil {
				log.Printf("shutdown TracerProvider: %v", tpErr)
			}
		}()

		signal.Notify(terminationChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		)
		defer grpcServer.GracefulStop()

		mux := runtime.NewServeMux()
		pb.RegisterTranslateServiceServer(grpcServer, &translate.TranslateServiceServer{})

		err = pb.RegisterTranslateServiceHandlerFromEndpoint(
			context.Background(),
			mux,
			"localhost:8080",
			[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
		if err != nil {
			log.Panic(err)
		}

		server := http.Server{
			Handler:           mux,
			ReadHeaderTimeout: time.Minute,
		}
		defer server.Shutdown(context.Background())

		l, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Panic(err)
		}

		m := cmux.New(l)
		// a different listener for HTTP1
		httpL := m.Match(cmux.HTTP1Fast())
		// a different listener for HTTP2 since gRPC uses HTTP2
		grpcL := m.Match(cmux.HTTP2())

		var terminated bool
		go func() {
			if httpErr := server.Serve(httpL); httpErr != nil && !terminated {
				log.Panic(httpErr)
			}
		}()

		go func() {
			if grpcErr := grpcServer.Serve(grpcL); grpcErr != nil && !terminated {
				log.Panic(grpcErr)
			}
		}()

		go func() {
			if cmuxErr := m.Serve(); cmuxErr != nil && !terminated {
				log.Panic(cmuxErr)
			}
		}()

		<-terminationChan
		terminated = true
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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find current dir.
		dir, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in current directory with name "translate.yaml".
		viper.AddConfigPath(dir)
		viper.SetConfigFile("translate.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
