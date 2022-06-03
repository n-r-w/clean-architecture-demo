.PHONY: build test run runbuild proto rebuild tidy race docker-up docker-down

build:
	go build -v -o . ./cmd/logserver

rebuild:
	go build -a -v -o . ./cmd/logserver

race:
	go run -race ./cmd/logserver

run:
	go run ./cmd/logserver -config-path ./config/server.toml

runbuild:
	./bin/logserver

tidy:
	go mod tidy

proto:
	protoc --proto_path=./api/proto --go_out=./internal/schema ./api/proto/log.proto

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

.DEFAULT_GOAL := run
