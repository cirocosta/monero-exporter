package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type OverallCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	info *daemon.GetInfoResult
}

var _ CustomCollector = (*OverallCollector)(nil)

func NewOverallCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *OverallCollector {
	return &OverallCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *OverallCollector) Name() string {
	return "overall"
}

func (c *OverallCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	c.collect()

	return nil
}

func (c *OverallCollector) fetchData(ctx context.Context) error {
	res, err := c.client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("get netstats: %w", err)
	}

	c.info = res

	return nil
}

func (c *OverallCollector) collect() {
	now := time.Now()

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_uptime_seconds_total",
			"for how long this node has been up",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(now.
			Sub(time.Unix(int64(c.info.StartTime), 0)).
			Seconds()),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_alternative_blocks",
			"number of blocks alternative to the longest",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.AltBlocksCount),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_offline",
			"whether the node is offline",
			nil, nil,
		),
		prometheus.GaugeValue,
		boolToFloat64(c.info.Offline),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_mainnet",
			"whether the node is connected to mainnet",
			nil, nil,
		),
		prometheus.GaugeValue,
		boolToFloat64(c.info.Mainnet),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_block_size_limit_bytes",
			"maximum hard limit of a block",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.BlockSizeLimit),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_block_size_median_bytes",
			"current median size for computing dynamic fees",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.BlockSizeMedian),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_synchronized",
			"whether the node's chain is in sync with the network",
			nil, nil,
		),
		prometheus.GaugeValue,
		boolToFloat64(c.info.Synchronized),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_height",
			"current height of the chain",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.Height),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_target_height",
			"target height to achieve to be considered in sync",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.TargetHeight),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_rpc_connections",
			"number of rpc connections being served by the node",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.RPCConnectionsCount),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_database_size_bytes",
			"size of the monero database",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.DatabaseSize),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_free_space_bytes",
			"amount of free space in the partition where "+
				"monero's database is in",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.FreeSpace),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_info_transactions_total",
			"total number of transactions seen so far",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.info.TxCount),
	)

}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
