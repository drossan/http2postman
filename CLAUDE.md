# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`http2postman` is a Go CLI tool that bidirectionally converts between IntelliJ HTTP Client `.http` files and Postman Collection v2.1.0 JSON format. Built with Cobra CLI framework.

## Commands

```sh
# Build
go build -mod=vendor -o bin/http2postman .

# Run tests
go test ./... -v

# Run a single test
go test ./internal/parser/ -run TestParseHTTPFile_SingleRequest

# Test coverage
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# Vet
go vet ./...

# Format
gofmt -w . && goimports -w .

# Lint (requires golangci-lint)
golangci-lint run ./...
```

Dependencies are vendored (`-mod=vendor` flag required for build).

## Architecture

**Current state:** All logic lives in `cmd/` with `map[string]interface{}` for data. Being migrated to Clean Architecture.

**Target architecture** (see `.claude/specs/01-architecture.md`):

- **`cmd/`** — Thin CLI layer. Commands only parse args, call services, handle errors. Use `RunE` (not `Run`).
- **`internal/model/`** — Typed domain structs for Postman collections, HTTP files, environments. No `map[string]interface{}`.
- **`internal/parser/`** — Parses `.http` files and Postman JSON into domain models.
- **`internal/converter/`** — Transforms between HTTP and Postman models.
- **`internal/writer/`** — Writes Postman JSON and `.http` files to disk.
- **`internal/fs/`** — `FileSystem` interface with OS and in-memory implementations. All file I/O in business logic goes through this interface.

## Specs

Detailed specs live in `.claude/specs/`. **Always read the relevant spec before implementing.** Index at `.claude/specs/00-index.md`.

## Key Rules

- **TDD mandatory**: Write failing test first, then implement, then refactor.
- **No `map[string]interface{}`** for domain data — use typed structs from `internal/model/`.
- **Type assertions require comma-ok** pattern — bare `.(type)` assertions are forbidden.
- **Wrap errors with context**: `fmt.Errorf("parsing %s: %w", path, err)` — never bare `return err`.
- **Never `fmt.Println` for errors** in business logic — return errors, only `cmd/` presents them.
- **Filesystem access via interface** — `internal/` packages never import `os` for file operations directly.
- **Functions max ~30 lines**, max 3 parameters. Use option structs if more needed.
- **Guard clauses** — early returns over nested if/else.
- **No `strings.Title`** (deprecated) — use `cases.Title(language.Und).String()` from `golang.org/x/text`.
- **Sentinel errors** in `internal/model/errors.go` for known business cases.
- **Table-driven tests** with `t.Run` subtests. Test fixtures in `testdata/` directories.

## Build & Release

GoReleaser handles cross-compilation (linux/darwin/windows, multiple archs) and Homebrew tap publishing to `drossan/homebrew-tools`. Version info injected via ldflags into `main.version` and `main.commit`.
