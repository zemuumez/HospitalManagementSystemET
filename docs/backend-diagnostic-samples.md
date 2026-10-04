# Diagnostic sample traceability

Migration 026 records each collected sample independently. Existing sample references are backfilled using the original collected-event actor/time where available; unknown historical actor/time remain null rather than being invented.

Lab technicians use the existing `POST /v1/diagnostic-orders/{id}/actions` endpoint:

- `collect` from ordered requires a unique sampleReference and current version.
- `reject_sample` from collected/processing requires a reason. It changes the order to sample_rejected, retaining the sample and a signed rejection record. It is prohibited if any result has already been recorded, including a reviewer-rejected result; use the result amendment workflow instead.
- `recollect` from sample_rejected requires a new sampleReference and current version. It returns the order to collected so processing must be explicitly started again.

Every reference stays globally reserved, including rejected samples. Concurrent recollections serialize on the order; one succeeds and the other receives a stale conflict. A rejected sample cannot be processed. An assigned doctor may cancel a rejected order using the existing reason-required cancel action. Administrators cannot perform clinical sample actions.

`GET /v1/diagnostic-orders/{id}/samples?page=1` returns 25 collections with collector/time, order version and rejection actor/reason/time. It is available to administrators, lab technicians and the assigned doctor. Patients do not receive internal sample-rejection history through this endpoint. Existing released report scope remains unchanged. Collection/rejection rows reject SQL update/delete; transitions and reads are audited.

Tests cover lab-versus-doctor action permissions, rejection/process denial, globally unique old references, concurrent recollection, cross-doctor/patient read denial, preserved rejection reasons, result-before-rejection protection and immutable history. Full Go tests and vet pass. This records authorized staff decisions; it does not automatically judge sample quality or replace laboratory policies.
