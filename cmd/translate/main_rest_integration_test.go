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
	query.Add("schema", req.GetSchema().String())

	u := url.URL{
		Scheme:   "http",
		Host:     host + ":" + port,
		Path:     fmt.Sprintf("v1/services/%s/files/%s", req.GetServiceId(), req.GetLanguage()),
		RawQuery: query.Encode(),
	}

	body, contentType := attachFile(req.GetData(), t)

	r, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), body)
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
	query.Add("schema", req.GetSchema().String())

	u := url.URL{
		Scheme:   "http",
		Host:     host + ":" + port,
		Path:     fmt.Sprintf("v1/services/%s/files/%s", req.GetServiceId(), req.GetLanguage()),
		RawQuery: query.Encode(),
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
	happyRequest := randUploadTranslationFileReq(t, service.GetId())

	// PUT /v1/services/{service_id}/files
	happyRequestNoLang := &translatev1.UploadTranslationFileRequest{
		ServiceId: service.GetId(),
		// NG Localize has language in the file.
		Data:   randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE, rand.Language()),
		Schema: translatev1.Schema_JSON_NG_LOCALIZE,
	}

	invalidArgumentMissingServiceRequest := randUploadTranslationFileReq(t, service.GetId())
	invalidArgumentMissingServiceRequest.ServiceId = ""

	notFoundServiceIDRequest := randUploadTranslationFileReq(t, gofakeit.UUID())

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
	uploadReq := randUploadTranslationFileReq(t, service.GetId())

	_, err := client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	// Change translation and upload again with the same language and serviceID
	uploadReq.Data = randUploadData(t, uploadReq.GetSchema(), language.MustParse(uploadReq.GetLanguage()))

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

	uploadRequest := randUploadTranslationFileReq(t, service.GetId())

	_, err := client.UploadTranslationFile(ctx, uploadRequest)
	require.NoError(t, err, "create test translation file")

	// Requests

	happyRequest := randDownloadRequest(service.GetId(), uploadRequest.GetLanguage())

	happyReqNoTranslationServiceID := randDownloadRequest(gofakeit.UUID(), uploadRequest.GetLanguage())

	happyReqNoTranslationLanguage := randDownloadRequest(service.GetId(), rand.Language().String())
	// Ensure that the language is not the same as the uploaded one.
	for happyReqNoTranslationLanguage.GetLanguage() == uploadRequest.GetLanguage() {
		happyReqNoTranslationLanguage.Language = rand.Language().String()
	}

	unspecifiedSchemaRequest := randDownloadRequest(service.GetId(), uploadRequest.GetLanguage())
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
			name:         "Happy path no translation with language",
			request:      happyReqNoTranslationLanguage,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Happy path no translation with Service ID",
			request:      happyReqNoTranslationServiceID,
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

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
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
		Path:   "v1/services/" + service.GetId(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(putBodyBytes))
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
		Path:   "v1/services/" + service.GetId(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u.String(), bytes.NewReader(patchBodyBytes))
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
				Path:   "v1/services/" + tt.service.GetId(),
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
				Path:   "v1/services/" + tt.service.GetId(),
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	require.NoError(t, err, "create request")

	resp, err := otelhttp.DefaultClient.Do(req)
	require.NoError(t, err, "do request")

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ------------------Translation------------------

