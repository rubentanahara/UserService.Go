# user_service

Gin + MongoDB user service. Create and fetch users, bcrypt-hashed passwords, unique email/username.

## Requirements

- Go 1.26+
- Docker (for MongoDB)

## Run

Start MongoDB:

```bash
docker run -d --name user_service-mongo -p 27017:27017 mongo:7
```

Start the server:

```bash
go run ./cmd/server
```

Server listens on `:5001`, connects to `mongodb://localhost:27017`, database `user_service`.

## Endpoints

| Method | Path            | Description                          |
|--------|-----------------|---------------------------------------|
| GET    | `/ping`         | liveness smoke check                  |
| GET    | `/health/live`  | process is up                         |
| GET    | `/health/ready` | process up and MongoDB reachable      |
| POST   | `/users`        | create user (`username`, `email`, `password`) |
| GET    | `/users/:id`    | fetch user by id                      |
| PUT    | `/users/:id`    | update user (`username`, `email`)     |
| DELETE | `/users/:id`    | delete user                           |

`POST /users` example:

```bash
curl -X POST localhost:5001/users \
  -H 'Content-Type: application/json' \
  -d '{"username":"jane","email":"jane@example.com","password":"password123"}'
```

Returns `409` on create/update if the email or username is already taken, `404` if the id doesn't exist.

## Structure

```
cmd/server/main.go       # wiring: Mongo client, indexes, router
internal/user/model.go   # User struct
internal/user/repository.go  # Mongo access, unique indexes
internal/user/service.go     # business logic, password hashing
internal/user/handler.go     # gin routes
```

## Conventions

Project rules live in [`.claude/rules/`](.claude/rules), indexed in [`CLAUDE.md`](CLAUDE.md).
