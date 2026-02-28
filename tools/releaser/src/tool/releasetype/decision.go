package releasetype

import (
	"fmt"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func applyFileCategorySignals(buckets changeBuckets, signals *releaseSignals) {
	output.Verbose("Applying non-PHP category signals")
	if len(buckets.composerFiles) > 0 {
		markPatch(signals, "composer.json/lock changed")
		output.VeryVerboseList("Composer files", buckets.composerFiles, 10)
	}
	if len(buckets.views) > 0 {
		markMinor(signals, fmt.Sprintf("views changed (%d)", len(buckets.views)))
		output.VeryVerboseList("View files", buckets.views, 10)
	}
	if len(buckets.migrations) > 0 {
		markMinor(signals, fmt.Sprintf("migrations changed (%d)", len(buckets.migrations)))
		output.VeryVerboseList("Migration files", buckets.migrations, 10)
	}
	if len(buckets.configs) > 0 {
		markMinor(signals, fmt.Sprintf("configs changed (%d)", len(buckets.configs)))
		output.VeryVerboseList("Config files", buckets.configs, 10)
	}
}

func applyFinalDecision(cfg *shared.Config, buckets changeBuckets, signals *releaseSignals) {
	output.Blank()
	output.Verbose("Applying final release-type decision")
	signals.emitRules()

	if buckets.hasOnlyDocs() {
		files := buckets.docsFiles()
		output.Info(fmt.Sprintf("🧪 Only docs changed (%d files) → %s", len(files), output.SemverLabel("patch")))
		output.VeryVerboseList("Doc files", files, 20)
		cfg.Type = "patch"
		return
	}

	switch {
	case signals.major:
		output.Info("🧨 Detected " + output.SemverLabel("major") + " changes")
		cfg.Type = "major"
	case signals.minor:
		output.Info("✨ Detected " + output.SemverLabel("minor") + " changes")
		cfg.Type = "minor"
	default:
		output.Info("🐛 Only safe changes → " + output.SemverLabel("patch"))
		cfg.Type = "patch"
	}
}

func markMajor(signals *releaseSignals, message string) {
	signals.major = true
	signals.addGlobalRule(output.SemverLabel("major") + " | " + message)
}

func markMinor(signals *releaseSignals, message string) {
	signals.minor = true
	signals.addGlobalRule(output.SemverLabel("minor") + " | " + message)
}

func markPatch(signals *releaseSignals, message string) {
	signals.addGlobalRule(output.SemverLabel("patch") + " | " + message)
}

func markMajorForFile(signals *releaseSignals, file, reason, snippet string) {
	signals.major = true
	signals.addFileRule(file, "major", reason, snippet)
}

func markMinorForFile(signals *releaseSignals, file, reason, snippet string) {
	signals.minor = true
	signals.addFileRule(file, "minor", reason, snippet)
}
