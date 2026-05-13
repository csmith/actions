package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testChangelog = `# Changelog

## Unreleased

## 1.2.0 - 2024-03-01

### Added
- New feature X
- New feature Y

### Fixed
- Bug fix Z

## 1.1.0 - 2024-02-01

### Added
- Feature A

## 1.0.0 - 2024-01-01

Initial release.
`

const testChangelogBracketed = `# Changelog

## [Unreleased]

## [1.2.0] - 2024-03-01

### Added
- New feature X

## [1.1.0] - 2024-02-01

### Added
- Feature A
`

func TestFindChangelogSection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		versions []string
		expected string
	}{
		{
			name:     "exact version match",
			content:  testChangelog,
			versions: []string{"1.2.0"},
			expected: "### Added\n- New feature X\n- New feature Y\n\n### Fixed\n- Bug fix Z",
		},
		{
			name:     "version with v prefix matches without",
			content:  testChangelog,
			versions: []string{"v1.2.0"},
			expected: "### Added\n- New feature X\n- New feature Y\n\n### Fixed\n- Bug fix Z",
		},
		{
			name:     "middle section",
			content:  testChangelog,
			versions: []string{"1.1.0"},
			expected: "### Added\n- Feature A",
		},
		{
			name:     "last section",
			content:  testChangelog,
			versions: []string{"1.0.0"},
			expected: "Initial release.",
		},
		{
			name:     "no match returns empty",
			content:  testChangelog,
			versions: []string{"9.9.9"},
			expected: "",
		},
		{
			name:     "matches first of multiple versions",
			content:  testChangelog,
			versions: []string{"1.1.0", "1.2.0"},
			expected: "### Added\n- Feature A",
		},
		{
			name:     "falls back to second version",
			content:  testChangelog,
			versions: []string{"9.9.9", "1.1.0"},
			expected: "### Added\n- Feature A",
		},
		{
			name:     "empty content",
			content:  "",
			versions: []string{"1.0.0"},
			expected: "",
		},
		{
			name:     "empty versions",
			content:  testChangelog,
			versions: []string{},
			expected: "",
		},
		{
			name:     "heading without date",
			content:  "## 3.0.0\n\nContent here\n",
			versions: []string{"3.0.0"},
			expected: "Content here",
		},
		{
			name:     "bracketed headings",
			content:  testChangelogBracketed,
			versions: []string{"1.2.0"},
			expected: "### Added\n- New feature X",
		},
		{
			name:     "bracketed headings with v prefix",
			content:  testChangelogBracketed,
			versions: []string{"v1.2.0"},
			expected: "### Added\n- New feature X",
		},
		{
			name:     "heading with v-prefix in changelog",
			content:  "## [v2.0.0]\n\nSome changes\n",
			versions: []string{"2.0.0"},
			expected: "Some changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindChangelogSection(tt.content, tt.versions...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
