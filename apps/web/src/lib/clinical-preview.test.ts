import test from "node:test";
import assert from "node:assert/strict";
import {
  changeClinicalField,
  dependentOptions,
  validateClinicalPreview,
  previewSlots,
} from "./clinical-preview.ts";
test("changing clinical parents clears stale child selections", () => {
  assert.equal(
    changeClinicalField(
      { doctor_id: "Dr. Avery Reed", timeslot: "09:00" },
      "department_id",
      "Cardiology",
    ).doctor_id,
    "",
  );
  assert.equal(
    changeClinicalField(
      { case_id: "Alex Morgan — CASE-001" },
      "patient_id",
      "Jamie Wilson",
    ).case_id,
    "",
  );
  assert.equal(
    changeClinicalField({ bed_id: "General — G01" }, "bed_type_id", "ICU")
      .bed_id,
    "",
  );
  assert.deepEqual(
    dependentOptions("doctor_id", { department_id: "Cardiology" }),
    ["Dr. Robin Patel"],
  );
});
test("unavailable slots and inconsistent dates are rejected", () => {
  assert.deepEqual(previewSlots("Dr. Avery Reed", "2026-10-04"), []);
  assert.notEqual(
    validateClinicalPreview("appointments", {
      doctor_id: "Dr. Robin Patel",
      opd_date: "2026-10-05",
      timeslot: "09:00",
    }),
    "",
  );
  assert.equal(
    validateClinicalPreview("appointments", {
      doctor_id: "Dr. Robin Patel",
      opd_date: "2026-10-05",
      timeslot: "14:00",
    }),
    "",
  );
  assert.notEqual(
    validateClinicalPreview("patient-admissions", {
      admission_date: "2026-10-05",
      discharge_date: "2026-10-04",
    }),
    "",
  );
});

test("case-only forms still offer cases without a patient selector", () => {
  assert.ok(dependentOptions("case_id", {})!.length > 0);
  assert.deepEqual(dependentOptions("case_id", { patient_id: "" }), []);
});
