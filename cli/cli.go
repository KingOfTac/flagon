package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type Handler func(ctx context.Context) error
type Hook func(ctx context.Context) error
type Middleware func(next Handler) Handler

type HookPhase int

const (
	BeforeRun HookPhase = iota
	AfterRun
	BeforeCommand
	AfterCommand
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

type App struct {
	Logger *log.Logger
	Data   map[string]any
}

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

type PluginRegistrar interface {
	RegisterCommand(parentPath []string, cmd *Command) error
	Use(m Middleware)
	Hook(phase HookPhase, h Hook)
	FindCommand(path ...string) (*Command, bool)
}

type Option func(*CLI)

func WithWriters(out, err io.Writer) Option {
	return func(c *CLI) {
		if out != nil {
			c.out = out
		}
		if err != nil {
			c.err = err
		}
	}
}

func WithLogger(l *log.Logger) Option {
	return func(c *CLI) {
		c.app.Logger = l
	}
}

func WithAppData(m map[string]any) Option {
	return func(c *CLI) {
		c.app.Data = m
	}
}

func WithHelpCommandName(name string) Option {
	return func(c *CLI) {
		if strings.TrimSpace(name) != "" {
			c.HelpCommandName = name
		}
	}
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

func (c *CLI) Use(m Middleware) {
	if m != nil {
		c.Middleware = append(c.Middleware, m)
	}
}

func (c *CLI) Hook(phase HookPhase, h Hook) {
	if h == nil {
		return
	}

	c.hooks[phase] = append(c.hooks[phase], h)
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
		fmt.Fprintln(w, "Arguments:")
		for _, a := range cmd.Args {
			suffix := ""
			if a.Optional {
				suffix += " (optional)"
			}
			if a.Variadic {
				suffix += " (variadic)"
			}
			fmt.Fprintf(w, "  %s\t%s%s\n", a.Name, a.Description, suffix)
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

	var flagLines []string
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == "h" || f.Name == "help" {
			return
		}
		flagLines = append(flagLines, fmt.Sprintf("  -%s\t%s (default %q)", f.Name, f.Usage, f.DefValue))
	})
	if len(flagLines) > 0 {
		sort.Strings(flagLines)
		fmt.Fprintln(w, "Flags:")
		for _, line := range flagLines {
			fmt.Fprintln(w, line)
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
		sort.Slice(subs, func(i, j int) bool {
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
			fmt.Fprintf(w, "  %s\t%s\n", sub.Name, desc)
		}
		fmt.Fprintln(w)
	}

	return nil
}

func applyMiddleware(h Handler, m []Middleware) Handler {
	if h == nil {
		return nil
	}

	for i := len(m) - 1; i >= 0; i-- {
		if m[i] == nil {
			continue
		}
		h = m[i](h)
	}
	return h
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
