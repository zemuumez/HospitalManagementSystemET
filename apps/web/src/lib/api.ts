export type User = { id: string; name: string; role: string };
export type Patient = {
  userId: string;
  clinicianId: string;
  id: string;
  mrn: string;
  givenName: string;
  familyName: string;
  dateOfBirth: string;
  phone: string;
  createdAt: string;
};
export type Message = {
  id: string;
  channel: string;
  recipient: string;
  status: string;
  createdAt: string;
};
export class ApiError extends Error {
  status: number;
  code?: string;
  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/hms/${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  });
  const data = await response.json();
  if (!response.ok) {
    if (response.status === 401) window.location.assign("/login");
    throw new ApiError(data.error ?? "Unable to complete request", response.status, data.code);
  }
  return data;
}
