package lua

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/yuin/gopher-lua"
)

type Engine struct {
	L         *lua.LState
	registrar cli.PluginRegistrar
}

func NewEngine(registrar cli.PluginRegistrar) *Engine {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})

	e := &Engine{
		L:         L,
		registrar: registrar,
	}

	e.installSandbox()
	e.installAPI()

	return e
}

func (e *Engine) Close() {
	e.L.Close()
}

func (e *Engine) LoadFile(path string) error {
	if err := e.L.DoFile(path); err != nil {
		return fmt.Errorf("lua plugin error (%s): %w", path, err)
	}
	return nil
}

func (e *Engine) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".lua" {
			if err := e.LoadFile(filepath.Join(dir, entry.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}
