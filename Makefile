install:
	go install -v ./cmd/monero-exporter

run:
	monero-exporter \
		--monero-addr=http://localhost:18081 \
		--bind-addr=:9000 \
		--geoip-filepath=./hack/geoip.mmdb

test:
	go test ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run \
		--config=.golangci.yaml


table-of-contents:
	doctoc --notitle ./README.md
