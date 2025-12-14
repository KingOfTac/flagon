package cli

import (
	"context"
	"io"
	"log"
	"strings"
)

type appKeyType struct{}
type commandKeyType struct{}
type argsKeyType struct{}
type flagsKeyType struct{}

var appKey = appKeyType{}
var commandKey = commandKeyType{}
var argsKey = argsKeyType{}
var flagsKey = flagsKeyType{}

type App struct {
	Logger *log.Logger
	Data   map[string]any
}

type Option func(*CLI)

func WithWriters(out, err io.Writer) Option {
	return func(c *CLI) {
		if out != nil {
			c.out = out
		}
		if err != nil {
			c.err = err
		}
	}
}

func WithLogger(l *log.Logger) Option {
	return func(c *CLI) {
		c.app.Logger = l
	}
}

func WithAppData(m map[string]any) Option {
	return func(c *CLI) {
		c.app.Data = m
	}
}

func WithHelpCommandName(name string) Option {
	return func(c *CLI) {
		if strings.TrimSpace(name) != "" {
			c.HelpCommandName = name
		}
	}
}

func AppFromContext(ctx context.Context) *App {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(appKey)
	if v == nil {
		return nil
	}
	return v.(*App)
}

func CurrentCommand(ctx context.Context) *Command {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(commandKey)
	if v == nil {
		return nil
	}
	return v.(*Command)
}

func Args(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(argsKey)
	if v == nil {
		return nil
	}
	return v.([]string)
}

func Flags(ctx context.Context) map[string]any {
	if ctx == nil {
		return map[string]any{}
	}
	v := ctx.Value(flagsKey)
	if v == nil {
		return map[string]any{}
	}
	return v.(map[string]any)
}
