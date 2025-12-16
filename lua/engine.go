package lua

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/yuin/gopher-lua"
)

const (
	DefaultCallStackSize = 128
	DefaultRegistrySize  = 256
	DefaultScriptTimeout = 5 * time.Second
)

type Engine struct {
	L             *lua.LState
	registrar     cli.PluginRegistrar
	LastCommand   *cli.Command
	ScriptTimeout time.Duration
}

type EngineOption func(*Engine)

func WithScriptTimeout(d time.Duration) EngineOption {
	return func(e *Engine) {
		e.ScriptTimeout = d
	}
}

func NewEngine(registrar cli.PluginRegistrar, opts ...EngineOption) *Engine {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:  true,
		CallStackSize: DefaultCallStackSize,
		RegistrySize:  DefaultRegistrySize,
	})

	e := &Engine{
		L:             L,
		registrar:     registrar,
		ScriptTimeout: DefaultScriptTimeout,
	}

	for _, opt := range opts {
		opt(e)
	}

	e.installSandbox()
	e.installAPI()

	return e
}

func (e *Engine) Close() {
	e.L.Close()
}

func (e *Engine) DoString(script string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.ScriptTimeout)
	defer cancel()

	e.L.SetContext(ctx)
	defer e.L.RemoveContext()

	if err := e.L.DoString(script); err != nil {
		return err
	}
	return nil
}

func (e *Engine) LoadFile(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.ScriptTimeout)
	defer cancel()

	e.L.SetContext(ctx)
	defer e.L.RemoveContext()

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
