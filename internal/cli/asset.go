package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxAssetBytes = 256 << 20

	assetFetchTimeout = 3 * time.Minute
)

var assetExtensions = map[string]string{
	"video/mp4":       "mp4",
	"video/webm":      "webm",
	"video/quicktime": "mov",

	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
}

func WriteAsset(dir, prefix string, data []byte, mediaType string) (string, error) {
	ext, ok := assetExtensions[strings.ToLower(strings.TrimSpace(mediaType))]
	if !ok {
		return "", fmt.Errorf("cannot store media type %q", mediaType)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	stamp := time.Now().UTC().Format("20060102-150405")

	tmp, err := os.CreateTemp(dir, prefix+"-"+stamp+"-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}

	name := tmp.Name()
	defer func() {
		_ = os.Remove(name)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()

		return "", fmt.Errorf("write output file: %w", err)
	}

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()

		return "", fmt.Errorf("set output permissions: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close output file: %w", err)
	}

	final := filepath.Join(dir, strings.TrimSuffix(filepath.Base(name), ".tmp")+"."+ext)

	if err := os.Rename(name, final); err != nil {
		return "", fmt.Errorf("store output file: %w", err)
	}

	name = ""

	return final, nil
}

func DownloadAsset(ctx context.Context, source, fallbackMediaType string) ([]byte, string, error) {
	return downloadAsset(ctx, source, fallbackMediaType, MaxAssetBytes)
}

func downloadAsset(
	ctx context.Context, source, fallbackMediaType string, limit int64,
) ([]byte, string, error) {
	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", fmt.Errorf("%q is not a URL: %w", source, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, "", fmt.Errorf("%q is not an http URL", source)
	}

	ctx, cancel := context.WithTimeout(ctx, assetFetchTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build request for %s: %w", source, err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("download %s: %w", source, err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download %s: %s", source, response.Status)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, "", fmt.Errorf("download %s: %w", source, err)
	}
	if int64(len(data)) > limit {
		return nil, "", fmt.Errorf("the asset exceeds the %d byte cap", limit)
	}

	mediaType, err := mediaTypeOf(response.Header.Get("Content-Type"), data, fallbackMediaType)
	if err != nil {
		return nil, "", fmt.Errorf("download %s: %w", source, err)
	}

	return data, mediaType, nil
}

func DetectAssetMediaType(data []byte) (string, bool) {
	detected := http.DetectContentType(data)

	return normalizeMediaType(detected)
}

func mediaTypeOf(header string, data []byte, fallback string) (string, error) {
	if mediaType, ok := normalizeMediaType(header); ok {
		return mediaType, nil
	}

	if mediaType, ok := DetectAssetMediaType(data); ok {
		return mediaType, nil
	}

	if mediaType, ok := normalizeMediaType(fallback); ok {
		return mediaType, nil
	}

	return "", fmt.Errorf("unrecognised media type %q", header)
}

func normalizeMediaType(candidate string) (string, bool) {
	mediaType, _, _ := strings.Cut(candidate, ";")
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))

	_, ok := assetExtensions[mediaType]

	return mediaType, ok
}
