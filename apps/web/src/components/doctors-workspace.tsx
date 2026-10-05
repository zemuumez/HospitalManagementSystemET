"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  X,
  ChevronDown,
  Copy,
  Clock,
  Calendar,
  AlertCircle,
  ArrowLeft,
  Check,
} from "lucide-react";
import { useLanguage } from "./language";

export type DoctorTab =
  | "doctors"
  | "doctor-departments"
  | "schedules"
  | "doctor-holidays"
  | "holidays"
  | "breaks";

interface DoctorItem {
  id: string;
  name: string;
  email: string;
  phone: string;
  department: string;
  qualification: string;
  status: boolean;
  avatarColor: string;
  initials: string;
}

interface DepartmentItem {
  id: string;
  title: string;
  doctorsCount: number;
}

interface ScheduleRow {
  day: string;
  from: string;
  to: string;
}

interface DoctorScheduleItem {
  id: string;
  doctorName: string;
  perPatientTime: string;
  days: ScheduleRow[];
}

interface HolidayItem {
  id: string;
  doctorName: string;
  date: string;
  reason: string;
}

interface BreakItem {
  id: string;
  doctorName: string;
  doctorEmail: string;
  initials: string;
  color: string;
  breakFrom: string;
  breakTo: string;
  dateType: string; // "Every Day" or "23 May 2025"
}

const INITIAL_DOCTORS: DoctorItem[] = [
  {
    id: "doc-1",
    name: "Dr. Nero Patrick",
    email: "neropatrick@gmail.com",
    phone: "+1 555-0192",
    department: "Cardiology",
    qualification: "MD, FACC",
    status: true,
    avatarColor: "#f59e0b",
    initials: "NP",
  },
  {
    id: "doc-2",
    name: "Dr. Muhammad Haseeb",
    email: "haseebmuhammad016@gmail.com",
    phone: "+1 555-0183",
    department: "Neurology",
    qualification: "MBBS, FCPS",
    status: true,
    avatarColor: "#10b981",
    initials: "MH",
  },
  {
    id: "doc-3",
    name: "Dr. Harish Mohan",
    email: "vatsal@gmail.com",
    phone: "+1 555-0144",
    department: "Orthopedics",
    qualification: "MS Ortho",
    status: true,
    avatarColor: "#3b82f6",
    initials: "HM",
  },
  {
    id: "doc-4",
    name: "Dr. John Avery",
    email: "haticadaci@mailinator.com",
    phone: "+1 555-0175",
    department: "General Surgery",
    qualification: "FRCS",
    status: true,
    avatarColor: "#6366f1",
    initials: "JA",
  },
  {
    id: "doc-5",
    name: "Dr. Mark Cruise",
    email: "aby2@gmail.com",
    phone: "+1 555-0166",
    department: "Pediatrics",
    qualification: "MD Pediatrics",
    status: true,
    avatarColor: "#8b5cf6",
    initials: "MC",
  },
  {
    id: "doc-6",
    name: "Dr. Kashif Khan",
    email: "kashif@gmail.com",
    phone: "+1 555-0127",
    department: "Dermatology",
    qualification: "MD Dermatology",
    status: true,
    avatarColor: "#06b6d4",
    initials: "DK",
  },
  {
    id: "doc-7",
    name: "Dr. Shikha Pandey",
    email: "shikha@gmail.com",
    phone: "+1 555-0198",
    department: "Gynecology",
    qualification: "MS OB/GYN",
    status: true,
    avatarColor: "#ec4899",
    initials: "SP",
  },
];

const INITIAL_DEPARTMENTS: DepartmentItem[] = [
  { id: "dep-1", title: "Cardiology", doctorsCount: 4 },
  { id: "dep-2", title: "Neurology", doctorsCount: 3 },
  { id: "dep-3", title: "Orthopedics", doctorsCount: 5 },
  { id: "dep-4", title: "Pediatrics", doctorsCount: 6 },
  { id: "dep-5", title: "Dermatology", doctorsCount: 2 },
  { id: "dep-6", title: "Gynecology", doctorsCount: 4 },
];

