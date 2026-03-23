package cli

import (
	"testing"

	"releaser/tool/shared"
)

func TestParseArgs_TypeImpliesForce(t *testing.T) {
	for _, releaseType := range []string{"major", "minor", "patch"} {
		t.Run(releaseType, func(t *testing.T) {
			cfg := &shared.Config{Follow: true}
			if err := ParseArgs(cfg, []string{releaseType}, "releaser"); err != nil {
				t.Fatalf("ParseArgs returned error: %v", err)
			}
			if !cfg.TypeSet {
				t.Fatalf("expected TypeSet=true")
			}
			if cfg.Type != releaseType {
				t.Fatalf("expected type %s, got %s", releaseType, cfg.Type)
			}
			if !cfg.Force {
				t.Fatalf("expected Force=true when type is explicitly set")
			}
		})
	}
}

func TestParseArgs_ForceFlagWithoutType(t *testing.T) {
	cfg := &shared.Config{Follow: true}
	if err := ParseArgs(cfg, []string{"--force"}, "releaser"); err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !cfg.Force {
		t.Fatalf("expected Force=true")
	}
	if cfg.TypeSet {
		t.Fatalf("expected TypeSet=false")
	}
}
