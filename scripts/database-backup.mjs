import {
  createCipheriv,
  createDecipheriv,
  randomBytes,
  createHash,
} from "node:crypto";
import { createReadStream, createWriteStream } from "node:fs";
import {
  mkdir,
  readFile,
  writeFile,
  rm,
  rmdir,
  mkdtemp,
  readdir,
  stat,
} from "node:fs/promises";
import { resolve, join, basename } from "node:path";
import { tmpdir } from "node:os";
import { pipeline } from "node:stream/promises";
import { execFileSync } from "node:child_process";

const mode = process.argv[2];
if (!["backup", "restore"].includes(mode)) throw Error("Use backup or restore");
const key = Buffer.from(process.env.BACKUP_ENCRYPTION_KEY || "", "base64");
if (key.length !== 32)
  throw Error(
    "BACKUP_ENCRYPTION_KEY must be a base64-encoded random 32-byte key",
  );
const url = new URL(
  mode === "restore"
    ? process.env.RESTORE_DATABASE_URL
    : process.env.DATABASE_URL,
);
if (!["postgres:", "postgresql:"].includes(url.protocol))
  throw Error("PostgreSQL URL required");
const env = {
  ...process.env,
  PGHOST: url.hostname,
  PGPORT: url.port || "5432",
  PGUSER: decodeURIComponent(url.username),
  PGPASSWORD: decodeURIComponent(url.password),
  PGDATABASE: decodeURIComponent(url.pathname.slice(1)),
};
if (url.searchParams.has("sslmode"))
  env.PGSSLMODE = url.searchParams.get("sslmode");
function run(name, args) {
  const executable = process.env.PG_BIN
    ? join(
        process.env.PG_BIN,
        name + (process.platform === "win32" ? ".exe" : ""),
      )
    : name;
  try {
    return execFileSync(executable, args, {
      env,
      encoding: "utf8",
      windowsHide: true,
      stdio: ["ignore", "pipe", "pipe"],
    }).trim();
  } catch {
    throw Error(`${name} failed; no successful backup/restore is claimed`);
  }
}
async function hash(path) {
  const h = createHash("sha256");
  for await (const chunk of createReadStream(path)) h.update(chunk);
  return h.digest("hex");
}
const temp = await mkdtemp(join(tmpdir(), "hms-backup-"));
try {
  const dump = join(temp, "database.dump");
  if (mode === "backup") {
    const directory = resolve(
      process.argv[3] || process.env.BACKUP_DIR || "backups",
    );
    await mkdir(directory, { recursive: true, mode: 0o700 });
    const retention = Number(process.env.BACKUP_RETENTION_DAYS || 30);
    if (!Number.isInteger(retention) || retention < 1 || retention > 3650)
      throw Error("Invalid retention days");
    run("pg_dump", [
      "--format=custom",
      "--no-owner",
      "--no-privileges",
      `--file=${dump}`,
    ]);
    const name = `hms_backup_${Date.now()}_${randomBytes(6).toString("hex")}.enc`;
    const file = join(directory, name);
    const iv = randomBytes(12),
      cipher = createCipheriv("aes-256-gcm", key, iv);
    try {
      await pipeline(
        createReadStream(dump),
        cipher,
        createWriteStream(file, { flags: "wx", mode: 0o600 }),
      );
      const manifest = {
        version: 2,
        algorithm: "aes-256-gcm",
        file: name,
        iv: iv.toString("base64"),
        tag: cipher.getAuthTag().toString("base64"),
        sha256: await hash(file),
        rawSha256: await hash(dump),
        createdAt: new Date().toISOString(),
      };
      await writeFile(file + ".json", JSON.stringify(manifest, null, 2), {
        flag: "wx",
        mode: 0o600,
      });
    } catch (error) {
      await rm(file, { force: true });
      throw error;
    }
    // Only paired files created by this format are eligible for retention.
    for (const entry of await readdir(directory)) {
      if (!/^hms_backup_[0-9]+_[a-f0-9]{12}\.enc$/.test(entry)) continue;
      const path = join(directory, entry);
      if ((await stat(path)).mtimeMs > Date.now() - retention * 86400000)
        continue;
      try {
        await stat(path + ".json");
      } catch {
        continue;
      }
      await rm(path);
      await rm(path + ".json");
    }
    console.log(JSON.stringify({ backup: file, manifest: file + ".json" }));
  } else {
    const file = resolve(process.argv[3] || "");
    const m = JSON.parse(await readFile(file + ".json", "utf8"));
    if (
      m.version !== 2 ||
      m.algorithm !== "aes-256-gcm" ||
      m.file !== basename(file) ||
      m.sha256 !== (await hash(file))
    )
      throw Error("Backup manifest/integrity verification failed");
    const iv = Buffer.from(m.iv, "base64"),
      tag = Buffer.from(m.tag, "base64");
    if (iv.length !== 12 || tag.length !== 16)
      throw Error("Invalid encryption metadata");
    const decipher = createDecipheriv("aes-256-gcm", key, iv);
    decipher.setAuthTag(tag);
    await pipeline(
      createReadStream(file),
      decipher,
      createWriteStream(dump, { flags: "wx", mode: 0o600 }),
    );
    if ((await hash(dump)) !== m.rawSha256)
      throw Error("Decrypted backup verification failed");
    const objects = run("psql", [
      "-X",
      "-A",
      "-t",
      "-v",
      "ON_ERROR_STOP=1",
      "-c",
      `SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname NOT IN ('pg_catalog','information_schema') AND n.nspname NOT LIKE 'pg_toast%' AND c.relkind IN ('r','v','m','S','f')`,
    ]);
    if (objects !== "0")
      throw Error(
        "Restore target must be an empty database; existing databases are never cleaned or overwritten",
      );
    run("pg_restore", [
      "--exit-on-error",
      "--single-transaction",
      "--no-owner",
      "--no-privileges",
      "--dbname",
      env.PGDATABASE,
      dump,
    ]);
    console.log(
      JSON.stringify({
        restored: true,
        database: env.PGDATABASE,
        integrityVerified: true,
      }),
    );
  }
} finally {
  await rm(join(temp, "database.dump"), { force: true });
  await rmdir(temp);
  key.fill(0);
}
