# Contributing to Rate Limiter

Thank you for your interest in contributing to the Rate Limiter project! 🎉

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:
- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS

### Suggesting Features

We welcome feature suggestions! Please open an issue describing:
- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Write tests** for any new functionality
3. **Ensure tests pass**: `go test ./...`
4. **Run race detector**: `go test -race ./...`
5. **Format your code**: `go fmt ./...`
6. **Update documentation** if needed
7. **Submit your PR** with a clear description

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write clear, descriptive commit messages
- Add comments for exported functions and types
- Keep functions focused and testable

### Testing Requirements

- All new code must have tests
- Maintain or improve code coverage
- Include both unit and integration tests
- Test concurrent scenarios where applicable

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/rate-limiter.git
cd rate-limiter

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Questions?

Feel free to open an issue for any questions or clarifications!

Thank you for contributing! 🙏
