# Releaser Improvements Backlog

This file lists potential improvements for `tools/releaser`.
Each item is independent and can be referenced by its section number.

## 1. New Features

### 1.1 Path-Based Release-Type Priority Rules
- Goal: Make major/minor decisions more predictable before deep PHP heuristics.
- Scope:
  - Add explicit path-based rules that run first.
  - Major examples: deletions/renames in `app/Contracts/**`, route removals, destructive migrations.
  - Minor examples: new migrations (non-destructive), new controllers/routes, additions in service/action layers.
- Benefit: Clearer and less surprising release classification.

### 1.2 Migration Breaking-Change Detection
- Goal: Mark destructive schema changes as major.
- Scope:
  - Detect patterns like drop/rename table/column.
  - Treat additive schema changes as minor.
- Benefit: Better semantic versioning for DB-facing changes.

### 1.3 Add `--dry-run` Mode
- Goal: Show planned actions without mutating git or GitHub.
- Scope:
  - Add `--dry-run` flag.
  - Show old/new version, tag actions, release payload summary.
  - Skip `git tag`, `git push`, and GitHub `POST`.
- Benefit: Safe preflight validation.

### 1.4 Add `--non-interactive` Mode
- Goal: Support CI use without prompts.
- Scope:
  - Add `--non-interactive` flag.
  - Disable all `Ask(...)` prompts.
  - Fail early when required input is missing.
- Benefit: Reliable automation.

### 1.5 Add `--base-dir` CLI Option
- Goal: Explicitly control repository path.
- Scope:
  - Parse and validate `--base-dir <path>`.
  - Keep autodetect as fallback.
- Benefit: Better scripting and fewer path surprises.

### 1.6 Add Custom Release Notes Overrides
- Goal: Let users provide release notes directly.
- Scope:
  - Add `--notes` and/or `--notes-file`.
  - Bypass auto-generated changelog when provided.
- Benefit: Better editorial control.

### 1.7 Add `--draft` and `--prerelease` Flags
- Goal: Support staging and prerelease workflows.
- Scope:
  - Pass flags directly to GitHub release payload.
- Benefit: Wider workflow compatibility.

### 1.8 Add Range Control (`--from-tag` / `--to-ref`)
- Goal: Build notes and type detection over explicit refs.
- Scope:
  - Default remains latest release tag to `HEAD`.
  - Allow overriding source/target refs.
- Benefit: Better for hotfix/backport branches.

### 1.9 Add Safeguard for Existing GitHub Release by Tag
- Goal: Make re-runs idempotent at release creation step.
- Scope:
  - Check release existence by tag before `POST`.
  - Optionally skip or update existing release.
- Benefit: Prevent duplicate-run failures.

### 1.10 Add Optional JSON Output Mode
- Goal: Produce machine-readable results for CI/tooling.
- Scope:
  - Add `--json` output option with structured fields.
- Benefit: Easier integration with automation.

### 1.11 Add Config File Support (`releaser.yaml`)
- Goal: Centralize project-specific defaults.
- Scope:
  - Optional config file with defaults for rules, prompts, output mode, token sources.
- Benefit: Repeatable behavior across teams/repos.

## 2. Code

### 2.1 Scope PHP API Detection to Explicit Directories
- Goal: Reduce false positives by only scanning API-relevant PHP surfaces.
- Scope:
  - Include configurable directories (e.g. `app/Contracts/**`, `app/Services/**`, `app/Actions/**`).
  - Exclude noisy internal-only paths by default.
- Benefit: Better signal-to-noise for major/minor detection.

### 2.2 Harden PHP Signature Parsing
- Goal: Correctly detect signature changes in real-world PHP syntax.
- Scope:
  - Support `public static function`, references, typed params/returns, multiline signatures.
  - Ignore docblocks/attributes while preserving signature semantics.
- Benefit: Fewer missed breaking changes and fewer false positives.

### 2.3 Controller Rule Refinement
- Goal: Avoid over-classifying controller edits as minor.
- Scope:
  - Controller additions => minor.
  - Controller modifications => patch unless another rule marks major/minor.
- Benefit: More accurate version bumps for common bugfixes.

### 2.4 Refine “Only Docs/Tests” Patch Rule
- Goal: Make non-code patch classification explicit and predictable.
- Scope:
  - Cover docs/tests/readme-like changes clearly.
  - Avoid accidental patch fallback for unrelated non-PHP files unless intended.
- Benefit: More deterministic patch decisions.

### 2.5 Improve GitHub API Robustness
- Goal: Provide clearer failures and avoid hangs.
- Scope:
  - Add HTTP timeout.
  - Parse GitHub error JSON fields for better messages.
  - Surface rate-limit context.
- Benefit: Easier diagnosis and higher reliability.

### 2.6 Add Retry Policy for Transient GitHub Failures
- Goal: Handle flaky network/API conditions gracefully.
- Scope:
  - Retry safe `GET` requests with backoff.
  - Keep release `POST` idempotent-aware.
- Benefit: Fewer failed release runs due to temporary issues.

### 2.7 Add Structured Error Types
- Goal: Normalize failure handling.
- Scope:
  - Introduce typed errors for expected states (missing token, uncommitted changes, diverged branch, etc.).
  - Centralize user-facing guidance mapping.
- Benefit: Cleaner and more maintainable error handling.

## 3. Testing

### 3.1 Add Unit Tests for Core Deterministic Logic
- Goal: Protect parsing and versioning behavior.
- Scope:
  - Test CLI arg parsing.
  - Test version parse/bump and prompt-validation helpers.
- Benefit: Reduced regression risk.

### 3.2 Add Integration Tests with Temporary Git Repos
- Goal: Verify end-to-end git flow behavior.
- Scope:
  - Test clean tree checks, tag existence, ahead/behind handling, no-upstream branch.
- Benefit: Confidence before real releases.

## 4. Style and UX

### 4.1 Align Composer Messaging with Actual Behavior
- Goal: Ensure logs match decision logic.
- Scope:
  - Reword or adjust `composer.json/lock changed -> patch` message.
  - Only claim patch when no higher-priority rule triggered.
- Benefit: Output that users can trust.

### 4.2 Add Verbosity Levels (`--quiet`, `--verbose`)
- Goal: Control output detail.
- Scope:
  - Quiet mode for minimal logs.
  - Verbose mode for command/API trace summaries.
- Benefit: Better UX for both humans and CI logs.

### 4.3 Improve Prompt UX and Validation
- Goal: Make prompts cleaner and safer.
- Scope:
  - Standardize spacing around prompts.
  - Re-prompt on invalid release type input.
- Benefit: Fewer accidental user errors.

### 4.4 Add Release Summary Before Mutations
- Goal: Show a final checkpoint before state changes.
- Scope:
  - Print repo/base dir/old/new version/type/tag summary.
  - Confirm once unless forced/non-interactive.
- Benefit: Reduces operator mistakes.

### 4.5 Improve Message Wording and Consistency
- Goal: Make output clearer and professional.
- Scope:
  - Fix typos and awkward messages.
  - Standardize capitalization and punctuation.
- Benefit: Better day-to-day UX.
