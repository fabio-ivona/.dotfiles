package githubapi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func GetCurrentVersion(cfg *shared.Config) error {
	label := "Fetching latest GitHub release"
	output.Info(label + "...")
	output.Verbose("Repository: " + cfg.Repo)

	resp, err := request("GET", "https://api.github.com/repos/"+cfg.Repo+"/releases/latest", cfg.Token, nil)
	if err != nil {
		output.ReplaceLastLine(label + " ⚠")
		output.Warn("Failed to call GitHub API for latest release")
		return err
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		output.ReplaceLastLine(label + " ⚠")
		output.Warn("Failed to fetch latest tag from GitHub")
		fmt.Println(string(resp))
		return err
	}
	if payload.TagName == "" {
		output.ReplaceLastLine(label + " ⚠")
		output.Warn("Failed to fetch latest tag from GitHub")
		fmt.Println(string(resp))
		return errors.New("missing tag_name")
	}

	cfg.OldTag = payload.TagName
	cfg.OldVer = strings.TrimPrefix(cfg.OldTag, "v")
	output.ReplaceLastLine(label + ": " + cfg.OldTag + " ✔")
	return nil
}

func CreateRelease(cfg *shared.Config) error {
	output.Info("Creating GitHub release " + cfg.NewTag + "...")
	output.Verbose("Release compare URL: https://github.com/" + cfg.Repo + "/compare/" + cfg.OldTag + "..." + cfg.NewTag)
	payload := map[string]any{
		"tag_name":   cfg.NewTag,
		"name":       cfg.NewTag,
		"body":       cfg.Changes,
		"draft":      false,
		"prerelease": false,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := request("POST", "https://api.github.com/repos/"+cfg.Repo+"/releases", cfg.Token, b)
	if err != nil {
		output.Warn("Failed to call GitHub API for release creation")
		return err
	}

	var out struct {
		HTMLURL     string `json:"html_url"`
		PublishedAt string `json:"published_at"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		output.Warn("Failed to create GitHub release; response was:")
		fmt.Println(string(resp))
		return err
	}
	if out.HTMLURL == "" {
		output.Warn("Failed to create GitHub release; response was:")
		fmt.Println(string(resp))
		return errors.New("missing html_url")
	}

	cfg.Release = out.HTMLURL
	cfg.Published = out.PublishedAt
	return nil
}

func FollowReleaseWorkflow(cfg *shared.Config) error {
	label := "Following release workflow status"
	output.Info(label + "...")

	since := time.Now().UTC().Add(-2 * time.Minute)
	if cfg.Published != "" {
		publishedAt, err := time.Parse(time.RFC3339, cfg.Published)
		if err == nil {
			since = publishedAt.Add(-30 * time.Second)
		}
	}

	deadline := time.Now().Add(2 * time.Minute)
	for {
		run, found, err := latestReleaseWorkflowRun(cfg, since)
		if err != nil {
			output.ReplaceLastLine(label + " ⚠")
			output.Warn("Failed to query GitHub Actions runs")
			return err
		}
		if found {
			output.ReplaceLastLine(label + ": found run ✔")
			if run.HTMLURL != "" {
				output.Continue("Run: " + run.HTMLURL)
			}
			return followWorkflowRunUntilTerminal(cfg, run)
		}
		if time.Now().After(deadline) {
			output.ReplaceLastLine(label + " ⚠")
			return errors.New("no release workflow run found within timeout")
		}
		time.Sleep(5 * time.Second)
	}
}

func followWorkflowRunUntilTerminal(cfg *shared.Config, run workflowRun) error {
	previous := ""
	spinnerIndex := 0
	enterCh := startEnterListener()
	if enterCh != nil {
		output.Continue("Press Enter to stop following.")
	}
	for {
		if enterCh != nil {
			select {
			case <-enterCh:
				output.Warn("Stopped following workflow status.")
				return nil
			default:
			}
		}

		currentRun, err := getWorkflowRun(cfg, run.ID)
		if err != nil {
			output.Warn("Failed to fetch workflow run status")
			return err
		}

		current := normalizeWorkflowStatus(currentRun.Status, currentRun.Conclusion)
		if current != previous {
			message := "Workflow status: " + output.WorkflowStatus(current) + " " + statusSymbol(current)
			if previous != "" {
				message = "Workflow status: " + output.WorkflowStatus(previous) + " -> " + output.WorkflowStatus(current) + " " + statusSymbol(current)
			}

			if current == "completed" {
				output.Success(message)
			} else if current == "failed" {
				output.Error(message)
			} else if current == "skipped" {
				output.Warn(message)
			} else {
				output.Info(message)
			}
			previous = current
		} else if current == "running" {
			output.Info("Workflow status: " + output.WorkflowStatus(current) + " " + spinnerFrame(spinnerIndex))
			spinnerIndex++
		}

		if current == "completed" || current == "skipped" || current == "failed" {
			return nil
		}

		if enterCh == nil {
			time.Sleep(5 * time.Second)
		} else {
			select {
			case <-enterCh:
				output.Warn("Stopped following workflow status.")
				return nil
			case <-time.After(5 * time.Second):
			}
		}
	}
}

func spinnerFrame(index int) string {
	frames := []string{"|", "/", "-", `\`}
	return frames[index%len(frames)]
}

func startEnterListener() <-chan struct{} {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil || (stdinInfo.Mode()&os.ModeCharDevice) == 0 {
		return nil
	}

	ch := make(chan struct{})
	go func() {
		reader := bufio.NewReader(os.Stdin)
		_, err := reader.ReadString('\n')
		if err == nil || errors.Is(err, io.EOF) {
			close(ch)
		}
	}()
	return ch
}

func latestReleaseWorkflowRun(cfg *shared.Config, since time.Time) (workflowRun, bool, error) {
	resp, err := request("GET", "https://api.github.com/repos/"+cfg.Repo+"/actions/runs?event=release&per_page=20", cfg.Token, nil)
	if err != nil {
		return workflowRun{}, false, err
	}

	var payload struct {
		WorkflowRuns []workflowRun `json:"workflow_runs"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		return workflowRun{}, false, err
	}

	for _, run := range payload.WorkflowRuns {
		createdAt, err := time.Parse(time.RFC3339, run.CreatedAt)
		if err != nil {
			continue
		}
		if createdAt.Before(since) {
			continue
		}
		return run, true, nil
	}

	return workflowRun{}, false, nil
}

func getWorkflowRun(cfg *shared.Config, id int64) (workflowRun, error) {
	resp, err := request("GET", fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d", cfg.Repo, id), cfg.Token, nil)
	if err != nil {
		return workflowRun{}, err
	}

	var run workflowRun
	if err := json.Unmarshal(resp, &run); err != nil {
		return workflowRun{}, err
	}
	return run, nil
}

func normalizeWorkflowStatus(status, conclusion string) string {
	switch strings.ToLower(status) {
	case "queued", "requested", "waiting", "pending":
		return "queued"
	case "in_progress":
		return "running"
	case "completed":
		switch strings.ToLower(conclusion) {
		case "skipped":
			return "skipped"
		case "success", "neutral":
			return "completed"
		case "":
			return "completed"
		default:
			return "failed"
		}
	default:
		return "queued"
	}
}

func statusSymbol(status string) string {
	switch status {
	case "queued":
		return "⌛"
	case "running":
		return "⌛"
	case "completed":
		return "✔"
	case "skipped":
		return "⏭"
	case "failed":
		return "⚠"
	default:
		return "⌛"
	}
}

type workflowRun struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	CreatedAt  string `json:"created_at"`
	HTMLURL    string `json:"html_url"`
}

func request(method, url, token string, body []byte) ([]byte, error) {
	output.Verbose("GitHub API request: " + method + " " + url)
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		output.Verbose("GitHub API non-success status: " + resp.Status)
		return b, fmt.Errorf("github api status %s", resp.Status)
	}
	output.Verbose("GitHub API response status: " + resp.Status)
	return b, nil
}
