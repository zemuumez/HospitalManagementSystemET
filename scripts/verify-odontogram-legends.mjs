import { chromium } from "@playwright/test";
import assert from "node:assert/strict";
import fs from "node:fs";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
import { randomUUID } from "node:crypto";

async function main() {
  console.log("Starting Odontogram Legends Playwright test...");

  // 1. Connect to PostgreSQL to create a test admin user for login
  const dbUrl =
    process.env.DATABASE_URL ||
    "postgresql://hms:424d48acc68f829bc0cd4f7b1f07f2328b706dc863002e63@127.0.0.1:5433/hms";
  console.log("Connecting to PostgreSQL at:", dbUrl.split("@")[1]);
  const db = new Pool({ connectionString: dbUrl });

  const userId = randomUUID();
  const testEmail = `test-legend-${userId.slice(0, 8)}@example.com`;
  const testPassword = `Passw0rd!${userId.slice(0, 6)}`;
  const hash = await hashPassword(testPassword);

  try {
    await db.query(
      'INSERT INTO "user"(id, name, email, "emailVerified") VALUES($1, $2, $3, true)',
      [userId, "Odontogram Tester", testEmail],
    );
    await db.query(
      'INSERT INTO account(id, "accountId", "providerId", "userId", password) VALUES($1, $2, \'credential\', $2, $3)',
      [randomUUID(), userId, hash],
    );
    await db.query(
      "INSERT INTO staff_access(user_id, role) VALUES($1, 'admin')",
      [userId],
    );
    console.log("Created test admin user:", testEmail);
  } catch (err) {
    console.error("DB setup note:", err.message);
  }

  // 2. Locate Chrome/Edge executable
  const chromePaths = [
    process.env.HMS_CHROME_PATH,
    "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
    "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
    "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
    "C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe",
  ].filter(Boolean);

  let executablePath;
  for (const cp of chromePaths) {
    if (fs.existsSync(cp)) {
      executablePath = cp;
      break;
    }
  }
  console.log("Using browser at:", executablePath || "bundled chromium");

  const browser = await chromium.launch({
    headless: true,
    ...(executablePath ? { executablePath } : {}),
  });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  try {
    // 3. Login
    console.log("Navigating to http://127.0.0.1:3000/login");
    await page.goto("http://127.0.0.1:3000/login");
    await page.getByLabel("Email address", { exact: false }).fill(testEmail);
    await page.getByLabel("Password", { exact: false }).fill(testPassword);
    await page.getByRole("button", { name: "Sign In", exact: true }).click();
    await page.waitForURL("**/dashboard");
    console.log("Logged in successfully, arrived at dashboard!");

    // 4. Visit odontogram page
    console.log("Navigating to http://127.0.0.1:3000/modules/odontogram");
    await page.goto("http://127.0.0.1:3000/modules/odontogram");
    await page.waitForSelector(".legacy-table");
    console.log("Odontogram table is visible!");

    // 5. Open Manage Legends modal
    console.log("Opening Odontogram Legends modal...");
    const legendsBtn = page.getByRole("button", { name: "Odontogram Legends" });
    await legendsBtn.click();
    await page.waitForSelector("#manage-legends-title");
    console.log("Odontogram Legends Management modal opened!");

    // 6. Click New Legend inside modal
    console.log("Clicking New Legend button...");
    const newLegendBtn = page
      .getByRole("button", { name: "New Legend" })
      .first();
    await newLegendBtn.click();
    await page.waitForSelector("#legend-form-title");
    console.log("Create New Odontogram Legend modal opened!");

    // 7. Fill in legend details
    console.log("Filling in New Legend form...");
    await page.getByPlaceholder("e.g. F, V, RCT, IMP").fill("V");
    await page
      .getByPlaceholder("e.g. Composite Restoration, Ceramic Veneer")
      .fill("Porcelain Veneer");
    await page.locator("input[type='color']").fill("#8b5cf6");
    await page
      .getByPlaceholder("Optional clinical notes")
      .fill("Aesthetic anterior veneer");

    // 8. Save legend
    console.log("Saving new legend...");
    await page.getByRole("button", { name: "Save Legend" }).click();

    // 9. Verify it appears in the legends list
    console.log(
      "Checking that Porcelain Veneer appears in the legends table...",
    );
    await page.waitForSelector("td:has-text('Porcelain Veneer')");
    console.log("PASS: Porcelain Veneer found in legends list!");

    // 10. Close legends modal
    await page.getByRole("button", { name: "Close" }).last().click();

    // 11. Verify quick legends key on main page contains V
    const legendKeyText = await page.textContent(".form-card-container");
    assert.ok(
      legendKeyText.includes("V:"),
      "Main page legend strip should include new V legend",
    );
    console.log(
      "PASS: Main page quick legend key contains V: Porcelain Veneer!",
    );

    // 12. Open Alex Morgan's chart for editing
    console.log("Opening Alex Morgan odontogram editor...");
    await page
      .getByRole("button", { name: "Edit odontogram Alex Morgan", exact: true })
      .click();
    await page.waitForSelector("#odontogram-title");

    // 13. Verify the new legend V button exists in the palette!
    console.log("Checking palette button for V tooth 1...");
    const vBtn = page.getByRole("button", { name: "V tooth 1", exact: true });
    await vBtn.click();

    // 14. Save the chart
    console.log("Saving Alex Morgan's chart...");
    await page.getByRole("button", { name: "Save", exact: true }).click();

    // 15. Re-open Alex Morgan's chart to verify persistence
    console.log("Re-opening Alex Morgan's chart to verify fill color...");
    await page
      .getByRole("button", { name: "Edit odontogram Alex Morgan", exact: true })
      .click();
    await page.waitForSelector("#odontogram-title");

    const tooth1Fill = await page.locator("#Tooth1").getAttribute("fill");
    console.log("Tooth1 fill attribute is:", tooth1Fill);
    assert.equal(
      tooth1Fill?.toLowerCase(),
      "#8b5cf6",
      "Tooth1 should have the custom legend color #8b5cf6",
    );
    console.log(
      "PASS: Custom legend applied to tooth and persisted correctly!",
    );

    // 16. Also test default K condition for regression check
    const kBtn = page.getByRole("button", { name: "K tooth 1", exact: true });
    await kBtn.click();
    await page.getByRole("button", { name: "Save", exact: true }).click();

    await page
      .getByRole("button", { name: "Edit odontogram Alex Morgan", exact: true })
      .click();
    const tooth1FillK = await page.locator("#Tooth1").getAttribute("fill");
    console.log("Tooth1 fill with K is:", tooth1FillK);
    assert.equal(
      tooth1FillK?.toLowerCase(),
      "#e91e63",
      "Tooth1 with K should have #e91e63",
    );
    console.log("PASS: Standard K legend test works as expected!");

    await page.getByRole("button", { name: "Cancel", exact: true }).click();

    // 17. Test editing an existing legend (change name or notes)
    console.log("Testing editing of legend...");
    await legendsBtn.click();
    await page.waitForSelector("#manage-legends-title");
    const editVBtn = page.getByRole("button", {
      name: "Edit legend Porcelain Veneer",
    });
    await editVBtn.click();
    await page.waitForSelector("#legend-form-title");
    await page
      .getByPlaceholder("e.g. Composite Restoration, Ceramic Veneer")
      .fill("Laminate Veneer");
    await page.getByRole("button", { name: "Save Legend" }).click();
    await page.waitForSelector("td:has-text('Laminate Veneer')");
    console.log("PASS: Edited legend name persisted!");

    // 18. Reset to defaults
    console.log("Testing Reset Defaults...");
    page.once("dialog", (dialog) => dialog.accept());
    await page.getByRole("button", { name: "Reset Defaults" }).click();
    await page.waitForTimeout(300);
    await page.getByRole("button", { name: "Close" }).last().click();

    console.log("ALL ODONTOGRAM LEGEND TESTS PASSED SUCCESSFULLY! 🎉");
  } catch (err) {
    console.error("TEST FAILED:", err);
    process.exit(1);
  } finally {
    await browser.close();
    try {
      // Clean up test user
      await db.query('DELETE FROM account WHERE "userId" = $1', [userId]);
      await db.query("DELETE FROM staff_access WHERE user_id = $1", [userId]);
      await db.query('DELETE FROM "user" WHERE id = $1', [userId]);
      await db.end();
    } catch {}
  }
}

main();
