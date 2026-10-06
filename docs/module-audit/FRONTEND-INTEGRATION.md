# Frontend integration source register

All inspected component/hospital-page files. A fetch count does not establish working integration; local storage for theme/language is harmless. Signals require action-level review, and helper/transitive API calls may live in imported .ts files.

| File | Direct API call sites | Explicit mutation method sites | Storage/fixture/preview signals |
| --- | --- | --- | --- |
| `apps/web/src/app/(hospital)/account/page.tsx` | 0 | 0 | 0 |
| `apps/web/src/app/(hospital)/communications/page.tsx` | 2 | 1 | 0 |
| `apps/web/src/app/(hospital)/dashboard/page.tsx` | 8 | 0 | 0 |
| `apps/web/src/app/(hospital)/layout.tsx` | 0 | 0 | 0 |
| `apps/web/src/app/(hospital)/live-patients/page.tsx` | 2 | 1 | 0 |
| `apps/web/src/app/(hospital)/modules/[slug]/page.tsx` | 0 | 0 | 0 |
| `apps/web/src/app/(hospital)/patients/page.tsx` | 0 | 0 | 0 |
| `apps/web/src/app/(hospital)/portal/[section]/page.tsx` | 0 | 0 | 1 |
| `apps/web/src/app/(hospital)/security-preview/page.tsx` | 0 | 0 | 1 |
| `apps/web/src/components/attendance-workspace.tsx` | 2 | 0 | 2 |
| `apps/web/src/components/bed-management-workspace.tsx` | 9 | 4 | 9 |
| `apps/web/src/components/billing-workspace.tsx` | 6 | 2 | 8 |
| `apps/web/src/components/blood-bank-workspace.tsx` | 11 | 3 | 2 |
| `apps/web/src/components/brand.tsx` | 0 | 0 | 0 |
| `apps/web/src/components/connected-billing.tsx` | 9 | 3 | 0 |
| `apps/web/src/components/connected-clinical.tsx` | 9 | 3 | 0 |
| `apps/web/src/components/connected-scheduling.tsx` | 10 | 6 | 0 |
| `apps/web/src/components/diagnosis-workspace.tsx` | 3 | 1 | 0 |
| `apps/web/src/components/doctors-workspace.tsx` | 6 | 3 | 6 |
| `apps/web/src/components/encounter-register.tsx` | 4 | 1 | 2 |
| `apps/web/src/components/front-office-workspace.tsx` | 12 | 7 | 2 |
| `apps/web/src/components/inventory-workspace.tsx` | 7 | 4 | 1 |
| `apps/web/src/components/language.tsx` | 0 | 0 | 2 |
| `apps/web/src/components/legacy-extras.tsx` | 0 | 0 | 19 |
| `apps/web/src/components/legacy-screen.tsx` | 0 | 0 | 20 |
| `apps/web/src/components/live-consultation-workspace.tsx` | 7 | 3 | 3 |
| `apps/web/src/components/medicines-workspace.tsx` | 9 | 5 | 5 |
| `apps/web/src/components/modal.tsx` | 0 | 0 | 0 |
| `apps/web/src/components/odontogram-register.tsx` | 3 | 1 | 10 |
| `apps/web/src/components/pathology-workspace.tsx` | 7 | 4 | 3 |
| `apps/web/src/components/patients-workspace.tsx` | 5 | 2 | 3 |
| `apps/web/src/components/phone-auth.tsx` | 1 | 1 | 0 |
| `apps/web/src/components/prescriptions-workspace.tsx` | 9 | 3 | 4 |
| `apps/web/src/components/public-site.tsx` | 0 | 0 | 10 |
| `apps/web/src/components/review-workspace.tsx` | 1 | 0 | 0 |
| `apps/web/src/components/role-portal.tsx` | 0 | 0 | 2 |
| `apps/web/src/components/services-workspace.tsx` | 10 | 5 | 1 |
| `apps/web/src/components/workspace.tsx` | 1 | 0 | 15 |

