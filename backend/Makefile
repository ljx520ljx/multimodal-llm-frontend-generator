.PHONY: run build test clean tidy

# Default target
run:
	go run cmd/server/main.go

# Build binary
build:
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Tidy dependencies
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Development with hot reload (requires air)
dev:
	air
