import { Pool } from "pg";
import { randomBytes } from "node:crypto";
import { spawnSync } from "node:child_process";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
const base = new URL(process.env.DATABASE_URL);
if (!["localhost", "127.0.0.1"].includes(base.hostname))
  throw Error("Restore QA must use loopback PostgreSQL");
const admin = new Pool({ connectionString: base.href });
const token = randomBytes(10).toString("hex");
const source = "hms_restore_source_" + token,
  target = "hms_restore_target_" + token;
const url = (name) => {
  const u = new URL(base);
  u.pathname = "/" + name;
  u.search = "";
  return u.href;
};
const directory = resolve(".local", "restore-drill-" + token);
await mkdir(directory, { recursive: true });
const env = {
  ...process.env,
  DATABASE_URL: url(source),
  RESTORE_DATABASE_URL: url(target),
  BACKUP_ENCRYPTION_KEY: randomBytes(32).toString("base64"),
};
let fixture, restored;
const created = [];
function call(mode, file, changes = {}) {
  return spawnSync(
    process.execPath,
    ["scripts/database-backup.mjs", mode, file],
    { env: { ...env, ...changes }, encoding: "utf8", windowsHide: true },
  );
}
try {
  for (const name of [source, target]) {
    await admin.query(`CREATE DATABASE "${name}"`);
    created.push(name);
  }
  fixture = new Pool({ connectionString: url(source) });
  await fixture.query(
    "CREATE TABLE patient(id integer PRIMARY KEY,name text NOT NULL); INSERT INTO patient VALUES(1,'Synthetic restore fixture'); CREATE TABLE audit_event(id integer PRIMARY KEY,action text); INSERT INTO audit_event VALUES(1,'synthetic.created')",
  );
  await fixture.end();
  fixture = null;
  const backup = call("backup", directory);
  if (backup.status !== 0) throw Error("Backup test failed: " + backup.stderr);
  const { backup: file } = JSON.parse(backup.stdout);
  if (
    call("restore", file, {
      BACKUP_ENCRYPTION_KEY: randomBytes(32).toString("base64"),
    }).status === 0
  )
    throw Error("Wrong decryption key accepted");
  const original = await readFile(file);
  const corrupt = Buffer.from(original);
  corrupt[0] ^= 1;
  await writeFile(file, corrupt);
  if (call("restore", file).status === 0)
    throw Error("Corrupt backup accepted");
  await writeFile(file, original);
  const result = call("restore", file);
  if (result.status !== 0) throw Error("Restore test failed: " + result.stderr);
  restored = new Pool({ connectionString: url(target) });
  const p = await restored.query("SELECT * FROM patient");
  const a = await restored.query("SELECT * FROM audit_event");
  if (
    p.rows.length !== 1 ||
    p.rows[0].name !== "Synthetic restore fixture" ||
    a.rows.length !== 1
  )
    throw Error("Restored data mismatch");
  if (call("restore", file).status === 0)
    throw Error("Nonempty target was overwritten");
  console.log(
    "PASS: encrypted backup, authenticated restore, exact data reconciliation, wrong-key/corruption rejection, and nonempty-target protection",
  );
} finally {
  if (fixture) await fixture.end();
  if (restored) await restored.end();
  for (const name of created.reverse()) {
    if (!/^hms_restore_(source|target)_[a-f0-9]{20}$/.test(name))
      throw Error("Unsafe QA cleanup name");
    await admin.query(`DROP DATABASE "${name}"`);
  }
  await admin.end();
}
