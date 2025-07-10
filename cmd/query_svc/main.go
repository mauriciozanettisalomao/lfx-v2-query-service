// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	querysvcapi "github.com/linuxfoundation/lfx-v2-query-service"
	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/mock"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/opensearch"
	logging "github.com/linuxfoundation/lfx-v2-query-service/pkg/log"

	"goa.design/clue/debug"
)

func init() {
	// slog is the standard library logger, we use it to log errors and
	logging.InitStructureLogConfig()
}

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		hostF     = flag.String("host", "localhost", "Server host (valid values: localhost)")
		domainF   = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
		httpPortF = flag.String("http-port", "", "HTTP port (overrides host HTTP port specified in service design)")
		secureF   = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")

		// Search source implementation configuration
		searchSource = flag.String("search-source", "mock", "Search implementation to use (mock, elasticsearch)")

		// OpenSearch configuration flags
		opensearchURL   = flag.String("opensearch-url", "http://localhost:9200", "Opensearch URL")
		opensearchIndex = flag.String("opensearch-index", "lfx-resources", "Opensearch index name")
	)
	flag.Parse()

	ctx := context.Background()
	slog.InfoContext(ctx, "Starting query service",
		"host", *hostF,
		"http-port", *httpPortF,
	)

	// Initialize the resource searcher based on configuration
	var (
		resourceSearcher domain.ResourceSearcher
		err              error
	)

	switch *searchSource {
	case "mock":
		slog.InfoContext(ctx, "initializing mock resource searcher")
		resourceSearcher = mock.NewMockResourceSearcher()

	case "opensearch":
		slog.InfoContext(ctx, "initializing opensearch resource searcher")
		opensearchConfig := opensearch.Config{
			URL:   *opensearchURL,
			Index: *opensearchIndex,
		}

		resourceSearcher, err = opensearch.NewSearcher(ctx, opensearchConfig)
		if err != nil {
			log.Fatalf("failed to initialize OpenSearch searcher: %v", err)
		}

	default:
		log.Fatalf("unsupported search implementation: %s", *searchSource)
	}

	// Initialize the services.
	var (
		querySvcSvc querysvc.Service
	)
	{
		querySvcSvc = querysvcapi.NewQuerySvc(resourceSearcher)
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var (
		querySvcEndpoints *querysvc.Endpoints
	)
	{
		querySvcEndpoints = querysvc.NewEndpoints(querySvcSvc)
		querySvcEndpoints.Use(debug.LogPayloads())
	}

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

	// Start the servers and send errors (if any) to the error channel.
	switch *hostF {
	case "localhost":
		{
			addr := "http://localhost:8080"
			u, err := url.Parse(addr)
			if err != nil {
				log.Fatalf("invalid URL %#v, error: %v\n", addr, err)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					log.Fatalf("invalid URL %#v, error: %v\n", u.Host, err)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "8080")
			}
			handleHTTPServer(ctx, u, querySvcEndpoints, &wg, errc, *dbgF)
		}

	default:
		log.Fatalf("invalid host argument: %v", *hostF)
	}

	// Wait for signal.
	slog.InfoContext(ctx, "received shutdown signal, stopping servers",
		"signal", <-errc,
	)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	slog.InfoContext(ctx, "exited")
}
