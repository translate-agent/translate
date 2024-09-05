package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func newLsCmd(svc *Service) *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			timeout, err := cmd.InheritedFlags().GetDuration("timeout")
			if err != nil {
				return fmt.Errorf("list services: get cli parameter 'timeout': %w", err)
			}

			ctx, cancelFunc := context.WithTimeout(cmd.Context(), timeout)
			defer cancelFunc()

			resp, err := svc.client.ListServices(ctx, &translatev1.ListServicesRequest{})
			if err != nil {
				return fmt.Errorf("list services: send gRPC request: %w", err)
			}

			headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt := color.New(color.FgYellow).SprintfFunc()
			tbl := table.New("ID", "Name")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for _, v := range resp.GetServices() {
				tbl.AddRow(v.GetId(), v.GetName())
			}

			tbl.WithWriter(cmd.OutOrStdout())
			tbl.Print()

			return nil
		},
	}

	return lsCmd
}
