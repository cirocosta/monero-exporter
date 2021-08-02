package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type NetStatsCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	stats *daemon.GetNetStatsResult
}

var _ CustomCollector = (*NetStatsCollector)(nil)

func NewNetStatsCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *NetStatsCollector {
	return &NetStatsCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *NetStatsCollector) Name() string {
	return "net"
}

func (c *NetStatsCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	c.collectRxTx()

	return nil
}

func (c *NetStatsCollector) fetchData(ctx context.Context) error {
	res, err := c.client.GetNetStats(ctx)
	if err != nil {
		return fmt.Errorf("get netstats: %w", err)
	}

	c.stats = res

	return nil
}

func (c *NetStatsCollector) collectRxTx() {
	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_net_rx_bytes",
			"number of bytes received by this node",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.stats.TotalBytesIn),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_net_tx_bytes",
			"number of bytes received by this node",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.stats.TotalBytesOut),
	)
}
