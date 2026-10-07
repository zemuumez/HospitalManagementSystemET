import { Pool } from "pg";
import { createRequire } from "node:module";
const require = createRequire(import.meta.url);
import { randomBytes } from "node:crypto";
import { spawn, execFileSync } from "node:child_process";
import {
  cpSync,
  existsSync,
  mkdirSync,
  openSync,
  closeSync,
  readdirSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import { resolve } from "node:path";
import { createServer } from "node:net";
import { setTimeout as delay } from "node:timers/promises";

const root = process.cwd();
const manualMode = process.argv.includes("--manual");
const moduleAuditMode = process.argv.includes("--module-audit");
const endpointRepairMode = process.argv.includes("--endpoint-repairs");
const moduleUiAuditMode = process.argv.includes("--module-ui-audit");
const operationalMode = process.argv.includes("--operational");
const firebaseMode = process.argv.includes("--firebase");
const invitationsMode = process.argv.includes("--invitations");
const integrationMode = process.argv.includes("--integration");
const recoveryMode = process.argv.includes("--recovery");
const dsn = new URL(process.env.DATABASE_URL);
if (!["127.0.0.1", "localhost"].includes(dsn.hostname))
  throw Error("Browser QA requires a loopback PostgreSQL database");
const token = randomBytes(12).toString("hex");
const schema = `hms_browser_${token}`;
const admin = new Pool({ connectionString: dsn.href });
const children = [];
const logs = [];
let isolated;
let created = false;
async function port() {
  const server = createServer();
  await new Promise((ok, fail) => {
    server.once("error", fail);
    server.listen(0, "127.0.0.1", ok);
  });
  const value = server.address().port;
  await new Promise((ok) => server.close(ok));
  return value;
}
function start(command, args, env, cwd, label) {
  const fd = openSync(resolve(root, `.local/qa-${token}-${label}.log`), "w");
  logs.push(fd);
  const child = spawn(command, args, {
    cwd,
    env,
    stdio: ["ignore", fd, fd],
    windowsHide: true,
    detached: process.platform !== "win32",
  });
  child.on("error", () => {});
  children.push(child);
  return child;
}
async function ready(url, child) {
  for (let attempt = 0; attempt < 180; attempt++) {
    if (child.exitCode !== null || !child.pid)
      throw Error("Isolated QA service exited; inspect its local QA log");
    try {
      const response = await fetch(url, { signal: AbortSignal.timeout(2000) });
      if (response.ok) return;
    } catch {}
    await delay(500);
  }
  throw Error("Isolated QA service readiness timeout");
}
try {
  mkdirSync(resolve(root, ".local"), { recursive: true });
  mkdirSync(resolve(root, "services/api/bin"), { recursive: true });
  await admin.query(`CREATE SCHEMA ${schema}`);
  created = true;
  dsn.searchParams.set("options", `--search_path=${schema}`);
  isolated = new Pool({ connectionString: dsn.href });
  const client = await isolated.connect();
  try {
    for (const file of readdirSync(resolve(root, "db/migrations"))
      .filter((f) => f.endsWith(".sql"))
      .sort()) {
      await client.query("BEGIN");
      try {
        await client.query(
          readFileSync(resolve(root, "db/migrations", file), "utf8"),
        );
        await client.query("COMMIT");
      } catch (error) {
        await client.query("ROLLBACK");
        throw error;
      }
    }
  } finally {
    client.release();
  }
  const webRoot = resolve(root, `.local/qa-${token}/web`);
  mkdirSync(webRoot, { recursive: true });
  for (const name of [
    "src",
    "public",
    "package.json",
    "next.config.ts",
    "tsconfig.json",
    "next-env.d.ts",
    "postcss.config.mjs",
  ]) {
    const source = resolve(root, "apps/web", name);
    if (existsSync(source))
      cpSync(source, resolve(webRoot, name), { recursive: true });
  }
  const webPort = await port(),
    apiPort = await port();
  const env = {
    ...process.env,
    DATABASE_URL: dsn.href,
    BETTER_AUTH_URL: `http://127.0.0.1:${webPort}`,
    BETTER_AUTH_SECRET: randomBytes(48).toString("hex"),
    GO_API_URL: `http://127.0.0.1:${apiPort}`,
    API_ADDR: `127.0.0.1:${apiPort}`,
    APP_ENV: "development",
    NODE_ENV: "development",
    HMS_TEST_ISOLATED_SCHEMA: schema,
    GOPATH: resolve(root, ".cache/go"),
    GOMODCACHE: resolve(root, ".cache/go-mod"),
    GOCACHE: resolve(root, ".cache/go-build"),
  };
  for (const key of Object.keys(env))
    if (key.includes("FIREBASE") || key === "GOOGLE_APPLICATION_CREDENTIALS")
      env[key] = "";
  if (firebaseMode) {
    const authPort = await port();
    env.FIREBASE_PROJECT_ID = `demo-hms-${token.slice(0, 12)}`;
    env.FIREBASE_AUTH_EMULATOR_HOST = `127.0.0.1:${authPort}`;
    env.FIREBASE_CLI_DISABLE_USAGE_TRACKING = "true";
    env.XDG_CONFIG_HOME = resolve(webRoot, "../firebase-config");
    env.CI = "true";
    const config = resolve(webRoot, "../firebase.json");
    writeFileSync(
      config,
      JSON.stringify({
        emulators: {
          auth: { host: "127.0.0.1", port: authPort },
          hub: { host: "127.0.0.1", port: await port() },
          logging: { host: "127.0.0.1", port: await port() },
          ui: { enabled: false },
          singleProjectMode: true,
        },
      }),
    );
    const emulator = start(
      process.execPath,
      [
        require.resolve("firebase-tools/lib/bin/firebase.js"),
        "emulators:start",
        "--only",
        "auth",
        "--project",
        env.FIREBASE_PROJECT_ID,
        "--config",
        config,
        "--non-interactive",
      ],
      env,
      resolve(webRoot, ".."),
      "firebase",
    );
    await ready(
      `http://${env.FIREBASE_AUTH_EMULATOR_HOST}/emulator/v1/projects/${env.FIREBASE_PROJECT_ID}/config`,
      emulator,
    );
  }
  if (recoveryMode || integrationMode || invitationsMode) {
    Object.assign(env, {
      SMTP_HOST: "127.0.0.1",
      SMTP_PORT: "1025",
      SMTP_SECURE: "false",
      SMTP_USER: "",
      SMTP_PASSWORD: "",
      MAIL_FROM: "noreply@hms.local",
      SMTP_FROM: "noreply@hms.local",
      SMS_PROVIDER: "capture",
      TWILIO_ACCOUNT_SID: "",
      TWILIO_AUTH_TOKEN: "",
      TWILIO_FROM: "",
    });
  }
  execFileSync(
    "go",
    [
      "run",
      resolve(root, "scripts/reconcile_import.go"),
      "-out",
      resolve(root, `.local/reconciliation-${token}.json`),
    ],
    {
      cwd: resolve(root, "services/api"),
      env,
      stdio: "inherit",
      windowsHide: true,
    },
  );
  const binary = resolve(
    root,
    `services/api/bin/qa-api-${token}${process.platform === "win32" ? ".exe" : ""}`,
  );
  execFileSync("go", ["build", "-o", binary, "./cmd/api"], {
    cwd: resolve(root, "services/api"),
    env,
    stdio: "inherit",
    windowsHide: true,
  });
  const api = start(binary, [], env, root, "api");
  if (integrationMode || invitationsMode) {
    const workerBinary = resolve(
      root,
      `services/api/bin/qa-worker-${token}${process.platform === "win32" ? ".exe" : ""}`,
    );
    execFileSync("go", ["build", "-o", workerBinary, "./cmd/worker"], {
      cwd: resolve(root, "services/api"),
      env,
      stdio: "inherit",
      windowsHide: true,
    });
    start(workerBinary, [], env, root, "worker");
  }
  const web = start(
    process.execPath,
    [
      require.resolve("next/dist/bin/next"),
      "dev",
      "--hostname",
      "127.0.0.1",
      "--port",
      String(webPort),
    ],
    env,
    webRoot,
    "web",
  );
  await Promise.all([
    ready(env.GO_API_URL + "/readyz", api),
    ready(env.BETTER_AUTH_URL + "/login", web),
  ]);
  if (manualMode) {
    const { hashPassword } = await import("better-auth/crypto");
    const email = `manual-${token}@example.test`;
    const password = randomBytes(24).toString("hex") + "Aa1!";
    await isolated.query('INSERT INTO "user"(id,name,email) VALUES($1,$2,$3)', [
      token,
      "Isolated QA admin",
      email,
    ]);
    await isolated.query(
      `INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,'credential',$2,$3)`,
      [token + "-account", token, await hashPassword(password)],
    );
    await isolated.query(
      "INSERT INTO staff_access(user_id,role) VALUES($1,'admin')",
      [token],
    );
    console.log(
      JSON.stringify({ url: env.BETTER_AUTH_URL, email, password, schema }),
    );
    console.log(
      "Manual QA ready. Press Enter to stop services and remove this isolated schema.",
    );
    await new Promise((done) => {
      process.stdin.resume();
      process.stdin.once("data", done);
    });
    process.stdin.pause();
  } else {
    await new Promise((ok, fail) => {
      const test = spawn(
        process.execPath,
        [
          endpointRepairMode
            ? "scripts/verify-endpoint-repairs.mjs"
            : moduleUiAuditMode
              ? "scripts/verify-module-pages.mjs"
              : moduleAuditMode
                ? "scripts/verify-module-contracts.mjs"
                : operationalMode
                  ? "scripts/verify-all-operational-workspaces.mjs"
                  : invitationsMode
                    ? "scripts/verify-invitations.mjs"
                    : integrationMode
                      ? "scripts/integration.mjs"
                      : firebaseMode
                        ? "scripts/verify-firebase.mjs"
                        : recoveryMode
                          ? "scripts/verify-recovery.mjs"
                          : "scripts/verify-connected.mjs",
        ],
        {
          cwd: root,
          env,
          stdio: "inherit",
          windowsHide: true,
        },
      );
      children.push(test);
      test.once("error", fail);
      test.once("exit", (code) =>
        code === 0
          ? ok()
          : fail(Error("Isolated integration verification failed")),
      );
    });
  }
} finally {
  for (const child of children.reverse()) {
    if (!child.pid || child.exitCode !== null) continue;
    try {
      if (process.platform === "win32")
        execFileSync("taskkill", ["/PID", String(child.pid), "/T", "/F"], {
          stdio: "ignore",
          windowsHide: true,
        });
      else process.kill(-child.pid, "SIGKILL");
    } catch {
      child.kill();
    }
  }
  for (const fd of logs) closeSync(fd);
  await isolated?.end();
  if (created && /^hms_browser_[a-f0-9]{24}$/.test(schema))
    await admin.query(`DROP SCHEMA ${schema} CASCADE`);
  await admin.end();
  console.log("Isolated QA schema removed; development records retained.");
}
