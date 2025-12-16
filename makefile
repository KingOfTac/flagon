# Makefile for flagon

SHELL=powershell.exe

.PHONY: test

# Run the application
run: build
	./flagon.exe

# Test all packages
test:
	cd cli && go test
	cd cli-lua && go test

# Build WASM version
build-wasm:
	$$env:GOOS='js'; $$env:GOARCH='wasm'; go build -o ./examples/public/flagon.wasm ./wasm
