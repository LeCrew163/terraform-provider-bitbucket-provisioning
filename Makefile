default: build

.PHONY: build
build:
	go build -o terraform-provider-bitbucket-dc

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/colab.internal.sldo.cloud/alpina/bitbucket-dc/0.1.0/darwin_arm64
	mv terraform-provider-bitbucket-dc ~/.terraform.d/plugins/colab.internal.sldo.cloud/alpina/bitbucket-dc/0.1.0/darwin_arm64/

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -cover ./... -timeout 120m

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

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
	tfplugindocs generate --provider-name bitbucket-dc

.PHONY: clean
clean:
	rm -f terraform-provider-bitbucket-dc
	rm -rf dist/

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: release-test
release-test:
	goreleaser release --snapshot --clean

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the provider binary"
	@echo "  install      - Install provider locally for testing"
	@echo "  test         - Run unit tests"
	@echo "  testacc      - Run acceptance tests"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  generate     - Run go generate"
	@echo "  generate-client - Generate API client from OpenAPI spec"
	@echo "  docs         - Generate documentation"
	@echo "  clean        - Clean build artifacts"
	@echo "  tidy         - Tidy Go modules"
	@echo "  vendor       - Vendor dependencies"
	@echo "  release-test - Test release process locally"
