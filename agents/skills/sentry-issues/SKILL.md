---
name: sentry-issues
description: Use when working with Sentry issues, especially to read issue details based on prompts like "fix sentry <PROJECT>-<ID>", "implement sentry issue <PROJECT>-<ID>", "show details for sentry issue <PROJECT>-<ID>".
---

# Sentry Issues

## Overview

Use `sentry` command to read details for a specific issue from Sentry. Assume `sentry` is installed and authenticated; warn the user if it is not.

If you are told to "find the cause" or "inspect" an issue, you should not apply any fix, just debug and find the root causes of the bug and issue a cleare explaination of the causes and offer possible solutions

## Quick flow

1. Check `sentry` availability and auth state; if missing/unauthenticated, warn and tell the user to install/authenticate.
2. Use sentry commands read details.
3. Analyze the issue details and use them to find the root causes of the issue
4. Give a clear explaination of the issue, its causes and possible solutions
5. Unless you are told to "find the cause" or "inspect" an issue, fix it (but only if you are sure about the solution, otherwise ask)

## Commands

If required command fails due to sandbox/network/permission constraints, rerun the same command with escalated permissions (sandbox_permissions=require_escalated) and a short justification. 

### Issues

- Details: `sentry issue view <issue ID>`

## Prompt mapping

- "fix sentry <PROJECT>-<ID>" -> `sentry issue view <PROJECT>-<ID>`
- "fix sentry issue <PROJECT>-<ID>" -> `sentry issue view <PROJECT>-<ID>`
- "find the cause of sentry issue <PROJECT>-<ID>" -> `sentry issue view <PROJECT>-<ID>`
- "inspect sentry issue <PROJECT>-<ID>" -> `sentry issue view <PROJECT>-<ID>`
