package lua

import (
	"github.com/kingoftac/flagon/cli"
	lua "github.com/yuin/gopher-lua"
)

func (e *Engine) installSandbox() {
	L := e.L

	lua.OpenBase(L)
	lua.OpenTable(L)
	lua.OpenString(L)
	lua.OpenMath(L)

	dangerous := []string{
		"dofile",
		"loadfile",
		"load",
		"require",
		"collectgarbage",
	}

	for _, name := range dangerous {
		L.SetGlobal(name, lua.LNil)
	}

	// Intercept print to use CLI logger
	cli := e.registrar.(*cli.CLI)
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		var msg string
		for i := 1; i <= top; i++ {
			if i > 1 {
				msg += "\t"
			}
			msg += L.ToString(i)
		}
		cli.App().Logger.Println(msg)
		return 0
	}))
}
