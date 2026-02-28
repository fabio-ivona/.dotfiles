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
	output.Verbose("PHP diff stats for " + file + ": added=" + strconv.Itoa(len(diff.added)) + " removed=" + strconv.Itoa(len(diff.removed)))

	changedMethods, sameSignatureMethods := detectSignatureChanges(diff, signals)
	output.Verbose("Method signature map sizes for " + file + ": changed=" + strconv.Itoa(len(changedMethods)) + " same-signature=" + strconv.Itoa(len(sameSignatureMethods)))

	evaluateAddedAPI(diff, changedMethods, sameSignatureMethods, signals)
	evaluateRemovedAPI(diff, changedMethods, sameSignatureMethods, signals)
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

func detectSignatureChanges(diff phpDiff, signals *releaseSignals) (map[string]bool, map[string]bool) {
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
				markMajor(signals, "- changed parameters for "+name+" → MAJOR ["+removedLine+"]")
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

func evaluateAddedAPI(diff phpDiff, changedMethods, sameSignatureMethods map[string]bool, signals *releaseSignals) {
	for _, line := range diff.added {
		if reTypeDecl.MatchString(line) {
			markMinor(signals, "- added class/trait/interface/enum → Minor ["+line+"]")
		}
		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if !changedMethods[name] && !sameSignatureMethods[name] {
				markMinor(signals, "- added public method → Minor ["+line+"]")
			}
		}
		if rePublicProperty.MatchString(line) {
			markMinor(signals, "- added public property → Minor ["+line+"]")
		}
		if rePublicConst.MatchString(line) {
			markMinor(signals, "- added public constant → Minor ["+line+"]")
		}
	}
}

func evaluateRemovedAPI(diff phpDiff, changedMethods, sameSignatureMethods map[string]bool, signals *releaseSignals) {
	addedVisibility := collectFunctionVisibilities(diff.added)

	for _, line := range diff.removed {
		if reTypeDecl.MatchString(line) {
			markMajor(signals, "- removed class/trait/interface/enum → MAJOR ["+line+"]")
		}

		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if !changedMethods[name] && !sameSignatureMethods[name] {
				markMajor(signals, "- removed public method → MAJOR ["+line+"]")
			}
		}

		if m := reVisibilityFunction.FindStringSubmatch(line); len(m) == 3 {
			oldVisibility := m[1]
			name := m[2]
			for _, newVisibility := range addedVisibility[name] {
				if oldVisibility != newVisibility {
					markMajor(signals, "- visibility changed for "+name+" → MAJOR ["+line+"]")
				}
			}
		}

		if rePublicProperty.MatchString(line) {
			markMajor(signals, "- removed public property → MAJOR ["+line+"]")
		}
		if rePublicConst.MatchString(line) {
			markMajor(signals, "- removed public constant → MAJOR ["+line+"]")
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

func evaluateControllerRule(file string, signals *releaseSignals) {
	if strings.HasPrefix(file, "app/Http/Controllers/") && (signals.major || signals.minor) {
		markMinor(signals, "- controller change detected → Minor ["+file+"]")
	}
}
