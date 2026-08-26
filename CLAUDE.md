# user_service

Go service. Rules live in `.claude/rules/`, split by topic:

- [naming.md](.claude/rules/naming.md) — identifiers, package naming
- [error-handling.md](.claude/rules/error-handling.md) — error wrapping, sentinels, panic policy
- [structure-and-idioms.md](.claude/rules/structure-and-idioms.md) — package layout, idiomatic Go
- [concurrency.md](.claude/rules/concurrency.md) — goroutines, context, mutex
- [yagni.md](.claude/rules/yagni.md) — non-negotiable scope constraints
- [testing.md](.claude/rules/testing.md) — scoped to `*_test.go`
- [service-layer.md](.claude/rules/service-layer.md) — scoped to `service/` and `*service*.go`
- [logging.md](.claude/rules/logging.md) — `slog` usage, request-id, boundary logging

## Structure

`cmd/server` — entrypoint, Wire DI, config. `internal/user` — handler → service → repository.
`internal/auth` — JWT issue/verify. `internal/logging` — slog + request-id middleware.
`internal/ratelimit` — per-IP limiter. Full endpoint list: [README.md](README.md).

## Commands

```bash
make run               # needs JWT_SECRET + Mongo running
make test              # unit tests, mocked deps
make test-integration  # + testcontainers Mongo (needs Docker)
make generate          # regen wire + mocks after signature changes
make check             # fmt + vet + staticcheck + lint + test
```

## Generated code

`cmd/server/wire_gen.go` and `internal/user/*_mock_test.go` are generated
(`go:generate` — wire + mockgen), never hand-edit. After changing a
provider or an interface signature, run `go generate ./...`.
CI fails if this diverges from committed output.
