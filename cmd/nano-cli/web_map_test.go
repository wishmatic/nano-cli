package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestWebMapSendsRequestAndPrintsResult(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK,
		`{"success": true, "links": ["https://example.com/page"]}`)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "map", "https://example.com", "--sitemap", "only", "--limit", "5",
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)

	if request.Path != "/api/v1/firecrawl" {
		t.Errorf("path = %s, want /api/v1/firecrawl", request.Path)
	}
	if auth := request.Header.Get("Authorization"); auth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want Bearer test-key", auth)
	}

	body := request.jsonBody(t)
	if body["operation"] != "map" {
		t.Errorf("operation = %v, want map", body["operation"])
	}
	if body["url"] != "https://example.com" {
		t.Errorf("url = %v, want https://example.com", body["url"])
	}
	if body["sitemap"] != "only" {
		t.Errorf("sitemap = %v, want only", body["sitemap"])
	}
	if body["limit"] != float64(5) {
		t.Errorf("limit = %v, want 5", body["limit"])
	}

	if !strings.Contains(out, `"links"`) || !strings.Contains(out, "https://example.com/page") {
		t.Errorf("stdout does not preserve the upstream payload:\n%s", out)
	}
}
