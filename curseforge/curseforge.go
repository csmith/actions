package curseforge

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
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

type gameVersion struct {
	ID                int    `json:"id"`
	GameVersionTypeID int    `json:"gameVersionTypeID"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	APIVersion        string `json:"apiVersion"`
}

type metadata struct {
	Changelog                string   `json:"changelog"`
	ChangelogType            string   `json:"changelogType"`
	DisplayName              string   `json:"displayName"`
	GameVersions             []int    `json:"gameVersions"`
	GameVersionNames         []string `json:"gameVersionsNames"`
	ReleaseType              string   `json:"releaseType"`
	IsMarkedForManualRelease bool     `json:"isMarkedForManualRelease"`
}

func Run(ctx *common.Context, apiToken, projectID, path, changelogFile string) error {
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

	interfaceVersions, err := extractInterfaceFromZip(filePath)
	if err != nil {
		return fmt.Errorf("failed to read interface version from zip: %w", err)
	}

	slog.Info("Found interface versions in TOC", "interface", interfaceVersions)

	var changelogStr string
	changelogSection, err := ctx.Changelog(changelogFile, version, ctx.Tag())
	if err == nil {
		changelogStr = changelogSection
	}

	slog.Info("Uploading to CurseForge", "file", filePath, "project", projectID, "version", version)

	gameVersions, err := fetchGameVersions(apiToken)
	if err != nil {
		return fmt.Errorf("failed to fetch game versions: %w", err)
	}

	matchedVersions := matchVersions(gameVersions, interfaceVersions)
	if len(matchedVersions) == 0 {
		return fmt.Errorf("no game versions matched interface versions %v", interfaceVersions)
	}

	versionIDs := make([]int, len(matchedVersions))
	versionNames := make([]string, len(matchedVersions))
	for i, v := range matchedVersions {
		versionIDs[i] = v.ID
		versionNames[i] = v.Name
		slog.Info("Matched game version", "id", v.ID, "name", v.Name, "apiVersion", v.APIVersion)
	}

	if err := upload(apiToken, projectID, version, changelogStr, versionIDs, versionNames, filePath); err != nil {
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

func extractInterfaceFromZip(zipPath string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	var topLevelFolder string
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			parts := strings.SplitN(f.Name, "/", 2)
			if len(parts) < 2 {
				continue
			}
			candidate := parts[0]
			if topLevelFolder == "" {
				topLevelFolder = candidate
			} else if candidate != topLevelFolder {
				return nil, fmt.Errorf("zip contains multiple top-level folders (%q and %q), expected exactly one", topLevelFolder, candidate)
			}
		}
	}

	if topLevelFolder == "" {
		return nil, fmt.Errorf("zip appears to be empty")
	}

	tocName := topLevelFolder + "/" + topLevelFolder + ".toc"
	for _, f := range r.File {
		if f.Name != tocName {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open %s in zip: %w", f.Name, err)
		}
		defer rc.Close()

		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if after, ok := strings.CutPrefix(line, "## Interface:"); ok {
				raw := strings.TrimSpace(after)
				var versions []string
				for v := range strings.SplitSeq(raw, ",") {
					v = strings.TrimSpace(v)
					if v != "" {
						versions = append(versions, v)
					}
				}
				if len(versions) == 0 {
					return nil, fmt.Errorf("## Interface: found in %s but no valid versions", f.Name)
				}
				return versions, nil
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
		}

		return nil, fmt.Errorf("## Interface: not found in %s", f.Name)
	}

	return nil, fmt.Errorf("expected toc file %s not found in zip", tocName)
}

func fetchGameVersions(apiToken string) ([]gameVersion, error) {
	req, err := http.NewRequest("GET", "https://wow.curseforge.com/api/game/versions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Api-Token", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("game versions API returned status %d: %s", resp.StatusCode, string(body))
	}

	var versions []gameVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse game versions: %w", err)
	}

	return versions, nil
}

func matchVersions(versions []gameVersion, interfaceVersions []string) []gameVersion {
	want := make(map[string]bool)
	for _, iv := range interfaceVersions {
		want[iv] = true
	}
	var matched []gameVersion
	for _, v := range versions {
		if want[v.APIVersion] {
			matched = append(matched, v)
		}
	}
	return matched
}

func upload(apiToken, projectID, version, changelog string, gameVersionIDs []int, gameVersionNames []string, filePath string) error {
	meta := metadata{
		Changelog:                changelog,
		ChangelogType:            "markdown",
		DisplayName:              version,
		GameVersions:             gameVersionIDs,
		GameVersionNames:         gameVersionNames,
		ReleaseType:              "release",
		IsMarkedForManualRelease: false,
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("metadata", string(metaJSON)); err != nil {
		return fmt.Errorf("failed to write metadata field: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("failed to write file contents: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("https://wow.curseforge.com/api/projects/%s/upload-file", projectID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-Api-Token", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	slog.Info("Response", "status", resp.StatusCode, "body", string(body))
	return nil
}
