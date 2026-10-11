# Smart Cards upgrade review — 2026-10-11

Submitted commit: `0e302c6`. Decision: changes requested for one remaining upgrade issue.

Source inspection confirms that completed request replay no longer falls through to issuance and that the browser retains its submission key after uncertain responses. Migration 058 addresses databases that applied the original 057 without the request table.

## U1 — High: reconcile the intermediate 057 table's cascading foreign key

The published intermediate migration 057 at `b65cc98` created `smart_card_generation_request.template_id REFERENCES patient_id_card_template(id) ON DELETE CASCADE`. Migration 058 uses only `CREATE TABLE IF NOT EXISTS` and index creation. When that table already exists, PostgreSQL leaves its original foreign key intact. Omitting the foreign key from the new CREATE statement does not remove the existing constraint.

Consequently, deleting a template on an upgraded intermediate database deletes its completed generation request records. The same key can then be reused with a different template without the required payload-conflict rejection. This contradicts the durable request-history contract and the claimed replay-after-template-deletion behavior.

The new upgrade test covers the reverted original 057, not the intermediate published 057 that already created the table. Fresh-schema tests cannot detect this discrepancy.

Required correction: add an explicit corrective migration that removes the template foreign key from existing request tables while preserving their rows and uniqueness constraints. Use a new migration if 058 has already been applied. Cover fresh installation, original-057 upgrade, intermediate-057 upgrade, and already-applied-058 upgrade. Seed a completed request before upgrading; delete its template afterward and assert that request identity survives, identical replay does not issue cards, and changed-payload reuse conflicts.

Independent PostgreSQL verification: ran `go test ./internal/adapters/postgres -run 'TestSmartCardsReviewIntermediate057|TestSmartCards_' -count=1` with the database enabled. The submitted suites reported no failures. The additional isolated-schema regression restored the actual table DDL from `b65cc98`, applied 058, and inspected `pg_constraint`; it failed with `058 leaves the intermediate 057 template DELETE CASCADE constraint intact; request history is not durable` (2.70s; combined run exit 1, 27.837s). The temporary regression was removed after recording the evidence.

Main and Odontogram were not modified. Browser and frontend gates were not independently rerun during this review.
