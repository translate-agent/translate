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
	addr := "http://localhost:8080/v1/files"

	type args struct {
		language string
		text     []byte
	}

	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "Happy Path",
			args: args{
				language: "lv-lv",
				text: []byte(`{
					"messages":[
						 {
								"id":"1",
								"meaning":"When you great someone",
								"message":"hello",
								"translation":"čau",
								"fuzzy":false
						 }
					]
			 }`),
			},
			want: http.StatusOK,
		},
		{
			name: "Missing language",
			args: args{
				text: []byte(`{
					"messages":[
						 {
								"id":"1",
								"meaning":"When you great someone",
								"message":"hello",
								"translation":"čau",
								"fuzzy":false
						 }
					]
			 }`),
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Missing file",
			args: args{
				language: "lv-lv",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Invalid language",
			args: args{
				language: "xyz-ZY-Latn",
				text: []byte(`{
					"messages":[
						 {
								"id":"1",
								"meaning":"When you great someone",
								"message":"hello",
								"translation":"čau",
								"fuzzy":false
						 }
					]
			 }`),
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, contentType, err := attachFile(tt.args.text, t)
			if !assert.NoError(t, err) {
				return
			}

			req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/%s", addr, tt.args.language), body)
			if !assert.NoError(t, err) {
				return
			}

			req.Header.Add("Content-Type", contentType)
			client := &http.Client{}
			resp, err := client.Do(req)

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
