SOURCES := $(shell find . -name '*.go' -type f -not -path './vendor/*'  -not -path '*/mocks/*')
TEST_OPTS := -covermode=atomic $(TEST_OPTS)

IMAGE_NAME = boilerplate

# Database
MYSQL_ADDRESS ?= localhost:3306
MYSQL_USER ?= kurio
MYSQL_PASSWORD ?= supersecret
MYSQL_DATABASE ?= myDB
MONGO_URI ?= mongodb://localhost:27017

# Dependencies Management

.PHONY: vendor
vendor: go.mod go.sum
	@GO111MODULE=on go get ./...

# Linter
.PHONY: lint-prepare
lint-prepare:
	@echo "Installing golangci-lint"
	@GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: vendor
	GO111MODULE=on golangci-lint run ./...

.PHONY: mockery-prepare
mockery-prepare:
	@echo "Installing mockery"
	@GO111MODULE=off go get -u github.com/vektra/mockery/.../

# Testing
.PHONY: unittest
unittest: vendor
	GO111MODULE=on go test -short $(TEST_OPTS) ./...

.PHONY: test
test: vendor
	GO111MODULE=on go test $(TEST_OPTS) ./...

# Build and Installation
.PHONY: install
install: vendor
	@GO111MODULE=on go install ./...

.PHONY: uninstall
uninstall:
	@echo "Removing binaries and libraries"
	@GO111MODULE=on go clean -i ./...

boilerplate: vendor $(SOURCES)
	GO111MODULE=on go build -o boilerplate github.com/kurio/boilerplate-go/app


# Database Migration
.PHONY: migrate-prepare
migrate-prepare:
	@GO111MODULE=off go get -tags 'mysql' -u github.com/golang-migrate/migrate/cmd/migrate

.PHONY: migrate-up
migrate-up:
	@migrate -database "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_ADDRESS))/$(MYSQL_DATABASE)" \
	-path=internal/mysql/migrations up

.PHONY: migrate-down
migrate-down:
	@migrate -database "mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_ADDRESS))/$(MYSQL_DATABASE)" \
	-path=internal/mysql/migrations down

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
	@docker stop boilerplate_mysql

.PHONY: mongo-up
mongo-up:
	@docker-compose up -d mongo

.PHONY: mongo-down
mongo-down:
	@docker stop boilerplate_mongo

.PHONY: docker
docker: vendor $(SOURCES)
	@cp ~/.ssh/id_rsa .keys/
	@docker build -t $(IMAGE_NAME) .

.PHONY: run
run:
	@docker-compose up -d

.PHONY: stop
stop:
	@docker-compose down

# Mock
MyRepository: filename.go
	@mockery -name=MyRepository

MyService: filename.go
	@mockery -name=MyService
