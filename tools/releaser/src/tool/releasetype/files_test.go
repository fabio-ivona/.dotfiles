package releasetype

import "testing"

func TestCollectChangedFiles_Empty(t *testing.T) {
	buckets, empty := collectChangedFiles("\n\n")
	if !empty {
		t.Fatalf("expected empty=true")
	}
	if len(buckets.phpFiles) != 0 {
		t.Fatalf("expected no php files")
	}
}

func TestCollectChangedFiles_ExcludesTestsFromPHPHeuristics(t *testing.T) {
	raw := "tests/Unit/FooTest.php\napp/Services/Foo.php\ndocs/readme.md\ndatabase/migrations/2026_01_01_create_users.php\nconfig/app.php\nresources/views/home.blade.php\ncomposer.json\n"

	buckets, empty := collectChangedFiles(raw)
	if empty {
		t.Fatalf("expected empty=false")
	}

	for _, f := range buckets.phpFiles {
		if len(f) >= 6 && f[:6] == "tests/" {
			t.Fatalf("tests file must not be part of php heuristics: %s", f)
		}
	}

	if len(buckets.docs) != 1 {
		t.Fatalf("expected 1 docs file, got %d", len(buckets.docs))
	}
	if len(buckets.migrations) != 1 {
		t.Fatalf("expected 1 migration file, got %d", len(buckets.migrations))
	}
	if len(buckets.configs) != 1 {
		t.Fatalf("expected 1 config file, got %d", len(buckets.configs))
	}
	if len(buckets.views) != 1 {
		t.Fatalf("expected 1 view file, got %d", len(buckets.views))
	}
	if len(buckets.composerFiles) != 1 {
		t.Fatalf("expected 1 composer file, got %d", len(buckets.composerFiles))
	}
}

func TestChangeBuckets_HasOnlyDocs(t *testing.T) {
	if !(changeBuckets{docs: []string{"README.md"}}).hasOnlyDocs() {
		t.Fatalf("expected hasOnlyDocs for docs-only changes")
	}
	if (changeBuckets{phpFiles: []string{"app/Foo.php"}, docs: []string{"README.md"}}).hasOnlyDocs() {
		t.Fatalf("expected hasOnlyDocs=false when php files exist")
	}
}
