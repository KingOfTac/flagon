//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/kingoftac/flagon/cli-lua"
)

var (
	globalCLI *cli.CLI
	output    *bytes.Buffer
)

func init() {
	// Create a global CLI instance for the browser
	output = &bytes.Buffer{}
	globalCLI = cli.New(&cli.Command{
		Name:        "flagon-wasm",
		Description: "Flagon CLI running in WebAssembly",
	}, cli.WithWriters(output, output), cli.WithLogger(log.New(output, "", 0)))
}

// runCommand executes a CLI command and returns the output
func runCommand(cmd string) string {
	output.Reset()

	args := parseCommand(cmd)
	if len(args) == 0 {
		return "No command provided"
	}

	err := globalCLI.Run(args)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	result := output.String()
	if result == "" {
		result = "Command executed successfully"
	}

	return result
}

// loadLuaPlugin loads a Lua script as a plugin
func loadLuaPlugin(script string) string {
	output.Reset()

	engine := lua.NewEngine(globalCLI)

	// Load script from string instead of file
	err := engine.L.DoString(script)
	if err != nil {
		return fmt.Sprintf("Failed to load Lua script: %v", err)
	}

	result := output.String()
	if result == "" {
		result = "Lua plugin loaded successfully"
	}

	return result
}

// parseCommand splits a command string into arguments
func parseCommand(cmd string) []string {
	// Simple parsing - can be enhanced for quotes, etc.
	return strings.Fields(cmd)
}

func main() {
	// WASM entry point - register functions for JavaScript interop
	js.Global().Set("runCommand", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return "No command provided"
		}
		cmd := args[0].String()
		return runCommand(cmd)
	}))

	js.Global().Set("loadLuaPlugin", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return "No script provided"
		}
		script := args[0].String()
		return loadLuaPlugin(script)
	}))

	// Keep the program running for JavaScript calls
	select {}
}
