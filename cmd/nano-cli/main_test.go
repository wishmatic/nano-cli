package main

import (
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestHelpDoesNotRequireCredentials(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "help")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	for _, command := range []string{"login", "logout", "serve"} {
		if !strings.Contains(out, command) {
			t.Errorf("help output does not mention %q:\n%s", command, out)
		}
	}
}

func TestHelpForASubcommand(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "help", "login")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !strings.Contains(out, "nano-cli login") {
		t.Errorf("help output = %q, want the login usage", out)
	}
}
