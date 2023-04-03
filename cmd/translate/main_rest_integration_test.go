//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

// TODO Currently, we manually create requests for the REST API.
// We could use a client generated from the OpenAPI specification to simplify testing and integration.

// -------------Translation File-------------.

func attachFile(text []byte, t *testing.T) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", "test.json")
	require.NoError(t, err, "create form file")

	_, err = part.Write(text)
	require.NoError(t, err, "write to part")

	return body, writer.FormDataContentType()
}

func gRPCUploadFileToRESTReq(ctx context.Context, t *testing.T, req *translatev1.UploadTranslationFileRequest) *http.Request {
	t.Helper()

	query := url.Values{}
	query.Add("schema", req.Schema.String())

	u := url.URL{
		Scheme:   "http",
		Host:     host + ":" + port,
		Path:     fmt.Sprintf("v1/services/%s/files/%s", req.ServiceId, req.Language),
		RawQuery: query.Encode(),
	}

	body, contentType := attachFile(req.Data, t)

	r, err := http.NewRequestWithContext(ctx, "PUT", u.String(), body)
	require.NoError(t, err, "create request")

	r.Header.Add("Content-Type", contentType)

	return r
}

func gRPCDownloadFileToRESTReq(ctx context.Context, t *testing.T, req *translatev1.DownloadTranslationFileRequest) *http.Request {
	t.Helper()

	query := url.Values{}
	query.Add("schema", req.Schema.String())

	u := url.URL{
		Scheme:   "http",
		Host:     host + ":" + port,
		Path:     fmt.Sprintf("v1/services/%s/files/%s", req.ServiceId, req.Language),
		RawQuery: query.Encode(),
	}

	r, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	require.NoError(t, err, "create request")

	return r
}

func Test_UploadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	// Requests

	happyRequest := randUploadRequest(t, service.Id)

	invalidArgumentRequest := randUploadRequest(t, service.Id)
	invalidArgumentRequest.Language = ""

	notFoundServiceIDRequest := randUploadRequest(t, gofakeit.UUID())

	tests := []struct {
		name     string
		request  *http.Request
		expected uint
	}{
		{
			name:     "Happy Path",
			request:  gRPCUploadFileToRESTReq(ctx, t, happyRequest),
			expected: http.StatusOK,
		},
		{
			name:     "Invalid argument missing language",
			request:  gRPCUploadFileToRESTReq(ctx, t, invalidArgumentRequest),
			expected: http.StatusBadRequest,
		},
		{
			name:     "Not found service ID",
			request:  gRPCUploadFileToRESTReq(ctx, t, notFoundServiceIDRequest),
			expected: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := http.DefaultClient.Do(tt.request)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()
			respBody, _ := ioutil.ReadAll(resp.Body)

			actual := resp.StatusCode
			assert.Equal(t, int(tt.expected), actual, "body: %s", respBody)
		})
	}
}

func Test_UploadTranslationFileDifferentLanguages_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	uploadRequest := randUploadRequest(t, service.Id)

	for i := 0; i < 3; i++ {
		uploadRequest.Language = gofakeit.LanguageBCP()

		req := gRPCUploadFileToRESTReq(ctx, t, uploadRequest)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "do request")

		respBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		actual := resp.StatusCode
		expected := http.StatusOK

		require.Equal(t, int(expected), actual, "body: %s", respBody)
	}
}

func Test_UploadTranslationFileUpdateFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	// Upload initial
	uploadReq := randUploadRequest(t, service.Id)

	_, err = client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	// Change messages and upload again with the same language and serviceID
	uploadReq.Data, _ = randUploadData(t, uploadReq.Schema)

	req := gRPCUploadFileToRESTReq(ctx, t, uploadReq)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	actual := resp.StatusCode
	expected := http.StatusOK

	require.Equal(t, int(expected), actual, "body: %s", respBody)
}

func Test_DownloadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	uploadRequest := randUploadRequest(t, service.Id)

	_, err = client.UploadTranslationFile(ctx, uploadRequest)
	require.NoError(t, err, "create test translation file")

	// Requests

	happyRequest := randDownloadRequest(service.Id, uploadRequest.Language)

	invalidArgumentRequest := randDownloadRequest(service.Id, uploadRequest.Language)
	invalidArgumentRequest.Schema = translatev1.Schema_UNSPECIFIED

	notFoundIDRequest := randDownloadRequest(gofakeit.UUID(), uploadRequest.Language)

	notFoundLanguageRequest := randDownloadRequest(service.Id, gofakeit.LanguageBCP())

	tests := []struct {
		name     string
		request  *http.Request
		expected uint
	}{
		{
			name:     "Happy path",
			request:  gRPCDownloadFileToRESTReq(ctx, t, happyRequest),
			expected: http.StatusOK,
		},
		{
			name:     "Invalid argument unspecified schema",
			request:  gRPCDownloadFileToRESTReq(ctx, t, invalidArgumentRequest),
			expected: http.StatusBadRequest,
		},
		{
			name:     "Not found ID",
			request:  gRPCDownloadFileToRESTReq(ctx, t, notFoundIDRequest),
			expected: http.StatusNotFound,
		},
		{
			name:     "Not found language",
			request:  gRPCDownloadFileToRESTReq(ctx, t, notFoundLanguageRequest),
			expected: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := http.DefaultClient.Do(tt.request)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()
			respBody, _ := ioutil.ReadAll(resp.Body)

			actual := resp.StatusCode
			assert.Equal(t, int(tt.expected), actual, "body: %s", respBody)
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
