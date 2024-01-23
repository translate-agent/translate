package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

// file formats.
const (
	arb  = "arb"
	json = "json"
	po   = "po"
	xlf  = "xlf"
)

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

			ctx, cancelFunc := context.WithTimeout(cmd.Context(), timeout)
			defer cancelFunc()

			client, err := newClientConn(ctx, cmd)
			if err != nil {
				return fmt.Errorf("download file: new GRPC client connection: %w", err)
			}

			serviceID, err := cmd.Flags().GetString("service")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'service': %w", err)
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

			fileName := serviceID + "_" + language

			switch translateSchema {
			case translatev1.Schema_UNSPECIFIED:
				return errors.New("download file: unspecified file schema")
			case translatev1.Schema_ARB:
				fileName += "." + arb
			case translatev1.Schema_JSON_NG_LOCALIZE, translatev1.Schema_JSON_NGX_TRANSLATE, translatev1.Schema_GO:
				fileName += "." + json
			case translatev1.Schema_PO:
				fileName += "." + po
			case translatev1.Schema_XLIFF_12, translatev1.Schema_XLIFF_2:
				fileName += "." + xlf
			}

			if err = os.WriteFile(filepath.Join(path, fileName), res.GetData(), 0644); err != nil { //nolint:gomnd,gofumpt,gosec
				return fmt.Errorf("download file: write file to path: %w", err)
			}

			if _, err = fmt.Fprintln(cmd.OutOrStdout(), "File downloaded successfully."); err != nil {
				return fmt.Errorf("download file: output response to stdout: %w", err)
			}

			return nil
		},
	}

	downloadFlags := downloadCmd.Flags()
	downloadFlags.String("service", "", "service UUID")
	downloadFlags.String("path", "", "download folder path")
	downloadFlags.String("language", "", "translation language in BCP47 format")
	downloadFlags.Var(&schemaFlag, "schema",
		`translate schema, allowed: 'json_ng_localize', 'json_ngx_translate', 'go', 'arb', 'po', 'xliff_12', 'xliff_2'`)

	if err := downloadCmd.MarkFlagRequired("service"); err != nil {
		log.Panicf("download file cmd: set field 'service' as required: %v", err)
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
