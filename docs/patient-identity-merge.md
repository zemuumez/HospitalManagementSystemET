# Patient identity consolidation

Administrator-only `POST /v1/patients/merge` consolidates identity using `primaryPatientId`, `mergedPatientId` and a required reason. It does not rewrite the patient IDs of clinical, prescription or financial records, delete either MRN, or alter signed history.

Migration 043 stores one-hop aliases. Merge requests serialize with portal-link edits, lock patient rows in deterministic order, and flatten earlier alias groups into the selected primary. Repeating an already completed merge is a no-op. The immutable merge event retains original patient, profile and contact snapshots. Primary demographics remain authoritative; SMS/email consent is reduced to the most restrictive profile consent and data-sharing consent is reset to false for explicit review. The profile change receives a retained revision.

Distinct linked portal owners, active encounters, booked/arrived appointments or active queue entries block a merge with a state conflict. Resolve those relationships explicitly before retrying; the operation never chooses an owner automatically. Routine identity merging cannot override a conflicting account or rewrite past treatment attribution.

`GET /v1/patients/{id}/identities` returns the retained IDs/MRNs and primary identifier to an authorized user. Reading a retired patient's main profile resolves to the primary profile. Patient portal ownership is resolved across the group for historical clinical, invoice, diagnostic, pharmacy, ambulance and consultation access. Patient-specific history lists include retained identity members. New clinical/financial work using a retired patient UUID is rejected by database triggers; clients should use the primary ID returned by profile/identity lookup. Legacy complaint rows that use account IDs rather than patient UUIDs retain their original account-based meaning.

Corrections to a mistaken merge require administrative reconciliation against retained snapshots; there is no casual delete/unmerge endpoint that could silently revoke or disclose another account's history. Treat merge reasons and verified identity evidence as operational requirements. Database restore is disaster recovery, not a routine way to undo one merge.

Database tests cover concurrent repeated merges, preserved encounter/invoice references, portal access to historical invoices, alias profile resolution, one merge event, prevention of new work against retired IDs, active-care and owner conflict rejection. No real patient data was used in these tests.
