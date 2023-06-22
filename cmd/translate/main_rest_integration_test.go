//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TODO Currently, we manually create requests for the REST API.
// We could use a client generated from the OpenAPI specification to simplify testing and integration.

// -------------Translation File-------------.

func attachFile(text []byte, t *testing.T) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", "test.json")
	require.NoError(t, err, "create form file")

	_, err = part.Write(text)
	require.NoError(t, err, "write to part")

	return &body, writer.FormDataContentType()
}

func gRPCUploadFileToRESTReq(
	ctx context.Context,
	t *testing.T,
	req *translatev1.UploadTranslationFileRequest,
) *http.Request {
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

func gRPCDownloadFileToRESTReq(
	ctx context.Context,
	t *testing.T,
	req *translatev1.DownloadTranslationFileRequest,
) *http.Request {
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

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	// Requests

	happyRequest := randUploadRequest(t, service.Id)

	missingLanguageRequest := randUploadRequest(t, service.Id)
	missingLanguageRequest.Language = ""

	notFoundServiceIDRequest := randUploadRequest(t, gofakeit.UUID())

	tests := []struct {
		request      *translatev1.UploadTranslationFileRequest
		name         string
		expectedCode int
	}{
		{
			name:         "Happy Path",
			request:      happyRequest,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Bad request missing language",
			request:      missingLanguageRequest,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Not found service ID",
			request:      notFoundServiceIDRequest,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := otelhttp.DefaultClient.Do(gRPCUploadFileToRESTReq(ctx, t, tt.request))
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			// Read the response to give error message on failure
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "body: %s", string(respBody))
		})
	}
}

func Test_UploadTranslationFileUpdateFile_REST(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	// Upload initial
	uploadReq := randUploadRequest(t, service.Id)

	_, err := client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	// Change messages and upload again with the same language and serviceID
	uploadReq.Data, _ = randUploadData(t, uploadReq.Schema)

	resp, err := otelhttp.DefaultClient.Do(gRPCUploadFileToRESTReq(ctx, t, uploadReq))
	require.NoError(t, err, "do request")

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", string(respBody))
}

func Test_DownloadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	uploadRequest := randUploadRequest(t, service.Id)

	_, err := client.UploadTranslationFile(ctx, uploadRequest)
	require.NoError(t, err, "create test translation file")

	// Requests

	happyRequest := randDownloadRequest(service.Id, uploadRequest.Language)

	happyReqNoMessagesServiceID := randDownloadRequest(gofakeit.UUID(), uploadRequest.Language)

	happyReqNoMessagesLanguage := randDownloadRequest(service.Id, gofakeit.LanguageBCP())
	// Ensure that the language is not the same as the uploaded one.
	for happyReqNoMessagesLanguage.Language == uploadRequest.Language {
		happyReqNoMessagesLanguage.Language = gofakeit.LanguageBCP()
	}

	unspecifiedSchemaRequest := randDownloadRequest(service.Id, uploadRequest.Language)
	unspecifiedSchemaRequest.Schema = translatev1.Schema_UNSPECIFIED

	tests := []struct {
		request      *translatev1.DownloadTranslationFileRequest
		name         string
		expectedCode int
	}{
		{
			name:         "Happy path",
			request:      happyRequest,
			expectedCode: http.StatusOK,
		},

		{
			name:         "File not found, no messages with language",
			request:      happyReqNoMessagesLanguage,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "File not found, no messages with Service ID",
			request:      happyReqNoMessagesServiceID,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Bad request unspecified schema",
			request:      unspecifiedSchemaRequest,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := otelhttp.DefaultClient.Do(gRPCDownloadFileToRESTReq(ctx, t, tt.request))
			require.NoError(t, err, "do request")

			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "body: %s", string(respBody))
		})
	}
}

// ------------------Service------------------

// POST.
func Test_CreateService_REST(t *testing.T) {
	t.Parallel()

	_, subtest := testutil.Trace(t)

	// Prepare
	serviceWithID := randService()

	serviceWithoutID := randService()
	serviceWithoutID.Id = ""

	serviceMalformedID := randService()
	serviceMalformedID.Id += "_FAIL"

	tests := []struct {
		service      *translatev1.Service
		name         string
		expectedCode int
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
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.service)
			require.NoError(t, err, "marshal service")

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services",
			}

			req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewBuffer(body))
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

type restUpdateBody struct {
	Name string `json:"name,omitempty"`
}

// PUT.
func Test_UpdateServiceAllFields_REST(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
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

	resp, err := otelhttp.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// PATCH.
func Test_UpdateServiceSpecificField_REST(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
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

	resp, err := otelhttp.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// GET.
func Test_GetService_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "Prepare test service")

	tests := []struct {
		service      *translatev1.Service
		name         string
		expectedCode int
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
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.service.Id,
			}

			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

// DELETE.
func Test_DeleteService_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "Prepare test service")

	tests := []struct {
		service      *translatev1.Service
		name         string
		expectedCode int
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
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.service.Id,
			}

			req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

// GET (list).
func Test_ListServices_REST(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	u := url.URL{
		Scheme: "http",
		Host:   host + ":" + port,
		Path:   "v1/services",
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	require.NoError(t, err, "create request")

	resp, err := otelhttp.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ------------------Messages------------------

// GET.
func Test_GetMessages_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	n := gofakeit.IntRange(1, 5)
	langTags := make([]string, 0, n)

	for i := 0; i < n; i++ {
		uploadRequest := randUploadRequest(t, service.Id)
		_, err := client.UploadTranslationFile(ctx, uploadRequest)
		require.NoError(t, err, "create test translation file")

		langTags = append(langTags, uploadRequest.Language)
	}

	// Requests

	langTagsQuery := url.Values{}
	langTagsQuery.Add("languages", strings.Join(langTags, ","))

	tests := []struct {
		name         string
		query        string
		expectedCode int
	}{
		{
			name:         "Happy Path, get all languages",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Happy Path, filter languages",
			query:        langTagsQuery.Encode(),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Bad request, filter by unknown language format",
			query:        gofakeit.BeerName(),
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme:   "http",
				Host:     host + ":" + port,
				Path:     "v1/services/" + service.Id + "/messages",
				RawQuery: tt.query,
			}

			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}
