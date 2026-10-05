import fs from "node:fs";
import { chromium } from "@playwright/test";
(async () => {
  const email = process.env.HMS_TEST_EMAIL,
    password = process.env.HMS_TEST_PASSWORD;
  if (!email || !password)
    throw Error("Set HMS_TEST_EMAIL and HMS_TEST_PASSWORD for local visual QA");
  const b = await chromium.launch({
    headless: true,
    ...(process.env.HMS_CHROME_PATH
      ? { executablePath: process.env.HMS_CHROME_PATH }
      : {}),
  });
  const p = await b.newPage({ viewport: { width: 1440, height: 1000 } });
  for (const pattern of ["**/api/hms/**", "**/api/staff**"])
    await p.route(pattern, (route) =>
      route.request().method() === "GET" ? route.continue() : route.abort(),
    );
  const errors = [];
  p.on("pageerror", (e) => errors.push(e.message));
  await p.goto("http://127.0.0.1:3000/login");
  await p.getByLabel("Email address", { exact: false }).fill(email);
  await p.getByLabel("Password", { exact: false }).fill(password);
  await p.getByRole("button", { name: "Sign In", exact: true }).click();
  await p.waitForURL("**/dashboard");
  await p.locator(".dashboard-widget").first().waitFor();
  await p.getByRole("button", { name: "Use dark theme", exact: true }).click();
  fs.mkdirSync(".local/reference-repair", { recursive: true });
  for (const slug of [
    "dashboard",
    "patient-id-card-template",
    "generate-patient-id-card",
    "odontogram",
    "attendance",
    "attendance-report",
    "attendance-shifts",
    "attendance-assignments",
    "attendance-leaves",
    "attendance-requests",
    "manage-attendance",
    "manual-billing-payments",
    "advance-payments",
    "payment-reports",
    "payments",
    "invoices",
    "accounts",
    "employee-payrolls",
    "bills",
    "bed-status",
    "bed-assigns",
    "beds",
    "bed-types",
  ]) {
    await p.goto(
      "http://127.0.0.1:3000/" +
        (slug === "dashboard" ? slug : "modules/" + slug),
    );
    await p.locator('[data-ready="true"]').waitFor();
    await p.screenshot({
      path: ".local/reference-repair/" + slug + ".png",
      fullPage: true,
    });
  }
  await p.goto("http://127.0.0.1:3000/modules/patient-id-card-template");
  await p
    .getByRole("switch", { name: "Standard email", exact: true })
    .uncheck();
  await p.reload();
  await p.locator('[data-ready="true"]').waitFor();
  if (
    await p
      .getByRole("switch", { name: "Standard email", exact: true })
      .isChecked()
  )
    throw Error("Card visibility did not persist");
  await p.goto("http://127.0.0.1:3000/modules/generate-patient-id-card");
  await p
    .getByRole("button", { name: "View card Alex Morgan", exact: true })
    .click();
  if (
    await p
      .getByRole("dialog")
      .locator("dt")
      .filter({ hasText: "Email" })
      .count()
  )
    throw Error("Hidden card email was rendered");
  const download = p.waitForEvent("download");
  await p.getByRole("button", { name: "Download", exact: true }).click();
  await (await download).saveAs(".local/reference-repair/card.png");
  await p.getByRole("button", { name: "Close", exact: true }).click();
  await p.goto("http://127.0.0.1:3000/modules/odontogram");
  await p.getByRole("button", { name: "Add Odontogram", exact: true }).click();
  await p.getByRole("dialog").waitFor();
  await p.screenshot({
    path: ".local/reference-repair/odontogram-modal.png",
    fullPage: true,
  });
  await p.getByRole("button", { name: "Cancel", exact: true }).click();
  await p
    .getByRole("button", { name: "Edit odontogram Alex Morgan", exact: true })
    .click();
  await p.getByRole("button", { name: "K tooth 1", exact: true }).click();
  await p.getByRole("button", { name: "Save", exact: true }).click();
  await p.reload();
  await p
    .getByRole("button", { name: "Edit odontogram Alex Morgan", exact: true })
    .click();
  if ((await p.locator("#Tooth1").getAttribute("fill")) !== "#e91e63")
    throw Error("Dental marking did not persist");
  await p.getByRole("button", { name: "Cancel", exact: true }).click();
  await p.goto("http://127.0.0.1:3000/modules/attendance-leaves");
  await p
    .getByRole("button", { name: "Approve Alex Morgan", exact: true })
    .click();
  await p.reload();
  await p.getByRole("cell", { name: "Approved", exact: true }).waitFor();
  await p.goto("http://127.0.0.1:3000/modules/attendance-shifts");
  await p.getByRole("button", { name: "Add Shifts", exact: true }).click();
  await p.getByLabel("Name", { exact: false }).first().fill("QA Evening");
  await p.getByLabel("Code", { exact: false }).fill("QA");
  await p.getByLabel("Start Time", { exact: false }).fill("16:00");
  await p.getByLabel("End Time", { exact: false }).fill("22:00");
  await p.getByLabel("Status", { exact: false }).selectOption("Active");
  await p.getByRole("button", { name: "Save", exact: true }).click();
  await p.getByRole("cell", { name: "QA Evening", exact: true }).waitFor();
  await p.goto("http://127.0.0.1:3000/modules/attendance");
  if ((await p.locator(".legacy-submenu a").count()) !== 6)
    throw Error("Attendance tabs missing");
  for (const kind of ["ipd", "opd"]) {
    await p.goto(`http://127.0.0.1:3000/modules/${kind}-patient-departments`);
    await p
      .getByRole("button", {
        name: `New ${kind.toUpperCase()} Patient`,
        exact: true,
      })
      .click();
    await p.locator(".encounter-intake").waitFor();
    await p
      .locator('.encounter-intake [role="status"]')
      .waitFor({ state: "hidden" });
    await p.screenshot({
      path: `.local/reference-repair/${kind}-form.png`,
      fullPage: true,
    });
    await p.getByRole("button", { name: "Cancel", exact: true }).click();
  }
  await p.setViewportSize({ width: 1440, height: 1000 });
  await p.goto("http://127.0.0.1:3000/modules/manual-billing-payments");
  await p.locator('[data-ready="true"]').waitFor();
  await p.locator(".billing-search-box input").fill("Trith");
  if ((await p.locator(".billing-table tbody tr").count()) !== 3)
    throw Error("Manual billing search did not filter correctly");
  await p.locator(".billing-search-box input").fill("");
  await p.goto("http://127.0.0.1:3000/modules/invoices");
  await p.locator('[data-ready="true"]').waitFor();
  await p.getByRole("button", { name: "New Invoice", exact: true }).click();
  await p.locator("#billing-form-title").waitFor();
  await p.getByRole("button", { name: "Cancel", exact: true }).click();
  await p.goto("http://127.0.0.1:3000/modules/bed-assigns");
  await p.locator('[data-ready="true"]').waitFor();
  await p.locator(".billing-search-box input").fill("Vinay");
  if ((await p.locator(".billing-table tbody tr").count()) !== 1)
    throw Error("Bed assigns search did not filter correctly");
  await p.locator(".billing-search-box input").fill("");
  await p.goto("http://127.0.0.1:3000/modules/bed-types");
  await p.locator('[data-ready="true"]').waitFor();
  await p.getByRole("button", { name: "New Bed Type", exact: true }).click();
  await p.locator("#bed-type-modal-title").waitFor();
  await p.getByRole("button", { name: "Cancel", exact: true }).click();
  await p.goto("http://127.0.0.1:3000/modules/attendance");
  await p.setViewportSize({ width: 390, height: 844 });
  await p.screenshot({
    path: ".local/reference-repair/attendance-mobile.png",
    fullPage: true,
  });
  if (await p.evaluate(() => document.documentElement.scrollWidth > innerWidth))
    throw Error("Mobile overflow");
  if (errors.length) throw Error(errors.join("\n"));
  console.log(
    "PASS: dashboard, cards, dental dialog, six attendance tabs, shift create, mobile overflow and browser errors",
  );
  await b.close();
})().catch((e) => {
  console.error(e.message);
  process.exit(1);
});
