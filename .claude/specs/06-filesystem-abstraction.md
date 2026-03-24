# Spec 06: Abstracción de Filesystem

## Objetivo

Abstraer el acceso a filesystem detrás de una interfaz para permitir testing sin tocar disco y cumplir con Dependency Inversion (SOLID).

## Interfaz

```go
// internal/fs/filesystem.go

package fs

import "io/fs"

// FileSystem abstrae las operaciones de filesystem necesarias para el proyecto.
type FileSystem interface {
    // ReadFile lee el contenido completo de un archivo.
    ReadFile(path string) ([]byte, error)

    // WriteFile escribe contenido a un archivo, creando directorios padre si es necesario.
    WriteFile(path string, data []byte, perm fs.FileMode) error

    // Walk recorre un árbol de directorios ejecutando fn por cada entrada.
    Walk(root string, fn WalkFunc) error

    // Stat retorna información del archivo. Retorna error si no existe.
    Stat(path string) (fs.FileInfo, error)

    // MkdirAll crea un directorio y todos sus padres.
    MkdirAll(path string, perm fs.FileMode) error

    // FileExists retorna true si el archivo existe.
    FileExists(path string) bool
}

// WalkFunc es la firma del callback para Walk.
type WalkFunc func(path string, info fs.FileInfo, err error) error
```

## Implementación real

```go
// internal/fs/os_filesystem.go

package fs

import (
    "io/fs"
    "os"
    "path/filepath"
)

// OSFileSystem implementa FileSystem usando el filesystem real del SO.
type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
    return &OSFileSystem{}
}

func (o *OSFileSystem) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func (o *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, os.ModePerm); err != nil {
        return err
    }
    return os.WriteFile(path, data, perm)
}

func (o *OSFileSystem) Walk(root string, fn WalkFunc) error {
    return filepath.Walk(root, fn)
}

func (o *OSFileSystem) Stat(path string) (fs.FileInfo, error) {
    return os.Stat(path)
}

func (o *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
    return os.MkdirAll(path, perm)
}

func (o *OSFileSystem) FileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}
```

## Implementación in-memory para tests

```go
// internal/fs/memory_filesystem.go

package fs

// MemoryFileSystem implementa FileSystem en memoria para tests.
type MemoryFileSystem struct {
    Files map[string][]byte
    Dirs  map[string]bool
}

func NewMemoryFileSystem() *MemoryFileSystem {
    return &MemoryFileSystem{
        Files: make(map[string][]byte),
        Dirs:  make(map[string]bool),
    }
}

// ... implementar cada método operando sobre los maps
```

## Inyección de dependencias

Los servicios reciben el filesystem como dependencia:

```go
// En el parser
type HTTPFileParser struct {
    fs fs.FileSystem
}

func NewHTTPFileParser(filesystem fs.FileSystem) *HTTPFileParser {
    return &HTTPFileParser{fs: filesystem}
}

func (p *HTTPFileParser) ParseFile(path string) (*model.HTTPFile, error) {
    content, err := p.fs.ReadFile(path)
    // ...
}
```

```go
// En cmd/export.go — inyección desde CLI
parser := parser.NewHTTPFileParser(fs.NewOSFileSystem())
```

## Reglas

1. **Ningún paquete de `internal/` importa `os` directamente** para operaciones de archivo. Solo a través de la interfaz.
2. **La única excepción** es `internal/fs/os_filesystem.go` que implementa la interfaz.
3. **Tests unitarios siempre usan `MemoryFileSystem`**. Solo tests de integración usan `OSFileSystem`.
4. **`WriteFile` crea directorios padre** automáticamente — no requerir `MkdirAll` previo en el código de negocio.
