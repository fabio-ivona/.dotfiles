package gitops

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func CheckUncommittedChanges(cfg *shared.Config) error {
	label := "Checking for uncommitted changes"
	output.Info(label + "...")
	out, err := Run(cfg.BaseDir, "status", "--porcelain")
	if err != nil {
		output.ReplaceLastLine(label + " ✖")
		output.Warn("Failed to check for uncommitted changes")
		return err
	}
	if strings.TrimSpace(out) != "" {
		output.ReplaceLastLine(label + " ⚠")
		output.Warn("⚠️  There are uncommitted changes in your working directory:")
		fmt.Print(out)
		return errors.New("uncommitted changes")
	}
	output.ReplaceLastLine(label + " ✔")
	return nil
}

func GetRepository(cfg *shared.Config) error {
	label := "Detecting repository"
	output.Info(label + "...")
	out, err := Run(cfg.BaseDir, "config", "--get", "remote.origin.url")
	if err != nil {
		output.ReplaceLastLine(label + " ⚠")
		output.Warn("Failed to detect repository (git config remote.origin.url)")
		return err
	}

	url := strings.TrimSpace(out)
	var repo string
	if strings.Contains(url, "github.com:") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			repo = parts[1]
		}
	} else if idx := strings.Index(url, "github.com/"); idx >= 0 {
		repo = url[idx+len("github.com/"):]
	} else {
		output.ReplaceLastLine(label + " ⚠")
		output.Warn("Remote origin does not look like a GitHub URL: " + url)
		return errors.New("non-github remote")
	}

	repo = strings.TrimSuffix(repo, ".git")
	cfg.Repo = repo
	output.ReplaceLastLine(label + ": " + cfg.Repo + " ✔")
	return nil
}

func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
