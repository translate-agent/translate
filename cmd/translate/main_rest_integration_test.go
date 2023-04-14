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

	service := prepareService(ctx, t)

	// Requests

	happyRequest := randUploadRequest(t, service.Id)

	missingLanguageRequest := randUploadRequest(t, service.Id)
	missingLanguageRequest.Language = ""

	notFoundServiceIDRequest := randUploadRequest(t, gofakeit.UUID())

	tests := []struct {
		name         string
		request      *http.Request
		expectedCode uint
	}{
		{
			name:         "Happy Path",
			request:      gRPCUploadFileToRESTReq(ctx, t, happyRequest),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Bad request missing language",
			request:      gRPCUploadFileToRESTReq(ctx, t, missingLanguageRequest),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Not found service ID",
			request:      gRPCUploadFileToRESTReq(ctx, t, notFoundServiceIDRequest),
			expectedCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := http.DefaultClient.Do(tt.request)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			// Read the response to give error message on failure
			respBody, _ := ioutil.ReadAll(resp.Body)

			actualCode := resp.StatusCode
			assert.Equal(t, int(tt.expectedCode), actualCode, "body: %s", string(respBody))
		})
	}
}

func Test_UploadTranslationFileDifferentLanguages_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := prepareService(ctx, t)

	uploadRequest := randUploadRequest(t, service.Id)

	for i := 0; i < 3; i++ {
		uploadRequest.Language = gofakeit.LanguageBCP()

		req := gRPCUploadFileToRESTReq(ctx, t, uploadRequest)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "do request")

		respBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		actualCode := resp.StatusCode
		expectedCode := http.StatusOK

		require.Equal(t, int(expectedCode), actualCode, "body: %s", string(respBody))
	}
}

func Test_UploadTranslationFileUpdateFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := prepareService(ctx, t)

	// Upload initial
	uploadReq := randUploadRequest(t, service.Id)

	_, err := client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	// Change messages and upload again with the same language and serviceID
	uploadReq.Data, _ = randUploadData(t, uploadReq.Schema)

	req := gRPCUploadFileToRESTReq(ctx, t, uploadReq)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	actualCode := resp.StatusCode
	expectedCode := http.StatusOK

	assert.Equal(t, int(expectedCode), actualCode, "body: %s", string(respBody))
}

func Test_DownloadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Prepare

	service := prepareService(ctx, t)

	uploadRequest := randUploadRequest(t, service.Id)

	_, err := client.UploadTranslationFile(ctx, uploadRequest)
	require.NoError(t, err, "create test translation file")

	// Requests

	happyRequest := randDownloadRequest(service.Id, uploadRequest.Language)

	unspecifiedSchemaRequest := randDownloadRequest(service.Id, uploadRequest.Language)
	unspecifiedSchemaRequest.Schema = translatev1.Schema_UNSPECIFIED

	notFoundServiceIDRequest := randDownloadRequest(gofakeit.UUID(), uploadRequest.Language)

	notFoundLanguageRequest := randDownloadRequest(service.Id, gofakeit.LanguageBCP())
	// Ensure that the language is not the same as the uploaded one.
	for notFoundLanguageRequest.Language == uploadRequest.Language {
		notFoundLanguageRequest.Language = gofakeit.LanguageBCP()
	}

	tests := []struct {
		name         string
		request      *http.Request
		expectedCode uint
	}{
		{
			name:         "Happy path",
			request:      gRPCDownloadFileToRESTReq(ctx, t, happyRequest),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Bad request unspecified schema",
			request:      gRPCDownloadFileToRESTReq(ctx, t, unspecifiedSchemaRequest),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Not found Service ID",
			request:      gRPCDownloadFileToRESTReq(ctx, t, notFoundServiceIDRequest),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Not found language",
			request:      gRPCDownloadFileToRESTReq(ctx, t, notFoundLanguageRequest),
			expectedCode: http.StatusNotFound,
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

			actualCode := resp.StatusCode
			assert.Equal(t, int(tt.expectedCode), actualCode, "body: %s", string(respBody))
		})
	}
}

// ------------------Service------------------

// POST.
func Test_CreateService_REST(t *testing.T) {
	t.Parallel()

	serviceWithID := randService()

	serviceWithoutID := randService()
	serviceWithoutID.Id = ""

	serviceMalformedID := randService()
	serviceMalformedID.Id += "_FAIL"

	tests := []struct {
		service      *translatev1.Service
		name         string
		expectedCode uint
	}{
		{
			name:         "Happy path With ID",
			service:      serviceWithID,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Happy path Without ID",
			service:      serviceWithoutID,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid argument malformed ID",
			service:      serviceMalformedID,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.service)
			require.NoError(t, err, "marshal service")

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services",
			}

			req, err := http.NewRequestWithContext(context.Background(), "POST", u.String(), bytes.NewBuffer(body))
			require.NoError(t, err, "create request")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			actualCode := resp.StatusCode

			assert.Equal(t, int(tt.expectedCode), actualCode)
		})
	}
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
	require.NoError(t, err, "prepare test service")

	putBody := restUpdateBody{Name: gofakeit.FirstName()}

	putBodyBytes, err := json.Marshal(putBody)
	require.NoError(t, err, "marshal put body")

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services/" + service.Id,
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", u.String(), bytes.NewBuffer(putBodyBytes))
	require.NoError(t, err, "create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	actualCode := resp.StatusCode
	expectedCode := http.StatusOK

	assert.Equal(t, expectedCode, actualCode)
}

// PATCH.
func Test_UpdateServiceSpecificField_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "Prepare test service")

	patchBody := restUpdateBody{Name: gofakeit.FirstName()}

	patchBodyBytes, err := json.Marshal(patchBody)
	require.NoError(t, err, "marshal patch body")

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services/" + service.Id,
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", u.String(), bytes.NewReader(patchBodyBytes))
	require.NoError(t, err, "create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	actualCode := resp.StatusCode
	expectedCode := http.StatusOK

	assert.Equal(t, expectedCode, actualCode)
}

// GET.
func Test_GetService_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "Prepare test service")

	tests := []struct {
		service      *translatev1.Service
		name         string
		expectedCode uint
	}{
		{
			service:      service,
			name:         "Happy Path",
			expectedCode: http.StatusOK,
		},
		{
			service:      randService(),
			name:         "Not Found",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.service.Id,
			}

			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			require.NoError(t, err, "create request")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			actualCode := resp.StatusCode
			assert.Equal(t, int(tt.expectedCode), actualCode)
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
	require.NoError(t, err, "Prepare test service")

	tests := []struct {
		service      *translatev1.Service
		name         string
		expectedCode uint
	}{
		{
			service:      service,
			name:         "Happy Path",
			expectedCode: http.StatusOK,
		},
		{
			service:      randService(),
			name:         "Not Found",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.service.Id,
			}

			req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
			require.NoError(t, err, "create request")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			actualCode := resp.StatusCode
			assert.Equal(t, int(tt.expectedCode), actualCode)
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
	require.NoError(t, err, "create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	actualCode := resp.StatusCode
	expectedCode := http.StatusOK

	assert.Equal(t, expectedCode, actualCode)
}
