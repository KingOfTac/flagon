package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type Handler func(ctx context.Context) error

type CLI struct {
	app             *App
	ctx             context.Context
	Root            *Command
	out             io.Writer
	err             io.Writer
	hooks           map[HookPhase][]Hook
	Middleware      []Middleware
	HelpCommandName string
}

func New(root *Command, opts ...Option) *CLI {
	if root == nil {
		root = &Command{Name: "app"}
	}

	if root.Name == "" {
		root.Name = "app"
	}

	app := &App{
		Logger: log.New(os.Stdout, "[FLAGON] ", log.LstdFlags),
		Data:   map[string]any{},
	}

	c := &CLI{
		app:             app,
		ctx:             context.Background(),
		Root:            root,
		out:             os.Stdout,
		err:             os.Stderr,
		hooks:           map[HookPhase][]Hook{},
		HelpCommandName: "help",
	}

	for _, opt := range opts {
		opt(c)
	}

	c.ctx = context.WithValue(c.ctx, appKey, c.app)

	c.installHelpCommand()

	return c
}

func (c *CLI) App() *App {
	return c.app
}

func (c *CLI) Context() context.Context {
	return c.ctx
}

func (c *CLI) Run(args []string) error {
	for _, h := range c.hooks[BeforeRun] {
		if err := h(c.ctx); err != nil {
			return err
		}
	}

	var runErr error
	if len(args) == 0 {
		runErr = c.printHelp(c.Root)
	} else {
		runErr = c.execute(c.ctx, c.Root, args, nil)
	}

	for _, h := range c.hooks[AfterRun] {
		if err := h(c.ctx); err != nil && runErr == nil {
			runErr = err
		}
	}

	return runErr
}

func (c *CLI) execute(ctx context.Context, cmd *Command, args []string, parents []*Command) error {
	if len(args) > 0 {
		if sub := findSubcommand(cmd, args[0]); sub != nil {
			return c.execute(ctx, sub, args[1:], append(parents, cmd))
		}
	}

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.SetOutput((c.err))

	showHelp := false
	fs.BoolVar(&showHelp, "h", false, "show help")
	fs.BoolVar(&showHelp, "help", false, "show help")

	if cmd.Flags != nil {
		cmd.Flags(fs)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if showHelp {
		return c.printHelp(cmd)
	}

	parsedArgs := fs.Args()

	if err := validatePositionalArgs(cmd, parsedArgs); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, commandKey, cmd)
	ctx = context.WithValue(ctx, argsKey, parsedArgs)
	ctx = context.WithValue(ctx, flagsKey, snapshotFlags(fs))

	for _, h := range c.hooks[BeforeCommand] {
		if err := h(ctx); err != nil {
			return err
		}
	}

	for _, p := range parents {
		for _, h := range p.Before {
			if err := h(ctx); err != nil {
				return err
			}
		}
	}

	for _, h := range cmd.Before {
		if err := h(ctx); err != nil {
			return err
		}
	}

	if cmd.Handler == nil {
		return c.printHelp(cmd)
	}

	final := cmd.Handler
	final = applyMiddleware(final, cmd.Middleware)
	for i := len(parents) - 1; i >= 0; i-- {
		final = applyMiddleware(final, parents[i].Middleware)
	}
	final = applyMiddleware(final, c.Middleware)

	err := final(ctx)

	for _, h := range cmd.After {
		if hookErr := h(ctx); hookErr != nil && err == nil {
			err = hookErr
		}
	}

	for i := len(parents) - 1; i >= 0; i-- {
		for _, h := range parents[i].After {
			if hookErr := h(ctx); hookErr != nil && err == nil {
				err = hookErr
			}
		}
	}

	for _, h := range c.hooks[AfterCommand] {
		if hookErr := h(ctx); hookErr != nil && err == nil {
			err = hookErr
		}
	}

	return err
}

func (c *CLI) installHelpCommand() {
	if collides(c.Root, c.HelpCommandName) || c.Root.Name == c.HelpCommandName {
		return
	}

	help := &Command{
		Name:        c.HelpCommandName,
		Description: "Show help for a command",
		Summary:     "Show help",
		Args: []Arg{
			{Name: "path", Description: "Command path, e.g. project build", Optional: true, Variadic: true},
		},
		Handler: func(ctx context.Context) error {
			path := Args(ctx)
			if len(path) == 0 {
				return c.printHelp(c.Root)
			}
			target, ok := c.FindCommand(path...)
			if !ok {
				return fmt.Errorf("unknown command: %s", strings.Join(path, " "))
			}
			return c.printHelp(target)
		},
	}

	_ = c.RegisterCommand(nil, help)
}

