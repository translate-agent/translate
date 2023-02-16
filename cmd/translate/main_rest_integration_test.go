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
	addr := baseAddr + "/v1/files/"

	type params struct {
		Schema string
		URL    string
		Data   []byte
	}

	tests := []struct {
		name   string
		params params
		want   uint
	}{
		{
			name: "Happy Path",
			params: params{
				Schema: "GO",
				URL:    addr + "lv-LV",
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
			name: "Invalid argument",
			params: params{
				Schema: "GO",
				URL:    addr + "lv-LV-asd",
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

			req, err := http.NewRequestWithContext(ctx, "PUT", tt.params.URL, body)
			if !assert.NoError(t, err) {
				return
			}
			req.Header.Add("Content-Type", contentType)

			q := req.URL.Query()
			q.Add("schema", tt.params.Schema)
			req.URL.RawQuery = q.Encode()

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
	addr := baseAddr + "/v1/files/"

	type params struct {
		Schema string
		URL    string
	}

	tests := []struct {
		name   string
		params params
		want   uint
	}{
		{
			name: "Happy path",
			params: params{
				Schema: "GO",
				URL:    addr + "lv-LV",
			},
			want: http.StatusOK,
		},
		{
			name: "Invalid argument",
			params: params{
				Schema: "GO",
				URL:    addr + "lv-LV-asd",
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(ctx, "GET", tt.params.URL, nil)
			if !assert.NoError(t, err) {
				return
			}

			// Add Schema as query parameter
			q := req.URL.Query()
			q.Add("schema", tt.params.Schema)
			req.URL.RawQuery = q.Encode()

			resp, err := http.DefaultClient.Do(req)
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want, resp.StatusCode)
		})
	}
}
