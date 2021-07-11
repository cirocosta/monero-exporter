package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cirocosta/go-monero/pkg/constant"
	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
)

type TransactionPoolCollector struct {
	client   *daemon.Client
	metricsC chan<- prometheus.Metric

	txns  []*daemon.TransactionJSON
	pool  *daemon.GetTransactionPoolResult
	stats *daemon.GetTransactionPoolStatsResult
}

var _ CustomCollector = (*TransactionPoolCollector)(nil)

func NewTransactionPoolCollector(
	client *daemon.Client, metricsC chan<- prometheus.Metric,
) *TransactionPoolCollector {
	return &TransactionPoolCollector{
		client:   client,
		metricsC: metricsC,
	}
}

func (c *TransactionPoolCollector) Name() string {
	return "transaction_pool"
}

func (c *TransactionPoolCollector) Collect(ctx context.Context) error {
	err := c.fetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	c.collectSpentKeyImages()
	c.collectSize()
	c.collectTransactionsSize()
	c.collectTransactionsCount()
	c.collectTransactionsFee()
	c.collectTransactionsFeePerKb()
	c.collectTransactionsInputs()
	c.collectTransactionsOutputs()
	c.collectTransactionsAgeDistribution()
	c.collectWeirdCases()

	return nil
}

func (c *TransactionPoolCollector) fetchData(ctx context.Context) error {
	stats, err := c.client.GetTransactionPoolStats(ctx)
	if err != nil {
		return fmt.Errorf("get transactionpool stats: %w", err)
	}

	pool, err := c.client.GetTransactionPool(ctx)
	if err != nil {
		return fmt.Errorf("get transaction pool: %w", err)
	}

	c.stats = stats
	c.pool = pool

	c.txns = make([]*daemon.TransactionJSON, len(pool.Transactions))

	for idx, txn := range c.pool.Transactions {
		c.txns[idx] = new(daemon.TransactionJSON)

		err := json.Unmarshal([]byte(txn.TxJSON), c.txns[idx])
		if err != nil {
			return fmt.Errorf("unmarhsal tx json: %w", err)
		}
	}

	return nil
}

func (c *TransactionPoolCollector) collectSpentKeyImages() {
	desc := prometheus.NewDesc(
		"monero_transaction_pool_spent_key_images",
		"total number of key images spent across all transactions"+
			" in the pool",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(len(c.pool.SpentKeyImages)),
	)

}

func (c *TransactionPoolCollector) collectTransactionsCount() {
	desc := prometheus.NewDesc(
		"monero_transaction_pool_transactions",
		"number of transactions in the pool at the moment of "+
			"the scrape",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(len(c.pool.Transactions)),
	)

}

func (c *TransactionPoolCollector) collectSize() {
	desc := prometheus.NewDesc(
		"monero_transaction_pool_size_bytes",
		"total size of the transaction pool",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.stats.PoolStats.BytesTotal),
	)
}

func (c *TransactionPoolCollector) collectTransactionsSize() {
	summary := NewSummary()
	for _, txn := range c.pool.Transactions {
		summary.Insert(float64(txn.BlobSize))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_transaction_pool_transactions_size_bytes",
			"distribution of the size of the transactions "+
				"in the transaction pool",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *TransactionPoolCollector) collectTransactionsFeePerKb() {
	summary := NewSummary()
	for _, txn := range c.pool.Transactions {
		fee := float64(txn.Fee) / constant.MicroXMR
		size := float64(txn.BlobSize) / 1024

		summary.Insert(fee / size)
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_transaction_pool_fees_micronero_per_kb",
			"distribution of the feeperkb utilized for txns"+
				" in the pool",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *TransactionPoolCollector) collectTransactionsInputs() {
	summary := NewSummary()
	for _, txn := range c.txns {
		summary.Insert(float64(len(txn.Vin)))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_transaction_pool_transactions_inputs",
			"distribution of inputs in the pool",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *TransactionPoolCollector) collectTransactionsOutputs() {
	summary := NewSummary()
	for _, txn := range c.txns {
		summary.Insert(float64(len(txn.Vout)))
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_transaction_pool_transactions_outputs",
			"distribution of outputs in the pool",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *TransactionPoolCollector) collectTransactionsAgeDistribution() {
	now := time.Now()

	summary := NewSummary()
	for _, txn := range c.pool.Transactions {
		summary.Insert(
			now.Sub(time.Unix(txn.ReceiveTime, 0)).Seconds(),
		)
	}

	c.metricsC <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_transaction_pool_transactions_age",
			"distribution of for how long transactions have "+
				"been in the pool",
			nil, nil,
		),
		summary.Count(), summary.Sum(), summary.Quantiles(),
	)
}

func (c *TransactionPoolCollector) collectTransactionsFee() {
	desc := prometheus.NewDesc(
		"monero_transaction_pool_fees_monero",
		"total amount of fee being spent in the transaction pool",
		nil, nil,
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(c.stats.PoolStats.FeeTotal)/constant.XMR,
	)
}

func (c *TransactionPoolCollector) collectWeirdCases() {
	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_transaction_pool_failing_transactions",
			"number of transactions that are marked as failing",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.stats.PoolStats.NumFailing),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_transaction_pool_double_spends",
			"transactions doubly spending outputs",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.stats.PoolStats.NumDoubleSpends),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_transaction_pool_not_relayed",
			"number of transactions that have not been relayed",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.stats.PoolStats.NumNotRelayed),
	)

	c.metricsC <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"monero_transaction_pool_older_than_10m",
			"number of transactions that are older than 10m",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(c.stats.PoolStats.Num10M),
	)
}
