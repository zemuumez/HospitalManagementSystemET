# Retained audit review

Migration 019 blocks UPDATE, DELETE and TRUNCATE of `audit_event`. Existing actor foreign keys retain attribution. Events contain actor ID, action, resource ID and timestamp; clinical payloads, passwords and provider secrets are not added to this API.

`GET /v1/audit-events` is administrator-only, with exact optional `actorId`, `action`, `resourceId` filters and a positive `before` event ID cursor. It returns at most 50 events ordered by descending ID, plus `nextBefore` (zero when the page has fewer than 50 events). An empty page is possible after the last full page. Cursor pagination avoids duplicate rows when newer events arrive. Each successful review appends `audit.reviewed` without copying arbitrary filter text into the log. The API uses normal session introspection and no-store responses; no export or deletion route exists.

Database owners/superusers can disable triggers; this protection does not constitute tamper-proof storage against a database administrator. Production must use a separate migration owner, restricted application roles and protected off-host backups/archival. No automatic retention purge is configured: the hospital must approve retention periods and access procedures. The audit review UI remains Section 4 work.

Disposable-schema tests check all eight non-admin roles, exact filters, stable multi-page retrieval, review self-auditing, invalid cursors and rejected SQL update/delete/truncate. Full Go tests and vet pass.
