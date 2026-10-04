import { Pool } from "pg";
const globalPool = globalThis as unknown as { hmsPool?: Pool };
export const pool =
  globalPool.hmsPool ??
  new Pool({
    connectionString: process.env.DATABASE_URL,
    max: 10,
    connectionTimeoutMillis: 5000,
    idleTimeoutMillis: 30000,
  });
if (process.env.NODE_ENV !== "production") globalPool.hmsPool = pool;
