#include "yamlstar_plugin.h"

#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifndef PLUGIN_EXTENSION
#define PLUGIN_EXTENSION "so"
#endif

typedef uint64_t (*abi_fn)(void);
typedef int32_t (*manifest_fn)(uint8_t **, size_t *);
typedef int32_t (*format_fn)(const uint8_t *, size_t, const uint8_t *,
                            size_t, uint8_t **, size_t *);
typedef void (*free_fn)(uint8_t *);

static void fail(const char *message) {
    fprintf(stderr, "%s\n", message);
    exit(1);
}

static char *take_output(free_fn free_output, uint8_t *output,
                         size_t length) {
    char *text = malloc(length + 1);
    if (text == NULL) {
        fail("malloc failed");
    }
    memcpy(text, output, length);
    text[length] = '\0';
    free_output(output);
    return text;
}

int main(void) {
    const char *directory = getenv("YAMLSTAR_LIBRARY_PATH");
    if (directory == NULL) {
        fail("YAMLSTAR_LIBRARY_PATH is not set");
    }
    char path[4096];
    snprintf(path, sizeof(path), "%s/libyamlstar-plugin-yamlfmt.%s",
             directory, PLUGIN_EXTENSION);
    void *handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
    if (handle == NULL) {
        fail(dlerror());
    }

    abi_fn abi = (abi_fn)dlsym(handle, "yamlstar_plugin_v1_abi");
    manifest_fn manifest = (manifest_fn)dlsym(
        handle, "yamlstar_plugin_v1_manifest");
    format_fn format = (format_fn)dlsym(
        handle, "yamlstar_plugin_v1_format");
    free_fn free_output = (free_fn)dlsym(
        handle, "yamlstar_plugin_v1_free");
    if (abi == NULL || manifest == NULL || format == NULL ||
        free_output == NULL) {
        fail("plugin ABI symbol is missing");
    }
    if (abi() != 1) {
        fail("plugin ABI version is not 1");
    }

    uint8_t *output = NULL;
    size_t length = 0;
    if (manifest(&output, &length) != 0) {
        fail("manifest call failed");
    }
    char *text = take_output(free_output, output, length);
    if (strstr(text, ":api \"dumper-format\"") == NULL ||
        strstr(text, ":options-format \"json-v1\"") == NULL) {
        fail("manifest is incorrect");
    }
    free(text);

    static const char input[] = "root:\n  child:\n  - one\n";
    static const char options[] =
        "{\"formatter\":{\"type\":\"basic\",\"indent\":4}}";
    if (format((const uint8_t *)input, strlen(input),
               (const uint8_t *)options, strlen(options),
               &output, &length) != 0) {
        fail("format call failed");
    }
    text = take_output(free_output, output, length);
    if (strcmp(text, "root:\n    child:\n        - one\n") != 0) {
        fail("format output is incorrect");
    }
    free(text);

    static const char bad_options[] = "{\"formatter\":false}";
    if (format((const uint8_t *)input, strlen(input),
               (const uint8_t *)bad_options, strlen(bad_options),
               &output, &length) != 1) {
        fail("invalid options did not return status 1");
    }
    text = take_output(free_output, output, length);
    if (strstr(text, "\"type\":\"format\"") == NULL) {
        fail("format error response is incorrect");
    }
    free(text);

    if (format(NULL, 1, NULL, 0, &output, &length) != 2) {
        fail("invalid buffer did not return status 2");
    }
    free_output(output);
    dlclose(handle);
    puts("shared plugin ABI test passed");
    return 0;
}
