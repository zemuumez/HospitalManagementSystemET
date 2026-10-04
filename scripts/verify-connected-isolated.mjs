import { Pool } from "pg";
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
} from "node:fs";
import { resolve } from "node:path";
import { createServer } from "node:net";
import { setTimeout as delay } from "node:timers/promises";

const root = process.cwd();
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
  const binary = resolve(
    root,
    `services/api/bin/qa-api${process.platform === "win32" ? ".exe" : ""}`,
  );
  execFileSync("go", ["build", "-o", binary, "./cmd/api"], {
    cwd: resolve(root, "services/api"),
    env,
    stdio: "inherit",
    windowsHide: true,
  });
  const api = start(binary, [], env, root, "api");
  const web = start(
    process.execPath,
    [
      resolve(root, "node_modules/next/dist/bin/next"),
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
  await new Promise((ok, fail) => {
    const test = spawn(process.execPath, ["scripts/verify-connected.mjs"], {
      cwd: root,
      env,
      stdio: "inherit",
      windowsHide: true,
    });
    children.push(test);
    test.once("error", fail);
    test.once("exit", (code) =>
      code === 0 ? ok() : fail(Error("Connected browser verification failed")),
    );
  });
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
  console.log(
    "Isolated browser QA schema removed; development records retained.",
  );
}
