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
import { shiftRow, saveAttendanceShift } from "@/lib/attendance-shifts";
import {
  assignmentRow,
  saveShiftAssignment,
  leaveRow,
  saveLeaveRequest,
  updateLeaveStatus,
  recordRow,
  updateRecordApproval,
  clockIn,
  clockOut,
  startBreak,
  endBreak,
  StaffOption,
} from "@/lib/attendance-ops";

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
  {
    key: "default",
    label: "Is Default",
    options: ["Not Default", "Is Default"],
  },
  { key: "start", label: "Start Time", type: "time", required: true },
  { key: "end", label: "End Time", type: "time", required: true },
  { key: "grace", label: "Grace Minutes", type: "number" },
  { key: "break", label: "Break Minutes", type: "number" },
  { key: "halfDay", label: "Half Day Minutes", type: "number", required: true },
  { key: "fullDay", label: "Full Day Minutes", type: "number", required: true },
  { key: "status", label: "Status", options: ["Active", "Inactive"] },
];

const staffField: Field = {
  key: "staff",
  label: "Staff",
  options: [],
  required: true,
};

const shiftField: Field = {
  key: "shift",
  label: "Shift",
  options: [],
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
      staffField,
      shiftField,
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
      staffField,
      { key: "from", label: "From Date", type: "date", required: true },
      { key: "to", label: "To Date", type: "date", required: true },
      { key: "days", label: "Total Days", type: "number", required: true },
      {
        key: "type",
        label: "Type",
        options: ["Casual", "Annual", "Emergency", "Sick", "Other"],
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
      ["shift", "Shift"],
      ["status", "Status"],
      ["checkIn", "Check In"],
      ["checkOut", "Check Out"],
      ["late", "Late In"],
      ["early", "Early Out"],
      ["worked", "Worked"],
      ["overtime", "Over Time"],
      ["breakTime", "Break Time"],
      ["remarks", "Remarks"],
    ],
  },
  "manage-attendance": {
    title: "Manage Attendance",
    subtitle: "Admin can create and edit attendance records manually.",
    action: "Add Attendance",
    fields: [
      staffField,
      { key: "date", label: "Date", type: "date", required: true },
      shiftField,
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

const minute = (time: string) => {
  const [h, m] = time.split(":").map(Number);
  return (h || 0) * 60 + (m || 0);
};

const duration = (n: number) =>
  `${String(Math.floor(Math.max(0, n) / 60)).padStart(2, "0")}:${String(Math.max(0, n) % 60).padStart(2, "0")}`;

export function AttendanceWorkspace({ id }: { id: string }) {
  const { t } = useLanguage();

  const attendanceTabs = [
    {
      id: "attendance",
      label: "Attendance",
      href: "/modules/attendance",
    },
    {
      id: "attendance-report",
      label: "Daily Report",
      href: "/modules/attendance-report",
    },
    {
      id: "attendance-shifts",
      label: "Shifts",
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
  ];

  const [data, setData] = useState<Record<string, Row[]>>({
    "attendance-shifts": [],
    "attendance-assignments": [],
    "attendance-leaves": [],
    "attendance-requests": [],
    "attendance-report": [],
    "manage-attendance": [],
  });
  const [staffList, setStaffList] = useState<StaffOption[]>([]);
  const [summary, setSummary] = useState<Record<string, any>>({
    totalCount: 0,
    presentCount: 0,
    lateCount: 0,
    halfDayCount: 0,
    absentCount: 0,
    totalWorkedMinutes: 0,
    totalOvertimeMinutes: 0,
    pendingApprovalCount: 0,
  });
  const [todayRecord, setTodayRecord] = useState<Row | null>(null);
  const [todayOnBreak, setTodayOnBreak] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);

  const [search, setSearch] = useState("");
  const [filters, setFilters] = useState(false);
  const [status, setStatus] = useState("All");
  const [date, setDate] = useState("");
  const [shiftFilter, setShiftFilter] = useState("All");
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(10);
  const [editing, setEditing] = useState<Row | null>(null);
  const [view, setView] = useState<Row | null>(null);
  const [error, setError] = useState("");

  const [apiConnected, setApiConnected] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");
  const [saving, setSaving] = useState(false);

  const loadAttendanceData = useCallback(async () => {
    setIsSyncing(true);
    setApiErrorBanner("");
    try {
      const [
        shiftsRes,
        assignRes,
        leavesRes,
        requestsRes,
        recordsRes,
        staffRes,
        summaryRes,
        todayRes,
      ] = await Promise.all([
        fetch("/api/hms/attendance/shifts").catch(() => null),
        fetch("/api/hms/attendance/assignments").catch(() => null),
        fetch("/api/hms/attendance/leaves").catch(() => null),
        fetch("/api/hms/attendance/records?approvalStatus=submitted").catch(
          () => null,
        ),
        fetch(
          `/api/hms/attendance/records${date ? `?workDate=${encodeURIComponent(date)}` : ""}`,
        ).catch(() => null),
        fetch("/api/hms/attendance/staff").catch(() => null),
        fetch("/api/hms/attendance/summary").catch(() => null),
        fetch("/api/hms/attendance/today").catch(() => null),
      ]);

      const newData: Record<string, Row[]> = {
        "attendance-shifts": [],
        "attendance-assignments": [],
        "attendance-leaves": [],
        "attendance-requests": [],
        "attendance-report": [],
        "manage-attendance": [],
      };

      if (shiftsRes && shiftsRes.ok) {
        const d = await shiftsRes.json();
        const raw = Array.isArray(d) ? d : d.shifts || [];
        if (Array.isArray(raw)) {
          newData["attendance-shifts"] = raw.map(shiftRow);
        }
      }

      if (assignRes && assignRes.ok) {
        const d = await assignRes.json();
        const raw = Array.isArray(d) ? d : d.assignments || [];
        if (Array.isArray(raw)) {
          newData["attendance-assignments"] = raw.map(assignmentRow);
        }
      }

      if (leavesRes && leavesRes.ok) {
        const d = await leavesRes.json();
        const raw = Array.isArray(d) ? d : d.leaves || [];
        if (Array.isArray(raw)) {
          newData["attendance-leaves"] = raw.map(leaveRow);
        }
      }

      if (requestsRes && requestsRes.ok) {
        const d = await requestsRes.json();
        const raw = Array.isArray(d) ? d : d.records || [];
        if (Array.isArray(raw)) {
          newData["attendance-requests"] = raw.map(recordRow);
        }
      }

      if (recordsRes && recordsRes.ok) {
        const d = await recordsRes.json();
        const raw = Array.isArray(d) ? d : d.records || [];
        if (Array.isArray(raw)) {
          const mapped = raw.map(recordRow);
          newData["attendance-report"] = mapped;
          newData["manage-attendance"] = mapped;
        }
      }

      if (staffRes && staffRes.ok) {
        const d = await staffRes.json();
        const list = Array.isArray(d) ? d : d.staff || [];
        if (Array.isArray(list)) {
          setStaffList(list);
        }
      }

      if (summaryRes && summaryRes.ok) {
        const s = await summaryRes.json();
        if (s && typeof s === "object") {
          setSummary(s);
        }
      }

      if (todayRes && todayRes.ok) {
        const d = await todayRes.json();
        if (d && d.record) {
          const mapped = recordRow(d.record);
          setTodayRecord(mapped);
          const breaks = d.record.breaks || [];
          const openBreak = breaks.some((b: any) => !b.endAt);
          setTodayOnBreak(openBreak);
        } else {
          setTodayRecord(null);
          setTodayOnBreak(false);
        }
      }

      setData(newData);
      const connected = Boolean(shiftsRes?.ok && assignRes?.ok);
      setApiConnected(connected);
      if (!connected) {
        setApiErrorBanner(
          "Some attendance data could not be loaded. Retry before making changes.",
        );
      }
    } catch {
      setApiConnected(false);
      setApiErrorBanner("Attendance data could not be loaded. Please retry.");
    } finally {
      setIsSyncing(false);
    }
  }, [date]);

  useEffect(() => {
    loadAttendanceData();
  }, [loadAttendanceData]);

  const shifts = data["attendance-shifts"] || [];
  const config = configs[id] || configs["attendance-shifts"];
  const rows = (data[id] || []).filter(
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

  const handleClockIn = async () => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await clockIn();
      setApiSuccessBanner(t("Successfully clocked in today."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error ? err.message : "Failed to clock in.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleClockOut = async () => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await clockOut();
      setApiSuccessBanner(t("Successfully clocked out."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error ? err.message : "Failed to clock out.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleStartBreak = async () => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await startBreak();
      setApiSuccessBanner(t("Break started."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error ? err.message : "Failed to start break.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleEndBreak = async () => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await endBreak();
      setApiSuccessBanner(t("Break ended. Resumed shift."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error ? err.message : "Failed to end break.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleApproveLeave = async (row: Row) => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await updateLeaveStatus(
        row.id,
        "approved",
        Number(row.version) || 1,
        "Approved by admin",
      );
      setApiSuccessBanner(t("Leave request approved."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error ? err.message : "Failed to approve leave request.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleRejectLeave = async (row: Row) => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await updateLeaveStatus(
        row.id,
        "rejected",
        Number(row.version) || 1,
        "Rejected by admin",
      );
      setApiSuccessBanner(t("Leave request rejected."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error ? err.message : "Failed to reject leave request.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleApproveRecord = async (row: Row) => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await updateRecordApproval(row.id, "approved", "Approved by admin");
      setApiSuccessBanner(t("Attendance record approved."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error
          ? err.message
          : "Failed to approve attendance record.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  const handleRejectRecord = async (row: Row) => {
    setActionLoading(true);
    setApiErrorBanner("");
    try {
      await updateRecordApproval(row.id, "rejected", "Rejected by admin");
      setApiSuccessBanner(t("Attendance record rejected."));
      await loadAttendanceData();
    } catch (err) {
      setApiErrorBanner(
        err instanceof Error
          ? err.message
          : "Failed to reject attendance record.",
      );
    } finally {
      setActionLoading(false);
    }
  };

  function cell(row: Row, key: string) {
    if (key === "period") return `${row.from} - ${row.to || t("Open")}`;
    if (key === "time") return `${row.start} - ${row.end}`;
    if (key === "late") return row.late || "0m";
    if (key === "early") return row.early || "0m";
    if (key === "worked") return row.worked || "0m";
    if (key === "overtime") return row.overtime || "0m";
    if (key === "breakTime") return row.breakTime || "0m";
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
                        disabled={
                          id === "attendance-shifts" ||
                          id === "attendance-assignments"
                        }
                        title={t(
                          "Deletion is not connected yet. No hospital record will be removed.",
                        )}
                        onClick={() => {
                          if (window.confirm(t("Delete this record?")))
                            setData({
                              ...data,
                              [id]: data[id].filter((r) => r.id !== row.id),
                            });
                        }}
                      >
                        <Trash2 size={17} />
                      </button>
                    ) : (
                      <button
                        aria-label={`View ${row.staff || row.name}`}
                        onClick={() => setView(row)}
                      >
                        <Eye size={17} />
                      </button>
                    )}
                    {id === "attendance-leaves" && row.status === "Pending" && (
                      <>
                        <button
                          aria-label={`Approve ${row.staff}`}
                          disabled={actionLoading}
                          onClick={() => handleApproveLeave(row)}
                          title={t("Approve Leave")}
                        >
                          <CheckCircle color="#00c97b" size={17} />
                        </button>
                        <button
                          aria-label={`Reject ${row.staff}`}
                          disabled={actionLoading}
                          onClick={() => handleRejectLeave(row)}
                          title={t("Reject Leave")}
                        >
                          <XCircle color="#ffb400" size={17} />
                        </button>
                      </>
                    )}
                    {id === "attendance-requests" &&
                      (row.status === "Pending" ||
                        row.approvalStatus === "Submitted") && (
                        <>
                          <button
                            aria-label={`Approve ${row.staff}`}
                            disabled={actionLoading}
                            onClick={() => handleApproveRecord(row)}
                            title={t("Approve Attendance")}
                          >
                            <CheckCircle color="#00c97b" size={17} />
                          </button>
                          <button
                            aria-label={`Reject ${row.staff}`}
                            disabled={actionLoading}
                            onClick={() => handleRejectRecord(row)}
                            title={t("Reject Attendance")}
                          >
                            <XCircle color="#ffb400" size={17} />
                          </button>
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

  if (id === "attendance") {
    return (
      <section>
        {apiErrorBanner && (
          <p className="error" role="alert">
            {t(apiErrorBanner)}
          </p>
        )}
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
                ? t("Attendance operational with live hospital APIs.")
                : t("Connecting to hospital attendance services...")}
            </span>
            <span style={{ color: "#94a3b8" }}>•</span>
            <span style={{ color: "#cbd5e1" }}>
              {shifts.length} {t("shifts")} |{" "}
              {data["attendance-assignments"].length} {t("assignments")} |{" "}
              {data["attendance-leaves"].length} {t("leaves")} |{" "}
              {data["attendance-requests"].length} {t("requests")}
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

        {/* Today's Punch Card */}
        <div className="legacy-card" style={{ marginBottom: "20px" }}>
          <div className="page-heading">
            <div>
              <h2>{t("Today's Attendance")}</h2>
              <p className="text-muted">
                {todayRecord
                  ? `${t("Current Status")}: ${t(todayRecord.status || "Checked In")} (${todayRecord.checkIn || ""}${todayRecord.checkOut ? ` - ${todayRecord.checkOut}` : ""})`
                  : t("You have not clocked in today.")}
              </p>
            </div>
            <div style={{ display: "flex", gap: "10px" }}>
              {!todayRecord || !todayRecord.checkIn ? (
                <button
                  type="button"
                  className="primary"
                  disabled={actionLoading}
                  onClick={handleClockIn}
                >
                  <Clock size={16} />
                  {t("Clock In")}
                </button>
              ) : !todayRecord.checkOut ? (
                <>
                  {todayOnBreak ? (
                    <button
                      type="button"
                      className="secondary"
                      disabled={actionLoading}
                      onClick={handleEndBreak}
                    >
                      {t("End Break")}
                    </button>
                  ) : (
                    <button
                      type="button"
                      className="secondary"
                      disabled={actionLoading}
                      onClick={handleStartBreak}
                    >
                      {t("Take Break")}
                    </button>
                  )}
                  <button
                    type="button"
                    className="primary"
                    disabled={actionLoading}
                    onClick={handleClockOut}
                  >
                    <Clock size={16} />
                    {t("Clock Out")}
                  </button>
                </>
              ) : (
                <span className="badge status-present">
                  <CheckCircle
                    size={14}
                    style={{ display: "inline", marginRight: "4px" }}
                  />
                  {t("Completed for Today")}
                </span>
              )}
            </div>
          </div>
        </div>

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
            ["Total Staff", staffList.length || summary.totalCount || 0, "All"],
            ["Checked In", summary.presentCount || 0, "Active"],
            ["Late Today", summary.lateCount || 0, "Late"],
            [
              "On Leave",
              data["attendance-leaves"].filter((r) => r.status === "Approved")
                .length,
              "Away",
            ],
            ["Absent", summary.absentCount || 0, "Absent"],
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
                  {(r.staff || "S")
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
            {!data["attendance-leaves"].length && (
              <p className="text-muted" style={{ padding: "16px 0" }}>
                {t("No leave requests found.")}
              </p>
            )}
          </div>
        </div>
      </section>
    );
  }

  return (
    <section>
      {apiErrorBanner && (
        <p className="error" role="alert">
          {t(apiErrorBanner)}
        </p>
      )}
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
              ? t("Attendance operational with live hospital APIs.")
              : t("Connecting to hospital attendance services...")}
          </span>
          <span style={{ color: "#94a3b8" }}>•</span>
          <span style={{ color: "#cbd5e1" }}>
            {shifts.length} {t("shifts")} |{" "}
            {data["attendance-assignments"].length} {t("assignments")} |{" "}
            {data["attendance-leaves"].length} {t("leaves")} |{" "}
            {data["attendance-requests"].length} {t("requests")}
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
                  halfDay: "240",
                  fullDay: "480",
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
              if (saving) return;
              setError("");

              if (id === "attendance-shifts") {
                setSaving(true);
                try {
                  await saveAttendanceShift(editing);
                  await loadAttendanceData();
                  setApiSuccessBanner(t("Shift saved."));
                  setEditing(null);
                } catch (failure) {
                  setError(
                    failure instanceof Error
                      ? failure.message
                      : "Shift was not saved. Please retry.",
                  );
                } finally {
                  setSaving(false);
                }
                return;
              }

              if (id === "attendance-assignments") {
                if (editing.to && editing.from && editing.to < editing.from) {
                  setError("End date must be on or after the start date.");
                  return;
                }
                setSaving(true);
                try {
                  const staffMatch = staffList.find(
                    (s) => s.name === editing.staff || s.id === editing.staff,
                  );
                  const shiftMatch = shifts.find(
                    (s) => s.name === editing.shift || s.id === editing.shift,
                  );
                  if (!staffMatch) {
                    setError("Please select a valid staff member.");
                    setSaving(false);
                    return;
                  }
                  if (!shiftMatch) {
                    setError("Please select a valid shift.");
                    setSaving(false);
                    return;
                  }
                  await saveShiftAssignment({
                    staffId: staffMatch.id,
                    shiftId: shiftMatch.id,
                    effectiveFrom: editing.from,
                    effectiveTo: editing.to || undefined,
                    note: editing.note || "",
                    active: editing.status !== "Inactive",
                  });
                  await loadAttendanceData();
                  setApiSuccessBanner(t("Duty assignment saved."));
                  setEditing(null);
                } catch (failure) {
                  setError(
                    failure instanceof Error
                      ? failure.message
                      : "Duty assignment was not saved. Please retry.",
                  );
                } finally {
                  setSaving(false);
                }
                return;
              }

              if (id === "attendance-leaves") {
                if (editing.to && editing.from && editing.to < editing.from) {
                  setError("To date must be on or after the from date.");
                  return;
                }
                setSaving(true);
                try {
                  const staffMatch = staffList.find(
                    (s) => s.name === editing.staff || s.id === editing.staff,
                  );
                  await saveLeaveRequest({
                    staffId: staffMatch?.id,
                    fromDate: editing.from,
                    toDate: editing.to,
                    leaveType: editing.type || "Casual",
                    reason: editing.reason || "",
                  });
                  await loadAttendanceData();
                  setApiSuccessBanner(t("Leave request submitted."));
                  setEditing(null);
                } catch (failure) {
                  setError(
                    failure instanceof Error
                      ? failure.message
                      : "Leave request was not saved. Please retry.",
                  );
                } finally {
                  setSaving(false);
                }
                return;
              }

              if (id === "manage-attendance") {
                setSaving(true);
                try {
                  const staffMatch = staffList.find(
                    (s) => s.name === editing.staff || s.id === editing.staff,
                  );
                  const shiftMatch = shifts.find(
                    (s) => s.name === editing.shift || s.id === editing.shift,
                  );
                  if (!staffMatch || !shiftMatch) {
                    setError("Please select a valid staff member and shift.");
                    setSaving(false);
                    return;
                  }
                  const workDate =
                    editing.date || new Date().toISOString().slice(0, 10);
                  const checkInIso = `${workDate}T${editing.checkIn || "08:00"}:00Z`;
                  const checkOutIso = editing.checkOut
                    ? `${workDate}T${editing.checkOut}:00Z`
                    : undefined;
                  const res = await fetch("/api/hms/attendance/records", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                      staffId: staffMatch.id,
                      workDate,
                      shiftId: shiftMatch.id,
                      checkInAt: checkInIso,
                      ...(checkOutIso ? { checkOutAt: checkOutIso } : {}),
                      totalBreakMinutes: Number(editing.break || 0),
                      adminNotes: editing.remarks || "",
                      reason: "Manual admin creation",
                    }),
                  });
                  if (!res.ok) {
                    throw new Error(
                      `Manual attendance record could not be saved (${res.status}).`,
                    );
                  }
                  await loadAttendanceData();
                  setApiSuccessBanner(t("Attendance record saved."));
                  setEditing(null);
                } catch (failure) {
                  setError(
                    failure instanceof Error
                      ? failure.message
                      : "Attendance record was not saved. Please retry.",
                  );
                } finally {
                  setSaving(false);
                }
                return;
              }
            }}
          >
            <header className="modal-heading">
              <h2 id="attendance-editor-title">
                {t(
                  (data[id] || []).some((r) => r.id === editing.id)
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
                  {f.key === "shift" ? (
                    <select
                      className="field"
                      required={f.required}
                      value={editing[f.key] || ""}
                      onChange={(e) =>
                        setEditing({ ...editing, [f.key]: e.target.value })
                      }
                    >
                      <option value="">{t("Select Shift")}</option>
                      {shifts.map((s) => (
                        <option key={s.id} value={s.name}>
                          {s.name} ({s.code || s.start})
                        </option>
                      ))}
                    </select>
                  ) : f.key === "staff" ? (
                    <select
                      className="field"
                      required={f.required}
                      value={editing[f.key] || ""}
                      onChange={(e) =>
                        setEditing({ ...editing, [f.key]: e.target.value })
                      }
                    >
                      <option value="">{t("Select Staff")}</option>
                      {staffList.map((st) => (
                        <option key={st.id} value={st.name}>
                          {st.name} ({st.role})
                        </option>
                      ))}
                    </select>
                  ) : f.options ? (
                    <select
                      className="field"
                      required={f.required}
                      value={editing[f.key] || ""}
                      onChange={(e) =>
                        setEditing({ ...editing, [f.key]: e.target.value })
                      }
                    >
                      <option value="">{t("Select")}</option>
                      {f.options.map((o) => (
                        <option key={o} value={o}>
                          {t(o)}
                        </option>
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
              <button className="primary" disabled={saving}>
                {t(saving ? "Saving..." : "Save")}
              </button>
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
                {t(config.title)}: {view.staff || view.name}
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
