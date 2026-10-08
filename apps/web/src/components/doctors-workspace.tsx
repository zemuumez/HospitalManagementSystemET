"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  X,
  Copy,
  AlertCircle,
  CheckCircle2,
  RefreshCw,
  Eye,
  DollarSign,
} from "lucide-react";
import { useLanguage } from "./language";

export type DoctorTab =
  | "doctors"
  | "doctor-departments"
  | "schedules"
  | "doctor-schedules"
  | "doctor-holidays"
  | "holidays"
  | "breaks"
  | "doctor-opd-charges"
  | "opd-charges";

interface DoctorItem {
  id: string;
  name: string;
  email: string;
  phone: string;
  department: string;
  departmentId?: string;
  specialist?: string;
  qualification: string;
  designation?: string;
  gender?: string;
  dob?: string;
  bloodGroup?: string;
  address1?: string;
  address2?: string;
  city?: string;
  zip?: string;
  description?: string;
  status: boolean;
  opdCharge: number;
  appointmentCharge: number;
  slotMinutes: number;
  version: number;
  avatarColor: string;
  initials: string;
}

interface DepartmentItem {
  id: string;
  title: string;
  description?: string;
  doctorsCount: number;
}

interface ScheduleRow {
  day: string;
  from: string;
  to: string;
}

interface DoctorScheduleItem {
  id: string;
  doctorId: string;
  doctorName: string;
  perPatientTime: string;
  days: ScheduleRow[];
}

interface HolidayItem {
  id: string;
  doctorId: string;
  doctorName: string;
  date: string;
  reason: string;
}

interface BreakItem {
  id: string;
  doctorId: string;
  doctorName: string;
  doctorEmail: string;
  initials: string;
  color: string;
  breakFrom: string;
  breakTo: string;
  dateType: string;
  date?: string;
  everyDay: boolean;
}

interface OPDChargeItem {
  id: string;
  doctorId: string;
  doctorName: string;
  doctorDepartment: string;
  standardCharge: number;
  currencySymbol: string;
}

const DAYS_OF_WEEK = [
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
  "Sunday",
];

const AVATAR_COLORS = [
  "#f59e0b",
  "#10b981",
  "#3b82f6",
  "#6366f1",
  "#8b5cf6",
  "#06b6d4",
  "#ec4899",
];

