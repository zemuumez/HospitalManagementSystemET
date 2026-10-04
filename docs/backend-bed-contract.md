# Bed states and occupancy history

Existing `GET /v1/beds` now includes `state` and `version`. Availability requires an active, ready, unoccupied bed. Administrator `PATCH /v1/beds/{id}` accepts `{state, version, reason}`; states are ready, maintenance and unavailable. Occupied beds cannot be made unavailable. State changes leave immutable reason/actor/version events.

`POST /v1/encounters/{id}/transfer` accepts `{bedId, version, reason}`. Admin/reception and the assigned doctor can transfer active IPD encounters. The service locks the encounter and destination bed, rejects occupied or unavailable destinations, increments the encounter version and records the source/destination and destination tariff. Repeating a stale request returns STATE_CONFLICT, with no second transfer. Admission idempotency continues to compare the original admission bed after transfers.

`GET /v1/encounters/{id}/bed-history?page=1` returns at most 25 events, newest version first, scoped like encounter reads. Admission, transfer and discharge events are retained. Existing records are marked migration baselines; no missing history is invented. History records cannot be updated/deleted. The admission bed and admission tariff remain original snapshots. Transfers do not yet generate bed charges or invoices.

Tests cover maintenance rejection, occupied-bed maintenance rejection, stale state changes, role and assigned-doctor restrictions, concurrent transfers, original-request retries after a transfer, immutable history/original bed, discharge history, strict HTTP input and origin denial. Browser regression covers existing admission/discharge, with new transfer UI deferred to Section 4. Bed-type master data and full original assignment fields remain open.
