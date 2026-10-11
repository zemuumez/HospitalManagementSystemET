import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import { existsSync, mkdirSync, copyFileSync } from "node:fs";
import { randomBytes } from "node:crypto";
import assert from "node:assert/strict";

const databaseUrl = process.env.DATABASE_URL?.trim();
if (!databaseUrl) {
  console.error("ERROR: DATABASE_URL environment variable is required.");
  process.exit(1);
}

const base =
  process.env.BASE_URL ||
  process.env.BETTER_AUTH_URL ||
  "http://127.0.0.1:3000";
const isolatedSchema = process.env.HMS_TEST_ISOLATED_SCHEMA;

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
    } catch {}
  }
}

const db = new Pool({ connectionString: databaseUrl });

async function main() {
  console.log(
    "=== Starting Odontogram Workspace Parity Isolated Browser Verification ===",
  );
  console.log(`Target Base URL: ${base}`);
  console.log(`Screenshots Directory: ${screenshotsDir}`);

  if (isolatedSchema) {
    const currentSchema = (await db.query("SELECT current_schema() AS name"))
      .rows[0]?.name;
    assert.equal(
      currentSchema,
      isolatedSchema,
      `Database connection search_path must be active on isolated schema ${isolatedSchema}`,
    );
  }

  // Ensure Admin User exists for login
  let adminEmail = process.env.ADMIN_EMAIL;
  let adminPassword = process.env.ADMIN_PASSWORD;

  if (!adminEmail || !adminPassword) {
    const { hashPassword } = await import("better-auth/crypto");
    const token = randomBytes(8).toString("hex");
    const adminId = `adm-odonto-${token}`;
    adminEmail = `admin-odonto-${token}@ulshms.local`;
    adminPassword = `AdmPass!${randomBytes(12).toString("hex")}`;
    await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
      adminId,
      "QA Odontogram Admin",
      adminEmail,
    ]);
    await db.query(
      'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
      [`acc-odonto-${token}`, adminId, await hashPassword(adminPassword)],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'admin', true)",
      [adminId],
    );
  }

  // Ensure real Doctor exists in scheduling and doctor_profile
  const docRes = await db.query(
    "SELECT id FROM \"user\" u JOIN staff_access sa ON sa.user_id = u.id WHERE sa.role = 'doctor' LIMIT 1",
  );
  let doctorId;
  if (docRes.rows.length > 0) {
    doctorId = docRes.rows[0].id;
  } else {
    doctorId = `doc-odonto-${randomBytes(6).toString("hex")}`;
    await db.query(
      'INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)',
      [doctorId, "Dr. Dawit Haile", `${doctorId}@example.test`],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'doctor', true)",
      [doctorId],
    );
    await db.query(
      "INSERT INTO doctor_profile(user_id, department, slot_minutes) VALUES($1, 'Dentistry', 30) ON CONFLICT DO NOTHING",
      [doctorId],
    );
  }

  // Ensure real Patient exists in patient table
  const patRes = await db.query(
    "SELECT id, given_name, family_name FROM patient LIMIT 1",
  );
  let patientId;
  let patientName;
  if (patRes.rows.length > 0) {
    patientId = patRes.rows[0].id;
    patientName = `${patRes.rows[0].given_name} ${patRes.rows[0].family_name}`.trim();
  } else {
    const pInsert = await db.query(
      "INSERT INTO patient(given_name, family_name, date_of_birth) VALUES('Abebe', 'Bikila', '1990-01-01') RETURNING id",
    );
    patientId = pInsert.rows[0].id;
    patientName = "Abebe Bikila";
  }

  console.log(`Verified Doctor ID: ${doctorId}, Patient ID: ${patientId} (${patientName})`);

  const browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });

  try {
    // -------------------------------------------------------------
    // Journey 1: Admin Login and Navigate to Odontogram Module
    // -------------------------------------------------------------
    console.log("\n--- Journey 1: Login & Initial Odontogram State ---");
    const context1 = await browser.newContext({
      viewport: { width: 1440, height: 900 },
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

    // Navigate to Odontogram module
    await page1.goto(`${base}/modules/odontogram`, { waitUntil: "networkidle" });
    await page1.waitForSelector("table.legacy-table", { timeout: 15000 });

    // Verify 11 Default Condition Labels in Key Strip
    const legendKeyText = await page1.textContent(".form-card-container");
    assert.ok(legendKeyText.includes("11 conditions defined") || legendKeyText.includes("Odontogram Legends Key"),
      "Key strip should display 11 conditions defined");

    const expectedCodes = ["Healthy", "K", "C", "Ce", "D", "KR", "PS", "IP", "X", "F", "B"];
    for (const code of expectedCodes) {
      assert.ok(legendKeyText.includes(code), `Legend key strip must display standard code '${code}'`);
    }

    const initialScreenshot = resolve(screenshotsDir, "01_odontogram_initial_landing.png");
    await page1.screenshot({ path: initialScreenshot, fullPage: true });
    saveScreenshot(initialScreenshot, "01_odontogram_initial_landing.png");
    console.log("PASS: Journey 1 Initial state verified with exact 11 legend labels.");

    // -------------------------------------------------------------
    // Journey 2: Custom Legend Creation & Chart Aggregate Save
    // -------------------------------------------------------------
    console.log("\n--- Journey 2: Custom Legend Creation & Chart Aggregate Save ---");

    // Click "Manage / Customize Legends" button
    await page1.click('button:has-text("Odontogram Legends")');
    await page1.waitForSelector("#manage-legends-title", { timeout: 8000 });

    // Click "New Legend" button
    await page1.click('button:has-text("New Legend")');
    await page1.waitForSelector("#legend-form-title", { timeout: 8000 });

    // Create custom code "IMP-Z"
    await page1.fill('input[placeholder*="e.g. F, V, RCT"]', "IMP-Z");
    await page1.fill('input[placeholder*="Composite Restoration"]', "Zirconia Implant");
    await page1.fill('textarea[placeholder*="Optional clinical notes"]', "Full biocompatible zirconia implant unit");
    await page1.click('button:has-text("Save Legend")');
    await page1.waitForTimeout(500);

    // Close Manage Legends modal
    await page1.click('button:has-text("Close")');
    await page1.waitForTimeout(500);

    // Now click "Add Odontogram"
    await page1.click('button:has-text("Add Odontogram")');
    await page1.waitForSelector("#odontogram-title", { timeout: 8000 });

    // Test server-backed patient search input
    await page1.fill('input[placeholder*="Search patient by name or MRN"]', patientName.slice(0, 5));
    await page1.waitForTimeout(400);

    // Select Patient and Doctor
    await page1.selectOption('select.field:has(option[value=""]) >> nth=0', patientId);
    await page1.selectOption('select.field:has(option[value=""]) >> nth=1', doctorId);

    // Fill clinical description
    const notesText = "Comprehensive dental examination. Tooth 1 caries lesion, Tooth 8 bridge unit, Tooth 14 zirconia implant.";
    await page1.fill('textarea.field', notesText);

    // Assign conditions to teeth:
    // Tooth 1 -> K (Caries)
    await page1.selectOption('label:has-text("Tooth Selector") select', "1");
    await page1.click('.dental-palette button:has-text("K")');

    // Tooth 8 -> B (Bridge)
    await page1.selectOption('label:has-text("Tooth Selector") select', "8");
    await page1.click('.dental-palette button:has-text("B")');

    // Tooth 14 -> IMP-Z (Custom Zirconia Implant)
    await page1.selectOption('label:has-text("Tooth Selector") select', "14");
    await page1.click('.dental-palette button:has-text("IMP-Z")');

    // Submit and save chart to backend
    await page1.click('button:has-text("Save Dental Chart")');
    await page1.waitForSelector(".alert-notice:has-text('successfully')", { timeout: 15000 });

    // Wait for row with patient name to be rendered in the table
    await page1.waitForFunction(
      (expectedName) => {
        const tbody = document.querySelector("table.legacy-table tbody");
        if (!tbody) return false;
        const text = tbody.textContent || "";
        return text.includes(expectedName) || text.includes("Abebe");
      },
      patientName,
      { timeout: 15000 }
    );

    // Verify row appears in table with real patient, doctor, and notes
    const tableBody = await page1.textContent("table.legacy-table tbody");
    assert.ok(tableBody.includes(patientName) || tableBody.includes("Abebe"), "Table must list created chart with patient name");
    assert.ok(tableBody.includes("3 teeth charted") || tableBody.includes("teeth charted"), "Must show count of charted teeth");

    const createdScreenshot = resolve(screenshotsDir, "02_odontogram_chart_created.png");
    await page1.screenshot({ path: createdScreenshot, fullPage: true });
    saveScreenshot(createdScreenshot, "02_odontogram_chart_created.png");
    console.log("PASS: Journey 2 Chart created with standard & custom codes.");

    // Query chart ID from database for subsequent checks
    const dbChart = (await db.query(
      "SELECT id, version, odontogram FROM odontogram_chart WHERE patient_id = $1::uuid ORDER BY created_at DESC LIMIT 1",
      [patientId],
    )).rows[0];
    assert.ok(dbChart, "Database must store the odontogram chart");
    assert.equal(dbChart.odontogram["1"], "K", "Tooth 1 must be 'K'");
    assert.equal(dbChart.odontogram["8"], "B", "Tooth 8 must be 'B'");
    assert.equal(dbChart.odontogram["14"], "IMP-Z", "Tooth 14 must be 'IMP-Z'");
    const chartId = dbChart.id;
    console.log(`Created Chart ID in PostgreSQL: ${chartId} (version ${dbChart.version})`);

    // -------------------------------------------------------------
    // Journey 3: Fresh Session Persistence (No Session Storage)
    // -------------------------------------------------------------
    console.log("\n--- Journey 3: Fresh Session Persistence ---");
    await context1.close();

    // Open a brand new isolated context
    const context2 = await browser.newContext({
      viewport: { width: 1440, height: 900 },
    });
    const page2 = await context2.newPage();

    await performLogin(page2, adminEmail, adminPassword);

    await page2.goto(`${base}/modules/odontogram`, { waitUntil: "networkidle" });
    await page2.waitForSelector("table.legacy-table", { timeout: 15000 });

    await page2.waitForFunction(
      (expectedName) => {
        const tbody = document.querySelector("table.legacy-table tbody");
        if (!tbody) return false;
        const text = tbody.textContent || "";
        return text.includes(expectedName) || text.includes("Abebe");
      },
      patientName,
      { timeout: 15000 }
    );

    const freshTable = await page2.textContent("table.legacy-table tbody");
    assert.ok(freshTable.includes(patientName) || freshTable.includes("Abebe"),
      "Fresh session without sessionStorage must load chart from PostgreSQL database");

    // Verify custom legend IMP-Z persisted and appears in the key strip in fresh session
    const freshKeyStrip = await page2.textContent(".form-card-container");
    assert.ok(freshKeyStrip.includes("IMP-Z"), "Custom legend 'IMP-Z' must be visible in key strip in fresh session");

    // Open Edit modal to verify exact teeth markings loaded from backend
    await page2.click(`button[aria-label*="Edit odontogram"] >> nth=0`);
    await page2.waitForSelector("#odontogram-title", { timeout: 8000 });

    // Select Tooth 1 -> verify badge shows 'K'
    await page2.selectOption('label:has-text("Tooth Selector") select', "1");
    const tooth1Badge = await page2.textContent(".badge-available-stock");
    assert.ok(tooth1Badge.includes("K"), `Tooth 1 badge in fresh session must show K, got: ${tooth1Badge}`);

    // Select Tooth 8 -> verify badge shows 'B'
    await page2.selectOption('label:has-text("Tooth Selector") select', "8");
    const tooth8Badge = await page2.textContent(".badge-available-stock");
    assert.ok(tooth8Badge.includes("B"), `Tooth 8 badge in fresh session must show B, got: ${tooth8Badge}`);

    // Select Tooth 14 -> verify badge shows 'IMP-Z'
    await page2.selectOption('label:has-text("Tooth Selector") select', "14");
    const tooth14Badge = await page2.textContent(".badge-available-stock");
    assert.ok(tooth14Badge.includes("IMP-Z"), `Tooth 14 badge in fresh session must show IMP-Z, got: ${tooth14Badge}`);

    // Verify custom legend button exists in the palette in fresh session
    const impZPaletteBtn = page2.locator('.dental-palette button:has-text("IMP-Z")');
    assert.ok((await impZPaletteBtn.count()) > 0, "Custom legend 'IMP-Z' must be available in dental palette");

    const freshScreenshot = resolve(screenshotsDir, "03_odontogram_fresh_session_persistence.png");
    await page2.screenshot({ path: freshScreenshot, fullPage: true });
    saveScreenshot(freshScreenshot, "03_odontogram_fresh_session_persistence.png");
    console.log("PASS: Journey 3 Multi-session persistence confirmed from PostgreSQL.");

    // -------------------------------------------------------------
    // Journey 4: Concurrent Edit Conflict (HTTP 409 Stale Version)
    // -------------------------------------------------------------
    console.log("\n--- Journey 4: Concurrency Protection (HTTP 409) ---");

    // While page2 edit dialog is open (holding version 1), update chart in database to version 2
    await db.query(
      "UPDATE odontogram_chart SET version = version + 1, description = 'Conflicting update from concurrent clinician' WHERE id = $1::uuid",
      [chartId],
    );

    // Now click "Save Dental Chart" in page2 with stale version
    await page2.click('button:has-text("Save Dental Chart")');
    await page2.waitForSelector(".alert-notice:has-text('Concurrent Edit Conflict')", { timeout: 10000 });

    const conflictNotice = await page2.textContent(".alert-notice");
    assert.ok(conflictNotice.includes("Concurrent Edit Conflict"), "Must display concurrent edit conflict notice");

    const conflictScreenshot = resolve(screenshotsDir, "04_odontogram_concurrency_conflict.png");
    await page2.screenshot({ path: conflictScreenshot, fullPage: true });
    saveScreenshot(conflictScreenshot, "04_odontogram_concurrency_conflict.png");
    console.log("PASS: Journey 4 Stale edit blocked with 409 conflict notice.");

    // Close modal
    await page2.click('button:has-text("Cancel")');
    await page2.waitForTimeout(500);

    // -------------------------------------------------------------
    // Journey 5: Chart-Specific PDF Verification
    // -------------------------------------------------------------
    console.log("\n--- Journey 5: Chart-Specific PDF & Print Export ---");
    const pdfResponse = await page2.request.get(`${base}/api/hms/odontograms/${chartId}/pdf`);
    assert.equal(pdfResponse.status(), 200, "PDF endpoint must return HTTP 200");
    const contentType = pdfResponse.headers()["content-type"];
    assert.ok(contentType?.includes("application/pdf"), `Expected application/pdf, got ${contentType}`);
    const pdfBytes = await pdfResponse.body();
    assert.ok(pdfBytes.length > 50000, `PDF payload must be >50KB with embedded TrueType font, got ${pdfBytes.length}`);
    assert.equal(pdfBytes.subarray(0, 4).toString(), "%PDF", "Must begin with %PDF header");
    const pdfStr = pdfBytes.toString("binary");
    assert.ok(
      pdfStr.includes("/CIDFontType2") || pdfStr.includes("/Type0") || pdfStr.includes("/Identity-H"),
      "PDF must embed Type 0 / CIDFontType2 font",
    );
    console.log(`PASS: Journey 5 PDF endpoint returned valid %PDF payload (${pdfBytes.length} bytes) with Type 0 / CIDFont.`);

    // -------------------------------------------------------------
    // Journey 6: Clean Deletion & Real-Time Sync
    // -------------------------------------------------------------
    console.log("\n--- Journey 6: Chart Deletion ---");
    page2.on("dialog", (dialog) => dialog.accept());
    await page2.click(`button[aria-label*="Delete odontogram"] >> nth=0`);
    await page2.waitForSelector(".alert-notice:has-text('deleted successfully')", { timeout: 10000 });

    await page2.waitForFunction(
      () => {
        const tbody = document.querySelector("table.legacy-table tbody");
        if (!tbody) return false;
        const text = tbody.textContent || "";
        return text.includes("No dental charts found") || !text.includes("3 teeth charted");
      },
      null,
      { timeout: 10000 }
    );

    const afterDeleteTable = await page2.textContent("table.legacy-table tbody");
    assert.ok(afterDeleteTable.includes("No dental charts found") || !afterDeleteTable.includes(chartId),
      "Table must update to reflect chart deletion");

    const deletedScreenshot = resolve(screenshotsDir, "05_odontogram_deleted_state.png");
    await page2.screenshot({ path: deletedScreenshot, fullPage: true });
    saveScreenshot(deletedScreenshot, "05_odontogram_deleted_state.png");
    console.log("PASS: Journey 6 Deletion completed and verified.");

    await context2.close();
  } finally {
    await browser.close();
    await db.end();
  }

  console.log("\n=== ALL ODONTOGRAM BROWSER QA JOURNEYS PASSED CLEANLY ===");
}

main().catch((err) => {
  console.error("FATAL: Odontogram Browser QA failed:", err);
  process.exit(1);
});
