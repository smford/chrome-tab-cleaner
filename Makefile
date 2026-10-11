BINARY_NAME := chrome-tab-cleaner
CMD_DIR     := ./cmd/chrome-tab-cleaner
BUILD_DIR   := bin
GO          := go

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE        ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS     := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: all build run test clean lint fmt tidy pre-commit help

all: build

## build: Compile the CLI application binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) ./$(BINARY_NAME)
	@echo "==> Built ./$(BINARY_NAME)"

## run: Run the application CLI
run: build
	@./$(BINARY_NAME) $(ARGS)

## test: Run all tests
test:
	@echo "==> Running tests..."
	$(GO) test -v -race ./...

## fmt: Format all Go source files
fmt:
	@echo "==> Formatting code..."
	$(GO) fmt ./...

## lint: Run code formatting checks and go vet
lint:
	@echo "==> Running code standards and lint..."
	@unformatted=$$($(GO) fmt ./...); if [ -n "$$unformatted" ]; then echo "Unformatted files:"; echo "$$unformatted"; exit 1; fi
	$(GO) vet ./...

## tidy: Tidy and verify go.mod
tidy:
	@echo "==> Tidying module dependencies..."
	$(GO) mod tidy

## pre-commit: Run pre-commit hooks on all files
pre-commit:
	@echo "==> Running pre-commit hooks..."
	@command -v pre-commit >/dev/null && pre-commit run --all-files || echo "pre-commit not installed locally (brew install pre-commit)"

## clean: Remove build artifacts and temporary files
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(BINARY_NAME) dist
	@echo "==> Cleaned."

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
