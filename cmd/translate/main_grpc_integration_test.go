//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/server"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// -------------Translation File-------------.

func randUploadData(t *testing.T, schema translatev1.Schema, lang language.Tag) []byte {
	t.Helper()

	messages := rand.ModelMessages(3, nil, rand.WithLanguage(lang))

	data, err := server.MessagesToData(schema, messages)
	require.NoError(t, err, "convert rand messages to serialized data")

	return data
}

func randUploadRequest(t *testing.T, serviceID string) *translatev1.UploadTranslationFileRequest {
	t.Helper()

	schema := translatev1.Schema(gofakeit.IntRange(1, 7))
	lang := rand.Language()

	data := randUploadData(t, schema, lang)

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

	happyRequestNoLangInReq := &translatev1.UploadTranslationFileRequest{
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
		expectedCode codes.Code
	}{
		{
			name:         "Happy path",
			request:      happyRequest,
			expectedCode: codes.OK,
		},
		{
			name:         "Happy path no language in request",
			request:      happyRequestNoLangInReq,
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid argument missing service_id",
			request:      invalidArgumentMissingServiceRequest,
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
	uploadReq.Data = randUploadData(t, uploadReq.Schema, language.MustParse(uploadReq.Language))

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
		expectedCode codes.Code
	}{
		{
			name:         "Happy path",
			request:      happyRequest,
			expectedCode: codes.OK,
		},
		{
			name:         "Happy path no messages with language",
			request:      happyReqNoMessagesLanguage,
			expectedCode: codes.OK,
		},
		{
			name:         "Happy path no messages with Service ID",
			request:      happyReqNoMessagesServiceID,
			expectedCode: codes.OK,
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

func randMessages(t *testing.T, override *translatev1.Messages) *translatev1.Messages {
	t.Helper()

	lang := rand.Language().String()
	if override != nil {
		lang = override.Language
	}

	n := gofakeit.IntRange(1, 5)

	msgs := &translatev1.Messages{
		Language: lang,
		Messages: make([]*translatev1.Message, 0, n),
	}

	for i := 0; i < n; i++ {
		message := &translatev1.Message{
			Id:          gofakeit.SentenceSimple(),
			Message:     gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
			Fuzzy:       gofakeit.Bool(),
		}

		if gofakeit.Bool() {
			gofakeit.Slice(message.Positions)
		}

		msgs.Messages = append(msgs.Messages, message)
	}

	return msgs
}

func Test_CreateMessages_gRPC(t *testing.T) {
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
		request      *translatev1.CreateMessagesRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name: "Happy path, create messages",
			request: &translatev1.CreateMessagesRequest{
				ServiceId: service.Id,
				Messages:  randMessages(t, &translatev1.Messages{Language: langs[0].String()}),
			},
			expectedCode: codes.OK,
		},
		{
			name: "Happy path, empty messages.messages",
			request: &translatev1.CreateMessagesRequest{
				ServiceId: service.Id,
				Messages: &translatev1.Messages{
					Language: langs[1].String(),
				},
			},
			expectedCode: codes.OK,
		},
		{
			name: "Not found, service not found",
			request: &translatev1.CreateMessagesRequest{
				ServiceId: gofakeit.UUID(),
				Messages:  randMessages(t, nil),
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Invalid argument, messages not provided",
			request: &translatev1.CreateMessagesRequest{
				ServiceId: service.Id,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid argument, messages.language not provided",
			request: &translatev1.CreateMessagesRequest{
				ServiceId: service.Id,
				Messages: &translatev1.Messages{
					Language: "",
				},
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Already exists, service already has messages for specified language",
			request: &translatev1.CreateMessagesRequest{
				ServiceId: serviceWithMsgs.Id,
				Messages: &translatev1.Messages{
					Language: uploadReq.Language,
				},
			},
			expectedCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			msgs, err := client.CreateMessages(ctx, tt.request)
			if err != nil {
				require.Nil(t, msgs)
			}

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func Test_ListMessages_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	for i := 0; i < gofakeit.IntRange(1, 5); i++ {
		uploadRequest := randUploadRequest(t, service.Id)
		_, err := client.UploadTranslationFile(ctx, uploadRequest)
		require.NoError(t, err, "create test translation file")
	}

	// Requests

	tests := []struct {
		request      *translatev1.ListMessagesRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy path, get all messages",
			request:      &translatev1.ListMessagesRequest{ServiceId: service.Id},
			expectedCode: codes.OK,
		},
		{
			name:         "Happy path, service doesn't exist",
			request:      &translatev1.ListMessagesRequest{ServiceId: uuid.New().String()},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid argument, ServiceID not provided",
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

func Test_UpdateMessages_gRPC(t *testing.T) {
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
		expectedCode codes.Code
	}{
		{
			name:         "Happy Path",
			request:      happyReq,
			expectedCode: codes.OK,
		},
		{
			name:         "Message does not exists",
			request:      notFoundMessagesReq,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Service does not exists",
			request:      notFoundServiceID,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Invalid argument nil messages",
			request:      invalidArgumentNilMessagesReq,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Invalid argument und messages.language",
			request:      invalidArgumentUndMessagesLanguageReq,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := client.UpdateMessages(ctx, tt.request)

			if err == nil {
				require.NotNil(t, resp)
			}

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}
