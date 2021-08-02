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
	telemetryPath string
	bindAddr      string
	geoIPFilepath string
	moneroAddr    string
}

func (c *command) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monero-exporter",
		Short: "Prometheus exporter for monero metrics",
		RunE:  c.RunE,
	}

	cmd.Flags().StringVar(&c.bindAddr, "bind-addr",
		":9000", "address to bind the prometheus server to")

	cmd.Flags().StringVar(&c.telemetryPath, "telemetry-path",
		"/metrics", "endpoint at which prometheus metrics are served")

	cmd.Flags().StringVar(&c.moneroAddr, "monero-addr",
		"http://localhost:18081", "address of the monero instance to "+
			"collect info from")

	cmd.Flags().StringVar(&c.geoIPFilepath, "geoip-filepath",
		"", "filepath of a geoip database file for ip to country "+
			"resolution")
	_ = cmd.MarkFlagFilename("geoip-filepath")

	return cmd
}

func (c *command) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpcClient, err := rpc.NewClient(c.moneroAddr)
	if err != nil {
		return fmt.Errorf("new client '%s': %w", c.moneroAddr, err)
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
				return "", fmt.Errorf(
					"country '%s': %w", ip, err,
				)
			}

			return res.RegisteredCountry.IsoCode, nil
		}

		collectorOpts = append(collectorOpts,
			collector.WithCountryMapper(countryMapper),
		)
	}

	err = collector.Register(daemonClient, collectorOpts...)
	if err != nil {
		return fmt.Errorf("collector register: %w", err)
	}

	prometheusExporter, err := exporter.New(
		exporter.WithBindAddress(c.bindAddr),
		exporter.WithTelemetryPath(c.telemetryPath),
	)
	if err != nil {
		return fmt.Errorf("new exporter: %w", err)
	}
	defer prometheusExporter.Close()

	err = prometheusExporter.Run(ctx)
	if err != nil {
		return fmt.Errorf("prometheus exporter run: %w", err)
	}

	return nil
}
