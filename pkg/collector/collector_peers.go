package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type PeersCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	graylist  []daemon.Peer
	whitelist []daemon.Peer
}

var _ CustomCollector = (*PeersCollector)(nil)

func NewPeersCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *PeersCollector {
	return &PeersCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *PeersCollector) Name() string {
	return "peerlist"
}

func (c *PeersCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	c.collectPeersCount()
	c.collectPeersLastSeen()

	return nil
}

func (c *PeersCollector) fetchData(ctx context.Context) error {
	resp, err := c.client.GetPeerList(ctx)
	if err != nil {
		return fmt.Errorf("get peerlist: %w", err)
	}

	c.graylist = resp.GrayList
	c.whitelist = resp.WhiteList

	return nil
}

func (c *PeersCollector) collectPeersCount() {
	desc := prometheus.NewDesc(
		"monero_peerlist",
		"number of node entries in the peerlist",
		[]string{"type"}, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(len(c.whitelist)),
		"white",
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(len(c.graylist)),
		"gray",
	)
}

func (c *PeersCollector) collectPeersLastSeen() {
	now := time.Now()
	summary := NewSummary()

	for _, peer := range c.whitelist {
		summary.Insert(now.
			Sub(time.Unix(peer.LastSeen, 0)).
			Seconds())
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_peerlist_lastseen",
			"distribution of when our peers have been seen",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}
