default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.dev.tfrc:
	@GOBIN=$$(go env GOBIN); \
	printf 'provider_installation {\n  dev_overrides {\n    "registry.terraform.io/536tech/e2b" = "%s"\n  }\n  direct {}\n}\n' \
		"$${GOBIN:-$$(go env GOPATH)/bin}" > $@

terraform-test: install .dev.tfrc
	E2B_API_KEY=terraform-test TF_CLI_CONFIG_FILE=$(CURDIR)/.dev.tfrc terraform init -backend=false
	E2B_API_KEY=terraform-test TF_CLI_CONFIG_FILE=$(CURDIR)/.dev.tfrc terraform test

.PHONY: fmt lint test testacc terraform-test build install generate
