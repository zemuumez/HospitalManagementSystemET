import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";

export async function checkInsuranceUI({
  context,
  db,
  base,
  results,
  browser,
}) {
  const page = await context.newPage();
  page.on("dialog", (d) => d.accept());
  page.on("console", (msg) => {
    console.log(`[BROWSER LOG ${msg.type()}] ${msg.text()}`);
  });
  page.on("pageerror", (err) => console.log(`[PAGE CRASH] ${err.message}`));
  page.on("request", (req) => {
    if (req.url().includes("/api/hms/insurances")) {
      console.log(`[BROWSER REQ] ${req.method()} ${req.url()}`);
    }
  });
  page.on("response", (res) => {
    if (res.url().includes("/api/hms/insurances")) {
      console.log(`[BROWSER RES] ${res.status()} ${res.url()}`);
    }
  });

  const suffix = randomUUID().slice(0, 8);
  const insName = `Medicaid Plus ${suffix}`;
  const passed = (name) => {
    console.log(`[PASS] Insurances: ${name}`);
    results.push({ name: `Insurances: ${name}`, status: "pass" });
  };

  async function open() {
    const insPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        r.request().method() === "GET",
    );
    await page.goto(`${base}/modules/insurances`, {
      waitUntil: "domcontentloaded",
    });
    await insPromise;
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    await page.locator(".billing-toolbar input").waitFor();
  }

  try {
    // 1. Initial Empty State Check: Database must not display hardcoded/mock fixtures
    console.log("[1/10] Verifying insurances empty state...");
    await open();
    assert.equal(
      await page.getByRole("cell", { name: "BAJAJ", exact: true }).count(),
      0,
      "empty database must not show BAJAJ mock fixture",
    );
    assert.equal(
      await page
        .getByRole("cell", { name: "Zelda Walls", exact: true })
        .count(),
      0,
      "empty database must not show Zelda Walls mock fixture",
    );
    assert.equal(
      await page
        .getByRole("cell", { name: "Brooke Leblan", exact: true })
        .count(),
      0,
      "empty database must not show Brooke Leblan mock fixture",
    );
    assert.ok(
      await page
        .getByText("No insurance policies found in catalog")
        .isVisible(),
      "empty catalog must show empty notice",
    );
    passed("empty catalog displays zero hardcoded mock fixtures");

    // 2. Open full-page New Insurance form
    console.log("[2/10] Opening full-page New Insurance form...");
    const newInsBtn = page
      .locator(".billing-toolbar")
      .getByRole("button", { name: "New Insurance", exact: true });
    await newInsBtn.waitFor({ state: "visible" });
    await newInsBtn.click();

    // Verify create view container is ready
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    await page.getByRole("heading", { name: "New Insurance" }).waitFor();

    // Fill form fields
    await page.getByPlaceholder("Insurance", { exact: true }).fill(insName);
    await page.getByPlaceholder("Service Tax").fill("50.00");
    await page.locator('input[type="number"][max="100"]').fill("10");
    await page.getByPlaceholder("Insurance No").fill(`POL-${suffix}`);
    await page.getByPlaceholder("Insurance Code").fill(`COD-${suffix}`);
    await page.getByPlaceholder("Hospital Rate").fill("150.00");
    await page
      .getByPlaceholder("Remark")
      .fill("Comprehensive corporate coverage");

    // Fill first disease line
    await page
      .getByPlaceholder("Diseases Name")
      .nth(0)
      .fill("Cardiology Consultation");
    await page.getByPlaceholder("Diseases charge").nth(0).fill("200.00");

    // Add and fill second disease line
    await page.getByRole("button", { name: "Add", exact: true }).click();
    await page.getByPlaceholder("Diseases Name").nth(1).fill("ECG Scan");
    await page.getByPlaceholder("Diseases charge").nth(1).fill("100.00");

    // 3. Submit form and verify server calculation
    // Base = 50 + 150 + 200 + 100 = 500.00. Discount 10% = 50.00. Total = 450.00
    console.log("[3/10] Submitting create insurance form...");
    const createRespPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        r.request().method() === "POST",
    );
    await page.getByRole("button", { name: "Save", exact: true }).click();
    const createResp = await createRespPromise;
    console.log("[3/10] Create response received:", createResp.status());
    assert.equal(createResp.status(), 201);

    // Verify return to list view
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    const createdRow = page.locator("tr", { hasText: insName });
    await createdRow.waitFor();
    assert.ok(
      await createdRow.getByText("ETB 450.00").isVisible(),
      "total should reflect 450.00 ETB",
    );
    passed(
      "insurance created with multiple diseases and server-calculated total of 450.00 ETB",
    );

    // 4. Test page reload persistence
    console.log("[4/10] Verifying page reload persistence...");
    const reloadInsLoaded = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        r.request().method() === "GET",
    );
    await page.reload({ waitUntil: "domcontentloaded" });
    await reloadInsLoaded;
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    await page.locator("tr", { hasText: insName }).waitFor();

    const dbIns = (
      await db.query(
        "SELECT id, name, service_tax_minor, hospital_rate_minor, discount, total_minor, status FROM insurance WHERE name = $1",
        [insName],
      )
    ).rows;
    assert.equal(dbIns.length, 1);
    assert.equal(Number(dbIns[0].service_tax_minor), 5000);
    assert.equal(Number(dbIns[0].hospital_rate_minor), 15000);
    assert.equal(Number(dbIns[0].discount), 10);
    assert.equal(Number(dbIns[0].total_minor), 45000);
    assert.equal(Number(dbIns[0].status), 1);

    const dbDiseases = (
      await db.query(
        "SELECT id, disease_name, disease_charge_minor FROM insurance_disease WHERE insurance_id = $1 ORDER BY disease_charge_minor DESC",
        [dbIns[0].id],
      )
    ).rows;
    assert.equal(dbDiseases.length, 2);
    passed(
      "insurance policy and disease lines survive reload and persist in database",
    );

    // 5. Test Insurance Details modal
    console.log("[5/10] Verifying insurance details modal...");
    const viewBtn = page.locator("tr", { hasText: insName }).getByTitle("View");
    await viewBtn.click();
    const detailModal = page.locator(".modal-backdrop-custom");
    await detailModal
      .getByRole("heading", { name: "Insurance Details" })
      .waitFor();
    await detailModal.getByText("Comprehensive corporate coverage").waitFor();
    await detailModal.getByText("ETB 450.00").waitFor();
    await detailModal.getByText("Cardiology Consultation").waitFor();
    await detailModal.getByText("ECG Scan").waitFor();
    await detailModal.getByRole("button", { name: "Cancel" }).click();
    await detailModal.waitFor({ state: "hidden" });
    passed(
      "insurance details modal displays disease breakdown and authoritative totals",
    );

    // 6. Test Status Toggle via PATCH
    console.log("[6/10] Testing status toggle...");
    const statusToggle = page
      .locator("tr", { hasText: insName })
      .locator(".switch-toggle");

    // Toggle to inactive (0)
    const patch1Promise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/insurances/${dbIns[0].id}/status`) &&
        r.request().method() === "PATCH",
    );
    const get1Promise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        r.request().method() === "GET",
    );
    await statusToggle.click();
    const patch1Resp = await patch1Promise;
    assert.equal(patch1Resp.status(), 200);
    await get1Promise;

    const checkStatus0 = (
      await db.query("SELECT status FROM insurance WHERE id = $1", [
        dbIns[0].id,
      ])
    ).rows[0].status;
    assert.equal(Number(checkStatus0), 0);

    // Toggle back to active (1)
    const patch2Promise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/insurances/${dbIns[0].id}/status`) &&
        r.request().method() === "PATCH",
    );
    const get2Promise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        r.request().method() === "GET",
    );
    await statusToggle.click();
    const patch2Resp = await patch2Promise;
    assert.equal(patch2Resp.status(), 200);
    await get2Promise;

    const checkStatus1 = (
      await db.query("SELECT status FROM insurance WHERE id = $1", [
        dbIns[0].id,
      ])
    ).rows[0].status;
    assert.equal(Number(checkStatus1), 1);
    await page.waitForTimeout(300);
    passed(
      "status toggle updates active state atomically via PATCH and persists in database",
    );

    // 7. Test Editing via Modal
    console.log("[7/10] Testing insurance edit modal...");
    const editBtn = page.locator("tr", { hasText: insName }).getByTitle("Edit");
    await editBtn.click();
    const editModal = page.locator(".modal-backdrop-custom");
    await editModal.getByRole("heading", { name: "Edit Insurance" }).waitFor();

    // Update discount to 20%, hospital rate to 250.00
    // Base = 50 (tax) + 250 (rate) + 300 (diseases) = 600.00
    // Discount 20% = 120.00. Total = 480.00
    await editModal.locator('input[type="number"][max="100"]').fill("20");
    await editModal.getByPlaceholder("Hospital Rate").fill("250.00");

    const saveBtn = editModal.getByRole("button", {
      name: "Save",
      exact: true,
    });

    const updateRespPromise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/insurances/${dbIns[0].id}`) &&
        r.request().method() === "PUT",
    );
    await saveBtn.scrollIntoViewIfNeeded();
    await saveBtn.click();
    await editModal
      .locator("form")
      .evaluate((f) => f.requestSubmit())
      .catch(() => {});
    const updateResp = await updateRespPromise;
    console.log("[7/10] Update response received:", updateResp.status());
    assert.equal(updateResp.status(), 200);
    await editModal.waitFor({ state: "hidden" });

    const updatedRow = page.locator("tr", { hasText: insName });
    await updatedRow.waitFor();
    await updatedRow.getByText("ETB 480.00").waitFor();

    const updatedDbIns = (
      await db.query(
        "SELECT total_minor, hospital_rate_minor, discount FROM insurance WHERE id = $1",
        [dbIns[0].id],
      )
    ).rows[0];
    assert.equal(Number(updatedDbIns.discount), 20);
    assert.equal(Number(updatedDbIns.hospital_rate_minor), 25000);
    assert.equal(Number(updatedDbIns.total_minor), 48000);
    passed(
      "insurance editing updates rates, discount, and recalculates authoritative total of 480.00 ETB",
    );

    // 8. Error Handling: Duplicate name rejected with 409
    console.log("[8/10] Testing duplicate rejection without fake state...");
    await page
      .locator(".billing-toolbar")
      .getByRole("button", { name: "New Insurance", exact: true })
      .click();
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();
    await page.getByPlaceholder("Insurance", { exact: true }).fill(insName);
    await page.getByPlaceholder("Service Tax").fill("10.00");
    await page.getByPlaceholder("Insurance No").fill(`POL-DUP-${suffix}`);
    await page.getByPlaceholder("Insurance Code").fill(`COD-DUP-${suffix}`);
    await page.getByPlaceholder("Hospital Rate").fill("50.00");
    await page.getByPlaceholder("Diseases Name").nth(0).fill("Sample Disease");
    await page.getByPlaceholder("Diseases charge").nth(0).fill("50.00");

    const conflictPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        r.request().method() === "POST",
    );
    await page.getByRole("button", { name: "Save", exact: true }).click();
    const conflictResp = await conflictPromise;
    console.log("[8/10] Conflict response status:", conflictResp.status());
    assert.equal(conflictResp.status(), 409);

    // Form must remain open with error banner visible
    assert.ok(
      await page.getByRole("heading", { name: "New Insurance" }).isVisible(),
    );
    await page.getByRole("button", { name: "Cancel" }).click();
    await page.locator('.legacy-workspace[data-ready="true"]').waitFor();

    const countCheck = (
      await db.query("SELECT count(*) FROM insurance WHERE name = $1", [
        insName,
      ])
    ).rows[0].count;
    assert.equal(Number(countCheck), 1);
    passed(
      "duplicate insurance submission is rejected with 409 and never fabricates a local row",
    );

    // 9. In-Use Deletion Protection (409 Conflict)
    console.log("[9/10] Testing in-use deletion rejection (409 Conflict)...");
    const userRes = await db.query('SELECT id FROM "user" LIMIT 1');
    const userId = userRes.rows[0].id;
    const patRes = await db.query(
      `INSERT INTO patient (given_name, family_name, date_of_birth, phone)
       VALUES ('InsTest', 'Patient', '1992-05-10', '+251911998877')
       RETURNING id`,
    );
    const patientId = patRes.rows[0].id;

    const caseRes = await db.query(
      `INSERT INTO patient_case (patient_id, doctor_id, description, created_by)
       VALUES ($1, $2, 'In-use insurance test case', $2)
       RETURNING id`,
      [patientId, userId],
    );
    const caseId = caseRes.rows[0].id;

    const encRes = await db.query(
      `INSERT INTO encounter (kind, case_id, patient_id, doctor_id, admitted_at, status, request_key, created_by)
       VALUES ('opd', $1, $2, $3, now(), 'active', $4, $3)
       RETURNING id`,
      [caseId, patientId, userId, `enc-ins-key-${suffix}`],
    );
    const encId = encRes.rows[0].id;

    await db.query(
      `INSERT INTO ipd_admission_details (encounter_id, insurance_id)
       VALUES ($1, $2)`,
      [encId, dbIns[0].id],
    );

    // Attempt to delete: expect 409 RECORD_IN_USE
    const delBlockedPromise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/insurances/${dbIns[0].id}`) &&
        r.request().method() === "DELETE",
    );
    await page.locator("tr", { hasText: insName }).getByTitle("Delete").click();
    const delBlockedResp = await delBlockedPromise;
    console.log("[9/10] Delete blocked response:", delBlockedResp.status());
    assert.equal(delBlockedResp.status(), 409);

    // Insurance must still be in the table and database
    await page.locator("tr", { hasText: insName }).waitFor();
    const stillInDb = (
      await db.query("SELECT id FROM insurance WHERE id = $1", [dbIns[0].id])
    ).rows;
    assert.equal(stillInDb.length, 1);
    passed(
      "in-use insurance deletion returns 409 Conflict and preserves database record",
    );

    // 10. Unreferenced Deletion (200 / 204)
    console.log(
      "[10/10] Testing unreferenced deletion (200/204 No Content)...",
    );
    await db.query(
      "DELETE FROM ipd_admission_details WHERE encounter_id = $1",
      [encId],
    );
    await db.query("DELETE FROM encounter WHERE id = $1", [encId]);
    await db.query("DELETE FROM patient_case WHERE id = $1", [caseId]);
    await db.query("DELETE FROM patient WHERE id = $1", [patientId]);

    const delSuccessPromise = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/insurances/${dbIns[0].id}`) &&
        r.request().method() === "DELETE",
    );
    await page.locator("tr", { hasText: insName }).getByTitle("Delete").click();
    const delSuccessResp = await delSuccessPromise;
    console.log("[10/10] Delete success response:", delSuccessResp.status());
    assert.ok(
      [200, 204].includes(delSuccessResp.status()),
      `expected 200 or 204, got ${delSuccessResp.status()}`,
    );

    // Insurance should disappear from table
    await page.locator("tr", { hasText: insName }).waitFor({ state: "hidden" });
    const deletedDb = (
      await db.query("SELECT id FROM insurance WHERE id = $1", [dbIns[0].id])
    ).rows;
    assert.equal(deletedDb.length, 0);

    const deletedDiseases = (
      await db.query(
        "SELECT id FROM insurance_disease WHERE insurance_id = $1",
        [dbIns[0].id],
      )
    ).rows;
    assert.equal(deletedDiseases.length, 0);

    passed(
      "unreferenced insurance deletion succeeds and removes record and disease lines from database",
    );

    // 11. Pagination, Server Search Across Pages, and Export >25 (R4, R5)
    console.log(
      "[11/12] Testing insurance pagination, search across pages, and export with >30 records...",
    );
    const bulkInsIds = [];
    for (let i = 1; i <= 32; i++) {
      const iId = randomUUID();
      bulkInsIds.push(iId);
      const insTitle = `Bulk Insurance Policy ${String(i).padStart(2, "0")} ${suffix}`;
      await db.query(
        `INSERT INTO insurance (id, name, service_tax_minor, discount, insurance_no, insurance_code, hospital_rate_minor, remark, status, total_minor, created_at, updated_at)
         VALUES ($1, $2, 5000, 0, $3, $4, 15000, 'Bulk policy', 1, 20000, now() - interval '${35 - i} minutes', now())`,
        [
          iId,
          insTitle,
          `POL-BLK-${String(i).padStart(2, "0")}-${suffix}`,
          `COD-BLK-${String(i).padStart(2, "0")}-${suffix}`,
        ],
      );
      await db.query(
        `INSERT INTO insurance_disease (id, insurance_id, disease_name, disease_charge_minor)
         VALUES ($1, $2, 'Consultation', 5000)`,
        [randomUUID(), iId],
      );
    }

    // Reload page to view populated catalog
    await open();

    // Verify pagination controls on Page 1
    const paginationText = page.locator(".billing-pagination");
    await paginationText.waitFor();
    assert.ok(
      (await paginationText.innerText()).includes(
        "Showing 1 to 10 of 32 Results",
      ),
      "should show 1 to 10 of 32 results on page 1",
    );

    // Switch page size to 25
    const pageSizeSelect = page.locator(".billing-pagination select");
    const p25Promise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") && r.url().includes("limit=25"),
    );
    await pageSizeSelect.selectOption("25");
    await p25Promise;
    assert.ok(
      (await paginationText.innerText()).includes(
        "Showing 1 to 25 of 32 Results",
      ),
      "should show 1 to 25 of 32 results after page size change",
    );

    // Navigate to Page 2
    const pPage2Promise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") && r.url().includes("page=2"),
    );
    await page
      .locator(".billing-pagination-controls button", { hasText: "2" })
      .click();
    await pPage2Promise;
    assert.ok(
      (await paginationText.innerText()).includes(
        "Showing 26 to 32 of 32 Results",
      ),
      "page 2 should display records 26 to 32",
    );

    // Test server search for a record that sorts to later page
    const targetSearchIns = `Bulk Insurance Policy 31 ${suffix}`;
    const searchPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") && r.url().includes("search="),
    );
    await page.locator(".billing-toolbar input").fill(targetSearchIns);
    await searchPromise;
    await page.locator("tr", { hasText: targetSearchIns }).waitFor();
    assert.ok(
      (await paginationText.innerText()).includes(
        "Showing 1 to 1 of 1 Results",
      ),
      "search results should update total to 1",
    );

    // Clear search
    const clearPromise = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/insurances") &&
        !r.url().includes("search=Bulk+Insurance"),
    );
    await page.locator(".billing-toolbar input").fill("");
    await clearPromise;
    await page.waitForTimeout(300);

    // Verify complete export returns all 32+ insurances (R4)
    const exportRes = await page.request.get(
      `${base}/api/hms/insurances-export`,
    );
    assert.equal(exportRes.status(), 200);
    const exportData = await exportRes.json();
    assert.ok(
      exportData.insurances && exportData.insurances.length >= 32,
      `export must return all 32 records, got ${exportData?.insurances?.length}`,
    );
    passed(
      "insurance pagination, server search across pages, and complete export >25 (R4, R5)",
    );

    // 12. Error and Failure Isolation (R6)
    console.log(
      "[12/12] Testing partial failure isolation, error states, and retry (R6)...",
    );
    // Intercept insurances endpoint with HTTP 500
    await page.route("**/api/hms/insurances*", (route) => {
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Simulated upstream failure in insurances",
        }),
      });
    });

    // Trigger sync / reload
    await page.getByRole("button", { name: /Sync Backend/ }).click();
    await page.waitForTimeout(400);

    // Verify error banner is visible
    const errorBanner = page.locator(".alert-notice", {
      hasText: "Simulated upstream failure in insurances",
    });
    await errorBanner.waitFor();

    // Verify genuine empty notice is NOT visible
    assert.equal(
      await page.getByText("No insurance policies found in catalog").count(),
      0,
      "failed load must not display genuine empty catalog notice",
    );

    // Unroute and click Retry
    await page.unroute("**/api/hms/insurances*");
    const retryPromise = page.waitForResponse(
      (r) => r.url().includes("/api/hms/insurances") && r.status() === 200,
    );
    await errorBanner.getByRole("button", { name: "Retry" }).click();
    await retryPromise;

    // Verify table restored
    await page
      .locator("tr", { hasText: "Bulk Insurance Policy" })
      .first()
      .waitFor();
    passed(
      "insurance failure isolation, visible error alerts, and retry state recovery (R6)",
    );
  } finally {
    await page.close();
  }
}
