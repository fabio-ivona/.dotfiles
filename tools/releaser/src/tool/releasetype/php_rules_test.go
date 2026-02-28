package releasetype

import (
	"strings"
	"testing"
)

func TestEvaluateAddedAPI_SuppressesMethodWhenTypeAdded(t *testing.T) {
	file := "app/Foo.php"
	diff := phpDiff{added: []string{
		"class Foo {",
		"public function bar() {}",
	}}
	s := newReleaseSignals()

	evaluateAddedAPI(file, diff, map[string]bool{}, map[string]bool{}, s)

	rules := s.fileRules[file]
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].reason != "added class" {
		t.Fatalf("expected added class reason, got %q", rules[0].reason)
	}
}

func TestEvaluateRemovedAPI_SuppressesMethodWhenTypeRemoved(t *testing.T) {
	file := "app/Foo.php"
	diff := phpDiff{removed: []string{
		"class Foo {",
		"public function bar() {}",
	}}
	s := newReleaseSignals()

	evaluateRemovedAPI(file, diff, map[string]bool{}, map[string]bool{}, s)

	rules := s.fileRules[file]
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].reason != "removed class" {
		t.Fatalf("expected removed class reason, got %q", rules[0].reason)
	}
}

func TestTypeDeclarationChangedInPlace_NoAddRemoveTypeFinding(t *testing.T) {
	file := "app/Page.php"
	diff := phpDiff{
		added:   []string{"class Page extends NewBase {"},
		removed: []string{"class Page extends OldBase {"},
	}
	s := newReleaseSignals()

	evaluateAddedAPI(file, diff, map[string]bool{}, map[string]bool{}, s)
	evaluateRemovedAPI(file, diff, map[string]bool{}, map[string]bool{}, s)

	if len(s.fileRules[file]) != 0 {
		t.Fatalf("expected no add/remove type findings for in-place type declaration change")
	}
}

func TestDetectSignatureChanges_ReportsPairSnippet(t *testing.T) {
	file := "app/Foo.php"
	diff := phpDiff{
		removed: []string{"public function run($a)"},
		added:   []string{"public function run($a, $b = null)"},
	}
	s := newReleaseSignals()

	detectSignatureChanges(file, diff, s)

	if !s.major {
		t.Fatalf("expected major signal")
	}
	rules := s.fileRules[file]
	if len(rules) != 1 {
		t.Fatalf("expected exactly 1 file rule, got %d", len(rules))
	}
	if !strings.Contains(rules[0].snippet, "- public function run($a)") {
		t.Fatalf("expected removed signature in snippet, got %q", rules[0].snippet)
	}
	if !strings.Contains(rules[0].snippet, "+ public function run($a, $b = null)") {
		t.Fatalf("expected added signature in snippet, got %q", rules[0].snippet)
	}
}

func TestRenderSnippetCode_PreservesIndentation(t *testing.T) {
	line := "    public function up(): void   "
	got := renderSnippetCode(line, 120)
	if !strings.HasPrefix(got, "    public function") {
		t.Fatalf("expected leading indentation to be preserved, got %q", got)
	}
	if strings.HasSuffix(got, " ") {
		t.Fatalf("expected trailing spaces to be trimmed, got %q", got)
	}
}

func TestSnippetBlock_UpToFourLines(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e"}
	got := snippetBlock(lines, 0, 4, "+ ")
	parts := strings.Split(got, "\n")
	if len(parts) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(parts))
	}
}
