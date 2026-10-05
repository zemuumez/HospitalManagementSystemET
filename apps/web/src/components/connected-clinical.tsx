"use client";
import { useCallback, useEffect, useState } from "react";
import { api, type Patient } from "@/lib/api";
import { useLanguage } from "./language";
import { useIdentity } from "./workspace";
import {
  useResource,
  Table,
  Status,
  Editor,
  Paging,
  Input,
  errorText,
  type Doctor,
} from "./connected-scheduling";
type Bed = {
  id: string;
  name: string;
  type: string;
  chargeMinor: number;
  available: boolean;
};
type Case = {
  id: string;
  number: number;
  patientId: string;
  doctorId: string;
  patientName: string;
  doctorName: string;
  description: string;
};
export type Encounter = {
  patientEmail?: string;
  doctorEmail?: string;
  billStatus?: string;
  totalVisits?: number;
  intake?: {
    height: number;
    weight: number;
    bloodPressure: string;
    notes: string;
    identificationNumber: string;
    reference: string;
    oldPatient: boolean;
    standardChargeMinor: number;
    paymentMode: string;
  };
  id: string;
  number: number;
  kind: string;
  patientName: string;
  doctorName: string;
  doctorId: string;
  bedName: string;
  admittedAt: string;
  status: string;
  version: number;
  symptoms: string;
  dischargeSummary: string;
};
type Note = { id: string; authorName: string; body: string; signedAt: string };
type Mode = "beds" | "cases" | "ipd" | "opd";
const date = (value: string) =>
  new Date(value).toLocaleString("en-GB", { timeZone: "Africa/Addis_Ababa" });

