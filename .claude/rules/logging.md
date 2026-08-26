# Logging

- `log/slog` only — no other logging library.
- One `*slog.Logger` built once in `main.go`, injected into constructors (`NewHandler(service, logger)`), never a package-level/`init()` singleton.
- `context.Context` carries a request id, never the logger itself.
- Log at boundaries (HTTP handler, request middleware) — domain/service code returns wrapped errors, doesn't log them.
- Structured attributes via `slog.Group`, not flat key soup or string-formatted messages.
