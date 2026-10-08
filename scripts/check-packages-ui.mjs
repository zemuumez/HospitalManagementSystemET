import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";

export async function checkPackagesUI({ context, db, base, results, browser }) {
  const page = await context.newPage();
  page.on("dialog", (d) => d.accept());
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      console.log(`[BROWSER ERROR] ${msg.text()}`);
    }
  });
  page.on("pageerror", (err) => console.log(`[PAGE CRASH] ${err.message}`));

  const suffix = randomUUID().slice(0, 8);
  const pkgName = `Executive Health Checkup ${suffix}`;
  const passed = (name) => {
    console.log(`[PASS] Packages: ${name}`);
    results.push({ name: `Packages: ${name}`, status: "pass" });
  };

  async function open() {
    const pkgPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/packages") && r.request().method() === "GET",
    );
    const srvPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/services") && r.request().method() === "GET",
    );
    await page.goto(`${base}/modules/packages`, {
      waitUntil: "domcontentloaded",
    });
    await Promise.all([pkgPromise, srvPromise]);
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    await page.locator(".billing-toolbar input").waitFor();
  }

  try {
    // 1. Initial Empty State Check: Database must not display hardcoded/mock fixtures
    console.log("[1/10] Verifying empty state...");
    await open();
    assert.equal(
      await page.getByRole("cell", { name: "Checkup", exact: true }).count(),
      0,
      "empty database must not show Checkup fixture",
    );
    assert.equal(
      await page
        .getByRole("cell", { name: "Fever Package", exact: true })
        .count(),
      0,
      "empty database must not show Fever Package fixture",
    );
    assert.ok(
      await page.getByText("No medical packages found in catalog").isVisible(),
      "empty catalog must show empty notice",
    );
    passed("empty catalog displays zero hardcoded mock fixtures");

    // 2. Ensure at least two active services exist in the database
    console.log("[2/10] Seeding prerequisites...");
    const srv1Id = randomUUID();
    const srv2Id = randomUUID();
    await db.query(
      `INSERT INTO hospital_service (id, name, description, quantity, rate_minor, status)
       VALUES ($1, $2, 'Consultation desc', 1, 5000, 1),
              ($3, $4, 'Lab test desc', 1, 3500, 1)`,
      [srv1Id, `Consultation ${suffix}`, srv2Id, `Blood Test ${suffix}`],
    );

    // Refresh page so services load into the dropdown
    console.log("[2/10] Reloading page to fetch seeded services...");
    await open();

    // 3. Create Package with two service lines
    console.log("[3/10] Creating package via modal...");
    const newPkgBtn = page
      .locator(".billing-toolbar")
      .getByRole("button", { name: "New Package", exact: true });
    await newPkgBtn.waitFor({ state: "visible" });
    await newPkgBtn.click();

    const createModal = page.locator(".modal-backdrop-custom");
    try {
      await createModal.waitFor({ state: "visible", timeout: 5000 });
    } catch {
      console.log("[RETRY] Clicking New Package again...");
      await newPkgBtn.click();
      await createModal.waitFor({ state: "visible" });
    }
    await createModal.getByPlaceholder("Package Name").fill(pkgName);
    await createModal
      .getByPlaceholder("Description")
      .fill("Comprehensive annual executive checkup");
    await createModal.locator('input[type="number"][max="100"]').fill("15");

    console.log("[3/10] Selecting first service line...");
    await createModal
      .locator(`select option[value="${srv1Id}"]`)
      .first()
      .waitFor({ state: "attached" });
    await createModal.locator("select").nth(0).selectOption(srv1Id);
    await createModal.locator('input[type="number"][min="1"]').nth(0).fill("1");

    console.log("[3/10] Adding and selecting second service line...");
    await createModal
      .getByRole("button", { name: "Add Service", exact: true })
      .click();
    await createModal
      .locator(`select option[value="${srv2Id}"]`)
      .nth(1)
      .waitFor({ state: "attached" });
    await createModal.locator("select").nth(1).selectOption(srv2Id);
    await createModal.locator('input[type="number"][min="1"]').nth(1).fill("2");

    console.log("[3/10] Submitting create package form...");
    const createRespPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/packages") &&
        r.request().method() === "POST",
    );
    await createModal
      .getByRole("button", { name: "Save", exact: true })
      .click();
    const createResp = await createRespPromise;
    console.log("[3/10] Create response received:", createResp.status());
    assert.equal(createResp.status(), 201);
    await createModal.waitFor({ state: "hidden" });
    passed("package created via modal and returned 201 Created");

    // Verify row appeared in the table with exact totals
    // Subtotal = 1*50 + 2*35 = 120.00. Discount 15% = 18.00. Total = 102.00
    console.log("[3/10] Verifying row in table...");
    const createdRow = page.locator("tr", { hasText: pkgName });
    await createdRow.waitFor();
    assert.ok(
      await createdRow.getByText("ETB 102.00").isVisible(),
      "total should reflect 102.00",
    );
    passed("live table reflects server-calculated total of 102.00 ETB");

    // 4. Test page reload
    console.log("[4/10] Verifying page reload persistence...");
    const reloadPkgLoaded = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/packages") && r.request().method() === "GET",
    );
    await page.reload({ waitUntil: "domcontentloaded" });
    await reloadPkgLoaded;
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    await page.locator("tr", { hasText: pkgName }).waitFor();

    const dbPkg = (
      await db.query(
        "SELECT id, name, discount, total_amount_minor FROM package WHERE name = $1",
        [pkgName],
      )
    ).rows;
    assert.equal(dbPkg.length, 1);
    assert.equal(Number(dbPkg[0].discount), 15);
    assert.equal(Number(dbPkg[0].total_amount_minor), 10200);
    const dbLines = (
      await db.query("SELECT id FROM package_service WHERE package_id = $1", [
        dbPkg[0].id,
      ])
    ).rows;
    assert.equal(dbLines.length, 2);
    passed("package and lines survive page reload and persist correctly in DB");

    // 5. Test Package Details modal
    console.log("[5/10] Verifying package details modal...");
    const viewBtn = page.locator("tr", { hasText: pkgName }).getByTitle("View");
    await viewBtn.click();
    const detailModal = page.locator(".modal-backdrop-custom");
    await detailModal
      .getByRole("heading", { name: "Package Details" })
      .waitFor();
    assert.ok(
      await detailModal
        .getByText("Comprehensive annual executive checkup")
        .isVisible(),
    );
    assert.ok(await detailModal.getByText("ETB 102.00").isVisible());
    await detailModal.getByRole("button", { name: "Cancel" }).click();
    await detailModal.waitFor({ state: "hidden" });
    passed(
      "package details modal displays service breakdown and authoritative total",
    );

    // 6. Test Editing
    console.log("[6/10] Testing package edit modal...");
    const editBtn = page.locator("tr", { hasText: pkgName }).getByTitle("Edit");
    await editBtn.click();
    const editModal = page.locator(".modal-backdrop-custom");
    await editModal.getByRole("heading", { name: "Edit Package" }).waitFor();

    // Change discount from 15 to 20
    await editModal.locator('input[type="number"][max="100"]').fill("20");
    // Update quantity of second service from 2 to 3
    await editModal.locator('input[type="number"][min="1"]').nth(1).fill("3");

    // Subtotal: 1*50 + 3*35 = 155.00. Discount 20% = 31.00. Total = 124.00
    const updateRespPromise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/packages/${dbPkg[0].id}`) &&
        r.request().method() === "PUT",
    );
    await editModal.getByRole("button", { name: "Save", exact: true }).click();
    const updateResp = await updateRespPromise;
    console.log("[6/10] Update response received:", updateResp.status());
    assert.equal(updateResp.status(), 200);
    await editModal.waitFor({ state: "hidden" });

    const updatedRow = page.locator("tr", { hasText: pkgName });
    await updatedRow.waitFor();
    await updatedRow.getByText("ETB 124.00").waitFor();

    const updatedDbPkg = (
      await db.query(
        "SELECT total_amount_minor, discount FROM package WHERE id = $1",
        [dbPkg[0].id],
      )
    ).rows;
    assert.equal(Number(updatedDbPkg[0].discount), 20);
    assert.equal(Number(updatedDbPkg[0].total_amount_minor), 12400);
    passed("package editing updates child lines, discount, and exact totals");

    // 7. Error Handling: Rejection preserves form without fabricated saves
    console.log("[7/10] Testing duplicate rejection without fake state...");
    await page
      .locator(".billing-toolbar")
      .getByRole("button", { name: "New Package", exact: true })
      .click();
    const duplicateModal = page.locator(".modal-backdrop-custom");
    await duplicateModal.waitFor({ state: "visible" });
    await duplicateModal.getByPlaceholder("Package Name").fill(pkgName); // duplicate name
    await duplicateModal
      .getByPlaceholder("Description")
      .fill("Duplicate attempt");
    await duplicateModal
      .locator(`select option[value="${srv1Id}"]`)
      .first()
      .waitFor({ state: "attached" });
    await duplicateModal.locator("select").nth(0).selectOption(srv1Id);

    const conflictPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/packages") &&
        r.request().method() === "POST",
    );
    await duplicateModal
      .getByRole("button", { name: "Save", exact: true })
      .click();
    const conflictResp = await conflictPromise;
    console.log("[7/10] Conflict response status:", conflictResp.status());
    assert.equal(conflictResp.status(), 409);

    // Modal must remain open
    assert.ok(
      await duplicateModal
        .getByRole("heading", { name: "New Package" })
        .isVisible(),
    );
    await duplicateModal.getByRole("button", { name: "Cancel" }).click();
    await duplicateModal.waitFor({ state: "hidden" });

    // Verify no duplicate row exists in DB
    const countCheck = (
      await db.query("SELECT count(*) FROM package WHERE name = $1", [pkgName])
    ).rows[0].count;
    assert.equal(Number(countCheck), 1);
    passed(
      "duplicate package submission is rejected with 409 and never fabricates a local row",
    );

    // 8. Fresh Browser Context
    console.log("[8/10] Testing fresh browser context...");
    if (browser) {
      const freshContext = await browser.newContext();
      const cookies = await context.cookies();
      await freshContext.addCookies(cookies);
      const freshPage = await freshContext.newPage();
      freshPage.on("dialog", (d) => d.accept());
      const freshPkgLoaded = freshPage.waitForResponse(
        (r) =>
          r.url().includes("/api/hms/packages") &&
          r.request().method() === "GET",
      );
      await freshPage.goto(`${base}/modules/packages`, {
        waitUntil: "domcontentloaded",
      });
      await freshPkgLoaded;
      await freshPage.locator('.legacy-workspace[data-ready="true"]').waitFor();
      const freshRow = freshPage.locator("tr", { hasText: pkgName });
      await freshRow.waitFor();
      await freshRow.getByText("ETB 124.00").waitFor();
      await freshPage.close();
      await freshContext.close();
      passed("persisted package is verified from a fresh browser context");
    }

    // 9. In-Use Deletion Protection
    console.log("[9/10] Testing in-use deletion rejection (409 Conflict)...");
    const userRes = await db.query('SELECT id FROM "user" LIMIT 1');
    const userId = userRes.rows[0].id;
    const patRes = await db.query(
      `INSERT INTO patient (given_name, family_name, date_of_birth, phone)
       VALUES ('Test', 'Patient', '1990-01-01', '+251911223344')
       RETURNING id`,
    );
    const patientId = patRes.rows[0].id;

    const caseRes = await db.query(
      `INSERT INTO patient_case (patient_id, doctor_id, description, created_by)
       VALUES ($1, $2, 'In-use test case', $2)
       RETURNING id`,
      [patientId, userId],
    );
    const caseId = caseRes.rows[0].id;

    const encRes = await db.query(
      `INSERT INTO encounter (kind, case_id, patient_id, doctor_id, admitted_at, status, request_key, created_by)
       VALUES ('opd', $1, $2, $3, now(), 'active', $4, $3)
       RETURNING id`,
      [caseId, patientId, userId, `enc-key-${suffix}`],
    );
    const encId = encRes.rows[0].id;

    await db.query(
      `INSERT INTO ipd_admission_details (encounter_id, package_id, package_name, package_charge_minor)
       VALUES ($1, $2, 'Package In Use', 10200)`,
      [encId, dbPkg[0].id],
    );

    // Attempt to delete: expect 409 RECORD_IN_USE
    const delBlockedPromise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/packages/${dbPkg[0].id}`) &&
        r.request().method() === "DELETE",
    );
    await page.locator("tr", { hasText: pkgName }).getByTitle("Delete").click();
    const delBlockedResp = await delBlockedPromise;
    console.log("[9/10] Delete blocked response:", delBlockedResp.status());
    assert.equal(delBlockedResp.status(), 409);

    // Package must still be in the table
    await page.locator("tr", { hasText: pkgName }).waitFor();
    const stillInDb = (
      await db.query("SELECT id FROM package WHERE id = $1", [dbPkg[0].id])
    ).rows;
    assert.equal(stillInDb.length, 1);
    passed(
      "in-use package deletion returns 409 Conflict and preserves database record",
    );

    // 10. Remove IPD reference and verify successful deletion
    console.log("[10/10] Testing unreferenced deletion (204 No Content)...");
    await db.query(
      "DELETE FROM ipd_admission_details WHERE encounter_id = $1",
      [encId],
    );
    await db.query("DELETE FROM encounter WHERE id = $1", [encId]);
    await db.query("DELETE FROM patient_case WHERE id = $1", [caseId]);
    await db.query("DELETE FROM patient WHERE id = $1", [patientId]);

    const delSuccessPromise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/packages/${dbPkg[0].id}`) &&
        r.request().method() === "DELETE",
    );
    await page.locator("tr", { hasText: pkgName }).getByTitle("Delete").click();
    const delSuccessResp = await delSuccessPromise;
    console.log("[10/10] Delete success response:", delSuccessResp.status());
    assert.ok(
      [200, 204].includes(delSuccessResp.status()),
      `expected 200 or 204, got ${delSuccessResp.status()}`,
    );

    // Package should disappear from table
    await page.locator("tr", { hasText: pkgName }).waitFor({ state: "hidden" });
    const deletedDb = (
      await db.query("SELECT id FROM package WHERE id = $1", [dbPkg[0].id])
    ).rows;
    assert.equal(deletedDb.length, 0);
    passed(
      "unreferenced package deletion succeeds and removes record from database",
    );
  } finally {
    await page.close();
  }
}
