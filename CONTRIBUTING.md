# Contributing to manifestor

Thank you for taking the time to contribute! This document covers everything you need to get started.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

---

## Getting Started

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/<your-username>/manifestor.git
   cd manifestor
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/alanzng/manifestor.git
   ```

---

## Development Setup

**Requirements:**
- Go 1.22 or later
- [`golangci-lint`](https://golangci-lint.run/usage/install/) for linting

```bash
# Verify the build
go build ./...

# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run linter
golangci-lint run

# Run benchmarks
go test -bench=. -benchmem ./...
```

---

## Project Structure

```
manifestor/
├── cmd/manifestor/      # CLI entry point
├── dash/                # DASH parser, writer, builder, filter, options
├── hls/                 # HLS parser, writer, builder, filter, options
├── manifest/            # Unified auto-detect API
├── server/              # HTTP proxy server
├── testdata/            # Real-world fixture manifests
│   ├── hls/
│   └── dash/
├── .github/             # CI workflows, issue/PR templates
├── CHANGELOG.md
├── CONTRIBUTING.md
├── go.mod
└── README.md
```

---

## How to Contribute

### Bug fixes

1. Check existing [issues](https://github.com/alanzng/manifestor/issues) to avoid duplicates.
2. Open a new issue describing the bug (or pick an existing one).
3. Create a branch: `git checkout -b fix/<short-description>`
4. Write a failing test that reproduces the bug.
5. Fix the bug so the test passes.
6. Submit a pull request.

### New features

1. Open a [feature request issue](https://github.com/alanzng/manifestor/issues/new?template=feature_request.md) first to discuss the design.
2. Wait for maintainer feedback before investing significant time.
3. Create a branch: `git checkout -b feat/<short-description>`
4. Implement the feature with tests.
5. Submit a pull request.

### Documentation

Documentation improvements are always welcome — just open a PR directly.

---

## Coding Standards

- **No external dependencies** in `hls/`, `dash/`, `manifest/`, or `server/` packages.
- All exported types, functions, and constants must have **Go doc comments**.
- Follow standard Go formatting — run `gofmt -w .` before committing.
- Keep packages focused: parsers parse, writers write, builders build.
- No global state — all configuration via functional options or builder methods.
- Error values (e.g. `ErrNoVariantsRemain`) must be defined as package-level `var`s, not inline `errors.New()` calls.

---

## Testing

- Test coverage must remain **≥ 80%** on all core packages.
- Each filter and transformer must have a unit test using a **real-world fixture manifest** from `testdata/`.
- Use table-driven tests (`t.Run(...)`) for multi-case scenarios.
- Benchmarks go in `_test.go` files alongside the code they benchmark.

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Submitting a Pull Request

1. Rebase on the latest `main`:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```
2. Ensure all tests pass and linting is clean:
   ```bash
   go test -race ./...
   golangci-lint run
   ```
3. Update `CHANGELOG.md` under the `[Unreleased]` section.
4. Open a PR against `main` and fill in the pull request template.
5. A maintainer will review within a few business days.

---

## Reporting Bugs

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include:
- Go version (`go version`)
- OS and architecture
- Minimal reproducing manifest (or link to a public one)
- Expected vs. actual behaviour

---

## Requesting Features

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md). Describe:
- The problem you are trying to solve
- Your proposed solution
- Alternatives you considered
