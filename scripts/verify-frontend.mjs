import { chromium } from "@playwright/test";
import { readFileSync, mkdirSync, writeFileSync } from "node:fs";
import assert from "node:assert/strict";

// Only synthetic browser previews are mutated. Never run against a deployed hospital.
const base = process.env.HMS_BASE_URL || "http://127.0.0.1:3000";
if (!["127.0.0.1", "localhost", "[::1]"].includes(new URL(base).hostname))
  throw new Error("Frontend verification requires a loopback development URL.");
const email = process.env.HMS_TEST_EMAIL;
const password = process.env.HMS_TEST_PASSWORD;
if (!email || !password)
  throw new Error("Set HMS_TEST_EMAIL and HMS_TEST_PASSWORD.");
const onlyModules = (process.env.HMS_FRONTEND_MODULES || "")
  .split(",")
  .filter(Boolean);
const output = onlyModules.length
  ? ".local/frontend-verification-targeted"
  : ".local/frontend-verification";
mkdirSync(output, { recursive: true });
const browser = await chromium.launch({
  headless: true,
  ...(process.env.HMS_CHROME_PATH
    ? { executablePath: process.env.HMS_CHROME_PATH }
    : {}),
});
const context = await browser.newContext({
  viewport: { width: 1440, height: 1000 },
  acceptDownloads: true,
});
// Guard preview QA from ever submitting new persistent workflows.
for (const pattern of ["**/api/hms/**", "**/api/staff**"])
  await context.route(pattern, (route) =>
    route.request().method() === "GET"
      ? route.continue()
      : route.abort("blockedbyclient"),
  );
const connected = new Set([
  "users",
  "patients",
  "appointments",
  "schedules",
  "beds",
  "bed-status",
  "patient-cases",
  "ipd-patient-departments",
  "opd-patient-departments",
  "accounts",
  "invoices",
]);
const connectedPortals = new Set([
  "/portal/appointments",
  "/portal/ipd",
  "/portal/opd",
  "/portal/cases",
  "/portal/invoices",
]);
const page = await context.newPage();
page.setDefaultTimeout(12000);
const failures = [],
  passed = [],
  pageErrors = [];
