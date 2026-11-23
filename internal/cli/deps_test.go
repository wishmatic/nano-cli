package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/credential"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
)

type memoryStore struct {
	key     string
	loadErr error
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
	s.key = strings.TrimSpace(key)

	return nil
}

func (s *memoryStore) Delete() error {
	s.key = ""

	return nil
}

func TestOpenCredentialsUsesTheInjectedStore(t *testing.T) {
	store := &memoryStore{}

	got, err := OpenCredentials(Deps{Credentials: store})
	if err != nil {
		t.Fatalf("OpenCredentials() error: %v", err)
	}

	if got != store {
		t.Error("OpenCredentials() did not return the injected store")
	}
}

func newTestClient(t *testing.T) *nanogpt.Client {
	t.Helper()

	client, err := nanogpt.New(nanogpt.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("nanogpt.New() error: %v", err)
	}

	return client
}

func TestNanoClientUsesAnInjectedClient(t *testing.T) {
	client := newTestClient(t)

	got, err := NanoClient(Deps{Client: client})
	if err != nil {
		t.Fatalf("NanoClient() error: %v", err)
	}

	if got != client {
		t.Error("NanoClient() did not return the injected client")
	}
}

func TestNanoClientPrefersTheEnvironment(t *testing.T) {
	t.Setenv("NANOGPT_API_KEY", "environment-key")

	deps := Deps{Credentials: &memoryStore{loadErr: errors.New("should not be read")}}

	client, err := NanoClient(deps)
	if err != nil {
		t.Fatalf("NanoClient() error: %v", err)
	}

	if client == nil {
		t.Error("NanoClient() = nil, want a client")
	}
}

func TestNanoClientUsesTheStoredKey(t *testing.T) {
	t.Setenv("NANOGPT_API_KEY", "")

	client, err := NanoClient(Deps{Credentials: &memoryStore{key: "stored-key"}})
	if err != nil {
		t.Fatalf("NanoClient() error: %v", err)
	}

	if client == nil {
		t.Error("NanoClient() = nil, want a client")
	}
}

func TestNanoClientReportsAMissingKey(t *testing.T) {
	t.Setenv("NANOGPT_API_KEY", "")

	_, err := NanoClient(Deps{Credentials: &memoryStore{}})
	if err == nil || !strings.Contains(err.Error(), "nano-cli login") {
		t.Errorf("NanoClient() error = %v, want a login hint", err)
	}
}
