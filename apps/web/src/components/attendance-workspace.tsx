"use client";
import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import {
  Search,
  Filter,
  Pencil,
  Trash2,
  Eye,
  CheckCircle,
  XCircle,
  Users,
  Clock,
  Calendar,
  ClipboardList,
  X,
  CheckCircle2,
  RefreshCw,
  AlertCircle,
} from "lucide-react";
import { useLanguage } from "./language";
import { Modal } from "./modal";
import { people } from "@/lib/legacy";
type Row = { id: string; [key: string]: string };
type Field = {
  key: string;
  label: string;
  type?: string;
  options?: string[];
  required?: boolean;
};
const shiftFields: Field[] = [
  { key: "name", label: "Name", required: true },
  { key: "code", label: "Code", required: true },
  { key: "start", label: "Start Time", type: "time", required: true },
  { key: "end", label: "End Time", type: "time", required: true },
  { key: "grace", label: "Grace Minutes", type: "number" },
  { key: "break", label: "Break Minutes", type: "number" },
  {
    key: "default",
    label: "Is Default",
    options: ["Not Default", "Is Default"],
  },
  { key: "status", label: "Status", options: ["Active", "Inactive"] },
];
const staff: Field = {
  key: "staff",
  label: "Staff",
  options: people,
  required: true,
};
const shift: Field = {
  key: "shift",
  label: "Shift",
  options: ["Day Shift", "Night Shift"],
  required: true,
};
const configs: Record<
  string,
  {
    title: string;
    subtitle: string;
    action: string;
    fields: Field[];
    columns: [string, string][];
  }
