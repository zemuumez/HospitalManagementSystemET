import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import { existsSync, mkdirSync, writeFileSync } from "node:fs";
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
    "ERROR: verify-settings-workspace requires isolated schema execution.",
  );
  process.exit(1);
}

const screenshotsDir = process.env.SCREENSHOT_DIR
  ? resolve(process.env.SCREENSHOT_DIR)
  : resolve(process.cwd(), ".local/screenshots");
mkdirSync(screenshotsDir, { recursive: true });

const db = new Pool({ connectionString: databaseUrl });

async function main() {
  console.log(
    "=== Starting Settings Workspace Isolated Browser Verification ===",
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

  // Provision doctor for non-admin rejection test
  const { hashPassword } = await import("better-auth/crypto");
  const docToken = randomBytes(8).toString("hex");
  const docId = `doc-${docToken}`;
  const docEmail = `doc-${docToken}@ulshms.local`;
  const docPassword = `DocPass!${randomBytes(12).toString("hex")}`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    docId,
    "Dr. Test NonAdmin",
    docEmail,
  ]);
  await db.query(
    'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
    [`acc-doc-${docToken}`, docId, await hashPassword(docPassword)],
  );
  await db.query(
    "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'doctor', true)",
    [docId],
  );

  // Create sample 1x1 PNG file for attachment upload test
  const samplePngPath = resolve(screenshotsDir, "test_logo.png");
  const pngBuffer = Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
    "base64",
  );
  writeFileSync(samplePngPath, pngBuffer);

  const browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
  });

  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();
  page.on("console", (msg) =>
    console.log(`[BROWSER ${msg.type()}]`, msg.text()),
  );
  page.on("pageerror", (err) => console.log("[BROWSER ERROR]", err.message));

  async function performLogin(p, email, password) {
    await p.goto(`${base}/login`);
    await p.locator('input[name="email"]').fill(email);
    await p.locator('input[name="password"]').fill(password);
    await p.getByRole("button", { name: "Sign In", exact: true }).click();
    await p.waitForURL("**/dashboard", { timeout: 15000 }).catch(async () => {
      await p.goto(`${base}/dashboard`);
    });
  }

  try {
    // -----------------------------------------------------------------------
    // Journey 1: Admin Login & Workspace Navigation
    // -----------------------------------------------------------------------
    console.log("\n[Journey 1] Admin Login & Settings Workspace Navigation...");
    await performLogin(page, adminEmail, adminPassword);

    await page.goto(`${base}/modules/settings`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']", {
      timeout: 10000,
    });

    // Verify subtabs navigation items
    const subtabs = await page.$$eval(".module-subtab-link", (els) =>
      els.map((e) => e.textContent.trim()),
    );
    console.log("Found Settings Subtabs:", subtabs);
    assert(
      subtabs.includes("General Settings"),
      "Expected General Settings subtab",
    );
    assert(
      subtabs.includes("Hospital Schedule"),
      "Expected Hospital Schedule subtab",
    );
    assert(
      subtabs.includes("Modules Setting"),
      "Expected Modules Setting subtab",
    );
    assert(subtabs.includes("Currencies"), "Expected Currencies subtab");
    assert(
      subtabs.includes("Payment Gateways"),
      "Expected Payment Gateways subtab",
    );
    assert(
      subtabs.includes("Patient Queue Theme"),
      "Expected Patient Queue Theme subtab",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "01_settings_admin_landing.png"),
    });
    console.log(
      "✔ Journey 1 Passed: Admin landed on Settings Workspace with all tabs.",
    );

    // -----------------------------------------------------------------------
    // Journey 2: Required-Field Validation & Error Notice
    // -----------------------------------------------------------------------
    console.log("\n[Journey 2] Required-Field Validation & Error Notice...");
    const appNameInput = page.locator('label:has-text("App Name") input');
    await appNameInput.fill("");
    await page.click('button:has-text("Save Settings")');

    // Browser or component validation alert
    const alert = page.locator('div.alert-notice[role="alert"]');
    await alert.waitFor({ state: "visible", timeout: 5000 });
    const alertText = await alert.textContent();
    console.log("Validation Alert:", alertText);
    assert(
      alertText.includes("Application Name is required"),
      "Expected validation message",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "02_settings_validation_error.png"),
    });
    console.log(
      "✔ Journey 2 Passed: Required-field client validation stops submit with error alert.",
    );

    // -----------------------------------------------------------------------
    // Journey 3: General Settings Persistence Across Reload and Fresh Session
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 3] General Settings Persistence Across Reload and Fresh Session...",
    );
    await appNameInput.fill("Addis Central Hospital");
    await page
      .locator('label:has-text("Company Name") input')
      .fill("ACH Health System");
    await page
      .locator('label:has-text("Hospital Phone") input')
      .fill("+251911778899");
    await page.click('button:has-text("Save Settings")');

    const successNotice = page.locator('div[role="status"]');
    await successNotice.waitFor({ state: "visible", timeout: 5000 });
    const noticeText = await successNotice.textContent();
    assert(
      noticeText.includes("General settings saved successfully"),
      "Expected success notice",
    );
    assert(
      !noticeText.includes("Preview saved"),
      "Must NOT contain preview saved text",
    );

    // Verify PostgreSQL persistence
    const dbAppName = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'app_name'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbAppName,
      "Addis Central Hospital",
      "DB hospital_general_setting app_name must match",
    );

    // Reload page in current session
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const reloadedAppName = await page
      .locator('label:has-text("App Name") input')
      .inputValue();
    assert.equal(
      reloadedAppName,
      "Addis Central Hospital",
      "App Name must persist after reload",
    );

    // Fresh session in new incognito context
    const freshContext = await browser.newContext();
    const freshPage = await freshContext.newPage();
    await performLogin(freshPage, adminEmail, adminPassword);
    await freshPage.goto(`${base}/modules/settings`);
    await freshPage.waitForSelector(".legacy-workspace[data-ready='true']");
    const freshAppName = await freshPage
      .locator('label:has-text("App Name") input')
      .inputValue();
    assert.equal(
      freshAppName,
      "Addis Central Hospital",
      "App Name must persist in fresh session",
    );
    await freshContext.close();

    await page.screenshot({
      path: resolve(screenshotsDir, "03_settings_persisted.png"),
    });
    console.log(
      "✔ Journey 3 Passed: General settings persisted across reload and fresh session.",
    );

    // -----------------------------------------------------------------------
    // Journey 4: Optional Field Clearing
    // -----------------------------------------------------------------------
    console.log("\n[Journey 4] Optional Field Clearing...");
    const fbInput = page.locator('label:has-text("Facebook URL") input');
    await fbInput.fill("https://facebook.com/addiscentral");
    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector('div[role="status"]');

    let dbFb = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'facebook_url'",
      )
    ).rows[0]?.value;
    assert.equal(dbFb, "https://facebook.com/addiscentral");

    // Clear optional field
    await fbInput.fill("");
    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector('div[role="status"]');

    dbFb = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'facebook_url'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbFb,
      "",
      "Database must clear optional field when submitted empty",
    );

    await page.reload();
    const reloadedFb = await page
      .locator('label:has-text("Facebook URL") input')
      .inputValue();
    assert.equal(
      reloadedFb,
      "",
      "UI must reflect cleared optional field on reload",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "04_settings_cleared_optional.png"),
    });
    console.log(
      "✔ Journey 4 Passed: Optional field clearing persists empty string.",
    );

    // -----------------------------------------------------------------------
    // Journey 5: Provider Secret Redaction & Preservation (OpenAI Key)
    // -----------------------------------------------------------------------
    console.log("\n[Journey 5] Provider Secret Redaction & Preservation...");
    const secretInput = page.locator('label:has-text("Open AI Key") input');
    const initialSecretVal = await secretInput.inputValue();
    console.log("Initial secret input value:", initialSecretVal);

    // Enter new secret
    await secretInput.fill("sk-proj-testsecret123456789");
    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector('div[role="status"]');

    // Verify plaintext is stored in PostgreSQL
    const dbSecret = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'open_ai_key'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbSecret,
      "sk-proj-testsecret123456789",
      "PostgreSQL must store actual secret",
    );

    // Reload page: verify returned value is REDACTED placeholder [CONFIGURED]
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const redactedVal = await page
      .locator('label:has-text("Open AI Key") input')
      .inputValue();
    assert.equal(
      redactedVal,
      "[CONFIGURED]",
      "Browser must receive redacted [CONFIGURED], never plaintext",
    );

    // Save unchanged with [CONFIGURED]
    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector('div[role="status"]');

    // Verify database preserved original secret
    const preservedSecret = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'open_ai_key'",
      )
    ).rows[0]?.value;
    assert.equal(
      preservedSecret,
      "sk-proj-testsecret123456789",
      "Submitting placeholder must preserve existing secret in DB",
    );

    // Clear secret
    await page.click('button:has-text("Clear Secret")');
    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector('div[role="status"]');

    const clearedSecret = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'open_ai_key'",
      )
    ).rows[0]?.value;
    assert.equal(
      clearedSecret,
      "",
      "Explicitly clearing secret must persist empty string in DB",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "05_settings_secret_protection.png"),
    });
    console.log(
      "✔ Journey 5 Passed: Provider secret redacted, preserved on unchanged, cleared on empty.",
    );

    // -----------------------------------------------------------------------
    // Journey 6: Authorized Attachment Lifecycle (Logo & Favicon)
    // -----------------------------------------------------------------------
    console.log("\n[Journey 6] Authorized Attachment Lifecycle...");
    // Upload PNG logo using file input
    const logoFileInput = page
      .locator('input[type="file"][accept*="image/png"]')
      .first();
    await logoFileInput.setInputFiles(samplePngPath);
    await page.waitForSelector(
      'div[role="status"]:has-text("uploaded successfully")',
    );

    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector(
      'div[role="status"]:has-text("General settings saved successfully")',
    );

    const dbLogoUrl = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'logo_url'",
      )
    ).rows[0]?.value;
    console.log("Database stored logo_url:", dbLogoUrl);
    assert(
      dbLogoUrl.startsWith("/api/hms/attachments/"),
      "Logo URL must point to authorized attachment endpoint",
    );

    // Verify unauthenticated download of public logo attachment
    const downloadRes = await fetch(`${base}${dbLogoUrl}`);
    assert.equal(
      downloadRes.status,
      200,
      "Public logo attachment must be retrievable without auth",
    );
    const contentType = downloadRes.headers.get("content-type");
    assert(
      contentType.includes("image/png"),
      `Expected image/png, got ${contentType}`,
    );

    // Test logo removal
    await page.click('button:has-text("Remove")');
    await page.click('button:has-text("Save Settings")');
    await page.waitForSelector(
      'div[role="status"]:has-text("General settings saved successfully")',
    );

    const dbClearedLogo = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'logo_url'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbClearedLogo,
      "",
      "Logo URL must be cleared in DB upon removal",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "06_settings_logo_attachment.png"),
    });
    console.log(
      "✔ Journey 6 Passed: Public attachment lifecycle (upload, retrieval, removal) verified.",
    );

    // -----------------------------------------------------------------------
    // Journey 7: Hospital Schedule Persistence & Validation
    // -----------------------------------------------------------------------
    console.log("\n[Journey 7] Hospital Schedule Persistence & Validation...");
    await page.click('.module-subtab-link:has-text("Hospital Schedule")');
    await page.waitForSelector('h2:has-text("Hospital Schedule")');

    // Test validation: opening after closing
    const monOpen = page.locator('input[aria-label="Monday Opening time"]');
    const monClose = page.locator('input[aria-label="Monday Closing time"]');
    await monOpen.fill("18:00");
    await monClose.fill("09:00");
    await page.click('button:has-text("Save Schedule")');

    const schedAlert = page.locator('div.alert-notice[role="alert"]');
    await schedAlert.waitFor({ state: "visible", timeout: 5000 });
    const schedAlertText = await schedAlert.textContent();
    assert(
      schedAlertText.includes("Closing time must be after opening time"),
      "Expected hours validation",
    );

    // Fix valid times: Monday 08:30 to 17:30
    await monOpen.fill("08:30");
    await monClose.fill("17:30");
    await page.click('button:has-text("Save Schedule")');
    await page.waitForSelector(
      'div[role="status"]:has-text("Hospital schedule saved successfully")',
    );

    // Query DB
    const monSched = (
      await db.query(
        "SELECT start_time, end_time, is_closed FROM hospital_schedule_day WHERE day_of_week = 1",
      )
    ).rows[0];
    assert.equal(monSched.start_time, "08:30");
    assert.equal(monSched.end_time, "17:30");
    assert.equal(monSched.is_closed, false);

    await page.reload();
    const reloadedMonOpen = await page
      .locator('input[aria-label="Monday Opening time"]')
      .inputValue();
    assert.equal(
      reloadedMonOpen,
      "08:30",
      "Schedule opening time must persist after reload",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "07_hospital_schedule_persisted.png"),
    });
    console.log(
      "✔ Journey 7 Passed: Hospital schedule validation and atomic persistence verified.",
    );

    // -----------------------------------------------------------------------
    // Journey 8: Modules Setting Toggle & Persistence
    // -----------------------------------------------------------------------
    console.log("\n[Journey 8] Modules Setting Toggle & Persistence...");
    await page.click('.module-subtab-link:has-text("Modules Setting")');
    await page.waitForSelector('h2:has-text("Modules Setting")');

    const initialAppointmentActive = (
      await db.query(
        "SELECT is_active FROM hospital_module_setting WHERE module_key = 'appointments'",
      )
    ).rows[0]?.is_active;

    // Toggle appointments module
    const apptCard = page.locator('[data-module-key="appointments"]');
    const toggle = apptCard.locator("label");
    await toggle.click();
    await page.waitForSelector(
      'div[role="status"]:has-text("Module setting updated successfully")',
    );

    const dbUpdatedActive = (
      await db.query(
        "SELECT is_active FROM hospital_module_setting WHERE module_key = 'appointments'",
      )
    ).rows[0]?.is_active;
    assert.equal(
      dbUpdatedActive,
      !initialAppointmentActive,
      "Database is_active must toggle",
    );

    // Toggle back to active
    await toggle.click();
    await page.waitForSelector(
      'div[role="status"]:has-text("Module setting updated successfully")',
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "08_modules_setting_persisted.png"),
    });
    console.log(
      "✔ Journey 8 Passed: Modules setting toggles live API and persists.",
    );

    // -----------------------------------------------------------------------
    // Journey 9: Currencies Active Selection
    // -----------------------------------------------------------------------
    console.log("\n[Journey 9] Currencies Active Selection...");
    await page.click('.module-subtab-link:has-text("Currencies")');
    await page.waitForSelector('h2:has-text("Currencies")');

    // Make USD default
    const usdRow = page.locator('tr:has-text("US Dollar")');
    await usdRow.locator('button:has-text("Make Default")').click();
    await page.waitForSelector(
      'div[role="status"]:has-text("Active currency set to USD")',
    );

    const dbCurr = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'current_currency'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbCurr,
      "USD",
      "Database current_currency must be updated to USD",
    );

    // Set back to ETB
    const etbRow = page.locator('tr:has-text("Ethiopian Birr")');
    await etbRow.locator('button:has-text("Make Default")').click();
    await page.waitForSelector(
      'div[role="status"]:has-text("Active currency set to ETB")',
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "09_currency_settings_persisted.png"),
    });
    console.log(
      "✔ Journey 9 Passed: Currency selection updates active hospital currency.",
    );

    // -----------------------------------------------------------------------
    // Journey 10: Payment Gateways Secret Protection
    // -----------------------------------------------------------------------
    console.log("\n[Journey 10] Payment Gateways Secret Protection...");
    await page.click('.module-subtab-link:has-text("Payment Gateways")');
    await page.waitForSelector('h2:has-text("Payment Gateways")');

    await page
      .locator('label:has-text("Stripe Key") input')
      .fill("pk_test_stripe12345");
    await page
      .locator('label:has-text("Stripe Secret") input')
      .fill("sk_test_stripesecret999");
    await page.click('button:has-text("Save Payment Gateways")');
    await page.waitForSelector(
      'div[role="status"]:has-text("Payment gateway settings saved successfully")',
    );

    const dbStripeSecret = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'stripe_secret'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbStripeSecret,
      "sk_test_stripesecret999",
      "PostgreSQL must store actual Stripe secret",
    );

    // Reload page: verify redacted
    await page.reload();
    const redactedStripe = await page
      .locator('label:has-text("Stripe Secret") input')
      .inputValue();
    assert.equal(
      redactedStripe,
      "[CONFIGURED]",
      "Stripe secret must be redacted to [CONFIGURED]",
    );

    // Save unchanged
    await page.click('button:has-text("Save Payment Gateways")');
    await page.waitForSelector(
      'div[role="status"]:has-text("Payment gateway settings saved successfully")',
    );
    const preservedStripe = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'stripe_secret'",
      )
    ).rows[0]?.value;
    assert.equal(
      preservedStripe,
      "sk_test_stripesecret999",
      "Stripe secret must be preserved",
    );

    // Clear secret
    await page.click(
      'label:has-text("Stripe Secret") button:has-text("Clear")',
    );
    await page.click('button:has-text("Save Payment Gateways")');
    await page.waitForSelector(
      'div[role="status"]:has-text("Payment gateway settings saved successfully")',
    );
    const clearedStripe = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'stripe_secret'",
      )
    ).rows[0]?.value;
    assert.equal(clearedStripe, "", "Stripe secret must be cleared");

    await page.screenshot({
      path: resolve(screenshotsDir, "10_payment_gateways_persisted.png"),
    });
    console.log(
      "✔ Journey 10 Passed: Payment gateways secrets protected, preserved, and cleared.",
    );

    // -----------------------------------------------------------------------
    // Journey 11: Patient Queue Theme Persistence
    // -----------------------------------------------------------------------
    console.log("\n[Journey 11] Patient Queue Theme Persistence...");
    await page.click('.module-subtab-link:has-text("Patient Queue Theme")');
    await page.waitForSelector('h2:has-text("Patient Queue Theme")');

    await page
      .locator('label:has-text("Queue Display Message") input')
      .fill("Now Serving OPD Room 4");
    await page.selectOption(
      'label:has-text("Theme Preset") select',
      "emerald-care",
    );
    await page.click('button:has-text("Save Queue Theme")');
    await page.waitForSelector(
      'div[role="status"]:has-text("Patient queue theme saved successfully")',
    );

    const dbThemeMsg = (
      await db.query(
        "SELECT value FROM hospital_general_setting WHERE key = 'queue_theme_message'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbThemeMsg,
      "Now Serving OPD Room 4",
      "DB queue_theme_message must match",
    );

    await page.reload();
    const reloadedMsg = await page
      .locator('label:has-text("Queue Display Message") input')
      .inputValue();
    assert.equal(
      reloadedMsg,
      "Now Serving OPD Room 4",
      "Queue message must persist after reload",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "11_queue_theme_persisted.png"),
    });
    console.log(
      "✔ Journey 11 Passed: Patient queue theme persists to backend.",
    );

    // -----------------------------------------------------------------------
    // Journey 12: Front CMS Settings Persistence
    // -----------------------------------------------------------------------
    console.log("\n[Journey 12] Front CMS Settings Persistence...");
    await page.goto(`${base}/modules/front-settings`);
    await page.waitForSelector('h2:has-text("Front Setting Details")');

    const homeTitleInput = page.locator(
      'label:has-text("Home Page Title") input',
    );
    await homeTitleInput.fill("Comprehensive Compassionate Healthcare Center");
    await page.click('button:has-text("Save Front CMS Settings")');
    await page.waitForSelector(
      'div[role="status"]:has-text("Front CMS settings saved successfully")',
    );

    const dbHomeTitle = (
      await db.query(
        "SELECT value FROM front_cms_setting WHERE key = 'home_title'",
      )
    ).rows[0]?.value;
    assert.equal(
      dbHomeTitle,
      "Comprehensive Compassionate Healthcare Center",
      "Front CMS home_title in DB must match",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "12_front_cms_persisted.png"),
    });
    console.log(
      "✔ Journey 12 Passed: Front CMS settings persist to PostgreSQL.",
    );

    // -----------------------------------------------------------------------
    // Journey 13: Non-Admin Rejection
    // -----------------------------------------------------------------------
    console.log("\n[Journey 13] Non-Admin Rejection...");
    const docContext = await browser.newContext();
    const docPage = await docContext.newPage();
    await performLogin(docPage, docEmail, docPassword);

    await docPage.goto(`${base}/modules/settings`);
    await docPage.waitForSelector('h2:has-text("Access Denied")', {
      timeout: 5000,
    });
    const accessDeniedVisible = await docPage.isVisible(
      'h2:has-text("Access Denied")',
    );
    assert(
      accessDeniedVisible,
      "Non-admin doctor must be blocked by Access Denied banner",
    );

    // Also verify API 403 response
    const apiRes = await docPage.evaluate(async () => {
      const res = await fetch("/api/hms/general-settings");
      return { status: res.status };
    });
    console.log("Doctor direct API fetch status:", apiRes.status);
    assert.equal(
      apiRes.status,
      403,
      "Doctor direct GET /api/hms/general-settings must return 403 Forbidden",
    );

    await docPage.screenshot({
      path: resolve(screenshotsDir, "13_non_admin_rejection.png"),
    });
    await docContext.close();
    console.log(
      "✔ Journey 13 Passed: Non-admin rejected with Access Denied banner and 403 Forbidden.",
    );

    console.log("\n=======================================================");
    console.log("ALL 13 SETTINGS WORKSPACE JOURNEYS PASSED SUCCESSFULLY!");
    console.log("=======================================================\n");
  } finally {
    await browser.close();
    await db.end();
  }
}

main().catch((err) => {
  console.error("FATAL: Settings workspace verification failed:", err);
  process.exit(1);
});
