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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//nolint:gocognit
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

			serviceID, err := cmd.Flags().GetString("service")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'service': %w", err)
			}

			language, err := cmd.Flags().GetString("language")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'language': %w", err)
			}

			filePath, err := cmd.Flags().GetString("file")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'file': %w", err)
			}

			var original *bool

			if cmd.Flags().Changed("original") {
				v, flagErr := cmd.Flags().GetBool("original")
				if flagErr != nil {
					return fmt.Errorf("upload file: get cli parameter 'original': %w", err)
				}

				original = &v
			}

			populateTranslations, err := cmd.Flags().GetBool("populate_translations")
			if err != nil {
				return fmt.Errorf("upload file: get cli parameter 'populate_translations': %w", err)
			}

			var data []byte

			if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
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
					Language:             language,
					Data:                 data,
					Schema:               translateSchema,
					ServiceId:            serviceID,
					Original:             original,
					PopulateTranslations: populateTranslations,
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
	uploadFlags.String("service", "", "service UUID")
	uploadFlags.String("file", "", "local path or URL for the translation file")
	uploadFlags.String("language", "", "translation language")
	uploadFlags.Var(&schemaFlag, "schema",
		`translate schema, allowed: 'json_ng_localize', 'json_ngx_translate', 'go', 'arb', 'pot', 'xliff_12', 'xliff_2'`)
	uploadFlags.Bool("original", false, "file's language is an original language")
	uploadFlags.Bool("populate_translations", true, "populate translation messages from original file")

	if err := uploadCmd.MarkFlagRequired("service"); err != nil {
		log.Panicf("upload file cmd: set field 'service' as required: %v", err)
	}

	if err := uploadCmd.MarkFlagRequired("file"); err != nil {
		log.Panicf("upload file cmd: set field 'file' as required: %v", err)
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
		return nil, fmt.Errorf("validate file URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("prepare request to fetch file: %w", err)
	}

	resp, err := otelhttp.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch file: %w", err)
	}

	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch file status: got %s, expected 200", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read fetched file: %w", err)
	}

	return data, nil
}
