.PHONY: build
build:
	go build  -o ./backend/bin/server ./backend/cmd/server/.

.PHONY: run
run:
	./backend/bin/server

.PHONY: test
test:
	go test -v ./...
