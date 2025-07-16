# Copyright The Linux Foundation and each contributor to LFX.
# SPDX-License-Identifier: MIT

APP_NAME := lfx-query-svc
VERSION := $(shell git describe --tags --always)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse HEAD)

# Docker
DOCKER_REGISTRY := linuxfoundation## container registry ghcr.io/ ???
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG := $(VERSION)

# Go
GO_VERSION := 1.24.2
GOOS := linux
GOARCH := amd64

# Linting
GOLANGCI_LINT_VERSION := v1.64.6
LINT_TIMEOUT := 10m
LINT_TOOL=$(shell go env GOPATH)/bin/golangci-lint

##@ Development

.PHONY: setup-dev
setup-dev: ## Setup development tools
	@echo "Installing development tools..."
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go mod tidy

.PHONY: deps
deps: ## Install dependencies
	@echo "Installing dependencies..."
	go install goa.design/goa/v3/cmd/goa@latest

.PHONY: apigen
apigen: deps #@ Generate API code using Goa
	goa gen github.com/linuxfoundation/lfx-v2-query-service/design

.PHONY: lint
lint: ## Run golangci-lint with default settings
	@echo "Running golangci-lint..."
	@$(LINT_TOOL) run --config=.golangci.yml ./... && echo "==> Lint OK"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: build-local
build-local: ## Build the application for local OS
	@echo "Building application for local development..."
	go build \
		-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" \
		-o bin/$(APP_NAME) .

##@ Docker

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest


.PHONY: docker-run
docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	docker run \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-e OPENSEARCH_URL=http://opensearch-cluster-master.lfx.svc.cluster.local:9200 \
		-e NATS_URL=nats://nats.lfx.svc.cluster.local:4222 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

##@ Helm/Kubernetes

.PHONY: helm-render
helm-render: ## Render Helm templates
	@echo "Rendering Helm templates..."
	helm template lfx-query-svc deploy/charts --set image.tag=$(DOCKER_TAG)

.PHONY: helm-deploy
helm-deploy: ## Deploy application using Helm
	helm upgrade --install lfx-query-svc deploy/charts --set image.tag=$(DOCKER_TAG)