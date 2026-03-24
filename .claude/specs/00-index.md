# Specs del Proyecto http2postman

## Índice

| # | Spec | Descripción |
|---|------|-------------|
| 01 | [Arquitectura](01-architecture.md) | Estructura del proyecto, Clean Architecture, principios SOLID |
| 02 | [Modelos de Dominio](02-domain-models.md) | Structs tipados para Postman, HTTP files y environments |
| 03 | [Manejo de Errores](03-error-handling.md) | Error wrapping, validación, errores centinela, fail fast |
| 04 | [Testing (TDD)](04-testing.md) | Estrategia de testing, convenciones, cobertura mínima, casos obligatorios |
| 05 | [Clean Code](05-clean-code.md) | Naming, funciones, comentarios, formato, guard clauses |
| 06 | [Abstracción Filesystem](06-filesystem-abstraction.md) | Interfaz FileSystem, implementación real e in-memory |
| 07 | [Capa CLI](07-cli-layer.md) | Comandos thin, RunE, flags, version, protección sobreescritura |
| 08 | [Makefile y Tooling](08-makefile-and-tooling.md) | Build, test, lint, CI/CD, golangci-lint |
| 09 | [Patrones de Diseño](09-design-patterns.md) | Strategy, Builder, Factory, Composite aplicados al proyecto |

## Cómo usar estas specs

1. **Antes de implementar**: leer las specs relevantes para entender las convenciones.
2. **Al escribir código nuevo**: seguir TDD (Spec 04) y Clean Code (Spec 05).
3. **Al refactorizar**: migrar hacia la arquitectura objetivo (Spec 01) incrementalmente.
4. **En code review**: verificar cumplimiento de las specs aplicables.

## Prioridad de implementación

1. Modelos tipados (Spec 02) — base para todo lo demás
2. Abstracción filesystem (Spec 06) — habilita testing
3. Tests para código existente (Spec 04) — red de seguridad antes de refactorizar
4. Refactorizar parsers y converters (Specs 01, 03, 05)
5. Capa CLI (Spec 07) — version, flags, RunE
6. Makefile y tooling (Spec 08) — automatización
