# Makefile for flagon

.PHONY: build clean run test

# Build the executable
build:
	go build -o flagon.exe ./cmd

# Clean build artifacts
clean:
	rm -f flagon.exe

# Run the application
run: build
	./flagon.exe

# Test all packages
test:
	go test ./cli
	go test ./cli-lua
