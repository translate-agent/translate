package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	addSubcommandPalettes()
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate provides tools for interacting with translate agent service.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Panicf("display help: %v", err)
		}
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("execute root command: %w", err)
	}

	return nil
}

// helpers

func addSubcommandPalettes() {
	rootCmd.AddCommand(serviceCmd)
}
