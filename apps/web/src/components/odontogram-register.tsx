"use client";

import { useEffect, useState, useCallback } from "react";
import {
  Search,
  Pencil,
  Trash2,
  Printer,
  X,
  Palette,
  Plus,
  RotateCcw,
  Check,
  AlertCircle,
  Settings2,
  Loader2,
  RefreshCw,
  CheckCircle2,
} from "lucide-react";
import { Modal } from "./modal";
import { useLanguage } from "./language";
import { people } from "@/lib/legacy";
import toothSvg from "@/lib/odontogram-svg.json";
import { api, type Patient } from "@/lib/api";

export type OdontogramLegend = {
  id: string;
  code: string;
  name: string;
  color: string;
  description?: string;
  isDefault?: boolean;
};

export type Chart = {
  id: string;
  patientId?: string;
  patient: string;
  doctor: string;
  description: string;
  conditions: Record<number, string>;
};

export type ServerToothEntry = {
  id: string;
  patientId: string;
  encounterId?: string;
  toothNumber: number;
  condition: string;
  procedureNotes: string;
  diagnosedBy: string;
  createdAt: string;
  updatedAt: string;
};

export function legendCodeToDomainCondition(code: string): string {
  switch (code) {
    case "Healthy":
      return "healthy";
    case "K":
    case "Ce":
    case "D":
      return "caries";
    case "C":
    case "B":
      return "crown";
    case "KR":
      return "root_canal";
    case "IP":
      return "implant";
    case "X":
      return "extracted";
    case "F":
    case "PS":
      return "filled";
    default:
      return "caries";
  }
}

export function domainConditionToLegendCode(condition: string): string {
  switch (condition) {
    case "healthy":
      return "Healthy";
    case "caries":
      return "K";
    case "crown":
      return "C";
    case "root_canal":
      return "KR";
    case "implant":
      return "IP";
    case "extracted":
      return "X";
    case "filled":
      return "F";
    default:
      return "Healthy";
  }
}

const doctors = ["Dr. Avery Reed", "Dr. Robin Patel", "Dr. Quinn Parker"];

export const defaultLegends: OdontogramLegend[] = [
  {
    id: "leg-healthy",
    code: "Healthy",
    name: "Healthy Tooth",
    color: "#ffffff",
    description: "Sound tooth surface with no pathology",
    isDefault: true,
  },
  {
    id: "leg-k",
    code: "K",
    name: "Caries / Cavity",
    color: "#e91e63",
    description: "Active dental caries lesion",
    isDefault: true,
  },
  {
    id: "leg-c",
    code: "C",
    name: "Crown",
    color: "#3f51b5",
    description: "Artificial full crown or cap",
    isDefault: true,
  },
  {
    id: "leg-ce",
    code: "Ce",
    name: "Pulpitis / Deep Caries",
    color: "#009688",
    description: "Extensive caries with pulpal involvement",
    isDefault: true,
  },
  {
    id: "leg-d",
    code: "D",
    name: "Decayed / Denture",
    color: "#795548",
    description: "Grossly decayed or retained root",
    isDefault: true,
  },
  {
    id: "leg-kr",
    code: "KR",
    name: "Root Canal (Karies Radix)",
    color: "#607d8b",
    description: "Endodontic therapy / Root canal treated",
    isDefault: true,
  },
  {
    id: "leg-ps",
    code: "PS",
    name: "Porcelain / Prosthetic",
    color: "#ffc107",
    description: "Porcelain veneer or prosthetic replacement",
    isDefault: true,
  },
  {
    id: "leg-ip",
    code: "IP",
    name: "Implant",
    color: "#673ab7",
    description: "Titanium or ceramic dental implant",
    isDefault: true,
  },
  {
    id: "leg-x",
    code: "X",
    name: "Missing / Extracted",
    color: "#2196f3",
    description: "Congenitally missing or extracted tooth",
    isDefault: true,
  },
  {
    id: "leg-f",
    code: "F",
    name: "Composite Filling",
    color: "#10b981",
    description: "Tooth-colored composite restoration",
    isDefault: true,
  },
  {
    id: "leg-b",
    code: "B",
    name: "Bridge",
    color: "#f97316",
    description: "Fixed partial denture / pontic unit",
    isDefault: true,
  },
];

const PRESET_COLORS = [
  "#e91e63",
  "#3f51b5",
  "#009688",
  "#795548",
  "#607d8b",
  "#ffc107",
  "#673ab7",
  "#2196f3",
  "#10b981",
  "#f97316",
  "#8b5cf6",
  "#ec4899",
  "#06b6d4",
  "#84cc16",
  "#ef4444",
  "#64748b",
];

const chartStorageKey = "hms-odontogram-register-preview";
const legendsStorageKey = "hms-odontogram-legends-v2";

