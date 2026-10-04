import { groups, groupScreens, screenHref } from "./legacy";
import tabRoles from "./role-tabs.json";
export function roleCanSeeScreen(role: string, id: string) {
  const allowed = (tabRoles as Record<string, string[]>)[id];
  return role === "Admin" || !allowed || allowed.includes(role);
}
export type RoleLink = { title: string; href: string; group?: string };
// Ordering and employee/patient destinations follow layouts/menu.blade.php.
const menus: Record<string, string[]> = {
  Doctor: [
    "Appointments",
    "Odontogram",
    "Bed Management",
    "@doctors",
    "Schedules",
    "Prescriptions",
    "Documents",
    "Diagnosis",
    "@notices",
    "IPD - Patient In",
    "OPD - Patient Out",
    "Live Consultations",
    "@payrolls",
    "Patients",
    "Reports",
    "SMS / Mail",
  ],
  "Case Manager": [
    "@doctors",
    "@notices",
    "Live Meetings",
    "@payrolls",
    "Patients",
    "Services",
    "SMS / Mail",
  ],
  Receptionist: [
    "Appointments",
    "Patient Smart Cards",
    "Billings",
    "Doctors",
    "Diagnosis",
    "Enquiries",
    "Front Office",
    "@notices",
    "Complaint",
    "Hospital Charges",
    "IPD - Patient In",
    "OPD - Patient Out",
    "Live Meetings",
    "@payrolls",
    "Patients",
    "Pathology",
    "Radiology",
    "Services",
    "SMS / Mail",
  ],
  Pharmacist: [
    "@doctors",
    "@prescriptions",
    "@notices",
    "Live Meetings",
    "Medicines",
    "@payrolls",
    "Pathology",
    "Radiology",
    "SMS / Mail",
  ],
  Nurse: [
    "Bed Management",
    "IPD - Patient In",
    "OPD - Patient Out",
    "@notices",
    "Live Meetings",
    "@payrolls",
    "Medicines",
    "Prescriptions",
    "Diagnosis",
    "Reports",
    "Doctors",
  ],
  "Lab Technician": [
    "Blood Banks",
    "@doctors",
    "IPD - Patient In",
    "Diagnosis",
    "@notices",
    "Live Meetings",
    "Medicines",
    "@payrolls",
    "Pathology",
    "Radiology",
  ],
  Accountant: [
    "Billings",
    "Finance",
    "@notices",
    "Live Meetings",
    "@payrolls",
    "Services",
    "SMS / Mail",
  ],
  Patient: [
    "@appointments",
    "Odontogram",
    "@bills",
    "@documents",
    "@notices",
    "Complaint",
    "@ipd",
    "@opd",
    "@invoices",
    "@consultations",
    "@cases",
    "@admissions",
    "@prescriptions",
    "@vaccinations",
    "@reports",
  ],
};
export const portalSections: Record<
  string,
  { title: string; columns: string[]; source: string }
> = {
  appointments: {
    title: "Appointments",
    columns: ["Doctor", "Department", "Date", "Status"],
    source: "PatientAppoinmentDetailTable.php",
  },
  doctors: {
    title: "Doctors",
    columns: ["Doctor", "Department", "Phone", "Email"],
    source: "employees/doctors",
  },
  notices: {
    title: "Notice Boards",
    columns: ["Title", "Date", "Description"],
    source: "employees/notice_boards",
  },
  payrolls: {
    title: "My Payrolls",
    columns: [
      "Payroll ID",
      "Month",
      "Year",
      "Basic Salary",
      "Allowance",
      "Deductions",
      "Net Salary",
      "Status",
    ],
    source: "PayrollTable.php",
  },
  prescriptions: {
    title: "Prescriptions",
    columns: [
      "Patient",
      "Doctor",
      "Medical History",
      "Current Medication",
      "Health Insurance",
      "Low Income",
      "Reference",
      "Status",
    ],
    source: "patients_prescription_list; employee_prescription_list",
  },
  bills: {
    title: "Bills",
    columns: ["Bill Number", "Admission ID", "Date", "Amount", "Status"],
    source: "employees/bills",
  },
  invoices: {
    title: "Invoices",
    columns: ["Invoice Number", "Date", "Amount", "Status"],
    source: "employees/invoices",
  },
  documents: {
    title: "Documents",
    columns: ["Title", "Document Type", "Date"],
    source: "PatientDocumentTable.php",
  },
  ipd: {
    title: "IPD - Patient In",
    columns: ["IPD Number", "Doctor", "Admission Date", "Bed", "Bill Status"],
    source: "ipd_patient_list",
  },
  opd: {
    title: "OPD - Patient Out",
    columns: [
      "OPD Number",
      "Doctor",
      "Appointment Date",
      "Standard Charge",
      "Payment Mode",
      "Total Visits",
    ],
    source: "opd_patient_list",
  },
  cases: {
    title: "Patient Cases",
    columns: ["Case ID", "Doctor", "Case Date", "Fee", "Status"],
    source: "patients_cases_list",
  },
  admissions: {
    title: "Patient Admissions",
    columns: [
      "Admission ID",
      "Doctor",
      "Admission Date",
      "Discharge Date",
      "Status",
    ],
    source: "employees/patient_admissions",
  },
  vaccinations: {
    title: "Vaccinated Patients",
    columns: ["Vaccination", "Serial Number", "Dose Number", "Dose Given Date"],
    source: "patient_vaccinated_list",
  },
  reports: {
    title: "Reports",
    columns: ["Report Number", "Report Type", "Doctor", "Date"],
    source: "employees/patient_diagnosis_test",
  },
  consultations: {
    title: "Live Consultations",
    columns: ["Title", "Doctor", "Date", "Duration", "Status"],
    source: "live_consultations",
  },
};
export function roleNavigation(role: string): RoleLink[] {
  return (role === "Admin" ? groups : menus[role] || []).map((title) => {
    if (title.startsWith("@")) {
      const key = title.slice(1);
      return { title: portalSections[key].title, href: `/portal/${key}` };
    }
    const special: Record<string, string> = {
      Schedules: "schedules",
      "Live Meetings": "live-consultations-live-meetings",
      Complaint: "complaints",
    };
    const first = groupScreens(title).find((s) => roleCanSeeScreen(role, s.id));
    return {
      title,
      group: title,
      href: special[title]
        ? `/modules/${special[title]}`
        : first
          ? screenHref(first)
          : "/dashboard",
    };
  });
}
export function canPreviewPortal(role: string, section: string) {
  return roleNavigation(role).some(
    (link) => link.href === `/portal/${section}`,
  );
}
