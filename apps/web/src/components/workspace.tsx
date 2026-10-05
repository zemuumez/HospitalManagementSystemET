"use client";
import { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Activity,
  Bell,
  BedDouble,
  CalendarDays,
  ChartPie,
  ChevronDown,
  ClipboardList,
  CreditCard,
  Droplets,
  FileText,
  HeartPulse,
  Hospital,
  Mail,
  Menu,
  Moon,
  Sun,
  Pill,
  Search,
  Settings,
  ShieldPlus,
  Stethoscope,
  Users,
  X,
} from "lucide-react";
import { api, type User } from "@/lib/api";
import { LanguageSwitcher, useLanguage } from "./language";
import { authClient } from "@/lib/auth-client";
import {
  groups,
  screens,
  screenHref,
  previewRoles,
  visibleGroups,
  groupScreens,
} from "@/lib/legacy";
import {
  roleNavigation,
  portalSections,
  roleCanSeeScreen,
} from "@/lib/role-preview";
type Identity = { user: User; permissions: string[] };
const PreviewRoleContext = createContext("Admin");
export const usePreviewRole = () => useContext(PreviewRoleContext);
const IdentityContext = createContext<Identity | null>(null);
export const useIdentity = () => useContext(IdentityContext);
export function Workspace({ children }: { children: React.ReactNode }) {
  const { t } = useLanguage();
  const [disabledModules, setDisabledModules] = useState<string[]>([]);
  useEffect(() => {
    const update = () => {
      try {
        setDisabledModules(
          JSON.parse(localStorage.getItem("hms-disabled-modules") || "[]"),
        );
      } catch {}
    };
    update();
    window.addEventListener("hms-modules-updated", update);
    return () => window.removeEventListener("hms-modules-updated", update);
  }, []);
  const [collapsed, setCollapsed] = useState(false);
  const [dark, setDark] = useState(false);
  useEffect(() => {
    setDark(localStorage.getItem("hms-theme") === "dark");
  }, []);
  function toggleTheme() {
    setDark((value) => {
      localStorage.setItem("hms-theme", value ? "light" : "dark");
      return !value;
    });
  }
  const [identity, setIdentity] = useState<Identity | null>(null);
  const [role, setRole] = useState("Admin");
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [mobile, setMobile] = useState(false);
  const [profile, setProfile] = useState(false);
  const [notifications, setNotifications] = useState(false);
  const path = usePathname();
  const router = useRouter();
  useEffect(() => {
    const saved = sessionStorage.getItem("hms-preview-role");
    if (saved && previewRoles.includes(saved)) setRole(saved);
  }, []);
  function changeRole(next: string) {
    setRole(next);
    sessionStorage.setItem("hms-preview-role", next);
    router.push("/dashboard");
  }
  const isLive = [
    "/patients",
    "/live-patients",
    "/communications",
    "/account",
    "/modules/users",
    "/modules/schedules",
    "/modules/appointments",
    "/portal/appointments",
    "/portal/invoices",
    "/modules/invoices",
    "/modules/accounts",
    "/portal/ipd",
    "/portal/opd",
    "/portal/cases",
    "/modules/beds",
    "/modules/bed-status",
    "/modules/patient-cases",
    "/modules/ipd-patient-departments",
    "/modules/opd-patient-departments",
  ].includes(path);
  const selected = screens.find((s) => screenHref(s) === path);
  const group = selected?.group;
  useEffect(() => {
    api<Identity>("me")
      .then((result) => {
        setIdentity(result);
        if (!sessionStorage.getItem("hms-preview-role")) {
          const current = previewRoles.find(
            (r) => r.toLowerCase().replaceAll(" ", "_") === result.user.role,
          );
          if (current) setRole(current);
        }
      })
      .catch((e) => setError(e.message));
  }, []);
  useEffect(() => {
    setMobile(false);
    setProfile(false);
  }, [path]);
  return (
    <IdentityContext.Provider value={identity}>
      <PreviewRoleContext.Provider value={role}>
        <div
          data-ready={identity ? "true" : "false"}
          className={`legacy-shell ${collapsed ? "is-collapsed" : ""} ${dark ? "legacy-dark" : ""}`}
        >
          <a className="sr-only focus:not-sr-only" href="#main-content">
            {t("Skip to content")}
          </a>
          {mobile && (
            <button
              className="legacy-overlay"
              aria-label={t("Close navigation")}
              onClick={() => setMobile(false)}
            />
          )}
          <aside className={`legacy-sidebar ${mobile ? "is-open" : ""}`}>
            <div className="legacy-logo">
              <Link href="/dashboard">
                <Hospital size={34} color="#6571ff" />
                <strong>ULSHMS</strong>
              </Link>
              <button
                aria-label={t("Close menu")}
                onClick={() => setMobile(false)}
                className="mobile-only"
              >
                <X size={20} />
              </button>
              <button
                className="desktop-only"
                aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
                onClick={() => setCollapsed(!collapsed)}
              >
                <Menu size={18} />
              </button>
            </div>
            <div className="legacy-search">
              <Search size={16} />
              <input
                aria-label={t("Search menu")}
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <nav aria-label={t("Main navigation")}>
              <Link
                className={`legacy-nav ${path === "/dashboard" ? "active" : ""}`}
                href="/dashboard"
              >
                <ChartPie size={18} />
                <span>{t("Dashboard")}</span>
              </Link>
              {roleNavigation(role)
                .filter((link) => !disabledModules.includes(link.group || ""))
                .filter((link) =>
                  (link.title + " " + t(link.title))
                    .toLowerCase()
                    .includes(search.toLowerCase()),
                )
                .map((link) => (
                  <Link
                    key={link.href}
                    className={`legacy-nav ${path === link.href || (link.group && group === link.group) ? "active" : ""}`}
                    href={link.href}
                  >
                    {(() => {
                      const Icon =
                        (
                          {
                            "Patient Smart Cards": CreditCard,
                            Users,
                            Appointments: CalendarDays,
                            Attendance: Users,
                            "Manage Attendance": ClipboardList,
                            "IPD - Patient In": BedDouble,
                            "OPD - Patient Out": Stethoscope,
                            Billings: CreditCard,
                            "Bed Management": BedDouble,
                            "Blood Banks": Droplets,
                            Doctors: Stethoscope,
                            Medicines: Pill,
                            Patients: Users,
                            Settings,
                            Prescriptions: FileText,
                            Odontogram: HeartPulse,
                          } as Record<string, typeof FileText>
                        )[link.title] || FileText;
                      return <Icon size={18} />;
                    })()}
                    <span>{t(link.title)}</span>
                  </Link>
                ))}
            </nav>
          </aside>
          <div className="legacy-body">
            <header className="legacy-header">
              <button
                className="mobile-only"
                aria-label={t("Open navigation")}
                onClick={() => setMobile(true)}
              >
                <Menu size={22} />
              </button>
              <nav
                className="legacy-submenu"
                aria-label={t("Module navigation")}
              >
                {group ? (
                  groupScreens(group)
                    .filter((s) => roleCanSeeScreen(role, s.id))
                    .map((s) => (
                      <Link
                        key={s.id}
                        className={selected?.id === s.id ? "active" : ""}
                        href={screenHref(s)}
                      >
                        {t(s.title)}
                      </Link>
                    ))
                ) : (
                  <Link className="active" href={path}>
                    {path === "/dashboard"
                      ? t("Dashboard")
                      : path === "/account"
                        ? t("My Profile")
                        : path.startsWith("/portal/")
                          ? t(
                              portalSections[path.split("/")[2]]?.title ||
                                "Dashboard",
                            )
                          : t("Communications")}
                  </Link>
                )}
              </nav>
              <div className="legacy-header-actions">
                <LanguageSwitcher />
                <button
                  aria-label={dark ? "Use light theme" : "Use dark theme"}
                  onClick={toggleTheme}
                >
                  {dark ? <Sun size={19} /> : <Moon size={19} />}
                </button>
                <div className="relative">
                  <button
                    aria-label={t("Notifications")}
                    onClick={() => setNotifications(!notifications)}
                  >
                    <Bell size={20} color="#6571ff" />
                  </button>
                  {notifications && (
                    <div className="legacy-popover">
                      <strong>{t("Notifications")}</strong>
                      <p>{t("No new notifications.")}</p>
                    </div>
                  )}
                </div>
                <div className="relative">
                  <button
                    className="legacy-user"
                    onClick={() => setProfile(!profile)}
                    aria-expanded={profile}
                  >
                    <span className="avatar">
                      {identity?.user.name?.slice(0, 1) || "A"}
                    </span>
                    <span>{identity?.user.name || "Hospital account"}</span>
                    <ChevronDown size={14} />
                  </button>
                  {profile && (
                    <div className="legacy-popover">
                      <p>{identity?.user.role}</p>
                      <Link href="/account">{t("My Profile")}</Link>
                      <Link href="/security-preview">
                        {t("Two-Factor Authentication")}
                      </Link>
                      <Link href="/">{t("Front Site")}</Link>
                      <Link href="/communications">
                        {t("Live communications")}
                      </Link>
                      <button
                        onClick={async () => {
                          await authClient.signOut();
                          window.location.assign("/login");
                        }}
                      >
                        {t("Sign Out")}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </header>
            <div className="preview-strip">
              <span>
                {isLive
                  ? t("Connected hospital workspace")
                  : t(
                      "Frontend preview · Sample data · Changes stay in this browser tab",
                    )}
              </span>
              {!isLive && (
                <label className="role-preview">
                  {t("View as")}{" "}
                  <select
                    aria-label={t("Preview role")}
                    value={role}
                    onChange={(e) => changeRole(e.target.value)}
                  >
                    {previewRoles.map((r) => (
                      <option key={r} value={r}>
                        {t(r)}
                      </option>
                    ))}
                  </select>
                </label>
              )}
              <Link href={isLive ? "/dashboard" : "/patients"}>
                {isLive ? t("Dashboard") : t("Open live patient register")}
              </Link>
            </div>
            <main id="main-content" className="legacy-main">
              {error ? (
                <p className="error" role="alert">
                  {t(error)}
                </p>
              ) : (
                children
              )}
            </main>
            <footer className="legacy-footer">
              {t("Copyright ©")} {new Date().getFullYear()}{" "}
              {t("ULSHMS. All rights reserved.")}
              <span>{t("Hospital Management System")}</span>
            </footer>
          </div>
        </div>
      </PreviewRoleContext.Provider>
    </IdentityContext.Provider>
  );
}
