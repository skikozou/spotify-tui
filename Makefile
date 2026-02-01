.PHONY: build run debug clean install test

LOG_DIR := $(CURDIR)/log

# Build the application
build:
	go build -o spotify-tui ./cmd/spotify-tui

# Run the application
run: build
	./spotify-tui

# Run with debug logging
debug: build
	@mkdir -p $(LOG_DIR)
	./spotify-tui --debug --log-file=$(LOG_DIR)/spotify-tui.log

# Install dependencies
install:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -f spotify-tui
	go clean

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build   - Build the application"
	@echo "  run     - Build and run the application"
	@echo "  debug   - Build and run with debug logging (logs to ./log/)"
	@echo "  install - Install dependencies"
	@echo "  clean   - Clean build artifacts"
	@echo "  test    - Run tests"
	@echo "  fmt     - Format code"
	@echo "  lint    - Lint code"
