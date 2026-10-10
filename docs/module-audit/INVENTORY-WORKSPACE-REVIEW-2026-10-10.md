# Inventory workspace independent review

Reviewed branch: `feat/inventory-workspace-parity`, HEAD `8813c2962d487d6ea11732f5acacb59fbf9a0407`.
Verdict: **changes required before acceptance**. Settings integration is confirmed: `origin/main` is `5cb7f81`. This review does not modify main or the Inventory implementation.

## Independent verification

Full Go/PostgreSQL tests with `-count=1`, TypeScript typecheck, all 8 frontend tests, and formatting passed. All 19 isolated browser journeys passed with exit code 0, and isolated-schema cleanup succeeded. Findings below are based on code inspection and, for source parity, direct inspection of the supplied original HMS ZIP. They are not claims that the existing happy-path tests failed.

The backend already provides item-row serialization, transactionally coupled balance/movement/audit writes, idempotency, bounded cumulative returns, and non-restocked returns. Preserve these safeguards while addressing the following batch.

## I1 — Registers and selectors silently stop at 25 records (high)

`fetchInventoryData` in `apps/web/src/components/inventory-workspace.tsx:205` requests categories, items, and movements once without a page argument. All three repository list methods in `services/api/internal/adapters/postgres/inventory.go` enforce `LIMIT 25`. The component has no pagination or subsequent-page loading. Search, category filters, status filters, counts, item lookups, and movement-kind filtering operate only on those first pages.

Consequences: item/category 26 cannot be selected or found through client search; an older unreturned issue disappears once 25 newer movements exist, preventing return through the UI; receipts can disappear from the stocks tab when newer issues/returns occupy the combined movement page. Movement rows can also lose their item labels when the item is outside the first item page.

Required: implement server-backed pagination/search/filtering or a complete, explicitly bounded retrieval strategy that does not silently truncate. Fetch related display data independently of the visible item page. Counters must communicate actual totals or clearly labeled page counts. Add fixtures exceeding 25 categories, items, and mixed movement kinds, including an old outstanding issue that remains discoverable and returnable.

## I2 — Empty-category submission silently moves another category's stock (high)

`handleSaveStockReceive` and `handleSaveIssue` at component lines 396 and 420 fall back to `items[0]?.id` when both the selected item and filtered category items are empty. Their item selectors (around lines 1543 and 1760) are not required and contain no options for an empty category. `selectedIssueItem` uses the same global fallback.

Concrete code path: create category A with an item and category B without items; open Receive or Issue and select B. The item selector is empty, yet completing the other required fields and submitting sends A's item ID. The API correctly processes the submitted item, so the unintended stock change succeeds.

Required: validate an explicit item belonging to the selected category, clear invalid child selections when the category changes, disable submission when no eligible item exists, and remove the global fallback. Test both receive and issue for empty categories and stale child selections, asserting no request/movement/balance change.

## I3 — Source fields and correction actions are missing despite COMPLETE claims

Directly inspected source files:

- `hms/resources/views/issued_items/fields.blade.php`: department/user-type selection, recipient, issued-by, issued date, and return date are present. The delivered issue contract has recipient and reason but no business issue/due-return dates or issued-by field. `createdAt` is an audit timestamp, not a replacement for a user-entered business date. The UI displays raw recipient IDs rather than resolving the selected staff member's name.
- `hms/resources/views/item_stocks/fields.blade.php` and `edit_fields.blade.php`: receipt attachment upload is present. The delivered receipt contract and modal have no attachment field or lifecycle.
- `hms/resources/views/item_stocks/action.blade.php`: edit and delete actions are present. The new stocks tab only provides details. `hms/resources/views/issued_items/action.blade.php` also provides deletion; there is no corresponding authorized correction/void workflow in the delivered issue register.

Required: complete the field/action matrix and implement the missing source workflows without destroying audit history. An audited amendment/reversal/void operation is an acceptable way to retain ledger integrity while providing the original correction capability. Validate stock effects transactionally, reject unsafe corrections after consumption/returns, and preserve prior history. Wire receipt attachments through an authorized nonclinical storage lifecycle with appropriate access checks. Resolve recipient display names while retaining stable IDs. Do not invent standalone supplier/procurement modules: the inspected source uses supplier/store metadata.

Correct unsupported documentation claims: the movement schema has `store_name`, not `store`; issue/return UI payloads currently omit store metadata; lot and supplier-contact tracking are not represented by the inspected movement contract; insufficient balance returns `ErrStale`/409, not the documented `ErrInsufficient`/422. Mark any deliberately excluded source action explicitly instead of declaring complete parity.

## I4 — Editing rounds away a valid fractional reorder threshold

`openEditItem` at component line 387 applies `Math.round(item.reorderMilli / 1000)`, then the save handler multiplies the rounded value by 1000. A valid stored threshold of 1500 becomes 2000 even when the user edits only the description. Zero is accepted by the API but the input uses `min={1}`, preventing an unchanged zero-threshold item from being saved through the form.

Required: preserve the exact supported milli-unit precision, allow the backend-supported zero threshold, and align numeric controls with the documented unit contract. Test an unrelated edit of items with zero and fractional thresholds and assert the original values remain unchanged.

## I5 — Theme/localization claims exceed the actual UI and assertions

Journey 19 adds `dark` to `document.documentElement` and takes a screenshot, but the application theme is controlled by the shell's `legacy-dark` class and `hms-theme` setting. The test neither uses the real theme toggle nor asserts computed contrast. The supplied "dark mode" screenshot still shows a light workspace. The test checks a few Amharic labels, not all controls; the supplied Amharic screenshot still displays New Item, Sync, All Categories, Item Name, Reorder Level, and In Stock in English.

Required: use the real theme toggle, assert the shell enters the intended mode, verify representative table/form/modal colors, and capture fully loaded screenshots. Complete translations for the actual new controls and status messages, and cover each tab and its dialogs. Reuse canonical workspace navigation instead of duplicating the shell submenu. Update the report to match the actual executed journey names, quantities, and evidence scope.

## One correction batch

Address I1–I5 together with logical commits and targeted regressions. Retain isolated-schema execution and the accepted Settings worker-test isolation fix. Rerun the browser suite, full Go/PostgreSQL suite, typecheck, frontend tests, and formatting. Return the branch HEAD, corrected source matrix, and screenshots for independent acceptance. Keep Inventory off main; Smart Cards and Odontogram remain out of this batch.
