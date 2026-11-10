---
name: dependency-audit
description: Audit Composer and npm dependency vulnerabilities, automatically using Dock commands when the project is configured with .dock.
---

# Dependency Audit

Use this skill when the user asks to check, audit, or report dependency vulnerabilities in a project that uses Composer and/or npm. Do not use it to upgrade or remediate dependencies unless the user explicitly asks for remediation.

## Select the command environment

Start from the project root. If it does not contain `.dock`, inspect only its immediate parent. When that parent contains `.dock`, run every `dock` audit command from the parent directory. This fallback takes precedence over the normal directory selection.

Otherwise, run commands from the project root and check whether the `.dock` directory exists. Its presence means the project is configured to run dependency tooling through Dock.

| Project configuration | Composer audit | npm audit |
| --- | --- | --- |
| `.dock` exists | `dock composer audit --locked` | `dock npm audit` |
| `.dock` does not exist | `composer audit --locked` | `npm audit` |

When the chosen command directory contains `.dock`, always run both Dock audits: `dock composer audit --locked` and `dock npm audit`. Run the npm audit even when no npm manifest or lockfile is present in that directory. In a non-Dock project, run Composer only when `composer.lock` exists and npm only when `package-lock.json` exists.

## Docker sandbox access

Dock commands need access to the Docker socket. If an audit fails because the sandbox denies that access (for example, `permission denied while trying to connect to the docker API`), retry that same read-only audit command with `sandbox_permissions: "require_escalated"`. Treat Dock's generic `check service "cli" running state: exit status 1` the same way on the first attempt: in restricted environments it can mask an inability to inspect Docker rather than a genuinely stopped service. The approval request must say that it grants Docker-socket access solely to run the specified dependency audit.

Only after the escalated retry returns the same service-state error should it be reported as an operational blocker. Do not start services, build images, or change dependencies for an audit-only request.

## Reporting

Treat a non-zero exit status caused by reported vulnerabilities as an audit finding, not a command-execution failure. Clearly distinguish findings from operational failures such as a missing command, unavailable container, or registry/network error.

Return all findings in a single Markdown table, ordered from highest to lowest severity. Use exactly these columns: `sorgente`, `pacchetto vulnerabile`, `severity`, and `complessità del fix`. Use `composer` or `npm` as the source. Include the advisory identifier or link with the package name when the audit provides one.

Assess fix complexity from the audit output: use `bassa` when a compatible update or an automatic non-breaking fix is available; `media` when constraints or an application-level compatibility check are needed; and `alta` when the only known remediation is a breaking upgrade, package replacement, or the audit reports no available fix. If the audit gives insufficient information, write `da verificare` rather than guessing.

Add a detailed comment for every fix complexity "alta" e "da verificare"

Do not add `--fix`, `--force`, or dependency-update commands to an audit-only request. If an audit returns no vulnerabilities, report the audit commands that passed.
