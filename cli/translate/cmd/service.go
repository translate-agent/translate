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

func init() {
	flags := lsCmd.Flags()

	flags.StringP("address", "a", "localhost:8080",
		"address for the translate agent GRPC client")

	if err := viper.BindPFlag("address", flags.Lookup("address")); err != nil {
		log.Panicf("bind address flag: %v", err)
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
		addr, err := cmd.Flags().GetString("address")
		if err != nil {
			log.Panicf("list services: retrieve service address env variable: %v", err)
		}

		ctx := context.Background()

		client, err := newClientConn(ctx, addr)
		if err != nil {
			log.Panicf("list services: new GRPC client connection: %v", err)
		}

		resp, err := translatev1.NewTranslateServiceClient(client).ListServices(ctx, &translatev1.ListServicesRequest{})
		if err != nil {
			log.Panicf("list services: make GRPC request: %v", err)
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

func newClientConn(ctx context.Context, address string) (*grpc.ClientConn, error) {
	ctx, cancelFunc := context.WithTimeout(ctx, time.Second*10) //nolint:gomnd
	defer cancelFunc()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	// Wait for the server to start and establish a connection.
	return grpc.DialContext(ctx, address, opts...) //nolint:wrapcheck
}
