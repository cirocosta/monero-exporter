package collector

import (
	"context"
	"fmt"
	"math"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/bmizerany/perks/quantile"
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
		log:           zapr.NewLogger(defaultLogger.Named("collector")),
		countryMapper: func(_ net.IP) (string, error) { return "unknown", nil },
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
	// Because we can present the description of the metrics at collection time, we
	// don't need to write anything to the channel.
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

	for _, collector := range []struct {
		name string
		fn   CollectFunc
	}{
		{"info_stats", c.CollectInfoStats},
		{"mempool_stats", c.CollectMempoolStats},
		{"last_block_header", c.CollectLastBlockHeader},
		{"bans", c.CollectBans},
		{"peer_height_divergence", c.CollectPeerHeightDivergence},
		{"fee_estimate", c.CollectFeeEstimate},
		{"peers", c.CollectPeers},
		{"connections", c.CollectConnections},
		{"last_block_stats", c.CollectLastBlockStats},
		{"peers_live_time", c.CollectPeersLiveTime},
		{"net_stats", c.CollectNetStats},
		{"collect_rpc", c.CollectRPC},
	} {
		collector := collector

		g.Go(func() error {
			if err := collector.fn(ctx, ch); err != nil {
				return fmt.Errorf("collector fn '%s': %w", collector.name, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.log.Error(err, "wait")
	}
}

// CollectConnections.
//
func (c *Collector) CollectConnections(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetConnections(ctx)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	perCountryCounter := map[string]uint64{}
	for _, conn := range res.Connections {
		country, err := c.countryMapper(net.ParseIP(conn.Host))
		if err != nil {
			return fmt.Errorf("to country '%s': %w", conn.Host, err)
		}

		perCountryCounter[country]++
	}

	desc := prometheus.NewDesc(
		"monero_connections",
		"connections info",
		[]string{"country"}, nil,
	)

	for country, count := range perCountryCounter {
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(count),
			country,
		)
	}

	return nil
}

// CollectPeers.
//
func (c *Collector) CollectPeers(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetPeerList(ctx)
	if err != nil {
		return fmt.Errorf("get peer list: %w", err)
	}

	perCountryCounter := map[string]uint64{}
	for _, peer := range res.WhiteList {
		country, err := c.countryMapper(net.ParseIP(peer.Host))
		if err != nil {
			return fmt.Errorf("to country '%s': %w", peer.Host, err)
		}

		perCountryCounter[country]++
	}

	desc := prometheus.NewDesc(
		"monero_peers_new",
		"peers info",
		[]string{"country"}, nil,
	)

	for country, count := range perCountryCounter {
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(count),
			country,
		)
	}

	return nil
}

// CollectLastBlockHeader.
//
func (c *Collector) CollectLastBlockHeader(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetLastBlockHeader(ctx)
	if err != nil {
		return fmt.Errorf("get last block header: %w", err)
	}

	metrics, err := c.toMetrics("last_block_header", &res.BlockHeader)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

// CollectInfoStats.
//
func (c *Collector) CollectInfoStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("get transaction pool: %w", err)
	}

	metrics, err := c.toMetrics("info", res)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

func (c *Collector) CollectLastBlockStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	lastBlockHeaderResp, err := c.client.GetLastBlockHeader(ctx)
	if err != nil {
		return fmt.Errorf("get last block header: %w", err)
	}

	currentHeight := lastBlockHeaderResp.BlockHeader.Height
	block, err := c.client.GetBlock(ctx, daemon.GetBlockRequestParameters{
		Height: currentHeight,
	})
	if err != nil {
		return fmt.Errorf("get block '%d': %w", currentHeight, err)
	}

	blockJSON, err := block.InnerJSON()
	if err != nil {
		return fmt.Errorf("block inner json: %w", err)
	}

	txnsResp, err := c.client.GetTransactions(ctx, blockJSON.TxHashes)
	if err != nil {
		return fmt.Errorf("get txns: %w", err)
	}

	txns, err := txnsResp.GetTransactions()
	if err != nil {
		return fmt.Errorf("get transactions: %w", err)
	}

	phis := []float64{0.25, 0.50, 0.75, 0.90, 0.95, 0.99, 1}

	var (
		streamTxnSize    = quantile.NewTargeted(phis...)
		sumTxnSize       = float64(0)
		quantilesTxnSize = make(map[float64]float64, len(phis))

		streamTxnFee    = quantile.NewTargeted(phis...)
		sumTxnFee       = float64(0)
		quantilesTxnFee = make(map[float64]float64, len(phis))

		streamVin    = quantile.NewTargeted(phis...)
		sumVin       = float64(0)
		quantilesVin = make(map[float64]float64, len(phis))

		streamVout    = quantile.NewTargeted(phis...)
		sumVout       = float64(0)
		quantilesVout = make(map[float64]float64, len(phis))
	)

	for _, txn := range txnsResp.TxsAsHex {
		streamTxnSize.Insert(float64(len(txn)))
		sumTxnSize += float64(len(txn))
	}

	for _, txn := range txns {
		streamTxnFee.Insert(float64(txn.RctSignatures.Txnfee))
		sumTxnFee += float64(txn.RctSignatures.Txnfee)

		streamVin.Insert(float64(len(txn.Vin)))
		sumVin += float64(len(txn.Vin))

		streamVout.Insert(float64(len(txn.Vout)))
		sumVout += float64(len(txn.Vout))
	}

	for _, phi := range phis {
		quantilesTxnSize[phi] = streamTxnSize.Query(phi)
		quantilesTxnFee[phi] = streamTxnFee.Query(phi)
		quantilesVin[phi] = streamVin.Query(phi)
		quantilesVout[phi] = streamVout.Query(phi)
	}

	ch <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_last_block_txn_size",
			"distribution of tx sizes",
			nil, nil,
		),
		uint64(streamTxnSize.Count()),
		sumTxnSize,
		quantilesTxnSize,
	)

	ch <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_last_block_txn_fee",
			"distribution of outputs in last block",
			nil, nil,
		),
		uint64(streamTxnFee.Count()),
		sumTxnFee,
		quantilesTxnFee,
	)

	ch <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_last_block_vin",
			"distribution of inputs in last block",
			nil, nil,
		),
		uint64(streamVin.Count()),
		sumVin,
		quantilesVin,
	)

	ch <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_last_block_vout",
			"distribution of outputs in last block",
			nil, nil,
		),
		uint64(streamVout.Count()),
		sumVout,
		quantilesVout,
	)

	return nil
}

