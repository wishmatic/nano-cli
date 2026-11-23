package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestWebSearchSendsRequestAndPrintsResult(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK, `{
		"data": [{"title": "Example"}],
		"metadata": {
			"query": "hello world",
			"provider": "exa",
			"operation": "search",
			"depth": "deep",
			"outputType": "sourcedAnswer",
			"timestamp": "2025-01-01T00:00:00Z",
			"cost": 0.0021
		}
	}`)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "search", "hello", "world",
		"--provider", "exa", "--depth", "deep", "--output-type", "sourcedAnswer",
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)

	if request.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", request.Method)
	}
	if request.Path != "/api/web" {
		t.Errorf("path = %s, want /api/web", request.Path)
	}
	if key := request.Header.Get("x-api-key"); key != "test-key" {
		t.Errorf("x-api-key = %q, want test-key", key)
	}

	body := request.jsonBody(t)
	if body["query"] != "hello world" {
		t.Errorf("query = %v, want hello world", body["query"])
	}
	if body["provider"] != "exa" {
		t.Errorf("provider = %v, want exa", body["provider"])
	}
	if body["depth"] != "deep" {
		t.Errorf("depth = %v, want deep", body["depth"])
	}
	if body["outputType"] != "sourcedAnswer" {
		t.Errorf("outputType = %v, want sourcedAnswer", body["outputType"])
	}

	for _, want := range []string{"Provider: exa", "Cost: $0.0021", "Example"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, out)
		}
	}
}

func TestWebSearchReportsAnUpstreamError(t *testing.T) {
	server, _ := captureServer(t, http.StatusInternalServerError, "boom")

	_, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "search", "hello",
	)
	if err == nil ||
		!strings.Contains(err.Error(), "web search") ||
		!strings.Contains(err.Error(), "boom") {
		t.Errorf("Execute() error = %v, want the upstream failure", err)
	}
}

func TestWebSearchReportsMalformedJSON(t *testing.T) {
	server, _ := captureServer(t, http.StatusOK, `{`)

	_, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "search", "hello",
	)
	if err == nil || !strings.Contains(err.Error(), "decode /api/web") {
		t.Errorf("Execute() error = %v, want a decode failure", err)
	}
}

func TestWebSearchRejectsAnEmptyQuery(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{}, "web", "search", "   ")
	if err == nil || !strings.Contains(err.Error(), "search query is empty") {
		t.Errorf("Execute() error = %v, want an empty-query rejection", err)
	}
}
