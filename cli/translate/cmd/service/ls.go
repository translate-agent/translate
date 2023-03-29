package service

import (
	"context"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// lsCmd represents the lsCmd command.
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists services.",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := cmd.Flags().GetString("service.address")
		if err != nil {
			log.Panicf("list services: retrieve service address env variable: %s", err)
		}

		ctx := context.Background()

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}

		// Wait for the server to start and establish a connection.
		conn, err := grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			log.Panicf("list services: connect to service: %s", err)
		}

		client := translatev1.NewTranslateServiceClient(conn)

		resp, err := client.ListServices(ctx, &translatev1.ListServicesRequest{})
		if err != nil {
			log.Panicf("list services: grpc request: %s", err)
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

func init() {
	flags := lsCmd.Flags()

	flags.StringP("service.address", "a", "localhost:8080",
		"address for the translate agent GRPC client")

	if err := viper.BindPFlag("service.address", flags.Lookup("service.address")); err != nil {
		log.Panicf("bind address flag: %s", err)
	}

	ServiceCmd.AddCommand(lsCmd)
}
