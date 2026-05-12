package wowaddon

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, src, dst string) error {
	srcPath := ctx.ResolvePath(src)
	dstPath := ctx.ResolvePath(dst)

	name, err := addonName(srcPath)
	if err != nil {
		return err
	}

	tag := ctx.Tag()
	zipName := zipFileName(name, tag)
	zipPath := filepath.Join(dstPath, zipName)

	if err := os.MkdirAll(dstPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	f, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	count, err := addFiles(w, srcPath, name)
	if err != nil {
		return err
	}

	slog.Info("Created addon zip", "path", zipPath, "files", count)
	return nil
}

func addonName(src string) (string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return "", fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toc") {
			return strings.TrimSuffix(entry.Name(), ".toc"), nil
		}
	}

	return "", fmt.Errorf("no .toc file found in %s", src)
}

func zipFileName(name, tag string) string {
	if tag != "" {
		return fmt.Sprintf("%s-%s.zip", name, tag)
	}
	return fmt.Sprintf("%s.zip", name)
}

func addFiles(w *zip.Writer, src, prefix string) (int, error) {
	count := 0
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(prefix, rel)
		slog.Debug("Adding file", "source", path, "zip", zipPath)

		if err := addFile(w, path, zipPath); err != nil {
			return err
		}

		count++
		return nil
	})
	return count, err
}

func addFile(w *zip.Writer, path, zipPath string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", path, err)
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	out, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry for %s: %w", path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	_, err = io.Copy(out, f)
	return err
}
