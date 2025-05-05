MODULE_PKG := github.com/jeffrydegrande/solidair

# If the environment has GITHUB_SHA we're running in a Github action, otherwise we try to
# get the commit hash from git. This avoids having to copy .git into docker build context.
GIT_HASH := $(or $(shell echo $(GITHUB_SHA)),$(shell git rev-parse --short HEAD))
BUILD_DATE := $(shell date -u '+%Y%m%dT%H%M%S')
LDFLAGS := -ldflags="-s -w -X $(MODULE_PKG)/build.BuildDate=${BUILD_DATE} -X ${MODULE_PKG}/build.Commit=${GIT_HASH}"

BIN_DIR ?= $(shell pwd)/bin
BIN_NAME ?= solidair

.PHONY: default
default: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: docker
docker: ## Build docker image
	@ echo "▶️ Building solidair docker image for development."
	@ docker buildx build \
		--build-arg GITHUB_SHA=${GITHUB_SHA} \
		--target solidair-runtime \
		-t jeffrydegrande/solidair .
	@ echo ✅ success!

ci: clean build lint vet staticcheck test
.PHONY: ci

ci-full: ci security
.PHONY: ci-full

clean:
	@rm -fr coverage.html coverage.txt
	@rm -fr $(BIN_NAME)
.PHONY: clean

.PHONY: security
security: ## Run security checks
	@ echo "▶️ gosec ./..."
	@ gosec  ./...
	@ echo "✅ gosec golint ./..."

.PHONY: lint
LINT_ARGS ?= -v
LINT_TARGETS ?= ./...
lint: ## Lint Go code with the installed golangci-lint
	@ echo "▶️ golangci-lint run $(LINT_ARGS) $(LINT_TARGETS)"
	golangci-lint run $(LINT_ARGS) $(LINT_TARGETS)
	@ echo "✅ golangci-lint run"

vet:
	go vet
.PHONY: vet

.PHONY: staticcheck
STATICCHECK_TARGETS ?= ./...
staticcheck: ## Run staticcheck linter
	@ echo "▶️ gstaticcheck $(STATICCHECK_TARGETS)"
	CGO_ENABLED=0 staticcheck $(STATICCHECK_TARGETS)
	@ echo "✅ staticcheck $(STATICCHECK_TARGETS)"

.PHONY: test
TEST_TARGETS ?= ./...
TEST_ARGS ?= -v -coverprofile=coverage.txt
test: ## Test the Go modules within this package.
	@ echo ▶️ go test $(TEST_ARGS) $(TEST_TARGETS)
	BUNDEBUG=1 go test $(TEST_ARGS) $(TEST_TARGETS)
	@ echo ✅ success!

	@ echo ▶️ go tool cover -func=coverage.txt
	go tool cover -func=coverage.txt
	@ echo ✅ success!

	@ echo ▶️ go tool cover -html=coverage.txt -o coverage.html
	go tool cover -html=coverage.txt -o coverage.html
	@ echo ✅ success!

.PHONY: bench
bench: ## Run benchmarks
	@ echo ▶️ go test -bench=. -benchmem
	go test -bench=. -benchmem ./...
	@ echo ✅ success!

generated:
	go generate -v ./...

build: docs generated
	go build $(LDFLAGS) -o $(BIN_NAME)
.PHONY: build

production:
	go build $(LDFLAGS) -o $(BIN_NAME)

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build

deps: ## Install dependencies
	@ echo ▶️ install swaggo/swag
	@ go install github.com/swaggo/swag/cmd/swag@latest
	@ echo ✅ success!
	@ echo ▶️ install staticcheck
	@ go install honnef.co/go/tools/cmd/staticcheck@latest
	@ echo ✅ success!
	@ echo ▶️ install gosec
	@ go install github.com/securego/gosec/cmd/gosec@latest
	@ echo ✅ success!
	@ echo ▶️ install go-swagger
	@ go install github.com/go-swagger/go-swagger/cmd/swagger@latest
	@ echo ✅ success!
	@ echo ▶️ install go-templ
	@ go install github.com/a-h/templ/cmd/templ@latest
	@ echo ✅ success!
	@ echo ▶️ install golangci-lint
	@ go install github.com/golangci/golangci-lint/cmd/golangci-lint@HEAD
	@ echo ✅ success!


.PHONY: deps
