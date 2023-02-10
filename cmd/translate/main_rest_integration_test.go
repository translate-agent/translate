package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func attachFile(text []byte, t *testing.T) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", "test.json")
	if err != nil {
		log.Panic(err)
	}

	_, err = part.Write(text)
	if err != nil {
		log.Panic(err)
	}

	return body, writer.FormDataContentType()
}

func Test_UploadTranslationFile_REST(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	addr := "http://localhost:8080/v1/files"

	type args struct {
		language string
		text     []byte
	}

	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "Happy Path",
			args: args{
				language: "lv-lv",
				text: []byte(`{
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
			want: http.StatusOK,
		},
		{
			name: "Missing language",
			args: args{
				text: []byte(`{
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
			want: http.StatusBadRequest,
		},
		{
			name: "Missing file",
			args: args{
				language: "lv-lv",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Invalid language",
			args: args{
				language: "xyz-ZY-Latn",
				text: []byte(`{
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
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, contentType := attachFile(tt.args.text, t)

			req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/%s", addr, tt.args.language), body)
			assert.NoError(t, err)

			req.Header.Add("Content-Type", contentType)
			client := &http.Client{}
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
			}

			assert.EqualValues(t, tt.want, resp.StatusCode)
		})
	}
}
