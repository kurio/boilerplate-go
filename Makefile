# Build
BINARY=boilerplate

build: test
	go build -o ${BINARY} github.com/kurio/boilerplate-go/cmd/app

# Test
.PHONY: test
test: lint
	@echo "Run go test"
	go test ./... -v -race -cover

.PHONY: unittest
unittest:
	@echo "Run go test"
	go test -short -v ./...

# Linter
.PHONY: lint-prepare
lint-prepare:
	@echo "Installing golangci-lint"
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: vendor
	@echo "Run lint"
	@golangci-lint run \
		--exclude-use-default=false \
		--enable=golint \
		--enable=gocyclo \
		--enable=goconst \
		--enable=unconvert \
		./...
