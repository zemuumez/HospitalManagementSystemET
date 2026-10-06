// Page/transport audit of every source-catalog screen. This is NOT CRUD parity acceptance.
import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
import { chromium } from "@playwright/test";

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
    email = `pages-${id}@example.test`;
  await db.query(
    'INSERT INTO "user"(id,name,email,"emailVerified") VALUES($1,$2,$3,true)',
    [id, "Synthetic module page auditor", email],
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
  const cookies = login.headers.getSetCookie().map((c) => {
    const pair = c.split(";")[0];
    const eq = pair.indexOf("=");
    return { name: pair.slice(0, eq), value: pair.slice(eq + 1), url: base };
  });
  browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
  });
  const context = await browser.newContext();
  await context.addCookies(cookies);
  const catalog = JSON.parse(
    readFileSync("apps/web/src/lib/legacy-catalog.json", "utf8"),
  );
  const paths = ["/dashboard", ...catalog.map((s) => `/modules/${s.id}`)];
  for (const path of paths) {
    const page = await context.newPage();
    const row = {
      path,
      status: "pass",
      navigationStatus: null,
      headings: [],
      failedApi: [],
      pageErrors: [],
    };
    page.on("pageerror", (e) => row.pageErrors.push(e.message.slice(0, 400)));
    page.on("response", (r) => {
      if (r.url().includes("/api/hms/") && r.status() >= 400)
        row.failedApi.push({
          path: new URL(r.url()).pathname,
          status: r.status(),
        });
    });
    try {
      const response = await page.goto(base + path, {
        waitUntil: "networkidle",
        timeout: 45000,
      });
      row.navigationStatus = response?.status();
      await page
        .locator('[data-ready="true"]')
        .first()
        .waitFor({ timeout: 20000 });
      // Record initial render/API errors; no data mutation and no invented CRUD coverage.
      row.headings = await page.locator("h1,h2").allTextContents();
      row.visibleAlerts = await page
        .locator('[role="alert"]')
        .allTextContents();
      const body = await page.locator("body").innerText();
      assert.ok(
        !new URL(page.url()).pathname.startsWith("/login"),
        "redirected to login",
      );
      assert.equal(row.navigationStatus, 200);
      assert.ok(
        !/Application error:|This page could not be found/.test(body),
        "error page",
      );
      assert.equal(row.pageErrors.length, 0, "browser runtime error");
      if (row.failedApi.length) row.status = "api-failure";
    } catch (e) {
      row.status = "fail";
      row.error = String(e.message).slice(0, 500);
    }
    results.push(row);
    mkdirSync(".local/module-audit", { recursive: true });
    writeFileSync(
      ".local/module-audit/page-progress.json",
      JSON.stringify(results, null, 2),
    );
    await page.close();
    console.log(`${results.length}/${paths.length} ${path}: ${row.status}`);
    // Each navigation resolves the actual session. Stay below the real limiter.
    await new Promise((resolve) => setTimeout(resolve, 800));
  }
} finally {
  await browser?.close();
  await db.end();
  mkdirSync(".local/module-audit", { recursive: true });
  writeFileSync(
    ".local/module-audit/page-results.json",
    JSON.stringify(
      {
        date: new Date().toISOString(),
        scope:
          "Admin page navigation/initial API transport only; not all-role, CRUD or visual parity",
        results,
      },
      null,
      2,
    ),
  );
  console.log(
    JSON.stringify({
      pages: results.length,
      passed: results.filter((r) => r.status === "pass").length,
      apiFailures: results.filter((r) => r.status === "api-failure").length,
      failed: results.filter((r) => r.status === "fail").length,
    }),
  );
}
if (results.some((r) => r.status !== "pass")) process.exitCode = 1;
