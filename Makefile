PREFIX ?= /usr/local
CRYSTAL_FLAGS ?=
INSTALL_DIR=$(PREFIX)/bin

# Auto-detect pkg-config paths for brew-installed libraries (libarchive, webp)
# Works on both Intel Macs (/usr/local) and Apple Silicon (/opt/homebrew)
# Skipped entirely on Linux/Docker where brew is not available
empty :=
space := $(empty) $(empty)
HAS_BREW := $(shell which brew 2>/dev/null)
ifneq ($(HAS_BREW),)
  LIBARCHIVE_PREFIX := $(shell brew --prefix libarchive 2>/dev/null)
  WEBP_PREFIX := $(shell brew --prefix webp 2>/dev/null)
  EXTRA_PKG_CONFIG_PATHS := $(LIBARCHIVE_PREFIX)/lib/pkgconfig $(WEBP_PREFIX)/lib/pkgconfig
  ifneq ($(strip $(EXTRA_PKG_CONFIG_PATHS)),)
    _EXTRA_PKG := $(subst $(space),:,$(strip $(EXTRA_PKG_CONFIG_PATHS)))
    export PKG_CONFIG_PATH := $(_EXTRA_PKG)$(if $(PKG_CONFIG_PATH),:$(PKG_CONFIG_PATH))
  endif
endif

all: uglify | build

uglify:
	yarn
	yarn uglify

setup: libs
	yarn
	yarn gulp dev

build: libs
	crystal build src/mango.cr --release --progress --error-trace $(CRYSTAL_FLAGS)

static: uglify | libs
	crystal build src/mango.cr --release --progress --static --error-trace $(CRYSTAL_FLAGS)

libs:
	shards install --production
	$(MAKE) patch-libs
	$(MAKE) build-image-size

# Patches for Crystal 1.19+ compatibility with older shards
# These are applied after shards install to fix breaking changes:
#   - mg shard: DB::Database#driver was removed in crystal-db 0.14+
#   - archive shard: Crystal::System::FileInfo became a module in Crystal 1.19
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  SEDI = sed -i ''
else
  SEDI = sed -i
endif

patch-libs:
	@echo 'Patching lib/mg for crystal-db 0.14+ compatibility...'
	$(SEDI) 's|@db.driver.class.to_s == "SQLite3::Driver"|true|' lib/mg/src/mg/migration.cr
	@echo 'Patching lib/archive for Crystal 1.19+ compatibility...'
	grep -q 'require "file"' lib/archive/src/archive.cr || \
		{ echo 'require "file"' | cat - lib/archive/src/archive.cr >lib/archive/src/archive.cr.tmp && \
		  mv lib/archive/src/archive.cr.tmp lib/archive/src/archive.cr; }
	$(SEDI) 's|Crystal::System::FileInfo|::File::Info|g' lib/archive/src/archive.cr

# Build native extensions for image_size shard (libwebp + stbi)
# shards install --production skips postinstall scripts, so this must be run manually.
build-image-size:
	test -f lib/image_size/ext/libwebp/v1.1.0.tar.gz || \
		wget -q https://github.com/webmproject/libwebp/archive/v1.1.0.tar.gz -O lib/image_size/ext/libwebp/v1.1.0.tar.gz
	test -d lib/image_size/ext/libwebp/libwebp-1.1.0 || \
		tar xzf lib/image_size/ext/libwebp/v1.1.0.tar.gz -C lib/image_size/ext/libwebp/
	$(MAKE) -C lib/image_size/ext/libwebp/libwebp-1.1.0 -f makefile.unix
	$(MAKE) -C lib/image_size/ext/libwebp
	$(MAKE) -C lib/image_size/ext/stbi

run:
	crystal run src/mango.cr --error-trace $(CRYSTAL_FLAGS)

# Development mode with hot reload for frontend assets (LESS → CSS).
# Runs gulp in watch mode alongside the Crystal server.
# For full hot reload including Crystal source changes, install fswatch:
#   brew install fswatch                # macOS (Homebrew)
#   sudo port install fswatch           # macOS (MacPorts)
#   sudo apt install fswatch            # Linux
# Then use: make dev-full
dev:
	npx gulp dev
	@trap 'kill %1 2>/dev/null || true' EXIT; \
	npx gulp watch & \
	crystal run src/mango.cr --error-trace $(CRYSTAL_FLAGS)

# Full hot reload: watches LESS files AND Crystal source files.
# Requires fswatch (see install notes above).
dev-full:
	@command -v fswatch >/dev/null 2>&1 || { echo "Error: fswatch is not installed. Install it with: brew install fswatch"; exit 1; }
	npx gulp dev
	@trap 'kill 0' EXIT; \
	npx gulp watch & \
	while true; do \
		crystal run src/mango.cr --error-trace $(CRYSTAL_FLAGS) & \
		SERVER_PID=$$!; \
		fswatch -1 --include '\.cr$$' --latency 0.5 src/ 2>/dev/null; \
		echo "[dev] Restarting..."; \
		kill $$SERVER_PID 2>/dev/null; \
		wait $$SERVER_PID 2>/dev/null; \
		sleep 1; \
	done

test:
	crystal spec $(CRYSTAL_FLAGS)

check:
	crystal tool format --check
	./bin/ameba

arm32v7:
	crystal build src/mango.cr --release --progress --error-trace $(CRYSTAL_FLAGS) --cross-compile --target='arm-linux-gnueabihf' -o mango-arm32v7

arm64v8:
	crystal build src/mango.cr --release --progress --error-trace $(CRYSTAL_FLAGS) --cross-compile --target='aarch64-linux-gnu' -o mango-arm64v8

install:
	cp mango $(INSTALL_DIR)/mango

uninstall:
	rm -f $(INSTALL_DIR)/mango

cleandist:
	rm -rf dist
	rm -f yarn.lock
	rm -rf node_modules

clean:
	rm -f mango
