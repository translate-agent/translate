//go:build integration

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/cmd/client/cmd"
	translatesrv "go.expect.digital/translate/cmd/translate/service"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/server"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/testutil/rand"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const host = "localhost"

var (
	port = mustGetFreePort()
	addr = net.JoinHostPort(host, port)

	// client is a gRPC client for the translate service, used in tests to create resources, before testing the CLI.
	client translatev1.TranslateServiceClient
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	viper.Set("service.port", port)
	viper.Set("service.host", host)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		translatesrv.Serve()
	}()

	// ensure gRPC server is listening before running tests
	// wait for 300ms (6x50ms) for successful TCP connection
	for range 6 {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		defer conn.Close()

		break
	}

	// set up gRPC client
	closeConn := setUpClient()

	defer func() {
		if err := closeConn(); err != nil {
			log.Panicf("close client connection: %v", err)
		}
	}()

	// Run the tests.
	code := m.Run()

	// Send soft kill (termination) signal to process.
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		log.Panicf("send termination signal: %v", err)
	}

	// Wait for grpc server to stop.
	wg.Wait()

	return code
}

// setUpClient creates a gRPC client connection to the translate service and assigns it to the client variable.
func setUpClient() (closeConn func() error) {
	conn, err := grpc.NewClient(
		net.JoinHostPort(host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Panicf("create gRPC client: %v", err)
	}

	client = translatev1.NewTranslateServiceClient(conn)

	return conn.Close
}

func mustGetFreePort() string {
	// Listen on port 0 to have the operating system allocate an available port.
	l, err := net.Listen("tcp", host+":0")
	if err != nil {
		log.Panicf("get free port: %v", err)
	}
	defer l.Close()

	// Get the port number from the address that the Listener is listening on.
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		log.Panic("get free port address")
	}

	return strconv.Itoa(addr.Port)
}

func Test_ListServices_CLI(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "ls",
			"--address", addr,
		})

		require.ErrorContains(t, err, "no transport security set")
		assert.Nil(t, res)
	})
}

func Test_TranslationFileUpload_CLI(t *testing.T) {
	t.Parallel()

	t.Run("OK, file from local path", func(t *testing.T) {
		t.Parallel()
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
			"--service", service.GetId(),
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("OK, with local file and original flag", func(t *testing.T) {
		t.Parallel()
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
			"--original", "true",
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("OK, with local file, original=true populate=false", func(t *testing.T) {
		t.Parallel()
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
			"--original", "true",
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
			"--populate_translations", "false",
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("OK, file from URL", func(t *testing.T) {
		t.Parallel()
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
			"--service", service.GetId(),
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
				addr, service.GetId(), lang, translatev1.Schema_JSON_NG_LOCALIZE),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	// Translation has language tag, but CLI parameter 'language' is not set.
	t.Run("OK, local without lang parameter", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		require.NotNil(t, service)

		file, err := os.CreateTemp(t.TempDir(), "test")
		require.NoError(t, err)

		// Ng localise schema has language tag in the file.
		data, _ := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})

		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully.\n", string(res))
	})

	t.Run("error, malformed language", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		file, err := os.CreateTemp(t.TempDir(), "test")

		require.NoError(t, err)

		_, err = file.WriteString(`
		{
			"locale": "xyz-ZY-Latn",
			"translations": {
			"Hello": "Bonjour",
			"Welcome": "Bienvenue"
			}
		}`)

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

		require.ErrorContains(t, err, "well-formed but unknown")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unrecognized", func(t *testing.T) {
		t.Parallel()
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

		require.ErrorContains(t, err,
			"must be one of: json_ng_localize, json_ngx_translate, go, arb, po, xliff_12, xliff_2")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' unspecified", func(t *testing.T) {
		t.Parallel()
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

		require.ErrorContains(t, err,
			"must be one of: json_ng_localize, json_ngx_translate, go, arb, po, xliff_12, xliff_2")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--service", gofakeit.UUID(),
		})

		require.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	// Translation does not have language tag, and CLI parameter 'language' is not set.
	t.Run("error, language could not be determined", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		require.NotNil(t, service)

		file, err := os.CreateTemp(t.TempDir(), "test")
		require.NoError(t, err)

		// ngx translate schema does not have language tag in the file.
		data, _ := randUploadData(t, translatev1.Schema_JSON_NGX_TRANSLATE)

		_, err = file.Write(data)
		require.NoError(t, err)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})

		require.ErrorContains(t, err, "no language is set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--schema", "json_ng_localize",
			"--service", gofakeit.UUID(),
		})

		require.ErrorContains(t, err, "required flag(s) \"file\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'service' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", "test.json",
			"--language", "xyz-ZY-Latn",
			"--schema", "json_ng_localize",
		})

		require.ErrorContains(t, err, "required flag(s) \"service\" not set")
		assert.Nil(t, res)
	})
}

func Test_TranslationFileDownload_CLI(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
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
			"--service", service.GetId(),
		})

		require.NoError(t, err)
		require.Equal(t, "File uploaded successfully.\n", string(res))

		res, err = cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--schema", "xliff_12",
			"--service", service.GetId(),
			"--path", tempDir,
		})

		require.NoError(t, err)
		require.Equal(t, "File downloaded successfully.\n", string(res))

		_, err = os.Stat(filepath.Join(tempDir, service.GetId()+"_"+lang.String()+".xlf"))
		require.NoError(t, err)
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--schema", "xliff_12",
			"--service", gofakeit.UUID(),
			"--path", t.TempDir(),
		})

		require.ErrorContains(t, err, "required flag(s) \"language\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--service", gofakeit.UUID(),
			"--path", t.TempDir(),
		})

		require.ErrorContains(t, err, "required flag(s) \"schema\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'service' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--schema", "xliff_12",
			"--path", t.TempDir(),
		})

		require.ErrorContains(t, err, "required flag(s) \"service\" not set")
		assert.Nil(t, res)
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		res, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--schema", "xliff_12",
			"--service", gofakeit.UUID(),
		})

		require.ErrorContains(t, err, "required flag(s) \"path\" not set")
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

	translation := rand.ModelTranslation(3, nil, rand.WithSimpleMF2Messages())

	data, err := server.TranslationToData(schema, translation)
	require.NoError(t, err, "convert rand translation to serialized data")

	return data, translation.Language
}
