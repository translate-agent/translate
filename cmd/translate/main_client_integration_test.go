//nolint:paralleltest
package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/cmd/client/cmd"
)

func Test_ServiceLsCmd(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",
		})

		require.NoError(t, err)
		assert.Contains(t, string(res), "ID")
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

func Test_ServiceUploadCmd(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "test")
		require.NoError(t, err)

		_, err = file.Write([]byte(`
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
	 }`))

		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "lv-lv",
			"-f", file.Name(),
			"-s", "ng_localise",
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("error, malformed language tag", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "test")

		require.NoError(t, err)

		_, err = file.Write([]byte(`
		{
		  "locale": "xyz-ZY-Latn",
		  "translations": {
			"Hello": "Bonjour",
			"Welcome": "Bienvenue"
		  }
		}`))

		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", file.Name(),
			"-s", "ng_localise",
		})

		assert.ErrorContains(t, err, "well-formed but unknown")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unrecognized", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", "test.json",
			"-s", "unrecognized",
		})

		assert.ErrorContains(t, err, "flag: must be one of \"ng_localise\", \"ngx_translate\", \"go\", \"arb")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", "test.json",
		})

		assert.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-f", "test.json",
			"-s", "ng_localise",
		})

		assert.ErrorContains(t, err, "required flag(s) \"language\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-s", "ng_localise",
		})

		assert.ErrorContains(t, err, "required flag(s) \"file\" not set")
		assert.Nil(t, res)
	})
}
