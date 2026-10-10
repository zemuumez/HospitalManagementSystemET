import { randomUUID } from "node:crypto";
import { hashPassword } from "better-auth/crypto";
import { z } from "zod";
import { auth } from "@/lib/auth";
import { pool } from "@/lib/db";

export const runtime = "nodejs";
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
] as const;
const doctorHourSchema = z.object({
  weekday: z.number().int().min(0).max(6),
  startMinute: z.number().int().min(0).max(1440),
  endMinute: z.number().int().min(0).max(1440),
});

const createInput = z
  .object({
    name: z.string().trim().min(1).max(120),
    email: z
      .email()
      .max(254)
      .transform((v) => v.toLowerCase()),
    role: z.enum(roles),
    password: z.string().min(12).max(128),
    departmentId: z.string().uuid().optional(),
    department_id: z.string().uuid().optional(),
    specialist: z.string().trim().min(1).max(191).optional(),
    designation: z.string().trim().min(1).max(191).optional(),
    qualification: z.string().trim().min(1).max(191).optional(),
    gender: z.union([z.string(), z.number()]).optional(),
    dateOfBirth: z.string().max(32).optional(),
    dob: z.string().max(32).optional(),
    bloodGroup: z.string().max(32).optional(),
    blood_group: z.string().max(32).optional(),
    phone: z.string().max(32).optional(),
    address1: z.string().max(191).optional(),
    address2: z.string().max(191).optional(),
    city: z.string().max(191).optional(),
    zip: z.string().max(32).optional(),
    postalCode: z.string().max(32).optional(),
    description: z.string().max(2000).optional(),
    photoUrl: z.string().max(512).optional(),
    photo_url: z.string().max(512).optional(),
    appointmentCharge: z.number().min(0).optional(),
    opdCharge: z.number().min(0).optional(),
    slotMinutes: z.number().int().min(5).max(120).optional(),
    hours: z.array(doctorHourSchema).max(21).optional(),
  })
  .strict();

const changeInput = z
  .object({
    id: z.string().min(1).max(128),
    active: z.boolean().optional(),
    name: z.string().trim().min(1).max(120).optional(),
    email: z
      .email()
      .max(254)
      .transform((v) => v.toLowerCase())
      .optional(),
  })
  .strict()
  .refine(
    (d) =>
      d.active !== undefined || d.name !== undefined || d.email !== undefined,
    { message: "At least one field to change must be supplied" },
  );

