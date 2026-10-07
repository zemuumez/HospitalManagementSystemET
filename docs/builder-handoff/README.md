# HMS builder handoff — start here

Prepared 2026-10-08 against implementation commit `72c00a7`. This bundle is for an AI builder with the new repository but without the original Laravel project. The exported source evidence is included: do not require the user to supply the ZIP again.

## Mission and boundaries

Finish a single-hospital management system using Next.js/Tailwind, Go clean architecture and PostgreSQL. Preserve the original navigation order, screens, fields, tabs and user journeys before modernizing appearance. Support English and Amharic. Keep Better Auth sessions, a separate Firebase phone verification integration, operational SMS and Mailpit for development. External credentials remain blank until supplied; local fake-provider tests must still work.

The frontend looking similar does not establish backend completeness. A page rendering, an existing table, a route returning 200, or a generic CRUD repository is not proof of workflow parity. Do not invent records after failed requests, use sessionStorage as connected persistence, or report successful saves before the server commits.

## Reading order

1. [Implementation plan](IMPLEMENTATION-PLAN.md): current state, ordered tickets and completion gates.
2. [Product and actor journeys](PRODUCT-AND-ACTORS.md): what the hospital does and who does it.
3. [Workspace reference](WORKSPACES.md): every audited module family, its original obligations and gaps.
4. [Database design](DATABASE-DESIGN.md), then [complete table dictionary](../legacy-schema/TABLES.md).
5. [First ticket: packages and insurance](FIRST-TICKET.md).
6. [Security and acceptance](SECURITY-AND-ACCEPTANCE.md).
7. [Original screen field inventory](WORKSPACE-FIELDS.md), [model rules](../module-audit/MODEL-RULES.md), [route declarations](../module-audit/LEGACY-ACTIONS.md).

## Evidence and authority

**Source-confirmed:** schema-only SQL, table dictionary, migration inventory, model rules and extracted route/workflow evidence. The bundle contains 141 original tables, 1,202 columns, 144 foreign keys, 275 application migrations, 130 model records and 855 route declarations. Resource route declarations are not expanded into individual HTTP operations. The workflow index is a discovery aid, not full controller source or proof that every method was semantically reviewed.

**Reference-only:** screenshots and the read-only live-demo audit. The newer demo includes attendance features absent from the ZIP. Do not label attendance rules as Laravel-source verified. Do not assume the demo and ZIP are the same release.

**Target-tested:** [repair register](../module-audit/REPAIRS.md) and its JSON evidence describe specific tested behavior. Historical audit failures remain in their original reports; subsequent repairs supersede only the named findings.

**Proposed:** the implementation order, safe replacements for legacy weaknesses, role decisions and missing schema/API designs in this handoff. Resolve uncertain business rules in a decision record and test the chosen rule. Do not convert an assumption into an original fact.

This document supersedes the branch/uncommitted-work snapshot in the older `AI-HANDOFF-BACKEND-INTEGRATION-QA.md`. Retain that document's useful technical material but do not redo or revert repairs already committed.

## What is already done

- Original schema, migration, relationship, model-rule, route and module acceptance exports are committed.
- Scoped doctor-absence and inventory-movement registers were repaired; settings proxy and atomic bulk settings writes were repaired; CMS permission advertising was corrected in `bead0c9`.
- Inventory category/item creation and receive/issue now persist and reload server state. Errors stay visible, requests resist duplicate submission, and recipients are active staff IDs. Fake records and local stock mutations were removed in `72c00a7`.
- Latest combined isolated Better Auth/Next/Go/browser evidence passed 55 checks. The full Go suite and eight frontend tests passed at that repair checkpoint. These are previous results, not tests rerun for this documentation delivery.
- Inventory edit/archive, returns/write-offs UI, supplier/store parity and full pagination remain unfinished. Packages/insurance, cards, dental history and encounter child workflows remain substantial gaps.

## How to resume

Inspect `git status`, branch/log and current migrations before changing files. Preserve other contributors' work. Start with FIRST-TICKET.md, implement one complete vertical workflow, test it with real isolated PostgreSQL and browser reload, update the evidence/checklist, and commit that verified step. Push to main only through a reviewed, non-destructive fast-forward or normal merge consistent with the user's existing authorization. Never force-push or overwrite a dirty checkout.

The portable archive includes all documentation/evidence and the source-derived field catalog. It does not contain the Laravel application, database rows, `.env` secrets or a runnable replacement for the new repository.