> = {
  "attendance-shifts": {
    title: "Attendance Shifts",
    subtitle: "Create and manage shift templates for staff duty plans.",
    action: "Add Shifts",
    fields: shiftFields,
    columns: [
      ["name", "Name"],
      ["code", "Code"],
      ["staffCount", "Total Staff"],
      ["time", "Time"],
      ["grace", "Grace Minutes"],
      ["break", "Break Minutes"],
      ["default", "Is Default"],
      ["status", "Status"],
    ],
  },
  "attendance-assignments": {
    title: "Duty Assignments",
    subtitle: "Assign a shift to staff members and manage active duty periods.",
    action: "Add Duty Assignments",
    fields: [
      staff,
      shift,
      { key: "from", label: "Effective From", type: "date", required: true },
      { key: "to", label: "Effective To", type: "date" },
      { key: "status", label: "Status", options: ["Active", "Inactive"] },
      { key: "note", label: "Note", type: "textarea" },
    ],
    columns: [
      ["staff", "Staff"],
      ["shift", "Shifts"],
      ["period", "Effective Period"],
      ["status", "Status"],
      ["note", "Note"],
    ],
  },
  "attendance-leaves": {
    title: "Leave Requests",
    subtitle: "Review and approve staff leave requests.",
    action: "Leave Request",
    fields: [
      staff,
      { key: "from", label: "From Date", type: "date", required: true },
      { key: "to", label: "To Date", type: "date", required: true },
      { key: "days", label: "Total Days", type: "number", required: true },
      {
        key: "type",
        label: "Type",
        options: ["Casual", "Annual", "Emergency", "Sick"],
      },
      { key: "reason", label: "Reason", type: "textarea", required: true },
    ],
    columns: [
      ["staff", "Staff"],
      ["period", "Period"],
      ["days", "Total Days"],
      ["type", "Type"],
      ["status", "Status"],
      ["reason", "Reason"],
      ["approver", "Approver"],
    ],
  },
  "attendance-requests": {
    title: "Attendance Requests",
    subtitle: "Review staff attendance corrections.",
    action: "",
    fields: [],
    columns: [
      ["staff", "Staff"],
      ["date", "Date"],
      ["shift", "Shifts"],
      ["checkIn", "Check In"],
      ["checkOut", "Check Out"],
      ["reason", "Reason"],
      ["status", "Status"],
    ],
  },
  "attendance-report": {
    title: "Attendance Report",
    subtitle:
      "Review daily attendance by date, shift and status from one place.",
    action: "",
    fields: [],
    columns: [
      ["staff", "Staff"],
      ["role", "Roles"],
      ["shift", "Shift"],
      ["status", "Status"],
      ["checkIn", "Check In"],
      ["checkOut", "Check Out"],
      ["late", "Late In"],
      ["early", "Early Out"],
      ["worked", "Worked"],
      ["overtime", "Over Time"],
      ["break", "Break"],
      ["leaveType", "Leave Type"],
      ["reason", "Reason"],
    ],
  },
  "manage-attendance": {
    title: "Manage Attendance",
    subtitle: "Admin can create and edit attendance records manually.",
    action: "Add Attendance",
    fields: [
      staff,
      { key: "date", label: "Date", type: "date", required: true },
      shift,
      { key: "checkIn", label: "Check In", type: "time", required: true },
      { key: "checkOut", label: "Check Out", type: "time" },
      { key: "break", label: "Break Minutes", type: "number" },
      { key: "remarks", label: "Remarks", type: "textarea" },
    ],
    columns: [
      ["staff", "Staff"],
      ["date", "Date"],
      ["shift", "Shift"],
      ["status", "Status"],
      ["checkIn", "Check In"],
      ["checkOut", "Check Out"],
      ["late", "Late"],
      ["early", "Early Out"],
      ["worked", "Worked"],
      ["overtime", "Over Time"],
      ["breakTime", "Break Time"],
      ["remarks", "Remarks"],
    ],
  },
};
const initial: Record<string, Row[]> = {
  "attendance-shifts": [
    {
      id: "day",
      name: "Day Shift",
      code: "DS",
      start: "08:00",
      end: "17:00",
      grace: "10",
      break: "45",
      default: "Is Default",
      status: "Active",
      staffCount: "6",
    },
    {
      id: "night",
      name: "Night Shift",
      code: "NS",
      start: "20:00",
      end: "05:00",
      grace: "10",
      break: "45",
      default: "Not Default",
      status: "Active",
      staffCount: "0",
    },
  ],
  "attendance-assignments": people.map((staff, i) => ({
    id: `d-${i}`,
    staff,
    shift: "Day Shift",
    from: "2026-10-01",
    to: "",
    status: "Active",
    note: "Auto assigned default shift",
  })),
  "attendance-leaves": [
    {
      id: "l-1",
      staff: people[0],
      from: "2026-10-06",
      to: "2026-10-07",
      days: "2",
      type: "Emergency",
      reason: "Sample leave request",
      status: "Pending",
      approver: "N/A",
    },
  ],
  "attendance-requests": [
    {
      id: "r-1",
      staff: people[1],
      date: "2026-10-04",
      shift: "Day Shift",
      checkIn: "08:30",
      checkOut: "17:00",
      reason: "Missed clock out",
      status: "Pending",
    },
  ],
  "manage-attendance": people.map((staff, i) => ({
    id: `m-${i}`,
    staff,
    date: "2026-10-04",
    shift: "Day Shift",
    checkIn: i === 1 ? "08:30" : "08:00",
    checkOut: "17:00",
    status: i === 1 ? "Late" : "Present",
    break: "45",
    remarks: "N/A",
    role: i % 2 ? "Nurse" : "Doctor",
  })),
};
const minute = (time: string) => {
  const [h, m] = time.split(":").map(Number);
  return h * 60 + m;
};
const duration = (n: number) =>
  `${String(Math.floor(Math.max(0, n) / 60)).padStart(2, "0")}:${String(Math.max(0, n) % 60).padStart(2, "0")}`;
