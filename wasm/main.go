//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/kingoftac/flagon/cli"
	"github.com/kingoftac/flagon/lua"
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
func loadLuaPlugin(script string) interface{} {
	output.Reset()

	engine := lua.NewEngine(globalCLI)

	// Load script from string instead of file
	err := engine.L.DoString(script)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Failed to load Lua script: %v", err)}
	}

	// Retrieve metadata from the last registered command
	if engine.LastCommand == nil {
		return `{"error": "No command found in Lua script"}`
	}

	// Serialize the command struct to a map
	cmd := engine.LastCommand
	args := make([]map[string]interface{}, len(cmd.Args))
	for i, arg := range cmd.Args {
		args[i] = map[string]interface{}{
			"name":        arg.Name,
			"description": arg.Description,
			"optional":    arg.Optional,
			"variadic":    arg.Variadic,
		}
	}

	data := map[string]interface{}{
		"name":        cmd.Name,
		"description": cmd.Description,
		"summary":     cmd.Summary,
		"hidden":      cmd.Hidden,
		"aliases":     cmd.Aliases,
		"args":        args,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Error serializing command: %v", err)
	}

	return string(jsonData)
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
