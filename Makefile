MAKES-COMMIT := a7b80ec8f10ac700693278a559f315fccd50b5ad
M ?= .cache/makes
$(shell test -d $M || { \
  git clone -q https://github.com/makeplus/makes $M && \
  git -C $M checkout -q $(MAKES-COMMIT); \
})

GO-VERSION := 1.27.1

include $M/init.mk
include $M/go.mk
include $M/clean.mk
include $M/shell.mk

VERSION := 0.1.0
PLUGIN := yamlstar-plugin-yamlfmt
LIB := lib/lib$(PLUGIN).$(SO)
GO-YAML-ROOT ?= $(abspath ../..)

ifneq ($(wildcard $(GO-YAML-ROOT)/internal/libyaml/plugin_dumper_format.go),)
TEST-WORK := $(abspath .cache/go.work)
TEST-GOWORK := GOWORK=$(TEST-WORK)
endif

MAKES-CLEAN := lib .cache/abi-test .cache/go.work .cache/go.work.sum

default:: build

build: $(LIB)

test: $(LIB) $(TEST-WORK)
	$(TEST-GOWORK) $(GO) test ./...
	$(CC) -Wall -Wextra -Werror -Iinclude \
	  -DPLUGIN_EXTENSION='"$(SO)"' test/abi.c \
	  $(if $(IS-MACOS),,-ldl) -o .cache/abi-test
	YAMLSTAR_LIBRARY_PATH=$(abspath lib) .cache/abi-test

$(LIB): plugin.go cmd/shared/main.go plugin.edn go.mod go.sum $(GO)
	@mkdir -p $(dir $@)
	$(GO) build -buildmode=c-shared -o $@ ./cmd/shared

ifneq ($(TEST-WORK),)
$(TEST-WORK): go.mod $(GO)
	@mkdir -p $(dir $@)
	cd $(dir $@) && $(GO) work init .. $(GO-YAML-ROOT)
endif

format:
	$(GO) fmt ./...

install: $(LIB)
	install -d $(PREFIX)/lib
	install -m 755 $(LIB) $(PREFIX)/lib/
