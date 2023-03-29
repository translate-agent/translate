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
	flags := lsCmd.Flags()

	flags.StringP("address", "a", "localhost:8080", `"translate" service address as "host:port"`)
	flags.BoolP("insecure", "i", false, `disable transport security (default false)`)
	flags.DurationP("timeout", "t", establishConnTimeout, `timeout for establishing connection with "translate" service`)

	if err := viper.BindPFlag("address", flags.Lookup("address")); err != nil {
		log.Panicf("bind address flag: %v", err)
	}

	if err := viper.BindPFlag("insecure", flags.Lookup("insecure")); err != nil {
		log.Panicf("bind insecure flag: %v", err)
	}

	if err := viper.BindPFlag("timeout", flags.Lookup("timeout")); err != nil {
		log.Panicf("bind timeout flag: %v", err)
	}

	serviceCmd.AddCommand(lsCmd)
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