// POST.
func Test_CreateTranslation_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	langs := rand.Languages(4)

	service := createService(ctx, t)
	createTranslation(ctx, t, service.GetId(),
		&translatev1.Translation{Original: true, Language: langs[0].String()})

	serviceWithTranslations := createService(ctx, t)
	uploadReq := randUploadTranslationFileReq(t, serviceWithTranslations.GetId())
	_, err := client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	tests := []struct {
		translation  *translatev1.Translation
		name         string
		serviceID    string
		expectedCode int
	}{
		{
			name:         "Happy path, create translation",
			serviceID:    service.GetId(),
			translation:  randTranslation(t, &translatev1.Translation{Language: langs[1].String()}),
			expectedCode: http.StatusOK,
		},
		{
			name:      "Happy path, empty translation.messages",
			serviceID: service.GetId(),
			translation: &translatev1.Translation{
				Language: langs[2].String(),
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not found, service not found",
			serviceID:    gofakeit.UUID(),
			translation:  randTranslation(t, nil),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Bad request, translation not provided",
			serviceID:    service.GetId(),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "Bad request, translation.language not provided",
			serviceID: service.GetId(),
			translation: &translatev1.Translation{
				Language: "",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "Status conflict, service already has translation for specified language",
			serviceID: serviceWithTranslations.GetId(),
			translation: &translatev1.Translation{
				Language: uploadReq.GetLanguage(),
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:      "Status conflict, service already has original translation",
			serviceID: service.GetId(),
			translation: &translatev1.Translation{
				Original: true,
				Language: langs[3].String(),
			},
			expectedCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.translation)
			require.NoError(t, err, "marshal translation")

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.serviceID + "/translations",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func Test_UpdateTranslation_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)
	langs := rand.Languages(3)

	createTranslation(ctx, t, service.GetId(), &translatev1.Translation{Original: true, Language: langs[0].String()})
	createTranslation(ctx, t, service.GetId(), &translatev1.Translation{Original: false, Language: langs[1].String()})

	happyReq := createUpdateTranslationReq(t, service.GetId(), &translatev1.Translation{Language: langs[1].String()})

	// different language without translation
	notFoundTranslationReq := createUpdateTranslationReq(t,
		service.GetId(), &translatev1.Translation{Language: langs[2].String()})

	notFoundServiceID := createUpdateTranslationReq(t,
		gofakeit.UUID(), &translatev1.Translation{Language: langs[1].String()})

	invalidArgumentNilTranslationReq := &translatev1.UpdateTranslationRequest{ServiceId: service.GetId()}

	invalidArgumentUndTranslationLanguageReq := createUpdateTranslationReq(t, gofakeit.UUID(), nil)
	invalidArgumentUndTranslationLanguageReq.Translation.Language = ""

	originalAlreadyExistsReq := createUpdateTranslationReq(t,
		service.GetId(), &translatev1.Translation{Language: langs[1].String(), Original: true})

	tests := []struct {
		request      *translatev1.UpdateTranslationRequest
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
			request:      notFoundTranslationReq,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Service does not exists",
			request:      notFoundServiceID,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Invalid argument nil translation",
			request:      invalidArgumentNilTranslationReq,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid argument und translation.language",
			request:      invalidArgumentUndTranslationLanguageReq,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Status conflict, service already has original translation",
			request:      originalAlreadyExistsReq,
			expectedCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.request.GetTranslation())
			require.NoError(t, err, "marshal translation")

			language := langs[0].String()

			if tt.request.GetTranslation() != nil {
				language = tt.request.GetTranslation().GetLanguage()
			}

			u := url.URL{
				Scheme: "http",
				Host:   host + ":" + port,
				Path:   "v1/services/" + tt.request.GetServiceId() + "/translations/" + language,
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(body))
			require.NoError(t, err, "create request")

			resp, err := otelhttp.DefaultClient.Do(req)
			require.NoError(t, err, "do request")

			defer resp.Body.Close()

			assert.Equal(t, int(tt.expectedCode), resp.StatusCode)
		})
	}
}

// GET.
func Test_GetTranslations_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	for i := 0; i < gofakeit.IntRange(1, 5); i++ {
		uploadRequest := randUploadTranslationFileReq(t, service.GetId())
		_, err := client.UploadTranslationFile(ctx, uploadRequest)
		require.NoError(t, err, "create test translation file")
	}

	tests := []struct {
		serviceID    string
		name         string
		expectedCode int
	}{
		{
			serviceID:    service.GetId(),
			name:         "Happy Path, get all translations",
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
				Path:   "v1/services/" + tt.serviceID + "/translations",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