const INITIAL_BREAKS: BreakItem[] = [
  {
    id: "brk-1",
    doctorName: "Nero Patrick",
    doctorEmail: "neropatrick@gmail.com",
    initials: "NP",
    color: "#f59e0b",
    breakFrom: "12:00:00",
    breakTo: "13:00:00",
    dateType: "Every Day",
  },
  {
    id: "brk-2",
    doctorName: "Muhammad Haseeb",
    doctorEmail: "haseebmuhammad016@gmail.com",
    initials: "MH",
    color: "#10b981",
    breakFrom: "12:00:00",
    breakTo: "08:00:00",
    dateType: "Every Day",
  },
  {
    id: "brk-3",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    initials: "HM",
    color: "#3b82f6",
    breakFrom: "05:00:00",
    breakTo: "00:05:00",
    dateType: "Every Day",
  },
  {
    id: "brk-4",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    initials: "HM",
    color: "#3b82f6",
    breakFrom: "13:00:00",
    breakTo: "13:30:00",
    dateType: "Every Day",
  },
  {
    id: "brk-5",
    doctorName: "John Avery",
    doctorEmail: "haticadaci@mailinator.com",
    initials: "JA",
    color: "#6366f1",
    breakFrom: "01:05:00",
    breakTo: "23:05:00",
    dateType: "Every Day",
  },
  {
    id: "brk-6",
    doctorName: "Mark Cruise",
    doctorEmail: "aby2@gmail.com",
    initials: "MC",
    color: "#8b5cf6",
    breakFrom: "01:05:00",
    breakTo: "00:05:00",
    dateType: "23 May 2025",
  },
  {
    id: "brk-7",
    doctorName: "Dr Kashif Khan",
    doctorEmail: "kashif@gmail.com",
    initials: "DK",
    color: "#06b6d4",
    breakFrom: "05:30:00",
    breakTo: "05:45:00",
    dateType: "Every Day",
  },
  {
    id: "brk-8",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    initials: "HM",
    color: "#3b82f6",
    breakFrom: "00:10:00",
    breakTo: "01:05:00",
    dateType: "Every Day",
  },
  {
    id: "brk-9",
    doctorName: "Shikha Pandey",
    doctorEmail: "shikha@gmail.com",
    initials: "SP",
    color: "#ec4899",
    breakFrom: "01:00:00",
    breakTo: "02:00:00",
    dateType: "Every Day",
  },
  {
    id: "brk-10",
    doctorName: "Mark Cruise",
    doctorEmail: "aby2@gmail.com",
    initials: "MC",
    color: "#8b5cf6",
    breakFrom: "20:05:00",
    breakTo: "06:05:00",
    dateType: "Every Day",
  },
];

const INITIAL_SCHEDULES: DoctorScheduleItem[] = [
  {
    id: "sch-1",
    doctorName: "Dr. Nero Patrick",
    perPatientTime: "00:15:00",
    days: [
      { day: "Monday", from: "09:00:00", to: "17:00:00" },
      { day: "Tuesday", from: "09:00:00", to: "17:00:00" },
      { day: "Wednesday", from: "09:00:00", to: "17:00:00" },
      { day: "Thursday", from: "09:00:00", to: "17:00:00" },
      { day: "Friday", from: "09:00:00", to: "15:00:00" },
    ],
  },
  {
    id: "sch-2",
    doctorName: "Dr. Muhammad Haseeb",
    perPatientTime: "00:20:00",
    days: [
      { day: "Monday", from: "10:00:00", to: "18:00:00" },
      { day: "Wednesday", from: "10:00:00", to: "18:00:00" },
      { day: "Friday", from: "10:00:00", to: "18:00:00" },
    ],
  },
];

const INITIAL_HOLIDAYS: HolidayItem[] = [
  {
    id: "hol-1",
    doctorName: "Dr. Nero Patrick",
    date: "15 Oct 2026",
    reason: "Medical Conference",
  },
  {
    id: "hol-2",
    doctorName: "Dr. Harish Mohan",
    date: "20 Oct 2026",
    reason: "Annual Leave",
  },
];

const DAYS_OF_WEEK = [
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
  "Sunday",
];

