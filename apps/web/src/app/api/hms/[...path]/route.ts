import { NextRequest } from "next/server";
export const runtime = "nodejs";
async function proxy(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> },
) {
  const { path } = await context.params;
  const allowed =
    path.length === 1 &&
    [
      "me",
      "overview",
      "patients",
      "messages",
      "doctors",
      "slots",
      "appointments",
      "beds",
      "cases",
      "encounters",
    ].includes(path[0]);
  const appointment =
    path.length === 2 &&
    ["appointments", "patients"].includes(path[0]) &&
    /^[0-9a-f-]{36}$/i.test(path[1]);
  const encounter =
    path.length === 3 &&
    path[0] === "encounters" &&
    /^[0-9a-f-]{36}$/i.test(path[1]) &&
    ["notes", "discharge"].includes(path[2]);
  if (!allowed && !appointment && !encounter)
    return Response.json({ error: "Not found" }, { status: 404 });
  const origin = process.env.BETTER_AUTH_URL ?? "http://127.0.0.1:3000";
  if (request.method !== "GET" && request.headers.get("origin") !== origin)
    return Response.json({ error: "Origin not allowed" }, { status: 403 });
  let body: string | undefined;
  if (request.method !== "GET") {
    const reader = request.body?.getReader();
    const chunks: Uint8Array[] = [];
    let bytes = 0;
    if (reader) {
      while (true) {
        const chunk = await reader.read();
        if (chunk.done) break;
        bytes += chunk.value.byteLength;
        if (bytes > 32768) {
          await reader.cancel();
          return Response.json({ error: "Request too large" }, { status: 413 });
        }
        chunks.push(chunk.value);
      }
    }
    body = Buffer.concat(chunks).toString("utf8");
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
          "Content-Type": "application/json",
          Cookie: request.headers.get("cookie") ?? "",
          Origin: origin,
          "Idempotency-Key": request.headers.get("idempotency-key") ?? "",
        },
      },
    );
    return new Response(response.body, {
      status: response.status,
      headers: {
        "Content-Type": "application/json",
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
export { proxy as GET, proxy as POST, proxy as PATCH };
