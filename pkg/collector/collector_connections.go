package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type ConnectionsCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	connections *daemon.GetConnectionsResult
}

var _ CustomCollector = (*ConnectionsCollector)(nil)

func NewConnectionsCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *ConnectionsCollector {
	return &ConnectionsCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *ConnectionsCollector) Name() string {
	return "connections"
}

func (c *ConnectionsCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	c.collectConnectionsCount()
	c.collectHeightDistribution()
	c.collectDataRates()
	c.collectConnectionAges()

	return nil
}

func (c *ConnectionsCollector) collectConnectionAges() {
	summary := NewSummary()

	for _, conn := range c.connections.Connections {
		summary.Insert(float64(conn.LiveTime))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_p2p_connections_age",
			"distribution of age of the connections we have",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *ConnectionsCollector) collectDataRates() {
	summaryRx := NewSummary()
	summaryTx := NewSummary()

	for _, conn := range c.connections.Connections {
		summaryRx.Insert(float64(conn.RecvCount) / float64(conn.LiveTime))
		summaryTx.Insert(float64(conn.SendCount) / float64(conn.LiveTime))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_p2p_connections_rx_rate_bps",
			"distribution of data receive rate in bytes/s",
			nil, nil,
		),
		summaryRx.Count(), summaryRx.Sum(), summaryRx.Quantiles(),
	)

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_p2p_connections_tx_rate_bps",
			"distribution of data transmit rate in bytes/s",
			nil, nil,
		),
		summaryTx.Count(), summaryTx.Sum(), summaryTx.Quantiles(),
	)
}

func (c *ConnectionsCollector) collectHeightDistribution() {
	summary := NewSummary()
	for _, conn := range c.connections.Connections {
		summary.Insert(float64(conn.Height))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_p2p_connections_height",
			"distribution the height of the peers "+
				"connected to/from us",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *ConnectionsCollector) collectConnectionsCount() {
	desc := prometheus.NewDesc(
		"monero_p2p_connections",
		"number of connections to/from this node",
		[]string{"type", "state"}, nil,
	)

	type key struct {
		ttype string
		state string
	}

	counters := map[key]float64{}

	for _, conn := range c.connections.Connections {
		ttype := "in"
		if !conn.Incoming {
			ttype = "out"
		}

		counters[key{ttype, conn.State}]++
	}

	for k, v := range counters {
		c.metricsC <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			v,
			k.ttype, k.state,
		)
	}
}

func (c *ConnectionsCollector) fetchData(ctx context.Context) error {
	res, err := c.client.GetConnections(ctx)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	c.connections = res
	return nil
}
