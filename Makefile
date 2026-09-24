MAKES-COMMIT := a7b80ec8f10ac700693278a559f315fccd50b5ad
M ?= .cache/makes
$(shell test -d $M || { \
  git clone -q https://github.com/makeplus/makes $M && \
  git -C $M checkout -q $(MAKES-COMMIT); \
})

GO-VERSION := 1.27.1

include $M/init.mk
include $M/perl.mk
include $M/go.mk
include $M/clean.mk
include $M/shell.mk

VERSION := 0.1.0
RELEASE-TAG ?= v$(VERSION)
PLUGIN := yamlstar-plugin-yamlfmt
LIB := lib/lib$(PLUGIN).$(SO)
RELEASE-ARCH-linux-int64 := x64
RELEASE-ARCH-linux-arm64 := aarch64
RELEASE-ARCH-macos-int64 := x64
RELEASE-ARCH-macos-arm64 := arm64
RELEASE-ARCH := $(RELEASE-ARCH-$(OS-NAME)-$(ARCH-NAME))
RELEASE-PLATFORM := $(OS-NAME)-$(RELEASE-ARCH)
RELEASE-NAME := $(PLUGIN)-$(RELEASE-TAG)-$(RELEASE-PLATFORM)
RELEASE-ROOT := dist/$(RELEASE-NAME)
RELEASE-ARCHIVE := $(RELEASE-ROOT).tar.xz

ifeq (,$(RELEASE-ARCH))
$(error Unsupported release platform: $(OS-NAME)-$(ARCH-NAME))
endif

MAKES-CLEAN := \
  lib \
  dist \
  .cache/abi-test \
  .cache/go.work \
  .cache/go.work.sum \
  .cache/release-test

default:: build

build: $(LIB)

test: $(LIB)
	GOWORK=off $(GO) test ./...
	$(CC) -Wall -Wextra -Werror -Iinclude \
	  -DPLUGIN_EXTENSION='"$(SO)"' test/abi.c \
	  $(if $(IS-MACOS),,-ldl) -o .cache/abi-test
	YAMLSTAR_LIBRARY_PATH=$(abspath lib) .cache/abi-test

test-race: $(GO)
	GOWORK=off $(GO) test -race ./...

$(LIB): plugin.go cmd/shared/main.go plugin.edn go.mod go.sum $(GO)
	@mkdir -p $(dir $@)
	GOWORK=off $(GO) build -buildmode=c-shared -o $@ ./cmd/shared

format:
	GOWORK=off $(GO) fmt ./...

install: $(LIB)
	install -d $(PREFIX)/lib
	install -m 755 $(LIB) $(PREFIX)/lib/

release: test-release

release-check: $(PERL)
	@version=$$($(PERL) -ne \
	  'print $$1 if /:version "([0-9]+\.[0-9]+\.[0-9]+)"/' \
	  plugin.edn); \
	  test "$$version" = "$(VERSION)" && \
	  test "$(RELEASE-TAG)" = "v$(VERSION)" || { \
	    echo "plugin.edn, VERSION, and RELEASE-TAG do not match" >&2; \
	    exit 1; \
	  }

$(RELEASE-ARCHIVE): release-check $(LIB) plugin.edn License ReadMe.md
	rm -rf $(RELEASE-ROOT)
	mkdir -p $(RELEASE-ROOT)/lib
	cp $(LIB) $(RELEASE-ROOT)/lib/
	cp plugin.edn License ReadMe.md $(RELEASE-ROOT)/
	cd dist && tar -cJf $(notdir $@) $(RELEASE-NAME)

test-release: $(RELEASE-ARCHIVE)
	rm -rf .cache/release-test
	mkdir -p .cache/release-test
	tar -xf $(RELEASE-ARCHIVE) -C .cache/release-test
	test -f .cache/release-test/$(RELEASE-NAME)/lib/$(notdir $(LIB))
	$(CC) -Wall -Wextra -Werror -Iinclude \
	  -DPLUGIN_EXTENSION='"$(SO)"' test/abi.c \
	  $(if $(IS-MACOS),,-ldl) -o .cache/release-test/abi-test
	YAMLSTAR_LIBRARY_PATH=$(abspath .cache/release-test/$(RELEASE-NAME)/lib) \
	  .cache/release-test/abi-test
