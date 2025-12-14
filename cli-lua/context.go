package lua

import (
	"context"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/yuin/gopher-lua"
)

func newLuaContext(ctx interface{}, L *lua.LState) *lua.LTable {
	t := L.NewTable()

	args := L.NewTable()
	for _, a := range cli.Args(ctx.(context.Context)) {
		args.Append(lua.LString(a))
	}

	t.RawSetString("args", args)

	app := cli.AppFromContext(ctx.(context.Context))
	t.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		level := L.CheckString(1)
		msg := L.CheckString(2)
		if app != nil && app.Logger != nil {
			app.Logger.Printf("[%s] %s", level, msg)
		}
		return 0
	}))

	t.RawSetString("next", L.NewFunction(func(L *lua.LState) int {
		// do nothing
		return 0
	}))

	return t
}

func luaHandler(fn *lua.LFunction, L *lua.LState) cli.Handler {
	return func(ctx context.Context) error {
		L.Push(fn)
		L.Push(newLuaContext(ctx, L))

		if err := L.PCall(1, 0, nil); err != nil {
			return err
		}
		return nil
	}
}

func luaMiddleware(fn *lua.LFunction, L *lua.LState) cli.Middleware {
	return func(next cli.Handler) cli.Handler {
		return func(ctx context.Context) error {
			L.Push(fn)
			L.Push(newLuaContext(ctx, L))

			if err := L.PCall(1, 0, nil); err != nil {
				return err
			}

			return next(ctx)
		}
	}
}