export function AttendanceWorkspace({ id }: { id: string }) {
  const { t } = useLanguage();

  const attendanceTabs = [
    {
      id: "attendance",
      label: "Attendance Dashboard",
      href: "/modules/attendance",
    },
    {
      id: "attendance-shifts",
      label: "Attendance Shifts",
      href: "/modules/attendance-shifts",
    },
    {
      id: "attendance-assignments",
      label: "Duty Assignments",
      href: "/modules/attendance-assignments",
    },
    {
      id: "attendance-leaves",
      label: "Leave Requests",
      href: "/modules/attendance-leaves",
    },
    {
      id: "attendance-requests",
      label: "Attendance Requests",
      href: "/modules/attendance-requests",
    },
    {
      id: "attendance-report",
      label: "Attendance Report",
      href: "/modules/attendance-report",
    },
    {
      id: "manage-attendance",
      label: "Manage Attendance",
      href: "/modules/manage-attendance",
    },
  ];

  const [data, setData] = useState(initial),
    [ready, setReady] = useState(false),
    [search, setSearch] = useState(""),
    [filters, setFilters] = useState(false),
    [status, setStatus] = useState("All"),
    [date, setDate] = useState(""),
    [shiftFilter, setShiftFilter] = useState("All"),
    [page, setPage] = useState(1),
    [size, setSize] = useState(10),
    [editing, setEditing] = useState<Row | null>(null),
    [view, setView] = useState<Row | null>(null),
    [error, setError] = useState("");

  const [apiConnected, setApiConnected] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");

  const loadAttendanceData = useCallback(async () => {
    setIsSyncing(true);
    let connected = false;
    try {
      const [shiftsRes, assignRes] = await Promise.all([
        fetch("/api/hms/attendance/shifts").catch(() => null),
        fetch("/api/hms/attendance/assignments").catch(() => null),
      ]);

      if (shiftsRes && shiftsRes.ok) {
        const d = await shiftsRes.json();
        const raw = Array.isArray(d) ? d : d.shifts || [];
        if (raw.length > 0) {
          const mappedShifts: Row[] = raw.map((s: any) => ({
            id: s.id,
            name: s.name,
            code:
              s.name
                .split(" ")
                .map((w: string) => w[0])
                .join("")
                .toUpperCase() || "SH",
            start: (s.startTime || "08:00:00").slice(0, 5),
            end: (s.endTime || "17:00:00").slice(0, 5),
            grace: String(s.gracePeriodMinutes ?? 10),
            break: String(s.breakDurationMinutes ?? 45),
            default: s.active ? "Is Default" : "Not Default",
            status: s.active ? "Active" : "Inactive",
            staffCount: "0",
          }));
          setData((prev) => ({
            ...prev,
            "attendance-shifts": mappedShifts,
          }));
          connected = true;
        }
      }

      if (assignRes && assignRes.ok) {
        const d = await assignRes.json();
        const raw = Array.isArray(d) ? d : d.assignments || [];
        if (raw.length > 0) {
          const mappedAssigns: Row[] = raw.map((a: any) => ({
            id: a.id,
            staff: a.staffName || a.staffId || "Staff Member",
            shift: a.shiftName || "Day Shift",
            from: a.effectiveFrom || "2026-10-01",
            to: a.effectiveTo || "",
            status: "Active",
            note: "Active assignment",
          }));
          setData((prev) => ({
            ...prev,
            "attendance-assignments": mappedAssigns,
          }));
          connected = true;
        }
      }

      setApiConnected(connected);
    } catch {
      setApiConnected(false);
    } finally {
      setIsSyncing(false);
    }
  }, []);

  useEffect(() => {
    loadAttendanceData();
  }, [loadAttendanceData]);

  useEffect(() => {
    try {
      setData(
        JSON.parse(
          sessionStorage.getItem("hms-attendance-workspace-preview") || "null",
        ) || initial,
      );
    } catch {}
    setReady(true);
  }, []);
  useEffect(() => {
    if (ready)
      sessionStorage.setItem(
        "hms-attendance-workspace-preview",
        JSON.stringify(data),
      );
  }, [data, ready]);
  const source = id === "attendance-report" ? "manage-attendance" : id;
  const config = configs[id];
  const shifts = data["attendance-shifts"] || [];
  const rows = (data[source] || []).filter(
    (row) =>
      Object.values(row)
        .join(" ")
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (status === "All" || row.status === status) &&
      (!date || row.date === date || row.from === date) &&
      (shiftFilter === "All" || row.shift === shiftFilter),
  );
  const pages = Math.max(1, Math.ceil(rows.length / size));
  const current = Math.min(page, pages);
  function update(row: Row) {
    setData((old) => ({
      ...old,
      [source]: (old[source] || []).map((r) => (r.id === row.id ? row : r)),
    }));
  }
  function cell(row: Row, key: string) {
    if (key === "period") return `${row.from} - ${row.to || t("Open")}`;
    if (key === "time") return `${row.start} - ${row.end}`;
    const sh = shifts.find((s) => s.name === row.shift);
    const start = minute(row.checkIn || "00:00"),
      scheduled = minute(sh?.start || "08:00");
    let end = minute(row.checkOut || "00:00"),
      finish = minute(sh?.end || "17:00");
    if (end < start) end += 1440;
    if (finish < scheduled) finish += 1440;
    const worked = row.checkOut
      ? Math.max(0, end - start - Number(row.break || 0))
      : 0;
    if (key === "late") return duration(Math.max(0, start - scheduled));
    if (key === "early")
      return duration(row.checkOut ? Math.max(0, finish - end) : 0);
    if (key === "worked") return duration(worked);
    if (key === "overtime")
      return duration(row.checkOut ? Math.max(0, end - finish) : 0);
    if (key === "breakTime") return duration(Number(row.break || 0));
    return row[key] || "N/A";
  }
  const badge = (value: string) => (
    <span className={`badge attendance-badge status-${value.toLowerCase()}`}>
      {t(value)}
    </span>
  );
  const table = (items: Row[], columns: [string, string][], actions = true) => (
    <div className="legacy-table-wrap">
      <table className="legacy-table">
        <thead>
          <tr>
            {columns.map(([key, label]) => (
              <th key={key}>{t(label)}</th>
            ))}
            {actions && <th>{t("Action")}</th>}
          </tr>
        </thead>
        <tbody>
          {items.map((row) => (
            <tr key={row.id}>
              {columns.map(([key]) => (
                <td key={key}>
                  {key === "status" ? badge(row.status) : cell(row, key)}
                </td>
              ))}
              {actions && (
                <td>
                  <div className="row-actions">
                    {(id === "attendance-shifts" ||
                      id === "manage-attendance") && (
                      <button
                        aria-label={`Edit ${row.name || row.staff}`}
                        onClick={() => {
                          setError("");
                          setEditing({ ...row });
                        }}
                      >
                        <Pencil size={17} />
                      </button>
                    )}
                    {id === "attendance-shifts" ||
                    id === "attendance-assignments" ? (
                      <button
                        aria-label={`Delete ${row.name || row.staff}`}
                        onClick={() => {
                          if (window.confirm(t("Delete this record?")))
                            setData({
                              ...data,
                              [source]: data[source].filter(
                                (r) => r.id !== row.id,
                              ),
                            });
                        }}
                      >
                        <Trash2 size={17} />
                      </button>
                    ) : (
                      <button
                        aria-label={`View ${row.staff}`}
                        onClick={() => setView(row)}
                      >
                        <Eye size={17} />
                      </button>
                    )}
                    {["attendance-leaves", "attendance-requests"].includes(
                      id,
                    ) &&
                      row.status === "Pending" && (
                        <>
                          <button
                            aria-label={`Approve ${row.staff}`}
                            onClick={() =>
                              update({
                                ...row,
                                status: "Approved",
                                approver: "Preview Admin",
                              })
                            }
                          >
                            <CheckCircle color="#00c97b" size={17} />
                          </button>
                          {id === "attendance-leaves" && (
                            <button
                              aria-label={`Reject ${row.staff}`}
                              onClick={() =>
                                update({
                                  ...row,
                                  status: "Rejected",
                                  approver: "Preview Admin",
                                })
                              }
                            >
                              <XCircle color="#ffb400" size={17} />
                            </button>
                          )}
                        </>
                      )}
                  </div>
                </td>
              )}
            </tr>
          ))}
          {!items.length && (
            <tr>
              <td colSpan={columns.length + (actions ? 1 : 0)}>
                {t("No records found")}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
  if (id === "attendance")
    return (
      <section>
        {/* Subtabs Navigation */}
        <div className="module-subtabs-nav" style={{ marginBottom: "20px" }}>
          {attendanceTabs.map((tab) => {
            const isActive = id === tab.id;
            return (
              <Link
                key={tab.id}
                href={tab.href}
                className={`subtab-btn ${isActive ? "active" : ""}`}
              >
                {t(tab.label)}
              </Link>
            );
          })}
        </div>

        {/* Live Backend Connection Banner */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            background: apiConnected
              ? "rgba(16, 185, 129, 0.08)"
              : "rgba(59, 130, 246, 0.08)",
            border: `1px solid ${apiConnected ? "rgba(16, 185, 129, 0.3)" : "rgba(59, 130, 246, 0.25)"}`,
            borderRadius: "8px",
            padding: "10px 16px",
            marginBottom: "16px",
            fontSize: "13px",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            {apiConnected ? (
              <CheckCircle2 size={16} color="#10b981" />
            ) : (
              <AlertCircle size={16} color="#3b82f6" />
            )}
            <span
              style={{
                fontWeight: 500,
                color: apiConnected ? "#10b981" : "#60a5fa",
              }}
            >
              {apiConnected
                ? t("Connected to PostgreSQL Backend (/v1/attendance)")
                : t("Local Clinical Preview Mode (Attendance Ready)")}
            </span>
            <span style={{ color: "#94a3b8" }}>•</span>
            <span style={{ color: "#cbd5e1" }}>
              {shifts.length} {t("shifts")} |{" "}
              {data["attendance-assignments"].length} {t("assignments")}
            </span>
          </div>

          <button
            type="button"
            onClick={() => loadAttendanceData()}
            disabled={isSyncing}
            style={{
              background: "transparent",
              border: "1px solid #475569",
              color: "#e2e8f0",
              borderRadius: "6px",
              padding: "4px 10px",
              fontSize: "12px",
              cursor: "pointer",
              display: "inline-flex",
              alignItems: "center",
              gap: "6px",
            }}
            title={t("Sync Attendance")}
          >
            <RefreshCw size={12} className={isSyncing ? "animate-spin" : ""} />
            {isSyncing ? t("Syncing...") : t("Sync Data")}
          </button>
        </div>

        {apiSuccessBanner && (
          <div
            style={{
              background: "rgba(16, 185, 129, 0.15)",
              border: "1px solid #10b981",
              color: "#34d399",
              padding: "10px 14px",
              borderRadius: "6px",
              marginBottom: "14px",
              fontSize: "13px",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <span>{apiSuccessBanner}</span>
            <button
              onClick={() => setApiSuccessBanner("")}
              style={{
                background: "transparent",
                border: "none",
                color: "#34d399",
                cursor: "pointer",
              }}
            >
              <X size={14} />
            </button>
          </div>
        )}

        <div className="page-heading">
          <div>
            <h1>{t("Attendance Dashboard")}</h1>
            <p className="text-muted">
              {t(
                "Monitor staff attendance, shifts, and leave requests from one place.",
              )}
            </p>
          </div>
        </div>
        <div className="attendance-metrics">
          {[
            ["Total Staff", people.length, "All"],
            ["Checked In", 0, "Active"],
            ["Late Today", 0, "Late"],
            ["On Leave", 0, "Away"],
            ["Absent", people.length, "Absent"],
            [
              "Active Shifts",
              shifts.filter((s) => s.status === "Active").length,
              "Shifts",
            ],
            [
              "Pending Leaves",
              data["attendance-leaves"].filter((r) => r.status === "Pending")
                .length,
              "Pending",
            ],
            [
              "Shift Assignments",
              data["attendance-assignments"].length,
              "Assigned",
            ],
          ].map(([name, n, label], i) => {
            const Icon = [
              Users,
              CheckCircle,
              Clock,
              Users,
              Users,
              Calendar,
              ClipboardList,
              ClipboardList,
            ][i];
            return (
              <div className={`attendance-metric metric-${i}`} key={name}>
                <div>
                  <span>
                    <Icon size={18} />
                  </span>
                  {badge(String(label))}
                </div>
                <small>{t(String(name))}</small>
                <strong>{n}</strong>
              </div>
            );
          })}
        </div>
        <div className="attendance-dashboard-columns">
          <div className="legacy-card">
            <div className="page-heading">
              <h2>{t("Shift Summary")}</h2>
              <Link className="secondary" href="/modules/attendance-shifts">
                {t("Manage Shifts")}
              </Link>
            </div>
            {table(
              shifts,
              [
                ["name", "Name"],
                ["time", "Time"],
                ["grace", "Grace Minutes"],
                ["break", "Break Minutes"],
                ["default", "Is Default"],
                ["status", "Status"],
              ],
              false,
            )}
          </div>
          <div className="legacy-card">
            <div className="page-heading">
              <h2>{t("Recent Leave Requests")}</h2>
              <Link className="secondary" href="/modules/attendance-leaves">
                {t("All Leave")}
              </Link>
            </div>
            {data["attendance-leaves"].slice(0, 5).map((r) => (
              <div className="attendance-leave" key={r.id}>
                <span className="avatar">
                  {r.staff
                    .split(" ")
                    .map((n) => n[0])
                    .join("")}
                </span>
                <div>
                  <strong>{r.staff}</strong>
                  <small>
                    {t(r.type)} {t("Leave")}
                  </small>
                  <small>
                    {r.from} - {r.to}
                  </small>
                </div>
                {badge(r.status)}
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  return (
    <section>
      {/* Subtabs Navigation */}
      <div className="module-subtabs-nav" style={{ marginBottom: "20px" }}>
        {attendanceTabs.map((tab) => {
          const isActive = id === tab.id;
          return (
            <Link
              key={tab.id}
              href={tab.href}
              className={`subtab-btn ${isActive ? "active" : ""}`}
            >
              {t(tab.label)}
            </Link>
          );
        })}
      </div>

      {/* Live Backend Connection Banner */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          background: apiConnected
            ? "rgba(16, 185, 129, 0.08)"
            : "rgba(59, 130, 246, 0.08)",
          border: `1px solid ${apiConnected ? "rgba(16, 185, 129, 0.3)" : "rgba(59, 130, 246, 0.25)"}`,
          borderRadius: "8px",
          padding: "10px 16px",
          marginBottom: "16px",
          fontSize: "13px",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          {apiConnected ? (
            <CheckCircle2 size={16} color="#10b981" />
          ) : (
            <AlertCircle size={16} color="#3b82f6" />
          )}
          <span
            style={{
              fontWeight: 500,
              color: apiConnected ? "#10b981" : "#60a5fa",
            }}
          >
            {apiConnected
              ? t("Connected to PostgreSQL Backend (/v1/attendance)")
              : t("Local Clinical Preview Mode (Attendance Ready)")}
          </span>
          <span style={{ color: "#94a3b8" }}>•</span>
          <span style={{ color: "#cbd5e1" }}>
            {shifts.length} {t("shifts")} |{" "}
            {data["attendance-assignments"].length} {t("assignments")}
          </span>
        </div>

        <button
          type="button"
          onClick={() => loadAttendanceData()}
          disabled={isSyncing}
          style={{
            background: "transparent",
            border: "1px solid #475569",
            color: "#e2e8f0",
            borderRadius: "6px",
            padding: "4px 10px",
            fontSize: "12px",
            cursor: "pointer",
            display: "inline-flex",
            alignItems: "center",
            gap: "6px",
          }}
          title={t("Sync Attendance")}
        >
          <RefreshCw size={12} className={isSyncing ? "animate-spin" : ""} />
          {isSyncing ? t("Syncing...") : t("Sync Data")}
        </button>
      </div>

      {apiSuccessBanner && (
        <div
          style={{
            background: "rgba(16, 185, 129, 0.15)",
            border: "1px solid #10b981",
            color: "#34d399",
            padding: "10px 14px",
            borderRadius: "6px",
            marginBottom: "14px",
            fontSize: "13px",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>{apiSuccessBanner}</span>
          <button
            onClick={() => setApiSuccessBanner("")}
            style={{
              background: "transparent",
              border: "none",
              color: "#34d399",
              cursor: "pointer",
            }}
          >
            <X size={14} />
          </button>
        </div>
      )}

      <div className="page-heading">
        <div>
          <h1>{t(config.title)}</h1>
          <p className="text-muted">{t(config.subtitle)}</p>
        </div>
      </div>
      <div className="table-toolbar">
        <label className="table-search">
          <Search size={16} />
          <input
            placeholder={t("Search")}
            aria-label={t("Search attendance")}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </label>
        <div className="flex gap-3">
          <button
            className="primary"
            aria-label={t("Filter attendance")}
            aria-expanded={filters}
            onClick={() => setFilters(!filters)}
          >
            <Filter size={18} />
          </button>
          {config.action && (
            <button
              className="primary"
              onClick={() => {
                setError("");
                setEditing({
                  id: crypto.randomUUID(),
                  status: id === "attendance-leaves" ? "Pending" : "Active",
                  default: "Not Default",
                  grace: "10",
                  break: "45",
                });
              }}
            >
              {t(config.action)}
            </button>
          )}
        </div>
      </div>
      {filters && (
        <div className="attendance-filters">
          <label>
            {t("Date")}
            <input
              className="field"
              type="date"
              value={date}
              onChange={(e) => {
                setDate(e.target.value);
                setPage(1);
              }}
            />
          </label>
          <label>
            {t("Shift")}
            <select
              className="field"
              value={shiftFilter}
              onChange={(e) => {
                setShiftFilter(e.target.value);
                setPage(1);
              }}
            >
              {["All", ...shifts.map((s) => s.name)].map((s) => (
                <option key={s}>{s}</option>
              ))}
            </select>
          </label>
          <label>
            {t("Status")}
            <select
              className="field"
              value={status}
              onChange={(e) => {
                setStatus(e.target.value);
                setPage(1);
              }}
            >
              {[
                "All",
                "Active",
                "Inactive",
                "Pending",
                "Approved",
                "Rejected",
                "Present",
                "Late",
                "Absent",
              ].map((s) => (
                <option key={s}>{s}</option>
              ))}
            </select>
          </label>
          <button
            className="secondary"
            onClick={() => {
              setDate("");
              setStatus("All");
              setShiftFilter("All");
              setPage(1);
            }}
          >
            {t("Reset")}
          </button>
        </div>
      )}
      {table(rows.slice((current - 1) * size, current * size), config.columns)}
      <div className="reference-pagination">
        <label>
          {t("Show")}
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
          {t("Showing")} {rows.length ? (current - 1) * size + 1 : 0} -{" "}
          {Math.min(current * size, rows.length)} {t("of")} {rows.length}{" "}
          {t("Results")}
        </span>
        <div>
          {Array.from({ length: pages }, (_, i) => (
            <button
              className={current === i + 1 ? "primary" : "secondary"}
              key={i}
              onClick={() => setPage(i + 1)}
            >
              {i + 1}
            </button>
          ))}
        </div>
      </div>
      {editing && (
        <Modal
          titleId="attendance-editor-title"
          onClose={() => setEditing(null)}
          wide
        >
          <form
            className="p-6"
            onSubmit={async (e) => {
              e.preventDefault();
              if (editing.to && editing.from && editing.to < editing.from) {
                setError("End date must be on or after the start date.");
                return;
              }
              let row = { ...editing };
              if (id === "manage-attendance") {
                const sh = shifts.find((s) => s.name === row.shift);
                row.status =
                  minute(row.checkIn) >
                  minute(sh?.start || "08:00") + Number(sh?.grace || 0)
                    ? "Late"
                    : "Present";
              }

              if (id === "attendance-shifts") {
                try {
                  const startTime =
                    (row.start || "08:00") +
                    (row.start.split(":").length === 2 ? ":00" : "");
                  const endTime =
                    (row.end || "17:00") +
                    (row.end.split(":").length === 2 ? ":00" : "");
                  const res = await fetch("/api/hms/attendance/shifts", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                      name: row.name,
                      startTime,
                      endTime,
                      gracePeriodMinutes: Number(row.grace || 10),
                      breakDurationMinutes: Number(row.break || 45),
                      active: row.status !== "Inactive",
                    }),
                  });
                  if (res.ok) {
                    setApiSuccessBanner(
                      t("Shift saved and committed to database!"),
                    );
                  }
                } catch {
                  // preview fallback
                }
              }

              setData((old) => ({
                ...old,
                [source]: (old[source] || []).some((r) => r.id === row.id)
                  ? old[source].map((r) => (r.id === row.id ? row : r))
                  : [row, ...(old[source] || [])],
              }));
              setEditing(null);
            }}
          >
            <header className="modal-heading">
              <h2 id="attendance-editor-title">
                {t(
                  data[source]?.some((r) => r.id === editing.id)
                    ? "Edit " + config.title
                    : config.action,
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
            {error && (
              <p className="error" role="alert">
                {t(error)}
              </p>
            )}
            <div className="legacy-form">
              {config.fields.map((f) => (
                <label key={f.key}>
                  <span className="label">
                    {t(f.label)}:{" "}
                    {f.required && <b className="text-red-500">*</b>}
                  </span>
                  {f.options ? (
                    <select
                      className="field"
                      required={f.required}
                      value={editing[f.key] || ""}
                      onChange={(e) =>
                        setEditing({ ...editing, [f.key]: e.target.value })
                      }
                    >
                      <option value="">{t("Select")}</option>
                      {(f.key === "shift"
                        ? shifts.map((s) => s.name)
                        : f.options
                      ).map((o) => (
                        <option key={o}>{o}</option>
                      ))}
                    </select>
                  ) : f.type === "textarea" ? (
                    <textarea
                      className="field"
                      required={f.required}
                      value={editing[f.key] || ""}
                      onChange={(e) =>
                        setEditing({ ...editing, [f.key]: e.target.value })
                      }
                    />
                  ) : (
                    <input
                      className="field"
                      required={f.required}
                      type={f.type || "text"}
                      min={f.type === "number" ? 0 : undefined}
                      step={f.key === "days" ? 0.5 : undefined}
                      value={editing[f.key] || ""}
                      onChange={(e) =>
                        setEditing({ ...editing, [f.key]: e.target.value })
                      }
                    />
                  )}
                </label>
              ))}
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
      {view && (
        <Modal
          titleId="attendance-view-title"
          onClose={() => setView(null)}
          wide
        >
          <div className="p-6">
            <header className="modal-heading">
              <h2 id="attendance-view-title">
                {t(config.title)}: {view.staff}
              </h2>
              <button aria-label="Close" onClick={() => setView(null)}>
                <X size={20} />
              </button>
            </header>
            <dl className="legacy-form">
              {config.columns.map(([key, label]) => (
                <div key={key}>
                  <dt className="label">{t(label)}</dt>
                  <dd>{cell(view, key)}</dd>
                </div>
              ))}
            </dl>
          </div>
        </Modal>
      )}
    </section>
  );
}
