"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { useLanguage } from "./language";
import { useIdentity } from "./workspace";
import { api } from "@/lib/api";
import { groups } from "@/lib/legacy";
import {
  Pencil,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Lock,
  ShieldAlert,
  Eye,
  EyeOff,
} from "lucide-react";

export const SETTINGS_TABS = [
  { id: "settings", label: "General Settings", href: "/modules/settings" },
  {
    id: "hospital-schedule",
    label: "Hospital Schedule",
    href: "/modules/hospital-schedule",
  },
  {
    id: "modules-setting",
    label: "Modules Setting",
    href: "/modules/modules-setting",
  },
  {
    id: "currency-settings",
    label: "Currencies",
    href: "/modules/currency-settings",
  },
  {
    id: "operation-categories",
    label: "Operation Categories",
    href: "/modules/operation-categories",
  },
  { id: "operations", label: "Operations", href: "/modules/operations" },
  {
    id: "payment-gateway",
    label: "Payment Gateways",
    href: "/modules/payment-gateway",
  },
  {
    id: "add-custom-fields",
    label: "Add Custom Fields",
    href: "/modules/add-custom-fields",
  },
  {
    id: "patient-queue-theme",
    label: "Patient Queue Theme",
    href: "/modules/patient-queue-theme",
  },
];

export const FRONT_CMS_TABS = [
  {
    id: "front-settings",
    label: "Front CMS Settings",
    href: "/modules/front-settings",
  },
  {
    id: "front-cms-services",
    label: "Front CMS Services",
    href: "/modules/front-cms-services",
  },
  {
    id: "notice-boards",
    label: "Notice Boards",
    href: "/modules/notice-boards",
  },
  { id: "testimonials", label: "Testimonials", href: "/modules/testimonials" },
  { id: "complaints", label: "Complaints", href: "/modules/complaints" },
];

