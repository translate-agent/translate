//go:build integration

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil"
	"go.expect.digital/translate/pkg/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const host = "localhost"

var (
	port       string
	client     translatev1.TranslateServiceClient
	testTracer oteltrace.Tracer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	tp, err := tracer.TracerProvider(ctx)
	if err != nil {
		log.Panicf("set tracer provider: %v", err)
	}

	defer tp.ForceFlush(ctx)

	testTracer = tp.Tracer("go.expect.digital/translate/cmd/translate")

	// start the translate service

	port = mustGetFreePort()

	viper.Set("service.port", port)
	viper.Set("service.host", host)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		main()
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

	tp.ForceFlush(ctx)

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

// trace provides integration test span.
func trace(ctx context.Context, t *testing.T) context.Context {
	return testutil.Trace(ctx, t, testTracer)
}

// subtest runs parallel subtest with a trace instrumentation.
func subtest(ctx context.Context, t *testing.T, name string, f func(context.Context, *testing.T)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()

		ctx = trace(ctx, t)

		f(ctx, t)
	})
}
