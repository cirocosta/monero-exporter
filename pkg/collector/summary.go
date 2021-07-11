package collector

import "github.com/beorn7/perks/quantile"

// defaultQuantiles is the default quantiles to compute for a given data stream
// that we want to summarize.
//
// these (quantile -> epsilon) will be used by default by any Summary unless
// initialized with the `WithQuantiles` option to override it.
//
var defaultQuantiles = map[float64]float64{
	0.05: 0.01,
	0.10: 0.01,
	0.25: 0.01,
	0.50: 0.01,
	0.75: 0.01,
	0.90: 0.01,
	0.95: 0.01,
	0.99: 0.01,
	1.00: 0.01,
}

type Summary struct {
	count     uint64
	sum       float64
	quantiles map[float64]float64

	stream   *quantile.Stream
	computed bool
}

type SummaryOption func(s *Summary)

func WithQuantiles(v map[float64]float64) SummaryOption {
	return func(s *Summary) {
		s.quantiles = v
	}
}

func NewSummary(opts ...SummaryOption) *Summary {
	summary := &Summary{
		count:     uint64(0),
		sum:       float64(0),
		quantiles: cloneMap(defaultQuantiles),
	}

	for _, opt := range opts {
		opt(summary)
	}

	summary.stream = quantile.NewTargeted(summary.quantiles)

	return summary
}

func (s *Summary) Insert(v float64) {
	s.sum += v
	s.stream.Insert(v)
	s.count++
}

func (s *Summary) Count() uint64 {
	s.compute()
	return s.count
}

func (s *Summary) Quantiles() map[float64]float64 {
	s.compute()
	return s.quantiles
}

func (s *Summary) Sum() float64 {
	s.compute()
	return s.sum
}

func (s *Summary) compute() {
	if s.computed {
		return
	}

	for phi := range s.quantiles {
		s.quantiles[phi] = s.stream.Query(phi)
	}
}

func cloneMap(o map[float64]float64) map[float64]float64 {
	m := make(map[float64]float64, len(o))
	for k, v := range o {
		m[k] = v
	}

	return m
}
