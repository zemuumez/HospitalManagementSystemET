"use client";
import { useEffect, useState } from "react";
import { Download, Eye, Pencil, Plus, Search, Trash2, X } from "lucide-react";
import extractedCmsFields from "@/lib/front-cms-fields.json";
import { Modal } from "./modal";
import { useLanguage } from "./language";
import {
  defaultFrontSettings,
  defaultGeneralSettings,
  FRONT_KEY,
  GENERAL_KEY,
} from "@/lib/front-settings";
import { groups, people } from "@/lib/legacy";

export const extraScreens = [
  "front-settings",
  "settings",
  "modules-setting",
  "patient-queue-theme",
  "attendance",
  "manage-attendance",
  "patient-id-card-template",
  "generate-patient-id-card",
];
export function LegacyExtras({ id }: { id: string }) {
  if (id === "front-settings" || id === "settings")
    return <SettingsForm cms={id === "front-settings"} />;
  if (id === "modules-setting") return <ModulesSettings />;
  if (id === "patient-queue-theme") return <QueueTheme />;
  if (id === "attendance" || id === "manage-attendance")
    return <Attendance manage={id === "manage-attendance"} />;
  return <SmartCards templates={id === "patient-id-card-template"} />;
}
type SettingField = {
  key: string;
  label: string;
  type?: string;
  options?: string[];
};
const cmsFields: Record<string, SettingField[]> = {
  Home: [
    { key: "home_page_image", label: "Home Page Image", type: "image" },
    {
      key: "home_page_certified_doctor_image",
      label: "Home Page Certified Doctor Image",
      type: "image",
    },
    {
      key: "home_page_experience",
      label: "Home Page Experience",
      type: "number",
    },
    { key: "home_page_title", label: "Home Page Title" },
    {
      key: "home_page_description",
      label: "Home Page Description",
      type: "textarea",
    },
    { key: "home_page_box_title", label: "Home Page Box Title" },
    {
      key: "home_page_box_description",
      label: "Home Page Box Description",
      type: "textarea",
    },
    {
      key: "home_page_certified_doctor_title",
      label: "Certified Doctor Title",
    },
    {
      key: "home_page_certified_doctor_text",
      label: "Certified Doctor Text",
      type: "textarea",
    },
  ],
  "About Us": [
    { key: "about_title", label: "Title" },
    { key: "about_description", label: "Description", type: "textarea" },
  ],
  Appointment: [
    { key: "appointment_title", label: "Title" },
    { key: "appointment_description", label: "Description", type: "textarea" },
  ],
  "Terms & Conditions": [
    { key: "terms", label: "Terms & Conditions", type: "textarea" },
  ],
  Map: [
    { key: "map_address", label: "Address" },
    { key: "map_url", label: "Map URL", type: "url" },
  ],
};
const generalFields: SettingField[] = [
  { key: "app_name", label: "Application Name" },
  { key: "company_name", label: "Company Name" },
  { key: "hospital_email", label: "Hospital Email", type: "email" },
  { key: "hospital_phone", label: "Hospital Phone", type: "tel" },
  { key: "hospital_start_day", label: "Hospital Start Day", type: "number" },
  { key: "hospital_start_time", label: "Hospital Start Time", type: "time" },
  { key: "hospital_address", label: "Hospital Address" },
  { key: "currency", label: "Currency", options: ["ETB", "USD", "EUR", "KES"] },
  {
    key: "country_code",
    label: "Country Code",
    options: ["+251", "+254", "+1", "+44", "+91"],
  },
  { key: "default_language", label: "Default Language", options: ["en", "am"] },
  { key: "about_us", label: "About Us", type: "textarea" },
  { key: "app_logo", label: "Application Logo", type: "image" },
  { key: "favicon", label: "Favicon", type: "image" },
];
function SettingsForm({ cms }: { cms: boolean }) {
  const { t, setLocale } = useLanguage();
  const [tab, setTab] = useState("Home");
  const [values, setValues] = useState(
    cms ? defaultFrontSettings : defaultGeneralSettings,
  );
  const [notice, setNotice] = useState("");
  const key = cms ? FRONT_KEY : GENERAL_KEY;
  useEffect(() => {
    try {
      const saved = localStorage.getItem(key);
      if (saved) setValues((old) => ({ ...old, ...JSON.parse(saved) }));
    } catch {}
  }, [key]);
  const fields = cms
    ? [
        ...cmsFields[tab],
        ...(
          (extractedCmsFields as Record<string, SettingField[]>)[tab] || []
        ).filter((f) => !cmsFields[tab].some((x) => x.key === f.key)),
      ]
    : generalFields;
  async function imageFile(field: string, file: File | undefined) {
    if (!file) return;
    if (
      !["image/png", "image/jpeg"].includes(file.type) ||
      file.size > 1500000
    ) {
      setNotice("Choose a PNG or JPEG image smaller than 1.5 MB.");
      return;
    }
    const reader = new FileReader();
    reader.onload = () =>
      setValues((old) => ({ ...old, [field]: String(reader.result) }));
    reader.readAsDataURL(file);
  }
  return (
    <section>
      {cms && (
        <div className="detail-tabs cms-tabs">
          {Object.keys(cmsFields).map((name) => (
            <button
              key={name}
              className={tab === name ? "active" : ""}
              onClick={() => {
                setTab(name);
                setNotice("");
              }}
            >
              {t(name)}
            </button>
          ))}
        </div>
      )}
      <form
        className="legacy-card settings-form"
        onSubmit={(e) => {
          e.preventDefault();
          try {
            localStorage.setItem(key, JSON.stringify(values));
            if (!cms) setLocale(values.default_language === "am" ? "am" : "en");
            setNotice(t("Preview saved. No hospital data was changed."));
          } catch {
            setNotice(
              "Browser storage is full. Choose smaller preview images.",
            );
          }
        }}
      >
        {cms && <h2>{t("Front Setting Details")}</h2>}
        {notice && (
          <p className="success mb-5" role="status">
            {t(notice)}
          </p>
        )}
        <div
          className={`legacy-form ${cms && tab === "Home" ? "cms-home-form" : ""}`}
        >
          {fields.map((f) => (
            <label
              key={f.key}
              className={f.type === "textarea" ? "form-span" : ""}
            >
              <span className="label">
                {t(f.label)}: <b className="text-red-500">*</b>
              </span>
              {f.type === "image" ? (
                <>
                  <span className="setting-image-picker">
                    <img
                      src={values[f.key] || "/legacy/front/home.png"}
                      alt={t(f.label)}
                    />
                    <span>
                      <Pencil size={13} />
                    </span>
                    <input
                      aria-label={f.label}
                      type="file"
                      accept="image/png,image/jpeg"
                      onChange={(e) => imageFile(f.key, e.target.files?.[0])}
                    />
                  </span>
                  <small className="block text-muted mt-3">
                    {t("Allowed file types: png, jpg, jpeg.")}
                  </small>
                </>
              ) : f.options ? (
                <select
                  className="field"
                  value={values[f.key] || ""}
                  onChange={(e) =>
                    setValues({ ...values, [f.key]: e.target.value })
                  }
                >
                  {f.options.map((o) => (
                    <option key={o} value={o}>
                      {o === "en" ? t("English") : o === "am" ? "አማርኛ" : o}
                    </option>
                  ))}
                </select>
              ) : f.type === "textarea" ? (
                <textarea
                  className="field"
                  rows={4}
                  value={values[f.key] || ""}
                  onChange={(e) =>
                    setValues({ ...values, [f.key]: e.target.value })
                  }
                />
              ) : (
                <input
                  className="field"
                  type={f.type || "text"}
                  min={f.type === "number" ? 0 : undefined}
                  value={values[f.key] || ""}
                  onChange={(e) =>
                    setValues({ ...values, [f.key]: e.target.value })
                  }
                />
              )}
            </label>
          ))}
        </div>
        <div className="flex justify-end mt-7">
          <button className="primary">{t("Save")}</button>
        </div>
      </form>
    </section>
  );
}
function ModulesSettings() {
  const { t } = useLanguage();
  const [disabled, setDisabled] = useState<string[]>([]);
  const [notice, setNotice] = useState("");
  useEffect(() => {
    try {
      setDisabled(
        JSON.parse(localStorage.getItem("hms-disabled-modules") || "[]"),
      );
    } catch {}
  }, []);
  return (
    <section className="legacy-card">
      <h2>{t("Modules Setting")}</h2>
      <p className="form-preview-note">
        {t("Preview visibility settings. API permissions are unchanged.")}
      </p>
      <div className="module-toggle-grid">
        {groups
          .filter((g) => g !== "Settings")
          .map((g) => (
            <label key={g}>
              <span>{t(g)}</span>
              <input
                type="checkbox"
                checked={!disabled.includes(g)}
                onChange={(e) =>
                  setDisabled(
                    e.target.checked
                      ? disabled.filter((x) => x !== g)
                      : [...disabled, g],
                  )
                }
              />
            </label>
          ))}
      </div>
      <button
        className="primary mt-6"
        onClick={() => {
          localStorage.setItem(
            "hms-disabled-modules",
            JSON.stringify(disabled),
          );
          window.dispatchEvent(new Event("hms-modules-updated"));
          setNotice(t("Preview saved. No hospital data was changed."));
        }}
      >
        {t("Save")}
      </button>
      {notice && (
        <p className="success mt-4" role="status">
          {t(notice)}
        </p>
      )}
    </section>
  );
}
function QueueTheme() {
  const { t } = useLanguage();
  const [color, setColor] = useState("#6571ff");
  const [message, setMessage] = useState("Please wait for your number");
  const [saved, setSaved] = useState(false);
  useEffect(() => {
    try {
      const saved = JSON.parse(
        localStorage.getItem("hms-queue-theme") || "null",
      );
      if (saved) {
        setColor(saved.color);
        setMessage(saved.message);
      }
    } catch {}
  }, []);
  return (
    <section className="legacy-card">
      <h2>{t("Patient Queue Theme")}</h2>
      <div className="legacy-form mt-6">
        <label>
          <span className="label">{t("Header Color")}</span>
          <input
            className="field"
            type="color"
            value={color}
            onChange={(e) => setColor(e.target.value)}
          />
        </label>
        <label>
          <span className="label">{t("Message")}</span>
          <input
            className="field"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
          />
        </label>
      </div>
      <div className="queue-preview">
        <header style={{ background: color }}>
          {t("ULSHMS · Patient Queue")}
        </header>
        <h3>{t("Now serving")}</h3>
        <strong>024</strong>
        <p>Dr. Avery Reed · Room 2</p>
        <footer>{t(message)}</footer>
      </div>
      <button
        className="primary"
        onClick={() => {
          localStorage.setItem(
            "hms-queue-theme",
            JSON.stringify({ color, message }),
          );
          setSaved(true);
        }}
      >
        {t("Save")}
      </button>
      {saved && (
        <p className="success mt-4" role="status">
          {t("Preview saved. No hospital data was changed.")}
        </p>
      )}
    </section>
  );
}

