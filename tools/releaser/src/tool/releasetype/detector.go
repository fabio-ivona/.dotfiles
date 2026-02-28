package releasetype

import (
	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/shared"
)

func Detect(cfg *shared.Config) error {
	output.Info("Detecting release type from git diff since " + cfg.OldTag + "...")
	_, _ = gitops.Run(cfg.BaseDir, "fetch", "--tags")

	changed, err := gitops.Run(cfg.BaseDir, "diff", "--name-only", cfg.OldTag+"..HEAD")
	if err != nil {
		output.Warn("Failed to run git diff for changed files")
		return err
	}

	buckets, empty := categorizeChangedFiles(changed)
	if empty {
		output.Info("No code changes detected → patch")
		cfg.Type = "patch"
		return nil
	}

	indicators := &releaseIndicators{}
	for _, file := range buckets.phpFiles {
		analyzePHPFile(cfg, file, indicators)
	}

	applyFileTypeSignals(buckets, indicators)
	finalizeReleaseType(cfg, buckets, indicators)
	return nil
}