func (c *Collector) CollectPeerHeightDivergence(ctx context.Context, ch chan<- prometheus.Metric) error {
	blockCountRes, err := c.client.GetBlockCount(ctx)
	if err != nil {
		return fmt.Errorf("get block count: %w", err)
	}

	res, err := c.client.GetConnections(ctx)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	phis := []float64{0.25, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85, 0.90, 0.95, 0.99}
	stream := quantile.NewTargeted(phis...)

	sum := float64(0)
	ourHeight := blockCountRes.Count
	for _, conn := range res.Connections {
		diff := math.Abs(float64(ourHeight - uint64(conn.Height)))

		stream.Insert(diff)
		sum += diff
	}

	quantiles := make(map[float64]float64, len(phis))
	for _, phi := range phis {
		quantiles[phi] = stream.Query(phi)
	}

	desc := prometheus.NewDesc(
		"monero_height_divergence",
		"how much our peers diverge from us in block height",
		nil, nil,
	)

	ch <- prometheus.MustNewConstSummary(
		desc,
		uint64(stream.Count()),
		sum,
		quantiles,
	)

	return nil
}

func (c *Collector) CollectPeersLiveTime(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetConnections(ctx)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	var (
		phis      = []float64{0.25, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85, 0.90, 0.95, 0.99}
		sum       = float64(0)
		stream    = quantile.NewTargeted(phis...)
		quantiles = make(map[float64]float64, len(phis))
	)

	for _, conn := range res.Connections {
		stream.Insert(float64(conn.LiveTime))
		sum += float64(conn.LiveTime)
	}

	for _, phi := range phis {
		quantiles[phi] = stream.Query(phi)
	}

	desc := prometheus.NewDesc(
		"monero_connections_livetime",
		"peers livetime distribution",
		nil, nil,
	)

	ch <- prometheus.MustNewConstSummary(
		desc,
		uint64(stream.Count()),
		sum,
		quantiles,
	)

	return nil
}

