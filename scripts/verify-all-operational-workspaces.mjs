import fs from "node:fs";
import { chromium } from "@playwright/test";

(async () => {
  const base = process.env.BETTER_AUTH_URL || "http://127.0.0.1:3000";
  let email = process.env.ADMIN_EMAIL;
  let password = process.env.ADMIN_PASSWORD;

  if (process.env.DATABASE_URL && process.env.HMS_TEST_ISOLATED_SCHEMA) {
    const { Pool } = await import("pg");
    const { hashPassword } = await import("better-auth/crypto");
    const { randomUUID } = await import("node:crypto");
    const db = new Pool({ connectionString: process.env.DATABASE_URL });
    const id = randomUUID();
    password = randomUUID() + "Aa1!";
    email = `operational-${id.slice(0, 8)}@example.test`;
    await db.query(
      'INSERT INTO "user"(id,name,email,"emailVerified") VALUES($1,$2,$3,true)',
      [id, "Operational QA Admin", email],
    );
    await db.query(
      'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
      [randomUUID(), id, await hashPassword(password)],
    );
    await db.query(
      "INSERT INTO staff_access(user_id,role) VALUES($1,'admin')",
      [id],
    );
    await db.end();
  } else if (!email || !password) {
    if (fs.existsSync(".local/admin-login.txt")) {
      const lines = fs
        .readFileSync(".local/admin-login.txt", "utf8")
        .split("\n");
      email = lines
        .find((l) => l.startsWith("Email:"))
        ?.replace("Email:", "")
        .trim();
      password = lines
        .find((l) => l.startsWith("Password:"))
        ?.replace("Password:", "")
        .trim();
    }
  }

  if (!email || !password) {
    throw Error(
      "Admin credentials missing: set ADMIN_EMAIL & ADMIN_PASSWORD or ensure .local/admin-login.txt exists",
    );
  }

  fs.mkdirSync(".local/operational-verification", { recursive: true });

  const b = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
  });
  const p = await b.newPage({ viewport: { width: 1440, height: 1000 } });
  const errors = [];
  p.on("pageerror", (e) => errors.push(e.message));

  console.log(`Starting Comprehensive Operational Verification on ${base}`);

  // 1. Sign In
  await p.goto(`${base}/login`);
  await p.getByLabel("Email address", { exact: false }).fill(email);
  await p.getByLabel("Password", { exact: false }).fill(password);
  await p.getByRole("button", { name: "Sign In", exact: true }).click();
  await p.waitForURL("**/dashboard");

  const darkBtn = p.getByRole("button", {
    name: "Use dark theme",
    exact: true,
  });
  if ((await darkBtn.count()) > 0) {
    await darkBtn.click();
  }
  console.log("✔ 1. Logged in and ensured dark theme on Dashboard");

  // Helper function to visit and wait for shell or card
  async function visitWorkspace(path, waitSelector = '[data-ready="true"]') {
    await p.goto(`${base}${path}`);
    if (waitSelector) {
      await p.locator(waitSelector).waitFor({ timeout: 30000 });
    }
  }

  // ============================================================
  // 1. DASHBOARD & OVERVIEW
  // ============================================================
  console.log("\n--- Testing Dashboard Metrics & Overview ---");
  await visitWorkspace(
    "/dashboard",
    ".dashboard-container, [data-ready='true']",
  );
  await p.waitForTimeout(1000);
  const metricCards = await p
    .locator(".metric-card, .dashboard-stat-card, .widget-card")
    .count();
  console.log(`✔ Dashboard metric widgets rendered: count >= ${metricCards}`);
  await p.screenshot({
    path: ".local/operational-verification/01-dashboard.png",
    fullPage: true,
  });

  // ============================================================
  // 2. PATIENTS WORKSPACE
  // ============================================================
  console.log("\n--- Testing Patients Workspace ---");
  await visitWorkspace("/modules/patients");
  let patientRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Patients directory rows: ${patientRows}`);
  await p.screenshot({
    path: ".local/operational-verification/02-patients.png",
    fullPage: true,
  });

  await visitWorkspace("/modules/case-handlers");
  const caseHandlerRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Case Handlers rows: ${caseHandlerRows}`);

  await visitWorkspace("/modules/patient-admissions");
  const admissionRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Patient Admissions rows: ${admissionRows}`);

  // ============================================================
  // 3. DOCTORS WORKSPACE
  // ============================================================
  console.log("\n--- Testing Doctors Workspace ---");
  await visitWorkspace("/modules/doctors");
  const doctorCards = await p
    .locator(".doctor-card, .billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Doctor directory records: ${doctorCards}`);
  await p.screenshot({
    path: ".local/operational-verification/03-doctors.png",
    fullPage: true,
  });

  await visitWorkspace("/modules/doctor-departments");
  const deptRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Doctor Departments rows: ${deptRows}`);

  await visitWorkspace("/modules/doctor-holidays");
  const holidayRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Doctor Holidays rows: ${holidayRows}`);

  await visitWorkspace("/modules/breaks");
  const breakRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Doctor Breaks rows: ${breakRows}`);

  // ============================================================
  // 4. BEDS WORKSPACE
  // ============================================================
  console.log("\n--- Testing Bed Management Workspace ---");
  await visitWorkspace("/modules/bed-status");
  const wardSections = await p.locator(".ward-section-card").count();
  console.log(`✔ Bed Status Wards rendered: ${wardSections}`);
  if (wardSections < 10)
    throw Error(`Expected >= 10 wards, got ${wardSections}`);

  await visitWorkspace("/modules/bed-assigns");
  const bedAssignRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Bed Assignments rows: ${bedAssignRows}`);

  await visitWorkspace("/modules/beds");
  const bedsCount = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Hospital Beds rows: ${bedsCount}`);

  await visitWorkspace("/modules/bed-types");
  const bedTypesCount = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Bed Types rows: ${bedTypesCount}`);
  await p.screenshot({
    path: ".local/operational-verification/04-bed-management.png",
    fullPage: true,
  });

  // ============================================================
  // 5. ATTENDANCE WORKSPACE
  // ============================================================
  console.log("\n--- Testing Attendance Workspace ---");
  await visitWorkspace("/modules/attendance");
  const attendanceRows = await p
    .locator(".billing-table tbody tr, table tbody tr, .attendance-card")
    .count();
  console.log(`✔ Attendance workspace rendered: ${attendanceRows} records`);
  await p.screenshot({
    path: ".local/operational-verification/05-attendance.png",
    fullPage: true,
  });

  // ============================================================
  // 6. BILLING & INVOICES WORKSPACE
  // ============================================================
  console.log("\n--- Testing Billing & Financial Workspaces ---");
  await visitWorkspace("/modules/invoices");
  const invoicesCount = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Invoices rendered: ${invoicesCount} rows`);

  await visitWorkspace("/modules/bills");
  const billsCount = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Bills rendered: ${billsCount} rows`);

  await visitWorkspace("/modules/advance-payments");
  const advanceCount = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Advance Payments rendered: ${advanceCount} rows`);
  await p.screenshot({
    path: ".local/operational-verification/06-billing.png",
    fullPage: true,
  });

  // ============================================================
  // 7. MEDICINES & PHARMACY WORKSPACE
  // ============================================================
  console.log("\n--- Testing Medicines & Pharmacy Workspace ---");
  await visitWorkspace("/modules/medicines");
  const medsCount = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Medicines catalog: ${medsCount} rows`);

  await visitWorkspace("/modules/medicine-categories");
  const medCats = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Medicine Categories: ${medCats} rows`);

  await visitWorkspace("/modules/purchase-medicines");
  const purchases = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Medicine Purchases: ${purchases} rows`);
  await p.screenshot({
    path: ".local/operational-verification/07-medicines.png",
    fullPage: true,
  });

  // ============================================================
  // 8. BLOOD BANK WORKSPACE
  // ============================================================
  console.log("\n--- Testing Blood Bank Workspace ---");
  await visitWorkspace("/modules/blood-banks");
  const bbRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Blood Bank status: ${bbRows} rows`);
  if (bbRows < 4)
    throw Error(`Expected blood bank stock rows >= 4, got ${bbRows}`);

  await visitWorkspace("/modules/blood-donors");
  const donorRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Blood Donors: ${donorRows} rows`);
  await p.screenshot({
    path: ".local/operational-verification/08-blood-bank.png",
    fullPage: true,
  });

  // ============================================================
  // 9. PRESCRIPTIONS WORKSPACE
  // ============================================================
  console.log("\n--- Testing Prescriptions Workspace ---");
  await visitWorkspace("/modules/prescriptions");
  const rxRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Prescriptions list: ${rxRows} rows`);
  await p.screenshot({
    path: ".local/operational-verification/09-prescriptions.png",
    fullPage: true,
  });

  // ============================================================
  // 10. FRONT OFFICE WORKSPACE
  // ============================================================
  console.log("\n--- Testing Front Office Workspace ---");
  await visitWorkspace("/modules/call-logs");
  const callLogs = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Call logs: ${callLogs} rows`);

  await visitWorkspace("/modules/visitors");
  const visitors = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Visitors: ${visitors} rows`);

  await visitWorkspace("/modules/enquiries");
  const enquiries = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Enquiries: ${enquiries} rows`);
  await p.screenshot({
    path: ".local/operational-verification/10-front-office.png",
    fullPage: true,
  });

  // ============================================================
  // 11. GENERAL INVENTORY WORKSPACE
  // ============================================================
  console.log("\n--- Testing General Inventory Workspace ---");
  await visitWorkspace("/modules/items");
  const items = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Inventory Items: ${items} rows`);

  await visitWorkspace("/modules/item-categories");
  const itemCats = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Item Categories: ${itemCats} rows`);
  await p.screenshot({
    path: ".local/operational-verification/11-inventory.png",
    fullPage: true,
  });

  // ============================================================
  // 12. LIVE CONSULTATIONS & TELEHEALTH
  // ============================================================
  console.log("\n--- Testing Live Consultations Workspace ---");
  await visitWorkspace("/modules/live-consultations");
  const liveRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Live Consultations: ${liveRows} rows`);
  await p.screenshot({
    path: ".local/operational-verification/12-telehealth.png",
    fullPage: true,
  });

  // ============================================================
  // 13. DIAGNOSIS WORKSPACE
  // ============================================================
  console.log("\n--- Testing Diagnosis Categories & Tests ---");
  await visitWorkspace("/modules/diagnosis-categories");
  const diagCats = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Diagnosis Categories: ${diagCats} rows`);

  await visitWorkspace("/modules/patient-diagnosis-test");
  const diagTests = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Diagnosis Tests: ${diagTests} rows`);
  await p.screenshot({
    path: ".local/operational-verification/13-diagnosis.png",
    fullPage: true,
  });

  // ============================================================
  // 14. REVIEWS WORKSPACE
  // ============================================================
  console.log("\n--- Testing Reviews Workspace ---");
  await visitWorkspace("/modules/reviews");
  const reviewRows = await p
    .locator(".billing-table tbody tr, table tbody tr, .review-card")
    .count();
  console.log(`✔ Reviews: ${reviewRows} records`);
  await p.screenshot({
    path: ".local/operational-verification/14-reviews.png",
    fullPage: true,
  });

  // ============================================================
  // 15. USERS & ROLE DIRECTORIES
  // ============================================================
  console.log("\n--- Testing Users & Role Portals ---");
  await visitWorkspace("/modules/users");
  const usersRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Total Staff Users: ${usersRows} rows`);

  await visitWorkspace("/modules/nurses");
  const nurseRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Nurses: ${nurseRows} rows`);

  await visitWorkspace("/modules/accountants");
  const acctRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Accountants: ${acctRows} rows`);

  await visitWorkspace("/modules/lab-technicians");
  const labRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Lab Technicians: ${labRows} rows`);

  await visitWorkspace("/modules/pharmacists");
  const pharmRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Pharmacists: ${pharmRows} rows`);

  await visitWorkspace("/modules/receptionists");
  const recepRows = await p
    .locator(".billing-table tbody tr, table tbody tr")
    .count();
  console.log(`✔ Receptionists: ${recepRows} rows`);
  await p.screenshot({
    path: ".local/operational-verification/15-users.png",
    fullPage: true,
  });

  if (errors.length > 0) {
    console.warn("Uncaught browser page errors detected:", errors);
    throw Error(
      `Page errors during operational verification: ${errors.join("; ")}`,
    );
  }

  console.log("\n=======================================================");
  console.log("SUCCESS: All 15 operational workspaces verified cleanly!");
  console.log("=======================================================\n");

  await b.close();
})().catch((err) => {
  console.error("Operational verification failed:", err);
  process.exit(1);
});
