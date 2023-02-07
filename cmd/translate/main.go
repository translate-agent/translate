package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pb "go.expect.digital/translate/pkg/server/translate/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TranslateServiceServer struct {
	pb.UnimplementedTranslateServiceServer
}

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "translate",
	Short: "Enables translation for Cloud-native systems",
	Long:  `Enables translation for Cloud-native systems`,
	Run: func(cmd *cobra.Command, args []string) {
		grpcSever := grpc.NewServer()
		mux := runtime.NewServeMux()
		err := pb.RegisterTranslateServiceHandlerFromEndpoint(
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

		l, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Panic(err)
		}

		m := cmux.New(l)
		// a different listener for HTTP1
		httpL := m.Match(cmux.HTTP1Fast())
		// a different listener for HTTP2 since gRPC uses HTTP2
		grpcL := m.Match(cmux.HTTP2())

		go func() {
			if err := server.Serve(httpL); err != nil {
				log.Fatal(err)
			}
		}()

		go func() {
			if err := grpcSever.Serve(grpcL); err != nil {
				log.Fatal(err)
			}
		}()

		if err := m.Serve(); err != nil {
			log.Panic(err)
		}
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
