"use client";

import { useEffect, useState, useCallback } from "react";
import {
  Search,
  Pencil,
  Trash2,
  Printer,
  FileDown,
  X,
  Palette,
  Plus,
  RotateCcw,
  AlertCircle,
  Settings2,
  Loader2,
  RefreshCw,
  CheckCircle2,
} from "lucide-react";
import { Modal } from "./modal";
import { useLanguage } from "./language";
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
  patientId: string;
  patient: string;
  patientMrn?: string;
  doctorId: string;
  doctor: string;
  encounterId?: string;
  description: string;
  conditions: Record<number, string>;
  version: number;
  createdBy?: string;
  updatedBy?: string;
  createdAt?: string;
  updatedAt?: string;
  isLegacy?: boolean;
};

export type ServerOdontogramChart = {
  id: string;
  patientId: string;
  patientName?: string;
  patientMrn?: string;
  doctorId: string;
  doctorName?: string;
  encounterId?: string;
  description: string;
  odontogram: Record<string, string>;
  version: number;
  createdBy: string;
  updatedBy: string;
  createdAt: string;
  updatedAt: string;
  isLegacy: boolean;
};

export type DoctorOption = {
  id: string;
  name: string;
  department?: string;
};

export type ServerToothEntry = {
  id: string;
  patientId: string;
  chartId?: string;
  encounterId?: string;
  toothNumber: number;
  condition: string;
  procedureNotes: string;
  diagnosedBy: string;
  createdAt: string;
  updatedAt: string;
};

// Legacy helper functions preserved for test and audit tool compatibility
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

const legendsStorageKey = "hms-odontogram-legends-v2";

