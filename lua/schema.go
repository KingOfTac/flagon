package lua

import (
	"fmt"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/yuin/gopher-lua"
)

func decodeCommand(L *lua.LState, t *lua.LTable) (*cli.Command, error) {
	name := getStringField(t, "name", true)
	desc := getStringField(t, "description", false)

	cmd := &cli.Command{
		Name:        name,
		Description: desc,
	}

	if args := t.RawGetString("args"); args != lua.LNil {
		arr := args.(*lua.LTable)
		arr.ForEach(func(_ lua.LValue, v lua.LValue) {
			at := v.(*lua.LTable)
			cmd.Args = append(cmd.Args, cli.Arg{
				Name:        getStringField(at, "name", true),
				Description: getStringField(at, "description", false),
				Optional:    getBoolField(at, "optional"),
				Variadic:    getBoolField(at, "variadic"),
			})
		})
	}

	if h := t.RawGetString("handler"); h != lua.LNil {
		fn := h.(*lua.LFunction)
		cmd.Handler = luaHandler(fn, L)
	}

	if mw := t.RawGetString("middleware"); mw != lua.LNil {
		arr := mw.(*lua.LTable)
		arr.ForEach(func(_ lua.LValue, v lua.LValue) {
			fn := v.(*lua.LFunction)
			cmd.Middleware = append(cmd.Middleware, luaMiddleware(fn, L))
		})
	}

	return cmd, nil
}

func getStringField(t *lua.LTable, key string, required bool) string {
	v := t.RawGetString(key)
	if v == lua.LNil {
		if required {
			panic(fmt.Sprintf("missing required field '%s'", key))
		}
		return ""
	}
	return v.String()
}

func getBoolField(t *lua.LTable, key string) bool {
	v := t.RawGetString(key)
	if v == lua.LNil {
		return false
	}
	return lua.LVAsBool(v)
}
