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
  CalendarDays,
} from "lucide-react";
import { useState } from "react";
import { usePreviewRole } from "@/components/workspace";
import { visibleGroups, screens } from "@/lib/legacy";
import { RoleDashboard } from "@/components/role-portal";
const widgets = [
  ["Total Invoices", "12,500", "invoices", CreditCard],
  ["Total Bills", "8,250", "bills", FileText],
  ["Total Payments", "9,650", "payments", Banknote],
  ["Advance Payments", "2,850", "advanced-payments", Wallet],
  ["Available Beds", "16", "bed-status", BedDouble],
  ["Total Patients", "6", "patients", Users],
  ["Total Doctors", "3", "doctors", Stethoscope],
  ["Appointments", "6", "appointments", CalendarDays],
] as const;
export default function Dashboard() {
  const { t } = useLanguage();
  const role = usePreviewRole();
  const allowed = visibleGroups(role);
  const [period, setPeriod] = useState("This Month");
  if (role !== "Admin") return <RoleDashboard />;
  return (
    <section>
      <div className="page-heading">
        <h1>{t("Dashboard")}</h1>
        <select
          className="field !w-auto"
          aria-label="Report period"
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
        >
          <option value="This Month">{t("This Month")}</option>
          <option value="Last Month">{t("Last Month")}</option>
          <option value="This Year">{t("This Year")}</option>
        </select>
      </div>
      <div className="dashboard-widgets">
        {widgets
          .filter((w) =>
            allowed.includes(screens.find((s) => s.id === w[2])?.group || ""),
          )
          .map(([title, value, slug, Icon], i) => (
            <Link
              href={slug === "patients" ? "/patients" : `/modules/${slug}`}
              className="dashboard-widget"
              key={t(title)}
            >
              <span
                style={{
                  background: ["#6571ff", "#0ac074", "#ff8717", "#0099fb"][
                    i % 4
                  ],
                }}
              >
                <Icon size={28} />
              </span>
              <div>
                <strong
                  style={{
                    color: ["#6571ff", "#0ac074", "#ff8717", "#0099fb"][i % 4],
                  }}
                >
                  {value}
                </strong>
                <h3>{t(title)}</h3>
              </div>
            </Link>
          ))}
      </div>
      <div className="dashboard-columns">
        <div className="legacy-card">
          <div className="page-heading">
            <h2>{t("Income and Expense Report")}</h2>
            <span className="badge">{period}</span>
          </div>
          <div className="chart-legend">
            <span>● Income</span>
            <span>● Expense</span>
          </div>
          <div
            className="bar-chart"
            aria-label={`Sample income and expenses for ${period}`}
          >
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
              <div className="bar-group" key={m}>
                <div className="bars">
                  <span
                    style={{
                      height: `${25 + ((i * 17 + (period === "Last Month" ? 25 : 0)) % 65)}%`,
                    }}
                  />
                  <span style={{ height: `${15 + ((i * 11) % 50)}%` }} />
                </div>
                <small>{m}</small>
              </div>
            ))}
          </div>
        </div>
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
                <h3>{n}</h3>
                <p>October {i + 1}, 2026 · Sample notice</p>
              </div>
            </div>
          ))}
        </div>
      </div>
      <div className="legacy-card mt-6">
        <div className="page-heading">
          <h2>{t("Upcoming Appointments")}</h2>
          <Link className="text-brand" href="/modules/appointments">
            View All
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
                    <Link className="record-link" href="/modules/appointments">
                      <span className="avatar">{n[0]}</span>
                      {n}
                    </Link>
                  </td>
                  <td>Dr. Avery Reed</td>
                  <td>General Medicine</td>
                  <td>Oct {i + 5}, 2026 · 09:00 AM</td>
                  <td>
                    <span className="badge">Confirmed</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
