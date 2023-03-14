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

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	tpb "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	host string
	port string

	client tpb.TranslateServiceClient
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

	client = tpb.NewTranslateServiceClient(conn)
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

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    *tpb.UploadTranslationFileRequest
		name     string
		expected codes.Code
	}{
		{
			name: "Happy path",
			input: &tpb.UploadTranslationFileRequest{
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
				Schema: tpb.Schema_GO,
			},
			expected: codes.OK,
		},
		{
			name: "Missing language",
			input: &tpb.UploadTranslationFileRequest{
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
				Schema: tpb.Schema_GO,
			},
			expected: codes.InvalidArgument,
		},
		{
			name:     "Missing data",
			input:    &tpb.UploadTranslationFileRequest{Language: "lv-lv"},
			expected: codes.InvalidArgument,
		},
		{
			name: "Invalid language",
			input: &tpb.UploadTranslationFileRequest{
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
				Schema: tpb.Schema_GO,
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
		input    *tpb.DownloadTranslationFileRequest
		name     string
		expected codes.Code
	}{
		{
			name:     "Happy path",
			input:    &tpb.DownloadTranslationFileRequest{Language: "lv-lv", Schema: tpb.Schema_GO},
			expected: codes.OK,
		},
		{
			name:     "Invalid argument",
			input:    &tpb.DownloadTranslationFileRequest{Schema: tpb.Schema_GO},
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
