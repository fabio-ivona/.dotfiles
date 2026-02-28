package version

import (
	"fmt"
	"strings"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func Bump(cfg *shared.Config) error {
	if !cfg.TypeSet {
		answer := output.Ask(fmt.Sprintf("Please confirm auto-detected release type [major|minor|patch] (default: %s): ", defaultType(cfg.Type)))
		if answer != "" {
			cfg.Type = answer
		}
	} else {
		output.Info("Using provided release type: " + cfg.Type)
	}

	major, minor, patch := parse(cfg.OldVer)
	output.Verbose(fmt.Sprintf("Parsed current version: major=%d minor=%d patch=%d", major, minor, patch))
	switch cfg.Type {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch", "":
		patch++
		cfg.Type = "patch"
	default:
		return fmt.Errorf("Invalid release type: %s", cfg.Type)
	}

	cfg.NewVer = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	if strings.HasPrefix(cfg.OldTag, "v") {
		cfg.NewTag = "v" + cfg.NewVer
	} else {
		cfg.NewTag = cfg.NewVer
	}
	output.Info(fmt.Sprintf("Bumping new %s version from %s to %s", cfg.Type, cfg.OldTag, cfg.NewTag))
	return nil
}

func DefaultYes(ans string) string {
	if ans == "" {
		return "y"
	}
	return ans
}

func defaultType(t string) string {
	if t == "" {
		return "patch"
	}
	return t
}

func parse(v string) (int, int, int) {
	parts := strings.Split(v, ".")
	major, minor, patch := 0, 0, 0
	if len(parts) > 0 {
		major = atoi(parts[0])
	}
	if len(parts) > 1 {
		minor = atoi(parts[1])
	}
	if len(parts) > 2 {
		patch = atoi(parts[2])
	}
	return major, minor, patch
}

func atoi(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}
