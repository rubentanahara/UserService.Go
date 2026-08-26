---
paths:
  - "**/service/**"
  - "**/*service*.go"
---

# Service Layer Design

- Layers: `Handler → Service → Repository → DB`. Service = business logic only, no HTTP, no SQL.
- Define the service interface from the consumer's perspective.
- Inject dependencies via constructor (`NewUserService(repo, hasher, logger)`).
- One thing per method — e.g. `Create` doesn't also send a welcome email.
- Validate input at the boundary (handler or service entry), not scattered across the codebase.
- Repository is an interface — swappable backend (Mongo, Postgres, etc.) without touching service code.
