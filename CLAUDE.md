# user_service

Go service. Rules live in `.claude/rules/`, split by topic:

- [naming.md](.claude/rules/naming.md) — identifiers, package naming
- [error-handling.md](.claude/rules/error-handling.md) — error wrapping, sentinels, panic policy
- [structure-and-idioms.md](.claude/rules/structure-and-idioms.md) — package layout, idiomatic Go
- [concurrency.md](.claude/rules/concurrency.md) — goroutines, context, mutex
- [yagni.md](.claude/rules/yagni.md) — non-negotiable scope constraints
- [testing.md](.claude/rules/testing.md) — scoped to `*_test.go`
- [service-layer.md](.claude/rules/service-layer.md) — scoped to `service/` and `*service*.go`
