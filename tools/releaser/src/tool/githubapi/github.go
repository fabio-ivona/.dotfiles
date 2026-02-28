package githubapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func GetCurrentVersion(cfg *shared.Config) error {
	label := "Fetching latest GitHub release"
	output.Info(label + "...")

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
		HTMLURL string `json:"html_url"`
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
	return nil
}

func request(method, url, token string, body []byte) ([]byte, error) {
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
		return b, fmt.Errorf("github api status %s", resp.Status)
	}
	return b, nil
}
