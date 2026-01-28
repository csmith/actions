package imagetags

import (
	"slices"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestParseVersions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "valid ordinary release versions",
			input:    []string{"v1.2.3", "v2.0.0", "v0.1.0"},
			expected: []string{"1.2.3", "2.0.0", "0.1.0"},
		},
		{
			name:     "valid versions without v prefix",
			input:    []string{"1.2.3", "2.0.0"},
			expected: []string{"1.2.3", "2.0.0"},
		},
		{
			name:     "filters out versions with metadata",
			input:    []string{"v1.2.3+001", "v2.0.0+exp", "v1.2.3"},
			expected: []string{"1.2.3"},
		},
		{
			name:     "filters out versions with prerelease",
			input:    []string{"v1.2.3-beta", "v2.0.0-alpha", "v1.2.3"},
			expected: []string{"1.2.3"},
		},
		{
			name:     "filters out versions with both metadata and prerelease",
			input:    []string{"v1.2.3-beta+001", "v2.0.0-alpha+exp.sha.5114f85"},
			expected: []string{},
		},
		{
			name:     "handles invalid versions gracefully",
			input:    []string{"not-a-version", "v1.2.3", "also-invalid"},
			expected: []string{"1.2.3"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "mixed versions",
			input: []string{
				"v1.0.0",
				"v1.1.0-beta",
				"v1.2.0+build123",
				"v2.0.0",
				"invalid",
				"v2.1.0-rc1+abc123",
			},
			expected: []string{"1.0.0", "2.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersions(slices.Values(tt.input))

			actual := make([]string, len(result))
			for i, v := range result {
				actual[i] = v.Original()
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name              string
		targetVersion     string
		availableVersions []string
		expected          []string
	}{
		{
			name:              "first release - gets all tags",
			targetVersion:     "1.0.0",
			availableVersions: []string{},
			expected:          []string{"1.0.0", "1.0", "1", "latest"},
		},
		{
			name:              "3.1.2 with no newer versions - gets all tags",
			targetVersion:     "3.1.2",
			availableVersions: []string{"1.0.0", "2.0.0", "3.0.0", "3.1.0", "3.1.1"},
			expected:          []string{"3.1.2", "3.1", "3", "latest"},
		},
		{
			name:              "2.9.2 when v3 exists - gets 2.9.2 and 2.9 only",
			targetVersion:     "2.9.2",
			availableVersions: []string{"3.0.0"},
			expected:          []string{"2.9.2", "2.9", "2"},
		},
		{
			name:              "2.9.2 when v3 exists and newer v2 minor exists - gets 2.9.2 and 2.9",
			targetVersion:     "2.9.2",
			availableVersions: []string{"2.10.0", "3.0.0"},
			expected:          []string{"2.9.2", "2.9"},
		},
		{
			name:              "3.1.2 when 3.1.3 exists - gets 3.1.2 only",
			targetVersion:     "3.1.2",
			availableVersions: []string{"3.1.3"},
			expected:          []string{"3.1.2"},
		},
		{
			name:              "3.1.2 when 3.2.0 exists - gets 3.1.2 and 3.1",
			targetVersion:     "3.1.2",
			availableVersions: []string{"3.2.0"},
			expected:          []string{"3.1.2", "3.1"},
		},
		{
			name:              "3.1.2 when 4.0.0 exists - gets 3.1.2, 3.1, 3",
			targetVersion:     "3.1.2",
			availableVersions: []string{"4.0.0"},
			expected:          []string{"3.1.2", "3.1", "3"},
		},
		{
			name:              "3.1.2 with newer patch but same minor/major - gets 3.1.2 only",
			targetVersion:     "3.1.2",
			availableVersions: []string{"3.1.5"},
			expected:          []string{"3.1.2"},
		},
		{
			name:              "3.1.2 with newer minor but same major - gets 3.1.2 and 3.1",
			targetVersion:     "3.1.2",
			availableVersions: []string{"3.5.0"},
			expected:          []string{"3.1.2", "3.1"},
		},
		{
			name:              "multiple versions in history",
			targetVersion:     "2.5.3",
			availableVersions: []string{"1.0.0", "1.5.0", "2.0.0", "2.5.0", "2.5.2", "3.0.0"},
			expected:          []string{"2.5.3", "2.5", "2"},
		},
		{
			name:              "only major tag when newer patch and minor exist",
			targetVersion:     "1.2.3",
			availableVersions: []string{"1.2.4", "1.3.0"},
			expected:          []string{"1.2.3"},
		},
		{
			name:              "pre-release version with metadata - only exact tag",
			targetVersion:     "1.2.3-beta+001",
			availableVersions: []string{},
			expected:          []string{"1.2.3-beta+001"},
		},
		{
			name:              "pre-release version without metadata - only exact tag",
			targetVersion:     "1.2.3-beta",
			availableVersions: []string{},
			expected:          []string{"1.2.3-beta"},
		},
		{
			name:              "version with metadata only - only exact tag",
			targetVersion:     "1.2.3+001",
			availableVersions: []string{},
			expected:          []string{"1.2.3+001"},
		},
		{
			name:              "pre-release target ignores available versions",
			targetVersion:     "2.0.0-rc1",
			availableVersions: []string{"2.0.0", "2.1.0", "3.0.0"},
			expected:          []string{"2.0.0-rc1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := version.Must(version.NewVersion(tt.targetVersion))
			available := make([]*version.Version, len(tt.availableVersions))
			for i, v := range tt.availableVersions {
				available[i] = version.Must(version.NewVersion(v))
			}

			result := resolve(target, available)
			assert.Equal(t, tt.expected, result)
		})
	}
}
