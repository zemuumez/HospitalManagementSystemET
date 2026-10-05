import fs from "node:fs";
import { chromium } from "@playwright/test";

(async () => {
  const lines = fs.readFileSync(".local/admin-login.txt", "utf8").split("\n");
  const email = lines
    .find((l) => l.startsWith("Email:"))
    ?.replace("Email:", "")
    .trim();
  const password = lines
    .find((l) => l.startsWith("Password:"))
    ?.replace("Password:", "")
    .trim();

  if (!email || !password) throw Error("Admin credentials missing");

  const b = await chromium.launch({
    headless: true,
    executablePath:
      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
  });
  const p = await b.newPage({ viewport: { width: 1440, height: 1000 } });

  const errors = [];
  p.on("pageerror", (e) => errors.push(e.message));

  // 1. Sign in
  await p.goto("http://127.0.0.1:3000/login");
  await p.getByLabel("Email address", { exact: false }).fill(email);
  await p.getByLabel("Password", { exact: false }).fill(password);
  await p.getByRole("button", { name: "Sign In", exact: true }).click();
  await p.waitForURL("**/dashboard");
  await p.getByRole("button", { name: "Use dark theme", exact: true }).click();
  console.log("1. Logged in and set dark theme");

  // 2. Bed Status
  await p.goto("http://127.0.0.1:3000/modules/bed-status");
  await p.locator('[data-ready="true"]').waitFor();
  await p.locator(".legacy-dark").waitFor();
  const wardCount = await p.locator(".ward-section-card").count();
  console.log(`2. Bed status loaded: ${wardCount} wards found`);
  if (wardCount < 10) throw Error(`Expected >= 10 wards, got ${wardCount}`);
  await p.screenshot({
    path: ".local/reference-repair/bed-status.png",
    fullPage: true,
  });

  // Test popover hover on occupied bed
  const occupiedBed = p.locator(".bed-status-item.occupied").first();
  await occupiedBed.hover();
  await p.locator(".bed-hover-popover").waitFor();
  console.log("2b. Occupied bed hover popover verified");

  // 3. Bed Assigns
  await p.goto("http://127.0.0.1:3000/modules/bed-assigns");
  await p.locator('[data-ready="true"]').waitFor();
  const assignsRows = await p.locator(".billing-table tbody tr").count();
  console.log(`3. Bed assigns loaded: ${assignsRows} rows`);
  if (assignsRows === 0) throw Error("Bed assigns table is empty");

  // Search filter check
  await p.locator(".billing-search-box input").fill("Vinay");
  if ((await p.locator(".billing-table tbody tr").count()) !== 1) {
    throw Error("Bed assigns search filter did not match Vinay");
  }
  await p.locator(".billing-search-box input").fill("");

  // Toggle switch check
  const toggleInput = p.locator(".switch-toggle input").first();
  const initialChecked = await toggleInput.isChecked();
  await p.locator(".switch-toggle").first().click();
  if ((await toggleInput.isChecked()) === initialChecked) {
    throw Error("Toggle switch did not change state");
  }

  // Open & Cancel New Bed Assign modal
  await p.getByRole("button", { name: "New Bed Assign", exact: true }).click();
  await p.locator("#bed-assign-modal-title").waitFor();
  await p.getByRole("button", { name: "Cancel", exact: true }).click();
  await p.screenshot({
    path: ".local/reference-repair/bed-assigns.png",
    fullPage: true,
  });

  // 4. Beds
  await p.goto("http://127.0.0.1:3000/modules/beds");
  await p.locator('[data-ready="true"]').waitFor();
  const bedsRows = await p.locator(".billing-table tbody tr").count();
  console.log(`4. Beds loaded: ${bedsRows} rows`);
  if (bedsRows === 0) throw Error("Beds table is empty");

  // Dropdown check
  await p.locator(".dropdown-trigger").click();
  await p.locator(".dropdown-action-menu").waitFor();
  await p.getByRole("button", { name: "New Bed", exact: true }).click();
  await p.locator("#bed-modal-title").waitFor();
  await p.getByRole("button", { name: "Cancel", exact: true }).click();
  await p.screenshot({
    path: ".local/reference-repair/beds.png",
    fullPage: true,
  });

  // 5. Bed Types
  await p.goto("http://127.0.0.1:3000/modules/bed-types");
  await p.locator('[data-ready="true"]').waitFor();
  await p.locator(".legacy-dark").waitFor();
  const bedTypesRows = await p.locator(".billing-table tbody tr").count();
  console.log(`5. Bed types loaded: ${bedTypesRows} rows`);
  if (bedTypesRows === 0) throw Error("Bed types table is empty");

  await p.screenshot({
    path: ".local/reference-repair/bed-types.png",
    fullPage: true,
  });

  // 6. New Bed Type Modal (matches Screenshot 10)
  await p.getByRole("button", { name: "New Bed Type", exact: true }).click();
  await p.locator("#bed-type-modal-title").waitFor();
  await p.locator('input[placeholder="Bed Type"]').fill("QA Neonatal Unit");
  await p
    .locator('textarea[placeholder="Description"]')
    .fill("Automated QA test bed type description");
  await p.screenshot({
    path: ".local/reference-repair/bed-types-modal.png",
    fullPage: true,
  });
  await p.getByRole("button", { name: "Save", exact: true }).click();

  // Verify added
  await p
    .locator(".bed-type-title-link", { hasText: "QA Neonatal Unit" })
    .waitFor();
  console.log("6. New Bed Type added and rendered in table!");

  if (errors.length) {
    throw Error(`Browser page errors encountered:\n${errors.join("\n")}`);
  }

  await b.close();
  console.log("SUCCESS: All Bed Management tabs and modals verified cleanly!");
})();
