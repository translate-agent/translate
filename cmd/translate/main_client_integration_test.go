//go:build integration

//nolint:paralleltest
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/cmd/client/cmd"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func Test_ListServices(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", addr,
			"-i", "true",
		})

		require.NoError(t, err)
		assert.Contains(t, string(res), "ID")
	})

	t.Run("error, no transport security set", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", addr,
		})

		assert.ErrorContains(t, err, "no transport security set")
		assert.Nil(t, res)
	})
}

func Test_TranslationFileUpload(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		service := createService(context.Background(), t)
		require.NotNil(t, service)

		file, err := os.CreateTemp(t.TempDir(), "test")
		require.NoError(t, err)

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-l", lang.String(),
			"-f", file.Name(),
			"-s", "json_ng_localize",
			"-u", service.Id,
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
			"-a", addr,
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", file.Name(),
			"-s", "json_ng_localize",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "well-formed but unknown")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unrecognized", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", "test.json",
			"-s", "unrecognized",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err,
			"must be one of: json_ng_localize, json_ngx_translate, go, arb, pot, xliff_12, xliff_2")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unspecified", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", "test.json",
			"-s", "unspecified",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err,
			"must be one of: json_ng_localize, json_ngx_translate, go, arb, pot, xliff_12, xliff_2")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-f", "test.json",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-f", "test.json",
			"-s", "json_ng_localize",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"language\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-l", "xyz-ZY-Latn",
			"-s", "json_ng_localize",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"file\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'serviceID' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-f", "test.json",
			"-l", "xyz-ZY-Latn",
			"-s", "json_ng_localize",
		})

		assert.ErrorContains(t, err, "required flag(s) \"serviceID\" not set")
		assert.Nil(t, res)
	})
}

func Test_TranslationFileDownload(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		service := createService(context.Background(), t)
		require.NotNil(t, service)

		tempDir := t.TempDir()

		file, err := os.CreateTemp(tempDir, "test")
		require.NoError(t, err)

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams([]string{
			"service", "upload",
			"-a", addr,
			"-i", "true",

			"-l", lang.String(),
			"-f", file.Name(),
			"-s", "json_ng_localize",
			"-u", service.Id,
		})

		require.NoError(t, err)
		require.Equal(t, "File uploaded successfully.\n", string(res))

		res, err = cmd.ExecuteWithParams([]string{
			"service", "download",
			"-a", addr,
			"-i", "true",

			"-l", lang.String(),
			"-s", "xliff_12",
			"-u", service.Id,
			"-p", tempDir,
		})

		require.NoError(t, err)
		require.Equal(t, "File downloaded successfully.\n", string(res))

		_, err = os.Stat(filepath.Join(tempDir, service.Id+"_"+lang.String()+".xlf"))
		assert.NoError(t, err)
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "download",
			"-a", addr,
			"-i", "true",

			"-s", "xliff_12",
			"-u", gofakeit.UUID(),
			"-p", t.TempDir(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"language\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "download",
			"-a", addr,
			"-i", "true",

			"-l", "lv-lv",
			"-u", gofakeit.UUID(),
			"-p", t.TempDir(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'serviceID' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "download",
			"-a", addr,
			"-i", "true",

			"-l", "lv-lv",
			"-s", "xliff_12",
			"-p", t.TempDir(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"serviceID\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "download",
			"-a", addr,
			"-i", "true",

			"-l", "lv-lv",
			"-s", "xliff_12",
			"-u", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"path\" not set")
		assert.Nil(t, res)
	})
}
