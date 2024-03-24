package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koenbollen/logging"

	_ "modernc.org/sqlite"
)

// Route a logical collection of logic of the API
type Route func(context.Context, *http.ServeMux, *Dependencies) error

// SetupRoutes will combine all the routes into a simple http.ServeMux and
// add a health check route
func SetupRoutes(ctx context.Context, deps *Dependencies, routes ...Route) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		logging.IgnoreRequest(r)
		if deps.DB.PingContext(r.Context()) != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	for _, route := range routes {
		if err := route(ctx, mux, deps); err != nil {
			return nil, fmt.Errorf("failed to add route: %w", err)
		}
	}
	return mux, nil
}

// Main will handle the setup of dependencies, routes and the http server. Start
// the server and wait for a the context to be cancelled to shutdown the server.
func Main(ctx context.Context, component string, routes ...Route) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := logging.New(ctx, "go-tested-api-with-sqlite", component)
	ctx = logging.WithLogger(ctx, logger)
	logger.Info("init")
	defer logger.Info("fin")

	config, err := ConfigFromEnv(ctx)
	if err != nil {
		logger.Error("failed to load config", "err", err)
		return
	}

	deps, err := Setup(ctx, config)
	if err != nil {
		logger.Error("failed to setup dependencies", "err", err)
		return
	}

	mux, err := SetupRoutes(ctx, deps, routes...)
	if err != nil {
		logger.Error("failed to setup routes", "err", err)
		return
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: logging.Middleware(mux, logger),

		BaseContext: func(net.Listener) context.Context {
			return ctx
		},

		// Better timeouts for load balancers:
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  650 * time.Second,
	}
	if addr, ok := os.LookupEnv("ADDR"); ok {
		server.Addr = addr
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to listen or serve", "err", err)
			return
		}
	}()
	logger.Info("listening", "addr", server.Addr)

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 29*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown server", "err", err)
		return
	}
}
