package exporter

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Exporter is responsible for bringing up a web server that collects metrics
// that have been globally registered via prometheus collectors (e.g., see
// `pkg/collector`).
//
type Exporter struct {
	// ListenAddress is the full address used by prometheus
	// to listen for scraping requests.
	//
	// Examples:
	// - :8080
	// - 127.0.0.2:1313
	//
	listenAddress string

	// TelemetryPath configures the path under which
	// the prometheus metrics are reported.
	//
	// For instance:
	// - /metrics
	// - /telemetry
	//
	telemetryPath string

	// listener is the TCP listener used by the webserver. `nil` if no
	// server is running.
	//
	listener net.Listener

	log logr.Logger
}

// Option.
//
type Option func(e *Exporter)

// New.
//
func New(opts ...Option) (*Exporter, error) {
	defaultLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("zap new development: %w", err)
	}

	e := &Exporter{
		listenAddress: ":9000",
		telemetryPath: "/metrics",
		log:           zapr.NewLogger(defaultLogger.Named("exporter")),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

// Run initiates the HTTP server to serve the metrics.
//
// ps.: this is a BLOCKING method - make sure you either make use of goroutines
// to not block if needed.
//
func (e *Exporter) Run(ctx context.Context) error {
	var err error

	e.listener, err = net.Listen("tcp", e.listenAddress)
	if err != nil {
		return fmt.Errorf("listen on '%s': %w", e.listenAddress, err)
	}

	doneChan := make(chan error, 1)

	go func() {
		defer close(doneChan)

		e.log.WithValues(
			"addr", e.listenAddress,
			"path", e.telemetryPath,
		).Info("listening")

		http.Handle(e.telemetryPath, promhttp.Handler())
		if err := http.Serve(e.listener, nil); err != nil {
			doneChan <- fmt.Errorf(
				"failed listening on address %s: %w",
				e.listenAddress, err,
			)
		}
	}()

	select {
	case err = <-doneChan:
		if err != nil {
			return fmt.Errorf("donechan err: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("ctx err: %w", ctx.Err())
	}

	return nil
}

// Close gracefully closes the tcp listener associated with it.
//
func (e *Exporter) Close() (err error) {
	if e.listener == nil {
		return nil
	}

	e.log.Info("closing")
	if err := e.listener.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}
