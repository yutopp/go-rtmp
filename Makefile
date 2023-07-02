FMT_DIRS:=$(shell go list -f {{.Dir}} ./...)

.PHONY: all
all: check test

.PHONY: check
check: fmt lint vet

.PHONY: download-ci-tools
download-ci-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.53.1
	curl -sSfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s v0.14.2

.PHONY: fmt
fmt:
	@gofmt -l -w -s $(FMT_DIRS)
	@goimports -w $(FMT_DIRS)

.PHONY: lint
lint:
	./bin/golangci-lint run

.PHONY: lint-ci
lint-ci:
	./bin/golangci-lint run | \
	./bin/reviewdog -f=golangci-lint -reporter=github-pr-review -filter-mode=nofilter

.PHONY: vet
vet:
	go vet $$(go list ./... | grep -v /vendor/)

.PHONY: test
test:
	go test -cover -coverprofile=coverage.txt -covermode=atomic -v -race -timeout 10s ./...

.PHONY: bench
bench:
	go test -bench . -benchmem -gcflags="-m -m -l" ./...

.PHONY: example
example:
	make -C example/server_demo
	make -C example/server_relay_demo
	make -C example/client_demo
