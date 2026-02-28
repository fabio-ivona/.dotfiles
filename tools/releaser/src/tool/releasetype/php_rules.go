package releasetype

import (
	"regexp"
	"strconv"
	"strings"

	"releaser/tool/gitops"
	"releaser/tool/output"
	"releaser/tool/shared"
)

var (
	rePublicFunction     = regexp.MustCompile(`^\s*public\s+function\s+([A-Za-z0-9_]+)\s*\((.*)\)`)
	rePublicFunctionName = regexp.MustCompile(`^\s*public\s+function\s+([A-Za-z0-9_]+)\s*\(`)
	reTypeDecl           = regexp.MustCompile(`\b(class|interface|trait|enum)\s+[A-Za-z0-9_]+`)
	reTypeDeclNamed      = regexp.MustCompile(`\b(class|interface|trait|enum)\s+([A-Za-z0-9_]+)`)
	rePublicProperty     = regexp.MustCompile(`public\s+\$[A-Za-z0-9_]+`)
	rePublicConst        = regexp.MustCompile(`public\s+const\s+[A-Za-z0-9_]+`)
	reVisibilityFunction = regexp.MustCompile(`(public|protected|private)\s+function\s+([A-Za-z0-9_]+)`)
)

type phpDiff struct {
	removed []string
	added   []string
}

func analyzePHPFile(cfg *shared.Config, file string, signals *releaseSignals) {
	diffText, _ := gitops.Run(cfg.BaseDir, "diff", cfg.OldTag+"..HEAD", "--", file)
	diff := parsePHPDiff(diffText)
	output.VeryVerbose("PHP diff stats for " + file + ": added=" + strconv.Itoa(len(diff.added)) + " removed=" + strconv.Itoa(len(diff.removed)))

	changedMethods, sameSignatureMethods := detectSignatureChanges(file, diff, signals)
	output.VeryVerbose("Method signature map sizes for " + file + ": changed=" + strconv.Itoa(len(changedMethods)) + " same-signature=" + strconv.Itoa(len(sameSignatureMethods)))

	evaluateAddedAPI(file, diff, changedMethods, sameSignatureMethods, signals)
	evaluateRemovedAPI(file, diff, changedMethods, sameSignatureMethods, signals)
	evaluateControllerRule(file, signals)
}

func parsePHPDiff(raw string) phpDiff {
	d := phpDiff{}
	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "-"):
			d.removed = append(d.removed, strings.TrimPrefix(line, "-"))
		case strings.HasPrefix(line, "+"):
			d.added = append(d.added, strings.TrimPrefix(line, "+"))
		}
	}
	return d
}

func detectSignatureChanges(file string, diff phpDiff, signals *releaseSignals) (map[string]bool, map[string]bool) {
	changedMethods := map[string]bool{}
	sameSignatureMethods := map[string]bool{}
	addedSignatures := collectAddedSignatures(diff.added)

	for _, removedLine := range diff.removed {
		name, oldParams, ok := extractPublicSignature(removedLine)
		if !ok {
			continue
		}

		for _, newParams := range addedSignatures[name] {
			if oldParams != newParams {
				changedMethods[name] = true
				sameSignatureMethods[name] = false
				markMajorForFile(signals, file, formatRuleMessage("changed parameters for "+name, removedLine))
			} else if !changedMethods[name] {
				sameSignatureMethods[name] = true
			}
		}
	}

	return changedMethods, sameSignatureMethods
}

func collectAddedSignatures(lines []string) map[string][]string {
	out := make(map[string][]string)
	for _, line := range lines {
		name, params, ok := extractPublicSignature(line)
		if !ok {
			continue
		}
		out[name] = append(out[name], params)
	}
	return out
}

func extractPublicSignature(line string) (name string, params string, ok bool) {
	matches := rePublicFunction.FindStringSubmatch(line)
	if len(matches) != 3 {
		return "", "", false
	}
	return matches[1], matches[2], true
}

