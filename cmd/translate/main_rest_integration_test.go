package main

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	type params struct {
		fileSchema string
		path       string
		data       []byte
	}

	tests := []struct {
		name   string
		params params
		want   uint
	}{
		{
			name: "Happy Path",
			params: params{
				fileSchema: "GO",
				path:       "v1/files/lv-LV",
				data: []byte(`{
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
				fileSchema: "GO",
				path:       "v1/files/lv-LV-asd",
				data: []byte(`{
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

			query := url.Values{}
			query.Add("schema", tt.params.fileSchema)

			u := url.URL{
				Scheme:   "http",
				Host:     host + ":" + port,
				Path:     tt.params.path,
				RawQuery: query.Encode(),
			}

			body, contentType, err := attachFile(tt.params.data, t)
			if !assert.NoError(t, err) {
				return
			}

			req, err := http.NewRequestWithContext(context.Background(), "PUT", u.String(), body)
			if !assert.NoError(t, err) {
				return
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

	type params struct {
		fileSchema string
		path       string
	}

	tests := []struct {
		name   string
		params params
		want   uint
	}{
		{
			name: "Happy path",
			params: params{
				fileSchema: "GO",
				path:       "v1/files/lv-LV",
			},
			want: http.StatusOK,
		},
		{
			name: "Invalid argument",
			params: params{
				fileSchema: "GO",
				path:       "v1/files/lv-LV-asd",
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			query := url.Values{}
			query.Add("schema", tt.params.fileSchema)

			u := url.URL{
				Scheme:   "http",
				Host:     host + ":" + port,
				Path:     tt.params.path,
				RawQuery: query.Encode(),
			}

			req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), nil)
			if !assert.NoError(t, err) {
				return
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
