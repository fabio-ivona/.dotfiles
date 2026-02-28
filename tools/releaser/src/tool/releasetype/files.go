package releasetype

import "strings"

func collectChangedFiles(raw string) (changeBuckets, bool) {
	files := trimNonEmptyLines(raw)
	if len(files) == 0 {
		return changeBuckets{}, true
	}

	buckets := changeBuckets{}
	for _, file := range files {
		switch {
		case strings.HasSuffix(file, ".php"):
			buckets.phpFiles = append(buckets.phpFiles, file)
		}

		switch {
		case strings.HasPrefix(file, "database/migrations/"):
			buckets.migrations = append(buckets.migrations, file)
		case strings.HasPrefix(file, "tests/"):
			buckets.tests = append(buckets.tests, file)
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