## apps/web/src/app/(hospital)/communications/page.tsx

Direct calls

```tsx
16: api<{ messages: Message[] }>("messages")
55: await api("messages", {
```

Persistence/fixture signals

```tsx

```


## apps/web/src/app/(hospital)/dashboard/page.tsx

Direct calls

```tsx
71: fetch("/api/hms/overview"),
72: fetch("/api/hms/patients"),
73: fetch("/api/hms/doctors"),
74: fetch("/api/hms/beds"),
75: fetch("/api/hms/invoices"),
76: fetch("/api/staff"),
77: fetch("/api/hms/notices"),
78: fetch("/api/hms/enquiries"),
```

Persistence/fixture signals

```tsx

```


## apps/web/src/app/(hospital)/live-patients/page.tsx

Direct calls

```tsx
27: api<{ patients: Patient[] }>(
45: const result = await api<Patient>("patients", {
```

Persistence/fixture signals

```tsx

```


## apps/web/src/app/(hospital)/portal/[section]/page.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
5: import { portalSections } from "@/lib/role-preview";
```


## apps/web/src/app/(hospital)/security-preview/page.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
30: <p className="form-preview-note">
```


## apps/web/src/components/attendance-workspace.tsx

Direct calls

```tsx
347: fetch("/api/hms/attendance/shifts").catch(() => null),
348: fetch("/api/hms/attendance/assignments").catch(() => null),
```

Persistence/fixture signals

```tsx
617: "Shifts and assignments loaded. Other attendance screens remain previews.",
846: "Shifts and assignments loaded. Other attendance screens remain previews.",
```


## apps/web/src/components/bed-management-workspace.tsx

Direct calls

```tsx
963: const bedsRes = await api<{ beds: ApiBed[] }>("beds?page=1");
1019: const typesRes = await api<{ bedTypes: ApiBedType[] }>(
1038: const rep = await api<ApiBedOccupancyReport>("bed-occupancy/report");
1044: const patRes = await api<{ patients: Patient[] }>("patients?page=1");
1050: const encRes = await api<{ encounters: any[] }>(
1145: await api<ApiBedType>(`bed-types/${editingItem.id}`, {
1156: const created = await api<ApiBedType>("bed-types", {
1232: const created = await api<ApiBed>("beds", {
1339: await api("bed-assignments", {
```

Persistence/fixture signals

