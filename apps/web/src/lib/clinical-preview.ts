// Synthetic relationships, matching the original dependent-select interactions.
// Real availability, charges and stock must be supplied and revalidated by Go.
export const previewDoctors = [
  { name: "Dr. Avery Reed", department: "General Medicine", charge: "250" },
  { name: "Dr. Robin Patel", department: "Cardiology", charge: "400" },
  { name: "Dr. Quinn Parker", department: "Paediatrics", charge: "300" },
];
export const previewBeds: Record<string, string[]> = {
  General: ["General — G01", "General — G02"],
  Private: ["Private — P01"],
  ICU: ["ICU — I01"],
};
export function previewSlots(doctor: string, date: string) {
  if (
    !doctor ||
    !/^\d{4}-\d{2}-\d{2}$/.test(date) ||
    new Date(date + "T12:00:00").getDay() === 0
  )
    return [];
  return doctor === "Dr. Robin Patel"
    ? ["14:00", "14:30", "15:00"]
    : ["09:00", "09:30", "10:00", "10:30"];
}
export function dependentOptions(
  key: string,
  values: Record<string, string>,
): string[] | undefined {
  if (key === "doctor_id" && "department_id" in values)
    return previewDoctors
      .filter((d) => d.department === values.department_id)
      .map((d) => d.name);
  if (key === "bed_type_id") return Object.keys(previewBeds);
  if (key === "bed_id" && "bed_type_id" in values)
    return previewBeds[values.bed_type_id] || [];
  if (key === "case_id")
    return values.patient_id
      ? [`${values.patient_id} — CASE-001`, `${values.patient_id} — CASE-002`]
      : [];
  if (key === "timeslot")
    return previewSlots(values.doctor_id, values.opd_date);
  if (key === "next_visit_time") return ["Days", "Weeks", "Months"];
  return undefined;
}
export function changeClinicalField(
  values: Record<string, string>,
  key: string,
  value: string,
) {
  const next = { ...values, [key]: value };
  if (key === "department_id") {
    next.doctor_id = "";
    next.timeslot = "";
    next.appointment_charge = "";
  }
  if (key === "doctor_id") {
    next.timeslot = "";
    next.standard_charge =
      previewDoctors.find((d) => d.name === value)?.charge || "";
    next.appointment_charge = next.standard_charge;
  }
  if (key === "opd_date") next.timeslot = "";
  if (key === "patient_id") next.case_id = "";
  if (key === "bed_type_id") next.bed_id = "";
  return next;
}
export function validateClinicalPreview(
  id: string,
  v: Record<string, string>,
): string {
  if (
    v.admission_date &&
    v.discharge_date &&
    v.discharge_date < v.admission_date
  )
    return "Discharge date cannot be before admission date.";
  if (v.dob && v.dob > new Date().toISOString().slice(0, 10))
    return "Date of birth cannot be in the future.";
  if (
    (id === "appointments" || id === "appointment-calendars") &&
    !previewSlots(v.doctor_id, v.opd_date).includes(v.timeslot)
  )
    return "Choose an available appointment time.";
  if (
    v.bed_type_id &&
    v.bed_id &&
    !previewBeds[v.bed_type_id]?.includes(v.bed_id)
  )
    return "Choose a bed from the selected bed type.";
  if (v.case_id && v.patient_id && !v.case_id.startsWith(v.patient_id + " — "))
    return "Choose a case for the selected patient.";
  return "";
}
