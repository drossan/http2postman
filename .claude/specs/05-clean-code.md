# Spec 05: Clean Code y Convenciones

## Objetivo

Establecer convenciones de código limpio y estilo consistente para todo el proyecto.

## Naming

### Paquetes
- Nombres cortos, en minúsculas, una sola palabra: `parser`, `model`, `converter`, `writer`.
- No usar `utils`, `helpers`, `common`, `misc`.

### Funciones
- Verbos descriptivos: `ParseHTTPFile`, `ConvertToPostman`, `WriteCollection`.
- Funciones que retornan bool: prefijo `Is`/`Has`: `IsFolder()`, `HasAuth()`.
- Constructores: `New<Type>`: `NewCollection(name string) PostmanCollection`.

### Variables
- Nombres descriptivos, sin abreviaciones excepto las convencionales (`err`, `ok`, `ctx`, `i`).
- No usar `data`, `info`, `temp`, `result` como nombres genéricos.

```go
// MAL
d, err := os.ReadFile(p)
var r []map[string]interface{}

// BIEN
content, err := os.ReadFile(path)
var requests []HTTPRequest
```

### Constantes
- Agrupar en bloques con nombre del esquema:

```go
const (
    PostmanSchemaV210 = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    HTTPSectionSeparator = "###"
)
```

## Funciones

### Tamaño
- Máximo ~30 líneas por función. Si supera, extraer subfunciones con nombre descriptivo.
- El `export.go` actual tiene funciones de 40+ líneas que deben refactorizarse.

### Parámetros
- Máximo 3 parámetros. Si se necesitan más, agrupar en un struct de opciones:

```go
// MAL
func createHTTPRequestFile(filePath string, item map[string]interface{}, auth map[string]interface{}) error

// BIEN
func WriteHTTPRequest(req model.HTTPRequest, opts WriteOptions) error

type WriteOptions struct {
    BasePath string
    Force    bool  // sobreescribir existentes
}
```

### Retornos
- No retornar más de 2 valores (resultado + error).
- Si se necesitan más, usar un struct de resultado.

## Comentarios

- No comentar lo obvio. El código debe ser autoexplicativo.
- Comentar el "por qué", no el "qué".
- Documentar todas las funciones y tipos exportados con godoc:

```go
// ParseHTTPFile lee un archivo .http y retorna sus requests parseados.
// El archivo puede contener múltiples requests separados por "###".
func ParseHTTPFile(path string) (*HTTPFile, error) {
```

- Eliminar comentarios de código muerto (como los `// Uncomment...` en root.go).

## Formato

- Usar `gofmt` / `goimports` siempre (automatizado en Makefile y pre-commit).
- Imports agrupados en 3 bloques: stdlib, externo, interno:

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/drossan/http2postman/internal/model"
)
```

## Funciones deprecadas

- **No usar `strings.Title`** — usar `cases.Title(language.Und).String()` del paquete `golang.org/x/text`.
- **No usar `io/ioutil`** — usar `os.ReadFile`, `os.WriteFile`, `io.ReadAll`.

## Constantes mágicas

Eliminar strings/números mágicos. Extraer a constantes con nombre:

```go
// MAL
if body["mode"] == "raw" {
sections := strings.Split(string(content), "###")

// BIEN
if body.Mode == BodyModeRaw {
sections := strings.Split(string(content), HTTPSectionSeparator)
```

## Zero values útiles

Diseñar structs para que su zero value sea usable cuando sea posible:

```go
// PostmanCollection con zero value válido (info vacío, items vacíos)
collection := model.PostmanCollection{
    Info: model.PostmanInfo{
        Name:   name,
        Schema: model.PostmanSchemaV210,
    },
}
// collection.Item es nil, que se serializa como [] con json tag adecuado
```

## Guard clauses

Preferir early returns sobre if/else anidados:

```go
// MAL
func processFile(path string) error {
    if path != "" {
        content, err := os.ReadFile(path)
        if err == nil {
            // 20 líneas de lógica
        } else {
            return err
        }
    } else {
        return errors.New("path required")
    }
}

// BIEN
func processFile(path string) error {
    if path == "" {
        return errors.New("path required")
    }
    content, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("reading %s: %w", path, err)
    }
    // lógica principal sin indentación
}
```
