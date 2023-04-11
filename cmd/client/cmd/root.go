package cmd

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	addSubcommandPalettes()
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate provides tools for interacting with translate agent service.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return fmt.Errorf("display help: %w", err)
		}
		return nil
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("execute root command: %w", err)
	}

	return nil
}

// ExecuteWithParams executes root command using passed in command parameters,
// returns output result bytes and error.
func ExecuteWithParams(params []string) ([]byte, error) {
	cmd := rootCmd
	cmd.SetArgs(params)

	var output []byte
	buf := bytes.NewBuffer(output)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := rootCmd.Execute(); err != nil {
		buf.WriteString(err.Error())
		return nil, fmt.Errorf("execute root command with params: %w", err)
	}

	return buf.Bytes(), nil
}

// helpers

func addSubcommandPalettes() {
	rootCmd.AddCommand(serviceCmd)
}
