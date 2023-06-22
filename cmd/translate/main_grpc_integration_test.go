//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/translate"
	"golang.org/x/text/language"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// -------------Translation File-------------.

func randUploadData(t *testing.T, schema translatev1.Schema) ([]byte, language.Tag) {
	t.Helper()

	n := gofakeit.IntRange(1, 5)
	lang := language.MustParse(gofakeit.LanguageBCP())
	messages := model.Messages{
		Language: lang,
		Messages: make([]model.Message, 0, n),
	}

	for i := 0; i < n; i++ {
		message := model.Message{ID: gofakeit.SentenceSimple(), Description: gofakeit.SentenceSimple()}
		messages.Messages = append(messages.Messages, message)
	}

	data, err := translate.MessagesToData(schema, messages)
	require.NoError(t, err, "convert rand messages to serialized data")

	return data, lang
}

func randUploadRequest(t *testing.T, serviceID string) *translatev1.UploadTranslationFileRequest {
	t.Helper()

	schema := translatev1.Schema(gofakeit.IntRange(1, 7))
	data, lang := randUploadData(t, schema)

	return &translatev1.UploadTranslationFileRequest{
		ServiceId: serviceID,
		Language:  lang.String(),
		Data:      data,
		Schema:    schema,
	}
}

// createService creates a random service, and calls the CreateService RPC.
func createService(ctx context.Context, t *testing.T) *translatev1.Service {
	t.Helper()

	ctx, span := testutil.Tracer().Start(ctx, "test: create service")
	defer span.End()

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	return service
}

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	// Requests

	happyRequest := randUploadRequest(t, service.Id)

	invalidArgumentMissingLangRequest := randUploadRequest(t, service.Id)
	invalidArgumentMissingLangRequest.Language = ""

	notFoundServiceIDRequest := randUploadRequest(t, gofakeit.UUID())

	tests := []struct {
		request      *translatev1.UploadTranslationFileRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy path",
			request:      happyRequest,
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid argument missing language",
			request:      invalidArgumentMissingLangRequest,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Not found service ID",
			request:      notFoundServiceIDRequest,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.UploadTranslationFile(ctx, tt.request)

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func Test_UploadTranslationFileUpdateFile_gRPC(t *testing.T) {
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

	_, err = client.UploadTranslationFile(ctx, uploadReq)

	assert.Equal(t, codes.OK, status.Code(err))
}

func randDownloadRequest(serviceID, lang string) *translatev1.DownloadTranslationFileRequest {
	return &translatev1.DownloadTranslationFileRequest{
		ServiceId: serviceID,
		Language:  lang,
		Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
	}
}

func Test_DownloadTranslationFile_gRPC(t *testing.T) {
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
		expectedCode codes.Code
	}{
		{
			name:         "Happy path",
			request:      happyRequest,
			expectedCode: codes.OK,
		},
		{
			name:         "File not found, no messages with language",
			request:      happyReqNoMessagesLanguage,
			expectedCode: codes.NotFound,
		},
		{
			name:         "File not found, no messages with Service ID",
			request:      happyReqNoMessagesServiceID,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Invalid argument unspecified schema",
			request:      unspecifiedSchemaRequest,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.DownloadTranslationFile(ctx, tt.request)

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

// ------------------Service------------------

func randService() *translatev1.Service {
	return &translatev1.Service{
		Id:   gofakeit.UUID(),
		Name: gofakeit.FirstName(),
	}
}

func Test_CreateService_gRPC(t *testing.T) {
	t.Parallel()

	_, subtest := testutil.Trace(t)

	// Prepare
	serviceWithID := randService()

	serviceWithoutID := randService()
	serviceWithoutID.Id = ""

	serviceMalformedID := randService()
	serviceMalformedID.Id += "_FAIL"

	tests := []struct {
		request      *translatev1.CreateServiceRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy path With ID",
			request:      &translatev1.CreateServiceRequest{Service: serviceWithID},
			expectedCode: codes.OK,
		},
		{
			name:         "Happy path Without ID",
			request:      &translatev1.CreateServiceRequest{Service: serviceWithoutID},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid argument malformed ID",
			request:      &translatev1.CreateServiceRequest{Service: serviceMalformedID},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.CreateService(ctx, tt.request)

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func Test_UpdateService_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	tests := []struct {
		request         *translatev1.UpdateServiceRequest
		serviceToUpdate *translatev1.Service
		name            string
		expectedCode    codes.Code
	}{
		{
			name:            "Happy path all fields",
			expectedCode:    codes.OK,
			serviceToUpdate: createService(ctx, t),
			request: &translatev1.UpdateServiceRequest{
				Service:    randService(),
				UpdateMask: nil,
			},
		},
		{
			name:            "Happy path one field",
			expectedCode:    codes.OK,
			serviceToUpdate: createService(ctx, t),
			request: &translatev1.UpdateServiceRequest{
				Service: randService(),
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"name"},
				},
			},
		},
		{
			name:            "Invalid field in update mask",
			expectedCode:    codes.InvalidArgument,
			serviceToUpdate: createService(ctx, t),
			request: &translatev1.UpdateServiceRequest{
				Service: randService(),
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"invalid_field"},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			// Change the ID to the one of the service that was created in the prepare step.
			tt.request.Service.Id = tt.serviceToUpdate.Id

			_, err := client.UpdateService(ctx, tt.request)

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func Test_GetService_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "Prepare test service")

	tests := []struct {
		request      *translatev1.GetServiceRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy Path",
			request:      &translatev1.GetServiceRequest{Id: service.Id},
			expectedCode: codes.OK,
		},
		{
			name:         "Not found",
			request:      &translatev1.GetServiceRequest{Id: gofakeit.UUID()},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.GetService(ctx, tt.request)

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func Test_DeleteService_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "Prepare test service")

	tests := []struct {
		request      *translatev1.DeleteServiceRequest
		name         string
		expectedCode codes.Code
	}{
		{
			request:      &translatev1.DeleteServiceRequest{Id: service.Id},
			name:         "Happy Path",
			expectedCode: codes.OK,
		},
		{
			request:      &translatev1.DeleteServiceRequest{Id: gofakeit.UUID()},
			name:         "Not found",
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.DeleteService(ctx, tt.request)

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func Test_ListServices_gRPC(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	_, err := client.ListServices(ctx, &translatev1.ListServicesRequest{})

	assert.Equal(t, codes.OK, status.Code(err))
}

// ------------------Messages------------------

func Test_ListMessages_gRPC(t *testing.T) {
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

	tests := []struct {
		request      *translatev1.ListMessagesRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy path",
			request:      &translatev1.ListMessagesRequest{ServiceId: service.Id},
			expectedCode: codes.OK,
		},
		{
			name:         "Happy path, service doesn't exist",
			request:      &translatev1.ListMessagesRequest{ServiceId: uuid.New().String()},
			expectedCode: codes.OK,
		},
		{
			name: "Happy path, filter language",
			request: &translatev1.ListMessagesRequest{
				ServiceId: service.Id,
				Languages: []string{gofakeit.LanguageBCP()},
			},
			expectedCode: codes.OK,
		},
		{
			name: "Happy path, filter existing languages",
			request: &translatev1.ListMessagesRequest{
				ServiceId: uuid.New().String(),
				Languages: langTags,
			},
			expectedCode: codes.OK,
		},
		{
			name:         "ServiceID not provided",
			request:      &translatev1.ListMessagesRequest{},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := client.ListMessages(ctx, tt.request)

			if err == nil {
				require.NotNil(t, resp)
			}

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}
