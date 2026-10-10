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

  let adminId = "";
  const { hashPassword } = await import("better-auth/crypto");

  if (!adminEmail || !adminPassword) {
    const token = randomBytes(8).toString("hex");
    adminId = `adm-${token}`;
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
  } else {
    const existingAdmin = await db.query(
      'SELECT id FROM "user" WHERE email = $1',
      [adminEmail],
    );
    adminId = existingAdmin.rows[0]?.id || "";
  }

  const insertMovementFixture = async ({
    itemId,
    kind,
    quantityMilli,
    deltaMilli,
    recipientId = null,
    originalId = null,
    supplier = "",
    storeName = "Main Hospital Store",
    reference = "",
    costMinor = 0,
    restock = false,
    reason = "Verification fixture",
    issuedDate = "",
    department = "",
    issuedBy = "",
    returnDueDate = "",
    attachmentUrl = "",
    createdAtClause = "NOW()",
  }) => {
    const reqKey = randomBytes(16).toString("hex");
    const reqHash = randomBytes(16).toString("hex");
    const ref =
      reference ||
      (kind === "receive" ? `REF-${randomBytes(6).toString("hex")}` : "");
    const supp = supplier || (kind === "receive" ? "Standard Supplier" : "");

    return db.query(
      `INSERT INTO inventory_movement(
        item_id, kind, quantity_milli, delta_milli, recipient_id, original_id,
        supplier, store_name, reference, cost_minor, restock, reason,
        actor_id, request_key, request_hash, issued_date, department, issued_by,
        return_due_date, attachment_url, created_at
      ) VALUES (
        $1, $2, $3, $4, $5, $6,
        $7, $8, $9, $10, $11, $12,
        $13, $14, $15, $16, $17, $18,
        $19, $20, ${createdAtClause}
      ) RETURNING id`,
      [
        itemId,
        kind,
        quantityMilli,
        deltaMilli,
        recipientId,
        originalId,
        supp,
        storeName,
        ref,
        costMinor,
        restock,
        reason,
        adminId,
        reqKey,
        reqHash,
        issuedDate || "",
        department || "",
        issuedBy || "",
        returnDueDate || "",
        attachmentUrl || "",
      ],
    );
  };

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

    // Select the category and newly created item
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(categoryId);
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

    // Select Category and Item, verify available badge
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(categoryId);
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
      .getByLabel("Category", { exact: true })
      .selectOption(categoryId);
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
      .getByLabel("Category", { exact: true })
      .selectOption(categoryId);
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
    // Journey 19: Theme & Amharic Localization Parity (Finding I5 Regression)
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 19] Theme & Amharic Localization Parity (Finding I5)...",
    );
    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    // Real Shell Dark Theme Toggle verification:
    const themeBtn = page.locator(
      'button[aria-label="Use dark theme"], button[aria-label="Use light theme"]',
    );
    await themeBtn.waitFor({ state: "visible" });
    const currentAria = await themeBtn.getAttribute("aria-label");
    if (currentAria === "Use dark theme") {
      await themeBtn.click();
    }
    // Assert .legacy-shell enters .legacy-dark
    await page.waitForSelector(".legacy-shell.legacy-dark", { timeout: 10000 });
    const storedTheme = await page.evaluate(() =>
      localStorage.getItem("hms-theme"),
    );
    assert.equal(
      storedTheme,
      "dark",
      "localStorage hms-theme must be set to 'dark'",
    );

    // Verify computed contrast/styling on workspace and cards
    const cardBg = await page.$eval(
      ".inventory-card-bg",
      (el) => window.getComputedStyle(el).backgroundColor,
    );
    console.log("Verified Dark mode computed card background:", cardBg);
    assert(
      cardBg.includes("18, 21, 31") ||
        cardBg.includes("26, 29, 45") ||
        cardBg.includes("17, 24, 39"),
      `Expected dark card background (#12151f), got ${cardBg}`,
    );

    await page.screenshot({
      path: resolve(screenshotsDir, "12_inventory_dark_mode.png"),
    });
    console.log(
      "✔ Captured authentic dark mode screenshot: 12_inventory_dark_mode.png",
    );

    // Toggle back to light mode for Amharic check
    await themeBtn.click();
    await page.waitForSelector(".legacy-shell:not(.legacy-dark)", {
      timeout: 10000,
    });

    // Amharic Localization verification:
    await page.evaluate(() => localStorage.setItem("hms-language", "am"));
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    // Verify live indicator and tab labels
    await page
      .getByText("ክምችት ተጭኗል", { exact: false })
      .waitFor({ timeout: 10000 });
    assert(
      (await page.getByText("ዕቃዎች").count()) >= 1,
      "Amharic 'ዕቃዎች' (Items) must appear",
    );
    assert(
      (await page.getByText("የዕቃ ምድቦች").count()) >= 1,
      "Amharic 'የዕቃ ምድቦች' (Item Categories) must appear",
    );
    assert(
      (await page.getByText("የዕቃ ክምችት").count()) >= 1,
      "Amharic 'የዕቃ ክምችት' (Item Stocks) must appear",
    );
    assert(
      (await page.getByText("የተሰጡ ዕቃዎች").count()) >= 1,
      "Amharic 'የተሰጡ ዕቃዎች' (Issued Items) must appear",
    );

    // Verify action buttons
    assert(
      (await page.getByRole("button", { name: "አዲስ ዕቃ" }).count()) >= 1,
      "Amharic 'አዲስ ዕቃ' button must appear",
    );
    assert(
      (await page.getByRole("button", { name: "ክምችት አስምር" }).count()) >= 1,
      "Amharic 'ክምችት አስምር' button must appear",
    );

    // Verify table headers
    assert(
      (await page.getByText("የዕቃ ስም").count()) >= 1,
      "Header 'የዕቃ ስም' must appear",
    );
    assert(
      (await page.getByText("ምድብ").count()) >= 1,
      "Header 'ምድብ' must appear",
    );
    assert(
      (await page.getByText("የድጋሚ ማዘዣ መጠን").count()) >= 1,
      "Header 'የድጋሚ ማዘዣ መጠን' must appear",
    );
    assert(
      (await page.getByText("በክምችት ላይ").count()) >= 1,
      "Header 'በክምችት ላይ' must appear",
    );

    // Verify category filter option & search placeholder
    assert(
      (await page.getByText("ሁሉንም ምድቦች").count()) >= 1,
      "'ሁሉንም ምድቦች' must appear",
    );
    const searchPlaceholder = await page
      .locator('input[aria-label="የዕቃ ፍለጋ"]')
      .getAttribute("placeholder");
    assert(
      searchPlaceholder?.includes("ዕቃዎችን ይፈልጉ"),
      "Search placeholder must be in Amharic",
    );

    // Open "New Item" modal in Amharic and verify form fields
    await page.getByRole("button", { name: "አዲስ ዕቃ" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    assert(
      (await page.getByRole("dialog").getByText("አዲስ የዕቃ ምዝገባ").count()) >= 1,
      "Modal title 'አዲስ የዕቃ ምዝገባ' must appear",
    );
    assert(
      (await page
        .getByRole("dialog")
        .getByRole("button", { name: "ዕቃ መዝግብ" })
        .count()) >= 1,
      "Modal button 'ዕቃ መዝግብ' must appear",
    );
    await page.getByRole("dialog").getByRole("button", { name: "ሰርዝ" }).click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Check "Item Stocks" tab
    await page.getByRole("button", { name: "የዕቃ ክምችት" }).click();
    await page.waitForTimeout(500);
    assert(
      (await page.getByRole("button", { name: "አዲስ ክምችት ተቀበል" }).count()) >= 1,
      "'አዲስ ክምችት ተቀበል' must appear",
    );

    // Check "Issued Items" tab
    await page.getByRole("button", { name: "የተሰጡ ዕቃዎች" }).click();
    await page.waitForTimeout(500);
    assert(
      (await page.getByRole("button", { name: "ዕቃ ስጥ" }).count()) >= 1,
      "'ዕቃ ስጥ' must appear",
    );

    // Return to Items tab for complete Amharic screenshot
    await page.locator('button[data-tab="items"]').click();
    await page.waitForTimeout(500);
    await page.screenshot({
      path: resolve(screenshotsDir, "13_inventory_amharic_localization.png"),
    });
    console.log(
      "✔ Captured authentic Amharic screenshot: 13_inventory_amharic_localization.png",
    );

    // Restore English
    await page.evaluate(() => localStorage.setItem("hms-language", "en"));
    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    console.log(
      "✔ Journey 19 Passed: Dark mode and Amharic localization verified with real controls.",
    );

    // -----------------------------------------------------------------------
    // Journey 20: Finding I1 Regression — Pagination & Silent Cutoff Prevention (>25 Records)
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 20] Finding I1: Pagination & Silent Cutoff Prevention (>25 Records)...",
    );
    const bulkSuffix = randomBytes(4).toString("hex");

    // 1. Seed 28 categories directly in PostgreSQL
    const bulkCategoryIds = [];
    for (let i = 1; i <= 28; i++) {
      const cName = `Bulk Cat ${bulkSuffix} ${i.toString().padStart(2, "0")}`;
      const cRes = await db.query(
        "INSERT INTO inventory_category(name, description) VALUES($1, $2) RETURNING id",
        [cName, `Bulk category ${i}`],
      );
      bulkCategoryIds.push(cRes.rows[0].id);
    }

    // 2. Seed 28 items in PostgreSQL (Item 27 is a special overflow item)
    const bulkItemIds = [];
    let overflowItemId = "";
    const overflowItemName = `Special Overflow Forceps ${bulkSuffix}`;
    for (let i = 1; i <= 28; i++) {
      const iName =
        i === 27
          ? overflowItemName
          : `Bulk Item ${bulkSuffix} ${i.toString().padStart(2, "0")}`;
      const catId = bulkCategoryIds[(i - 1) % bulkCategoryIds.length];
      const initialBal = i === 1 ? 20000 : 0;
      await db.query("BEGIN");
      const iRes = await db.query(
        "INSERT INTO inventory_item(category_id, name, unit, reorder_milli, balance_milli) VALUES($1, $2, 'Piece', 5000, $3) RETURNING id",
        [catId, iName, initialBal],
      );
      const newId = iRes.rows[0].id;
      bulkItemIds.push(newId);
      if (i === 27) overflowItemId = newId;

      if (initialBal > 0) {
        await insertMovementFixture({
          itemId: newId,
          kind: "receive",
          quantityMilli: initialBal,
          deltaMilli: initialBal,
          reference: `INIT-${bulkSuffix}-${i}`,
          supplier: "Bulk Initial Supplier",
          reason: "Initial bulk stock",
          createdAtClause: "NOW() - interval '70 days'",
        });
      }
      await db.query("COMMIT");
    }

    // 3. Seed an Old Outstanding Issue on Item 1 from 60 days ago
    await db.query("BEGIN");
    const oldIssueRes = await insertMovementFixture({
      itemId: bulkItemIds[0],
      kind: "issue",
      quantityMilli: 5000,
      deltaMilli: -5000,
      recipientId: staffId,
      reason: "Old Outstanding Issue 60 Days Ago",
      issuedDate: "2026-08-10",
      createdAtClause: "NOW() - interval '60 days'",
    });
    const oldIssueId = oldIssueRes.rows[0].id;
    await db.query(
      "UPDATE inventory_item SET balance_milli = 15000 WHERE id = $1",
      [bulkItemIds[0]],
    );
    await db.query("COMMIT");

    // 4. Seed 27 newer movements (receive) so movements total > 28
    for (let i = 1; i <= 27; i++) {
      const targetItemId = bulkItemIds[i % bulkItemIds.length];
      await db.query("BEGIN");
      await insertMovementFixture({
        itemId: targetItemId,
        kind: "receive",
        quantityMilli: 10000,
        deltaMilli: 10000,
        reference: `BULK-REC-${bulkSuffix}-${i}`,
        supplier: "Bulk Supplier",
        reason: "Routine Restock",
        createdAtClause: `NOW() - interval '${28 - i} hours'`,
      });
      await db.query(
        "UPDATE inventory_item SET balance_milli = balance_milli + 10000 WHERE id = $1",
        [targetItemId],
      );
      await db.query("COMMIT");
    }
    console.log(
      `Seeded 28 categories, 28 items, and movements including older issue ${oldIssueId}`,
    );

    // 5. Test Categories Pagination
    await page.goto(`${base}/modules/item-categories`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page
      .getByText("Page 1 of 2", { exact: false })
      .waitFor({ timeout: 10000 });
    await page.getByRole("button", { name: "Next" }).click();
    await page
      .getByText("Page 2 of 2", { exact: false })
      .waitFor({ timeout: 10000 });
    await page.getByRole("button", { name: "Previous" }).click();
    await page
      .getByText("Page 1 of 2", { exact: false })
      .waitFor({ timeout: 10000 });

    // 6. Test Items Pagination & Server-Backed Search for Overflow Item
    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page
      .getByText("Page 1 of 2", { exact: false })
      .waitFor({ timeout: 10000 });

    const searchInput = page.locator('input[aria-label="Search Items"]');
    await searchInput.fill(overflowItemName);
    await page
      .getByRole("cell", { name: overflowItemName, exact: true })
      .waitFor({ timeout: 10000 });
    console.log("✔ Server-backed search found overflow item beyond page 1");
    await searchInput.fill("");

    // 7. Test Dialog Dropdowns (Large Catalog Retrieval with limit=500)
    await page.goto(`${base}/modules/item-stocks`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page.getByRole("button", { name: "Receive New Stock" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    const overflowCatId = bulkCategoryIds[26 % bulkCategoryIds.length];
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(overflowCatId);
    const receiveItemOptions = await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .locator("option")
      .allTextContents();
    assert(
      receiveItemOptions.some((opt) => opt.includes(overflowItemName)),
      `Overflow item ${overflowItemName} must be available in Receive modal dropdown`,
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // 8. Test Discoverability of Old Outstanding Issue & Return Workflow
    await page.goto(`${base}/modules/issued-items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const issueSearchInput = page.locator(
      'input[aria-label="Search Movements"]',
    );
    await issueSearchInput.fill("Old Outstanding Issue");
    await page.waitForSelector(`tr:has-text("Old Outstanding Issue")`, {
      timeout: 10000,
    });
    const oldIssueRow = page.locator("tr", {
      hasText: "Old Outstanding Issue",
    });
    await oldIssueRow.getByRole("button", { name: "Return Item" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Return", { exact: true })
      .fill("5");
    const returnSubmitRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Return" })
      .click();
    const retSaved = await returnSubmitRes;
    assert.equal(
      retSaved.status(),
      201,
      `Return old issue failed: ${await retSaved.text()}`,
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    const balRestored = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [bulkItemIds[0]],
    );
    assert.equal(
      Number(balRestored.rows[0].balance_milli),
      20000,
      "Balance must be restored to 20 units",
    );
    console.log(
      "✔ Journey 20 Passed: Pagination, server-backed search, large catalogs, and older issue return verified.",
    );

    // -----------------------------------------------------------------------
    // Journey 21: Finding I2 Regression — Empty-Category & Stale Selection Protection
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 21] Finding I2: Empty-Category & Stale Selection Protection...",
    );
    const catSuffix = randomBytes(4).toString("hex");
    const catARes = await db.query(
      "INSERT INTO inventory_category(name, description) VALUES($1, 'Has items') RETURNING id",
      [`Category Alpha ${catSuffix}`],
    );
    const catAId = catARes.rows[0].id;
    await db.query("BEGIN");
    const itemARes = await db.query(
      "INSERT INTO inventory_item(category_id, name, unit, reorder_milli, balance_milli) VALUES($1, $2, 'Piece', 1000, 50000) RETURNING id",
      [catAId, `Alpha Stethoscope ${catSuffix}`],
    );
    const itemAId = itemARes.rows[0].id;
    await insertMovementFixture({
      itemId: itemAId,
      kind: "receive",
      quantityMilli: 50000,
      deltaMilli: 50000,
      reference: `ALPHA-REC-${catSuffix}`,
      supplier: "Alpha Supplier",
      reason: "Initial alpha stock",
    });
    await db.query("COMMIT");

    const catBRes = await db.query(
      "INSERT INTO inventory_category(name, description) VALUES($1, 'Empty category') RETURNING id",
      [`Category Beta Empty ${catSuffix}`],
    );
    const catBId = catBRes.rows[0].id;

    // 1. Receive Stock Modal
    await page.goto(`${base}/modules/item-stocks`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page.getByRole("button", { name: "Receive New Stock" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catBId);
    assert(
      await page
        .getByRole("dialog")
        .getByLabel("Item", { exact: true })
        .isDisabled(),
      "Item selector must be disabled for empty category",
    );
    assert(
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Confirm Receive" })
        .isDisabled(),
      "Confirm Receive must be disabled for empty category",
    );

    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catAId);
    assert(
      !(await page
        .getByRole("dialog")
        .getByLabel("Item", { exact: true })
        .isDisabled()),
      "Item selector must be enabled for Category Alpha",
    );
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemAId);

    // Switch Category back to empty Category Beta: must CLEAR selection and disable
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catBId);
    const itemBVal = await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .inputValue();
    assert.equal(
      itemBVal,
      "",
      "Selected item must be cleared when switching to empty category",
    );
    assert(
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Confirm Receive" })
        .isDisabled(),
      "Submit button must be disabled",
    );

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // 2. Issue Item Modal
    await page.goto(`${base}/modules/issued-items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page.getByRole("button", { name: "Issue Item" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catBId);
    assert(
      await page
        .getByRole("dialog")
        .getByLabel("Item", { exact: true })
        .isDisabled(),
      "Item selector must be disabled in Issue modal for empty category",
    );
    assert(
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Confirm Issue" })
        .isDisabled(),
      "Confirm Issue must be disabled for empty category",
    );

    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catAId);
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemAId);
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catBId);
    const issueItemBVal = await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .inputValue();
    assert.equal(
      issueItemBVal,
      "",
      "Selected item must be cleared in Issue modal when switching to empty category",
    );
    assert(
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Confirm Issue" })
        .isDisabled(),
      "Confirm Issue must be disabled",
    );

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Assert Database Integrity: Item A balance remains strictly unchanged (50,000 milli)
    const checkItemA = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemAId],
    );
    assert.equal(
      Number(checkItemA.rows[0].balance_milli),
      50000,
      "Item A balance must remain 50,000 milli; no unintended movements",
    );
    console.log(
      "✔ Journey 21 Passed: Empty category disables submission and clears child selection; no unintended stock mutations.",
    );

    // -----------------------------------------------------------------------
    // Journey 22: Finding I3 Regression — Source Fields, Attachments & Void Workflows
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 22] Finding I3: Source Fields, Attachments & Void Workflows...",
    );
    const i3Suffix = randomBytes(4).toString("hex");
    const testAttachUrl =
      "https://files.ulshms.local/receipts/batch-invoice-001.pdf";

    // 1. Stock Receipt with Attachment & Extended Metadata
    await page.goto(`${base}/modules/item-stocks`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page.getByRole("button", { name: "Receive New Stock" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catAId);
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemAId);
    await page
      .getByRole("dialog")
      .getByLabel("Quantity", { exact: true })
      .fill("20");
    await page
      .getByRole("dialog")
      .getByLabel("Unit Cost", { exact: true })
      .fill("45.00");
    await page
      .getByRole("dialog")
      .getByLabel("Supplier", { exact: true })
      .fill("Apex Med Supplies");
    await page
      .getByRole("dialog")
      .getByLabel("Store Name", { exact: true })
      .fill("Emergency Sub-Store");
    const recRef = `REC-ATTACH-${i3Suffix}`;
    await page
      .getByRole("dialog")
      .getByLabel("Reference", { exact: true })
      .fill(recRef);
    await page
      .getByRole("dialog")
      .getByLabel("Attachment URL", { exact: true })
      .fill(testAttachUrl);
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Reason"]')
      .fill("Urgent pediatric batch");

    const receiveI3Res = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Receive" })
      .click();
    assert.equal((await receiveI3Res).status(), 201);
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify DB persisted extended fields
    const recDb = await db.query(
      "SELECT attachment_url, supplier, store_name, reference FROM inventory_movement WHERE reference=$1",
      [recRef],
    );
    assert.equal(
      recDb.rows[0].attachment_url,
      testAttachUrl,
      "Attachment URL must be stored in database",
    );
    assert.equal(recDb.rows[0].supplier, "Apex Med Supplies");
    assert.equal(recDb.rows[0].store_name, "Emergency Sub-Store");

    // Verify UI displays details modal with attachment link
    const recRow = page.locator("tr", { hasText: recRef });
    await recRow.locator('button[aria-label="Stock Details"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    const viewAttachLink = page
      .getByRole("dialog")
      .getByRole("link", { name: "View Receipt" });
    assert.equal(
      await viewAttachLink.getAttribute("href"),
      testAttachUrl,
      "View Receipt link must match attachment URL",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Close" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // 2. Issue Item with Department, Business Dates, Issued By, and Staff Name Resolution
    await page.goto(`${base}/modules/issued-items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    await page.getByRole("button", { name: "Issue Item" }).click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(catAId);
    await page
      .getByRole("dialog")
      .getByLabel("Item", { exact: true })
      .selectOption(itemAId);
    await page
      .getByRole("dialog")
      .getByLabel("Department", { exact: true })
      .selectOption("Pediatrics Ward");
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Issue", { exact: true })
      .fill("10");
    await page
      .getByRole("dialog")
      .getByLabel("Recipient", { exact: true })
      .selectOption(staffId);
    await page
      .getByRole("dialog")
      .getByLabel("Issued Date", { exact: true })
      .fill("2026-10-09");
    await page
      .getByRole("dialog")
      .getByLabel("Return Due Date", { exact: true })
      .fill("2026-10-25");
    const issueReason = `Clinical ward loan ${i3Suffix}`;
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Reason"]')
      .fill(issueReason);

    const issueI3Res = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Issue" })
      .click();
    assert.equal((await issueI3Res).status(), 201);
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // Verify DB stored department, issued_date, return_due_date, issued_by
    const issueDb = await db.query(
      "SELECT department, issued_date, return_due_date, issued_by, recipient_id FROM inventory_movement WHERE reason=$1",
      [issueReason],
    );
    assert.equal(issueDb.rows[0].department, "Pediatrics Ward");
    assert.equal(issueDb.rows[0].issued_date, "2026-10-09");
    assert.equal(issueDb.rows[0].return_due_date, "2026-10-25");
    assert.equal(issueDb.rows[0].recipient_id, staffId);

    // Verify UI displays resolved staff display name ("Nurse Genet Lemma"), not raw staff ID
    const issueRow = page.locator("tr", { hasText: issueReason });
    await issueRow.waitFor({ timeout: 10000 });
    assert(
      (await issueRow.getByText("Nurse Genet Lemma").count()) >= 1,
      "Issued item row must display resolved staff name 'Nurse Genet Lemma'",
    );

    // Open Details Modal and verify extended business dates & department
    await issueRow.locator('button[aria-label="View issue record"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    assert(
      (await page.getByRole("dialog").getByText("Nurse Genet Lemma").count()) >=
        1,
      "Details modal must show recipient display name",
    );
    assert(
      (await page.getByRole("dialog").getByText("Pediatrics Ward").count()) >=
        1,
      "Details modal must show department",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Close" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // 3. Void Workflows:
    // 3a. Void Receipt Rejection when Stock Consumed:
    await db.query("BEGIN");
    const voidConsumeItemRes = await db.query(
      "INSERT INTO inventory_item(category_id, name, unit, reorder_milli, balance_milli) VALUES($1, $2, 'Piece', 1000, 2000) RETURNING id",
      [catAId, `Void Consume Test ${i3Suffix}`],
    );
    const voidConsumeItemId = voidConsumeItemRes.rows[0].id;
    await insertMovementFixture({
      itemId: voidConsumeItemId,
      kind: "receive",
      quantityMilli: 10000,
      deltaMilli: 10000,
      reference: `VOID-FAIL-${i3Suffix}`,
      supplier: "Supplier",
      reason: "Initial fixture receipt",
    });
    await insertMovementFixture({
      itemId: voidConsumeItemId,
      kind: "issue",
      quantityMilli: 8000,
      deltaMilli: -8000,
      recipientId: staffId,
      reason: "Issued 8 units",
    });
    await db.query("COMMIT");

    await page.goto(`${base}/modules/item-stocks`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const voidFailRow = page.locator("tr", {
      hasText: `VOID-FAIL-${i3Suffix}`,
    });
    await voidFailRow.locator('button[aria-label="Void Receipt"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Void Reason"]')
      .fill("Erroneous receipt");
    const voidFailRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Void" })
      .click();
    assert.equal(
      (await voidFailRes).status(),
      409,
      "Void receipt must be rejected with 409 when remaining balance < receipt quantity",
    );
    await page
      .getByRole("dialog")
      .getByRole("alert")
      .waitFor({ state: "visible" });
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Cancel" })
      .click();
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    // 3b. Successful Void Receipt on Unconsumed Batch:
    await db.query("BEGIN");
    const freshVoidItemRes = await db.query(
      "INSERT INTO inventory_item(category_id, name, unit, reorder_milli, balance_milli) VALUES($1, $2, 'Piece', 1000, 15000) RETURNING id",
      [catAId, `Fresh Void Test ${i3Suffix}`],
    );
    const freshVoidItemId = freshVoidItemRes.rows[0].id;
    await insertMovementFixture({
      itemId: freshVoidItemId,
      kind: "receive",
      quantityMilli: 15000,
      deltaMilli: 15000,
      reference: `VOID-SUCCEED-${i3Suffix}`,
      supplier: "Supplier",
      reason: "Initial fixture receipt",
    });
    await db.query("COMMIT");

    await page.reload();
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const voidSucceedRow = page.locator("tr", {
      hasText: `VOID-SUCCEED-${i3Suffix}`,
    });
    await voidSucceedRow.locator('button[aria-label="Void Receipt"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Void Reason"]')
      .fill("Vendor shipment cancelled");
    const voidSucceedRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Void" })
      .click();
    assert.equal(
      (await voidSucceedRes).status(),
      201,
      "Void receipt succeeds for unconsumed stock",
    );
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    const balAfterVoid = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [freshVoidItemId],
    );
    assert.equal(
      Number(balAfterVoid.rows[0].balance_milli),
      0,
      "Balance must be reduced to 0 by void writeoff",
    );
    const writeoffQuery = await db.query(
      "SELECT id, kind, delta_milli FROM inventory_movement WHERE item_id=$1 AND kind='writeoff'",
      [freshVoidItemId],
    );
    assert.equal(
      writeoffQuery.rowCount,
      1,
      "Audited writeoff movement recorded",
    );
    assert.equal(
      Number(writeoffQuery.rows[0].delta_milli),
      -15000,
      "Writeoff delta is -15,000 milli",
    );

    // 3c. Successful Void Issue:
    await page.goto(`${base}/modules/issued-items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");
    const activeIssueRow = page.locator("tr", { hasText: issueReason });
    await activeIssueRow.locator('button[aria-label="Void Issue"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Void Reason"]')
      .fill("Order cancelled by head nurse");
    const voidIssueRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Void" })
      .click();
    assert.equal((await voidIssueRes).status(), 201, "Void issue succeeds");
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    const balItemAfterIssueVoid = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [itemAId],
    );
    assert.equal(
      Number(balItemAfterIssueVoid.rows[0].balance_milli),
      70000,
      "Balance must be 70,000 milli after void issue reversal",
    );
    console.log(
      "✔ Journey 22 Passed: Source fields, receipt attachments, staff name resolution, and void/reversal workflows verified.",
    );

    // -----------------------------------------------------------------------
    // Journey 23: Finding I4 Regression — Exact Fractional & Zero Reorder Thresholds
    // -----------------------------------------------------------------------
    console.log(
      "\n[Journey 23] Finding I4: Fractional & Zero Reorder Thresholds...",
    );
    const i4Suffix = randomBytes(4).toString("hex");
    const fracItemName = `Fractional Scalpel ${i4Suffix}`;
    const zeroItemName = `Zero Threshold Syringe ${i4Suffix}`;

    await db.query(
      "INSERT INTO inventory_item(category_id, name, unit, reorder_milli, balance_milli, description) VALUES($1, $2, 'Piece', 1500, 0, 'Original fractional description')",
      [catAId, fracItemName],
    );
    await db.query(
      "INSERT INTO inventory_item(category_id, name, unit, reorder_milli, balance_milli, description) VALUES($1, $2, 'Piece', 0, 0, 'Original zero description')",
      [catAId, zeroItemName],
    );

    await page.goto(`${base}/modules/items`);
    await page.waitForSelector(".legacy-workspace[data-ready='true']");

    const thresholdSearchInput = page.locator(
      'input[aria-label="Search Items"]',
    );
    await thresholdSearchInput.fill(fracItemName);

    // 1. Edit Fractional Item: only change description
    const fracRow = page.locator("tr", { hasText: fracItemName });
    await fracRow.locator('button[aria-label="Edit Item"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    const fracReorderVal = await page
      .getByRole("dialog")
      .getByLabel("Reorder Level", { exact: true })
      .inputValue();
    assert.equal(
      fracReorderVal,
      "1.5",
      "Reorder level input must display exact 1.5 without integer rounding",
    );

    const updatedFracDesc =
      "Updated description for fractional item without threshold rounding";
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Description"]')
      .fill(updatedFracDesc);
    const fracEditRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/items") &&
        r.request().method() === "PATCH",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: /Update Item|Save Item/ })
      .click();
    assert.equal((await fracEditRes).status(), 200);
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    const fracDb = await db.query(
      "SELECT reorder_milli, description FROM inventory_item WHERE name=$1",
      [fracItemName],
    );
    assert.equal(
      Number(fracDb.rows[0].reorder_milli),
      1500,
      "Reorder level in DB must remain exactly 1500 milli (1.5 units)",
    );
    assert.equal(fracDb.rows[0].description, updatedFracDesc);

    // 2. Edit Zero Threshold Item: only change description
    await thresholdSearchInput.fill(zeroItemName);
    const zeroRow = page.locator("tr", { hasText: zeroItemName });
    await zeroRow.locator('button[aria-label="Edit Item"]').click();
    await page.getByRole("dialog").waitFor({ state: "visible" });

    const zeroReorderVal = await page
      .getByRole("dialog")
      .getByLabel("Reorder Level", { exact: true })
      .inputValue();
    assert.equal(
      zeroReorderVal,
      "0",
      "Reorder level input must display exact 0",
    );

    const updatedZeroDesc =
      "Updated description for zero threshold item without rejection";
    await page
      .getByRole("dialog")
      .locator('textarea[aria-label="Description"]')
      .fill(updatedZeroDesc);
    const zeroEditRes = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/items") &&
        r.request().method() === "PATCH",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: /Update Item|Save Item/ })
      .click();
    assert.equal((await zeroEditRes).status(), 200);
    await page.getByRole("dialog").waitFor({ state: "hidden" });

    const zeroDb = await db.query(
      "SELECT reorder_milli, description FROM inventory_item WHERE name=$1",
      [zeroItemName],
    );
    assert.equal(
      Number(zeroDb.rows[0].reorder_milli),
      0,
      "Reorder level in DB must remain exactly 0 milli (0 units)",
    );
    assert.equal(zeroDb.rows[0].description, updatedZeroDesc);

    console.log(
      "✔ Journey 23 Passed: Fractional (1500 milli) and zero (0 milli) reorder thresholds preserved without rounding drift.",
    );

    console.log("\n=======================================================");
    console.log("ALL 23 INVENTORY WORKSPACE JOURNEYS PASSED SUCCESSFULLY!");
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
