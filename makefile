# -------------------------------------------------
# Project configuration
# -------------------------------------------------
GOOS ?= $(shell go env GOOS)
GOROOT := $(shell go env GOROOT)
WASM_OUT := ./examples/public/flagon.wasm
WASM_JS := ./examples/public/wasm_exec.js
WASM_JS_SRC := $(firstword $(wildcard $(GOROOT)/lib/wasm/wasm_exec.js) $(wildcard $(GOROOT)/misc/wasm/wasm_exec.js))

ifeq ($(WASM_JS_SRC),)
$(error Could not find wasm_exec.js in $(GOROOT)/lib/wasm or $(GOROOT)/misc/wasm)
endif

# -------------------------------------------------
# Go environment
# -------------------------------------------------

ifeq ($(GOOS),windows)
SHELL := powershell.exe
SHELLFLAGS := -NoProfile -Command
ENV_WASM := $$env:GOOS='js'; $$env:GOARCH='wasm';
COPY_WASM_JS := Copy-Item '$(WASM_JS_SRC)' '$(WASM_JS)' -Force
WASM_BUILD_CMD := $(ENV_WASM) go build -o '$(WASM_OUT)' ./wasm
else
ENV_WASM := GOOS=js GOARCH=wasm
COPY_WASM_JS := cp '$(WASM_JS_SRC)' '$(WASM_JS)'
WASM_BUILD_CMD := $(ENV_WASM) go build -o '$(WASM_OUT)' ./wasm
endif

# -------------------------------------------------
# Fuzz test configuration
# -------------------------------------------------
FUZZ_TIME ?= 30s
FUZZ_TEST ?= FuzzCLIRun

# -------------------------------------------------
# Targets
# -------------------------------------------------

.PHONY: test fuzz fuzz-all wasm-env build-wasm

test:
ifeq ($(GOOS),windows)
	cd cli; go test -fuzz=^$$;
	cd lua; go test -fuzz=^$$;
else
	cd cli && go test -fuzz='^$$'
	cd lua && go test -fuzz='^$$'
endif

fuzz:
ifeq ($(GOOS),windows)
	cd cli; go test -fuzz=$(FUZZ_TEST) -fuzztime=$(FUZZ_TIME);
	cd lua; go test -fuzz=$(FUZZ_TEST) -fuzztime=$(FUZZ_TIME);
else
	cd cli && go test -fuzz=$(FUZZ_TEST) -fuzztime=$(FUZZ_TIME)
	cd lua && go test -fuzz=$(FUZZ_TEST) -fuzztime=$(FUZZ_TIME)
endif

fuzz-all:
ifeq ($(GOOS),windows)
	cd cli; go test -fuzz=FuzzCLIRun -fuzztime=$(FUZZ_TIME);
	cd cli; go test -fuzz=FuzzValidatePositionalArgs -fuzztime=$(FUZZ_TIME);
	cd cli; go test -fuzz=FuzzFindSubcommand -fuzztime=$(FUZZ_TIME);
	cd cli; go test -fuzz=FuzzCollides -fuzztime=$(FUZZ_TIME);
	cd cli; go test -fuzz=FuzzContextFunctions -fuzztime=$(FUZZ_TIME);
	cd cli; go test -fuzz=FuzzMiddlewareChain -fuzztime=$(FUZZ_TIME);
	cd lua; go test -fuzz=FuzzLuaScriptExecution -fuzztime=$(FUZZ_TIME);
	cd lua; go test -fuzz=FuzzDecodeCommand -fuzztime=$(FUZZ_TIME);
	cd lua; go test -fuzz=FuzzLuaHandler -fuzztime=$(FUZZ_TIME);
	cd lua; go test -fuzz=FuzzLuaMiddleware -fuzztime=$(FUZZ_TIME);
else
	cd cli && go test -fuzz=FuzzCLIRun -fuzztime=$(FUZZ_TIME)
	cd cli && go test -fuzz=FuzzValidatePositionalArgs -fuzztime=$(FUZZ_TIME)
	cd cli && go test -fuzz=FuzzFindSubcommand -fuzztime=$(FUZZ_TIME)
	cd cli && go test -fuzz=FuzzCollides -fuzztime=$(FUZZ_TIME)
	cd cli && go test -fuzz=FuzzContextFunctions -fuzztime=$(FUZZ_TIME)
	cd cli && go test -fuzz=FuzzMiddlewareChain -fuzztime=$(FUZZ_TIME)
	cd lua && go test -fuzz=FuzzLuaScriptExecution -fuzztime=$(FUZZ_TIME)
	cd lua && go test -fuzz=FuzzDecodeCommand -fuzztime=$(FUZZ_TIME)
	cd lua && go test -fuzz=FuzzLuaHandler -fuzztime=$(FUZZ_TIME)
	cd lua && go test -fuzz=FuzzLuaMiddleware -fuzztime=$(FUZZ_TIME)
endif

wasm-env:
ifeq ($(GOOS),windows)
	$(SHELL) $(SHELLFLAGS) "$(COPY_WASM_JS)"
else
	$(COPY_WASM_JS)
endif

build-wasm: wasm-env
	$(WASM_BUILD_CMD)