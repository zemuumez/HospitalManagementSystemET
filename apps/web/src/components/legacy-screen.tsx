"use client";
import { useEffect, useState, useRef } from "react";
import {
  ArrowDownUp,
  BedDouble,
  ChevronLeft,
  ChevronRight,
  Download,
  Eye,
  Pencil,
  Plus,
  Search,
  Trash2,
  X,
} from "lucide-react";
import { useLanguage } from "./language";
import { Modal } from "./modal";
import {
  changeClinicalField,
  dependentOptions,
  validateClinicalPreview,
} from "@/lib/clinical-preview";
import odontogramSvg from "@/lib/odontogram-svg.json";
import {
  type Field,
  type Screen,
  type PreviewRow,
  optionsFor,
  seedRows,
  screens,
  people,
} from "@/lib/legacy";

function FieldInput({
  field: f,
  value,
  onChange,
  values = {},
}: {
  values?: Record<string, string>;
  field: Field;
  value: string;
  onChange: (v: string) => void;
}) {
  const { t } = useLanguage();
  const type = /password/.test(f.key) ? "password" : f.type;
  const customOptions = f.key.startsWith("custom_")
    ? (f.source || "")
        .split(/[,\n]/)
        .map((v) => v.trim())
        .filter(Boolean)
    : undefined;
  if (f.key === "gender")
    return (
      <fieldset>
        <legend className="label">
          Gender: <b className="text-red-500">*</b>
        </legend>
        <div className="flex gap-6 pt-3">
          {["Male", "Female"].map((g) => (
            <label className="flex gap-2 items-center" key={g}>
              <input
                type="radio"
                name="gender"
                value={g}
                checked={value === g}
                onChange={() => onChange(g)}
                required={f.required}
              />
              {g}
            </label>
          ))}
        </div>
      </fieldset>
    );
  if (f.key === "status")
    return (
      <label>
        <span className="label">Status:</span>
        <input
          className="legacy-checkbox"
          type="checkbox"
          checked={value !== "Inactive"}
          onChange={(e) => onChange(e.target.checked ? "Active" : "Inactive")}
        />
      </label>
    );
  return (
    <label className={type === "textarea" ? "form-span" : ""}>
      <span className="label">
        {t(f.label)}: {f.required && <b className="text-red-500">*</b>}
      </span>
      {type === "select" ? (
        <select
          className="field"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={f.required}
        >
          <option value="">Select {t(f.label)}</option>
          {(
            customOptions ||
            dependentOptions(f.key, values) ||
            optionsFor(f)
          ).map((o) => (
            <option key={o} value={o}>
              {t(o)}
            </option>
          ))}
        </select>
      ) : type === "textarea" ? (
        <textarea
          className="field"
          rows={3}
          value={value}
          required={f.required}
          onChange={(e) => onChange(e.target.value)}
        />
      ) : (
        <input
          className="field"
          type={
            [
              "date",
              "number",
              "email",
              "password",
              "tel",
              "file",
              "color",
            ].includes(type)
              ? type
              : "text"
          }
          {...(type === "file" ? {} : { value })}
          autoComplete="off"
          min={type === "number" ? 0 : undefined}
          required={f.required}
          onChange={(e) =>
            onChange(
              type === "file"
                ? e.target.files?.[0]?.name || ""
                : e.target.value,
            )
          }
          placeholder={t(f.label)}
        />
      )}
    </label>
  );
}

