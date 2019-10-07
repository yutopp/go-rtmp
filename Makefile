REVISION = $(shell git rev-parse --short HEAD)
GOLANGCI_LINT=hack/bin/golangci-lint
GO111MODULE=on

all: pre test

pre: fmt lint vet

fmt:
	go fmt ./...

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

vet:
	go vet $$(go list ./... | grep -v /vendor/)

test:
	go test -cover -v -race -timeout 10s ./...

bench:
	go test -bench . -benchmem -gcflags="-m -m -l" ./...

example:
	go build -i -v -o dist/server_demo ./example/server_demo/...
	go build -i -v -o dist/client_demo ./example/client_demo/...

$(GOLANGCI_LINT):
	cd ./hack; \
	go build -v \
		-o ./bin/golangci-lint \
		github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: all pre fmt lint vet test bench example