func (c *Collector) CollectNetStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetNetStats(ctx)
	if err != nil {
		return fmt.Errorf("get fee estimate: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_net_total_in_bytes",
			"network statistics",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(res.TotalBytesIn),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_net_total_out_bytes",
			"network statistics",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(res.TotalBytesOut),
	)

	return nil
}

func (c *Collector) CollectFeeEstimate(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetFeeEstimate(ctx, 1)
	if err != nil {
		return fmt.Errorf("get fee estimate: %w", err)
	}

	desc := prometheus.NewDesc(
		"monero_fee_estimate",
		"fee estimate for 1 grace block",
		nil, nil,
	)

	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(res.Fee),
	)

	return nil
}

func (c *Collector) CollectRPC(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.RPCAccessTracking(ctx)
	if err != nil {
		return fmt.Errorf("rpc access tracking: %w", err)
	}

	descCount := prometheus.NewDesc(
		"monero_rpc_count",
		"todo",
		[]string{"method"}, nil,
	)

	descTime := prometheus.NewDesc(
		"monero_rpc_time",
		"todo",
		[]string{"method"}, nil,
	)

	for _, d := range res.Data {
		ch <- prometheus.MustNewConstMetric(
			descCount,
			prometheus.CounterValue,
			float64(d.Count),
			d.RPC,
		)

		ch <- prometheus.MustNewConstMetric(
			descTime,
			prometheus.CounterValue,
			float64(d.Time),
			d.RPC,
		)
	}

	return nil
}

func (c *Collector) CollectBans(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetBans(ctx)
	if err != nil {
		return fmt.Errorf("get bans: %w", err)
	}

	desc := prometheus.NewDesc(
		"monero_bans",
		"number of nodes banned",
		nil, nil,
	)

	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(len(res.Bans)),
	)

	return nil
}

func (c *Collector) CollectMempoolStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetTransactionPoolStats(ctx)
	if err != nil {
		return fmt.Errorf("get transaction pool: %w", err)
	}

	metrics, err := c.toMetrics("mempool", &res.PoolStats)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

func (c *Collector) toMetrics(ns string, res interface{}) ([]prometheus.Metric, error) {
	var (
		metrics = []prometheus.Metric{}
		v       = reflect.ValueOf(res).Elem()
		err     error
	)

	for i := 0; i < v.NumField(); i++ {
		observation := float64(0)
		field := v.Field(i)

		switch field.Type().Kind() {
		case reflect.Bool:
			if field.Bool() {
				observation = float64(1)
			}

		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Uintptr:

			observation, err = strconv.ParseFloat(fmt.Sprintf("%v", field.Interface()), 64)
			if err != nil {
				return nil, fmt.Errorf("parse float: %w", err)
			}

		default:
			c.log.Info("ignoring",
				"field", v.Type().Field(i).Name,
				"type", field.Type().Kind().String(),
			)

			continue
		}

		tag := v.Type().Field(i).Tag.Get("json")

		metrics = append(metrics, prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"monero_"+ns+"_"+tag,
				"info for "+tag,
				nil, nil,
			),
			prometheus.GaugeValue,
			observation,
		))
	}

	return metrics, nil
}
