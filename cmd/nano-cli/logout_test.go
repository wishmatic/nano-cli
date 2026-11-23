package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestLogoutDeletesTheStoredKey(t *testing.T) {
	store := &memoryStore{key: "secret-key"}

	out, _, err := runCommand(t, cli.Deps{Credentials: store}, "logout")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !store.deleted || store.key != "" {
		t.Errorf("store after logout = %+v, want the key deleted", store)
	}

	if !strings.Contains(out, "logged out") {
		t.Errorf("stdout = %q, want a logout confirmation", out)
	}
}

func TestLogoutSucceedsWithoutAStoredKey(t *testing.T) {
	store := &memoryStore{}

	if _, _, err := runCommand(t, cli.Deps{Credentials: store}, "logout"); err != nil {
		t.Errorf("Execute() error: %v", err)
	}
}

func TestLogoutReportsADeleteFailure(t *testing.T) {
	store := &memoryStore{deleteErr: errors.New("delete failed")}

	_, _, err := runCommand(t, cli.Deps{Credentials: store}, "logout")
	if err == nil || !strings.Contains(err.Error(), "delete API key") {
		t.Errorf("Execute() error = %v, want a delete failure", err)
	}
}
