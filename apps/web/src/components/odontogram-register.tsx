"use client";
import { useEffect, useState } from "react";
import { Search, Pencil, Trash2, Printer, X } from "lucide-react";
import { Modal } from "./modal";
import { useLanguage } from "./language";
import { people } from "@/lib/legacy";
import toothSvg from "@/lib/odontogram-svg.json";
type Chart = {
  id: string;
  patient: string;
  doctor: string;
  description: string;
  conditions: Record<number, string>;
};
const doctors = ["Dr. Avery Reed", "Dr. Robin Patel", "Dr. Quinn Parker"];
const colors: Record<string, string> = {
  Healthy: "#ffffff",
  K: "#e91e63",
  C: "#3f51b5",
  Ce: "#009688",
  D: "#795548",
  KR: "#607d8b",
  PS: "#ffc107",
  IP: "#673ab7",
  X: "#2196f3",
};
const key = "hms-odontogram-register-preview";
function drawing(conditions: Chart["conditions"]) {
  return toothSvg.replace(
    /(<polygon id="Tooth(\d+)"\s+fill=")[^"]*(")/g,
    (_m, p, n, s) => p + (colors[conditions[Number(n)]] || "#fff") + s,
  );
}
export function OdontogramRegister() {
  const { t } = useLanguage();
  const [rows, setRows] = useState<Chart[]>([]),
    [ready, setReady] = useState(false),
    [search, setSearch] = useState(""),
    [editing, setEditing] = useState<Chart | null>(null),
    [tooth, setTooth] = useState(1),
    [page, setPage] = useState(1),
    [size, setSize] = useState(10);
  useEffect(() => {
    try {
      setRows(
        JSON.parse(sessionStorage.getItem(key) || "null") ||
          people.map((patient, i) => ({
            id: `sample-${i}`,
            patient,
            doctor: doctors[i % 3],
            description: "Sample dental chart",
            conditions: {},
          })),
      );
    } catch {}
    setReady(true);
  }, []);
  useEffect(() => {
    if (ready) sessionStorage.setItem(key, JSON.stringify(rows));
  }, [rows, ready]);
  const filtered = rows.filter((r) =>
    `${r.patient} ${r.doctor}`.toLowerCase().includes(search.toLowerCase()),
  );
  const avatar = (name: string) => (
    <span className="record-link">
      <span className="avatar">
        {name
          .split(" ")
          .map((n) => n[0])
          .slice(0, 2)
          .join("")}
      </span>
      <span>
        {name}
        <small>{name.toLowerCase().replaceAll(" ", ".")}@example.invalid</small>
      </span>
    </span>
  );
  function print(row: Chart) {
    const win = window.open("", "_blank");
    if (!win) return;
    const d = win.document;
    d.title = "Odontogram";
    const heading = d.createElement("h1");
    heading.textContent = `${row.patient} - ${row.doctor}`;
    d.body.append(heading);
    const p = d.createElement("p");
    p.textContent = row.description;
    d.body.append(p);
    const chart = d.createElement("div");
    chart.innerHTML = drawing(row.conditions);
    chart.style.width = "350px";
    d.body.append(chart);
    win.print();
  }
  return (
    <section>
      <div className="table-toolbar">
        <label className="table-search">
          <Search size={16} />
          <input
            aria-label={t("Search odontograms")}
            placeholder={t("Search")}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </label>
        <button
          className="primary"
          onClick={() => {
            setTooth(1);
            setEditing({
              id: crypto.randomUUID(),
              patient: "",
              doctor: "",
              description: "",
              conditions: {},
            });
          }}
        >
          {t("Add Odontogram")}
        </button>
      </div>
      <div className="legacy-table-wrap">
        <table className="legacy-table">
          <thead>
            <tr>
              {["Patient", "Doctor", "Action"].map((h) => (
                <th key={h}>{t(h)}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.slice((page - 1) * size, page * size).map((row) => (
              <tr key={row.id}>
                <td>{avatar(row.patient)}</td>
                <td>{avatar(row.doctor)}</td>
                <td>
                  <div className="row-actions">
                    <button
                      aria-label={`Edit odontogram ${row.patient}`}
                      onClick={() => setEditing(structuredClone(row))}
                    >
                      <Pencil size={17} />
                    </button>
                    <button
                      aria-label={`Delete odontogram ${row.patient}`}
                      onClick={() => {
                        if (window.confirm(t("Delete this odontogram?")))
                          setRows(rows.filter((r) => r.id !== row.id));
                      }}
                    >
                      <Trash2 size={17} />
                    </button>
                    <button
                      className="print-action"
                      aria-label={`Print odontogram ${row.patient}`}
                      onClick={() => print(row)}
                    >
                      <Printer size={17} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {!filtered.length && (
              <tr>
                <td colSpan={3}>{t("No records found")}</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="reference-pagination">
        <label>
          {t("Show")}{" "}
          <select
            aria-label="Rows per page"
            className="field"
            value={size}
            onChange={(e) => {
              setSize(Number(e.target.value));
              setPage(1);
            }}
          >
            {[10, 25, 50].map((n) => (
              <option key={n}>{n}</option>
            ))}
          </select>
        </label>
        <span>
          {t("Showing")}{" "}
          {filtered.length
            ? Math.min((page - 1) * size + 1, filtered.length)
            : 0}{" "}
          - {Math.min(page * size, filtered.length)} {t("of")} {filtered.length}{" "}
          {t("Results")}
        </span>
        <div>
          {Array.from({ length: Math.ceil(filtered.length / size) }, (_, i) => (
            <button
              className={page === i + 1 ? "primary" : "secondary"}
              key={i}
              onClick={() => setPage(i + 1)}
            >
              {i + 1}
            </button>
          ))}
        </div>
      </div>
      {editing && (
        <Modal titleId="odontogram-title" wide onClose={() => setEditing(null)}>
          <form
            className="odontogram-editor"
            onSubmit={(e) => {
              e.preventDefault();
              setRows((old) =>
                old.some((r) => r.id === editing.id)
                  ? old.map((r) => (r.id === editing.id ? editing : r))
                  : [editing, ...old],
              );
              setEditing(null);
            }}
          >
            <header className="modal-heading">
              <h2 id="odontogram-title">
                {t(
                  rows.some((r) => r.id === editing.id)
                    ? "Edit Odontogram"
                    : "Add Odontogram",
                )}
              </h2>
              <button
                type="button"
                aria-label="Close"
                onClick={() => setEditing(null)}
              >
                <X size={20} />
              </button>
            </header>
            <div className="dental-editor-grid">
              <div>
                {(["patient", "doctor"] as const).map((field) => (
                  <label key={field}>
                    <span className="label">
                      {t(field === "patient" ? "Patient" : "Doctor")}:{" "}
                      <b className="text-red-500">*</b>
                    </span>
                    <select
                      required
                      className="field"
                      value={editing[field]}
                      onChange={(e) =>
                        setEditing({ ...editing, [field]: e.target.value })
                      }
                    >
                      <option value="">
                        {t(
                          field === "patient"
                            ? "Select Patient"
                            : "Select Doctor",
                        )}
                      </option>
                      {(field === "patient" ? people : doctors).map((name) => (
                        <option key={name}>{name}</option>
                      ))}
                    </select>
                  </label>
                ))}
                <label>
                  <span className="label">
                    {t("Description")}: <b className="text-red-500">*</b>
                  </span>
                  <textarea
                    className="field"
                    required
                    rows={5}
                    value={editing.description}
                    onChange={(e) =>
                      setEditing({ ...editing, description: e.target.value })
                    }
                  />
                </label>
              </div>
              <div>
                <div className="dental-palette">
                  {Object.entries(colors).map(([code, color]) => (
                    <button
                      type="button"
                      key={code}
                      title={`${code}: tooth ${tooth}`}
                      aria-label={`${code} tooth ${tooth}`}
                      style={{
                        background: color,
                        color: code === "Healthy" ? "#222" : "white",
                      }}
                      onClick={() =>
                        setEditing({
                          ...editing,
                          conditions: { ...editing.conditions, [tooth]: code },
                        })
                      }
                    >
                      {code}
                    </button>
                  ))}
                </div>
                <label>
                  <span className="label">{t("Tooth")}</span>
                  <select
                    className="field"
                    value={tooth}
                    onChange={(e) => setTooth(Number(e.target.value))}
                  >
                    {Array.from({ length: 32 }, (_, i) => (
                      <option key={i}>{i + 1}</option>
                    ))}
                  </select>
                </label>
                <div
                  className="original-odontogram"
                  onClick={(e) => {
                    const n = (e.target as Element)
                      .closest("[data-key]")
                      ?.getAttribute("data-key");
                    if (n) setTooth(Number(n));
                  }}
                  dangerouslySetInnerHTML={{
                    __html: drawing(editing.conditions),
                  }}
                />
              </div>
            </div>
            <footer className="modal-footer">
              <button className="primary">{t("Save")}</button>
              <button
                type="button"
                className="secondary"
                onClick={() => setEditing(null)}
              >
                {t("Cancel")}
              </button>
            </footer>
          </form>
        </Modal>
      )}
    </section>
  );
}
