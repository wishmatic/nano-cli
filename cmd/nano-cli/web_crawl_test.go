package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestWebCrawlSendsAllSettings(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK,
		`{"status": "completed", "data": [{"url": "https://example.com/page"}]}`)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "crawl", "https://example.com",
		"--format", "markdown", "--format", "html",
		"--proxy", "stealth",
		"--only-main-content", "--only-clean-content=false",
		"--sitemap", "include",
		"--limit", "10",
		"--crawl-entire-domain", "--allow-subdomains",
		"--max-discovery-depth", "3",
		"--wait-for-finish", "60",
		"--max-returned-pages", "5",
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)

	if request.Path != "/api/v1/firecrawl" {
		t.Errorf("path = %s, want /api/v1/firecrawl", request.Path)
	}

	body := request.jsonBody(t)
	checks := map[string]any{
		"operation":         "crawl",
		"url":               "https://example.com",
		"proxy":             "stealth",
		"onlyMainContent":   true,
		"onlyCleanContent":  false,
		"sitemap":           "include",
		"limit":             float64(10),
		"crawlEntireDomain": true,
		"allowSubdomains":   true,
		"maxDiscoveryDepth": float64(3),
		"waitForFinishSecs": float64(60),
		"maxReturnedPages":  float64(5),
	}

	for field, want := range checks {
		if body[field] != want {
			t.Errorf("%s = %v, want %v", field, body[field], want)
		}
	}

	formats, ok := body["formats"].([]any)
	if !ok || len(formats) != 2 || formats[0] != "markdown" || formats[1] != "html" {
		t.Errorf("formats = %v, want markdown and html", body["formats"])
	}

	if !strings.Contains(out, `"status": "completed"`) ||
		!strings.Contains(out, "https://example.com/page") {
		t.Errorf("stdout does not preserve the upstream payload:\n%s", out)
	}
}