type AttendanceRow = {
  id: string;
  staff: string;
  date: string;
  shift: string;
  status: string;
  checkIn: string;
  checkOut: string;
  breakMinutes: number;
};
const attendanceSeed: AttendanceRow[] = people.map((staff, i) => ({
  id: `A-${i}`,
  staff,
  date: `2026-10-0${i + 1}`,
  shift: "Day Shift",
  status: i === 1 ? "Late" : "Present",
  checkIn: i === 1 ? "08:30" : "08:00",
  checkOut: "17:00",
  breakMinutes: 60,
}));
const minutes = (time: string) => {
  const [h, m] = time.split(":").map(Number);
  return h * 60 + m;
};
const duration = (m: number) =>
  `${String(Math.floor(Math.max(0, m) / 60)).padStart(2, "0")}:${String(Math.max(0, m) % 60).padStart(2, "0")}`;
function Attendance({ manage }: { manage: boolean }) {
  const { t } = useLanguage();
  const [rows, setRows] = useState(attendanceSeed);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState("All");
  const [editing, setEditing] = useState<AttendanceRow | null>(null);
  const [notice, setNotice] = useState("");
  const [loaded, setLoaded] = useState(false);
  useEffect(() => {
    try {
      const saved = sessionStorage.getItem("hms-attendance-preview");
      if (saved) setRows(JSON.parse(saved));
    } catch {}
    setLoaded(true);
  }, []);
  useEffect(() => {
    if (loaded)
      sessionStorage.setItem("hms-attendance-preview", JSON.stringify(rows));
  }, [rows, loaded]);
  const shown = rows.filter(
    (r) =>
      (manage || r.staff === people[0]) &&
      r.staff.toLowerCase().includes(search.toLowerCase()) &&
      (filter === "All" || filter === r.status),
  );
  const columns = [
    "Staff",
    "Date",
    "Shift",
    "Status",
    "Check In",
    "Check Out",
    "Late",
    "Early Out",
    "Worked",
    "Over Time",
    "Break Time",
  ];
  function clock() {
    const now = new Date();
    const date = now.toLocaleDateString("en-CA");
    const time = now.toTimeString().slice(0, 5);
    const open = rows.find(
      (r) => r.staff === people[0] && r.date === date && !r.checkOut,
    );
    if (open)
      setRows(
        rows.map((r) => (r.id === open.id ? { ...r, checkOut: time } : r)),
      );
    else
      setRows([
        {
          id: crypto.randomUUID(),
          staff: people[0],
          date,
          shift: "Day Shift",
          status: "Present",
          checkIn: time,
          checkOut: "",
          breakMinutes: 0,
        },
        ...rows,
      ]);
    setNotice(t("Preview saved. No hospital data was changed."));
  }
  return (
    <section>
      <div className="page-heading">
        <div>
          <h1>{t(manage ? "Manage Attendance" : "Attendance")}</h1>
          {manage && (
            <p className="text-muted mt-1">
              {t("Admin can create and edit attendance records manually.")}
            </p>
          )}
        </div>
        {!manage && (
          <button className="primary" onClick={clock}>
            {t("Check In")} / {t("Check Out")}
          </button>
        )}
      </div>
      {notice && (
        <p className="success mb-5" role="status">
          {t(notice)}
        </p>
      )}
      <div className="table-toolbar">
        <div className="table-search">
          <Search size={18} />
          <input
            aria-label={t("Search attendance")}
            placeholder={t("Search")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="flex gap-3">
          <select
            aria-label={t("Attendance status")}
            className="field !w-auto"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          >
            {["All", "Present", "Late", "Absent"].map((v) => (
              <option key={v} value={v}>
                {t(v)}
              </option>
            ))}
          </select>
          {manage && (
            <button
              className="primary"
              onClick={() =>
                setEditing({
                  id: crypto.randomUUID(),
                  staff: people[0],
                  date: "2026-10-04",
                  shift: "Day Shift",
                  status: "Present",
                  checkIn: "08:00",
                  checkOut: "17:00",
                  breakMinutes: 60,
                })
              }
            >
              {t("Add Attendance")}
            </button>
          )}
        </div>
      </div>
      <div className="legacy-table-wrap">
        <table className="legacy-table reference-striped">
          <thead>
            <tr>
              {columns.map((c) => (
                <th key={c}>{t(c)}</th>
              ))}
              {manage && <th>{t("Action")}</th>}
            </tr>
          </thead>
          <tbody>
            {shown.map((r) => {
              const shiftStart = r.shift === "Night Shift" ? 1200 : 480;
              const shiftEnd = r.shift === "Night Shift" ? 1740 : 1020;
              const start = minutes(r.checkIn);
              let end = minutes(r.checkOut);
              if (end < start) end += 1440;
              const worked = r.checkOut ? end - start - r.breakMinutes : 0;
              return (
                <tr key={r.id}>
                  <td>{r.staff}</td>
                  <td>{r.date}</td>
                  <td>{t(r.shift)}</td>
                  <td>
                    <span
                      className={`badge ${r.status === "Late" ? "attendance-late" : ""}`}
                    >
                      {t(r.status)}
                    </span>
                  </td>
                  <td>{r.checkIn || "N/A"}</td>
                  <td>{r.checkOut || "N/A"}</td>
                  <td className="text-red-500">
                    {duration(start - shiftStart)}
                  </td>
                  <td className="text-red-500">
                    {r.checkOut ? duration(shiftEnd - end) : "00:00"}
                  </td>
                  <td>{duration(worked)}</td>
                  <td className="text-emerald-500">{duration(worked - 480)}</td>
                  <td>{duration(r.breakMinutes)}</td>
                  {manage && (
                    <td>
                      <button
                        aria-label={`Edit attendance ${r.staff}`}
                        className="text-brand"
                        onClick={() => setEditing({ ...r })}
                      >
                        <Pencil size={16} />
                      </button>
                    </td>
                  )}
                </tr>
              );
            })}
            {!shown.length && (
              <tr>
                <td colSpan={12}>{t("No matching records found")}</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {editing && (
        <Modal titleId="attendance-title" onClose={() => setEditing(null)}>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              setRows((old) =>
                old.some((r) => r.id === editing.id)
                  ? old.map((r) => (r.id === editing.id ? editing : r))
                  : [editing, ...old],
              );
              setEditing(null);
              setNotice(t("Preview saved. No hospital data was changed."));
            }}
          >
            <div className="modal-heading">
              <h2 id="attendance-title">{t("Add Attendance")}</h2>
              <button
                type="button"
                aria-label={t("Close")}
                onClick={() => setEditing(null)}
              >
                <X />
              </button>
            </div>
            <div className="modal-content legacy-form">
              {(
                [
                  "staff",
                  "date",
                  "shift",
                  "status",
                  "checkIn",
                  "checkOut",
                  "breakMinutes",
                ] as const
              ).map((key, i) => (
                <label key={key}>
                  <span className="label">
                    {t(
                      [
                        "Staff",
                        "Date",
                        "Shift",
                        "Status",
                        "Check In",
                        "Check Out",
                        "Break Time",
                      ][i],
                    )}
                  </span>
                  {["staff", "shift", "status"].includes(key) ? (
                    <select
                      className="field"
                      value={editing[key]}
                      onChange={(e) =>
                        setEditing({ ...editing, [key]: e.target.value })
                      }
                    >
                      {(key === "staff"
                        ? people
                        : key === "shift"
                          ? ["Day Shift", "Night Shift"]
                          : ["Present", "Late", "Absent"]
                      ).map((o) => (
                        <option key={o} value={o}>
                          {t(o)}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      className="field"
                      required={key !== "checkOut"}
                      type={
                        key === "date"
                          ? "date"
                          : key === "breakMinutes"
                            ? "number"
                            : "time"
                      }
                      min={key === "breakMinutes" ? 0 : undefined}
                      value={editing[key]}
                      onChange={(e) =>
                        setEditing({
                          ...editing,
                          [key]:
                            key === "breakMinutes"
                              ? Number(e.target.value)
                              : e.target.value,
                        })
                      }
                    />
                  )}
                </label>
              ))}
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
        </Modal>
      )}
    </section>
  );
}

type CardTemplate = {
  name: string;
  color: string;
  email: boolean;
  phone: boolean;
  dob: boolean;
  blood: boolean;
  address: boolean;
  uniqueId?: boolean;
};
const templateSeed: CardTemplate[] = [
  {
    name: "Standard",
    color: "#f4bf16",
    email: true,
    phone: true,
    dob: true,
    blood: true,
    address: true,
    uniqueId: true,
  },
  {
    name: "VIP",
    color: "#6571ff",
    email: true,
    phone: true,
    dob: false,
    blood: true,
    address: false,
    uniqueId: true,
  },
];
type CardRow = { id: string; patient: number; template: string };
function SmartCards({ templates }: { templates: boolean }) {
  const { t } = useLanguage();
  const [list, setList] = useState<CardTemplate[]>(() => {
    if (typeof window !== "undefined") {
      try {
        const stored = sessionStorage.getItem("hms-card-templates");
        if (stored) return JSON.parse(stored);
      } catch {}
    }
    return templateSeed;
  });
  const [cards, setCards] = useState<CardRow[]>(() => {
    if (typeof window !== "undefined") {
      try {
        const stored = sessionStorage.getItem("hms-smart-cards");
        if (stored) return JSON.parse(stored);
      } catch {}
    }
    return people.map((_, i) => ({
      id: `CARD-${i + 1}`,
      patient: i,
      template: "Standard",
    }));
  });
  const [editing, setEditing] = useState<CardTemplate | null>(null);
  const [originalName, setOriginalName] = useState<string | null>(null);
  const [templateError, setTemplateError] = useState("");
  const [create, setCreate] = useState(false);
  const [selected, setSelected] = useState<CardRow | null>(null);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1),
    [pageSize, setPageSize] = useState(10);
  const [ready, setReady] = useState(false);
  useEffect(() => {
    try {
      setList(
        JSON.parse(sessionStorage.getItem("hms-card-templates") || "null") ||
          templateSeed,
      );
      setCards(
        JSON.parse(sessionStorage.getItem("hms-smart-cards") || "null") ||
          people.map((_, i) => ({
            id: `CARD-${i + 1}`,
            patient: i,
            template: "Standard",
          })),
      );
    } catch {}
    setReady(true);
  }, []);
  useEffect(() => {
    if (ready) {
      sessionStorage.setItem("hms-card-templates", JSON.stringify(list));
      sessionStorage.setItem("hms-smart-cards", JSON.stringify(cards));
    }
  }, [list, cards, ready]);
  const filteredTemplates = list.filter((row) =>
    row.name.toLowerCase().includes(search.toLowerCase()),
  );
  const filteredCards = cards.filter((row) =>
    people[row.patient].toLowerCase().includes(search.toLowerCase()),
  );
  const count = templates ? filteredTemplates.length : filteredCards.length;
  const pages = Math.max(1, Math.ceil(count / pageSize));
  const current = Math.min(page, pages);
  const selectedTemplate =
    list.find((x) => x.name === selected?.template) ||
    list[0] ||
    templateSeed[0];
  return (
    <section className="smart-card-screen">
      {templateError && !editing && (
        <p role="alert" className="error">
          {t(templateError)}
        </p>
      )}
      <div className="page-heading">
        <h1>
          {t(
            templates
              ? "Smart Patient Card Templates"
              : "Generate Patient Smart Cards",
          )}
        </h1>
        <button
          className="primary"
          onClick={() =>
            templates
              ? (setOriginalName(null),
                setTemplateError(""),
                setEditing({ ...templateSeed[0], name: "" }))
              : setCreate(true)
          }
        >
          <Plus size={16} />
          {t(
            templates
              ? "New Patient Smart Card Template"
              : "New Patient Smart Card",
          )}
        </button>
      </div>
      <div className="table-toolbar">
        <div className="table-search">
          <Search size={18} />
          <input
            aria-label={t("Search cards")}
            placeholder={t("Search")}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>
      </div>
      <div className="legacy-table-wrap">
        <table className="legacy-table reference-striped">
          <thead>
            <tr>
              <th>{t(templates ? "Name" : "Patients")}</th>
              <th>{templates ? t("Color") : t("Patient Unique ID")}</th>
              {templates ? (
                [
                  "Email",
                  "Phone",
                  "Date of Birth",
                  "Blood Group",
                  "Address",
                  "Patient Unique ID",
                ].map((label) => <th key={label}>{t(label)}</th>)
              ) : (
                <th>{t("Template Name")}</th>
              )}
              <th>{t("Action")}</th>
            </tr>
          </thead>
          <tbody>
            {templates
              ? filteredTemplates
                  .slice((current - 1) * pageSize, current * pageSize)
                  .map((template) => (
                    <tr key={template.name}>
                      <td>{template.name}</td>
                      <td>
                        <span
                          className="color-swatch"
                          style={{ background: template.color }}
                        />
                      </td>
                      {(
                        [
                          "email",
                          "phone",
                          "dob",
                          "blood",
                          "address",
                          "uniqueId",
                        ] as const
                      ).map((key) => (
                        <td key={key}>
                          <input
                            className="reference-switch"
                            type="checkbox"
                            role="switch"
                            aria-label={`${template.name} ${key}`}
                            checked={template[key] !== false}
                            onChange={(e) => {
                              const next = list.map((row) =>
                                row.name === template.name
                                  ? { ...row, [key]: e.target.checked }
                                  : row,
                              );
                              setList(next);
                              try {
                                sessionStorage.setItem(
                                  "hms-card-templates",
                                  JSON.stringify(next),
                                );
                              } catch {}
                            }}
                          />
                        </td>
                      ))}
                      <td>
                        <button
                          className="text-brand"
                          aria-label={`Edit ${template.name}`}
                          onClick={() => {
                            setOriginalName(template.name);
                            setTemplateError("");
                            setEditing({ ...template });
                          }}
                        >
                          <Pencil size={17} />
                        </button>
                        <button
                          className="text-red-500 ml-3"
                          aria-label={`Delete ${template.name}`}
                          onClick={() => {
                            if (
                              cards.some((c) => c.template === template.name)
                            ) {
                              setTemplateError(
                                "This template is used by patient cards. Remove those cards before deleting the template.",
                              );
                              return;
                            }
                            if (window.confirm(`Delete ${template.name}?`))
                              setList(
                                list.filter(
                                  (row) => row.name !== template.name,
                                ),
                              );
                          }}
                        >
                          <Trash2 size={17} />
                        </button>
                      </td>
                    </tr>
                  ))
              : filteredCards
                  .slice((current - 1) * pageSize, current * pageSize)
                  .map((card) => (
                    <tr key={card.id}>
                      <td>
                        <button
                          className="record-link"
                          onClick={() => setSelected(card)}
                        >
                          <span className="avatar">
                            {people[card.patient]
                              .split(" ")
                              .map((n) => n[0])
                              .join("")}
                          </span>
                          <span>
                            {people[card.patient]}
                            <small>
                              {t("sample")}
                              {card.patient + 1}@example.invalid
                            </small>
                          </span>
                        </button>
                      </td>
                      <td>
                        <span className="badge card-id">
                          DEMO-{String(card.patient + 1).padStart(4, "0")}
                        </span>
                      </td>
                      <td>
                        <span className="badge">{card.template}</span>
                      </td>
                      <td>
                        <div className="row-actions">
                          <button
                            aria-label={`View card ${people[card.patient]}`}
                            onClick={() => setSelected(card)}
                          >
                            <Eye size={18} />
                          </button>
                          <button
                            aria-label={`Download card ${people[card.patient]}`}
                            onClick={() =>
                              downloadCard(
                                card,
                                list.find((t) => t.name === card.template) ||
                                  templateSeed[0],
                              )
                            }
                          >
                            <Download size={18} />
                          </button>
                          <button
                            aria-label={`Delete card ${people[card.patient]}`}
                            onClick={() => {
                              if (
                                window.confirm(t("Delete this patient card?"))
                              )
                                setCards(cards.filter((c) => c.id !== card.id));
                            }}
                          >
                            <Trash2 size={18} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
          </tbody>
        </table>
      </div>
      <div className="reference-pagination">
        <label>
          {t("Show")}
          <select
            className="field"
            aria-label="Rows per page"
            value={pageSize}
            onChange={(e) => {
              setPageSize(Number(e.target.value));
              setPage(1);
            }}
          >
            {[10, 25, 50].map((n) => (
              <option key={n}>{n}</option>
            ))}
          </select>
        </label>
        <span>
          {t("Showing")} {count ? (current - 1) * pageSize + 1 : 0} -{" "}
          {Math.min(current * pageSize, count)} {t("of")} {count} {t("Results")}
        </span>
        <div>
          {Array.from({ length: pages }, (_, i) => (
            <button
              key={i}
              className={current === i + 1 ? "primary" : "secondary"}
              onClick={() => setPage(i + 1)}
            >
              {i + 1}
            </button>
          ))}
        </div>
      </div>
      {editing && (
        <Modal titleId="template-title" onClose={() => setEditing(null)}>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              const name = editing.name.trim();
              if (
                !name ||
                list.some(
                  (x) =>
                    x.name.toLowerCase() === name.toLowerCase() &&
                    x.name !== originalName,
                )
              ) {
                setTemplateError("Template name must be unique.");
                return;
              }
              const updated = { ...editing, name };
              setList((old) =>
                originalName
                  ? old.map((x) => (x.name === originalName ? updated : x))
                  : [...old, updated],
              );
              if (originalName)
                setCards((old) =>
                  old.map((c) =>
                    c.template === originalName ? { ...c, template: name } : c,
                  ),
                );
              setEditing(null);
            }}
          >
            <div className="modal-heading">
              <h2 id="template-title">{t("Template")}</h2>
              <button
                type="button"
                aria-label={t("Close")}
                onClick={() => setEditing(null)}
              >
                <X />
              </button>
            </div>
            <div className="modal-content">
              {templateError && (
                <p className="error" role="alert">
                  {t(templateError)}
                </p>
              )}
              <div className="legacy-form">
                <label>
                  <span className="label">{t("Template Name")}</span>
                  <input
                    className="field"
                    required
                    value={editing.name}
                    onChange={(e) =>
                      setEditing({ ...editing, name: e.target.value })
                    }
                  />
                </label>
                <label>
                  <span className="label">{t("Header Color")}</span>
                  <input
                    className="field"
                    type="color"
                    value={editing.color}
                    onChange={(e) =>
                      setEditing({ ...editing, color: e.target.value })
                    }
                  />
                </label>
                {(["email", "phone", "dob", "blood", "address"] as const).map(
                  (key, i) => (
                    <label className="flex gap-3" key={key}>
                      <input
                        type="checkbox"
                        checked={editing[key]}
                        onChange={(e) =>
                          setEditing({ ...editing, [key]: e.target.checked })
                        }
                      />
                      {t(
                        [
                          "Email",
                          "Phone",
                          "Date of Birth",
                          "Blood Group",
                          "Address",
                        ][i],
                      )}
                    </label>
                  ),
                )}
              </div>
              <SmartCard patient={0} template={editing} />
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
        </Modal>
      )}
      {create && (
        <Modal titleId="card-create-title" onClose={() => setCreate(false)}>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              const f = new FormData(e.currentTarget);
              setCards([
                ...cards,
                {
                  id: crypto.randomUUID(),
                  patient: Number(f.get("patient")),
                  template: String(f.get("template")),
                },
              ]);
              setCreate(false);
            }}
          >
            <div className="modal-heading">
              <h2 id="card-create-title">{t("New Patient Smart Card")}</h2>
            </div>
            <div className="modal-content legacy-form">
              <label>
                <span className="label">{t("Patient")}</span>
                <select className="field" name="patient">
                  {people.map((p, i) => (
                    <option key={p} value={i}>
                      {p}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span className="label">{t("Template")}</span>
                <select className="field" name="template">
                  {list.map((x) => (
                    <option key={x.name}>{x.name}</option>
                  ))}
                </select>
              </label>
            </div>
            <div className="modal-footer">
              <button className="primary">{t("Save")}</button>
              <button
                type="button"
                className="secondary"
                onClick={() => setCreate(false)}
              >
                {t("Cancel")}
              </button>
            </div>
          </form>
        </Modal>
      )}
      {selected && (
        <Modal titleId="smart-card-title" onClose={() => setSelected(null)}>
          <div className="smart-card-modal">
            <button
              className="smart-card-close"
              aria-label={t("Close card")}
              onClick={() => setSelected(null)}
            >
              <X size={19} />
            </button>
            <h2 id="smart-card-title" className="sr-only">
              {t("Patient Smart Cards")}
            </h2>
            <SmartCard patient={selected.patient} template={selectedTemplate} />
            <div className="modal-footer">
              <button
                className="secondary"
                onClick={() => downloadCard(selected, selectedTemplate)}
              >
                <Download size={16} />
                {t("Download")}
              </button>
              <button className="secondary" onClick={() => setSelected(null)}>
                {t("Close")}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </section>
  );
}
function SmartCard({
  patient,
  template,
}: {
  patient: number;
  template: CardTemplate;
}) {
  const { t } = useLanguage();
  return (
    <article className="smart-card">
      <header style={{ background: template.color }}>
        <strong>ULSHMS</strong>
        <span>{t("Addis Ababa, Ethiopia · Hospital Management System")}</span>
      </header>
      <div className="smart-card-details">
        <div className="card-portrait">
          {people[patient]
            .split(" ")
            .map((n) => n[0])
            .join("")}
        </div>
        <dl>
          <dt>{t("Patient Name")}:</dt>
          <dd>{people[patient]}</dd>
          {[
            [template.email, "Email", `sample${patient + 1}@example.invalid`],
            [template.phone, "Phone", `+25190000010${patient}`],
            [template.dob, "Date of Birth", "2000-08-08"],
            [template.blood, "Blood Group", "AB+"],
            [template.address, "Address", "Addis Ababa, Ethiopia"],
          ].map(
            ([show, label, value]) =>
              show && (
                <div className="card-field" key={String(label)}>
                  <dt>{t(String(label))}:</dt>
                  <dd>{value}</dd>
                </div>
              ),
          )}
        </dl>
        <div className="card-qr">
          <img
            src={`/legacy/front/card-${patient + 1}.svg`}
            alt={`Demo QR identifier ${patient + 1}`}
          />
          {template.uniqueId !== false && (
            <strong>DEMO-{String(patient + 1).padStart(4, "0")}</strong>
          )}
        </div>
      </div>
    </article>
  );
}
async function downloadCard(card: CardRow, template: CardTemplate) {
  const canvas = document.createElement("canvas");
  canvas.width = 1000;
  canvas.height = 460;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.fillStyle = "#fff";
  ctx.fillRect(0, 0, 1000, 460);
  ctx.fillStyle = template.color;
  ctx.fillRect(0, 0, 1000, 100);
  ctx.fillStyle = "#fff";
  ctx.font = "bold 30px Arial";
  ctx.fillText("ULSHMS", 35, 60);
  ctx.font = "18px Arial";
  ctx.fillText("Hospital Management System · Patient Smart Identity Card", 220, 60);
  ctx.fillStyle = "#333";
  ctx.font = "24px Arial";
  const rows = [
    ["Patient Name", people[card.patient]],
    ...(template.email
      ? [["Email", `sample${card.patient + 1}@example.invalid`]]
      : []),
    ...(template.phone ? [["Phone", `+25190000010${card.patient}`]] : []),
    ...(template.dob ? [["Date of Birth", "2000-08-08"]] : []),
    ...(template.blood ? [["Blood Group", "AB+"]] : []),
    ...(template.address ? [["Address", "Addis Ababa, Ethiopia"]] : []),
  ];
  rows.forEach(([label, value], i) =>
    ctx.fillText(`${label}:  ${value}`, 35, 145 + i * 45),
  );
  const qr = new Image();
  qr.src = `/legacy/front/card-${card.patient + 1}.svg`;
  await qr.decode();
  ctx.drawImage(qr, 790, 150, 170, 170);
  ctx.font = "20px Arial";
  if (template.uniqueId !== false)
    ctx.fillText(`DEMO-${String(card.patient + 1).padStart(4, "0")}`, 805, 350);
  const a = document.createElement("a");
  a.href = canvas.toDataURL("image/png");
  a.download = `preview-patient-card-${card.patient + 1}.png`;
  a.click();
}
