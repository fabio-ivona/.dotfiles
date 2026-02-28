package releasetype

import (
	"fmt"
	"regexp"
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
	reVisibilityFn       = regexp.MustCompile(`(public|protected|private)\s+function\s+([A-Za-z0-9_]+)`)
)

func analyzePHPFile(cfg *shared.Config, file string, indicators *releaseIndicators) {
	diff, _ := gitops.Run(cfg.BaseDir, "diff", cfg.OldTag+"..HEAD", "--", file)
	removed, added := splitDiffLines(diff)

	changedMethods, sameSignatureMethods := detectMethodSignatureChanges(removed, added, indicators)
	evaluateAddedLines(added, changedMethods, sameSignatureMethods, indicators)
	evaluateRemovedLines(removed, added, changedMethods, sameSignatureMethods, indicators)

	if strings.HasPrefix(file, "app/Http/Controllers/") && (indicators.major || indicators.minor) {
		output.Info("- controller change detected → Minor [" + file + "]")
		indicators.minor = true
	}
}

func splitDiffLines(diff string) ([]string, []string) {
	var removed []string
	var added []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "-") {
			removed = append(removed, strings.TrimPrefix(line, "-"))
		} else if strings.HasPrefix(line, "+") {
			added = append(added, strings.TrimPrefix(line, "+"))
		}
	}
	return removed, added
}

func detectMethodSignatureChanges(removed, added []string, indicators *releaseIndicators) (map[string]bool, map[string]bool) {
	changedMethods := map[string]bool{}
	sameSignatureMethods := map[string]bool{}

	for _, line := range removed {
		matches := rePublicFunction.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		fname := matches[1]
		paramsOld := matches[2]
		reSame := regexp.MustCompile(fmt.Sprintf(`^\s*public\s+function\s+%s\s*\((.*)\)`, regexp.QuoteMeta(fname)))
		for _, a := range added {
			m := reSame.FindStringSubmatch(a)
			if len(m) != 2 {
				continue
			}
			paramsNew := m[1]
			if paramsOld != paramsNew {
				changedMethods[fname] = true
				sameSignatureMethods[fname] = false
				output.Info("- changed parameters for " + fname + " → MAJOR [" + line + "]")
				indicators.major = true
			} else if !changedMethods[fname] {
				sameSignatureMethods[fname] = true
			}
		}
	}

	return changedMethods, sameSignatureMethods
}

func evaluateAddedLines(added []string, changedMethods, sameSignatureMethods map[string]bool, indicators *releaseIndicators) {
	for _, line := range added {
		if reTypeDecl.MatchString(line) {
			output.Info("- added class/trait/interface/enum → Minor [" + line + "]")
			indicators.minor = true
		}
		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			fname := m[1]
			if !changedMethods[fname] && !sameSignatureMethods[fname] {
				output.Info("- added public method → Minor [" + line + "]")
				indicators.minor = true
			}
		}
		if rePublicProperty.MatchString(line) {
			output.Info("- added public property → Minor [" + line + "]")
			indicators.minor = true
		}
		if rePublicConst.MatchString(line) {
			output.Info("- added public constant → Minor [" + line + "]")
			indicators.minor = true
		}
	}
}

func evaluateRemovedLines(removed, added []string, changedMethods, sameSignatureMethods map[string]bool, indicators *releaseIndicators) {
	for _, line := range removed {
		if reTypeDecl.MatchString(line) {
			output.Info("- removed class/trait/interface/enum → MAJOR [" + line + "]")
			indicators.major = true
		}
		if m := rePublicFunctionName.FindStringSubmatch(line); len(m) == 2 {
			fname := m[1]
			if !changedMethods[fname] && !sameSignatureMethods[fname] {
				output.Info("- removed public method → MAJOR [" + line + "]")
				indicators.major = true
			}
		}

		if m := reVisibilityFn.FindStringSubmatch(line); len(m) == 3 {
			vis1 := m[1]
			fname := m[2]
			reSameVisibility := regexp.MustCompile(fmt.Sprintf(`(public|protected|private)\s+function\s+%s`, regexp.QuoteMeta(fname)))
			for _, a := range added {
				if m2 := reSameVisibility.FindStringSubmatch(a); len(m2) == 2 {
					vis2 := m2[1]
					if vis1 != vis2 {
						output.Info("- visibility changed for " + fname + " → MAJOR [" + line + "]")
						indicators.major = true
					}
				}
			}
		}

		if rePublicProperty.MatchString(line) {
			output.Info("- removed public property → MAJOR [" + line + "]")
			indicators.major = true
		}
		if rePublicConst.MatchString(line) {
			output.Info("- removed public constant → MAJOR [" + line + "]")
			indicators.major = true
		}
	}
}
