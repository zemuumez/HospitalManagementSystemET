import { NextRequest } from "next/server";

export const runtime = "nodejs";

export async function GET(request: NextRequest) {
  const token = request.nextUrl.searchParams.get("token") || "";
  const goApiUrl = process.env.GO_API_URL || "http://127.0.0.1:8080";
  try {
    const res = await fetch(
      `${goApiUrl}/v1/smart-cards/verify?token=${encodeURIComponent(token)}`,
      {
        cache: "no-store",
      },
    );
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json(
      { valid: false, error: "Verification service unavailable" },
      { status: 502 },
    );
  }
}
