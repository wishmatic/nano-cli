package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestWebScrapeSendsRequestAndPrintsResults(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK, `{
		"results": [
			{"url": "https://a.example", "success": true, "title": "A", "markdown": "# A"},
			{"url": "https://b.example", "success": false, "error": "blocked"}
		],
		"summary": {"stealthModeUsed": true}
	}`)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "scrape", "https://a.example", "https://b.example",
		"--provider", "linkup", "--stealth",
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)

	if request.Path != "/api/scrape-urls" {
		t.Errorf("path = %s, want /api/scrape-urls", request.Path)
	}
	if auth := request.Header.Get("Authorization"); auth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want Bearer test-key", auth)
	}

	body := request.jsonBody(t)
	urls, ok := body["urls"].([]any)
	if !ok || len(urls) != 2 || urls[0] != "https://a.example" || urls[1] != "https://b.example" {
		t.Errorf("urls = %v, want both URLs", body["urls"])
	}
	if body["provider"] != "linkup" {
		t.Errorf("provider = %v, want linkup", body["provider"])
	}
	if body["stealthMode"] != true {
		t.Errorf("stealthMode = %v, want true", body["stealthMode"])
	}

	for _, want := range []string{
		"URL: https://a.example", "Markdown:", "# A",
		"URL: https://b.example", "Error: blocked", "stealthModeUsed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, out)
		}
	}
}

func TestWebScrapeRejectsTooManyURLs(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{},
		"web", "scrape", "a", "b", "c", "d", "e", "f",
	)
	if err == nil || !strings.Contains(err.Error(), "1 to 5") {
		t.Errorf("Execute() error = %v, want a URL-count rejection", err)
	}
}

func TestFirecrawlScrapeSendsRequest(t *testing.T) {
	server, requests := captureServer(
		t, http.StatusOK, `{"success": true, "data": {"markdown": "hi"}}`,
	)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "scrape", "--firecrawl",
		"--format", "markdown", "--format", "html", "--proxy", "stealth",
		"--only-main-content", "--only-clean-content=false",
		"https://example.com",
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
	if body["operation"] != "scrape" {
		t.Errorf("operation = %v, want scrape", body["operation"])
	}
	if body["url"] != "https://example.com" {
		t.Errorf("url = %v, want https://example.com", body["url"])
	}
	if body["proxy"] != "stealth" {
		t.Errorf("proxy = %v, want stealth", body["proxy"])
	}
	if body["onlyMainContent"] != true {
		t.Errorf("onlyMainContent = %v, want true", body["onlyMainContent"])
	}
	if body["onlyCleanContent"] != false {
		t.Errorf("onlyCleanContent = %v, want explicit false", body["onlyCleanContent"])
	}

	formats, ok := body["formats"].([]any)
	if !ok || len(formats) != 2 || formats[0] != "markdown" || formats[1] != "html" {
		t.Errorf("formats = %v, want markdown and html", body["formats"])
	}

	if !strings.Contains(out, `"success": true`) || !strings.Contains(out, `"markdown": "hi"`) {
		t.Errorf("stdout does not preserve the upstream payload:\n%s", out)
	}
}

func TestFirecrawlScrapeOmitsUnsetBooleans(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK, `{}`)

	if _, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"web", "scrape", "--firecrawl", "https://example.com",
	); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	body := nextRequest(t, requests).jsonBody(t)

	if _, ok := body["onlyMainContent"]; ok {
		t.Errorf("onlyMainContent = %v, want it omitted", body["onlyMainContent"])
	}
	if _, ok := body["onlyCleanContent"]; ok {
		t.Errorf("onlyCleanContent = %v, want it omitted", body["onlyCleanContent"])
	}
}

func TestFirecrawlScrapeRequiresExactlyOneURL(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{},
		"web", "scrape", "--firecrawl", "https://a.example", "https://b.example",
	)
	if err == nil || !strings.Contains(err.Error(), "exactly one URL") {
		t.Errorf("Execute() error = %v, want an exactly-one-URL rejection", err)
	}
}
