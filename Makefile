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
	go test -cover ./...

example: dist/server_demo

dist/server_demo:
	go build -i -v -o $@ ./example/server_demo/...

.PHONY: all pre fmt test vet lint example dist/server_demo
