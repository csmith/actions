package wowinterface

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, apiKey, addonID, path string) error {
	resolved := ctx.ResolvePath(path)
	matches, err := filepath.Glob(resolved)
	if err != nil {
		return fmt.Errorf("invalid glob pattern %q: %w", path, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no files matched pattern %q", path)
	}
	if len(matches) > 1 {
		return fmt.Errorf("expected exactly one file, but pattern %q matched %d files", path, len(matches))
	}

	filePath := matches[0]
	version, err := extractVersion(filePath)
	if err != nil {
		return err
	}

	slog.Info("Uploading to WowInterface", "file", filePath, "addon", addonID, "version", version)

	if err := upload(apiKey, addonID, version, filePath); err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	slog.Info("Upload complete")
	return nil
}

func extractVersion(path string) (string, error) {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".zip")
	idx := strings.LastIndex(name, "-")
	if idx < 0 {
		return "", fmt.Errorf("cannot extract version from filename %q: expected modname-version.zip format", filepath.Base(path))
	}
	return name[idx+1:], nil
}

func upload(apiKey, addonID, version, filePath string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("id", addonID); err != nil {
		return fmt.Errorf("failed to write id field: %w", err)
	}

	if err := w.WriteField("version", version); err != nil {
		return fmt.Errorf("failed to write version field: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	part, err := w.CreateFormFile("updatefile", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("failed to write file contents: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.wowinterface.com/addons/update", &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-api-token", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	slog.Info("Response", "status", resp.StatusCode, "body", string(body))
	return nil
}
