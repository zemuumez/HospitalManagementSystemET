import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { mkdirSync, writeFileSync } from "node:fs";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
import { chromium } from "@playwright/test";
import { checkInventoryUI } from "./check-inventory-ui.mjs";

assert.match(
  process.env.HMS_TEST_ISOLATED_SCHEMA || "",
  /^hms_browser_[a-f0-9]{24}$/,
);
const base = process.env.BETTER_AUTH_URL;
assert.ok(["127.0.0.1", "localhost"].includes(new URL(base).hostname));
const db = new Pool({ connectionString: process.env.DATABASE_URL });
const results = [];
let browser;
try {
  const id = randomUUID(),
    password = randomUUID() + "Aa1!",
    email = `repair-${id}@example.test`;
  await db.query(
    'INSERT INTO "user"(id,name,email,"emailVerified") VALUES($1,$2,$3,true)',
    [id, "Synthetic endpoint auditor", email],
  );
  await db.query(
    'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
    [randomUUID(), id, await hashPassword(password)],
  );
  await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,'admin')", [
    id,
  ]);
  const login = await fetch(`${base}/api/auth/sign-in/email`, {
    method: "POST",
    headers: { origin: base, "content-type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  assert.equal(login.status, 200);
  const pairs = login.headers.getSetCookie().map((x) => x.split(";")[0]);
  const cookie = pairs.join("; ");
  async function request(path, method = "GET", body, authenticated = true) {
    await new Promise((resolve) => setTimeout(resolve, 700));
    return fetch(`${base}/api/hms/${path}`, {
      method,
      headers: {
        cookie: authenticated ? cookie : "",
        origin: base,
        "content-type": "application/json",
      },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  }
  async function status(path, role, expected, method = "GET", body) {
    const response = await request(path, method, body);
    assert.equal(response.status, expected, `${role} ${method} ${path}`);
    results.push({
      name: `${role} ${method} ${path}`,
      status: "pass",
      httpStatus: response.status,
    });
    return response;
  }
  for (const role of [
    "admin",
    "doctor",
    "patient",
    "nurse",
    "receptionist",
    "pharmacist",
    "accountant",
    "case_manager",
    "lab_technician",
  ]) {
    // Go must resolve current role on every request, not trust a cached client role.
    await db.query("UPDATE staff_access SET role=$2 WHERE user_id=$1", [
      id,
      role,
    ]);
    await status("doctor-absences", role, role === "admin" ? 200 : 403);
    await status("inventory/movements", role, role === "admin" ? 200 : 403);
    await status("general-settings", role, role === "patient" ? 403 : 200);
    await status(
      "general-settings",
      role,
      role === "admin" ? 200 : 403,
      "PUT",
      { audit_name: "Verified" },
    );
  }
  await db.query("UPDATE staff_access SET role='admin' WHERE user_id=$1", [id]);
  for (const path of [
    "general-settings",
    "doctor-absences",
    "inventory/movements",
  ]) {
    assert.equal((await request(path, "GET", undefined, false)).status, 401);
    results.push({ name: `anonymous denied ${path}`, status: "pass" });
  }
  await status("general-settings", "admin", 422, "PUT", {
    audit_name: "Partial",
    "": "Invalid",
  });
  const settings = await (await request("general-settings")).json();
  assert.equal(settings.general_settings.audit_name, "Verified");
  results.push({
    name: "settings rejected form preserves persisted values",
    status: "pass",
  });
  await status("inventory/movements?itemId=invalid", "admin", 422);
  await status("doctor-absences?page=0", "admin", 422);
  browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
  });
  const context = await browser.newContext();
  await context.addCookies(
    pairs.map((pair) => ({
      name: pair.slice(0, pair.indexOf("=")),
      value: pair.slice(pair.indexOf("=") + 1),
      url: base,
    })),
  );
  for (const slug of [
    "doctors",
    "doctor-departments",
    "items",
    "item-categories",
    "item-stocks",
    "issued-items",
  ]) {
    const page = await context.newPage();
    const failures = [];
    page.on("response", (r) => {
      if (r.url().includes("/api/hms/") && r.status() >= 400)
        failures.push(`${r.status()} ${new URL(r.url()).pathname}`);
    });
    const expected = slug.startsWith("doctor")
      ? "doctor-absences"
      : "inventory/movements";
    const loaded = page.waitForResponse(
      (r) => new URL(r.url()).pathname === `/api/hms/${expected}`,
    );
    await page.goto(`${base}/modules/${slug}`, {
      waitUntil: "domcontentloaded",
    });
    assert.equal((await loaded).status(), 200);
    await page.locator('[data-ready="true"]').first().waitFor();
    assert.deepEqual(failures, [], slug);
    results.push({
      name: `${slug} authenticated initial API transport`,
      status: "pass",
    });
    await page.close();
  }
  await checkInventoryUI({ context, db, base, recipientId: id, results });
  console.log(
    `Endpoint and inventory repairs: ${results.length} checks passed.`,
  );
} finally {
  await browser?.close();
  await db.end();
  mkdirSync(".local/module-audit", { recursive: true });
  writeFileSync(
    ".local/module-audit/endpoint-repairs.json",
    JSON.stringify({ date: new Date().toISOString(), results }, null, 2),
  );
}
