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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
	return &translatev1.DownloadTranslationFileRequest{
		ServiceId: serviceID,
		Language:  lang,
		Schema:    translatev1.Schema(gofakeit.IntRange(1, 7)),
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

	invalidArgumentRequest := randDownloadRequest(service.Id, uploadRequest.Language)
	invalidArgumentRequest.Schema = translatev1.Schema_UNSPECIFIED

	notFoundIDRequest := randDownloadRequest(gofakeit.UUID(), uploadRequest.Language)

	notFoundLanguageRequest := randDownloadRequest(service.Id, gofakeit.LanguageBCP())

	tests := []struct {
		input        *translatev1.DownloadTranslationFileRequest
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Happy path",
			input:        happyRequest,
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid argument",
			input:        invalidArgumentRequest,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Not found ID",
			input:        notFoundIDRequest,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Not found language",
			input:        notFoundLanguageRequest,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.DownloadTranslationFile(ctx, tt.input)

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

	service := randService()

	_, err := client.CreateService(context.Background(), &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	actual := status.Code(err)
	expected := codes.OK

	assert.Equal(t, expected, actual, "want codes.%s got codes.%s", expected, actual)
}

func Test_UpdateServiceAllFields_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	service.Name = gofakeit.FirstName()

	_, err = client.UpdateService(ctx, &translatev1.UpdateServiceRequest{
		Service:    service,
		UpdateMask: nil,
	})
	if !assert.NoError(t, err) {
		return
	}

	actual := status.Code(err)
	expected := codes.OK

	assert.Equal(t, expected, actual, "want codes.%s got codes.%s", expected, actual)
}

func Test_UpdateServiceSpecificField_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service := randService()

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	service.Name = gofakeit.FirstName()

	_, err = client.UpdateService(ctx, &translatev1.UpdateServiceRequest{
		Service:    service,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
	})
	if !assert.NoError(t, err) {
		return
	}

	actual := status.Code(err)
	expected := codes.OK

	assert.Equal(t, expected, actual, "want codes.%s got codes.%s", expected, actual)
}

func Test_GetService_gRPC(t *testing.T) {
	t.Parallel()

	service := randService()

	_, err := client.CreateService(context.Background(), &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	tests := []struct {
		input    *translatev1.GetServiceRequest
		name     string
		expected codes.Code
	}{
		{
			input:    &translatev1.GetServiceRequest{Id: service.Id},
			name:     "Happy Path",
			expected: codes.OK,
		},
		{
			input:    &translatev1.GetServiceRequest{Id: gofakeit.UUID()},
			name:     "Not found",
			expected: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.GetService(context.Background(), tt.input)

			actual := status.Code(err)
			assert.Equal(t, tt.expected, actual, "want codes.%s got codes.%s", tt.expected, actual)
		})
	}
}

func Test_DeleteService_gRPC(t *testing.T) {
	t.Parallel()

	service := randService()

	_, err := client.CreateService(context.Background(), &translatev1.CreateServiceRequest{Service: service})
	if !assert.NoError(t, err, "Prepare test data") {
		return
	}

	tests := []struct {
		input    *translatev1.DeleteServiceRequest
		name     string
		expected codes.Code
	}{
		{
			input:    &translatev1.DeleteServiceRequest{Id: service.Id},
			name:     "Happy Path",
			expected: codes.OK,
		},
		{
			input:    &translatev1.DeleteServiceRequest{Id: gofakeit.UUID()},
			name:     "Not found",
			expected: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.DeleteService(context.Background(), tt.input)

			actual := status.Code(err)
			assert.Equal(t, tt.expected, actual, "want codes.%s got codes.%s", tt.expected, actual)
		})
	}
}

func Test_ListServices_gRPC(t *testing.T) {
	t.Parallel()

	_, err := client.ListServices(context.Background(), &translatev1.ListServicesRequest{})

	expected := codes.OK
	actual := status.Code(err)

	assert.Equal(t, expected, actual, "want codes.%s got codes.%s", expected, actual)
}
