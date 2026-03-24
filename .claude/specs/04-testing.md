# Spec 04: Testing (TDD)

## Objetivo

Establecer una estrategia de testing completa siguiendo TDD. Todo código nuevo debe tener tests escritos ANTES de la implementación.

## Flujo TDD obligatorio

1. **Red**: Escribir el test que describe el comportamiento esperado. Ejecutar y verificar que falla.
2. **Green**: Escribir el mínimo código necesario para que el test pase.
3. **Refactor**: Mejorar el código manteniendo los tests en verde.

## Estructura de tests

Los tests se ubican junto al código que testean, con el sufijo `_test.go`:

```
internal/
├── parser/
│   ├── httpfile_parser.go
│   ├── httpfile_parser_test.go
│   ├── postman_parser.go
│   └── postman_parser_test.go
├── converter/
│   ├── http_to_postman.go
│   ├── http_to_postman_test.go
│   ├── postman_to_http.go
│   └── postman_to_http_test.go
└── writer/
    ├── postman_writer.go
    ├── postman_writer_test.go
    ├── httpfile_writer.go
    └── httpfile_writer_test.go
```

## Convenciones de tests

### Naming

Usar el patrón `Test<Función>_<Escenario>`:

```go
func TestParseHTTPFile_SingleRequest(t *testing.T) { ... }
func TestParseHTTPFile_MultipleRequests(t *testing.T) { ... }
func TestParseHTTPFile_EmptyFile(t *testing.T) { ... }
func TestParseHTTPFile_MalformedHeaders(t *testing.T) { ... }
```

### Table-driven tests

Usar table-driven tests cuando hay múltiples escenarios para la misma función:

```go
func TestParseHTTPFile(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *model.HTTPFile
        wantErr  bool
    }{
        {
            name:  "single GET request",
            input: "# My Request\nGET https://api.example.com/users\n",
            expected: &model.HTTPFile{
                Requests: []model.HTTPRequest{
                    {Name: "My Request", Method: "GET", URL: "https://api.example.com/users"},
                },
            },
        },
        {
            name:    "empty file",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := parser.ParseHTTPContent(tt.input)
            if tt.wantErr {
                if err == nil {
                    t.Fatal("expected error but got nil")
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            // assertions...
        })
    }
}
```

### Testdata

Archivos de fixture se almacenan en `testdata/` dentro de cada paquete:

```
internal/parser/testdata/
├── simple_get.http
├── multiple_requests.http
├── with_headers_and_body.http
├── malformed.http
├── simple_collection.json
└── nested_collection.json
```

Se cargan con `os.ReadFile("testdata/simple_get.http")` (Go ejecuta tests con cwd = directorio del paquete).

## Cobertura mínima requerida

| Paquete | Cobertura mínima |
|---------|-----------------|
| model/ | 80% |
| parser/ | 90% |
| converter/ | 90% |
| writer/ | 80% |

## Casos de test obligatorios

### Parser HTTP
- Request GET simple sin headers ni body
- Request POST con headers y body JSON
- Request con múltiples secciones (separadas por ###)
- Archivo con formato inválido (sin método, sin URL)
- Archivo vacío
- Headers malformados (sin ":")
- Body multilínea

### Parser Postman
- Colección con un request simple
- Colección con folders anidados
- Request con Bearer auth
- Request con form-data body
- Request con URL como string vs objeto
- Colección con variables
- JSON inválido
- Colección sin campo "item"

### Converter HTTP → Postman
- Conversión de un request simple
- Preservación de la jerarquía de directorios como folders
- Inclusión de variables de entorno
- Formato del nombre del grupo

### Converter Postman → HTTP
- Conversión de un request simple
- Herencia de auth del folder padre
- Sanitización de nombres de archivo
- Manejo de form-data vs raw body

### Writer
- Escritura de colección JSON válida
- Escritura de archivos .http con formato correcto
- No sobreescribir archivos existentes sin flag --force

## Tests de integración (roundtrip)

Test end-to-end que verifica la ida y vuelta:

```go
func TestRoundtrip_ExportThenImport(t *testing.T) {
    // 1. Crear archivos .http temporales
    // 2. Ejecutar export → genera collection.json
    // 3. Ejecutar import sobre ese JSON → genera archivos .http
    // 4. Comparar archivos originales con los generados
}
```

## Ejecución

```bash
# Todos los tests
make test

# Con cobertura
make test-coverage

# Un paquete específico
go test ./internal/parser/...

# Un test específico
go test ./internal/parser/ -run TestParseHTTPFile_SingleRequest
```

## Reglas

1. **No se mergea código sin tests.** Todo PR debe incluir tests para el código nuevo.
2. **Los tests no dependen de filesystem real** (excepto tests de integración). Usar la interfaz FileSystem con implementación in-memory.
3. **Los tests no dependen del orden de ejecución.** Cada test es independiente.
4. **No se usa `t.Log` para assertions.** Usar `t.Fatal`, `t.Fatalf`, `t.Error`, `t.Errorf`.
5. **Tests de error verifican el mensaje o tipo de error**, no solo que `err != nil`.
