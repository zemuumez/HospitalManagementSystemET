"use client";
import { useLanguage } from "@/components/language";
import Link from "next/link";
import {
  CreditCard,
  FileText,
  Banknote,
  Wallet,
  BedDouble,
  Users,
  Stethoscope,
  ShieldPlus,
  BriefcaseMedical,
  FlaskConical,
  Pill,
  UserRound,
  LockKeyhole,
} from "lucide-react";
import { usePreviewRole } from "@/components/workspace";
import { RoleDashboard } from "@/components/role-portal";
const widgets = [
  ["Invoice Amount", "12,500", "invoices", CreditCard],
  ["Bill Amount", "8,250", "bills", FileText],
  ["Payment Amount", "9,650", "payments", Banknote],
  ["Advance Payment Amount", "2,850", "advanced-payments", Wallet],
  ["Available Beds", "16", "bed-status", BedDouble],
  ["Doctors", "3", "doctors", Stethoscope],
  ["Patients", "6", "patients", Users],
  ["Nurses", "0", "nurses", ShieldPlus],
  ["Admins", "1", "users", LockKeyhole],
  ["Accountants", "0", "accountants", BriefcaseMedical],
  ["Lab Technicians", "0", "lab-technicians", FlaskConical],
  ["Pharmacists", "0", "pharmacists", Pill],
] as const;
export default function Dashboard() {
  const { t } = useLanguage();
  const role = usePreviewRole();
  if (role !== "Admin") return <RoleDashboard />;
  return (
    <section>
      <h1 className="sr-only">{t("Dashboard")}</h1>
      <div className="dashboard-widgets">
        {widgets.map(([title, value, slug, Icon], i) => (
          <Link
            href={slug === "patients" ? "/patients" : `/modules/${slug}`}
            className="dashboard-widget"
            key={t(title)}
          >
            <span>
              <Icon size={28} />
            </span>
            <div>
              <strong>{value}</strong>
              <h3>{t(title)}</h3>
            </div>
          </Link>
        ))}
      </div>
      <div className="dashboard-columns">
        <div className="legacy-card">
          <div className="page-heading">
            <h2>{t("Yearly Income Expense Chart")}</h2>
            <span className="text-muted">{new Date().getFullYear()}</span>
          </div>
          <div className="chart-legend">
            <span>{t("● Income")}</span>
            <span>{t("● Expense")}</span>
          </div>
          <svg
            className="income-chart"
            viewBox="0 0 800 320"
            role="img"
            aria-label="Sample yearly income and expense chart"
          >
            {[0, 1, 2, 3, 4].map((i) => (
              <g key={i}>
                <line
                  x1="60"
                  y1={30 + i * 58}
                  x2="780"
                  y2={30 + i * 58}
                  stroke="currentColor"
                  opacity=".12"
                />
                <text x="4" y={35 + i * 58} fill="currentColor" fontSize="10">
                  {(16000 - i * 4000).toLocaleString()}
                </text>
              </g>
            ))}
            <path
              d="M60 262 L125 262 L190 262 C215 262 224 75 255 55 C282 35 297 262 320 262 L780 262"
              fill="none"
              stroke="#9966ff"
              strokeWidth="2"
            />
            <path
              d="M60 264 L780 264"
              stroke="#3285ef"
              fill="none"
              strokeWidth="2"
            />
            {[
              "Jan",
              "Feb",
              "Mar",
              "Apr",
              "May",
              "Jun",
              "Jul",
              "Aug",
              "Sep",
              "Oct",
              "Nov",
              "Dec",
            ].map((m, i) => (
              <g key={m}>
                <circle cx={60 + i * 65} cy="262" r="3" fill="#9966ff" />
                <text
                  x={60 + i * 65}
                  y="287"
                  textAnchor="middle"
                  fill="currentColor"
                  fontSize="11"
                >
                  {t(m)}
                </text>
              </g>
            ))}
            <text
              x="410"
              y="315"
              textAnchor="middle"
              fill="currentColor"
              fontSize="11"
            >
              {t("Month")}
            </text>
          </svg>
        </div>
        <div className="dashboard-aside">
          <Link className="dashboard-widget" href="/modules/receptionists">
            <span>
              <UserRound size={28} />
            </span>
            <div>
              <strong>0</strong>
              <h3>{t("Receptionists")}</h3>
            </div>
          </Link>
          <div className="legacy-card">
            <h2>{t("Notice Boards")}</h2>
            {[
              "Welcome to the hospital frontend preview",
              "Monthly staff meeting",
              "Stock review and inventory checks",
            ].map((n, i) => (
              <div className="notice-row" key={n}>
                <span className="notice-dot" />
                <div>
                  <h3>{t(n)}</h3>
                  <p>
                    {t("October")}
                    {i + 1}, 2026 · Sample notice
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="dashboard-bottom">
        <div className="legacy-card">
          <h2>{t("Enquiries")}</h2>
          <p className="empty-enquiries">{t("No Enquiries Yet")}</p>
        </div>
        <div className="legacy-card">
          <div className="page-heading">
            <h2>{t("Appointments")}</h2>
            <Link className="text-brand" href="/modules/appointments">
              {t("View All")}
            </Link>
          </div>
          <div className="legacy-table-wrap">
            <table className="legacy-table">
              <thead>
                <tr>
                  <th>{t("Patient")}</th>
                  <th>{t("Doctor")}</th>
                  <th>{t("Department")}</th>
                  <th>{t("Date")}</th>
                  <th>{t("Status")}</th>
                </tr>
              </thead>
              <tbody>
                {["Alex Morgan", "Jamie Wilson", "Taylor Davis"].map((n, i) => (
                  <tr key={n}>
                    <td>
                      <Link
                        className="record-link"
                        href="/modules/appointments"
                      >
                        <span className="avatar">{n[0]}</span>
                        {n}
                      </Link>
                    </td>
                    <td>Dr. Avery Reed</td>
                    <td>{t("General Medicine")}</td>
                    <td>
                      {t("Oct")}
                      {i + 5}, 2026 · 09:00 AM
                    </td>
                    <td>
                      <span className="badge">{t("Confirmed")}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </section>
  );
}
