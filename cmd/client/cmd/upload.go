package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

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

			ctx, cancelFunc := context.WithTimeout(cmd.Context(), timeout)
			defer cancelFunc()

			client, err := newClientConn(ctx, cmd)
			if err != nil {
				return fmt.Errorf("upload file: new GRPC client connection: %w", err)
			}

			serviceID, err := cmd.Flags().GetString("serviceID")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'serviceID': %w", err)
			}

			language, err := cmd.Flags().GetString("language")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'language': %w", err)
			}

			filePath, err := cmd.Flags().GetString("file")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'file': %w", err)
			}

			var data []byte

			if strings.HasPrefix(filePath, "http") || strings.HasPrefix(filePath, "https") {
				if data, err = readFileFromURL(ctx, filePath); err != nil {
					return fmt.Errorf("upload file: read file from URL: %w", err)
				}
			} else {
				if data, err = os.ReadFile(filePath); err != nil {
					return fmt.Errorf("upload file: read file from local path: %w", err)
				}
			}

			translateSchema, err := schemaFlag.ToTranslateSchema()
			if err != nil {
				return fmt.Errorf("upload file: schema to translate schema: %w", err)
			}

			if _, err = translatev1.NewTranslateServiceClient(client).UploadTranslationFile(ctx,
				&translatev1.UploadTranslationFileRequest{
					Language: language, Data: data, Schema: translateSchema, ServiceId: serviceID,
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
	uploadFlags.String("serviceID", "", "service UUID")
	uploadFlags.String("file", "", "translation file file path or URL")
	uploadFlags.String("language", "", "translation language")
	uploadFlags.Var(&schemaFlag, "schema",
		`translate schema, allowed: 'json_ng_localize', 'json_ngx_translate', 'go', 'arb', 'pot', 'xliff_12', 'xliff_2'`)

	if err := uploadCmd.MarkFlagRequired("serviceID"); err != nil {
		log.Panicf("upload file cmd: set field 'serviceID' as required: %v", err)
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

// readFileFromURL reads translation file from URL.
func readFileFromURL(ctx context.Context, filePath string) ([]byte, error) {
	u, err := url.ParseRequestURI(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare new GET request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do GET request: %w", err)
	}

	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body data: %w", err)
	}

	return data, nil
}
