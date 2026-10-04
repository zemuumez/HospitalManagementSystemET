"use client";
import { useState } from "react";
import Link from "next/link";
import { Eye, Search, X, Printer } from "lucide-react";
import { usePreviewRole } from "./workspace";
import { useLanguage } from "./language";
import { Modal } from "./modal";
import {
  canPreviewPortal,
  portalSections,
  roleNavigation,
} from "@/lib/role-preview";
const sampleValue = (column: string, i: number, role: string): string => {
  const values: Record<string, string> = {
    Patient:
      role === "Patient"
        ? "Alex Morgan"
        : ["Alex Morgan", "Jamie Wilson", "Taylor Davis"][i],
    Doctor: ["Dr. Avery Reed", "Dr. Robin Patel", "Dr. Quinn Parker"][i],
    Department: ["General Medicine", "Cardiology", "Paediatrics"][i],
    Phone: "+251900000000",
    Email: `doctor${i + 1}@example.invalid`,
    Title: ["Monthly staff meeting", "Stock review", "Hospital notice"][i],
    Description: "Sample notice for frontend review",
    Status: "Completed",
    "Bill Status": "Paid",
    Month: ["September", "August", "July"][i],
    Year: "2026",
    "Basic Salary": "10000",
    Allowance: "1500",
    Deductions: "500",
    "Net Salary": "11000",
    "Medical History": "No known allergies",
    "Current Medication": "See prescription details",
    "Health Insurance": "No",
    "Low Income": "No",
    Reference: "General Medicine",
    Amount: "1200",
    Fee: "250",
    Bed: "General — G01",
    "Standard Charge": "250",
    "Payment Mode": "Cash",
    "Total Visits": String(i + 1),
    Vaccination: "Hepatitis B",
    "Dose Number": String(i + 1),
    "Report Type": ["Diagnosis", "Investigation", "Operation"][i],
    "Document Type": "Medical Report",
    Duration: "30 minutes",
  };
  if (/Date/.test(column)) return `2026-0${9 - i}-15`;
  return values[column] || `DEMO-${i + 1}`;
};
export function RolePortal({ section }: { section: string }) {
  const role = usePreviewRole();
  const { t } = useLanguage();
  const config = portalSections[section];
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [selected, setSelected] = useState<number | null>(null);
  const [tab, setTab] = useState("Overview");
  const [descending, setDescending] = useState(false);
  if (!canPreviewPortal(role, section))
    return (
      <section className="legacy-card">
        <h1>{t("Select a role to preview this screen")}</h1>
        <Link href="/dashboard">{t("Dashboard")}</Link>
      </section>
    );
  const rows = [0, 1, 2].filter(
    (i) =>
      config.columns.some((c) =>
        sampleValue(c, i, role).toLowerCase().includes(search.toLowerCase()),
      ) &&
      (!status || sampleValue("Status", i, role) === status),
  );
  if (descending) rows.reverse();
  const tabs =
    section === "ipd"
      ? [
          "Overview",
          "Diagnosis",
          "Consultant Register",
          "Charges",
          "Prescriptions",
          "Payments",
          "Timeline",
        ]
      : section === "opd"
        ? ["Overview", "Visits", "Diagnosis", "Prescriptions", "Timeline"]
        : ["Overview"];
  return (
    <section>
      <div className="page-heading">
        <h1>{t(config.title)}</h1>
        <span className="badge">
          {t(role === "Patient" ? "My Records" : "Staff Records")}
        </span>
      </div>
      <p className="muted mb-5">
        {t("Synthetic records for this role only. No live patient data.")}
      </p>
      <div className="page-heading">
        <label className="legacy-search">
          <Search size={16} />
          <input
            aria-label={t("Search records")}
            placeholder={t("Search")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </label>
        <select
          className="field !w-auto"
          aria-label={t("Status")}
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">{t("All Statuses")}</option>
          {["Completed", "Pending"].map((s) => (
            <option key={s} value={s}>
              {t(s)}
            </option>
          ))}
        </select>
      </div>
      <div className="legacy-table-wrap">
        <table className="legacy-table">
          <thead>
            <tr>
              {config.columns.map((c) => (
                <th key={c}>
                  <button onClick={() => setDescending(!descending)}>
                    {t(c)} ↕
                  </button>
                </th>
              ))}
              <th>{t("Action")}</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((i) => (
              <tr key={i}>
                {config.columns.map((c) => (
                  <td key={c}>{t(sampleValue(c, i, role))}</td>
                ))}
                <td>
                  <button
                    className="text-brand"
                    aria-label={`${t("View")} ${i + 1}`}
                    onClick={() => {
                      setSelected(i);
                      setTab("Overview");
                    }}
                  >
                    <Eye size={19} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && <p className="p-6">{t("No records found")}</p>}
      </div>
      <p className="muted mt-4">
        {t("Total Records")}: {rows.length}
      </p>
      {selected !== null && (
        <Modal titleId="portal-detail-title" onClose={() => setSelected(null)}>
          <div className="p-6">
            <div className="page-heading">
              <h2 id="portal-detail-title">
                {t(config.title)} · DEMO-{selected + 1}
              </h2>
              <button aria-label={t("Close")} onClick={() => setSelected(null)}>
                <X />
              </button>
            </div>
            <div className="legacy-tabs" role="tablist">
              {tabs.map((label) => (
                <button
                  role="tab"
                  aria-selected={tab === label}
                  className={tab === label ? "active" : ""}
                  key={label}
                  onClick={() => setTab(label)}
                >
                  {t(label)}
                </button>
              ))}
            </div>
            <div role="tabpanel">
              {tab === "Overview" ? (
                <dl className="detail-grid">
                  {config.columns.map((c) => (
                    <div key={c}>
                      <dt className="muted">{t(c)}</dt>
                      <dd>{t(sampleValue(c, selected, role))}</dd>
                    </div>
                  ))}
                </dl>
              ) : (
                <p className="p-4">{t("No records found")}</p>
              )}
            </div>
            {section === "prescriptions" && (
              <div className="mt-5">
                <h3>{t("Prescription Details")}</h3>
                <p>{t("Sample prescription for layout review only")}</p>
                <dl>
                  <dt>{t("Medicine")}</dt>
                  <dd>{t("Sample medicine")}</dd>
                  <dt>{t("Instructions")}</dt>
                  <dd>
                    {t("Follow the prescribing clinician’s instructions")}
                  </dd>
                </dl>
              </div>
            )}
            <button
              className="btn-secondary mt-5"
              onClick={() => window.print()}
            >
              <Printer size={16} />
              {t("Print")}
            </button>
          </div>
        </Modal>
      )}
    </section>
  );
}
export function RoleDashboard() {
  const role = usePreviewRole();
  const { t } = useLanguage();
  const links = roleNavigation(role);
  const tiles =
    role === "Patient"
      ? [
          {
            title: "Total Appointments",
            href: "/portal/appointments",
            value: "3",
          },
          {
            title: "Today’s Appointments",
            href: "/portal/appointments",
            value: "0",
          },
          {
            title: "Total Meetings",
            href: "/portal/consultations",
            value: "3",
          },
          { title: "Total Bills", href: "/portal/bills", value: "3,600" },
        ]
      : links.slice(0, 4).map((l) => ({ ...l, value: "3" }));
  return (
    <section>
      <div className="page-heading">
        <h1>{t("Dashboard")}</h1>
        <span className="badge">{t(role)}</span>
      </div>
      <div className="dashboard-widgets">
        {tiles.map((tile, i) => (
          <Link key={tile.title} href={tile.href} className="dashboard-widget">
            <span
              style={{
                background: ["#6571ff", "#0ac074", "#ff8717", "#0099fb"][i],
              }}
            >
              <Eye />
            </span>
            <div>
              <strong>{tile.value}</strong>
              <h3>{t(tile.title)}</h3>
            </div>
          </Link>
        ))}
      </div>
      <div className="legacy-card mt-6">
        <h2>
          {t(role === "Patient" ? "Recent Appointments" : "My Workspace")}
        </h2>
        <div className="role-shortcuts">
          {links.map((l) => (
            <Link className="btn-secondary" key={l.href} href={l.href}>
              {t(l.title)}
            </Link>
          ))}
        </div>
        {role === "Patient" && (
          <div className="legacy-table-wrap">
            <table className="legacy-table">
              <thead>
                <tr>
                  {["Doctor", "Department", "Date", "Status"].map((c) => (
                    <th key={c}>{t(c)}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {[0, 1, 2].map((i) => (
                  <tr key={i}>
                    {["Doctor", "Department", "Date", "Status"].map((c) => (
                      <td key={c}>{t(sampleValue(c, i, role))}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}
