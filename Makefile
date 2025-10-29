# Default target: build, test, lint
all: build test lint

# Build the application
build:
	@echo "Building..."
	@go build -ldflags="-s -w" -o bin/bkp .

# Run the application
run:
	@echo "Running..."
	@go run .

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Lint the code
lint:
	@echo "Linting..."
	@golangci-lint run

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f bin/bkp

.PHONY: all build run test lint clean