function json(body: unknown, status = 200) {
  return Response.json(body, {
    status,
    headers: { "Cache-Control": "no-store" },
  });
}
async function administrator(request: Request) {
  const session = await auth.api.getSession({ headers: request.headers });
  if (!session) return null;
  const result = await pool.query(
    "SELECT user_id FROM staff_access WHERE user_id=$1 AND role='admin' AND active",
    [session.user.id],
  );
  return result.rowCount ? session.user.id : null;
}
export async function GET(request: Request) {
  try {
    const actor = await administrator(request);
    if (!actor) return json({ error: "Administrator access required" }, 403);
    const url = new URL(request.url),
      page = Number(url.searchParams.get("page") || 1),
      limit = Math.min(
        500,
        Math.max(1, Number(url.searchParams.get("limit") || 25)),
      ),
      search = url.searchParams.get("search") || "";
    if (
      !Number.isInteger(page) ||
      page < 1 ||
      page > 1000 ||
      search.length > 80
    )
      return json({ error: "Invalid search" }, 422);
    const result = await pool.query(
      `SELECT u.id,u.name,u.email,a.role,a.active FROM "user" u JOIN staff_access a ON a.user_id=u.id WHERE ($1='' OR strpos(lower(u.name||' '||u.email),lower($1))>0) ORDER BY u.name,u.id LIMIT $3 OFFSET $2`,
      [search, (page - 1) * limit, limit],
    );
    return json({ users: result.rows, page, pageSize: limit });
  } catch {
    return json({ error: "Unable to load hospital users" }, 503);
  }
}
async function mutate(request: Request) {
  if (
    request.headers.get("origin") !==
    (process.env.BETTER_AUTH_URL ?? "http://127.0.0.1:3000")
  )
    return json({ error: "Origin not allowed" }, 403);
  try {
    const actor = await administrator(request);
    if (!actor) return json({ error: "Administrator access required" }, 403);
    const reader = request.body?.getReader();
    let body = "",
      size = 0;
    const decoder = new TextDecoder();
    if (reader) {
      while (true) {
        const chunk = await reader.read();
        if (chunk.done) break;
        size += chunk.value.byteLength;
        if (size > 8192) {
          await reader.cancel();
          return json({ error: "Request too large" }, 413);
        }
        body += decoder.decode(chunk.value, { stream: true });
      }
      body += decoder.decode();
    }
    let input: unknown;
    try {
      input = JSON.parse(body);
    } catch {
      return json({ error: "Invalid JSON" }, 400);
    }
    const creating = request.method === "POST";
    const parsed = creating
      ? createInput.safeParse(input)
      : changeInput.safeParse(input);
    if (!parsed.success)
      return json(
        {
          error:
            "Check the supplied fields. Passwords need at least 12 characters.",
        },
        422,
      );

    if (creating) {
      const data = parsed.data as z.infer<typeof createInput>;
      if (data.role === "doctor") {
        const deptId = data.departmentId || data.department_id;
        const spec = data.specialist?.trim();
        const desig = data.designation?.trim();
        const qual = data.qualification?.trim();
        const rawGender =
          data.gender !== undefined && data.gender !== null
            ? String(data.gender).trim().toLowerCase()
            : "";
        if (
          !deptId ||
          !spec ||
          !desig ||
          !qual ||
          !rawGender ||
          (rawGender !== "male" &&
            rawGender !== "female" &&
            rawGender !== "other" &&
            rawGender !== "0" &&
            rawGender !== "1")
        ) {
          return json(
            {
              error:
                "Doctor profiles require a valid department, specialist, designation, qualification, and gender.",
            },
            422,
          );
        }
        const dobStr = data.dateOfBirth || data.dob;
        if (dobStr) {
          const d = new Date(dobStr);
          if (isNaN(d.getTime()) || d.getFullYear() < 1850 || d > new Date()) {
            return json({ error: "Invalid date of birth" }, 422);
          }
        }
        if (data.hours && data.hours.length > 0) {
          const slot = data.slotMinutes || 60;
          for (let idx = 0; idx < data.hours.length; idx++) {
            const h = data.hours[idx];
            if (h.endMinute - h.startMinute < slot) {
              return json(
                {
                  error:
                    "Working hours duration must be at least slot duration",
                },
                422,
              );
            }
            for (let otherIdx = 0; otherIdx < idx; otherIdx++) {
              const o = data.hours[otherIdx];
              if (
                h.weekday === o.weekday &&
                h.startMinute < o.endMinute &&
                o.startMinute < h.endMinute
              ) {
                return json(
                  { error: "Overlapping working hours are not permitted" },
                  422,
                );
              }
            }
          }
        }
      }
    }

    const hash = creating
      ? await hashPassword(createInput.parse(input).password)
      : null;
    const db = await pool.connect();
    try {
      await db.query("BEGIN");
      // Serialize access changes to protect the final administrator and in-flight checks.
      await db.query("SELECT pg_advisory_xact_lock(72841022)");
      const current = await db.query(
        "SELECT user_id FROM staff_access WHERE user_id=$1 AND role='admin' AND active FOR UPDATE",
        [actor],
      );
      if (!current.rowCount) {
        await db.query("ROLLBACK");
        return json({ error: "Administrator access required" }, 403);
      }
      let id: string;
      if (creating) {
        const data = parsed.data as z.infer<typeof createInput>;
        const deptId = data.departmentId || data.department_id;

        if (data.role === "doctor") {
          id = randomUUID();
          await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$2,$3)', [
            id,
            data.name,
            data.email,
          ]);
          await db.query(
            `INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,'credential',$2,$3)`,
            [randomUUID(), id, hash],
          );
          await db.query(
            "INSERT INTO staff_access(user_id,role) VALUES($1,$2)",
            [id, data.role],
          );

          const deptCheck = await db.query(
            "SELECT title, archived FROM doctor_department WHERE id=$1",
            [deptId],
          );
          if (!deptCheck.rowCount || deptCheck.rows[0].archived) {
            await db.query("ROLLBACK");
            return json(
              { error: "Valid non-archived department is required" },
              422,
            );
          }
          const deptTitle = deptCheck.rows[0].title;

          const slotMinutes = data.slotMinutes || 60;
          await db.query(
            `INSERT INTO doctor_profile(
              user_id, department, department_id, specialist, designation, qualification,
              description, photo_url, opd_charge, appointment_charge, slot_minutes, version
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1)`,
            [
              id,
              deptTitle,
              deptId,
              data.specialist!.trim(),
              data.designation!.trim(),
              data.qualification!.trim(),
              data.description?.trim() || "",
              data.photoUrl || data.photo_url || "",
              data.opdCharge || 0,
              data.appointmentCharge || 0,
              slotMinutes,
            ],
          );

          let genderStr = String(data.gender).trim().toLowerCase();
          if (genderStr === "0") genderStr = "male";
          else if (genderStr === "1") genderStr = "female";

          const dob = data.dateOfBirth || data.dob || "";
          const bloodGroup = data.bloodGroup || data.blood_group || "";
          const details = {
            phone: data.phone || "",
            gender: genderStr,
            dateOfBirth: dob,
            bloodGroup,
            designation: data.designation!.trim(),
            qualification: data.qualification!.trim(),
            specialty: data.specialist!.trim(),
            address1: data.address1 || "",
            address2: data.address2 || "",
            city: data.city || "",
            postalCode: data.zip || data.postalCode || "",
          };

          await db.query(
            `INSERT INTO staff_profile(user_id, details, version)
             VALUES($1, $2, 1)
             ON CONFLICT (user_id) DO UPDATE SET details = EXCLUDED.details`,
            [id, JSON.stringify(details)],
          );

          const hours =
            data.hours && data.hours.length > 0
              ? data.hours
              : Array.from({ length: 7 }, (_, weekday) => ({
                  weekday,
                  startMinute: 600,
                  endMinute: 1170,
                }));

          for (const h of hours) {
            await db.query(
              "INSERT INTO doctor_hours(doctor_id, weekday, start_minute, end_minute) VALUES($1, $2, $3, $4)",
              [id, h.weekday, h.startMinute, h.endMinute],
            );
          }

          await db.query(
            "INSERT INTO audit_event(actor_id, action, resource_id) VALUES($1, 'doctor.created', $2)",
            [actor, id],
          );
        } else {
          id = randomUUID();
          await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$2,$3)', [
            id,
            data.name,
            data.email,
          ]);
          await db.query(
            `INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,'credential',$2,$3)`,
            [randomUUID(), id, hash],
          );
          await db.query(
            "INSERT INTO staff_access(user_id,role) VALUES($1,$2)",
            [id, data.role],
          );
        }
      } else {
        const data = parsed.data as z.infer<typeof changeInput>;
        id = data.id;
        const target = await db.query(
          "SELECT role,active FROM staff_access WHERE user_id=$1 FOR UPDATE",
          [id],
        );
        if (!target.rowCount) {
          await db.query("ROLLBACK");
          return json({ error: "User not found" }, 404);
        }
        if (data.name !== undefined) {
          await db.query('UPDATE "user" SET name=$1 WHERE id=$2', [
            data.name,
            id,
          ]);
        }
        if (data.email !== undefined) {
          await db.query('UPDATE "user" SET email=$1 WHERE id=$2', [
            data.email,
            id,
          ]);
        }
        if (data.active !== undefined) {
          if (id === actor && !data.active) {
            await db.query("ROLLBACK");
            return json({ error: "You cannot disable your own account" }, 409);
          }
          if (target.rows[0].role === "admin" && !data.active) {
            const count = await db.query(
              "SELECT count(*)::int AS count FROM staff_access WHERE role='admin' AND active",
            );
            if (count.rows[0].count <= 1) {
              await db.query("ROLLBACK");
              return json(
                { error: "The hospital must retain an active administrator" },
                409,
              );
            }
          }
          await db.query("UPDATE staff_access SET active=$2 WHERE user_id=$1", [
            id,
            data.active,
          ]);
          if (!data.active)
            await db.query('DELETE FROM session WHERE "userId"=$1', [id]);
        }
      }
      await db.query(
        "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,$2,$3)",
        [
          actor,
          creating ? "identity.provisioned" : "identity.access_changed",
          id,
        ],
      );
      await db.query("COMMIT");
      return json({ id }, creating ? 201 : 200);
    } catch (error) {
      await db.query("ROLLBACK");
      if ((error as { code?: string }).code === "23505")
        return json({ error: "This email is already registered" }, 409);
      throw error;
    } finally {
      db.release();
    }
  } catch {
    return json({ error: "Unable to update hospital users" }, 503);
  }
}

