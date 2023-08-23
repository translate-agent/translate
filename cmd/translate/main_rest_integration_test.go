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
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/text/language"
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

	// PUT /v1/services/{service_id}/files/{language}
	happyRequest := randUploadRequest(t, service.Id)

	// PUT /v1/services/{service_id}/files
	happyRequestNoLang := &translatev1.UploadTranslationFileRequest{
		ServiceId: service.Id,
		// NG Localize has language in the file.
		Data:   randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE, rand.Language()),
		Schema: translatev1.Schema_JSON_NG_LOCALIZE,
	}

	invalidArgumentMissingServiceRequest := randUploadRequest(t, service.Id)
	invalidArgumentMissingServiceRequest.ServiceId = ""

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
			name:         "Happy Path no language in path",
			request:      happyRequestNoLang,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Bad request missing service_id",
			request:      invalidArgumentMissingServiceRequest,
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
	uploadReq.Data = randUploadData(t, uploadReq.Schema, language.MustParse(uploadReq.Language))

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

	happyReqNoMessagesLanguage := randDownloadRequest(service.Id, rand.Language().String())
	// Ensure that the language is not the same as the uploaded one.
	for happyReqNoMessagesLanguage.Language == uploadRequest.Language {
		happyReqNoMessagesLanguage.Language = rand.Language().String()
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
			name:         "Happy path no messages with language",
			request:      happyReqNoMessagesLanguage,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Happy path no messages with Service ID",
			request:      happyReqNoMessagesServiceID,
			expectedCode: http.StatusOK,
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

// POST.
func Test_CreateMessages_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare

	service := createService(ctx, t)
	langs := rand.Languages(2)

	serviceWithMsgs := createService(ctx, t)
	uploadReq := randUploadRequest(t, serviceWithMsgs.Id)
	_, err := client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	tests := []struct {
		messages     *translatev1.Messages
		name         string
		serviceID    string
		expectedCode int
	}{
		{
			name:         "Happy path, create messages",
			serviceID:    service.Id,
			messages:     randMessages(t, &translatev1.Messages{Language: langs[0].String()}),
			expectedCode: http.StatusOK,
		},
		{
			name:      "Happy path, empty messages.messages",
			serviceID: service.Id,
			messages: &translatev1.Messages{
				Language: langs[1].String(),
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not found, service not found",
			serviceID:    gofakeit.UUID(),
			messages:     randMessages(t, nil),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Bad request, messages not provided",
			serviceID:    service.Id,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "Bad request, messages.language not provided",
			serviceID: service.Id,
			messages: &translatev1.Messages{
				Language: "",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "Status conflict, service already has messages for specified language",
			serviceID: serviceWithMsgs.Id,
			messages: &translatev1.Messages{
				Language: uploadReq.Language,
			},
			expectedCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.messages)
			require.NoError(t, err, "marshal messages")

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.serviceID + "/messages",
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

func Test_UpdateMessages_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)
	langs := rand.Languages(2)

	_, err := client.CreateMessages(ctx, &translatev1.CreateMessagesRequest{
		ServiceId: service.Id,
		Messages:  randMessages(t, &translatev1.Messages{Language: langs[0].String()}),
	})
	require.NoError(t, err, "create test messages")

	// helper for update request generation
	randUpdateMessageReq := func(lang string) *translatev1.UpdateMessagesRequest {
		if lang == "" {
			lang = rand.Language().String()
		}

		return &translatev1.UpdateMessagesRequest{
			ServiceId: service.Id,
			Messages:  randMessages(t, &translatev1.Messages{Language: lang}),
		}
	}

	happyReq := randUpdateMessageReq(langs[0].String()) // uploaded messages language

	notFoundMessagesReq := randUpdateMessageReq(langs[1].String()) // different language without messages

	notFoundServiceID := randUpdateMessageReq("")
	notFoundServiceID.ServiceId = gofakeit.UUID()

	invalidArgumentNilMessagesReq := randUpdateMessageReq("")
	invalidArgumentNilMessagesReq.Messages = nil

	invalidArgumentUndMessagesLanguageReq := randUpdateMessageReq("")
	invalidArgumentUndMessagesLanguageReq.Messages.Language = ""

	tests := []struct {
		request      *translatev1.UpdateMessagesRequest
		name         string
		expectedCode uint
	}{
		{
			name:         "Happy Path",
			request:      happyReq,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Message does not exists",
			request:      notFoundMessagesReq,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Service does not exists",
			request:      notFoundServiceID,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Invalid argument nil messages",
			request:      invalidArgumentNilMessagesReq,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid argument und messages.language",
			request:      invalidArgumentUndMessagesLanguageReq,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.request.Messages)
			require.NoError(t, err, "marshal messages")

			language := langs[0].String()

			if tt.request.Messages != nil {
				language = tt.request.Messages.Language
			}

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.request.ServiceId + "/messages/" + language,
			}

			req, err := http.NewRequestWithContext(ctx, "PUT", u.String(), bytes.NewBuffer(body))
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, int(tt.expectedCode), resp.StatusCode)
		})
	}
}

// GET.
func Test_GetMessages_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	for i := 0; i < gofakeit.IntRange(1, 5); i++ {
		uploadRequest := randUploadRequest(t, service.Id)
		_, err := client.UploadTranslationFile(ctx, uploadRequest)
		require.NoError(t, err, "create test translation file")
	}

	tests := []struct {
		serviceID    string
		name         string
		expectedCode int
	}{
		{
			serviceID:    service.Id,
			name:         "Happy Path, get all messages",
			expectedCode: http.StatusOK,
		},
		{
			serviceID:    gofakeit.UUID(),
			name:         "Happy path, service doesn't exist",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Bad request, ServiceID not provided",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.serviceID + "/messages",
			}

			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			if err == nil {
				require.NotEmpty(t, resp.Body)
			}

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}