```tsx
106: const initialWards: WardGroup[] = [
615: const initialBedAssigns: BedAssignRow[] = [
763: const initialBeds: BedRow[] = [
863: const initialBedTypes: BedTypeRow[] = [
941: const [wards, setWards] = useState<WardGroup[]>(initialWards);
943: useState<BedAssignRow[]>(initialBedAssigns);
944: const [beds, setBeds] = useState<BedRow[]>(initialBeds);
945: const [bedTypes, setBedTypes] = useState<BedTypeRow[]>(initialBedTypes);
1552: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/billing-workspace.tsx

Direct calls

```tsx
789: fetch("/api/hms/invoices"),
790: fetch("/api/hms/charge-accounts"),
791: fetch("/api/hms/payrolls"),
792: fetch("/api/hms/billing-patients"),
964: const res = await fetch("/api/hms/charge-accounts", {
996: const res = await fetch("/api/hms/invoices", {
```

Persistence/fixture signals

```tsx
725: const initialDatasets: Record<string, BillingRow[]> = {
760: useState<Record<string, BillingRow[]>>(initialDatasets);
885: // Load persisted modifications from localStorage
888: const saved = localStorage.getItem("hms-billing-workspace");
901: localStorage.setItem("hms-billing-workspace", JSON.stringify(data));
906: return data[canonicalSlug] || initialDatasets[canonicalSlug] || [];
1027: // Local fallback for smooth experience
1126: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/blood-bank-workspace.tsx

Direct calls

```tsx
488: fetch("/api/hms/blood-bank").catch(() => null),
489: fetch("/api/hms/blood-donors").catch(() => null),
490: fetch("/api/hms/blood-donations").catch(() => null),
491: fetch("/api/hms/blood-issues").catch(() => null),
492: fetch("/api/hms/patients").catch(() => null),
493: fetch("/api/hms/doctors").catch(() => null),
710: const res = await fetch("/api/hms/blood-donors", {
792: const res = await fetch("/api/hms/blood-donations", {
817: fetch("/api/hms/blood-bank")
876: const res = await fetch("/api/hms/blood-issues", {
920: fetch("/api/hms/blood-bank")
```

Persistence/fixture signals

```tsx
606: // dual offline preview fallback
1090: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/connected-billing.tsx

Direct calls

```tsx
62: () => api<{ accounts: Account[] }>(`charge-accounts?page=${page}`),
71: await api("charge-accounts", {
144: () => api<{ invoices: Invoice[] }>(`invoices?page=${page}`),
237: await api<{ patients: Patient[] }>(
249: const out = await api<{ accounts: Account[] }>(
288: await api("invoices", {
462: api<Invoice>(`invoices/${id}`),
463: api<{ payments: Payment[] }>(`invoices/${id}/payments`),
491: await api(`invoices/${id}/payments`, {
```

Persistence/fixture signals

```tsx

```


## apps/web/src/components/connected-clinical.tsx

Direct calls

```tsx
74: api<{ beds?: Bed[]; cases?: Case[]; encounters?: Encounter[] }>(
217: doctors: (await api<{ doctors: Doctor[] }>("doctors")).doctors,
228: api<{ cases: Case[] }>(`cases?page=${referencePage}`),
230: ? api<{ beds: Bed[] }>(`beds?page=${bedPage}`)
245: await api<{ patients: Patient[] }>(
283: await api(mode === "beds" || mode === "cases" ? mode : "encounters", {
514: ? api<{ notes: Note[] }>(`encounters/${encounter.id}/notes`)
525: await api(`encounters/${encounter.id}/notes`, {
543: await api(`encounters/${encounter.id}/discharge`, {
```

Persistence/fixture signals

```tsx

```


## apps/web/src/components/connected-scheduling.tsx

Direct calls

```tsx
39: const response = await fetch(`/api/staff${query}`, {
405: const load = useCallback(() => api<{ doctors: Doctor[] }>("doctors"), []),
429: await api("doctors", { method: "POST", body: JSON.stringify(editing) });
695: () => api<{ appointments: Appointment[] }>(`appointments?page=${page}`),
700: () => api<{ doctors: Doctor[] }>("doctors"),
707: ? api<{ slots: string[] }>(
724: const r = await api<{ patients: Patient[] }>(
737: await api("appointments", {
768: await api(`appointments/${row.id}`, {
1026: await api(`patients/${patient.id}`, {
```

Persistence/fixture signals

```tsx

```


## apps/web/src/components/diagnosis-workspace.tsx

Direct calls

```tsx
281: fetch("/api/hms/diagnostic-categories", { credentials: "same-origin" })
302: fetch("/api/hms/diagnostic-tests", { credentials: "same-origin" })
377: await fetch("/api/hms/diagnostic-categories", {
```

Persistence/fixture signals

```tsx

```


## apps/web/src/components/doctors-workspace.tsx

Direct calls

```tsx
396: fetch("/api/hms/doctors").catch(() => null),
397: fetch("/api/hms/doctor-departments").catch(() => null),
398: fetch("/api/hms/doctor-absences").catch(() => null),
496: const res = await fetch("/api/hms/doctors", {
555: const res = await fetch("/api/hms/doctor-departments", {
599: const res = await fetch("/api/hms/doctor-absences", {
```

Persistence/fixture signals

```tsx
519: setApiErrorBanner(err.error || t("Saved locally in preview mode."));
522: setApiErrorBanner(t("Server offline. Doctor added to local preview."));
566: setApiErrorBanner(err.error || t("Saved locally in preview mode."));
570: t("Server offline. Department added to local preview."),
613: err.error || t("Holiday preserved in local preview."),
618: t("Server offline. Holiday preserved in preview mode."),
```


## apps/web/src/components/encounter-register.tsx

Direct calls

```tsx
35: api<{ encounters: Encounter[] }>(`encounters?kind=${kind}&page=${page}`),
252: const batch = await api<{ cases: Case[] }>(`cases?page=${page}`);
258: const batch = await api<{ beds: Bed[] }>(`beds?page=${page}`);
305: await api("encounters", {
```

Persistence/fixture signals

```tsx
268: const [initialTime] = useState(() =>
443: { defaultValue: initialTime },
```


## apps/web/src/components/front-office-workspace.tsx

Direct calls

```tsx
284: fetch("/api/hms/call-logs"),
285: fetch("/api/hms/visitors"),
286: fetch("/api/hms/postals"),
287: fetch("/api/hms/enquiries"),
288: fetch("/api/hms/complaints"),
378: const res = await fetch("/api/hms/call-logs", {
426: const res = await fetch("/api/hms/visitors", {
474: const res = await fetch("/api/hms/postals", {
518: const res = await fetch("/api/hms/enquiries", {
554: await fetch(`/api/hms/enquiries/${id}/read`, { method: "PUT" });
569: const res = await fetch("/api/hms/complaints", {
609: await fetch(`/api/hms/complaints/${resolveTarget.id}/resolve`, {
```

Persistence/fixture signals

```tsx
163: // Seed default fallback data
704: : t("Dual-mode local preview")}
```


## apps/web/src/components/inventory-workspace.tsx

Direct calls

```tsx
223: fetch("/api/hms/inventory/categories"),
224: fetch("/api/hms/inventory/items"),
225: fetch("/api/hms/inventory/movements"),
280: const res = await fetch("/api/hms/inventory/categories", {
321: const res = await fetch("/api/hms/inventory/items", {
378: const res = await fetch("/api/hms/inventory/movements", {
447: const res = await fetch("/api/hms/inventory/movements", {
```

Persistence/fixture signals

```tsx
583: : t("Dual-mode local preview")}
```


## apps/web/src/components/language.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
15: const saved = localStorage.getItem("hms-language");
23: localStorage.setItem("hms-language", l);
```


## apps/web/src/components/legacy-extras.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
120: const saved = localStorage.getItem(key);
169: localStorage.setItem(key, JSON.stringify(values));
174: "Browser storage is full. Choose smaller preview images.",
268: JSON.parse(localStorage.getItem("hms-disabled-modules") || "[]"),
275: <p className="form-preview-note">
301: localStorage.setItem(
327: localStorage.getItem("hms-queue-theme") || "null",
357: <div className="queue-preview">
369: localStorage.setItem(
423: const saved = sessionStorage.getItem("hms-attendance-preview");
430: sessionStorage.setItem("hms-attendance-preview", JSON.stringify(rows));
758: const stored = sessionStorage.getItem("hms-card-templates");
767: const stored = sessionStorage.getItem("hms-smart-cards");
789: JSON.parse(sessionStorage.getItem("hms-card-templates") || "null") ||
793: JSON.parse(sessionStorage.getItem("hms-smart-cards") || "null") ||
805: sessionStorage.setItem("hms-card-templates", JSON.stringify(list));
806: sessionStorage.setItem("hms-smart-cards", JSON.stringify(cards));
928: sessionStorage.setItem(
1367: a.download = `preview-patient-card-${card.patient + 1}.png`;
```


## apps/web/src/components/legacy-screen.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
22: } from "@/lib/clinical-preview";
191: const storage = `hms-preview-v2:${s.id}:${scope}`;
194: const saved = sessionStorage.getItem(storage);
199: sessionStorage.getItem("hms-preview-v2:add-custom-fields:") || "[]",
226: sessionStorage.setItem(storage, JSON.stringify(rows));
317: setFormError("This appointment time is already booked in the preview.");
369: a.download = `preview-${s.id}.csv`;
439: setNotice("Settings saved for preview only.");
776: <p className="form-preview-note">
919: <h2 id="delete-title">{t("Delete preview record?")}</h2>
924: {t("from this preview?")}
982: <div className="patient-card-preview">
1258: JSON.parse(sessionStorage.getItem(`hms-discharge:${patient}`) || "{}"),
1267: sessionStorage.setItem(
1403: sessionStorage.getItem(`hms-odontogram-preview:${patient}`) || "{}",
1477: <p className="form-preview-note">
1496: sessionStorage.setItem(
1497: `hms-odontogram-preview:${patient}`,
1546: sessionStorage.getItem("hms-hospital-schedule") || "null",
1564: sessionStorage.setItem("hms-hospital-schedule", JSON.stringify(days));
```


## apps/web/src/components/live-consultation-workspace.tsx

Direct calls

```tsx
214: fetch("/api/hms/live-consultations", { credentials: "same-origin" })
254: fetch("/api/hms/live-meetings", { credentials: "same-origin" })
289: fetch("/api/hms/doctors", { credentials: "same-origin" })
303: fetch("/api/hms/patients", { credentials: "same-origin" })
434: const res = await fetch("/api/hms/live-consultations", {
507: const res = await fetch("/api/hms/live-meetings", {
568: await fetch("/api/hms/live-consultations/provider-settings", {
```

Persistence/fixture signals

```tsx
474: // fallback
543: // fallback
579: // silent fallback
```


## apps/web/src/components/medicines-workspace.tsx

Direct calls

```tsx
521: fetch("/api/hms/medicines"),
522: fetch("/api/hms/medicine-categories"),
523: fetch("/api/hms/medicine-brands"),
524: fetch("/api/hms/patients"),
606: const res = await fetch("/api/hms/medicine-categories", {
648: const res = await fetch("/api/hms/medicine-brands", {
709: const res = await fetch("/api/hms/medicines", {
788: const res = await fetch("/api/hms/patients", {
908: const res = await fetch("/api/hms/invoices", {
```

Persistence/fixture signals

```tsx
629: // Local fallback
678: // Local fallback
748: // Local fallback
820: // Local fallback
1554: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/odontogram-register.tsx

Direct calls

```tsx
272: const res = await api<{ patients: Patient[] }>("patients?page=1");
301: const res = await api<{ teeth: ServerToothEntry[] }>(
876: await api(`patients/${editing.patientId}/odontogram`, {
```

Persistence/fixture signals

```tsx
214: const chartStorageKey = "hms-odontogram-register-preview";
224: // Ensure fallback defaults exist
290: // Backend not yet populated with patients or in preview mode
321: // Local fallback
330: const storedLegends = sessionStorage.getItem(legendsStorageKey);
342: JSON.parse(sessionStorage.getItem(chartStorageKey) || "null") ||
358: if (ready) sessionStorage.setItem(chartStorageKey, JSON.stringify(rows));
364: sessionStorage.setItem(legendsStorageKey, JSON.stringify(legends));
502: sessionStorage.setItem(legendsStorageKey, JSON.stringify(defaultLegends));
592: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/pathology-workspace.tsx

Direct calls

```tsx
223: fetch("/api/hms/diagnostic-categories?kind=pathology").catch(
226: fetch("/api/hms/diagnostic-units").catch(() => null),
227: fetch("/api/hms/diagnostic-tests").catch(() => null),
327: const res = await fetch("/api/hms/diagnostic-tests", {
384: const res = await fetch("/api/hms/diagnostic-categories", {
441: const res = await fetch("/api/hms/diagnostic-units", {
499: const res = await fetch("/api/hms/diagnostic-tests", {
```

Persistence/fixture signals

```tsx
307: // dual offline preview fallback
356: // Dual fallback
619: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/patients-workspace.tsx

Direct calls

```tsx
121: const pRes = await api<{ patients: Patient[] }>("patients?page=1");
141: const cRes = await api<{ cases: any[] }>("cases?page=1");
161: const encRes = await api<{ encounters: any[] }>(
741: const created = await api<Patient>("patients", {
793: await api("cases", {
```

Persistence/fixture signals

```tsx
1094: <div className="avatar-preview-box">
1572: <div className="avatar-preview-box">
1922: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/phone-auth.tsx

Direct calls

```tsx
70: const response = await fetch(
```

Persistence/fixture signals

```tsx

```


## apps/web/src/components/prescriptions-workspace.tsx

Direct calls

```tsx
307: fetch("/api/hms/prescriptions").catch(() => null),
308: fetch("/api/hms/patients").catch(() => null),
309: fetch("/api/hms/doctors").catch(() => null),
310: fetch("/api/hms/medicines").catch(() => null),
311: fetch("/api/hms/medicine-categories").catch(() => null),
312: fetch("/api/hms/medicine-brands").catch(() => null),
608: await fetch(`/api/hms/prescriptions/${rx.id}/status`, {
690: const res = await fetch("/api/hms/prescriptions", {
787: await fetch("/api/hms/medicines", {
```

Persistence/fixture signals

```tsx
614: // Local preview fallback maintained
735: err.error || t("Prescription saved locally in preview mode."),
740: t("Server offline. Prescription recorded in preview mode."),
804: // offline fallback
```


## apps/web/src/components/public-site.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
15: import { previewSlots } from "@/lib/clinical-preview";
62: const saved = localStorage.getItem(FRONT_KEY);
68: if (page === "appointment" && !previewSlots(doctor, date).length) {
417: <p className="form-preview-note">
432: <p className="form-preview-note">
434: "Frontend preview · This form does not submit real requests or create accounts.",
507: disabled={!previewSlots(doctor, date).length}
511: previewSlots(doctor, date).length
516: {previewSlots(doctor, date).map((time) => (
574: <small>{t("Frontend preview · Sample content")}</small>
```


## apps/web/src/components/review-workspace.tsx

Direct calls

```tsx
83: fetch("/api/hms/reviews")
```

Persistence/fixture signals

```tsx

```


## apps/web/src/components/role-portal.tsx

Direct calls

```tsx

```

Persistence/fixture signals

```tsx
20: } from "@/lib/role-preview";
74: <h1>{t("Select a role to preview this screen")}</h1>
```


## apps/web/src/components/services-workspace.tsx

Direct calls

```tsx
361: fetch("/api/hms/services"),
362: fetch("/api/hms/ambulances"),
363: fetch("/api/hms/ambulance-calls"),
364: fetch("/api/hms/packages"),
365: fetch("/api/hms/insurances"),
493: const res = await fetch("/api/hms/services", {
553: const res = await fetch("/api/hms/ambulances", {
630: const res = await fetch("/api/hms/ambulance-calls", {
683: const res = await fetch("/api/hms/packages", {
732: const res = await fetch("/api/hms/insurances", {
```

Persistence/fixture signals

```tsx
1069: : t("Local preview mode · Syncing locally")}
```


## apps/web/src/components/workspace.tsx

Direct calls

```tsx
223: api<Identity>("me")
```

Persistence/fixture signals

```tsx
42: previewRoles,
50: } from "@/lib/role-preview";
63: JSON.parse(localStorage.getItem("hms-disabled-modules") || "[]"),
74: setDark(localStorage.getItem("hms-theme") === "dark");
78: localStorage.setItem("hms-theme", value ? "light" : "dark");
92: const saved = sessionStorage.getItem("hms-preview-role");
93: if (saved && previewRoles.includes(saved)) setRole(saved);
97: sessionStorage.setItem("hms-preview-role", next);
226: if (!sessionStorage.getItem("hms-preview-role")) {
227: const current = previewRoles.find(
421: <Link href="/security-preview">
441: <div className="preview-strip">
446: "Frontend preview · Sample data · Changes stay in this browser tab",
450: <label className="role-preview">
457: {previewRoles.map((r) => (
```