const deleteInput = z
  .object({
    id: z.string().min(1).max(128),
  })
  .strict();

async function remove(request: Request) {
  try {
    let body: unknown;
    try {
      body = await request.json();
    } catch {
      return json({ error: "Invalid JSON" }, 400);
    }
    const parsed = deleteInput.safeParse(body);
    if (!parsed.success) return json({ error: "Valid ID is required" }, 422);
    const { id } = parsed.data;

    const actor = await administrator(request);
    if (!actor)
      return json(
        { error: "Only administrators can delete staff members" },
        403,
      );
    if (id === actor)
      return json({ error: "You cannot delete your own account" }, 409);

    const db = await pool.connect();
    try {
      await db.query("BEGIN");
      const target = await db.query(
        "SELECT role FROM staff_access WHERE user_id=$1 FOR UPDATE",
        [id],
      );
      if (!target.rowCount) {
        await db.query("ROLLBACK");
        return json({ error: "Staff member not found" }, 404);
      }

      // Check clinical and payroll dependencies
      const inUseCheck = await db.query(
        `SELECT
          EXISTS(SELECT 1 FROM patient_case WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM encounter WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM appointment WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM birth_report WHERE delivered_by=$1)
          OR EXISTS(SELECT 1 FROM death_report WHERE certified_by=$1)
          OR EXISTS(SELECT 1 FROM investigation_report WHERE investigated_by=$1)
          OR EXISTS(SELECT 1 FROM operation_report WHERE surgeon_id=$1)
          OR EXISTS(SELECT 1 FROM prescription WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM ipd_admission_details i JOIN encounter e ON e.id=i.encounter_id WHERE e.doctor_id=$1)
          OR EXISTS(SELECT 1 FROM employee_payroll WHERE user_id=$1)
          OR EXISTS(SELECT 1 FROM opd_follow_up WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM patient_queue WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM public_appointment_request WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM live_consultation WHERE doctor_id=$1)
          OR EXISTS(SELECT 1 FROM patient_referral WHERE referred_by=$1)
          OR EXISTS(SELECT 1 FROM patient_odontogram_entry WHERE diagnosed_by=$1)
          OR EXISTS(SELECT 1 FROM clinical_note WHERE author_id=$1)
          OR EXISTS(SELECT 1 FROM patient WHERE user_id=$1 OR clinician_user_id=$1)
          OR EXISTS(SELECT 1 FROM audit_event WHERE actor_id=$1)
          AS in_use`,
        [id],
      );
      if (inUseCheck.rows[0].in_use) {
        await db.query("ROLLBACK");
        return json(
          {
            error: "Doctor is in use by clinical records and cannot be deleted",
            code: "RECORD_IN_USE",
          },
          409,
        );
      }

      await db.query("DELETE FROM doctor_hours WHERE doctor_id=$1", [id]);
      await db.query("DELETE FROM doctor_absence WHERE doctor_id=$1", [id]);
      await db.query("DELETE FROM doctor_profile WHERE user_id=$1", [id]);
      await db.query("DELETE FROM staff_profile_revision WHERE user_id=$1", [
        id,
      ]);
      await db.query("DELETE FROM staff_role_event WHERE user_id=$1", [id]);
      await db.query("DELETE FROM staff_profile WHERE user_id=$1", [id]);
      await db.query("DELETE FROM staff_access WHERE user_id=$1", [id]);
      await db.query('DELETE FROM session WHERE "userId"=$1', [id]);
      await db.query(
        "DELETE FROM verification WHERE value=$1 AND (identifier LIKE '2fa-%' OR identifier LIKE 'trust-device-%')",
        [id],
      );
      await db.query('DELETE FROM account WHERE "userId"=$1', [id]);
      await db.query(
        "INSERT INTO audit_event(actor_id, action, resource_id) VALUES($1, 'doctor.deleted', $2)",
        [actor, id],
      );
      await db.query('DELETE FROM "user" WHERE id=$1', [id]);
      await db.query("COMMIT");
      return json({ id, deleted: true }, 200);
    } catch (error) {
      await db.query("ROLLBACK");
      throw error;
    } finally {
      db.release();
    }
  } catch {
    return json({ error: "Unable to delete staff member" }, 503);
  }
}

export { mutate as POST, mutate as PATCH, remove as DELETE };
