"use client";
import { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
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
  const [identity, setIdentity] = useState<Identity | null>(null);
  const [role, setRole] = useState("Admin");
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [mobile, setMobile] = useState(false);
  const [profile, setProfile] = useState(false);
  const [notifications, setNotifications] = useState(false);
  const path = usePathname();
  const isLive = ["/live-patients", "/communications", "/account"].includes(
    path,
  );
  const selected = screens.find((s) => screenHref(s) === path);
  const group = selected?.group;
  useEffect(() => {
    api<Identity>("me")
      .then(setIdentity)
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
          className={`legacy-shell ${collapsed ? "is-collapsed" : ""} ${dark ? "legacy-dark" : ""}`}
        >
          <a className="sr-only focus:not-sr-only" href="#main-content">
            Skip to content
          </a>
          {mobile && (
            <button
              className="legacy-overlay"
              aria-label="Close navigation"
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
                aria-label="Close menu"
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
                aria-label="Search menu"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <nav aria-label="Main navigation">
              <Link
                className={`legacy-nav ${path === "/dashboard" ? "active" : ""}`}
                href="/dashboard"
              >
                <ChartPie size={18} />
                <span>{t("Dashboard")}</span>
              </Link>
              {visibleGroups(role)
                .filter((g) => !disabledModules.includes(g))
                .filter(
                  (g) =>
                    (g + " " + t(g))
                      .toLowerCase()
                      .includes(search.toLowerCase()) ||
                    screens.some(
                      (s) =>
                        s.group === g &&
                        (s.title + " " + t(s.title))
                          .toLowerCase()
                          .includes(search.toLowerCase()),
                    ),
                )
                .map((g) => {
                  const Icon =
                    (
                      {
                        "Patient Smart Cards": CreditCard,
                        Users: Users,
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
                        Settings: Settings,
                      } as Record<string, typeof FileText>
                    )[g] || FileText;
                  const first = groupScreens(g).find(
                    (s) =>
                      s.group === g &&
                      ((s.title + " " + t(s.title))
                        .toLowerCase()
                        .includes(search.toLowerCase()) ||
                        (g + " " + t(g))
                          .toLowerCase()
                          .includes(search.toLowerCase())),
                  )!;
                  return (
                    <Link
                      key={g}
                      className={`legacy-nav ${group === g ? "active" : ""}`}
                      href={screenHref(first)}
                    >
                      <Icon size={18} />
                      <span>{t(g)}</span>
                    </Link>
                  );
                })}
            </nav>
          </aside>
          <div className="legacy-body">
            <header className="legacy-header">
              <button
                className="mobile-only"
                aria-label="Open navigation"
                onClick={() => setMobile(true)}
              >
                <Menu size={22} />
              </button>
              <nav className="legacy-submenu" aria-label="Module navigation">
                {group ? (
                  groupScreens(group).map((s) => (
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
                      ? "Dashboard"
                      : path === "/account"
                        ? "My Profile"
                        : "Communications"}
                  </Link>
                )}
              </nav>
              <div className="legacy-header-actions">
                <LanguageSwitcher />
                <button
                  aria-label={dark ? "Use light theme" : "Use dark theme"}
                  onClick={() => setDark(!dark)}
                >
                  {dark ? <Sun size={19} /> : <Moon size={19} />}
                </button>
                <div className="relative">
                  <button
                    aria-label="Notifications"
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
                      <Link href="/">{t("Front Site")}</Link>
                      <Link href="/communications">Live communications</Link>
                      <button
                        onClick={async () => {
                          await authClient.signOut();
                          window.location.assign("/login");
                        }}
                      >
                        Sign Out
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </header>
            <div className="preview-strip">
              <span>
                {isLive
                  ? "Connected hospital workspace"
                  : "Frontend preview · Sample data · Changes stay in this browser tab"}
              </span>
              {!isLive && (
                <label className="role-preview">
                  View as{" "}
                  <select
                    aria-label="Preview role"
                    value={role}
                    onChange={(e) => setRole(e.target.value)}
                  >
                    {previewRoles.map((r) => (
                      <option key={r}>{r}</option>
                    ))}
                  </select>
                </label>
              )}
              <Link href={isLive ? "/patients" : "/live-patients"}>
                {isLive
                  ? "Return to frontend preview"
                  : "Open live patient register"}
              </Link>
            </div>
            <main id="main-content" className="legacy-main">
              {error ? (
                <p className="error" role="alert">
                  {error}
                </p>
              ) : (
                children
              )}
            </main>
            <footer className="legacy-footer">
              Copyright © {new Date().getFullYear()} ULSHMS. All rights
              reserved.<span>Hospital Management System</span>
            </footer>
          </div>
        </div>
      </PreviewRoleContext.Provider>
    </IdentityContext.Provider>
  );
}