export function LegacyScreen({
  screen: s,
  scope = "",
}: {
  screen: Screen;
  scope?: string;
}) {
  const { t } = useLanguage();
  const [rows, setRows] = useState<PreviewRow[]>(() => seedRows(s));
  const [ready, setReady] = useState(false);
  const [customFields, setCustomFields] = useState<Field[]>([]);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState("All");
  const [sort, setSort] = useState("");
  const [descending, setDescending] = useState(false);
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(10);
  const [editing, setEditing] = useState<PreviewRow | null>(null);
  const [view, setView] = useState<PreviewRow | null>(null);
  const [deleting, setDeleting] = useState<PreviewRow | null>(null);
  const [notice, setNotice] = useState("");
  const [formError, setFormError] = useState("");
  const [tab, setTab] = useState("Overview");
  const [calendar, setCalendar] = useState(s.id === "appointment-calendars");
  const [month, setMonth] = useState(9);
  const [year, setYear] = useState(2026);
  const storage = `hms-preview-v2:${s.id}:${scope}`;
  useEffect(() => {
    try {
      const saved = sessionStorage.getItem(storage);
      if (saved) setRows(JSON.parse(saved));
    } catch {}
    try {
      const definitions: PreviewRow[] = JSON.parse(
        sessionStorage.getItem("hms-preview-v2:add-custom-fields:") || "[]",
      );
      setCustomFields(
        definitions
          .filter((r) => r.values.module === s.group && r.status !== "Inactive")
          .map((r) => ({
            key: `custom_${r.id}`,
            label: r.values.field_name,
            type:
              (
                {
                  Number: "number",
                  Date: "date",
                  Textarea: "textarea",
                  Select: "select",
                } as Record<string, string>
              )[r.values.field_type] || "text",
            required: r.values.is_required === "Yes",
            source: r.values.values || "",
          })),
      );
    } catch {}
    setReady(true);
  }, [storage]);
  useEffect(() => {
    if (ready)
      try {
        sessionStorage.setItem(storage, JSON.stringify(rows));
      } catch {}
  }, [rows, ready, storage]);
  if (!ready)
    return (
      <div className="legacy-card" aria-busy="true">
        Loading {s.title}…
      </div>
    );
  const columns = s.columns
    .filter((c) => !/password|image|file/i.test(c))
    .slice(0, 7);
  const filtered = rows
    .filter(
      (r) =>
        (filter === "All" || r.status === filter) &&
        Object.values(r.values)
          .join(" ")
          .toLowerCase()
          .includes(search.toLowerCase()),
    )
    .sort((a, b) =>
      sort
        ? (a.values[sort] || "").localeCompare(b.values[sort] || "") *
          (descending ? -1 : 1)
        : 0,
    );
  const pages = Math.max(1, Math.ceil(filtered.length / size));
  const current = Math.min(page, pages);
  const shown = filtered.slice((current - 1) * size, current * size);
  const FormContainer = s.fields.length > 8 ? FullPageForm : Modal;
  const DetailContainer = [
    "patients",
    "doctors",
    "ipd-patient-departments",
    "opd-patient-departments",
    "invoices",
    "bills",
    "medicine-bills",
    "purchase-medicines",
  ].includes(s.id)
    ? FullPageForm
    : Modal;
  const readonly = [
    "appointment-transaction",
    "payment-reports",
    "used-medicine",
    "manual-bill-payments",
    "bed-status",
    "hospital-schedule",
  ].includes(s.id);
  const newRow = () => {
    setFormError("");
    window.scrollTo(0, 0);
    setEditing({
      id: `DEMO-${crypto.randomUUID().slice(0, 8)}`,
      values: Object.fromEntries(
        s.fields
          .filter((f) => ["department_id", "bed_type_id"].includes(f.key))
          .map((f) => [f.key, ""]),
      ),
      status: "Active",
    });
    setNotice("");
  };
  function save(e: React.FormEvent) {
    e.preventDefault();
    if (!editing) return;
    if (
      editing.values.password !== editing.values.password_confirmation &&
      editing.values.password_confirmation !== undefined
    ) {
      setFormError("Password confirmation must match the password.");
      return;
    }
    const validation = validateClinicalPreview(s.id, editing.values);
    if (validation) {
      setFormError(validation);
      return;
    }
    if (
      (s.id === "appointments" || s.id === "appointment-calendars") &&
      rows.some(
        (r) =>
          r.id !== editing.id &&
          r.values.doctor_id === editing.values.doctor_id &&
          r.values.opd_date === editing.values.opd_date &&
          r.values.timeslot === editing.values.timeslot,
      )
    ) {
      setFormError("This appointment time is already booked in the preview.");
      return;
    }
    const values = { ...editing.values };
    if (values._lineItems) {
      try {
        const billing = JSON.parse(values._lineItems);
        values.Amount = (
          Math.max(
            0,
            billing.items.reduce(
              (sum: number, i: { qty: number; rate: number }) =>
                sum + i.qty * i.rate,
              0,
            ) - billing.discount,
          ) *
          (1 + billing.tax / 100)
        ).toFixed(2);
      } catch {}
    }
    s.fields.forEach((f) => {
      if (f.type === "password") delete values[f.key];
      else values[f.label] = values[f.key] || "";
    });
    if (s.id === "patients")
      values.Patients =
        `${values.first_name || ""} ${values.last_name || ""}`.trim();
    const record = {
      ...editing,
      values,
      status: values.status || editing.status,
    };
    setRows((old) =>
      old.some((r) => r.id === record.id)
        ? old.map((r) => (r.id === record.id ? record : r))
        : [record, ...old],
    );
    setEditing(null);
    setNotice("Saved in this frontend preview. No hospital data was changed.");
  }
  function exportRows() {
    const quote = (v: string) =>
      `"${(/^[=+\-@]/.test(v) ? "'" : "") + v.replaceAll('"', '""')}"`;
    const csv = [
      columns,
      ...filtered.map((r) => columns.map((c) => r.values[c] || "")),
    ]
      .map((r) => r.map(quote).join(","))
      .join("\r\n");
    const url = URL.createObjectURL(new Blob([csv], { type: "text/csv" }));
    const a = document.createElement("a");
    a.href = url;
    a.download = `preview-${s.id}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }
  const detailTabs =
    s.id === "ipd-patient-departments"
      ? [
          "Overview",
          "Diagnosis",
          "Consultant Register",
          "Operations",
          "Charges",
          "Prescriptions",
          "Timeline",
          "Payment",
          "Bill",
          "Discharge",
        ]
      : s.id === "opd-patient-departments"
        ? ["Overview", "Visits", "Diagnosis", "Prescriptions", "Timeline"]
        : s.id === "patients"
          ? [
              "Overview",
              "Cases",
              "Admissions",
              "Appointments",
              "Prescriptions",
              "Documents",
              "Invoices",
              "Vaccinations",
            ]
          : ["Overview"];
  return (
    <section>
      <div className="page-heading">
        <h1>{t(s.title)}</h1>
        <div className="flex gap-2">
          <button className="secondary" onClick={exportRows}>
            <Download size={16} />
            {t("Export")}
          </button>
          <button className="primary" hidden={readonly} onClick={newRow}>
            <Plus size={16} />
            {t("New")} {t(s.title.replace(/ies$/, "y").replace(/s$/, ""))}
          </button>
        </div>
      </div>
      {notice && (
        <p className="success mb-5" role="status">
          {notice}
        </p>
      )}
      {s.id === "bed-status" ? (
        <BedBoard
          onSelect={(r) => {
            setView(r);
            setTab("Overview");
          }}
        />
      ) : s.id === "odontogram" ? (
        <Odontogram />
      ) : s.id === "hospital-schedule" ? (
        <Schedule />
      ) : s.id === "settings" ||
        s.id === "front-settings" ||
        s.id === "payment-gateway" ? (
        <div className="legacy-card">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              setNotice("Settings saved for preview only.");
            }}
          >
            <div className="legacy-form">
              {[...s.fields, ...customFields].map((f) => (
                <FieldInput
                  key={f.key}
                  field={f}
                  value={rows[0]?.values[f.key] || ""}
                  onChange={(v) =>
                    setRows((old) =>
                      old.map((r, i) =>
                        i === 0
                          ? { ...r, values: { ...r.values, [f.key]: v } }
                          : r,
                      ),
                    )
                  }
                />
              ))}
            </div>
            <button className="primary mt-6">{t("Save")}</button>
          </form>
        </div>
      ) : (
        <>
          <div className="table-toolbar">
            <div className="table-search">
              <Search size={18} />
              <input
                aria-label={`Search ${s.title}`}
                placeholder={t("Search")}
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
              />
            </div>
            <div className="flex flex-wrap gap-3">
              {s.group === "Appointments" && (
                <button
                  className="secondary"
                  onClick={() => setCalendar(!calendar)}
                >
                  {calendar ? "List View" : "Calendar View"}
                </button>
              )}
              <select
                aria-label="Filter by status"
                className="field !w-auto"
                value={filter}
                onChange={(e) => {
                  setFilter(e.target.value);
                  setPage(1);
                }}
              >
                <option value="All">{t("All")}</option>
                <option value="Active">{t("Active")}</option>
                <option value="Inactive">{t("Inactive")}</option>
                <option value="Pending">{t("Pending")}</option>
                <option value="Completed">{t("Completed")}</option>
              </select>
            </div>
          </div>
          {calendar ? (
            <div className="legacy-card">
              <div className="page-heading">
                <h2>
                  {new Date(year, month).toLocaleDateString("en", {
                    month: "long",
                    year: "numeric",
                  })}
                </h2>
                <div className="flex gap-2">
                  <button
                    className="secondary"
                    aria-label="Previous month"
                    onClick={() => {
                      const d = new Date(year, month - 1);
                      setMonth(d.getMonth());
                      setYear(d.getFullYear());
                    }}
                  >
                    <ChevronLeft size={16} />
                  </button>
                  <button
                    className="secondary"
                    aria-label="Next month"
                    onClick={() => {
                      const d = new Date(year, month + 1);
                      setMonth(d.getMonth());
                      setYear(d.getFullYear());
                    }}
                  >
                    <ChevronRight size={16} />
                  </button>
                </div>
              </div>
              <div className="calendar-grid">
                {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
                  <strong key={d}>{d}</strong>
                ))}
                {Array.from(
                  { length: new Date(year, month, 1).getDay() },
                  (_, i) => (
                    <div key={`empty-${i}`} />
                  ),
                )}
                {Array.from(
                  { length: new Date(year, month + 1, 0).getDate() },
                  (_, i) => (
                    <button
                      className="calendar-day"
                      key={i}
                      onClick={() => {
                        newRow();
                      }}
                    >
                      <span>{i + 1}</span>
                      {month === 9 && year === 2026 && i < rows.length && (
                        <small>
                          {people[i % 6]}
                          <br />
                          09:00 · Consultation
                        </small>
                      )}
                    </button>
                  ),
                )}
              </div>
            </div>
          ) : (
            <div className="legacy-table-wrap">
              <table className="legacy-table">
                <thead>
                  <tr>
                    {columns.map((c) => (
                      <th key={c}>
                        <button
                          onClick={() => {
                            setSort(c);
                            setDescending(!descending);
                          }}
                        >
                          {t(c)}
                          <ArrowDownUp size={12} />
                        </button>
                      </th>
                    ))}
                    {!columns.some((c) => /status/i.test(c)) && (
                      <th>{t("Status")}</th>
                    )}
                    <th>{t("Action")}</th>
                  </tr>
                </thead>
                <tbody>
                  {shown.map((r) => (
                    <tr key={r.id}>
                      {columns.map((c, i) => (
                        <td key={c}>
                          {/status/i.test(c) ? (
                            <button
                              aria-label={`Toggle status ${r.id}`}
                              className={`status-switch ${r.status === "Active" ? "on" : ""}`}
                              onClick={() =>
                                setRows((old) =>
                                  old.map((x) =>
                                    x.id === r.id
                                      ? {
                                          ...x,
                                          status:
                                            x.status === "Active"
                                              ? "Inactive"
                                              : "Active",
                                        }
                                      : x,
                                  ),
                                )
                              }
                            />
                          ) : i === 0 ? (
                            <button
                              className="record-link"
                              onClick={() => {
                                setView(r);
                                setTab("Overview");
                              }}
                            >
                              {/patient|doctor|user|employee|nurse|pharmacist|receptionist/i.test(
                                c,
                              ) && (
                                <span className="avatar">
                                  {(r.values[c] || "S").slice(0, 1)}
                                </span>
                              )}
                              <span>
                                {r.values[c] || r.id}
                                {s.id === "patients" && (
                                  <small>{r.values.email}</small>
                                )}
                              </span>
                            </button>
                          ) : (
                            r.values[c] || "—"
                          )}
                        </td>
                      ))}
                      {!columns.some((c) => /status/i.test(c)) && (
                        <td>
                          <span
                            className={`badge ${r.status === "Inactive" ? "muted" : ""}`}
                          >
                            {t(r.status)}
                          </span>
                        </td>
                      )}
                      <td>
                        <div className="row-actions">
                          <button
                            title="View"
                            aria-label={`View ${r.id}`}
                            onClick={() => {
                              setView(r);
                              setTab("Overview");
                            }}
                          >
                            <Eye size={17} />
                          </button>
                          <button
                            title="Edit"
                            aria-label={`Edit ${r.id}`}
                            onClick={() =>
                              setEditing({ ...r, values: { ...r.values } })
                            }
                          >
                            <Pencil size={16} />
                          </button>
                          <button
                            title="Delete"
                            aria-label={`Delete ${r.id}`}
                            onClick={() => setDeleting(r)}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {!shown.length && (
                    <tr>
                      <td colSpan={columns.length + 2} className="empty-state">
                        No matching records found
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
              <div className="table-footer">
                <span>
                  Showing {filtered.length ? (current - 1) * size + 1 : 0} to{" "}
                  {Math.min(current * size, filtered.length)} of{" "}
                  {filtered.length} results
                </span>
                <div className="flex items-center gap-3">
                  <select
                    className="field !w-auto"
                    aria-label="Rows per page"
                    value={size}
                    onChange={(e) => {
                      setSize(Number(e.target.value));
                      setPage(1);
                    }}
                  >
                    {[5, 10, 25, 50].map((n) => (
                      <option key={n}>{n}</option>
                    ))}
                  </select>
                  <button
                    aria-label="Previous page"
                    disabled={current === 1}
                    onClick={() => setPage(current - 1)}
                  >
                    <ChevronLeft size={17} />
                  </button>
                  <span className="page-number">{current}</span>
                  <button
                    aria-label="Next page"
                    disabled={current === pages}
                    onClick={() => setPage(current + 1)}
                  >
                    <ChevronRight size={17} />
                  </button>
                </div>
              </div>
            </div>
          )}
        </>
      )}
      {editing && (
        <FormContainer
          titleId="record-form-title"
          onClose={() => setEditing(null)}
        >
          <form onSubmit={save}>
            <div className="modal-heading">
              <h2 id="record-form-title">
                {rows.some((r) => r.id === editing.id) ? "Edit" : "New"}{" "}
                {t(s.title.replace(/s$/, ""))}
              </h2>
              <button
                type="button"
                aria-label="Close form"
                onClick={() => setEditing(null)}
              >
                <X size={20} />
              </button>
            </div>
            <div className="modal-content">
              {formError && (
                <p className="error mb-4" role="alert">
                  {formError}
                </p>
              )}
              <p className="form-preview-note">
                Preview form · Use sample information only
              </p>
              <div className="legacy-form">
                {s.fields.map((f) => (
                  <FieldInput
                    key={f.key}
                    field={f}
                    values={editing.values}
                    value={editing.values[f.key] || ""}
                    onChange={(v) =>
                      setEditing({
                        ...editing,
                        values: changeClinicalField(editing.values, f.key, v),
                      })
                    }
                  />
                ))}
              </div>
              {s.group === "Patient ID Card" && (
                <PatientCard values={editing.values} />
              )}
              {/invoice|bill|purchase|prescription|package/.test(s.id) && (
                <LineItems
                  prescription={s.id.includes("prescription")}
                  purchase={s.id.includes("purchase")}
                  saved={editing.values._lineItems}
                  onChange={(v) =>
                    setEditing((old) =>
                      old
                        ? { ...old, values: { ...old.values, _lineItems: v } }
                        : old,
                    )
                  }
                />
              )}
            </div>
            <div className="modal-footer">
              <button className="primary">{t("Save")}</button>
              <button
                className="secondary"
                type="button"
                onClick={() => setEditing(null)}
              >
                {t("Cancel")}
              </button>
            </div>
          </form>
        </FormContainer>
      )}
      {view && (
        <DetailContainer
          titleId="record-detail-title"
          onClose={() => setView(null)}
        >
          <div className="modal-heading">
            <h2 id="record-detail-title">{s.title} Details</h2>
            <button aria-label="Close details" onClick={() => setView(null)}>
              <X size={20} />
            </button>
          </div>
          <div className="modal-content">
            <div className="detail-summary">
              <span className="avatar large">
                {(view.values[columns[0]] || "S").slice(0, 1)}
              </span>
              <div>
                <h2>{view.values[columns[0]] || view.id}</h2>
                <p>
                  {view.id} <span className="badge">{t(view.status)}</span>
                </p>
              </div>
            </div>
            <div className="detail-tabs">
              {detailTabs.map((tabName) => (
                <button
                  key={tabName}
                  className={tab === tabName ? "active" : ""}
                  onClick={() => setTab(tabName)}
                >
                  {t(tabName)}
                </button>
              ))}
            </div>
            {tab === "Overview" ? (
              <dl className="detail-grid">
                {s.fields
                  .filter((f) => f.type !== "password")
                  .map((f) => (
                    <div key={f.key}>
                      <dt>{t(f.label)}</dt>
                      <dd>{view.values[f.key] || "—"}</dd>
                    </div>
                  ))}
              </dl>
            ) : (
              <ClinicalTab
                key={tab}
                name={tab}
                patient={view.id}
                parent={s.id}
                admissionDate={view.values.admission_date}
              />
            )}
            {/invoice|bill|prescription/.test(s.id) && (
              <LineItems
                prescription={s.id.includes("prescription")}
                saved={view.values._lineItems}
                readOnly
              />
            )}
            {s.group === "Patient ID Card" && (
              <PatientCard values={view.values} />
            )}
          </div>
          <div className="modal-footer">
            <button
              className="primary"
              onClick={() => {
                setEditing({ ...view, values: { ...view.values } });
                setView(null);
              }}
            >
              {t("Edit")}
            </button>
            <button className="secondary" onClick={() => window.print()}>
              {t("Print")}
            </button>
            <button className="secondary" onClick={() => setView(null)}>
              {t("Close")}
            </button>
          </div>
        </DetailContainer>
      )}
      {deleting && (
        <Modal titleId="delete-title" onClose={() => setDeleting(null)}>
          <div className="modal-heading">
            <h2 id="delete-title">Delete preview record?</h2>
          </div>
          <div className="modal-content">
            Remove {deleting.id} from this preview?
          </div>
          <div className="modal-footer">
            <button
              className="primary !bg-red-500"
              onClick={() => {
                setRows((old) => old.filter((r) => r.id !== deleting.id));
                setDeleting(null);
                setNotice("Preview record deleted.");
              }}
            >
              {t("Delete")}
            </button>
            <button className="secondary" onClick={() => setDeleting(null)}>
              {t("Cancel")}
            </button>
          </div>
        </Modal>
      )}
    </section>
  );
}

function FullPageForm({
  children,
  onClose,
  titleId,
}: {
  children: React.ReactNode;
  onClose: () => void;
  titleId: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    ref.current?.scrollIntoView({ block: "start" });
    ref.current?.focus();
  }, []);
  return (
    <div
      ref={ref}
      tabIndex={-1}
      aria-labelledby={titleId}
      className="full-editor legacy-card"
      onKeyDown={(e) => {
        if (e.key === "Escape") {
          e.stopPropagation();
          onClose();
        }
      }}
    >
      {children}
    </div>
  );
}
function PatientCard({ values }: { values: Record<string, string> }) {
  return (
    <div className="patient-card-preview">
      <header
        style={{
          background: /^#[0-9a-f]{6}$/i.test(values.header_color || "")
            ? values.header_color
            : "#6571ff",
        }}
      >
        ULSHMS <span>Patient Identification Card</span>
      </header>
      <div>
        <span className="avatar large">A</span>
        <section>
          <strong>{values.patient || "Alex Morgan"}</strong>
          <p>Patient ID: DEMO-0001</p>
          {values.show_email !== "No" && <p>sample1@example.invalid</p>}
          {values.show_phone !== "No" && <p>+254700000100</p>}
          {values.show_blood_group !== "No" && <p>Blood Group: A+</p>}
        </section>
      </div>
      <footer>Sample card · Not valid for patient identification</footer>
    </div>
  );
}
function LineItems({
  prescription = false,
  purchase = false,
  saved,
  onChange,
  readOnly = false,
}: {
  prescription?: boolean;
  purchase?: boolean;
  saved?: string;
  onChange?: (v: string) => void;
  readOnly?: boolean;
}) {
  type Item = {
    id: number;
    name: string;
    qty: number;
    rate: number;
    lot?: string;
    expiry?: string;
    frequency?: string;
    instructions?: string;
  };
  const initial = () => {
    try {
      return saved ? JSON.parse(saved) : null;
    } catch {
      return null;
    }
  };
  const [items, setItems] = useState<Item[]>(
    () => initial()?.items || [{ id: 1, name: "", qty: 1, rate: 0 }],
  );
  const [discount, setDiscount] = useState<number>(
    () => initial()?.discount || 0,
  );
  const [tax, setTax] = useState<number>(() => initial()?.tax || 0);
  useEffect(() => {
    onChange?.(JSON.stringify({ items, discount, tax }));
  }, [items, discount, tax]);
  const subtotal = items.reduce((sum, i) => sum + i.qty * i.rate, 0);
  const total = Math.max(0, subtotal - discount) * (1 + tax / 100);
  const change = (id: number, key: string, value: string | number) =>
    setItems(items.map((i) => (i.id === id ? { ...i, [key]: value } : i)));
  return (
    <section className="line-items">
      <h3>{prescription ? "Medicines" : "Items"}</h3>
      {items.map((item, index) => (
        <div key={item.id}>
          <div className="line-item">
            <label>
              <span className="label">
                {prescription || purchase ? "Medicine" : "Item"}
              </span>
              <input
                className="field"
                disabled={readOnly}
                value={item.name}
                onChange={(e) => change(item.id, "name", e.target.value)}
              />
            </label>
            <label>
              <span className="label">Quantity</span>
              <input
                className="field"
                disabled={readOnly}
                type="number"
                min="1"
                value={item.qty}
                onChange={(e) => change(item.id, "qty", Number(e.target.value))}
              />
            </label>
            <label>
              <span className="label">{prescription ? "Days" : "Rate"}</span>
              <input
                className="field"
                disabled={readOnly}
                type="number"
                min="0"
                value={item.rate}
                onChange={(e) =>
                  change(item.id, "rate", Number(e.target.value))
                }
              />
            </label>
            {!readOnly && (
              <button
                type="button"
                aria-label={`Remove item ${index + 1}`}
                onClick={() => setItems(items.filter((i) => i.id !== item.id))}
              >
                <Trash2 size={17} />
              </button>
            )}
          </div>
          {purchase && (
            <div className="legacy-form mb-5">
              <label>
                <span className="label">Lot Number</span>
                <input
                  className="field"
                  disabled={readOnly}
                  value={item.lot || ""}
                  onChange={(e) => change(item.id, "lot", e.target.value)}
                />
              </label>
              <label>
                <span className="label">Expiry Date</span>
                <input
                  className="field"
                  disabled={readOnly}
                  type="date"
                  value={item.expiry || ""}
                  onChange={(e) => change(item.id, "expiry", e.target.value)}
                />
              </label>
            </div>
          )}
          {prescription && (
            <div className="legacy-form mb-5">
              <label>
                <span className="label">Dose Interval</span>
                <select
                  className="field"
                  disabled={readOnly}
                  value={item.frequency || ""}
                  onChange={(e) => change(item.id, "frequency", e.target.value)}
                >
                  <option value="">Select interval</option>
                  <option>Once daily</option>
                  <option>Twice daily</option>
                  <option>Three times daily</option>
                </select>
              </label>
              <label>
                <span className="label">Instructions</span>
                <input
                  className="field"
                  disabled={readOnly}
                  value={item.instructions || ""}
                  onChange={(e) =>
                    change(item.id, "instructions", e.target.value)
                  }
                />
              </label>
            </div>
          )}
        </div>
      ))}
      {!readOnly && (
        <button
          className="secondary"
          type="button"
          onClick={() =>
            setItems([...items, { id: Date.now(), name: "", qty: 1, rate: 0 }])
          }
        >
          <Plus size={16} />
          Add Item
        </button>
      )}
      {!prescription && (
        <div className="invoice-totals">
          <label>
            Discount{" "}
            <input
              className="field"
              disabled={readOnly}
              type="number"
              min="0"
              value={discount}
              onChange={(e) => setDiscount(Number(e.target.value))}
            />
          </label>
          <label>
            Tax (%){" "}
            <input
              className="field"
              disabled={readOnly}
              type="number"
              min="0"
              value={tax}
              onChange={(e) => setTax(Number(e.target.value))}
            />
          </label>
          <strong>Total: {total.toFixed(2)}</strong>
        </div>
      )}
    </section>
  );
}
function ClinicalTab({
  name,
  patient,
  parent,
  admissionDate,
}: {
  admissionDate?: string;
  name: string;
  patient: string;
  parent: string;
}) {
  const prefix = parent.startsWith("opd") ? "opd" : "ipd";
  const ids: Record<string, string> = {
    Diagnosis: `${prefix}-diagnoses`,
    "Consultant Register": "ipd-consultant-registers",
    Operations: "ipd-operation",
    Charges: "ipd-charges",
    Prescriptions:
      parent === "patients" ? "prescriptions" : `${prefix}-prescriptions`,
    Timeline: `${prefix}-timelines`,
    Payment: "ipd-payments",
    Bill: "ipd-bills",
    Cases: "patient-cases",
    Admissions: "patient-admissions",
    Appointments: "appointments",
    Documents: "documents",
    Invoices: "invoices",
    Vaccinations: "vaccinated-patients",
    Visits: "opd-patient-departments",
  };
  const screen = screens.find((s) => s.id === ids[name]);
  if (screen)
    return (
      <div className="clinical-tab">
        <LegacyScreen screen={screen} scope={patient} />
      </div>
    );
  return <Discharge patient={patient} admissionDate={admissionDate} />;
}
function Discharge({
  patient,
  admissionDate,
}: {
  patient: string;
  admissionDate?: string;
}) {
  const { t } = useLanguage();
  const [saved, setSaved] = useState(false);
  const [summary, setSummary] = useState<Record<string, string>>({});
  useEffect(() => {
    try {
      setSummary(
        JSON.parse(sessionStorage.getItem(`hms-discharge:${patient}`) || "{}"),
      );
    } catch {}
  }, [patient]);
  return (
    <form
      className="py-6"
      onSubmit={(e) => {
        e.preventDefault();
        sessionStorage.setItem(
          `hms-discharge:${patient}`,
          JSON.stringify(summary),
        );
        setSaved(true);
      }}
    >
      <div className="legacy-form">
        {[
          "Discharge Date",
          "Discharge Status",
          "Diagnosis",
          "Treatment Given",
          "Advice On Discharge",
          "Follow Up",
        ].map((label) => (
          <label key={label}>
            <span className="label">{label}</span>
            {label === "Discharge Date" ? (
              <input
                className="field"
                type="date"
                required
                min={admissionDate}
                value={summary[label] || ""}
                onChange={(e) =>
                  setSummary({ ...summary, [label]: e.target.value })
                }
              />
            ) : label === "Discharge Status" ? (
              <select
                className="field"
                value={summary[label] || "Recovered"}
                onChange={(e) =>
                  setSummary({ ...summary, [label]: e.target.value })
                }
              >
                <option>Recovered</option>
                <option>Referred</option>
                <option>Discharged on request</option>
                <option>Death</option>
              </select>
            ) : (
              <textarea
                className="field"
                rows={3}
                value={summary[label] || ""}
                onChange={(e) =>
                  setSummary({ ...summary, [label]: e.target.value })
                }
              />
            )}
          </label>
        ))}
      </div>
      <button className="primary mt-5">Save Discharge Summary</button>
      {saved && (
        <p className="success mt-4" role="status">
          Sample discharge summary saved for {patient}.
        </p>
      )}
    </form>
  );
}
function BedBoard({ onSelect }: { onSelect: (r: PreviewRow) => void }) {
  const { t } = useLanguage();
  const [filter, setFilter] = useState("All");
  return (
    <div className="legacy-card">
      <div className="page-heading">
        <h2>Bed Status</h2>
        <select
          className="field !w-auto"
          aria-label="Bed availability"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        >
          <option value="All">{t("All")}</option>
          <option>Available</option>
          <option>Occupied</option>
        </select>
      </div>
      {["General Ward", "Private Ward", "ICU"].map((ward, j) => (
        <section className="mb-8" key={ward}>
          <h3>{ward}</h3>
          <div className="bed-grid">
            {Array.from({ length: 8 }, (_, i) => ({
              id: `${j + 1}-${i + 1}`,
              occupied: i % 3 === 0,
            }))
              .filter(
                (b) =>
                  filter === "All" || b.occupied === (filter === "Occupied"),
              )
              .map((b) => (
                <button
                  key={b.id}
                  className={`bed-tile ${b.occupied ? "occupied" : ""}`}
                  onClick={() =>
                    onSelect({
                      id: b.id,
                      status: b.occupied ? "Occupied" : "Available",
                      values: {
                        Name: `${ward} — ${b.id}`,
                        name: `${ward} — ${b.id}`,
                        description: b.occupied
                          ? "Sample patient assignment"
                          : "Available bed",
                      },
                    })
                  }
                >
                  <BedDouble size={30} />
                  <strong>{b.id}</strong>
                  <small>{b.occupied ? "Occupied" : "Available"}</small>
                </button>
              ))}
          </div>
        </section>
      ))}
    </div>
  );
}
function Odontogram() {
  const { t } = useLanguage();
  const [selected, setSelected] = useState(1);
  const [conditions, setConditions] = useState<Record<number, string>>({});
  const [saved, setSaved] = useState(false);
  const choices = [
    ["Healthy", "#ffffff"],
    ["K", "#e91e63"],
    ["C", "#3f51b5"],
    ["Ce", "#009688"],
    ["D", "#795548"],
    ["KR", "#607d8b"],
    ["PS", "#ffc107"],
    ["IP", "#673ab7"],
  ];
  const svg = odontogramSvg.replace(
    /(<polygon id="Tooth(\d+)"\s+fill=")[^"]*(")/g,
    (_m, prefix, n, suffix) =>
      prefix +
      (choices.find((c) => c[0] === conditions[Number(n)])?.[1] || "#fff") +
      suffix,
  );
  return (
    <div className="legacy-card">
      <div className="legacy-form">
        <label>
          <span className="label">Patient:</span>
          <select className="field">
            {people.map((p) => (
              <option key={p}>{p}</option>
            ))}
          </select>
        </label>
        <label>
          <span className="label">Tooth:</span>
          <select
            className="field"
            value={selected}
            onChange={(e) => setSelected(Number(e.target.value))}
          >
            {Array.from({ length: 32 }, (_, i) => (
              <option key={i} value={i + 1}>
                {i + 1}
              </option>
            ))}
          </select>
        </label>
      </div>
      <div className="dental-palette">
        {choices.map(([code, color]) => (
          <button
            key={code}
            className="secondary"
            style={{
              borderColor: color,
              background: color,
              color: code === "Healthy" ? "#212529" : "white",
            }}
            onClick={() => {
              setConditions({ ...conditions, [selected]: code });
              setSaved(false);
            }}
          >
            {code}
          </button>
        ))}
      </div>
      <p className="form-preview-note">
        Select a tooth, then choose its marking. Original chart and marking
        palette.
      </p>
      <div
        className="original-odontogram"
        onClick={(e) => {
          const n = (e.target as Element)
            .closest("[data-key]")
            ?.getAttribute("data-key");
          if (n) setSelected(Number(n));
        }}
        dangerouslySetInnerHTML={{ __html: svg }}
      />
      <div className="flex gap-3">
        <button
          className="primary"
          onClick={() => {
            sessionStorage.setItem(
              "hms-odontogram-preview",
              JSON.stringify(conditions),
            );
            setSaved(true);
          }}
        >
          {t("Save")}
        </button>
        <button
          className="secondary"
          onClick={() => {
            setConditions({});
            setSaved(false);
          }}
        >
          Reset
        </button>
      </div>
      {saved && (
        <p className="success mt-4" role="status">
          Sample dental chart saved.
        </p>
      )}
    </div>
  );
}
function Schedule() {
  const { t } = useLanguage();
  const [saved, setSaved] = useState(false);
  return (
    <form
      className="legacy-card"
      onSubmit={(e) => {
        e.preventDefault();
        setSaved(true);
      }}
    >
      <h2>Hospital Schedule</h2>
      {[
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
        "Sunday",
      ].map((d) => (
        <div className="schedule-row" key={d}>
          <label>
            <input type="checkbox" defaultChecked={d !== "Sunday"} /> {d}
          </label>
          <input
            aria-label={`${d} opening time`}
            className="field"
            type="time"
            defaultValue="08:00"
          />
          <span>to</span>
          <input
            aria-label={`${d} closing time`}
            className="field"
            type="time"
            defaultValue="17:00"
          />
        </div>
      ))}
      <button className="primary mt-5">{t("Save")}</button>
      {saved && (
        <p role="status" className="success mt-4">
          Preview schedule saved.
        </p>
      )}
    </form>
  );
}
