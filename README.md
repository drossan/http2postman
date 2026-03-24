# http2postman

`http2postman` es una herramienta de línea de comandos (CLI) que permite convertir archivos HTTP en colecciones de
Postman y viceversa. Facilita la gestión y el intercambio de colecciones de solicitudes HTTP entre diferentes entornos
y equipos.

## Instalación

### Usando Homebrew (macOS)

```sh
brew tap drossan/homebrew-tools
brew install http2postman
```

### Desde el código fuente

```sh
git clone https://github.com/drossan/http2postman.git
cd http2postman
make build
```

El binario se genera en `bin/http2postman`.

### Requisitos

- Go 1.22 o superior

## Uso

### Exportar solicitudes HTTP a una colección de Postman

```sh
http2postman export [directorio] [flags]
```

| Flag | Descripción | Default |
|------|-------------|---------|
| `-o, --output` | Ruta del archivo de salida | `import_postman_collection.json` |
| `-f, --force` | Sobreescribir si el archivo ya existe | `false` |

Lee los archivos `.http` en el directorio especificado y genera una colección de Postman en formato JSON.

**Ejemplo:**

Dada la siguiente estructura de directorios:

```
http-requests/
├── backend/
│   ├── auth.http
│   └── users.http
└── frontend/
    └── pages.http
```

```sh
http2postman export http-requests
http2postman export http-requests -o my_collection.json --force
```

### Importar una colección de Postman a archivos HTTP

```sh
http2postman import [archivo.json] [flags]
```

| Flag | Descripción | Default |
|------|-------------|---------|
| `-o, --output` | Directorio de salida | `http-requests` |
| `-f, --force` | Sobreescribir archivos existentes | `false` |

Lee una colección de Postman en formato JSON y genera archivos `.http` replicando la estructura de carpetas.

**Ejemplo:**

```sh
http2postman import collection.json
http2postman import collection.json -o my-requests --force
```

### Ver versión

```sh
http2postman version
```

## Formato de archivos .http

La herramienta usa el formato de IntelliJ HTTP Client. Cada archivo puede contener múltiples requests separados
por `###`:

```http
# Obtener usuarios
GET https://api.example.com/users
Authorization: Bearer {{token}}
Content-Type: application/json

###

# Crear usuario
POST https://api.example.com/users
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "name": "John",
  "email": "john@example.com"
}
```

## Variables de entorno

Si existe un archivo `http-client.env.json` en el directorio de los archivos `.http` (o en un directorio padre), las
variables se incluirán automáticamente en la colección de Postman exportada.

```json
{
  "dev": {
    "host": "https://dev.api.example.com",
    "token": "dev-token"
  },
  "prod": {
    "host": "https://api.example.com",
    "token": "prod-token"
  }
}
```

## Notas

- Soporta encabezados y cuerpos (raw y form-data) en las solicitudes HTTP.
- En la importación, si una solicitud no contiene autorización, se hereda del elemento padre (Bearer token).
- Se omiten claves vacías en datos de formularios al generar archivos `.http`.

## Plataformas soportadas

| OS | Arquitecturas |
|----|---------------|
| Linux | amd64, arm64, 386, arm |
| macOS | amd64, arm64 (Universal Binary) |
| Windows | amd64, arm64, 386, arm |

Las versiones estables e inestables están disponibles en la
[página de releases](https://github.com/drossan/http2postman/releases).

## Desarrollo

```sh
make build          # Compilar
make test           # Ejecutar tests
make test-coverage  # Tests con cobertura
make vet            # Análisis estático
make check          # fmt + vet + tests
make clean          # Limpiar artefactos
```

## Licencia

Este proyecto está bajo la licencia MIT. Ver [LICENSE](LICENSE) para más detalles.
