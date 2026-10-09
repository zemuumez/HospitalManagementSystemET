import { NextRequest } from "next/server";
export const runtime = "nodejs";
const allowedRoot = new Set([
  "me",
  "overview",
  "patients",
  "messages",
  "doctors",
  "slots",
  "appointments",
  "beds",
  "bed-types",
  "bed-assignments",
  "bed-transfers",
  "cases",
  "case-handlers",
  "encounters",
  "invoices",
  "charge-accounts",
  "billing-patients",
  "medicines",
  "medicine-batches",
  "medication-orders",
  "pharmacy-movements",
  "medicine-categories",
  "medicine-brands",
  "medicine-bills",
  "purchase-medicines",
  "used-medicines",
  "blood-bank",
  "blood-banks",
  "blood-donors",
  "blood-donations",
  "blood-issues",
  "ambulances",
  "ambulance-calls",
  "services",
  "packages",
  "packages-export",
  "insurances",
  "insurances-export",
  "operations",
  "operation-categories",
  "pathology-categories",
  "pathology-units",
  "pathology-parameters",
  "pathology-tests",
  "radiology-categories",
  "radiology-tests",
  "live-consultations",
  "live-meetings",
  "zoom-credentials",
  "attendance",
  "attendance-shifts",
  "attendance-leaves",
  "attendance-duties",
  "call-logs",
  "visitors",
  "postal-receives",
  "postal-dispatches",
  "postals",
  "complaints",
  "enquiries",
  "notices",
  "notice-boards",
  "inventory",
  "front-settings",
  "front-cms-settings",
  "settings",
  "general-settings",
  "hospital-schedules",
  "modules-setting",
  "module-settings",
  "attachments",
  "reviews",
  "doctor-departments",
  "doctor-schedules",
  "doctor-holidays",
  "doctor-breaks",
  "doctor-opd-charges",
  "doctor-absences",
  "employee-payrolls",
  "payrolls",
  "expenses",
  "incomes",
  "expense-heads",
  "income-heads",
  "prescriptions",
  "diagnostics",
  "diagnostic-categories",
  "diagnostic-units",
  "diagnostic-tests",
  "diagnostic-orders",
  "diagnostic-templates",
  "charges",
  "charge-categories",
  "vaccines",
  "vital-reports",
  "patient-queues",
  "documents",
  "document-types",
  "smart-cards",
  "addons",
  "sms",
  "mail",
  "email-templates",
  "sms-templates",
]);

function isAllowedPath(path: string[]): boolean {
  if (path.length === 0) return false;
  for (const seg of path) {
    if (!seg || seg.includes("..") || !/^[a-zA-Z0-9_\-\.]+$/.test(seg)) {
      return false;
    }
  }

  if (path.length === 1) {
    return allowedRoot.has(path[0]);
  }

  if (path.length === 2) {
    if (path[0] === "bed-occupancy" && path[1] === "report") return true;
    if (
      path[0] === "attendance" &&
      [
        "check-in",
        "check-out",
        "shifts",
        "assignments",
        "today",
        "clock-in",
        "clock-out",
        "records",
        "summary",
      ].includes(path[1])
    )
      return true;
    if (
      path[0] === "finance" &&
      ["expenses", "incomes", "expense-heads", "income-heads"].includes(path[1])
    )
      return true;
    if (
      path[0] === "cms" &&
      ["testimonials", "notice-boards"].includes(path[1])
    )
      return true;
    if (path[0] === "care-team") return true;
    return allowedRoot.has(path[0]);
  }

  if (path.length === 3) {
    if (path[0] === "finance" && path[1] === "reports" && path[2] === "summary")
      return true;
    if (path[0] === "patients") {
      return [
        "odontogram",
        "follow-ups",
        "referrals",
        "vital-reports",
        "vaccinations",
      ].includes(path[2]);
    }
    if (path[0] === "encounters") {
      return [
        "notes",
        "discharge",
        "care-team",
        "consultations",
        "ipd-details",
        "billing-summary",
        "clearance",
      ].includes(path[2]);
    }
    if (path[0] === "invoices") {
      return ["payments", "refunds"].includes(path[2]);
    }
    if (path[0] === "medication-orders" && path[2] === "cancel") {
      return true;
    }
    if (
      (path[0] === "doctors" ||
        path[0] === "prescriptions" ||
        path[0] === "live-consultations" ||
        path[0] === "live-meetings" ||
        path[0] === "insurances") &&
      path[2] === "status"
    ) {
      return true;
    }
    if (
      (path[0] === "diagnostic-orders" || path[0] === "pharmacy-movements") &&
      path[2] === "invoice"
    ) {
      return true;
    }
    if (path[0] === "employee-payrolls") {
      return path[2] === "slip";
    }
    if (path[0] === "attendance") {
      if (path[1] === "breaks" && ["start", "end"].includes(path[2]))
        return true;
      if (path[1] === "records" && ["history", "approval"].includes(path[2]))
        return true;
      if (path[1] === "shifts") return true;
    }
    if (
      path[0] === "doctor-departments" &&
      ["archive", "revisions"].includes(path[2])
    ) {
      return true;
    }
    if (path[0] === "complaints" && path[2] === "resolve") {
      return true;
    }
    if (path[0] === "enquiries" && path[2] === "read") {
      return true;
    }
    if (path[0] === "inventory" && ["categories", "items"].includes(path[1])) {
      return true;
    }
    if (path[0] === "attachments" && path[2] === "content") {
      return true;
    }
  }

  return false;
}

