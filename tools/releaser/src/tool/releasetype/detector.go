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
		output.Info("No code changes detected → " + output.SemverLabel("patch"))
		cfg.Type = "patch"
		return nil
	}
	output.Verbose("Changed files by category: " + buckets.summary())
	logBucketDetails(buckets)

	signals := newReleaseSignals()
	analyzePHPChanges(cfg, buckets.phpFiles, signals)
	applyFileCategorySignals(buckets, signals)
	output.Verbose("Signals before final decision: major=" + boolString(signals.major) + " minor=" + boolString(signals.minor))
	applyFinalDecision(cfg, buckets, signals)
	output.Verbose("Final detected release type: " + cfg.Type)
	return nil
}

func analyzePHPChanges(cfg *shared.Config, phpFiles []string, signals *releaseSignals) {
	for _, file := range phpFiles {
		output.VeryVerbose("Analyzing PHP file: " + file)
		analyzePHPFile(cfg, file, signals)
	}
}

func logBucketDetails(buckets changeBuckets) {
	output.VeryVerboseList("PHP files", buckets.phpFiles, 10)
	output.VeryVerboseList("Migration files", buckets.migrations, 10)
	output.VeryVerboseList("Doc files", buckets.docs, 10)
	output.VeryVerboseList("Config files", buckets.configs, 10)
	output.VeryVerboseList("View files", buckets.views, 10)
	output.VeryVerboseList("Composer files", buckets.composerFiles, 10)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
