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
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/text/language"
	"google.golang.org/genproto/protobuf/field_mask"
)

// TODO: Currently, we manually create requests for the REST API.
// We could use a client generated from the OpenAPI specification to simplify testing and integration.

// TODO: DRY: Tests are the same with gRPC tests. We could combine them to avoid duplication.

// -------------Translation File-------------.

func attachFile(t *testing.T, text []byte) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", "test.json")
	if err != nil {
		t.Error(err)
		return nil, ""
	}

	_, err = part.Write(text)
	if err != nil {
		t.Error(err)
		return nil, ""
	}

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

	body, contentType := attachFile(t, req.GetData())

	r, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), body)
	if err != nil {
		t.Error(err)
		return nil
	}

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
	if err != nil {
		t.Error(err)
		return nil
	}

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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			resp, err := otelhttp.DefaultClient.Do(gRPCUploadFileToRESTReq(ctx, t, test.request))
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			// Read the response to give error message on failure
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
				return
			}

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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
	if err != nil {
		t.Error(err)
		return
	}

	// Change translation and upload again with the same language and serviceID
	uploadReq.Data = randUploadData(t, language.MustParse(uploadReq.GetLanguage()))

	resp, err := otelhttp.DefaultClient.Do(gRPCUploadFileToRESTReq(ctx, t, uploadReq))
	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func Test_DownloadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	uploadRequest := randUploadTranslationFileReq(t, service.GetId())

	_, err := client.UploadTranslationFile(ctx, uploadRequest)
	if err != nil {
		t.Error(err)
		return
	}

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
		wantCode int
	}{
		{
			name:     "Happy path",
			request:  happyRequest,
			wantCode: http.StatusOK,
		},

		{
			name:     "Happy path no translation with language",
			request:  happyReqNoTranslationLanguage,
			wantCode: http.StatusOK,
		},
		{
			name:     "Happy path no translation with Service ID",
			request:  happyReqNoTranslationServiceID,
			wantCode: http.StatusOK,
		},
		{
			name:     "Bad request unspecified schema",
			request:  unspecifiedSchemaRequest,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			resp, err := otelhttp.DefaultClient.Do(gRPCDownloadFileToRESTReq(ctx, t, test.request))
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			body, err := json.Marshal(test.service)
			if err != nil {
				t.Error(err)
				return
			}

			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := otelhttp.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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
	if err != nil {
		t.Error(err)
		return
	}

	putBody := restUpdateBody{Name: gofakeit.FirstName()}

	putBodyBytes, err := json.Marshal(putBody)
	if err != nil {
		t.Error(err)
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   "v1/services/" + service.GetId(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(putBodyBytes))
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := otelhttp.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// PATCH.
func Test_UpdateServiceSpecificField_REST(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if err != nil {
		t.Error(err)
		return
	}

	patchBody := restUpdateBody{Name: gofakeit.FirstName()}

	patchBodyBytes, err := json.Marshal(patchBody)
	if err != nil {
		t.Error(err)
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
		Path:   "v1/services/" + service.GetId(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u.String(), bytes.NewReader(patchBodyBytes))
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := otelhttp.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// GET.
func Test_GetService_REST(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	// Using gRPC client to create service
	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if err != nil {
		t.Error(err)
		return
	}

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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + test.service.GetId(),
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := otelhttp.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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
	if err != nil {
		t.Error(err)
		return
	}

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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + test.service.GetId(),
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := otelhttp.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := otelhttp.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
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
	if err != nil {
		t.Error(err)
		return
	}

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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			body, err := json.Marshal(test.translation)
			if err != nil {
				t.Error(err)
				return
			}

			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + test.serviceID + "/translations",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := otelhttp.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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
		wantCode int
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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			body, err := json.Marshal(test.request.GetTranslation())
			if err != nil {
				t.Error(err)
				return
			}

			language := langs[0].String()

			if test.request.GetTranslation() != nil {
				language = test.request.GetTranslation().GetLanguage()
			}

			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + test.request.GetServiceId() + "/translations/" + language,
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewBuffer(body))
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := otelhttp.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
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
		if err != nil {
			t.Error(err)
			return
		}
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

	for _, test := range tests {
		subtest(test.name, func(ctx context.Context, t *testing.T) { //nolint:thelper
			u := url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(host, port),
				Path:   "v1/services/" + test.serviceID + "/translations",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := otelhttp.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			n, err := io.Copy(io.Discard, resp.Body)
			if err != nil {
				t.Errorf("want no error, got %s", err)
				return
			}

			if n == 0 {
				t.Errorf("want response body, got empty")
			}

			if test.wantCode != resp.StatusCode {
				t.Errorf("want status code %d, got %d", test.wantCode, resp.StatusCode)
			}
		})
	}
}
