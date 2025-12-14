package cli

import "context"

type appKeyType struct{}
type commandKeyType struct{}
type argsKeyType struct{}
type flagsKeyType struct{}

var appKey = appKeyType{}
var commandKey = commandKeyType{}
var argsKey = argsKeyType{}
var flagsKey = flagsKeyType{}

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
