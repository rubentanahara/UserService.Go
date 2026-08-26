# user_service

Gin + MongoDB user service. Create, authenticate, and manage users; bcrypt-hashed passwords,
case-insensitive unique email/username, JWT-protected routes.

## Requirements

- Go 1.26+
- Docker (for MongoDB, or to run the whole stack via Compose)

## Run

### Via Docker Compose (app + MongoDB)

```bash
echo "JWT_SECRET=change-me" > .env
docker compose up --build
```

### Locally

Start MongoDB:

```bash
docker run -d --name user_service-mongo -p 27017:27017 mongo:7
```

Start the server:

```bash
JWT_SECRET=change-me go run ./cmd/server
```

## Configuration

| Env var                | Default                          | Required |
|-------------------------|-----------------------------------|----------|
| `JWT_SECRET`             | —                                 | yes      |
| `MONGO_URI`              | `mongodb://localhost:27017`       | no       |
| `MONGO_DB`               | `user_service`                    | no       |
| `PORT`                   | `5001`                            | no       |
| `MONGO_MAX_POOL_SIZE`    | `100`                             | no       |

## Endpoints

| Method | Path                     | Auth      | Description                                          |
|--------|--------------------------|-----------|-------------------------------------------------------|
| GET    | `/health/live`           | —         | process is up                                          |
| GET    | `/health/ready`          | —         | process up and MongoDB reachable                      |
| POST   | `/users`                 | —         | register (`username`, `email`, `password`)            |
| POST   | `/auth/login`            | —         | log in (`email`, `password`), returns a JWT           |
| GET    | `/users`                 | Bearer    | list users (`?limit=&offset=`, default 20/0)          |
| GET    | `/users/:id`             | Bearer    | fetch user by id                                       |
| PUT    | `/users/:id`             | Bearer    | update user (`username`, `email`)                     |
| PUT    | `/users/:id/password`    | Bearer    | change password (`old_password`, `new_password`)      |
| DELETE | `/users/:id`             | Bearer    | delete user                                             |

`POST /users` and `POST /auth/login` are rate-limited per IP. All `/users` routes require
`Authorization: Bearer <token>` except registration; any valid token can act on any user id.

Returns `409` on create/update if the email or username is already taken, `404` if the id
doesn't exist, `401` on bad credentials or a missing/invalid token.

```bash
curl -X POST localhost:5001/users \
  -H 'Content-Type: application/json' \
  -d '{"username":"jane","email":"jane@example.com","password":"password123"}'

TOKEN=$(curl -s -X POST localhost:5001/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"jane@example.com","password":"password123"}' | jq -r .token)

curl localhost:5001/users -H "Authorization: Bearer $TOKEN"
```

## Structure

```
cmd/server/main.go        # entrypoint: middleware chain, HTTP server, graceful shutdown
cmd/server/wire.go         # DI injector definition (wireinject build tag)
cmd/server/wire_gen.go     # generated DI wiring
cmd/server/config.go       # env-driven config providers
cmd/server/server.go       # Server struct, Mongo client/collection providers
internal/user/model.go     # User struct
internal/user/repository.go   # Mongo access, unique indexes
internal/user/service.go       # business logic, password hashing, auth
internal/user/handler.go       # gin routes, error mapping
internal/auth/             # JWT issuing and middleware
internal/logging/          # slog setup, request-id middleware
internal/ratelimit/        # per-IP rate limiting
```

## Testing

```bash
go test ./...                      # unit tests (mocked repository/service)
go test -tags=integration ./...    # + real MongoDB via testcontainers (needs Docker)
```

## Conventions

Project rules live in [`.claude/rules/`](.claude/rules), indexed in [`CLAUDE.md`](CLAUDE.md).
