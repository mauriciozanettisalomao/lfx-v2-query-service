// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	querysvcapi "github.com/linuxfoundation/lfx-v2-query-service/cmd/query_svc"
	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/port"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/mock"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/nats"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/opensearch"
	logging "github.com/linuxfoundation/lfx-v2-query-service/pkg/log"
	"goa.design/clue/debug"
)

const (
	defaultPort = "8080"
	// gracefulShutdownSeconds should be higher than NATS client
	// request timeout, and lower than the pod or liveness probe's
	// terminationGracePeriodSeconds.
	gracefulShutdownSeconds = 25
)

func init() {
	// slog is the standard library logger, we use it to log errors and
	logging.InitStructureLogConfig()
}

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		dbgF = flag.Bool("d", false, "enable debug logging")
		port = flag.String("p", defaultPort, "listen port")
		bind = flag.String("bind", "*", "interface to bind on")
	)
	flag.Usage = func() {
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	ctx := context.Background()
	slog.InfoContext(ctx, "Starting query service",
		"bind", *bind,
		"http-port", *port,
		"graceful-shutdown-seconds", gracefulShutdownSeconds,
	)

	// Initialize the resource searcher based on configuration
	resourceSearcher := searcherImpl(ctx)
	accessControlChecker := accessControlCheckerImpl(ctx)

	// Initialize the services.
	var (
		querySvcSvc querysvc.Service
	)
	{
		querySvcSvc = querysvcapi.NewQuerySvc(resourceSearcher, accessControlChecker)
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	querySvcEndpoints := querysvc.NewEndpoints(querySvcSvc)
	querySvcEndpoints.Use(debug.LogPayloads())

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	// Setup the JWT authentication which validates and parses the JWT token.
	querysvcapi.SetupJWTAuth(ctx)

	// Start the servers and send errors (if any) to the error channel.
	addr := ":" + *port
	if *bind != "*" {
		addr = *bind + ":" + *port
	}

	handleHTTPServer(ctx, addr, querySvcEndpoints, &wg, errc, *dbgF)

	// Wait for signal.
	slog.InfoContext(ctx, "received shutdown signal, stopping servers",
		"signal", <-errc,
	)

	// Send cancellation signal to the goroutines.
	cancel()

	// Create a timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), gracefulShutdownSeconds*time.Second)
	defer shutdownCancel()

	// Gracefully close the access control checker
	go func() {
		if accessControlChecker != nil {
			slog.InfoContext(shutdownCtx, "closing access control checker")
			if err := accessControlChecker.Close(); err != nil {
				slog.ErrorContext(shutdownCtx, "failed to close access control checker", "error", err)
			}
		}
	}()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.InfoContext(ctx, "graceful shutdown completed")
	case <-shutdownCtx.Done():
		slog.WarnContext(ctx, "graceful shutdown timed out")
	}

	slog.InfoContext(ctx, "exited")
}

func searcherImpl(ctx context.Context) port.ResourceSearcher {

	var (
		resourceSearcher port.ResourceSearcher
		err              error
	)

	// Search source implementation configuration
	searchSource := os.Getenv("SEARCH_SOURCE")
	if searchSource == "" {
		searchSource = "opensearch"
	}

	opensearchURL := os.Getenv("OPENSEARCH_URL")
	if opensearchURL == "" {
		opensearchURL = "http://localhost:9200"
	}

	opensearchIndex := os.Getenv("OPENSEARCH_INDEX")
	if opensearchIndex == "" {
		opensearchIndex = "resources"
	}

	switch searchSource {
	case "mock":
		slog.InfoContext(ctx, "initializing mock resource searcher")
		resourceSearcher = mock.NewMockResourceSearcher()

	case "opensearch":
		slog.InfoContext(ctx, "initializing opensearch resource searcher",
			"url", opensearchURL,
			"index", opensearchIndex,
		)
		opensearchConfig := opensearch.Config{
			URL:   opensearchURL,
			Index: opensearchIndex,
		}

		resourceSearcher, err = opensearch.NewSearcher(ctx, opensearchConfig)
		if err != nil {
			log.Fatalf("failed to initialize OpenSearch searcher: %v", err)
		}

	default:
		log.Fatalf("unsupported search implementation: %s", searchSource)
	}

	return resourceSearcher

}

func accessControlCheckerImpl(ctx context.Context) port.AccessControlChecker {

	var (
		accessControlChecker port.AccessControlChecker
		err                  error
	)

	// Access control implementation configuration
	accessControlSource := os.Getenv("ACCESS_CONTROL_SOURCE")
	if accessControlSource == "" {
		accessControlSource = "nats"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	natsTimeout := os.Getenv("NATS_TIMEOUT")
	if natsTimeout == "" {
		natsTimeout = "10s"
	}
	natsTimeoutDuration, err := time.ParseDuration(natsTimeout)
	if err != nil {
		log.Fatalf("invalid NATS timeout duration: %v", err)
	}

	natsMaxReconnect := os.Getenv("NATS_MAX_RECONNECT")
	if natsMaxReconnect == "" {
		natsMaxReconnect = "3"
	}
	natsMaxReconnectInt, err := strconv.Atoi(natsMaxReconnect)
	if err != nil {
		log.Fatalf("invalid NATS max reconnect value %s: %v", natsMaxReconnect, err)
	}

	natsReconnectWait := os.Getenv("NATS_RECONNECT_WAIT")
	if natsReconnectWait == "" {
		natsReconnectWait = "2s"
	}
	natsReconnectWaitDuration, err := time.ParseDuration(natsReconnectWait)
	if err != nil {
		log.Fatalf("invalid NATS reconnect wait duration %s : %v", natsReconnectWait, err)
	}

	//natsReconnectWait := flag.Duration("nats-reconnect-wait", 2*time.Second, "NATS reconnection wait time")

	// Initialize the access control checker based on configuration
	switch accessControlSource {
	case "mock":
		slog.InfoContext(ctx, "initializing mock access control checker")
		accessControlChecker = mock.NewMockAccessControlChecker()

	case "nats":
		slog.InfoContext(ctx, "initializing NATS access control checker")
		natsConfig := nats.Config{
			URL:           natsURL,
			Timeout:       natsTimeoutDuration,
			MaxReconnect:  natsMaxReconnectInt,
			ReconnectWait: natsReconnectWaitDuration,
		}

		accessControlChecker, err = nats.NewAccessControlChecker(ctx, natsConfig)
		if err != nil {
			log.Fatalf("failed to initialize NATS access control checker: %v", err)
		}

	default:
		log.Fatalf("unsupported access control implementation: %s", accessControlSource)
	}

	return accessControlChecker
}
