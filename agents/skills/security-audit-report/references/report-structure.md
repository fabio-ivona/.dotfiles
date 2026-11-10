# Standalone audit report

Adapt length and layout to the evidence and audience. Keep these information categories visible without forcing a fixed number of sections, pages or findings.

## Decision summary

Identify project, date and assessed snapshot. State the most important consequences and priorities first. Give counts by severity only after final reconciliation, separating recommendations and conditional risks where the counts would otherwise mislead. State whether anonymous compromise or other critical outcomes were actually demonstrated.

## Scope and method

Explain stack, reviewed components, source/runtime access, checks run and unavailable checks. Describe synthetic fixtures, copied code, harness substitutions and database engine. Distinguish a test of a function from an end-to-end API/browser test. Clarify whether source was changed, production was accessed or external actions were performed.

## Findings

For each material finding include:

- Stable identifier, clear title, qualitative severity and evidence status.
- Relative file and line references for the assessed snapshot.
- Root cause and exact attacker prerequisites or triggering conditions.
- A minimal reproducible observation with expected versus observed behavior. Redact secrets; use synthetic actors and records.
- Practical impact and boundaries: what was demonstrated, what was inferred and what still depends on policy or runtime.
- A correction addressing the cause, plus observable regression acceptance criteria.

Use a concise table for navigation when the report is long. Keep speculative hardening separate from findings with concrete evidence. Include confirmed effective controls where they materially limit an attack or invalidate a prior report.

## Dependencies and vulnerability classes

Cite authoritative advisories near each version claim. Give the local version evidence, affected range, relevant invocation/configuration, fixed version information and limits of exploit validation. Do not call a library CVE an application exploit unless its prerequisites are met.

When relevant, state explicit conclusions for SQL injection, XSS, authorization, file access and anonymous takeover. Prefer “no concrete path identified in the inspected source” to “the application is secure.”

## Remediation and evidence appendix

Order work by impact and reachability. Separate source fixes from permission-policy decisions and operational checks. Include the essential runtime/browser/database validations needed to close findings.

Provide a test inventory, meaningful results and a Git commit or hashes of key assessed sources where useful. Hashing a configuration file is not necessary; never print its secrets. Reproduction artifacts intended for handoff should run outside production and document prerequisites. Otherwise describe the procedure within the report instead of relying on ephemeral files.

## Markdown and PDF acceptance

Deliver the full report in Markdown and PDF. Maintain the Markdown as the editable content source and derive the PDF from the same substantive content. Include every finding and all supporting sections in both; do not substitute a shortened findings list for the Markdown report. Keep format-specific navigation appropriate: heading links for Markdown, page references for PDF.

Both documents are readable, self-contained and free of credentials and unnecessary personal data. Validate Markdown headings, fenced code blocks, tables and links. Compare the extracted PDF text with Markdown for completeness and consistency of IDs, severity, evidence status, scope, limitations, references and test outcomes. Check extracted text and every rendered page. Validate finding counts, evidence labels, source lines, link destinations, page references, footers and absence of empty or clipped pages. If layout changes, recheck pagination before delivery.
