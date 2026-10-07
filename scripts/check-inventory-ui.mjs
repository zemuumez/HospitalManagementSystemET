import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";

export async function checkInventoryUI({
  context,
  db,
  base,
  recipientId,
  results,
}) {
  const page = await context.newPage();
  const suffix = randomUUID().slice(0, 8),
    category = `QA category ${suffix}`,
    item = `QA supply ${suffix}`;
  const passed = (name) =>
    results.push({ name: `Inventory: ${name}`, status: "pass" });
  async function open(slug) {
    await page.goto(`${base}/modules/${slug}`, {
      waitUntil: "domcontentloaded",
    });
    await page.getByText("Inventory loaded", { exact: true }).waitFor();
  }
  async function submit(name) {
    const response = page.waitForResponse(
      (r) =>
        r.url().includes("/api/hms/inventory/") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name, exact: true })
      .click();
    const saved = await response;
    assert.equal(saved.status(), 201, await saved.text());
    await page.getByRole("dialog").waitFor({ state: "hidden" });
  }
  try {
    await open("item-categories");
    assert.equal(
      await page
        .getByRole("cell", { name: "Surgical Supplies", exact: true })
        .count(),
      0,
      "empty database must not show fixtures",
    );
    await page
      .getByRole("button", { name: "New Item Category", exact: true })
      .click();
    await page
      .getByRole("dialog")
      .getByLabel("Category Name", { exact: true })
      .fill(category);
    await submit("Save Category");
    await page.reload({ waitUntil: "domcontentloaded" });
    await page.getByRole("cell", { name: category, exact: true }).waitFor();
    const cat = (
      await db.query("SELECT id FROM inventory_category WHERE name=$1", [
        category,
      ])
    ).rows;
    assert.equal(cat.length, 1);
    passed("category created once and survives reload");

    await open("items");
    await page.getByRole("button", { name: "New Item", exact: true }).click();
    await page
      .getByRole("dialog")
      .getByLabel("Item Name", { exact: true })
      .fill(item);
    await page
      .getByRole("dialog")
      .getByLabel("Category", { exact: true })
      .selectOption(cat[0].id);
    await submit("Save Item");
    const stock = (
      await db.query(
        "SELECT id,balance_milli FROM inventory_item WHERE name=$1",
        [item],
      )
    ).rows;
    assert.equal(stock.length, 1);
    assert.equal(Number(stock[0].balance_milli), 0);
    passed("item is persisted without fabricated opening stock");

    await open("item-stocks");
    await page
      .getByRole("button", { name: "Receive New Stock", exact: true })
      .click();
    await page
      .getByRole("dialog")
      .getByLabel("Quantity", { exact: true })
      .fill("10");
    await page
      .getByRole("dialog")
      .getByLabel("Reference", { exact: true })
      .fill(`QA-PO-${suffix}`);
    await submit("Confirm Receive");
    await open("issued-items");
    await page.getByRole("button", { name: "Issue Item", exact: true }).click();
    await page
      .getByRole("dialog")
      .getByLabel("Recipient", { exact: true })
      .selectOption(recipientId);
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Issue", { exact: true })
      .fill("3");
    await submit("Confirm Issue");
    let balance = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [stock[0].id],
    );
    assert.equal(Number(balance.rows[0].balance_milli), 7000);
    const movements = await db.query(
      "SELECT kind,quantity_milli FROM inventory_movement WHERE item_id=$1 ORDER BY created_at",
      [stock[0].id],
    );
    assert.deepEqual(
      movements.rows.map((x) => [x.kind, Number(x.quantity_milli)]),
      [
        ["receive", 10000],
        ["issue", 3000],
      ],
    );
    await page.reload({ waitUntil: "domcontentloaded" });
    await page.getByRole("cell", { name: item, exact: true }).waitFor();
    passed(
      "receive and staff issue persist immutable movements and reconcile balance",
    );

    await page.getByRole("button", { name: "Issue Item", exact: true }).click();
    await page
      .getByRole("dialog")
      .getByLabel("Recipient", { exact: true })
      .selectOption(recipientId);
    await page
      .getByRole("dialog")
      .getByLabel("Quantity to Issue", { exact: true })
      .fill("999");
    const rejected = page.waitForResponse(
      (r) =>
        r.url().includes("/inventory/movements") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Confirm Issue", exact: true })
      .click();
    assert.equal((await rejected).status(), 409);
    await page.getByRole("dialog").getByRole("alert").waitFor();
    balance = await db.query(
      "SELECT balance_milli FROM inventory_item WHERE id=$1",
      [stock[0].id],
    );
    assert.equal(Number(balance.rows[0].balance_milli), 7000);
    passed("over-issue keeps editor open and does not change balance");

    await open("item-categories");
    await page
      .getByRole("button", { name: "New Item Category", exact: true })
      .click();
    const rejectedName = `Rejected ${suffix}`;
    await page
      .getByRole("dialog")
      .getByLabel("Category Name", { exact: true })
      .fill(rejectedName);
    await page.route("**/api/hms/inventory/categories", (route) =>
      route.request().method() === "POST"
        ? route.fulfill({
            status: 422,
            contentType: "application/json",
            body: JSON.stringify({ error: "Synthetic validation failure" }),
          })
        : route.continue(),
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save Category", exact: true })
      .click();
    await page
      .getByRole("dialog")
      .getByRole("alert")
      .filter({ hasText: "Synthetic validation failure" })
      .waitFor();
    assert.equal(
      await page.getByRole("cell", { name: rejectedName, exact: true }).count(),
      0,
    );
    assert.equal(
      (
        await db.query("SELECT id FROM inventory_category WHERE name=$1", [
          rejectedName,
        ])
      ).rowCount,
      0,
    );
    passed("HTTP failure preserves form and never creates a local success row");
    await page.unroute("**/api/hms/inventory/categories");
    await page.route("**/api/hms/inventory/categories", (route) =>
      route.request().method() === "POST" ? route.abort() : route.continue(),
    );
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Save Category", exact: true })
      .click();
    await page
      .getByRole("dialog")
      .getByRole("alert")
      .filter({ hasNotText: "Synthetic validation failure" })
      .waitFor();
    assert.equal(
      (
        await db.query("SELECT id FROM inventory_category WHERE name=$1", [
          rejectedName,
        ])
      ).rowCount,
      0,
    );
    passed("network failure preserves form and does not fabricate persistence");
  } finally {
    await page.close();
  }
}
