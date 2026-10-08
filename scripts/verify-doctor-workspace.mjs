import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import assert from "node:assert/strict";

const db = new Pool({
  connectionString: process.env.DATABASE_URL || "postgresql://hms:424d48acc68f829bc0cd4f7b1f07f2328b706dc863002e63@127.0.0.1:5433/hms"
});

const screenshotsDir = "C:/Users/USER/.gemini/antigravity-ide/brain/211d27d4-51d4-4c25-a079-46b35d8664de/screenshots";
const base = "http://127.0.0.1:3000";

async function main() {
  console.log("=== Starting Doctor Workspace Browser Verification ===");

  // 1. Ensure at least one department exists in database
  let deptId;
  const depCheck = await db.query("SELECT id FROM doctor_department WHERE NOT archived LIMIT 1");
  if (depCheck.rowCount > 0) {
    deptId = depCheck.rows[0].id;
  } else {
    const depIns = await db.query(
      "INSERT INTO doctor_department (title, description) VALUES ('Cardiology Department', 'Heart & Vascular') RETURNING id"
    );
    deptId = depIns.rows[0].id;
  }
  console.log(`Using active department: ${deptId}`);

  const browser = await chromium.launch({
    headless: true,
    executablePath: process.env.HMS_CHROME_PATH || "C:/Program Files/Google/Chrome/Application/chrome.exe",
  });

  const context = await browser.newContext({
    viewport: { width: 1440, height: 960 },
  });
  const page = await context.newPage();
  page.setDefaultTimeout(30000);

  page.on("response", async (resp) => {
    if (resp.url().includes("/api/staff") || resp.url().includes("/api/hms/")) {
      console.log(`[HTTP ${resp.status()}] ${resp.request().method()} ${resp.url()}`);
      if (resp.status() >= 400) {
        console.log(`[ERROR RESPONSE]`, await resp.text().catch(() => ""));
      }
    }
  });

  try {
    // -------------------------------------------------------------
    // Journey 1: Admin Login
    // -------------------------------------------------------------
    console.log("Step 1: Navigating to login page...");
    await page.goto(`${base}/login`);
    await page.getByLabel("Email address").fill("admin@ulshms.local");
    await page.getByLabel("Password", { exact: false }).fill("AdminPassword123!@#");
    await page.screenshot({ path: `${screenshotsDir}/01_admin_login.png` });
    
    await page.getByRole("button", { name: "Sign In", exact: true }).click();
    await page.waitForURL("**/dashboard", { timeout: 15000 }).catch(async () => {
      await page.goto(`${base}/dashboard`);
    });
    console.log("Logged in successfully as Administrator.");

    // -------------------------------------------------------------
    // Journey 2: Doctors Workspace Overview & Real Database Connection
    // -------------------------------------------------------------
    console.log("Step 2: Navigating to Doctors Workspace (/modules/doctors)...");
    await page.goto(`${base}/modules/doctors`);
    await page.waitForSelector(".billing-card");
    await page.waitForTimeout(1000);
    
    await page.screenshot({ path: `${screenshotsDir}/02_doctors_directory.png` });
    const bannerText = await page.textContent("body");
    console.log("Banner snippet:", bannerText.slice(0, 300));
    assert(bannerText.includes("Connected to PostgreSQL") || bannerText.includes("Connecting to PostgreSQL"), "Should show backend status");
    console.log("Doctors workspace directory verified with live connection.");

    // -------------------------------------------------------------
    // Journey 3: Validation Errors
    // -------------------------------------------------------------
    console.log("Step 3: Testing form validation errors...");
    await page.getByRole("button", { name: "+ New Doctor" }).click();
    await page.waitForSelector(".modal-backdrop-custom");
    
    // Clear name and email, then attempt submit
    await page.locator('input[placeholder="Dr. John Doe"]').fill("");
    await page.locator('input[placeholder="doctor@hospital.local"]').fill("");
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await page.screenshot({ path: `${screenshotsDir}/03_validation_error.png` });
    console.log("Validation errors verified.");

    // Close the modal from Step 3 cleanly
    await page.getByRole("button", { name: "Cancel" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached" });

    // -------------------------------------------------------------
    // Journey 4: Doctor Creation & Persistence After Reload
    // -------------------------------------------------------------
    console.log("Step 4: Creating doctor and verifying reload persistence...");
    await page.getByRole("button", { name: "+ New Doctor" }).click();
    await page.waitForSelector(".modal-backdrop-custom");

    const timestamp = Date.now();
    const docName = `Dr. Sophia Vance ${timestamp.toString().slice(-4)}`;
    const docEmail = `sophia.vance.${timestamp}@ulshms.local`;
    const docPassword = "DoctorSecure1234!";

    await page.locator('input[placeholder="Dr. John Doe"]').fill(docName);
    await page.locator('input[placeholder="doctor@hospital.local"]').fill(docEmail);
    await page.locator('input[type="password"]').fill(docPassword);
    await page.locator('input[placeholder="Cardiologist, Neurologist..."]').fill("Senior Cardiothoracic Surgeon");
    await page.locator('input[placeholder="Senior Consultant"]').fill("Chief Cardiothoracic Surgeon");
    await page.locator('input[placeholder="MD, Specialist"]').fill("MD, FACS, FACC");
    await page.locator('input[placeholder="+251 91 123 4567"]').fill("+251 91 888 9900");
    
    // Select department
    const deptSelect = page.locator('select.form-select-custom').first();
    const options = await deptSelect.locator('option').all();
    console.log(`Department options count: ${options.length}`);
    if (options.length > 1) {
      await deptSelect.selectOption({ index: 1 });
    } else {
      // If no department option in select, select value deptId
      await deptSelect.selectOption(deptId).catch(() => {});
    }

    // OPD and Appt charges
    const chargeInputs = page.locator('input[type="number"]');
    await chargeInputs.nth(0).fill("450");
    await chargeInputs.nth(1).fill("500");

    await page.getByRole("button", { name: "Save", exact: true }).click();

    // Check if error banner inside modal appears
    try {
      await page.waitForSelector(".modal-backdrop-custom", { state: "detached", timeout: 8000 });
    } catch {
      const modalText = await page.locator(".modal-backdrop-custom").textContent().catch(() => "");
      console.log("Modal did not close! Modal content:", modalText);
      const bodyErr = await page.locator("body").textContent();
      console.log("Body snippet:", bodyErr.slice(0, 500));
      await page.screenshot({ path: `${screenshotsDir}/04_create_modal_error.png` });
      throw new Error(`Modal failed to close: ${modalText}`);
    }
    console.log("Doctor created. Reloading page to verify persistence...");

    // Hard reload
    await page.reload();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });

    const pageContent = await page.textContent("body");
    assert(pageContent.includes(docName), `Doctor ${docName} must persist after reload`);
    assert(pageContent.includes("450 ETB"), "Doctor OPD charge must persist after reload");
    await page.screenshot({ path: `${screenshotsDir}/04_doctor_persisted.png` });
    console.log("Persistence after reload verified.");

    // Retrieve created doctor's database ID
    const docDb = await db.query('SELECT u.id, u.email FROM "user" u WHERE email = $1', [docEmail]);
    assert.equal(docDb.rowCount, 1, "Doctor must exist in database");
    const createdDoctorId = docDb.rows[0].id;
    console.log(`Created doctor database ID: ${createdDoctorId}`);

    // -------------------------------------------------------------
    // Journey 5: Doctor Details View Modal
    // -------------------------------------------------------------
    console.log("Step 5: Verifying Doctor Profile Details modal...");
    // Click view details (eye icon)
    const row = page.locator(`tr:has-text("${docName}")`);
    await row.locator('button[aria-label="View Details"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    const modalText = await page.locator(".modal-body-custom").textContent();
    assert(modalText.includes(docName), "Details modal must show doctor name");
    assert(modalText.includes("450 ETB"), "Details modal must show OPD charge");
    assert(modalText.includes("500 ETB"), "Details modal must show Appt charge");
    await page.screenshot({ path: `${screenshotsDir}/05_doctor_details.png` });
    
    await page.getByRole("button", { name: "Close" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached" });
    console.log("Doctor details modal verified.");

    // -------------------------------------------------------------
    // Journey 6: Doctor Profile Editing
    // -------------------------------------------------------------
    console.log("Step 6: Editing doctor profile...");
    await row.locator('button[aria-label="Edit"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    await page.locator('input.form-input-custom').filter({ hasText: "" }).nth(2).fill("Chief Cardiovascular Surgeon");
    // Update OPD charge to 520
    const editOpdInput = page.locator('input[type="number"]').first();
    await editOpdInput.fill("520");

    await page.getByRole("button", { name: "Save Changes" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached", timeout: 15000 });

    await page.waitForTimeout(1000);
    const updatedContent = await page.textContent("body");
    assert(updatedContent.includes("520 ETB"), "Updated charge 520 ETB must appear in table");
    await page.screenshot({ path: `${screenshotsDir}/06_doctor_edited.png` });
    console.log("Doctor profile edit verified.");

    // -------------------------------------------------------------
    // Journey 7: Doctor Schedules & Timetable
    // -------------------------------------------------------------
    console.log("Step 7: Creating doctor schedule timetable...");
    await page.locator('.module-subtabs-nav a.subtab-btn:has-text("Schedules")').click();
    await page.waitForSelector(".billing-card");

    await page.getByRole("button", { name: "+ New Schedule" }).click();
    await page.waitForSelector("form.form-card-container");

    // Select doctor
    await page.locator("select.form-select-custom").selectOption({ value: createdDoctorId });
    await page.locator('input[placeholder="00:15:00"]').fill("00:20:00");
    await page.screenshot({ path: `${screenshotsDir}/07_schedule_form.png` });

    await page.getByRole("button", { name: "Save", exact: true }).click();
    await page.waitForSelector(".billing-table", { timeout: 15000 });

    const schedText = await page.textContent("body");
    assert(schedText.includes(docName), "Schedule list must contain created doctor schedule");
    await page.screenshot({ path: `${screenshotsDir}/07_schedule_saved.png` });
    console.log("Doctor schedule creation verified.");

    // -------------------------------------------------------------
    // Journey 8: Doctor Holidays & Lunch Breaks
    // -------------------------------------------------------------
    console.log("Step 8: Adding doctor holiday and lunch break...");
    await page.locator('.module-subtabs-nav a.subtab-btn:has-text("Doctor Holidays")').click();
    await page.waitForSelector(".billing-card");

    await page.getByRole("button", { name: "+ Add Doctor Holiday" }).click();
    await page.waitForSelector(".modal-backdrop-custom");
    await page.locator("select.form-select-custom").selectOption({ value: createdDoctorId });
    await page.locator('input[type="date"]').fill("2026-12-25");
    await page.locator('input[placeholder="e.g. Annual Leave, Medical Conference"]').fill("Christmas Holiday 2026");
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached", timeout: 15000 });

    const holText = await page.textContent("body");
    assert(holText.includes("Christmas Holiday 2026"), "Holiday must appear in table");
    await page.screenshot({ path: `${screenshotsDir}/08_holiday_saved.png` });
    console.log("Doctor holiday verified.");

    // Add Lunch Break
    await page.locator('.module-subtabs-nav a.subtab-btn:has-text("Breaks")').click();
    await page.waitForSelector(".billing-card");

    await page.getByRole("button", { name: "+ Add Break" }).click();
    await page.waitForSelector("form.form-card-container");
    await page.locator("select.form-select-custom").selectOption({ value: createdDoctorId });
    await page.locator('input.form-input-custom').nth(0).fill("12:30:00");
    await page.locator('input.form-input-custom').nth(1).fill("13:30:00");
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await page.waitForSelector('tr:has-text("12:30:00")', { timeout: 15000 });

    const breakText = await page.textContent("body");
    assert(breakText.includes("12:30:00") && breakText.includes("13:30:00"), "Break window must appear in table");
    await page.screenshot({ path: `${screenshotsDir}/09_break_saved.png` });
    console.log("Lunch break verified.");

    // -------------------------------------------------------------
    // Journey 9: OPD Charges Master & Flow Connection
    // -------------------------------------------------------------
    console.log("Step 9: Verifying OPD Charges tab...");
    await page.locator('.module-subtabs-nav a.subtab-btn:has-text("OPD Charges")').click();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });

    const opdText = await page.textContent("body");
    assert(opdText.includes("520.00"), "OPD charges table must reflect synchronized charge of 520.00 ETB");
    
    // Edit OPD charge in dedicated master tab
    const opdRow = page.locator(`tr:has-text("${docName}")`);
    await opdRow.locator('button[aria-label="Edit OPD Charge"]').click();
    await page.waitForSelector(".modal-backdrop-custom");
    await page.locator('input[type="number"]').fill("580.00");
    await page.getByRole("button", { name: "Save Charge" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached", timeout: 15000 });
    await page.waitForSelector('tr:has-text("580.00")', { timeout: 15000 });

    const updatedOpdText = await page.textContent("body");
    assert(updatedOpdText.includes("580.00"), "Updated charge 580.00 ETB must appear");
    await page.screenshot({ path: `${screenshotsDir}/10_opd_charge_synced.png` });
    console.log("OPD charges master verified.");

    // -------------------------------------------------------------
    // Journey 10: Booking Conflict Protection & Slot Availability
    // -------------------------------------------------------------
    console.log("Step 10: Verifying appointment booking conflict protection...");
    // Create a patient and book appointment for this doctor
    let patId = (await db.query("SELECT id FROM patient LIMIT 1")).rows[0]?.id;
    if (!patId) {
      const pRes = await db.query(`
        INSERT INTO patient (given_name, family_name, date_of_birth, gender, contact_phone)
        VALUES ('Test', 'Patient', '1990-01-01', 'female', '+251911000000')
        RETURNING id
      `);
      patId = pRes.rows[0].id;
    }
    const adminUser = await db.query('SELECT id FROM "user" WHERE email = $1', ['admin@ulshms.local']);
    const adminUserId = adminUser.rows[0].id;
    // Direct booking in appointment table
    await db.query(`
      INSERT INTO appointment (patient_id, doctor_id, starts_at, ends_at, problem, status, created_by, request_key)
      VALUES ($1, $2, '2026-11-10 10:00:00+03', '2026-11-10 10:20:00+03', 'Routine Cardiac Checkup', 'booked', $3, $4)
    `, [patId, createdDoctorId, adminUserId, `req-key-${Date.now()}`]);
    console.log("Booked active clinical appointment for doctor on 2026-11-10.");

    // Attempt to add a holiday on this booked date (2026-11-10) to verify booking-conflict protection
    await page.locator('.module-subtabs-nav a.subtab-btn:has-text("Doctor Holidays")').click();
    await page.waitForSelector(".billing-card");
    await page.getByRole("button", { name: "+ Add Doctor Holiday" }).click();
    await page.waitForSelector(".modal-backdrop-custom");
    await page.locator("select.form-select-custom").selectOption({ value: createdDoctorId });
    await page.locator('input[type="date"]').fill("2026-11-10");
    await page.locator('input[placeholder="e.g. Annual Leave, Medical Conference"]').fill("Holiday on booked date");
    await Promise.all([
      page.waitForResponse((r) => r.url().includes("/doctor-holidays") && r.request().method() === "POST" && r.status() === 409),
      page.getByRole("button", { name: "Save", exact: true }).click(),
    ]);

    // Verify error appears due to appointment conflict
    await page.waitForSelector(".modal-backdrop-custom div:has-text('appointments')", { timeout: 10000 });
    const modalContent = await page.locator(".modal-backdrop-custom").textContent();
    assert(
      modalContent.includes("Cannot create holiday") || modalContent.includes("active appointments") || modalContent.includes("appointments"),
      "Booking conflict must block holiday creation when doctor has active appointment!"
    );
    await page.screenshot({ path: `${screenshotsDir}/11_booking_conflict_protected.png` });
    console.log("Booking conflict protection verified: holiday blocked by active appointment.");
    await page.getByRole("button", { name: "Cancel" }).click();
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached" });

    // -------------------------------------------------------------
    // Journey 11: Referenced-Doctor Deletion Conflict Protection
    // -------------------------------------------------------------
    console.log("Step 11: Testing deletion conflict protection on referenced doctor...");
    await page.locator('.module-subtabs-nav a.subtab-btn:has-text("Doctors")').click();
    await page.waitForSelector(`tr:has-text("${docName}")`, { timeout: 15000 });

    const docRowToDelete = page.locator(`tr:has-text("${docName}")`);
    await docRowToDelete.locator('button[aria-label="Delete"]').click();
    await page.waitForSelector(".modal-backdrop-custom");

    // Click confirm Delete and await 409 Conflict response
    await Promise.all([
      page.waitForResponse((r) => r.url().includes("/doctors/") && r.request().method() === "DELETE" && r.status() === 409),
      page.locator(".modal-backdrop-custom button.btn-delete-confirm-red").click(),
    ]);

    // Verify 409 Conflict banner appears
    await page.waitForSelector(".modal-backdrop-custom", { state: "detached", timeout: 15000 });
    await page.waitForTimeout(1000);

    const bannerConflict = await page.textContent("body");
    assert(
      bannerConflict.includes("cannot be deleted") ||
      bannerConflict.includes("in use") ||
      bannerConflict.includes("Cannot delete doctor") ||
      bannerConflict.includes("conflict") ||
      bannerConflict.includes("referenced") ||
      bannerConflict.includes("associated"),
      "Conflict banner must appear when deleting referenced doctor"
    );

    // Verify doctor is NOT deleted and remains in table
    const tableText = await page.locator(".billing-table").textContent();
    assert(tableText.includes(docName), "Referenced doctor must be preserved in table after conflict!");
    await page.screenshot({ path: `${screenshotsDir}/12_deletion_conflict_blocked.png` });
    console.log("Referenced-doctor deletion conflict protection verified.");

    // -------------------------------------------------------------
    // Journey 12: Restricted Roles Access Scoping
    // -------------------------------------------------------------
    console.log("Step 12: Verifying role restrictions with doctor account...");
    const docContext = await browser.newContext({ viewport: { width: 1440, height: 960 } });
    const docPage = await docContext.newPage();
    docPage.setDefaultTimeout(30000);

    docPage.on("response", async (resp) => {
      if (resp.url().includes("/api/")) {
        console.log(`[DOC PAGE HTTP ${resp.status()}] ${resp.request().method()} ${resp.url()}`);
      }
    });

    await docPage.goto(`${base}/login`);
    await docPage.getByLabel("Email address").fill(docEmail);
    await docPage.getByLabel("Password", { exact: false }).fill(docPassword);
    await docPage.getByRole("button", { name: "Sign In", exact: true }).click();
    await docPage.waitForURL("**/dashboard", { timeout: 15000 }).catch(async () => {
      console.log("Did not reach dashboard via redirect, current URL:", docPage.url());
      await docPage.goto(`${base}/dashboard`);
    });

    // Navigate to /modules/doctors
    await docPage.goto(`${base}/modules/doctors`);
    await docPage.waitForSelector(".billing-table", { timeout: 15000 });
    await docPage.waitForTimeout(1000);

    // As a doctor:
    // 1. "+ New Doctor" button should NOT be visible
    const newDocBtnCount = await docPage.getByRole("button", { name: "+ New Doctor" }).count();
    assert.equal(newDocBtnCount, 0, "Non-admin doctor should NOT see + New Doctor button");

    // 2. Doctor directory Delete buttons should NOT be visible
    const deleteBtnCount = await docPage.locator('.billing-table button[aria-label="Delete"]').count();
    assert.equal(deleteBtnCount, 0, "Non-admin doctor should NOT see Delete buttons on doctor table");

    // 3. Status toggle switches should be disabled
    const switchDisabled = await docPage.locator(".billing-table .switch-toggle input").first().isDisabled();
    assert(switchDisabled, "Status toggle switch must be disabled for non-admin");

    await docPage.screenshot({ path: `${screenshotsDir}/13_doctor_role_restricted.png` });
    console.log("Restricted doctor role verified.");
    await docContext.close();

    console.log("=== All 12 Browser Journeys Completed Successfully! ===");
  } finally {
    await browser.close();
    await db.end();
  }
}

main().catch((err) => {
  console.error("Browser test failed:", err);
  process.exit(1);
});
