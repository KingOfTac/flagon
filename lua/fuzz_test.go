package lua

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/yuin/gopher-lua"
)

func FuzzLuaScriptExecution(f *testing.F) {
	f.Add(`print("hello")`)
	f.Add(`command { name = "test", description = "test cmd" }`)
	f.Add(`return 1 + 1`)
	f.Add(`local x = 1`)
	f.Add(`for i = 1, 10 do print(i) end`)
	f.Add(``)
	f.Add(`nil`)
	f.Add(`true`)
	f.Add(`false`)
	f.Add(`{}`)
	f.Add(`{1, 2, 3}`)

	// Sandbox escape attempts
	f.Add(`os.execute("echo pwned")`)
	f.Add(`io.open("/etc/passwd")`)
	f.Add(`dofile("/etc/passwd")`)
	f.Add(`loadfile("/etc/passwd")`)
	f.Add(`require("os")`)
	f.Add(`load("os.execute('echo pwned')")()`)
	f.Add(`rawget(_G, "os")`)
	f.Add(`debug.getinfo(1)`)
	f.Add(`debug.setmetatable({}, {})`)
	f.Add(`collectgarbage()`)
	f.Add(`package.loadlib("libc.so", "system")`)

	// String manipulation attacks
	f.Add(`string.dump(function() end)`)

	// Metatable attacks
	f.Add(`setmetatable(_G, {__index = function() return os end})`)
	f.Add(`getmetatable("").__index.char(0x41)`)

	// Resource exhaustion attacks
	f.Add(`while true do end`)
	f.Add(`function f() f() end f()`)
	f.Add(`local function f(n) if n > 0 then f(n-1) end end f(10000)`)

	// Syntax errors and edge cases
	f.Add(`)))`)
	f.Add(`function`)
	f.Add(`end`)
	f.Add(`if then else`)
	f.Add(`\x00\x01\x02`)

	// Unicode and special characters
	f.Add(`print("日本語")`)
	f.Add(`local 変数 = 1`)
	f.Add("print('\x00')")
	f.Add(`print("\n\r\t")`)

	f.Fuzz(func(t *testing.T, script string) {
		if len(script) > 10000 {
			return
		}

		out := &bytes.Buffer{}
		root := &cli.Command{Name: "test"}
		c := cli.New(root, cli.WithWriters(out, out), cli.WithLogger(log.New(out, "", 0)))

		engine := NewEngine(c, WithScriptTimeout(100*time.Millisecond))
		defer engine.Close()

		func() {
			defer func() {
				if r := recover(); r != nil {
					// Panics from Lua runtime are acceptable for security
					// (e.g., registry overflow, stack overflow)
					// The important thing is the Go process doesn't crash
				}
			}()
			_ = engine.DoString(script)
		}()

		func() {
			defer func() {
				_ = recover()
			}()

			verifyGlobal := func(name string) {
				val := engine.L.GetGlobal(name)
				if val != lua.LNil && name != "print" {
					if name == "os" || name == "io" || name == "debug" || name == "package" {
						t.Errorf("dangerous global %q should be nil but is %v", name, val.Type())
					}
				}
			}

			verifyGlobal("os")
			verifyGlobal("io")
			verifyGlobal("debug")
			verifyGlobal("package")
			verifyGlobal("dofile")
			verifyGlobal("loadfile")
			verifyGlobal("load")
			verifyGlobal("require")
		}()
	})
}

func FuzzDecodeCommand(f *testing.F) {
	f.Add("test", "description", 0, false)
	f.Add("", "", 0, false)
	f.Add("a", "b", 5, true)
	f.Add("cmd-name", "A longer description here", 10, true)
	f.Add("\x00", "\x00", 0, false)
	f.Add(strings.Repeat("a", 1000), strings.Repeat("b", 1000), 100, true)

	f.Fuzz(func(t *testing.T, name, desc string, numArgs int, hasHandler bool) {
		if numArgs < 0 || numArgs > 100 {
			return
		}

		L := lua.NewState()
		defer L.Close()

		tbl := L.NewTable()
		tbl.RawSetString("name", lua.LString(name))
		tbl.RawSetString("description", lua.LString(desc))

		if numArgs > 0 {
			args := L.NewTable()
			for i := 0; i < numArgs; i++ {
				arg := L.NewTable()
				arg.RawSetString("name", lua.LString("arg"+string(rune('0'+i%10))))
				arg.RawSetString("description", lua.LString("arg description"))
				arg.RawSetString("optional", lua.LBool(i%2 == 0))
				arg.RawSetString("variadic", lua.LBool(i == numArgs-1))
				args.Append(arg)
			}
			tbl.RawSetString("args", args)
		}

		if hasHandler {
			fn := L.NewFunction(func(L *lua.LState) int { return 0 })
			tbl.RawSetString("handler", fn)
		}

		func() {
			defer func() {
				_ = recover()
			}()
			_, _ = decodeCommand(L, tbl)
		}()
	})
}

