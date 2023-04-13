//go:build integration

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/translate"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/text/language"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	host string
	port string

	client translatev1.TranslateServiceClient
)

func mustGetFreePort() string {
	// Listen on port 0 to have the operating system allocate an available port.
	l, err := net.Listen("tcp", host+":0")
	if err != nil {
		log.Panicf("get free port: %v", err)
	}
	defer l.Close()

	// Get the port number from the address that the Listener is listening on.
	addr := l.Addr().(*net.TCPAddr)

	return fmt.Sprint(addr.Port)
}

func TestMain(m *testing.M) {
	host = "localhost"
	port = mustGetFreePort()

	viper.Set("service.port", port)
	viper.Set("service.host", host)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		main()
	}()

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithBlock(),
	}
	// Wait for the server to start and establish a connection.
	conn, err := grpc.DialContext(context.Background(), host+":"+port, grpcOpts...)
	if err != nil {
		log.Panicf("create connection: %v", err)
	}

	client = translatev1.NewTranslateServiceClient(conn)
	// Run the tests.
	code := m.Run()
	// Send soft kill (termination) signal to process.
	err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		log.Panicf("send termination signal: %v", err)
	}
	// Wait for main() to finish cleanup.
	wg.Wait()
	conn.Close()

	os.Exit(code)
}

// -------------Translation File-------------.

func randUploadData(t *testing.T, schema translatev1.Schema) ([]byte, language.Tag) {
	t.Helper()

	messagesCount := gofakeit.IntRange(1, 5)

	lang := language.MustParse(gofakeit.LanguageBCP())

	messages := model.Messages{
		Language: lang,
		Messages: make([]model.Message, 0, messagesCount),
	}

	for i := 0; i < messagesCount; i++ {
		messages.Messages = append(messages.Messages, model.Message{
			ID:          gofakeit.SentenceSimple(),
			Description: gofakeit.SentenceSimple(),
		},
		)
	}

	data, err := translate.MessagesToData(schema, messages)
	require.NoError(t, err, "convert rand messages to serialized data")

	return data, lang
}

func randUploadRequest(t *testing.T, serviceID string) *translatev1.UploadTranslationFileRequest {
	t.Helper()

	schema := translatev1.Schema(gofakeit.IntRange(1, 7))
	for schema == translatev1.Schema_POT {
		schema = translatev1.Schema(gofakeit.IntRange(1, 7))
	}

	data, lang := randUploadData(t, schema)

	return &translatev1.UploadTranslationFileRequest{
		ServiceId: serviceID,
		Language:  lang.String(),
		Data:      data,
		Schema:    schema,
	}
}

func prepareService(ctx context.Context, t *testing.T) *translatev1.Service {
	t.Helper()

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	return service
}

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := prepareService(ctx, t)

	// Requests

	happyRequest := randUploadRequest(t, service.Id)

	invalidArgumentRequest := randUploadRequest(t, service.Id)
	invalidArgumentRequest.Language = ""

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
			name:         "Invalid argument No language",
			request:      invalidArgumentRequest,
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.UploadTranslationFile(ctx, tt.request)

			actualCode := status.Code(err)
			assert.Equal(t, tt.expectedCode, actualCode, "want codes.%s got codes.%s\nerr: %s", tt.expectedCode, actualCode, err)
		})
	}
}

func Test_UploadTranslationFileDifferentLanguages_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := prepareService(ctx, t)

	uploadRequest := randUploadRequest(t, service.Id)

	for i := 0; i < 3; i++ {
		uploadRequest.Language = gofakeit.LanguageBCP()

		_, err := client.UploadTranslationFile(ctx, uploadRequest)

		expectedCode := codes.OK
		actualCode := status.Code(err)

		require.Equal(t, expectedCode, actualCode, "want codes.%s got codes.%s\nerr: %s", expectedCode, actualCode, err)

	}
}

func Test_UploadTranslationFileUpdateFile_gRPC(t *testing.T) {
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

	_, err = client.UploadTranslationFile(ctx, uploadReq)

	expectedCode := codes.OK
	actualCode := status.Code(err)

	assert.Equal(t, expectedCode, actualCode, "want codes.%s got codes.%s\nerr: %s", expectedCode, actualCode, err)
}

func randDownloadRequest(serviceID, lang string) *translatev1.DownloadTranslationFileRequest {
	schema := translatev1.Schema(gofakeit.IntRange(1, 7))
	for schema == translatev1.Schema_POT {
		schema = translatev1.Schema(gofakeit.IntRange(1, 7))
	}

	return &translatev1.DownloadTranslationFileRequest{
		ServiceId: serviceID,
		Language:  lang,
		Schema:    schema,
	}
}

func Test_DownloadTranslationFile_gRPC(t *testing.T) {
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
			name:         "Invalid argument unspecified schema",
			request:      unspecifiedSchemaRequest,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Not found Service ID",
			request:      notFoundServiceIDRequest,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Not found language",
			request:      notFoundLanguageRequest,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.DownloadTranslationFile(ctx, tt.request)

			actualCode := status.Code(err)
			assert.Equal(t, tt.expectedCode, actualCode, "want codes.%s got codes.%s\nerr: %s", tt.expectedCode, actualCode, err)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.CreateService(context.Background(), tt.request)

			actualCode := status.Code(err)

			assert.Equal(t, tt.expectedCode, actualCode, "want codes.%s got codes.%s", tt.expectedCode, actualCode)
		})
	}
}

func Test_UpdateService_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	test := []struct {
		request      *translatev1.UpdateServiceRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy path all fields",
			expectedCode: codes.OK,
			request: &translatev1.UpdateServiceRequest{
				Service:    randService(),
				UpdateMask: nil,
			},
		},
		{
			name:         "Happy path one field",
			expectedCode: codes.OK,
			request: &translatev1.UpdateServiceRequest{
				Service: randService(),
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"name"},
				},
			},
		},
		{
			name:         "Invalid field in update mask",
			expectedCode: codes.InvalidArgument,
			request: &translatev1.UpdateServiceRequest{
				Service: randService(),
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"invalid_field"},
				},
			},
		},
	}

	for _, tt := range test {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: tt.request.Service})
			require.NoError(t, err, "Prepare test service")

			_, err = client.UpdateService(ctx, tt.request)

			actualCode := status.Code(err)

			assert.Equal(t, tt.expectedCode, actualCode, "want codes.%s got codes.%s", tt.expectedCode, actualCode)
		})
	}
}

func Test_GetService_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.GetService(ctx, tt.request)

			actualCode := status.Code(err)
			assert.Equal(t, tt.expectedCode, actualCode, "want codes.%s got codes.%s", tt.expectedCode, actualCode)
		})
	}
}

func Test_DeleteService_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.DeleteService(ctx, tt.request)

			actualCode := status.Code(err)
			assert.Equal(t, tt.expectedCode, actualCode, "want codes.%s got codes.%s", tt.expectedCode, actualCode)
		})
	}
}

func Test_ListServices_gRPC(t *testing.T) {
	t.Parallel()

	_, err := client.ListServices(context.Background(), &translatev1.ListServicesRequest{})

	expectedCode := codes.OK
	actualCode := status.Code(err)

	assert.Equal(t, expectedCode, actualCode, "want codes.%s got codes.%s", expectedCode, actualCode)
}
