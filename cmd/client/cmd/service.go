package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// file formats
const (
	arb  = ".arb"
	json = ".json"
	pot  = ".po"
	xlf  = ".xlf"
)

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

	serviceCmd.AddCommand(newUploadCmd())
	serviceCmd.AddCommand(newDownloadCmd())
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
				return fmt.Errorf("list services: get cli parameter 'timeout': %w", err)
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
			tbl := table.New("ID", "Name")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for _, v := range resp.Services {
				tbl.AddRow(v.Id, v.Name)
			}

			tbl.WithWriter(cmd.OutOrStdout())
			tbl.Print()

			return nil
		},
	}

	return lsCmd
}

func newUploadCmd() *cobra.Command {
	var schemaFlag schema

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload file to translate agent service",
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := cmd.InheritedFlags().GetDuration("timeout")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'timeout': %w", err)
			}

			ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
			defer cancelFunc()

			client, err := newClientConn(ctx, cmd)
			if err != nil {
				return fmt.Errorf("upload file: new GRPC client connection: %w", err)
			}

			serviceID, err := cmd.Flags().GetString("serviceUUID")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'serviceUUID': %w", err)
			}

			language, err := cmd.Flags().GetString("language")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'language': %w", err)
			}

			filePath, err := cmd.Flags().GetString("file")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'file': %w", err)
			}

			fileID, err := cmd.Flags().GetString("fileUUID")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'fileUUID': %w", err)
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("upload file: read file from path: %w", err)
			}

			translateSchema, err := schemaFlag.ToTranslateSchema()
			if err != nil {
				return fmt.Errorf("upload file: schema to translate schema: %w", err)
			}

			if _, err = translatev1.NewTranslateServiceClient(client).UploadTranslationFile(ctx,
				&translatev1.UploadTranslationFileRequest{
					Language: language, Data: data, Schema: translateSchema, ServiceId: serviceID, TranslationFileId: fileID,
				}); err != nil {
				return fmt.Errorf("upload file: send GRPC request: %w", err)
			}

			if _, err = fmt.Fprintln(cmd.OutOrStdout(), "File uploaded successfully."); err != nil {
				return fmt.Errorf("upload file: output response to stdout: %w", err)
			}

			return nil
		},
	}

	uploadFlags := uploadCmd.Flags()
	uploadFlags.StringP("serviceUUID", "u", "", "service UUID")
	uploadFlags.StringP("fileUUID", "p", "", "translation file UUID")
	uploadFlags.StringP("file", "f", "", "file path")
	uploadFlags.StringP("language", "l", "", "translation language")
	uploadFlags.VarP(&schemaFlag, "schema", "s",
		`translate schema, allowed: 'json_ng_localize', 'json_ngx_translate', 'go', 'arb', 'pot', 'xliff_12', 'xliff_2'`)

	if err := uploadCmd.MarkFlagRequired("serviceUUID"); err != nil {
		log.Panicf("upload file cmd: set field 'serviceUUID' as required: %v", err)
	}

	if err := uploadCmd.MarkFlagRequired("file"); err != nil {
		log.Panicf("upload file cmd: set field 'file' as required: %v", err)
	}

	if err := uploadCmd.MarkFlagRequired("language"); err != nil {
		log.Panicf("upload file cmd: set field 'language' as required: %v", err)
	}

	if err := uploadCmd.MarkFlagRequired("schema"); err != nil {
		log.Panicf("upload file cmd: set field 'schema' as required: %v", err)
	}

	if err := viper.BindPFlags(uploadFlags); err != nil {
		log.Panicf("upload file cmd: bind flags: %v", err)
	}

	return uploadCmd
}

func newDownloadCmd() *cobra.Command {
	var schemaFlag schema

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download file from translate agent service",
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := cmd.InheritedFlags().GetDuration("timeout")
			if err != nil {
				return fmt.Errorf("download file: get cli parameter 'timeout': %w", err)
			}

			ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
			defer cancelFunc()

			client, err := newClientConn(ctx, cmd)
			if err != nil {
				return fmt.Errorf("download file: new GRPC client connection: %w", err)
			}

			serviceID, err := cmd.Flags().GetString("serviceUUID")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'serviceUUID': %w", err)
			}

			path, err := cmd.Flags().GetString("path")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'path': %w", err)
			}

			language, err := cmd.Flags().GetString("language")
			if err != nil {
				return fmt.Errorf("download file: get cli parameter 'language': %w", err)
			}

			translateSchema, err := schemaFlag.ToTranslateSchema()
			if err != nil {
				return fmt.Errorf("download file: schema to translate schema: %w", err)
			}

			res, err := translatev1.NewTranslateServiceClient(client).DownloadTranslationFile(ctx,
				&translatev1.DownloadTranslationFileRequest{
					Language: language, Schema: translateSchema, ServiceId: serviceID,
				})
			if err != nil {
				return fmt.Errorf("download file: send GRPC request: %w", err)
			}

			fileName := serviceID

			switch translateSchema {
			case translatev1.Schema_UNSPECIFIED:
				return errors.New("download file: unspecified file schema")
			case translatev1.Schema_ARB:
				fileName += arb
			case translatev1.Schema_JSON_NG_LOCALIZE, translatev1.Schema_JSON_NGX_TRANSLATE,
				translatev1.Schema_GO:
				fileName += json
			case translatev1.Schema_POT:
				fileName += pot
			case translatev1.Schema_XLIFF_12, translatev1.Schema_XLIFF_2:
				fileName += xlf
			}

			if err = os.WriteFile(filepath.Join(path, fileName), res.Data, 0600); err != nil { //nolint:gomnd,gofumpt
				return fmt.Errorf("download file: write file to path: %w", err)
			}

			if _, err = fmt.Fprintln(cmd.OutOrStdout(), "File downloaded successfully."); err != nil {
				return fmt.Errorf("download file: output response to stdout: %w", err)
			}

			return nil
		},
	}

	downloadFlags := downloadCmd.Flags()
	downloadFlags.StringP("serviceUUID", "u", "", "service UUID")
	downloadFlags.StringP("path", "p", "", "download folder path")
	downloadFlags.StringP("language", "l", "", "translation language")
	downloadFlags.VarP(&schemaFlag, "schema", "s",
		`translate schema, allowed: 'json_ng_localize', 'json_ngx_translate', 'go', 'arb', 'pot', 'xliff_12', 'xliff_2'`)

	if err := downloadCmd.MarkFlagRequired("serviceUUID"); err != nil {
		log.Panicf("download file cmd: set field 'serviceUUID' as required: %v", err)
	}

	if err := downloadCmd.MarkFlagRequired("path"); err != nil {
		log.Panicf("download file cmd: set field 'path' as required: %v", err)
	}

	if err := downloadCmd.MarkFlagRequired("language"); err != nil {
		log.Panicf("download file cmd: set field 'language' as required: %v", err)
	}

	if err := downloadCmd.MarkFlagRequired("schema"); err != nil {
		log.Panicf("download file cmd: set field 'schema' as required: %v", err)
	}

	if err := viper.BindPFlags(downloadFlags); err != nil {
		log.Panicf("download file cmd: bind flags: %v", err)
	}

	return downloadCmd
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
func (s *schema) String() string {
	return string(*s)
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
func (s *schema) Type() string {
	return "schema"
}

func (s schema) ToTranslateSchema() (translatev1.Schema, error) {
	if schemaNum, ok := translatev1.Schema_value[strings.ToUpper(s.String())]; ok {
		return translatev1.Schema(schemaNum), nil
	}

	return translatev1.Schema_UNSPECIFIED, errors.New("schema doesn't match translate schema patterns")
}
