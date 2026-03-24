# Spec 07: Capa CLI (cmd/)

## Objetivo

Los comandos CLI deben ser thin wrappers que solo manejan input del usuario y delegan a la lógica de negocio. Nada de lógica de parseo, conversión, o IO directo.

## Cambios requeridos

### 1. Usar RunE en lugar de Run

```go
var exportCmd = &cobra.Command{
    Use:   "export [directory]",
    Short: "Export HTTP requests to a Postman collection",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        dir := args[0]

        name, err := promptCollectionName(cmd.InOrStdin())
        if err != nil {
            return fmt.Errorf("reading collection name: %w", err)
        }

        exporter := converter.NewHTTPToPostmanExporter(
            parser.NewHTTPFileParser(fs.NewOSFileSystem()),
            writer.NewPostmanWriter(fs.NewOSFileSystem()),
        )

        outputPath, err := exporter.Export(dir, name)
        if err != nil {
            return err
        }

        cmd.Printf("Collection exported to %s\n", outputPath)
        return nil
    },
}
```

### 2. Añadir comando version

```go
// cmd/version.go

var (
    version = "dev"
    commit  = "none"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version of http2postman",
    Run: func(cmd *cobra.Command, args []string) {
        cmd.Printf("http2postman %s (commit: %s)\n", version, commit)
    },
}
```

Y en `main.go`:
```go
var (
    version = "dev"
    commit  = "none"
)

func main() {
    cmd.SetVersionInfo(version, commit)
    cmd.Execute()
}
```

Esto conecta con los ldflags de `.goreleaser.yaml` que ya inyectan `main.version` y `main.commit`.

### 3. Flags útiles

```go
// En export
exportCmd.Flags().StringP("output", "o", "import_postman_collection.json", "Output file path")
exportCmd.Flags().BoolP("force", "f", false, "Overwrite output file if exists")

// En import
importCmd.Flags().StringP("output", "o", "http-requests", "Output directory path")
importCmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
```

### 4. Protección de sobreescritura

Antes de escribir, verificar si el archivo/directorio existe:

```go
if !force && fsys.FileExists(outputPath) {
    return fmt.Errorf("output %s already exists, use --force to overwrite", outputPath)
}
```

### 5. Prompt de stdin testeable

El prompt actual usa `bufio.NewReader(os.Stdin)` directamente. Hacerlo testeable inyectando el reader:

```go
func promptCollectionName(reader io.Reader) (string, error) {
    scanner := bufio.NewScanner(reader)
    fmt.Print("Enter the name for the Postman collection: ")
    if !scanner.Scan() {
        if err := scanner.Err(); err != nil {
            return "", fmt.Errorf("reading input: %w", err)
        }
        return "", model.ErrEmptyCollectionName
    }
    name := strings.TrimSpace(scanner.Text())
    if name == "" {
        return "", model.ErrEmptyCollectionName
    }
    return name, nil
}
```

## Reglas

1. **Los archivos de `cmd/` no importan `os` directamente** (excepto para `os.Stdin` que se pasa como `io.Reader`).
2. **Toda lógica de negocio se delega** a paquetes de `internal/`.
3. **Los mensajes al usuario usan `cmd.Printf`**, no `fmt.Println`.
4. **Los errores se retornan con `RunE`**, no se imprimen manualmente.
5. **Flags para toda configuración** — no hardcodear paths de salida.
