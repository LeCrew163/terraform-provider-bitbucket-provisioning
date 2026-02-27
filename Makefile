default: build

# ── Build info ───────────────────────────────────────────────────────────────
BINARY       := terraform-provider-bitbucket-provisioning
PROVIDER_NS  := bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning
VERSION      := 0.10.0

OS   := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
# Goreleaser capitalises the OS name in archive filenames (e.g. Darwin, Linux)
OS_CAP := $(shell go env GOOS | sed 's/^./\u&/')
PLUGIN_DIR := $(HOME)/.terraform.d/plugins/$(PROVIDER_NS)/$(VERSION)/$(OS)_$(ARCH)

ARTIFACTORY_BASE := http://art01.sldnet.de:8081/artifactory/terraform

# ── Core ─────────────────────────────────────────────────────────────────────
.PHONY: build
build:
	go build -o $(BINARY)

.PHONY: install
install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY) $(PLUGIN_DIR)/

# Download a released binary from Artifactory and install it locally.
# Usage: make install-remote VERSION=0.10.0
.PHONY: install-remote
install-remote:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make install-remote VERSION=x.y.z"; exit 1; fi
	$(eval REMOTE_ZIP := $(BINARY)_$(VERSION)_$(OS_CAP)_$(ARCH).zip)
	$(eval REMOTE_URL := $(ARTIFACTORY_BASE)/$(BINARY)/$(VERSION)/$(REMOTE_ZIP))
	$(eval REMOTE_PLUGIN_DIR := $(HOME)/.terraform.d/plugins/$(PROVIDER_NS)/$(VERSION)/$(OS)_$(ARCH))
	@echo "Downloading $(REMOTE_URL) ..."
	@curl -fSL "$(REMOTE_URL)" -o /tmp/$(REMOTE_ZIP)
	@mkdir -p $(REMOTE_PLUGIN_DIR)
	@unzip -o /tmp/$(REMOTE_ZIP) "$(BINARY)_v$(VERSION)" -d $(REMOTE_PLUGIN_DIR)/
	@mv $(REMOTE_PLUGIN_DIR)/$(BINARY)_v$(VERSION) $(REMOTE_PLUGIN_DIR)/$(BINARY)
	@rm /tmp/$(REMOTE_ZIP)
	@echo "Installed $(BINARY) v$(VERSION) → $(REMOTE_PLUGIN_DIR)"

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf dist/

# ── Tests ────────────────────────────────────────────────────────────────────
.PHONY: test
test:
	go test -v -cover ./...

# Run acceptance tests against a live Bitbucket instance.
# Required environment variables:
#   BITBUCKET_BASE_URL   e.g. http://localhost:7990
#   BITBUCKET_USERNAME + BITBUCKET_PASSWORD  (or BITBUCKET_TOKEN)
# Note: if both BITBUCKET_TOKEN and username/password env vars are set, unset
#       BITBUCKET_TOKEN first to avoid the "Multiple Authentication Methods" error.
.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -cover ./... -timeout 120m

# End-to-end test: start Docker Compose, build+install provider, run Terraform.
# Copy .env.local.example to .env.local and fill in BITBUCKET_LICENSE first.
.PHONY: test-local
test-local:
	bash scripts/test-local.sh

# ── Docker ───────────────────────────────────────────────────────────────────
.PHONY: docker-up
docker-up:
	docker compose --env-file .env.local up -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-clean
docker-clean:
	docker compose down -v

.PHONY: docker-logs
docker-logs:
	docker compose logs -f bitbucket

# ── Code quality ─────────────────────────────────────────────────────────────
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

# ── Code generation ──────────────────────────────────────────────────────────
.PHONY: generate
generate:
	go generate ./...

.PHONY: generate-client
generate-client:
	@echo "Generating API client from OpenAPI spec..."
	@mkdir -p internal/client/generated
	openapi-generator generate \
		-i specs/bitbucket-openapi.json \
		-g go \
		-o internal/client/generated \
		--skip-validate-spec \
		--additional-properties=packageName=bitbucket,isGoSubmodule=false
	@echo "Client generation complete"

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate --provider-name bitbucketdc

# ── Module management ────────────────────────────────────────────────────────
.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: release-test
release-test:
	goreleaser release --snapshot --clean

# ── Help ─────────────────────────────────────────────────────────────────────
.PHONY: help
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Core:"
	@echo "  build            Build the provider binary ($(BINARY))"
	@echo "  install          Build and install to $(PLUGIN_DIR)"
	@echo "  install-remote   Download released binary from Artifactory (VERSION=x.y.z)"
	@echo "  clean            Remove build artifacts"
	@echo ""
	@echo "Tests:"
	@echo "  test             Run unit tests"
	@echo "  testacc          Run acceptance tests (needs BITBUCKET_BASE_URL + credentials)"
	@echo "  test-local       Full end-to-end test via Docker Compose"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up        Start Bitbucket + PostgreSQL (requires .env.local)"
	@echo "  docker-down      Stop containers (keep volumes)"
	@echo "  docker-clean     Stop containers and delete volumes"
	@echo "  docker-logs      Tail Bitbucket container logs"
	@echo ""
	@echo "Code quality:"
	@echo "  fmt              Format Go code"
	@echo "  vet              Run go vet"
	@echo "  lint             Run golangci-lint"
	@echo ""
	@echo "Generation:"
	@echo "  generate         Run go generate"
	@echo "  generate-client  Regenerate API client from OpenAPI spec"
	@echo "  docs             Generate provider documentation"
	@echo ""
	@echo "Module:"
	@echo "  tidy             Tidy go.mod / go.sum"
	@echo "  vendor           Vendor dependencies"
	@echo "  release-test     Test goreleaser release build"
	@echo ""
