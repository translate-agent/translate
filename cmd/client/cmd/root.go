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

var conn *grpc.ClientConn

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "translate",
		TraverseChildren:  true,
		Short:             "Translate provides tools for interacting with translate agent service",
		PersistentPreRunE: rootCmdPersistentPreRunE,
		RunE:              func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
		// TODO: Add graceful connection close, in PersistentPostRunE
	}

	rootCmd.AddCommand(newServiceCmd())

	return rootCmd
}

func rootCmdPersistentPreRunE(cmd *cobra.Command, _ []string) error {
	address, err := cmd.InheritedFlags().GetString("address")
	if err != nil {
		return fmt.Errorf("get cli parameter 'address': %w", err)
	}

	connIsInsecure, err := cmd.InheritedFlags().GetBool("insecure")
	if err != nil {
		return fmt.Errorf("get cli parameter 'insecure': %w", err)
	}

	var opts []grpc.DialOption

	if connIsInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err = grpc.NewClient(address, opts...)
	if err != nil {
		return fmt.Errorf("create gRPC client: %w", err)
	}

	return nil
}

func Execute(ctx context.Context) error {
	if err := newRootCmd().ExecuteContext(ctx); err != nil { //nolint:contextcheck
		return fmt.Errorf("execute root command: %w", err)
	}

	return nil
}

// ExecuteWithParams executes root command using passed in command parameters,
// returns output result bytes and error.
func ExecuteWithParams(ctx context.Context, params []string) ([]byte, error) {
	rootCmd := newRootCmd() //nolint:contextcheck
	rootCmd.SetArgs(params)

	var output []byte
	buf := bytes.NewBuffer(output)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		buf.WriteString(err.Error())
		return nil, fmt.Errorf("execute root command with params: %w", err)
	}

	return buf.Bytes(), nil
}

// helpers

type schema string

// String is used both by fmt.Print and by Cobra in help text.
func (s schema) String() string {
	return string(s)
}

// Set must have pointer receiver so it doesn't change the value of a copy.
func (s *schema) Set(v string) error {
	v = strings.ToUpper(v)

	if sv, ok := translatev1.Schema_value[v]; ok && sv != 0 {
		*s = schema(v)
		return nil
	}

	availableSchemas := make([]string, len(translatev1.Schema_name))
	for schemaValue, schemaName := range translatev1.Schema_name {
		availableSchemas[schemaValue] = strings.ToLower(schemaName)
	}

	return fmt.Errorf("invalid schema value: must be one of: %s", strings.Join(availableSchemas[1:], ", "))
}

// Type is only used in help text.
func (s schema) Type() string {
	return "schema"
}

func (s schema) ToTranslateSchema() (translatev1.Schema, error) {
	if v, ok := translatev1.Schema_value[s.String()]; ok {
		return translatev1.Schema(v), nil
	}

	return translatev1.Schema_UNSPECIFIED, errors.New("schema doesn't match translate schema patterns")
}