export function DoctorsWorkspace({ id }: { id: string }) {
  const router = useRouter();
  const { t } = useLanguage();

  const activeTab: DoctorTab = (
    [
      "doctors",
      "doctor-departments",
      "schedules",
      "doctor-holidays",
      "holidays",
      "breaks",
    ].includes(id)
      ? id === "holidays"
        ? "doctor-holidays"
        : id
      : "doctors"
  ) as DoctorTab;

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);

  // Data states
  const [doctors, setDoctors] = useState<DoctorItem[]>(INITIAL_DOCTORS);
  const [departments, setDepartments] =
    useState<DepartmentItem[]>(INITIAL_DEPARTMENTS);
  const [breaks, setBreaks] = useState<BreakItem[]>(INITIAL_BREAKS);
  const [schedules, setSchedules] =
    useState<DoctorScheduleItem[]>(INITIAL_SCHEDULES);
  const [holidays, setHolidays] = useState<HolidayItem[]>(INITIAL_HOLIDAYS);

  // Sub-views
  const [viewMode, setViewMode] = useState<
    "list" | "new-schedule" | "add-break"
  >("list");

  // New Schedule form state (Screenshot 174330)
  const [scheduleDoctor, setScheduleDoctor] = useState("");
  const [perPatientTime, setPerPatientTime] = useState("");
  const [scheduleDays, setScheduleDays] = useState<ScheduleRow[]>(
    DAYS_OF_WEEK.map((day) => ({
      day,
      from: "00:00:00",
      to: "00:00:00",
    })),
  );

  // Add Break form state (Screenshot 174558)
  const [breakDoctor, setBreakDoctor] = useState("");
  const [breakDateMode, setBreakDateMode] = useState<
    "Every Day" | "Single Day"
  >("Every Day");
  const [breakSingleDate, setBreakSingleDate] = useState("");
  const [breakFrom, setBreakFrom] = useState("00:00:00");
  const [breakTo, setBreakTo] = useState("00:00:00");

  // Delete modal
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    type: string;
    name: string;
  } | null>(null);

  // Add Department modal
  const [showAddDept, setShowAddDept] = useState(false);
  const [deptTitle, setDeptTitle] = useState("");

  // Copy Monday times to all other days
  const handleCopyTimesToDay = (targetIndex: number) => {
    const monday = scheduleDays[0];
    const updated = [...scheduleDays];
    updated[targetIndex] = {
      ...updated[targetIndex],
      from: monday.from,
      to: monday.to,
    };
    setScheduleDays(updated);
  };

  const handleSaveSchedule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!scheduleDoctor) {
      alert("Please select a doctor");
      return;
    }
    const newSch: DoctorScheduleItem = {
      id: `sch-${Date.now()}`,
      doctorName: scheduleDoctor,
      perPatientTime: perPatientTime || "00:15:00",
      days: scheduleDays,
    };
    setSchedules([newSch, ...schedules]);
    setViewMode("list");
  };

  const handleSaveBreak = (e: React.FormEvent) => {
    e.preventDefault();
    if (!breakDoctor) {
      alert("Please select a doctor");
      return;
    }
    const doc = doctors.find((d) => d.name === breakDoctor);
    const newBreak: BreakItem = {
      id: `brk-${Date.now()}`,
      doctorName: breakDoctor.replace(/^Dr\.\s*/, ""),
      doctorEmail: doc?.email || "doctor@hospital.local",
      initials: doc?.initials || "DR",
      color: doc?.avatarColor || "#5b73e8",
      breakFrom: breakFrom || "12:00:00",
      breakTo: breakTo || "13:00:00",
      dateType:
        breakDateMode === "Every Day"
          ? "Every Day"
          : breakSingleDate || "Today",
    };
    setBreaks([newBreak, ...breaks]);
    setViewMode("list");
    setBreakDoctor("");
    setBreakFrom("00:00:00");
    setBreakTo("00:00:00");
  };

  const confirmDelete = () => {
    if (!deleteTarget) return;
    if (deleteTarget.type === "break") {
      setBreaks(breaks.filter((b) => b.id !== deleteTarget.id));
    } else if (deleteTarget.type === "doctor") {
      setDoctors(doctors.filter((d) => d.id !== deleteTarget.id));
    } else if (deleteTarget.type === "department") {
      setDepartments(departments.filter((d) => d.id !== deleteTarget.id));
    } else if (deleteTarget.type === "schedule") {
      setSchedules(schedules.filter((s) => s.id !== deleteTarget.id));
    }
    setDeleteTarget(null);
  };

  // SCREEN: New Schedule Form (Screenshot 174330)
  if (viewMode === "new-schedule") {
    return (
      <div className="legacy-page-container" style={{ padding: "24px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "24px",
          }}
        >
          <h2 style={{ fontSize: "20px", fontWeight: 600, color: "#f1f5f9" }}>
            {t("New Schedule")}
          </h2>
          <button
            type="button"
            className="btn-action-secondary"
            onClick={() => setViewMode("list")}
          >
            {t("Back")}
          </button>
        </div>

        <form onSubmit={handleSaveSchedule} className="form-card-container">
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "24px",
              marginBottom: "28px",
            }}
          >
            <div>
              <label className="form-label-custom">
                {t("Doctor")}: <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <select
                required
                className="form-select-custom"
                value={scheduleDoctor}
                onChange={(e) => setScheduleDoctor(e.target.value)}
              >
                <option value="">{t("Select Doctor Name")}</option>
                {doctors.map((d) => (
                  <option key={d.id} value={d.name}>
                    {d.name} ({d.department})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="form-label-custom">
                {t("Per Patient Time")}:{" "}
                <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <input
                required
                placeholder={t("Per Patient Time")}
                className="form-input-custom"
                value={perPatientTime}
                onChange={(e) => setPerPatientTime(e.target.value)}
              />
            </div>
          </div>

          {/* Schedule Table */}
          <div className="table-responsive" style={{ marginBottom: "28px" }}>
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th style={{ width: "35%" }}>
                    {t("AVAILABLE ON")}:{" "}
                    <span style={{ color: "#ef4444" }}>*</span>
                  </th>
                  <th style={{ width: "30%" }}>
                    {t("AVAILABLE FROM")}:{" "}
                    <span style={{ color: "#ef4444" }}>*</span>
                  </th>
                  <th style={{ width: "25%" }}>
                    {t("AVAILABLE TO")}:{" "}
                    <span style={{ color: "#ef4444" }}>*</span>
                  </th>
                  <th style={{ width: "10%" }} className="text-end">
                    {t("ACTION")}
                  </th>
                </tr>
              </thead>
              <tbody>
                {scheduleDays.map((row, idx) => (
                  <tr key={row.day}>
                    <td>
                      <input
                        readOnly
                        className="form-input-custom"
                        value={row.day}
                      />
                    </td>
                    <td>
                      <input
                        className="form-input-custom"
                        value={row.from}
                        onChange={(e) => {
                          const updated = [...scheduleDays];
                          updated[idx].from = e.target.value;
                          setScheduleDays(updated);
                        }}
                      />
                    </td>
                    <td>
                      <input
                        className="form-input-custom"
                        value={row.to}
                        onChange={(e) => {
                          const updated = [...scheduleDays];
                          updated[idx].to = e.target.value;
                          setScheduleDays(updated);
                        }}
                      />
                    </td>
                    <td className="text-end">
                      {idx > 0 && (
                        <button
                          type="button"
                          className="btn-icon-blue"
                          style={{
                            width: "36px",
                            height: "36px",
                            borderRadius: "6px",
                          }}
                          title="Copy Monday times"
                          onClick={() => handleCopyTimesToDay(idx)}
                        >
                          <Copy size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div style={{ display: "flex", justifyContent: "flex-end" }}>
            <button type="submit" className="btn-action-blue">
              {t("Save")}
            </button>
          </div>
        </form>
      </div>
    );
  }

  // SCREEN: Add Break Form (Screenshot 174558)
  if (viewMode === "add-break") {
    return (
      <div className="legacy-page-container" style={{ padding: "24px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "24px",
          }}
        >
          <h2 style={{ fontSize: "20px", fontWeight: 600, color: "#f1f5f9" }}>
            {t("Add Break")}
          </h2>
          <button
            type="button"
            className="btn-action-secondary"
            onClick={() => setViewMode("list")}
          >
            {t("Back")}
          </button>
        </div>

        <form onSubmit={handleSaveBreak} className="form-card-container">
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "24px",
              alignItems: "center",
              marginBottom: "24px",
            }}
          >
            <div>
              <label className="form-label-custom">
                {t("Doctor")}: <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <select
                required
                className="form-select-custom"
                value={breakDoctor}
                onChange={(e) => setBreakDoctor(e.target.value)}
              >
                <option value="">{t("Doctor")}</option>
                {doctors.map((d) => (
                  <option key={d.id} value={d.name}>
                    {d.name}
                  </option>
                ))}
              </select>
            </div>

            <div style={{ display: "flex", gap: "24px", marginTop: "16px" }}>
              <label
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  cursor: "pointer",
                  color: "#cbd5e1",
                  fontSize: "14px",
                }}
              >
                <input
                  type="radio"
                  name="breakMode"
                  checked={breakDateMode === "Every Day"}
                  onChange={() => setBreakDateMode("Every Day")}
                />
                {t("Every Day")}
              </label>

              <label
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  cursor: "pointer",
                  color: "#cbd5e1",
                  fontSize: "14px",
                }}
              >
                <input
                  type="radio"
                  name="breakMode"
                  checked={breakDateMode === "Single Day"}
                  onChange={() => setBreakDateMode("Single Day")}
                />
                {t("Single Day")}
              </label>
            </div>
          </div>

          {breakDateMode === "Single Day" && (
            <div style={{ marginBottom: "24px", maxWidth: "48%" }}>
              <label className="form-label-custom">{t("Date")}:</label>
              <input
                type="date"
                className="form-input-custom"
                value={breakSingleDate}
                onChange={(e) => setBreakSingleDate(e.target.value)}
              />
            </div>
          )}

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "24px",
              marginBottom: "28px",
            }}
          >
            <div>
              <label className="form-label-custom">
                {t("From")}: <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <input
                required
                className="form-input-custom"
                value={breakFrom}
                onChange={(e) => setBreakFrom(e.target.value)}
              />
            </div>

            <div>
              <label className="form-label-custom">
                {t("To")}: <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <input
                required
                className="form-input-custom"
                value={breakTo}
                onChange={(e) => setBreakTo(e.target.value)}
              />
            </div>
          </div>

          <div style={{ display: "flex", gap: "12px" }}>
            <button type="submit" className="btn-action-blue">
              {t("Save")}
            </button>
            <button
              type="button"
              className="btn-action-secondary"
              onClick={() => setViewMode("list")}
            >
              {t("Cancel")}
            </button>
          </div>
        </form>
      </div>
    );
  }

  // DEFAULT SCREEN: Sub-Tab Lists
  return (
    <div className="legacy-page-container" style={{ padding: "24px" }}>
      {/* Toolbar */}
      <div className="billing-toolbar">
        <div className="billing-search-box">
          <Search size={16} className="search-icon" />
          <input
            placeholder={t("Search")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        <div className="billing-actions">
          {activeTab === "doctors" && (
            <button
              className="btn-action-blue"
              onClick={() => {
                const name = prompt("Enter Doctor Name:");
                if (name) {
                  setDoctors([
                    {
                      id: `doc-${Date.now()}`,
                      name,
                      email: `${name.toLowerCase().replace(/\s+/g, "")}@hospital.local`,
                      phone: "+1 555-0100",
                      department: "General Medicine",
                      qualification: "MD",
                      status: true,
                      avatarColor: "#5b73e8",
                      initials: name.slice(0, 2).toUpperCase(),
                    },
                    ...doctors,
                  ]);
                }
              }}
            >
              + {t("New Doctor")}
            </button>
          )}

          {activeTab === "doctor-departments" && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddDept(true)}
            >
              + {t("New Doctor Department")}
            </button>
          )}

          {activeTab === "schedules" && (
            <button
              className="btn-action-blue"
              onClick={() => setViewMode("new-schedule")}
            >
              + {t("New Schedule")}
            </button>
          )}

          {activeTab === "doctor-holidays" && (
            <button
              className="btn-action-blue"
              onClick={() => {
                const reason = prompt("Enter Holiday Reason:");
                if (reason) {
                  setHolidays([
                    {
                      id: `hol-${Date.now()}`,
                      doctorName: "Dr. Nero Patrick",
                      date: "25 Oct 2026",
                      reason,
                    },
                    ...holidays,
                  ]);
                }
              }}
            >
              + {t("Add Doctor Holiday")}
            </button>
          )}

          {activeTab === "breaks" && (
            <button
              className="btn-action-blue"
              onClick={() => setViewMode("add-break")}
            >
              {t("Add Break")}
            </button>
          )}
        </div>
      </div>

      {/* Main Card */}
      <div className="billing-card">
        {/* TAB 1: Doctors */}
        {activeTab === "doctors" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("DEPARTMENT")} ↕</th>
                  <th>{t("PHONE")} ↕</th>
                  <th>{t("QUALIFICATION")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {doctors
                  .filter(
                    (d) =>
                      d.name.toLowerCase().includes(search.toLowerCase()) ||
                      d.department.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((doc) => (
                    <tr key={doc.id}>
                      <td>
                        <div className="patient-cell">
                          <div
                            className="avatar-circle"
                            style={{ background: doc.avatarColor }}
                          >
                            {doc.initials}
                          </div>
                          <div className="patient-info">
                            <span className="link-cyan">{doc.name}</span>
                            <span className="patient-email">{doc.email}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="badge-blue-pill">
                          {doc.department}
                        </span>
                      </td>
                      <td style={{ color: "#cbd5e1" }}>{doc.phone}</td>
                      <td style={{ color: "#cbd5e1" }}>{doc.qualification}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={doc.status}
                            onChange={() =>
                              setDoctors(
                                doctors.map((item) =>
                                  item.id === doc.id
                                    ? { ...item, status: !item.status }
                                    : item,
                                ),
                              )
                            }
                          />
                          <span className="slider round" />
                        </label>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-edit"
                            aria-label="Edit"
                            onClick={() => {
                              const newDept = prompt(
                                "Enter Department:",
                                doc.department,
                              );
                              if (newDept) {
                                setDoctors(
                                  doctors.map((d) =>
                                    d.id === doc.id
                                      ? { ...d, department: newDept }
                                      : d,
                                  ),
                                );
                              }
                            }}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: doc.id,
                                type: "doctor",
                                name: doc.name,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 2: Doctor Departments */}
        {activeTab === "doctor-departments" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR DEPARTMENT")} ↕</th>
                  <th>{t("DOCTORS")} ↕</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {departments
                  .filter((dep) =>
                    dep.title.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((dep) => (
                    <tr key={dep.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {dep.title}
                      </td>
                      <td>
                        <span className="badge-blue-pill">
                          {dep.doctorsCount} {t("Doctors")}
                        </span>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: dep.id,
                                type: "department",
                                name: dep.title,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 3: Schedules */}
        {activeTab === "schedules" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("PER PATIENT TIME")} ↕</th>
                  <th>{t("SCHEDULE DAYS")}</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {schedules
                  .filter((sch) =>
                    sch.doctorName.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((sch) => (
                    <tr key={sch.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {sch.doctorName}
                      </td>
                      <td>
                        <span className="badge-blue-pill">
                          {sch.perPatientTime}
                        </span>
                      </td>
                      <td>
                        <div
                          style={{
                            display: "flex",
                            flexWrap: "wrap",
                            gap: "6px",
                          }}
                        >
                          {sch.days
                            .filter((d) => d.from !== "00:00:00")
                            .map((d) => (
                              <span
                                key={d.day}
                                style={{
                                  fontSize: "12px",
                                  background: "#1e2230",
                                  padding: "3px 8px",
                                  borderRadius: "4px",
                                  border: "1px solid #2b3040",
                                }}
                              >
                                {d.day} ({d.from}-{d.to})
                              </span>
                            ))}
                        </div>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: sch.id,
                                type: "schedule",
                                name: `Schedule for ${sch.doctorName}`,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 4: Doctor Holidays */}
        {activeTab === "doctor-holidays" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("DATE")} ↕</th>
                  <th>{t("REASON")}</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {holidays
                  .filter((h) =>
                    h.doctorName.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((h) => (
                    <tr key={h.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {h.doctorName}
                      </td>
                      <td>
                        <span className="badge-blue-pill">{h.date}</span>
                      </td>
                      <td style={{ color: "#cbd5e1" }}>{h.reason}</td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setHolidays(
                                holidays.filter((it) => it.id !== h.id),
                              )
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 5: Breaks (Screenshot 174535) */}
        {activeTab === "breaks" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("DOCTOR")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("BREAK FROM")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("BREAK TO")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("DATE")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {breaks
                  .filter((b) =>
                    b.doctorName.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((brk) => (
                    <tr key={brk.id}>
                      <td>
                        <div className="patient-cell">
                          <div
                            className="avatar-circle"
                            style={{ background: brk.color }}
                          >
                            {brk.initials}
                          </div>
                          <div className="patient-info">
                            <span className="link-cyan">{brk.doctorName}</span>
                            <span className="patient-email">
                              {brk.doctorEmail}
                            </span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="badge-blue-pill">{brk.breakFrom}</span>
                      </td>
                      <td>
                        <span className="badge-blue-pill">{brk.breakTo}</span>
                      </td>
                      <td>
                        <span className="badge-blue-pill">{brk.dateType}</span>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: brk.id,
                                type: "break",
                                name: `Break for ${brk.doctorName}`,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Footer */}
        <div className="billing-footer">
          <div className="billing-pagination-info">
            <span>{t("Show")}</span>
            <select
              className="billing-page-size-select"
              value={pageSize}
              onChange={(e) => setPageSize(Number(e.target.value))}
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
            </select>
            <span>
              {t("Showing")}{" "}
              {activeTab === "doctors"
                ? doctors.length
                : activeTab === "doctor-departments"
                  ? departments.length
                  : activeTab === "schedules"
                    ? schedules.length
                    : activeTab === "doctor-holidays"
                      ? holidays.length
                      : breaks.length}{" "}
              {t("Results")}
            </span>
          </div>

          <div className="billing-pagination-controls">
            <button className="billing-page-btn" disabled>
              ‹
            </button>
            <button className="billing-page-btn is-active">1</button>
            <button className="billing-page-btn">2</button>
            <button className="billing-page-btn">›</button>
          </div>
        </div>
      </div>

      {/* MODAL: New Doctor Department */}
      {showAddDept && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "480px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("New Doctor Department")}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowAddDept(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                if (!deptTitle) return;
                setDepartments([
                  ...departments,
                  {
                    id: `dep-${Date.now()}`,
                    title: deptTitle,
                    doctorsCount: 0,
                  },
                ]);
                setShowAddDept(false);
                setDeptTitle("");
              }}
              className="modal-body-custom"
            >
              <div
                className="form-group-custom"
                style={{ marginBottom: "20px" }}
              >
                <label className="form-label-custom">
                  {t("Title")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder={t("Title")}
                  className="form-input-custom"
                  value={deptTitle}
                  onChange={(e) => setDeptTitle(e.target.value)}
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button type="submit" className="btn-action-blue">
                  {t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowAddDept(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteTarget && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "450px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("Confirm Delete")}</h3>
              <button
                className="modal-close-btn"
                onClick={() => setDeleteTarget(null)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <div
                style={{
                  display: "flex",
                  gap: "14px",
                  alignItems: "flex-start",
                }}
              >
                <AlertCircle size={28} color="#ef4444" />
                <div>
                  <p style={{ margin: "0 0 8px 0" }}>
                    {t("Are you sure you want to delete")} &quot;
                    <strong>{deleteTarget.name}</strong>&quot;?
                  </p>
                  <span style={{ fontSize: "12px", color: "#64748b" }}>
                    {t("This action cannot be undone.")}
                  </span>
                </div>
              </div>
              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                  marginTop: "24px",
                }}
              >
                <button
                  type="button"
                  className="btn-delete-confirm-red"
                  onClick={confirmDelete}
                >
                  {t("Delete")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setDeleteTarget(null)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