export function DoctorsWorkspace({ id }: { id: string }) {
  const { t } = useLanguage();

  const rawTab = id as DoctorTab;
  const activeTab: DoctorTab =
    rawTab === "holidays"
      ? "doctor-holidays"
      : rawTab === "doctor-schedules"
        ? "schedules"
        : rawTab === "opd-charges"
          ? "doctor-opd-charges"
          : [
                "doctors",
                "doctor-departments",
                "schedules",
                "doctor-holidays",
                "breaks",
                "doctor-opd-charges",
              ].includes(rawTab)
            ? rawTab
            : "doctors";

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);

  // Subtabs navigation
  const tabs = [
    { id: "doctors", label: "Doctors", href: "/modules/doctors" },
    {
      id: "doctor-departments",
      label: "Doctor Departments",
      href: "/modules/doctor-departments",
    },
    { id: "schedules", label: "Schedules", href: "/modules/schedules" },
    {
      id: "doctor-holidays",
      label: "Doctor Holidays",
      href: "/modules/doctor-holidays",
    },
    { id: "breaks", label: "Breaks", href: "/modules/breaks" },
    {
      id: "doctor-opd-charges",
      label: "OPD Charges",
      href: "/modules/doctor-opd-charges",
    },
  ];

  // Data states - strictly initialized from database, zero mock fixtures
  const [doctors, setDoctors] = useState<DoctorItem[]>([]);
  const [departments, setDepartments] = useState<DepartmentItem[]>([]);
  const [schedules, setSchedules] = useState<DoctorScheduleItem[]>([]);
  const [holidays, setHolidays] = useState<HolidayItem[]>([]);
  const [breaks, setBreaks] = useState<BreakItem[]>([]);
  const [opdCharges, setOpdCharges] = useState<OPDChargeItem[]>([]);

  // Backend live sync state
  const [apiConnected, setApiConnected] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");
  const [currentUser, setCurrentUser] = useState<{ id: string; role: string } | null>(null);
  const isAdmin = currentUser?.role === "admin";

  // Modals & sub-views
  const [viewMode, setViewMode] = useState<
    "list" | "new-schedule" | "add-break"
  >("list");

  // Add Doctor modal state
  const [showAddDoctor, setShowAddDoctor] = useState(false);
  const [docName, setDocName] = useState("");
  const [docEmail, setDocEmail] = useState("");
  const [docPassword, setDocPassword] = useState("DoctorPass1234!");
  const [docDeptId, setDocDeptId] = useState("");
  const [docSpecialist, setDocSpecialist] = useState("");
  const [docDesignation, setDocDesignation] = useState("");
  const [docQual, setDocQual] = useState("");
  const [docGender, setDocGender] = useState("");
  const [docPhone, setDocPhone] = useState("");
  const [docDob, setDocDob] = useState("");
  const [docBloodGroup, setDocBloodGroup] = useState("");
  const [docAddress1, setDocAddress1] = useState("");
  const [docAddress2, setDocAddress2] = useState("");
  const [docCity, setDocCity] = useState("");
  const [docZip, setDocZip] = useState("");
  const [docDescription, setDocDescription] = useState("");
  const [docOpdCharge, setDocOpdCharge] = useState("0");
  const [docApptCharge, setDocApptCharge] = useState("0");
  const [docSlotMinutes, setDocSlotMinutes] = useState("60");

  const resetAddDoctorForm = () => {
    setDocName("");
    setDocEmail("");
    setDocPassword("DoctorPass1234!");
    setDocDeptId("");
    setDocSpecialist("");
    setDocDesignation("");
    setDocQual("");
    setDocGender("");
    setDocPhone("");
    setDocDob("");
    setDocBloodGroup("");
    setDocAddress1("");
    setDocAddress2("");
    setDocCity("");
    setDocZip("");
    setDocDescription("");
    setDocOpdCharge("0");
    setDocApptCharge("0");
    setDocSlotMinutes("60");
  };

  // Edit Doctor modal state
  const [editingDoctor, setEditingDoctor] = useState<DoctorItem | null>(null);
  const [editDocName, setEditDocName] = useState("");
  const [editDocDeptId, setEditDocDeptId] = useState("");
  const [editDocSpecialist, setEditDocSpecialist] = useState("");
  const [editDocDesignation, setEditDocDesignation] = useState("");
  const [editDocQual, setEditDocQual] = useState("");
  const [editDocGender, setEditDocGender] = useState("");
  const [editDocPhone, setEditDocPhone] = useState("");
  const [editDocDob, setEditDocDob] = useState("");
  const [editDocBloodGroup, setEditDocBloodGroup] = useState("");
  const [editDocAddress1, setEditDocAddress1] = useState("");
  const [editDocAddress2, setEditDocAddress2] = useState("");
  const [editDocCity, setEditDocCity] = useState("");
  const [editDocZip, setEditDocZip] = useState("");
  const [editDocDescription, setEditDocDescription] = useState("");
  const [editDocOpdCharge, setEditDocOpdCharge] = useState("0");
  const [editDocApptCharge, setEditDocApptCharge] = useState("0");
  const [editDocSlotMinutes, setEditDocSlotMinutes] = useState("60");

  // Doctor Details modal state
  const [viewingDoctor, setViewingDoctor] = useState<DoctorItem | null>(null);

  // Add Department modal state
  const [showAddDept, setShowAddDept] = useState(false);
  const [deptTitle, setDeptTitle] = useState("");
  const [deptDesc, setDeptDesc] = useState("");

  // Add Holiday modal state
  const [showAddHoliday, setShowAddHoliday] = useState(false);
  const [holidayDoc, setHolidayDoc] = useState("");
  const [holidayDate, setHolidayDate] = useState("");
  const [holidayReason, setHolidayReason] = useState("");

  // New Schedule form state
  const [scheduleDoctorId, setScheduleDoctorId] = useState("");
  const [perPatientTime, setPerPatientTime] = useState("01:00:00");
  const [scheduleDays, setScheduleDays] = useState<ScheduleRow[]>(
    DAYS_OF_WEEK.map((day) => ({
      day,
      from: day === "Sunday" ? "00:00:00" : "09:00:00",
      to: day === "Sunday" ? "00:00:00" : "17:00:00",
    })),
  );

  // Add Break form state
  const [breakDoctorId, setBreakDoctorId] = useState("");
  const [breakDateMode, setBreakDateMode] = useState<"Every Day" | "Single Day">(
    "Every Day",
  );
  const [breakSingleDate, setBreakSingleDate] = useState("");
  const [breakFrom, setBreakFrom] = useState("12:00:00");
  const [breakTo, setBreakTo] = useState("13:00:00");

  // Add / Edit OPD Charge modal state
  const [showOpdModal, setShowOpdModal] = useState(false);
  const [opdModalDoctorId, setOpdModalDoctorId] = useState("");
  const [opdModalAmount, setOpdModalAmount] = useState("350");

  // Delete target state
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    type: "doctor" | "department" | "schedule" | "holiday" | "break" | "opd-charge";
    name: string;
  } | null>(null);

  // Load real records from PostgreSQL API
  const loadDoctorsData = useCallback(async () => {
    setIsSyncing(true);
    let connected = false;
    try {
      const [docRes, depRes, schedRes, holRes, breakRes, opdRes, meRes] =
        await Promise.all([
          fetch("/api/hms/doctors?status=all").catch(() => null),
          fetch("/api/hms/doctor-departments").catch(() => null),
          fetch("/api/hms/doctor-schedules").catch(() => null),
          fetch("/api/hms/doctor-holidays").catch(() => null),
          fetch("/api/hms/doctor-breaks").catch(() => null),
          fetch("/api/hms/doctor-opd-charges").catch(() => null),
          fetch("/api/hms/me").catch(() => null),
        ]);

      if (meRes && meRes.ok) {
        const meData = await meRes.json().catch(() => null);
        const u = meData?.user || meData;
        if (u) setCurrentUser({ id: u.id, role: u.role });
      }

      if (docRes && docRes.ok) {
        const data = await docRes.json();
        const raw = Array.isArray(data) ? data : data.doctors || [];
        setDoctors(
          raw.map((d: any, idx: number) => {
            const rawName = d.name || `Doctor ${d.id.slice(0, 6)}`;
            const cleanName = rawName.startsWith("Dr.")
              ? rawName
              : `Dr. ${rawName}`;
            return {
              id: d.id,
              name: cleanName,
              email: d.email || "",
              phone: d.phone || "",
              department: d.department || "",
              departmentId: d.departmentId || d.department_id || "",
              specialist: d.specialist || "",
              qualification: d.qualification || "",
              designation: d.designation || "",
              gender: d.gender || "",
              dob: d.dob || d.dateOfBirth || "",
              bloodGroup: d.bloodGroup || "",
              address1: d.address1 || "",
              address2: d.address2 || "",
              city: d.city || "",
              zip: d.zip || "",
              description: d.description || "",
              status:
                d.active !== undefined
                  ? Boolean(d.active)
                  : d.status === 1 || d.status === true,
              opdCharge: Number(d.opdCharge ?? d.opd_charge ?? 0),
              appointmentCharge: Number(
                d.appointmentCharge ?? d.appointment_charge ?? 0,
              ),
              slotMinutes: Number(d.slotMinutes ?? d.slot_minutes ?? 60),
              version: Number(d.version ?? 1),
              avatarColor: AVATAR_COLORS[idx % AVATAR_COLORS.length],
              initials:
                cleanName
                  .replace(/^Dr\.\s*/, "")
                  .slice(0, 2)
                  .toUpperCase() || "DR",
            };
          }),
        );
        connected = true;
      }

      if (depRes && depRes.ok) {
        const data = await depRes.json();
        const raw = Array.isArray(data) ? data : data.departments || [];
        setDepartments(
          raw.map((dept: any) => ({
            id: dept.id,
            title: dept.title || dept.name || "Department",
            description: dept.description || "",
            doctorsCount: dept.doctorsCount ?? dept.doctors_count ?? 0,
          })),
        );
        connected = true;
      }

      if (schedRes && schedRes.ok) {
        const data = await schedRes.json();
        const raw = Array.isArray(data) ? data : data.schedules || [];
        setSchedules(
          raw.map((s: any) => ({
            id: s.id,
            doctorId: s.doctorId,
            doctorName: s.doctorName || `Dr. (${s.doctorId?.slice(0, 6)})`,
            perPatientTime: s.perPatientTime || "00:15:00",
            days: Array.isArray(s.days) ? s.days : [],
          })),
        );
        connected = true;
      }

      if (holRes && holRes.ok) {
        const data = await holRes.json();
        const raw = Array.isArray(data) ? data : data.holidays || [];
        setHolidays(
          raw.map((h: any) => ({
            id: h.id,
            doctorId: h.doctorId,
            doctorName: h.doctorName || `Dr. (${h.doctorId?.slice(0, 6)})`,
            date: h.date || "",
            reason: h.reason || "Scheduled Leave",
          })),
        );
        connected = true;
      }

      if (breakRes && breakRes.ok) {
        const data = await breakRes.json();
        const raw = Array.isArray(data) ? data : data.breaks || [];
        setBreaks(
          raw.map((b: any, idx: number) => {
            const rawName = b.doctorName || "Doctor";
            const cleanName = rawName.replace(/^Dr\.\s*/, "");
            return {
              id: b.id,
              doctorId: b.doctorId,
              doctorName: cleanName,
              doctorEmail: b.doctorEmail || "",
              initials: cleanName.slice(0, 2).toUpperCase() || "DR",
              color: AVATAR_COLORS[idx % AVATAR_COLORS.length],
              breakFrom: b.breakFrom || "12:00:00",
              breakTo: b.breakTo || "13:00:00",
              dateType:
                b.dateType || (b.everyDay ? "Every Day" : b.date || "Today"),
              date: b.date || "",
              everyDay: b.everyDay ?? true,
            };
          }),
        );
        connected = true;
      }

      if (opdRes && opdRes.ok) {
        const data = await opdRes.json();
        const raw = Array.isArray(data) ? data : data.charges || [];
        setOpdCharges(
          raw.map((c: any) => ({
            id: c.id,
            doctorId: c.doctorId,
            doctorName: c.doctorName || `Dr. (${c.doctorId?.slice(0, 6)})`,
            doctorDepartment: c.doctorDepartment || "",
            standardCharge: Number(c.standardCharge ?? 0),
            currencySymbol: c.currencySymbol || "ETB",
          })),
        );
        connected = true;
      }

      setApiConnected(connected);
    } catch {
      setApiConnected(false);
    } finally {
      setIsSyncing(false);
    }
  }, []);

  useEffect(() => {
    loadDoctorsData();
  }, [loadDoctorsData]);

  // Handle Create Doctor via real atomic /api/staff endpoint
  const handleCreateDoctor = async (e: React.FormEvent) => {
    e.preventDefault();
    if (
      !docName.trim() ||
      !docEmail.trim() ||
      !docDeptId ||
      !docSpecialist.trim() ||
      !docDesignation.trim() ||
      !docQual.trim() ||
      !docGender
    ) {
      setApiErrorBanner(
        t(
          "Please fill in all required doctor fields: name, email, department, specialist, designation, qualification, and gender.",
        ),
      );
      return;
    }
    if (docPassword.length < 12) {
      setApiErrorBanner(
        t("Check the supplied fields. Passwords need at least 12 characters."),
      );
      return;
    }

    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    const cleanName = docName.startsWith("Dr.") ? docName : `Dr. ${docName}`;

    try {
      const res = await fetch("/api/staff", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: cleanName,
          email: docEmail.trim().toLowerCase(),
          role: "doctor",
          password: docPassword,
          departmentId: docDeptId,
          specialist: docSpecialist.trim(),
          designation: docDesignation.trim(),
          qualification: docQual.trim(),
          gender: docGender,
          phone: docPhone.trim() || undefined,
          dateOfBirth: docDob || undefined,
          bloodGroup: docBloodGroup || undefined,
          address1: docAddress1.trim() || undefined,
          address2: docAddress2.trim() || undefined,
          city: docCity.trim() || undefined,
          zip: docZip.trim() || undefined,
          description: docDescription.trim() || undefined,
          opdCharge: parseFloat(docOpdCharge) || 0,
          appointmentCharge: parseFloat(docApptCharge) || 0,
          slotMinutes: parseInt(docSlotMinutes) || 60,
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(
          t("Doctor account and clinical profile created successfully!"),
        );
        setShowAddDoctor(false);
        resetAddDoctorForm();
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(
          err.error ||
            err.message ||
            t("Failed to create doctor. Verification rejected."),
        );
      }
    } catch {
      setApiErrorBanner(t("Network error creating doctor."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Open Edit Doctor Modal
  const openEditDoctor = (doc: DoctorItem) => {
    setEditingDoctor(doc);
    setEditDocName(doc.name);
    setEditDocDeptId(doc.departmentId || "");
    setEditDocSpecialist(doc.specialist || doc.department || "");
    setEditDocDesignation(doc.designation || "");
    setEditDocQual(doc.qualification || "");
    setEditDocGender(doc.gender || "");
    setEditDocPhone(doc.phone || "");
    setEditDocDob(doc.dob || "");
    setEditDocBloodGroup(doc.bloodGroup || "");
    setEditDocAddress1(doc.address1 || "");
    setEditDocAddress2(doc.address2 || "");
    setEditDocCity(doc.city || "");
    setEditDocZip(doc.zip || "");
    setEditDocDescription(doc.description || "");
    setEditDocOpdCharge(String(doc.opdCharge));
    setEditDocApptCharge(String(doc.appointmentCharge));
    setEditDocSlotMinutes(String(doc.slotMinutes || 60));
  };

  // Handle Edit Doctor via PUT /api/hms/doctors/{id}
  const handleUpdateDoctor = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingDoctor) return;

    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const res = await fetch(`/api/hms/doctors/${editingDoctor.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: editDocName.trim(),
          departmentId: editDocDeptId || undefined,
          specialist: editDocSpecialist.trim(),
          designation: editDocDesignation.trim() || undefined,
          qualification: editDocQual.trim(),
          gender: editDocGender || undefined,
          phone: editDocPhone.trim() || undefined,
          dateOfBirth: editDocDob || undefined,
          bloodGroup: editDocBloodGroup || undefined,
          address1: editDocAddress1.trim() || undefined,
          address2: editDocAddress2.trim() || undefined,
          city: editDocCity.trim() || undefined,
          zip: editDocZip.trim() || undefined,
          description: editDocDescription.trim() || undefined,
          opdCharge: parseFloat(editDocOpdCharge) || 0,
          appointmentCharge: parseFloat(editDocApptCharge) || 0,
          slotMinutes: parseInt(editDocSlotMinutes) || 60,
          version: editingDoctor.version,
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Doctor updated successfully!"));
        setEditingDoctor(null);
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(err.error || t("Failed to update doctor profile."));
      }
    } catch {
      setApiErrorBanner(t("Network error updating doctor profile."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Toggle Doctor Status via PATCH /api/hms/doctors/{id}/status
  const handleToggleDoctorStatus = async (doc: DoctorItem) => {
    const nextStatus = !doc.status;
    try {
      const res = await fetch(`/api/hms/doctors/${doc.id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ active: nextStatus }),
      });

      if (res.ok) {
        setDoctors((prev) =>
          prev.map((d) => (d.id === doc.id ? { ...d, status: nextStatus } : d)),
        );
        setApiSuccessBanner(
          nextStatus
            ? t("Doctor activated successfully!")
            : t("Doctor deactivated successfully!"),
        );
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(err.error || t("Failed to update doctor status."));
      }
    } catch {
      setApiErrorBanner(t("Network error updating doctor status."));
    }
  };

  // Handle Create Department
  const handleCreateDepartment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!deptTitle.trim()) return;

    setIsSubmitting(true);
    try {
      const res = await fetch("/api/hms/doctor-departments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: deptTitle.trim(),
          description: deptDesc.trim() || undefined,
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Doctor department created successfully!"));
        setShowAddDept(false);
        setDeptTitle("");
        setDeptDesc("");
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(err.error || t("Failed to create department."));
      }
    } catch {
      setApiErrorBanner(t("Network error creating department."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Create Doctor Holiday
  const handleCreateHoliday = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!holidayDoc || !holidayReason.trim() || !holidayDate) return;

    setIsSubmitting(true);
    try {
      const res = await fetch("/api/hms/doctor-holidays", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          doctorId: holidayDoc,
          date: holidayDate,
          reason: holidayReason.trim(),
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Holiday leave booked successfully!"));
        setShowAddHoliday(false);
        setHolidayReason("");
        setHolidayDate("");
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        if (res.status === 409) {
          setApiErrorBanner(
            t(
              "Cannot create holiday: doctor has active appointments on this date.",
            ),
          );
        } else {
          setApiErrorBanner(
            err.error ||
              t("Failed to create holiday. Overlapping appointments may exist."),
          );
        }
      }
    } catch {
      setApiErrorBanner(t("Network error creating doctor holiday."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Copy Monday times to target day in schedule form
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

  // Handle Save Doctor Schedule
  const handleSaveSchedule = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!scheduleDoctorId) {
      alert(t("Please select a doctor"));
      return;
    }

    setIsSubmitting(true);
    try {
      const res = await fetch("/api/hms/doctor-schedules", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          doctorId: scheduleDoctorId,
          perPatientTime: perPatientTime || "00:15:00",
          days: scheduleDays,
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Doctor timetable saved successfully!"));
        setViewMode("list");
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(
          err.error ||
            t("Failed to save schedule. Check for existing appointment conflicts."),
        );
      }
    } catch {
      setApiErrorBanner(t("Network error saving timetable."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Save Lunch Break
  const handleSaveBreak = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!breakDoctorId) {
      alert(t("Please select a doctor"));
      return;
    }

    setIsSubmitting(true);
    try {
      const res = await fetch("/api/hms/doctor-breaks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          doctorId: breakDoctorId,
          breakFrom: breakFrom || "12:00:00",
          breakTo: breakTo || "13:00:00",
          everyDay: breakDateMode === "Every Day",
          date: breakDateMode === "Single Day" ? breakSingleDate : undefined,
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Lunch break saved successfully!"));
        setViewMode("list");
        setBreakDoctorId("");
        setBreakFrom("12:00:00");
        setBreakTo("13:00:00");
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(
          err.error ||
            t("Failed to save break. Overlapping appointments may exist."),
        );
      }
    } catch {
      setApiErrorBanner(t("Network error saving lunch break."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Save OPD Charge
  const handleSaveOpdCharge = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!opdModalDoctorId) return;

    setIsSubmitting(true);
    try {
      const res = await fetch("/api/hms/doctor-opd-charges", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          doctorId: opdModalDoctorId,
          standardCharge: parseFloat(opdModalAmount) || 0,
          currencySymbol: "ETB",
        }),
      });

      if (res.ok) {
        setApiSuccessBanner(
          t("Doctor OPD charge saved and synchronized successfully!"),
        );
        setShowOpdModal(false);
        setOpdModalDoctorId("");
        setOpdModalAmount("350");
        await loadDoctorsData();
      } else {
        const err = await res.json().catch(() => ({}));
        setApiErrorBanner(err.error || t("Failed to save doctor OPD charge."));
      }
    } catch {
      setApiErrorBanner(t("Network error saving OPD charge."));
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Confirmed Deletion with real API endpoints and 409 Conflict Protection
  const confirmDelete = async () => {
    if (!deleteTarget) return;

    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      if (deleteTarget.type === "doctor") {
        const res = await fetch(`/api/hms/doctors/${deleteTarget.id}`, {
          method: "DELETE",
        });

        if (res.ok) {
          setApiSuccessBanner(t("Doctor deleted successfully."));
          setDoctors((prev) => prev.filter((d) => d.id !== deleteTarget.id));
        } else {
          const err = await res.json().catch(() => ({}));
          // Conflict protection: referenced doctor cannot be deleted!
          setApiErrorBanner(
            err.error ||
              t(
                "Cannot delete doctor: referenced by appointments, admissions, or payroll records.",
              ),
          );
        }
      } else if (deleteTarget.type === "department") {
        const res = await fetch(
          `/api/hms/doctor-departments/${deleteTarget.id}/archive`,
          { method: "POST" },
        );
        if (res.ok) {
          setApiSuccessBanner(t("Department archived successfully."));
          setDepartments((prev) => prev.filter((d) => d.id !== deleteTarget.id));
        } else {
          const err = await res.json().catch(() => ({}));
          setApiErrorBanner(err.error || t("Failed to archive department."));
        }
      } else if (deleteTarget.type === "schedule") {
        const res = await fetch(
          `/api/hms/doctor-schedules/${deleteTarget.id}`,
          { method: "DELETE" },
        );
        if (res.ok) {
          setApiSuccessBanner(t("Schedule removed successfully."));
          setSchedules((prev) =>
            prev.filter(
              (s) => s.doctorId !== deleteTarget.id && s.id !== deleteTarget.id,
            ),
          );
        } else {
          const err = await res.json().catch(() => ({}));
          setApiErrorBanner(
            err.error ||
              t(
                "Cannot delete schedule: active future appointments are booked for this doctor.",
              ),
          );
        }
      } else if (deleteTarget.type === "holiday") {
        const res = await fetch(`/api/hms/doctor-holidays/${deleteTarget.id}`, {
          method: "DELETE",
        });
        if (res.ok) {
          setApiSuccessBanner(t("Doctor holiday removed successfully."));
          setHolidays((prev) => prev.filter((h) => h.id !== deleteTarget.id));
        } else {
          const err = await res.json().catch(() => ({}));
          setApiErrorBanner(err.error || t("Failed to delete holiday."));
        }
      } else if (deleteTarget.type === "break") {
        const res = await fetch(`/api/hms/doctor-breaks/${deleteTarget.id}`, {
          method: "DELETE",
        });
        if (res.ok) {
          setApiSuccessBanner(t("Doctor lunch break removed successfully."));
          setBreaks((prev) => prev.filter((b) => b.id !== deleteTarget.id));
        } else {
          const err = await res.json().catch(() => ({}));
          setApiErrorBanner(err.error || t("Failed to delete lunch break."));
        }
      } else if (deleteTarget.type === "opd-charge") {
        const res = await fetch(
          `/api/hms/doctor-opd-charges/${deleteTarget.id}`,
          { method: "DELETE" },
        );
        if (res.ok) {
          setApiSuccessBanner(t("Doctor OPD charge reset to 0."));
          setOpdCharges((prev) =>
            prev.filter((c) => c.doctorId !== deleteTarget.id),
          );
          await loadDoctorsData();
        } else {
          const err = await res.json().catch(() => ({}));
          setApiErrorBanner(err.error || t("Failed to reset OPD charge."));
        }
      }
    } catch {
      setApiErrorBanner(t("An unexpected network error occurred."));
    } finally {
      setIsSubmitting(false);
      setDeleteTarget(null);
    }
  };

  // SUB-SCREEN: New Schedule Form
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
                value={scheduleDoctorId}
                onChange={(e) => setScheduleDoctorId(e.target.value)}
              >
                <option value="">{t("Select Doctor Name")}</option>
                {doctors.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name} ({d.department || "No Dept"})
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
                placeholder="00:15:00"
                className="form-input-custom"
                value={perPatientTime}
                onChange={(e) => setPerPatientTime(e.target.value)}
              />
            </div>
          </div>

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
                          title={t("Copy Monday times")}
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
            <button
              type="submit"
              className="btn-action-blue"
              disabled={isSubmitting}
            >
              {isSubmitting ? t("Saving...") : t("Save")}
            </button>
          </div>
        </form>
      </div>
    );
  }

  // SUB-SCREEN: Add Break Form
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
                value={breakDoctorId}
                onChange={(e) => setBreakDoctorId(e.target.value)}
              >
                <option value="">{t("Select Doctor")}</option>
                {doctors.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name} ({d.department || "No Dept"})
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
                required
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
            <button
              type="submit"
              className="btn-action-blue"
              disabled={isSubmitting}
            >
              {isSubmitting ? t("Saving...") : t("Save")}
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

  // DEFAULT VIEW: Sub-tab Lists
  return (
    <div className="legacy-page-container" style={{ padding: "24px" }}>
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

      {apiErrorBanner && (
        <div
          style={{
            background: "rgba(239, 68, 68, 0.15)",
            border: "1px solid #ef4444",
            color: "#f87171",
            padding: "10px 14px",
            borderRadius: "6px",
            marginBottom: "14px",
            fontSize: "13px",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>{apiErrorBanner}</span>
          <button
            onClick={() => setApiErrorBanner("")}
            style={{
              background: "transparent",
              border: "none",
              color: "#f87171",
              cursor: "pointer",
            }}
          >
            <X size={14} />
          </button>
        </div>
      )}

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
          {activeTab === "doctors" && isAdmin && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddDoctor(true)}
            >
              + {t("New Doctor")}
            </button>
          )}

          {activeTab === "doctor-departments" && isAdmin && (
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
              onClick={() => setShowAddHoliday(true)}
            >
              + {t("Add Doctor Holiday")}
            </button>
          )}

          {activeTab === "breaks" && (
            <button
              className="btn-action-blue"
              onClick={() => setViewMode("add-break")}
            >
              + {t("Add Break")}
            </button>
          )}

          {activeTab === "doctor-opd-charges" && (
            <button
              className="btn-action-blue"
              onClick={() => {
                setOpdModalDoctorId(doctors[0]?.id || "");
                setOpdModalAmount("350");
                setShowOpdModal(true);
              }}
            >
              + {t("Set OPD Charge")}
            </button>
          )}
        </div>
      </div>

      {/* Main Table Card */}
      <div className="billing-card">
        {/* TAB 1: Doctors */}
        {activeTab === "doctors" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("DEPARTMENT")} ↕</th>
                  <th>{t("SPECIALIST")} ↕</th>
                  <th>{t("PHONE")} ↕</th>
                  <th>{t("OPD CHARGE")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {doctors.length === 0 ? (
                  <tr>
                    <td colSpan={7} style={{ textAlign: "center", padding: "32px", color: "#94a3b8" }}>
                      {t("No doctor records found in database.")}
                    </td>
                  </tr>
                ) : (
                  doctors
                    .filter(
                      (d) =>
                        d.name.toLowerCase().includes(search.toLowerCase()) ||
                        d.department.toLowerCase().includes(search.toLowerCase()) ||
                        (d.specialist && d.specialist.toLowerCase().includes(search.toLowerCase())),
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
                              <button
                                type="button"
                                className="link-cyan"
                                style={{
                                  background: "none",
                                  border: "none",
                                  padding: 0,
                                  cursor: "pointer",
                                  textAlign: "left",
                                }}
                                onClick={() => setViewingDoctor(doc)}
                              >
                                {doc.name}
                              </button>
                              <span className="patient-email">{doc.email}</span>
                            </div>
                          </div>
                        </td>
                        <td>
                          <span className="badge-blue-pill">
                            {doc.department || t("General")}
                          </span>
                        </td>
                        <td style={{ color: "#cbd5e1" }}>
                          {doc.specialist || doc.qualification || "-"}
                        </td>
                        <td style={{ color: "#cbd5e1" }}>{doc.phone || "-"}</td>
                        <td style={{ color: "#34d399", fontWeight: 500 }}>
                          {doc.opdCharge} ETB
                        </td>
                        <td>
                          <label className="switch-toggle">
                            <input
                              type="checkbox"
                              checked={doc.status}
                              disabled={!isAdmin}
                              onChange={() => handleToggleDoctorStatus(doc)}
                            />
                            <span className="slider round" />
                          </label>
                        </td>
                        <td className="text-end">
                          <div
                            className="action-buttons"
                            style={{ justifyContent: "flex-end", gap: "8px" }}
                          >
                            <button
                              className="action-btn-edit"
                              aria-label="View Details"
                              title={t("View Profile")}
                              onClick={() => setViewingDoctor(doc)}
                            >
                              <Eye size={16} />
                            </button>
                            {(isAdmin || doc.id === currentUser?.id) && (
                              <button
                                className="action-btn-edit"
                                aria-label="Edit"
                                title={t("Edit Doctor")}
                                onClick={() => openEditDoctor(doc)}
                              >
                                <Edit2 size={16} />
                              </button>
                            )}
                            {isAdmin && (
                              <button
                                className="action-btn-delete"
                                aria-label="Delete"
                                title={t("Delete Doctor")}
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
                            )}
                          </div>
                        </td>
                      </tr>
                    ))
                )}
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
                {departments.length === 0 ? (
                  <tr>
                    <td colSpan={3} style={{ textAlign: "center", padding: "32px", color: "#94a3b8" }}>
                      {t("No doctor departments found.")}
                    </td>
                  </tr>
                ) : (
                  departments
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
                          {isAdmin && (
                            <div
                              className="action-buttons"
                              style={{ justifyContent: "flex-end" }}
                            >
                              <button
                                className="action-btn-delete"
                                aria-label="Archive Department"
                                title={t("Archive Department")}
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
                          )}
                        </td>
                      </tr>
                    ))
                )}
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
                {schedules.length === 0 ? (
                  <tr>
                    <td colSpan={4} style={{ textAlign: "center", padding: "32px", color: "#94a3b8" }}>
                      {t("No doctor timetables found.")}
                    </td>
                  </tr>
                ) : (
                  schedules
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
                              aria-label="Delete Schedule"
                              title={t("Delete Schedule")}
                              onClick={() =>
                                setDeleteTarget({
                                  id: sch.doctorId || sch.id,
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
                    ))
                )}
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
                {holidays.length === 0 ? (
                  <tr>
                    <td colSpan={4} style={{ textAlign: "center", padding: "32px", color: "#94a3b8" }}>
                      {t("No doctor holidays registered.")}
                    </td>
                  </tr>
                ) : (
                  holidays
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
                              aria-label="Delete Holiday"
                              title={t("Delete Holiday")}
                              onClick={() =>
                                setDeleteTarget({
                                  id: h.id,
                                  type: "holiday",
                                  name: `Holiday on ${h.date} for ${h.doctorName}`,
                                })
                              }
                            >
                              <Trash2 size={16} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 5: Breaks */}
        {activeTab === "breaks" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("BREAK FROM")} ↕</th>
                  <th>{t("BREAK TO")} ↕</th>
                  <th>{t("DATE")} ↕</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {breaks.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ textAlign: "center", padding: "32px", color: "#94a3b8" }}>
                      {t("No doctor breaks scheduled.")}
                    </td>
                  </tr>
                ) : (
                  breaks
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
                              aria-label="Delete Break"
                              title={t("Delete Break")}
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
                    ))
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 6: Doctor OPD Charges Master */}
        {activeTab === "doctor-opd-charges" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("DEPARTMENT")} ↕</th>
                  <th>{t("STANDARD OPD CHARGE")} ↕</th>
                  <th>{t("CURRENCY")}</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {opdCharges.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ textAlign: "center", padding: "32px", color: "#94a3b8" }}>
                      {t("No doctor OPD charges registered.")}
                    </td>
                  </tr>
                ) : (
                  opdCharges
                    .filter(
                      (c) =>
                        c.doctorName.toLowerCase().includes(search.toLowerCase()) ||
                        c.doctorDepartment.toLowerCase().includes(search.toLowerCase()),
                    )
                    .map((charge) => (
                      <tr key={charge.id}>
                        <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                          {charge.doctorName}
                        </td>
                        <td>
                          <span className="badge-blue-pill">
                            {charge.doctorDepartment || t("General")}
                          </span>
                        </td>
                        <td style={{ color: "#34d399", fontWeight: 600 }}>
                          {charge.standardCharge.toFixed(2)}
                        </td>
                        <td style={{ color: "#94a3b8" }}>
                          {charge.currencySymbol}
                        </td>
                        <td className="text-end">
                          {isAdmin && (
                            <div
                              className="action-buttons"
                              style={{ justifyContent: "flex-end", gap: "8px" }}
                            >
                              <button
                                className="action-btn-edit"
                                aria-label="Edit OPD Charge"
                                title={t("Edit Charge")}
                                onClick={() => {
                                  setOpdModalDoctorId(charge.doctorId);
                                  setOpdModalAmount(String(charge.standardCharge));
                                  setShowOpdModal(true);
                                }}
                              >
                                <Edit2 size={16} />
                              </button>
                              <button
                                className="action-btn-delete"
                                aria-label="Reset OPD Charge"
                                title={t("Reset Charge")}
                                onClick={() =>
                                  setDeleteTarget({
                                    id: charge.doctorId,
                                    type: "opd-charge",
                                    name: `OPD charge for ${charge.doctorName}`,
                                  })
                                }
                              >
                                <Trash2 size={16} />
                              </button>
                            </div>
                          )}
                        </td>
                      </tr>
                    ))
                )}
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
                      : activeTab === "breaks"
                        ? breaks.length
                        : opdCharges.length}{" "}
              {t("Results")}
            </span>
          </div>

          <div className="billing-pagination-controls">
            <button className="billing-page-btn" disabled>
              ‹
            </button>
            <button className="billing-page-btn is-active">1</button>
            <button className="billing-page-btn">›</button>
          </div>
        </div>
      </div>

      {/* MODAL: New Doctor */}
      {showAddDoctor && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{
              maxWidth: "680px",
              width: "100%",
              maxHeight: "90vh",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("New Doctor")}</h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowAddDoctor(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={handleCreateDoctor}
              className="modal-body-custom"
              style={{ overflowY: "auto", padding: "20px 24px" }}
            >
              {apiErrorBanner && (
                <div
                  style={{
                    marginBottom: "16px",
                    padding: "10px 14px",
                    background: "rgba(239, 68, 68, 0.15)",
                    border: "1px solid #ef4444",
                    borderRadius: "6px",
                    color: "#fca5a5",
                    fontSize: "13px",
                  }}
                >
                  {apiErrorBanner}
                </div>
              )}

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Doctor Name")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Dr. John Doe"
                    className="form-input-custom"
                    value={docName}
                    onChange={(e) => setDocName(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Email")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    type="email"
                    required
                    placeholder="doctor@hospital.local"
                    className="form-input-custom"
                    value={docEmail}
                    onChange={(e) => setDocEmail(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Password")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    type="password"
                    required
                    minLength={12}
                    className="form-input-custom"
                    value={docPassword}
                    onChange={(e) => setDocPassword(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Department")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <select
                    required
                    className="form-select-custom"
                    value={docDeptId}
                    onChange={(e) => setDocDeptId(e.target.value)}
                  >
                    <option value="">{t("Select Department")}</option>
                    {departments.map((dep) => (
                      <option key={dep.id} value={dep.id}>
                        {dep.title}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Specialist")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Cardiologist, Neurologist..."
                    className="form-input-custom"
                    value={docSpecialist}
                    onChange={(e) => setDocSpecialist(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Designation")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Senior Consultant"
                    className="form-input-custom"
                    value={docDesignation}
                    onChange={(e) => setDocDesignation(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Qualification")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="MBBS, MD"
                    className="form-input-custom"
                    value={docQual}
                    onChange={(e) => setDocQual(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Gender")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <select
                    required
                    className="form-select-custom"
                    value={docGender}
                    onChange={(e) => setDocGender(e.target.value)}
                  >
                    <option value="">{t("Select Gender")}</option>
                    <option value="male">{t("Male")}</option>
                    <option value="female">{t("Female")}</option>
                    <option value="other">{t("Other")}</option>
                  </select>
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Date of Birth")}:</label>
                  <input
                    type="date"
                    className="form-input-custom"
                    value={docDob}
                    onChange={(e) => setDocDob(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Blood Group")}:</label>
                  <select
                    className="form-select-custom"
                    value={docBloodGroup}
                    onChange={(e) => setDocBloodGroup(e.target.value)}
                  >
                    <option value="">{t("Select Blood Group")}</option>
                    <option value="A+">A+</option>
                    <option value="A-">A-</option>
                    <option value="B+">B+</option>
                    <option value="B-">B-</option>
                    <option value="AB+">AB+</option>
                    <option value="AB-">AB-</option>
                    <option value="O+">O+</option>
                    <option value="O-">O-</option>
                  </select>
                </div>
                <div>
                  <label className="form-label-custom">{t("Phone")}:</label>
                  <input
                    placeholder="+251 91 123 4567"
                    className="form-input-custom"
                    value={docPhone}
                    onChange={(e) => setDocPhone(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Slot Duration")}:</label>
                  <select
                    className="form-select-custom"
                    value={docSlotMinutes}
                    onChange={(e) => setDocSlotMinutes(e.target.value)}
                  >
                    <option value="15">15 {t("minutes")}</option>
                    <option value="30">30 {t("minutes")}</option>
                    <option value="45">45 {t("minutes")}</option>
                    <option value="60">60 {t("minutes")}</option>
                    <option value="90">90 {t("minutes")}</option>
                    <option value="120">120 {t("minutes")}</option>
                  </select>
                </div>
                <div>
                  <label className="form-label-custom">{t("OPD Charge (ETB)")}:</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    className="form-input-custom"
                    value={docOpdCharge}
                    onChange={(e) => setDocOpdCharge(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Appt Charge (ETB)")}:</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    className="form-input-custom"
                    value={docApptCharge}
                    onChange={(e) => setDocApptCharge(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Address Line 1")}:</label>
                  <input
                    placeholder="Street address"
                    className="form-input-custom"
                    value={docAddress1}
                    onChange={(e) => setDocAddress1(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Address Line 2")}:</label>
                  <input
                    placeholder="Apartment, suite, unit"
                    className="form-input-custom"
                    value={docAddress2}
                    onChange={(e) => setDocAddress2(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("City")}:</label>
                  <input
                    placeholder="City"
                    className="form-input-custom"
                    value={docCity}
                    onChange={(e) => setDocCity(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Zip Code")}:</label>
                  <input
                    placeholder="Postal code"
                    className="form-input-custom"
                    value={docZip}
                    onChange={(e) => setDocZip(e.target.value)}
                  />
                </div>
              </div>

              <div style={{ marginBottom: "20px" }}>
                <label className="form-label-custom">{t("Description / Bio")}:</label>
                <textarea
                  rows={2}
                  placeholder="Doctor description, medical background, notes..."
                  className="form-textarea-custom"
                  value={docDescription}
                  onChange={(e) => setDocDescription(e.target.value)}
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? t("Creating...") : t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowAddDoctor(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: Edit Doctor */}
      {editingDoctor && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{
              maxWidth: "680px",
              width: "100%",
              maxHeight: "90vh",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("Edit Doctor")} - {editingDoctor.name}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setEditingDoctor(null)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={handleUpdateDoctor}
              className="modal-body-custom"
              style={{ overflowY: "auto", padding: "20px 24px" }}
            >
              {apiErrorBanner && (
                <div
                  style={{
                    marginBottom: "16px",
                    padding: "10px 14px",
                    background: "rgba(239, 68, 68, 0.15)",
                    border: "1px solid #ef4444",
                    borderRadius: "6px",
                    color: "#fca5a5",
                    fontSize: "13px",
                  }}
                >
                  {apiErrorBanner}
                </div>
              )}

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Doctor Name")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    className="form-input-custom"
                    value={editDocName}
                    onChange={(e) => setEditDocName(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Department")}:</label>
                  <select
                    className="form-select-custom"
                    value={editDocDeptId}
                    onChange={(e) => setEditDocDeptId(e.target.value)}
                  >
                    <option value="">{t("Select Department")}</option>
                    {departments.map((dep) => (
                      <option key={dep.id} value={dep.id}>
                        {dep.title}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Specialist")}:</label>
                  <input
                    className="form-input-custom"
                    value={editDocSpecialist}
                    onChange={(e) => setEditDocSpecialist(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Designation")}:</label>
                  <input
                    className="form-input-custom"
                    value={editDocDesignation}
                    onChange={(e) => setEditDocDesignation(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Qualification")}:</label>
                  <input
                    className="form-input-custom"
                    value={editDocQual}
                    onChange={(e) => setEditDocQual(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Gender")}:</label>
                  <select
                    className="form-select-custom"
                    value={editDocGender}
                    onChange={(e) => setEditDocGender(e.target.value)}
                  >
                    <option value="">{t("Select Gender")}</option>
                    <option value="male">{t("Male")}</option>
                    <option value="female">{t("Female")}</option>
                    <option value="other">{t("Other")}</option>
                  </select>
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Date of Birth")}:</label>
                  <input
                    type="date"
                    className="form-input-custom"
                    value={editDocDob}
                    onChange={(e) => setEditDocDob(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Blood Group")}:</label>
                  <select
                    className="form-select-custom"
                    value={editDocBloodGroup}
                    onChange={(e) => setEditDocBloodGroup(e.target.value)}
                  >
                    <option value="">{t("Select Blood Group")}</option>
                    <option value="A+">A+</option>
                    <option value="A-">A-</option>
                    <option value="B+">B+</option>
                    <option value="B-">B-</option>
                    <option value="AB+">AB+</option>
                    <option value="AB-">AB-</option>
                    <option value="O+">O+</option>
                    <option value="O-">O-</option>
                  </select>
                </div>
                <div>
                  <label className="form-label-custom">{t("Phone")}:</label>
                  <input
                    className="form-input-custom"
                    value={editDocPhone}
                    onChange={(e) => setEditDocPhone(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Slot Duration")}:</label>
                  <select
                    className="form-select-custom"
                    value={editDocSlotMinutes}
                    onChange={(e) => setEditDocSlotMinutes(e.target.value)}
                  >
                    <option value="15">15 {t("minutes")}</option>
                    <option value="30">30 {t("minutes")}</option>
                    <option value="45">45 {t("minutes")}</option>
                    <option value="60">60 {t("minutes")}</option>
                    <option value="90">90 {t("minutes")}</option>
                    <option value="120">120 {t("minutes")}</option>
                  </select>
                </div>
                <div>
                  <label className="form-label-custom">{t("OPD Charge (ETB)")}:</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    className="form-input-custom"
                    value={editDocOpdCharge}
                    onChange={(e) => setEditDocOpdCharge(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Appt Charge (ETB)")}:</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    className="form-input-custom"
                    value={editDocApptCharge}
                    onChange={(e) => setEditDocApptCharge(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Address Line 1")}:</label>
                  <input
                    placeholder="Street address"
                    className="form-input-custom"
                    value={editDocAddress1}
                    onChange={(e) => setEditDocAddress1(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Address Line 2")}:</label>
                  <input
                    placeholder="Apartment, suite, unit"
                    className="form-input-custom"
                    value={editDocAddress2}
                    onChange={(e) => setEditDocAddress2(e.target.value)}
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("City")}:</label>
                  <input
                    placeholder="City"
                    className="form-input-custom"
                    value={editDocCity}
                    onChange={(e) => setEditDocCity(e.target.value)}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Zip Code")}:</label>
                  <input
                    placeholder="Postal code"
                    className="form-input-custom"
                    value={editDocZip}
                    onChange={(e) => setEditDocZip(e.target.value)}
                  />
                </div>
              </div>

              <div style={{ marginBottom: "20px" }}>
                <label className="form-label-custom">{t("Description / Bio")}:</label>
                <textarea
                  rows={2}
                  placeholder="Doctor description, medical background, notes..."
                  className="form-textarea-custom"
                  value={editDocDescription}
                  onChange={(e) => setEditDocDescription(e.target.value)}
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? t("Saving...") : t("Save Changes")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setEditingDoctor(null)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: View Doctor Profile Details */}
      {viewingDoctor && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{
              maxWidth: "620px",
              width: "100%",
              maxHeight: "90vh",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("Doctor Profile")} - {viewingDoctor.name}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setViewingDoctor(null)}
              >
                <X size={18} />
              </button>
            </div>
            <div
              className="modal-body-custom"
              style={{ overflowY: "auto", padding: "20px 24px" }}
            >
              <div
                style={{
                  display: "flex",
                  gap: "16px",
                  alignItems: "center",
                  marginBottom: "20px",
                  paddingBottom: "16px",
                  borderBottom: "1px solid #334155",
                }}
              >
                <div
                  className="avatar-circle"
                  style={{
                    background: viewingDoctor.avatarColor,
                    width: "56px",
                    height: "56px",
                    fontSize: "20px",
                  }}
                >
                  {viewingDoctor.initials}
                </div>
                <div>
                  <h4
                    className="doctor-modal-heading"
                    style={{ margin: "0 0 4px 0", fontSize: "18px" }}
                  >
                    {viewingDoctor.name}
                  </h4>
                  <div className="doctor-modal-sub" style={{ fontSize: "13px" }}>
                    {viewingDoctor.email}{" "}
                    {viewingDoctor.phone ? `• ${viewingDoctor.phone}` : ""}
                  </div>
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "12px",
                  fontSize: "13px",
                  marginBottom: "16px",
                }}
              >
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Department")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.department || "General Medicine"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Specialist")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.specialist || viewingDoctor.qualification || "-"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Designation")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.designation || "-"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Qualification")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.qualification || "-"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Gender")}: </span>
                  <strong
                    className="doctor-detail-value"
                    style={{ textTransform: "capitalize" }}
                  >
                    {viewingDoctor.gender || "-"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Date of Birth")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.dob || "-"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Blood Group")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.bloodGroup || "-"}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Status")}: </span>
                  <strong
                    style={{
                      color: viewingDoctor.status ? "#10b981" : "#ef4444",
                    }}
                  >
                    {viewingDoctor.status ? t("Active") : t("Inactive")}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">
                    {t("OPD Consultation Charge")}:{" "}
                  </span>
                  <strong style={{ color: "#10b981" }}>
                    {viewingDoctor.opdCharge} ETB
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">
                    {t("Appointment Charge")}:{" "}
                  </span>
                  <strong style={{ color: "#38bdf8" }}>
                    {viewingDoctor.appointmentCharge} ETB
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Slot Duration")}: </span>
                  <strong className="doctor-detail-value">
                    {viewingDoctor.slotMinutes} {t("minutes")}
                  </strong>
                </div>
                <div className="doctor-detail-box">
                  <span className="doctor-detail-label">{t("Account ID")}: </span>
                  <code style={{ fontSize: "11px", wordBreak: "break-all" }}>
                    {viewingDoctor.id}
                  </code>
                </div>
              </div>

              {(viewingDoctor.address1 || viewingDoctor.city) && (
                <div
                  className="doctor-detail-box"
                  style={{ marginBottom: "12px", fontSize: "13px" }}
                >
                  <span className="doctor-detail-label">{t("Address")}: </span>
                  <span className="doctor-detail-value">
                    {[
                      viewingDoctor.address1,
                      viewingDoctor.address2,
                      viewingDoctor.city,
                      viewingDoctor.zip,
                    ]
                      .filter(Boolean)
                      .join(", ")}
                  </span>
                </div>
              )}

              {viewingDoctor.description && (
                <div
                  className="doctor-detail-box"
                  style={{ marginBottom: "16px", fontSize: "13px" }}
                >
                  <div
                    className="doctor-detail-label"
                    style={{ marginBottom: "4px" }}
                  >
                    {t("Description / Bio")}:
                  </div>
                  <div
                    className="doctor-detail-value"
                    style={{ lineHeight: "1.5" }}
                  >
                    {viewingDoctor.description}
                  </div>
                </div>
              )}

              <div style={{ display: "flex", justifyContent: "flex-end" }}>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setViewingDoctor(null)}
                >
                  {t("Close")}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

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
              onSubmit={handleCreateDepartment}
              className="modal-body-custom"
            >
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Title")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder={t("Department Title")}
                  className="form-input-custom"
                  value={deptTitle}
                  onChange={(e) => setDeptTitle(e.target.value)}
                />
              </div>

              <div
                className="form-group-custom"
                style={{ marginBottom: "20px" }}
              >
                <label className="form-label-custom">
                  {t("Description")}:
                </label>
                <input
                  placeholder={t("Department Description")}
                  className="form-input-custom"
                  value={deptDesc}
                  onChange={(e) => setDeptDesc(e.target.value)}
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? t("Saving...") : t("Save")}
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

      {/* MODAL: Add Doctor Holiday */}
      {showAddHoliday && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "500px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("Add Doctor Holiday")}</h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowAddHoliday(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreateHoliday} className="modal-body-custom">
              {apiErrorBanner && (
                <div
                  style={{
                    background: "rgba(239, 68, 68, 0.15)",
                    border: "1px solid #ef4444",
                    color: "#f87171",
                    padding: "8px 12px",
                    borderRadius: "6px",
                    marginBottom: "14px",
                    fontSize: "13px",
                  }}
                >
                  {apiErrorBanner}
                </div>
              )}
              <div style={{ marginBottom: "16px" }}>
                <label className="form-label-custom">
                  {t("Doctor")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select
                  required
                  className="form-select-custom"
                  value={holidayDoc}
                  onChange={(e) => setHolidayDoc(e.target.value)}
                >
                  <option value="">{t("Select Doctor")}</option>
                  {doctors.map((doc) => (
                    <option key={doc.id} value={doc.id}>
                      {doc.name} ({doc.department || "General"})
                    </option>
                  ))}
                </select>
              </div>

              <div style={{ marginBottom: "16px" }}>
                <label className="form-label-custom">
                  {t("Date")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  type="date"
                  required
                  className="form-input-custom"
                  value={holidayDate}
                  onChange={(e) => setHolidayDate(e.target.value)}
                />
              </div>

              <div style={{ marginBottom: "20px" }}>
                <label className="form-label-custom">
                  {t("Reason")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder={t("e.g. Annual Leave, Medical Conference")}
                  className="form-input-custom"
                  value={holidayReason}
                  onChange={(e) => setHolidayReason(e.target.value)}
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? t("Saving...") : t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowAddHoliday(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: Set / Edit Doctor OPD Charge */}
      {showOpdModal && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "480px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("Set Doctor OPD Charge")}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowOpdModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveOpdCharge} className="modal-body-custom">
              <div style={{ marginBottom: "16px" }}>
                <label className="form-label-custom">
                  {t("Doctor")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select
                  required
                  className="form-select-custom"
                  value={opdModalDoctorId}
                  onChange={(e) => setOpdModalDoctorId(e.target.value)}
                >
                  <option value="">{t("Select Doctor")}</option>
                  {doctors.map((doc) => (
                    <option key={doc.id} value={doc.id}>
                      {doc.name} ({doc.department || "General"})
                    </option>
                  ))}
                </select>
              </div>

              <div style={{ marginBottom: "20px" }}>
                <label className="form-label-custom">
                  {t("Standard OPD Charge (ETB)")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  required
                  placeholder="350.00"
                  className="form-input-custom"
                  value={opdModalAmount}
                  onChange={(e) => setOpdModalAmount(e.target.value)}
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? t("Saving...") : t("Save Charge")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowOpdModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: Delete Confirmation */}
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
                    {deleteTarget.type === "doctor"
                      ? t(
                          "This doctor cannot be deleted if active clinical or payroll records reference them.",
                        )
                      : t("This action cannot be undone.")}
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
                  disabled={isSubmitting}
                  onClick={confirmDelete}
                >
                  {isSubmitting ? t("Deleting...") : t("Delete")}
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
