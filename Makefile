REVISION = $(shell git rev-parse --short HEAD)

all: pre test

pre: fmt lint vet

fmt:
	go fmt ./...

lint:
	golint $$(go list ./... | grep -v /vendor/)

vet:
	go vet $$(go list ./... | grep -v /vendor/)

test:
	go test -cover -v -race ./...

bench:
	go test -bench . -benchmem -gcflags="-m -l" ./...

dep-init:
	dep ensure

dep-update:
	dep ensure -update

example: dist/server_demo

dist/server_demo:
	go build -i -v -o $@ ./example/server_demo/...

.PHONY: all pre fmt lint vet test bench dep-init dep-update example dist/server_demo
