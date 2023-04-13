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
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
									"meaning":"When you greet someone",
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
									"meaning":"When you greet someone",
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
									"meaning":"When you greet someone",
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
