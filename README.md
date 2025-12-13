<p align="center">
  <a href="https://github.com/kingoftac/gork">
    <picture>
      <source srcset=".github/media/flagon_dark.png" media="(prefers-color-scheme: dark)">
      <source srcset=".github/media/flagon_light.png" media="(prefers-color-scheme: light)">
      <img src=".github/media/flagon_light.png" alt="flagon logo">
    </picture>
  </a>
</p>
<p align="center">A zero dependency, declarative CLI framework for GO</p>

---

# Installation

```bash
go get github.com/kingoftac/flagon
```

# Usage
```go
func main() {
	cli := NewCLI(&Command{
		Name:        "app",
		Description: "Example declarative CLI",
		Commands: []*Command{
			{
				Name:        "build",
				Description: "Build the project",
				Args: []Arg{
					{Name: "path", Description: "Path to project"},
					{Name: "target", Description: "Build target", Optional: true},
				},
				Flags: func(fs *flag.FlagSet) {
					fs.Bool("release", false, "enable release mode")
				},
				Handler: func(ctx context.Context) error {
					app := AppFromContext(ctx)
					cmd := CurrentCommand(ctx)
					args := Args(ctx)

					app.Logger.Println("command:", cmd.Name)
					app.Logger.Println("args:", args)
					return nil
				},
			},
		},
	})

	if err := cli.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
```