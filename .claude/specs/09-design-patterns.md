# Spec 09: Patrones de Diseño Aplicados

## Objetivo

Documentar los patrones de diseño que se aplican en el proyecto y cuándo usarlos.

## 1. Strategy Pattern — Parsers y Writers

Diferentes estrategias de parsing/writing intercambiables a través de interfaces:

```go
// Parser strategy
type RequestParser interface {
    Parse(content []byte) ([]model.HTTPRequest, error)
}

// Writer strategy
type CollectionWriter interface {
    Write(collection model.PostmanCollection, outputPath string) error
}
```

Esto permite añadir nuevos formatos (ej: cURL, Insomnia) implementando la interfaz sin modificar código existente (Open/Closed).

## 2. Builder Pattern — Construcción de PostmanCollection

La colección se construye incrementalmente. Un builder encapsula la lógica de inserción en la jerarquía:

```go
type CollectionBuilder struct {
    collection model.PostmanCollection
}

func NewCollectionBuilder(name string) *CollectionBuilder {
    return &CollectionBuilder{
        collection: model.PostmanCollection{
            Info: model.PostmanInfo{
                Name:   name,
                Schema: model.PostmanSchemaV210,
            },
        },
    }
}

// AddRequestToPath añade un request en la ruta de folders especificada.
// Crea los folders intermedios si no existen.
func (b *CollectionBuilder) AddRequestToPath(path []string, requests []model.PostmanItem) {
    // Navegar/crear la jerarquía de folders
    // Insertar los requests en el folder final
}

// AddVariables añade variables de entorno a la colección.
func (b *CollectionBuilder) AddVariables(vars []model.PostmanVar) {
    b.collection.Variable = append(b.collection.Variable, vars...)
}

// Build retorna la colección construida.
func (b *CollectionBuilder) Build() model.PostmanCollection {
    return b.collection
}
```

Esto reemplaza la función `addToCollection` actual que manipula maps con type assertions peligrosas.

## 3. Factory Method — Creación de requests

Encapsular la creación de requests Postman desde diferentes fuentes:

```go
// NewPostmanItemFromHTTPRequest crea un PostmanItem desde un HTTPRequest.
func NewPostmanItemFromHTTPRequest(req model.HTTPRequest) model.PostmanItem {
    item := model.PostmanItem{
        Name: req.Name,
        Request: &model.PostmanReq{
            Method: req.Method,
            URL:    model.PostmanURL{Raw: req.URL},
        },
    }

    for _, h := range req.Headers {
        item.Request.Header = append(item.Request.Header, model.PostmanHeader{
            Key: h.Key, Value: h.Value,
        })
    }

    if req.Body != "" {
        item.Request.Body = &model.PostmanBody{
            Mode: "raw",
            Raw:  req.Body,
        }
    }

    return item
}
```

## 4. Composite Pattern — Estructura jerárquica de PostmanItem

`PostmanItem` ya implementa naturalmente el patrón Composite: un item puede ser una hoja (request) o un compuesto (folder con sub-items):

```go
type PostmanItem struct {
    Name    string         `json:"name"`
    Item    []PostmanItem  `json:"item,omitempty"`    // Composite: sub-items
    Request *PostmanReq    `json:"request,omitempty"` // Leaf: request
}

func (i PostmanItem) IsFolder() bool {
    return i.Request == nil
}
```

## 5. Template Method — Flujo de Export/Import

El flujo de export e import siguen pasos similares que se pueden abstraer:

```
Export: Leer fuente → Parsear → Convertir → Escribir destino
Import: Leer fuente → Parsear → Convertir → Escribir destino
```

Cada paso es una implementación específica, pero el flujo es el mismo.

## Patrones a NO usar

- **Singleton**: No. Usar inyección de dependencias en su lugar.
- **Observer**: Innecesario para la escala actual del proyecto.
- **Abstract Factory**: Over-engineering para solo dos formatos.

## Reglas

1. **Aplicar patrones solo cuando resuelvan un problema real**, no por anticipación.
2. **Preferir composición sobre herencia** (Go no tiene herencia, pero aplica a embedding).
3. **Cada patrón debe simplificar el código**, no añadir complejidad. Si el patrón hace el código más difícil de entender, no usarlo.
