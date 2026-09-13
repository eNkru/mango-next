PREFIX ?= /usr/local
INSTALL_DIR=$(PREFIX)/bin
GO_DIR := go
BINARY := mango

.PHONY: all build static run test check clean install uninstall frontend-install frontend-build frontend-check

all:
	@$(MAKE) check
	@$(MAKE) test
	@$(MAKE) build

frontend-install:
	npm ci

frontend-build:
	npm run build

frontend-check:
	npm run typecheck
	npm run check

build: frontend-build
	cd $(GO_DIR) && go build -o ../$(BINARY) ./cmd/mango/

static: frontend-build
	cd $(GO_DIR) && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../$(BINARY) ./cmd/mango/

run: frontend-build
	cd $(GO_DIR) && go run ./cmd/mango/

test:
	cd $(GO_DIR) && go test ./...

check: frontend-check
	cd $(GO_DIR) && go vet ./...

clean:
	rm -f $(BINARY) mango-go

install: build
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)
