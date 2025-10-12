//go:build integration

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/spf13/viper"
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
		dialer := new(net.Dialer)

		conn, err := dialer.DialContext(context.Background(), "tcp", addr)
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
		err := closeConn()
		if err != nil {
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
	var lc net.ListenConfig

	l, err := lc.Listen(context.Background(), "tcp", host+":0")
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

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "ls",
			"--address", addr,
			"--insecure", "true",
		})
		if err != nil {
			t.Error(err)
			return
		}

		if !bytes.Contains(output, []byte("ID")) {
			t.Errorf("want output to contain 'ID', got '%s'", string(output))
		}
	})

	t.Run("error, no transport security set", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "ls",
			"--address", addr,
		})

		if want := "no transport security set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})
}

//nolint:gocognit,cyclop
func Test_TranslationFileUpload_CLI(t *testing.T) {
	t.Parallel()

	t.Run("OK, file from local path", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Error(err)
			return
		}

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}
	})

	t.Run("OK, with local file and original flag", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Error(err)
			return
		}

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--original", "true",
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}
	})

	t.Run("OK, with local file, original=true populate=false", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Error(err)
			return
		}

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
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
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}
	})

	t.Run("OK, file from URL", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		tempDir := t.TempDir()

		file, err := os.CreateTemp(tempDir, "test")
		if err != nil {
			t.Error(err)
			return
		}

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}

		// upload file using link to previously uploaded translation file.
		output, err = cmd.ExecuteWithParams(ctx, []string{
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
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}
	})

	// Translation has language tag, but CLI parameter 'language' is not set.
	t.Run("OK, local without lang parameter", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Error(err)
			return
		}

		// Ng localise schema has language tag in the file.
		data, _ := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}
	})

	t.Run("error, malformed language", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Error(err)
			return
		}

		_, err = file.WriteString(`
		{
			"locale": "xyz-ZY-Latn",
			"translations": {
			"Hello": "Bonjour",
			"Welcome": "Bienvenue"
			}
		}`)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", gofakeit.UUID(),
		})

		if want := "well-formed but unknown"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'schema' unrecognized", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--schema", "unrecognized",
			"--service", gofakeit.UUID(),
		})

		want := "must be one of: json_ng_localize, json_ngx_translate, go, arb, po, xliff_12, xliff_2"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'schema' unspecified", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--schema", "unspecified",
			"--service", gofakeit.UUID(),
		})

		want := "must be one of: json_ng_localize, json_ngx_translate, go, arb, po, xliff_12, xliff_2"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--file", "test.json",
			"--service", gofakeit.UUID(),
		})
		if want := "required flag(s) \"schema\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	// Translation does not have language tag, and CLI parameter 'language' is not set.
	t.Run("error, language could not be determined", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		file, err := os.CreateTemp(t.TempDir(), "test")
		if err != nil {
			t.Error(err)
			return
		}

		// ngx translate schema does not have language tag in the file.
		data, _ := randUploadData(t, translatev1.Schema_JSON_NGX_TRANSLATE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})

		if want := "no language is set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", "xyz-ZY-Latn",
			"--schema", "json_ng_localize",
			"--service", gofakeit.UUID(),
		})

		if want := "required flag(s) \"file\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'service' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--file", "test.json",
			"--language", "xyz-ZY-Latn",
			"--schema", "json_ng_localize",
		})

		if want := "required flag(s) \"service\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got %s", output)
		}
	})
}

//nolint:gocognit
func Test_TranslationFileDownload_CLI(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		service := createService(ctx, t)
		if service == nil {
			t.Error("want service, got nil")
			return
		}

		tempDir := t.TempDir()

		file, err := os.CreateTemp(tempDir, "test")
		if err != nil {
			t.Error(err)
			return
		}

		data, lang := randUploadData(t, translatev1.Schema_JSON_NG_LOCALIZE)

		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
			return
		}

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "upload",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--file", file.Name(),
			"--schema", "json_ng_localize",
			"--service", service.GetId(),
		})
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File uploaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}

		output, err = cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", lang.String(),
			"--schema", "xliff_12",
			"--service", service.GetId(),
			"--path", tempDir,
		})
		if err != nil {
			t.Error(err)
			return
		}

		if want := "File downloaded successfully.\n"; string(output) != want {
			t.Errorf("want output '%s', got '%s'", want, output)
		}

		_, err = os.Stat(filepath.Join(tempDir, service.GetId()+"_"+lang.String()+".xlf"))
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("error, path parameter 'language' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--schema", "xliff_12",
			"--service", gofakeit.UUID(),
			"--path", t.TempDir(),
		})

		if want := "required flag(s) \"language\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'schema' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--service", gofakeit.UUID(),
			"--path", t.TempDir(),
		})

		if want := "required flag(s) \"schema\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'service' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--schema", "xliff_12",
			"--path", t.TempDir(),
		})

		if want := "required flag(s) \"service\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
	})

	t.Run("error, path parameter 'path' missing", func(t *testing.T) {
		t.Parallel()
		ctx, _ := testutil.Trace(t)

		output, err := cmd.ExecuteWithParams(ctx, []string{
			"service", "download",
			"--address", addr,
			"--insecure", "true",

			"--language", "lv-lv",
			"--schema", "xliff_12",
			"--service", gofakeit.UUID(),
		})

		if want := "required flag(s) \"path\" not set"; !strings.Contains(err.Error(), want) {
			t.Errorf("want error '%s' to contain '%s'", err, want)
		}

		if len(output) > 0 {
			t.Errorf("want empty output, got '%s'", output)
		}
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
	if err != nil {
		t.Error(err)
		return nil
	}

	return service
}

func randUploadData(t *testing.T, schema translatev1.Schema) ([]byte, language.Tag) {
	t.Helper()

	translation := rand.ModelTranslation(3, nil, rand.WithSimpleMF2Messages())

	data, err := server.TranslationToData(schema, translation)
	if err != nil {
		t.Error(err)
		return nil, language.Und
	}

	return data, translation.Language
}
