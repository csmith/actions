package common

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var versionHeadingRe = regexp.MustCompile(`^## \[?v?([^\]\s]+)\]?`)

func (c *Context) Changelog(filename string, versions ...string) (string, error) {
	content, err := os.ReadFile(c.ResolvePath(filename))
	if err != nil {
		return "", fmt.Errorf("failed to read changelog: %w", err)
	}
	return FindChangelogSection(string(content), versions...), nil
}

func FindChangelogSection(content string, versions ...string) string {
	normalized := make([]string, len(versions))
	for i, v := range versions {
		normalized[i] = strings.TrimPrefix(v, "v")
	}

	for _, target := range normalized {
		if section := extractSection(content, target); section != "" {
			return section
		}
	}

	return ""
}

func extractSection(content, target string) string {
	lines := strings.Split(content, "\n")
	var sectionLines []string
	found := false

	for _, line := range lines {
		matches := versionHeadingRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			if found {
				break
			}
			if matches[1] == target {
				found = true
			}
			continue
		}
		if found {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Trim(strings.Join(sectionLines, "\n"), "\n")
}
