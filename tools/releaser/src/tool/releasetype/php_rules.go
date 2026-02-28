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

type signatureLine struct {
	params string
	line   string
}

type visibilityLine struct {
	visibility string
	line       string
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
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			d.removed = append(d.removed, strings.TrimPrefix(line, "-"))
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
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

		for _, addedSig := range addedSignatures[name] {
			if oldParams != addedSig.params {
				changedMethods[name] = true
				sameSignatureMethods[name] = false
				markMajorForFile(signals, file, "changed parameters for "+name, diffPairSnippet(removedLine, addedSig.line))
			} else if !changedMethods[name] {
				sameSignatureMethods[name] = true
			}
		}
	}

	return changedMethods, sameSignatureMethods
}

func collectAddedSignatures(lines []string) map[string][]signatureLine {
	out := make(map[string][]signatureLine)
	for _, line := range lines {
		name, params, ok := extractPublicSignature(line)
		if !ok {
			continue
		}
		out[name] = append(out[name], signatureLine{params: params, line: line})
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

	for i, line := range diff.added {
		if reTypeDecl.MatchString(line) {
			kind, name, _ := extractTypeDecl(line)
			if _, exists := removedTypes[typeDeclKey(kind, name)]; exists {
				output.VeryVerbose("Skipping added type declaration for " + kind + " " + name + " (declaration changed in place)")
				continue
			}
			markMinorForFile(signals, file, "added "+kind, snippetBlock(diff.added, i, 4, "+ "))
		}
		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if !typeAdded && !changedMethods[name] && !sameSignatureMethods[name] {
				markMinorForFile(signals, file, "added public method", snippetBlock(diff.added, i, 4, "+ "))
			}
		}
		if rePublicProperty.MatchString(line) {
			markMinorForFile(signals, file, "added public property", snippetBlock(diff.added, i, 4, "+ "))
		}
		if rePublicConst.MatchString(line) {
			markMinorForFile(signals, file, "added public constant", snippetBlock(diff.added, i, 4, "+ "))
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

	for i, line := range diff.removed {
		if reTypeDecl.MatchString(line) {
			kind, name, _ := extractTypeDecl(line)
			if _, exists := addedTypes[typeDeclKey(kind, name)]; exists {
				output.VeryVerbose("Skipping removed type declaration for " + kind + " " + name + " (declaration changed in place)")
				continue
			}
			markMajorForFile(signals, file, "removed "+kind, snippetBlock(diff.removed, i, 4, "- "))
		}

		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if !typeRemoved && !changedMethods[name] && !sameSignatureMethods[name] {
				markMajorForFile(signals, file, "removed public method", snippetBlock(diff.removed, i, 4, "- "))
			}
		}

		if m := reVisibilityFunction.FindStringSubmatch(line); len(m) == 3 {
			oldVisibility := m[1]
			name := m[2]
			for _, newVisibility := range addedVisibility[name] {
				if oldVisibility != newVisibility.visibility {
					markMajorForFile(signals, file, "visibility changed for "+name, diffPairSnippet(line, newVisibility.line))
				}
			}
		}

		if rePublicProperty.MatchString(line) {
			markMajorForFile(signals, file, "removed public property", snippetBlock(diff.removed, i, 4, "- "))
		}
		if rePublicConst.MatchString(line) {
			markMajorForFile(signals, file, "removed public constant", snippetBlock(diff.removed, i, 4, "- "))
		}
	}
}

func collectFunctionVisibilities(lines []string) map[string][]visibilityLine {
	out := make(map[string][]visibilityLine)
	for _, line := range lines {
		m := reVisibilityFunction.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		out[m[2]] = append(out[m[2]], visibilityLine{visibility: m[1], line: line})
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

func evaluateControllerRule(file string, signals *releaseSignals) {
	if strings.HasPrefix(file, "app/Http/Controllers/") && (signals.major || signals.minor) {
		markMinorForFile(signals, file, "controller change detected", "")
	}
}

func snippetBlock(lines []string, start, maxLines int, prefix string) string {
	if start < 0 || start >= len(lines) {
		return ""
	}
	if maxLines <= 0 {
		maxLines = 4
	}

	end := start + maxLines
	if end > len(lines) {
		end = len(lines)
	}

	var out []string
	for i := start; i < end; i++ {
		out = append(out, prefix+renderSnippetCode(lines[i], 140))
	}
	return strings.Join(out, "\n")
}

func diffPairSnippet(removedLine, addedLine string) string {
	var out []string
	if strings.TrimSpace(removedLine) != "" {
		out = append(out, "- "+renderSnippetCode(removedLine, 140))
	}
	if strings.TrimSpace(addedLine) != "" {
		out = append(out, "+ "+renderSnippetCode(addedLine, 140))
	}
	return strings.Join(out, "\n")
}

func renderSnippetCode(line string, max int) string {
	if max <= 0 {
		max = 140
	}
	code := strings.TrimRight(strings.ReplaceAll(line, "\t", "    "), " \t\r")
	if len(code) <= max {
		return code
	}
	if max <= 3 {
		return code[:max]
	}
	return code[:max-3] + "..."
}
