package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/cirocosta/go-monero/pkg/rpc"
	"github.com/jessevdk/go-flags"
	"github.com/oschwald/geoip2-golang"

	"github.com/cirocosta/monero-exporter/pkg/collector"
	"github.com/cirocosta/monero-exporter/pkg/exporter"
)

type MetricsCommand struct {
	// nolint:lll
	MonerodAddress string `long:"monerod-address" default:"http://localhost:18081" required:"true" description:"address of monerod rpc (restricted if possible)"`
	GeoIPFile      string `long:"geoip-file" description:"filepath of geoip database"`
}

func main() {
	cmd := &MetricsCommand{}

	if _, err := flags.Parse(cmd); err != nil {
		os.Exit(1)
	}

	if err := cmd.Execute(nil); err != nil {
		panic(err)
	}
}

func (c *MetricsCommand) Execute(_ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prometheusExporter, err := exporter.New()
	if err != nil {
		return fmt.Errorf("new exporter: %w", err)
	}
	defer prometheusExporter.Close()

	daemonClient, err := rpc.NewClient(c.MonerodAddress)
	if err != nil {
		return fmt.Errorf("new client '%s': %w", c.MonerodAddress, err)
	}

	collectorOpts := []collector.Option{}

	if c.GeoIPFile != "" {
		db, err := geoip2.Open(c.GeoIPFile)
		if err != nil {
			return fmt.Errorf("geoip open: %w", err)
		}
		defer db.Close()

		countryMapper := func(ip net.IP) (string, error) {
			res, err := db.Country(ip)
			if err != nil {
				return "", fmt.Errorf("country '%s': %w", ip, err)
			}

			return res.RegisteredCountry.IsoCode, nil
		}

		collectorOpts = append(collectorOpts, collector.WithCountryMapper(countryMapper))
	}

	if err := collector.Register(daemonClient, collectorOpts...); err != nil {
		return fmt.Errorf("collector register: %w", err)
	}

	if err := prometheusExporter.Run(ctx); err != nil {
		return fmt.Errorf("prometheus exporter run: %w", err)
	}

	return nil
}
