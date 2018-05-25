REVISION = $(shell git rev-parse --short HEAD)

.PHONY: all pre fmt test vet lint example dist/server_demo

all: pre test

pre: fmt vet lint

fmt:
	go fmt ./...

vet:
	go vet $$(go list ./... | grep -v /vendor/)

lint:
	golint $$(go list ./... | grep -v /vendor/)

test:
	go test -cover ./...

example: dist/server_demo

dist/server_demo:
	go build -i -v -o $@ ./example/server_demo/...
