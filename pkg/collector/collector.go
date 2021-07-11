package collector

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

// CountryMapper defines the signature of a function that given an IP,
// translates it into a country name.
//
//	f(ip) -> CN
//
type CountryMapper func(net.IP) (string, error)

// Collector implements the prometheus Collector interface, providing monero
// metrics whenever a prometheus scrape is received.
//
type Collector struct {
	// client is a Go client that communicated with a `monero` daemon via
	// plain HTTP(S) RPC.
	//
	client *daemon.Client

	// countryMapper is a function that knows how to translate IPs to
	// country codes.
	//
	// optional: if nil, no country-mapping will take place.
	//
	countryMapper CountryMapper

	log logr.Logger
}

// ensure that we implement prometheus' collector interface.
//
var _ prometheus.Collector = &Collector{}

// Option is a type used by functional arguments to mutate the collector to
// override default behavior.
//
type Option func(c *Collector)

// WithCountryMapper is a functional argument that overrides the default no-op
// country mapper.
//
func WithCountryMapper(v CountryMapper) func(c *Collector) {
	return func(c *Collector) {
		c.countryMapper = v
	}
}

func defaultCountryMapper(_ net.IP) (string, error) {
	return "unknown", nil
}

// Register registers this collector with the global prometheus collectors
// registry making it available for an exporter to collect our metrics.
//
func Register(client *daemon.Client, opts ...Option) error {
	defaultLogger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("zap new development: %w", err)
	}

	c := &Collector{
		client:        client,
		log:           zapr.NewLogger(defaultLogger),
		countryMapper: defaultCountryMapper,
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := prometheus.Register(c); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	return nil
}

// CollectFunc defines a standardized signature for functions that want to
// expose metrics for collection.
//
type CollectFunc func(ctx context.Context, ch chan<- prometheus.Metric) error

// Describe implements the Describe function of the Collector interface.
//
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	// Because we can present the description of the metrics at collection
	// time, we don't need to write anything to the channel.
}

type CustomCollector interface {
	Name() string
	Collect(ctx context.Context) error
}

// Collect implements the Collect function of the Collector interface.
//
// Here is where all of the calls to a monero rpc endpoint is made, each being
// wrapped in its own function, all being called concurrently.
//
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var g *errgroup.Group

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	g, ctx = errgroup.WithContext(ctx)

	for _, collector := range []CustomCollector{
		NewLastBlockStatsCollector(c.client, ch),
		NewTransactionPoolCollector(c.client, ch),
		NewRPCCollector(c.client, ch),
		NewConnectionsCollector(c.client, ch),
		NewPeersCollector(c.client, ch),
		NewNetStatsCollector(c.client, ch),
		NewOverallCollector(c.client, ch),
	} {
		collector := collector

		g.Go(func() error {
			if err := collector.Collect(ctx); err != nil {
				return fmt.Errorf("%s collect: %w",
					collector.Name(), err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.log.Error(err, "wait")
	}
}
