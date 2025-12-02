# Contributing to Valet

First off, thank you for considering contributing to Valet! It's people like you that make this project better.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct: be respectful, inclusive, and constructive.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (code snippets, JSON data, etc.)
- **Describe the behavior you observed and what you expected**
- **Include your Go version** (`go version`)
- **Include your OS and version**

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Use a clear and descriptive title**
- **Provide a detailed description of the suggested enhancement**
- **Explain why this enhancement would be useful**
- **List any alternative solutions you've considered**

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Write tests** for any new functionality
3. **Ensure all tests pass** (`make test-race`)
4. **Run the linter** (`make lint`)
5. **Update documentation** if needed
6. **Write a clear commit message**

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Make (optional, but recommended)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/valet.git
cd valet

# Add upstream remote
git remote add upstream https://github.com/ezartsh/valet.git

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
make test

# Run tests with race detection
make test-race

# Run benchmarks
make bench

# Run linter
make lint
```

### Project Structure

```
valet/
├── adapters.go          # Database adapters (SQL, GORM, SQLX, Bun)
├── array.go             # Array validator
├── boolean.go           # Boolean validator
├── cache.go             # Regex cache
├── db.go                # Database types and interfaces
├── errors.go            # Error types
├── file.go              # File validator
├── number.go            # Number validators (Float, Int, Int64)
├── object.go            # Object validator
├── string.go            # String validator
├── types.go             # Common types and interfaces
├── utils.go             # Utility functions
├── validate.go          # Main validation logic
├── *_test.go            # Test files
├── benchmark_test.go    # Benchmark tests
└── database_integration_test.go  # DB integration tests
```

## Coding Guidelines

### Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Write descriptive variable and function names
- Add comments for exported functions and types

### Testing

- Write tests for all new functionality
- Aim for high test coverage
- Include both positive and negative test cases
- Use table-driven tests where appropriate
- Run race detection tests

### Performance

- Be mindful of allocations
- Use `sync.Pool` for frequently allocated objects
- Pre-allocate slices and maps when size is known
- Run benchmarks before and after changes
- Document any performance trade-offs

### Documentation

- Update README.md for new features
- Add GoDoc comments for exported items
- Include examples in documentation
- Keep CHANGELOG.md updated

## Commit Messages

Use clear and meaningful commit messages:

```
feat: add UUID validation to string validator
fix: handle nil values in array validation
perf: reduce allocations in batch key generation
docs: update README with new examples
test: add tests for edge cases in number validation
refactor: simplify error handling in validate.go
```

Prefixes:
- `feat:` - New feature
- `fix:` - Bug fix
- `perf:` - Performance improvement
- `docs:` - Documentation only
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Questions?

Feel free to open an issue for any questions or discussions.

Thank you for contributing! 🎉
