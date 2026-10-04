# Nursing assignments and vital observations

Source IPD/OPD models store height, weight and a single `bp` field with inconsistent integer/string casts. The rewrite uses explicit centimeter/kilogram/Celsius/mmHg/bpm/percentage fields and paired systolic/diastolic values. Migration of historical BP strings requires a reviewed mapping; this increment does not infer missing readings or clinical meaning.

## Access and routes

- `GET /v1/nursing-encounters?page=1`: authenticated nurse, only encounters with active assignment; 25 per page.
- `GET /v1/encounters/{id}/nurses`: administrator, assigned doctor or actively assigned nurse; at most 100 retained roster entries.
- `POST /v1/encounters/{id}/nurses`: admin or assigned doctor; `{nurseId, active, version, reason}`. New assignments use version zero; edits use the current version. Nurses must have an active nurse role for activation. Inactive staff can be removed. Closed encounters permit removal but not new activation. Every change retains actor/reason/time; roster limit is 100 distinct nurses per encounter.
- `GET /v1/encounters/{id}/vitals?page=1`: admin, assigned doctor or actively assigned nurse; 25 immutable records per page, newest signing time first.
- `POST /v1/encounters/{id}/vitals`: assigned doctor/nurse, active encounter, Idempotency-Key required (16-80 characters). Accepts `{observedAt, measurements, note, correctionOf, correctionReason}`. Measurements use optional `heightCm`, `weightKg`, `temperatureC`, `systolicMmHg`, `diastolicMmHg`, `pulseBpm`, `respiratoryRate` (breaths/minute) and `oxygenPercent`.

At least one finite, nonnegative measurement is required. The technical magnitude ceiling is 100000; percentage cannot exceed 100. These are storage/input checks, not clinical normal ranges or alerts. BP requires both components. Observation time cannot precede admission or be in the future. Note/reason limits are 2000 characters. Each record stores its author and server signing time.

Corrections are new records referencing one observation in the same encounter and require a reason. The original remains readable and immutable. One direct successor per observation prevents competing correction branches; subsequent correction can reference the latest successor. Retries with the same key/data return the same record, changed data conflicts, and authorization is rechecked before returning a retry. Removing nurse access serializes against observation writing on the encounter lock.

No patient/reception/accountant/lab/pharmacy observation access is enabled. Patient release and broader clinical team roles remain pending, along with clinician-defined plausibility rules, alerts, attachments, post-discharge addenda and UI integration.

Verification covers unassigned/cross-encounter denial, scoped nursing lists, revocation, original/retry/correction behavior, correction concurrency, database mutation rejection, closed-encounter denial, finite/boundary inputs and HTTP role/input checks.
