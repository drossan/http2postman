# Spec 03: Manejo de Errores

## Objetivo

Establecer un manejo de errores consistente, con contexto suficiente para diagnosticar problemas, siguiendo las convenciones idiomáticas de Go.

## Reglas

### 1. Siempre envolver errores con contexto

Usar `fmt.Errorf` con el verbo `%w` para mantener la cadena de errores:

```go
// MAL
return err

// BIEN
return fmt.Errorf("parsing HTTP file %s: %w", path, err)
```

### 2. Nunca ignorar errores silenciosamente

```go
// MAL
name, _ := reader.ReadString('\n')

// BIEN
name, err := reader.ReadString('\n')
if err != nil {
    return "", fmt.Errorf("reading collection name from stdin: %w", err)
}
```

**Excepción:** Solo se permite ignorar errores con `_` cuando el comportamiento por defecto del zero-value es aceptable Y se documenta con un comentario:

```go
// ok to ignore: empty string is valid default for optional field
value, _ := headerMap["value"].(string)
```

### 3. Nunca usar fmt.Println/Printf para reportar errores

Los errores se retornan, no se imprimen. Solo la capa CLI (cmd/) decide cómo presentar errores al usuario:

```go
// MAL (en lógica de negocio)
fmt.Printf("Invalid header format: %s\n", line)
continue

// BIEN (retornar error o acumular warnings)
warnings = append(warnings, fmt.Sprintf("invalid header format at line %d: %s", lineNum, line))
```

### 4. Type assertions siempre con comma-ok

```go
// MAL (panic si falla)
group := item.(map[string]interface{})

// BIEN
group, ok := item.(map[string]interface{})
if !ok {
    return fmt.Errorf("expected map but got %T", item)
}
```

**Nota:** Con la migración a structs tipados (Spec 02), las type assertions desaparecerán casi por completo. Esta regla aplica a cualquier caso residual.

### 5. Errores centinela para casos conocidos

Definir errores centinela en `internal/model/errors.go` para casos de negocio predecibles:

```go
package model

import "errors"

var (
    ErrInvalidHTTPFormat    = errors.New("invalid HTTP request format")
    ErrInvalidURLFormat     = errors.New("invalid URL line format")
    ErrInvalidCollection    = errors.New("invalid Postman collection format")
    ErrEnvFileNotFound      = errors.New("http-client.env.json not found")
    ErrEmptyCollectionName  = errors.New("collection name cannot be empty")
)
```

### 6. Validación temprana (fail fast)

Validar inputs al inicio de cada función pública, antes de hacer trabajo:

```go
func ParseHTTPFile(path string) (*HTTPFile, error) {
    if path == "" {
        return nil, fmt.Errorf("path is required")
    }
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading file %s: %w", path, err)
    }
    if len(content) == 0 {
        return nil, fmt.Errorf("file %s is empty", path)
    }
    // ... continuar
}
```

### 7. Uso de RunE en Cobra en lugar de Run

Para propagar errores correctamente desde los comandos CLI:

```go
// MAL
Run: func(cmd *cobra.Command, args []string) {
    if err := doSomething(); err != nil {
        fmt.Println("Error:", err)
    }
}

// BIEN
RunE: func(cmd *cobra.Command, args []string) error {
    if err := doSomething(); err != nil {
        return fmt.Errorf("export failed: %w", err)
    }
    return nil
}
```
