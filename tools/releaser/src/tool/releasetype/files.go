package releasetype

import "strings"

type changedBuckets struct {
	phpFiles      []string
	migrations    []string
	tests         []string
	docs          []string
	configs       []string
	views         []string
	composerFiles []string
}

type releaseIndicators struct {
	major bool
	minor bool
}

func categorizeChangedFiles(changed string) (changedBuckets, bool) {
	changed = strings.TrimSpace(changed)
	if changed == "" {
		return changedBuckets{}, true
	}

	buckets := changedBuckets{}
	for _, file := range strings.Split(changed, "\n") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		if strings.HasSuffix(file, ".php") {
			buckets.phpFiles = append(buckets.phpFiles, file)
		}
		if strings.HasPrefix(file, "database/migrations/") {
			buckets.migrations = append(buckets.migrations, file)
		}
		if strings.HasPrefix(file, "tests/") {
			buckets.tests = append(buckets.tests, file)
		}
		if strings.HasPrefix(file, "docs/") || strings.HasSuffix(file, ".md") || strings.HasSuffix(file, ".rst") || strings.HasSuffix(file, ".txt") {
			buckets.docs = append(buckets.docs, file)
		}
		if strings.HasPrefix(file, "config/") {
			buckets.configs = append(buckets.configs, file)
		}
		if strings.HasPrefix(file, "resources/views") {
			buckets.views = append(buckets.views, file)
		}
		if file == "composer.json" || file == "composer.lock" {
			buckets.composerFiles = append(buckets.composerFiles, file)
		}
	}

	return buckets, false
}
