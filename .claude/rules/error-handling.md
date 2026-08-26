# Error Handling

- Return `error` as the last value, check immediately.
- Wrap with context: `fmt.Errorf("fetch user %s: %w", id, err)`.
- Sentinel errors: `var ErrNotFound = errors.New("not found")`, compare with `errors.Is`.
- Custom error types: use `errors.As`.
- No panic in library code. Panic only for unrecoverable programmer error.
- Never discard an error with `_` unless truly don't care (and it's obvious why).
