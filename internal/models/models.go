package models

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

type App struct {
	Logger *log.Logger
}

type appKeyType struct{}
type commandKeyType struct{}
type argsKeyType struct{}

var appKey = appKeyType{}
var commandKey = commandKeyType{}
var argsKey = argsKeyType{}

type Handler func(ctx context.Context) error

type Arg struct {
	Name        string
	Description string
	Optional    bool
}

type Command struct {
	Name        string
	Description string

	Args     []Arg
	Flags    func(fs *flag.FlagSet)
	Handler  Handler
	Commands []*Command
}

type CLI struct {
	app  *App
	ctx  context.Context
	Root *Command
}

func NewCLI(cliName string, root *Command) *CLI {
	app := &App{
		Logger: log.New(os.Stdout, fmt.Sprintf("[%s] ", cliName), log.LstdFlags),
	}

	ctx := context.WithValue(context.Background(), appKey, app)

	return &CLI{
		app:  app,
		ctx:  ctx,
		Root: root,
	}
}

func (c *CLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.printHelp(c.Root)
	}
	return c.execute(ctx, c.Root, args)
}

func (c *CLI) execute(ctx context.Context, cmd *Command, args []string) error {
	if len(args) > 0 {
		for _, sub := range cmd.Commands {
			if sub.Name == args[0] {
				return c.execute(ctx, sub, args[1:])
			}
		}
	}

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	if cmd.Flags != nil {
		cmd.Flags(fs)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	parsedArgs := fs.Args()

	if len(cmd.Args) > 0 {
		required := 0
		for _, a := range cmd.Args {
			if !a.Optional {
				required++
			}
		}

		if len(parsedArgs) < required {
			return fmt.Errorf("missing required arguments")
		}
	}

	ctx = context.WithValue(ctx, argsKey, parsedArgs)
	ctx = context.WithValue(ctx, commandKey, cmd)

	if cmd.Handler == nil {
		return c.printHelp(cmd)
	}

	return cmd.Handler(ctx)
}

func (c *CLI) printHelp(cmd *Command) error {
	fmt.Println(cmd.Name)
	fmt.Println(cmd.Description)
	fmt.Println()

	if len(cmd.Args) > 0 {
		fmt.Println("Arguments:")
		for _, arg := range cmd.Args {
			opt := ""
			if arg.Optional {
				opt = " (optional)"
			}
			fmt.Printf("  %s\t%s%s\n", arg.Name, arg.Description, opt)
		}
		fmt.Println()
	}

	if len(cmd.Commands) > 0 {
		fmt.Println("Commands:")
		for _, sub := range cmd.Commands {
			fmt.Printf("  %s\t%s\n", sub.Name, sub.Description)
		}
	}

	return nil
}

func AppFromContext(ctx context.Context) *App {
	return ctx.Value(appKey).(*App)
}

func CurrentCommand(ctx context.Context) *Command {
	return ctx.Value(commandKey).(*Command)
}

func Args(ctx context.Context) []string {
	return ctx.Value(argsKey).([]string)
}
