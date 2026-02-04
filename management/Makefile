# Конфигурация
CONFIG_PATH ?= ./config/local.json
BINARY_NAME = main.exe
BINARY_PATH = ./tmp/$(BINARY_NAME)
CMD_PATH = ./cmd/main

DOCKER_COMPOSE = ../deployments/docker/docker-compose.yml

postgres:
	docker compose -f $(DOCKER_COMPOSE) up -d

psql:
	docker compose -f $(DOCKER_COMPOSE) exec postgres psql -U makhkets -d management

generate:
	@powershell -NoProfile -Command "if (-Not (Test-Path -Path '../proto/gen/go')) { New-Item -ItemType Directory -Force -Path '../proto/gen/go' | Out-Null }"
	protoc -I ../proto --go_out=../proto/gen/go --go_opt=paths=source_relative --go-grpc_out=../proto/gen/go --go-grpc_opt=paths=source_relative ../proto/management/*.proto ../proto/sso/*.proto
