package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
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
	serviceFlags.String("address", "localhost:8080", `"translate" service address as "host:port"`)
	serviceFlags.Bool("insecure", false, `disable transport security (default false)`)
	serviceFlags.Duration("timeout", cmdTimeout, `command execution timeout`)

	serviceCmd.AddCommand(newUploadCmd())
	serviceCmd.AddCommand(newDownloadCmd())
	serviceCmd.AddCommand(newLsCmd())

	return serviceCmd
}
