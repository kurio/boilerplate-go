SOURCES := $(shell find . -name '*.go' -type f -not -path './vendor/*'  -not -path '*/mocks/*')
TEST_OPTS := -covermode=atomic $(TEST_OPTS)

IMAGE_NAME = goboilerplate
BINARY_NAME = goboilerplate

# Database
MYSQL_ADDRESS ?= localhost:3306
MYSQL_USER ?= kurio
MYSQL_PASSWORD ?= supersecret
MYSQL_DATABASE ?= myDB
MONGO_URI ?= mongodb://localhost:27017

.PHONY: goboilerplate
goboilerplate: vendor $(SOURCES)
	@echo "Building binary"
	@GO111MODULE=on go build -o $(BINARY_NAME) github.com/kurio/boilerplate-go/cmd/goboilerplate

# Dependencies Management

.PHONY: vendor
vendor: go.mod go.sum
	@GO111MODULE=on go get ./...

# Linter
.PHONY: lint-prepare
lint-prepare:
	@echo "Installing golangci-lint"
	@wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.31.0

.PHONY: lint
lint: vendor
	golangci-lint run ./...

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
