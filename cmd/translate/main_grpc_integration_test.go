package main

import (
	"context"
	"log"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	transferProtocol string // http or https
	addr             string
	port             string

	client pb.TranslateServiceClient
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		main()
	}()

	// Wait until viper finishes initialization.
	for viper.GetUint("service.port") == 0 {
	}

	addr = viper.GetString("service.address")
	port = viper.GetString("service.port")
	transferProtocol = viper.GetString("service.protocol")

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithBlock(),
	}
	// Wait for the server to start and establish a connection.
	conn, err := grpc.DialContext(context.Background(), ":"+port, grpcOpts...)
	if err != nil {
		log.Panic(err)
	}

	client = pb.NewTranslateServiceClient(conn)

	// Run the tests.
	code := m.Run()
	// Send soft kill (termination) signal to process.
	err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		log.Panic(err)
	}
	// Wait for main() to finish cleanup.
	wg.Wait()
	conn.Close()

	os.Exit(code)
}

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		req  *pb.UploadTranslationFileRequest
		name string
		want codes.Code
	}{
		{
			name: "Happy path",
			req: &pb.UploadTranslationFileRequest{
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
				Schema: pb.Schema_GO,
			},
			want: codes.OK,
		},
		{
			name: "Missing language",
			req: &pb.UploadTranslationFileRequest{
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
				Schema: pb.Schema_GO,
			},
			want: codes.InvalidArgument,
		},
		{
			name: "Missing data",
			req:  &pb.UploadTranslationFileRequest{Language: "lv-lv"},
			want: codes.InvalidArgument,
		},
		{
			name: "Invalid language",
			req: &pb.UploadTranslationFileRequest{
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
				Schema: pb.Schema_GO,
			},
			want: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.UploadTranslationFile(ctx, tt.req)

			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}

func Test_DownloadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		req  *pb.DownloadTranslationFileRequest
		name string
		want codes.Code
	}{
		{
			name: "Happy path",
			req:  &pb.DownloadTranslationFileRequest{Language: "lv-lv"},
			want: codes.OK,
		},
		{
			name: "Invalid argument",
			req:  &pb.DownloadTranslationFileRequest{},
			want: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.DownloadTranslationFile(ctx, tt.req)

			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}
