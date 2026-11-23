package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/credential"
)

type memoryStore struct {
	key       string
	loadErr   error
	saveErr   error
	deleteErr error
	deleted   bool
}

func (s *memoryStore) Load() (string, error) {
	if s.loadErr != nil {
		return "", s.loadErr
	}
	if s.key == "" {
		return "", credential.ErrNotFound
	}

	return s.key, nil
}

func (s *memoryStore) Save(key string) error {
	if s.saveErr != nil {
		return s.saveErr
	}

	s.key = strings.TrimSpace(key)

	return nil
}

func (s *memoryStore) Delete() error {
	s.deleted = true

	if s.deleteErr != nil {
		return s.deleteErr
	}

	s.key = ""

	return nil
}

func runCommand(t *testing.T, deps cli.Deps, args ...string) (string, string, error) {
	t.Helper()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	deps.Out = out
	deps.Err = errOut

	command := newRoot(deps)
	command.SetArgs(args)

	err := command.Execute()

	return out.String(), errOut.String(), err
}
