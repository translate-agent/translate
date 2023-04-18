package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// helpers

func newClientConn(ctx context.Context, cmd *cobra.Command) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	address, err := cmd.InheritedFlags().GetString("address")
	if err != nil {
		return nil, fmt.Errorf("get cli parameter 'address': %w", err)
	}

	connIsInsecure, err := cmd.InheritedFlags().GetBool("insecure")
	if err != nil {
		return nil, fmt.Errorf("get cli parameter 'insecure': %w", err)
	}

	if connIsInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Wait for the server to start and establish a connection.
	return grpc.DialContext(ctx, address, opts...) //nolint:wrapcheck
}

type schema string

// String is used both by fmt.Print and by Cobra in help text.
func (s schema) String() string {
	return string(s)
}

// Set must have pointer receiver so it doesn't change the value of a copy.
func (s *schema) Set(v string) error {
	switch v {
	case "json_ng_localize", "json_ngx_translate", "go", "arb", "pot", "xliff_12", "xliff_2":
		*s = schema(v)
		return nil
	default:
		return errors.New(
			`must be one of "json_ng_localize", "json_ngx_translate", "go", "arb", "pot", "xliff_12", "xliff_2"`)
	}
}

// Type is only used in help text.
func (s schema) Type() string {
	return "schema"
}

func (s schema) ToTranslateSchema() (translatev1.Schema, error) {
	if schemaNum, ok := translatev1.Schema_value[strings.ToUpper(s.String())]; ok {
		return translatev1.Schema(schemaNum), nil
	}

	return translatev1.Schema_UNSPECIFIED, errors.New("schema doesn't match translate schema patterns")
}
