SERVER_HOST ?= 127.0.0.1
SERVER_PORT ?= 4201
PROF_PORT ?= 2222
LOG_LEVEL ?= DEBUG

build-start-server: build-server start-server

build-server:
	go build -ldflags "-s -w" -o ./bin/server.app ./cmd/todo-service/main.go

start-server:
	SERVER_HOST=$(SERVER_HOST) SERVER_PORT=$(SERVER_PORT) PROF_PORT=$(PROF_PORT) LOG_LEVEL=$(LOG_LEVEL) ./bin/server.app