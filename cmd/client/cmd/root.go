package cmd

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "translate",
		TraverseChildren: true,
		Short:            "Translate provides tools for interacting with translate agent service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return fmt.Errorf("display help: %w", err)
			}
			return nil
		},
	}

	rootCmd.AddCommand(newServiceCmd())

	return rootCmd
}

func Execute() error {
	if err := newRootCmd().Execute(); err != nil {
		return fmt.Errorf("execute root command: %w", err)
	}

	return nil
}

// ExecuteWithParams executes root command using passed in command parameters,
// returns output result bytes and error.
func ExecuteWithParams(params []string) ([]byte, error) {
	rootCmd := newRootCmd()
	rootCmd.SetArgs(params)

	var output []byte
	buf := bytes.NewBuffer(output)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	if err := rootCmd.Execute(); err != nil {
		buf.WriteString(err.Error())
		return nil, fmt.Errorf("execute root command with params: %w", err)
	}

	return buf.Bytes(), nil
}
