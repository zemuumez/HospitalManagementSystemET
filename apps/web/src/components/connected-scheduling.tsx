"use client";
import { useCallback, useEffect, useState } from "react";
import { Plus, X } from "lucide-react";
import { api, type Patient } from "@/lib/api";
import { Modal } from "./modal";
import { useIdentity } from "./workspace";
import { useLanguage } from "./language";

type Staff = {
  id: string;
  name: string;
  email: string;
  role: string;
  active: boolean;
};
type Hours = { weekday: number; startMinute: number; endMinute: number };
type Doctor = {
  id: string;
  name: string;
  department: string;
  slotMinutes: number;
  version: number;
  hours: Hours[];
};
type Appointment = {
  id: string;
  patientId: string;
  doctorId: string;
  patientName: string;
  doctorName: string;
  startsAt: string;
  endsAt: string;
  problem: string;
  status: string;
  version: number;
};
type LinkedPatient = Patient & { userId: string; clinicianId: string };
async function staffRequest<T>(query = "", init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/staff${query}`, {
    ...init,
    headers: { "Content-Type": "application/json" },
    cache: "no-store",
  });
  const body = await response.json();
  if (!response.ok) throw new Error(body.error || "Unable to load users");
  return body;
}
function errorText(error: unknown) {
  return error instanceof Error
    ? error.message
    : "Unable to complete the request";
}
function useResource<T>(loader: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(true),
    [revision, setRevision] = useState(0);
  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    loader()
      .then((value) => {
        if (active) setData(value);
      })
      .catch((e) => {
        if (active) {
          setData(null);
          setError(errorText(e));
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [loader, revision]);
  return { data, error, loading, reload: () => setRevision((v) => v + 1) };
}
function Table({
  head,
  children,
}: {
  head: string[];
  children: React.ReactNode;
}) {
  const { t } = useLanguage();
  return (
    <div className="legacy-table-wrap">
      <table className="legacy-table">
        <thead>
          <tr>
            {head.map((h) => (
              <th key={h}>{t(h)}</th>
            ))}
          </tr>
        </thead>
        <tbody>{children}</tbody>
      </table>
    </div>
  );
}
function Status({
  loading,
  error,
  empty,
}: {
  loading: boolean;
  error: string;
  empty: boolean;
}) {
  const { t } = useLanguage();
  return error ? (
    <p className="error" role="alert">
      {t(error)}
    </p>
  ) : loading ? (
    <p role="status" className="p-8">
      {t("Loading…")}
    </p>
  ) : empty ? (
    <p className="legacy-card p-8">{t("No records found")}</p>
  ) : null;
}
function Editor({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  const { t } = useLanguage();
  return (
    <Modal titleId="connected-editor-title" onClose={onClose}>
      <div className="p-6">
        <div className="page-heading">
          <h2 id="connected-editor-title">{t(title)}</h2>
          <button type="button" onClick={onClose} aria-label={t("Close")}>
            <X size={20} />
          </button>
        </div>
        {children}
      </div>
    </Modal>
  );
}
function Paging({
  page,
  setPage,
  count,
  loading,
}: {
  page: number;
  setPage: (p: number) => void;
  count: number;
  loading: boolean;
}) {
  const { t } = useLanguage();
  return (
    <div className="flex justify-between items-center mt-5">
      <span>
        {t("Page")} {page}
      </span>
      <div className="flex gap-2">
        <button
          className="secondary"
          disabled={loading || page === 1}
          onClick={() => setPage(page - 1)}
        >
          {t("Previous")}
        </button>
        <button
          className="secondary"
          disabled={loading || count < 25}
          onClick={() => setPage(page + 1)}
        >
          {t("Next")}
        </button>
      </div>
    </div>
  );
}
function Input({
  label,
  ...props
}: { label: string } & React.InputHTMLAttributes<HTMLInputElement>) {
  const { t } = useLanguage();
  return (
    <label className="block">
      <span className="label">{t(label)}</span>
      <input className="field" {...props} />
    </label>
  );
}

export function ConnectedUsers() {
  const { t } = useLanguage(),
    identity = useIdentity();
  const [page, setPage] = useState(1),
    [search, setSearch] = useState(""),
    [query, setQuery] = useState(""),
    [open, setOpen] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const load = useCallback(
    () =>
      staffRequest<{ users: Staff[] }>(
        `?page=${page}&search=${encodeURIComponent(query)}`,
      ),
    [page, query],
  );
  const resource = useResource(load);
  async function create(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await staffRequest("", {
        method: "POST",
        body: JSON.stringify(Object.fromEntries(new FormData(e.currentTarget))),
      });
      setOpen(false);
      resource.reload();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  async function toggle(user: Staff) {
    setBusy(true);
    setError("");
    try {
      await staffRequest("", {
        method: "PATCH",
        body: JSON.stringify({ id: user.id, active: !user.active }),
      });
      resource.reload();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <div className="page-heading">
        <h1>{t("Users")}</h1>
        {identity?.permissions.includes("staff.manage") && (
          <button
            className="primary"
            onClick={() => {
              setError("");
              setOpen(true);
            }}
          >
            <Plus size={16} />
            {t("New User")}
          </button>
        )}
      </div>
      <form
        className="flex gap-3 mb-6"
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1);
          setQuery(search);
        }}
      >
        <input
          className="field max-w-sm"
          placeholder={t("Search")}
          aria-label={t("Search users")}
          maxLength={80}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <button className="secondary">{t("Search")}</button>
      </form>
      {!open && error && (
        <p className="error" role="alert">
          {t(error)}
        </p>
      )}
      <Status {...resource} empty={!resource.data?.users.length} />
      {!resource.loading && resource.data && (
        <Table head={["Name", "Email", "Role", "Status", "Action"]}>
          {resource.data.users.map((u) => (
            <tr key={u.id}>
              <td>{u.name}</td>
              <td>{u.email}</td>
              <td>{t(u.role.replaceAll("_", " "))}</td>
              <td>{t(u.active ? "Active" : "Inactive")}</td>
              <td>
                <button
                  className="secondary"
                  disabled={busy || u.id === identity?.user.id}
                  onClick={() => toggle(u)}
                >
                  {t(u.active ? "Deactivate" : "Activate")}
                </button>
              </td>
            </tr>
          ))}
        </Table>
      )}
      <Paging
        page={page}
        setPage={setPage}
        count={resource.data?.users.length || 0}
        loading={resource.loading}
      />
      {open && (
        <Editor
          title="New User"
          onClose={() => {
            if (!busy) setOpen(false);
          }}
        >
          <form onSubmit={create} className="space-y-4">
            {error && (
              <p className="error" role="alert">
                {t(error)}
              </p>
            )}
            <Input label="Name" name="name" required maxLength={120} />
            <Input
              label="Email"
              name="email"
              type="email"
              required
              maxLength={254}
            />
            <label className="block">
              <span className="label">{t("Role")}</span>
              <select
                aria-label={t("Role")}
                name="role"
                className="field"
                required
                defaultValue=""
              >
                <option value="">{t("Select Role")}</option>
                {[
                  "admin",
                  "doctor",
                  "patient",
                  "nurse",
                  "receptionist",
                  "pharmacist",
                  "accountant",
                  "case_manager",
                  "lab_technician",
                ].map((r) => (
                  <option key={r} value={r}>
                    {t(r.replaceAll("_", " "))}
                  </option>
                ))}
              </select>
            </label>
            <Input
              label="Password"
              name="password"
              type="password"
              required
              minLength={12}
              maxLength={128}
              autoComplete="new-password"
            />
            <p className="text-sm text-muted">
              {t(
                "Use at least 12 characters. The user can reset their password from the sign-in page.",
              )}
            </p>
            <button className="primary" disabled={busy}>
              {t(busy ? "Saving…" : "Save")}
            </button>
          </form>
        </Editor>
      )}
    </>
  );
}

export function ConnectedSchedules() {
  const { t } = useLanguage(),
    identity = useIdentity();
  const load = useCallback(() => api<{ doctors: Doctor[] }>("doctors"), []),
    resource = useResource(load);
  const [editing, setEditing] = useState<Doctor | null>(null),
    [staff, setStaff] = useState<Staff[]>([]),
    [search, setSearch] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  async function findDoctors() {
    setError("");
    try {
      const r = await staffRequest<{ users: Staff[] }>(
        `?search=${encodeURIComponent(search)}`,
      );
      setStaff(r.users.filter((u) => u.role === "doctor" && u.active));
    } catch (e) {
      setError(errorText(e));
    }
  }
  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (!editing) return;
    setBusy(true);
    setError("");
    try {
      await api("doctors", { method: "POST", body: JSON.stringify(editing) });
      setEditing(null);
      resource.reload();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  function changeHours(index: number, patch: Partial<Hours>) {
    setEditing((d) =>
      d
        ? {
            ...d,
            hours: d.hours.map((h, i) =>
              i === index ? { ...h, ...patch } : h,
            ),
          }
        : null,
    );
  }
  const timeValue = (m: number) =>
    `${String(Math.floor(m / 60)).padStart(2, "0")}:${String(m % 60).padStart(2, "0")}`;
  const minutes = (s: string) => {
    const [h, m] = s.split(":").map(Number);
    return h * 60 + m;
  };
  return (
    <>
      <div className="page-heading">
        <h1>{t("Schedules")}</h1>
        {identity?.user.role === "admin" && (
          <button
            className="primary"
            onClick={() => {
              setError("");
              setStaff([]);
              setEditing({
                id: "",
                name: "",
                department: "",
                slotMinutes: 30,
                version: 0,
                hours: [],
              });
            }}
          >
            <Plus size={16} />
            {t("New Schedule")}
          </button>
        )}
      </div>
      <p className="mb-5 text-sm text-muted">
        {t(
          "Times use Africa/Addis_Ababa (EAT). Add separate periods for morning and afternoon breaks.",
        )}
      </p>
      <Status {...resource} empty={!resource.data?.doctors.length} />
      {resource.data && !resource.loading && (
        <Table head={["Doctor", "Department", "Per Patient Time", "Action"]}>
          {resource.data.doctors.map((d) => (
            <tr key={d.id}>
              <td>{d.name}</td>
              <td>{d.department}</td>
              <td>
                {d.slotMinutes} {t("Minutes")}
              </td>
              <td>
                {(identity?.user.role === "admin" ||
                  identity?.user.id === d.id) && (
                  <button
                    className="secondary"
                    onClick={() => {
                      setError("");
                      setEditing(structuredClone(d));
                    }}
                  >
                    {t("Edit")}
                  </button>
                )}
              </td>
            </tr>
          ))}
        </Table>
      )}
      {editing && (
        <Editor
          title="Doctor Schedule"
          onClose={() => {
            if (!busy) setEditing(null);
          }}
        >
          <form onSubmit={save} className="space-y-4">
            {error && (
              <p className="error" role="alert">
                {t(error)}
              </p>
            )}
            {editing.version === 0 ? (
              <>
                <label className="block">
                  <span className="label">{t("Search doctor accounts")}</span>
                  <div className="flex gap-2">
                    <input
                      className="field"
                      aria-label={t("Search doctor accounts")}
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      maxLength={80}
                    />
                    <button
                      type="button"
                      className="secondary"
                      onClick={findDoctors}
                    >
                      {t("Search")}
                    </button>
                  </div>
                </label>
                <select
                  required
                  aria-label={t("Doctor")}
                  className="field"
                  value={editing.id}
                  onChange={(e) =>
                    setEditing({ ...editing, id: e.target.value })
                  }
                >
                  <option value="">{t("Select Doctor")}</option>
                  {staff.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.name} — {u.email}
                    </option>
                  ))}
                </select>
              </>
            ) : (
              <p>{editing.name}</p>
            )}
            <Input
              label="Department"
              required
              maxLength={100}
              value={editing.department}
              onChange={(e) =>
                setEditing({ ...editing, department: e.target.value })
              }
            />
            <Input
              label="Per Patient Time (minutes)"
              type="number"
              min={5}
              max={120}
              required
              value={editing.slotMinutes}
              onChange={(e) =>
                setEditing({ ...editing, slotMinutes: Number(e.target.value) })
              }
            />
            {editing.hours.map((h, i) => (
              <div key={i} className="flex flex-wrap gap-2 items-center">
                <select
                  aria-label={t("Day")}
                  className="field !w-28"
                  value={h.weekday}
                  onChange={(e) =>
                    changeHours(i, { weekday: Number(e.target.value) })
                  }
                >
                  {[
                    "Sunday",
                    "Monday",
                    "Tuesday",
                    "Wednesday",
                    "Thursday",
                    "Friday",
                    "Saturday",
                  ].map((day, n) => (
                    <option key={day} value={n}>
                      {t(day)}
                    </option>
                  ))}
                </select>
                <input
                  type="time"
                  required
                  aria-label={t("From")}
                  className="field !w-28"
                  value={timeValue(h.startMinute)}
                  onChange={(e) =>
                    changeHours(i, { startMinute: minutes(e.target.value) })
                  }
                />
                <input
                  type="time"
                  required
                  aria-label={t("To")}
                  className="field !w-28"
                  value={timeValue(h.endMinute)}
                  onChange={(e) =>
                    changeHours(i, { endMinute: minutes(e.target.value) })
                  }
                />
                <button
                  type="button"
                  className="secondary"
                  aria-label={t("Remove period")}
                  onClick={() =>
                    setEditing({
                      ...editing,
                      hours: editing.hours.filter((_, n) => n !== i),
                    })
                  }
                >
                  <X size={14} />
                </button>
              </div>
            ))}
            <button
              className="secondary"
              type="button"
              disabled={editing.hours.length >= 21}
              onClick={() =>
                setEditing({
                  ...editing,
                  hours: [
                    ...editing.hours,
                    { weekday: 1, startMinute: 540, endMinute: 720 },
                  ],
                })
              }
            >
              {t("Add period")}
            </button>
            <p className="text-sm text-muted">
              {t(
                "Schedule changes require future appointments to be cancelled first.",
              )}
            </p>
            <button className="primary" disabled={busy}>
              {t(busy ? "Saving…" : "Save")}
            </button>
          </form>
        </Editor>
      )}
    </>
  );
}

export function ConnectedAppointments() {
  const { t } = useLanguage(),
    identity = useIdentity();
  const [page, setPage] = useState(1),
    [open, setOpen] = useState(false),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [patientSearch, setPatientSearch] = useState(""),
    [patients, setPatients] = useState<Patient[]>([]),
    [patientId, setPatientId] = useState(""),
    [doctorId, setDoctorId] = useState(""),
    [date, setDate] = useState(""),
    [startsAt, setStartsAt] = useState(""),
    [problem, setProblem] = useState(""),
    [notifySms, setNotifySms] = useState(false),
    [requestKey, setRequestKey] = useState("");
  const load = useCallback(
      () => api<{ appointments: Appointment[] }>(`appointments?page=${page}`),
      [page],
    ),
    resource = useResource(load);
  const doctorLoad = useCallback(
      () => api<{ doctors: Doctor[] }>("doctors"),
      [],
    ),
    doctors = useResource(doctorLoad);
  const slotsLoad = useCallback(
      () =>
        doctorId && date
          ? api<{ slots: string[] }>(
              `slots?doctorId=${encodeURIComponent(doctorId)}&date=${date}`,
            )
          : Promise.resolve({ slots: [] }),
      [doctorId, date],
    ),
    slots = useResource(slotsLoad);
  useEffect(() => {
    setStartsAt("");
  }, [doctorId, date]);
  // A changed request gets a new key; a retry of the same request keeps its key.
  useEffect(() => {
    setRequestKey(crypto.randomUUID());
  }, [patientId, doctorId, startsAt, problem, notifySms, open]);
  async function findPatients() {
    setError("");
    try {
      const r = await api<{ patients: Patient[] }>(
        `patients?search=${encodeURIComponent(patientSearch)}`,
      );
      setPatients(r.patients);
    } catch (e) {
      setError(errorText(e));
    }
  }
  async function book(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api("appointments", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": requestKey,
        },
        body: JSON.stringify({
          patientId,
          doctorId,
          startsAt,
          problem,
          notifySms,
        }),
      });
      setOpen(false);
      setProblem("");
      setPatientId("");
      setStartsAt("");
      resource.reload();
      slots.reload();
    } catch (e) {
      setError(errorText(e));
      slots.reload();
    } finally {
      setBusy(false);
    }
  }
  async function change(row: Appointment, status: string) {
    setBusy(true);
    setError("");
    try {
      await api(`appointments/${row.id}`, {
        method: "PATCH",
        body: JSON.stringify({ status, version: row.version }),
      });
      resource.reload();
      slots.reload();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  function actions(row: Appointment) {
    const now = Date.now(),
      start = Date.parse(row.startsAt);
    if (identity?.user.role === "patient")
      return row.status === "booked" && start > now ? ["cancelled"] : [];
    if (!["booked", "arrived"].includes(row.status)) return [];
    return row.status === "arrived"
      ? [...(start <= now ? ["completed"] : []), "cancelled"]
      : [
          ...(start <= now + 86400000 ? ["arrived"] : []),
          ...(start <= now ? ["no_show"] : []),
          "cancelled",
        ];
  }
  return (
    <>
      <div className="page-heading">
        <h1>{t("Appointments")}</h1>
        {identity?.permissions.includes("appointments.book") && (
          <button
            className="primary"
            onClick={() => {
              setError("");
              setOpen(true);
            }}
          >
            <Plus size={16} />
            {t("New Appointment")}
          </button>
        )}
      </div>
      {!open && error && (
        <p className="error" role="alert">
          {t(error)}
        </p>
      )}
      <Status {...resource} empty={!resource.data?.appointments.length} />
      {resource.data && !resource.loading && (
        <Table head={["Patient", "Doctor", "Date", "Status", "Action"]}>
          {resource.data.appointments.map((row) => (
            <tr key={row.id}>
              <td>{row.patientName}</td>
              <td>{row.doctorName}</td>
              <td>
                {new Date(row.startsAt).toLocaleString("en-GB", {
                  timeZone: "Africa/Addis_Ababa",
                })}{" "}
                EAT
              </td>
              <td>
                <span className="badge">
                  {t(row.status.replaceAll("_", " "))}
                </span>
              </td>
              <td>
                <div className="flex flex-wrap gap-2">
                  {actions(row).map((status) => (
                    <button
                      key={status}
                      disabled={busy}
                      className="secondary"
                      onClick={() => change(row, status)}
                    >
                      {t(
                        {
                          arrived: "Mark arrived",
                          completed: "Complete",
                          cancelled: "Cancel",
                          no_show: "No show",
                        }[status] || status,
                      )}
                    </button>
                  ))}
                </div>
              </td>
            </tr>
          ))}
        </Table>
      )}
      <Paging
        page={page}
        setPage={setPage}
        count={resource.data?.appointments.length || 0}
        loading={resource.loading}
      />
      {open && (
        <Editor
          title="New Appointment"
          onClose={() => {
            if (!busy) setOpen(false);
          }}
        >
          <form onSubmit={book} className="space-y-4">
            {error && (
              <p className="error" role="alert">
                {t(error)}
              </p>
            )}
            <label className="block">
              <span className="label">{t("Search patients")}</span>
              <div className="flex gap-2">
                <input
                  className="field"
                  aria-label={t("Search patients")}
                  value={patientSearch}
                  onChange={(e) => setPatientSearch(e.target.value)}
                  maxLength={80}
                />
                <button
                  className="secondary"
                  type="button"
                  onClick={findPatients}
                >
                  {t("Search")}
                </button>
              </div>
            </label>
            <select
              className="field"
              required
              aria-label={t("Patient")}
              value={patientId}
              onChange={(e) => setPatientId(e.target.value)}
            >
              <option value="">{t("Select Patient")}</option>
              {patients.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.givenName} {p.familyName} — {p.mrn}
                </option>
              ))}
            </select>
            {doctors.error && <p className="error">{t(doctors.error)}</p>}
            <select
              className="field"
              required
              aria-label={t("Doctor")}
              value={doctorId}
              onChange={(e) => setDoctorId(e.target.value)}
            >
              <option value="">{t("Select Doctor")}</option>
              {doctors.data?.doctors.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name} — {d.department}
                </option>
              ))}
            </select>
            <Input
              label="Date"
              type="date"
              required
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
            {slots.error && <p className="error">{t(slots.error)}</p>}
            <label className="block">
              <span className="label">{t("Available slots")} (EAT)</span>
              <select
                className="field"
                required
                disabled={slots.loading}
                aria-label={t("Available slots (EAT)")}
                value={startsAt}
                onChange={(e) => setStartsAt(e.target.value)}
              >
                <option value="">
                  {t(slots.loading ? "Loading…" : "Select time")}
                </option>
                {slots.data?.slots.map((slot) => (
                  <option key={slot} value={slot}>
                    {new Date(slot).toLocaleTimeString("en-GB", {
                      timeZone: "Africa/Addis_Ababa",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </option>
                ))}
              </select>
            </label>
            {doctorId &&
              date &&
              !slots.loading &&
              !slots.data?.slots.length && <p>{t("No available slots")}</p>}
            <label className="block">
              <span className="label">{t("Problem")}</span>
              <textarea
                className="field"
                maxLength={2000}
                value={problem}
                onChange={(e) => setProblem(e.target.value)}
              />
            </label>
            <label className="flex gap-2">
              <input
                type="checkbox"
                checked={notifySms}
                onChange={(e) => setNotifySms(e.target.checked)}
              />
              {t("Send SMS confirmation")}
            </label>
            <button
              className="primary"
              disabled={busy || slots.loading || !startsAt}
            >
              {t(busy ? "Saving…" : "Save")}
            </button>
          </form>
        </Editor>
      )}
    </>
  );
}

export function PatientAccessEditor({
  patient,
  onClose,
  onSaved,
}: {
  patient: LinkedPatient;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useLanguage();
  const [search, setSearch] = useState(""),
    [users, setUsers] = useState<Staff[]>([]),
    [userId, setUserId] = useState(patient.userId),
    [clinicianId, setClinicianId] = useState(patient.clinicianId),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function find() {
    try {
      setError("");
      setUsers(
        (
          await staffRequest<{ users: Staff[] }>(
            `?search=${encodeURIComponent(search)}`,
          )
        ).users,
      );
    } catch (e) {
      setError(errorText(e));
    }
  }
  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api(`patients/${patient.id}`, {
        method: "PATCH",
        body: JSON.stringify({ userId, clinicianId }),
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
      title="Patient access"
      onClose={() => {
        if (!busy) onClose();
      }}
    >
      <form onSubmit={save} className="space-y-4">
        {error && <p className="error">{t(error)}</p>}
        <p>
          {patient.givenName} {patient.familyName} — {patient.mrn}
        </p>
        <div className="flex gap-2">
          <input
            className="field"
            aria-label={t("Search users")}
            maxLength={80}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button type="button" className="secondary" onClick={find}>
            {t("Search")}
          </button>
        </div>
        <label className="block">
          <span className="label">{t("Patient portal account")}</span>
          <select
            className="field"
            disabled={!!patient.userId}
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
          >
            <option value="">{t("Not linked")}</option>
            {patient.userId && (
              <option value={patient.userId}>
                {t("Existing linked account")}
              </option>
            )}
            {users
              .filter(
                (u) =>
                  u.role === "patient" && u.active && u.id !== patient.userId,
              )
              .map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name} — {u.email}
                </option>
              ))}
          </select>
        </label>
        <label className="block">
          <span className="label">{t("Assigned doctor")}</span>
          <select
            className="field"
            value={clinicianId}
            onChange={(e) => setClinicianId(e.target.value)}
          >
            <option value="">{t("Not assigned")}</option>
            {patient.clinicianId && (
              <option value={patient.clinicianId}>
                {t("Existing assigned doctor")}
              </option>
            )}
            {users
              .filter(
                (u) =>
                  u.role === "doctor" &&
                  u.active &&
                  u.id !== patient.clinicianId,
              )
              .map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name} — {u.email}
                </option>
              ))}
          </select>
        </label>
        <button className="primary" disabled={busy}>
          {t(busy ? "Saving…" : "Save")}
        </button>
      </form>
    </Editor>
  );
}
