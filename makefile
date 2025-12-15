# Makefile for flagon

SHELL=powershell.exe

.PHONY: build clean run test

# Build the executable
build:
	go build -o flagon.exe ./cmd

# Clean build artifacts
clean:
	if exist flagon.exe del flagon.exe
	if exist flagon.wasm del flagon.wasm

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
