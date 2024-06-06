# http2postman

`http2postman` es una herramienta de línea de comandos (CLI) que permite convertir archivos HTTP en colecciones de
Postman y viceversa. Esta herramienta facilita la gestión y el intercambio de colecciones de solicitudes HTTP entre
diferentes entornos y equipos.

## Instalación

### Desde el código fuente

1. Clona este repositorio:
    ```sh
    git clone https://github.com/tuusuario/http2postman.git
    cd http2postman
    ```

2. Construye la herramienta:
    ```sh
    go build -o http2postman
    ```

### Usando Homebrew en macOS

1. Añade el repositorio tap:
    ```sh
    brew tap drossan/homebrew-tools
    ```

2. Instala la herramienta:
    ```sh
    brew install http2postman
    ```

## Uso

### Exportar solicitudes HTTP a una colección de Postman

```sh
./http2postman export [directorio]

## Uso

### Exportar solicitudes HTTP a una colección de Postman

```bash
./http2postman export [directorio]
```

Este comando lee los archivos HTTP en el directorio especificado y crea una colección de Postman en formato JSON.

Ejemplo
Supongamos que tienes la siguiente estructura de directorios:

```text
/http-requests/
|-- backend
|   |-- auth.http
|   |-- users.http
```

Ejecuta el siguiente comando:

```bash
./http2postman export http-requests
```

La herramienta te pedirá que ingreses un nombre para la colección de Postman. Una vez ingresado, se generará un archivo
postman_collection.json con la estructura y contenido de las solicitudes HTTP.

Esto generará un archivo .json con el que ya puedes importan tu colección http en postman!

### Importar una colección de Postman a archivos HTTP (Experimental)

```bash
./http2postman import http-requests
```

Este comando lee una colección de Postman en formato JSON y crea archivos HTTP en el directorio http-requests,
replicando la estructura de la colección.

Ejemplo
Supongamos que tienes un archivo import_postman_collection.json con la colección de Postman. Ejecuta el siguiente
comando:

```bash
./http2postman import import_postman_collection.json
```

Esto creará archivos HTTP en el directorio http-requests según la estructura y contenido de la colección de Postman.

## Notas

La herramienta soporta encabezados y cuerpos en las solicitudes HTTP.
Si una solicitud en la colección de Postman no contiene una autorización, la herramienta buscará una autorización en los
elementos padre y la agregará a la solicitud.
La herramienta omitirá claves vacías en los datos de formularios al generar los archivos HTTP.

## Versiones Inestables

Además de las versiones estables, http2postman también proporciona versiones inestables que puedes probar. Estas
versiones están disponibles para los siguientes entornos:

- Linux
- macOS
- Windows

Puedes descargar las versiones inestables desde la página de lanzamientos del repositorio en GitHub.
