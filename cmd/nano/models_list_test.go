package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

const imageModelsBody = `{
	"data": [
		{
			"id": "nano-banana", "name": "Nano Banana", "owned_by": "google",
			"description": "Fast image editing.", "tags": ["editing"]
		},
		{
			"id": "hidream", "name": "HiDream", "owned_by": "hidream",
			"description": "Illustration and anime.", "tags": ["anime"]
		},
		{
			"id": "p-image", "name": "P-Image", "description": "Cheap drafts.",
			"tags": ["text-to-image"]
		}
	]
}`

func TestModelsListDefaultsToBothKinds(t *testing.T) {
	imageBody := `{"data": [{"id": "nano-banana", "name": "Nano Banana", "owned_by": "google"}]}`
	videoBody := `{"data": [{"id": "veo2-video", "name": "Veo 2", "owned_by": "google"}]}`

	server, requests := captureServerFunc(t, func(r *http.Request) (int, string) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/image-models") {
			return http.StatusOK, imageBody
		}

		return http.StatusOK, videoBody
	})

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)}, "models", "list")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	imageRequest := nextRequest(t, requests)
	if imageRequest.Path != "/api/v1/image-models" || imageRequest.Query != "detailed=true" {
		t.Errorf(
			"image request = %s?%s, want the detailed image list",
			imageRequest.Path, imageRequest.Query,
		)
	}
	if imageRequest.Method != http.MethodGet {
		t.Errorf("image method = %s, want GET", imageRequest.Method)
	}
	if auth := imageRequest.Header.Get("Authorization"); auth != "Bearer test-key" {
		t.Errorf("image Authorization = %q, want Bearer test-key", auth)
	}

	videoRequest := nextRequest(t, requests)
	if videoRequest.Path != "/api/v1/video-models" || videoRequest.Query != "detailed=true" {
		t.Errorf(
			"video request = %s?%s, want the detailed video list",
			videoRequest.Path, videoRequest.Query,
		)
	}

	for _, want := range []string{"IMAGE MODELS", "nano-banana", "VIDEO MODELS", "veo2-video"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, out)
		}
	}
}

func TestModelsListFiltersAndLimits(t *testing.T) {
	t.Run("query matches tags case-insensitively", func(t *testing.T) {
		out := runImageList(t, "--query", "ANIME")

		for _, want := range []string{"hidream", "1 of 3 models shown (1 matched)"} {
			if !strings.Contains(out, want) {
				t.Errorf("stdout does not contain %q:\n%s", want, out)
			}
		}

		for _, unwanted := range []string{"nano-banana", "p-image"} {
			if strings.Contains(out, unwanted) {
				t.Errorf("stdout contains filtered model %q:\n%s", unwanted, out)
			}
		}
	})

	t.Run("limit caps the shown models", func(t *testing.T) {
		out := runImageList(t, "--limit", "2")

		for _, want := range []string{"nano-banana", "hidream", "2 of 3 models shown (3 matched)"} {
			if !strings.Contains(out, want) {
				t.Errorf("stdout does not contain %q:\n%s", want, out)
			}
		}

		if strings.Contains(out, "p-image") {
			t.Errorf("stdout contains a model past the limit:\n%s", out)
		}
	})

	t.Run("zero uses the default limit", func(t *testing.T) {
		out := runImageList(t, "--limit", "0")

		for _, want := range []string{"nano-banana", "hidream", "p-image", "3 of 3 models shown"} {
			if !strings.Contains(out, want) {
				t.Errorf("stdout does not contain %q:\n%s", want, out)
			}
		}
	})
}

func runImageList(t *testing.T, args ...string) string {
	t.Helper()

	server, _ := captureServer(t, http.StatusOK, imageModelsBody)
	commandArgs := append([]string{"models", "list", "image"}, args...)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)}, commandArgs...)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	return out
}

func TestModelsListRejectsAnInvalidKind(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{}, "models", "list", "audio")
	if err == nil || !strings.Contains(err.Error(), "not one of") {
		t.Errorf("Execute() error = %v, want an invalid-kind rejection", err)
	}
}

func TestModelsListRejectsALimitOutsideTheCap(t *testing.T) {
	for _, limit := range []string{"-1", "101"} {
		_, _, err := runCommand(t, cli.Deps{}, "models", "list", "image", "--limit", limit)
		if err == nil || !strings.Contains(err.Error(), "1 to 100") {
			t.Errorf("Execute() error for limit %s = %v, want the accepted range", limit, err)
		}
	}
}

func TestModelsListReportsAnUpstreamFailure(t *testing.T) {
	server, _ := captureServer(t, http.StatusUnauthorized, `{"error":"bad key"}`)

	_, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)}, "models", "list", "image")
	if err == nil ||
		!strings.Contains(err.Error(), "models list image") ||
		!strings.Contains(err.Error(), "401") {
		t.Errorf("Execute() error = %v, want the upstream failure", err)
	}
}
