package release

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/shared"
	"releaser/tool/version"
)

func CreateTag(cfg *shared.Config) error {
	if !cfg.Force {
		fmt.Printf("Are you sure you want to create a new %s %s? [Y/n] ", cfg.Type, cfg.NewTag)
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(ans)
		switch strings.ToLower(version.DefaultYes(ans)) {
		case "y", "yes":
			// continue
		default:
			output.Info("Aborted.")
			return errors.New("aborted")
		}
	}

	output.Info("Creating new tag " + cfg.NewTag + " and pushing...")
	if _, err := gitops.Run(cfg.BaseDir, "tag", cfg.NewTag); err != nil {
		output.Warn("Failed to create/push tag " + cfg.NewTag)
		return err
	}
	if _, err := gitops.Run(cfg.BaseDir, "push"); err != nil {
		output.Warn("Failed to create/push tag " + cfg.NewTag)
		return err
	}
	if _, err := gitops.Run(cfg.BaseDir, "push", "--tags"); err != nil {
		output.Warn("Failed to create/push tag " + cfg.NewTag)
		return err
	}
	return nil
}

func BuildChanges(cfg *shared.Config) error {
	output.Info("Detecting changes for release notes...")
	log, err := gitops.Run(cfg.BaseDir, "log", cfg.OldTag+"..HEAD", "--pretty=format:%s####%an")
	if err != nil || strings.TrimSpace(log) == "" {
		cfg.Changes = "## What's Changed\n\n"
		if err != nil {
			cfg.Changes += "No commits found since " + cfg.OldTag + "\n\n"
		} else {
			cfg.Changes += "No commits found\n\n"
		}
		cfg.Changes += fmt.Sprintf("**Full Changelog**: https://github.com/%s/compare/%s...%s", cfg.Repo, cfg.OldTag, cfg.NewTag)
		return nil
	}

	var body strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "####", 2)
		msg := parts[0]
		author := ""
		if len(parts) == 2 {
			author = parts[1]
		}
		if author == "dependabot[bot]" {
			author = "dependabot"
		}
		body.WriteString("- **")
		body.WriteString(msg)
		body.WriteString("**")
		if author != "" {
			body.WriteString(" by ")
			body.WriteString(author)
		}
		body.WriteString("\n")
	}

	cfg.Changes = "## What's Changed\n\n" + body.String() + "\n**Full Changelog**: https://github.com/" + cfg.Repo + "/compare/" + cfg.OldTag + "..." + cfg.NewTag
	return nil
}
