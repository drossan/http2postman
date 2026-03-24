# Spec 01: Arquitectura y Estructura del Proyecto

## Objetivo

Reorganizar el proyecto siguiendo principios SOLID y Clean Architecture, separando responsabilidades en capas bien definidas.

## Estructura objetivo

```
http2postman/
├── cmd/                          # Capa de presentación (CLI)
│   ├── root.go                   # Comando raíz
│   ├── export.go                 # Comando export (solo orquestación CLI)
│   ├── import.go                 # Comando import (solo orquestación CLI)
│   └── version.go                # Comando version
├── internal/                     # Lógica interna (no exportable)
│   ├── model/                    # Entidades del dominio
│   │   ├── postman.go            # Structs tipados de Postman Collection v2.1.0
│   │   ├── httpfile.go           # Structs para representar archivos .http
│   │   └── environment.go        # Structs para http-client.env.json
│   ├── parser/                   # Parsers (Single Responsibility)
│   │   ├── httpfile_parser.go    # Parsea archivos .http → modelo interno
│   │   ├── httpfile_parser_test.go
│   │   ├── postman_parser.go     # Parsea JSON Postman → modelo interno
│   │   └── postman_parser_test.go
│   ├── converter/                # Conversores entre modelos
│   │   ├── http_to_postman.go    # Modelo HTTP → Modelo Postman
│   │   ├── http_to_postman_test.go
│   │   ├── postman_to_http.go    # Modelo Postman → Modelo HTTP
│   │   └── postman_to_http_test.go
│   ├── writer/                   # Escritores de salida
│   │   ├── postman_writer.go     # Escribe colección Postman a JSON
│   │   ├── postman_writer_test.go
│   │   ├── httpfile_writer.go    # Escribe archivos .http a disco
│   │   └── httpfile_writer_test.go
│   └── fs/                       # Abstracción de filesystem
│       ├── filesystem.go         # Interfaz FileSystem
│       └── os_filesystem.go      # Implementación real con os/filepath
├── main.go
├── go.mod
├── go.sum
├── Makefile
├── LICENSE
└── README.md
```

## Principios aplicados

### Single Responsibility (S)
- **cmd/**: Solo maneja input del usuario y delega a servicios.
- **parser/**: Solo parsea formatos de entrada.
- **converter/**: Solo transforma entre modelos.
- **writer/**: Solo escribe output a disco.

### Open/Closed (O)
- Nuevos formatos de auth (Basic, API Key, OAuth2) se añaden extendiendo el modelo y los converters, sin modificar el código existente.
- Nuevos formatos de salida se implementan con nuevos writers.

### Liskov Substitution (L)
- La interfaz `FileSystem` permite sustituir el FS real por uno en memoria para tests.

### Interface Segregation (I)
- Interfaces pequeñas y específicas: `Parser`, `Converter`, `Writer`.
- No se fuerza a implementar métodos que no se necesitan.

### Dependency Inversion (D)
- Los comandos CLI dependen de interfaces, no de implementaciones concretas.
- El filesystem se inyecta como dependencia, no se usa `os` directamente en la lógica de negocio.

## Reglas

1. **Nunca usar `map[string]interface{}`** para representar datos de dominio. Siempre structs tipados.
2. **Nunca hacer type assertions sin verificación** (comma-ok pattern obligatorio).
3. **Nunca usar `fmt.Println` para errores** — los errores se retornan, no se imprimen.
4. **Los comandos CLI solo orquestan**: reciben args, llaman al servicio, manejan el error.
5. **Todo acceso a filesystem se hace a través de la interfaz** `FileSystem`.
