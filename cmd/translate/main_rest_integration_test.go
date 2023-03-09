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
		name     string
		input    params
		expected uint
	}{
		{
			name: "Happy Path",
			input: params{
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
				}`),
			},
			expected: http.StatusOK,
		},
		{
			name: "Invalid argument",
			input: params{
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
				}`),
			},
			expected: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			query := url.Values{}
			query.Add("schema", tt.input.fileSchema)

			u := url.URL{
				Scheme:   "http",
				Host:     host + ":" + port,
				Path:     tt.input.path,
				RawQuery: query.Encode(),
			}

			body, contentType, err := attachFile(tt.input.data, t)
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

			assert.EqualValues(t, tt.expected, resp.StatusCode)
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
		name     string
		input    params
		expected uint
	}{
		{
			name: "Happy path",
			input: params{
				fileSchema: "GO",
				path:       "v1/files/lv-LV",
			},
			expected: http.StatusOK,
		},
		{
			name: "Invalid argument",
			input: params{
				fileSchema: "GO",
				path:       "v1/files/lv-LV-asd",
			},
			expected: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			query := url.Values{}
			query.Add("schema", tt.input.fileSchema)

			u := url.URL{
				Scheme:   "http",
				Host:     host + ":" + port,
				Path:     tt.input.path,
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

			assert.EqualValues(t, tt.expected, resp.StatusCode)
		})
	}
}
