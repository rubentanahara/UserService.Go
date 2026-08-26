# Structure & Idioms

- `cmd/` for main packages, `internal/` for private/unimportable code.
- Package per feature, not package per layer, for internal code.
- Accept interfaces, return structs.
- Small interfaces, defined by the consumer, not the producer.
- `go fmt` non-negotiable, no style debates.
- Avoid `init()` unless registering (drivers, plugins) — hides control flow.
- No global mutable state.
- Struct embedding for composition, not inheritance-mimicry.
- Zero values should be useful (`var buf bytes.Buffer` just works).
- Don't over-generalize with generics — only when duplication is real (3+ concrete uses).
