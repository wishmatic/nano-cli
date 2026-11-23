package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestWebYouTubeSendsRequestAndPrintsResult(t *testing.T) {
	for _, name := range []string{"youtube", "yt"} {
		t.Run(name, func(t *testing.T) {
			server, requests := captureServer(t, http.StatusOK, `{
				"transcripts": [
					{"url": "https://youtu.be/a", "success": true, "title": "A", "transcript": "hello"},
					{"url": "https://youtu.be/b", "success": false, "error": "no captions"}
				],
				"summary": {"requested": 2, "processed": 2, "successful": 1, "failed": 1, "totalCost": 0.01}
			}`)

			out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
				"web", name, "https://youtu.be/a", "https://youtu.be/b",
			)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			request := nextRequest(t, requests)

			if request.Path != "/api/youtube-transcribe" {
				t.Errorf("path = %s, want /api/youtube-transcribe", request.Path)
			}
			if key := request.Header.Get("x-api-key"); key != "test-key" {
				t.Errorf("x-api-key = %q, want test-key", key)
			}

			body := request.jsonBody(t)
			urls, ok := body["urls"].([]any)
			if !ok || len(urls) != 2 || urls[0] != "https://youtu.be/a" || urls[1] != "https://youtu.be/b" {
				t.Errorf("urls = %v, want both URLs", body["urls"])
			}

			for _, want := range []string{
				"URL: https://youtu.be/a", "Transcript:", "hello",
				"URL: https://youtu.be/b", "Error: no captions",
				"failed=1", "cost=$0.01",
			} {
				if !strings.Contains(out, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestWebYouTubeRejectsTooManyURLs(t *testing.T) {
	urls := make([]string, 11)
	for index := range urls {
		urls[index] = fmt.Sprintf("https://youtu.be/%d", index)
	}

	args := append([]string{"web", "youtube"}, urls...)

	_, _, err := runCommand(t, cli.Deps{}, args...)
	if err == nil || !strings.Contains(err.Error(), "1 to 10") {
		t.Errorf("Execute() error = %v, want a URL-count rejection", err)
	}
}
