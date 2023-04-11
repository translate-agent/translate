package main

import (
	"fmt"
	"strings"

	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/cmd/client/cmd"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func Test_ServiceLs(t *testing.T) {
	t.Skip()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",
		})

		if !assert.NoError(t, err) {
			return
		}

		assert.Contains(t, string(res), "TOTAL")
	})

	t.Run("error, no transport security set", func(t *testing.T) {
		t.Parallel()

		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", fmt.Sprintf("%s:%s", host, port),
		})

		assert.ErrorContains(t, err, "no transport security set")
		assert.Nil(t, res)
	})
}

func Test_ServiceUpload(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		file, err := os.CreateTemp("./", "test")
		if !assert.NoError(t, err) {
			return
		}

		defer os.Remove(file.Name())

		if _, err = file.Write([]byte(`{
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
	 }`)); !assert.NoError(t, err) {
			return
		}

		wd, err := os.Getwd()
		if !assert.NoError(t, err) {
			return
		}

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "lv-lv",
			"-p", fmt.Sprintf("%s/%s", wd, strings.TrimPrefix(file.Name(), ".")),
			"-s", fmt.Sprintf("%d", translatev1.Schema_GO),
		})

		assert.NoError(t, err)
		assert.Contains(t, string(res), "uploaded successfully")
	})
}
