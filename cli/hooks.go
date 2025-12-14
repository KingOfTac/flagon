package cli

import "context"

type Hook func(ctx context.Context) error

type HookPhase int

const (
	BeforeRun HookPhase = iota
	AfterRun
	BeforeCommand
	AfterCommand
)

func (c *CLI) Hook(phase HookPhase, h Hook) {
	if h == nil {
		return
	}

	c.hooks[phase] = append(c.hooks[phase], h)
}
