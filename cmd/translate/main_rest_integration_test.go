package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"go.expect.digital/translate/pkg/translate"
)

const baseAddr = "http://localhost:8080"

func TestMain(m *testing.M) {
	go main()

	// Ensure that a connection can be established.
	conn, err := net.DialTimeout("tcp", "localhost:8080", time.Second)
	if err != nil {
		log.Panic(err)
	}

	conn.Close()

	os.Exit(m.Run())
}

func attachFile(text []byte, t *testing.T) (*bytes.Buffer, string, error) {
	t.Helper()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", "test.json")
	if err != nil {
		return nil, "", fmt.Errorf("creating form file: %w", err)
	}

	_, err = part.Write(text)
	if err != nil {
		return nil, "", fmt.Errorf("writing to file: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

func Test_UploadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	addr := baseAddr + "/v1/files"

	tests := []struct {
		name   string
		params translate.UploadParams
		want   uint
	}{
		{
			name: "Happy Path",
			params: translate.UploadParams{
				Language: translate.LanguageData{Str: "lv-LV"},
				Schema:   pb.Schema_GO,
				Data: []byte(`{
					"messages": [
						{
							"id": "1",
							"meaning": "When you great someone",
							"message": "hello",
							"translation": "čau",
							"fuzzy": false
						}
					]
				}
				`),
			},
			want: http.StatusOK,
		},
		{
			name: "Missing language",
			params: translate.UploadParams{
				Language: translate.LanguageData{},
				Schema:   pb.Schema_GO,
				Data: []byte(`{
					"messages": [
						{
							"id": "1",
							"meaning": "When you great someone",
							"message": "hello",
							"translation": "čau",
							"fuzzy": false
						}
					]
				}
				`),
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Missing file",
			params: translate.UploadParams{
				Language: translate.LanguageData{Str: "lv-LV"},
				Data:     []byte{},
				Schema:   pb.Schema_GO,
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Malformed language tag",
			params: translate.UploadParams{
				Language: translate.LanguageData{Str: "xyz-ZY-Latn"},
				Schema:   pb.Schema_GO,
				Data: []byte(`{
					"messages": [
						{
							"id": "1",
							"meaning": "When you great someone",
							"message": "hello",
							"translation": "čau",
							"fuzzy": false
						}
					]
				}
				`),
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, contentType, err := attachFile(tt.params.Data, t)
			if !assert.NoError(t, err) {
				return
			}

			req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/%s", addr, tt.params.Language.Str), body)
			if !assert.NoError(t, err) {
				return
			}

			// Add Schema as query parameter, if it is specified
			if tt.params.Schema != pb.Schema_UNSPECIFIED {
				q := req.URL.Query()
				q.Add("schema", tt.params.Schema.String())
				req.URL.RawQuery = q.Encode()
			}

			req.Header.Add("Content-Type", contentType)

			resp, err := http.DefaultClient.Do(req)
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want, resp.StatusCode)
		})
	}
}

func Test_DownloadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	addr := baseAddr + "/v1/files"

	tests := []struct {
		name   string
		params translate.DownloadParams
		want   uint
	}{
		{
			name: "Happy path",
			params: translate.DownloadParams{
				Language: translate.LanguageData{Str: "lv-LV"},
			},
			want: http.StatusOK,
		},
		{
			name:   "Invalid argument",
			params: translate.DownloadParams{},
			want:   http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s", addr, tt.params.Language.Str), nil)
			if !assert.NoError(t, err) {
				return
			}

			// Add Schema as query parameter, if it is specified
			if tt.params.Schema != pb.Schema_UNSPECIFIED {
				q := req.URL.Query()
				q.Add("schema", tt.params.Schema.String())
				req.URL.RawQuery = q.Encode()
			}

			resp, err := http.DefaultClient.Do(req)
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want, resp.StatusCode)
		})
	}
}