func evaluateAddedAPI(file string, diff phpDiff, changedMethods, sameSignatureMethods map[string]bool, signals *releaseSignals) {
	typeAdded := hasNetTypeAddition(diff)
	if typeAdded {
		output.VeryVerbose("Type declaration added in " + file + "; suppressing added-method entries for this file")
	}
	removedTypes := collectTypeDecls(diff.removed)

	for _, line := range diff.added {
		if reTypeDecl.MatchString(line) {
			kind, name, _ := extractTypeDecl(line)
			if _, exists := removedTypes[typeDeclKey(kind, name)]; exists {
				output.VeryVerbose("Skipping added type declaration for " + kind + " " + name + " (declaration changed in place)")
				continue
			}
			markMinorForFile(signals, file, formatRuleMessage("added "+kind, line))
		}
		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if !typeAdded && !changedMethods[name] && !sameSignatureMethods[name] {
				markMinorForFile(signals, file, formatRuleMessage("added public method", line))
			}
		}
		if rePublicProperty.MatchString(line) {
			markMinorForFile(signals, file, formatRuleMessage("added public property", line))
		}
		if rePublicConst.MatchString(line) {
			markMinorForFile(signals, file, formatRuleMessage("added public constant", line))
		}
	}
}

func evaluateRemovedAPI(file string, diff phpDiff, changedMethods, sameSignatureMethods map[string]bool, signals *releaseSignals) {
	addedVisibility := collectFunctionVisibilities(diff.added)
	typeRemoved := hasNetTypeRemoval(diff)
	if typeRemoved {
		output.VeryVerbose("Type declaration removed in " + file + "; suppressing removed-method entries for this file")
	}
	addedTypes := collectTypeDecls(diff.added)

	for _, line := range diff.removed {
		if reTypeDecl.MatchString(line) {
			kind, name, _ := extractTypeDecl(line)
			if _, exists := addedTypes[typeDeclKey(kind, name)]; exists {
				output.VeryVerbose("Skipping removed type declaration for " + kind + " " + name + " (declaration changed in place)")
				continue
			}
			markMajorForFile(signals, file, formatRuleMessage("removed "+kind, line))
		}

		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if !typeRemoved && !changedMethods[name] && !sameSignatureMethods[name] {
				markMajorForFile(signals, file, formatRuleMessage("removed public method", line))
			}
		}

		if m := reVisibilityFunction.FindStringSubmatch(line); len(m) == 3 {
			oldVisibility := m[1]
			name := m[2]
			for _, newVisibility := range addedVisibility[name] {
				if oldVisibility != newVisibility {
					markMajorForFile(signals, file, formatRuleMessage("visibility changed for "+name, line))
				}
			}
		}

		if rePublicProperty.MatchString(line) {
			markMajorForFile(signals, file, formatRuleMessage("removed public property", line))
		}
		if rePublicConst.MatchString(line) {
			markMajorForFile(signals, file, formatRuleMessage("removed public constant", line))
		}
	}
}

func collectFunctionVisibilities(lines []string) map[string][]string {
	out := make(map[string][]string)
	for _, line := range lines {
		m := reVisibilityFunction.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		visibility := m[1]
		name := m[2]
		out[name] = append(out[name], visibility)
	}
	return out
}

func hasNetTypeAddition(diff phpDiff) bool {
	addedTypes := collectTypeDecls(diff.added)
	removedTypes := collectTypeDecls(diff.removed)
	for key := range addedTypes {
		if _, exists := removedTypes[key]; !exists {
			return true
		}
	}
	return false
}

func hasNetTypeRemoval(diff phpDiff) bool {
	addedTypes := collectTypeDecls(diff.added)
	removedTypes := collectTypeDecls(diff.removed)
	for key := range removedTypes {
		if _, exists := addedTypes[key]; !exists {
			return true
		}
	}
	return false
}

func collectTypeDecls(lines []string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, line := range lines {
		kind, name, ok := extractTypeDecl(line)
		if !ok {
			continue
		}
		out[typeDeclKey(kind, name)] = struct{}{}
	}
	return out
}

func extractTypeDecl(line string) (kind string, name string, ok bool) {
	m := reTypeDeclNamed.FindStringSubmatch(line)
	if len(m) == 3 {
		return m[1], m[2], true
	}
	return "", "", false
}

func typeDeclKey(kind, name string) string {
	return kind + ":" + name
}

func typeKeyword(line string) string {
	kind, _, ok := extractTypeDecl(line)
	if ok {
		return kind
	}
	return "type"
}

func evaluateControllerRule(file string, signals *releaseSignals) {
	if strings.HasPrefix(file, "app/Http/Controllers/") && (signals.major || signals.minor) {
		markMinorForFile(signals, file, "controller change detected")
	}
}

func formatRuleMessage(rule, line string) string {
	return output.PrimaryText(rule) + " | " + output.SecondaryText(compactSnippet(line, 120))
}

func compactSnippet(line string, max int) string {
	if max <= 0 {
		max = 120
	}
	snippet := strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
	if len(snippet) <= max {
		return snippet
	}
	return snippet[:max-3] + "..."
}
