APP_DIR         ?= ./backend
CMD_DIR         ?= $(APP_DIR)/cmd/server
ONLINECONF_CMD  ?= $(APP_DIR)/cmd/onlineconf
BIN_DIR         ?= $(APP_DIR)/bin
BIN_NAME        ?= server
ONLINECONF_BIN  ?= onlineconf
BIN_PATH        ?= $(BIN_DIR)/$(BIN_NAME)
ONLINECONF_PATH ?= $(BIN_DIR)/$(ONLINECONF_BIN)

GO              ?= go

.PHONY: all
all: build build-onlineconf

## build: Компиляция основного бинарного файла server
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_PATH) $(CMD_DIR)

## build-onlineconf: Компиляция бинарного файла сервиса onlineconf
.PHONY: build-onlineconf
build-onlineconf:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(ONLINECONF_PATH) $(ONLINECONF_CMD)

## run: Сборка и запуск основного приложения
.PHONY: run
run: build
	$(BIN_PATH)

## run-onlineconf: Сборка и запуск сервиса onlineconf
.PHONY: run-onlineconf
run-onlineconf: build-onlineconf
	$(ONLINECONF_PATH)

## clean: Удаление собранных бинарников
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

## help: Вывод списка доступных команд
.PHONY: help
help:
	@echo "Доступные команды:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':'

.PHONY: test
test:
	go test -coverprofile=$(BIN_DIR)/coverage.out ./... >/dev/null 2>&1
	go tool cover -func=$(BIN_DIR)/coverage.out | grep "total"
