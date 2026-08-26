# Concurrency

- Don't start a goroutine without knowing how it stops.
- Pass `context.Context` as the first param, use for cancellation/timeout/deadline only — not for optional params.
- `context.Context` never stored in a struct field — pass explicitly.
- `sync.Mutex` zero-value is ready to use, don't init.
- Guard shared state; prefer channels for orchestration, mutex for simple state protection.
- Always `defer wg.Done()` / `defer mu.Unlock()` right after acquire.
- Benchmark before optimizing (`go test -bench`), don't guess.
