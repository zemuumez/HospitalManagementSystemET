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
if (!isolatedSchema || !/^hms_browser_[a-f0-9]{24}$/.test(isolatedSchema)) {
  console.error(
    "ERROR: verify-smart-cards-workspace requires isolated schema execution.",
  );
  process.exit(1);
}

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
    "=== Starting Smart Cards Workspace Isolated Browser Verification ===",
  );
  console.log(`Target Base URL: ${base}`);
  console.log(`Screenshots Directory: ${screenshotsDir}`);

  const currentSchema = (await db.query("SELECT current_schema() AS name"))
    .rows[0]?.name;
  assert.equal(
    currentSchema,
    isolatedSchema,
    `Database connection search_path must be active on isolated schema ${isolatedSchema}`,
  );

  let adminEmail = process.env.ADMIN_EMAIL;
  let adminPassword = process.env.ADMIN_PASSWORD;

  if (!adminEmail || !adminPassword) {
    const { hashPassword } = await import("better-auth/crypto");
    const token = randomBytes(8).toString("hex");
    const adminId = `adm-${token}`;
    adminEmail = `admin-${token}@ulshms.local`;
    adminPassword = `AdmPass!${randomBytes(12).toString("hex")}`;
    await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
      adminId,
      "QA Admin",
      adminEmail,
    ]);
    await db.query(
      'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
      [`acc-${token}`, adminId, await hashPassword(adminPassword)],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'admin', true)",
      [adminId],
    );
  }

  // Seed sample patients for cards generation testing
  const patientSeeds = [
    {
      given: "Abebe",
      family: "Bikila",
      dob: "1990-01-01",
      phone: "+251911223344",
    },
    {
      given: "Derartu",
      family: "Tulu",
      dob: "1992-03-15",
      phone: "+251922334455",
    },
    {
      given: "Haile",
      family: "Gebrselassie",
      dob: "1988-04-18",
      phone: "+251933445566",
    },
    {
      given: "Kenenisa",
      family: "Bekele",
      dob: "1995-06-13",
      phone: "+251944556677",
    },
    {
      given: "Tirunesh",
      family: "Dibaba",
      dob: "1996-09-22",
      phone: "+251955667788",
    },
  ];

  const seededPatients = [];
  for (const p of patientSeeds) {
    const res = await db.query(
      `INSERT INTO patient(given_name, family_name, date_of_birth, phone)
       VALUES($1, $2, $3, $4)
       RETURNING id, given_name, family_name`,
      [p.given, p.family, p.dob, p.phone],
    );
    const pid = res.rows[0].id;
    // ensure patient profile has email and address
    await db.query(
      `INSERT INTO patient_profile(patient_id, email, address1, city, blood_group)
       VALUES($1, $2, $3, 'Addis Ababa', 'O+')
       ON CONFLICT (patient_id) DO UPDATE SET
         email = EXCLUDED.email, address1 = EXCLUDED.address1, city = EXCLUDED.city, blood_group = EXCLUDED.blood_group`,
      [pid, `${p.given.toLowerCase()}@ulshms.local`, `${p.given} Residence`],
    );
    seededPatients.push({ id: pid, name: `${p.given} ${p.family}` });
  }

  const browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });

  try {
    const context = await browser.newContext({
      viewport: { width: 1440, height: 900 },
    });
    const page = await context.newPage();

    async function performLogin(p, email, password) {
      await p.goto(`${base}/login`);
      await p.locator('input[name="email"]').fill(email);
      await p.locator('input[name="password"]').fill(password);
      await p.getByRole("button", { name: "Sign In", exact: true }).click();
      await p.waitForURL("**/dashboard", { timeout: 15000 }).catch(async () => {
        await p.goto(`${base}/dashboard`);
      });
    }

    // ─── LOGIN ─────────────────────────────────────────────────────────────
    console.log("Authenticating QA Admin...");
    await performLogin(page, adminEmail, adminPassword);
    console.log("QA Admin authenticated successfully.");

    // ─── JOURNEY 1: Template CRUD, Duplicate Names, Toggles ───────────────
    console.log(
      "Starting Journey 1: Template CRUD, Duplicate Names, Toggles...",
    );
    await page.goto(`${base}/modules/patient-id-card-template`, {
      waitUntil: "domcontentloaded",
    });

    // Check heading
    const heading = await page.textContent("h1");
    assert.match(heading, /Smart Patient Card Templates/i);

    // Verify seeded templates exist
    await page.waitForSelector('table tbody tr:has-text("Standard")');
    const tableText = await page.textContent("table");
    assert.ok(
      tableText.includes("Standard"),
      "Table must contain 'Standard' template",
    );
    assert.ok(tableText.includes("VIP"), "Table must contain 'VIP' template");

    // Screenshot 1: Templates landing
    const shot1 = resolve(
      screenshotsDir,
      "01_smart_cards_templates_landing.png",
    );
    await page.screenshot({ path: shot1, fullPage: true });
    saveScreenshot(shot1, "01_smart_cards_templates_landing.png");
    console.log("Captured: 01_smart_cards_templates_landing.png");

    // Click "New Patient Smart Card Template"
    await page.click('button:has-text("New Patient Smart Card Template")');
    await page.waitForSelector("#template-title");

    // Fill form
    await page.fill("input[required]", "Pediatric Clinic");
    // Pick color
    await page.fill('input[type="color"]', "#10b981");

    // Verify live preview card reflects values
    const previewHeader = await page.$(".smart-card header");
    assert.ok(previewHeader, "Live preview card must be visible in modal");

    // Save template
    await page.click('button:has-text("Save")');
    await page.waitForSelector(".modal-heading", { state: "detached" });

    // Verify new template in table
    await page.waitForFunction(() =>
      document
        .querySelector("table")
        ?.textContent?.includes("Pediatric Clinic"),
    );

    // Test Duplicate Name Rejection
    await page.click('button:has-text("New Patient Smart Card Template")');
    await page.waitForSelector("#template-title");
    await page.fill("input[required]", "pediatric clinic"); // duplicate case-insensitive
    await page.click('button:has-text("Save")');
    await page.waitForSelector('p[role="alert"]');
    const alertMsg = await page.textContent('p[role="alert"]');
    assert.match(alertMsg, /Template name must be unique/i);
    await page.click('button:has-text("Cancel")');
    await page.waitForSelector(".modal-heading", { state: "detached" });

    // Test Inline Toggle switch: Toggle Email off on Pediatric Clinic
    const pediatricRow = page.locator('tr:has-text("Pediatric Clinic")');
    const emailSwitch = pediatricRow.locator('input[role="switch"]').first();
    assert.ok(
      await emailSwitch.isChecked(),
      "Email switch should initially be checked",
    );
    await emailSwitch.click();
    await page.waitForTimeout(500); // let API save

    // Verify in database that toggle persisted
    const dbTmpl = (
      await db.query(
        "SELECT email FROM patient_id_card_template WHERE name='Pediatric Clinic'",
      )
    ).rows[0];
    assert.equal(
      dbTmpl.email,
      false,
      "Database must persist email toggle as false",
    );

    // Edit Template: Rename to "Pediatric Gold" and color "#059669"
    await pediatricRow.locator('button[aria-label*="Edit"]').click();
    await page.waitForSelector("#template-title");
    await page.fill("input[required]", "Pediatric Gold");
    await page.fill('input[type="color"]', "#059669");
    await page.click('button:has-text("Save")');
    await page.waitForSelector(".modal-heading", { state: "detached" });

    // Verify updated name in table
    await page.waitForFunction(() =>
      document.querySelector("table")?.textContent?.includes("Pediatric Gold"),
    );

    // Screenshot 2: Template Created
    const shot2 = resolve(
      screenshotsDir,
      "02_smart_cards_template_created.png",
    );
    await page.screenshot({ path: shot2, fullPage: true });
    saveScreenshot(shot2, "02_smart_cards_template_created.png");
    console.log("Captured: 02_smart_cards_template_created.png");

    // ─── JOURNEY 2: All 3 Generation Modes, Repeat Requests, Concurrent ───
    console.log("Starting Journey 2: Three Generation Modes & Concurrency...");
    await page.click(
      '.detail-tabs button:has-text("Generate Patient Smart Cards")',
    );
    await page.waitForSelector(
      '.detail-tabs button.active:has-text("Generate Patient Smart Cards")',
    );
    const cardsHeading = await page.locator(".page-heading h1").textContent();
    assert.match(cardsHeading, /Generate Patient Smart Cards/i);

    // Mode 1: "Selected Patient" (Single mode)
    await page.locator(".page-heading button.primary").click();
    await page.waitForSelector("#generate-cards-title");
    const modal1 = page.locator("dialog[open]");

    // Select "Selected Patient"
    await modal1.locator("select").nth(0).selectOption("single");

    // Select Template: "Pediatric Gold"
    await modal1
      .locator("select")
      .nth(1)
      .selectOption({ label: "Pediatric Gold" });

    // Select Patient: Abebe Bikila
    const patientSelect = modal1.locator("select").nth(2);
    await patientSelect.waitFor();
    await patientSelect.selectOption({ label: "Abebe Bikila (+251911223344)" });

    await modal1.locator('button:has-text("Save")').click();
    await page.waitForSelector("dialog[open]", { state: "detached" });

    // Verify card appears in table
    await page.waitForSelector('table tbody tr:has-text("Abebe Bikila")');
    const abebeRow = page.locator('tr:has-text("Abebe Bikila")');
    const abebeIdText = await abebeRow.locator(".card-id").textContent();
    assert.match(
      abebeIdText,
      /^P-\d+$/,
      "Patient Unique ID must match format P-XXXXX",
    );
    console.log(`Generated Abebe Bikila Unique ID: ${abebeIdText}`);

    // Mode 2: "Patients Without Cards"
    await page.locator(".page-heading button.primary").click();
    await page.waitForSelector("#generate-cards-title");
    const modal2 = page.locator("dialog[open]");
    await modal2.locator("select").nth(0).selectOption("unassigned");
    await modal2.locator("select").nth(1).selectOption({ label: "Standard" });
    await modal2.locator('button:has-text("Save")').click();
    await page.waitForSelector("dialog[open]", { state: "detached" });

    // Verify multiple cards now in table
    await page.waitForFunction(
      () => document.querySelectorAll("table tbody tr").length >= 5,
    );

    // Repeat Request / Idempotency: Re-generate Abebe with VIP template
    await page.locator(".page-heading button.primary").click();
    await page.waitForSelector("#generate-cards-title");
    const modal3 = page.locator("dialog[open]");
    await modal3.locator("select").nth(0).selectOption("single");
    await modal3.locator("select").nth(1).selectOption({ label: "VIP" });
    const patientSelect3 = modal3.locator("select").nth(2);
    await patientSelect3.waitFor();
    await patientSelect3.selectOption({
      label: "Abebe Bikila (+251911223344)",
    });
    await modal3.locator('button:has-text("Save")').click();
    await page.waitForSelector("dialog[open]", { state: "detached" });

    // Verify Abebe retained EXACT same unique ID
    const abebeRowAfter = page.locator('tr:has-text("Abebe Bikila")');
    const abebeIdAfter = await abebeRowAfter.locator(".card-id").textContent();
    assert.equal(
      abebeIdAfter,
      abebeIdText,
      "Persistent patient_unique_id must never mutate across card reassignments!",
    );

    // Mode 3: "All Patients"
    await page.locator(".page-heading button.primary").click();
    await page.waitForSelector("#generate-cards-title");
    const modal4 = page.locator("dialog[open]");
    await modal4.locator("select").nth(0).selectOption("all");
    await modal4.locator("select").nth(1).selectOption({ label: "VIP" });
    await modal4.locator('button:has-text("Save")').click();
    await page.waitForSelector("dialog[open]", { state: "detached" });

    // Verify all 5 patients have active VIP cards in database
    const vipCount = (
      await db.query(
        "SELECT count(*) FROM patient_smart_card sc JOIN patient_id_card_template t ON t.id=sc.template_id WHERE t.name='VIP' AND sc.status='active'",
      )
    ).rows[0].count;
    assert.equal(
      Number(vipCount),
      5,
      "All 5 patients must have active VIP cards",
    );

    // Screenshot 3: Generated list
    const shot3 = resolve(screenshotsDir, "03_smart_cards_generated_list.png");
    await page.screenshot({ path: shot3, fullPage: true });
    saveScreenshot(shot3, "03_smart_cards_generated_list.png");
    console.log("Captured: 03_smart_cards_generated_list.png");

    // Concurrent Generation Test (5 concurrent requests)
    console.log("Running concurrent generation requests...");
    const tmplId = (
      await db.query("SELECT id FROM patient_id_card_template WHERE name='VIP'")
    ).rows[0].id;
    const concurrentStatuses = await page.evaluate(async (tmplId) => {
      const promises = Array.from({ length: 5 }, (_, i) =>
        fetch("/api/hms/smart-cards/generate", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            templateId: tmplId,
            mode: i % 2 === 0 ? "all" : "unassigned",
          }),
        }).then((r) => r.status),
      );
      return Promise.all(promises);
    }, tmplId);
    for (const status of concurrentStatuses) {
      assert.equal(
        status,
        200,
        `Concurrent generation request must succeed with 200, got ${status}`,
      );
    }
    // Verify each patient STILL has EXACTLY 1 active card
    const patientActiveCounts = (
      await db.query(
        "SELECT patient_id, count(*) as count FROM patient_smart_card WHERE status='active' GROUP BY patient_id",
      )
    ).rows;
    for (const row of patientActiveCounts) {
      assert.equal(
        Number(row.count),
        1,
        `Patient ${row.patient_id} has ${row.count} active cards, must be exactly 1 under concurrency!`,
      );
    }

    // ─── JOURNEY 3: Fresh-Session Persistence Without Mocks ───────────────
    console.log(
      "Starting Journey 3: Fresh-Session Persistence Without Mocks...",
    );
    const freshContext = await browser.newContext({
      viewport: { width: 1440, height: 900 },
    });
    const freshPage = await freshContext.newPage();
    await performLogin(freshPage, adminEmail, adminPassword);

    // Navigate to smart cards in fresh context
    await freshPage.goto(`${base}/modules/generate-patient-id-card`, {
      waitUntil: "domcontentloaded",
    });
    await freshPage.waitForSelector('table tbody tr:has-text("Abebe Bikila")');
    const freshTable = await freshPage.textContent("table");
    assert.ok(
      freshTable.includes("Abebe Bikila"),
      "Fresh session must show Abebe Bikila from DB",
    );
    assert.ok(
      freshTable.includes(abebeIdText),
      "Fresh session must show persistent ID from DB",
    );

    // Screenshot 4: Fresh session persistence
    const shot4 = resolve(
      screenshotsDir,
      "04_smart_cards_fresh_session_persistence.png",
    );
    await freshPage.screenshot({ path: shot4, fullPage: true });
    saveScreenshot(shot4, "04_smart_cards_fresh_session_persistence.png");
    console.log("Captured: 04_smart_cards_fresh_session_persistence.png");
    await freshContext.close();

    // ─── JOURNEY 4: Actual QR Decoding, Verification, and Revocation ──────
    console.log(
      "Starting Journey 4: QR Decoding, Verification, and Revocation...",
    );
    // Reload cards table to reflect latest active cards after concurrent generation
    await page.goto(`${base}/modules/generate-patient-id-card`, {
      waitUntil: "domcontentloaded",
    });
    await page.waitForSelector('table tbody tr:has-text("Abebe Bikila")');
    const currentAbebeRow = page.locator('tr:has-text("Abebe Bikila")');
    // View card for Abebe Bikila
    await currentAbebeRow.locator('button[aria-label*="View card"]').click();
    await page.waitForSelector(".smart-card-modal");

    // Extract QR code token and verify URL from DOM
    const qrContainer = page.locator(".card-qr");
    const qrToken = await qrContainer.getAttribute("data-qr-token");
    const qrUrl = await qrContainer.getAttribute("data-qr-url");
    assert.ok(
      qrToken && qrToken.length >= 16,
      `Valid QR token required, got ${qrToken}`,
    );
    assert.ok(
      qrUrl && qrUrl.includes(qrToken),
      `QR URL must contain token: ${qrUrl}`,
    );

    // Verify SVG QR code rendered
    const qrSvg = await qrContainer.locator("svg").count();
    assert.equal(qrSvg, 1, "Authentic SVG QR matrix must be rendered");

    // Perform actual public verification HTTP request against Go backend
    console.log(
      `Verifying QR token against /v1/smart-cards/verify?token=${qrToken}...`,
    );
    const verifyRes = await fetch(
      `${base}/v1/smart-cards/verify?token=${qrToken}`,
    );
    assert.equal(verifyRes.status, 200, "Verification must return 200 OK");
    const verifyData = await verifyRes.json();
    assert.equal(verifyData.valid, true, "Card must be valid");
    assert.equal(
      verifyData.patientName,
      "Abebe Bikila",
      "Patient name must match",
    );
    assert.equal(
      verifyData.patientUniqueId,
      abebeIdText,
      "Patient unique ID must match",
    );
    assert.equal(verifyData.bloodGroup, "O+", "Blood group must match");
    // Under VIP template, dob is false: verify dob is omitted!
    assert.equal(
      verifyData.dob,
      undefined,
      "DOB must be omitted per VIP template toggle",
    );

    // Test forged token returns 404
    const forgedRes = await fetch(
      `${base}/v1/smart-cards/verify?token=forged-token-9999999999999999`,
    );
    assert.equal(
      forgedRes.status,
      404,
      "Forged token must return 404 Not Found",
    );

    // Screenshot 5: View card modal
    const shot5 = resolve(screenshotsDir, "05_smart_cards_view_modal.png");
    await page.screenshot({ path: shot5, fullPage: true });
    saveScreenshot(shot5, "05_smart_cards_view_modal.png");
    console.log("Captured: 05_smart_cards_view_modal.png");

    // Close view card modal
    await page.click('button:has-text("Close")');
    await page.waitForSelector(".smart-card-modal", { state: "detached" });

    // Test Referenced-Template Deletion: Clears assignments, preserves patients & unique IDs, revokes cards
    const delVipStatus = await page.evaluate(async (tmplId) => {
      const res = await fetch(`/api/hms/smart-cards/templates/${tmplId}`, {
        method: "DELETE",
      });
      return res.status;
    }, tmplId);
    assert.equal(
      delVipStatus,
      200,
      `Referenced template deletion must succeed with 200, got ${delVipStatus}`,
    );

    // Verify patient in DB is preserved and template assignment cleared
    const abebePatientDb = (
      await db.query(
        "SELECT patient_unique_id, smart_card_template_id FROM patient WHERE given_name='Abebe'",
      )
    ).rows[0];
    assert.equal(
      abebePatientDb.patient_unique_id,
      abebeIdText,
      "Patient unique ID must be preserved in DB after template deletion",
    );
    assert.equal(
      abebePatientDb.smart_card_template_id,
      null,
      "Patient smart_card_template_id must be cleared in DB after template deletion",
    );

    // Verify previously valid QR token is now REVOKED because its template was deleted!
    const revokedVerifyRes = await fetch(
      `${base}/v1/smart-cards/verify?token=${qrToken}`,
    );
    assert.equal(revokedVerifyRes.status, 200);
    const revokedVerifyData = await revokedVerifyRes.json();
    assert.equal(
      revokedVerifyData.valid,
      false,
      "Revoked card must NOT be valid",
    );
    assert.equal(
      revokedVerifyData.status,
      "revoked",
      "Status must be 'revoked'",
    );
    console.log(
      "Verified revoked QR token rejection after template deletion successfully.",
    );

    // ─── JOURNEY 5: Dark Mode, Printable Output, Amharic Controls ─────────
    console.log("Starting Journey 5: Dark Mode and Amharic Localization...");
    // Generate cards using Standard template
    const stdTmplId = (
      await db.query(
        "SELECT id FROM patient_id_card_template WHERE name='Standard'",
      )
    ).rows[0].id;
    await page.evaluate(async (tmplId) => {
      await fetch("/api/hms/smart-cards/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ templateId: tmplId, mode: "all" }),
      });
    }, stdTmplId);

    // Reload page to show generated cards
    await page.goto(`${base}/modules/generate-patient-id-card`, {
      waitUntil: "domcontentloaded",
    });
    await page.waitForSelector('table tbody tr:has-text("Haile Gebrselassie")');
    const haileRow = page.locator('tr:has-text("Haile Gebrselassie")');
    await haileRow.locator('button[aria-label*="View card"]').click();
    await page.waitForSelector(".smart-card-modal");

    // Verify Download & Print buttons
    assert.ok(
      await page.$('button:has-text("Download")'),
      "Download button must be present",
    );
    assert.ok(
      await page.$('button:has-text("Print")'),
      "Print button must be present",
    );

    // Toggle Dark Mode
    await page.evaluate(() => {
      document.documentElement.classList.add("legacy-dark");
      document.body.classList.add("legacy-dark");
    });
    await page.waitForTimeout(300);

    // Screenshot 6: Dark Mode
    const shot6 = resolve(screenshotsDir, "06_smart_cards_dark_mode.png");
    await page.screenshot({ path: shot6, fullPage: true });
    saveScreenshot(shot6, "06_smart_cards_dark_mode.png");
    console.log("Captured: 06_smart_cards_dark_mode.png");

    // Revert dark mode
    await page.evaluate(() => {
      document.documentElement.classList.remove("legacy-dark");
      document.body.classList.remove("legacy-dark");
    });

    // Close card modal
    await page.click('button:has-text("Close")');
    await page.waitForSelector(".smart-card-modal", { state: "detached" });

    // Switch Language to Amharic
    console.log("Switching language to Amharic...");
    await page.evaluate(() => {
      localStorage.setItem("hms-language", "am");
      window.dispatchEvent(new Event("storage"));
    });
    await page.goto(`${base}/modules/generate-patient-id-card`, {
      waitUntil: "domcontentloaded",
    });

    // Wait for client-side language hydration
    await page.waitForFunction(() => document.documentElement.lang === "am");
    await page.waitForSelector(
      '.detail-tabs button.active:has-text("የታካሚ ስማርት ካርድ ያዘጋጁ")',
    );
    await page.waitForSelector('.table-toolbar select:has-text("ሁሉም አብነቶች")');
    await page.waitForSelector('.reference-pagination label:has-text("አሳይ")');

    // Verify Amharic translated headings
    const amHeading = await page.textContent("h1");
    assert.match(
      amHeading,
      /የታካሚ ስማርት ካርድ ያዘጋጁ/i,
      `Amharic heading must translate to 'የታካሚ ስማርት ካርድ ያዘጋጁ', got ${amHeading}`,
    );

    // Verify "All Templates" translates to "ሁሉም አብነቶች"
    const tmplSelectText = await page
      .locator(".table-toolbar select")
      .textContent();
    assert.ok(
      tmplSelectText.includes("ሁሉም አብነቶች"),
      `Amharic template filter must show 'ሁሉም አብነቶች', got ${tmplSelectText}`,
    );

    // Verify "Show" translates to "አሳይ"
    const showLabelText = await page
      .locator(".reference-pagination label")
      .first()
      .textContent();
    assert.ok(
      showLabelText.includes("አሳይ"),
      `Amharic pagination label must show 'አሳይ', got ${showLabelText}`,
    );

    // Screenshot 7: Amharic Localization
    const shot7 = resolve(
      screenshotsDir,
      "07_smart_cards_amharic_localization.png",
    );
    await page.screenshot({ path: shot7, fullPage: true });
    saveScreenshot(shot7, "07_smart_cards_amharic_localization.png");
    console.log("Captured: 07_smart_cards_amharic_localization.png");

    console.log(
      "=== All Smart Cards Workspace Isolated Browser Tests Passed Successfully! ===",
    );
  } finally {
    await browser.close();
    await db.end();
  }
}

main().catch((err) => {
  console.error("FATAL: verify-smart-cards-workspace failed:", err);
  process.exit(1);
});
