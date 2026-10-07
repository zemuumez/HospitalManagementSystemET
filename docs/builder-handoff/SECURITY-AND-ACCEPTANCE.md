# Security and acceptance contract

These are implementation and test requirements, not a claim of regulatory certification or proof that all current code satisfies them. Inspect existing contracts and preserve established protections.

## Trust boundaries

- Better Auth owns authenticated sessions. Go must receive and verify trusted identity through the established integration; never trust arbitrary browser role/user headers.
- Firebase proves control of a phone in the configured project. Verify proof server-side, bind it to the intended account and prevent unauthorized account linking/replay. Operational SMS is a delivery capability, not authentication proof.
- Enforce authorization at HTTP/use-case and sensitive repository boundaries. A Next proxy or hidden button does not secure a directly reachable Go endpoint.
- Separate named permission, role policy, record ownership and state-transition checks. Exercise all of them for list, detail, mutation, print, export and file download.
- Session expiry, logout, account deactivation, role change and MFA changes must invalidate or restrict access as designed. Never preserve stale elevated authorization indefinitely.

## Data and transaction protections

- Validate type, length, enum, range, conditional requirements and referenced-record status server-side. Reject cross-parent child IDs and unexpected writable ownership fields.
- Use parameterized queries, bounded pagination and allowlisted sort/filter fields. Search must not bypass the scope predicate.
- Enforce FK/check/unique constraints alongside application rules. Use transactions/locking/version checks for balances, bed occupancy, scheduling, stock and approvals.
- Stable idempotency keys must prevent duplicate financial/stock/provider effects. A failed response may follow a committed transaction; retries must safely recover the result.
- Audit actor/action/resource/time/outcome without logging credentials, raw tokens or unnecessary clinical payloads. Couple required audit/outbox writes to the business transaction.
- Treat issued financial records and clinically significant history as retained records with explicit correction/reversal/archive policy. Do not blindly reproduce legacy destructive deletes.

## Browser, uploads and integrations

- Follow the established secure cookie, origin/CSRF and CORS contract. Test cross-origin mutations and unauthenticated API calls.
- Escape untrusted content and sanitize any permitted rich HTML. Avoid rendering stored CMS/notes as unrestricted HTML.
- Validate attachment size/type/content, storage key and parent authorization. Keep private media outside public static serving; authorize every download and avoid path traversal or user-controlled filesystem paths.
- Keep credentials in server-side configuration/secret storage, never NEXT_PUBLIC variables or settings responses. Commit only blank examples. Redact provider errors and logs.
- Authenticate provider callbacks, check expected provider/resource/amount/currency and make event processing idempotent. Do not mark payment successful from a browser redirect alone.
- Rate-limit authentication, recovery, phone proof, public enquiry/booking and messaging abuse paths. Apply consent/recipient controls and bounded retries to operational messaging.
- Do not import live patient data into tests. Synthetic records must be isolated and removed by the test harness, not by broad cleanup against shared schemas.

## Per-action acceptance matrix

For each module action create a row: action/route, original evidence, role, resource owner/context, allowed state, input fields, expected response, database effects, audit/outbox effects, UI result, test ID and status.

Run these cases:

1. Allowed actor creates a valid aggregate; verify all parent/child fields and reload in a new browser context.
2. Allowed edit preserves immutable identifiers/history and correctly adds/removes child lines.
3. Missing/invalid/conditional input produces a useful error with no partial writes.
4. Anonymous and each disallowed role are denied. Allowed-role actor targeting an unrelated patient/owner is denied.
5. Guess IDs, child IDs, file IDs and export filters; prove scope is not bypassed.
6. Force a write failure after a parent update; verify rollback of children/balance/audit as appropriate.
7. Repeat request/key, race two requests and simulate network loss; prove no duplicate or lost business effect.
8. Delete/archive/reverse unused and in-use records according to documented policy; verify history remains valid.
9. Test empty data, load failure, mutation failure, loading/duplicate submit, cancellation and pagination/search/sort.
10. Check original field/tab/navigation parity, both languages, accessible labels/focus/errors, print/export/download and currency/date/time presentation.

## Cross-module acceptance journeys

- Receptionist registration → booking → doctor visit → diagnosis/prescription → lab/pharmacy completion → bill/receipt → patient view.
- Case/admission → bed assignment → IPD charges/payment → final bill/discharge → bed available; race a second reservation.
- Package/insurance edit → admission selection → historical price unchanged after catalog update.
- Pharmacy purchase → usable stock → dispense → return/reversal → reconciled stock and money.
- Attendance request/leave → authorized approval → daily report/payroll input where supported; deny self-approval if chosen policy requires separation.
- CMS publish → anonymous website reload; private draft/attachments stay private.
- Outbox event → local Mailpit/SMS fake → retry/failure/recovery without duplicate business mutation.

## Evidence and release gate

Keep historical audit reports immutable and add dated repair evidence. Record commit, environment isolation, commands, counts, failed/skipped tests and cleanup. A skipped external-provider test remains skipped until credentials and a safe account are available; do not represent fake-provider success as live delivery. Use provider-setup.md for the user's credential tasks and the backup/deployment/cutover runbooks for operations.

The release gate includes all module action rows accounted for, no unresolved unauthorized access, no fake connected save, reconciled money/stock, migration/restore rehearsal, monitored worker failures and user acceptance of intentional legacy deviations. “World-class security” is an objective requiring continuing review, not a completion label produced by this checklist.
