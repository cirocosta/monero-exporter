install:
	go install -v ./cmd/monero-exporter

test:
	go test ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run --config=.golangci.yaml
