package cli

import (
	"errors"
	"fmt"
	"strings"
)

type PluginRegistrar interface {
	RegisterCommand(parentPath []string, cmd *Command) error
	Use(m Middleware)
	Hook(phase HookPhase, h Hook)
	FindCommand(path ...string) (*Command, bool)
}

func (c *CLI) FindCommand(path ...string) (*Command, bool) {
	if len(path) == 0 {
		return c.Root, true
	}

	cur := c.Root

	for _, p := range path {
		next := findSubcommand(cur, p)
		if next == nil {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

func (c *CLI) RegisterCommand(parentPath []string, cmd *Command) error {
	if cmd == nil {
		return errors.New("command cannot be nil")
	}

	if strings.TrimSpace(cmd.Name) == "" {
		return errors.New("command name cannot be empty")
	}

	parent, ok := c.FindCommand(parentPath...)
	if !ok {
		return fmt.Errorf("parent command not found: %v", parentPath)
	}

	if collides(parent, cmd.Name) {
		return fmt.Errorf("command name collision: %s", cmd.Name)
	}
	for _, a := range cmd.Aliases {
		if collides(parent, a) {
			return fmt.Errorf("command alias collision: %s", a)
		}
	}

	parent.Commands = append(parent.Commands, cmd)
	return nil
}
