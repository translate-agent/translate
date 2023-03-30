package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const establishConnTimeout = 5 * time.Second

func init() {
	// upload command uploadFlags
	uploadFlags := uploadCmd.Flags()
	uploadFlags.String("path", "", "path to file")
	uploadFlags.StringP("address", "a", "localhost:8080", `"translate" service address as "host:port"`)
	uploadFlags.BoolP("insecure", "i", false, `disable transport security (default false)`)
	uploadFlags.DurationP("timeout", "t", establishConnTimeout, `timeout for establishing connection with "translate" service`)

	// ls command lsFlags
	lsFlags := lsCmd.Flags()
	lsFlags.StringP("address", "a", "localhost:8080", `"translate" service address as "host:port"`)
	lsFlags.BoolP("insecure", "i", false, `disable transport security (default false)`)
	lsFlags.DurationP("timeout", "t", establishConnTimeout, `timeout for establishing connection with "translate" service`)

	if err := viper.BindPFlag("address", lsFlags.Lookup("address")); err != nil {
		log.Panicf("bind address flag: %v", err)
	}

	if err := viper.BindPFlag("insecure", lsFlags.Lookup("insecure")); err != nil {
		log.Panicf("bind insecure flag: %v", err)
	}

	if err := viper.BindPFlag("timeout", lsFlags.Lookup("timeout")); err != nil {
		log.Panicf("bind timeout flag: %v", err)
	}

	// add commands to serviceCmd
	serviceCmd.AddCommand(fileCmd)
	serviceCmd.AddCommand(lsCmd)

	// add commands to fileCmd
	fileCmd.AddCommand(uploadCmd)
}

// serviceCmd represents the service command.
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service holds list of commands for interacting with services.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Panicf("display help: %v", err)
		}
	},
}

// lsCmd represents the lsCmd command.
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists services.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		client, err := newClientConn(ctx, cmd)
		if err != nil {
			log.Panicf("list services: new GRPC client connection: %v", err)
		}

		resp, err := translatev1.NewTranslateServiceClient(client).ListServices(ctx, &translatev1.ListServicesRequest{})
		if err != nil {
			log.Panicf("list services: send GRPC request: %v", err)
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "ID", "Name"})

		tableRows := make([]table.Row, 0, len(resp.Services))

		for i := range resp.Services {
			tableRows = append(tableRows, table.Row{i + 1, resp.Services[i].Id, resp.Services[i].Name})
		}

		t.AppendRows(tableRows)
		t.AppendFooter(table.Row{"", "Total", len(resp.Services)})
		t.SetStyle(table.StyleLight)
		t.Render()
	},
}

// uploadCmd represents the uploadCmd command.
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads file to translate service.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		client, err := newClientConn(ctx, cmd)
		if err != nil {
			log.Panicf("list services: new GRPC client connection: %v", err)
		}

		filePath, err := cmd.Flags().GetDuration("path")
		if err != nil {
			log.Panicf("upload file: retrieve cli parameter 'path': %v", err)
		}

		resp, err := translatev1.NewTranslateServiceClient(client).UploadTranslationFile(context.Background())
		if err != nil {
			log.Panicf("upload file: send GRPC request: %v", err)
		}

	},
}

// fileCmd represents the service command.
var fileCmd = &cobra.Command{
	Use:   "service",
	Short: "File holds commands for file transfer in translate service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Panicf("display help: %v", err)
		}
	},
}

// helpers

func newClientConn(ctx context.Context, cmd *cobra.Command) (*grpc.ClientConn, error) {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		log.Panicf("list services: retrieve cli parameter 'address': %v", err)
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		log.Panicf("list services: retrieve cli parameter 'timeout': %v", err)
	}

	isInsecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		log.Panicf("list services: retrieve cli parameter 'insecure': %v", err)
	}

	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	if isInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Wait for the server to start and establish a connection.
	return grpc.DialContext(ctx, addr, opts...) //nolint:wrapcheck
}
