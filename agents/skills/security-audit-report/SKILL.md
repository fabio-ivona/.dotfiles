---
name: security-audit-report
description: Audit a project's source for security vulnerabilities, authorization flaws and functional bugs, verify findings with isolated evidence, and prepare matching Markdown and PDF technical reports. Use for comprehensive project security reviews and audit reports; not for ordinary bug fixes or dependency-only scans.
---

# Project security audit and report

Produce an evidence-based assessment of the current project as both a complete Markdown report and a shareable PDF. Use the user's language; default to Italian when the request is Italian. Respect an explicit request for a different format or a narrower scope.

## Scope and operating boundaries

- Start from the actual checkout: read applicable project instructions, identify stack, entry points, deployment assumptions, existing changes and available test/runtime tools. Use prior reports only as leads and recheck every retained finding against current source.
- Auditing does not imply fixing the application. Keep application source unchanged unless remediation is requested; write report artifacts and isolated reproductions separately.
- Use synthetic fixtures, a disposable database and local copies for reproductions. Do not connect to configured production databases, probe a public deployment, send messages, stress-test resources or run destructive actions without authorization covering those actions. Read project configuration selectively and redact secrets from all outputs.
- Prefer the project's supported runtime, containers and test commands. A database substitute, CLI harness or mocked frontend proves only the behavior it actually exercises. Record substitutions and their limitations.
- Continue useful source review when a runtime is unavailable. Do not silently claim equivalent verification. Follow environment approval rules for required downloads or writes outside permitted roots.

## Review and evidence

1. Map authentication, roles, trust boundaries, records, files, asynchronous jobs and offline state. Trace user-controlled inputs through validation, permission decisions, persistence and output. Read [review-checks.md](references/review-checks.md) for the relevant review prompts; adapt them to the stack.
2. Review direct endpoint access and lifecycle paths, not just visible UI. For each suspected issue, find a concrete path and required privileges. Distinguish intended shared access from a missing policy; flag the decision that needs confirmation.
3. Reproduce high-impact findings with minimal isolated checks where practical. Include a nearby positive control, such as a rejected unauthorized mutation or accepted legitimate update. Avoid exploits with real credentials, exfiltration or resource exhaustion; demonstrating the missing boundary is enough.
4. Identify bundled and managed dependency versions. Verify relevant vulnerabilities using primary maintainer advisories or official documentation. An affected version is not proof that the application reaches the vulnerable path: inspect configuration, invocation and runtime prerequisites. Do not treat a partial version review as an exhaustive dependency audit.
5. For every candidate maintain: stable ID, type, severity, evidence status, source location, prerequisites, reproduction/observation, impact, correction and regression acceptance criteria. Recheck references after edits or generation. Merge duplicate symptoms with the same cause.

Use these evidence statuses explicitly:

- **Reproduced in isolation:** the tested code path exhibits the reported behavior under documented synthetic conditions.
- **Confirmed in source:** a concrete path is visible but has not been exercised; explain the remaining uncertainty.
- **Conditional risk / policy decision:** impact depends on configuration, intended access, external behavior or an untested prerequisite.
- **Recommendation:** hardening or operational verification without a demonstrated defect.

Severity and evidence are separate. Explain severity through impact, privileges, reachability and user interaction. Do not invent critical findings, require a minimum count, assign unsupported CVSS numbers or claim a full compromise from an authenticated field-level abuse. Report no findings if that is what the evidence supports; state coverage limits.

## Shareable deliverables

Read [report-structure.md](references/report-structure.md) when composing the report. Prioritize reproducible issues and actionable corrections. Include explicit evidence-scoped conclusions for SQL injection, XSS, authorization and anonymous compromise when relevant to the project or the user's question. Absence of a reproduced path is not a guarantee of security.

Write the complete Markdown report first and use it as the content source for the PDF. Both formats must contain the same substantive content: summary, scope, all findings, evidence, limitations, positive controls, references, remediation priorities and appendices. The Markdown is an editable standalone report, not a findings-only extract or a link to the PDF. Presentation may differ, including page numbers in the PDF and heading links in Markdown.

Use the available PDF skill for authoring and its required artifact workflow. If it is unavailable, generate with an appropriate PDF library, extract text and render pages for visual inspection; disclose any unavailable verification. A Markdown file alone does not satisfy a requested PDF.

- Default outputs: `output/pdf/security-audit-YYYY-MM-DD.md` and `output/pdf/security-audit-YYYY-MM-DD.pdf`, with the same basename and actual assessment date, unless the user or project specifies different paths or formats. Do not overwrite an earlier report without intent to replace it.
- Include a project name, assessment date, snapshot identifier or relevant source hashes, and relative file/line references. Prefer content-driven pagination over a fixed page count or fixed findings per page.
- Never include live tokens, PINs, passwords, confidential database contents or irrelevant personal information. Use synthetic actor names and minimal reproduction details.
- Put supporting references near the claims, including clickable primary advisory links and access dates for dependency findings. A shareable report must stand on its own without chat history, memory citations or tool-internal reference IDs.
- Check Markdown structure, code fences, tables and links. Compare the final PDF text against the Markdown and reconcile finding IDs, counts, severities, evidence statuses, references, limitations and stated test outcomes. Keep subsequent content changes synchronized across both formats. Check PDF page references separately. Render and visually inspect every page for clipping, overlaps, unreadable tables and broken glyphs. After revisions, verify changed pages and pagination again.

Deliver links to both the complete Markdown report and the actual PDF with a concise summary of the highest priorities, what was verified, remaining production/browser limitations and whether application source changed. Follow the active artifact citation conventions. Retain useful reproductions when appropriate, but do not present temporary session paths as a portable test suite.
