package cli

import (
	"fmt"
	"path/filepath"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func Usage(bin string) {
	fmt.Printf("Usage: %s [major|minor|patch] [--force] [--no-follow] [--verbose|-v|-vv]\n\n", filepath.Base(bin))
	fmt.Println("Arguments:")
	fmt.Println("  major|minor|patch   Optional release type. If omitted, it will be detected")
	fmt.Println("                      from git diff (like your Laravel command). When provided,")
	fmt.Println("                      the confirmation prompt is skipped.")
	fmt.Println("Options:")
	fmt.Println("  --force             Don't ask confirmation before creating the tag.")
	fmt.Println("  --no-follow         Don't check the GitHub Actions workflow after publishing.")
	fmt.Println("  --verbose, -v       Enable verbose output.")
	fmt.Println("  -vv                 Enable very verbose output (trace-level).")
}

func ParseArgs(cfg *shared.Config, args []string, bin string) error {
	for len(args) > 0 {
		switch args[0] {
		case "major", "minor", "patch":
			cfg.Type = args[0]
			cfg.TypeSet = true
			args = args[1:]
		case "--force":
			cfg.Force = true
			args = args[1:]
		case "--no-follow":
			cfg.Follow = false
			args = args[1:]
		case "--verbose":
			if cfg.Verbosity < 1 {
				cfg.Verbosity = 1
			}
			args = args[1:]
		case "-v":
			if cfg.Verbosity < 2 {
				cfg.Verbosity++
			}
			args = args[1:]
		case "-vv":
			cfg.Verbosity = 2
			args = args[1:]
		case "-h", "--help":
			Usage(bin)
			output.Exit(0)
		default:
			return fmt.Errorf("Unknown argument: %s", args[0])
		}
	}
	return nil
}

func IsValidType(t string) bool {
	switch t {
	case "major", "minor", "patch":
		return true
	default:
		return false
	}
}
