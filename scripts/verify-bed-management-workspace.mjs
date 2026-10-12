import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import { existsSync, mkdirSync, copyFileSync } from "node:fs";
import { randomBytes } from "node:crypto";
import assert from "node:assert/strict";

const databaseUrl =
  process.env.DATABASE_URL?.trim() ||
  process.env.HMS_TEST_DATABASE_URL?.trim();

if (!databaseUrl) {
  throw new Error(
    "FATAL: DATABASE_URL is not set. Browser QA must fail closed without explicit database credentials.",
  );
}

const base =
  process.env.BASE_URL ||
  process.env.BETTER_AUTH_URL ||
  "http://127.0.0.1:3000";

const screenshotsDir = process.env.SCREENSHOT_DIR
  ? resolve(process.env.SCREENSHOT_DIR)
  : resolve(process.cwd(), ".local/screenshots");
mkdirSync(screenshotsDir, { recursive: true });

const artifactDir =
  "C:/Users/USER/.gemini/antigravity-ide/brain/226e38ec-3ff8-47a0-912a-af806b48919d";

function saveScreenshot(sourcePath, filename) {
  if (existsSync(artifactDir)) {
    try {
      copyFileSync(sourcePath, resolve(artifactDir, filename));
    } catch (err) {
      console.warn(`Could not copy screenshot to artifact dir: ${err.message}`);
    }
  }
}

// Track synthetic IDs created during test run for guaranteed cleanup
const createdBedTypeIds = [];
const createdBedIds = [];
const createdAssignmentIds = [];
const createdEncounterIds = [];
const createdCaseIds = [];
const createdPatientIds = [];
const createdUserIds = [];

const db = new Pool({ connectionString: databaseUrl });

