# Contributing to pdf-cli

Thank you for your interest in contributing to pdf-cli!

## Development Setup

### Prerequisites
- Go 1.21 or later
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- (Optional) Tesseract for native OCR testing

### Getting Started

```bash
# Clone the repository
git clone https://github.com/lgbarn/pdf-cli.git
cd pdf-cli

# Download dependencies
make deps

# Run tests
make test

# Build
make build
```

## Development Workflow

### Before Making Changes

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Ensure tests pass: `make test`
3. Ensure linting passes: `make lint`

### Making Changes

1. Write tests first (TDD encouraged)
2. Implement the minimal code to pass tests
3. Run `make check-all` to verify everything passes
4. Commit with descriptive messages

### Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Use `make lint-fix` to auto-fix linting issues
- Keep functions focused and under 50 lines when possible

### Commit Messages

Follow conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Test changes
- `chore:` Maintenance
- `refactor:` Code restructuring

Example: `feat: add --dry-run flag to merge command`

## Testing

### Running Tests

```bash
make test           # Run all tests
make test-coverage  # Run with coverage report
make test-race      # Run with race detection
make coverage       # Show coverage percentage
```

### Writing Tests

- Place tests in `*_test.go` files alongside source
- Use table-driven tests for multiple cases
- Use mocks from `internal/testing/` for isolation
- Test both success and error paths

Example:
```go
func TestParsePages(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    []int
        wantErr bool
    }{
        {"single page", "1", []int{1}, false},
        {"range", "1-3", []int{1, 2, 3}, false},
        {"invalid", "abc", nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParsePages(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Process

1. Ensure all checks pass: `make check-all`
2. Update documentation if needed
3. Add tests for new functionality
4. Keep PRs focused on a single change
5. Respond to review feedback promptly

## Project Structure

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## Questions?

Open an issue for questions or discussions.
