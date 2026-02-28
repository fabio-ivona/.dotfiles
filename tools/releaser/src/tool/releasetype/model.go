package releasetype

import (
	"fmt"
	"sort"
	"strings"

	"releaser/tool/output"
)

type changeBuckets struct {
	phpFiles      []string
	migrations    []string
	docs          []string
	configs       []string
	views         []string
	composerFiles []string
}

type releaseSignals struct {
	major       bool
	minor       bool
	globalRules []string
	fileRules   map[string][]string
}

func newReleaseSignals() *releaseSignals {
	return &releaseSignals{
		fileRules: make(map[string][]string),
	}
}

func (s *releaseSignals) addGlobalRule(rule string) {
	if strings.TrimSpace(rule) == "" {
		return
	}
	s.globalRules = append(s.globalRules, rule)
}

func (s *releaseSignals) addFileRule(file, rule string) {
	file = strings.TrimSpace(file)
	rule = strings.TrimSpace(rule)
	if file == "" || rule == "" {
		return
	}
	s.fileRules[file] = append(s.fileRules[file], rule)
}

func (s *releaseSignals) emitRules() {
	for _, rule := range s.globalRules {
		output.Info("- " + rule)
	}

	if len(s.fileRules) == 0 {
		return
	}

	files := make([]string, 0, len(s.fileRules))
	for file := range s.fileRules {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		output.Info(output.AccentText("[" + file + "]"))
		for _, rule := range s.fileRules[file] {
			output.Info("  - " + rule)
		}
	}
}

func (b changeBuckets) hasOnlyDocs() bool {
	return len(b.phpFiles) == 0 && len(b.docs) > 0
}

func (b changeBuckets) docsFiles() []string {
	return append([]string{}, b.docs...)
}

func (b changeBuckets) summary() string {
	return fmt.Sprintf(
		"php=%d migrations=%d docs=%d configs=%d views=%d composer=%d",
		len(b.phpFiles),
		len(b.migrations),
		len(b.docs),
		len(b.configs),
		len(b.views),
		len(b.composerFiles),
	)
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
