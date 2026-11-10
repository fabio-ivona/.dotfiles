package gitops

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
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
	return run(dir, nil, "", args...)
}

// RunRemote retries GitHub SSH authentication failures using the API token.
// Configuration and credentials apply only to the retried child process.
func RunRemote(cfg *shared.Config, args ...string) (string, error) {
	out, err := Run(cfg.BaseDir, args...)
	if err == nil || cfg.Token == "" || !strings.Contains(strings.ToLower(out), "permission denied (publickey)") {
		return out, err
	}

	output.Info("SSH key authentication failed; retrying GitHub access over HTTPS with GITHUB_TOKEN.")
	auth := "AUTHORIZATION: basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+cfg.Token))
	env := os.Environ()
	// Append to any inherited command-scoped Git configuration.
	count, _ := strconv.Atoi(os.Getenv("GIT_CONFIG_COUNT"))
	settings := [][2]string{
		{"url.https://github.com/.insteadOf", "git@github.com:"},
		{"url.https://github.com/.insteadOf", "ssh://git@github.com/"},
		{"http.https://github.com/.extraheader", ""},
		{"http.https://github.com/.extraheader", auth},
		{"credential.helper", ""},
	}
	for i, setting := range settings {
		env = append(env, fmt.Sprintf("GIT_CONFIG_KEY_%d=%s", count+i, setting[0]), fmt.Sprintf("GIT_CONFIG_VALUE_%d=%s", count+i, setting[1]))
	}
	env = append(env, fmt.Sprintf("GIT_CONFIG_COUNT=%d", count+len(settings)), "GIT_TERMINAL_PROMPT=0")
	return run(cfg.BaseDir, env, cfg.Token, args...)
}

func run(dir string, env []string, token string, args ...string) (string, error) {
	output.VeryVerbose("git -C " + dir + " " + strings.Join(args, " "))
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if token != "" {
		redacted := strings.ReplaceAll(string(out), token, "[REDACTED]")
		out = []byte(strings.ReplaceAll(redacted, base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token)), "[REDACTED]"))
	}
	outputText := strings.TrimSpace(string(out))
	if err != nil {
		if outputText == "" {
			return string(out), fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
		}
		return string(out), fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, outputText)
	}
	return string(out), nil
}

func TagExists(dir, tag string) (bool, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "-q", "--verify", "refs/tags/"+tag)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	outputText := strings.TrimSpace(string(out))
	if outputText == "" {
		return false, fmt.Errorf("failed to check local tag %s: %w", tag, err)
	}
	return false, fmt.Errorf("failed to check local tag %s: %w: %s", tag, err, outputText)
}

func RemoteTagExists(cfg *shared.Config, tag string) (bool, error) {
	out, err := RunRemote(cfg, "ls-remote", "--tags", "origin", "refs/tags/"+tag)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

func AheadBehind(dir string) (ahead int, behind int, hasUpstream bool, err error) {
	output.Verbose("Checking branch ahead/behind status")
	upstreamCmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if out, upstreamErr := upstreamCmd.CombinedOutput(); upstreamErr != nil {
		var exitErr *exec.ExitError
		if errors.As(upstreamErr, &exitErr) && exitErr.ExitCode() == 128 {
			return 0, 0, false, nil
		}
		outputText := strings.TrimSpace(string(out))
		if outputText == "" {
			return 0, 0, false, fmt.Errorf("failed to resolve upstream branch: %w", upstreamErr)
		}
		return 0, 0, false, fmt.Errorf("failed to resolve upstream branch: %w: %s", upstreamErr, outputText)
	}

	out, err := Run(dir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, true, err
	}

	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) != 2 {
		return 0, 0, true, fmt.Errorf("unexpected rev-list output for ahead/behind: %q", strings.TrimSpace(out))
	}

	ahead, err = strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, true, fmt.Errorf("invalid ahead count %q: %w", fields[0], err)
	}
	behind, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, true, fmt.Errorf("invalid behind count %q: %w", fields[1], err)
	}

	return ahead, behind, true, nil
}
