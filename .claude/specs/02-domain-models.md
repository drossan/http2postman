# Spec 02: Modelos de Dominio Tipados

## Objetivo

Eliminar el uso de `map[string]interface{}` y definir structs tipados que representen fielmente las estructuras de datos del proyecto. Esto previene panics por type assertions fallidas y da seguridad en compilación.

## Modelos Postman Collection v2.1.0

```go
// internal/model/postman.go

package model

// PostmanCollection representa una colección Postman v2.1.0 completa.
type PostmanCollection struct {
    Info     PostmanInfo     `json:"info"`
    Item     []PostmanItem   `json:"item"`
    Variable []PostmanVar    `json:"variable,omitempty"`
}

// PostmanInfo contiene los metadatos de la colección.
type PostmanInfo struct {
    Name        string `json:"name"`
    PostmanID   string `json:"_postman_id"`
    Description string `json:"description"`
    Schema      string `json:"schema"`
}

// PostmanItem puede ser un folder (tiene Item) o un request (tiene Request).
type PostmanItem struct {
    Name    string         `json:"name"`
    Item    []PostmanItem  `json:"item,omitempty"`
    Request *PostmanReq    `json:"request,omitempty"`
    Auth    *PostmanAuth   `json:"auth,omitempty"`
}

// IsFolder retorna true si el item es un folder (contiene sub-items).
func (i PostmanItem) IsFolder() bool {
    return i.Request == nil && len(i.Item) > 0
}

// PostmanReq representa una petición HTTP en formato Postman.
type PostmanReq struct {
    Method string          `json:"method"`
    URL    PostmanURL      `json:"url"`
    Header []PostmanHeader `json:"header,omitempty"`
    Body   *PostmanBody    `json:"body,omitempty"`
    Auth   *PostmanAuth    `json:"auth,omitempty"`
}

// PostmanURL representa una URL en formato Postman (puede ser string o struct).
type PostmanURL struct {
    Raw string `json:"raw"`
}

// PostmanHeader representa un header HTTP.
type PostmanHeader struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

// PostmanBody representa el body de una petición.
type PostmanBody struct {
    Mode     string           `json:"mode"`
    Raw      string           `json:"raw,omitempty"`
    FormData []PostmanFormData `json:"formdata,omitempty"`
}

// PostmanFormData representa un campo de formulario.
type PostmanFormData struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

// PostmanAuth representa configuración de autenticación.
type PostmanAuth struct {
    Type   string          `json:"type"`
    Bearer []PostmanKV     `json:"bearer,omitempty"`
}

// PostmanKV es un par key-value genérico.
type PostmanKV struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

// PostmanVar representa una variable de colección.
type PostmanVar struct {
    Key   string `json:"key"`
    Value string `json:"value"`
    Type  string `json:"type"`
}
```

## Modelos HTTP File

```go
// internal/model/httpfile.go

package model

// HTTPFile representa un archivo .http parseado con sus requests.
type HTTPFile struct {
    Path     string        // Ruta relativa al directorio base
    Requests []HTTPRequest
}

// HTTPRequest representa una petición HTTP individual dentro de un archivo .http.
type HTTPRequest struct {
    Name    string
    Method  string
    URL     string
    Headers []HTTPHeader
    Body    string
}

// HTTPHeader representa un header HTTP.
type HTTPHeader struct {
    Key   string
    Value string
}
```

## Modelo Environment

```go
// internal/model/environment.go

package model

// Environment representa el contenido de http-client.env.json.
// La clave exterior es el nombre del entorno, la interior son las variables.
type Environment map[string]map[string]string
```

## Reglas

1. **Todos los campos JSON deben tener tags `json`** con `omitempty` donde corresponda.
2. **No se permite `interface{}`** en los modelos. Si un campo puede ser de varios tipos, modelar con un custom UnmarshalJSON.
3. **PostmanURL necesita un `UnmarshalJSON` custom** porque Postman a veces envía la URL como string y otras como objeto. Manejar ambos casos.
4. **Los modelos son structs puros** — sin dependencias externas, sin lógica de IO.
5. **Cada modelo en su propio archivo** dentro de `internal/model/`.
