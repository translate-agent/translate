package main

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func createConnection(ctx context.Context, t *testing.T) *grpc.ClientConn {
	t.Helper()

	conn, err := grpc.DialContext(
		ctx,
		"localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Panic("Failed to dial bufnet: %w", err)
	}

	return conn
}

func Test_UploadTranslationFile_gRPC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type args struct {
		req *pb.UploadTranslationFileRequest
	}

	tests := []struct {
		args args
		name string
		want codes.Code
	}{
		{
			name: "Happy path",
			args: args{
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
				},
			},
			want: codes.OK,
		},
		{
			name: "Missing language",
			args: args{
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
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "Missing data",
			args: args{
				req: &pb.UploadTranslationFileRequest{
					Language: "lv-lv",
				},
			},
			want: codes.InvalidArgument,
		},
		{
			name: "Invalid language",
			args: args{
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
				},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn := createConnection(ctx, t)
			defer conn.Close()

			client := pb.NewTranslateServiceClient(conn)
			_, err := client.UploadTranslationFile(ctx, tt.args.req)

			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}
