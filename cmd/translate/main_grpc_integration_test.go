package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func createConnection(ctx context.Context, t *testing.T) (*grpc.ClientConn, error) {
	t.Helper()

	conn, err := grpc.DialContext(
		ctx,
		"localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("creating connection: %w", err)
	}

	return conn, nil
}

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

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

			ctx := context.Background()

			conn, err := createConnection(ctx, t)
			if !assert.NoError(t, err) {
				return
			}

			defer conn.Close()

			client := pb.NewTranslateServiceClient(conn)
			_, err = client.UploadTranslationFile(ctx, tt.req)

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

			conn, err := createConnection(ctx, t)
			if !assert.NoError(t, err) {
				return
			}

			defer conn.Close()

			client := pb.NewTranslateServiceClient(conn)
			_, err = client.DownloadTranslationFile(ctx, tt.req)

			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}
