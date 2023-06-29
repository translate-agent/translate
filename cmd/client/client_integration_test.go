//go:build integration

//nolint:paralleltest
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/cmd/client/cmd"
	translatesrv "go.expect.digital/translate/cmd/translate/service"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/translate"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const host = "localhost"

var (
	addr   string
	client translatev1.TranslateServiceClient
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	port := mustGetFreePort()
	addr = fmt.Sprintf("%s:%s", host, port)

	viper.Set("service.port", port)
	viper.Set("service.host", host)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		translatesrv.Serve()
	}()

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithBlock(),
	}
	// Wait for the server to start and establish a connection.
	conn, err := grpc.DialContext(ctx, host+":"+port, grpcOpts...)
	if err != nil {
		log.Panicf("create connection to gRPC server: %v", err)
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

	// Close the connection and tracer.
	if err := conn.Close(); err != nil {
		log.Panicf("close gRPC client connection: %v", err)
	}

	os.Exit(code)
}

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

func Test_ListServices_CLI(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "ls",
			"--address", addr,
			"--insecure", "true",
		})

		require.NoError(t, err)
		assert.Contains(t, string(res), "ID")
	})

	t.Run("error, no transport security set", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "ls",
			"--address", addr,
		})

		assert.ErrorContains(t, err, "no transport security set")
		assert.Nil(t, res)
	})
}

func Test_TranslationFileUpload_CLI(t *testing.T) {
	t.Run("OK, file from local path", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		require.NotNil(t, service)

		file, err := os.CreateTemp(t.TempDir(), "test")
		require.NoError(t, err)

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.Id,
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("OK, file from URL", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		require.NotNil(t, service)

		tempDir := t.TempDir()

		file, err := os.CreateTemp(tempDir, "test")
		require.NoError(t, err)

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.Id,
		})

		require.NoError(t, err)
		require.Equal(t, "File uploaded successfully.\n", string(res))

		// upload file using link to previously uploaded translation file.

		res, err = cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", fmt.Sprintf(
				"http://%s/v1/services/%s/files/%s?schema=%s",
				addr, service.Id, lang, translatev1.Schema_JSON_NG_LOCALIZE),
			"--schema", "json_ng_localize",
			"--service", service.Id,
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("error, malformed language", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

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

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "well-formed but unknown")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unrecognized", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--schema", "unrecognized",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err,
			"must be one of: json_ng_localize, json_ngx_translate, go, arb, pot, xliff_12, xliff_2")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unspecified", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--schema", "unspecified",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err,
			"must be one of: json_ng_localize, json_ngx_translate, go, arb, pot, xliff_12, xliff_2")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", "test.json",
			"--schema", "json_ng_localize",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"language\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--schema", "json_ng_localize",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"file\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'service' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", "test.json",
			"--language", "xyz-ZY-Latn",
			"--schema", "json_ng_localize",
		})

		assert.ErrorContains(t, err, "required flag(s) \"service\" not set")
		assert.Nil(t, res)
	})
}

func Test_TranslationFileDownload_CLI(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		require.NotNil(t, service)

		tempDir := t.TempDir()

		file, err := os.CreateTemp(tempDir, "test")
		require.NoError(t, err)

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.Id,
		})

		require.NoError(t, err)
		require.Equal(t, "File uploaded successfully.\n", string(res))

		res, err = cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--schema", "xliff_12",
			"--service", service.Id,
			"--path", tempDir,
		})

		require.NoError(t, err)
		require.Equal(t, "File downloaded successfully.\n", string(res))

		_, err = os.Stat(filepath.Join(tempDir, service.Id+"_"+lang.String()+".xlf"))
		assert.NoError(t, err)
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--schema", "xliff_12",
			"--service", gofakeit.UUID(),
			"--path", t.TempDir(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"language\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--service", gofakeit.UUID(),
			"--path", t.TempDir(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'service' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--schema", "xliff_12",
			"--path", t.TempDir(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"service\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--schema", "xliff_12",
			"--service", gofakeit.UUID(),
		})

		assert.ErrorContains(t, err, "required flag(s) \"path\" not set")
		assert.Nil(t, res)
	})
}

// helpers

func randService(t *testing.T) *translatev1.Service {
	t.Helper()

	return &translatev1.Service{
		Id:   gofakeit.UUID(),
		Name: gofakeit.FirstName(),
	}
}

// createService creates a random service, and calls the CreateService RPC.
func createService(ctx context.Context, t *testing.T) *translatev1.Service {
	t.Helper()

	ctx, span := testutil.Tracer().Start(ctx, "test: create service")
	defer span.End()

	service := randService(t)

	_, err := client.CreateService(ctx, &translatev1.CreateServiceRequest{Service: service})
	require.NoError(t, err, "create test service")

	return service
}

func randUploadData(t *testing.T, schema translatev1.Schema) ([]byte, language.Tag) {
	t.Helper()

	n := gofakeit.IntRange(1, 5)
	lang := language.MustParse(gofakeit.LanguageBCP())
	messages := model.Messages{
		Language: lang,
		Messages: make([]model.Message, 0, n),
	}

	for i := 0; i < n; i++ {
		message := model.Message{ID: gofakeit.SentenceSimple(), Description: gofakeit.SentenceSimple()}
		messages.Messages = append(messages.Messages, message)
	}

	data, err := translate.MessagesToData(schema, messages)
	require.NoError(t, err, "convert rand messages to serialized data")

	return data, lang
}
