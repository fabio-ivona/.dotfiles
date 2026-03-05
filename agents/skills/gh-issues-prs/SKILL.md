---
name: gh-issues-prs
description: Use when working with GitHub issues or pull requests via the gh CLI, especially to list, search, or read issue/PR details based on prompts like "fix #123", "implement issue #45", "review PR #78", or "show details for issue/PR <number>".
---

# GH Issues + PRs

## Overview

Use `gh` to fetch issues and PRs from a GitHub repo: list, search, and read details for a specific number. Assume `gh` is installed and authenticated; warn the user if it is not.

## Quick flow

1. Determine repo context.
   - If the user specifies a repo, use `-R owner/repo`.
   - Otherwise operate in the current repo.
2. Check `gh` availability and auth state; if missing/unauthenticated, warn and tell the user to install/authenticate.
3. Use issue/PR commands to list, search, or read details.

## Commands

### Issues
- List: `gh issue list`
- Search: `gh issue list --search "<query>"`
- Details: `gh issue view <number>`

### PRs
- List: `gh pr list`
- Search: `gh pr list --search "<query>"`
- Details: `gh pr view <number>`

## Prompt mapping

- "implement the feature in issue #123" -> `gh issue view 123`
- "fix #456" -> treat as issue unless user says PR
- "review #78" -> `gh pr view 78`
- "list issues about login" -> `gh issue list --search "login"`
- "find PRs by alice" -> `gh pr list --search "author:alice"`
