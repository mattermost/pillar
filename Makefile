################################################################################
##                             VERSION PARAMS                                 ##
################################################################################

## Docker Build Versions
DOCKER_BUILDER_SERVER_IMAGE = golang:1.14.4
DOCKER_BASE_IMAGE = alpine:3.12

################################################################################

export GOBIN ?= $(PWD)/bin
GO ?= $(shell command -v go 2> /dev/null)
GOFLAGS ?= $(GOFLAGS:)
PILLAR_IMAGE ?= mattermost/pillar:test

# Build
MACHINE = $(shell uname -m)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH = $(shell git rev-parse HEAD)

# Tests
TESTS=.

export GO111MODULE=on

## Checks the code style, tests, builds and bundles.
all: check-style

## Build the docker image for pillar
.PHONY: build-image
build-image:
	@echo Building Pillar Container Image
	docker build \
	--build-arg DOCKER_BUILDER_SERVER_IMAGE=$(DOCKER_BUILDER_SERVER_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	--build-arg ENV=cloud \
	--build-arg GITHUB_USERNAME=$(GITHUB_USERNAME) \
	--build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) \
	. -t $(PILLAR_IMAGE) \
	--no-cache

## Build the Container image for local development.
.PHONY: dev-start
dev-start:
	@echo Starting local development
	docker-compose up -d

## Shutdown the development environment.
.PHONY: dev-stop
dev-stop:
	@echo Shutting down the local environment
	docker-compose down

## Clean the development environment.
.PHONY: dev-clean
dev-clean:
	@echo Cleaning the local environment
	docker-compose kill

## Runs govet and gofmt against all packages.
.PHONY: check-style
check-style: govet lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
.PHONY: lint
lint:
	@echo Running lint
	env GO111MODULE=off $(GO) get -u golang.org/x/lint/golint
	$(GOBIN)/golint -set_exit_status $(./... | grep -v /blapi/)
	@echo lint success

## Runs govet against all packages.
.PHONY: vet
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

.PHONY: build
build: ## Build pillar.
	@echo Building pillar
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o build/_output/bin/cws  ./cmd/cws

### Generate mocks
.PHONY: mocks
mocks:
	$(GO) install github.com/golang/mock/mockgen
	$(GOBIN)/mockgen -package mock -destination=mock/cloud.go -source=api/context.go CloudClient
