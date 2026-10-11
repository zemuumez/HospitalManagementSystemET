# Smart Cards corrections review — 2026-10-11

Reviewed implementation: `a6c81e9939d411014d48a649cfc46c5b5109c7de`.

Decision: changes requested. This is an independent source review, not a rerun of the reported quality gates. No database or browser tests were executed in this review worktree. The reported passing suites do not cover the cases below.

The template-deletion transaction now clears assignments and revokes associated credentials; the requested Amharic keys are present. These corrections do not resolve the following issues.

## R1 — High: enforce patient scope on the card list

`PatientExtensionsService.ListIssuedCards` permits doctors and actors with `patients.read`, while `Store.ListIssuedCards` ignores its actor argument and queries every active card. Its response includes QR bearer tokens and patient demographics. A doctor can retrieve cards belonging to patients assigned to another clinician. Existing patient authorization uses the clinician/portal-owner scope in `postgres/store.go` and `AuthorizePatientRecord`.

Apply that scope to both the count and page queries, or restrict the endpoint to the intended administrative roles. Test two clinicians with different patients and confirm that neither rows, totals, searches nor QR tokens disclose the other clinician's patients. Include patient-role coverage if that role has read permission.

## R2 — High: stale assignment deletion changes the current assignment

`Store.DeleteCardAssignment` reads a card's status but never checks it. After generating card A and reassigning the patient to card B, deleting revoked card A unconditionally clears `patient.smart_card_template_id`, while active card B remains valid. A stale browser tab or replayed delete therefore corrupts the current assignment. It also locks the card before the patient, opposite to generation's patient-then-card order.

Only clear the assignment when deleting the current active credential; make stale deletion a safe no-op or conflict. Use a consistent lock order across issuance, assignment deletion and template deletion. Test A→B→delete A, repeated deletion and concurrent generation/deletion, checking both patient assignment and verification results.

## R3 — High: missing templates expand public QR disclosure

`VerifySmartCardQR` treats every NULL template toggle as allowed (`tEmail == nil || *tEmail`, and similarly for phone, DOB, blood group and address). Pre-migration cards have NULL template IDs, and the still-supported `IssueSmartCard` path continues creating cards without a template ID. Those cards now disclose additional demographics through the public verification route without an explicit template policy.

Define a conservative legacy-card disclosure policy and do not interpret a missing template as permission for every field. Route existing issuance through the authoritative template/assignment lifecycle or explicitly preserve a compatible restricted legacy contract. Test upgraded cards and cards issued through `/v1/patients/{id}/smart-cards`, including public verification and reassignment.

## R4 — Medium: generation is not retry-safe

`GenerateSmartCards` always revokes active credentials and generates random replacements for all/selected modes. The HTTP handler has no generation idempotency contract. Repeating a successful request after a lost response invalidates the credential returned by the first request and creates another card. Patient row locking does not make that replay idempotent.

Add durable request idempotency with payload matching, or an explicit equivalent retry contract that preserves the originally issued credentials. Verify sequential and concurrent retries return the same issuance and keep its QR valid, while a deliberate new reassignment still revokes the old credential.

## R5 — Medium: the QR acceptance test never decodes the QR

Journey 4 in `scripts/verify-smart-cards-workspace.mjs` reads `data-qr-token` and `data-qr-url` and checks for an SVG. It then verifies the extracted token directly. This passes even if the custom QR encoder produces an unreadable matrix, despite the journey being labeled actual QR decoding.

Decode the rendered QR pixels using an independent decoder and verify the decoded URL/token. Exercise the downloaded PNG as well, including enabled/disabled fields, representative long names/addresses and Amharic output. The PNG renderer currently hardcodes English labels and its heading instead of using the selected language, so localize that output too.

## Scope and acceptance

The branch additionally modifies Inventory's frontend and numeric JSON decoding (`inventory-workspace.tsx`, `domain/inventory.go`, and its tests). Identify and separately justify these changes, or remove them from this Smart Cards delivery; the previously accepted Inventory contract should not change unnoticed.

Submit one consolidated correction batch with regression evidence for R1–R5 and rerun the existing PostgreSQL, frontend and isolated browser gates. Keep main and Odontogram unchanged until acceptance.