async function proxy(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> },
) {
  const { path } = await context.params;
  if (!isAllowedPath(path))
    return Response.json({ error: "Not found" }, { status: 404 });
  const origin = process.env.BETTER_AUTH_URL ?? "http://127.0.0.1:3000";
  if (request.method !== "GET" && request.headers.get("origin") !== origin)
    return Response.json({ error: "Origin not allowed" }, { status: 403 });
  const isAttachmentUpload =
    path[0] === "attachments" && request.method === "POST";
  const maxBytes = isAttachmentUpload ? 26 * 1024 * 1024 : 32768;
  let body: BodyInit | undefined;
  const reqContentType =
    request.headers.get("content-type") || "application/json";
  if (request.method !== "GET") {
    const reader = request.body?.getReader();
    const chunks: Uint8Array[] = [];
    let bytes = 0;
    if (reader) {
      while (true) {
        const chunk = await reader.read();
        if (chunk.done) break;
        bytes += chunk.value.byteLength;
        if (bytes > maxBytes) {
          await reader.cancel();
          return Response.json({ error: "Request too large" }, { status: 413 });
        }
        chunks.push(chunk.value);
      }
    }
    const merged = Buffer.concat(chunks);
    body = isAttachmentUpload ? merged : merged.toString("utf8");
  }
  try {
    const response = await fetch(
      `${process.env.GO_API_URL ?? "http://127.0.0.1:8080"}/v1/${path.join("/")}${request.nextUrl.search}`,
      {
        method: request.method,
        body,
        cache: "no-store",
        signal: AbortSignal.timeout(16000),
        redirect: "error",
        headers: {
          "Content-Type": reqContentType,
          Cookie: request.headers.get("cookie") ?? "",
          Origin: origin,
          "Idempotency-Key": request.headers.get("idempotency-key") ?? "",
        },
      },
    );
    if (response.status === 204) {
      return new Response(null, {
        status: 204,
        headers: {
          "Cache-Control": "no-store",
        },
      });
    }
    const contentType =
      response.headers.get("content-type") || "application/json";
    const buf = await response.arrayBuffer();
    return new Response(buf, {
      status: response.status,
      headers: {
        "Content-Type": contentType,
        "Cache-Control": "no-store",
      },
    });
  } catch {
    return Response.json(
      {
        error: "The hospital service is unavailable. Please try again shortly.",
      },
      { status: 503 },
    );
  }
}
export {
  proxy as GET,
  proxy as POST,
  proxy as PATCH,
  proxy as PUT,
  proxy as DELETE,
};
