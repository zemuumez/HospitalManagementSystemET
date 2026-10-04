# Detailed review ledger

This ledger distinguishes inventory from completed semantic review. The machine-readable `review/source-index.json` inventories 3,599 source/configuration/asset entries and retains per-file pending status. Selected inspection in this session does not close an entire domain. The original archive is authoritative for excluded bundled/public/storage files.

| Review batch | Inspected so far | Still required before closing the batch |
|---|---|---|
| Specification | All 40 pages of extracted text; rendered layout overview | Resolve role conflicts, SaaS meaning, ambiguous financial and clinical rules with user |
| Identity/access | Kernel, route provider, role seed, selected role groups, module check, CSRF exclusions | All authentication controllers, user model/helpers, permissions seeds, 2FA flow, policies, session and API authorization paths |
| Appointments | Repository, patient API methods, model status definitions, scheduler | Full web/doctor APIs, requests, schedules/breaks/holidays, public booking, Livewire tables, browser scripts, schema constraints, payment coupling |
| Patients/documents/cards | Public QR lookup and beginning of QR view; route/module inventory | All demographic/case/admission relations, CRUD and exports, private file access, card templates, patient identity handling |
| Beds/IPD | Bed controller/repository, IPD bill controller/request/model/repository, selected IPD detail UI | All admission, transfer, discharge, clinical submodules, request rules, migrations, JS bill calculations and patient/API presentation |
| OPD | PDF and route/view inventory | Entire visit/revisit/diagnosis/prescription/timeline flow, charges and patient visibility |
| Prescriptions/pharmacy | Prescription repository and suggestion-entry search | Full prescribing/dispensing/stock/billing paths, dosage semantics, purchasing, reversals, expiry/batches, AI data handling |
| Diagnostics/reports/vaccines | PDF and model/controller/view inventory | Full code, fields, units/results, assignment/release, attachments, report generation and export ownership |
| Blood/inventory | Issued item repository | Blood stock/donations/issues, item receipt and return controllers/requests/constraints, all UI and exports |
| Finance/payments/payroll | IPD bill path and provider inventory | Other bills/invoices/advances, manual approval, every gateway callback/webhook, reconciliation, refunds, payroll, money and currency semantics |
| Communications | SMS repository, appointment reminders, console schedule | Mail templates/attachments, notifications, consultation/meeting access, OAuth/calendar, delivery retries/preferences and provider credentials |
| Public CMS/front office/services | PDF, routes and view inventory | All public content/forms, visitor/postal/call records, insurance/packages, ambulance assignment and exports |
| Extra modules | Dental, queue, complaints, email templates and add-on route discovery | Full behavior and user decision on parity; safe replacement of runtime add-on installation |
| UI/localization | Main layout, selected IPD view, dependency declarations, PDF UI descriptions | Every authored Blade/JS/Sass interaction, all role menus/Livewire tables, locale completeness, runtime visual captures and responsive/RTL behavior |
| Database/migration | Complete inventory; tenant-key searches | Every migration/seeder/factory, resulting schema, indexes/FKs, ownership and legacy data import strategy |
| Dependencies/generated files | Complete archive inventory; Composer/npm manifests | Separate authored code from copied libraries, lockfile/license/security review; literal dependency source audit remains out of completed scope |
| Tests/runtime | Five test/support file inventory; sample feature/unit tests read | Runnable legacy environment with safe configuration and synthetic data, behavior baselines, meaningful regression tests |

## How to close each batch

Read every authored file in the batch and record its hash and review status. For each operation, trace entry point, validation, authorization, use of helpers, persistence, side effects, UI behavior, exports and failure cases. Record business invariants and deviations between PDF and code. Add source-referenced acceptance examples, including negative permissions and concurrent/retried operations where relevant. Runtime observations must be labeled separately from static analysis.

Do not mark a file complete because it appeared in a search result, its function names were indexed, a tool output was truncated, or a neighboring file was reviewed. Distinguish vendored/generated source from authored code without quietly removing it from the user's requested scope.

## Next audit priority

Finish identity and record ownership first, including public QR access and API filter/update paths. Then trace patient/case/admission relationships and appointment availability, followed by bed occupancy/discharge and financial calculation rules. These decisions underpin the schema and security model for the rebuild.
