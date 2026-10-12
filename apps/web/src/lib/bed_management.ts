export interface SaveBedTypeInput {
  id?: string;
  name: string;
  description: string;
  active?: boolean;
  version?: number;
}

export interface SaveBedInput {
  id?: string;
  name: string;
  type: string;
  typeId?: string;
  wardName?: string;
  chargeMinor: number;
  state?: string;
  description?: string;
  version?: number;
}

export interface AssignBedInput {
  patientId: string;
  encounterId: string;
  bedId: string;
  notes?: string;
  version: number;
}

export async function saveBedType(
  input: SaveBedTypeInput,
  customFetch: typeof fetch = fetch,
) {
  const isEdit = Boolean(input.id);
  const url = isEdit
    ? `/api/hms/bed-types/${input.id}`
    : `/api/hms/bed-types`;
  const method = isEdit ? "PATCH" : "POST";
  const body = {
    name: input.name.trim(),
    description: input.description.trim(),
    active: input.active ?? true,
    ...(isEdit && input.version ? { version: input.version } : {}),
  };

  const res = await customFetch(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Failed to save bed type");
  }
  return data;
}

export async function deleteBedType(
  id: string,
  version: number,
  customFetch: typeof fetch = fetch,
) {
  if (!version || version <= 0) {
    throw new Error("A positive bed type version is required");
  }
  const res = await customFetch(`/api/hms/bed-types/${id}?version=${version}`, {
    method: "DELETE",
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Failed to delete bed type");
  }
  return data;
}

export async function saveBed(
  input: SaveBedInput,
  customFetch: typeof fetch = fetch,
) {
  const isEdit = Boolean(input.id);
  const url = isEdit ? `/api/hms/beds/${input.id}` : `/api/hms/beds`;
  const method = isEdit ? "PATCH" : "POST";
  const body = {
    name: input.name.trim(),
    type: input.type.trim(),
    typeId: input.typeId,
    wardName: input.wardName?.trim(),
    chargeMinor: input.chargeMinor,
    description: input.description?.trim(),
    state: input.state || "available",
    ...(isEdit && input.version ? { version: input.version } : {}),
  };

  const res = await customFetch(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Failed to save bed");
  }
  return data;
}

export async function deleteBed(
  id: string,
  customFetch: typeof fetch = fetch,
) {
  const res = await customFetch(`/api/hms/beds/${id}`, {
    method: "DELETE",
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Failed to delete bed");
  }
  return data;
}

export async function assignBed(
  input: AssignBedInput,
  customFetch: typeof fetch = fetch,
) {
  if (!input.patientId || !input.encounterId || !input.bedId) {
    throw new Error("Patient ID, Encounter ID, and Bed ID are all required");
  }
  if (!input.version || input.version <= 0) {
    throw new Error("A positive encounter version is required");
  }

  const res = await customFetch(`/api/hms/bed-assignments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      patientId: input.patientId,
      encounterId: input.encounterId,
      bedId: input.bedId,
      notes: input.notes?.trim() || "",
      version: input.version,
    }),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Failed to assign bed");
  }
  return data;
}

export async function dischargeBed(
  assignmentId: string,
  version: number,
  notes: string = "",
  customFetch: typeof fetch = fetch,
) {
  if (!version || version <= 0) {
    throw new Error("A positive encounter version is required");
  }
  const res = await customFetch(`/api/hms/bed-assignments/${assignmentId}/discharge`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ version, dischargeNotes: notes.trim() }),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Failed to discharge bed");
  }
  return data;
}
