// Package collector provides the core functionality of this exporter.
//
// It implements the Prometheus collector interface, providing `monero` metrics
// whenever a request hits this exporter, allowing us to not have to rely on a
// particular interval defined in this exporter (instead, rely on prometheus'
// scrape interval).
//
package collector
