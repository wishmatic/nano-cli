package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestGenerateVideoWritesTheDownloadedFile(t *testing.T) {
	t.Setenv("IMAGE_URL_MAP", "")

	var serverURL string
	server, requests := captureServerFunc(t, func(r *http.Request) (int, string) {
		switch r.URL.Path {
		case "/input.png":
			return http.StatusOK, string(testPNG)
		case "/api/generate-video":
			return http.StatusOK, `{"runId":"run-1","status":"queued","model":"veo2-video"}`
		case "/api/video/status":
			return http.StatusOK, fmt.Sprintf(
				`{"status":"completed","videoUrl":%q,"data":{"cost":0.25}}`,
				serverURL+"/video.mp4",
			)
		case "/video.mp4":
			return http.StatusOK, string(testVideo)
		default:
			return http.StatusNotFound, ""
		}
	})
	serverURL = server.URL

	dir := t.TempDir()

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"gen", "video", "a wave",
		"--model", "veo2-video",
		"--image", server.URL+"/input.png",
		"--negative-prompt", "blur",
		"--duration", "5",
		"--aspect-ratio", "16:9",
		"--resolution", "720p",
		"--seed", "7",
		"--generate-audio",
		"--out", dir,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	submit := nextRequestPath(t, requests, "/api/generate-video")
	if submit.Method != http.MethodPost {
		t.Errorf("submit method = %s, want POST", submit.Method)
	}

	body := submit.jsonBody(t)
	checks := map[string]any{
		"model":           "veo2-video",
		"prompt":          "a wave",
		"negative_prompt": "blur",
		"duration":        "5",
		"aspect_ratio":    "16:9",
		"resolution":      "720p",
		"mode":            "image-to-video",
		"seed":            float64(7),
		"generateAudio":   true,
	}
	for field, want := range checks {
		if body[field] != want {
			t.Errorf("%s = %v, want %v", field, body[field], want)
		}
	}

	image, ok := body["imageDataUrl"].(string)
	if !ok || !strings.HasPrefix(image, "data:image/png;base64,") {
		t.Errorf("imageDataUrl = %v, want the resolved image", body["imageDataUrl"])
	}

	status := nextRequestPath(t, requests, "/api/video/status")
	if status.Query != "requestId=run-1" {
		t.Errorf("status query = %q, want requestId=run-1", status.Query)
	}

	files, err := filepath.Glob(filepath.Join(dir, "nano-video-*.mp4"))
	if err != nil {
		t.Fatalf("Glob() error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("files = %v, want one video", files)
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != string(testVideo) {
		t.Errorf("video content = %q, want the downloaded bytes", data)
	}

	if !strings.Contains(out, dir) || !strings.Contains(out, "Cost: $0.25") {
		t.Errorf("stdout does not contain the path and cost:\n%s", out)
	}
}

func TestGenerateVideoLeavesNoFileWhenTheDownloadFails(t *testing.T) {
	t.Setenv("IMAGE_URL_MAP", "")

	var serverURL string
	server, _ := captureServerFunc(t, func(r *http.Request) (int, string) {
		switch r.URL.Path {
		case "/api/generate-video":
			return http.StatusOK, `{"runId":"run-1","status":"queued","model":"veo2-video"}`
		case "/api/video/status":
			return http.StatusOK, fmt.Sprintf(
				`{"status":"completed","videoUrl":%q,"data":{"cost":0.25}}`,
				serverURL+"/video.mp4",
			)
		case "/video.mp4":
			return http.StatusInternalServerError, "boom"
		default:
			return http.StatusNotFound, ""
		}
	})
	serverURL = server.URL

	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)},
		"gen", "video", "a wave", "--model", "veo2-video", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "gen video") {
		t.Fatalf("Execute() error = %v, want a download failure", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("output directory contains %d files, want none", len(entries))
	}
}

func TestGenerateVideoRequiresAModel(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{}, "gen", "video", "a wave")
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Errorf("Execute() error = %v, want the missing-model rejection", err)
	}
}