export function ConnectedClinical({ mode }: { mode: Mode }) {
  const { t } = useLanguage(),
    identity = useIdentity(),
    [page, setPage] = useState(1),
    [open, setOpen] = useState(false),
    [selected, setSelected] = useState<Encounter | null>(null);
  const loader = useCallback(
      () =>
        api<{ beds?: Bed[]; cases?: Case[]; encounters?: Encounter[] }>(
          mode === "beds" || mode === "cases"
            ? `${mode}?page=${page}`
            : `encounters?kind=${mode}&page=${page}`,
        ),
      [mode, page],
    ),
    resource = useResource(loader);
  const rows =
    resource.data?.beds ||
    resource.data?.cases ||
    resource.data?.encounters ||
    [];
  const title = {
    beds: "Beds",
    cases: "Patient Cases",
    ipd: "IPD - Patient In",
    opd: "OPD - Patient Out",
  }[mode];
  const canCreate =
    mode === "beds"
      ? identity?.user.role === "admin"
      : identity?.permissions.includes("clinical.admit");
  return (
    <>
      <div className="page-heading">
        <h1>{t(title)}</h1>
        {canCreate && (
          <button className="primary" onClick={() => setOpen(true)}>
            {t(
              mode === "beds"
                ? "New Bed"
                : mode === "cases"
                  ? "New Case"
                  : mode === "ipd"
                    ? "New IPD Patient"
                    : "New OPD Patient",
            )}
          </button>
        )}
      </div>
      <Status {...resource} empty={!rows.length} />
      {!resource.loading &&
        resource.data &&
        (mode === "beds" ? (
          <Table head={["Name", "Bed Type", "Charge (ETB)", "Status"]}>
            {resource.data.beds?.map((b) => (
              <tr key={b.id}>
                <td>{b.name}</td>
                <td>{b.type}</td>
                <td>{(b.chargeMinor / 100).toFixed(2)}</td>
                <td>
                  <span className="badge">
                    {t(b.available ? "Available" : "Occupied")}
                  </span>
                </td>
              </tr>
            ))}
          </Table>
        ) : mode === "cases" ? (
          <Table head={["Case ID", "Patient", "Doctor", "Description"]}>
            {resource.data.cases?.map((c) => (
              <tr key={c.id}>
                <td>CASE-{String(c.number).padStart(6, "0")}</td>
                <td>{c.patientName}</td>
                <td>{c.doctorName}</td>
                <td>{c.description}</td>
              </tr>
            ))}
          </Table>
        ) : (
          <Table
            head={[
              "Number",
              "Patient",
              "Doctor",
              "Bed",
              "Admission Date",
              "Status",
              "Action",
            ]}
          >
            {resource.data.encounters?.map((e) => (
              <tr key={e.id}>
                <td>
                  {mode.toUpperCase()}-{String(e.number).padStart(6, "0")}
                </td>
                <td>{e.patientName}</td>
                <td>{e.doctorName}</td>
                <td>{e.bedName || "—"}</td>
                <td>{date(e.admittedAt)} EAT</td>
                <td>{t(e.status)}</td>
                <td>
                  <button className="secondary" onClick={() => setSelected(e)}>
                    {t("View")}
                  </button>
                </td>
              </tr>
            ))}
          </Table>
        ))}
      <Paging
        page={page}
        setPage={setPage}
        count={rows.length}
        loading={resource.loading}
      />
      {open && (
        <CreateClinical
          mode={mode}
          onClose={() => setOpen(false)}
          onSaved={resource.reload}
        />
      )}{" "}
      {selected && (
        <EncounterDetails
          encounter={selected}
          onClose={() => setSelected(null)}
          onSaved={resource.reload}
        />
      )}
    </>
  );
}
function CreateClinical({
  mode,
  onClose,
  onSaved,
}: {
  mode: Mode;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useLanguage();
  const [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [patients, setPatients] = useState<Patient[]>([]),
    [search, setSearch] = useState(""),
    [referencePage, setReferencePage] = useState(1),
    [bedPage, setBedPage] = useState(1);
  const load = useCallback(async () => {
    if (mode === "cases")
      return {
        doctors: (await api<{ doctors: Doctor[] }>("doctors")).doctors,
        cases: [] as Case[],
        beds: [] as Bed[],
      };
    if (mode === "beds")
      return {
        doctors: [] as Doctor[],
        cases: [] as Case[],
        beds: [] as Bed[],
      };
    const [cases, beds] = await Promise.all([
      api<{ cases: Case[] }>(`cases?page=${referencePage}`),
      mode === "ipd"
        ? api<{ beds: Bed[] }>(`beds?page=${bedPage}`)
        : Promise.resolve({ beds: [] as Bed[] }),
    ]);
    return { ...cases, ...beds, doctors: [] as Doctor[] };
  }, [mode, referencePage, bedPage]);
  const resource = useResource(load);
  const [request, setRequest] = useState<{ key: string; body: string } | null>(
      null,
    ),
    [admittedAt] = useState(() => new Date().toISOString());
  async function find() {
    setError("");
    try {
      setPatients(
        (
          await api<{ patients: Patient[] }>(
            `patients?search=${encodeURIComponent(search)}`,
          )
        ).patients,
      );
    } catch (e) {
      setError(errorText(e));
    }
  }
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const form = Object.fromEntries(new FormData(e.currentTarget));
      let payload: unknown;
      if (mode === "beds") {
        const money = String(form.charge);
        if (!/^\d{1,7}(\.\d{1,2})?$/.test(money))
          throw Error("Enter an amount with at most two decimal places");
        const [whole, fraction = ""] = money.split(".");
        payload = {
          name: form.name,
          type: form.type,
          chargeMinor: Number(whole) * 100 + Number(fraction.padEnd(2, "0")),
        };
      } else if (mode === "cases") payload = form;
      else
        payload = {
          kind: mode,
          caseId: form.caseId,
          bedId: mode === "ipd" ? form.bedId : "",
          admittedAt,
          symptoms: form.symptoms,
        };
      const body = JSON.stringify(payload),
        key = request?.body === body ? request.key : crypto.randomUUID();
      setRequest({ key, body });
      await api(mode === "beds" || mode === "cases" ? mode : "encounters", {
        method: "POST",
        headers: { "Idempotency-Key": key },
        body,
      });
      onSaved();
      onClose();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Editor
      title={
        mode === "beds"
          ? "New Bed"
          : mode === "cases"
            ? "New Case"
            : "Patient admission"
      }
      onClose={() => {
        if (!busy) onClose();
      }}
    >
      <form className="space-y-4" onSubmit={save}>
        {error && (
          <p className="error" role="alert">
            {t(error)}
          </p>
        )}
        {resource.error && (
          <p className="error" role="alert">
            {t(resource.error)}
          </p>
        )}
        {mode === "beds" ? (
          <>
            <Input label="Name" name="name" required maxLength={80} />
            <Input label="Bed Type" name="type" required maxLength={80} />
            <Input
              label="Charge (ETB)"
              name="charge"
              inputMode="decimal"
              required
              pattern="[0-9]{1,7}(\.[0-9]{1,2})?"
            />
          </>
        ) : mode === "cases" ? (
          <>
            <div className="flex gap-2">
              <Input
                label="Search patients"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                maxLength={80}
              />
              <button type="button" className="secondary" onClick={find}>
                {t("Search")}
              </button>
            </div>
            <label>
              <span className="label">{t("Patient")}</span>
              <select
                aria-label={t("Patient")}
                className="field"
                name="patientId"
                required
                defaultValue=""
              >
                <option value="">{t("Select Patient")}</option>
                {patients.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.givenName} {p.familyName} — {p.mrn}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span className="label">{t("Doctor")}</span>
              <select
                aria-label={t("Doctor")}
                className="field"
                name="doctorId"
                required
                defaultValue=""
              >
                <option value="">{t("Select Doctor")}</option>
                {resource.data?.doctors.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span className="label">{t("Description")}</span>
              <textarea className="field" name="description" maxLength={2000} />
            </label>
          </>
        ) : (
          <>
            <label>
              <span className="label">{t("Patient Case")}</span>
              <select
                aria-label={t("Patient Case")}
                className="field"
                name="caseId"
                key={`case-${referencePage}`}
                required
                defaultValue=""
              >
                <option value="">{t("Select Case")}</option>
                {resource.data?.cases.map((c) => (
                  <option key={c.id} value={c.id}>
                    CASE-{c.number} — {c.patientName} — {c.doctorName}
                  </option>
                ))}
              </select>
            </label>
            {mode === "ipd" && (
              <label>
                <span className="label">{t("Bed")}</span>
                <select
                  aria-label={t("Bed")}
                  className="field"
                  name="bedId"
                  key={`bed-${bedPage}`}
                  required
                  defaultValue=""
                >
                  <option value="">{t("Select Bed")}</option>
                  {resource.data?.beds
                    .filter((b) => b.available)
                    .map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.name} — {b.type}
                      </option>
                    ))}
                </select>
              </label>
            )}
            <div className="flex gap-2">
              <button
                type="button"
                className="secondary"
                disabled={referencePage === 1 || resource.loading}
                onClick={() => setReferencePage((p) => p - 1)}
              >
                {t("Previous choices")}
              </button>
              <button
                type="button"
                className="secondary"
                disabled={
                  resource.loading ||
                  ((resource.data?.cases.length || 0) < 25 &&
                    (resource.data?.beds.length || 0) < 25)
                }
                onClick={() => setReferencePage((p) => p + 1)}
              >
                {t("More choices")}
              </button>
            </div>
            {mode === "ipd" && (
              <div className="flex gap-2">
                <button
                  type="button"
                  className="secondary"
                  disabled={bedPage === 1 || resource.loading}
                  onClick={() => setBedPage((p) => p - 1)}
                >
                  {t("Previous beds")}
                </button>
                <button
                  type="button"
                  className="secondary"
                  disabled={
                    resource.loading || (resource.data?.beds.length || 0) < 25
                  }
                  onClick={() => setBedPage((p) => p + 1)}
                >
                  {t("More beds")}
                </button>
              </div>
            )}
            <p className="text-sm text-muted">
              {t("Admission time")} {date(admittedAt)} EAT
            </p>
            <label>
              <span className="label">{t("Symptoms")}</span>
              <textarea className="field" name="symptoms" maxLength={2000} />
            </label>
          </>
        )}
        <button
          className="primary"
          disabled={busy || resource.loading || !!resource.error}
        >
          {t(busy ? "Saving…" : "Save")}
        </button>
      </form>
    </Editor>
  );
}
export function EncounterDetails({
  encounter,
  onClose,
  onSaved,
}: {
  encounter: Encounter;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useLanguage(),
    identity = useIdentity(),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [body, setBody] = useState(""),
    [summary, setSummary] = useState(""),
    [noteKey, setNoteKey] = useState("");
  const canRead =
      identity && ["admin", "doctor", "patient"].includes(identity.user.role),
    canSign =
      identity?.user.role === "doctor" &&
      identity.user.id === encounter.doctorId &&
      encounter.status === "active";
  const loader = useCallback(
      () =>
        canRead
          ? api<{ notes: Note[] }>(`encounters/${encounter.id}/notes`)
          : Promise.resolve({ notes: [] }),
      [canRead, encounter.id],
    ),
    resource = useResource(loader);
  useEffect(() => setNoteKey(crypto.randomUUID()), [body]);
  async function sign(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api(`encounters/${encounter.id}/notes`, {
        method: "POST",
        headers: { "Idempotency-Key": noteKey },
        body: JSON.stringify({ body }),
      });
      setBody("");
      resource.reload();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  async function discharge(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api(`encounters/${encounter.id}/discharge`, {
        method: "POST",
        body: JSON.stringify({ summary, version: encounter.version }),
      });
      onSaved();
      onClose();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Editor
      title="Patient encounter"
      onClose={() => {
        if (!busy) onClose();
      }}
    >
      <div className="space-y-5">
        <p>
          {encounter.patientName} · {encounter.doctorName}
        </p>
        <p>
          {t(encounter.status)} · {date(encounter.admittedAt)} EAT
        </p>
        <p className="whitespace-pre-wrap">{encounter.symptoms}</p>
        {encounter.intake && (
          <dl className="legacy-form">
            {Object.entries(encounter.intake).map(([key, value]) => (
              <div key={key}>
                <dt className="label">{t(key.replace(/([A-Z])/g, " $1"))}</dt>
                <dd>
                  {typeof value === "boolean"
                    ? t(value ? "Yes" : "No")
                    : String(value || "N/A")}
                </dd>
              </div>
            ))}
          </dl>
        )}
        {error && (
          <p className="error" role="alert">
            {t(error)}
          </p>
        )}
        {canRead && (
          <>
            <h3>{t("Clinical Notes")}</h3>
            <Status {...resource} empty={!resource.data?.notes.length} />
            {resource.data?.notes.map((n) => (
              <article key={n.id} className="border rounded p-3">
                <p className="whitespace-pre-wrap">{n.body}</p>
                <small>
                  {n.authorName} · {date(n.signedAt)} EAT
                </small>
              </article>
            ))}
          </>
        )}
        {canSign && (
          <>
            <form onSubmit={sign} className="space-y-3">
              <label>
                <span className="label">{t("Clinical Note")}</span>
                <textarea
                  className="field"
                  required
                  maxLength={10000}
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                />
              </label>
              <p className="text-sm text-muted">
                {t(
                  "Signed notes cannot be edited or deleted. Add a new note for corrections.",
                )}
              </p>
              <button className="primary" disabled={busy}>
                {t("Sign note")}
              </button>
            </form>
            <form onSubmit={discharge} className="space-y-3 border-t pt-4">
              <label>
                <span className="label">{t("Discharge Summary")}</span>
                <textarea
                  className="field"
                  required
                  maxLength={10000}
                  value={summary}
                  onChange={(e) => setSummary(e.target.value)}
                />
              </label>
              <button className="primary" disabled={busy}>
                {t("Discharge patient")}
              </button>
            </form>
          </>
        )}
        {encounter.dischargeSummary && (
          <section>
            <h3>{t("Discharge Summary")}</h3>
            <p className="whitespace-pre-wrap">{encounter.dischargeSummary}</p>
          </section>
        )}
      </div>
    </Editor>
  );
}
