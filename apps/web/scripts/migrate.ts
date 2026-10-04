import { readFile, readdir } from "node:fs/promises";
import { resolve } from "node:path";
import { pool } from "../src/lib/db.ts";
async function main() {
  const client = await pool.connect();
  try {
    await client.query("SELECT pg_advisory_lock(72841021)");
    await client.query(
      "CREATE TABLE IF NOT EXISTS schema_migration(name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())",
    );
    const folder = resolve(process.cwd(), "../../db/migrations");
    for (const name of (await readdir(folder))
      .filter((n) => n.endsWith(".sql"))
      .sort()) {
      if (
        (
          await client.query(
            "SELECT name FROM schema_migration WHERE name=$1",
            [name],
          )
        ).rowCount
      )
        continue;
      await client.query("BEGIN");
      try {
        await client.query(await readFile(resolve(folder, name), "utf8"));
        await client.query("INSERT INTO schema_migration(name) VALUES($1)", [
          name,
        ]);
        await client.query("COMMIT");
        console.log(`Applied ${name}`);
      } catch (error) {
        await client.query("ROLLBACK");
        throw error;
      }
    }
  } finally {
    await client.query("SELECT pg_advisory_unlock(72841021)");
    client.release();
    await pool.end();
  }
}
main().catch(() => {
  console.error("Migration failed. Check database access and migration SQL.");
  process.exitCode = 1;
});
