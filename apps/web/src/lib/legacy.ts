import source from "./legacy-catalog.json";
export type Field = {
  key: string;
  label: string;
  type: string;
  required: boolean;
  source?: string;
};
export type Screen = {
  id: string;
  title: string;
  group: string;
  fields: Field[];
  columns: string[];
  source: string;
  extracted: boolean;
};
const fields = (definitions: string): Field[] =>
  definitions.split("|").map((entry) => {
    const [key, type = "text", required = ""] = entry.split(":");
    return {
      key,
      label: key.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase()),
      type,
      required: required === "required",
    };
  });
const overrides: Record<string, Partial<Screen>> = {
  "live-consultations-live-meetings": {
    title: "Live Meetings",
    fields: fields(
      "title:text:required|meeting_date:date:required|duration_minutes:number:required|host:select|members:select|description:textarea",
    ),
    columns: ["Title", "Host", "Meeting Date", "Duration", "Status"],
  },
  "ipd-consultant-registers": {
    fields: fields(
      "applied_date:date:required|doctor:select:required|instruction:textarea:required|instruction_date:date:required",
    ),
    columns: ["Applied Date", "Doctor", "Instruction", "Instruction Date"],
  },
  "ipd-bills": {
    fields: fields(
      "bill_date:date:required|payment_mode:select|discount:number|tax:number|note:textarea",
    ),
    columns: ["Bill Date", "Amount", "Payment Mode", "Status"],
  },
  "opd-prescriptions": {
    fields: fields("header_note:textarea|footer_note:textarea"),
    columns: ["Date", "Doctor", "Notes"],
  },
  "appointment-calendars": {
    fields: source.find((s) => s.id === "appointments")!.fields,
  },
  "appointment-transaction": {
    columns: [
      "Patient",
      "Transaction ID",
      "Payment Type",
      "Amount",
      "Date",
      "Status",
    ],
  },
  "patient-queues": {
    fields: fields(
      "patient:select:required|doctor:select:required|date:date:required|queue_number:number:required|status:select",
    ),
    columns: ["Patient", "Doctor", "Queue Number", "Date", "Status"],
  },
  "payment-reports": {
    columns: ["Payment Date", "Patient", "Payment Type", "Amount", "Status"],
  },
  "manual-bill-payments": {
    columns: ["Patient", "Bill Number", "Amount", "Payment Date", "Status"],
  },
  enquiries: {
    fields: fields(
      "full_name:text:required|email:email:required|phone:tel|subject:text:required|message:textarea:required",
    ),
    columns: ["Full Name", "Email", "Phone", "Subject", "Date"],
  },
  "goole-meet-consultation": {
    fields: fields(
      "title:text:required|patient:select:required|doctor:select:required|consultation_date:date:required|duration_minutes:number:required|description:textarea",
    ),
    columns: ["Title", "Patient", "Doctor", "Date", "Duration", "Status"],
  },
  "purchase-medicines": {
    fields: fields(
      "supplier:text:required|purchase_date:date:required|invoice_number:text:required|payment_mode:select|note:textarea",
    ),
    columns: [
      "Invoice Number",
      "Supplier",
      "Purchase Date",
      "Amount",
      "Status",
    ],
  },
  "used-medicine": { columns: ["Medicine", "Patient", "Quantity", "Date"] },
  "medicine-bills": {
    fields: fields(
      "patient:select:required|doctor:select:required|bill_date:date:required|payment_mode:select|note:textarea",
    ),
    columns: [
      "Bill Number",
      "Patient",
      "Doctor",
      "Bill Date",
      "Amount",
      "Status",
    ],
  },
  emails: {
    title: "Mail",
    fields: fields(
      "recipient:email:required|subject:text:required|message:textarea:required",
    ),
    columns: ["Recipient", "Subject", "Date", "Status"],
  },
  "email-template": {
    fields: fields(
      "name:text:required|subject:text:required|body:textarea:required",
    ),
    columns: ["Name", "Subject", "Status"],
  },
  "add-custom-fields": {
    fields: fields(
      "field_name:text:required|module:select:required|field_type:select:required|values:textarea|is_required:select|grid:number",
    ),
    columns: ["Field Name", "Module", "Field Type", "Is Required"],
  },
  "add-on": {
    title: "Add-ons",
    fields: fields("name:text:required|version:text|package:file:required"),
    columns: ["Name", "Version", "Status"],
  },
  "patient-id-card-template": {
    fields: fields(
      "name:text:required|header_color:color|show_email:select|show_phone:select|show_blood_group:select|show_address:select",
    ),
    columns: ["Name", "Header Color", "Status"],
  },
  "generate-patient-id-card": {
    fields: fields("patient:select:required|template:select:required"),
    columns: ["Patient", "Template", "Created At"],
  },
};
const extra = (id: string, title: string, group: string): Screen => ({
  id,
  title,
  group,
  fields: [],
  columns: [],
  source: "User screenshots, October 2026",
  extracted: false,
});
const groupNames: Record<string, string> = {
  "Patient ID Card": "Patient Smart Cards",
  "Add-ons": "AddOn",
  Billing: "Billings",
  "Blood Bank": "Blood Banks",
  Medicine: "Medicines",
};
export const screens: Screen[] = [
  ...source.map((s) => ({ ...s, ...overrides[s.id] })),
  extra("attendance", "Attendance", "Attendance"),
  extra("manage-attendance", "Manage Attendance", "Manage Attendance"),
  extra("modules-setting", "Modules Setting", "Settings"),
  extra("patient-queue-theme", "Patient Queue Theme", "Settings"),
  {
    ...source.find((s) => s.id === "services")!,
    id: "front-cms-services",
    title: "Front CMS Services",
    group: "Front CMS",
  },
].map((s) => ({
  ...s,
  group:
    s.id === "ipd-patient-departments"
      ? "IPD - Patient In"
      : s.id === "opd-patient-departments"
        ? "OPD - Patient Out"
        : s.id === "email-template"
          ? "Email Templates"
          : s.id === "complaints"
            ? "Front CMS"
            : s.id === "services"
              ? "Services"
              : groupNames[s.group] || s.group,
  title:
    (
      {
        "patient-id-card-template": "Smart Patient Card Templates",
        "generate-patient-id-card": "Generate Patient Smart Cards",
        "front-settings": "CMS",
        settings: "General",
        "payment-gateway": "Payment Gateways",
        "add-custom-fields": "Custom Field",
        complaints: "Complaint",
        "add-on": "AddOn",
      } as Record<string, string>
    )[s.id] || s.title,
}));
// Source menu order, with screenshot-confirmed attendance add-on entries.
export const groups = [
  "Patient Smart Cards",
  "Users",
  "Odontogram",
  "AddOn",
  "Appointments",
  "Attendance",
  "Manage Attendance",
  "IPD - Patient In",
  "OPD - Patient Out",
  "Billings",
  "Bed Management",
  "Blood Banks",
  "Documents",
  "Doctors",
  "Prescriptions",
  "Diagnosis",
  "Enquiries",
  "Finance",
  "Front Office",
  "Front CMS",
  "Hospital Charges",
  "Inventory",
  "Live Consultations",
  "Medicines",
  "Patients",
  "Pathology",
  "Reports",
  "Radiology",
  "Services",
  "SMS / Mail",
  "Email Templates",
  "Settings",
  "Vaccinations",
];
const tabOrder: Record<string, string[]> = {
  Settings: [
    "settings",
    "hospital-schedule",
    "modules-setting",
    "currency-settings",
    "operation-categories",
    "operations",
    "payment-gateway",
    "add-custom-fields",
    "patient-queue-theme",
  ],
  "Front CMS": [
    "front-settings",
    "front-cms-services",
    "notice-boards",
    "testimonials",
    "complaints",
  ],
  Diagnosis: ["diagnosis-categories", "patient-diagnosis-test"],
  Appointments: [
    "appointments",
    "appointment-transaction",
    "patient-queues",
    "appointment-calendars",
  ],
  Patients: [
    "patients",
    "patient-cases",
    "case-handlers",
    "patient-admissions",
  ],
  Services: [
    "insurances",
    "packages",
    "services",
    "ambulances",
    "ambulance-calls",
  ],
};
export function groupScreens(group: string) {
  const items = screens.filter((s) => s.group === group);
  const order = tabOrder[group];
  return order
    ? items.sort((a, b) => {
        const rank = (id: string) =>
          order.includes(id) ? order.indexOf(id) : 999;
        return rank(a.id) - rank(b.id);
      })
    : items;
}
export const previewRoles = [
  "Admin",
  "Doctor",
  "Patient",
  "Nurse",
  "Receptionist",
  "Pharmacist",
  "Accountant",
  "Case Manager",
  "Lab Technician",
];
const roleGroups: Record<string, string[]> = {
  Doctor: [
    "Appointments",
    "IPD / OPD",
    "Patients",
    "Prescriptions",
    "Diagnosis",
    "Documents",
    "Doctors",
    "Live Consultations",
    "Pathology",
    "Radiology",
    "Reports",
    "Odontogram",
  ],
  Patient: [
    "Appointments",
    "IPD / OPD",
    "Prescriptions",
    "Diagnosis",
    "Documents",
    "Billing",
    "Vaccinations",
    "Live Consultations",
  ],
  Nurse: [
    "Patients",
    "IPD / OPD",
    "Bed Management",
    "Blood Bank",
    "Documents",
    "Reports",
    "Vaccinations",
  ],
  Receptionist: [
    "Patient ID Card",
    "Appointments",
    "Patients",
    "IPD / OPD",
    "Bed Management",
    "Front Office",
    "Enquiries",
    "SMS / Mail",
    "Doctors",
  ],
  Pharmacist: ["Medicine", "Prescriptions", "Inventory"],
  Accountant: ["Billing", "Finance", "Services"],
  "Case Manager": ["Patients", "Appointments", "Bed Management", "Documents"],
  "Lab Technician": [
    "Pathology",
    "Radiology",
    "Diagnosis",
    "Blood Bank",
    "Reports",
    "IPD / OPD",
  ],
};
export function visibleGroups(role: string) {
  return role === "Admin"
    ? groups
    : groups.filter(
        (g) =>
          (roleGroups[role] || [])
            .flatMap((original) =>
              original === "IPD / OPD"
                ? ["IPD - Patient In", "OPD - Patient Out"]
                : [groupNames[original] || original],
            )
            .includes(g) || g === "Attendance",
      );
}
export function screenHref(s: Screen) {
  return s.id === "patients" ? "/patients" : `/modules/${s.id}`;
}
export const people = [
  "Alex Morgan",
  "Jamie Wilson",
  "Taylor Davis",
  "Jordan Lee",
  "Casey Brown",
  "Sam Williams",
];
export function optionsFor(f: Field): string[] {
  const k = `${f.key} ${f.label}`.toLowerCase();
  if (/department/.test(k))
    return ["General Medicine", "Cardiology", "Orthopaedics", "Paediatrics"];
  if (/blood/.test(k))
    return ["A+", "A−", "B+", "B−", "AB+", "AB−", "O+", "O−"];
  if (/gender/.test(k)) return ["Male", "Female"];
  if (/status/.test(k)) return ["Active", "Inactive", "Pending", "Completed"];
  if (/show_|is_required/.test(k)) return ["Yes", "No"];
  if (/field_type/.test(k))
    return [
      "Text",
      "Number",
      "Date",
      "Select",
      "Multi Select",
      "Textarea",
      "Toggle",
    ];
  if (/module/.test(k)) return groups;
  if (/template/.test(k)) return ["Standard Patient Card", "Hospital Card"];
  if (/doctor|consultant/.test(k))
    return ["Dr. Avery Reed", "Dr. Robin Patel", "Dr. Quinn Parker"];
  if (/patient|donor|user|employee|person/.test(k)) return people;
  if (/department/.test(k))
    return ["General Medicine", "Cardiology", "Orthopaedics", "Paediatrics"];
  if (/bed/.test(k))
    return ["General — G01", "General — G02", "Private — P01", "ICU — I01"];
  if (/payment|mode/.test(k)) return ["Cash", "Bank transfer", "Card", "Other"];
  if (/medicine/.test(k))
    return ["Paracetamol 500 mg", "Amoxicillin 250 mg", "Cetirizine 10 mg"];
  if (/frequency/.test(k))
    return ["Once daily", "Twice daily", "Three times daily"];
  if (/unit/.test(k)) return ["mg", "mL", "g/dL", "mmol/L"];
  if (/category|type/.test(k))
    return ["General", "Consultation", "Laboratory", "Surgery"];
  return ["General", "Standard", "Other"];
}
export type PreviewRow = {
  id: string;
  values: Record<string, string>;
  status: string;
};
export function seedRows(s: Screen): PreviewRow[] {
  return Array.from({ length: 6 }, (_, i) => {
    const values: Record<string, string> = {};
    for (const f of s.fields)
      values[f.key] =
        f.type === "select"
          ? optionsFor(f)[i % optionsFor(f).length]
          : f.type === "date"
            ? `2026-10-0${i + 1}`
            : f.type === "number"
              ? String((i + 1) * 100)
              : /first_name/.test(f.key)
                ? people[i].split(" ")[0]
                : /last_name/.test(f.key)
                  ? people[i].split(" ")[1]
                  : /email/.test(f.key)
                    ? `sample${i + 1}@example.invalid`
                    : /phone/.test(f.key)
                      ? `+25470000010${i}`
                      : /name/.test(f.key)
                        ? `${s.title.replace(/s$/, "")} ${i + 1}`
                        : /password|image|file/.test(f.key)
                          ? ""
                          : /description|note|remark/.test(f.key)
                            ? "Sample record for frontend review."
                            : `${i + 1}`;
    for (const f of s.fields) {
      if (f.type === "text") {
        if (
          /ipd_number|opd_number|bill_number|invoice_number|case_id|reference_no/.test(
            f.key,
          )
        )
          values[f.key] =
            `${s.id.split("-")[0].toUpperCase()}-${String(i + 1).padStart(4, "0")}`;
        else if (/report_type/.test(f.key))
          values[f.key] = ["Blood test", "Chest X-ray", "Lab investigation"][
            i % 3
          ];
        else if (/symptom/.test(f.key))
          values[f.key] = ["Headache", "Fever", "Fatigue"][i % 3];
        else if (/address/.test(f.key))
          values[f.key] = [
            "12 Sample Road",
            "4 Example Avenue",
            "8 Preview Lane",
          ][i % 3];
        else if (/city/.test(f.key)) values[f.key] = "Nairobi";
        else if (/instruction/.test(f.key))
          values[f.key] = "Sample instructions for review";
        else if (/title|subject/.test(f.key))
          values[f.key] = `${s.title} — sample ${i + 1}`;
        else if (f.key === "name" && s.id === "medicines")
          values[f.key] = [
            "Paracetamol 500 mg",
            "Amoxicillin 250 mg",
            "Cetirizine 10 mg",
          ][i % 3];
      }
    }
    s.columns.forEach((c) => {
      if (values[c]) return;
      const match = s.fields.find((f) => f.label === c);
      if (match) {
        values[c] = values[match.key];
        return;
      }
      values[c] = /patient|donor|employee|user|name/i.test(c)
        ? people[i]
        : /doctor|consultant/i.test(c)
          ? "Dr. Avery Reed"
          : /date/i.test(c)
            ? `2026-10-0${i + 1}`
            : /amount|charge|salary|price|balance/i.test(c)
              ? ((i + 1) * 250).toLocaleString()
              : /blood/i.test(c)
                ? ["A+", "O+", "B+"][i % 3]
                : /phone/i.test(c)
                  ? `+25470000010${i}`
                  : /status/i.test(c)
                    ? i === 4
                      ? "Inactive"
                      : "Active"
                    : /email/i.test(c)
                      ? `sample${i + 1}@example.invalid`
                      : `${c} ${i + 1}`;
    });
    return {
      id: `DEMO-${String(i + 1).padStart(4, "0")}`,
      values,
      status: i === 4 ? "Inactive" : "Active",
    };
  });
}
