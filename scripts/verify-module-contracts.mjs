// Characterizes current authenticated contracts; legacy parity findings are reported separately.
// Only the isolated wrapper may launch this. It drops all generated fixtures on exit.
import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { mkdirSync, writeFileSync, readFileSync } from "node:fs";
import vm from "node:vm";
import { stripTypeScriptTypes } from "node:module";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
import { setTimeout as delay } from "node:timers/promises";
import { chromium } from "@playwright/test";

assert.match(
  process.env.HMS_TEST_ISOLATED_SCHEMA || "",
  /^hms_browser_[a-f0-9]{24}$/,
);
const base = process.env.BETTER_AUTH_URL;
assert.ok(["127.0.0.1", "localhost"].includes(new URL(base).hostname));
const db = new Pool({ connectionString: process.env.DATABASE_URL });
const results = [];
const findings = [];
const roles = [
  "admin",
  "doctor",
  "patient",
  "nurse",
  "receptionist",
  "pharmacist",
  "accountant",
  "case_manager",
  "lab_technician",
];
const staff = roles.filter((r) => r !== "patient");
const clinical = ["admin", "doctor", "patient", "receptionist"];
const probes = [
  ["patients", clinical],
  ["cases", clinical],
  ["encounters?kind=ipd", clinical],
  ["encounters?kind=opd", clinical],
  ["bed-types", ["admin", "doctor", "nurse", "receptionist"]],
  ["attendance/shifts", staff],
  ["ambulances", ["admin", "doctor", "nurse", "receptionist", "case_manager"]],
  [
    "charge-categories",
    ["admin", "doctor", "nurse", "receptionist", "accountant"],
  ],
  ["operations", ["admin", "doctor", "nurse", "receptionist"]],
  [
    "complaints",
    ["admin", "doctor", "nurse", "receptionist", "case_manager", "patient"],
  ],
  ["invoices", ["admin", "accountant", "patient"]],
  ["expenses", ["admin", "accountant"]],
  ["medicine-categories", ["admin", "doctor", "pharmacist"]],
  [
    "blood-bank",
    [
      "admin",
      "doctor",
      "nurse",
      "lab_technician",
      "pharmacist",
      "receptionist",
    ],
  ],
  ["live-consultations", ["admin", "doctor", "nurse", "patient"]],
  ["prescriptions", ["admin", "doctor", "nurse", "pharmacist", "patient"]],
];
async function request(path, actor, method = "GET", body) {
  // Respect the real 100/min session rate limiter; do not disable it for QA.
  await delay(700);
  return fetch(`${base}/api/hms/${path}`, {
    method,
    headers: {
      cookie: actor?.cookie || "",
      origin: base,
      "content-type": "application/json",
      "idempotency-key": randomUUID(),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}
async function check(name, fn) {
  try {
    await fn();
    results.push({ name, status: "pass" });
  } catch (e) {
    results.push({
      name,
      status: "fail",
      detail: String(e.message).slice(0, 1200),
    });
  }
}
async function fixture(role) {
  const id = randomUUID(),
    email = `module-${id}@example.test`,
    password = randomUUID() + "Aa1!";
  await db.query(
    'INSERT INTO "user"(id,name,email,"emailVerified") VALUES($1,$2,$3,true)',
    [id, `Synthetic ${role}`, email],
  );
  await db.query(
    'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
    [randomUUID(), id, await hashPassword(password)],
  );
  await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,$2)", [
    id,
    role,
  ]);
  const signIn = () =>
    fetch(`${base}/api/auth/sign-in/email`, {
      method: "POST",
      headers: { origin: base, "content-type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
  let response = await signIn();
  if (response.status === 429) {
    console.log(
      "Authentication rate limit observed; waiting for its window before continuing fixtures.",
    );
    await delay(61000);
    response = await signIn();
  }
  assert.equal(response.status, 200, `fixture login ${role}`);
  return {
    id,
    cookie: response.headers
      .getSetCookie()
      .map((c) => c.split(";")[0])
      .join("; "),
  };
}
try {
  const actors = {};
  for (const role of roles) actors[role] = await fixture(role);
  for (const [path, allowed] of probes) {
    console.log(`Checking module access: ${path}`);
    await check(`GET ${path}: anonymous denied`, async () =>
      assert.equal((await request(path)).status, 401),
    );
    for (const role of roles) {
      await check(`GET ${path}: ${role}`, async () => {
        const response = await request(path, actors[role]);
        assert.equal(
          response.status,
          allowed.includes(role) ? 200 : 403,
          `observed ${response.status}; expected ${allowed.includes(role) ? 200 : 403}`,
        );
        if (response.ok)
          assert.match(response.headers.get("content-type") || "", /json/);
      });
    }
  }
  let patient;
  await check("patient create and reload", async () => {
    const response = await request("patients", actors.admin, "POST", {
      givenName: "Module",
      familyName: "Audit",
      dateOfBirth: "2000-01-01",
    });
    assert.equal(response.status, 201);
    patient = await response.json();
    const list = await (
      await request("patients?search=Module", actors.admin)
    ).json();
    assert.ok(list.patients.some((p) => p.id === patient.id));
  });
  if (patient?.id) {
    for (const role of roles.filter((r) => !["admin", "doctor"].includes(r))) {
      await check(`odontogram write denied: ${role}`, async () =>
        assert.equal(
          (
            await request(
              `patients/${patient.id}/odontogram`,
              actors[role],
              "POST",
              {
                toothNumber: 1,
                condition: "caries",
                procedureNotes: "Synthetic audit",
              },
            )
          ).status,
          403,
        ),
      );
    }
    await check("odontogram unrelated doctor denied", async () => {
      const response = await request(
        `patients/${patient.id}/odontogram`,
        actors.doctor,
        "POST",
        { toothNumber: 1, condition: "caries" },
      );
      assert.ok(
        [403, 404].includes(response.status),
        `observed ${response.status}`,
      );
    });
    await check("odontogram invalid tooth rejected", async () =>
      assert.equal(
        (
          await request(
            `patients/${patient.id}/odontogram`,
            actors.admin,
            "POST",
            { toothNumber: 99, condition: "caries" },
          )
        ).status,
        422,
      ),
    );
    await check(
      "odontogram same-tooth persistence characterization",
      async () => {
        let first;
        for (const condition of ["caries", "filled"]) {
          const response = await request(
            `patients/${patient.id}/odontogram`,
            actors.admin,
            "POST",
            {
              toothNumber: 1,
              condition,
              procedureNotes: `Synthetic ${condition}`,
            },
          );
          assert.equal(response.status, 200);
          const saved = await response.json();
          if (!first) first = saved.id;
          else assert.equal(saved.id, first);
        }
        const rows = await db.query(
          "SELECT condition FROM patient_odontogram_entry WHERE patient_id=$1 AND tooth_number=1",
          [patient.id],
        );
        assert.equal(rows.rowCount, 1);
        assert.equal(rows.rows[0].condition, "filled");
        findings.push({
          id: "ODONTO-CHART-HISTORY",
          status: "confirmed-gap",
          evidence:
            "Two saves for one tooth return the same ID and leave one updated DB row; original creates separately identified chart records.",
        });
      },
    );
    await check(
      "original chart payload cannot be accepted as current tooth payload",
      async () => {
        const response = await request(
          `patients/${patient.id}/odontogram`,
          actors.admin,
          "POST",
          {
            patient_id: patient.id,
            doctor_id: actors.doctor.id,
            description: "Synthetic chart",
            odontogram: { 1: "K" },
          },
        );
        assert.equal(response.status, 400);
        findings.push({
          id: "ODONTO-AGGREGATE-CONTRACT",
          status: "confirmed-gap",
          evidence:
            "Current endpoint rejects original chart aggregate payload; requires explicit chart adapter/entity.",
        });
      },
    );
  }
  await check("frontend dental code round-trip characterization", async () => {
    const source = readFileSync(
      "apps/web/src/components/odontogram-register.tsx",
      "utf8",
    );
    const portion = source.slice(
      source.indexOf("export function legendCodeToDomainCondition"),
      source.indexOf("const doctors ="),
    );
    const js = stripTypeScriptTypes(
      portion.replaceAll("export function", "function"),
    );
    const context = vm.createContext({});
    vm.runInContext(js, context);
    const loss = ["K", "Ce", "D", "C", "B", "KR", "IP", "X", "F", "PS"]
      .map((code) => ({
        code,
        saved: context.legendCodeToDomainCondition(code),
        restored: context.domainConditionToLegendCode(
          context.legendCodeToDomainCondition(code),
        ),
      }))
      .filter((r) => r.restored !== r.code);
    assert.ok(loss.length > 0);
    findings.push({
      id: "ODONTO-CODE-LOSS",
      status: "confirmed-gap",
      evidence: loss,
    });
  });
  for (const path of [
    "packages",
    "insurances",
    "general-settings",
    "blood-banks",
  ]) {
    await check(`integration route characterization: ${path}`, async () => {
      const response = await request(path, actors.admin);
      assert.ok(
        response.status < 500,
        `unexpected server error ${response.status}`,
      );
      if (response.status === 404)
        findings.push({
          id: `ROUTE-${path.toUpperCase()}`,
          status: "confirmed-gap",
          evidence: `Authenticated GET /api/hms/${path} returned 404. Check proxy path and Go handler; frontend caller and API must agree.`,
        });
    });
  }
  await check(
    "smart-card template browser persistence characterization",
    async () => {
      const browser = await chromium.launch({
        headless: true,
        executablePath:
          process.env.HMS_CHROME_PATH ||
          "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
      });
      try {
        const cookies = actors.admin.cookie.split("; ").map((c) => {
          const index = c.indexOf("=");
          return {
            name: c.slice(0, index),
            value: c.slice(index + 1),
            url: base,
          };
        });
        const context = await browser.newContext();
        await context.addCookies(cookies);
        const page = await context.newPage();
        const mutations = [];
        page.on("request", (r) => {
          if (
            r.url().includes("/api/hms/") &&
            ["POST", "PUT", "PATCH", "DELETE"].includes(r.method())
          )
            mutations.push(r.method());
        });
        await page.goto(`${base}/modules/patient-id-card-template`);
        await page
          .getByRole("button", {
            name: "New Patient Smart Card Template",
            exact: true,
          })
          .click();
        const name = `Audit template ${randomUUID().slice(0, 8)}`;
        await page.getByLabel("Template Name", { exact: true }).fill(name);
        await page.getByRole("button", { name: "Save", exact: true }).click();
        await page.getByRole("cell", { name, exact: true }).waitFor();
        await page.reload();
        await page.getByRole("cell", { name, exact: true }).waitFor();
        assert.equal(mutations.length, 0);
        const fresh = await browser.newContext();
        await fresh.addCookies(cookies);
        const second = await fresh.newPage();
        await second.goto(`${base}/modules/patient-id-card-template`);
        await second
          .getByRole("button", {
            name: "New Patient Smart Card Template",
            exact: true,
          })
          .waitFor();
        assert.equal(
          await second.getByRole("cell", { name, exact: true }).count(),
          0,
        );
        findings.push({
          id: "CARDS-SESSION-ONLY",
          status: "confirmed-gap",
          evidence:
            "UI-created template survives same-tab reload, sends zero HMS mutations and disappears in a fresh browser context using the same admin account.",
        });
      } finally {
        await browser.close();
      }
    },
  );
} finally {
  await db.end();
  mkdirSync(".local/module-audit", { recursive: true });
  writeFileSync(
    ".local/module-audit/http-results.json",
    JSON.stringify(
      {
        date: new Date().toISOString(),
        scope:
          "Current policy status probes and targeted persistence characterization; not complete original-role parity or every CRUD.",
        results,
        findings,
      },
      null,
      2,
    ),
  );
  console.log(
    JSON.stringify(
      {
        passed: results.filter((r) => r.status === "pass").length,
        failed: results.filter((r) => r.status === "fail"),
        findings,
      },
      null,
      2,
    ),
  );
}
if (results.some((r) => r.status === "fail")) process.exitCode = 1;
