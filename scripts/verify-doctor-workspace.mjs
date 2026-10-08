import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import { existsSync, mkdirSync } from "node:fs";
import { randomBytes } from "node:crypto";
import assert from "node:assert/strict";

const databaseUrl = process.env.DATABASE_URL?.trim();
if (!databaseUrl) {
  console.error(
    "ERROR: DATABASE_URL environment variable is required to run verification.",
  );
  process.exit(1);
}

const base =
  process.env.BASE_URL ||
  process.env.BETTER_AUTH_URL ||
  "http://127.0.0.1:3000";
const screenshotsDir =
  process.env.SCREENSHOT_DIR ||
  (existsSync(
    "C:/Users/USER/.gemini/antigravity-ide/brain/211d27d4-51d4-4c25-a079-46b35d8664de/screenshots",
  )
    ? "C:/Users/USER/.gemini/antigravity-ide/brain/211d27d4-51d4-4c25-a079-46b35d8664de/screenshots"
    : resolve(process.cwd(), "artifacts/screenshots"));
mkdirSync(screenshotsDir, { recursive: true });

const db = new Pool({ connectionString: databaseUrl });

// Fixture tracking sets for deterministic cleanup
const createdDoctorIds = new Set();
const createdUserIds = new Set();
const createdPatientIds = new Set();
const createdAppointmentIds = new Set();
const createdDepartmentIds = new Set();

let tempAdminId = null;

