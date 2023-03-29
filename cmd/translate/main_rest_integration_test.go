//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

// -------------Translation File-------------.

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
							"meaning": "When you greet someone",
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
							"meaning": "When you greet someone",
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

			actual := resp.StatusCode
			assert.Equal(t, int(tt.expected), actual)
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

			actual := resp.StatusCode
			assert.Equal(t, int(tt.expected), actual)
		})
	}
}

// ------------------Service------------------

// POST.
func Test_CreateService_REST(t *testing.T) {
	t.Parallel()

	service := randService()

	body, err := json.Marshal(service)
	if !assert.NoError(t, err) {
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services",
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", u.String(), bytes.NewBuffer(body))
	if !assert.NoError(t, err) {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	actual := resp.StatusCode
	expected := http.StatusOK

	assert.Equal(t, expected, actual)
}

type restUpdateBody struct {
	Name string `json:"name,omitempty"`
}

// PUT.
func Test_UpdateServiceAllFields_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	putBody := restUpdateBody{Name: gofakeit.FirstName()}

	putBodyBytes, err := json.Marshal(putBody)
	if !assert.NoError(t, err) {
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services/" + service.Id,
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", u.String(), bytes.NewBuffer(putBodyBytes))
	if !assert.NoError(t, err) {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	actual := resp.StatusCode
	expected := http.StatusOK

	assert.Equal(t, expected, actual)
}

// PATCH.
func Test_UpdateServiceSpecificField_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	patchBody := restUpdateBody{Name: gofakeit.FirstName()}

	patchBodyBytes, err := json.Marshal(patchBody)
	if !assert.NoError(t, err) {
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services/" + service.Id,
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", u.String(), bytes.NewReader(patchBodyBytes))
	if !assert.NoError(t, err) {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	actual := resp.StatusCode
	expected := http.StatusOK

	assert.Equal(t, expected, actual)
}

// GET.
func Test_GetService_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	tests := []struct {
		serviceID string
		name      string
		expected  uint
	}{
		{
			serviceID: service.Id,
			name:      "Happy Path",
			expected:  http.StatusOK,
		},
		{
			serviceID: gofakeit.UUID(),
			name:      "Not Found",
			expected:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.serviceID,
			}

			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			if !assert.NoError(t, err) {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			actual := resp.StatusCode
			assert.Equal(t, int(tt.expected), actual)
		})
	}
}

// DELETE.
func Test_DeleteService_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	tests := []struct {
		serviceID string
		name      string
		expected  uint
	}{
		{
			serviceID: service.Id,
			name:      "Happy Path",
			expected:  http.StatusOK,
		},
		{
			serviceID: gofakeit.UUID(),
			name:      "Not Found",
			expected:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.serviceID,
			}

			req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
			if !assert.NoError(t, err) {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			actual := resp.StatusCode
			assert.Equal(t, int(tt.expected), actual)
		})
	}
}

// GET (list).
func Test_ListServices_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services",
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if !assert.NoError(t, err) {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	actual := resp.StatusCode
	expected := http.StatusOK

	assert.Equal(t, expected, actual)
}