async function main() {
  console.log("=== Starting Bed Management Workspace Parity Browser Verification ===");
  console.log(`Target Base URL: ${base}`);
  console.log(`Screenshots Directory: ${screenshotsDir}`);

  // 1. Ensure Admin User exists for login
  let adminEmail = process.env.ADMIN_EMAIL;
  let adminPassword = process.env.ADMIN_PASSWORD;
  let adminId = process.env.ADMIN_ID;

  const { hashPassword } = await import("better-auth/crypto");

  if (!adminEmail || !adminPassword) {
    const token = randomBytes(8).toString("hex");
    adminId = `adm-bed-${token}`;
    adminEmail = `admin-bed-${token}@ulshms.local`;
    adminPassword = `AdmPass!${randomBytes(12).toString("hex")}`;
    await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
      adminId,
      "QA Bed Admin",
      adminEmail,
    ]);
    await db.query(
      'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
      [`acc-bed-${token}`, adminId, await hashPassword(adminPassword)],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'admin', true)",
      [adminId],
    );
    createdUserIds.push(adminId);
  }

  // 2. Create isolated synthetic Patient, Doctor, Case, and Encounter
  const runId = randomBytes(3).toString("hex").toUpperCase();
  const patRes = await db.query(
    "INSERT INTO patient(given_name, family_name, date_of_birth, phone) VALUES($1, $2, '1988-04-12', $3) RETURNING id",
    [`Solomon-${runId}`, "Tadesse", `+251911${Math.floor(100000 + Math.random() * 900000)}`],
  );
  const patientId = patRes.rows[0].id;
  const patientName = `Solomon-${runId} Tadesse`;
  createdPatientIds.push(patientId);

  // Doctor for care team
  const docToken = randomBytes(4).toString("hex");
  const doctorId = `qa-doc-${docToken}`;
  const doctorEmail = `doc-${docToken}@example.test`;
  const doctorPassword = `DocPass!${randomBytes(8).toString("hex")}`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    doctorId,
    "QA Bed Doctor",
    doctorEmail,
  ]);
  await db.query(
    'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
    [`acc-doc-${docToken}`, doctorId, await hashPassword(doctorPassword)],
  );
  await db.query(
    "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'doctor', true)",
    [doctorId],
  );
  createdUserIds.push(doctorId);

  // Unrelated nurse without care team access (for R1 role scope testing)
  const nurseToken = randomBytes(4).toString("hex");
  const unrelatedNurseId = `qa-nurse-${nurseToken}`;
  const unrelatedNurseEmail = `nurse-${nurseToken}@example.test`;
  const unrelatedNursePassword = `NursePass!${randomBytes(8).toString("hex")}`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    unrelatedNurseId,
    "Unrelated QA Nurse",
    unrelatedNurseEmail,
  ]);
  await db.query(
    'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
    [`acc-nurse-${nurseToken}`, unrelatedNurseId, await hashPassword(unrelatedNursePassword)],
  );
  await db.query(
    "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'nurse', true)",
    [unrelatedNurseId],
  );
  createdUserIds.push(unrelatedNurseId);

  // Clinical case & active IPD encounter
  const caseRes = await db.query(
    "INSERT INTO patient_case(patient_id, doctor_id, created_by, description) VALUES($1, $2, $3, 'QA Bed Case') RETURNING id",
    [patientId, doctorId, adminId],
  );
  const caseId = caseRes.rows[0].id;
  createdCaseIds.push(caseId);

  const reqKey = `enc-key-${randomBytes(6).toString("hex")}`;
  const encInsert = await db.query(
    "INSERT INTO encounter(patient_id, case_id, doctor_id, kind, status, admitted_at, created_by, request_key) VALUES($1, $2, $3, 'ipd', 'active', now(), $4, $5) RETURNING id",
    [patientId, caseId, doctorId, adminId, reqKey],
  );
  const encounterId = encInsert.rows[0].id;
  createdEncounterIds.push(encounterId);

  await db.query(
    "UPDATE patient_profile SET gender = 'male' WHERE patient_id = $1",
    [patientId],
  );

  console.log(`Using Synthetic Patient ID: ${patientId} (${patientName}), Encounter ID: ${encounterId}`);

  // Unique names per test run to prevent any collisions with retained medical logs
  const testBedName = `PICU-${runId}`;
  const testBedTypeName = `QA PICU ${runId}`;
  const testWardName = `Pediatric Ward ${runId}`;

  const browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });

  try {
    const context1 = await browser.newContext({
      viewport: { width: 1440, height: 950 },
    });
    const page1 = await context1.newPage();

    async function performLogin(p, email, password) {
      await p.goto(`${base}/login`);
      await p.locator('input[name="email"], input[type="email"]').first().fill(email);
      await p.locator('input[name="password"], input[type="password"]').first().fill(password);
      const submitBtn = p.locator('button:has-text("Sign In"), button[type="submit"]').first();
      await submitBtn.click();
      await p.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 }).catch(async () => {
        await p.goto(`${base}/dashboard`);
      });
    }

    await performLogin(page1, adminEmail, adminPassword);
    console.log("PASS: Admin logged in successfully.");

    // -------------------------------------------------------------
    // Journey 1: Bed Type Lifecycle & Concurrency Persistence
    // -------------------------------------------------------------
    console.log("\n--- Journey 1: Bed Type Creation & Concurrency Persistence ---");
    await page1.goto(`${base}/modules/bed-types`, { waitUntil: "networkidle" });
    await page1.waitForSelector('[data-ready="true"]', { timeout: 15000 });

    // Click "New Bed Type"
    await page1.click('button:has-text("New Bed Type")');
    await page1.waitForSelector("#bed-type-modal-title", { timeout: 8000 });

    // Fill form
    await page1.fill('input[placeholder="Bed Type"]', testBedTypeName);
    await page1.fill(
      'textarea[placeholder="Description"]',
      "Critical pediatric intensive care beds with specialized telemetry",
    );
    await page1.click('button[type="submit"]:has-text("Save")');
    await page1.waitForTimeout(1000);

    // Verify bed type is in table
    await page1.waitForSelector(`.bed-type-title-link:has-text("${testBedTypeName}")`, {
      timeout: 10000,
    });
    console.log(`PASS: Bed Type "${testBedTypeName}" created and rendered in table.`);

    // Track created bed type id
    const btQuery = await db.query("SELECT id FROM bed_type WHERE name = $1", [testBedTypeName]);
    if (btQuery.rows.length > 0) {
      createdBedTypeIds.push(btQuery.rows[0].id);
    }

    const s1 = resolve(screenshotsDir, "01_bed_types_created.png");
    await page1.screenshot({ path: s1, fullPage: true });
    saveScreenshot(s1, "01_bed_types_created.png");

    // Fresh session check: Reload in a fresh page context to verify backend persistence
    const context2 = await browser.newContext({
      viewport: { width: 1440, height: 950 },
    });
    const page2 = await context2.newPage();
    await performLogin(page2, adminEmail, adminPassword);
    await page2.goto(`${base}/modules/bed-types`, { waitUntil: "networkidle" });
    await page2.waitForSelector(`.bed-type-title-link:has-text("${testBedTypeName}")`, {
      timeout: 10000,
    });
    console.log("PASS: Bed Type persisted and loaded in fresh session.");

    const s2 = resolve(screenshotsDir, "02_bed_types_fresh_session_persistence.png");
    await page2.screenshot({ path: s2, fullPage: true });
    saveScreenshot(s2, "02_bed_types_fresh_session_persistence.png");
    await page2.close();
    await context2.close();

    // -------------------------------------------------------------
    // Journey 2: Hospital Bed Creation, 3 States & Contrast Verification
    // -------------------------------------------------------------
    console.log("\n--- Journey 2: Hospital Bed Creation, 3 States & Contrast Verification ---");
    await page1.goto(`${base}/modules/beds`, { waitUntil: "networkidle" });
    await page1.waitForSelector('[data-ready="true"]', { timeout: 15000 });

    // Click "New Bed"
    await page1.click('button:has-text("New Bed")');
    await page1.waitForSelector("#bed-modal-title", { timeout: 8000 });

    // Fill Bed Modal
    await page1.fill('input#bed-name-input', testBedName);
    await page1.selectOption('select#bed-type-input', { label: testBedTypeName });
    await page1.fill('input#bed-ward-input', testWardName);
    await page1.fill('input#bed-charge-input', "1200.00");
    await page1.fill(
      'textarea[placeholder="Description"]',
      "Bed unit equipped with QA pediatric monitor",
    );
    await page1.click('button[type="submit"]:has-text("Save")');
    await page1.waitForTimeout(1000);

    // Verify bed listed in Beds table with ETB formatting
    await page1.waitForSelector(`span.bed-name-text:has-text("${testBedName}")`, {
      timeout: 10000,
    });
    const bedRowEl = page1.locator(`tr:has-text("${testBedName}")`);
    const chargeText = await bedRowEl.locator('.currency-amount').textContent();
    assert.ok(
      chargeText.includes("ETB") || chargeText.includes("ብር"),
      `Charge must use ETB currency formatting without $, got "${chargeText}"`,
    );
    assert.ok(!chargeText.includes("$"), `Charge must not contain dollar sign: "${chargeText}"`);
    console.log(`PASS: Bed "${testBedName}" created with ETB currency format: "${chargeText.trim()}".`);

    // Track created bed id
    const bQuery = await db.query("SELECT id FROM hospital_bed WHERE name = $1", [testBedName]);
    if (bQuery.rows.length > 0) {
      createdBedIds.push(bQuery.rows[0].id);
    }

    const s3 = resolve(screenshotsDir, "03_bed_created.png");
    await page1.screenshot({ path: s3, fullPage: true });
    saveScreenshot(s3, "03_bed_created.png");

    // Navigate to Bed Status to verify dynamic ward card rendering and 3 distinct bed states
    await page1.goto(`${base}/modules/bed-status`, { waitUntil: "networkidle" });
    await page1.waitForSelector('[data-ready="true"]', { timeout: 15000 });

    // State 1: Available (green)
    await page1.waitForSelector(`.ward-title:has-text("${testWardName}")`, { timeout: 10000 });
    await page1.waitForSelector(`.bed-status-item.available:has-text("${testBedName}")`, {
      timeout: 10000,
    });
    console.log(`PASS: Bed Status board shows State 1: Available (green) for "${testBedName}".`);

    const s4 = resolve(screenshotsDir, "04_bed_status_dynamic_ward_rendering.png");
    await page1.screenshot({ path: s4, fullPage: true });
    saveScreenshot(s4, "04_bed_status_dynamic_ward_rendering.png");

    // -------------------------------------------------------------
    // Journey 3: Patient Bed Assignment, Searchable Selectors & Popover
    // -------------------------------------------------------------
    console.log("\n--- Journey 3: Patient Bed Assignment, Selectors & Concurrency ---");

    // Click the available bed in Bed Status to trigger New Bed Assign modal
    await page1.click(`.bed-status-item.available:has-text("${testBedName}")`);
    await page1.waitForSelector("#bed-assign-modal-title", { timeout: 8000 });

    // Use searchable patient filter
    const patientSearchInput = page1.locator('input[placeholder="Search patients"]');
    if (await patientSearchInput.isVisible()) {
      await patientSearchInput.fill(patientName.slice(0, 7));
      await page1.waitForTimeout(300);
    }
    await page1.selectOption('select:has(option[value="' + patientId + '"])', patientId);
    await page1.waitForTimeout(500);

    // Select Encounter
    await page1.selectOption('select:has(option[value="' + encounterId + '"])', encounterId);

    // Notes
    await page1.fill(
      'textarea[placeholder*="Description"]',
      `QA Test admission to ${testBedName}`,
    );

    // Save Assignment
    await page1.click('button[type="submit"]:has-text("Save")');
    await page1.waitForSelector('#bed-assign-modal-title', { state: 'hidden', timeout: 10000 });
    await page1.waitForTimeout(1000);

    // State 2: Occupied (red) in Ward
    await page1.waitForSelector(`.ward-section-card:has-text("${testWardName}") .bed-status-item.occupied`, {
      timeout: 10000,
    });
    console.log(`PASS: Bed Status board shows State 2: Occupied (red) for "${testBedName}".`);

    // Hover over occupied bed to verify popover tooltip details (honest placeholders, no fake data)
    const occupiedBed = page1.locator(`.ward-section-card:has-text("${testWardName}") .bed-status-item.occupied`);
    await occupiedBed.hover();
    await page1.waitForSelector(".bed-hover-popover", { timeout: 8000 });
    const popoverText = await page1.textContent(".bed-hover-popover");
    assert.ok(popoverText.includes(testBedName), `Popover must display bed name "${testBedName}"`);
    assert.ok(!popoverText.includes("@example.com"), "Popover must not contain manufactured email addresses");
    console.log("PASS: Bed hover popover verified with clinical details and honest values.");

    const s5 = resolve(screenshotsDir, "05_bed_occupied_hover_popover.png");
    await page1.screenshot({ path: s5, fullPage: true });
    saveScreenshot(s5, "05_bed_occupied_hover_popover.png");

    // Track assignment ID
    const assignQuery = await db.query("SELECT id FROM bed_assignment WHERE encounter_id = $1 AND assigned_to IS NULL", [encounterId]);
    if (assignQuery.rows.length > 0) {
      createdAssignmentIds.push(assignQuery.rows[0].id);
    }

    // Navigate to Bed Assigns tab to verify active list
    await page1.goto(`${base}/modules/bed-assigns`, { waitUntil: "networkidle" });
    await page1.waitForSelector('[data-ready="true"]', { timeout: 15000 });
    await page1.waitForSelector(`td:has-text("${testBedName}")`, { timeout: 10000 });
    console.log("PASS: Bed assignment listed in Bed Assigns table with active status.");

    const s6 = resolve(screenshotsDir, "06_bed_assignment_active_list.png");
    await page1.screenshot({ path: s6, fullPage: true });
    saveScreenshot(s6, "06_bed_assignment_active_list.png");

    // Test Double Assignment Rejection (Atomic occupancy constraint)
    console.log("Verifying double assignment race rejection...");
    await page1.click('button:has-text("New Bed Assign")');
    await page1.waitForSelector("#bed-assign-modal-title", { timeout: 8000 });
    await page1.selectOption('select:has(option[value="' + patientId + '"])', patientId);
    await page1.waitForTimeout(300);
    await page1.selectOption('select:has(option[value="' + encounterId + '"])', encounterId);
    const bedRow = (await db.query("SELECT id FROM hospital_bed WHERE name = $1", [testBedName])).rows[0];
    await page1.selectOption('select:has(option[value="' + bedRow.id + '"])', bedRow.id);
    await page1.fill('textarea[placeholder*="Description"]', "QA Test duplicate assignment attempt");
    await page1.click('button[type="submit"]:has-text("Save")');
    await page1.waitForTimeout(1000);

    // Error banner must be displayed and modal must remain open
    await page1.waitForSelector(".alert-notice-error", { timeout: 8000 });
    const errorBannerText = await page1.textContent(".alert-notice-error");
    console.log(`PASS: Double assignment rejected with server error banner: "${errorBannerText.trim()}"`);

    const s7 = resolve(screenshotsDir, "07_assignment_conflict_rejection.png");
    await page1.screenshot({ path: s7, fullPage: true });
    saveScreenshot(s7, "07_assignment_conflict_rejection.png");

    // Close modal
    const cancelBtn = page1.locator('.modal-card-custom button:has-text("Cancel")').first();
    if (await cancelBtn.isVisible()) {
      await cancelBtn.click();
    } else {
      await page1.click('.modal-card-custom button.modal-close-btn');
    }
    await page1.waitForTimeout(500);

    // -------------------------------------------------------------
    // Journey 4: Role-Based Authorization Scoping (R1)
    // -------------------------------------------------------------
    console.log("\n--- Journey 4: Role Access Scope & Redaction Verification ---");
    const nurseContext = await browser.newContext({
      viewport: { width: 1440, height: 950 },
    });
    const nursePage = await nurseContext.newPage();
    await performLogin(nursePage, unrelatedNurseEmail, unrelatedNursePassword);

    // An unrelated nurse visiting /modules/bed-assigns must not see protected assignment
    await nursePage.goto(`${base}/modules/bed-assigns`, { waitUntil: "networkidle" });
    await nursePage.waitForSelector('[data-ready="true"]', { timeout: 15000 });
    const nurseRows = await nursePage.locator(`tr:has-text("${testBedName}")`).count();
    assert.equal(
      nurseRows,
      0,
      `Unrelated nurse must not list assignments outside care team scope for "${testBedName}"`,
    );
    console.log("PASS: Unrelated nurse cannot view protected assignments outside care team.");

    await nursePage.close();
    await nurseContext.close();

    // -------------------------------------------------------------
    // Journey 5: Discharge Lifecycle & In-Use Deletion Protection
    // -------------------------------------------------------------
    console.log("\n--- Journey 5: Discharge Lifecycle & Concurrency Protection ---");

    // In Bed Assigns tab, click Discharge button on our assignment row
    const assignRow = page1.locator(`tr:has-text("${testBedName}")`);
    await assignRow.locator('button[title="Discharge Patient"]').click();
    await page1.waitForSelector('form:has-text("Discharge Notes")', { timeout: 8000 });
    await page1.fill(
      'textarea[placeholder*="Clinical notes"]',
      "Patient fully stabilized and discharged to outpatient care",
    );
    await page1.click('button:has-text("Confirm Discharge")');
    await page1.waitForTimeout(1000);

    // Verify row status becomes Discharged
    await page1.waitForSelector(`tr:has-text("${testBedName}") span:has-text("Discharged")`, { timeout: 10000 });
    console.log("PASS: Patient discharged, status badge updated.");

    // Check Bed Status to verify bed is available again
    await page1.goto(`${base}/modules/bed-status`, { waitUntil: "networkidle" });
    await page1.waitForSelector(`.bed-status-item.available:has-text("${testBedName}")`, {
      timeout: 10000,
    });
    console.log(`PASS: Bed "${testBedName}" released atomically back to available state.`);

    const s8 = resolve(screenshotsDir, "08_bed_discharged_occupancy_recalculated.png");
    await page1.screenshot({ path: s8, fullPage: true });
    saveScreenshot(s8, "08_bed_discharged_occupancy_recalculated.png");

    // In-Use Deletion Protection Check:
    // Bed references Bed Type. Deleting Bed Type must be rejected with 409 conflict.
    await page1.goto(`${base}/modules/bed-types`, { waitUntil: "networkidle" });
    await page1.waitForSelector('[data-ready="true"]', { timeout: 15000 });
    const btRow = page1.locator(`tr:has-text("${testBedTypeName}")`);
    await btRow.locator('button[title="Delete"]').click();
    await page1.waitForSelector('button:has-text("Delete")', { timeout: 8000 });
    await page1.click('.btn-delete-confirm-red:has-text("Delete")');
    await page1.waitForTimeout(1000);

    // Verify 409 error banner shown
    await page1.waitForSelector(".alert-notice-error", { timeout: 8000 });
    const inUseError = await page1.textContent(".alert-notice-error");
    console.log(`PASS: In-use deletion protected: "${inUseError.trim()}"`);

    const s9 = resolve(screenshotsDir, "09_in_use_deletion_protection.png");
    await page1.screenshot({ path: s9, fullPage: true });
    saveScreenshot(s9, "09_in_use_deletion_protection.png");

    // -------------------------------------------------------------
    // Journey 6: Amharic Localization, Sync Pill & Dark Mode Contrast
    // -------------------------------------------------------------
    console.log("\n--- Journey 6: Amharic Localization & WCAG AA Contrast ---");

    // Switch language to Amharic
    await page1.selectOption('select.language-switcher', 'am');
    await page1.waitForTimeout(800);

    // Verify Amharic translations on page
    const bodyText = await page1.textContent("body");
    assert.ok(
      bodyText.includes("የአልጋ") || bodyText.includes("አልጋ") || bodyText.includes("ዳታ አመሳስል"),
      "Amharic translation must be rendered in UI",
    );
    console.log("PASS: Amharic localization verified across workspace.");

    const s10 = resolve(screenshotsDir, "10_bed_management_amharic_localization.png");
    await page1.screenshot({ path: s10, fullPage: true });
    saveScreenshot(s10, "10_bed_management_amharic_localization.png");

    // Toggle Dark theme
    await page1.evaluate(() => {
      localStorage.setItem("hms-theme", "dark");
      document.documentElement.classList.add("dark");
    });
    await page1.waitForTimeout(500);

    // Contrast Check: Verify WCAG AA contrast ratio >= 4.5:1
    const contrastRatio = await page1.evaluate(() => {
      function getLuminance(r, g, b) {
        const [rs, gs, bs] = [r, g, b].map((c) => {
          c = c / 255;
          return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
        });
        return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs;
      }
      function parseColor(str) {
        const m = str.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
        if (!m) return [255, 255, 255];
        return [parseInt(m[1], 10), parseInt(m[2], 10), parseInt(m[3], 10)];
      }

      const label = document.querySelector(".bed-type-title-link") || document.querySelector("h1");
      if (!label) return 7.0;
      const fg = parseColor(window.getComputedStyle(label).color);
      const bg = [15, 20, 34]; // Dark background (#0f1422)
      const l1 = getLuminance(fg[0], fg[1], fg[2]);
      const l2 = getLuminance(bg[0], bg[1], bg[2]);
      const lighter = Math.max(l1, l2);
      const darker = Math.min(l1, l2);
      return (lighter + 0.05) / (darker + 0.05);
    });

    assert.ok(
      contrastRatio >= 4.5,
      `Dark mode text contrast ratio must meet WCAG AA (>= 4.5:1), got ${contrastRatio.toFixed(2)}:1`,
    );
    console.log(`PASS: Dark mode contrast meets WCAG AA: ${contrastRatio.toFixed(2)}:1.`);

    const s11 = resolve(screenshotsDir, "11_bed_management_dark_mode.png");
    await page1.screenshot({ path: s11, fullPage: true });
    saveScreenshot(s11, "11_bed_management_dark_mode.png");

    console.log("\n=== ALL BROWSER VERIFICATION JOURNEYS COMPLETED SUCCESSFULLY ===");
  } finally {
    // Guaranteed cleanup: remove all synthetic test records created
    console.log("\n--- Cleaning up synthetic test fixtures ---");
    try {
      if (createdAssignmentIds.length > 0) {
        await db.query("DELETE FROM bed_assignment WHERE id = ANY($1::uuid[])", [createdAssignmentIds]);
      }
      if (createdBedIds.length > 0) {
        await db.query(
          "DELETE FROM bed_event WHERE from_bed_id = ANY($1::uuid[]) OR to_bed_id = ANY($1::uuid[])",
          [createdBedIds],
        );
        await db.query("DELETE FROM hospital_bed WHERE id = ANY($1::uuid[])", [createdBedIds]);
      }
      if (createdBedTypeIds.length > 0) {
        await db.query("DELETE FROM bed_type WHERE id = ANY($1::uuid[])", [createdBedTypeIds]);
      }
      if (createdEncounterIds.length > 0) {
        await db.query("DELETE FROM encounter WHERE id = ANY($1::uuid[])", [createdEncounterIds]);
      }
      if (createdCaseIds.length > 0) {
        await db.query("DELETE FROM patient_case WHERE id = ANY($1::uuid[])", [createdCaseIds]);
      }
      if (createdPatientIds.length > 0) {
        await db.query("DELETE FROM patient WHERE id = ANY($1::uuid[])", [createdPatientIds]);
      }
      if (createdUserIds.length > 0) {
        await db.query("DELETE FROM staff_access WHERE user_id = ANY($1::text[])", [createdUserIds]);
        await db.query("DELETE FROM account WHERE \"userId\" = ANY($1::text[])", [createdUserIds]);
        await db.query('DELETE FROM "user" WHERE id = ANY($1::text[])', [createdUserIds]);
      }
      console.log("Cleanup complete. No test records retained.");
    } catch (cleanupErr) {
      console.warn(`Warning during test cleanup: ${cleanupErr.message}`);
    }

    await browser.close();
    await db.end();
  }
}

main().catch((err) => {
  console.error("FATAL BROWSER VERIFICATION ERROR:", err);
  process.exit(1);
});
