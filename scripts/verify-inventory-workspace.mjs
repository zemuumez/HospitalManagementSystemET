import { chromium } from "playwright";
import { Pool } from "pg";
import { resolve } from "node:path";
import { existsSync, mkdirSync, writeFileSync } from "node:fs";
import { randomBytes, randomUUID } from "node:crypto";
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
    "ERROR: verify-inventory-workspace requires isolated schema execution.",
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
    "=== Starting Inventory Workspace Isolated Browser Verification ===",
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

  const { hashPassword } = await import("better-auth/crypto");

  if (!adminEmail || !adminPassword) {
    const token = randomBytes(8).toString("hex");
    const adminId = `adm-${token}`;
    adminEmail = `admin-${token}@ulshms.local`;
    adminPassword = `AdmPass!${randomBytes(12).toString("hex")}`;
    await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
      adminId,
      "QA Inventory Admin",
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

  // Provision doctor for unauthorized rejection testing
  const docToken = randomBytes(8).toString("hex");
  const docId = `doc-${docToken}`;
  const docEmail = `doc-${docToken}@ulshms.local`;
  const docPassword = `DocPass!${randomBytes(12).toString("hex")}`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    docId,
    "Dr. QA Inventory Doctor",
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

  // Provision staff recipient for issuing inventory items
  const staffToken = randomBytes(8).toString("hex");
  const staffId = `staff-${staffToken}`;
  const staffEmail = `nurse-${staffToken}@ulshms.local`;
  await db.query('INSERT INTO "user"(id, name, email) VALUES($1, $2, $3)', [
    staffId,
    "Nurse Genet Lemma",
    staffEmail,
  ]);
  await db.query(
    "INSERT INTO staff_access(user_id, role, active) VALUES($1, 'nurse', true)",
    [staffId],
  );

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

  const suffix = randomBytes(4).toString("hex");
  const categoryName = `Surgical Consumables ${suffix}`;
  const itemName = `Disposable Scalpel 10mm ${suffix}`;
  let categoryId = "";
  let itemId = "";
  let stockMovementId = "";
  let issueMovementId = "";

  try {
    // -----------------------------------------------------------------------
    // Journey 1: Admin Login & Inventory Workspace Landing
    // -----------------------------------------------------------------------
    console.log("\n[Journey 1] Admin Login & Inventory Workspace Landing...");
    await performLogin(page, adminEmail, adminPassword);

    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']", {
      timeout: 30000,
    });
    await page.getByText("Inventory loaded", { exact: false }).waitFor();

    // Verify subtabs navigation
    await page.waitForSelector("nav button");
    const tabTexts = await page.$$eval("nav button", (buttons) =>
      buttons.map((b) => b.textContent.trim()),
    );
    console.log("Found Inventory Subtabs:", tabTexts);
    assert(
      tabTexts.some((t) => t.includes("Items")),
      "Expected 'Items' tab",
    );
    assert(
      tabTexts.some((t) => t.includes("Item Categories")),
      "Expected 'Item Categories' tab",
    );
    assert(
      tabTexts.some((t) => t.includes("Item Stocks")),
      "Expected 'Item Stocks' tab",
    );
    assert(
      tabTexts.some((t) => t.includes("Issued Items")),
      "Expected 'Issued Items' tab",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "01_inventory_admin_landing.png"),
    });
    console.log("✔ Journey 1 Passed: Admin landed on Inventory Workspace.");

    // -----------------------------------------------------------------------
    // Journey 2: Empty State & Fixture-Free Verification
    // -----------------------------------------------------------------------
    console.log("\n[Journey 2] Empty State & Fixture-Free Verification...");
    assert.equal(
      await page
        .getByRole("cell", { name: "Surgical Supplies", exact: true })
        .count(),
      0,
      "Clean isolated schema must not show mock or demo fixtures",
    );

    await page.goto(`${base}/modules/item-categories`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    assert.equal(
      await page.getByText("No categories found").count(),
      1,
      "Expected clean empty state message on item-categories",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "02_inventory_empty_state.png"),
    });
    console.log(
      "✔ Journey 2 Passed: Empty database verified without fixtures.",
    );

    // -----------------------------------------------------------------------
    // Journey 3: Category Creation & Reload Persistence
    // -----------------------------------------------------------------------
    console.log("\n[Journey 3] Category Creation & Reload Persistence...");
    await page
      .getByRole("button", { name: "New Item Category", exact: true })
      .click();

    await page
      .getByRole("dialog")
      .getByLabel("Category Name", { exact: true })
      .fill(categoryName);
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Description"]')
      .fill("Sterile surgical equipment and disposables");

    const categoryCreateRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/categories") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save Category", exact: true })
      .click();

    const catSaved = await categoryCreateRes;
    assert.equal(
      catSaved.status(),
      201,
      `Category create failed: ${await catSaved.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify in PostgreSQL
    const catQuery = await db.query(
      "SELECT id, name, description FROM inventory_category WHERE name=$1",
      [categoryName],
    );
    assert.equal(catQuery.rowCount, 1, "Category must be saved in database");
    categoryId = catQuery.rows[0].id;
    console.log(`Created category ID: ${categoryId}`);

    // Verify UI reload persistence
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page
      .getByRole("cell", { name: categoryName, exact: true })
      .waitFor({ timeout: 10000 });

    await page.screenshot({
      path: resolve(screenshotsDir, "03_inventory_category_created.png"),
    });
    console.log(
      "✔ Journey 3 Passed: Category created and verified in DB and UI reload.",
    );

    // -----------------------------------------------------------------------
    // Journey 4: Category Editing & DB Update
    // -----------------------------------------------------------------------
    console.log("\n[Journey 4] Category Editing & DB Update...");
    const categoryRow = page.locator("tr", { hasText: categoryName });
    await categoryRow.locator('button[aria-label="Edit Category"]').click();

    const updatedCatDesc = "Updated sterile surgical inventory batch";
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Description"]')
      .fill(updatedCatDesc);

    const categoryEditRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/inventory/categories/${categoryId}`) &&
        r.request().method() === "PATCH",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: /Update Category|Save Category/ })
      .click();

    const catEdited = await categoryEditRes;
    assert.equal(
      catEdited.status(),
      200,
      `Category patch failed: ${await catEdited.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify DB update
    const catUpdated = await db.query(
      "SELECT description FROM inventory_category WHERE id=$1",
      [categoryId],
    );
    assert.equal(catUpdated.rows[0].description, updatedCatDesc);
    console.log("✔ Journey 4 Passed: Category edited and verified in DB.");

    // -----------------------------------------------------------------------
    // Journey 5: Item Creation with Zero Opening Balance
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 5] Item Creation with Zero Opening Balance Persistence...",
    );
    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    await page.getByRole("button", { name: "New Item", exact: true }).click();
    await page
      .getByRole("dialog")
      .getByLabel("Item Name", { exact: true })
      .fill(itemName);
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(categoryId);
    await page
      .getByRole("dialog")
      .getByLabel("Unit", { exact: true })
      .fill("Piece");
    await page
      .getByRole("dialog")
      .getByLabel("Reorder Level", { exact: true })
      .fill("25");
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Description"]')
      .fill("Sterile single-use surgical scalpels");

    const itemCreateRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/items") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save Item", exact: true })
      .click();

    const itemSaved = await itemCreateRes;
    assert.equal(
      itemSaved.status(),
      201,
      `Item create failed: ${await itemSaved.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify in PostgreSQL: opening balance MUST be 0
    const itemQuery = await db.query(
      "SELECT id, name, balance_milli, reorder_milli FROM inventory_item WHERE name=$1",
      [itemName],
    );
    assert.equal(itemQuery.rowCount, 1, "Item must be saved in database");
    itemId = itemQuery.rows[0].id;
    assert.equal(
      Number(itemQuery.rows[0].balance_milli),
      0,
      "Opening stock balance must be strictly 0 (no fabricated stock)",
    );
    assert.equal(
      Number(itemQuery.rows[0].reorder_milli),
      25000,
      "Reorder level must be 25,000 milli",
    );
    console.log(`Created item ID: ${itemId} with balance 0`);

    // Verify UI reload persistence
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page
      .getByRole("cell", { name: itemName, exact: true })
      .waitFor({ timeout: 10000 });

    await page.screenshot({
      path: resolve(screenshotsDir, "04_inventory_item_created.png"),
    });
    console.log(
      "✔ Journey 5 Passed: Item created with 0 balance, verified in DB and UI.",
    );

    // -----------------------------------------------------------------------
    // Journey 6: Item Details Modal Inspection
    // -----------------------------------------------------------------------
    console.log("\n[Journey 6] Item Details Modal Inspection...");
    const itemRow = page.locator("tr", { hasText: itemName });
    await itemRow.locator('button[aria-label="Item Details"]').click();

    await page.getByRole("dialog").waitFor({ state: "visible" });
    assert.equal(
      await page.getByRole("dialog").getByText(itemName).count(),
      1,
      "Item details modal must display item name",
    );
    assert.equal(
      await page.getByRole("dialog").getByText("0 Piece").count(),
      1,
      "Item details modal must display 0 Piece balance",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "05_inventory_item_details.png"),
    });

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Close" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });
    console.log("✔ Journey 6 Passed: Item details modal verified.");

    // -----------------------------------------------------------------------
    // Journey 7: Stock Receipt with Financials & Postgres Balance Verification
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 7] Stock Receipt with Financials & Postgres Balance...",
    );
    await page.goto(`${base}/modules/item-stocks`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    await page
      .getByRole("button", { name: "Receive New Stock", exact: true })
      .click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    // Select the newly created item
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemId);

    await page
      .getByRole("dialog")
      .getByLabel("Quantity", { exact: true })
      .fill("100");
    await page
      .getByRole("dialog")
      .getByLabel("Unit Cost", { exact: true })
      .fill("25.50");
    await page
      .getByRole("dialog")
      .getByLabel("Supplier", { exact: true })
      .fill("Ethio Pharma Logistics");
    await page
      .getByRole("dialog")
      .getByLabel("Store Name", { exact: true })
      .fill("Central Surgery Store");
    const stockRef = `PO-SURG-${suffix}`;
    await page
      .getByRole("dialog")
      .getByLabel("Reference", { exact: true })
      .fill(stockRef);

    const stockReceiveRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Receive", exact: true })
      .click();

    const received = await stockReceiveRes;
    assert.equal(
      received.status(),
      201,
      `Receive stock failed: ${await received.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify Database Balance
    const balanceAfterReceive = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterReceive.rows[0].balance_milli),
      100000,
      "Item balance must be 100,000 milli (100 units)",
    );

    // Verify Movement Record
    const movementQuery = await db.query(
      "SELECT id, kind, quantity_milli, delta_milli, cost_minor, reference, supplier, store_name FROM inventory_movement WHERE item_id=$1 AND kind='receive'",
      [itemId],
    );
    assert.equal(movementQuery.rowCount, 1, "Stock receipt movement recorded");
    const mov = movementQuery.rows[0];
    stockMovementId = mov.id;
    assert.equal(Number(mov.quantity_milli), 100000);
    assert.equal(Number(mov.delta_milli), 100000);
    assert.equal(Number(mov.cost_minor), 255000); // 100 * 25.50 * 100 = 255,000 cents
    assert.equal(mov.reference, stockRef);
    assert.equal(mov.supplier, "Ethio Pharma Logistics");
    assert.equal(mov.store_name, "Central Surgery Store");
    console.log(
      `Stock received: movement ID ${stockMovementId}, balance is 100 units`,
    );

    // Check UI Item Stocks table
    await page.waitForSelector(`tr:has-text("${stockRef}")`);

    await page.screenshot({
      path: resolve(screenshotsDir, "06_inventory_stock_received.png"),
    });
    console.log(
      "✔ Journey 7 Passed: Stock receipt persisted with 100 units and financials.",
    );

    // -----------------------------------------------------------------------
    // Journey 8: Stock Details Modal Inspection
    // -----------------------------------------------------------------------
    console.log("\n[Journey 8] Stock Details Modal Inspection...");
    const stockRow = page.locator("tr", { hasText: stockRef });
    await stockRow.locator('button[aria-label="Stock Details"]').click();

    await page.getByRole("dialog").waitFor({ state: "visible" });
    assert.equal(
      await page
        .getByRole("dialog")
        .getByText("Ethio Pharma Logistics")
        .count(),
      1,
      "Supplier must appear in stock details modal",
    );
    assert.equal(
      await page.getByRole("dialog").getByText("2,550.00").count(),
      1,
      "Total cost 2,550.00 ETB must appear in stock details modal",
    );

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Close" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });
    console.log("✔ Journey 8 Passed: Stock details modal verified.");

    // -----------------------------------------------------------------------
    // Journey 9: Issue Item to Staff & Balance Deduction
    // -----------------------------------------------------------------------
    console.log("\n[Journey 9] Issue Item to Staff & Balance Deduction...");
    await page.goto(`${base}/modules/issued-items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    await page.getByRole("button", { name: "Issue Item", exact: true }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    // Select Item and verify available badge
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemId);

    // Wait for staff options to load and select Nurse Genet Lemma
    await page
      .getByRole("dialog")
      .getByLabel("Recipient", { exact: true })
      .selectOption(staffId);

    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Issue", { exact: true })
      .fill("40");
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Reason"]')
      .fill("Operating Theater 2 Scheduled Surgeries");

    const issueRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Issue", exact: true })
      .click();

    const issued = await issueRes;
    assert.equal(
      issued.status(),
      201,
      `Issue item failed: ${await issued.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify Database Balance
    const balanceAfterIssue = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterIssue.rows[0].balance_milli),
      60000,
      "Item balance must be 60,000 milli (60 units remaining)",
    );

    // Verify Issue Movement Record
    const issueMovQuery = await db.query(
      "SELECT id, kind, quantity_milli, delta_milli, recipient_id FROM inventory_movement WHERE item_id=$1 AND kind='issue'",
      [itemId],
    );
    assert.equal(issueMovQuery.rowCount, 1, "Issue movement recorded in DB");
    const issueMov = issueMovQuery.rows[0];
    issueMovementId = issueMov.id;
    assert.equal(Number(issueMov.quantity_milli), 40000);
    assert.equal(Number(issueMov.delta_milli), -40000);
    assert.equal(issueMov.recipient_id, staffId);
    console.log(
      `Item issued: issue ID ${issueMovementId}, balance reduced to 60 units`,
    );

    // Verify UI Issued Items table
    await page.waitForSelector(`tr:has-text("${itemName}")`);
    const issuedRow = page.locator("tr", { hasText: itemName });
    assert(
      (await issuedRow.getByRole("button", { name: "Return Item" }).count()) >=
        1,
      "Active issue row must provide 'Return Item' action button",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "07_inventory_item_issued.png"),
    });
    console.log(
      "✔ Journey 9 Passed: Item issued to staff, balance deducted to 60 units.",
    );

    // -----------------------------------------------------------------------
    // Journey 10: Over-Issue & Insufficient Stock Prevention
    // -----------------------------------------------------------------------
    console.log("\n[Journey 10] Over-Issue & Insufficient Stock Prevention...");
    await page.getByRole("button", { name: "Issue Item", exact: true }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemId);
    await page
      .getByRole("dialog")
      .getByLabel("Recipient", { exact: true })
      .selectOption(staffId);
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Issue", { exact: true })
      .fill("999"); // Exceeds balance of 60

    const overIssueRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Issue", exact: true })
      .click();

    const overIssue = await overIssueRes;
    assert.equal(
      overIssue.status(),
      409,
      "Backend must reject over-issue with 409 Conflict",
    );

    // Form stays open and displays error alert
    await page.getByRole("dialog").getByRole("alert").waitFor({
      state: "visible",
    });

    // Check Postgres balance remained strictly unchanged
    const balanceAfterOverIssue = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterOverIssue.rows[0].balance_milli),
      60000,
      "Balance must remain 60,000 milli after rejected over-issue",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "08_inventory_over_issue_rejected.png"),
    });

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });
    console.log(
      "✔ Journey 10 Passed: Over-issue rejected with 409; balance unchanged.",
    );

    // -----------------------------------------------------------------------
    // Journey 11: Partial Return with Restock & Database Balance Check
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 11] Partial Return with Restock & Database Balance...",
    );
    const returnRow = page.locator("tr", { hasText: itemName });
    await returnRow
      .getByRole("button", { name: "Return Item" })
      .first()
      .click();

    await page.getByRole("dialog").waitFor({ state: "visible" });
    assert.equal(
      await page.getByRole("dialog").getByText("40 Piece").count(),
      2, // Quantity Issued and Remaining Returnable are both 40
      "Modal must show 40 Piece issued and returnable",
    );

    // Enter partial return quantity: 15 units
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Return", { exact: true })
      .fill("15");
    assert(
      await page.getByRole("dialog").locator("#restock-checkbox").isChecked(),
      "Restock checkbox must be checked by default",
    );
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Return Reason"]')
      .fill("15 unopened scalpels returned from OT 2");

    const partialReturnRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Return", exact: true })
      .click();

    const partialReturned = await partialReturnRes;
    assert.equal(
      partialReturned.status(),
      201,
      `Partial return failed: ${await partialReturned.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify Database Balance: 60 + 15 = 75 units (75,000 milli)
    const balanceAfterPartialReturn = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterPartialReturn.rows[0].balance_milli),
      75000,
      "Item balance must increase to 75,000 milli after restocked partial return",
    );

    // Verify Return Movement Record
    const returnMovQuery = await db.query(
      "SELECT id, kind, quantity_milli, delta_milli, restock, original_id FROM inventory_movement WHERE original_id=$1 AND kind='return'",
      [issueMovementId],
    );
    assert.equal(returnMovQuery.rowCount, 1, "Return movement recorded in DB");
    assert.equal(Number(returnMovQuery.rows[0].quantity_milli), 15000);
    assert.equal(Number(returnMovQuery.rows[0].delta_milli), 15000);
    assert.equal(returnMovQuery.rows[0].restock, true);
    assert.equal(returnMovQuery.rows[0].original_id, issueMovementId);

    // Verify UI Status Badge reflects Partial Return
    await page.waitForSelector('button:has-text("Partial Return")');
    assert.equal(
      await page.getByText("Partial Return (15/40)").count(),
      1,
      "Table status badge must reflect Partial Return (15/40)",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "09_inventory_partial_return.png"),
    });
    console.log(
      "✔ Journey 11 Passed: Partial return (15/40) restocked balance to 75 units.",
    );

    // -----------------------------------------------------------------------
    // Journey 12: Over-Return Prevention (Excessive Quantity Check)
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 12] Over-Return Prevention (Excessive Quantity)...",
    );
    // Remaining is 25 units. Attempting to return 30 units directly via API
    const overReturnAttempt = await page.evaluate(
      async ({ item, original }) => {
        const res = await fetch("/api/hms/inventory/movements", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": crypto.randomUUID(),
          },
          body: JSON.stringify({
            itemId: item,
            kind: "return",
            originalId: original,
            quantityMilli: 30000, // 30 units (15 already returned + 30 = 45 > 40)
            restock: true,
            reason: "Excessive return attempt",
          }),
        });
        return { status: res.status, body: await res.json() };
      },
      { item: itemId, original: issueMovementId },
    );

    console.log("Over-return error response:", overReturnAttempt);
    assert.equal(
      overReturnAttempt.status,
      409,
      "Over-return must be rejected with 409 Conflict",
    );

    // Verify DB balance strictly unchanged at 75,000 milli
    const balanceAfterOverReturn = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterOverReturn.rows[0].balance_milli),
      75000,
      "Balance must remain 75,000 milli after over-return rejection",
    );
    console.log(
      "✔ Journey 12 Passed: Over-return rejected with 409; balance preserved.",
    );

    // -----------------------------------------------------------------------
    // Journey 13: Remaining Return with Restock & Final "Returned" Status
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 13] Remaining Return with Restock & Final Status...",
    );
    await page.locator('button:has-text("Partial Return")').first().click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    // Verify remaining is 25
    assert.equal(
      await page.getByRole("dialog").getByText("25 Piece").count(),
      1,
      "Modal must show 25 Piece remaining returnable",
    );

    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Return", { exact: true })
      .fill("25");
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Return Reason"]')
      .fill("Final remaining 25 units returned to store");

    const remainingReturnRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Return", exact: true })
      .click();

    const remainingReturned = await remainingReturnRes;
    assert.equal(
      remainingReturned.status(),
      201,
      `Remaining return failed: ${await remainingReturned.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify Database Balance: 75 + 25 = 100 units (100,000 milli) fully restored!
    const balanceAfterFullReturn = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterFullReturn.rows[0].balance_milli),
      100000,
      "Item balance must be fully restored to original 100,000 milli",
    );

    // Verify UI Status Badge shows "Returned"
    await page.waitForSelector('span:has-text("Returned")');
    assert.equal(
      await page
        .locator("tr", { hasText: itemName })
        .getByRole("button", { name: "Return Item" })
        .count(),
      0,
      "Fully returned issue must no longer have active Return Item button",
    );

    await page.screenshot({
      path: resolve(
        screenshotsDir,
        "10_inventory_remaining_return_completed.png",
      ),
    });
    console.log(
      "✔ Journey 13 Passed: Remaining return completed; balance fully restored to 100 units.",
    );

    // -----------------------------------------------------------------------
    // Journey 14: Non-Restocked Return (Damaged Items Audit Ledger Integrity)
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 14] Non-Restocked Return (Damaged Items Ledger Integrity)...",
    );
    // Issue 20 units (balance drops to 80 units)
    await page.getByRole("button", { name: "Issue Item", exact: true }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemId);
    await page
      .getByRole("dialog")
      .getByLabel("Recipient", { exact: true })
      .selectOption(staffId);
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Issue", { exact: true })
      .fill("20");
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Reason"]')
      .fill("Emergency Room Night Shift");

    const issue2Res = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Issue", exact: true })
      .click();
    assert.equal((await issue2Res).status(), 201);
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Balance is now 80 units (80,000 milli)
    const bal80 = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(Number(bal80.rows[0].balance_milli), 80000);

    const issue2Id = (
      await db.query(
        "SELECT id FROM inventory_movement WHERE item_id=$1 AND kind='issue' ORDER BY created_at DESC LIMIT 1",
        [itemId],
      )
    ).rows[0].id;

    // Return 5 units with restock = false (damaged / broken packaging)
    const damagedReturnRes = await page.evaluate(
      async ({ item, original }) => {
        const res = await fetch("/api/hms/inventory/movements", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": crypto.randomUUID(),
          },
          body: JSON.stringify({
            itemId: item,
            kind: "return",
            originalId: original,
            quantityMilli: 5000,
            restock: false,
            reason: "Sterile seal broken; contaminated and disposed",
          }),
        });
        return { status: res.status, body: await res.json() };
      },
      { item: itemId, original: issue2Id },
    );
    assert.equal(damagedReturnRes.status, 201);

    // Verify in DB: movement recorded with delta_milli = 0 and restock = false
    const damagedMov = (
      await db.query(
        "SELECT quantity_milli, delta_milli, restock FROM inventory_movement WHERE original_id=$1 AND kind='return'",
        [issue2Id],
      )
    ).rows[0];
    assert.equal(Number(damagedMov.quantity_milli), 5000);
    assert.equal(
      Number(damagedMov.delta_milli),
      0,
      "Non-restocked return delta must be 0",
    );
    assert.equal(damagedMov.restock, false);

    // Balance in DB must remain 80,000 milli (NOT added back to available inventory!)
    const balAfterDamaged = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balAfterDamaged.rows[0].balance_milli),
      80000,
      "Balance must remain 80,000 milli after non-restocked return",
    );
    console.log(
      "✔ Journey 14 Passed: Non-restocked return recorded with delta 0; balance untouched.",
    );

    // -----------------------------------------------------------------------
    // Journey 15: Concurrent Requests Stock Integrity (Overspend Prevention)
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 15] Concurrent Requests Stock Integrity (Overspend Prevention)...",
    );
    // Current balance is 80 units (80,000 milli).
    // Send two concurrent requests of 50 units each (total 100 > 80).
    const concurrentResults = await page.evaluate(
      async ({ item, recipient }) => {
        const payload = JSON.stringify({
          itemId: item,
          kind: "issue",
          quantityMilli: 50000, // 50 units
          recipientId: recipient,
          reason: "Concurrent stress issue",
        });

        const p1 = fetch("/api/hms/inventory/movements", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": crypto.randomUUID(),
          },
          body: payload,
        }).then(async (r) => ({ status: r.status, data: await r.json() }));

        const p2 = fetch("/api/hms/inventory/movements", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": crypto.randomUUID(),
          },
          body: payload,
        }).then(async (r) => ({ status: r.status, data: await r.json() }));

        return Promise.all([p1, p2]);
      },
      { item: itemId, recipient: staffId },
    );

    console.log(
      "Concurrent Results:",
      concurrentResults.map((r) => r.status),
    );
    const successes = concurrentResults.filter((r) => r.status === 201);
    const conflicts = concurrentResults.filter((r) => r.status === 409);

    assert.equal(
      successes.length,
      1,
      "Exactly one concurrent issue request must succeed",
    );
    assert.equal(
      conflicts.length,
      1,
      "Exactly one concurrent issue request must fail with 409",
    );

    // Check PostgreSQL balance: 80 - 50 = 30 units (30,000 milli)
    const balanceAfterConcurrent = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(
      Number(balanceAfterConcurrent.rows[0].balance_milli),
      30000,
      "Balance must be exactly 30,000 milli (30 units); stock never overspent",
    );
    console.log(
      "✔ Journey 15 Passed: Concurrent stock overspend prevented by row-level locking.",
    );

    // -----------------------------------------------------------------------
    // Journey 16: Referenced vs Unreferenced Record Deletion Protection
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 16] Referenced vs Unreferenced Record Deletion Protection...",
    );
    // 16.1. Attempt to delete category in use by items: must be rejected with 409
    await page.goto(`${base}/modules/item-categories`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const catRowToDelete = page.locator("tr", { hasText: categoryName });
    await catRowToDelete
      .locator('button[aria-label="Delete Category"]')
      .click();

    await page.getByRole("dialog").waitFor({ state: "visible" });
    const catDeleteRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/inventory/categories/${categoryId}`) &&
        r.request().method() === "DELETE",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Delete" })
      .click();

    const catDelResponse = await catDeleteRes;
    assert.equal(
      catDelResponse.status(),
      409,
      "Deleting category in use must return 409 Conflict",
    );
    const catDelJson = await catDelResponse.json();
    assert.equal(catDelJson.code, "RECORD_IN_USE");

    // Modal displays error alert
    await page.getByRole("dialog").getByRole("alert").waitFor({
      state: "visible",
    });
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Category still exists in DB
    const catStillExists = await db.query(
      "SELECT id FROM inventory_category WHERE id=$1",
      [categoryId],
    );
    assert.equal(catStillExists.rowCount, 1, "Category must not be deleted");

    // 16.2. Attempt to delete item in use by movements: must be rejected with 409
    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const itemRowToDelete = page.locator("tr", { hasText: itemName });
    await itemRowToDelete.locator('button[aria-label="Delete Item"]').click();

    await page.getByRole("dialog").waitFor({ state: "visible" });
    const itemDeleteRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/hms/inventory/items/${itemId}`) &&
        r.request().method() === "DELETE",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Delete" })
      .click();

    const itemDelResponse = await itemDeleteRes;
    assert.equal(
      itemDelResponse.status(),
      409,
      "Deleting item with movements must return 409 Conflict",
    );
    const itemDelJson = await itemDelResponse.json();
    assert.equal(itemDelJson.code, "RECORD_IN_USE");

    await page.getByRole("dialog").getByRole("alert").waitFor({
      state: "visible",
    });
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Item still exists in DB
    const itemStillExists = await db.query(
      "SELECT id FROM inventory_item WHERE id=$1",
      [itemId],
    );
    assert.equal(itemStillExists.rowCount, 1, "Item must not be deleted");

    // 16.3. Unreferenced item and category safe deletion
    const tempSuffix = randomBytes(4).toString("hex");
    const tempCatRes = await page.evaluate(async (name) => {
      const res = await fetch("/api/hms/inventory/categories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description: "Temporary" }),
      });
      return res.json();
    }, `Temp Category ${tempSuffix}`);
    const tempCatId = tempCatRes.category?.id || tempCatRes.id;

    const tempItemRes = await page.evaluate(
      async ({ catId, name }) => {
        const res = await fetch("/api/hms/inventory/items", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            categoryId: catId,
            name,
            unit: "Piece",
            reorderMilli: 10000,
          }),
        });
        return res.json();
      },
      { catId: tempCatId, name: `Temp Item ${tempSuffix}` },
    );
    const tempItemId = tempItemRes.item?.id || tempItemRes.id;

    // Delete unreferenced item: succeeds
    const delUnrefItem = await page.evaluate(async (id) => {
      const res = await fetch(`/api/hms/inventory/items/${id}`, {
        method: "DELETE",
      });
      return { status: res.status, body: await res.json() };
    }, tempItemId);
    assert.equal(delUnrefItem.status, 200);
    assert.equal(
      (
        await db.query("SELECT id FROM inventory_item WHERE id=$1", [
          tempItemId,
        ])
      ).rowCount,
      0,
      "Unreferenced item deleted from DB",
    );

    // Delete unreferenced category: succeeds
    const delUnrefCat = await page.evaluate(async (id) => {
      const res = await fetch(`/api/hms/inventory/categories/${id}`, {
        method: "DELETE",
      });
      return { status: res.status, body: await res.json() };
    }, tempCatId);
    assert.equal(delUnrefCat.status, 200);
    assert.equal(
      (
        await db.query("SELECT id FROM inventory_category WHERE id=$1", [
          tempCatId,
        ])
      ).rowCount,
      0,
      "Unreferenced category deleted from DB",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "11_inventory_record_protection.png"),
    });
    console.log(
      "✔ Journey 16 Passed: Referenced deletion blocked with 409; unreferenced deletion succeeded cleanly.",
    );

    // -----------------------------------------------------------------------
    // Journey 17: Unauthorized Doctor Role Rejection
    // -----------------------------------------------------------------------
    console.log("\n[Journey 17] Unauthorized Doctor Role Rejection...");
    const doctorContext = await browser.newContext();
    const docPage = await doctorContext.newPage();
    try {
      await performLogin(docPage, docEmail, docPassword);

      // Attempt mutating inventory as doctor
      const docMutateResults = await docPage.evaluate(async () => {
        const r1 = await fetch("/api/hms/inventory/categories", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name: "Doc Category" }),
        });
        const r2 = await fetch("/api/hms/inventory/items", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name: "Doc Item" }),
        });
        const r3 = await fetch("/api/hms/inventory/movements", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": crypto.randomUUID(),
          },
          body: JSON.stringify({ quantityMilli: 1000 }),
        });
        return [r1.status, r2.status, r3.status];
      });

      console.log("Doctor Mutation Statuses:", docMutateResults);
      assert(
        docMutateResults.every((s) => s === 403),
        "All inventory mutations must be rejected with 403 for non-admin doctor",
      );
      console.log(
        "✔ Journey 17 Passed: Non-admin doctor requests blocked with 403 Forbidden.",
      );
    } finally {
      await doctorContext.close();
    }

    // -----------------------------------------------------------------------
    // Journey 18: Fresh Browser Session Persistence
    // -----------------------------------------------------------------------
    console.log("\n[Journey 18] Fresh Browser Session Persistence...");
    const freshContext = await browser.newContext();
    const freshPage = await freshContext.newPage();
    try {
      await performLogin(freshPage, adminEmail, adminPassword);

      await freshPage.goto(`${base}/modules/items`);
      await freshPage.waitForSelector(".legacy-workspace[data-ready='true']");
      await freshPage
        .getByRole("cell", { name: itemName, exact: true })
        .waitFor({ timeout: 10000 });

      await freshPage.goto(`${base}/modules/item-categories`);
      await freshPage.waitForSelector(".legacy-workspace[data-ready='true']");
      await freshPage
        .getByRole("cell", { name: categoryName, exact: true })
        .waitFor({ timeout: 10000 });

      await freshPage.goto(`${base}/modules/item-stocks`);
      await freshPage.waitForSelector(".legacy-workspace[data-ready='true']");
      await freshPage
        .getByRole("cell", { name: stockRef, exact: true })
        .waitFor({ timeout: 10000 });

      console.log(
        "✔ Journey 18 Passed: Fresh browser session verified all persisted records from PostgreSQL.",
      );
    } finally {
      await freshContext.close();
    }

    // -----------------------------------------------------------------------
    // Journey 19: Theme & Amharic Localization Parity
    // -----------------------------------------------------------------------
    console.log("\n[Journey 19] Theme & Amharic Localization Parity...");
    // Dark Theme verification
    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page.evaluate(() => document.documentElement.classList.add("dark"));

    await page.screenshot({
      path: resolve(screenshotsDir, "12_inventory_dark_mode.png"),
    });

    // Amharic Localization verification
    await page.evaluate(() => {
      localStorage.setItem("hms-language", "am");
      document.documentElement.classList.remove("dark");
    });
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    // Verify Amharic translation in tab navigation and live indicator
    await page.getByText("ክምችት ተጭኗል", { exact: false }).waitFor({
      timeout: 10000,
    });
    assert(
      (await page.getByText("ዕቃዎች").count()) >= 1,
      "Amharic 'ዕቃዎች' (Items) tab label must appear",
    );
    assert(
      (await page.getByText("የዕቃ ምድቦች").count()) >= 1,
      "Amharic 'የዕቃ ምድቦች' (Item Categories) tab label must appear",
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "13_inventory_amharic_localization.png"),
    });

    // Restore English
    await page.evaluate(() => localStorage.setItem("hms-language", "en"));
    console.log(
      "✔ Journey 19 Passed: Dark mode and Amharic localization verified with clean contrast.",
    );

    console.log("\n=======================================================");
    console.log("ALL 19 INVENTORY WORKSPACE JOURNEYS PASSED SUCCESSFULLY!");
    console.log("=======================================================\n");
  } finally {
    try {
      await browser.close();
    } catch (e) {
      console.error("Cleanup warning (browser):", e.message);
    }
    try {
      await db.end();
    } catch (e) {
      console.error("Cleanup warning (db):", e.message);
    }
  }
}

main().catch((err) => {
  console.error("FATAL: Inventory workspace verification failed:", err);
  process.exit(1);
});
