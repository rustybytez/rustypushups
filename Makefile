.PHONY: build run dev test lint fmt tidy sqlc-gen docker-build docker-run

build:
	go build ./...

run:
	go run ./cmd/server

dev:
	go tool air -c .air.toml

test:
	go test ./...

test-pkg:
	go test ./$(PKG)/...

lint:
	go tool golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy

sqlc-gen:
	go tool sqlc generate

docker-build:
	docker build -t rustypushups .

docker-run:
	docker compose up --build
