package service

import (
	"log"

	"github.com/spf13/cobra"
)

// serviceCmd represents the service command.
var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service holds list of commands for interacting with services.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Panicf("display help: %v", err)
		}
	},
}
