package releasetype

import (
	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/shared"
)

func Detect(cfg *shared.Config) error {
	output.Info("Detecting release type from git diff since " + cfg.OldTag + "...")
	_, _ = gitops.Run(cfg.BaseDir, "fetch", "--tags")

	changedFilesRaw, err := gitops.Run(cfg.BaseDir, "diff", "--name-only", cfg.OldTag+"..HEAD")
	if err != nil {
		output.Warn("Failed to run git diff for changed files")
		return err
	}

	buckets, empty := collectChangedFiles(changedFilesRaw)
	if empty {
		output.Info("No code changes detected → patch")
		cfg.Type = "patch"
		return nil
	}

	signals := &releaseSignals{}
	analyzePHPChanges(cfg, buckets.phpFiles, signals)
	applyFileCategorySignals(buckets, signals)
	applyFinalDecision(cfg, buckets, signals)
	return nil
}

func analyzePHPChanges(cfg *shared.Config, phpFiles []string, signals *releaseSignals) {
	for _, file := range phpFiles {
		analyzePHPFile(cfg, file, signals)
	}
}
