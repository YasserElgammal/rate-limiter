.PHONY: test test-verbose test-race test-cover bench run-basic run-http fmt vet check clean deps build-examples help

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Run tests with coverage
test-cover:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run basic example
run-basic:
	go run examples/basic/main.go

# Run HTTP server example
run-http:
	go run examples/http_server/main.go

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet test-race

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build examples
build-examples:
	go build -o bin/basic examples/basic/main.go
	go build -o bin/http_server examples/http_server/main.go

# Help
help:
	@echo "Available targets:"
	@echo "  test           - Run all tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-cover     - Run tests with coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  run-basic      - Run basic example"
	@echo "  run-http       - Run HTTP server example"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  check          - Run fmt, vet, and test-race"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  build-examples - Build example binaries"