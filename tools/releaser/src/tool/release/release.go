package release

import (
	"errors"
	"fmt"
	"strings"

	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/shared"
	"releaser/tool/version"
)

func CreateTag(cfg *shared.Config) error {
	if !cfg.Force {
		ans := output.Ask(fmt.Sprintf("Are you sure you want to create a new %s %s? [Y/n] ", cfg.Type, cfg.NewTag))
		switch strings.ToLower(version.DefaultYes(ans)) {
		case "y", "yes":
			// continue
		default:
			output.Info("Aborted.")
			return errors.New("aborted")
		}
	}

	localTagExists, err := gitops.TagExists(cfg.BaseDir, cfg.NewTag)
	if err != nil {
		output.Warn("Failed to check if tag already exists locally")
		return err
	}
	if localTagExists {
		output.Info("Tag " + cfg.NewTag + " already exists locally; skipping tag creation.")
	} else {
		output.Info("Creating new tag " + cfg.NewTag + "...")
		if _, err := gitops.Run(cfg.BaseDir, "tag", cfg.NewTag); err != nil {
			output.Warn("Failed to create tag " + cfg.NewTag)
			return err
		}
	}

	ahead, behind, hasUpstream, err := gitops.AheadBehind(cfg.BaseDir)
	if err != nil {
		output.Warn("Failed to determine branch sync state")
		return err
	}

	shouldPushCommits := true
	if hasUpstream {
		switch {
		case ahead == 0 && behind > 0:
			output.Info(fmt.Sprintf("Branch is behind upstream by %d commit(s); skipping commit push.", behind))
			shouldPushCommits = false
		case ahead == 0 && behind == 0:
			output.Info("Branch is up to date with upstream; skipping commit push.")
			shouldPushCommits = false
		case ahead > 0 && behind > 0:
			return fmt.Errorf("branch has diverged from upstream (ahead %d, behind %d); run git pull --rebase and retry", ahead, behind)
		}
	}

	if shouldPushCommits {
		output.Info("Pushing commits...")
		if _, err := gitops.Run(cfg.BaseDir, "push"); err != nil {
			output.Warn("Failed to push commits")
			return err
		}
	}

	remoteTagExists, err := gitops.RemoteTagExists(cfg.BaseDir, cfg.NewTag)
	if err != nil {
		output.Warn("Failed to check if tag already exists on origin")
		return err
	}
	if remoteTagExists {
		output.Info("Tag " + cfg.NewTag + " already exists on origin; skipping tag push.")
		return nil
	}

	output.Info("Pushing tag " + cfg.NewTag + "...")
	if _, err := gitops.Run(cfg.BaseDir, "push", "origin", cfg.NewTag); err != nil {
		output.Warn("Failed to push tag " + cfg.NewTag)
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
