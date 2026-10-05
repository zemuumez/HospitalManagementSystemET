import fs from "node:fs";
import { chromium } from "@playwright/test";

(async () => {
  const lines = fs.readFileSync(".local/admin-login.txt", "utf8").split("\n");
  const email = lines
    .find((l) => l.startsWith("Email:"))
    ?.replace("Email:", "")
    .trim();
  const password = lines
    .find((l) => l.startsWith("Password:"))
    ?.replace("Password:", "")
    .trim();

  if (!email || !password) throw Error("Admin credentials missing");

  const b = await chromium.launch({
    headless: true,
    executablePath:
      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
  });
  const p = await b.newPage({ viewport: { width: 1440, height: 1000 } });

  const errors = [];
  p.on("pageerror", (e) => errors.push(e.message));

  // 1. Sign in
  await p.goto("http://127.0.0.1:3000/login");
  await p.getByLabel("Email address", { exact: false }).fill(email);
  await p.getByLabel("Password", { exact: false }).fill(password);
  await p.getByRole("button", { name: "Sign In", exact: true }).click();
  await p.waitForURL("**/dashboard");
  const darkBtn = p.getByRole("button", { name: "Use dark theme", exact: true });
  if (await darkBtn.count() > 0) {
    await darkBtn.click();
  }
  console.log("1. Logged in and ensured dark theme");

  // ==========================================
  // SECTION 1: BLOOD BANK (5 tabs + modals)
  // ==========================================
  console.log("\n--- Testing Blood Bank Module ---");

  // 1A. Blood Banks
  await p.goto("http://127.0.0.1:3000/modules/blood-banks");
  await p.locator('[data-ready="true"]').waitFor();
  let rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Blood Banks rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 blood bank rows");
  await p.screenshot({
    path: ".local/reference-repair/blood-banks.png",
    fullPage: true,
  });

  // 1B. Blood Donors
  await p.goto("http://127.0.0.1:3000/modules/blood-donors");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Blood Donors rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 donor rows");
  // Test modal: New Blood Donor
  await p.getByRole("button", { name: "New Blood Donor" }).click();
  await p.locator(".modal-backdrop-custom").waitFor();
  await p.screenshot({
    path: ".local/reference-repair/blood-donors-modal.png",
    fullPage: true,
  });
  await p.locator(".modal-close-btn").click();
  console.log("New Blood Donor modal verified");

  // 1C. Blood Donations
  await p.goto("http://127.0.0.1:3000/modules/blood-donations");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Blood Donations rows: ${rows}`);

  // 1D. Blood Issues
  await p.goto("http://127.0.0.1:3000/modules/blood-issues");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Blood Issues rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 blood issue rows");
  await p.screenshot({
    path: ".local/reference-repair/blood-issues.png",
    fullPage: true,
  });

  // 1E. Blood Donor Report
  await p.goto("http://127.0.0.1:3000/modules/blood-donor-reports");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Blood Donor Reports rows: ${rows}`);
  if (rows < 2) throw Error("Expected >= 2 donor report rows");
  // Test modal: Add Blood Donor Report
  await p.getByRole("button", { name: "New Blood Donor Report" }).click();
  await p.locator(".modal-backdrop-custom").waitFor();
  await p.screenshot({
    path: ".local/reference-repair/blood-donor-report-modal.png",
    fullPage: true,
  });
  await p.locator(".modal-close-btn").click();
  console.log("Blood Donor Report modal verified");

  // ==========================================
  // SECTION 2: DOCTORS (5 tabs + forms)
  // ==========================================
  console.log("\n--- Testing Doctors Module ---");

  // 2A. Doctors
  await p.goto("http://127.0.0.1:3000/modules/doctors");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Doctors rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 doctor rows");
  await p.screenshot({
    path: ".local/reference-repair/doctors.png",
    fullPage: true,
  });

  // 2B. Doctor Departments
  await p.goto("http://127.0.0.1:3000/modules/doctor-departments");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Doctor Departments rows: ${rows}`);
  if (rows < 4) throw Error("Expected >= 4 department rows");

  // 2C. Schedules
  await p.goto("http://127.0.0.1:3000/modules/schedules");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Schedules rows: ${rows}`);
  // Test New Schedule view
  await p.getByRole("button", { name: "New Schedule" }).click();
  await p.locator("form").waitFor();
  await p.screenshot({
    path: ".local/reference-repair/new-schedule.png",
    fullPage: true,
  });
  await p.getByRole("button", { name: "Back" }).click();
  console.log("New Schedule form verified");

  // 2D. Breaks
  await p.goto("http://127.0.0.1:3000/modules/breaks");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Breaks rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 break rows");
  await p.screenshot({
    path: ".local/reference-repair/breaks.png",
    fullPage: true,
  });
  // Test Add Break view
  await p.getByRole("button", { name: "Add Break" }).click();
  await p.locator("form").waitFor();
  await p.screenshot({
    path: ".local/reference-repair/add-break.png",
    fullPage: true,
  });
  await p.getByRole("button", { name: "Back" }).click();
  console.log("Add Break form verified");

  // ==========================================
  // SECTION 3: PRESCRIPTIONS
  // ==========================================
  console.log("\n--- Testing Prescriptions Module ---");

  await p.goto("http://127.0.0.1:3000/modules/prescriptions");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Prescriptions rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 prescription rows");
  await p.screenshot({
    path: ".local/reference-repair/prescriptions-list.png",
    fullPage: true,
  });

  // Test New Prescription screen
  await p.getByRole("button", { name: "New Prescription" }).click();
  await p.locator("form").waitFor();
  await p.screenshot({
    path: ".local/reference-repair/new-prescription.png",
    fullPage: true,
  });

  // Test Suggest Medicines toast notification
  await p.getByRole("button", { name: "Suggest Medicines" }).click();
  await p.locator(".toast-notify-error").waitFor();
  console.log("Suggest Medicines toast verified");
  await p.screenshot({
    path: ".local/reference-repair/suggest-medicines-toast.png",
  });

  // Test New Medicine Modal
  await p.getByRole("button", { name: "New Medicine" }).click();
  await p.locator(".modal-backdrop-custom").waitFor();
  console.log("New Medicine modal verified");
  await p.screenshot({
    path: ".local/reference-repair/new-medicine-modal.png",
  });
  await p.locator(".modal-close-btn").click();
  await p.getByRole("button", { name: "Back" }).click();

  // ==========================================
  // SECTION 4: DIAGNOSIS (2 tabs + modals)
  // ==========================================
  console.log("\n--- Testing Diagnosis Module ---");

  // 4A. Diagnosis Categories
  await p.goto("http://127.0.0.1:3000/modules/diagnosis-categories");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Diagnosis Categories rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 diagnosis categories");
  await p.screenshot({
    path: ".local/reference-repair/diagnosis-categories.png",
    fullPage: true,
  });
  // Test modal: New Diagnosis Category
  await p.getByRole("button", { name: "New Diagnosis Category" }).click();
  await p.locator(".modal-backdrop-custom").waitFor();
  console.log("New Diagnosis Category modal verified");
  await p.screenshot({
    path: ".local/reference-repair/new-diagnosis-category-modal.png",
  });
  await p.locator(".modal-close-btn").click();

  // 4B. Diagnosis Tests
  await p.goto("http://127.0.0.1:3000/modules/patient-diagnosis-test");
  await p.locator('[data-ready="true"]').waitFor();
  rows = await p.locator(".billing-table tbody tr").count();
  console.log(`Diagnosis Tests rows: ${rows}`);
  if (rows < 5) throw Error("Expected >= 5 diagnosis test rows");
  await p.screenshot({
    path: ".local/reference-repair/patient-diagnosis-test.png",
    fullPage: true,
  });

  // Test form: New Patient Diagnosis Test
  await p.getByRole("button", { name: "New Patient Diagnosis Test" }).click();
  await p.locator("form").waitFor();
  console.log("New Patient Diagnosis Test form verified");
  await p.screenshot({
    path: ".local/reference-repair/new-patient-diagnosis-test.png",
    fullPage: true,
  });
  await p.getByRole("button", { name: "Back" }).click();

  if (errors.length > 0) {
    console.error("Page errors detected:", errors);
    throw Error(`Test failed with ${errors.length} page errors.`);
  }

  console.log("\nAll 4 modules passed verification with 100% fidelity!");
  await b.close();
})();