export function SettingsWorkspace({ id }: { id: string }) {
  const { t } = useLanguage();
  const identity = useIdentity();
  const isAdmin = identity?.user?.role === "admin";
  const isFrontCms = id === "front-settings";
  const tabs = isFrontCms ? FRONT_CMS_TABS : SETTINGS_TABS;

  // Non-admin rejection banner
  if (identity && !isAdmin) {
    return (
      <div className="legacy-workspace" data-ready="true">
        <div className="module-subtabs-nav">
          {tabs.map((tab) => (
            <Link
              key={tab.id}
              href={tab.href}
              className={`module-subtab-link ${id === tab.id ? "active" : ""}`}
            >
              {t(tab.label)}
            </Link>
          ))}
        </div>
        <div
          className="legacy-card p-6"
          style={{ borderColor: "#ef4444", borderLeftWidth: 4 }}
        >
          <div className="flex items-center gap-3 text-red-500 mb-2">
            <ShieldAlert size={24} />
            <h2 className="text-lg font-semibold text-red-500 m-0">
              {t("Access Denied")}
            </h2>
          </div>
          <p className="text-muted mb-4">
            {t(
              "Administrator privileges are required to view or configure hospital settings.",
            )}
          </p>
          <Link href="/dashboard" className="btn-action-blue inline-block">
            {t("Return to Dashboard")}
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="legacy-workspace" data-ready="true">
      <div className="module-subtabs-nav">
        {tabs.map((tab) => (
          <Link
            key={tab.id}
            href={tab.href}
            className={`module-subtab-link ${id === tab.id ? "active" : ""}`}
          >
            {t(tab.label)}
          </Link>
        ))}
      </div>

      {id === "settings" && <GeneralSettingsTab />}
      {id === "hospital-schedule" && <HospitalScheduleTab />}
      {id === "modules-setting" && <ModulesSettingTab />}
      {id === "currency-settings" && <CurrencySettingsTab />}
      {id === "payment-gateway" && <PaymentGatewayTab />}
      {id === "patient-queue-theme" && <PatientQueueThemeTab />}
      {id === "front-settings" && <FrontCmsSettingsTab />}
    </div>
  );
}

// ---------------------------------------------------------------------------
// 1. General Settings Tab
// ---------------------------------------------------------------------------

function GeneralSettingsTab() {
  const { t, setLocale } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");
  const [showSecret, setShowSecret] = useState(false);

  const [form, setForm] = useState({
    app_name: "",
    company_name: "",
    hospital_email: "",
    hospital_phone: "",
    hospital_from_day: "Monday",
    hospital_from_time: "08:00",
    hospital_address: "",
    current_currency: "ETB",
    country_phone: "+251",
    default_lang: "en",
    about_us: "",
    app_logo: "",
    favicon: "",
    facebook_url: "",
    twitter_url: "",
    instagram_url: "",
    linkedIn_url: "",
    open_ai_key: "",
    model_name: "gpt-4o",
    custom_serial_prefix: "HMS-",
  });

  const logoInputRef = useRef<HTMLInputElement>(null);
  const faviconInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    let active = true;
    api<Record<string, string>>("general-settings")
      .then((data) => {
        if (!active) return;
        setForm((prev) => ({
          ...prev,
          app_name: data.app_name || prev.app_name || "ULSHMS",
          company_name:
            data.company_name ||
            data.hospital_name ||
            prev.company_name ||
            "ULSHMS",
          hospital_email:
            data.hospital_email ||
            prev.hospital_email ||
            "hospital@example.invalid",
          hospital_phone:
            data.hospital_phone || prev.hospital_phone || "+251900000000",
          hospital_from_day:
            data.hospital_from_day ||
            data.hospital_start_day ||
            prev.hospital_from_day ||
            "Monday",
          hospital_from_time:
            data.hospital_from_time ||
            data.hospital_start_time ||
            prev.hospital_from_time ||
            "08:00",
          hospital_address:
            data.hospital_address ||
            prev.hospital_address ||
            "Addis Ababa, Ethiopia",
          current_currency:
            data.current_currency ||
            data.currency ||
            prev.current_currency ||
            "ETB",
          country_phone:
            data.country_phone ||
            data.country_code ||
            prev.country_phone ||
            "+251",
          default_lang:
            data.default_lang ||
            data.default_language ||
            prev.default_lang ||
            "en",
          about_us:
            data.about_us || prev.about_us || "Hospital Management System",
          app_logo: data.app_logo || data.logo_url || prev.app_logo || "",
          favicon: data.favicon || data.favicon_url || prev.favicon || "",
          facebook_url: data.facebook_url ?? prev.facebook_url,
          twitter_url: data.twitter_url ?? prev.twitter_url,
          instagram_url: data.instagram_url ?? prev.instagram_url,
          linkedIn_url: data.linkedIn_url ?? prev.linkedIn_url,
          open_ai_key: data.open_ai_key ?? prev.open_ai_key,
          model_name: data.model_name ?? prev.model_name,
          custom_serial_prefix:
            data.custom_serial_prefix ?? prev.custom_serial_prefix,
        }));
        setLoading(false);
      })
      .catch((err) => {
        if (!active) return;
        setErrorNotice(err.message || "Failed to load general settings.");
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function handleFileUpload(
    field: "app_logo" | "favicon",
    file: File | undefined,
  ) {
    if (!file) return;
    setErrorNotice("");
    setSuccessNotice("");
    if (
      !["image/png", "image/jpeg", "image/webp", "image/x-icon"].includes(
        file.type,
      )
    ) {
      setErrorNotice(
        "Please choose a PNG, JPEG, WEBP, or ICO image smaller than 25 MB.",
      );
      return;
    }
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("isPublic", "true");
      const res = await fetch("/api/hms/attachments", {
        method: "POST",
        body: formData,
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || "Attachment upload failed");
      }
      const data = await res.json();
      const contentUrl = `/api/hms/attachments/${data.token}/content`;
      setForm((prev) => ({ ...prev, [field]: contentUrl }));
      setSuccessNotice(
        `${field === "app_logo" ? "Application logo" : "Favicon"} uploaded successfully.`,
      );
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to upload image.");
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSuccessNotice("");
    setErrorNotice("");

    // Required fields validation
    if (!form.app_name.trim()) {
      setErrorNotice("Application Name is required.");
      return;
    }
    if (!form.company_name.trim()) {
      setErrorNotice("Company Name is required.");
      return;
    }
    if (!form.hospital_email.trim() || !form.hospital_email.includes("@")) {
      setErrorNotice("A valid Hospital Email is required.");
      return;
    }
    if (!form.hospital_phone.trim()) {
      setErrorNotice("Hospital Phone is required.");
      return;
    }
    if (!form.hospital_address.trim()) {
      setErrorNotice("Hospital Address is required.");
      return;
    }
    if (!form.about_us.trim()) {
      setErrorNotice("About Us description is required.");
      return;
    }

    setSaving(true);
    try {
      const payload: Record<string, string> = {
        app_name: form.app_name,
        company_name: form.company_name,
        hospital_name: form.company_name,
        hospital_email: form.hospital_email,
        hospital_phone: form.hospital_phone,
        hospital_from_day: form.hospital_from_day,
        hospital_start_day: form.hospital_from_day,
        hospital_from_time: form.hospital_from_time,
        hospital_start_time: form.hospital_from_time,
        hospital_address: form.hospital_address,
        current_currency: form.current_currency,
        currency: form.current_currency,
        country_phone: form.country_phone,
        country_code: form.country_phone,
        default_lang: form.default_lang,
        default_language: form.default_lang,
        about_us: form.about_us,
        app_logo: form.app_logo,
        logo_url: form.app_logo,
        favicon: form.favicon,
        favicon_url: form.favicon,
        facebook_url: form.facebook_url,
        twitter_url: form.twitter_url,
        instagram_url: form.instagram_url,
        linkedIn_url: form.linkedIn_url,
        open_ai_key: form.open_ai_key,
        model_name: form.model_name,
        custom_serial_prefix: form.custom_serial_prefix,
      };

      await api("general-settings", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      if (form.default_lang === "am" || form.default_lang === "en") {
        setLocale(form.default_lang);
      }

      setSuccessNotice("General settings saved successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to save general settings.");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="legacy-card p-6">
        <p className="text-muted">{t("Loading settings...")}</p>
      </div>
    );
  }

  return (
    <div className="legacy-card settings-card">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-lg font-semibold m-0">{t("General Settings")}</h2>
      </div>

      {successNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(34, 197, 94, 0.1)",
            borderColor: "#22c55e",
            color: "#22c55e",
          }}
          role="status"
        >
          <CheckCircle2 size={18} />
          <span>{t(successNotice)}</span>
        </div>
      )}

      {errorNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(239, 68, 68, 0.1)",
            borderColor: "#ef4444",
            color: "#ef4444",
          }}
          role="alert"
        >
          <AlertCircle size={18} />
          <span>{t(errorNotice)}</span>
        </div>
      )}

      <form onSubmit={handleSave} className="legacy-form">
        <label>
          <span className="label">
            {t("App Name")}: <b className="text-red-500">*</b>
          </span>
          <input
            className="field"
            type="text"
            required
            value={form.app_name}
            onChange={(e) => setForm({ ...form, app_name: e.target.value })}
          />
        </label>

        <label>
          <span className="label">
            {t("Company Name")}: <b className="text-red-500">*</b>
          </span>
          <input
            className="field"
            type="text"
            required
            value={form.company_name}
            onChange={(e) => setForm({ ...form, company_name: e.target.value })}
          />
        </label>

        <label>
          <span className="label">
            {t("Hospital Email")}: <b className="text-red-500">*</b>
          </span>
          <input
            className="field"
            type="email"
            required
            value={form.hospital_email}
            onChange={(e) =>
              setForm({ ...form, hospital_email: e.target.value })
            }
          />
        </label>

        <label>
          <span className="label">
            {t("Hospital Phone")}: <b className="text-red-500">*</b>
          </span>
          <input
            className="field"
            type="tel"
            required
            value={form.hospital_phone}
            onChange={(e) =>
              setForm({ ...form, hospital_phone: e.target.value })
            }
          />
        </label>

        <label>
          <span className="label">
            {t("Hospital From Day")}: <b className="text-red-500">*</b>
          </span>
          <select
            className="field"
            value={form.hospital_from_day}
            onChange={(e) =>
              setForm({ ...form, hospital_from_day: e.target.value })
            }
          >
            {[
              "Monday",
              "Tuesday",
              "Wednesday",
              "Thursday",
              "Friday",
              "Saturday",
              "Sunday",
            ].map((day) => (
              <option key={day} value={day}>
                {t(day)}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="label">
            {t("Hospital From Time")}: <b className="text-red-500">*</b>
          </span>
          <input
            className="field"
            type="time"
            required
            value={form.hospital_from_time}
            onChange={(e) =>
              setForm({ ...form, hospital_from_time: e.target.value })
            }
          />
        </label>

        <label className="form-span">
          <span className="label">
            {t("Address")}: <b className="text-red-500">*</b>
          </span>
          <input
            className="field"
            type="text"
            required
            value={form.hospital_address}
            onChange={(e) =>
              setForm({ ...form, hospital_address: e.target.value })
            }
          />
        </label>

        <label>
          <span className="label">
            {t("Currency")}: <b className="text-red-500">*</b>
          </span>
          <select
            className="field"
            value={form.current_currency}
            onChange={(e) =>
              setForm({ ...form, current_currency: e.target.value })
            }
          >
            {["ETB", "USD", "EUR", "KES"].map((cur) => (
              <option key={cur} value={cur}>
                {cur}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="label">{t("Country Code")}:</span>
          <select
            className="field"
            value={form.country_phone}
            onChange={(e) =>
              setForm({ ...form, country_phone: e.target.value })
            }
          >
            {["+251", "+254", "+1", "+44", "+91"].map((code) => (
              <option key={code} value={code}>
                {code}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="label">{t("Default Language")}:</span>
          <select
            className="field"
            value={form.default_lang}
            onChange={(e) => setForm({ ...form, default_lang: e.target.value })}
          >
            <option value="en">{t("English")}</option>
            <option value="am">አማርኛ</option>
          </select>
        </label>

        <label className="form-span">
          <span className="label">
            {t("About Us")}: <b className="text-red-500">*</b>
          </span>
          <textarea
            className="field"
            rows={4}
            required
            value={form.about_us}
            onChange={(e) => setForm({ ...form, about_us: e.target.value })}
          />
        </label>

        {/* Logo Attachment Upload/Replace/Remove */}
        <label>
          <span className="label">{t("App Logo")}:</span>
          <div className="flex items-center gap-4 mt-2">
            <span className="setting-image-picker relative inline-block">
              <img
                src={form.app_logo || "/legacy/hms-logo.png"}
                alt="App Logo Preview"
                style={{
                  width: 90,
                  height: 50,
                  objectFit: "contain",
                  background: "#1e2433",
                  borderRadius: 6,
                  padding: 4,
                }}
              />
              <button
                type="button"
                className="absolute -top-2 -right-2 bg-blue-600 text-white rounded-full p-1"
                title="Change Logo"
                onClick={() => logoInputRef.current?.click()}
              >
                <Pencil size={12} />
              </button>
            </span>
            <input
              ref={logoInputRef}
              type="file"
              accept="image/png,image/jpeg,image/webp"
              style={{ display: "none" }}
              onChange={(e) =>
                handleFileUpload("app_logo", e.target.files?.[0])
              }
            />
            {form.app_logo && (
              <button
                type="button"
                className="text-xs text-red-400 hover:text-red-300 flex items-center gap-1 border border-red-500/30 rounded px-2 py-1"
                onClick={() => setForm({ ...form, app_logo: "" })}
              >
                <Trash2 size={12} />
                {t("Remove")}
              </button>
            )}
          </div>
          <small className="block text-muted mt-1">
            {t("Authorized attachment (PNG, JPG, WEBP under 25 MB).")}
          </small>
        </label>

        {/* Favicon Attachment Upload/Replace/Remove */}
        <label>
          <span className="label">{t("Favicon")}:</span>
          <div className="flex items-center gap-4 mt-2">
            <span className="setting-image-picker relative inline-block">
              <img
                src={form.favicon || "/legacy/hms-logo.png"}
                alt="Favicon Preview"
                style={{
                  width: 40,
                  height: 40,
                  objectFit: "contain",
                  background: "#1e2433",
                  borderRadius: 6,
                  padding: 4,
                }}
              />
              <button
                type="button"
                className="absolute -top-2 -right-2 bg-blue-600 text-white rounded-full p-1"
                title="Change Favicon"
                onClick={() => faviconInputRef.current?.click()}
              >
                <Pencil size={12} />
              </button>
            </span>
            <input
              ref={faviconInputRef}
              type="file"
              accept="image/png,image/jpeg,image/webp,image/x-icon"
              style={{ display: "none" }}
              onChange={(e) => handleFileUpload("favicon", e.target.files?.[0])}
            />
            {form.favicon && (
              <button
                type="button"
                className="text-xs text-red-400 hover:text-red-300 flex items-center gap-1 border border-red-500/30 rounded px-2 py-1"
                onClick={() => setForm({ ...form, favicon: "" })}
              >
                <Trash2 size={12} />
                {t("Remove")}
              </button>
            )}
          </div>
          <small className="block text-muted mt-1">
            {t("Icon format (PNG, ICO under 25 MB).")}
          </small>
        </label>

        {/* Social URLs (Optional) */}
        <label>
          <span className="label">{t("Facebook URL")}:</span>
          <input
            className="field"
            type="url"
            value={form.facebook_url}
            onChange={(e) =>
              setForm({ ...form, facebook_url: e.target.value })
            }
          />
        </label>

        <label>
          <span className="label">{t("Twitter URL")}:</span>
          <input
            className="field"
            type="url"
            value={form.twitter_url}
            onChange={(e) => setForm({ ...form, twitter_url: e.target.value })}
          />
        </label>

        <label>
          <span className="label">{t("Instagram URL")}:</span>
          <input
            className="field"
            type="url"
            value={form.instagram_url}
            onChange={(e) =>
              setForm({ ...form, instagram_url: e.target.value })
            }
          />
        </label>

        <label>
          <span className="label">{t("LinkedIn URL")}:</span>
          <input
            className="field"
            type="url"
            value={form.linkedIn_url}
            onChange={(e) => setForm({ ...form, linkedIn_url: e.target.value })}
          />
        </label>

        {/* Provider Secrets Protection: Open AI Key */}
        <label className="form-span">
          <div className="flex justify-between items-center mb-1">
            <span className="label flex items-center gap-1.5">
              <Lock size={14} className="text-amber-400" />
              {t("Open AI Key")}:
            </span>
            <div className="flex items-center gap-2">
              {form.open_ai_key === "[CONFIGURED]" && (
                <span className="text-xs px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                  {t("Configured & Redacted")}
                </span>
              )}
              {form.open_ai_key && (
                <button
                  type="button"
                  className="text-xs text-muted hover:text-red-400"
                  onClick={() => setForm({ ...form, open_ai_key: "" })}
                >
                  {t("Clear Secret")}
                </button>
              )}
            </div>
          </div>
          <div className="relative">
            <input
              className="field pr-10"
              type={showSecret ? "text" : "password"}
              placeholder={
                form.open_ai_key === "[CONFIGURED]"
                  ? "[CONFIGURED]"
                  : "sk-..."
              }
              value={form.open_ai_key}
              onChange={(e) =>
                setForm({ ...form, open_ai_key: e.target.value })
              }
            />
            <button
              type="button"
              className="absolute right-3 top-2.5 text-muted hover:text-foreground"
              onClick={() => setShowSecret(!showSecret)}
            >
              {showSecret ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
          <small className="block text-muted mt-1">
            {t(
              "Sensitive secret. Stored securely and never returned to browser in plaintext. Leave unchanged to keep current credential, or clear to remove.",
            )}
          </small>
        </label>

        <label>
          <span className="label">{t("Model Name")}:</span>
          <input
            className="field"
            type="text"
            value={form.model_name}
            onChange={(e) => setForm({ ...form, model_name: e.target.value })}
          />
        </label>

        <label>
          <span className="label">{t("Custom Serial Prefix")}:</span>
          <input
            className="field"
            type="text"
            value={form.custom_serial_prefix}
            onChange={(e) =>
              setForm({ ...form, custom_serial_prefix: e.target.value })
            }
          />
        </label>

        <div className="form-span flex justify-end gap-3 mt-6">
          <button
            type="submit"
            className="btn-action-blue flex items-center gap-2"
            disabled={saving}
          >
            {saving ? t("Saving...") : t("Save Settings")}
          </button>
        </div>
      </form>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 2. Hospital Schedule Tab
// ---------------------------------------------------------------------------

interface DaySchedule {
  day_of_week: number;
  day_name: string;
  start_time: string;
  end_time: string;
  is_closed: boolean;
}

const DAY_NAMES = [
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
  "Sunday",
];

function HospitalScheduleTab() {
  const { t } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");

  const [days, setDays] = useState<DaySchedule[]>(() =>
    DAY_NAMES.map((name, i) => ({
      day_of_week: i + 1,
      day_name: name,
      start_time: "08:00",
      end_time: i === 6 ? "12:00" : i === 5 ? "13:00" : "17:00",
      is_closed: i === 6,
    })),
  );

  useEffect(() => {
    let active = true;
    api<Array<any>>("hospital-schedules")
      .then((data) => {
        if (!active) return;
        if (Array.isArray(data) && data.length > 0) {
          const map = new Map<number, any>();
          data.forEach((d) => map.set(d.day_of_week, d));
          setDays(
            DAY_NAMES.map((name, i) => {
              const dow = i + 1;
              const item = map.get(dow);
              return {
                day_of_week: dow,
                day_name: name,
                start_time: item?.start_time || "08:00",
                end_time: item?.end_time || "17:00",
                is_closed: item?.is_closed ?? (dow === 7),
              };
            }),
          );
        }
        setLoading(false);
      })
      .catch((err) => {
        if (!active) return;
        setErrorNotice(err.message || "Failed to load hospital schedules.");
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  function updateDay(index: number, patch: Partial<DaySchedule>) {
    setDays((prev) =>
      prev.map((d, i) => (i === index ? { ...d, ...patch } : d)),
    );
    setSuccessNotice("");
    setErrorNotice("");
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSuccessNotice("");
    setErrorNotice("");

    // Validate times for open days
    for (const d of days) {
      if (!d.is_closed) {
        if (d.end_time <= d.start_time) {
          setErrorNotice(
            `${t(d.day_name)}: ${t("Closing time must be after opening time.")}`,
          );
          return;
        }
      }
    }

    setSaving(true);
    try {
      const payload = days.map((d) => ({
        day_of_week: d.day_of_week,
        start_time: d.start_time,
        end_time: d.end_time,
        is_closed: d.is_closed,
      }));

      await api("hospital-schedules", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      setSuccessNotice("Hospital schedule saved successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to save hospital schedule.");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="legacy-card p-6">
        <p className="text-muted">{t("Loading schedule...")}</p>
      </div>
    );
  }

  return (
    <div className="legacy-card">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-lg font-semibold m-0">{t("Hospital Schedule")}</h2>
      </div>

      {successNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(34, 197, 94, 0.1)",
            borderColor: "#22c55e",
            color: "#22c55e",
          }}
          role="status"
        >
          <CheckCircle2 size={18} />
          <span>{t(successNotice)}</span>
        </div>
      )}

      {errorNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(239, 68, 68, 0.1)",
            borderColor: "#ef4444",
            color: "#ef4444",
          }}
          role="alert"
        >
          <AlertCircle size={18} />
          <span>{t(errorNotice)}</span>
        </div>
      )}

      <form onSubmit={handleSave}>
        <div className="flex flex-col gap-4">
          {days.map((d, i) => (
            <div
              key={d.day_name}
              className="flex items-center gap-4 p-3 rounded border border-gray-800 bg-[#161c28]"
            >
              <label className="flex items-center gap-2 min-w-[140px] cursor-pointer">
                <input
                  type="checkbox"
                  checked={!d.is_closed}
                  onChange={(e) =>
                    updateDay(i, { is_closed: !e.target.checked })
                  }
                />
                <span
                  className={
                    d.is_closed ? "text-muted" : "font-medium text-white"
                  }
                >
                  {t(d.day_name)}
                </span>
              </label>

              <div className="flex items-center gap-2 flex-1">
                <input
                  aria-label={`${t(d.day_name)} ${t("Opening time")}`}
                  className="field max-w-[140px]"
                  type="time"
                  disabled={d.is_closed}
                  required={!d.is_closed}
                  value={d.start_time}
                  onChange={(e) => updateDay(i, { start_time: e.target.value })}
                />
                <span className="text-muted text-sm">{t("to")}</span>
                <input
                  aria-label={`${t(d.day_name)} ${t("Closing time")}`}
                  className="field max-w-[140px]"
                  type="time"
                  disabled={d.is_closed}
                  required={!d.is_closed}
                  value={d.end_time}
                  onChange={(e) => updateDay(i, { end_time: e.target.value })}
                />
              </div>

              {d.is_closed && (
                <span className="text-xs px-2 py-1 rounded bg-red-500/10 text-red-400 border border-red-500/20">
                  {t("Closed")}
                </span>
              )}
            </div>
          ))}
        </div>

        <div className="flex justify-end mt-6">
          <button
            type="submit"
            className="btn-action-blue"
            disabled={saving}
          >
            {saving ? t("Saving...") : t("Save Schedule")}
          </button>
        </div>
      </form>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 3. Modules Setting Tab
// ---------------------------------------------------------------------------

function ModulesSettingTab() {
  const { t } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");

  const [moduleMap, setModuleMap] = useState<
    Record<string, { id: string; name: string; isActive: boolean }>
  >({});

  useEffect(() => {
    let active = true;
    api<{ modules_setting: Array<any> }>("modules-setting")
      .then((data) => {
        if (!active) return;
        const map: Record<
          string,
          { id: string; name: string; isActive: boolean }
        > = {};
        if (data && Array.isArray(data.modules_setting)) {
          data.modules_setting.forEach((ms) => {
            map[ms.module_key] = {
              id: ms.id,
              name: ms.name,
              isActive: ms.is_active,
            };
          });
        }
        setModuleMap(map);
        setLoading(false);
      })
      .catch((err) => {
        if (!active) return;
        setErrorNotice(err.message || "Failed to load module settings.");
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function handleToggle(moduleKey: string, nextActive: boolean) {
    setSuccessNotice("");
    setErrorNotice("");

    // Optimistically update
    setModuleMap((prev) => ({
      ...prev,
      [moduleKey]: {
        ...(prev[moduleKey] || {
          id: moduleKey,
          name: moduleKey,
          isActive: true,
        }),
        isActive: nextActive,
      },
    }));

    try {
      await api(`modules-setting/${moduleKey}`, {
        method: "PUT",
        body: JSON.stringify({ is_active: nextActive }),
      });

      // Update localStorage disabled list for workspace shell navigation
      const disabled = Object.entries(moduleMap)
        .filter(([k, v]) => (k === moduleKey ? !nextActive : !v.isActive))
        .map(([k, v]) => v.name || k);
      localStorage.setItem("hms-disabled-modules", JSON.stringify(disabled));
      window.dispatchEvent(new Event("hms-modules-updated"));

      setSuccessNotice("Module setting updated successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to update module setting.");
    }
  }

  if (loading) {
    return (
      <div className="legacy-card p-6">
        <p className="text-muted">{t("Loading modules...")}</p>
      </div>
    );
  }

  return (
    <div className="legacy-card">
      <div className="flex justify-between items-center mb-4">
        <div>
          <h2 className="text-lg font-semibold m-0">{t("Modules Setting")}</h2>
          <p className="text-muted text-sm mt-1">
            {t(
              "Enable or disable operational modules across clinical workflows.",
            )}
          </p>
        </div>
      </div>

      {successNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(34, 197, 94, 0.1)",
            borderColor: "#22c55e",
            color: "#22c55e",
          }}
          role="status"
        >
          <CheckCircle2 size={18} />
          <span>{t(successNotice)}</span>
        </div>
      )}

      {errorNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(239, 68, 68, 0.1)",
            borderColor: "#ef4444",
            color: "#ef4444",
          }}
          role="alert"
        >
          <AlertCircle size={18} />
          <span>{t(errorNotice)}</span>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
        {Object.entries(moduleMap).map(([key, ms]) => (
          <div
            key={key}
            className="flex items-center justify-between p-3.5 rounded border border-gray-800 bg-[#161c28]"
          >
            <div>
              <span className="font-medium text-white block">
                {t(ms.name || key)}
              </span>
              <span className="text-xs text-muted font-mono">{key}</span>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                className="sr-only peer"
                checked={ms.isActive}
                onChange={(e) => handleToggle(key, e.target.checked)}
              />
              <div className="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>
        ))}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 4. Currencies Tab
// ---------------------------------------------------------------------------

const CURRENCY_LIST = [
  { code: "ETB", name: "Ethiopian Birr", symbol: "Br" },
  { code: "USD", name: "US Dollar", symbol: "$" },
  { code: "EUR", name: "Euro", symbol: "€" },
  { code: "GBP", name: "British Pound", symbol: "£" },
  { code: "KES", name: "Kenyan Shilling", symbol: "KSh" },
  { code: "INR", name: "Indian Rupee", symbol: "₹" },
];

function CurrencySettingsTab() {
  const { t } = useLanguage();
  const [activeCurrency, setActiveCurrency] = useState("ETB");
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");

  useEffect(() => {
    api<Record<string, string>>("general-settings")
      .then((data) => {
        setActiveCurrency(data.current_currency || data.currency || "ETB");
      })
      .catch(() => {});
  }, []);

  async function handleSelect(code: string) {
    setSuccessNotice("");
    setErrorNotice("");
    setSaving(true);
    try {
      await api("general-settings", {
        method: "POST",
        body: JSON.stringify({
          current_currency: code,
          currency: code,
        }),
      });
      setActiveCurrency(code);
      setSuccessNotice(`Active currency set to ${code} successfully.`);
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to update currency.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="legacy-card">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-lg font-semibold m-0">{t("Currencies")}</h2>
          <p className="text-muted text-sm mt-1">
            {t(
              "Select the active hospital operating currency used for billing and encounters.",
            )}
          </p>
        </div>
      </div>

      {successNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(34, 197, 94, 0.1)",
            borderColor: "#22c55e",
            color: "#22c55e",
          }}
          role="status"
        >
          <CheckCircle2 size={18} />
          <span>{t(successNotice)}</span>
        </div>
      )}

      {errorNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(239, 68, 68, 0.1)",
            borderColor: "#ef4444",
            color: "#ef4444",
          }}
          role="alert"
        >
          <AlertCircle size={18} />
          <span>{t(errorNotice)}</span>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="legacy-table w-full">
          <thead>
            <tr>
              <th>{t("Currency Name")}</th>
              <th>{t("Currency Icon")}</th>
              <th>{t("Currency Code")}</th>
              <th className="text-right">{t("Action")}</th>
            </tr>
          </thead>
          <tbody>
            {CURRENCY_LIST.map((c) => {
              const isCurrent = activeCurrency === c.code;
              return (
                <tr key={c.code} className={isCurrent ? "bg-blue-900/10" : ""}>
                  <td className="font-medium text-white">{t(c.name)}</td>
                  <td className="font-semibold text-blue-400">{c.symbol}</td>
                  <td>
                    <span className="font-mono text-sm px-2 py-0.5 rounded bg-gray-800 border border-gray-700">
                      {c.code}
                    </span>
                  </td>
                  <td className="text-right">
                    {isCurrent ? (
                      <span className="inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 font-medium">
                        <CheckCircle2 size={12} />
                        {t("Active Currency")}
                      </span>
                    ) : (
                      <button
                        type="button"
                        className="btn-action-blue py-1 px-3 text-xs"
                        disabled={saving}
                        onClick={() => handleSelect(c.code)}
                      >
                        {t("Make Default")}
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 5. Payment Gateways Tab (Strict Secret Redaction & Preservation)
// ---------------------------------------------------------------------------

function PaymentGatewayTab() {
  const { t } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");

  const [form, setForm] = useState({
    stripe_key: "",
    stripe_secret: "",
    paypal_client_id: "",
    paypal_secret: "",
    paypal_mode: "sandbox",
    razorpay_key: "",
    razorpay_secret: "",
    flutterwave_public_key: "",
    flutterwave_secret_key: "",
    phonepe_merchant_id: "",
    phonepe_merchant_user_id: "",
    phonepe_env: "UAT",
    phonepe_salt_key: "",
    phonepe_salt_index: "1",
    phonepe_merchant_transaction_id: "",
    paystack_public_key: "",
    paystack_secret_key: "",
  });

  useEffect(() => {
    let active = true;
    api<Record<string, string>>("general-settings")
      .then((data) => {
        if (!active) return;
        setForm((prev) => ({
          ...prev,
          stripe_key: data.stripe_key || "",
          stripe_secret: data.stripe_secret || "",
          paypal_client_id: data.paypal_client_id || "",
          paypal_secret: data.paypal_secret || "",
          paypal_mode: data.paypal_mode || "sandbox",
          razorpay_key: data.razorpay_key || "",
          razorpay_secret: data.razorpay_secret || "",
          flutterwave_public_key: data.flutterwave_public_key || "",
          flutterwave_secret_key: data.flutterwave_secret_key || "",
          phonepe_merchant_id: data.phonepe_merchant_id || "",
          phonepe_merchant_user_id: data.phonepe_merchant_user_id || "",
          phonepe_env: data.phonepe_env || "UAT",
          phonepe_salt_key: data.phonepe_salt_key || "",
          phonepe_salt_index: data.phonepe_salt_index || "1",
          phonepe_merchant_transaction_id:
            data.phonepe_merchant_transaction_id || "",
          paystack_public_key: data.paystack_public_key || "",
          paystack_secret_key: data.paystack_secret_key || "",
        }));
        setLoading(false);
      })
      .catch((err) => {
        if (!active) return;
        setErrorNotice(err.message || "Failed to load gateway settings.");
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSuccessNotice("");
    setErrorNotice("");
    setSaving(true);

    try {
      await api("general-settings", {
        method: "POST",
        body: JSON.stringify(form),
      });

      setSuccessNotice("Payment gateway settings saved successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to save payment gateway settings.");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="legacy-card p-6">
        <p className="text-muted">{t("Loading payment gateways...")}</p>
      </div>
    );
  }

  return (
    <div className="legacy-card">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-lg font-semibold m-0">{t("Payment Gateways")}</h2>
          <p className="text-muted text-sm mt-1">
            {t(
              "Configure online payment providers. Sensitive secret keys are masked and preserved.",
            )}
          </p>
        </div>
      </div>

      {successNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(34, 197, 94, 0.1)",
            borderColor: "#22c55e",
            color: "#22c55e",
          }}
          role="status"
        >
          <CheckCircle2 size={18} />
          <span>{t(successNotice)}</span>
        </div>
      )}

      {errorNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(239, 68, 68, 0.1)",
            borderColor: "#ef4444",
            color: "#ef4444",
          }}
          role="alert"
        >
          <AlertCircle size={18} />
          <span>{t(errorNotice)}</span>
        </div>
      )}

      <form onSubmit={handleSave} className="flex flex-col gap-8">
        {/* Stripe Section */}
        <div className="border border-gray-800 rounded-lg p-5 bg-[#141a24]">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
            <span>Stripe</span>
          </h3>
          <div className="legacy-form">
            <label>
              <span className="label">{t("Stripe Key")}</span>
              <input
                className="field"
                type="text"
                placeholder="pk_live_..."
                value={form.stripe_key}
                onChange={(e) =>
                  setForm({ ...form, stripe_key: e.target.value })
                }
              />
            </label>
            <label>
              <div className="flex justify-between items-center mb-1">
                <span className="label flex items-center gap-1">
                  <Lock size={12} className="text-amber-400" />
                  {t("Stripe Secret")}
                </span>
                {form.stripe_secret === "[CONFIGURED]" && (
                  <button
                    type="button"
                    className="text-xs text-red-400 hover:text-red-300"
                    onClick={() => setForm({ ...form, stripe_secret: "" })}
                  >
                    {t("Clear")}
                  </button>
                )}
              </div>
              <input
                className="field"
                type="password"
                placeholder={
                  form.stripe_secret === "[CONFIGURED]"
                    ? "[CONFIGURED]"
                    : "sk_live_..."
                }
                value={form.stripe_secret}
                onChange={(e) =>
                  setForm({ ...form, stripe_secret: e.target.value })
                }
              />
            </label>
          </div>
        </div>

        {/* PayPal Section */}
        <div className="border border-gray-800 rounded-lg p-5 bg-[#141a24]">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
            <span>PayPal</span>
          </h3>
          <div className="legacy-form">
            <label>
              <span className="label">{t("PayPal Client ID")}</span>
              <input
                className="field"
                type="text"
                value={form.paypal_client_id}
                onChange={(e) =>
                  setForm({ ...form, paypal_client_id: e.target.value })
                }
              />
            </label>
            <label>
              <div className="flex justify-between items-center mb-1">
                <span className="label flex items-center gap-1">
                  <Lock size={12} className="text-amber-400" />
                  {t("PayPal Secret")}
                </span>
                {form.paypal_secret === "[CONFIGURED]" && (
                  <button
                    type="button"
                    className="text-xs text-red-400 hover:text-red-300"
                    onClick={() => setForm({ ...form, paypal_secret: "" })}
                  >
                    {t("Clear")}
                  </button>
                )}
              </div>
              <input
                className="field"
                type="password"
                placeholder={
                  form.paypal_secret === "[CONFIGURED]"
                    ? "[CONFIGURED]"
                    : ""
                }
                value={form.paypal_secret}
                onChange={(e) =>
                  setForm({ ...form, paypal_secret: e.target.value })
                }
              />
            </label>
            <label>
              <span className="label">{t("PayPal Mode")}</span>
              <select
                className="field"
                value={form.paypal_mode}
                onChange={(e) =>
                  setForm({ ...form, paypal_mode: e.target.value })
                }
              >
                <option value="sandbox">Sandbox</option>
                <option value="live">Live</option>
              </select>
            </label>
          </div>
        </div>

        {/* Razorpay Section */}
        <div className="border border-gray-800 rounded-lg p-5 bg-[#141a24]">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
            <span>Razorpay</span>
          </h3>
          <div className="legacy-form">
            <label>
              <span className="label">{t("Razorpay Key")}</span>
              <input
                className="field"
                type="text"
                value={form.razorpay_key}
                onChange={(e) =>
                  setForm({ ...form, razorpay_key: e.target.value })
                }
              />
            </label>
            <label>
              <div className="flex justify-between items-center mb-1">
                <span className="label flex items-center gap-1">
                  <Lock size={12} className="text-amber-400" />
                  {t("Razorpay Secret")}
                </span>
                {form.razorpay_secret === "[CONFIGURED]" && (
                  <button
                    type="button"
                    className="text-xs text-red-400 hover:text-red-300"
                    onClick={() => setForm({ ...form, razorpay_secret: "" })}
                  >
                    {t("Clear")}
                  </button>
                )}
              </div>
              <input
                className="field"
                type="password"
                placeholder={
                  form.razorpay_secret === "[CONFIGURED]"
                    ? "[CONFIGURED]"
                    : ""
                }
                value={form.razorpay_secret}
                onChange={(e) =>
                  setForm({ ...form, razorpay_secret: e.target.value })
                }
              />
            </label>
          </div>
        </div>

        {/* Flutterwave Section */}
        <div className="border border-gray-800 rounded-lg p-5 bg-[#141a24]">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
            <span>Flutterwave</span>
          </h3>
          <div className="legacy-form">
            <label>
              <span className="label">{t("Flutterwave Public Key")}</span>
              <input
                className="field"
                type="text"
                value={form.flutterwave_public_key}
                onChange={(e) =>
                  setForm({ ...form, flutterwave_public_key: e.target.value })
                }
              />
            </label>
            <label>
              <div className="flex justify-between items-center mb-1">
                <span className="label flex items-center gap-1">
                  <Lock size={12} className="text-amber-400" />
                  {t("Flutterwave Secret Key")}
                </span>
                {form.flutterwave_secret_key === "[CONFIGURED]" && (
                  <button
                    type="button"
                    className="text-xs text-red-400 hover:text-red-300"
                    onClick={() =>
                      setForm({ ...form, flutterwave_secret_key: "" })
                    }
                  >
                    {t("Clear")}
                  </button>
                )}
              </div>
              <input
                className="field"
                type="password"
                placeholder={
                  form.flutterwave_secret_key === "[CONFIGURED]"
                    ? "[CONFIGURED]"
                    : ""
                }
                value={form.flutterwave_secret_key}
                onChange={(e) =>
                  setForm({ ...form, flutterwave_secret_key: e.target.value })
                }
              />
            </label>
          </div>
        </div>

        {/* PhonePe Section */}
        <div className="border border-gray-800 rounded-lg p-5 bg-[#141a24]">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
            <span>PhonePe</span>
          </h3>
          <div className="legacy-form">
            <label>
              <span className="label">{t("Phonepe Merchant ID")}</span>
              <input
                className="field"
                type="text"
                value={form.phonepe_merchant_id}
                onChange={(e) =>
                  setForm({ ...form, phonepe_merchant_id: e.target.value })
                }
              />
            </label>
            <label>
              <span className="label">{t("Phonepe Salt Index")}</span>
              <input
                className="field"
                type="text"
                value={form.phonepe_salt_index}
                onChange={(e) =>
                  setForm({ ...form, phonepe_salt_index: e.target.value })
                }
              />
            </label>
            <label>
              <div className="flex justify-between items-center mb-1">
                <span className="label flex items-center gap-1">
                  <Lock size={12} className="text-amber-400" />
                  {t("Phonepe Salt Key")}
                </span>
                {form.phonepe_salt_key === "[CONFIGURED]" && (
                  <button
                    type="button"
                    className="text-xs text-red-400 hover:text-red-300"
                    onClick={() => setForm({ ...form, phonepe_salt_key: "" })}
                  >
                    {t("Clear")}
                  </button>
                )}
              </div>
              <input
                className="field"
                type="password"
                placeholder={
                  form.phonepe_salt_key === "[CONFIGURED]"
                    ? "[CONFIGURED]"
                    : ""
                }
                value={form.phonepe_salt_key}
                onChange={(e) =>
                  setForm({ ...form, phonepe_salt_key: e.target.value })
                }
              />
            </label>
          </div>
        </div>

        {/* Paystack Section */}
        <div className="border border-gray-800 rounded-lg p-5 bg-[#141a24]">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
            <span>Paystack</span>
          </h3>
          <div className="legacy-form">
            <label>
              <span className="label">{t("Paystack Public Key")}</span>
              <input
                className="field"
                type="text"
                value={form.paystack_public_key}
                onChange={(e) =>
                  setForm({ ...form, paystack_public_key: e.target.value })
                }
              />
            </label>
            <label>
              <div className="flex justify-between items-center mb-1">
                <span className="label flex items-center gap-1">
                  <Lock size={12} className="text-amber-400" />
                  {t("Paystack Secret Key")}
                </span>
                {form.paystack_secret_key === "[CONFIGURED]" && (
                  <button
                    type="button"
                    className="text-xs text-red-400 hover:text-red-300"
                    onClick={() =>
                      setForm({ ...form, paystack_secret_key: "" })
                    }
                  >
                    {t("Clear")}
                  </button>
                )}
              </div>
              <input
                className="field"
                type="password"
                placeholder={
                  form.paystack_secret_key === "[CONFIGURED]"
                    ? "[CONFIGURED]"
                    : ""
                }
                value={form.paystack_secret_key}
                onChange={(e) =>
                  setForm({ ...form, paystack_secret_key: e.target.value })
                }
              />
            </label>
          </div>
        </div>

        <div className="flex justify-end mt-4">
          <button
            type="submit"
            className="btn-action-blue"
            disabled={saving}
          >
            {saving ? t("Saving...") : t("Save Payment Gateways")}
          </button>
        </div>
      </form>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 6. Patient Queue Theme Tab
// ---------------------------------------------------------------------------

function PatientQueueThemeTab() {
  const { t } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");

  const [color, setColor] = useState("#6571ff");
  const [message, setMessage] = useState("Please wait for your number");
  const [theme, setTheme] = useState("modern-blue");

  useEffect(() => {
    let active = true;
    api<Record<string, string>>("general-settings")
      .then((data) => {
        if (!active) return;
        setColor(data.queue_theme_color || "#6571ff");
        setMessage(data.queue_theme_message || "Please wait for your number");
        setTheme(data.queue_theme || "modern-blue");
        setLoading(false);
      })
      .catch((err) => {
        if (!active) return;
        setErrorNotice(err.message || "Failed to load queue theme settings.");
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSuccessNotice("");
    setErrorNotice("");
    setSaving(true);

    try {
      await api("general-settings", {
        method: "POST",
        body: JSON.stringify({
          queue_theme_color: color,
          queue_theme_message: message,
          queue_theme: theme,
        }),
      });

      // Also persist to localStorage for immediate client-side component preview
      localStorage.setItem(
        "hms-queue-theme",
        JSON.stringify({ color, message, theme }),
      );

      setSuccessNotice("Patient queue theme saved successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to save queue theme.");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="legacy-card p-6">
        <p className="text-muted">{t("Loading queue theme...")}</p>
      </div>
    );
  }

  return (
    <div className="legacy-card">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-lg font-semibold m-0">
            {t("Patient Queue Theme")}
          </h2>
          <p className="text-muted text-sm mt-1">
            {t("Customize the display header and colors for OPD patient queues.")}
          </p>
        </div>
      </div>

      {successNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(34, 197, 94, 0.1)",
            borderColor: "#22c55e",
            color: "#22c55e",
          }}
          role="status"
        >
          <CheckCircle2 size={18} />
          <span>{t(successNotice)}</span>
        </div>
      )}

      {errorNotice && (
        <div
          className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
          style={{
            backgroundColor: "rgba(239, 68, 68, 0.1)",
            borderColor: "#ef4444",
            color: "#ef4444",
          }}
          role="alert"
        >
          <AlertCircle size={18} />
          <span>{t(errorNotice)}</span>
        </div>
      )}

      <form onSubmit={handleSave} className="legacy-form">
        <label>
          <span className="label">{t("Theme Preset")}</span>
          <select
            className="field"
            value={theme}
            onChange={(e) => {
              setTheme(e.target.value);
              if (e.target.value === "modern-blue") setColor("#6571ff");
              if (e.target.value === "emerald-care") setColor("#10b981");
              if (e.target.value === "slate-clean") setColor("#64748b");
            }}
          >
            <option value="modern-blue">Modern Blue</option>
            <option value="emerald-care">Emerald Care</option>
            <option value="slate-clean">Slate Clean</option>
          </select>
        </label>

        <label>
          <span className="label">{t("Header Color")}</span>
          <div className="flex items-center gap-3">
            <input
              type="color"
              className="field h-10 w-20 p-1 cursor-pointer"
              value={color}
              onChange={(e) => setColor(e.target.value)}
            />
            <span className="font-mono text-sm text-muted">{color}</span>
          </div>
        </label>

        <label className="form-span">
          <span className="label">{t("Queue Display Message")}</span>
          <input
            className="field"
            type="text"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
          />
        </label>

        {/* Live Preview Box */}
        <div className="form-span mt-4 p-5 rounded-lg border border-gray-800 bg-[#141a24]">
          <span className="text-xs uppercase tracking-wider text-muted font-semibold block mb-2">
            {t("Live Display Preview")}
          </span>
          <div
            className="p-4 rounded text-white font-medium text-center shadow-lg"
            style={{ backgroundColor: color }}
          >
            {message}
          </div>
        </div>

        <div className="form-span flex justify-end mt-6">
          <button
            type="submit"
            className="btn-action-blue"
            disabled={saving}
          >
            {saving ? t("Saving...") : t("Save Queue Theme")}
          </button>
        </div>
      </form>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 7. Front CMS Settings Tab (Real Persistence & Attachment Uploads)
// ---------------------------------------------------------------------------

function FrontCmsSettingsTab() {
  const { t } = useLanguage();
  const [tab, setTab] = useState("Home");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");
  const [errorNotice, setErrorNotice] = useState("");

  const [cmsValues, setCmsValues] = useState<Record<string, string>>({
    home_title: "Comprehensive Compassionate Healthcare in Addis Ababa",
    home_description:
      "Delivering premier specialized clinical care, diagnostics, and 24/7 emergency response.",
    home_page_image: "/legacy/front/home.png",
    home_page_experience: "15",
    about_title: "Leading Healthcare Innovation in East Africa",
    about_description:
      "Our hospital unites international diagnostic standards, compassionate bedside nursing, and multidisciplinary surgical excellence.",
    appointment_title: "Book an Appointment",
    appointment_description:
      "Schedule a consultation with our experienced clinical specialists.",
    terms:
      "Patient confidentiality, ethical standards, and emergency admission guidelines.",
    map_address: "Bole Sub-City, Addis Ababa, Ethiopia",
    map_url: "https://maps.google.com",
  });

  useEffect(() => {
    let active = true;
    api<Array<{ key: string; value: string; type: string }>>(
      "front-cms-settings",
    )
      .then((data) => {
        if (!active) return;
        if (Array.isArray(data)) {
          const map: Record<string, string> = {};
          data.forEach((item) => {
            map[item.key] = item.value;
          });
          setCmsValues((prev) => ({ ...prev, ...map }));
        }
        setLoading(false);
      })
      .catch((err) => {
        if (!active) return;
        setErrorNotice(err.message || "Failed to load Front CMS settings.");
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function handleImageUpload(field: string, file: File | undefined) {
    if (!file) return;
    setErrorNotice("");
    setSuccessNotice("");
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("isPublic", "true");
      const res = await fetch("/api/hms/attachments", {
        method: "POST",
        body: formData,
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || "Image upload failed");
      }
      const data = await res.json();
      setCmsValues((prev) => ({
        ...prev,
        [field]: `/api/hms/attachments/${data.token}/content`,
      }));
      setSuccessNotice("Image uploaded successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to upload image.");
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSuccessNotice("");
    setErrorNotice("");
    setSaving(true);

    try {
      // Build payload for all CMS settings
      const payload = Object.entries(cmsValues).map(([k, v]) => ({
        key: k,
        value: v,
        type: k.startsWith("about")
          ? "about"
          : k.startsWith("appointment")
            ? "appointment"
            : k.startsWith("terms")
              ? "terms"
              : k.startsWith("map")
                ? "map"
                : "home",
      }));

      await api("front-cms-settings", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      setSuccessNotice("Front CMS settings saved successfully.");
    } catch (err: any) {
      setErrorNotice(err.message || "Failed to save Front CMS settings.");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="legacy-card p-6">
        <p className="text-muted">{t("Loading Front CMS...")}</p>
      </div>
    );
  }

  const cmsTabList = [
    "Home",
    "About Us",
    "Appointment",
    "Terms & Conditions",
    "Map",
  ];

  return (
    <div>
      <div className="detail-tabs cms-tabs mb-4">
        {cmsTabList.map((name) => (
          <button
            key={name}
            type="button"
            className={tab === name ? "active" : ""}
            onClick={() => {
              setTab(name);
              setSuccessNotice("");
              setErrorNotice("");
            }}
          >
            {t(name)}
          </button>
        ))}
      </div>

      <div className="legacy-card">
        <h2 className="text-lg font-semibold mb-4">
          {t("Front Setting Details")} - {t(tab)}
        </h2>

        {successNotice && (
          <div
            className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
            style={{
              backgroundColor: "rgba(34, 197, 94, 0.1)",
              borderColor: "#22c55e",
              color: "#22c55e",
            }}
            role="status"
          >
            <CheckCircle2 size={18} />
            <span>{t(successNotice)}</span>
          </div>
        )}

        {errorNotice && (
          <div
            className="alert-notice p-3 mb-4 rounded border flex items-center gap-2"
            style={{
              backgroundColor: "rgba(239, 68, 68, 0.1)",
              borderColor: "#ef4444",
              color: "#ef4444",
            }}
            role="alert"
          >
            <AlertCircle size={18} />
            <span>{t(errorNotice)}</span>
          </div>
        )}

        <form onSubmit={handleSave} className="legacy-form">
          {tab === "Home" && (
            <>
              <label className="form-span">
                <span className="label">{t("Home Page Title")}</span>
                <input
                  className="field"
                  type="text"
                  value={cmsValues.home_title || ""}
                  onChange={(e) =>
                    setCmsValues({ ...cmsValues, home_title: e.target.value })
                  }
                />
              </label>
              <label className="form-span">
                <span className="label">{t("Home Page Description")}</span>
                <textarea
                  className="field"
                  rows={3}
                  value={cmsValues.home_description || ""}
                  onChange={(e) =>
                    setCmsValues({
                      ...cmsValues,
                      home_description: e.target.value,
                    })
                  }
                />
              </label>
              <label>
                <span className="label">{t("Years of Experience")}</span>
                <input
                  className="field"
                  type="number"
                  value={cmsValues.home_page_experience || "10"}
                  onChange={(e) =>
                    setCmsValues({
                      ...cmsValues,
                      home_page_experience: e.target.value,
                    })
                  }
                />
              </label>
              <label>
                <span className="label">{t("Home Page Hero Image")}</span>
                <div className="flex items-center gap-4 mt-2">
                  <img
                    src={
                      cmsValues.home_page_image || "/legacy/front/home.png"
                    }
                    alt="Hero Preview"
                    style={{
                      width: 100,
                      height: 60,
                      objectFit: "cover",
                      borderRadius: 6,
                    }}
                  />
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) =>
                      handleImageUpload(
                        "home_page_image",
                        e.target.files?.[0],
                      )
                    }
                  />
                </div>
              </label>
            </>
          )}

          {tab === "About Us" && (
            <>
              <label className="form-span">
                <span className="label">{t("About Title")}</span>
                <input
                  className="field"
                  type="text"
                  value={cmsValues.about_title || ""}
                  onChange={(e) =>
                    setCmsValues({ ...cmsValues, about_title: e.target.value })
                  }
                />
              </label>
              <label className="form-span">
                <span className="label">{t("About Description")}</span>
                <textarea
                  className="field"
                  rows={5}
                  value={cmsValues.about_description || ""}
                  onChange={(e) =>
                    setCmsValues({
                      ...cmsValues,
                      about_description: e.target.value,
                    })
                  }
                />
              </label>
            </>
          )}

          {tab === "Appointment" && (
            <>
              <label className="form-span">
                <span className="label">{t("Appointment Title")}</span>
                <input
                  className="field"
                  type="text"
                  value={cmsValues.appointment_title || ""}
                  onChange={(e) =>
                    setCmsValues({
                      ...cmsValues,
                      appointment_title: e.target.value,
                    })
                  }
                />
              </label>
              <label className="form-span">
                <span className="label">{t("Appointment Description")}</span>
                <textarea
                  className="field"
                  rows={4}
                  value={cmsValues.appointment_description || ""}
                  onChange={(e) =>
                    setCmsValues({
                      ...cmsValues,
                      appointment_description: e.target.value,
                    })
                  }
                />
              </label>
            </>
          )}

          {tab === "Terms & Conditions" && (
            <label className="form-span">
              <span className="label">{t("Terms & Conditions Content")}</span>
              <textarea
                className="field"
                rows={8}
                value={cmsValues.terms || ""}
                onChange={(e) =>
                  setCmsValues({ ...cmsValues, terms: e.target.value })
                }
              />
            </label>
          )}

          {tab === "Map" && (
            <>
              <label className="form-span">
                <span className="label">{t("Map Address")}</span>
                <input
                  className="field"
                  type="text"
                  value={cmsValues.map_address || ""}
                  onChange={(e) =>
                    setCmsValues({ ...cmsValues, map_address: e.target.value })
                  }
                />
              </label>
              <label className="form-span">
                <span className="label">{t("Map Embed URL")}</span>
                <input
                  className="field"
                  type="url"
                  value={cmsValues.map_url || ""}
                  onChange={(e) =>
                    setCmsValues({ ...cmsValues, map_url: e.target.value })
                  }
                />
              </label>
            </>
          )}

          <div className="form-span flex justify-end mt-6">
            <button
              type="submit"
              className="btn-action-blue"
              disabled={saving}
            >
              {saving ? t("Saving...") : t("Save Front CMS Settings")}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
