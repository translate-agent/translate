package cmd

import (
	"context"
	"fmt"
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

const cmdTimeout = 10 * time.Second

func init() {
	serviceFlags := serviceCmd.PersistentFlags()

	serviceFlags.StringP("address", "a", "localhost:8080", `"translate" service address as "host:port"`)
	serviceFlags.BoolP("insecure", "i", false, `disable transport security (default false)`)
	serviceFlags.DurationP("timeout", "t", cmdTimeout, `command execution timeout`)

	if err := viper.BindPFlags(serviceFlags); err != nil {
		log.Panicf("bind flags: %v", err)
	}

	serviceCmd.AddCommand(lsCmd)
}

// serviceCmd represents the service command.
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Panicf("display help: %v", err)
		}
	},
}

// lsCmd represents the lsCmd command.
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List services",
	Run: func(cmd *cobra.Command, args []string) {
		timeout, err := cmd.InheritedFlags().GetDuration("timeout")
		if err != nil {
			log.Panicf("list services: retrieve cli parameter 'timeout': %v", err)
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
		defer cancelFunc()

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
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	address, err := cmd.InheritedFlags().GetString("address")
	if err != nil {
		return nil, fmt.Errorf("retrieve cli parameter 'insecure': %w", err)
	}

	isInsecure, err := cmd.InheritedFlags().GetBool("insecure")
	if err != nil {
		return nil, fmt.Errorf("retrieve cli parameter 'insecure': %w", err)
	}

	if isInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Wait for the server to start and establish a connection.
	return grpc.DialContext(ctx, address, opts...) //nolint:wrapcheck
}
