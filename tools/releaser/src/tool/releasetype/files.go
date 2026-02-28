package releasetype

import (
	"strconv"
	"strings"

	"releaser/tool/output"
)

func collectChangedFiles(raw string) (changeBuckets, bool) {
	files := trimNonEmptyLines(raw)
	if len(files) == 0 {
		return changeBuckets{}, true
	}
	output.Verbose("Changed file count: " + strconv.Itoa(len(files)))
	output.VeryVerboseList("Changed files", files, 20)

	buckets := changeBuckets{}
	for _, file := range files {
		switch {
		case strings.HasSuffix(file, ".php") && !strings.HasPrefix(file, "tests/"):
			buckets.phpFiles = append(buckets.phpFiles, file)
		}

		switch {
		case strings.HasPrefix(file, "database/migrations/"):
			buckets.migrations = append(buckets.migrations, file)
		case isDocLikeFile(file):
			buckets.docs = append(buckets.docs, file)
		case strings.HasPrefix(file, "config/"):
			buckets.configs = append(buckets.configs, file)
		case strings.HasPrefix(file, "resources/views"):
			buckets.views = append(buckets.views, file)
		}

		if file == "composer.json" || file == "composer.lock" {
			buckets.composerFiles = append(buckets.composerFiles, file)
		}
	}

	return buckets, false
}

func isDocLikeFile(file string) bool {
	return strings.HasPrefix(file, "docs/") ||
		strings.HasSuffix(file, ".md") ||
		strings.HasSuffix(file, ".rst") ||
		strings.HasSuffix(file, ".txt")
}
