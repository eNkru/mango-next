PREFIX ?= /usr/local
INSTALL_DIR=$(PREFIX)/bin
GO_DIR := go
BINARY := mango

.PHONY: all build static run test check clean install uninstall go-build go-static go-test go-check go-run go-all

all: build

build:
	cd $(GO_DIR) && go build -o ../$(BINARY) ./cmd/mango/

static:
	cd $(GO_DIR) && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../$(BINARY) ./cmd/mango/

run:
	cd $(GO_DIR) && go run ./cmd/mango/

test:
	cd $(GO_DIR) && go test ./...

check:
	cd $(GO_DIR) && go vet ./...

clean:
	rm -f $(BINARY) mango-go

install: build
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)

# Aliases kept for existing docs / muscle memory
go-build: build
go-static: static
go-test: test
go-check: check
go-run: run
go-all: check test build
