package releasetype

import (
	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/shared"
)

func Detect(cfg *shared.Config) error {
	output.Info("Detecting release type from git diff since " + cfg.OldTag + "...")
	output.Verbose("Release type diff range: " + cfg.OldTag + "..HEAD")
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
	output.Verbose("Changed files by category: " + buckets.summary())

	signals := &releaseSignals{}
	analyzePHPChanges(cfg, buckets.phpFiles, signals)
	applyFileCategorySignals(buckets, signals)
	output.Verbose("Signals before final decision: major=" + boolString(signals.major) + " minor=" + boolString(signals.minor))
	applyFinalDecision(cfg, buckets, signals)
	output.Verbose("Final detected release type: " + cfg.Type)
	return nil
}

func analyzePHPChanges(cfg *shared.Config, phpFiles []string, signals *releaseSignals) {
	for _, file := range phpFiles {
		output.Verbose("Analyzing PHP file: " + file)
		analyzePHPFile(cfg, file, signals)
	}
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
