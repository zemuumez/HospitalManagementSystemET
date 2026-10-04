import { chromium } from "@playwright/test";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
import { randomUUID } from "node:crypto";
import { mkdirSync } from "node:fs";
import assert from "node:assert/strict";

const base = process.env.BETTER_AUTH_URL;
if (
  !base ||
  !["127.0.0.1", "localhost"].includes(new URL(base).hostname) ||
  !["127.0.0.1", "localhost"].includes(
    new URL(process.env.DATABASE_URL).hostname,
  )
)
  throw Error("Requires local development web and database");
const db = new Pool({ connectionString: process.env.DATABASE_URL });
const id = randomUUID(),
  password = randomUUID() + "Aa1!",
  doctorPassword = randomUUID() + "Aa1!",
  email = `browser-${id}@example.test`,
  doctorEmail = `doctor-${id}@example.test`,
  family = `Browser${id.slice(0, 8)}`;
let browser;
try {
  await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$2,$3)', [
    id,
    "Browser integration admin",
    email,
  ]);
  await db.query(
    'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
    [randomUUID(), id, await hashPassword(password)],
  );
  await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,'admin')", [
    id,
  ]);
  browser = await chromium.launch({
    headless: true,
    executablePath:
      process.env.HMS_CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
  });
  const page = await browser.newPage({
      viewport: { width: 1440, height: 1000 },
    }),
    errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  page.setDefaultTimeout(15000);
  async function visit(path) {
    await page.goto(base + path);
    await page.locator('.legacy-shell[data-ready="true"]').waitFor();
  }
  await page.goto(base + "/login");
  await page.getByLabel("Email address").fill(email);
  await page.getByLabel("Password", { exact: false }).fill(password);
  await page.getByRole("button", { name: "Sign In", exact: true }).click();
  await page.waitForURL("**/dashboard");
  await visit("/modules/users");
  await page.getByRole("button", { name: "New User", exact: true }).click();
  let dialog = page.getByRole("dialog");
  await dialog.getByLabel("Name", { exact: true }).fill("Browser test doctor");
  await dialog.getByLabel("Email", { exact: true }).fill(doctorEmail);
  await dialog.getByLabel("Role", { exact: true }).selectOption("doctor");
  await dialog.getByLabel("Password", { exact: true }).fill(doctorPassword);
  await dialog.getByRole("button", { name: "Save", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
  await visit("/modules/schedules");
  await page.getByRole("button", { name: "New Schedule", exact: true }).click();
  dialog = page.getByRole("dialog");
  await dialog.getByLabel("Search doctor accounts").fill(doctorEmail);
  await dialog.getByRole("button", { name: "Search", exact: true }).click();
  await dialog
    .getByLabel("Doctor", { exact: true })
    .selectOption({ label: `Browser test doctor — ${doctorEmail}` });
  await dialog
    .getByLabel("Department", { exact: true })
    .fill("General medicine");
  await dialog.getByRole("button", { name: "Add period", exact: true }).click();
  const day = new Date(Date.now() + 3 * 86400000).toISOString().slice(0, 10),
    weekday = new Date(day + "T12:00:00+03:00").getUTCDay();
  await dialog.getByLabel("Day", { exact: true }).selectOption(String(weekday));
  await dialog.getByRole("button", { name: "Save", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
  await visit("/patients");
  await page
    .getByRole("button", { name: "Register patient", exact: true })
    .click();
  dialog = page.getByRole("dialog");
  await dialog.getByLabel("First name", { exact: true }).fill("Synthetic");
  await dialog.getByLabel("Last name", { exact: true }).fill(family);
  await dialog.getByLabel("Date of birth", { exact: true }).fill("2000-01-01");
  await dialog
    .getByRole("button", { name: "Register patient", exact: true })
    .click();
  await dialog.waitFor({ state: "hidden" });
  await visit("/modules/appointments");
  await page
    .getByRole("button", { name: "New Appointment", exact: true })
    .click();
  dialog = page.getByRole("dialog");
  await dialog.getByLabel("Search patients", { exact: true }).fill(family);
  await dialog.getByRole("button", { name: "Search", exact: true }).click();
  await dialog
    .getByLabel("Patient", { exact: true })
    .selectOption({ index: 1 });
  await dialog
    .getByLabel("Doctor", { exact: true })
    .selectOption({ label: "Browser test doctor — General medicine" });
  await dialog.getByLabel("Date", { exact: true }).fill(day);
  await dialog
    .getByLabel("Available slots (EAT)", { exact: true })
    .selectOption({ index: 1 });
  await dialog
    .getByLabel("Problem", { exact: true })
    .fill("Synthetic browser verification");
  await dialog.getByRole("button", { name: "Save", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
  await page.reload();
  await page
    .getByRole("cell", { name: `Synthetic ${family}`, exact: true })
    .waitFor();
  assert.equal(
    (
      await db.query(
        "SELECT count(*)::int AS count FROM appointment WHERE created_by=$1",
        [id],
      )
    ).rows[0].count,
    1,
  );
  mkdirSync(".local/connected-verification", { recursive: true });
  await page.screenshot({
    path: ".local/connected-verification/appointments.png",
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({
    path: ".local/connected-verification/mobile.png",
    fullPage: true,
  });
  assert.ok(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth + 1,
    ),
  );

  await page.setViewportSize({ width: 1440, height: 1000 });
  await visit("/modules/beds");
  await page.getByRole("button", { name: "New Bed", exact: true }).click();
  dialog = page.getByRole("dialog");
  await dialog.getByLabel("Name", { exact: true }).fill(`Bed-${family}`);
  await dialog.getByLabel("Bed Type", { exact: true }).fill("General");
  await dialog.getByLabel("Charge (ETB)", { exact: true }).fill("123.45");
  await dialog.getByRole("button", { name: "Save", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
  await visit("/modules/patient-cases");
  await page.getByRole("button", { name: "New Case", exact: true }).click();
  dialog = page.getByRole("dialog");
  await dialog.getByLabel("Search patients", { exact: true }).fill(family);
  await dialog.getByRole("button", { name: "Search", exact: true }).click();
  await dialog
    .getByLabel("Patient", { exact: true })
    .selectOption({ index: 1 });
  await dialog
    .getByLabel("Doctor", { exact: true })
    .selectOption({ label: "Browser test doctor" });
  await dialog
    .getByLabel("Description", { exact: true })
    .fill("Synthetic case");
  await dialog.getByRole("button", { name: "Save", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
  await visit("/modules/ipd-patient-departments");
  await page
    .getByRole("button", { name: "New IPD Patient", exact: true })
    .click();
  dialog = page.getByRole("dialog");
  await dialog
    .getByLabel("Patient Case", { exact: true })
    .selectOption({ index: 1 });
  await dialog
    .getByLabel("Bed", { exact: true })
    .selectOption({ label: `Bed-${family} \u2014 General` });
  await dialog
    .getByLabel("Symptoms", { exact: true })
    .fill("Synthetic admission");
  await dialog.getByRole("button", { name: "Save", exact: true }).click();
  await dialog.waitFor({ state: "hidden" });
  const signout = await page.request.post(base + "/api/auth/sign-out", {
    headers: { origin: base },
    data: {},
  });
  assert.equal(signout.status(), 200);
  await page.goto(base + "/login");
  await page.getByLabel("Email address").fill(doctorEmail);
  await page.getByLabel("Password", { exact: false }).fill(doctorPassword);
  await page.getByRole("button", { name: "Sign In", exact: true }).click();
  await page.waitForURL("**/dashboard");
  await visit("/modules/ipd-patient-departments");
  await page.getByRole("button", { name: "View", exact: true }).click();
  dialog = page.getByRole("dialog");
  await dialog
    .getByLabel("Discharge Summary", { exact: true })
    .fill("Synthetic discharge. Test complete.");
  await dialog
    .getByRole("button", { name: "Discharge patient", exact: true })
    .click();
  await dialog.waitFor({ state: "hidden" });
  await page.reload();
  await page.getByRole("cell", { name: "discharged", exact: true }).waitFor();
  await visit("/modules/beds");
  await page
    .getByRole("row")
    .filter({ hasText: `Bed-${family}` })
    .getByText("Available", { exact: true })
    .waitFor();
  console.log(
    "PASS: browser bed and case creation, IPD admission, doctor sign-in, discharge persistence and released bed.",
  );
  assert.deepEqual(errors, []);
  console.log(
    "PASS: browser staff creation, doctor schedule, patient registration, booking, persistence after reload, mobile layout and no page errors.",
  );
} finally {
  await browser?.close();
  const users = (
    await db.query('SELECT id FROM "user" WHERE id=$1 OR email=$2', [
      id,
      doctorEmail,
    ])
  ).rows.map((u) => u.id);
  await db.query("DELETE FROM appointment WHERE created_by=$1", [id]);
  await db.query("DELETE FROM encounter WHERE created_by=$1", [id]);
  await db.query("DELETE FROM patient_case WHERE created_by=$1", [id]);
  await db.query("DELETE FROM hospital_bed WHERE created_by=$1", [id]);
  await db.query("DELETE FROM patient WHERE family_name=$1", [family]);
  await db.query("DELETE FROM doctor_hours WHERE doctor_id=ANY($1::text[])", [
    users,
  ]);
  await db.query("DELETE FROM doctor_profile WHERE user_id=ANY($1::text[])", [
    users,
  ]);
  await db.query("DELETE FROM audit_event WHERE actor_id=ANY($1::text[])", [
    users,
  ]);
  await db.query('DELETE FROM "user" WHERE id=ANY($1::text[])', [users]);
  await db.end();
}