export function getLegendColorMap(
  legends: OdontogramLegend[],
): Record<string, string> {
  const map: Record<string, string> = {};
  for (const leg of legends) {
    map[leg.code] = leg.color;
  }
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
  const [totalCount, setTotalCount] = useState(0);
  const [loadingCharts, setLoadingCharts] = useState(true);
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

  // Persistent Clinical Data
  const [patients, setPatients] = useState<Patient[]>([]);
  const [doctorsList, setDoctorsList] = useState<DoctorOption[]>([]);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeletingId, setIsDeletingId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState("");
  const [successBanner, setSuccessBanner] = useState("");

  // Load legends from session
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
  }, []);

  // Save legends to session
  const saveLegends = useCallback((newLegends: OdontogramLegend[]) => {
    setLegends(newLegends);
    try {
      sessionStorage.setItem(legendsStorageKey, JSON.stringify(newLegends));
    } catch {}
  }, []);

  // Load clinical doctors & patients
  const loadClinicalMetadata = useCallback(async () => {
    try {
      const [patRes, docRes] = await Promise.allSettled([
        api<{ patients: Patient[] }>("patients?page=1&size=100"),
        api<{ doctors: DoctorOption[] }>("doctors"),
      ]);
      if (patRes.status === "fulfilled" && patRes.value?.patients) {
        setPatients(patRes.value.patients);
      }
      if (docRes.status === "fulfilled" && docRes.value?.doctors) {
        setDoctorsList(docRes.value.doctors);
      }
    } catch {
      // Non-blocking
    }
  }, []);

  // Load odontogram charts from backend
  const loadCharts = useCallback(async () => {
    setLoadingCharts(true);
    setErrorMessage("");
    try {
      const queryParams = new URLSearchParams({
        page: String(page),
        size: String(size),
      });
      if (search.trim()) {
        queryParams.set("search", search.trim());
      }
      const res = await api<{ charts: ServerOdontogramChart[]; total: number }>(
        `odontograms?${queryParams.toString()}`,
      );
      if (res?.charts && Array.isArray(res.charts)) {
        const mapped: Chart[] = res.charts.map((c) => {
          const conditions: Record<number, string> = {};
          if (c.odontogram) {
            for (const [k, v] of Object.entries(c.odontogram)) {
              conditions[Number(k)] = v;
            }
          }
          return {
            id: c.id,
            patientId: c.patientId,
            patient: c.patientName || `Patient ${c.patientId.slice(0, 8)}`,
            patientMrn: c.patientMrn,
            doctorId: c.doctorId,
            doctor: c.doctorName || `Doctor ${c.doctorId.slice(0, 8)}`,
            encounterId: c.encounterId,
            description: c.description,
            conditions,
            version: c.version,
            createdBy: c.createdBy,
            updatedBy: c.updatedBy,
            createdAt: c.createdAt,
            updatedAt: c.updatedAt,
            isLegacy: c.isLegacy,
          };
        });
        setRows(mapped);
        setTotalCount(res.total ?? mapped.length);
      } else {
        setRows([]);
        setTotalCount(0);
      }
    } catch (err: any) {
      setErrorMessage(
        err?.message || t("Failed to load odontograms from server"),
      );
      setRows([]);
      setTotalCount(0);
    } finally {
      setLoadingCharts(false);
    }
  }, [page, size, search, t]);

  useEffect(() => {
    loadClinicalMetadata();
  }, [loadClinicalMetadata]);

  useEffect(() => {
    loadCharts();
  }, [loadCharts]);

  const colorMap = getLegendColorMap(legends);

  const avatar = (name: string, subtext?: string) => (
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
        {subtext && <small>{subtext}</small>}
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

    saveLegends(updatedLegends);

    if (oldCode && oldCode !== trimmedCode && editing) {
      const nextConditions = { ...editing.conditions };
      let changed = false;
      for (const [tNum, cCode] of Object.entries(nextConditions)) {
        if (cCode === oldCode) {
          nextConditions[Number(tNum)] = trimmedCode;
          changed = true;
        }
      }
      if (changed) {
        setEditing({ ...editing, conditions: nextConditions });
      }
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
        t("Are you sure you want to delete this odontogram legend?"),
      )
    ) {
      saveLegends(legends.filter((l) => l.id !== legendId));
    }
  }

  function handleResetLegends() {
    if (
      window.confirm(
        t("Reset all odontogram legends to the 11 default clinical labels?"),
      )
    ) {
      saveLegends(defaultLegends);
    }
  }

  // Delete Chart
  async function handleDeleteChart(chart: Chart) {
    if (
      !window.confirm(
        t("Are you sure you want to delete this dental examination chart?"),
      )
    ) {
      return;
    }
    setIsDeletingId(chart.id);
    setErrorMessage("");
    try {
      await api(`odontograms/${chart.id}`, { method: "DELETE" });
      await loadCharts();
      setSuccessBanner(t("Dental chart deleted successfully."));
    } catch (err: any) {
      setErrorMessage(err?.message || t("Failed to delete dental chart."));
    } finally {
      setIsDeletingId(null);
    }
  }

  // Print Chart (In-browser full clinical report)
  function handlePrint(row: Chart) {
    const win = window.open("", "_blank");
    if (!win) return;
    const d = win.document;
    d.title = `Odontogram Report - ${row.patient}`;

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
      <div class="meta">Hospital Dental Examination Chart • Chart ID: ${row.id} • Version ${row.version} • Generated on ${new Date().toLocaleDateString()}</div>
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

  // Open Chart-specific PDF
  function handleDownloadPdf(row: Chart) {
    window.open(`/api/hms/odontograms/${row.id}/pdf`, "_blank");
  }

  return (
    <section>
      {/* Top Banner / Alerts */}
      <div className="mb-3 d-flex flex-column gap-2">
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
            placeholder={t("Search patient or doctor")}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </label>

        <div className="d-flex align-items-center gap-2">
          <button
            type="button"
            className="btn-icon-link fs-7 d-flex align-items-center gap-1 px-2 py-1 border rounded"
            onClick={() => loadCharts()}
            disabled={loadingCharts}
            title={t("Refresh list from server")}
          >
            <RefreshCw
              size={14}
              className={loadingCharts ? "animate-spin" : ""}
            />
            <span>{loadingCharts ? t("Loading...") : t("Refresh")}</span>
          </button>

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

          {/* Add Odontogram Button */}
          <button
            className="primary"
            onClick={() => {
              setTooth(1);
              setEditing({
                id: "",
                patientId: patients.length > 0 ? patients[0].id : "",
                patient:
                  patients.length > 0
                    ? `${patients[0].givenName} ${patients[0].familyName}`.trim()
                    : "",
                doctorId: doctorsList.length > 0 ? doctorsList[0].id : "",
                doctor: doctorsList.length > 0 ? doctorsList[0].name : "",
                description: "",
                conditions: {},
                version: 0,
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
              <th>{t("Patient")}</th>
              <th>{t("Attending Doctor")}</th>
              <th>{t("Clinical Notes")}</th>
              <th>{t("Status / Conditions")}</th>
              <th>{t("Action")}</th>
            </tr>
          </thead>
          <tbody>
            {loadingCharts ? (
              <tr>
                <td colSpan={5} className="text-center py-4 text-secondary">
                  <div className="d-inline-flex align-items-center gap-2">
                    <Loader2 size={18} className="animate-spin" />
                    <span>{t("Loading dental examination charts...")}</span>
                  </div>
                </td>
              </tr>
            ) : rows.length > 0 ? (
              rows.map((row) => {
                const conditionCount = Object.keys(row.conditions).length;
                return (
                  <tr key={row.id}>
                    <td>
                      {avatar(
                        row.patient,
                        row.patientMrn ? `MRN: ${row.patientMrn}` : undefined,
                      )}
                    </td>
                    <td>{avatar(row.doctor)}</td>
                    <td
                      style={{
                        maxWidth: "240px",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                      title={row.description}
                    >
                      <span className="fs-7 text-secondary">
                        {row.description || t("Routine Dental Examination")}
                      </span>
                    </td>
                    <td>
                      <div className="d-flex align-items-center gap-2">
                        {row.isLegacy && (
                          <span
                            className="badge-available-stock fs-8 px-2 py-0"
                            style={{
                              backgroundColor: "#f59e0b22",
                              color: "#d97706",
                              borderColor: "#f59e0b",
                            }}
                          >
                            {t("Legacy Chart")}
                          </span>
                        )}
                        <span className="fs-8 text-secondary">
                          {conditionCount}{" "}
                          {conditionCount === 1
                            ? t("tooth charted")
                            : t("teeth charted")}
                        </span>
                        <span className="fs-8 text-secondary font-monospace">
                          v{row.version}
                        </span>
                      </div>
                    </td>
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
                          disabled={isDeletingId === row.id}
                          onClick={() => handleDeleteChart(row)}
                          title={t("Delete odontogram")}
                        >
                          {isDeletingId === row.id ? (
                            <Loader2 size={17} className="animate-spin" />
                          ) : (
                            <Trash2 size={17} />
                          )}
                        </button>
                        <button
                          className="print-action"
                          aria-label={`Print odontogram ${row.patient}`}
                          onClick={() => handlePrint(row)}
                          title={t("Print odontogram report")}
                        >
                          <Printer size={17} />
                        </button>
                        <button
                          className="btn-icon-link"
                          aria-label={`PDF export ${row.patient}`}
                          onClick={() => handleDownloadPdf(row)}
                          title={t("Download PDF")}
                        >
                          <FileDown size={17} />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            ) : (
              <tr>
                <td colSpan={5} className="text-center py-4 text-secondary">
                  {t("No dental charts found.")}
                </td>
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
          {totalCount > 0 ? Math.min((page - 1) * size + 1, totalCount) : 0} -{" "}
          {Math.min(page * size, totalCount)} {t("of")} {totalCount}{" "}
          {t("Results")}
        </span>
        <div>
          {Array.from(
            { length: Math.ceil(totalCount / size) || 1 },
            (_, i) => (
              <button
                className={page === i + 1 ? "primary" : "secondary"}
                key={i}
                onClick={() => setPage(i + 1)}
              >
                {i + 1}
              </button>
            ),
          )}
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

              if (!editing.patientId) {
                setErrorMessage(t("Please select a patient"));
                return;
              }
              if (!editing.doctorId) {
                setErrorMessage(t("Please select an attending doctor"));
                return;
              }

              setIsSaving(true);
              setErrorMessage("");
              setSuccessBanner("");

              const odontogramMap: Record<number, string> = {};
              for (const [tNumStr, code] of Object.entries(
                editing.conditions,
              )) {
                odontogramMap[Number(tNumStr)] = code;
              }

              const payload = {
                patientId: editing.patientId,
                doctorId: editing.doctorId,
                encounterId: editing.encounterId || undefined,
                description:
                  editing.description.trim() ||
                  "Dental Examination Chart",
                odontogram: odontogramMap,
                version: editing.version,
              };

              try {
                if (editing.id) {
                  await api(`odontograms/${editing.id}`, {
                    method: "PUT",
                    body: JSON.stringify(payload),
                  });
                  await loadCharts();
                  setSuccessBanner(t("Dental chart updated successfully."));
                } else {
                  await api("odontograms", {
                    method: "POST",
                    body: JSON.stringify(payload),
                  });
                  await loadCharts();
                  setSuccessBanner(t("New dental chart created successfully."));
                }
                setEditing(null);
              } catch (err: any) {
                const msg = err?.message || "";
                const isConflict =
                  err?.status === 409 ||
                  err?.code?.includes("CONFLICT") ||
                  msg.toLowerCase().includes("conflict") ||
                  msg.toLowerCase().includes("already used") ||
                  msg.toLowerCase().includes("record changed") ||
                  msg.includes("409");

                if (isConflict) {
                  setErrorMessage(
                    t(
                      "Concurrent Edit Conflict: A newer version of this chart was saved by another session. Please close, refresh, and review changes.",
                    ),
                  );
                } else {
                  setErrorMessage(
                    msg || t("Failed to save dental chart to server."),
                  );
                }
              } finally {
                setIsSaving(false);
              }
            }}
          >
            <header className="modal-heading">
              <h2 id="odontogram-title">
                {t(editing.id ? "Edit Odontogram Chart" : "Add Odontogram Chart")}
                {editing.id && (
                  <span className="ms-2 fs-7 text-secondary fw-normal font-monospace">
                    (v{editing.version})
                  </span>
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
                  </span>
                  <select
                    required
                    className="field"
                    value={editing.patientId}
                    onChange={(e) => {
                      const selId = e.target.value;
                      const pat = patients.find((p) => p.id === selId);
                      setEditing({
                        ...editing,
                        patientId: selId,
                        patient: pat
                          ? `${pat.givenName} ${pat.familyName}`.trim()
                          : selId,
                        patientMrn: pat?.mrn,
                      });
                    }}
                  >
                    <option value="">{t("Select Patient")}</option>
                    {patients.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.givenName} {p.familyName} ({p.mrn})
                      </option>
                    ))}
                  </select>
                </label>

                <label className="mt-3">
                  <span className="label">
                    {t("Attending Doctor")}: <b className="text-red-500">*</b>
                  </span>
                  <select
                    required
                    className="field"
                    value={editing.doctorId}
                    onChange={(e) => {
                      const selId = e.target.value;
                      const doc = doctorsList.find((d) => d.id === selId);
                      setEditing({
                        ...editing,
                        doctorId: selId,
                        doctor: doc?.name || selId,
                      });
                    }}
                  >
                    <option value="">{t("Select Attending Doctor")}</option>
                    {doctorsList.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name} {d.department ? `(${d.department})` : ""}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="mt-3">
                  <span className="label">
                    {t("Clinical Examination Notes")}: <b className="text-red-500">*</b>
                  </span>
                  <textarea
                    className="field"
                    required
                    rows={4}
                    placeholder={t(
                      "Record clinical diagnosis, periodontal remarks, treatment plan...",
                    )}
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
                {/* Condition Palette */}
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
                  <span className="label">{t("Tooth Selector")}</span>
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
                  t("Save Dental Chart")
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
