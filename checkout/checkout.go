package checkout

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context) error {
	if err := os.MkdirAll(ctx.Workspace, 0755); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	if err := os.Chdir(ctx.Workspace); err != nil {
		return fmt.Errorf("failed to change to workspace: %w", err)
	}

	token := "x-access-token:" + ctx.Token
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	repoURL := fmt.Sprintf("%s/%s.git", ctx.ServerURL, ctx.Repository)
	cmd := exec.Command("git", "-c", fmt.Sprintf("http.extraHeader=Authorization: basic %s", encodedToken), "clone", repoURL, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, output)
	}

	cmd = exec.Command("git", "checkout", ctx.SHA)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, output)
	}

	return nil
}
