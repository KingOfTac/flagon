package cli

import (
	"bytes"
	"context"
	"flag"
	"log"
	"testing"
)

func FuzzCLIRun(f *testing.F) {
	f.Add("")
	f.Add("help")
	f.Add("--help")
	f.Add("-h")
	f.Add("nonexistent")
	f.Add("cmd --flag=value")
	f.Add("cmd arg1 arg2 arg3")
	f.Add("cmd --flag value --other")
	f.Add("a b c d e f g h i j k l m n o p")
	f.Add("--")
	f.Add("-- --help")
	f.Add("-")
	f.Add("---")
	f.Add("cmd\x00with\x00nulls")
	f.Add("cmd\twith\ttabs")
	f.Add("cmd\nwith\nnewlines")
	f.Add("'quoted arg'")
	f.Add("\"double quoted\"")
	f.Add("cmd --flag='value with spaces'")
	f.Add("cmd -f=")
	f.Add("cmd --=value")
	f.Add("=value")
	f.Add("--=")
	f.Add(string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, input string) {
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		root := &Command{
			Name:        "test",
			Description: "Test command",
			Handler: func(ctx context.Context) error {
				return nil
			},
			Commands: []*Command{
				{
					Name:        "sub",
					Description: "Subcommand",
					Args: []Arg{
						{Name: "arg1", Description: "First arg"},
						{Name: "arg2", Description: "Second arg", Optional: true},
						{Name: "rest", Description: "Rest args", Variadic: true},
					},
					Flags: func(fs *flag.FlagSet) {
						fs.String("name", "", "name flag")
						fs.Int("count", 0, "count flag")
						fs.Bool("verbose", false, "verbose flag")
					},
					Handler: func(ctx context.Context) error {
						_ = Args(ctx)
						_ = Flags(ctx)
						_ = CurrentCommand(ctx)
						return nil
					},
				},
			},
		}

		cli := New(root,
			WithWriters(out, errOut),
			WithLogger(log.New(out, "", 0)),
		)

		args := splitArgs(input)

		_ = cli.Run(args)
	})
}

func FuzzValidatePositionalArgs(f *testing.F) {
	f.Add("", 0, false, false)
	f.Add("arg1", 1, false, false)
	f.Add("arg1 arg2 arg3", 2, true, false)
	f.Add("a b c d e f g h i j", 5, false, true)

	f.Fuzz(func(t *testing.T, argsStr string, numRequired int, hasOptional, hasVariadic bool) {
		if numRequired < 0 || numRequired > 20 {
			return
		}

		var cmdArgs []Arg
		for i := 0; i < numRequired; i++ {
			cmdArgs = append(cmdArgs, Arg{
				Name:     "required" + string(rune('0'+i)),
				Optional: false,
			})
		}
		if hasOptional {
			cmdArgs = append(cmdArgs, Arg{
				Name:     "optional",
				Optional: true,
			})
		}
		if hasVariadic {
			cmdArgs = append(cmdArgs, Arg{
				Name:     "variadic",
				Variadic: true,
			})
		}

		cmd := &Command{
			Name: "test",
			Args: cmdArgs,
		}

		parsed := splitArgs(argsStr)

		_ = validatePositionalArgs(cmd, parsed)
	})
}

func FuzzFindSubcommand(f *testing.F) {
	f.Add("sub")
	f.Add("alias1")
	f.Add("")
	f.Add("nonexistent")
	f.Add("\x00")
	f.Add("sub\x00extra")
	f.Add(string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, token string) {
		cmd := &Command{
			Name: "root",
			Commands: []*Command{
				{Name: "sub", Aliases: []string{"alias1", "a"}},
				{Name: "hidden", Hidden: true},
				{Name: "other"},
				nil,
			},
		}

		_ = findSubcommand(cmd, token)
		_ = findSubcommand(nil, token)
	})
}

func FuzzCollides(f *testing.F) {
	f.Add("test")
	f.Add("sub")
	f.Add("alias")
	f.Add("")
	f.Add("\x00")
	f.Add("help")

	f.Fuzz(func(t *testing.T, nameOrAlias string) {
		parent := &Command{
			Name: "root",
			Commands: []*Command{
				{Name: "sub", Aliases: []string{"alias1", "a"}},
				{Name: "other"},
				nil,
			},
		}

		_ = collides(parent, nameOrAlias)
		_ = collides(nil, nameOrAlias)
	})
}

func FuzzContextFunctions(f *testing.F) {
	f.Add(true, true, true)
	f.Add(false, false, false)
	f.Add(true, false, true)

	f.Fuzz(func(t *testing.T, hasApp, hasCmd, hasArgs bool) {
		ctx := context.Background()

		if hasApp {
			ctx = context.WithValue(ctx, appKey, &App{
				Logger: log.New(&bytes.Buffer{}, "", 0),
				Data:   map[string]any{"key": "value"},
			})
		}

		if hasCmd {
			ctx = context.WithValue(ctx, commandKey, &Command{Name: "test"})
		}

		if hasArgs {
			ctx = context.WithValue(ctx, argsKey, []string{"arg1", "arg2"})
		}

		_ = AppFromContext(ctx)
		_ = CurrentCommand(ctx)
		_ = Args(ctx)
		_ = Flags(ctx)

		_ = AppFromContext(nil)
		_ = CurrentCommand(nil)
		_ = Args(nil)
		_ = Flags(nil)
	})
}

func FuzzMiddlewareChain(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(5)
	f.Add(10)
	f.Add(100)

	f.Fuzz(func(t *testing.T, numMiddleware int) {
		if numMiddleware < 0 || numMiddleware > 1000 {
			return
		}

		var middleware []Middleware
		for i := 0; i < numMiddleware; i++ {
			middleware = append(middleware, func(next Handler) Handler {
				return func(ctx context.Context) error {
					return next(ctx)
				}
			})
		}

		handler := func(ctx context.Context) error {
			return nil
		}

		final := applyMiddleware(handler, middleware)
		if final != nil {
			_ = final(context.Background())
		}
	})
}

func splitArgs(s string) []string {
	if s == "" {
		return nil
	}

	var args []string
	var current []byte
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' || c == '\'':
			if !inQuote {
				inQuote = true
				quoteChar = c
			} else if c == quoteChar {
				inQuote = false
				quoteChar = 0
			} else {
				current = append(current, c)
			}
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			if inQuote {
				current = append(current, c)
			} else if len(current) > 0 {
				args = append(args, string(current))
				current = nil
			}
		default:
			current = append(current, c)
		}
	}

	if len(current) > 0 {
		args = append(args, string(current))
	}

	return args
}
