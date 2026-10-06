"use client";

import { useEffect, useState } from "react";
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

export default function Dashboard() {
  const { t } = useLanguage();
  const role = usePreviewRole();

  const [counts, setCounts] = useState({
    invoices: "12,500",
    bills: "8,250",
    payments: "9,650",
    advancePayments: "2,850",
    availableBeds: "16",
    doctors: "3",
    patients: "6",
    nurses: "0",
    admins: "1",
    accountants: "0",
    labTechnicians: "0",
    pharmacists: "0",
    receptionists: "0",
  });

  const [notices, setNotices] = useState<
    Array<{ id: string; title: string; date?: string }>
  >([
    {
      id: "1",
      title: "Hospital Clinical Operations Normal",
      date: "October 6, 2026",
    },
    {
      id: "2",
      title: "Monthly Staff & Departmental Review",
      date: "October 7, 2026",
    },
    {
      id: "3",
      title: "Pharmacy & Blood Bank Inventory Audit",
      date: "October 8, 2026",
    },
  ]);

  const [enquiries, setEnquiries] = useState<
    Array<{ id: string; name: string; message: string }>
  >([]);

  useEffect(() => {
    // Fetch live dashboard metrics from database APIs
    Promise.allSettled([
      fetch("/api/hms/patients"),
      fetch("/api/hms/doctors"),
      fetch("/api/hms/beds"),
      fetch("/api/hms/invoices"),
      fetch("/api/staff"),
      fetch("/api/hms/notices"),
      fetch("/api/hms/enquiries"),
    ])
      .then(
        async ([resPat, resDoc, resBeds, resInv, resStaff, resNot, resEnq]) => {
          const next = { ...counts };

          if (resPat.status === "fulfilled" && resPat.value.ok) {
            const data = await resPat.value.json();
            if (Array.isArray(data.patients))
              next.patients = String(data.patients.length);
            else if (data.total !== undefined)
              next.patients = String(data.total);
          }

          if (resDoc.status === "fulfilled" && resDoc.value.ok) {
            const data = await resDoc.value.json();
            if (Array.isArray(data.doctors))
              next.doctors = String(data.doctors.length);
          }

          if (resBeds.status === "fulfilled" && resBeds.value.ok) {
            const data = await resBeds.value.json();
            if (Array.isArray(data.beds)) {
              const avail = data.beds.filter((b: any) => b.available).length;
              next.availableBeds = String(avail);
            }
          }

          if (resInv.status === "fulfilled" && resInv.value.ok) {
            const data = await resInv.value.json();
            if (Array.isArray(data.invoices)) {
              const totalMinor = data.invoices.reduce(
                (acc: number, inv: any) => acc + (inv.totalMinor || 0),
                0,
              );
              next.invoices = (totalMinor / 100).toLocaleString();
            }
          }

          if (resStaff.status === "fulfilled" && resStaff.value.ok) {
            const data = await resStaff.value.json();
            if (Array.isArray(data.users)) {
              const users = data.users;
              next.admins = String(
                users.filter((u: any) => u.role === "admin").length || 1,
              );
              next.nurses = String(
                users.filter((u: any) => u.role === "nurse").length,
              );
              next.accountants = String(
                users.filter((u: any) => u.role === "accountant").length,
              );
              next.labTechnicians = String(
                users.filter((u: any) => u.role === "lab_technician").length,
              );
              next.pharmacists = String(
                users.filter((u: any) => u.role === "pharmacist").length,
              );
              next.receptionists = String(
                users.filter((u: any) => u.role === "receptionist").length,
              );
            }
          }

          if (resNot.status === "fulfilled" && resNot.value.ok) {
            const data = await resNot.value.json();
            const list = data.notices || data.notice_boards;
            if (Array.isArray(list) && list.length > 0) {
              setNotices(
                list.slice(0, 4).map((n: any) => ({
                  id: n.id,
                  title: n.title,
                  date: n.created_at
                    ? new Date(n.created_at).toLocaleDateString()
                    : "October 2026",
                })),
              );
            }
          }

          if (resEnq.status === "fulfilled" && resEnq.value.ok) {
            const data = await resEnq.value.json();
            if (Array.isArray(data.enquiries)) {
              setEnquiries(
                data.enquiries.slice(0, 3).map((e: any) => ({
                  id: e.id,
                  name: e.full_name || e.name || "Enquiry",
                  message: e.message || "",
                })),
              );
            }
          }

          setCounts(next);
        },
      )
      .catch(() => {});
  }, []);

  if (role !== "Admin") return <RoleDashboard />;

  const widgets = [
    ["Invoice Amount", counts.invoices, "invoices", CreditCard],
    ["Bill Amount", counts.bills, "bills", FileText],
    ["Payment Amount", counts.payments, "payments", Banknote],
    [
      "Advance Payment Amount",
      counts.advancePayments,
      "advanced-payments",
      Wallet,
    ],
    ["Available Beds", counts.availableBeds, "bed-status", BedDouble],
    ["Doctors", counts.doctors, "doctors", Stethoscope],
    ["Patients", counts.patients, "patients", Users],
    ["Nurses", counts.nurses, "nurses", ShieldPlus],
    ["Admins", counts.admins, "users", LockKeyhole],
    ["Accountants", counts.accountants, "accountants", BriefcaseMedical],
    ["Lab Technicians", counts.labTechnicians, "lab-technicians", FlaskConical],
    ["Pharmacists", counts.pharmacists, "pharmacists", Pill],
  ] as const;

  return (
    <section>
      <h1 className="sr-only">{t("Dashboard")}</h1>
      <div className="dashboard-widgets">
        {widgets.map(([title, value, slug, Icon]) => (
          <Link
            href={slug === "patients" ? "/patients" : `/modules/${slug}`}
            className="dashboard-widget"
            key={title}
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
            aria-label="Yearly income and expense chart"
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
              <strong>{counts.receptionists}</strong>
              <h3>{t("Receptionists")}</h3>
            </div>
          </Link>
          <div className="legacy-card">
            <h2>{t("Notice Boards")}</h2>
            {notices.map((n) => (
              <div className="notice-row" key={n.id}>
                <span className="notice-dot" />
                <div>
                  <h3>{t(n.title)}</h3>
                  <p>{n.date || "Active Notice"}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="dashboard-bottom">
        <div className="legacy-card">
          <div className="page-heading">
            <h2>{t("Enquiries")}</h2>
            <Link className="text-brand" href="/modules/enquiries">
              {t("View All")}
            </Link>
          </div>
          {enquiries.length === 0 ? (
            <p className="empty-enquiries">{t("No Enquiries Yet")}</p>
          ) : (
            <div className="space-y-2 pt-2">
              {enquiries.map((e) => (
                <div
                  key={e.id}
                  className="p-2.5 rounded-lg bg-muted/40 text-xs"
                >
                  <div className="font-semibold">{e.name}</div>
                  <div className="text-muted-foreground truncate">
                    {e.message}
                  </div>
                </div>
              ))}
            </div>
          )}
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
                {["Abebe Bekele", "Tigist Haile", "Dawit Wolde"].map((n, i) => (
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
                      {i + 6}, 2026 · 09:00 AM
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
