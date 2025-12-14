package cli

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	cmd := &Command{Name: "test"}
	cli := New(cmd)
	if cli == nil {
		t.Fatal("New returned nil")
	}
	if cli.Root.Name != "test" {
		t.Errorf("Expected root name 'test', got %s", cli.Root.Name)
	}
}

func TestArgs(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, argsKey, []string{"arg1", "arg2"})
	args := Args(ctx)
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
	if args[0] != "arg1" || args[1] != "arg2" {
		t.Errorf("Args mismatch: %v", args)
	}
}

func TestCurrentCommand(t *testing.T) {
	ctx := context.Background()
	cmd := &Command{Name: "test"}
	ctx = context.WithValue(ctx, commandKey, cmd)
	current := CurrentCommand(ctx)
	if current == nil {
		t.Fatal("CurrentCommand returned nil")
	}
	if current.Name != "test" {
		t.Errorf("Expected command name 'test', got %s", current.Name)
	}
}

func TestAppFromContext(t *testing.T) {
	ctx := context.Background()
	app := &App{}
	ctx = context.WithValue(ctx, appKey, app)
	got := AppFromContext(ctx)
	if got != app {
		t.Errorf("AppFromContext returned wrong app")
	}
}

func TestValidatePositionalArgs(t *testing.T) {
	cmd := &Command{
		Args: []Arg{
			{Name: "required"},
			{Name: "optional", Optional: true},
		},
	}

	// Valid: one arg for required
	err := validatePositionalArgs(cmd, []string{"val"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Invalid: no args
	err = validatePositionalArgs(cmd, []string{})
	if err == nil {
		t.Error("Expected error for missing required arg")
	}

	// Invalid: too many args
	err = validatePositionalArgs(cmd, []string{"a", "b", "c"})
	if err == nil {
		t.Error("Expected error for too many args")
	}
}

func TestFindSubcommand(t *testing.T) {
	cmd := &Command{
		Commands: []*Command{
			{Name: "sub1"},
			{Name: "sub2", Aliases: []string{"s2"}},
		},
	}

	found := findSubcommand(cmd, "sub1")
	if found == nil || found.Name != "sub1" {
		t.Error("Failed to find sub1")
	}

	found = findSubcommand(cmd, "s2")
	if found == nil || found.Name != "sub2" {
		t.Error("Failed to find alias s2")
	}

	found = findSubcommand(cmd, "nonexistent")
	if found != nil {
		t.Error("Should not find nonexistent command")
	}
}

func TestCollides(t *testing.T) {
	parent := &Command{
		Commands: []*Command{
			{Name: "existing"},
		},
	}

	if !collides(parent, "existing") {
		t.Error("Should detect collision with existing name")
	}

	if collides(parent, "new") {
		t.Error("Should not detect collision with new name")
	}
}
