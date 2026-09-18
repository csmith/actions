package setupbun

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"chameth.com/actions/common"
)

func TestInstall(t *testing.T) {
	ctx, workspace := newTestContext(t)

	source := filepath.Join(t.TempDir(), "bun")
	writeExecutableFixture(t, source, true)

	require.NoError(t, install(ctx, "tools/bun", source))

	binPath := filepath.Join(workspace, "tools", "bun", "bin")

	installedPath := filepath.Join(binPath, "bun")
	info, err := os.Stat(installedPath)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0111, "installed binary should be executable")

	output, err := exec.Command(installedPath, "test").CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "bun-fixture test\n", string(output))

	aliasPath := filepath.Join(binPath, "bunx")
	linkTarget, err := os.Readlink(aliasPath)
	require.NoError(t, err)
	assert.Equal(t, "bun", linkTarget)

	output, err = exec.Command(aliasPath, "install").CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "bun-fixture install\n", string(output))

	pathContent, err := os.ReadFile(ctx.PathFile)
	require.NoError(t, err)
	assert.Equal(t, binPath+"\n", string(pathContent))
}

func TestInstallIsIdempotent(t *testing.T) {
	ctx, workspace := newTestContext(t)

	source := filepath.Join(t.TempDir(), "bun")
	writeExecutableFixture(t, source, true)

	require.NoError(t, install(ctx, "tools/bun", source))
	require.NoError(t, install(ctx, "tools/bun", source))

	output, err := exec.Command(filepath.Join(workspace, "tools", "bun", "bin", "bunx")).CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "bun-fixture\n", string(output))
}

func TestInstallMissingSource(t *testing.T) {
	ctx, workspace := newTestContext(t)

	source := filepath.Join(t.TempDir(), "missing")

	err := install(ctx, "tools/bun", source)

	assert.ErrorContains(t, err, "failed to copy Bun binary")

	_, statErr := os.Stat(filepath.Join(workspace, "tools", "bun", "bin", "bun"))
	assert.True(t, os.IsNotExist(statErr), "binary should not be installed")

	_, statErr = os.Stat(ctx.PathFile)
	assert.True(t, os.IsNotExist(statErr), "nothing should be added to PATH")
}

func TestCopyExecutableSetsPermissions(t *testing.T) {
	source := filepath.Join(t.TempDir(), "bun")
	writeExecutableFixture(t, source, false)

	destination := filepath.Join(t.TempDir(), "installed")
	require.NoError(t, copyExecutable(source, destination))

	info, err := os.Stat(destination)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func newTestContext(t *testing.T) (*common.Context, string) {
	t.Helper()
	workspace := t.TempDir()
	return &common.Context{
		Workspace: workspace,
		PathFile:  filepath.Join(t.TempDir(), "path.txt"),
	}, workspace
}

func writeExecutableFixture(t *testing.T, path string, executable bool) {
	t.Helper()
	mode := os.FileMode(0644)
	if executable {
		mode = 0755
	}
	script := "#!/bin/sh\necho bun-fixture \"$@\"\n"
	require.NoError(t, os.WriteFile(path, []byte(script), mode))
}
