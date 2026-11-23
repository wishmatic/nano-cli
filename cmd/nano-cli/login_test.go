package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/credential"
)

func TestLoginStoresTrimmedStdin(t *testing.T) {
	store := &memoryStore{}

	out, errOut, err := runCommand(t, cli.Deps{
		In:          strings.NewReader("  secret-key\n"),
		Credentials: store,
	}, "login")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if store.key != "secret-key" {
		t.Errorf("stored key = %q, want secret-key", store.key)
	}

	if !strings.Contains(out, "logged in") {
		t.Errorf("stdout = %q, want a login confirmation", out)
	}

	if strings.Contains(out+errOut, "secret-key") {
		t.Errorf("output leaked the API key: stdout %q, stderr %q", out, errOut)
	}
}

func TestLoginRejectsMultipleLinesOfPipedInput(t *testing.T) {
	store := &memoryStore{}

	_, _, err := runCommand(t, cli.Deps{
		In:          strings.NewReader("key-from-pipe\nignored\n"),
		Credentials: store,
	}, "login")
	if err == nil || !strings.Contains(err.Error(), "single line") {
		t.Fatalf("Execute() error = %v, want a single-line rejection", err)
	}

	if store.key != "" {
		t.Errorf("stored key = %q, want nothing", store.key)
	}

	if _, _, err := runCommand(t, cli.Deps{
		In:          strings.NewReader("  one-line-key  \n"),
		Credentials: store,
	}, "login"); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if store.key != "one-line-key" {
		t.Errorf("stored key = %q, want one-line-key", store.key)
	}
}

func TestLoginRejectsAnEmptyKey(t *testing.T) {
	store := &memoryStore{}

	_, _, err := runCommand(t, cli.Deps{
		In:          strings.NewReader(" \n\t "),
		Credentials: store,
	}, "login")
	if !errors.Is(err, credential.ErrEmptyKey) {
		t.Fatalf("Execute() error = %v, want ErrEmptyKey", err)
	}

	if store.key != "" {
		t.Errorf("stored key = %q, want nothing", store.key)
	}
}

func TestLoginReportsAStoreFailure(t *testing.T) {
	store := &memoryStore{saveErr: errors.New("keyring exploded")}

	_, _, err := runCommand(t, cli.Deps{
		In:          strings.NewReader("secret-key\n"),
		Credentials: store,
	}, "login")
	if err == nil || !strings.Contains(err.Error(), "store API key") {
		t.Errorf("Execute() error = %v, want a store failure", err)
	}
}
