package releasetype

import (
	"fmt"
	"strings"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func applyFileTypeSignals(buckets changedBuckets, indicators *releaseIndicators) {
	if len(buckets.composerFiles) > 0 {
		output.Info("- composer.json/lock changed → patch")
	}
	if len(buckets.views) > 0 {
		output.Info("- new views → Minor [" + strings.Join(buckets.views, " ") + "]")
		indicators.minor = true
	}
	if len(buckets.migrations) > 0 {
		output.Info("- new migrations → Minor [" + strings.Join(buckets.migrations, " ") + "]")
		indicators.minor = true
	}
	if len(buckets.configs) > 0 {
		output.Info("- new configs → Minor [" + strings.Join(buckets.configs, " ") + "]")
		indicators.minor = true
	}
}

func finalizeReleaseType(cfg *shared.Config, buckets changedBuckets, indicators *releaseIndicators) {
	fmt.Println()
	if len(buckets.phpFiles) == 0 && (len(buckets.tests) > 0 || len(buckets.docs) > 0) {
		all := append([]string{}, buckets.tests...)
		all = append(all, buckets.docs...)
		output.Info("🧪 Only tests/docs changed → PATCH [" + strings.Join(all, ", ") + "]")
		cfg.Type = "patch"
		return
	}

	if indicators.major {
		output.Info("🧨 Detected MAJOR changes")
		cfg.Type = "major"
	} else if indicators.minor {
		output.Info("✨ Detected Minor changes")
		cfg.Type = "minor"
	} else {
		output.Info("🐛 Only safe changes → PATCH")
		cfg.Type = "patch"
	}
}
