# Contributing to 4ebur-net

Thank you for your interest in contributing to 4ebur-net! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Performance Guidelines](#performance-guidelines)

## Code of Conduct

Be respectful, professional, and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Docker (optional, for container testing)
- golangci-lint (for code quality checks)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/onixus/4ebur-net.git
cd 4ebur-net

# Install dependencies
make deps

# Install development tools
make install-tools

# Run tests to verify setup
make test
```

## Development Workflow

1. **Fork the repository**

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Write clean, readable code
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   make test
   make lint
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `test:` - Adding or updating tests
   - `perf:` - Performance improvements
   - `refactor:` - Code refactoring
   - `chore:` - Maintenance tasks

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**

## Coding Standards

### Go Code Style

- Follow official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Use `goimports` for import organization
- Pass all `golangci-lint` checks

### Code Organization

```
4ebur-net/
├── cmd/           # Application entry points
├── internal/      # Private application code
├── pkg/           # Public libraries
└── ...            # Other directories
```

### Naming Conventions

- Use descriptive names
- Follow Go naming conventions (camelCase for variables, PascalCase for exported items)
- Avoid abbreviations unless widely understood

### Comments

- Add comments for exported functions, types, and constants
- Explain "why", not "what" in comments
- Keep comments up-to-date with code changes

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Bad
if err != nil {
    panic(err)
}
```

## Testing

### Writing Tests

- Write table-driven tests when possible
- Test both success and failure cases
- Use meaningful test names
- Aim for >80% code coverage

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench

# Run specific test
go test -v -run TestFeatureName ./...
```

### Benchmarks

- Add benchmarks for performance-critical code
- Compare before/after performance

```go
func BenchmarkFeature(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Code to benchmark
    }
}
```

## Pull Request Process

1. **Ensure CI passes**
   - All tests pass
   - Linters pass
   - Coverage doesn't decrease

2. **Update documentation**
   - Update README if needed
   - Add/update comments
   - Update CHANGELOG (if exists)

3. **Fill out PR template**
   - Describe changes
   - Reference related issues
   - Add screenshots if UI changes

4. **Wait for review**
   - Address reviewer comments
   - Make requested changes
   - Re-request review when ready

5. **Squash commits** (if requested)
   ```bash
   git rebase -i HEAD~n
   ```

## Performance Guidelines

### This is a High-Load Proxy

- **Think about performance first**
- Avoid allocations in hot paths
- Use object pooling (`sync.Pool`) for frequently allocated objects
- Profile before optimizing
- Benchmark performance-critical changes

### Performance Testing

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Common Performance Patterns

```go
// Use sync.Pool for buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 32*1024))
    },
}

// Reuse HTTP clients
var client = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        1000,
        MaxIdleConnsPerHost: 100,
    },
}
```

## Security

- **Never commit secrets** (API keys, passwords, certificates)
- Use environment variables for configuration
- Follow least privilege principle
- Validate all inputs
- Handle errors securely

## Documentation

- Keep README up-to-date
- Document all exported functions
- Add examples where helpful
- Update architecture diagrams if structure changes

## Questions?

If you have questions:

- Check existing issues
- Ask in discussions
- Create a new issue with `question` label

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
