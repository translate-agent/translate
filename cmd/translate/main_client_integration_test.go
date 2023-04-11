//nolint:paralleltest
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/cmd/client/cmd"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func Test_ServiceLs(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
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
		file, err := os.CreateTemp(t.TempDir(), "test")
		if !assert.NoError(t, err) {
			return
		}

		if _, err = file.Write([]byte(`
		{
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

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "lv-lv",
			"-p", file.Name(),
			"-s", fmt.Sprintf("%d", translatev1.Schema_GO),
		})

		assert.NoError(t, err)
		assert.Contains(t, string(res), filepath.Base(file.Name())+" uploaded successfully")
	})

	t.Run("error, malformed language tag", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "test")
		if !assert.NoError(t, err) {
			return
		}

		if _, err = file.Write([]byte(`
		{
		  "locale": "xyz-ZY-Latn",
		  "translations": {
			"Hello": "Bonjour",
			"Welcome": "Bienvenue"
		  }
		}`)); !assert.NoError(t, err) {
			return
		}

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "lv-lx",
			"-p", file.Name(),
			"-s", fmt.Sprintf("%d", translatev1.Schema_GO),
		})

		assert.ErrorContains(t, err, "well-formed but unknown")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "test")
		if !assert.NoError(t, err) {
			return
		}

		if _, err = file.Write([]byte(`
		{
		  "locale": "xyz-ZY-Latn",
		  "translations": {
			"Hello": "Bonjour",
			"Welcome": "Bienvenue"
		  }
		}`)); !assert.NoError(t, err) {
			return
		}

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "lv-lx",
			"-p", file.Name(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})
}
