package main

import (
	"context"
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/spf13/cobra"

	"github.com/cirocosta/go-monero/pkg/rpc"
	"github.com/cirocosta/go-monero/pkg/rpc/daemon"
	"github.com/cirocosta/monero-exporter/pkg/collector"
	"github.com/cirocosta/monero-exporter/pkg/exporter"
)

type command struct {
	address       string
	geoIPFilepath string
}

func (c *command) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monero-exporter",
		Short: "Prometheus exporter for monero metrics",
		RunE:  c.RunE,
	}

	cmd.Flags().StringVar(&c.address, "address",
		"", "address of the monero node to collect metrics from")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringVar(&c.geoIPFilepath, "geoip-filepath",
		"", "filepath of a geoip database file for ip to country resolution")
	cmd.MarkFlagFilename("geoip-filepath")

	return cmd
}

func (c *command) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prometheusExporter, err := exporter.New()
	if err != nil {
		return fmt.Errorf("new exporter: %w", err)
	}
	defer prometheusExporter.Close()

	rpcClient, err := rpc.NewClient(c.address)
	if err != nil {
		return fmt.Errorf("new client '%s': %w", c.address, err)
	}

	daemonClient := daemon.NewClient(rpcClient)

	collectorOpts := []collector.Option{}

	if c.geoIPFilepath != "" {
		db, err := geoip2.Open(c.geoIPFilepath)
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
