package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const cmdTimeout = 10 * time.Second

func newServiceCmd() *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Manage services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return fmt.Errorf("display help: %w", err)
			}
			return nil
		},
	}

	serviceFlags := serviceCmd.PersistentFlags()

	serviceFlags.StringP("address", "a", "localhost:8080", `"translate" service address as "host:port"`)
	serviceFlags.BoolP("insecure", "i", false, `disable transport security (default false)`)
	serviceFlags.DurationP("timeout", "t", cmdTimeout, `command execution timeout`)

	if err := viper.BindPFlags(serviceFlags); err != nil {
		log.Panicf("service cmd: bind flags: %v", err)
	}

	serviceCmd.AddCommand(newLsCmd())

	return serviceCmd
}

func newLsCmd() *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List services",
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := cmd.InheritedFlags().GetDuration("timeout")
			if err != nil {
				return fmt.Errorf("list services: retrieve cli parameter 'timeout': %w", err)
			}

			ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
			defer cancelFunc()

			client, err := newClientConn(ctx, cmd)
			if err != nil {
				return fmt.Errorf("list services: new GRPC client connection: %w", err)
			}

			resp, err := translatev1.NewTranslateServiceClient(client).ListServices(ctx, &translatev1.ListServicesRequest{})
			if err != nil {
				return fmt.Errorf("list services: send GRPC request: %w", err)
			}

			headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt := color.New(color.FgYellow).SprintfFunc()
			tbl := table.New("#", "ID", "Name")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for i, v := range resp.Services {
				tbl.AddRow(i+1, v.Id, v.Name)
			}

			tbl.AddRow("", "TOTAL", len(resp.Services))

			tbl.WithWriter(cmd.OutOrStdout())
			tbl.Print()

			return nil
		},
	}

	return lsCmd
}

// helpers

func newClientConn(ctx context.Context, cmd *cobra.Command) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	address, err := cmd.InheritedFlags().GetString("address")
	if err != nil {
		return nil, fmt.Errorf("retrieve cli parameter 'address': %w", err)
	}

	connIsInsecure, err := cmd.InheritedFlags().GetBool("insecure")
	if err != nil {
		return nil, fmt.Errorf("retrieve cli parameter 'insecure': %w", err)
	}

	if connIsInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Wait for the server to start and establish a connection.
	return grpc.DialContext(ctx, address, opts...) //nolint:wrapcheck
}
