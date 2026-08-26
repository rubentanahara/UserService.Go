---
paths:
  - "**/*_test.go"
---

# Testing

- Table-driven tests, `t.Run` subtests.
- Same package as code under test (white-box) unless testing public API only (`package foo_test`).
- Prefer interface + fake struct over deep mock hierarchies.
- `go vet`, `staticcheck`, `golangci-lint` must pass in CI, non-negotiable.
