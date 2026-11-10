# Review prompts

Use the parts that match the project's architecture. Track what was reviewed, tested and left outside scope; a checklist is not evidence of a vulnerability or full coverage.

## Identity, permissions and confidentiality

- Trace login, registration, invitation issuance/consumption, password or PIN reset, provisional credentials, session renewal and logout. Check expiry, account disablement, privilege changes, revocation and throttling. Consider whether account-level lockout enables denial of legitimate login.
- Compare UI restrictions with server authorization across APIs, policies, resources, actions, exports, background jobs and file endpoints. Test ordinary, elevated and external/tenant actors as relevant.
- Examine both creation and update: immutable author/tenant/owner fields, allowlists, mass assignment, field-level permissions, deletion and bulk operations. Do not accept client-supplied ownership or delegation as authority.
- Check read projections, full and incremental sync, hidden fields, stale caches and downloaded files. Determine intended visibility before labeling broad shared reads as unauthorized.
- Associate attachments with visible records. Verify path construction, filename validation, MIME checks, content disposition and cache behavior. Random names prevent easy enumeration but do not establish record authorization. Configuration files such as .htaccess require runtime verification before claiming effective web protection.

## Integrity and lifecycle bugs

- Trace create/update/delete, imports, corrections, restore and queued processing. Check closed accounting periods, approvals and confirmations on the server, including date changes across boundaries.
- Inspect database constraints, transactions, optimistic revisions, missing revision fields and time-of-check/time-of-use gaps. Evaluate concurrency on the actual database engine where needed.
- For offline sync, reconcile each request with explicit per-operation acknowledgement. Look for silent continue, batch slicing, partial failure, latest-write overwrites, cursor/snapshot races, deletion tombstones and retries that resurrect data.
- Check identity and permission changes while requests are in flight; ensure cache and queued writes are scoped to the correct actor. Test logout followed by late responses and downgraded roles with unchanged records.
- Validate nested types, allowed enums, real dates, time order, numeric ranges, foreign references and schema limits. Reject excessive input explicitly instead of silently truncating business data.
- Distinguish timestamp storage, presentation timezone, day boundaries and date-only business fields. Use project conventions; do not impose a timezone solely because a prior project used it.

## Injection and resources

- Inventory dynamic SQL and parameter binding. Trace untrusted values into SQL identifiers as well as values; fixed SQL or correctly bound values are not SQL injection.
- Inspect HTML, attributes, URL schemes, CSS, inline event handlers, template execution, shell commands, filesystem paths, deserialization and server-side URL fetching. Escaping is context-specific; sanitization or a suspicious sink alone is not a complete exploit proof.
- Review exported CSV and spreadsheet strings for formula injection, separately from HTML and SQL escaping. Generating a dangerous cell does not prove a specific spreadsheet program executes it.
- Check whether files or other external side effects occur before authorization or survive failed transactions. Use a small harmless fixture to prove an orphan write; do not fill a disk.
- Assess cumulative quotas, image processing, archives, payload size, unrestricted registration and expensive queries. Do not report every missing limit as exploitable denial of service without a concrete resource path.

## Components and operations

- Inventory manifest/lockfile dependencies and identifiable bundled libraries. For CVEs verify version, maintainer advisory, patch, configuration and reachability. Distinguish rendering, extraction and other library paths.
- Review installer exposure, update/rollback authorization and atomicity, secrets, debug output, headers, service-worker caching and public artifacts. Do not assume a deployment uses Apache just because an .htaccess exists.
- Review backup consistency, partial writes, retention, attachment integrity and restore prerequisites. Source code cannot establish that a cron runs, off-site copies exist or recovery works. Frame missing operational evidence as a verification item.
- Run syntax/static checks and focused tests appropriate to the code. Tests passing do not prove production configuration, browser behavior, external integrations or absence of vulnerabilities.
