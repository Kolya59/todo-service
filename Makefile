SERVER_HOST ?= 127.0.0.1
SERVER_PORT ?= 4201
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DB_USER ?= kolya59
DB_PASSWORD ?= 12334566w
DB_NAME ?= todo
PROF_PORT ?= 2222
LOG_LEVEL ?= debug

all: build-server start-server

build-server:
	go build -ldflags "-s -w" -o ./bin/server.app ./cmd/todo-service/main.go

start-server:
	SERVER_HOST=$(SERVER_HOST) SERVER_PORT=$(SERVER_PORT) DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) PROF_PORT=$(PROF_PORT) LOG_LEVEL=$(LOG_LEVEL) ./bin/server.app