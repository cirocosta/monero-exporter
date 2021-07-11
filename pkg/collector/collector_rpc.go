package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type RPCCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	accessTracking *daemon.RPCAccessTrackingResult
}

var _ CustomCollector = (*RPCCollector)(nil)

func NewRPCCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *RPCCollector {
	return &RPCCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *RPCCollector) Name() string {
	return "rpc"
}

func (c *RPCCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	c.collectRPC()

	return nil
}

func (c *RPCCollector) collectRPC() {
	countDesc := prometheus.NewDesc(
		"monero_rpc_hits_total",
		"number of hits that a particular rpc "+
			"method had since startup",
		[]string{"method"}, nil,
	)

	timeDesc := prometheus.NewDesc(
		"monero_rpc_seconds_total",
		"amount of time spent service the method "+
			"since startup",
		[]string{"method"}, nil,
	)

	for _, d := range c.accessTracking.Data {
		c.metricsC <- prometheus.MustNewConstMetric(
			countDesc,
			prometheus.GaugeValue,
			float64(d.Count),
			d.RPC,
		)

		c.metricsC <- prometheus.MustNewConstMetric(
			timeDesc,
			prometheus.GaugeValue,
			time.Duration(int64(d.Time)).Seconds(),
			d.RPC,
		)
	}
}

func (c *RPCCollector) fetchData(ctx context.Context) error {
	res, err := c.client.RPCAccessTracking(ctx)
	if err != nil {
		return fmt.Errorf("rpc access tracking: %w", err)
	}

	c.accessTracking = res

	return nil
}
