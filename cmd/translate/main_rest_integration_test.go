//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/expect"
	"go.expect.digital/translate/pkg/testutil/rand"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/text/language"
	"google.golang.org/genproto/protobuf/field_mask"
)

// TODO: Currently, we manually create requests for the REST API.
// We could use a client generated from the OpenAPI specification to simplify testing and integration.

// TODO: DRY: Tests are the same with gRPC tests. We could combine them to avoid duplication.

// -------------Translation File-------------.

func attachFile(text []byte, t *testing.T) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", "test.json")
	expect.NoError(t, err)

	_, err = part.Write(text)
	expect.NoError(t, err)

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
		Host:     net.JoinHostPort(host, port),
		Path:     fmt.Sprintf("v1/services/%s/files/%s", req.GetServiceId(), req.GetLanguage()),
		RawQuery: query.Encode(),
	}

	body, contentType := attachFile(req.GetData(), t)

	r, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), body)
	expect.NoError(t, err)

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
		Host:     net.JoinHostPort(host, port),
		Path:     fmt.Sprintf("v1/services/%s/files/%s", req.GetServiceId(), req.GetLanguage()),
		RawQuery: query.Encode(),
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	expect.NoError(t, err)

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
		Data:   randUploadData(t, rand.Language()),
		Schema: translatev1.Schema_PO,
	}

	invalidArgumentMissingServiceRequest := randUploadTranslationFileReq(t, service.GetId())
	invalidArgumentMissingServiceRequest.ServiceId = ""

	notFoundServiceIDRequest := randUploadTranslationFileReq(t, gofakeit.UUID())

	tests := []struct {
		request  *translatev1.UploadTranslationFileRequest
		name     string
		wantCode int
	}{
		{
			name:     "Happy Path",
			request:  happyRequest,
			wantCode: http.StatusOK,
		},
		{
			name:     "Happy Path no language in path",
			request:  happyRequestNoLang,
			wantCode: http.StatusOK,
		},
		{
			name:     "Bad request missing service_id",
			request:  invalidArgumentMissingServiceRequest,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Not found service ID",
			request:  notFoundServiceIDRequest,
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := otelhttp.DefaultClient.Do(gRPCUploadFileToRESTReq(ctx, t, tt.request))
			expect.NoError(t, err)

			defer resp.Body.Close()

			// Read the response to give error message on failure
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.wantCode, resp.StatusCode, "body: %s", string(respBody))
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
	expect.NoError(t, err)

	// Change translation and upload again with the same language and serviceID
	uploadReq.Data = randUploadData(t, language.MustParse(uploadReq.GetLanguage()))

	resp, err := otelhttp.DefaultClient.Do(gRPCUploadFileToRESTReq(ctx, t, uploadReq))
	expect.NoError(t, err)

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
	expect.NoError(t, err)

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
		request  *translatev1.DownloadTranslationFileRequest
		name     string
		wantcode int
	}{
		{
			name:     "Happy path",
			request:  happyRequest,
			wantcode: http.StatusOK,
		},

		{
			name:     "Happy path no translation with language",
			request:  happyReqNoTranslationLanguage,
			wantcode: http.StatusOK,
		},
		{
			name:     "Happy path no translation with Service ID",
			request:  happyReqNoTranslationServiceID,
			wantcode: http.StatusOK,
		},
		{
			name:     "Bad request unspecified schema",
			request:  unspecifiedSchemaRequest,
			wantcode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := otelhttp.DefaultClient.Do(gRPCDownloadFileToRESTReq(ctx, t, tt.request))
			expect.NoError(t, err)

			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.wantcode, resp.StatusCode, "body: %s", string(respBody))
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
		service  *translatev1.Service
		name     string
		wantCode int
	}{
		{
			name:     "Happy path With ID",
			service:  serviceWithID,
			wantCode: http.StatusOK,
		},
		{
			name:     "Happy path Without ID",
			service:  serviceWithoutID,
			wantCode: http.StatusOK,
		},
		{
			name:     "Invalid argument malformed ID",
			service:  serviceMalformedID,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.service)
			expect.NoError(t, err)

			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
			expect.NoError(t, err)

			resp, err := otelhttp.DefaultClient.Do(req)
			expect.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
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
	expect.NoError(t, err)

	putBody := restUpdateBody{Name: gofakeit.FirstName()}

	putBodyBytes, err := json.Marshal(putBody)
	expect.NoError(t, err)

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   "v1/services/" + service.GetId(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(putBodyBytes))
	expect.NoError(t, err)

	resp, err := otelhttp.DefaultClient.Do(req)
	expect.NoError(t, err)

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
	expect.NoError(t, err)

	patchBody := restUpdateBody{Name: gofakeit.FirstName()}

	patchBodyBytes, err := json.Marshal(patchBody)
	expect.NoError(t, err)

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   "v1/services/" + service.GetId(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u.String(), bytes.NewReader(patchBodyBytes))
	expect.NoError(t, err)

	resp, err := otelhttp.DefaultClient.Do(req)
	expect.NoError(t, err)

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
	expect.NoError(t, err)

	tests := []struct {
		service  *translatev1.Service
		name     string
		wantCode int
	}{
		{
			service:  service,
			name:     "Happy Path",
			wantCode: http.StatusOK,
		},
		{
			service:  randService(),
			name:     "Not Found",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + tt.service.GetId(),
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			expect.NoError(t, err)

			resp, err := otelhttp.DefaultClient.Do(req)
			expect.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
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
	expect.NoError(t, err)

	tests := []struct {
		service  *translatev1.Service
		name     string
		wantCode int
	}{
		{
			service:  service,
			name:     "Happy Path",
			wantCode: http.StatusOK,
		},
		{
			service:  randService(),
			name:     "Not Found",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + tt.service.GetId(),
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
			expect.NoError(t, err)

			resp, err := otelhttp.DefaultClient.Do(req)
			expect.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
		})
	}
}

// GET (list).
func Test_ListServices_REST(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   "v1/services",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	expect.NoError(t, err)

	resp, err := otelhttp.DefaultClient.Do(req)
	expect.NoError(t, err)

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
	expect.NoError(t, err)

	tests := []struct {
		translation *translatev1.Translation
		name        string
		serviceID   string
		wantCode    int
	}{
		{
			name:        "Happy path, create translation",
			serviceID:   service.GetId(),
			translation: randTranslation(t, &translatev1.Translation{Language: langs[1].String()}),
			wantCode:    http.StatusOK,
		},
		{
			name:      "Happy path, empty translation.messages",
			serviceID: service.GetId(),
			translation: &translatev1.Translation{
				Language: langs[2].String(),
			},
			wantCode: http.StatusOK,
		},
		{
			name:        "Not found, service not found",
			serviceID:   gofakeit.UUID(),
			translation: randTranslation(t, nil),
			wantCode:    http.StatusNotFound,
		},
		{
			name:      "Bad request, translation not provided",
			serviceID: service.GetId(),
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "Bad request, translation.language not provided",
			serviceID: service.GetId(),
			translation: &translatev1.Translation{
				Language: "",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:      "Status conflict, service already has translation for specified language",
			serviceID: serviceWithTranslations.GetId(),
			translation: &translatev1.Translation{
				Language: uploadReq.GetLanguage(),
			},
			wantCode: http.StatusConflict,
		},
		{
			name:      "Bad request, service already has original translation",
			serviceID: service.GetId(),
			translation: &translatev1.Translation{
				Original: true,
				Language: langs[3].String(),
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.translation)
			expect.NoError(t, err)

			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + tt.serviceID + "/translations",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
			expect.NoError(t, err)

			resp, err := otelhttp.DefaultClient.Do(req)
			expect.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
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

	happyReq := randUpdateTranslationReq(t, service.GetId(), &translatev1.Translation{Language: langs[1].String()}, nil)

	req := randUpdateTranslationReq(t, service.GetId(), &translatev1.Translation{
		Language: langs[1].String(),
		Messages: []*translatev1.Message{
			{
				Id: "Hello", Message: "World",
			},
		},
	},
		&field_mask.FieldMask{Paths: []string{"messages"}})

	// different language without translation
	notFoundTranslationReq := randUpdateTranslationReq(t,
		service.GetId(), &translatev1.Translation{Language: langs[2].String()}, nil)

	notFoundServiceID := randUpdateTranslationReq(t,
		gofakeit.UUID(), &translatev1.Translation{Language: langs[1].String()}, nil)

	invalidArgumentNilTranslationReq := &translatev1.UpdateTranslationRequest{ServiceId: service.GetId()}

	invalidArgumentUndTranslationLanguageReq := randUpdateTranslationReq(t, gofakeit.UUID(), nil, nil)
	invalidArgumentUndTranslationLanguageReq.Translation.Language = ""

	originalAlreadyExistsReq := randUpdateTranslationReq(t,
		service.GetId(), &translatev1.Translation{Language: langs[1].String(), Original: true}, nil)

	tests := []struct {
		request  *translatev1.UpdateTranslationRequest
		name     string
		wantCode uint
	}{
		{
			name:     "Happy Path",
			request:  happyReq,
			wantCode: http.StatusOK,
		},
		{
			name:     "Happy path update messages field",
			request:  req,
			wantCode: http.StatusOK,
		},
		{
			name:     "Message does not exists",
			request:  notFoundTranslationReq,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Service does not exists",
			request:  notFoundServiceID,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Invalid argument nil translation",
			request:  invalidArgumentNilTranslationReq,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Invalid argument und translation.language",
			request:  invalidArgumentUndTranslationLanguageReq,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Bad request, service already has original translation",
			request:  originalAlreadyExistsReq,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			body, err := json.Marshal(tt.request.GetTranslation())
			expect.NoError(t, err)

			language := langs[0].String()

			if tt.request.GetTranslation() != nil {
				language = tt.request.GetTranslation().GetLanguage()
			}

			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + tt.request.GetServiceId() + "/translations/" + language,
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(body))
			expect.NoError(t, err)

			resp, err := otelhttp.DefaultClient.Do(req)
			expect.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, int(tt.wantCode), resp.StatusCode)
		})
	}
}

// GET.
func Test_GetTranslations_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	for range gofakeit.IntRange(1, 5) {
		uploadRequest := randUploadTranslationFileReq(t, service.GetId())
		_, err := client.UploadTranslationFile(ctx, uploadRequest)
		expect.NoError(t, err)
	}

	tests := []struct {
		serviceID string
		name      string
		wantCode  int
	}{
		{
			serviceID: service.GetId(),
			name:      "Happy Path, get all translations",
			wantCode:  http.StatusOK,
		},
		{
			serviceID: gofakeit.UUID(),
			name:      "Happy path, service doesn't exist",
			wantCode:  http.StatusOK,
		},
		{
			name:     "Bad request, ServiceID not provided",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + tt.serviceID + "/translations",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			expect.NoError(t, err)

			resp, err := otelhttp.DefaultClient.Do(req)
			expect.NoError(t, err)

			defer resp.Body.Close()

			if err == nil {
				require.NotEmpty(t, resp.Body)
			}

			assert.Equal(t, tt.wantCode, resp.StatusCode)
		})
	}
}
