.PHONY: build run test clean docker lint fmt help

# Variables
BINARY_NAME=goguard
DOCKER_IMAGE=goguard
VERSION?=1.0.0

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) ./cmd/goguard

# Run the application
run: build
	./$(BINARY_NAME)

# Run with debug mode
run-debug: build
	GOGUARD_MODE=debug GOGUARD_LOG_LEVEL=debug ./$(BINARY_NAME)

# Run tests
test:
	go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Build Docker image
docker:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 \
		-e GOGUARD_LLM_API_KEY=$(GOGUARD_LLM_API_KEY) \
		$(DOCKER_IMAGE):latest

# Lint code
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Download dependencies
deps:
	go mod download
	go mod tidy

# Update dependencies
deps-update:
	go get -u ./...
	go mod tidy

# Generate mocks (if needed)
generate:
	go generate ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run           - Build and run the application"
	@echo "  run-debug     - Run in debug mode"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker        - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  help          - Show this help"
