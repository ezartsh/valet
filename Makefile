.PHONY: all test test-race test-cover bench lint vet fmt clean help

# Default target
all: fmt vet test

# Run tests
test:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race -v ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run specific benchmark
bench-%:
	go test -bench=$* -benchmem ./...

# Run linter (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...
	@echo "Code formatted"

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	rm -f *.test
	go clean -testcache

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Check for vulnerabilities (requires govulncheck)
vuln:
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Generate documentation
doc:
	@echo "Opening documentation in browser..."
	@which godoc > /dev/null || (echo "Installing godoc..." && go install golang.org/x/tools/cmd/godoc@latest)
	@echo "Visit http://localhost:6060/pkg/github.com/ezartsh/validet-zod/"
	godoc -http=:6060

# Run all checks before commit
pre-commit: fmt vet test-race lint
	@echo "All checks passed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  all         - Format, vet, and test (default)"
	@echo "  test        - Run tests"
	@echo "  test-race   - Run tests with race detection"
	@echo "  test-cover  - Run tests with coverage report"
	@echo "  bench       - Run all benchmarks"
	@echo "  bench-NAME  - Run specific benchmark (e.g., bench-String)"
	@echo "  lint        - Run golangci-lint"
	@echo "  vet         - Run go vet"
	@echo "  fmt         - Format code"
	@echo "  clean       - Clean build artifacts"
	@echo "  tidy        - Tidy go.mod"
	@echo "  deps        - Download dependencies"
	@echo "  vuln        - Check for vulnerabilities"
	@echo "  doc         - Open documentation in browser"
	@echo "  pre-commit  - Run all checks before commit"
	@echo "  help        - Show this help"
