package cli

import (
	"fmt"
	"path/filepath"

	"releaser/tool/output"
	"releaser/tool/shared"
)

func Usage(bin string) {
	fmt.Printf("Usage: %s [major|minor|patch] [--force] [--verbose]\n\n", filepath.Base(bin))
	fmt.Println("Arguments:")
	fmt.Println("  major|minor|patch   Optional release type. If omitted, it will be detected")
	fmt.Println("                      from git diff (like your Laravel command). When provided,")
	fmt.Println("                      the confirmation prompt is skipped.")
	fmt.Println("Options:")
	fmt.Println("  --force             Don't ask confirmation before creating the tag.")
	fmt.Println("  --verbose           Print debug details about each release step.")
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
		case "--verbose":
			cfg.Verbose = true
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