export function getLegendColorMap(
  legends: OdontogramLegend[],
): Record<string, string> {
  const map: Record<string, string> = {};
  for (const leg of legends) {
    map[leg.code] = leg.color;
  }
  // Ensure fallback defaults exist
  if (!map["Healthy"]) map["Healthy"] = "#ffffff";
  return map;
}

export function drawing(
  conditions: Chart["conditions"],
  colorMap: Record<string, string>,
) {
  return toothSvg.replace(
    /(<polygon id="Tooth(\d+)"\s+fill=")[^"]*(")/g,
    (_m, p, n, s) => p + (colorMap[conditions[Number(n)]] || "#fff") + s,
  );
}

export function OdontogramRegister() {
  const { t } = useLanguage();
  const [rows, setRows] = useState<Chart[]>([]);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState("");
  const [editing, setEditing] = useState<Chart | null>(null);
  const [tooth, setTooth] = useState(1);
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(10);

  // Legends state
  const [legends, setLegends] = useState<OdontogramLegend[]>(defaultLegends);
  const [manageLegendsOpen, setManageLegendsOpen] = useState(false);
  const [editingLegend, setEditingLegend] = useState<OdontogramLegend | null>(
    null,
  );
  const [legendFormError, setLegendFormError] = useState("");

  // Persistent API States (Section 4 Integration)
  const [remotePatients, setRemotePatients] = useState<
    { id: string; name: string; mrn: string }[]
  >([]);
  const [loadingRemote, setLoadingRemote] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [loadingTeeth, setLoadingTeeth] = useState(false);
  const [apiConnected, setApiConnected] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [successBanner, setSuccessBanner] = useState("");

  const loadPatients = useCallback(async () => {
    setLoadingRemote(true);
    setErrorMessage("");
    try {
      const res = await api<{ patients: Patient[] }>("patients?page=1");
      if (res?.patients && Array.isArray(res.patients) && res.patients.length > 0) {
        setRemotePatients(
          res.patients.map((p) => ({
            id: p.id,
            name: `${p.givenName} ${p.familyName}`.trim(),
            mrn: p.mrn,
          })),
        );
        setApiConnected(true);
      } else {
        setApiConnected(true);
      }
    } catch {
      // Backend not yet populated with patients or in preview mode
      setApiConnected(false);
    } finally {
      setLoadingRemote(false);
    }
  }, []);

  const loadPatientOdontogram = useCallback(async (patientId: string) => {
    if (!patientId) return;
    setLoadingTeeth(true);
    try {
      const res = await api<{ teeth: ServerToothEntry[] }>(
        `patients/${patientId}/odontogram`,
      );
      if (res?.teeth && Array.isArray(res.teeth)) {
        const loadedConditions: Record<number, string> = {};
        for (const tEntry of res.teeth) {
          loadedConditions[tEntry.toothNumber] = domainConditionToLegendCode(
            tEntry.condition,
          );
        }
        setEditing((prev) =>
          prev && prev.patientId === patientId
            ? { ...prev, conditions: { ...prev.conditions, ...loadedConditions } }
            : prev,
        );
      }
    } catch {
      // Local fallback
    } finally {
      setLoadingTeeth(false);
    }
  }, []);

  // Load rows, remote patients, and legends
  useEffect(() => {
    try {
      const storedLegends = sessionStorage.getItem(legendsStorageKey);
      if (storedLegends) {
        setLegends(JSON.parse(storedLegends));
      } else {
        setLegends(defaultLegends);
      }
    } catch {
      setLegends(defaultLegends);
    }

    try {
      setRows(
        JSON.parse(sessionStorage.getItem(chartStorageKey) || "null") ||
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
    loadPatients();
  }, [loadPatients]);

  // Sync rows
  useEffect(() => {
    if (ready) sessionStorage.setItem(chartStorageKey, JSON.stringify(rows));
  }, [rows, ready]);

  // Sync legends
  useEffect(() => {
    if (ready)
      sessionStorage.setItem(legendsStorageKey, JSON.stringify(legends));
  }, [legends, ready]);

  const colorMap = getLegendColorMap(legends);

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

  // Save new or edited legend
  function handleSaveLegend(e: React.FormEvent) {
    e.preventDefault();
    if (!editingLegend) return;

    const trimmedCode = editingLegend.code.trim();
    const trimmedName = editingLegend.name.trim();

    if (!trimmedCode) {
      setLegendFormError(t("Legend code / symbol is required"));
      return;
    }
    if (!trimmedName) {
      setLegendFormError(t("Legend name is required"));
      return;
    }

    // Check duplicate code (case-insensitive)
    const duplicate = legends.some(
      (l) =>
        l.id !== editingLegend.id &&
        l.code.toLowerCase() === trimmedCode.toLowerCase(),
    );
    if (duplicate) {
      setLegendFormError(
        t(
          "A legend with this code already exists. Please choose another code.",
        ),
      );
      return;
    }

    const isExisting = legends.some((l) => l.id === editingLegend.id);
    let oldCode = "";

    if (isExisting) {
      const existing = legends.find((l) => l.id === editingLegend.id);
      oldCode = existing?.code || "";
    }

    const updatedLegend: OdontogramLegend = {
      ...editingLegend,
      code: trimmedCode,
      name: trimmedName,
      color: editingLegend.color || "#10b981",
    };

    let updatedLegends: OdontogramLegend[];
    if (isExisting) {
      updatedLegends = legends.map((l) =>
        l.id === updatedLegend.id ? updatedLegend : l,
      );
    } else {
      updatedLegends = [...legends, updatedLegend];
    }

    setLegends(updatedLegends);

    // If code was changed, migrate existing conditions in current editor and stored charts
    if (oldCode && oldCode !== trimmedCode) {
      if (editing) {
        const nextConditions = { ...editing.conditions };
        for (const [tNum, cCode] of Object.entries(nextConditions)) {
          if (cCode === oldCode) nextConditions[Number(tNum)] = trimmedCode;
        }
        setEditing({ ...editing, conditions: nextConditions });
      }
      setRows((prev) =>
        prev.map((chart) => {
          const nextConds = { ...chart.conditions };
          let changed = false;
          for (const [tNum, cCode] of Object.entries(nextConds)) {
            if (cCode === oldCode) {
              nextConds[Number(tNum)] = trimmedCode;
              changed = true;
            }
          }
          return changed ? { ...chart, conditions: nextConds } : chart;
        }),
      );
    }

    setEditingLegend(null);
    setLegendFormError("");
  }

  function handleDeleteLegend(legendId: string) {
    const target = legends.find((l) => l.id === legendId);
    if (!target) return;
    if (target.code === "Healthy") {
      alert(t("The Healthy condition legend cannot be removed."));
      return;
    }
    if (
      window.confirm(
        t(
          `Are you sure you want to delete legend "${target.name}" (${target.code})?`,
        ),
      )
    ) {
      setLegends(legends.filter((l) => l.id !== legendId));
    }
  }

  function handleResetLegends() {
    if (
      window.confirm(
        t(
          "Reset all odontogram legends to default dental conditions? Custom changes will be restored.",
        ),
      )
    ) {
      setLegends(defaultLegends);
      sessionStorage.setItem(legendsStorageKey, JSON.stringify(defaultLegends));
    }
  }

  function print(row: Chart) {
    const win = window.open("", "_blank");
    if (!win) return;
    const d = win.document;
    d.title = "Odontogram Report - " + row.patient;

    const currentMap = getLegendColorMap(legends);

    d.body.innerHTML = `
      <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; padding: 30px; color: #1e293b; }
        h1 { margin-bottom: 6px; font-size: 24px; color: #0f172a; }
        .meta { color: #64748b; font-size: 14px; margin-bottom: 20px; }
        .desc { background: #f8fafc; border-left: 4px solid #6571ff; padding: 12px 16px; margin: 16px 0 24px; border-radius: 4px; font-size: 14px; }
        .chart-box { width: 360px; margin: 20px 0; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; background: #fff; }
        .legends-table { width: 100%; border-collapse: collapse; margin-top: 30px; font-size: 13px; }
        .legends-table th { background: #f1f5f9; text-align: left; padding: 8px 12px; border: 1px solid #cbd5e1; }
        .legends-table td { padding: 8px 12px; border: 1px solid #e2e8f0; vertical-align: middle; }
        .swatch { display: inline-block; width: 18px; height: 18px; border-radius: 4px; border: 1px solid #94a3b8; margin-right: 8px; vertical-align: middle; }
        .code-badge { font-family: monospace; font-weight: bold; background: #f1f5f9; padding: 2px 6px; border-radius: 4px; }
      </style>
      <h1>${row.patient} - ${row.doctor}</h1>
      <div class="meta">Hospital Dental Examination Chart • Generated on ${new Date().toLocaleDateString()}</div>
      <div class="desc"><strong>Clinical Notes:</strong> ${row.description || "No specific remarks"}</div>
      <div class="chart-box">
        ${drawing(row.conditions, currentMap)}
      </div>
      <h3>Odontogram Legend Guide</h3>
      <table class="legends-table">
        <thead>
          <tr>
            <th>Color</th>
            <th>Symbol / Code</th>
            <th>Condition Name</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          ${legends
            .map(
              (l) => `
            <tr>
              <td><span class="swatch" style="background-color: ${l.color}"></span> ${l.color}</td>
              <td><span class="code-badge">${l.code}</span></td>
              <td><strong>${l.name}</strong></td>
              <td>${l.description || "-"}</td>
            </tr>
          `,
            )
            .join("")}
        </tbody>
      </table>
    `;
    setTimeout(() => {
      win.print();
    }, 250);
  }

  return (
    <section>
      {/* Persistent Connection Status & Alerts (Section 4 Integration) */}
      <div className="mb-3 d-flex flex-column gap-2">
        <div
          className="d-flex justify-content-between align-items-center flex-wrap gap-2 px-3 py-2 rounded border"
          style={{
            backgroundColor: apiConnected
              ? "rgba(16, 185, 129, 0.08)"
              : "rgba(245, 158, 11, 0.08)",
            borderColor: apiConnected ? "#10b981" : "#f59e0b",
          }}
        >
          <div className="d-flex align-items-center gap-2">
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                backgroundColor: apiConnected ? "#10b981" : "#f59e0b",
                display: "inline-block",
              }}
            />
            <span className="fs-7 fw-semibold">
              {apiConnected
                ? t(
                    "Connected to Go/PostgreSQL Clinical Records (/v1/patients/{id}/odontogram)",
                  )
                : t("Local preview mode · Syncing locally")}
            </span>
            {remotePatients.length > 0 && (
              <span
                className="badge-available-stock fs-8 py-0 px-2"
                style={{ backgroundColor: "#10b98122", color: "#10b981" }}
              >
                {remotePatients.length} {t("Authoritative Patients Loaded")}
              </span>
            )}
          </div>
          <button
            type="button"
            className="btn-icon-link fs-7 d-flex align-items-center gap-1"
            onClick={loadPatients}
            disabled={loadingRemote}
            title={t("Refresh patient records from API")}
          >
            <RefreshCw
              size={13}
              className={loadingRemote ? "animate-spin" : ""}
            />
            <span>{loadingRemote ? t("Syncing...") : t("Sync Backend")}</span>
          </button>
        </div>

        {successBanner && (
          <div
            className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded"
            style={{
              borderColor: "#10b981",
              backgroundColor: "rgba(16, 185, 129, 0.1)",
              color: "#10b981",
            }}
          >
            <div className="d-flex align-items-center gap-2">
              <CheckCircle2 size={16} />
              <span className="fs-7">{successBanner}</span>
            </div>
            <button
              type="button"
              onClick={() => setSuccessBanner("")}
              className="btn-icon-link"
            >
              <X size={14} />
            </button>
          </div>
        )}

        {errorMessage && (
          <div
            className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded"
            style={{
              borderColor: "#ef4444",
              backgroundColor: "rgba(239, 68, 68, 0.1)",
              color: "#ef4444",
            }}
          >
            <div className="d-flex align-items-center gap-2">
              <AlertCircle size={16} />
              <span className="fs-7">{errorMessage}</span>
            </div>
            <button
              type="button"
              onClick={() => setErrorMessage("")}
              className="btn-icon-link"
            >
              <X size={14} />
            </button>
          </div>
        )}
      </div>

      {/* Top Toolbar */}
      <div className="table-toolbar d-flex justify-content-between align-items-center flex-wrap gap-2">
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

        <div className="d-flex align-items-center gap-2">
          {/* Manage Legends Trigger */}
          <button
            type="button"
            className="btn-action-blue"
            style={{ background: "#475569" }}
            onClick={() => setManageLegendsOpen(true)}
            title={t("Manage Odontogram Legends")}
          >
            <Palette size={16} />
            <span>{t("Odontogram Legends")}</span>
          </button>

          {/* Add Odontogram Primary Button */}
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
      </div>

      {/* Main Table */}
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
                      title={t("Edit odontogram")}
                    >
                      <Pencil size={17} />
                    </button>
                    <button
                      aria-label={`Delete odontogram ${row.patient}`}
                      onClick={() => {
                        if (window.confirm(t("Delete this odontogram?")))
                          setRows(rows.filter((r) => r.id !== row.id));
                      }}
                      title={t("Delete odontogram")}
                    >
                      <Trash2 size={17} />
                    </button>
                    <button
                      className="print-action"
                      aria-label={`Print odontogram ${row.patient}`}
                      onClick={() => print(row)}
                      title={t("Print odontogram")}
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

      {/* Pagination */}
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

      {/* Quick Visual Legends Key Strip on Main Page */}
      <div className="mt-4 p-3 border rounded form-card-container">
        <div className="d-flex justify-content-between align-items-center mb-2">
          <div className="d-flex align-items-center gap-2">
            <Palette size={16} className="text-primary" />
            <strong className="fs-6">{t("Odontogram Legends Key")}</strong>
            <span className="text-secondary fs-7">
              ({legends.length} {t("conditions defined")})
            </span>
          </div>
          <button
            type="button"
            className="btn-action-blue py-1 px-2 fs-7"
            onClick={() => setManageLegendsOpen(true)}
          >
            <Settings2 size={14} />
            <span>{t("Manage / Customize Legends")}</span>
          </button>
        </div>
        <div className="d-flex flex-wrap gap-2 mt-2">
          {legends.map((leg) => (
            <div
              key={leg.id}
              className="d-flex align-items-center gap-2 px-2 py-1 border rounded"
              style={{
                background: "rgba(255, 255, 255, 0.04)",
                borderColor: "#334155",
                fontSize: "12px",
              }}
              title={leg.description || leg.name}
            >
              <span
                style={{
                  width: "12px",
                  height: "12px",
                  borderRadius: "50%",
                  background: leg.color,
                  border: "1px solid rgba(0,0,0,0.2)",
                  display: "inline-block",
                }}
              />
              <span className="fw-bold">{leg.code}:</span>
              <span className="text-secondary">{t(leg.name)}</span>
            </div>
          ))}
        </div>
      </div>

      {/* =============================================================
          MODAL: ADD / EDIT ODONTOGRAM
          ============================================================= */}
      {editing && (
        <Modal titleId="odontogram-title" wide onClose={() => setEditing(null)}>
          <form
            className="odontogram-editor"
            onSubmit={async (e) => {
              e.preventDefault();
              if (!editing) return;
              setIsSaving(true);
              setErrorMessage("");
              setSuccessBanner("");

              let apiSaved = false;
              if (editing.patientId) {
                try {
                  for (const [tNumStr, code] of Object.entries(
                    editing.conditions,
                  )) {
                    const toothNumber = Number(tNumStr);
                    const condition = legendCodeToDomainCondition(code);
                    await api(`patients/${editing.patientId}/odontogram`, {
                      method: "POST",
                      body: JSON.stringify({
                        toothNumber,
                        condition,
                        procedureNotes: editing.description || "",
                      }),
                    });
                  }
                  apiSaved = true;
                } catch (err: any) {
                  console.warn("API odontogram save notice:", err);
                  setErrorMessage(
                    err?.message ||
                      t(
                        "Saved locally. Could not synchronize to remote patient record.",
                      ),
                  );
                }
              }

              setRows((old) =>
                old.some((r) => r.id === editing.id)
                  ? old.map((r) => (r.id === editing.id ? editing : r))
                  : [editing, ...old],
              );

              if (apiSaved) {
                setSuccessBanner(
                  t(
                    "Odontogram successfully saved to persistent patient records.",
                  ),
                );
              }
              setEditing(null);
              setIsSaving(false);
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
                <label>
                  <span className="label">
                    {t("Patient")}: <b className="text-red-500">*</b>
                    {loadingTeeth && (
                      <span className="ms-2 fs-7 text-primary d-inline-flex align-items-center gap-1">
                        <Loader2 size={13} className="animate-spin" />{" "}
                        {t("Loading clinical record...")}
                      </span>
                    )}
                  </span>
                  <select
                    required
                    className="field"
                    value={editing.patientId || editing.patient}
                    onChange={(e) => {
                      const val = e.target.value;
                      const matched = remotePatients.find((p) => p.id === val);
                      if (matched) {
                        setEditing({
                          ...editing,
                          patientId: matched.id,
                          patient: matched.name,
                        });
                        loadPatientOdontogram(matched.id);
                      } else {
                        setEditing({
                          ...editing,
                          patientId: undefined,
                          patient: val,
                        });
                      }
                    }}
                  >
                    <option value="">{t("Select Patient")}</option>
                    {remotePatients.length > 0 && (
                      <optgroup label={t("Authoritative Patient Records")}>
                        {remotePatients.map((p) => (
                          <option key={p.id} value={p.id}>
                            {p.name} ({p.mrn})
                          </option>
                        ))}
                      </optgroup>
                    )}
                    <optgroup label={t("Standard Patients")}>
                      {people.map((name) => (
                        <option key={name} value={name}>
                          {name}
                        </option>
                      ))}
                    </optgroup>
                  </select>
                </label>

                <label>
                  <span className="label">
                    {t("Doctor")}: <b className="text-red-500">*</b>
                  </span>
                  <select
                    required
                    className="field"
                    value={editing.doctor}
                    onChange={(e) =>
                      setEditing({ ...editing, doctor: e.target.value })
                    }
                  >
                    <option value="">{t("Select Doctor")}</option>
                    {doctors.map((name) => (
                      <option key={name} value={name}>
                        {name}
                      </option>
                    ))}
                  </select>
                </label>
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

                {/* Selected Tooth Summary in Editor */}
                <div className="p-3 border rounded mt-3 bg-opacity-10 bg-secondary">
                  <div className="d-flex justify-content-between align-items-center">
                    <span className="fw-semibold">
                      {t("Selected Tooth")}: #{tooth}
                    </span>
                    <span
                      className="badge-available-stock"
                      style={{
                        backgroundColor: `${colorMap[editing.conditions[tooth] || "Healthy"]}22`,
                        color:
                          colorMap[editing.conditions[tooth] || "Healthy"] ===
                          "#ffffff"
                            ? "#cbd5e1"
                            : colorMap[editing.conditions[tooth] || "Healthy"],
                        borderColor:
                          colorMap[editing.conditions[tooth] || "Healthy"],
                      }}
                    >
                      {editing.conditions[tooth]
                        ? `${editing.conditions[tooth]} - ${legends.find((l) => l.code === editing.conditions[tooth])?.name || t("Condition")}`
                        : t("Healthy Tooth")}
                    </span>
                  </div>
                </div>
              </div>

              <div>
                {/* Condition Palette with Quick Legend Management */}
                <div className="d-flex justify-content-between align-items-center mb-2">
                  <span className="label mb-0 fw-semibold">
                    {t("Condition Palette (Click to assign)")}:
                  </span>
                  <div className="d-flex gap-2">
                    <button
                      type="button"
                      className="btn-icon-link text-primary fs-7 d-flex align-items-center gap-1"
                      onClick={() => {
                        setEditingLegend({
                          id: crypto.randomUUID(),
                          code: "",
                          name: "",
                          color: "#10b981",
                          description: "",
                          isDefault: false,
                        });
                      }}
                      title={t("Create new condition legend")}
                    >
                      <Plus size={13} />
                      <span>{t("Add Legend")}</span>
                    </button>
                    <button
                      type="button"
                      className="btn-icon-link text-secondary fs-7"
                      onClick={() => setManageLegendsOpen(true)}
                      title={t("Manage all legends")}
                    >
                      <Settings2 size={13} />
                    </button>
                  </div>
                </div>

                <div className="dental-palette">
                  {legends.map((leg) => {
                    const isHealthy =
                      leg.code === "Healthy" ||
                      leg.color.toLowerCase() === "#ffffff";
                    const isSelected =
                      (editing.conditions[tooth] || "Healthy") === leg.code;

                    return (
                      <button
                        type="button"
                        key={leg.id}
                        title={`${leg.code}: ${leg.name} (Tooth ${tooth})`}
                        aria-label={`${leg.code} tooth ${tooth}`}
                        style={{
                          background: leg.color,
                          color: isHealthy ? "#222" : "#fff",
                          border: isSelected
                            ? "2px solid #6571ff"
                            : isHealthy
                              ? "1px solid #cbd5e1"
                              : "1px solid transparent",
                          boxShadow: isSelected
                            ? "0 0 0 2px rgba(101, 113, 255, 0.4)"
                            : "none",
                          fontWeight: isSelected ? 700 : 500,
                        }}
                        onClick={() =>
                          setEditing({
                            ...editing,
                            conditions: {
                              ...editing.conditions,
                              [tooth]: leg.code,
                            },
                          })
                        }
                      >
                        {leg.code}
                      </button>
                    );
                  })}
                </div>

                <label className="mt-3">
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
                    __html: drawing(editing.conditions, colorMap),
                  }}
                />
              </div>
            </div>

            <footer className="modal-footer">
              <button className="primary" disabled={isSaving}>
                {isSaving ? (
                  <span className="d-inline-flex align-items-center gap-1">
                    <Loader2 size={15} className="animate-spin" />
                    {t("Saving to backend...")}
                  </span>
                ) : (
                  t("Save")
                )}
              </button>
              <button
                type="button"
                className="secondary"
                disabled={isSaving}
                onClick={() => setEditing(null)}
              >
                {t("Cancel")}
              </button>
            </footer>
          </form>
        </Modal>
      )}

      {/* =============================================================
          MODAL: MANAGE ODONTOGRAM LEGENDS
          ============================================================= */}
      {manageLegendsOpen && (
        <Modal
          titleId="manage-legends-title"
          wide
          onClose={() => setManageLegendsOpen(false)}
        >
          <div className="odontogram-editor">
            <header className="modal-heading">
              <div className="d-flex align-items-center gap-2">
                <Palette size={20} className="text-primary" />
                <h2 id="manage-legends-title" className="m-0">
                  {t("Odontogram Legends Management")}
                </h2>
              </div>
              <button
                type="button"
                aria-label="Close"
                onClick={() => setManageLegendsOpen(false)}
              >
                <X size={20} />
              </button>
            </header>

            <div className="d-flex justify-content-between align-items-center my-3">
              <p className="text-secondary m-0 fs-7">
                {t(
                  "Configure condition codes, descriptions, and tooth chart fill colors used in odontogram evaluations.",
                )}
              </p>
              <div className="d-flex gap-2">
                <button
                  type="button"
                  className="btn-action-blue"
                  onClick={() => {
                    setEditingLegend({
                      id: crypto.randomUUID(),
                      code: "",
                      name: "",
                      color: "#10b981",
                      description: "",
                      isDefault: false,
                    });
                  }}
                >
                  <Plus size={15} />
                  <span>{t("New Legend")}</span>
                </button>
                <button
                  type="button"
                  className="btn-back-outline d-flex align-items-center gap-1"
                  onClick={handleResetLegends}
                  title={t("Restore system default legends")}
                >
                  <RotateCcw size={14} />
                  <span>{t("Reset Defaults")}</span>
                </button>
              </div>
            </div>

            {/* Legends Table */}
            <div className="table-responsive border rounded">
              <table className="billing-table w-100">
                <thead>
                  <tr>
                    <th style={{ width: "80px" }}>{t("COLOR")}</th>
                    <th style={{ width: "90px" }}>{t("SYMBOL / CODE")}</th>
                    <th>{t("CONDITION NAME")}</th>
                    <th>{t("DESCRIPTION")}</th>
                    <th style={{ width: "100px", textAlign: "right" }}>
                      {t("ACTIONS")}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {legends.map((leg) => (
                    <tr key={leg.id}>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <span
                            style={{
                              width: "22px",
                              height: "22px",
                              borderRadius: "4px",
                              background: leg.color,
                              border: "1px solid #64748b",
                              display: "inline-block",
                            }}
                          />
                          <span className="font-monospace fs-7 text-secondary">
                            {leg.color}
                          </span>
                        </div>
                      </td>
                      <td>
                        <span
                          className="px-2 py-1 rounded fw-bold font-monospace"
                          style={{
                            background: leg.color,
                            color:
                              leg.code === "Healthy" ||
                              leg.color.toLowerCase() === "#ffffff"
                                ? "#0f172a"
                                : "#ffffff",
                            border: "1px solid #475569",
                            fontSize: "12px",
                          }}
                        >
                          {leg.code}
                        </span>
                      </td>
                      <td>
                        <strong>{leg.name}</strong>
                        {leg.isDefault && (
                          <span className="badge-available-stock ms-2 py-0 px-1 fs-8">
                            {t("Standard")}
                          </span>
                        )}
                      </td>
                      <td className="text-secondary fs-7">
                        {leg.description || "—"}
                      </td>
                      <td style={{ textAlign: "right" }}>
                        <div className="d-flex justify-content-end gap-1">
                          <button
                            type="button"
                            className="btn-icon-blue-link"
                            aria-label={`Edit legend ${leg.name}`}
                            onClick={() =>
                              setEditingLegend(structuredClone(leg))
                            }
                            title={t("Edit legend")}
                          >
                            <Pencil size={15} />
                          </button>
                          {leg.code !== "Healthy" && (
                            <button
                              type="button"
                              className="btn-icon-danger"
                              aria-label={`Delete legend ${leg.name}`}
                              onClick={() => handleDeleteLegend(leg.id)}
                              title={t("Delete legend")}
                            >
                              <Trash2 size={15} />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <footer className="modal-footer mt-3">
              <button
                type="button"
                className="secondary"
                onClick={() => setManageLegendsOpen(false)}
              >
                {t("Close")}
              </button>
            </footer>
          </div>
        </Modal>
      )}

      {/* =============================================================
          MODAL: CREATE OR EDIT SINGLE LEGEND
          ============================================================= */}
      {editingLegend && (
        <Modal
          titleId="legend-form-title"
          onClose={() => {
            setEditingLegend(null);
            setLegendFormError("");
          }}
        >
          <form className="odontogram-editor" onSubmit={handleSaveLegend}>
            <header className="modal-heading">
              <div className="d-flex align-items-center gap-2">
                <Palette size={18} className="text-primary" />
                <h2 id="legend-form-title" className="m-0">
                  {t(
                    legends.some((l) => l.id === editingLegend.id)
                      ? "Edit Odontogram Legend"
                      : "Create New Odontogram Legend",
                  )}
                </h2>
              </div>
              <button
                type="button"
                aria-label="Close"
                onClick={() => {
                  setEditingLegend(null);
                  setLegendFormError("");
                }}
              >
                <X size={20} />
              </button>
            </header>

            {legendFormError && (
              <div className="alert alert-danger d-flex align-items-center gap-2 mb-3 p-2 text-danger border border-danger rounded">
                <AlertCircle size={16} />
                <span>{legendFormError}</span>
              </div>
            )}

            <div className="form-group-custom">
              <label>
                {t("Legend Code / Symbol")}:{" "}
                <span className="text-danger">*</span>
              </label>
              <input
                type="text"
                required
                maxLength={8}
                placeholder={t("e.g. F, V, RCT, IMP")}
                value={editingLegend.code}
                onChange={(e) => {
                  setEditingLegend({
                    ...editingLegend,
                    code: e.target.value.toUpperCase(),
                  });
                  setLegendFormError("");
                }}
              />
              <small className="text-secondary fs-8">
                {t(
                  "Short abbreviation displayed on dental chart buttons (1-8 characters).",
                )}
              </small>
            </div>

            <div className="form-group-custom">
              <label>
                {t("Condition Name")}: <span className="text-danger">*</span>
              </label>
              <input
                type="text"
                required
                placeholder={t("e.g. Composite Restoration, Ceramic Veneer")}
                value={editingLegend.name}
                onChange={(e) => {
                  setEditingLegend({
                    ...editingLegend,
                    name: e.target.value,
                  });
                  setLegendFormError("");
                }}
              />
            </div>

            <div className="form-group-custom">
              <label>
                {t("Tooth Fill Color")}: <span className="text-danger">*</span>
              </label>
              <div className="d-flex align-items-center gap-3">
                <input
                  type="color"
                  value={editingLegend.color || "#10b981"}
                  onChange={(e) =>
                    setEditingLegend({
                      ...editingLegend,
                      color: e.target.value,
                    })
                  }
                  style={{
                    width: "48px",
                    height: "38px",
                    padding: "2px",
                    cursor: "pointer",
                    borderRadius: "6px",
                    border: "1px solid #475569",
                  }}
                />
                <input
                  type="text"
                  placeholder="#hex"
                  value={editingLegend.color}
                  onChange={(e) =>
                    setEditingLegend({
                      ...editingLegend,
                      color: e.target.value,
                    })
                  }
                  style={{ width: "120px", fontFamily: "monospace" }}
                />
              </div>

              {/* Preset Color Swatches for easy selection */}
              <div className="mt-2">
                <span className="text-secondary fs-8 mb-1 d-block">
                  {t("Quick Dental Palette Swatches:")}
                </span>
                <div className="d-flex flex-wrap gap-2">
                  {PRESET_COLORS.map((c) => (
                    <button
                      type="button"
                      key={c}
                      onClick={() =>
                        setEditingLegend({ ...editingLegend, color: c })
                      }
                      style={{
                        width: "24px",
                        height: "24px",
                        background: c,
                        borderRadius: "4px",
                        border:
                          editingLegend.color?.toLowerCase() === c.toLowerCase()
                            ? "2px solid #ffffff"
                            : "1px solid rgba(0,0,0,0.3)",
                        cursor: "pointer",
                        outline:
                          editingLegend.color?.toLowerCase() === c.toLowerCase()
                            ? "2px solid #6571ff"
                            : "none",
                      }}
                      title={c}
                    />
                  ))}
                </div>
              </div>
            </div>

            <div className="form-group-custom">
              <label>{t("Clinical Description / Notes")}:</label>
              <textarea
                rows={3}
                placeholder={t(
                  "Optional clinical notes or indication criteria...",
                )}
                value={editingLegend.description || ""}
                onChange={(e) =>
                  setEditingLegend({
                    ...editingLegend,
                    description: e.target.value,
                  })
                }
              />
            </div>

            {/* Live Preview Card */}
            <div className="p-3 border rounded mt-3 form-card-container">
              <span className="text-secondary fs-8 d-block mb-1">
                {t("Live Button Preview:")}
              </span>
              <div className="d-flex align-items-center gap-3">
                <button
                  type="button"
                  style={{
                    background: editingLegend.color || "#10b981",
                    color:
                      editingLegend.code === "Healthy" ||
                      editingLegend.color?.toLowerCase() === "#ffffff"
                        ? "#0f172a"
                        : "#ffffff",
                    border: "1px solid #475569",
                    padding: "6px 14px",
                    borderRadius: "4px",
                    fontWeight: 600,
                  }}
                >
                  {editingLegend.code || t("CODE")}
                </button>
                <div>
                  <strong>{editingLegend.name || t("Condition Name")}</strong>
                  <div className="text-secondary fs-8">
                    {editingLegend.description || t("No notes")}
                  </div>
                </div>
              </div>
            </div>

            <footer className="modal-footer mt-4">
              <button className="primary" type="submit">
                {t("Save Legend")}
              </button>
              <button
                type="button"
                className="secondary"
                onClick={() => {
                  setEditingLegend(null);
                  setLegendFormError("");
                }}
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