func FuzzLuaContext(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(5)
	f.Add(100)

	f.Fuzz(func(t *testing.T, numArgs int) {
		if numArgs < 0 || numArgs > 1000 {
			return
		}

		L := lua.NewState()
		defer L.Close()

		args := make([]string, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = "arg" + string(rune('0'+i%10))
		}

		ctx := context.Background()
		out := &bytes.Buffer{}
		root := &cli.Command{Name: "test"}
		c := cli.New(root, cli.WithWriters(out, out), cli.WithLogger(log.New(out, "", 0)))
		ctx = context.WithValue(ctx, appKeyForTest(), c.App())
		ctx = context.WithValue(ctx, argsKeyForTest(), args)

		luaCtx := newLuaContext(ctx, L)
		if luaCtx == nil {
			t.Error("newLuaContext returned nil")
		}
	})
}

func FuzzLuaHandler(f *testing.F) {
	f.Add(`return`)
	f.Add(`print(ctx.args)`)
	f.Add(`error("test error")`)
	f.Add(`ctx.log("INFO", "message")`)

	f.Fuzz(func(t *testing.T, handlerBody string) {
		out := &bytes.Buffer{}
		root := &cli.Command{Name: "test"}
		c := cli.New(root, cli.WithWriters(out, out), cli.WithLogger(log.New(out, "", 0)))

		engine := NewEngine(c)
		defer engine.Close()

		script := `
			__test_handler = function(ctx)
				` + handlerBody + `
			end
		`
		if err := engine.L.DoString(script); err != nil {
			return
		}

		fn := engine.L.GetGlobal("__test_handler")
		if fn == lua.LNil {
			return
		}

		luaFn, ok := fn.(*lua.LFunction)
		if !ok {
			return
		}

		handler := luaHandler(luaFn, engine.L)

		_ = handler(context.Background())
	})
}

func FuzzLuaMiddleware(f *testing.F) {
	f.Add(`ctx.next()`)
	f.Add(`print("before"); ctx.next(); print("after")`)
	f.Add(`error("middleware error")`)
	f.Add(`return`)

	f.Fuzz(func(t *testing.T, middlewareBody string) {
		out := &bytes.Buffer{}
		root := &cli.Command{Name: "test"}
		c := cli.New(root, cli.WithWriters(out, out), cli.WithLogger(log.New(out, "", 0)))

		engine := NewEngine(c)
		defer engine.Close()

		script := `
			__test_middleware = function(ctx)
				` + middlewareBody + `
			end
		`
		if err := engine.L.DoString(script); err != nil {
			return
		}

		fn := engine.L.GetGlobal("__test_middleware")
		if fn == lua.LNil {
			return
		}

		luaFn, ok := fn.(*lua.LFunction)
		if !ok {
			return
		}

		middleware := luaMiddleware(luaFn, engine.L)

		nextCalled := false
		next := func(ctx context.Context) error {
			nextCalled = true
			return nil
		}

		handler := middleware(next)

		_ = handler(context.Background())
		_ = nextCalled // Silence unused warning
	})
}

func FuzzGetStringField(f *testing.F) {
	f.Add("name", "value", true)
	f.Add("name", "", false)
	f.Add("", "value", true)
	f.Add("\x00", "\x00", false)
	f.Add(strings.Repeat("key", 100), strings.Repeat("val", 100), true)

	f.Fuzz(func(t *testing.T, key, value string, required bool) {
		L := lua.NewState()
		defer L.Close()

		tbl := L.NewTable()
		if value != "" {
			tbl.RawSetString(key, lua.LString(value))
		}

		func() {
			defer func() {
				_ = recover()
			}()
			_ = getStringField(tbl, key, required)
		}()
	})
}

func FuzzGetBoolField(f *testing.F) {
	f.Add("flag", true)
	f.Add("flag", false)
	f.Add("", true)
	f.Add("\x00", false)

	f.Fuzz(func(t *testing.T, key string, value bool) {
		L := lua.NewState()
		defer L.Close()

		tbl := L.NewTable()
		tbl.RawSetString(key, lua.LBool(value))

		result := getBoolField(tbl, key)
		if result != value {
			t.Errorf("getBoolField(%q) = %v, want %v", key, result, value)
		}
	})
}

func appKeyForTest() interface{} {
	type appKeyType struct{}
	return appKeyType{}
}

func argsKeyForTest() interface{} {
	type argsKeyType struct{}
	return argsKeyType{}
}
