---
name: github-cli-workflow
description: Use GitHub CLI for GitHub operations, including issues, pull requests, releases, repositories, workflows, API calls, authentication, and sandbox connectivity diagnosis.
---

# GitHub CLI workflow

Use `gh` for GitHub operations. Confirm the repository context before commands that are not clearly scoped to the current checkout.

- A sandboxed `gh` connection error or apparent invalid token is not conclusive. Verify with `gh api user --jq .login`; if sandbox networking fails, retry once through authorized network access before calling authentication invalid.
- Use the API/CLI response as the source of truth for issues, PRs, releases, workflow runs, repositories, or user state. Do not infer missing remote state from local Git data.
- Treat all mutating operations — creating or editing issues, PRs, releases, labels, workflow dispatches, repository settings, and deletions — as requiring the user’s stated scope. Read-only inspection needs no extra confirmation.
- For Git-over-HTTPS authentication, run `gh auth setup-git`. If an SSH remote needs a missing agent, use a one-command HTTPS remote override for the push rather than changing the saved remote URL.
- For a pull request, inspect branch, working tree, diff, and configured signing before committing. Preserve commit signing unless the user explicitly authorizes disabling it; prefer a repository-local setting when they do. Push before creating the PR, include the actual verification, and return the PR URL.
