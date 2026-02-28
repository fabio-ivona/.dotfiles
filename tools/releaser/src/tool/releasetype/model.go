package releasetype

import "strings"

type changeBuckets struct {
	phpFiles      []string
	migrations    []string
	tests         []string
	docs          []string
	configs       []string
	views         []string
	composerFiles []string
}

type releaseSignals struct {
	major bool
	minor bool
}

func (b changeBuckets) hasOnlyDocsOrTests() bool {
	return len(b.phpFiles) == 0 && (len(b.tests) > 0 || len(b.docs) > 0)
}

func (b changeBuckets) docsAndTests() []string {
	all := append([]string{}, b.tests...)
	all = append(all, b.docs...)
	return all
}

func trimNonEmptyLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
