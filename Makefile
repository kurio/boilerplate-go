SOURCES := $(shell find . -name '*.go' -type f -not -path './vendor/*' -not -path '*/mocks/*')
TEST_OPTS := -covermode=atomic $(TEST_OPTS)

# Database
MYSQL_ADDRESS ?= localhost:3306
MYSQL_USER ?= kurio
MYSQL_PASSWORD ?= supersecret
MYSQL_DATABASE ?= myDB
MIGRATION_STEP ?=
MONGO_URI ?= mongodb://localhost:27017

.PHONY: build
build:
	@echo "Building binary"
	@go build -o goboilerplate github.com/kurio/boilerplate-go/cmd/goboilerplate

# Prepare for development environment
.PHONY: prepare-dev
prepare-dev: vendor lint-prepare mockery-prepare;

.PHONY: vendor
vendor: go.mod go.sum
	go get ./...

# Linter
.PHONY: lint-prepare
lint-prepare:
	@echo "Installing golangci-lint"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: mockery-prepare
mockery-prepare:
	@echo "Installing mockery"
	@go install github.com/vektra/mockery/v2@latest

# Linter & Tests
.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test $(TEST_OPTS) ./...

.PHONY: unittest
unittest:
	go test -short $(TEST_OPTS) ./...

# Database Migration
.PHONY: migrate-prepare
migrate-prepare:
	@go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: migrate-up
migrate-up:
	@migrate -database "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_ADDRESS))/$(MYSQL_DATABASE)" \
	-path=internal/mysql/migrations up $(MIGRATION_STEP)

.PHONY: migrate-down
migrate-down:
	@migrate -database "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_ADDRESS))/$(MYSQL_DATABASE)" \
	-path=internal/mysql/migrations down $(MIGRATION_STEP)

.PHONY: migrate-drop
migrate-drop:
	@migrate -database "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_ADDRESS))/$(MYSQL_DATABASE)" \
	-path=internal/mysql/migrations drop

# Docker
.PHONY: mysql-up
mysql-up:
	@docker-compose up -d mysql

.PHONY: mysql-down
mysql-down:
	@docker-compose stop mysql && docker-compose rm -f

.PHONY: mongo-up
mongo-up:
	@docker-compose up -d mongo

.PHONY: mongo-down
mongo-down:
	@docker-compose stop mongo && docker-compose rm -f

.PHONY: redis-up
redis-up:
	@docker-compose up -d redis

.PHONY: redis-down
redis-down:
	@docker-compose stop redis && docker-compose rm -f

.PHONY: otel-up
otel-up:
	@docker-compose up -d jaeger-all-in-one zipkin-all-in-one otel-collector prometheus

.PHONY: docker
docker:
	@docker-compose build

.PHONY: run
run:
	@docker-compose up -d

.PHONY: stop
stop:
	@docker-compose down -v

# Mock
MyRepository: filename.go
	@mockery -name=MyRepository

MyService: filename.go
	@mockery -name=MyService

# Scan for vulnerabilities
.PHONY: scan-fs
scan-fs:
	trivy fs --vuln-type 'library' --severity 'MEDIUM,HIGH,CRITICAL' --ignore-unfixed --security-checks vuln .

.PHONY: scan-image
scan-image:
	trivy image --vuln-type 'library' --severity 'MEDIUM,HIGH,CRITICAL' --ignore-unfixed --security-checks vuln goboilerplate

.PHONY: scan
	scan: scan-fs scan-image;
