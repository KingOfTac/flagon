package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/kingoftac/flagon/cli"
	lua "github.com/kingoftac/flagon/cli-lua"
)

func main() {
	c := cli.New(&cli.Command{
		Name:        "app",
		Description: "Example App",
		Commands: []*cli.Command{
			{
				Name:        "foo",
				Summary:     "foo n' stuff",
				Description: "Does foo n' stuff",
				Args: []cli.Arg{
					{Name: "bar", Description: "bar thing"},
				},
				Flags: func(fs *flag.FlagSet) {
					fs.Bool("baz", false, "enable baz mode")
				},
				Handler: func(ctx context.Context) error {
					a := cli.AppFromContext(ctx)
					a.Logger.Println("args:", cli.Args(ctx))
					a.Logger.Println("flags:", cli.Flags(ctx))
					return nil
				},
			},
		},
	}, cli.WithLogger(log.New(os.Stdout, "[App] ", log.LstdFlags)))

	c.Use(TimingMiddleware(c.App().Logger))
	c.Hook(cli.BeforeRun, func(ctx context.Context) error {
		c.App().Logger.Println("Starting up...")
		return nil
	})

	// Lua plugin integration example
	engine := lua.NewEngine(c)
	if err := engine.LoadFile("./cmd/plugin.lua"); err != nil {
		log.Fatal(err)
	}

	if err := c.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func TimingMiddleware(logger *log.Logger) cli.Middleware {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	return func(next cli.Handler) cli.Handler {
		return func(ctx context.Context) error {
			start := time.Now()
			err := next(ctx)
			logger.Printf("took=%s cmd=%s err=%v", time.Since(start), cli.CurrentCommand(ctx).Name, err)
			return err
		}
	}
}