func (c *CLI) printHelp(cmd *Command) error {
	if cmd == nil {
		return nil
	}

	w := c.out
	minSpacing := 8

	if cmd.Summary != "" {
		fmt.Fprintf(w, "%s - %s\n\n", cmd.Name, cmd.Summary)
	} else {
		fmt.Fprintf(w, "%s\n\n", cmd.Name)
	}

	if cmd.Description != "" {
		fmt.Fprintf(w, "%s\n\n", cmd.Description)
	}

	fmt.Fprintf(w, "Usage:\n  %s", cmd.Name)
	if len(cmd.Args) > 0 {
		for _, a := range cmd.Args {
			token := a.Name
			if a.Variadic {
				token = token + "..."
			}
			if a.Optional {
				fmt.Fprintf(w, " [%s]", token)
			} else {
				fmt.Fprintf(w, " <%s>", token)
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	if len(cmd.Args) > 0 {
		maxArgLen := 0
		for _, a := range cmd.Args {
			if len(a.Name) > maxArgLen {
				maxArgLen = len(a.Name)
			}
		}
		fmt.Fprintln(w, "Arguments:")
		for _, a := range cmd.Args {
			suffix := ""
			if a.Optional {
				suffix += " (optional)"
			}
			if a.Variadic {
				suffix += " (variadic)"
			}
			spacing := strings.Repeat(" ", minSpacing)
			fmt.Fprintf(w, "  %-*s%s%s%s\n", maxArgLen, a.Name, spacing, a.Description, suffix)
		}
		fmt.Fprintln(w)
	}

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	_ = fs.Bool("h", false, "show help")
	_ = fs.Bool("help", false, "show help")
	if cmd.Flags != nil {
		cmd.Flags(fs)
	}

	type flagInfo struct {
		name, usage, def string
	}
	var flags []flagInfo
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == "h" || f.Name == "help" {
			return
		}
		flags = append(flags, flagInfo{f.Name, f.Usage, f.DefValue})
	})
	if len(flags) > 0 {
		maxFlagLen := 0
		for _, fl := range flags {
			if len(fl.name) > maxFlagLen {
				maxFlagLen = len(fl.name)
			}
		}
		sort.Slice(flags, func(i, j int) bool {
			return flags[i].name < flags[j].name
		})
		fmt.Fprintln(w, "Flags:")
		for _, fl := range flags {
			spacing := strings.Repeat(" ", minSpacing)
			fmt.Fprintf(w, "  -%-*s%s%s (default %q)\n", maxFlagLen, fl.name, spacing, fl.usage, fl.def)
		}
		fmt.Fprintln(w)
	}

	var subs []*Command
	for _, sub := range cmd.Commands {
		if sub == nil || sub.Hidden {
			continue
		}
		subs = append(subs, sub)
	}
	if len(subs) > 0 {
		maxCmdLen := 0
		for _, sub := range subs {
			if len(sub.Name) > maxCmdLen {
				maxCmdLen = len(sub.Name)
			}
		}
		sort.Slice(subs, func(i, j int) bool {
			if subs[i].Name == c.HelpCommandName {
				return true
			}
			if subs[j].Name == c.HelpCommandName {
				return false
			}
			return subs[i].Name < subs[j].Name
		})
		fmt.Fprintln(w, "Commands:")
		for _, sub := range subs {
			desc := sub.Summary
			if desc == "" {
				desc = sub.Description
			}
			if desc == "" {
				desc = "-"
			}
			spacing := strings.Repeat(" ", minSpacing)
			fmt.Fprintf(w, "  %-*s%s%s\n", maxCmdLen, sub.Name, spacing, desc)
		}
		fmt.Fprintln(w)
	}

	return nil
}

func snapshotFlags(fs *flag.FlagSet) map[string]any {
	out := map[string]any{}
	if fs == nil {
		return out
	}

	fs.VisitAll(func(f *flag.Flag) {
		// Most stdlib flag values implement Get() any
		type getter interface{ Get() any }
		if g, ok := f.Value.(getter); ok {
			out[f.Name] = g.Get()
			return
		}

		// Fallback to string
		out[f.Name] = f.Value.String()
	})

	return out
}
