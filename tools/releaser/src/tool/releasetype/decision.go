package releasetype

import (
	"strings"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func applyFileCategorySignals(buckets changeBuckets, signals *releaseSignals) {
	if len(buckets.composerFiles) > 0 {
		output.Info("- composer.json/lock changed → patch")
	}
	if len(buckets.views) > 0 {
		markMinor(signals, "- new views → Minor ["+strings.Join(buckets.views, " ")+"]")
	}
	if len(buckets.migrations) > 0 {
		markMinor(signals, "- new migrations → Minor ["+strings.Join(buckets.migrations, " ")+"]")
	}
	if len(buckets.configs) > 0 {
		markMinor(signals, "- new configs → Minor ["+strings.Join(buckets.configs, " ")+"]")
	}
}

func applyFinalDecision(cfg *shared.Config, buckets changeBuckets, signals *releaseSignals) {
	output.Blank()

	if buckets.hasOnlyDocsOrTests() {
		output.Info("🧪 Only tests/docs changed → PATCH [" + strings.Join(buckets.docsAndTests(), ", ") + "]")
		cfg.Type = "patch"
		return
	}

	switch {
	case signals.major:
		output.Info("🧨 Detected MAJOR changes")
		cfg.Type = "major"
	case signals.minor:
		output.Info("✨ Detected Minor changes")
		cfg.Type = "minor"
	default:
		output.Info("🐛 Only safe changes → PATCH")
		cfg.Type = "patch"
	}
}

func markMajor(signals *releaseSignals, message string) {
	output.Info(message)
	signals.major = true
}

func markMinor(signals *releaseSignals, message string) {
	output.Info(message)
	signals.minor = true
}
