# Frontend parity audit — 4 October 2026

## Evidence and limits

Compared the five user screenshots, the supplied Laravel Blade/Livewire source, the frontend Sass and static assets, and the current Next.js screens. The public reference `https://hms.infyom.com/dashboard` was attempted through the web reader and Chrome. The reader could not access it; Chrome navigation timed out after 60 seconds. No reference-site records were modified. Consequently, this audit does **not** claim that the deployed demo's nine role sessions were inspected.

The screenshots show Attendance and Manage Attendance. Those entries and an attendance module are absent from the supplied archive's menu/source inventory. They appear to be a deployment/add-on difference. Their new UI is based on the screenshots, not an invented claim that it was ported from the archive.

## Confirmed gaps addressed in this pass

| Gap                                    | Original evidence                                                      | New frontend behavior                                                                                                                                                                                                                                                                   |
| -------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Missing public landing page            | `web/home/index.blade.php`, `web/layouts/header.blade.php`, front Sass | `/` is public, with original orange/navy front-site colors, original hospital/care-team illustrations, hero, appointment strip, four steps, services, doctors, navigation and footer.                                                                                                   |
| Missing public pages                   | `web/home/*`                                                           | About, services, doctors, appointment, working hours, testimonials, contact, terms, privacy and registration previews are reachable. Booking shortcuts carry doctor/date into the appointment form. Public forms clearly indicate they are previews.                                    |
| Missing language selection             | User request, login/general-settings screenshots                       | Persistent English/Amharic selection on public site, login and hospital header. `html.lang` updates; 265 translated labels cover navigation, common controls, public content, settings and attendance. Unknown labels fall back to English; complete clinical translation remains open. |
| Incorrect sidebar structure/order      | `layouts/menu.blade.php`, screenshots 3–5                              | Explicit relevance order: Dashboard, Patient Smart Cards, Users, Odontogram, AddOn, Appointments, Attendance, Manage Attendance, IPD, OPD, Billings, Beds, Blood Banks, etc. IPD/OPD are separate. No alphabetical sorting is used.                                                     |
| Wrong/incomplete tab groupings         | `layouts/sub_menu.blade.php`, screenshots 4–5                          | CMS: CMS, Front CMS Services, Notice Boards, Testimonials, Complaint. Settings: General, Hospital Schedule, Modules Setting, Currencies, Operation Categories, Operations, Payment Gateways, Custom Field, Patient Queue Theme. Diagnosis categories precede tests.                     |
| Simplified CMS                         | `front_settings/*`, screenshot 4                                       | Five second-level CMS tabs, image selection/preview, source-derived home fields, about/appointment/policy/map settings. Saved home text/images update the public preview in this browser.                                                                                               |
| Simplified settings                    | `settings/general.blade.php`, screenshot 5                             | Two-column general form, hospital details, currency, country code, default language, About Us and logo/favicon image controls. Added module-visibility controls and queue-theme preview.                                                                                                |
| Missing attendance                     | Screenshot 3                                                           | Attendance and Manage Attendance screens, sample check-in/out, searchable/filterable striped table, manual add/edit, day/night shifts, calculated worked/late/early/overtime/break values. No payroll or biometric backend implied.                                                     |
| Incomplete smart cards                 | Screenshot 1, patient-card source views                                | Templates with header color and optional fields; generated-card table; card preview with identity, DOB, blood group, address and real QR symbol; PNG download. QR payload identifies synthetic demo data only.                                                                          |
| Missing Remember Me / front-site entry | Screenshot 2                                                           | Remember Me is connected to Better Auth's email sign-in option; login and account menu link to the public site. Reference-demo automatic role login is not reproduced with embedded credentials. Existing authenticated role previews remain available.                                 |

## Independently identified remaining gaps

1. **Dedicated role screens:** `employees`, `employee_prescription_list`, `ipd_patient_list`, `opd_patient_list`, `patients_cases_list`, `patients_prescription_list` and `patient_vaccinated_list` have role-specific source views. The current role selector still filters a shared preview instead of reproducing every dedicated role dashboard/detail/action set. The public demo's role-login comparison remains unverified.
2. **Department administration:** the source `departments` view directory is separate from doctor departments and is not yet represented as a verified standalone workflow.
3. **Two-factor management:** source `two_auth` screens have no equivalent completed management screen. Firebase phone sign-in preparation is not a replacement for that screen.
4. **Clinical conditional behavior:** dependent selects, custom-field injection, revisits, discharge constraints, prescriptions and stock/patient/billing relationships still need individual workflow checks against controllers/JS, beyond generated form fields.
5. **Document outputs:** source PDF/report templates and patient smart-card scan destinations have not all been reproduced. Browser print/CSV and the new sample card PNG are not proof of PDF parity.
6. **Public-site details:** doctor-detail pages, live appointment availability, real patient registration/contact delivery, captcha and production policy content remain pending. Public forms currently simulate submission; no signup/authentication bypass was introduced.
7. **Localization:** untranslated specialized labels, validation/toast messages and local date/currency formatting remain. Amharic wording needs native-speaker review; there was no Amharic language pack in the archive to port.
8. **Preview persistence versus real operation:** settings, attendance and smart cards persist in browser storage only. File selection previews images locally. Production storage, cross-user synchronization, actual QR patient lookup, notifications and operational integrations remain backend work.

## Checks performed

- Eleven public URLs rendered without authentication.
- English → Amharic switching, document language, refresh persistence and switching back passed.
- Email sign-in and Remember Me passed without resetting any credentials.
- Exact expected sidebar prefix and all nine settings-tab labels/order passed.
- Attendance add/save, CMS setting → landing-page update, card display and PNG download passed.
- Public and hospital layouts passed 390px horizontal-overflow checks; desktop, mobile, Amharic, CMS, settings and card screenshots inspected.
- TypeScript and production build checked separately. Route coverage and screenshot review do not establish full original-feature parity.

Next audit priority: dedicated role screens, clinical dependent controls and original print/report layouts. The user should not have to discover those omissions through screenshots.
