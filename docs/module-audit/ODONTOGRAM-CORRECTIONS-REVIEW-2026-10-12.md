# Odontogram corrections review — 2026-10-12

Reviewed local commit: `d89af48`. Decision: changes requested.

The shared doctor-purpose bypass has been removed, positive update versions are enforced, compatibility calls select legacy charts and lock parents first, and patient search is now server-backed. Remaining findings follow.

## F1 — High: schema corrections exist only in the already-applied migration 060

The new `legacy_provenance` column, legacy-chart unique index and `odontogram_legend` table are added only to 060. The migration runner skips recorded filenames. A database that applied 060 from `9eaa114` receives none of these additions; chart reads now reference a missing column, legend queries fail, and compatibility inserts depend on a missing conflict index.

Add a new corrective migration covering fresh and already-applied-060 databases. Preserve all existing charts/children. Capture available legacy provenance before subsequent edits, without claiming that previously overwritten history can be reconstructed. Test the actual prior 060 schema with its migration marker, populated charts and differing tooth notes, then run the normal upgrade and all relevant operations. Handle any duplicate legacy charts left by the earlier implementation explicitly before creating the unique index.

## F2 — High: default legend seeds reinterpret existing clinical codes

Migration 060 seeds meanings that differ from the established UI definitions still present in `odontogram-register.tsx`: `KR` changes from Root Canal to Radicular Cyst, `Ce` from Pulpitis / Deep Caries to Cervical Caries, `D` from Decayed / Denture to Decalcification, and `PS` from Porcelain / Prosthetic to Post & Core. Existing stored codes therefore acquire different displayed meanings when the database legends load.

Preserve the established code-to-definition contract during this parity work. Any intended clinical terminology change requires an explicit reviewed mapping and historical interpretation strategy; retaining the same code string is not lossless when its meaning changes. Add exact definition assertions, not just code-count checks. Reconcile already-seeded databases without overwriting user-customized definitions indiscriminately.

## F3 — High: legend updates can orphan or reinterpret saved chart findings

`UpdateOdontogramLegend` changes `code`, name, color and description without checking references or maintaining snapshots. After charting `IMP-Z`, renaming that legend to `IMP-Y` leaves stored entries/JSON at `IMP-Z` with no matching legend. Changing its name/meaning also silently changes historical interpretation. Delete protection alone does not cover this path.

Use immutable/versioned chart legend snapshots or reject incompatible edits to referenced definitions. Protect the reference check against concurrent chart saves. Verify edits, deletes and fresh-session exports against existing charts, including legacy provenance references. The PDF currently emits codes alone, so custom-code meaning is not included in that exported report.

## F4 — Medium: PDF text metrics still allow horizontal overflow

The CIDFont dictionary in `httpapi/clinical_care.go` has neither `/W` nor `/DW`; PDF's default CID advance is 1000 units. At 10-point size, a 65-character notes line advances 650 points from x=65 on a 612-point page. The long bilingual heading at 16 points is also emitted as one unwrapped line. Embedding a font and adding pages does not supply its advance widths automatically. The wrapping function also never splits overlong individual words, and names/condition lines are not wrapped.

Emit actual font widths (or use a PDF library that does), wrap every variable-length field using measured page width, and split overlong tokens safely. Independently render long notes, the bilingual heading, long names and 64-character custom codes; verify visible text stays within page bounds and Amharic glyphs render correctly. Current tests check structural markers and page count, not rendered bounds. This finding is based on source inspection; no rendered PDF comparison was performed in this review.

## Scope

Independent verification: `go test ./internal/adapters/postgres -run TestOdontogram_ -count=1` with PostgreSQL enabled passed (exit 0, 33.071s). The findings above identify cases outside those checks. Main was not modified. These corrections should be delivered together, with upgrade regression coverage rather than another in-place-only migration edit. Browser/build claims in the delivery report were not independently rerun in this review.
