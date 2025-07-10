package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	querysvcapi "github.com/linuxfoundation/lfx-v2-query-service"
	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/elasticsearch"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/mock"
	logging "github.com/linuxfoundation/lfx-v2-query-service/pkg/log"

	"goa.design/clue/debug"
	"goa.design/clue/log"
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

		// Search implementation configuration
		searchImpl = flag.String("search-impl", "mock", "Search implementation to use (mock, elasticsearch)")

		// Elasticsearch configuration flags
		esURL      = flag.String("es-url", "http://localhost:9200", "Elasticsearch URL")
		esUsername = flag.String("es-username", "", "Elasticsearch username")
		esPassword = flag.String("es-password", "", "Elasticsearch password")
		esIndex    = flag.String("es-index", "lfx-resources", "Elasticsearch index name")
	)
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatTerminal
	}
	ctx := log.Context(context.Background(), log.WithFormat(format))
	if *dbgF {
		ctx = log.Context(ctx, log.WithDebug())
		log.Debugf(ctx, "debug logs enabled")
	}
	log.Print(ctx, log.KV{K: "http-port", V: *httpPortF})

	// Initialize the resource searcher based on configuration
	var resourceSearcher domain.ResourceSearcher
	var err error

	switch *searchImpl {
	case "mock":
		log.Printf(ctx, "initializing mock resource searcher")
		resourceSearcher = mock.NewMockResourceSearcher()

	case "elasticsearch":
		log.Printf(ctx, "initializing elasticsearch resource searcher")
		esConfig := elasticsearch.Config{
			URL:      *esURL,
			Username: *esUsername,
			Password: *esPassword,
			Index:    *esIndex,
		}

		resourceSearcher, err = elasticsearch.NewElasticsearchSearcherFromConfig(esConfig)
		if err != nil {
			log.Fatalf(ctx, err, "failed to initialize Elasticsearch searcher")
		}

	default:
		log.Fatalf(ctx, fmt.Errorf("unsupported search implementation: %s", *searchImpl), "valid options: mock, elasticsearch")
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
		querySvcEndpoints.Use(log.Endpoint)
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
				log.Fatalf(ctx, err, "invalid URL %#v\n", addr)
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
					log.Fatalf(ctx, err, "invalid URL %#v\n", u.Host)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "8080")
			}
			handleHTTPServer(ctx, u, querySvcEndpoints, &wg, errc, *dbgF)
		}

	default:
		log.Fatal(ctx, fmt.Errorf("invalid host argument: %q (valid hosts: localhost)", *hostF))
	}

	// Wait for signal.
	log.Printf(ctx, "exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	log.Printf(ctx, "exited")
}