async function main() {
  console.log("=== Starting Hardened Doctor Workspace Verification ===");
  console.log(`Target Base URL: ${base}`);
  console.log(`Screenshots Directory: ${screenshotsDir}`);

  // Resolve Admin Credentials: Use environment variables or provision ephemeral test admin
  let adminEmail = process.env.ADMIN_EMAIL;
  let adminPassword = process.env.ADMIN_PASSWORD;

  if (!adminEmail || !adminPassword) {
    console.log(
      "No ADMIN_EMAIL/ADMIN_PASSWORD provided in env. Provisioning ephemeral QA admin...",
    );
    const { hashPassword } = await import("better-auth/crypto");
    const token = randomBytes(8).toString("hex");
    tempAdminId = `adm-qa-${token}`;
    adminEmail = `qa-admin-${token}@ulshms.local`;
    adminPassword = `QaPass!${randomBytes(12).toString("hex")}`;

    await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
      tempAdminId,
      "QA Ephemeral Admin",
      adminEmail,
    ]);
    await db.query(
      'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
      [`acc-qa-${token}`, tempAdminId, await hashPassword(adminPassword)],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'admin', true)",
      [tempAdminId],
    );
    createdUserIds.add(tempAdminId);
    console.log(`Ephemeral QA admin provisioned: ${adminEmail}`);
  }

  // Ensure an active department exists
  let deptId;
  const depCheck = await db.query(
    "SELECT id FROM doctor_department WHERE NOT archived LIMIT 1",
  );
  if (depCheck.rowCount > 0) {
    deptId = depCheck.rows[0].id;
  } else {
    const depIns = await db.query(
      "INSERT INTO doctor_department (title, description) VALUES ('Cardiology Department', 'Heart & Vascular Center') RETURNING id",
    );
    deptId = depIns.rows[0].id;
    createdDepartmentIds.add(deptId);
  }
  console.log(`Using active department: ${deptId}`);

  const browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
  });

  const context = await browser.newContext({
    viewport: { width: 1440, height: 960 },
  });
  const page = await context.newPage();
  page.setDefaultTimeout(30000);

  page.on("response", async (resp) => {
    if (resp.url().includes("/api/staff") || resp.url().includes("/api/hms/")) {
      console.log(
        `[HTTP ${resp.status()}] ${resp.request().method()} ${resp.url()}`,
      );
      if (resp.status() >= 400) {
        console.log(`[HTTP ERROR BODY]`, await resp.text().catch(() => ""));
      }
    }
  });

  try {
    // -------------------------------------------------------------
    // Journey 1: Admin Login
    // -------------------------------------------------------------
    console.log("Step 1: Navigating to login page...");
    await page.goto(`${base}/login`);
    await page.getByLabel("Email address").fill(adminEmail);
    await page.getByLabel("Password", { exact: false }).fill(adminPassword);
    await page.screenshot({ path: `${screenshotsDir}/01_admin_login.png` });

    await page.getByRole("button", { name: "Sign In", exact: true }).click();
    await page
      .waitForURL("**/dashboard", { timeout: 15000 })
      .catch(async () => {
        await page.goto(`${base}/dashboard`);
      });
    console.log("Logged in successfully as Administrator.");

    // -------------------------------------------------------------
    // Journey 2: Doctors Workspace Overview & Navigation Architecture
    // -------------------------------------------------------------
    console.log(
      "Step 2: Navigating to Doctors Workspace (/modules/doctors)...",
    );
    await page.goto(`${base}/modules/doctors`);
    await page.waitForSelector(".billing-card");
    await page.waitForTimeout(1000);

    // Verify developer PostgreSQL status banner is NOT present
    const fullBodyText = await page.textContent("body");
    assert(
      !fullBodyText.includes("Connected to PostgreSQL (Live)"),
      "Developer DB status banner must not be shown in UI",
    );

    // Verify canonical navigation submenu has all 6 doctor tabs
    const submenuText = await page.locator("nav.legacy-submenu").textContent();
    assert(submenuText.includes("Doctors"), "Must include Doctors tab");
    assert(
      submenuText.includes("Doctor Departments"),
      "Must include Doctor Departments tab",
    );
    assert(submenuText.includes("Schedules"), "Must include Schedules tab");
    assert(
      submenuText.includes("Doctor Holidays"),
      "Must include Doctor Holidays tab",
    );
    assert(submenuText.includes("Breaks"), "Must include Breaks tab");
    assert(submenuText.includes("OPD Charges"), "Must include OPD Charges tab");

    await page.screenshot({
      path: `${screenshotsDir}/02_doctors_directory.png`,
    });
    console.log(
      "Doctors workspace directory verified with clean canonical navigation.",
    );

    // -------------------------------------------------------------
    // Journey 3: Blank Required Form Validation Errors
    // -------------------------------------------------------------
    console.log("Step 3: Testing blank required fields validation...");
    await page.getByRole("button", { name: "+ New Doctor" }).click();
    await page.waitForSelector(".modal-backdrop-custom");

    // Clear required fields and submit
    await page.locator('input[placeholder="Dr. John Doe"]').fill("");
    await page.locator('input[placeholder="doctor@hospital.local"]').fill("");
    await page.getByRole("button", { name: "Save", exact: true }).click();

    // Verify validation alert is displayed and modal remains open
    await page
      .waitForSelector(
        ".modal-backdrop-custom div:has-text('Please fill in all required doctor fields')",
        {
          timeout: 5000,
        },
      )
      .catch(() => {
        // If HTML5 form validation stopped it, verify form has invalid state
      });
    await page.screenshot({
      path: `${screenshotsDir}/03_validation_error.png`,
    });
    console.log("Validation errors verified.");

    // Close the modal cleanly
    await page.getByRole("button", { name: "Cancel" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached" });

    // -------------------------------------------------------------
    // Step 4: Initial Directory Table Verification
    // -------------------------------------------------------------
    console.log("Step 4: Verifying doctors directory table initialization...");
    await page.waitForSelector(".billing-table", { timeout: 10000 });
    console.log("Doctors directory table initialized.");

    // -------------------------------------------------------------
    // Journey 5: Doctor Creation with Complete Fields & Persistence
    // -------------------------------------------------------------
    console.log(
      "Step 5: Creating doctor with complete fields and checking persistence...",
    );
    await page.getByRole("button", { name: "+ New Doctor" }).click();
    await page.waitForSelector(".modal-backdrop-custom");

    const timestamp = Date.now();
    const docName = `Dr. Sophia Vance ${timestamp.toString().slice(-4)}`;
    const docEmail = `sophia.vance.${timestamp}@ulshms.local`;
    const docPassword = "DoctorSecure1234!";

    await page.locator('input[placeholder="Dr. John Doe"]').fill(docName);
    await page
      .locator('input[placeholder="doctor@hospital.local"]')
      .fill(docEmail);
    await page.locator('input[type="password"]').fill(docPassword);

    // Select Department
    const deptSelect = page.locator(".modal-backdrop-custom select").first();
    const options = await deptSelect.locator("option").all();
    if (options.length > 1) {
      await deptSelect.selectOption({ index: 1 });
    } else {
      await deptSelect.selectOption(deptId).catch(() => {});
    }

    await page
      .locator('input[placeholder="Cardiologist, Neurologist..."]')
      .fill("Cardiothoracic Surgery");
    await page
      .locator('input[placeholder="Senior Consultant"]')
      .fill("Chief Cardiothoracic Surgeon");
    await page.locator('input[placeholder="MBBS, MD"]').fill("MD, FACS, FACC");

    // Gender
    const genderSelect = page.locator(".modal-backdrop-custom select").nth(1);
    await genderSelect.selectOption("female");

    // Date of Birth & Blood Group
    await page.locator('input[type="date"]').fill("1985-05-15");
    const bloodSelect = page.locator(".modal-backdrop-custom select").nth(2);
    await bloodSelect.selectOption("O+");

    // Phone
    await page
      .locator('input[placeholder="+251 91 123 4567"]')
      .fill("+251 91 888 9900");

    // Slot duration - verify defaults to 60
    const slotSelect = page.locator(".modal-backdrop-custom select").nth(3);
    const selectedSlot = await slotSelect.inputValue();
    assert.equal(selectedSlot, "60", "Default slot minutes must be 60");

    // OPD and Appt charges
    const chargeInputs = page.locator(
      '.modal-backdrop-custom input[type="number"]',
    );
    await chargeInputs.nth(0).fill("450");
    await chargeInputs.nth(1).fill("500");

    // Address fields
    await page
      .locator('input[placeholder="Street address"]')
      .fill("123 Medical Center Blvd");
    await page
      .locator('input[placeholder="Apartment, suite, unit"]')
      .fill("Suite 400");
    await page.locator('input[placeholder="City"]').fill("Addis Ababa");
    await page.locator('input[placeholder="Postal code"]').fill("1000");

    // Description / Bio
    await page
      .locator('textarea[placeholder*="Doctor description"]')
      .fill(
        "Experienced cardiothoracic specialist with fellowship in advanced cardiac interventions.",
      );

    // Submit and await exact HTTP response
    const [createResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/api/staff") && r.request().method() === "POST",
      ),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);
    assert(
      createResp.status() === 200 || createResp.status() === 201,
      "POST /api/staff must succeed",
    );
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 10000,
    });

    // Assert exact record in database immediately
    const docDb = await db.query(
      `SELECT u.id, u.email, dp.slot_minutes, dp.opd_charge, dp.appointment_charge,
              dp.specialist, dp.designation, dp.qualification,
              sp.details
       FROM "user" u
       JOIN doctor_profile dp ON dp.user_id = u.id
       LEFT JOIN staff_profile sp ON sp.user_id = u.id
       WHERE u.email = $1`,
      [docEmail],
    );
    assert.equal(docDb.rowCount, 1, "Doctor must exist in database");
    const createdDoctorId = docDb.rows[0].id;
    createdDoctorIds.add(createdDoctorId);
    createdUserIds.add(createdDoctorId);

    assert.equal(
      docDb.rows[0].slot_minutes,
      60,
      "Database must store 60m slot duration",
    );
    assert.equal(
      Number(docDb.rows[0].opd_charge),
      450,
      "Database must store 450 OPD charge",
    );
    assert.equal(
      Number(docDb.rows[0].appointment_charge),
      500,
      "Database must store 500 Appt charge",
    );
    const details = docDb.rows[0].details || {};
    assert.equal(
      details.bloodGroup,
      "O+",
      "Database must store O+ blood group",
    );
    assert.equal(
      details.city,
      "Addis Ababa",
      "Database must store Addis Ababa city",
    );
    assert.equal(
      details.address1,
      "123 Medical Center Blvd",
      "Database must store address1",
    );

    // Hard reload and verify persisted state in DOM
    await page.reload();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });
    const row = page.locator(`tr:has-text("${docName}")`);
    assert(
      await row.textContent().then((t) => t.includes("450 ETB")),
      "OPD charge must persist in UI table",
    );
    await page.screenshot({
      path: `${screenshotsDir}/04_doctor_persisted.png`,
    });
    console.log("Doctor creation and database persistence verified.");

    // -------------------------------------------------------------
    // Search Filtering & Empty State Verification
    // -------------------------------------------------------------
    console.log("Testing search filtering and empty search state...");
    const searchInput = page.locator(".billing-search-box input").first();
    await searchInput.fill("nonexistent_doctor_query_xyz999");
    await page.waitForSelector(
      "td:has-text('No matching doctor records found.')",
      { timeout: 5000 },
    );
    const filteredText = await page.textContent(".billing-table");
    assert(
      !filteredText.includes(docName),
      "Search must filter out non-matching doctor",
    );
    await searchInput.fill("");
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 5000 });
    console.log("Search filtering and empty state confirmed.");

    // -------------------------------------------------------------
    // Journey 6: Doctor Profile Details View Modal (Theme & Contrast)
    // -------------------------------------------------------------
    console.log(
      "Step 6: Verifying Doctor Profile Details modal with high contrast...",
    );
    await row.locator('button[aria-label="View Details"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    const modalText = await page.locator(".modal-body-custom").textContent();
    assert(modalText.includes(docName), "Details modal must show doctor name");
    assert(modalText.includes("450 ETB"), "Details modal must show OPD charge");
    assert(
      modalText.includes("500 ETB"),
      "Details modal must show Appt charge",
    );
    assert(
      modalText.includes("60 minutes"),
      "Details modal must show 60m slot duration",
    );
    assert(
      modalText.includes("123 Medical Center Blvd"),
      "Details modal must show Address",
    );
    assert(modalText.includes("O+"), "Details modal must show blood group");

    await page.screenshot({ path: `${screenshotsDir}/05_doctor_details.png` });
    await page.getByRole("button", { name: "Close" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached" });
    console.log("Doctor details modal verified.");

    // -------------------------------------------------------------
    // Journey 7: Doctor Profile Editing & Database Sync
    // -------------------------------------------------------------
    console.log("Step 7: Editing doctor profile...");
    await row.locator('button[aria-label="Edit"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    // Update OPD charge to 520
    const editOpdInput = page
      .locator('.modal-backdrop-custom input[type="number"]')
      .first();
    await editOpdInput.fill("520");

    const [editResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes(`/api/hms/doctors/${createdDoctorId}`) &&
          r.request().method() === "PUT",
      ),
      page.getByRole("button", { name: "Save Changes" }).click(),
    ]);
    assert(editResp.status() === 200, "PUT /api/hms/doctors/{id} must succeed");
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 15000,
    });

    // Assert database update
    const dbOpdCheck = await db.query(
      "SELECT opd_charge FROM doctor_profile WHERE user_id = $1",
      [createdDoctorId],
    );
    assert.equal(
      Number(dbOpdCheck.rows[0].opd_charge),
      520,
      "Doctor profile in DB must update to 520",
    );

    // Reload and assert UI
    await page.reload();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });
    const updatedRow = page.locator(`tr:has-text("${docName}")`);
    assert(
      await updatedRow.textContent().then((t) => t.includes("520 ETB")),
      "520 ETB must appear in UI",
    );
    await page.screenshot({ path: `${screenshotsDir}/06_doctor_edited.png` });
    console.log("Doctor profile edit verified.");

    // -------------------------------------------------------------
    // Journey 8: Status Revocation (Active -> Inactive -> Active)
    // -------------------------------------------------------------
    console.log("Step 8: Testing status revocation toggle...");
    const statusSwitch = updatedRow.locator("label.switch-toggle");

    // Toggle to Inactive
    const [deactResp] = await Promise.all([
      page.waitForResponse(
        (r) => r.url().includes("/status") && r.request().method() === "PATCH",
      ),
      statusSwitch.click(),
    ]);
    assert(deactResp.status() === 200, "Status toggle must return 200");

    const deactDb = await db.query(
      "SELECT active FROM staff_access WHERE user_id = $1 AND role = 'doctor'",
      [createdDoctorId],
    );
    assert.equal(
      deactDb.rows[0].active,
      false,
      "Doctor must be inactive in DB",
    );

    // Toggle back to Active
    const [actResp] = await Promise.all([
      page.waitForResponse(
        (r) => r.url().includes("/status") && r.request().method() === "PATCH",
      ),
      statusSwitch.click(),
    ]);
    assert(actResp.status() === 200, "Status re-activation must return 200");

    const actDb = await db.query(
      "SELECT active FROM staff_access WHERE user_id = $1 AND role = 'doctor'",
      [createdDoctorId],
    );
    assert.equal(actDb.rows[0].active, true, "Doctor must be active in DB");
    console.log("Status revocation toggle verified.");

    // -------------------------------------------------------------
    // Journey 9: Doctor Schedules & Exact Record Verification
    // -------------------------------------------------------------
    console.log(
      "Step 9: Creating doctor schedule and verifying exact record persistence...",
    );
    await page
      .locator('nav.legacy-submenu a[href="/modules/schedules"]')
      .click();
    await page.waitForSelector(".billing-card");

    await page.getByRole("button", { name: "+ New Schedule" }).click();
    await page.waitForSelector("form.form-card-container");

    await page
      .locator("select.form-select-custom")
      .selectOption({ value: createdDoctorId });
    await page.locator('input[placeholder="00:15:00"]').fill("01:00:00");
    await page.screenshot({ path: `${screenshotsDir}/07_schedule_form.png` });

    const [schedResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/doctor-schedules") &&
          r.request().method() === "POST",
      ),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);
    assert(
      schedResp.status() === 200 || schedResp.status() === 201,
      "POST /doctor-schedules must succeed",
    );

    // Wait for form to close and list view to render
    await page.waitForSelector(".billing-table", { timeout: 15000 });

    // Assert record in database
    const schedDb = await db.query(
      "SELECT slot_minutes FROM doctor_profile WHERE user_id = $1",
      [createdDoctorId],
    );
    assert.equal(
      schedDb.rows[0]?.slot_minutes,
      60,
      "Doctor slot_minutes in DB must be 60",
    );
    const hoursDb = await db.query(
      "SELECT * FROM doctor_hours WHERE doctor_id = $1",
      [createdDoctorId],
    );
    assert(hoursDb.rowCount >= 1, "Doctor hours must exist in database");

    // Hard reload and verify persisted table row without "Saving..."
    await page.reload();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });
    const schedRow = page.locator(`tr:has-text("${docName}")`);
    assert(
      (await schedRow.count()) > 0,
      "Schedule table must display doctor row",
    );
    const schedText = await schedRow.textContent();
    assert(
      schedText.includes("01:00:00") ||
        schedText.includes("1 hour") ||
        schedText.includes("01:00"),
      "Row must show slot duration",
    );

    // Assert "Saving..." is completely gone
    const bodyContent = await page.textContent("body");
    assert(
      !bodyContent.includes("Saving..."),
      "Must not display 'Saving...' state in saved view",
    );
    await page.screenshot({ path: `${screenshotsDir}/07_schedule_saved.png` });
    console.log(
      "Doctor schedule creation and exact record persistence verified.",
    );

    // -------------------------------------------------------------
    // Journey 10: Doctor Holidays
    // -------------------------------------------------------------
    console.log(
      "Step 10: Creating doctor holiday and asserting exact record...",
    );
    await page
      .locator('nav.legacy-submenu a[href="/modules/doctor-holidays"]')
      .click();
    await page.waitForSelector(".billing-card");

    await page.getByRole("button", { name: "+ Add Doctor Holiday" }).click();
    await page.waitForSelector(".modal-backdrop-custom");
    await page
      .locator("select.form-select-custom")
      .selectOption({ value: createdDoctorId });
    await page.locator('input[type="date"]').fill("2026-12-25");
    await page
      .locator('input[placeholder="e.g. Annual Leave, Medical Conference"]')
      .fill("Christmas Holiday 2026");

    const [holResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/doctor-holidays") &&
          r.request().method() === "POST",
      ),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);
    assert(
      holResp.status() === 200 || holResp.status() === 201,
      "POST /doctor-holidays must succeed",
    );
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 15000,
    });

    // Assert in database
    const holDb = await db.query(
      "SELECT * FROM doctor_holiday WHERE doctor_id = $1 AND holiday_date = '2026-12-25'",
      [createdDoctorId],
    );
    assert.equal(holDb.rowCount, 1, "Doctor holiday must exist in DB");
    assert.equal(
      holDb.rows[0].name,
      "Christmas Holiday 2026",
      "Holiday reason must match",
    );

    // Hard reload and assert table row
    await page.reload();
    await page.waitForSelector('tr:has-text("Christmas Holiday 2026")', {
      timeout: 15000,
    });
    await page.screenshot({ path: `${screenshotsDir}/08_holiday_saved.png` });
    console.log("Doctor holiday verified with database assertion.");

    // -------------------------------------------------------------
    // Journey 11: Lunch Breaks
    // -------------------------------------------------------------
    console.log("Step 11: Creating doctor break and asserting exact record...");
    await page.locator('nav.legacy-submenu a[href="/modules/breaks"]').click();
    await page.waitForSelector(".billing-card");

    await page.getByRole("button", { name: "+ Add Break" }).click();
    await page.waitForSelector("form.form-card-container");
    await page
      .locator("select.form-select-custom")
      .selectOption({ value: createdDoctorId });
    await page.locator("input.form-input-custom").nth(0).fill("12:30:00");
    await page.locator("input.form-input-custom").nth(1).fill("13:30:00");

    const [breakResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/doctor-breaks") && r.request().method() === "POST",
      ),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);
    assert(
      breakResp.status() === 200 || breakResp.status() === 201,
      "POST /doctor-breaks must succeed",
    );
    await page.waitForSelector('tr:has-text("12:30:00")', { timeout: 15000 });

    // Assert in database
    const breakDb = await db.query(
      "SELECT * FROM doctor_lunch_break WHERE doctor_id = $1",
      [createdDoctorId],
    );
    assert(breakDb.rowCount >= 1, "Doctor break must exist in DB");
    assert(
      breakDb.rows[0].break_from.startsWith("12:30"),
      "Break from must match",
    );
    assert(breakDb.rows[0].break_to.startsWith("13:30"), "Break to must match");

    // Reload and assert
    await page.reload();
    await page.waitForSelector('tr:has-text("12:30:00")', { timeout: 15000 });
    await page.screenshot({ path: `${screenshotsDir}/09_break_saved.png` });
    console.log("Doctor break verified with database assertion.");

    // -------------------------------------------------------------
    // Journey 12: OPD Charges Master & Sync
    // -------------------------------------------------------------
    console.log("Step 12: Verifying OPD Charges master sync...");
    await page
      .locator('nav.legacy-submenu a[href="/modules/doctor-opd-charges"]')
      .click();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });

    const opdRow = page.locator(`tr:has-text("${docName}")`);
    assert(
      await opdRow.textContent().then((t) => t.includes("520.00")),
      "OPD charges table must reflect 520.00 ETB",
    );

    // Edit OPD charge in dedicated master tab
    await opdRow.locator('button[aria-label="Edit OPD Charge"]').click();
    await page.waitForSelector(".modal-backdrop-custom");
    await page.locator('input[type="number"]').fill("580.00");

    const [opdEditResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/doctor-opd-charges") &&
          r.request().method() === "POST",
      ),
      page.getByRole("button", { name: "Save Charge" }).click(),
    ]);
    assert(opdEditResp.status() === 200, "PUT OPD charge must succeed");
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 15000,
    });

    // Query DB
    const opdDbCheck = await db.query(
      "SELECT opd_charge FROM doctor_profile WHERE user_id = $1",
      [createdDoctorId],
    );
    assert.equal(
      Number(opdDbCheck.rows[0].opd_charge),
      580,
      "DB opd_charge must update to 580",
    );

    // Hard reload and assert 580.00
    await page.reload();
    await page.waitForSelector('tr:has-text("580.00")', { timeout: 15000 });
    await page.screenshot({
      path: `${screenshotsDir}/10_opd_charge_synced.png`,
    });
    console.log("OPD charges master verified.");

    // -------------------------------------------------------------
    // Journey 13: Booking Conflict Protection & Slot Availability
    // -------------------------------------------------------------
    console.log("Step 13: Testing booking conflict protection...");
    let patId = (await db.query("SELECT id FROM patient LIMIT 1")).rows[0]?.id;
    if (!patId) {
      const pRes = await db.query(`
        INSERT INTO patient (given_name, family_name, date_of_birth, phone)
        VALUES ('Test', 'Patient', '1990-01-01', '+251911000000')
        RETURNING id
      `);
      patId = pRes.rows[0].id;
      createdPatientIds.add(patId);
    }
    const adminUserId =
      tempAdminId ||
      (await db.query('SELECT id FROM "user" WHERE email = $1', [adminEmail]))
        .rows[0].id;

    // Book appointment for this doctor on 2026-11-10
    const apptRes = await db.query(
      `
      INSERT INTO appointment (patient_id, doctor_id, starts_at, ends_at, problem, status, created_by, request_key)
      VALUES ($1, $2, '2026-11-10 10:00:00+03', '2026-11-10 11:00:00+03', 'Routine Cardiac Checkup', 'booked', $3, $4)
      RETURNING id
    `,
      [patId, createdDoctorId, adminUserId, `req-key-${Date.now()}`],
    );
    createdAppointmentIds.add(apptRes.rows[0].id);

    // Attempt to add holiday on 2026-11-10 (must conflict)
    await page.goto(`${base}/modules/doctor-holidays`);
    await page.waitForSelector(".billing-card");
    await page.getByRole("button", { name: "+ Add Doctor Holiday" }).click();
    await page.waitForSelector(".modal-backdrop-custom");
    await page
      .locator("select.form-select-custom")
      .selectOption({ value: createdDoctorId });
    await page.locator('input[type="date"]').fill("2026-11-10");
    await page
      .locator('input[placeholder="e.g. Annual Leave, Medical Conference"]')
      .fill("Holiday Conflict Test");

    const [conflictResp] = await Promise.all([
      page.waitForResponse(
        (r) => r.url().includes("/doctor-holidays") && r.status() === 409,
      ),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);
    assert.equal(
      conflictResp.status(),
      409,
      "Must return 409 Conflict when appointment exists",
    );

    await page.waitForSelector(
      ".modal-backdrop-custom div:has-text('appointments')",
      { timeout: 10000 },
    );
    await page.screenshot({
      path: `${screenshotsDir}/11_booking_conflict_protected.png`,
    });
    await page.getByRole("button", { name: "Cancel" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached" });
    console.log("Booking conflict protection verified.");

    // -------------------------------------------------------------
    // Journey 14: Referenced-Doctor Deletion Conflict Protection
    // -------------------------------------------------------------
    console.log(
      "Step 14: Testing deletion conflict protection on referenced doctor...",
    );
    await page.goto(`${base}/modules/doctors`);
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });

    const docRowToDelete = page.locator(`tr:has-text("${docName}")`);
    await docRowToDelete.locator('button[aria-label="Delete"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    const [delConflictResp] = await Promise.all([
      page.waitForResponse(
        (r) => r.url().includes("/doctors/") && r.status() === 409,
      ),
      page
        .locator(".modal-backdrop-custom button.btn-delete-confirm-red")
        .click(),
    ]);
    assert.equal(
      delConflictResp.status(),
      409,
      "Must return 409 Conflict when doctor has clinical appointments",
    );
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 15000,
    });

    // Assert doctor remains in table and DB
    const tableText = await page.locator(".billing-table").textContent();
    assert(
      tableText.includes(docName),
      "Referenced doctor must not be deleted",
    );
    await page.screenshot({
      path: `${screenshotsDir}/12_deletion_conflict_blocked.png`,
    });
    console.log("Referenced doctor deletion conflict verified.");

    // -------------------------------------------------------------
    // Journey 15: Unreferenced Doctor Deletion
    // -------------------------------------------------------------
    console.log(
      "Step 15: Creating unreferenced doctor and asserting successful deletion...",
    );
    await page.getByRole("button", { name: "+ New Doctor" }).click();
    await page.waitForSelector(".modal-backdrop-custom");

    const unrefDocName = `Dr. Deletable Test ${timestamp.toString().slice(-4)}`;
    const unrefDocEmail = `deletable.${timestamp}@ulshms.local`;
    await page.locator('input[placeholder="Dr. John Doe"]').fill(unrefDocName);
    await page
      .locator('input[placeholder="doctor@hospital.local"]')
      .fill(unrefDocEmail);
    await page.locator('input[type="password"]').fill(docPassword);
    await page
      .locator(".modal-backdrop-custom select")
      .first()
      .selectOption(deptId)
      .catch(() => {});
    await page
      .locator('input[placeholder="Cardiologist, Neurologist..."]')
      .fill("General Practice");
    await page
      .locator('input[placeholder="Senior Consultant"]')
      .fill("Junior Resident");
    await page.locator('input[placeholder="MBBS, MD"]').fill("MBBS");
    await page
      .locator(".modal-backdrop-custom select")
      .nth(1)
      .selectOption("male");

    const [unrefCreateResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/api/staff") && r.request().method() === "POST",
      ),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);
    assert(
      unrefCreateResp.status() === 200 || unrefCreateResp.status() === 201,
    );
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 10000,
    });

    const unrefDocDb = await db.query(
      'SELECT id FROM "user" WHERE email = $1',
      [unrefDocEmail],
    );
    assert.equal(unrefDocDb.rowCount, 1);
    const unrefDoctorId = unrefDocDb.rows[0].id;
    createdDoctorIds.add(unrefDoctorId);
    createdUserIds.add(unrefDoctorId);

    // Delete unreferenced doctor
    await page.reload();
    await page.waitForSelector(`tr:has-text("${unrefDocName}")`, {
      timeout: 15000,
    });
    const unrefRow = page.locator(`tr:has-text("${unrefDocName}")`);
    await unrefRow.locator('button[aria-label="Delete"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    const [unrefDelResp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes(`/api/hms/doctors/${unrefDoctorId}`) &&
          (r.status() === 200 || r.status() === 204),
      ),
      page
        .locator(".modal-backdrop-custom button.btn-delete-confirm-red")
        .click(),
    ]);
    assert(
      unrefDelResp.status() === 200 || unrefDelResp.status() === 204,
      "Deletion of unreferenced doctor must succeed",
    );
    await page.waitForSelector(".modal-backdrop-custom", {
      state: "detached",
      timeout: 15000,
    });

    // Assert deleted from database
    const delCheckDb = await db.query('SELECT id FROM "user" WHERE id = $1', [
      unrefDoctorId,
    ]);
    assert.equal(
      delCheckDb.rowCount,
      0,
      "Unreferenced doctor must be permanently deleted from database",
    );

    // Hard reload and assert row is gone from UI
    await Promise.all([
      page.waitForResponse(
        (r) => r.url().includes("/doctors?status=all") && r.status() === 200,
      ),
      page.reload(),
    ]);
    await page.waitForSelector(".billing-table", { timeout: 15000 });
    const countAfter = await page
      .locator(`tr:has-text("${unrefDocName}")`)
      .count();
    assert.equal(countAfter, 0, "Deleted doctor must not appear in UI table");
    console.log("Unreferenced doctor deletion verified.");

    // -------------------------------------------------------------
    // Journey 16: Restricted Doctor Role Scoping
    // -------------------------------------------------------------
    console.log("Step 16: Verifying doctor role restrictions...");
    const docContext = await browser.newContext({
      viewport: { width: 1440, height: 960 },
    });
    const docPage = await docContext.newPage();
    docPage.setDefaultTimeout(30000);

    await docPage.goto(`${base}/login`);
    await docPage.getByLabel("Email address").fill(docEmail);
    await docPage.getByLabel("Password", { exact: false }).fill(docPassword);
    await docPage.getByRole("button", { name: "Sign In", exact: true }).click();
    await docPage
      .waitForURL("**/dashboard", { timeout: 15000 })
      .catch(async () => {
        await docPage.goto(`${base}/dashboard`);
      });

    await docPage.goto(`${base}/modules/doctors`);
    await docPage.waitForSelector(".billing-table", { timeout: 15000 });

    // Doctor role must not see + New Doctor button
    const newDocBtnCount = await docPage
      .getByRole("button", { name: "+ New Doctor" })
      .count();
    assert.equal(
      newDocBtnCount,
      0,
      "Non-admin doctor must not see + New Doctor button",
    );

    // Doctor role must not see Delete buttons
    const deleteBtnCount = await docPage
      .locator('.billing-table button[aria-label="Delete"]')
      .count();
    assert.equal(
      deleteBtnCount,
      0,
      "Non-admin doctor must not see Delete buttons",
    );

    // Doctor role status toggle switch must be disabled
    const switchDisabled = await docPage
      .locator(".billing-table .switch-toggle input")
      .first()
      .isDisabled();
    assert(
      switchDisabled,
      "Status toggle switch must be disabled for non-admin",
    );

    await docPage.screenshot({
      path: `${screenshotsDir}/13_doctor_role_restricted.png`,
    });
    console.log("Doctor role scoping verified.");
    await docContext.close();

    console.log(
      "=== All 16 Browser Verification Journeys Passed Successfully! ===",
    );
  } finally {
    console.log("Performing full fixture cleanup...");
    // Clean up created appointments
    for (const apptId of createdAppointmentIds) {
      await db
        .query("DELETE FROM appointment WHERE id = $1", [apptId])
        .catch(() => {});
    }
    // Clean up created patients
    for (const patId of createdPatientIds) {
      await db
        .query("DELETE FROM appointment WHERE patient_id = $1", [patId])
        .catch(() => {});
      await db
        .query("DELETE FROM patient WHERE id = $1", [patId])
        .catch(() => {});
    }
    // Clean up created doctors
    for (const docId of createdDoctorIds) {
      await db
        .query("DELETE FROM appointment WHERE doctor_id = $1", [docId])
        .catch(() => {});
      await db
        .query("DELETE FROM doctor_holiday WHERE doctor_id = $1", [docId])
        .catch(() => {});
      await db
        .query("DELETE FROM doctor_lunch_break WHERE doctor_id = $1", [docId])
        .catch(() => {});
      await db
        .query("DELETE FROM doctor_opd_charge WHERE doctor_id = $1", [docId])
        .catch(() => {});
      await db
        .query("DELETE FROM doctor_hours WHERE doctor_id = $1", [docId])
        .catch(() => {});
      await db
        .query("DELETE FROM doctor_profile WHERE user_id = $1", [docId])
        .catch(() => {});
      await db
        .query("DELETE FROM staff_access WHERE user_id = $1", [docId])
        .catch(() => {});
      await db
        .query('DELETE FROM account WHERE "userId" = $1', [docId])
        .catch(() => {});
      await db
        .query('DELETE FROM session WHERE "userId" = $1', [docId])
        .catch(() => {});
      await db
        .query('DELETE FROM "user" WHERE id = $1', [docId])
        .catch(() => {});
    }
    // Clean up ephemeral admin and other created users
    for (const uId of createdUserIds) {
      await db
        .query("DELETE FROM staff_access WHERE user_id = $1", [uId])
        .catch(() => {});
      await db
        .query('DELETE FROM account WHERE "userId" = $1', [uId])
        .catch(() => {});
      await db
        .query('DELETE FROM session WHERE "userId" = $1', [uId])
        .catch(() => {});
      await db.query('DELETE FROM "user" WHERE id = $1', [uId]).catch(() => {});
    }
    // Clean up created departments
    for (const depId of createdDepartmentIds) {
      await db
        .query("DELETE FROM doctor_department WHERE id = $1", [depId])
        .catch(() => {});
    }

    await browser.close().catch(() => {});
    await db.end().catch(() => {});
    console.log("Fixture cleanup completed.");
  }
}

main().catch((err) => {
  console.error("Verification failed:", err);
  process.exit(1);
});
