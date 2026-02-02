.PHONY: run dev build clean test lint tidy

# Конфигурация
CONFIG_PATH ?= ./config/local.json
BINARY_NAME = main.exe
BINARY_PATH = ./tmp/$(BINARY_NAME)
CMD_PATH = ./cmd/main

# docker-compose exec -it postgres psql -U makhkets -d management

# Запуск с hot-reload (air)
dev:
	air

# Запуск без hot-reload
run:
	go run $(CMD_PATH) -config=$(CONFIG_PATH)

# Сборка
build:
	go build -o $(BINARY_PATH) $(CMD_PATH)

# Запуск собранного бинарника
start: build
	$(BINARY_PATH) -config=$(CONFIG_PATH)

# Очистка
clean:
	rm -rf ./tmp/*
	go clean

# Тесты
test:
	go test -v ./...

# Линтер
lint:
	golangci-lint run

# Обновление зависимостей
tidy:
	go mod tidy

# Запуск с prod конфигом
run-prod:
	go run $(CMD_PATH) -config=./config/prod.json

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make dev       - Запуск с hot-reload (air)"
	@echo "  make run       - Запуск без hot-reload"
	@echo "  make build     - Сборка бинарника"
	@echo "  make start     - Сборка и запуск"
	@echo "  make clean     - Очистка tmp и кэша"
	@echo "  make test      - Запуск тестов"
	@echo "  make lint      - Запуск линтера"
	@echo "  make tidy      - go mod tidy"
	@echo "  make run-prod  - Запуск с prod конфигом"
