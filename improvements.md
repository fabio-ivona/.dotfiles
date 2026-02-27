# Release Type Detection Improvements

This document enumerates proposed improvements to the release-type detection in `shell/scripts/release.sh`. Each item has a stable number and title so it can be referenced later.

1. Path-Based Priority Rules (Major/Minor/Breaking)
Define a small, explicit set of path-based rules that run before the PHP diff heuristics. These rules should decisively mark major or minor changes based on Laravel conventions and typical API surfaces.
- Major triggers (examples):
  - Deletions/renames under `app/Contracts/**`.
  - Removals of route files under `routes/*.php`.
  - Removals or renames under `app/Http/Middleware/**`.
  - Destructive changes detected in migrations (see item 6).
- Minor triggers (examples):
  - New migrations (non-destructive).
  - New controllers or new routes.
  - Additions under `app/Services/**`, `app/Actions/**`.
  - New config files or new keys in `config/**`.
Goal: more predictable release-type classification before heuristics run.

2. Scope PHP API Detection to Explicit Directories
Limit signature-change heuristics to directories that represent public or semi-public API surfaces in Laravel apps.
- Suggested includes:
  - `app/Contracts/**`
  - `app/Services/**`
  - `app/Actions/**`
  - `app/Support/**`
  - Optional: `app/Http/Controllers/**` (if controllers are treated as external API)
- Suggested excludes:
  - `app/Models/**`, `app/Jobs/**`, `database/**` (except migrations), internal helpers
Goal: reduce false positives when internal code changes should be patch-level.

3. Robust PHP Signature Parsing (Regex Hardening)
Expand and harden method signature detection to handle typical PHP syntax and formatting:
- Support `public static function`, `public function &name`, and multiline signatures.
- Ignore attributes and docblocks, but capture the first signature line reliably.
- Allow return type hints (`): Type`) and parameter typing (`Type $arg`).
Goal: reduce missed signature changes and false positives due to formatting variations.

4. Controller Changes Should Not Imply Minor by Default
Remove the blanket rule that any controller modification implies a minor release.
- Suggested behavior:
  - Controller additions => minor.
  - Controller modifications => patch unless they match API-breaking rules.
Goal: avoid inflated minor releases for typical bug-fix controller edits.

5. Composer Change Messaging Should Match Behavior
The script currently logs that `composer.json`/`composer.lock` changes imply patch, but it does not enforce patch if minor/major rules already triggered.
- Options:
  - Remove or reword the message to avoid implying behavior.
  - Or, explicitly enforce patch only if no higher-priority rule triggers minor/major.
Goal: align logs with actual behavior.

6. Migration Breaking-Change Detection
Detect destructive schema changes in migrations and mark as major.
- Patterns to treat as major:
  - `dropTable`, `dropColumn`, `renameTable`, `renameColumn`
- Patterns to treat as minor:
  - `createTable`, `addColumn`
- Apply only to migration files changed since the last tag.
Goal: catch schema breaking changes as major even if no PHP API signatures changed.

7. Explicit “Only Docs/Tests” Patch Rule Refinement
Refine the "only docs/tests" patch rule to ensure it also covers other non-code change sets that are not intended to bump minor/major.
- Example: treat only `tests/**`, `docs/**`, `README*`, and other documentation changes as patch when no other rule triggers.
- Ensure other non-PHP paths (like CI config or `composer.lock`) do not incorrectly bypass this rule.
Goal: make the patch rule accurate and predictable for non-code changes.
