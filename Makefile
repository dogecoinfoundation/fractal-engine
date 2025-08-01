all: binaries

BIN_DIR ?= ./bin/
binaries:
	go build -o ${BIN_DIR} ./cmd/...

GO_TEST_EXTRA_FLAGS ?=
test:
	env ENV=test TZ=UTC go test ${GO_TEST_EXTRA_FLAGS} -p 1 -covermode=count -coverprofile=coverage.txt -timeout=30m  ./...

coverage:
	go tool cover -func=coverage.txt

coverage-html:
	go tool cover -html=coverage.txt

lint:
	golangci-lint run --fix
	buf lint

vet:
	go vet ./...

format:
	golangci-lint fmt
	buf format -w

tidy:
	go mod tidy

generate:
	go generate -v ./...
	buf build
	buf generate

checks: tidy format generate lint vet

ci-checks: checks
	git diff --exit-code

.PHONY: all binaries test lint vet format tidy generate checks ci-checks
