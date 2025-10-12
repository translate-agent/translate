//go:build integration

package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/viper"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const host = "localhost"

var (
	addr, port string
	client     translatev1.TranslateServiceClient
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) (code int) {
	// start the translate service
	port = mustGetFreePort()
	addr = net.JoinHostPort(host, port)

	viper.Set("service.port", port)
	viper.Set("service.host", host)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		main()
	}()

	// ensure gRPC server is listening before running tests
	// wait for 300ms (6x50ms) for successful TCP connection
	dialer := new(net.Dialer)
	ctx := context.Background()

	for range 6 {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		defer conn.Close()

		break
	}

	// set up gRPC client
	conn, err := grpc.NewClient(
		net.JoinHostPort(host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Panicf("create gRPC client: %v", err)
	}

	client = translatev1.NewTranslateServiceClient(conn)

	// Run the tests.
	code = m.Run()
	// Send soft kill (termination) signal to process.
	err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		log.Panicf("send termination signal: %v", err)
	}
	// Wait for main() to finish cleanup.
	wg.Wait()

	// Close the connection and tracer.
	err = conn.Close()
	if err != nil {
		log.Panicf("close gRPC client connection: %v", err)
	}

	return code
}

func mustGetFreePort() string {
	// Listen on port 0 to have the operating system allocate an available port.
	lc := new(net.ListenConfig)

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

// ptr returns pointer to the passed in value.
func ptr[T any](v T) *T {
	return &v
}
