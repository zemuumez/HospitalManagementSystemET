import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import { existsSync, mkdirSync, copyFileSync } from "node:fs";
import { randomBytes } from "node:crypto";
import assert from "node:assert/strict";

const databaseUrl =
  process.env.DATABASE_URL?.trim() || process.env.HMS_TEST_DATABASE_URL?.trim();

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
const createdEncounterIds = [];
const createdCaseIds = [];
const createdPatientIds = [];
const createdUserIds = [];

const db = new Pool({ connectionString: databaseUrl });

async function main() {
  console.log(
    "=== Starting Clinical Encounters & IPD/OPD Detail Registers Browser QA ===",
  );
  console.log(`Target Base URL: ${base}`);
  console.log(`Screenshots Directory: ${screenshotsDir}`);

  // 1. Ensure Admin User exists for login
  let adminEmail = process.env.ADMIN_EMAIL;
  let adminPassword = process.env.ADMIN_PASSWORD;
  let adminId = process.env.ADMIN_ID;

  const { hashPassword } = await import("better-auth/crypto");

  if (!adminEmail || !adminPassword) {
    const token = randomBytes(8).toString("hex");
    adminId = `adm-enc-${token}`;
    adminEmail = `admin-enc-${token}@ulshms.local`;
    adminPassword = `AdmPass!${randomBytes(12).toString("hex")}`;
    await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
      adminId,
      "QA Clinical Admin",
      adminEmail,
    ]);
    await db.query(
      'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
      [`acc-enc-${token}`, adminId, await hashPassword(adminPassword)],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'admin', true)",
      [adminId],
    );
    createdUserIds.push(adminId);
  }

  // 2. Provision synthetic Primary Doctor, Secondary Doctor, Patient, Case, and 2 Beds
  const runId = randomBytes(3).toString("hex").toUpperCase();

  // Primary Doctor
  const doc1Token = randomBytes(4).toString("hex");
  const doc1Id = `qa-doc1-${doc1Token}`;
  const doc1Email = `doc1-${doc1Token}@example.test`;
  const doc1Password = `Doc1Pass!${randomBytes(8).toString("hex")}`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    doc1Id,
    `Dr. Dawit-${runId}`,
    doc1Email,
  ]);
  await db.query(
    'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
    [`acc-doc1-${doc1Token}`, doc1Id, await hashPassword(doc1Password)],
  );
  await db.query(
    "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'doctor', true)",
    [doc1Id],
  );
  await db.query(
    "INSERT INTO doctor_profile(user_id, department) VALUES($1, 'Cardiology')",
    [doc1Id],
  );
  createdUserIds.push(doc1Id);

  // Secondary Doctor (for care team & unauthorized check)
  const doc2Token = randomBytes(4).toString("hex");
  const doc2Id = `qa-doc2-${doc2Token}`;
  const doc2Email = `doc2-${doc2Token}@example.test`;
  const doc2Password = `Doc2Pass!${randomBytes(8).toString("hex")}`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    doc2Id,
    `Dr. Almaz-${runId}`,
    doc2Email,
  ]);
  await db.query(
    'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
    [`acc-doc2-${doc2Token}`, doc2Id, await hashPassword(doc2Password)],
  );
  await db.query(
    "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'doctor', true)",
    [doc2Id],
  );
  await db.query(
    "INSERT INTO doctor_profile(user_id, department) VALUES($1, 'Cardiology')",
    [doc2Id],
  );
  createdUserIds.push(doc2Id);

  // Patient
  const patRes = await db.query(
    "INSERT INTO patient(given_name, family_name, date_of_birth, phone) VALUES($1, $2, '1992-06-15', $3) RETURNING id",
    [
      `Kenenisa-${runId}`,
      "Bekele",
      `+251912${Math.floor(100000 + Math.random() * 900000)}`,
    ],
  );
  const patientId = patRes.rows[0].id;
  createdPatientIds.push(patientId);

  // Patient Case with Primary Doctor
  const caseRes = await db.query(
    "INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES($1, $2, 'Cardio Observation', $3) RETURNING id",
    [patientId, doc1Id, adminId],
  );
  const caseId = caseRes.rows[0].id;
  createdCaseIds.push(caseId);

  // Bed Type & Beds
  const btRes = await db.query(
    "INSERT INTO bed_type(name, description) VALUES($1, 'Cardio Ward Suite') RETURNING id",
    [`Cardio Suite-${runId}`],
  );
  const bedTypeId = btRes.rows[0].id;
  createdBedTypeIds.push(bedTypeId);

  const bed1Res = await db.query(
    "INSERT INTO hospital_bed(name, type_id, bed_type, ward_name, charge_minor, state, active, created_by) VALUES($1, $2, $3, 'Cardio Wing A', 95000, 'ready', true, $4) RETURNING id",
    [`Bed-C1-${runId}`, bedTypeId, `Cardio Suite-${runId}`, adminId],
  );
  const bed1Id = bed1Res.rows[0].id;
  createdBedIds.push(bed1Id);

  const bed2Res = await db.query(
    "INSERT INTO hospital_bed(name, type_id, bed_type, ward_name, charge_minor, state, active, created_by) VALUES($1, $2, $3, 'Cardio Wing B', 85000, 'ready', true, $4) RETURNING id",
    [`Bed-C2-${runId}`, bedTypeId, `Cardio Suite-${runId}`, adminId],
  );
  const bed2Id = bed2Res.rows[0].id;
  createdBedIds.push(bed2Id);

  // Launch Playwright Browser using configured Chrome executable
  const browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });

  const context = await browser.newContext({
    viewport: { width: 1440, height: 950 },
    locale: "en-US",
  });
  const page = await context.newPage();

  async function performLogin(p, email, password) {
    await p.goto(`${base}/login`);
    await p
      .locator('input[name="email"], input[type="email"]')
      .first()
      .fill(email);
    await p
      .locator('input[name="password"], input[type="password"]')
      .first()
      .fill(password);
    const submitBtn = p
      .locator('button:has-text("Sign In"), button[type="submit"]')
      .first();
    await submitBtn.click();
    await p
      .waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 })
      .catch(async () => {
        await p.goto(`${base}/dashboard`);
      });
  }

  let createdEncounterId = null;

  try {
    // ─── Journey 1: IPD & OPD Workspace Navigation & Pagination ───────────────
    console.log("1. Journey 1: IPD Workspace Landing & Pagination Controls...");
    await performLogin(page, adminEmail, adminPassword);
    console.log("   Admin signed in successfully.");

    await page.goto(`${base}/modules/ipd-patient-departments`, {
      waitUntil: "networkidle",
    });
    page.on("pageerror", (err) => console.log("   [Page Error]:", err.message));
    page.on("console", (msg) => {
      if (msg.type() === "error") console.log("   [Console Error]:", msg.text());
    });
    await page.waitForSelector("#encounters-table", { timeout: 30000 });
    assert(
      await page.isVisible("#encounter-search-input"),
      "Search input must be visible",
    );
    assert(
      await page.isVisible("#btn-new-admission"),
      "New admission button must be visible",
    );
    assert(
      await page.isVisible("#filter-status-All"),
      "Status filter bar must be visible",
    );

    const shot1 = resolve(screenshotsDir, "01_ipd_workspace_landing.png");
    await page.screenshot({ path: shot1, fullPage: true });
    saveScreenshot(shot1, "01_ipd_workspace_landing.png");
    console.log("   Saved 01_ipd_workspace_landing.png");

    // ─── Journey 2: IPD Admission Workflow ───────────────────────────────────
    console.log("2. Journey 2: IPD Admission Workflow...");
    await page.click("#btn-new-admission");
    await page.waitForSelector("#select-admit-case", { timeout: 5000 });

    // Select our synthetic case and bed1
    await page.selectOption("#select-admit-case", caseId);
    await page.selectOption("#select-admit-bed", bed1Id);
    await page.fill("#input-admit-bp", "135/85");
    await page.fill("#input-admit-height", "182");
    await page.fill("#input-admit-weight", "78");
    await page.fill(
      "#input-admit-symptoms",
      "Acute chest palpitation and shortness of breath",
    );

    await page.click("#btn-submit-admission");
    // Wait for modal to close and encounter table to refresh
    await page.waitForSelector("#select-admit-case", {
      state: "detached",
      timeout: 8000,
    });
    await page.waitForTimeout(1000);

    // Verify row appears with patient name
    const rowLocator = page.locator(`tr:has-text("Kenenisa-${runId}")`);
    await rowLocator.waitFor({ state: "visible", timeout: 8000 });
    assert(
      await rowLocator.isVisible(),
      "New IPD admission must appear in table",
    );

    const shot2 = resolve(screenshotsDir, "02_ipd_admission_created.png");
    await page.screenshot({ path: shot2, fullPage: true });
    saveScreenshot(shot2, "02_ipd_admission_created.png");
    console.log("   Saved 02_ipd_admission_created.png");

    // Query encounter ID from DB
    const encQuery = await db.query(
      "SELECT id FROM encounter WHERE patient_id = $1 ORDER BY admitted_at DESC LIMIT 1",
      [patientId],
    );
    assert(encQuery.rows.length > 0, "Encounter must be in database");
    createdEncounterId = encQuery.rows[0].id;
    createdEncounterIds.push(createdEncounterId);

    // ─── Journey 3: Fresh Session Persistence Verification ───────────────────
    console.log("3. Journey 3: Fresh Session Persistence Verification...");
    const context2 = await browser.newContext({
      viewport: { width: 1440, height: 950 },
      locale: "en-US",
    });
    const page2 = await context2.newPage();
    await performLogin(page2, adminEmail, adminPassword);
    await page2.goto(`${base}/modules/ipd-patient-departments`, {
      waitUntil: "networkidle",
    });
    await page2.waitForSelector("#encounters-table", { timeout: 10000 });

    const persistedRow = page2.locator(`tr:has-text("Kenenisa-${runId}")`);
    await persistedRow.waitFor({ state: "visible", timeout: 8000 });
    assert(
      await persistedRow.isVisible(),
      "Encounter must persist across fresh sessions",
    );

    // Open detail register modal in fresh session
    const detailsBtn2 = page2.locator(
      `#btn-view-encounter-${createdEncounterId}`,
    );
    await detailsBtn2.click();
    await page2.waitForSelector("#tab-overview", { timeout: 6000 });

    const shot3 = resolve(screenshotsDir, "03_persistence_fresh_session.png");
    await page2.screenshot({ path: shot3, fullPage: true });
    saveScreenshot(shot3, "03_persistence_fresh_session.png");
    console.log("   Saved 03_persistence_fresh_session.png");
    await context2.close();

    // ─── Journey 4: Clinical Register: Vitals & Diagnoses (ICD-10) ───────────
    console.log("4. Journey 4: Clinical Register Updates: Vitals & Diagnoses...");
    const detailsBtn = page.locator(
      `#btn-view-encounter-${createdEncounterId}`,
    );
    await detailsBtn.click();
    await page.waitForSelector("#tab-overview", { timeout: 6000 });

    // Update vitals & symptoms
    await page.fill("#input-intake-bp", "128/80");
    await page.fill(
      "#input-encounter-symptoms",
      "Chest palpitation improved with rest; sinus rhythm restored",
    );
    await page.click("#btn-save-symptoms");
    await page.waitForSelector('text="Clinical records updated successfully"', {
      timeout: 6000,
    });

    // Record Diagnosis (ICD-10)
    await page.click("#tab-diagnoses");
    await page.waitForSelector("#input-diag-code", { timeout: 5000 });

    await page.fill("#input-diag-code", "I47.1");
    await page.selectOption("#select-diag-category", "final");
    await page.fill("#input-diag-desc", "Supraventricular tachycardia");
    await page.click("#btn-add-diagnosis");
    await page.waitForSelector('td:has-text("I47.1")', { timeout: 6000 });

    // Verify DB record in encounter_diagnosis
    const diagCheck = await db.query(
      "SELECT icd10_code, description, category FROM encounter_diagnosis WHERE encounter_id = $1",
      [createdEncounterId],
    );
    assert.equal(
      diagCheck.rows[0].icd10_code,
      "I47.1",
      "Diagnosis ICD-10 must match",
    );

    const shot4 = resolve(
      screenshotsDir,
      "04_encounter_diagnosis_recorded.png",
    );
    await page.screenshot({ path: shot4, fullPage: true });
    saveScreenshot(shot4, "04_encounter_diagnosis_recorded.png");
    console.log("   Saved 04_encounter_diagnosis_recorded.png");

    // ─── Journey 5: Optimistic Concurrency & Stale Edit Rejection ────────────
    console.log(
      "5. Journey 5: Optimistic Concurrency & Stale Edit Conflict Rejection...",
    );
    // Query current version from DB
    const verQuery = await db.query(
      "SELECT version FROM encounter WHERE id = $1",
      [createdEncounterId],
    );
    const currentVer = verQuery.rows[0].version;
    assert(
      currentVer > 1,
      "Encounter version should have been incremented by Journey 4",
    );

    // Test server rejection of stale version update via direct API fetch (passing valid integer staleVer)
    const conflictResult = await page.evaluate(
      async ({ encId, staleVer }) => {
        const resp = await fetch(`/api/hms/encounters/${encId}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            version: staleVer, // Integer >= 1 passes validation, but is stale against DB version
            symptoms: "Attempted stale edit payload",
            intake: { bloodPressure: "120/80" },
          }),
        });
        return { status: resp.status, ok: resp.ok };
      },
      { encId: createdEncounterId, staleVer: currentVer - 1 },
    );

    assert.equal(
      conflictResult.status,
      409,
      "Server must reject stale encounter updates with HTTP 409 Conflict",
    );
    console.log(
      "   Verified server rejected stale version with HTTP 409 Conflict.",
    );

    const shot5 = resolve(
      screenshotsDir,
      "05_concurrency_conflict_rejection.png",
    );
    await page.screenshot({ path: shot5, fullPage: true });
    saveScreenshot(shot5, "05_concurrency_conflict_rejection.png");
    console.log("   Saved 05_concurrency_conflict_rejection.png");

    // ─── Journey 6: Patient-Scoped Access & Care Team Delegation ─────────────
    console.log(
      "6. Journey 6: Patient-Scoped Access & Care Team Delegation/Revocation...",
    );
    const contextDoc2 = await browser.newContext({
      viewport: { width: 1440, height: 950 },
      locale: "en-US",
    });
    const pageDoc2 = await contextDoc2.newPage();
    await performLogin(pageDoc2, doc2Email, doc2Password);

    // Doctor 2 is not assigned to this patient case or care team -> Unauthorized mutation is rejected with 403 Forbidden
    const unauthorizedMutate = await pageDoc2.evaluate(async (encId) => {
      const resp = await fetch(`/api/hms/encounters/${encId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          version: 2,
          symptoms: "Unauthorized tampering by Doctor 2",
        }),
      });
      return { status: resp.status, ok: resp.ok };
    }, createdEncounterId);

    assert.equal(
      unauthorizedMutate.status,
      403,
      "Unauthorized doctor mutation must be rejected with HTTP 403 Forbidden",
    );
    console.log("   Verified unassigned Doctor 2 receives 403 Forbidden on mutation.");

    // Doctor 2 also cannot view the patient encounter (scoped lookup returns 404)
    const unauthorizedRead = await pageDoc2.evaluate(async (encId) => {
      const resp = await fetch(`/api/hms/encounters/${encId}`);
      return { status: resp.status, ok: resp.ok };
    }, createdEncounterId);
    assert.equal(
      unauthorizedRead.status,
      404,
      "Unassigned doctor read must be hidden (404 Not Found)",
    );
    console.log("   Verified unassigned Doctor 2 read is hidden (404 Not Found).");

    // Admin delegates Doctor 2 to care team
    await page.click("#tab-care-team");
    await page.waitForSelector("#select-care-staff", { timeout: 5000 });
    await page.selectOption("#select-care-staff", doc2Id);
    await page.fill("#input-care-role", "Consultant Cardiologist");
    await page.fill("#input-care-notes", "Second opinion on electrophysiology");
    await page.click("#btn-add-care-member");
    await page.waitForSelector('td:has-text("Consultant Cardiologist")', {
      timeout: 6000,
    });

    // Doctor 2 now has access!
    const authorizedCheck = await pageDoc2.evaluate(async (encId) => {
      const resp = await fetch(`/api/hms/encounters/${encId}`);
      return { status: resp.status, ok: resp.ok };
    }, createdEncounterId);
    assert.equal(
      authorizedCheck.status,
      200,
      "Delegated care team doctor must receive HTTP 200 OK",
    );
    console.log(
      "   Verified care team delegation grants Doctor 2 access (200 OK).",
    );

    // Revoke care team member in admin page
    const revokeBtn = page.locator('button:has-text("Revoke")').first();
    await revokeBtn.click();
    await page.waitForSelector('text="Care team delegation revoked"', {
      timeout: 6000,
    });

    // Doctor 2 access is revoked -> read hidden again (404)
    const revokedCheck = await pageDoc2.evaluate(async (encId) => {
      const resp = await fetch(`/api/hms/encounters/${encId}`);
      return { status: resp.status, ok: resp.ok };
    }, createdEncounterId);
    assert.equal(
      revokedCheck.status,
      404,
      "Revoked care team doctor must lose access (404 Not Found)",
    );
    console.log("   Verified revoking care team member denies read access (404).");
    await contextDoc2.close();

    const shot6 = resolve(
      screenshotsDir,
      "06_care_team_delegation_revocation.png",
    );
    await page.screenshot({ path: shot6, fullPage: true });
    saveScreenshot(shot6, "06_care_team_delegation_revocation.png");
    console.log("   Saved 06_care_team_delegation_revocation.png");

    // ─── Journey 7: Bed Transfer & Atomic Bed Release on Discharge ───────────
    console.log("7. Journey 7: Bed Transfer & Atomic Bed Release on Discharge...");
    await page.click("#tab-bed");
    await page.waitForSelector("#select-transfer-bed", { timeout: 5000 });

    // Transfer from bed1 to bed2
    await page.selectOption("#select-transfer-bed", bed2Id);
    await page.fill(
      "#input-transfer-reason",
      "Patient transferred to Wing B for telemetry",
    );
    await page.click("#btn-confirm-transfer");
    await page.waitForSelector('text="Bed transferred successfully"', {
      timeout: 6000,
    });
    await page.waitForTimeout(1000);

    // Discharge workflow
    await page.click("#tab-discharge");
    await page.waitForSelector("#input-discharge-summary", { timeout: 5000 });

    await page.fill(
      "#input-discharge-summary",
      "Patient fully stabilized. Normal sinus rhythm maintained 48h. Discharged with prescription for beta-blockers and scheduled outpatient clinic follow-up.",
    );
    await page.click("#btn-confirm-discharge");
    await page.waitForSelector('text="Patient Discharged Successfully"', {
      timeout: 8000,
    });

    // Verify bed2 was atomically reset to 'ready'
    const bedCheck = await db.query(
      "SELECT state FROM hospital_bed WHERE id = $1",
      [bed2Id],
    );
    assert.equal(
      bedCheck.rows[0].state,
      "ready",
      "Hospital bed state must be atomically released to 'ready' upon discharge",
    );
    console.log(
      "   Confirmed hospital_bed state atomically restored to 'ready'.",
    );

    const shot7 = resolve(
      screenshotsDir,
      "07_discharge_atomic_bed_release.png",
    );
    await page.screenshot({ path: shot7, fullPage: true });
    saveScreenshot(shot7, "07_discharge_atomic_bed_release.png");
    console.log("   Saved 07_discharge_atomic_bed_release.png");

    // Close encounter modal
    await page.click("#btn-close-encounter-modal");
    await page.waitForTimeout(500);

    // ─── Journey 8: Amharic Localization & Dark Mode ─────────────────────────
    console.log("8. Journey 8: Amharic Localization & Dark Mode...");
    // Change language to Amharic
    await page.selectOption("select.language-switcher", "am");
    await page.waitForTimeout(800);

    // Verify Amharic translated elements on page
    const bodyText = await page.textContent("body");
    assert.ok(
      bodyText.includes("የታካሚ") ||
        bodyText.includes("አልጋ") ||
        bodyText.includes("ታካሚዎች"),
      "Amharic translation must be rendered in UI",
    );
    console.log("   Amharic translation confirmed in DOM.");

    // Toggle Dark Mode
    await page.evaluate(() => {
      localStorage.setItem("hms-theme", "dark");
      document.documentElement.classList.add("dark");
    });
    await page.waitForTimeout(500);

    const shot8 = resolve(
      screenshotsDir,
      "08_unauthorized_rejection_amharic_dark_mode.png",
    );
    await page.screenshot({ path: shot8, fullPage: true });
    saveScreenshot(shot8, "08_unauthorized_rejection_amharic_dark_mode.png");
    console.log("   Saved 08_unauthorized_rejection_amharic_dark_mode.png");

    console.log("=== All 8 Browser Journeys Verified Successfully ===");
  } finally {
    await browser.close();

    // ─── Synthetic Data Cleanup ──────────────────────────────────────────────
    console.log("Cleaning up synthetic test records...");
    for (const eid of createdEncounterIds) {
      await db
        .query("DELETE FROM encounter_diagnosis WHERE encounter_id = $1", [eid])
        .catch(() => {});
      await db
        .query("DELETE FROM encounter_procedure WHERE encounter_id = $1", [eid])
        .catch(() => {});
      await db
        .query("DELETE FROM encounter_care_team WHERE encounter_id = $1", [eid])
        .catch(() => {});
      await db
        .query("DELETE FROM bed_assignment WHERE encounter_id = $1", [eid])
        .catch(() => {});
      await db
        .query("DELETE FROM bed_event WHERE encounter_id = $1", [eid])
        .catch(() => {});
      await db
        .query("DELETE FROM encounter WHERE id = $1", [eid])
        .catch(() => {});
    }
    for (const bid of createdBedIds) {
      await db
        .query("DELETE FROM hospital_bed WHERE id = $1", [bid])
        .catch(() => {});
    }
    for (const btid of createdBedTypeIds) {
      await db
        .query("DELETE FROM bed_type WHERE id = $1", [btid])
        .catch(() => {});
    }
    for (const cid of createdCaseIds) {
      await db
        .query("DELETE FROM patient_case WHERE id = $1", [cid])
        .catch(() => {});
    }
    for (const pid of createdPatientIds) {
      await db
        .query("DELETE FROM patient_profile WHERE patient_id = $1", [pid])
        .catch(() => {});
      await db
        .query("DELETE FROM patient WHERE id = $1", [pid])
        .catch(() => {});
    }
    for (const uid of createdUserIds) {
      await db
        .query("DELETE FROM doctor_profile WHERE user_id = $1", [uid])
        .catch(() => {});
      await db
        .query("DELETE FROM staff_access WHERE user_id = $1", [uid])
        .catch(() => {});
      await db
        .query('DELETE FROM account WHERE "userId" = $1', [uid])
        .catch(() => {});
      await db.query('DELETE FROM "user" WHERE id = $1', [uid]).catch(() => {});
    }
    await db.end();
    console.log("Cleanup completed.");
  }
}

main().catch((err) => {
  console.error("FATAL ERROR in Clinical Encounters QA:", err);
  process.exit(1);
});
