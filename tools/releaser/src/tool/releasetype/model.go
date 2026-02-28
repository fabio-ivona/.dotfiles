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
	fileRules   map[string][]fileRule
}

type fileRule struct {
	severity string
	reason   string
	snippet  string
}

func newReleaseSignals() *releaseSignals {
	return &releaseSignals{
		fileRules: make(map[string][]fileRule),
	}
}

func (s *releaseSignals) addGlobalRule(rule string) {
	if strings.TrimSpace(rule) == "" {
		return
	}
	s.globalRules = append(s.globalRules, rule)
}

func (s *releaseSignals) addFileRule(file, severity, reason, snippet string) {
	file = strings.TrimSpace(file)
	reason = strings.TrimSpace(reason)
	if file == "" || reason == "" {
		return
	}
	s.fileRules[file] = append(s.fileRules[file], fileRule{
		severity: strings.TrimSpace(severity),
		reason:   reason,
		snippet:  strings.TrimSpace(snippet),
	})
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
		if output.VerbosityLevel() >= 1 {
			renderBoxedFileRules(file, s.fileRules[file])
			continue
		}

		output.Info(output.AccentText("[" + file + "]"))
		for _, rule := range s.fileRules[file] {
			output.Continue("  - " + output.SemverLabel(rule.severity) + " | " + output.PrimaryText(rule.reason))
		}
	}
}

func renderBoxedFileRules(file string, rules []fileRule) {
	width := len(file) + 4
	if width < 32 {
		width = 32
	}
	if width > 72 {
		width = 72
	}

	output.Info(output.AccentText("┌─ " + file))
	for _, rule := range rules {
		border := output.AccentText("│")
		output.Continue(border)
		output.Continue(border + "  " + output.SemverLabel(rule.severity) + " - " + output.PrimaryText(rule.reason))
		if rule.snippet != "" {
			for _, line := range strings.Split(rule.snippet, "\n") {
				output.Continue(border + "    " + output.SecondaryText(line))
			}
		}
	}
	output.Continue(output.AccentText("└" + strings.Repeat("─", width)))
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
