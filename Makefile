.PHONY: run build test test-integration coverage coverage-html generate fmt vet lint staticcheck check docker-build compose-up compose-down clean

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -race

test-integration:
	go test -tags=integration ./... -race

coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out

generate:
	go generate ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

staticcheck:
	staticcheck ./...

check: fmt vet staticcheck lint test

docker-build:
	docker build -t user_service .

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v

clean:
	rm -rf bin
