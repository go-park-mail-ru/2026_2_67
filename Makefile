APP_DIR   ?= ./backend
CMD_DIR   ?= $(APP_DIR)/cmd/server
BIN_DIR   ?= $(APP_DIR)/bin
BIN_NAME  ?= server
BIN_PATH  ?= $(BIN_DIR)/$(BIN_NAME)

GO        ?= go

.PHONY: all
all: build

## build: Компиляция бинарного файла
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_PATH) $(CMD_DIR)/.

## run: Сборка (если нужно) и запуск приложения
.PHONY: run
run: build
	$(BIN_PATH)

## clean: Удаление собранных бинарников
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

## help: Вывод списка доступных команд
.PHONY: help
help:
	@echo "Доступные команды:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':'
