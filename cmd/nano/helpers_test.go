package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
	"github.com/wishmatic/nano-mcp/internal/credential"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
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

func newClientAt(t *testing.T, baseURL string) *nanogpt.Client {
	t.Helper()

	client, err := nanogpt.New(nanogpt.Config{APIKey: "test-key", BaseURL: baseURL})
	if err != nil {
		t.Fatalf("nanogpt.New() error: %v", err)
	}

	return client
}

type capturedRequest struct {
	Method string
	Path   string
	Header http.Header
	Body   []byte
}

func captureServer(
	t *testing.T, status int, response string,
) (*httptest.Server, <-chan capturedRequest) {
	t.Helper()

	requests := make(chan capturedRequest, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			http.Error(w, "read body", http.StatusInternalServerError)

			return
		}

		requests <- capturedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Header: r.Header.Clone(),
			Body:   body,
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	return server, requests
}

func nextRequest(t *testing.T, requests <-chan capturedRequest) capturedRequest {
	t.Helper()

	select {
	case request := <-requests:
		return request
	default:
		t.Fatal("upstream received no request")

		return capturedRequest{}
	}
}

func (r capturedRequest) jsonBody(t *testing.T) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatalf("decode request body %q: %v", r.Body, err)
	}

	return body
}