page.on("pageerror", (error) => pageErrors.push(error.message));
async function check(name, fn) {
  try {
    await fn();
    passed.push(name);
  } catch (error) {
    failures.push({ name, error: error.message });
    console.error("FAIL", name, error.message.slice(0, 220));
    await page
      .screenshot({
        path: output + "/failure-" + name.replace(/[^a-z0-9]/gi, "_") + ".png",
        fullPage: true,
      })
      .catch(() => {});
    console.error(
      await page
        .locator("input:invalid,select:invalid,textarea:invalid")
        .evaluateAll((items) =>
          items.map((el) => ({
            label: el.labels?.[0]?.textContent?.trim(),
            type: el.type,
            message: el.validationMessage,
          })),
        ),
    );
  }
}
async function visit(path) {
  const response = await page.goto(base + path, {
    waitUntil: "domcontentloaded",
  });
  assert.equal(response.status(), 200, path);
  await page.locator("body").waitFor();
  await page.waitForFunction(
    () =>
      document.querySelector(".public-site") ||
      document.querySelector(".legacy-login") ||
      document.querySelector('.legacy-shell[data-ready="true"]'),
  );
}
async function fillRequired(form) {
  const controls = form.locator("input,select,textarea");
  for (let i = 0; i < (await controls.count()); i++) {
    const input = controls.nth(i);
    const info = await input.evaluate((el) => ({
      required: el.required,
      type: el.type,
      tag: el.tagName,
      value: el.value,
      disabled: el.disabled,
    }));
    if (!info.required || info.disabled) continue;
    if (info.type === "radio") {
      if (!(await form.locator('input[type="radio"]:checked').count()))
        await input.check();
      continue;
    }
    if (info.type === "checkbox") {
      await input.check();
      continue;
    }
    if (info.value) continue;
    if (info.tag === "SELECT") {
      const value = await input
        .locator("option")
        .evaluateAll(
          (options) => options.find((o) => o.value && !o.disabled)?.value,
        );
      if (value) await input.selectOption(value);
    } else if (info.type === "file") {
      await input.setInputFiles({
        name: "preview.txt",
        mimeType: "text/plain",
        buffer: Buffer.from("Synthetic frontend fixture"),
      });
    } else if (info.type === "checkbox" || info.type === "radio")
      await input.check();
    else
      await input.fill(
        info.type === "date"
          ? "2026-10-05"
          : info.type === "number"
            ? "10"
            : info.type === "email"
              ? "fixture@example.invalid"
              : info.type === "tel"
                ? "+251900000001"
                : info.type === "password"
                  ? "Synthetic-fixture-123!"
                  : "Frontend fixture",
      );
  }
}
try {
  for (const path of [
    "/",
    "/about-us",
    "/our-services",
    "/doctors",
    "/doctors/1",
    "/doctors/2",
    "/doctors/3",
    "/appointment",
    "/working-hours",
    "/testimonials",
    "/contact-us",
    "/privacy-policy",
    "/terms-of-service",
    "/register",
  ])
    await check("public " + path, () => visit(path));
  await visit("/login");
  await page.getByLabel("Email address").fill(email);
  await page.getByLabel("Password", { exact: false }).fill(password);
  await page.getByRole("button", { name: "Sign In", exact: true }).click();
  await page.getByRole("heading", { name: "Dashboard", exact: true }).waitFor();

  const catalog = JSON.parse(
    readFileSync("apps/web/src/lib/legacy-catalog.json", "utf8"),
  );
  const special = new Set([
    "front-settings",
    "settings",
    "patient-id-card-template",
    "generate-patient-id-card",
    "odontogram",
    "bed-status",
    "hospital-schedule",
  ]);
  for (const [index, screen] of (process.env.HMS_FRONTEND_SKIP_CATALOG === "1"
    ? []
    : catalog.filter(
        (screen) => !onlyModules.length || onlyModules.includes(screen.id),
      )
  ).entries()) {
    await check("module " + screen.id, async () => {
      await visit(
        screen.id === "patients" ? "/patients" : "/modules/" + screen.id,
      );
      if (special.has(screen.id) || connected.has(screen.id)) return;
      await page.locator("main h1").first().waitFor();
      const create = page.getByRole("button", { name: /^New / }).first();
      if (!(await create.isVisible())) return;
      await create.click();
      const form = page.locator(".full-editor form, dialog[open] form").last();
      await form.waitFor();
      await fillRequired(form);
      await form.getByRole("button", { name: "Save", exact: true }).click();
      await page
        .getByRole("status")
        .filter({ hasText: "Saved in this frontend preview" })
        .waitFor();
    });
    if ((index + 1) % 20 === 0)
      console.log(
        `Checked ${index + 1}/${catalog.length} catalog routes/forms`,
      );
  }
  if (!onlyModules.length) {
    for (const id of [
      "attendance",
      "manage-attendance",
      "modules-setting",
      "patient-queue-theme",
      "front-cms-services",
    ])
      await check("extra " + id, () => visit("/modules/" + id));

    for (const role of [
      "Doctor",
      "Nurse",
      "Receptionist",
      "Pharmacist",
      "Lab Technician",
      "Accountant",
      "Case Manager",
      "Patient",
      "Admin",
    ]) {
      await check("role " + role, async () => {
        await visit("/dashboard");
        await page.getByLabel("Preview role").selectOption(role);
        await page.waitForFunction(
          (value) => sessionStorage.getItem("hms-preview-role") === value,
          role,
        );
        const links = await page
          .getByRole("navigation", { name: "Main navigation", exact: true })
          .locator("a")
          .evaluateAll((elements) =>
            elements.map((el) => el.getAttribute("href")),
          );
        for (const href of links) {
          await visit(href);
          if (href.startsWith("/portal/") && !connectedPortals.has(href)) {
            await page
              .getByRole("button", { name: "View 1", exact: true })
              .click();
            await page.getByRole("dialog").waitFor();
            await page.keyboard.press("Escape");
            await page.getByRole("dialog").waitFor({ state: "hidden" });
          }
        }
      });
      console.log("Checked role", role);
    }
    await check("schedule persistence and validation", async () => {
      await visit("/modules/hospital-schedule");
      await page.getByLabel("Monday Opening time").fill("10:00");
      await page.getByLabel("Monday Closing time").fill("09:00");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await page
        .getByText("Closing time must be after opening time.", { exact: true })
        .waitFor();
      await page.getByLabel("Monday Closing time").fill("18:00");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await page.reload();
      await page.waitForFunction(
        () =>
          document.querySelector('input[aria-label="Monday Opening time"]')
            ?.value === "10:00",
      );
    });
    await check("dental register chart isolation and persistence", async () => {
      await visit("/modules/odontogram");
      await page
        .getByRole("button", {
          name: "Edit odontogram Alex Morgan",
          exact: true,
        })
        .click();
      await page
        .getByRole("button", { name: "K tooth 1", exact: true })
        .click();
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await page
        .getByRole("button", {
          name: "Edit odontogram Jamie Wilson",
          exact: true,
        })
        .click();
      assert.equal(await page.locator("#Tooth1").getAttribute("fill"), "#fff");
      await page.getByRole("button", { name: "Cancel", exact: true }).click();
      await page.reload();
      await page
        .getByRole("button", {
          name: "Edit odontogram Alex Morgan",
          exact: true,
        })
        .click();
      assert.equal(
        await page.locator("#Tooth1").getAttribute("fill"),
        "#e91e63",
      );
      await page.getByRole("button", { name: "Cancel", exact: true }).click();
    });
    await check("card template rename updates cards", async () => {
      await visit("/modules/patient-id-card-template");
      await page
        .getByRole("button", { name: "Edit Standard", exact: true })
        .click();
      await page
        .getByLabel("Template Name", { exact: false })
        .fill("Renamed Standard");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await visit("/modules/generate-patient-id-card");
      await page
        .getByRole("cell", { name: "Renamed Standard", exact: true })
        .first()
        .waitFor();
      await page
        .getByRole("button", { name: "View card Alex Morgan", exact: true })
        .click();
      const download = page.waitForEvent("download");
      await page.getByRole("button", { name: "Download", exact: true }).click();
      await (await download).saveAs(output + "/smart-card.png");
    });
    await check("custom fields appear in configured module", async () => {
      await visit("/modules/add-custom-fields");
      await page.getByRole("button", { name: /^New / }).click();
      await page.getByLabel("Field Name", { exact: false }).fill("Triage Note");
      await page.getByLabel(/^Module:/).selectOption("Patients");
      await page
        .getByLabel("Field Type", { exact: false })
        .selectOption("Text");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await visit("/modules/patients");
      await page
        .getByRole("button", { name: "New Patient", exact: true })
        .click();
      await page
        .getByLabel("Triage Note", { exact: false })
        .fill("Synthetic triage note");
      await page.getByRole("button", { name: "Cancel", exact: true }).click();
    });
    await check("record edit delete search and CSV export", async () => {
      await visit("/modules/blood-banks");
      await page
        .getByRole("button", { name: "New Blood Bank", exact: true })
        .click();
      await page.getByLabel("Blood Group:").fill("TEST-CRUD");
      await page.getByLabel("Remained Bags:").fill("9");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await page.getByRole("status").waitFor();
      await page.reload();
      await page
        .getByRole("textbox", { name: "Search Blood Banks" })
        .fill("TEST-CRUD");
      await page.getByRole("button", { name: /^Edit DEMO-/ }).click();
      await page.getByLabel("Remained Bags:").fill("12");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      const csv = page.waitForEvent("download");
      await page.getByRole("button", { name: "Export", exact: true }).click();
      await (await csv).saveAs(output + "/blood-banks.csv");
      assert.ok(
        readFileSync(output + "/blood-banks.csv", "utf8").includes("TEST-CRUD"),
      );
      await page.getByRole("button", { name: /^Delete DEMO-/ }).click();
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Delete", exact: true })
        .click();
      await page
        .getByText("No matching records found", { exact: true })
        .waitFor();
    });
    // Persistent appointment conflicts are verified by integration.mjs and verify-connected.mjs.
    await check("role-specific billing tabs", async () => {
      await visit("/dashboard");
      await page.getByLabel("Preview role").selectOption("Receptionist");
      await page
        .getByRole("navigation", { name: "Main navigation", exact: true })
        .getByRole("link", { name: "Billings", exact: true })
        .click();
      await page.waitForURL("**/modules/bills");
      await page.getByRole("heading", { name: "Bills", exact: true }).waitFor();
      assert.deepEqual(
        (
          await page
            .getByRole("navigation", { name: "Module navigation", exact: true })
            .locator("a")
            .allTextContents()
        ).map((s) => s.trim()),
        ["Bills"],
      );
      await page.getByLabel("Preview role").selectOption("Admin");
    });
    await check("queue theme persistence", async () => {
      await visit("/modules/patient-queue-theme");
      await page.getByLabel("Message", { exact: true }).fill("Queue fixture");
      await page.getByRole("button", { name: "Save", exact: true }).click();
      await page.reload();
      await page.waitForFunction(
        () =>
          document.querySelector(
            '.legacy-form input[type="text"],.legacy-form input:not([type])',
          )?.value === "Queue fixture",
      );
    });
    for (const width of [1440, 390])
      for (const path of [
        "/",
        "/dashboard",
        "/patients",
        "/modules/manage-attendance",
        "/doctors/1",
      ]) {
        await check(`layout ${width} ${path}`, async () => {
          await page.setViewportSize({ width, height: 900 });
          await visit(path);
          assert.equal(
            await page.evaluate(
              () => document.documentElement.scrollWidth > innerWidth,
            ),
            false,
            "Horizontal page overflow",
          );
          await page.screenshot({
            path: `${output}/${width}-${path.replaceAll("/", "_") || "home"}.png`,
            fullPage: true,
          });
        });
      }
  }
  assert.deepEqual(pageErrors, [], "Browser runtime errors");
} finally {
  writeFileSync(
    output + "/results.json",
    JSON.stringify({ passed, failures, pageErrors }, null, 2),
  );
  await browser.close();
}
console.log(`${passed.length} checks passed; ${failures.length} failed.`);
if (failures.length) process.exitCode = 1;
