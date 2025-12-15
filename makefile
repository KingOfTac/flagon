# Makefile for flagon

.PHONY: build clean run test

# Build the executable
build:
	go build -o flagon.exe ./cmd

# Clean build artifacts
clean:
	rm -f flagon.exe flagon.wasm

# Run the application
run: build
	./flagon.exe

# Test all packages
test:
	cd cli && go test
	cd cli-lua && go test

# Build WASM version
build-wasm:
	GOOS=js GOARCH=wasm go build -o flagon.wasm ./wasm
