# Spec 08: Makefile y Tooling

## Objetivo

Proporcionar un Makefile con targets estándar para build, test, lint, y desarrollo local.

## Makefile

```makefile
.PHONY: build test test-coverage lint fmt vet clean run-export run-import

BINARY_NAME=http2postman
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS=-ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'"

## Build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

## Test

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo "---"
	@echo "To view HTML report: go tool cover -html=coverage.out"

test-short:
	go test ./... -short

## Code Quality

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

check: fmt vet lint test

## Clean

clean:
	rm -rf bin/ coverage.out

## Development helpers

run-export: build
	./bin/$(BINARY_NAME) export $(DIR)

run-import: build
	./bin/$(BINARY_NAME) import $(FILE)
```

## Herramientas requeridas

Documentar en README cómo instalar:

- **golangci-lint**: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **goimports**: `go install golang.org/x/tools/cmd/goimports@latest`

## Configuración golangci-lint

Crear `.golangci.yml` en la raíz:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck       # Detecta errores no verificados
    - govet          # Detecta problemas sutiles
    - staticcheck    # Análisis estático avanzado
    - unused         # Código no usado
    - gosimple       # Simplificaciones
    - ineffassign    # Asignaciones inefectivas
    - typecheck      # Errores de tipo
    - gocritic       # Estilo y performance
    - gofmt          # Formato estándar
    - goimports      # Imports ordenados

linters-settings:
  errcheck:
    check-type-assertions: true   # Detecta type assertions sin comma-ok

issues:
  exclude-dirs:
    - vendor
```

## CI/CD

Actualizar `.github/workflows/release.yml` para añadir un job de validación previo al release:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make check

  release:
    needs: test
    # ... goreleaser existente
```

## Reglas

1. **`make check` debe pasar antes de cualquier commit.** Incluye fmt, vet, lint y tests.
2. **CI ejecuta `make check` antes del release.**
3. **Coverage no debe bajar** — si se añade código, se añaden tests.
4. **No se commitean binarios** — el directorio `bin/` está en `.gitignore`.
