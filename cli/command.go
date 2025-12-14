package cli

import (
	"flag"
	"fmt"
)

type Arg struct {
	Name        string
	Description string
	Optional    bool
	Variadic    bool
}

type Command struct {
	Name        string
	Description string
	Summary     string
	Hidden      bool

	Aliases []string
	Args    []Arg
	Flags   func(fs *flag.FlagSet)

	Handler  Handler
	Commands []*Command

	Before []Hook
	After  []Hook

	Middleware []Middleware
}

func validatePositionalArgs(cmd *Command, parsed []string) error {
	if len(cmd.Args) == 0 {
		return nil
	}

	required := 0
	var hasVariadic bool

	for _, a := range cmd.Args {
		if a.Variadic {
			hasVariadic = true
		}
		if !a.Optional && !a.Variadic {
			required++
		}
	}

	if len(parsed) < required {
		return fmt.Errorf("missing required arguments (need %d)", required)
	}

	if !hasVariadic && len(parsed) > len(cmd.Args) {
		return fmt.Errorf("too many arguments (got %d, max %d)", len(parsed), len(cmd.Args))
	}

	return nil
}

func findSubcommand(cmd *Command, token string) *Command {
	if cmd == nil {
		return nil
	}

	for _, sub := range cmd.Commands {
		if sub == nil || sub.Hidden {
			continue
		}
		if sub.Name == token {
			return sub
		}
		for _, a := range sub.Aliases {
			if a == token {
				return sub
			}
		}
	}

	return nil
}

func collides(parent *Command, nameOrAlias string) bool {
	if parent == nil {
		return false
	}
	for _, sub := range parent.Commands {
		if sub == nil {
			continue
		}
		if sub.Name == nameOrAlias {
			return true
		}
		for _, a := range sub.Aliases {
			if a == nameOrAlias {
				return true
			}
		}
	}
	return false
}
