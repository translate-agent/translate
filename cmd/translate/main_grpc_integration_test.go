//go:build integration

package main

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/convert"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// -------------Translation File-------------.

// randUploadData generates a random translation and serializes it to a po file.
func randUploadData(t *testing.T, lang language.Tag) []byte {
	t.Helper()

	translation := rand.ModelTranslation(3, nil, rand.WithLanguage(lang), rand.WithSimpleMF2Messages())

	data, err := convert.ToPo(*translation)
	require.NoError(t, err, "convert rand translation to serialized data")

	return data
}

func randUploadTranslationFileReq(t *testing.T, serviceID string) *translatev1.UploadTranslationFileRequest {
	t.Helper()

	lang := rand.Language()
	poData := randUploadData(t, lang)

	return &translatev1.UploadTranslationFileRequest{
		ServiceId: serviceID,
		Language:  lang.String(),
		Data:      poData,
		Schema:    translatev1.Schema_PO,
		Original:  ptr(false),
	}
}

func randUpdateTranslationReq(t *testing.T, serviceID string, override *translatev1.Translation,
	updateMask *field_mask.FieldMask,
) *translatev1.UpdateTranslationRequest {
	t.Helper()

	req := translatev1.UpdateTranslationRequest{
		ServiceId:   serviceID,
		Translation: randTranslation(t, override),
		UpdateMask:  updateMask,
	}

	return &req
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

// createTranslation creates a random translation, and calls the CreateTranslation RPC.
func createTranslation(ctx context.Context, t *testing.T, serviceID string,
	override *translatev1.Translation,
) *translatev1.Translation { //nolint:unparam
	t.Helper()

	require.NotEmpty(t, serviceID)

	ctx, span := testutil.Tracer().Start(ctx, "test: create translation")
	defer span.End()

	translation, err := client.CreateTranslation(ctx, &translatev1.CreateTranslationRequest{
		ServiceId:   serviceID,
		Translation: randTranslation(t, override),
	})
	require.NoError(t, err, "create test translation")

	return translation
}

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	langs := rand.Languages(2)
	serviceWithExistingOriginal := createService(ctx, t)
	createTranslation(ctx, t,
		serviceWithExistingOriginal.GetId(), &translatev1.Translation{Original: true, Language: langs[0].String()})

	// Requests

	happyRequest := randUploadTranslationFileReq(t, service.GetId())

	happyRequestNoLangInReq := &translatev1.UploadTranslationFileRequest{
		ServiceId: service.GetId(),
		Data:      randUploadData(t, rand.Language()),
		Schema:    translatev1.Schema_PO,
	}

	invalidArgumentMissingServiceRequest := randUploadTranslationFileReq(t, service.GetId())
	invalidArgumentMissingServiceRequest.ServiceId = ""

	notFoundServiceIDRequest := randUploadTranslationFileReq(t, gofakeit.UUID())

	originalAlreadyExistsReq := &translatev1.UploadTranslationFileRequest{
		Original:  ptr(true),
		Language:  langs[1].String(),
		ServiceId: serviceWithExistingOriginal.GetId(),
		Data:      randUploadData(t, langs[1]),
		Schema:    translatev1.Schema_PO,
	}

	tests := []struct {
		request  *translatev1.UploadTranslationFileRequest
		name     string
		wantcode codes.Code
	}{
		{
			name:     "Happy path",
			request:  happyRequest,
			wantcode: codes.OK,
		},
		{
			name:     "Happy path no language in request",
			request:  happyRequestNoLangInReq,
			wantcode: codes.OK,
		},
		{
			name:     "Invalid argument missing service_id",
			request:  invalidArgumentMissingServiceRequest,
			wantcode: codes.InvalidArgument,
		},
		{
			name:     "Not found service ID",
			request:  notFoundServiceIDRequest,
			wantcode: codes.NotFound,
		},
		{
			name:     "Invalid argument, service already has original translation",
			request:  originalAlreadyExistsReq,
			wantcode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.UploadTranslationFile(ctx, tt.request)

			assert.Equal(t, tt.wantcode, status.Code(err), "want %s, got %s", tt.wantcode, status.Code(err))
		})
	}
}

func Test_UploadTranslationFileUpdateFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	// Upload initial

	uploadReq := randUploadTranslationFileReq(t, service.GetId())

	_, err := client.UploadTranslationFile(ctx, uploadReq)
	require.NoError(t, err, "create test translation file")

	// Change translation and upload again with the same language and serviceID
	uploadReq.Data = randUploadData(t, language.MustParse(uploadReq.GetLanguage()))

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
		request  *translatev1.DownloadTranslationFileRequest
		name     string
		wantCode codes.Code
	}{
		{
			name:     "Happy path",
			request:  happyRequest,
			wantCode: codes.OK,
		},
		{
			name:     "Happy path no translation with language",
			request:  happyReqNoTranslationLanguage,
			wantCode: codes.OK,
		},
		{
			name:     "Happy path no translation with Service ID",
			request:  happyReqNoTranslationServiceID,
			wantCode: codes.OK,
		},
		{
			name:     "Invalid argument unspecified schema",
			request:  unspecifiedSchemaRequest,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.DownloadTranslationFile(ctx, tt.request)

			assert.Equal(t, tt.wantCode, status.Code(err))
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
		request  *translatev1.CreateServiceRequest
		name     string
		wantCode codes.Code
	}{
		{
			name:     "Happy path With ID",
			request:  &translatev1.CreateServiceRequest{Service: serviceWithID},
			wantCode: codes.OK,
		},
		{
			name:     "Happy path Without ID",
			request:  &translatev1.CreateServiceRequest{Service: serviceWithoutID},
			wantCode: codes.OK,
		},
		{
			name:     "Invalid argument malformed ID",
			request:  &translatev1.CreateServiceRequest{Service: serviceMalformedID},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.CreateService(ctx, tt.request)

			assert.Equal(t, tt.wantCode, status.Code(err))
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
		wantCode        codes.Code
	}{
		{
			name:            "Happy path all fields",
			wantCode:        codes.OK,
			serviceToUpdate: createService(ctx, t),
			request: &translatev1.UpdateServiceRequest{
				Service:    randService(),
				UpdateMask: nil,
			},
		},
		{
			name:            "Happy path one field",
			wantCode:        codes.OK,
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
			wantCode:        codes.InvalidArgument,
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
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			// Change the ID to the one of the service that was created in the prepare step.
			tt.request.Service.Id = tt.serviceToUpdate.GetId()

			_, err := client.UpdateService(ctx, tt.request)

			assert.Equal(t, tt.wantCode, status.Code(err))
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
		request  *translatev1.GetServiceRequest
		name     string
		wantCode codes.Code
	}{
		{
			name:     "Happy Path",
			request:  &translatev1.GetServiceRequest{Id: service.GetId()},
			wantCode: codes.OK,
		},
		{
			name:     "Not found",
			request:  &translatev1.GetServiceRequest{Id: gofakeit.UUID()},
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.GetService(ctx, tt.request)

			assert.Equal(t, tt.wantCode, status.Code(err))
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
		request  *translatev1.DeleteServiceRequest
		name     string
		wantCode codes.Code
	}{
		{
			request:  &translatev1.DeleteServiceRequest{Id: service.GetId()},
			name:     "Happy Path",
			wantCode: codes.OK,
		},
		{
			request:  &translatev1.DeleteServiceRequest{Id: gofakeit.UUID()},
			name:     "Not found",
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			_, err := client.DeleteService(ctx, tt.request)

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func Test_ListServices_gRPC(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	_, err := client.ListServices(ctx, &translatev1.ListServicesRequest{})

	assert.Equal(t, codes.OK, status.Code(err))
}

// ------------------Translation------------------

func randTranslation(t *testing.T, override *translatev1.Translation) *translatev1.Translation {
	t.Helper()

	n := gofakeit.IntRange(0, 3)

	translation := &translatev1.Translation{
		Language: rand.Language().String(),
		Messages: make([]*translatev1.Message, 0, n),
	}

	if override != nil && override.GetLanguage() != "" {
		translation.Language = override.GetLanguage()
	}

	if override != nil && override.GetOriginal() {
		translation.Original = override.GetOriginal()
	}

	if override != nil && override.GetMessages() != nil {
		translation.Messages = override.GetMessages()
		return translation
	}

	for range n {
		message := &translatev1.Message{
			Id:          gofakeit.SentenceSimple(),
			Message:     gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
			Status:      translatev1.Message_Status(gofakeit.IntRange(0, 2)),
		}

		for range gofakeit.Number(0, 2) {
			v, _ := url.Parse(gofakeit.URL())
			lineNumber := gofakeit.Number(0, 10_000)
			message.Positions = append(message.GetPositions(), fmt.Sprintf("%s:%d", v.Path, lineNumber))
		}

		translation.Messages = append(translation.GetMessages(), message)
	}

	return translation
}

func Test_CreateTranslation_gRPC(t *testing.T) {
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
		request  *translatev1.CreateTranslationRequest
		name     string
		wantCode codes.Code
	}{
		{
			name: "Happy path, create translation",
			request: &translatev1.CreateTranslationRequest{
				ServiceId:   service.GetId(),
				Translation: randTranslation(t, &translatev1.Translation{Language: langs[1].String()}),
			},
			wantCode: codes.OK,
		},
		{
			name: "Happy path, empty translation.messages",
			request: &translatev1.CreateTranslationRequest{
				ServiceId: service.GetId(),
				Translation: &translatev1.Translation{
					Language: langs[2].String(),
				},
			},
			wantCode: codes.OK,
		},
		{
			name: "Not found, service not found",
			request: &translatev1.CreateTranslationRequest{
				ServiceId:   gofakeit.UUID(),
				Translation: randTranslation(t, nil),
			},
			wantCode: codes.NotFound,
		},
		{
			name: "Invalid argument, translation not provided",
			request: &translatev1.CreateTranslationRequest{
				ServiceId: service.GetId(),
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "Invalid argument, translation.language not provided",
			request: &translatev1.CreateTranslationRequest{
				ServiceId: service.GetId(),
				Translation: &translatev1.Translation{
					Language: "",
				},
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "Already exists, service already has translation for specified language",
			request: &translatev1.CreateTranslationRequest{
				ServiceId: serviceWithTranslations.GetId(),
				Translation: &translatev1.Translation{
					Language: uploadReq.GetLanguage(),
				},
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "Invalid argument, service already has original translation",
			request: &translatev1.CreateTranslationRequest{
				ServiceId: service.GetId(),
				Translation: &translatev1.Translation{
					Language: langs[3].String(),
					Original: true,
				},
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			translation, err := client.CreateTranslation(ctx, tt.request)
			if err != nil {
				require.Nil(t, translation)
			}

			assert.Equal(t, tt.wantCode, status.Code(err))

			if status.Code(err) == codes.OK {
				require.Equal(t, tt.request.GetTranslation().GetLanguage(), translation.GetLanguage())
			}
		})
	}
}

func Test_ListTranslations_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)

	for range gofakeit.IntRange(1, 5) {
		uploadRequest := randUploadTranslationFileReq(t, service.GetId())
		_, err := client.UploadTranslationFile(ctx, uploadRequest)
		require.NoError(t, err, "create test translation file")
	}

	// Requests

	tests := []struct {
		request  *translatev1.ListTranslationsRequest
		name     string
		wantCode codes.Code
	}{
		{
			name:     "Happy path, get all translations",
			request:  &translatev1.ListTranslationsRequest{ServiceId: service.GetId()},
			wantCode: codes.OK,
		},
		{
			name:     "Happy path, service doesn't exist",
			request:  &translatev1.ListTranslationsRequest{ServiceId: uuid.New().String()},
			wantCode: codes.OK,
		},
		{
			name:     "Invalid argument, ServiceID not provided",
			request:  &translatev1.ListTranslationsRequest{},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := client.ListTranslations(ctx, tt.request)

			if err == nil {
				require.NotNil(t, resp)
			}

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func Test_UpdateTranslationFromMask_gRPC(t *testing.T) {
	t.Parallel()

	ctx, _ := testutil.Trace(t)

	// Prepare
	lang := rand.Language()
	service := createService(ctx, t)

	// Add multiple messages and update one message.
	createTranslation(ctx, t, service.GetId(), &translatev1.Translation{
		Original: false,
		Language: lang.String(),
		Messages: []*translatev1.Message{
			{
				Id:      "Hello",
				Message: "Sveiks",
			},
		},
	})

	req := randUpdateTranslationReq(t, service.GetId(), &translatev1.Translation{
		Language: lang.String(),
		Messages: []*translatev1.Message{
			{
				Id:          "Bye",
				Message:     "World",
				Description: "farewell",
				Status:      translatev1.Message_UNTRANSLATED,
			},
			{
				Id:          "Hi",
				Message:     "Dog",
				Description: "greeting",
				Status:      translatev1.Message_UNTRANSLATED,
			},
			{
				Id:          "Hello",
				Message:     "Hi",
				Description: "greeting",
				Status:      translatev1.Message_UNTRANSLATED,
			},
		},
	}, &field_mask.FieldMask{Paths: []string{"messages"}})

	want := &translatev1.ListTranslationsResponse{
		Translations: []*translatev1.Translation{
			{
				Language: lang.String(),
				Original: false,
				Messages: []*translatev1.Message{
					{
						Id:          "Hello",
						Message:     "Hi",
						Description: "greeting",
						Status:      translatev1.Message_UNTRANSLATED,
					},
					{
						Id:          "Bye",
						Message:     "World",
						Description: "farewell",
						Status:      translatev1.Message_UNTRANSLATED,
					},
					{
						Id:          "Hi",
						Message:     "Dog",
						Description: "greeting",
						Status:      translatev1.Message_UNTRANSLATED,
					},
				},
			},
		},
	}

	resp, err := client.UpdateTranslation(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	got, err := client.ListTranslations(ctx, &translatev1.ListTranslationsRequest{
		ServiceId: service.GetId(),
	})

	require.NoError(t, err)

	assert.ElementsMatch(t, want.GetTranslations(), got.GetTranslations())
}

func Test_UpdateTranslation_gRPC(t *testing.T) {
	t.Parallel()

	ctx, subtest := testutil.Trace(t)

	// Prepare
	service := createService(ctx, t)
	langs := rand.Languages(3)

	createTranslation(ctx, t, service.GetId(), &translatev1.Translation{Original: true, Language: langs[0].String()})
	createTranslation(ctx, t, service.GetId(), &translatev1.Translation{Original: false, Language: langs[1].String()})

	happyReq := randUpdateTranslationReq(t, service.GetId(), &translatev1.Translation{Language: langs[1].String()}, nil)

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
		wantCode codes.Code
	}{
		{
			name:     "Happy Path update all",
			request:  happyReq,
			wantCode: codes.OK,
		},
		{
			name:     "Translation does not exists",
			request:  notFoundTranslationReq,
			wantCode: codes.NotFound,
		},
		{
			name:     "Service does not exists",
			request:  notFoundServiceID,
			wantCode: codes.NotFound,
		},
		{
			name:     "Invalid argument nil translation",
			request:  invalidArgumentNilTranslationReq,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Invalid argument und translation.language",
			request:  invalidArgumentUndTranslationLanguageReq,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Invalid argument, service already has original translation",
			request:  originalAlreadyExistsReq,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		subtest(tt.name, func(ctx context.Context, t *testing.T) {
			resp, err := client.UpdateTranslation(ctx, tt.request)

			if err == nil {
				require.NotNil(t, resp)
			}

			assert.Equal(t, tt.wantCode, status.Code(err))

			if tt.request == happyReq {
				matchingTranslationExistsInService(ctx, t, tt.request.GetServiceId(), resp)
			}
		})
	}
}

// matchingTranslationExistsInService checks incoming translation is equal to translation
// with same language returned from listTranslations.
func matchingTranslationExistsInService(
	ctx context.Context,
	t *testing.T,
	serviceID string,
	translation *translatev1.Translation,
) {
	t.Helper()

	resp, err := client.ListTranslations(ctx, &translatev1.ListTranslationsRequest{
		ServiceId: serviceID,
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp)

	var translationFromService *translatev1.Translation

	for i := range resp.GetTranslations() {
		if resp.GetTranslations()[i].GetLanguage() == translation.GetLanguage() {
			translationFromService = resp.GetTranslations()[i]
			break
		}
	}

	require.NotNil(t, translationFromService)
	require.Equal(t, translation.GetOriginal(), translationFromService.GetOriginal())
	require.Equal(t, translation.GetLanguage(), translationFromService.GetLanguage())
	require.ElementsMatch(t, translation.GetMessages(), translationFromService.GetMessages())
}
