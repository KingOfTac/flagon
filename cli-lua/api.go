package lua

import lua "github.com/yuin/gopher-lua"

func (e *Engine) installAPI() {
	L := e.L

	L.SetGlobal("command", L.NewFunction(e.luaCommand))
}

func (e *Engine) luaCommand(L *lua.LState) int {
	tbl := L.CheckTable(1)

	cmd, err := decodeCommand(L, tbl)
	if err != nil {
		L.RaiseError("command(): %s", err.Error())
		return 0
	}

	if err := e.registrar.RegisterCommand(nil, cmd); err != nil {
		L.RaiseError("command(): %s", err.Error())
		return 0
	}

	return 0
}
