<p align="center">
  <a href="https://github.com/kingoftac/gork">
    <picture>
      <source srcset=".github/media/flagon_dark.png" media="(prefers-color-scheme: dark)">
      <source srcset=".github/media/flagon_light.png" media="(prefers-color-scheme: light)">
      <img src=".github/media/flagon_light.png" alt="flagon logo">
    </picture>
  </a>
</p>
<p align="center">A declarative CLI framework for GO</p>

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Examples](#examples)
- [API Documentation](#api-documentation)
- [Lua Plugin System](#lua-plugin-system)
- [Contributing](#contributing)
- [License](#license)

---

# Installation

```bash
go get github.com/kingoftac/flagon/cli
```

# Getting Started

Flagon is a declarative CLI framework for Go that allows you to define commands, arguments, flags, and handlers in a structured way. It supports middleware, hooks, and Lua scripting for plugins.

## Basic Example

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/kingoftac/flagon/cli"
)

func main() {
	c := cli.New(&cli.Command{
		Name:        "myapp",
		Description: "My awesome CLI app",
		Commands: []*cli.Command{
			{
				Name:        "greet",
				Description: "Greet someone",
				Args: []cli.Arg{
					{Name: "name", Description: "Name to greet"},
				},
				Handler: func(ctx context.Context) error {
					name := cli.Args(ctx)[0]
					log.Printf("Hello, %s!", name)
					return nil
				},
			},
		},
	})

	if err := c.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
```

Run it:

```bash
go run main.go greet "World"
# Output: Hello, World!
```

# Examples

## Using Flags

```go
{
	Name: "build",
	Description: "Build the project",
	Flags: func(fs *flag.FlagSet) {
		fs.Bool("verbose", false, "enable verbose output")
		fs.String("output", "build", "output directory")
	},
	Handler: func(ctx context.Context) error {
		flags := cli.Flags(ctx)
		verbose := flags["verbose"].(bool)
		output := flags["output"].(string)
		
		if verbose {
			log.Println("Building with verbose output")
		}
		log.Printf("Output directory: %s", output)
		return nil
	},
}
```

## Middleware

```go
func LoggingMiddleware(logger *log.Logger) cli.Middleware {
	return func(next cli.Handler) cli.Handler {
		return func(ctx context.Context) error {
			cmd := cli.CurrentCommand(ctx)
			logger.Printf("Executing command: %s", cmd.Name)
			err := next(ctx)
			logger.Printf("Finished command: %s", cmd.Name)
			return err
		}
	}
}

func main() {
	c := cli.New(&cli.Command{
		Name: "myapp",
		Middleware: []cli.Middleware{
			LoggingMiddleware(log.Default()),
		},
		Commands: []*cli.Command{
			{
				Name: "test",
				Handler: func(ctx context.Context) error {
					log.Println("Running tests...")
					return nil
				},
			},
		},
	})
	// ...
}
```

## Hooks

```go
func main() {
	c := cli.New(&cli.Command{
		Name: "myapp",
		Commands: []*cli.Command{
			{
				Name: "deploy",
				Before: []cli.Hook{
					func(ctx context.Context) error {
						log.Println("Pre-deployment checks...")
						return nil
					},
				},
				After: []cli.Hook{
					func(ctx context.Context) error {
						log.Println("Cleanup after deployment...")
						return nil
					},
				},
				Handler: func(ctx context.Context) error {
					log.Println("Deploying application...")
					return nil
				},
			},
		},
	})

	// Global hooks
	c.Hook(cli.BeforeRun, func(ctx context.Context) error {
		log.Println("Application starting...")
		return nil
	})
	c.Hook(cli.AfterRun, func(ctx context.Context) error {
		log.Println("Application finished.")
		return nil
	})
	// ...
}
```

## Nested Commands

```go
c := cli.New(&cli.Command{
	Name: "myapp",
	Commands: []*cli.Command{
		{
			Name: "db",
			Description: "Database operations",
			Commands: []*cli.Command{
				{
					Name: "migrate",
					Description: "Run database migrations",
					Handler: func(ctx context.Context) error {
						log.Println("Running migrations...")
						return nil
					},
				},
				{
					Name: "seed",
					Description: "Seed database",
					Handler: func(ctx context.Context) error {
						log.Println("Seeding database...")
						return nil
					},
				},
			},
		},
	},
})
```

## Command-Specific Middleware

```go
{
	Name: "sensitive",
	Description: "Sensitive operation",
	Middleware: []cli.Middleware{
		func(next cli.Handler) cli.Handler {
			return func(ctx context.Context) error {
				log.Println("Checking permissions...")
				// Permission check logic
				return next(ctx)
			}
		},
	},
	Handler: func(ctx context.Context) error {
		log.Println("Performing sensitive operation...")
		return nil
	},
}
```

# API Documentation

## Core Types

### Command

Defines a CLI command with its properties:

```go
type Command struct {
	Name        string
	Description string
	Summary     string
	Hidden      bool
	Aliases     []string
	Args        []Arg
	Flags       func(fs *flag.FlagSet)
	Handler     Handler
	Commands    []*Command
	Before      []Hook
	After       []Hook
	Middleware  []Middleware
}
```

### Arg

Defines a command argument:

```go
type Arg struct {
	Name        string
	Description string
	Optional    bool
	Variadic    bool
}
```

### Handler

Function signature for command handlers:

```go
type Handler func(ctx context.Context) error
```

### Middleware

Function for wrapping handlers:

```go
type Middleware func(next Handler) Handler
```

### Hook

Function for lifecycle hooks:

```go
type Hook func(ctx context.Context) error
```

## Key Functions

### New

Creates a new CLI instance:

```go
func New(root *Command, opts ...Option) *CLI
```

### Run

Executes the CLI with given arguments:

```go
func (c *CLI) Run(args []string) error
```

### Context Helpers

- `AppFromContext(ctx)`: Get the app instance
- `CurrentCommand(ctx)`: Get current command
- `Args(ctx)`: Get positional arguments
- `Flags(ctx)`: Get flag values

## Options

- `WithLogger(log.Logger)`: Set custom logger
- `WithAppData(map[string]any)`: Set app data
- `WithWriters(out, err io.Writer)`: Set output writers

# Lua Plugin System

Flagon supports extending CLI functionality with Lua scripts for dynamic command definition.

## Installation
```bash
go get github.com/kingoftac/flagon/lua
```

## Loading Plugins

```go
package main

import (
	"log"
	"os"

	"github.com/kingoftac/flagon/cli"
	"github.com/kingoftac/flagon/lua"
)

func main() {
	c := cli.New(&cli.Command{
		Name: "myapp",
	})

	// Load Lua plugins
	engine := lua.NewEngine(c)
	if err := engine.LoadFile("plugins/myplugin.lua"); err != nil {
		log.Fatal(err)
	}

	if err := c.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
```

## Authoring Plugins

Create a `.lua` file to define commands:

```lua
command {
  name = "hello",
  description = "Say hello",

  args = {
    { name = "name", description = "Name to greet" }
  },

  flags = {
    -- Lua doesn't handle flags directly, use Go for complex flags
  },

  middleware = {
    function(ctx)
      ctx.log("info", "Before hello")
      ctx.next()
      ctx.log("info", "After hello")
    end
  },

  handler = function(ctx)
    print("Hello, " .. ctx.args[1] .. "!")
  end
}
```

### Plugin Context

In Lua handlers and middleware, `ctx` provides:

- `ctx.args`: Array of positional arguments
- `ctx.log(level, message)`: Log messages
- `ctx.next()`: Call next middleware/handler (middleware only)

### Lua Environment

Plugins run in a sandboxed Lua environment with:

- Base libraries: `table`, `string`, `math`
- Safe functions only (no `dofile`, `loadfile`, etc.)
- `print` redirected to CLI logger

# Contributing

We welcome contributions! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Development

```bash
# Clone the repo
git clone https://github.com/kingoftac/flagon.git
cd flagon

# Run tests
go test ./...

# Build
make build
```

## Reporting Issues

Please report bugs and request features via [GitHub Issues](https://github.com/kingoftac/flagon/issues).

# License

This project is licensed under the MIT License - see the LICENSE file for details.
