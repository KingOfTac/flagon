package lua

import (
	"context"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestNewLuaContext(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	luaCtx := newLuaContext(ctx, L)
	if luaCtx == nil {
		t.Fatal("newLuaContext returned nil")
	}

	// Check args table (should be empty since no args in context)
	args := luaCtx.RawGetString("args")
	if args.Type() != lua.LTTable {
		t.Error("args should be a table")
	}

	argsTable := args.(*lua.LTable)
	if argsTable.Len() != 0 {
		t.Errorf("Expected 0 args, got %d", argsTable.Len())
	}

	// Check log function
	logFn := luaCtx.RawGetString("log")
	if logFn.Type() != lua.LTFunction {
		t.Error("log should be a function")
	}

	// Check next function
	nextFn := luaCtx.RawGetString("next")
	if nextFn.Type() != lua.LTFunction {
		t.Error("next should be a function")
	}
}

func TestLuaHandler(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a simple Lua function that does nothing
	L.DoString(`testFunc = function() end`)
	fn := L.GetGlobal("testFunc").(*lua.LFunction)

	handler := luaHandler(fn, L)
	if handler == nil {
		t.Fatal("luaHandler returned nil")
	}

	// Test that handler can be called (should not error with valid function)
	err := handler(context.Background())
	if err != nil {
		t.Errorf("Handler errored: %v", err)
	}
}

func TestLuaMiddleware(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a simple Lua function that calls ctx.next()
	L.DoString(`testMiddleware = function(ctx) ctx.next() end`)
	fn := L.GetGlobal("testMiddleware").(*lua.LFunction)

	middleware := luaMiddleware(fn, L)
	if middleware == nil {
		t.Fatal("luaMiddleware returned nil")
	}

	wrapped := middleware(func(ctx context.Context) error {
		return nil
	})

	err := wrapped(context.Background())
	if err != nil {
		t.Errorf("Middleware errored: %v", err)
	}
}

func TestDecodeCommand(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	table := L.NewTable()
	table.RawSetString("name", lua.LString("testcmd"))
	table.RawSetString("description", lua.LString("test description"))

	cmd, err := decodeCommand(L, table)
	if err != nil {
		t.Errorf("decodeCommand failed: %v", err)
	}

	if cmd.Name != "testcmd" {
		t.Errorf("Expected name 'testcmd', got %s", cmd.Name)
	}

	if cmd.Description != "test description" {
		t.Errorf("Expected description 'test description', got %s", cmd.Description)
	}
}

func TestGetStringField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	table := L.NewTable()
	table.RawSetString("key", lua.LString("value"))

	val := getStringField(table, "key", false)
	if val != "value" {
		t.Errorf("Expected 'value', got %s", val)
	}

	val = getStringField(table, "missing", false)
	if val != "" {
		t.Errorf("Expected empty string for missing key, got %s", val)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for required missing key")
		}
	}()
	getStringField(table, "missing", true)
}

func TestGetBoolField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	table := L.NewTable()
	table.RawSetString("boolkey", lua.LTrue)

	val := getBoolField(table, "boolkey")
	if !val {
		t.Error("Expected true")
	}

	val = getBoolField(table, "missing")
	if val {
		t.Error("Expected false for missing key")
	}
}
