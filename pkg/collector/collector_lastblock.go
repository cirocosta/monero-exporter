package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/constant"
	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type LastBlockStatsCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	txns     []*daemon.TransactionJSON
	txnSizes []int
	header   daemon.BlockHeader
}

var _ CustomCollector = (*LastBlockStatsCollector)(nil)

func NewLastBlockStatsCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *LastBlockStatsCollector {
	return &LastBlockStatsCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *LastBlockStatsCollector) Name() string {
	return "lastblock"
}

func (c *LastBlockStatsCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch last block data: %w", err)
	}

	c.collectBlockSize()
	c.collectDifficulty()
	c.collectFees()
	c.collectHeight()
	c.collectReward()
	c.collectSubsidy()
	c.collectTransactionsCount()
	c.collectTransactionsFeePerKb()
	c.collectTransactionsInputs()
	c.collectTransactionsOutputs()
	c.collectTransactionsSize()

	return nil
}

func (c *LastBlockStatsCollector) fetchData(ctx context.Context) error {
	lastBlockHeaderResp, err := c.client.GetLastBlockHeader(ctx)
	if err != nil {
		return fmt.Errorf("get last block header: %w", err)
	}

	lastBlockHash := lastBlockHeaderResp.BlockHeader.Hash

	params := daemon.GetBlockRequestParameters{
		Hash: lastBlockHash,
	}
	blockResp, err := c.client.GetBlock(ctx, params)
	if err != nil {
		return fmt.Errorf("get block '%s': %w", lastBlockHash, err)
	}

	blockJSON, err := blockResp.InnerJSON()
	if err != nil {
		return fmt.Errorf("block inner json: %w", err)
	}

	txnsResp, err := c.client.GetTransactions(ctx, blockJSON.TxHashes)
	if err != nil {
		return fmt.Errorf("get txns: %w", err)
	}

	txnSizes := make([]int, len(txnsResp.Txs))
	for idx, t := range txnsResp.Txs {
		txnSizes[idx] = len(t.AsHex) / 2
	}

	txns, err := txnsResp.GetTransactions()
	if err != nil {
		return fmt.Errorf("get transactions: %w", err)
	}

	c.txns = txns
	c.txnSizes = txnSizes
	c.header = blockResp.BlockHeader

	return nil
}

func (c *LastBlockStatsCollector) collectBlockSize() {
	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_lastblock_size_bytes",
			"total size of the last block",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.header.BlockSize),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_lastblock_weight_bytes",
			"total weight of the last block",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.header.BlockWeight),
	)
}

func (c *LastBlockStatsCollector) collectDifficulty() {
	desc := prometheus.NewDesc(
		"monero_lastblock_difficulty",
		"difficulty used for the last block",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.header.Difficulty),
	)
}

func (c *LastBlockStatsCollector) collectFees() {
	desc := prometheus.NewDesc(
		"monero_lastblock_fees_monero",
		"total amount of fees included in this block",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.gatherFees(c.txns))/constant.XMR,
	)
}

func (c *LastBlockStatsCollector) collectHeight() {
	desc := prometheus.NewDesc(
		"monero_lastblock_height",
		"height of the last block",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.header.Height),
	)
}

func (c *LastBlockStatsCollector) collectReward() {
	desc := prometheus.NewDesc(
		"monero_lastblock_reward_monero",
		"total amount of rewards granted in the last block "+
			"(subsidy + fees)",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.header.Reward)/constant.XMR,
	)
}

func (c *LastBlockStatsCollector) collectSubsidy() {
	totalReward := float64(c.header.Reward)
	fees := float64(c.gatherFees(c.txns))
	subsidy := (totalReward - fees) / constant.XMR

	desc := prometheus.NewDesc(
		"monero_lastblock_subsidy_monero",
		"newly minted monero for this block",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		subsidy,
	)
}

func (c *LastBlockStatsCollector) collectTransactionsCount() {
	desc := prometheus.NewDesc(
		"monero_lastblock_transactions",
		"number of transactions seen in the last block",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.header.NumTxes),
	)
}

func (c *LastBlockStatsCollector) collectTransactionsFeePerKb() {
	summary := NewSummary()
	for idx, txn := range c.txns {
		fee := float64(txn.RctSignatures.Txnfee) / constant.MicroXMR
		size := float64(c.txnSizes[idx]) / 1024

		summary.Insert(fee / size)
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_lastblock_fees_micronero_per_kb",
			"distribution of the feeperkb utilized for txns",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *LastBlockStatsCollector) collectTransactionsSize() {
	summary := NewSummary()
	for _, size := range c.txnSizes {
		summary.Insert(float64(size))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_lastblock_transactions_size_bytes",
			"distribution of the size of the transactions included",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

// - transaction weight for a miner transaciton or a normal transaction with
//   two outputs is equal to the size in bytes.
// - a bulletproof occupies ((2*[log_2(64p)]+9)*32) bytes
//
//
// ref: https://github.com/monero-project/monero/blob/a9cb4c082f1a815076674b543d93159a2137540e/src/cryptonote_basic/cryptonote_format_utils.cpp#L106
//
func (c *LastBlockStatsCollector) computeTransactionWeight(idx int) float64 {
	const bpBase uint = 368

	txn := c.txns[idx]
	txnSize := uint(c.txnSizes[idx])

	numberOfOutputs := uint(len(txn.Vout))
	if numberOfOutputs <= 2 {
		return float64(txnSize)
	}

	nlr := uint(0)
	for {
		if (1 << nlr) < numberOfOutputs {
			break
		}
		nlr++
	}
	nlr += 6

	bpSize := 32 * (9 + 2*nlr)
	bpClawback := (bpBase*numberOfOutputs - bpSize) * 4 / 5

	return float64(txnSize + bpClawback)
}

func (c *LastBlockStatsCollector) collectTransactionsWeight() {
	summary := NewSummary()
	for idx := range c.txnSizes {
		summary.Insert(c.computeTransactionWeight(idx))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_lastblock_transactions_weight_bytes",
			"distribution of the weight of the transactions included",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *LastBlockStatsCollector) collectTransactionsInputs() {
	summary := NewSummary()
	for _, txn := range c.txns {
		summary.Insert(float64(len(txn.Vin)))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_lastblock_transactions_inputs",
			"distribution of inputs in the last block",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *LastBlockStatsCollector) collectTransactionsOutputs() {
	summary := NewSummary()
	for _, txn := range c.txns {
		summary.Insert(float64(len(txn.Vout)))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_lastblock_transactions_outputs",
			"distribution of outputs in the last block",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *LastBlockStatsCollector) collectVersions() {
	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_lastblock_version_major",
			"major version of the block format",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.header.MajorVersion),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_lastblock_version_minor",
			"minor version of the block format",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.header.MinorVersion),
	)
}

func (c *LastBlockStatsCollector) gatherFees(txns []*daemon.TransactionJSON) uint64 {
	fees := uint64(0)
	for _, txn := range txns {
		fees += txn.RctSignatures.Txnfee
	}

	return fees
}
