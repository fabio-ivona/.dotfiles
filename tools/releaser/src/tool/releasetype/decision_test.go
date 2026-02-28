package releasetype

import (
	"testing"

	"releaser/tool/shared"
)

func TestApplyFinalDecision_DocsOnlyPatch(t *testing.T) {
	cfg := &shared.Config{}
	b := changeBuckets{docs: []string{"README.md"}}
	s := newReleaseSignals()

	applyFinalDecision(cfg, b, s)

	if cfg.Type != "patch" {
		t.Fatalf("expected patch, got %s", cfg.Type)
	}
}

func TestApplyFinalDecision_MajorWins(t *testing.T) {
	cfg := &shared.Config{}
	b := changeBuckets{}
	s := newReleaseSignals()
	s.major = true
	s.minor = true

	applyFinalDecision(cfg, b, s)

	if cfg.Type != "major" {
		t.Fatalf("expected major, got %s", cfg.Type)
	}
}

func TestApplyFinalDecision_MinorWhenNoMajor(t *testing.T) {
	cfg := &shared.Config{}
	b := changeBuckets{}
	s := newReleaseSignals()
	s.minor = true

	applyFinalDecision(cfg, b, s)

	if cfg.Type != "minor" {
		t.Fatalf("expected minor, got %s", cfg.Type)
	}
}
