# Contributing to zaple-go

Thank you for your interest in contributing! This document explains how to get set up and what we expect from contributions.

## Getting started

1. **Fork** the repository and clone your fork.
2. **Install** Go 1.22 or later.
3. **Run tests** to confirm a clean baseline:

```bash
go test ./...
```

## Development workflow

### Running tests

```bash
# All tests
go test ./...

# With race detector (always run this before opening a PR)
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

We use [golangci-lint](https://golangci-lint.run/). Install it and run:

```bash
golangci-lint run
```

### Code style

- Follow standard Go formatting — run `gofmt -w .` before committing.
- Every exported type, function, and constant must have a godoc comment.
- Prefer table-driven tests and `httptest.Server` over live API calls.
- No external dependencies unless strictly necessary (stdlib is preferred).

## Pull request guidelines

- **One concern per PR** — bug fixes, new features, and refactors should be separate.
- **Reference an issue** in the PR description where applicable.
- **Include tests** for all new behaviour (aim for ≥80% coverage on new code).
- **Update the README** if you add a new method or option.
- Keep commits atomic and write meaningful commit messages.

## Reporting bugs

Open a GitHub issue and include:
- Go version (`go version`)
- OS and architecture
- Minimal reproduction case
- Expected vs actual behaviour

## Security

Do **not** open a public issue for security vulnerabilities. Email the maintainer directly instead.
