"use client";
import { useCallback, useState } from "react";
import { Search, Eye, Filter } from "lucide-react";
import { api } from "@/lib/api";
import { useLanguage } from "./language";
import { useIdentity } from "./workspace";
import {
  useResource,
  Status,
  Table,
  Paging,
  errorText,
} from "./connected-scheduling";
import { EncounterDetails, type Encounter } from "./connected-clinical";
type Case = {
  id: string;
  number: number;
  patientId: string;
  patientName: string;
  doctorId: string;
  doctorName: string;
};
type Bed = { id: string; name: string; type: string; available: boolean };
export function EncounterRegister({ kind }: { kind: "ipd" | "opd" }) {
  const { t } = useLanguage();
  const identity = useIdentity();
  const [page, setPage] = useState(1),
    [search, setSearch] = useState(""),
    [filter, setFilter] = useState(false),
    [status, setStatus] = useState("All"),
    [create, setCreate] = useState(false),
    [selected, setSelected] = useState<Encounter | null>(null);
  const load = useCallback(
    () =>
      api<{ encounters: Encounter[] }>(`encounters?kind=${kind}&page=${page}`),
    [kind, page],
  );
  const resource = useResource(load);
  const rows = (resource.data?.encounters || []).filter(
    (e) =>
      `${e.number} ${e.patientName} ${e.doctorName}`
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (status === "All" || e.status === status),
  );
  const person = (name: string, email?: string) => (
    <span className="record-link">
      <span className="avatar">
        {name
          .split(" ")
          .map((s) => s[0])
          .slice(0, 2)
          .join("")}
      </span>
      <span>
        {name}
        {email && <small>{email}</small>}
      </span>
    </span>
  );
  if (create)
    return (
      <EncounterRegistration
        kind={kind}
        onClose={() => setCreate(false)}
        onSaved={() => {
          resource.reload();
          setCreate(false);
        }}
      />
    );
  return (
    <section>
      <div className="table-toolbar">
        <label className="table-search">
          <Search size={16} />
          <input
            aria-label="Search loaded encounters"
            placeholder={t("Search")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </label>
        <div className="flex gap-3">
          {kind === "ipd" && (
            <button
              className="primary"
              aria-label="Filter encounters"
              aria-expanded={filter}
              onClick={() => setFilter(!filter)}
            >
              <Filter size={18} />
            </button>
          )}
          {identity?.permissions.includes("clinical.admit") && (
            <button className="primary" onClick={() => setCreate(true)}>
              {t(`New ${kind.toUpperCase()} Patient`)}
            </button>
          )}
        </div>
      </div>
      {filter && (
        <label className="block mb-5">
          {t("Status")}
          <select
            className="field"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            {["All", "active", "discharged"].map((s) => (
              <option key={s}>{s}</option>
            ))}
          </select>
        </label>
      )}
      <Status {...resource} empty={false} />
      <Table
        head={
          kind === "ipd"
            ? [
                "IPD No",
                "Patient",
                "Doctor",
                "Admission Date",
                "Bed",
                "Bill Status",
                "Action",
              ]
            : [
                "OPD No",
                "Patient",
                "Doctor",
                "Appointment Date",
                "Standard Charge",
                "Payment Mode",
                "Total Visits",
                "Status",
                "Action",
              ]
        }
      >
        {rows.map((row) => (
          <tr key={row.id}>
            <td>
              <button
                className="badge card-id"
                onClick={() => setSelected(row)}
              >
                {kind.toUpperCase()}-{String(row.number).padStart(6, "0")}
              </button>
            </td>
            <td>{person(row.patientName, row.patientEmail)}</td>
            <td>{person(row.doctorName, row.doctorEmail)}</td>
            <td>
              <span className="badge card-id encounter-date">
                {new Date(row.admittedAt).toLocaleTimeString("en-GB", {
                  timeZone: "Africa/Addis_Ababa",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
                <small>
                  {new Date(row.admittedAt).toLocaleDateString("en-GB", {
                    timeZone: "Africa/Addis_Ababa",
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  })}
                </small>
              </span>
            </td>
            {kind === "ipd" ? (
              <td>{row.bedName || "N/A"}</td>
            ) : (
              <>
                <td>
                  {row.intake
                    ? `${(row.intake.standardChargeMinor / 100).toFixed(2)} ETB`
                    : "N/A"}
                </td>
                <td>
                  <span className="badge">
                    {row.intake?.paymentMode || "N/A"}
                  </span>
                </td>
              </>
            )}
            {kind === "opd" && (
              <td>
                <span className="badge card-id">{row.totalVisits ?? 0}</span>
              </td>
            )}
            <td>
              <span
                className={`badge ${kind === "ipd" && row.billStatus === "Unpaid" ? "status-rejected" : ""}`}
              >
                {t(kind === "ipd" ? row.billStatus || "Unbilled" : row.status)}
              </span>
            </td>
            <td>
              <button
                className="text-brand"
                aria-label={`View ${row.patientName}`}
                onClick={() => setSelected(row)}
              >
                <Eye size={18} />
              </button>
            </td>
          </tr>
        ))}
        {!rows.length && !resource.loading && (
          <tr>
            <td colSpan={kind === "ipd" ? 7 : 9}>{t("No records found")}</td>
          </tr>
        )}
      </Table>
      <Paging
        page={page}
        setPage={setPage}
        count={resource.data?.encounters.length || 0}
        loading={resource.loading}
      />
      {selected && (
        <EncounterDetails
          encounter={selected}
          onClose={() => setSelected(null)}
          onSaved={resource.reload}
        />
      )}
    </section>
  );
}
function EncounterRegistration({
  kind,
  onClose,
  onSaved,
}: {
  kind: "ipd" | "opd";
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useLanguage();
  const [patient, setPatient] = useState(""),
    [caseId, setCase] = useState(""),
    [bedType, setBedType] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [key, setKey] = useState<{ body: string; key: string } | null>(null);
  const load = useCallback(async () => {
    const cases: Case[] = [],
      beds: Bed[] = [];
    for (let page = 1; ; page++) {
      const batch = await api<{ cases: Case[] }>(`cases?page=${page}`);
      cases.push(...batch.cases);
      if (batch.cases.length < 25) break;
    }
    if (kind === "ipd")
      for (let page = 1; ; page++) {
        const batch = await api<{ beds: Bed[] }>(`beds?page=${page}`);
        beds.push(...batch.beds);
        if (batch.beds.length < 25) break;
      }
    return { cases, beds };
  }, [kind]);
  const resource = useResource(load);
  const cases = resource.data?.cases || [],
    beds = resource.data?.beds || [];
  const chosen = cases.find((c) => c.id === caseId);
  const [initialTime] = useState(() =>
    new Date(Date.now() + 3 * 3600000).toISOString().slice(0, 16),
  );
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const f = Object.fromEntries(new FormData(e.currentTarget));
      const charge = String(f.charge || "0");
      if (!/^\d{1,8}(\.\d{1,2})?$/.test(charge))
        throw Error("Enter a charge with at most two decimal places.");
      const [whole, fraction = ""] = charge.split(".");
      const payload = {
        kind,
        caseId,
        bedId: kind === "ipd" ? f.bedId : "",
        admittedAt: new Date(String(f.admittedAt) + ":00+03:00").toISOString(),
        symptoms: f.symptoms,
        intake: {
          telephone: String(f.telephone || ""),
          taxReference: String(f.taxReference || ""),
          height: Number(f.height || 0),
          weight: Number(f.weight || 0),
          bloodPressure: f.bloodPressure,
          notes: f.notes,
          identificationNumber: String(f.identificationNumber || ""),
          reference: String(f.reference || ""),
          oldPatient: f.oldPatient === "on",
          standardChargeMinor:
            Number(whole) * 100 + Number(fraction.padEnd(2, "0")),
          paymentMode: String(f.paymentMode || ""),
        },
      };
      const body = JSON.stringify(payload),
        requestKey = key?.body === body ? key.key : crypto.randomUUID();
      setKey({ body, key: requestKey });
      await api("encounters", {
        method: "POST",
        headers: { "Idempotency-Key": requestKey },
        body,
      });
      onSaved();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  const input = (
    label: string,
    name: string,
    type = "text",
    required = false,
    extra: React.InputHTMLAttributes<HTMLInputElement> = {},
  ) => (
    <label>
      <span className="label">
        {t(label)}: {required && <b className="text-red-500">*</b>}
      </span>
      <input
        className="field"
        aria-label={t(label)}
        name={name}
        type={type}
        required={required}
        {...extra}
      />
    </label>
  );
  const text = (label: string, name: string, required = false) => (
    <label className="intake-wide">
      <span className="label">
        {t(label)}: {required && <b className="text-red-500">*</b>}
      </span>
      <textarea
        className="field"
        aria-label={t(label)}
        rows={3}
        name={name}
        required={required}
        maxLength={name === "notes" ? 4000 : 2000}
      />
    </label>
  );
  return (
    <section>
      <div className="page-heading">
        <h1>{t(`New ${kind.toUpperCase()} Patient`)}</h1>
        <button className="secondary" disabled={busy} onClick={onClose}>
          {t("Back")}
        </button>
      </div>
      <form
        className={`legacy-card encounter-intake intake-${kind}`}
        onSubmit={save}
      >
        {error && (
          <p className="error mb-5" role="alert">
            {t(error)}
          </p>
        )}
        <Status {...resource} empty={false} />
        <div className="intake-top">
          <label>
            <span className="label">
              {t("Patient")}: <b className="text-red-500">*</b>
            </span>
            <select
              className="field"
              required
              aria-label="Patient"
              value={patient}
              onChange={(e) => {
                setPatient(e.target.value);
                setCase("");
              }}
            >
              <option value="">{t("Select Patient")}</option>
              {[...new Map(cases.map((c) => [c.patientId, c])).values()].map(
                (c) => (
                  <option key={c.patientId} value={c.patientId}>
                    {c.patientName}
                  </option>
                ),
              )}
            </select>
          </label>
          <label>
            <span className="label">
              {t("Case")}: <b className="text-red-500">*</b>
            </span>
            <select
              aria-label="Patient Case"
              className="field"
              required
              disabled={!patient}
              value={caseId}
              onChange={(e) => setCase(e.target.value)}
            >
              <option value="">{t("Choose Case")}</option>
              {cases
                .filter((c) => c.patientId === patient)
                .map((c) => (
                  <option key={c.id} value={c.id}>
                    CASE-{c.number}
                  </option>
                ))}
            </select>
          </label>
          {input(`${kind.toUpperCase()} No`, "number", "text", false, {
            readOnly: true,
            value: t("Assigned on save"),
          })}
          {input("Height", "height", "number", false, {
            defaultValue: 0,
            min: 0,
            max: 300,
            step: 0.01,
          })}
          {input("Weight", "weight", "number", false, {
            defaultValue: 0,
            min: 0,
            max: 1000,
            step: 0.01,
          })}
          {input("Blood Pressure", "bloodPressure", "text", false, {
            maxLength: 40,
            placeholder: "120/80",
          })}
          {input(
            kind === "ipd" ? "Admission Date" : "Appointment Date",
            "admittedAt",
            "datetime-local",
            true,
            { defaultValue: initialTime },
          )}
          <label>
            <span className="label">
              {t("Doctor")}: <b className="text-red-500">*</b>
            </span>
            <select
              className="field"
              required
              value={chosen?.doctorId || ""}
              disabled={!chosen}
              onChange={() => {}}
            >
              <option value="">{t("Select Doctor")}</option>
              {chosen && (
                <option value={chosen.doctorId}>{chosen.doctorName}</option>
              )}
            </select>
          </label>
          {kind === "ipd" ? (
            <>
              <label>
                <span className="label">
                  {t("Bed Type")}: <b className="text-red-500">*</b>
                </span>
                <select
                  className="field"
                  required
                  aria-label="Bed Type"
                  value={bedType}
                  onChange={(e) => setBedType(e.target.value)}
                >
                  <option value="">{t("Select Bed Type")}</option>
                  {[
                    ...new Set(
                      beds.filter((b) => b.available).map((b) => b.type),
                    ),
                  ].map((type) => (
                    <option key={type}>{type}</option>
                  ))}
                </select>
              </label>
              <label>
                <span className="label">
                  {t("Bed")}: <b className="text-red-500">*</b>
                </span>
                <select
                  aria-label="Bed"
                  key={bedType}
                  className="field"
                  name="bedId"
                  required
                  disabled={!bedType}
                  defaultValue=""
                >
                  <option value="">{t("Choose Bed")}</option>
                  {beds
                    .filter((b) => b.available && b.type === bedType)
                    .map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.name}
                      </option>
                    ))}
                </select>
              </label>
            </>
          ) : (
            <>
              {input("Standard Charge (ETB)", "charge", "text", true, {
                inputMode: "decimal",
                pattern: "[0-9]{1,8}([.][0-9]{1,2})?",
              })}
              <label>
                <span className="label">
                  {t("Payment Mode")}: <b className="text-red-500">*</b>
                </span>
                <select
                  className="field"
                  name="paymentMode"
                  required
                  defaultValue=""
                >
                  <option value="">{t("Choose Payment")}</option>
                  {[
                    ["cash", "Cash"],
                    ["bank_transfer", "Bank transfer"],
                    ["card", "Card"],
                    ["other", "Other"],
                  ].map(([value, label]) => (
                    <option key={value} value={value}>
                      {t(label)}
                    </option>
                  ))}
                </select>
              </label>
            </>
          )}
        </div>
        {kind === "ipd" && (
          <div className="intake-secondary">
            {input("ID Number", "identificationNumber", "text", false, {
              maxLength: 200,
            })}
            {input("IPD Unique Id For Reference", "reference", "text", false, {
              maxLength: 200,
            })}
          </div>
        )}
        <div className="intake-text">
          {text("Symptoms", "symptoms")}
          {text("Notes", "notes")}
        </div>
        {kind === "opd" && (
          <div className="intake-secondary">
            {input("Tel", "telephone", "tel", true, { maxLength: 80 })}
            {input("TAX", "taxReference", "text", true, { maxLength: 200 })}
          </div>
        )}
        <label className="block mt-5">
          <span className="label">{t("Is Old Patient")}:</span>
          <input
            type="checkbox"
            className="reference-switch"
            name="oldPatient"
            role="switch"
          />
        </label>
        {!resource.loading && !cases.length && (
          <p className="mt-5">
            {t("Create a patient case before registering an admission.")}{" "}
            <a className="text-brand" href="/modules/patient-cases">
              {t("Patient Cases")}
            </a>
          </p>
        )}
        <footer className="modal-footer">
          <button
            className="primary"
            disabled={busy || resource.loading || !!resource.error}
          >
            {t(busy ? "Saving..." : "Save")}
          </button>
          <button
            className="secondary"
            type="button"
            disabled={busy}
            onClick={onClose}
          >
            {t("Cancel")}
          </button>
        </footer>
      </form>
    </section>
  );
}
