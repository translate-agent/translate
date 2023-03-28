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
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    *translatev1.UploadTranslationFileRequest
		name     string
		expected codes.Code
	}{
		{
			name: "Happy path",
			input: &translatev1.UploadTranslationFileRequest{
				Language: "lv-lv",
				Data: []byte(`{
						"language":"lv-lv",
						"messages":[
							 {
									"id":"1",
									"meaning":"When you great someone",
									"message":"hello",
									"translation":"čau",
									"fuzzy":false
							 }
						]
				 }`),
				Schema: translatev1.Schema_GO,
			},
			expected: codes.OK,
		},
		{
			name: "Missing language",
			input: &translatev1.UploadTranslationFileRequest{
				Data: []byte(`{
						"messages":[
							 {
									"id":"1",
									"meaning":"When you great someone",
									"message":"hello",
									"translation":"čau",
									"fuzzy":false
							 }
						]
				 }`),
				Schema: translatev1.Schema_GO,
			},
			expected: codes.InvalidArgument,
		},
		{
			name:     "Missing data",
			input:    &translatev1.UploadTranslationFileRequest{Language: "lv-lv"},
			expected: codes.InvalidArgument,
		},
		{
			name: "Invalid language",
			input: &translatev1.UploadTranslationFileRequest{
				Language: "xyz-ZY-Latn",
				Data: []byte(`{
						"messages":[
							 {
									"id":"1",
									"meaning":"When you great someone",
									"message":"hello",
									"translation":"čau",
									"fuzzy":false
							 }
						]
				 }`),
				Schema: translatev1.Schema_GO,
			},
			expected: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.UploadTranslationFile(context.Background(), tt.input)

			actual := status.Code(err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_DownloadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    *translatev1.DownloadTranslationFileRequest
		name     string
		expected codes.Code
	}{
		{
			name:     "Happy path",
			input:    &translatev1.DownloadTranslationFileRequest{Language: "lv-lv", Schema: translatev1.Schema_GO},
			expected: codes.OK,
		},
		{
			name:     "Invalid argument",
			input:    &translatev1.DownloadTranslationFileRequest{Schema: translatev1.Schema_GO},
			expected: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.DownloadTranslationFile(context.Background(), tt.input)

			actual := status.Code(err)
			assert.Equal(t, tt.expected, actual)
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
	if !assert.NoError(t, err, "client.CreateService method returned an error") {
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
	if !assert.NoError(t, err, "client.CreateService method returned an error") {
		return
	}

	service.Name = gofakeit.FirstName()

	_, err = client.UpdateService(ctx, &translatev1.UpdateServiceRequest{
		Service:    service,
		UpdateMask: nil,
	})
	if !assert.NoError(t, err, "client.UpdateService method returned an error") {
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
	if !assert.NoError(t, err, "client.CreateService method returned an error") {
		return
	}

	service.Name = gofakeit.FirstName()

	_, err = client.UpdateService(ctx, &translatev1.UpdateServiceRequest{
		Service:    service,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
	})
	if !assert.NoError(t, err, "client.UpdateService method returned an error") {
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
	if !assert.NoError(t, err, "client.CreateService method returned an error") {
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
	if !assert.NoError(t, err, "client.CreateService method returned an error") {
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
