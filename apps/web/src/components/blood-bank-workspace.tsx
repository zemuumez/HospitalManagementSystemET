"use client";

import { useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  X,
  ChevronDown,
  Mail,
  Download,
  AlertCircle,
  Clock,
} from "lucide-react";
import { useLanguage } from "./language";

export type BloodBankTab =
  | "blood-banks"
  | "blood-donors"
  | "blood-donations"
  | "blood-issues"
  | "blood-donor-reports";

interface BloodBankItem {
  id: string;
  bloodGroup: string;
  remainedBags: number;
}

interface BloodDonorItem {
  id: string;
  name: string;
  email: string;
  age: number | string;
  gender: "Male" | "Female";
  bloodGroup: string;
  lastDonationDate: string;
  lastDonationTime: string;
}

interface BloodDonationItem {
  id: string;
  donorName: string;
  bags: number;
  date?: string;
}

interface BloodIssueItem {
  id: string;
  patientName: string;
  patientEmail: string;
  patientInitials: string;
  patientColor: string;
  doctorName: string;
  doctorEmail: string;
  doctorInitials: string;
  doctorColor: string;
  donorName: string;
  issueTime: string;
  issueDate: string;
  bloodGroup: string;
  amount: string;
}

interface DiseaseResult {
  id: string;
  name: string;
  result: "Negative" | "Positive";
}

interface BloodDonorReportItem {
  id: string;
  donorName: string;
  bloodGroup: string;
  reportDate: string;
  diseases: DiseaseResult[];
}

const INITIAL_BLOOD_BANKS: BloodBankItem[] = [
  { id: "bb-1", bloodGroup: "A+", remainedBags: 14 },
  { id: "bb-2", bloodGroup: "A-", remainedBags: 6 },
  { id: "bb-3", bloodGroup: "B+", remainedBags: 22 },
  { id: "bb-4", bloodGroup: "B-", remainedBags: 4 },
  { id: "bb-5", bloodGroup: "AB+", remainedBags: 9 },
  { id: "bb-6", bloodGroup: "AB-", remainedBags: 2 },
  { id: "bb-7", bloodGroup: "O+", remainedBags: 35 },
  { id: "bb-8", bloodGroup: "O-", remainedBags: 8 },
];

const INITIAL_DONORS: BloodDonorItem[] = [
  {
    id: "bd-1",
    name: "annie",
    email: "N/A",
    age: "N/A",
    gender: "Female",
    bloodGroup: "O+",
    lastDonationTime: "07:16 AM",
    lastDonationDate: "9th Apr,2026",
  },
  {
    id: "bd-2",
    name: "kumar",
    email: "N/A",
    age: "N/A",
    gender: "Male",
    bloodGroup: "B",
    lastDonationTime: "12:00 AM",
    lastDonationDate: "14th Mar,2026",
  },
  {
    id: "bd-3",
    name: "Vinoth",
    email: "N/A",
    age: "N/A",
    gender: "Male",
    bloodGroup: "B",
    lastDonationTime: "12:00 AM",
    lastDonationDate: "14th Mar,2026",
  },
  {
    id: "bd-4",
    name: "Muhammad Haseeb",
    email: "N/A",
    age: "N/A",
    gender: "Male",
    bloodGroup: "B+",
    lastDonationTime: "06:02 PM",
    lastDonationDate: "29th Dec,2025",
  },
  {
    id: "bd-5",
    name: "Deenadhayalan",
    email: "N/A",
    age: "N/A",
    gender: "Male",
    bloodGroup: "A+",
    lastDonationTime: "12:00 AM",
    lastDonationDate: "7th Oct,2025",
  },
  {
    id: "bd-6",
    name: "Abhishek saini",
    email: "N/A",
    age: "N/A",
    gender: "Male",
    bloodGroup: "A+ b+",
    lastDonationTime: "08:48 PM",
    lastDonationDate: "22nd Nov,2025",
  },
  {
    id: "bd-7",
    name: "Ahmed",
    email: "N/A",
    age: "N/A",
    gender: "Male",
    bloodGroup: "0 +ve",
    lastDonationTime: "07:25 AM",
    lastDonationDate: "7th Oct,2025",
  },
  {
    id: "bd-8",
    name: "jskjf",
    email: "N/A",
    age: "N/A",
    gender: "Female",
    bloodGroup: "A-",
    lastDonationTime: "12:00 AM",
    lastDonationDate: "26th Mar,2025",
  },
  {
    id: "bd-9",
    name: "Rohit",
    email: "N/A",
    age: 23,
    gender: "Male",
    bloodGroup: "AB",
    lastDonationTime: "01:29 PM",
    lastDonationDate: "9th Jan,2025",
  },
  {
    id: "bd-10",
    name: "cake bake",
    email: "N/A",
    age: 23,
    gender: "Male",
    bloodGroup: "AB",
    lastDonationTime: "11:54 AM",
    lastDonationDate: "9th Jan,2025",
  },
];

const INITIAL_DONATIONS: BloodDonationItem[] = [
  { id: "bdo-1", donorName: "annie", bags: 1, date: "9th Apr,2026" },
  { id: "bdo-2", donorName: "kumar", bags: 2, date: "14th Mar,2026" },
  { id: "bdo-3", donorName: "Vinoth", bags: 1, date: "14th Mar,2026" },
  { id: "bdo-4", donorName: "Muhammad Haseeb", bags: 1, date: "29th Dec,2025" },
  { id: "bdo-5", donorName: "Deenadhayalan", bags: 2, date: "7th Oct,2025" },
  { id: "bdo-6", donorName: "Abhishek saini", bags: 1, date: "22nd Nov,2025" },
];

const INITIAL_ISSUES: BloodIssueItem[] = [
  {
    id: "bi-1",
    patientName: "ABBA ADAMU",
    patientEmail: "adamuabba9@gmail.com",
    patientInitials: "AA",
    patientColor: "#3b82f6",
    doctorName: "Annie Bsseor",
    doctorEmail: "admin@hmqqs.com",
    doctorInitials: "AB",
    doctorColor: "#60a5fa",
    donorName: "Vinoth",
    issueTime: "11:00 AM",
    issueDate: "4th Aug,2026",
    bloodGroup: "B",
    amount: "$599.00",
  },
  {
    id: "bi-2",
    patientName: "111111 111111",
    patientEmail: "111111@qwe.com",
    patientInitials: "11",
    patientColor: "#10b981",
    doctorName: "123456 123456",
    doctorEmail: "123456@gmail.com",
    doctorInitials: "11",
    doctorColor: "#10b981",
    donorName: "Abhishek saini",
    issueTime: "09:44 AM",
    issueDate: "10th Jul,2026",
    bloodGroup: "A+ b+",
    amount: "$33.00",
  },
  {
    id: "bi-3",
    patientName: "11111 1111",
    patientEmail: "sulapojim@mailinator.com",
    patientInitials: "11",
    patientColor: "#14b8a6",
    doctorName: "123456 123456",
    doctorEmail: "123456@gmail.com",
    doctorInitials: "11",
    doctorColor: "#14b8a6",
    donorName: "Vinoth",
    issueTime: "03:46 PM",
    issueDate: "8th Jun,2026",
    bloodGroup: "B",
    amount: "$60,000.00",
  },
  {
    id: "bi-4",
    patientName: "AA Ahmed",
    patientEmail: "hostmileso@gmail.com",
    patientInitials: "AA",
    patientColor: "#f59e0b",
    doctorName: "AARAV Singh",
    doctorEmail: "aarav@gmail.com",
    doctorInitials: "AS",
    doctorColor: "#3b82f6",
    donorName: "annie",
    issueTime: "12:00 PM",
    issueDate: "9th Apr,2026",
    bloodGroup: "O+",
    amount: "$3.00",
  },
  {
    id: "bi-5",
    patientName: "Aamer Idris",
    patientEmail: "aamer.idris91@gmail.com",
    patientInitials: "AI",
    patientColor: "#06b6d4",
    doctorName: "AAAAA BBBBB",
    doctorEmail: "abab8764350@gmail.com",
    doctorInitials: "AB",
    doctorColor: "#10b981",
    donorName: "Vinoth",
    issueTime: "07:01 PM",
    issueDate: "10th Mar,2026",
    bloodGroup: "B",
    amount: "$100.00",
  },
  {
    id: "bi-6",
    patientName: "Abeer Saleem",
    patientEmail: "mohammadabeer3@gmail.com",
    patientInitials: "AS",
    patientColor: "#6366f1",
    doctorName: "123456 123456",
    doctorEmail: "123456@gmail.com",
    doctorInitials: "11",
    doctorColor: "#10b981",
    donorName: "Vinoth",
    issueTime: "05:30 PM",
    issueDate: "13th Mar,2026",
    bloodGroup: "B",
    amount: "$100.00",
  },
  {
    id: "bi-7",
    patientName: "Kiro Karam",
    patientEmail: "kiro.karam@gmail.com",
    patientInitials: "KK",
    patientColor: "#0284c7",
    doctorName: "AAAAA BBBBB",
    doctorEmail: "abab8764350@gmail.com",
    doctorInitials: "AB",
    doctorColor: "#10b981",
    donorName: "abc",
    issueTime: "12:41 PM",
    issueDate: "12th Mar,2026",
    bloodGroup: "0 +ve",
    amount: "$343.00",
  },
  {
    id: "bi-8",
    patientName: "Thagira S",
    patientEmail: "thagira12345@gmail.com",
    patientInitials: "TS",
    patientColor: "#ef4444",
    doctorName: "123456 123456",
    doctorEmail: "123456@gmail.com",
    doctorInitials: "11",
    doctorColor: "#10b981",
    donorName: "Abhishek saini",
    issueTime: "05:30 PM",
    issueDate: "2nd Feb,2026",
    bloodGroup: "A+ b+",
    amount: "$500.00",
  },
  {
    id: "bi-9",
    patientName: "Abdul Basheer",
    patientEmail: "basheervettupara@gmail.com",
    patientInitials: "AB",
    patientColor: "#3b82f6",
    doctorName: "123456 123456",
    doctorEmail: "123456@gmail.com",
    doctorInitials: "11",
    doctorColor: "#10b981",
    donorName: "Ahmed",
    issueTime: "11:41 PM",
    issueDate: "4th Feb,2026",
    bloodGroup: "0 +ve",
    amount: "$100.00",
  },
  {
    id: "bi-10",
    patientName: "Abdulatif Akromov",
    patientEmail: "abdulatif@gmail.com",
    patientInitials: "AA",
    patientColor: "#2563eb",
    doctorName: "AARAV Singh",
    doctorEmail: "aarav@gmail.com",
    doctorInitials: "AS",
    doctorColor: "#3b82f6",
    donorName: "Abhishek saini",
    issueTime: "07:13 PM",
    issueDate: "23rd Jan,2026",
    bloodGroup: "A+ b+",
    amount: "$400.00",
  },
];

const INITIAL_REPORTS: BloodDonorReportItem[] = [
  {
    id: "bdr-1",
    donorName: "Jenil Savani",
    bloodGroup: "O-",
    reportDate: "04 Aug 2026",
    diseases: [
      { id: "d-1", name: "Malaria", result: "Negative" },
      { id: "d-2", name: "Taifoide", result: "Negative" },
    ],
  },
  {
    id: "bdr-2",
    donorName: "Vinoth",
    bloodGroup: "B",
    reportDate: "04 Aug 2026",
    diseases: [
      { id: "d-3", name: "Hepatitis", result: "Negative" },
      { id: "d-4", name: "Malaria", result: "Positive" },
    ],
  },
];

export function BloodBankWorkspace({ id }: { id: string }) {
  const router = useRouter();
  const { t } = useLanguage();

  const activeTab: BloodBankTab = (
    [
      "blood-banks",
      "blood-donors",
      "blood-donations",
      "blood-issues",
      "blood-donor-reports",
    ].includes(id)
      ? id
      : "blood-banks"
  ) as BloodBankTab;

  // Search & Pagination
  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);

  // Data states
  const [bloodBanks, setBloodBanks] =
    useState<BloodBankItem[]>(INITIAL_BLOOD_BANKS);
  const [donors, setDonors] = useState<BloodDonorItem[]>(INITIAL_DONORS);
  const [donations, setDonations] =
    useState<BloodDonationItem[]>(INITIAL_DONATIONS);
  const [issues, setIssues] = useState<BloodIssueItem[]>(INITIAL_ISSUES);
  const [reports, setReports] =
    useState<BloodDonorReportItem[]>(INITIAL_REPORTS);

  // Modals
  const [showAddDonor, setShowAddDonor] = useState(false);
  const [showAddReport, setShowAddReport] = useState(false);
  const [showAddBank, setShowAddBank] = useState(false);
  const [showAddDonation, setShowAddDonation] = useState(false);
  const [showAddIssue, setShowAddIssue] = useState(false);

  // Delete modal
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    type: string;
    name: string;
  } | null>(null);

  // New Blood Donor form state
  const [donorForm, setDonorForm] = useState({
    name: "",
    email: "",
    age: "",
    gender: "Male" as "Male" | "Female",
    bloodGroup: "O+",
    lastDonationDate: "",
  });

  // Add Blood Donor Report form state
  const [reportForm, setReportForm] = useState<{
    donorName: string;
    age: string;
    gender: string;
    bloodGroup: string;
    lastDonationDate: string;
    diseases: { name: string; result: "Negative" | "Positive" }[];
  }>({
    donorName: "",
    age: "",
    gender: "",
    bloodGroup: "",
    lastDonationDate: "",
    diseases: [{ name: "", result: "Negative" }],
  });

  // Handle donor selection in Add Report modal
  const handleSelectDonorForReport = (donorName: string) => {
    const found = donors.find((d) => d.name === donorName);
    if (found) {
      setReportForm({
        ...reportForm,
        donorName: found.name,
        age: String(found.age || "N/A"),
        gender: found.gender,
        bloodGroup: found.bloodGroup,
        lastDonationDate: `${found.lastDonationTime} ${found.lastDonationDate}`,
      });
    } else {
      setReportForm({
        ...reportForm,
        donorName,
        age: "",
        gender: "",
        bloodGroup: "",
        lastDonationDate: "",
      });
    }
  };

  const addDiseaseRow = () => {
    setReportForm({
      ...reportForm,
      diseases: [...reportForm.diseases, { name: "", result: "Negative" }],
    });
  };

  const removeDiseaseRow = (index: number) => {
    setReportForm({
      ...reportForm,
      diseases: reportForm.diseases.filter((_, i) => i !== index),
    });
  };

  const handleSaveDonor = (e: React.FormEvent) => {
    e.preventDefault();
    if (!donorForm.name) return;
    const newDonor: BloodDonorItem = {
      id: `bd-${Date.now()}`,
      name: donorForm.name,
      email: donorForm.email || "N/A",
      age: donorForm.age || "N/A",
      gender: donorForm.gender,
      bloodGroup: donorForm.bloodGroup,
      lastDonationTime: "12:00 PM",
      lastDonationDate: donorForm.lastDonationDate || "Today",
    };
    setDonors([newDonor, ...donors]);
    setShowAddDonor(false);
    setDonorForm({
      name: "",
      email: "",
      age: "",
      gender: "Male",
      bloodGroup: "O+",
      lastDonationDate: "",
    });
  };

  const handleSaveReport = (e: React.FormEvent) => {
    e.preventDefault();
    if (!reportForm.donorName) return;
    const newReport: BloodDonorReportItem = {
      id: `bdr-${Date.now()}`,
      donorName: reportForm.donorName,
      bloodGroup: reportForm.bloodGroup || "O+",
      reportDate: new Date().toLocaleDateString("en-GB", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      }),
      diseases: reportForm.diseases
        .filter((d) => d.name.trim())
        .map((d, i) => ({
          id: `d-${Date.now()}-${i}`,
          name: d.name,
          result: d.result,
        })),
    };
    setReports([newReport, ...reports]);
    setShowAddReport(false);
    setReportForm({
      donorName: "",
      age: "",
      gender: "",
      bloodGroup: "",
      lastDonationDate: "",
      diseases: [{ name: "", result: "Negative" }],
    });
  };

  const confirmDelete = () => {
    if (!deleteTarget) return;
    if (deleteTarget.type === "donor") {
      setDonors(donors.filter((d) => d.id !== deleteTarget.id));
    } else if (deleteTarget.type === "issue") {
      setIssues(issues.filter((i) => i.id !== deleteTarget.id));
    } else if (deleteTarget.type === "report") {
      setReports(reports.filter((r) => r.id !== deleteTarget.id));
    } else if (deleteTarget.type === "bank") {
      setBloodBanks(bloodBanks.filter((b) => b.id !== deleteTarget.id));
    } else if (deleteTarget.type === "donation") {
      setDonations(donations.filter((d) => d.id !== deleteTarget.id));
    }
    setDeleteTarget(null);
  };

  return (
    <div className="legacy-page-container" style={{ padding: "24px" }}>
      {/* Search & Actions Toolbar */}
      <div className="billing-toolbar">
        <div className="billing-search-box">
          <Search size={16} className="search-icon" />
          <input
            placeholder={t("Search")}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>

        <div className="billing-actions">
          {activeTab === "blood-banks" && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddBank(true)}
            >
              + {t("New Blood Group")}
            </button>
          )}

          {activeTab === "blood-donors" && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddDonor(true)}
            >
              + {t("New Blood Donor")}
            </button>
          )}

          {activeTab === "blood-donations" && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddDonation(true)}
            >
              + {t("New Blood Donation")}
            </button>
          )}

          {activeTab === "blood-issues" && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddIssue(true)}
            >
              + {t("New Blood Issue")}
            </button>
          )}

          {activeTab === "blood-donor-reports" && (
            <button
              className="btn-action-blue"
              onClick={() => setShowAddReport(true)}
            >
              + {t("New Blood Donor Report")}
            </button>
          )}
        </div>
      </div>

      {/* Main Table Card */}
      <div className="billing-card">
        {/* TAB 1: Blood Banks */}
        {activeTab === "blood-banks" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("BLOOD GROUP")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("REMAINED BAGS")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {bloodBanks
                  .filter((b) =>
                    b.bloodGroup.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((b) => (
                    <tr key={b.id}>
                      <td>
                        <span className="badge-blood-red">{b.bloodGroup}</span>
                      </td>
                      <td>
                        <span className="badge-blue-pill">
                          {b.remainedBags} Bags
                        </span>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-edit"
                            aria-label="Edit"
                            onClick={() => {
                              const newCount = prompt(
                                "Enter remained bags count:",
                                String(b.remainedBags),
                              );
                              if (newCount !== null) {
                                setBloodBanks(
                                  bloodBanks.map((item) =>
                                    item.id === b.id
                                      ? {
                                          ...item,
                                          remainedBags: Number(newCount) || 0,
                                        }
                                      : item,
                                  ),
                                );
                              }
                            }}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: b.id,
                                type: "bank",
                                name: b.bloodGroup,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 2: Blood Donors */}
        {activeTab === "blood-donors" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("NAME")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("EMAIL")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("AGE")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("GENDER")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("BLOOD GROUP")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("LAST DONATION DATE")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {donors
                  .filter(
                    (d) =>
                      d.name.toLowerCase().includes(search.toLowerCase()) ||
                      d.bloodGroup.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((d) => (
                    <tr key={d.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {d.name}
                      </td>
                      <td style={{ color: "#64748b" }}>{d.email}</td>
                      <td>
                        {d.age !== "N/A" ? (
                          <span className="badge-blue-pill">{d.age}</span>
                        ) : (
                          <span style={{ color: "#64748b" }}>N/A</span>
                        )}
                      </td>
                      <td>
                        <span
                          style={{
                            color: d.gender === "Male" ? "#10b981" : "#ec4899",
                            fontWeight: 500,
                          }}
                        >
                          {d.gender}
                        </span>
                      </td>
                      <td>
                        <span className="badge-blood-red">{d.bloodGroup}</span>
                      </td>
                      <td>
                        <div className="tx-date-badge">
                          <span
                            className="tx-time"
                            style={{ color: "#29b6f6" }}
                          >
                            {d.lastDonationTime}
                          </span>
                          <span
                            className="tx-date"
                            style={{ color: "#29b6f6" }}
                          >
                            {d.lastDonationDate}
                          </span>
                        </div>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-edit"
                            aria-label="Edit"
                            onClick={() => {
                              setDonorForm({
                                name: d.name,
                                email: d.email === "N/A" ? "" : d.email,
                                age: String(d.age === "N/A" ? "" : d.age),
                                gender: d.gender,
                                bloodGroup: d.bloodGroup,
                                lastDonationDate: d.lastDonationDate,
                              });
                              setShowAddDonor(true);
                            }}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: d.id,
                                type: "donor",
                                name: d.name,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 3: Blood Donations */}
        {activeTab === "blood-donations" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("DONOR NAME")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("BAGS")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {donations
                  .filter((d) =>
                    d.donorName.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((d) => (
                    <tr key={d.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {d.donorName}
                      </td>
                      <td>
                        <span className="badge-blue-pill">{d.bags} Bags</span>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: d.id,
                                type: "donation",
                                name: `${d.donorName} donation`,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 4: Blood Issues */}
        {activeTab === "blood-issues" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("PATIENT")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("DOCTOR")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("DONOR NAME")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("ISSUE DATE")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("BLOOD GROUP")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("AMOUNT")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {issues
                  .filter(
                    (i) =>
                      i.patientName
                        .toLowerCase()
                        .includes(search.toLowerCase()) ||
                      i.doctorName
                        .toLowerCase()
                        .includes(search.toLowerCase()) ||
                      i.donorName.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((issue) => (
                    <tr key={issue.id}>
                      <td>
                        <div className="patient-cell">
                          <div
                            className="avatar-circle"
                            style={{ background: issue.patientColor }}
                          >
                            {issue.patientInitials}
                          </div>
                          <div className="patient-info">
                            <span className="link-cyan">
                              {issue.patientName}
                            </span>
                            <span className="patient-email">
                              {issue.patientEmail}
                            </span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <div
                            className="avatar-circle"
                            style={{ background: issue.doctorColor }}
                          >
                            {issue.doctorInitials}
                          </div>
                          <div className="patient-info">
                            <span className="link-cyan">
                              {issue.doctorName}
                            </span>
                            <span className="patient-email">
                              {issue.doctorEmail}
                            </span>
                          </div>
                        </div>
                      </td>
                      <td style={{ color: "#cbd5e1" }}>{issue.donorName}</td>
                      <td>
                        <div className="tx-date-badge">
                          <span
                            className="tx-time"
                            style={{ color: "#29b6f6" }}
                          >
                            {issue.issueTime}
                          </span>
                          <span
                            className="tx-date"
                            style={{ color: "#29b6f6" }}
                          >
                            {issue.issueDate}
                          </span>
                        </div>
                      </td>
                      <td>
                        <span className="badge-blood-red">
                          {issue.bloodGroup}
                        </span>
                      </td>
                      <td className="amount-text">{issue.amount}</td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-edit"
                            aria-label="Edit"
                            onClick={() => {
                              const newAmt = prompt(
                                "Enter new amount:",
                                issue.amount,
                              );
                              if (newAmt) {
                                setIssues(
                                  issues.map((it) =>
                                    it.id === issue.id
                                      ? { ...it, amount: newAmt }
                                      : it,
                                  ),
                                );
                              }
                            }}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="action-btn-delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: issue.id,
                                type: "issue",
                                name: `Issue for ${issue.patientName}`,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* TAB 5: Blood Donor Report */}
        {activeTab === "blood-donor-reports" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("DONOR NAME")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("BLOOD GROUP")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("REPORT DATE")} ↕</span>
                  </th>
                  <th>{t("DISEASES")}</th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {reports
                  .filter((r) =>
                    r.donorName.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((r) => (
                    <tr key={r.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {r.donorName}
                      </td>
                      <td>
                        <span style={{ color: "#cbd5e1" }}>{r.bloodGroup}</span>
                      </td>
                      <td style={{ color: "#cbd5e1" }}>{r.reportDate}</td>
                      <td>
                        <div
                          style={{
                            display: "flex",
                            flexWrap: "wrap",
                            gap: "6px",
                          }}
                        >
                          {r.diseases.map((dis) => (
                            <span
                              key={dis.id}
                              className={
                                dis.result === "Negative"
                                  ? "badge-disease-neg"
                                  : "badge-disease-pos"
                              }
                            >
                              {dis.name}: {dis.result}
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
                            className="action-btn-mail"
                            title="Send Email"
                            aria-label="Send Email"
                            onClick={() =>
                              alert(
                                `Blood donor report for ${r.donorName} sent via email!`,
                              )
                            }
                          >
                            <Mail size={16} />
                          </button>
                          <button
                            className="action-btn-download"
                            title="Download PDF"
                            aria-label="Download PDF"
                            onClick={() =>
                              alert(`Downloading report for ${r.donorName}...`)
                            }
                          >
                            <Download size={16} />
                          </button>
                          <button
                            className="action-btn-edit"
                            title="Edit"
                            aria-label="Edit"
                            onClick={() => {
                              handleSelectDonorForReport(r.donorName);
                              setShowAddReport(true);
                            }}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="action-btn-delete"
                            title="Delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: r.id,
                                type: "report",
                                name: `Report for ${r.donorName}`,
                              })
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Footer / Pagination */}
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
              {activeTab === "blood-banks"
                ? bloodBanks.length
                : activeTab === "blood-donors"
                  ? donors.length
                  : activeTab === "blood-donations"
                    ? donations.length
                    : activeTab === "blood-issues"
                      ? issues.length
                      : reports.length}{" "}
              {t("Results")}
            </span>
          </div>

          <div className="billing-pagination-controls">
            <button className="billing-page-btn" disabled>
              ‹
            </button>
            <button className="billing-page-btn is-active">1</button>
            <button className="billing-page-btn">2</button>
            <button className="billing-page-btn">3</button>
            <button className="billing-page-btn">›</button>
          </div>
        </div>
      </div>

      {/* MODAL: New Blood Donor */}
      {showAddDonor && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "560px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("New Blood Donor")}</h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowAddDonor(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveDonor} className="modal-body-custom">
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Name")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder={t("Name")}
                  className="form-input-custom"
                  value={donorForm.name}
                  onChange={(e) =>
                    setDonorForm({ ...donorForm, name: e.target.value })
                  }
                />
              </div>

              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Email")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  type="email"
                  placeholder={t("Email")}
                  className="form-input-custom"
                  value={donorForm.email}
                  onChange={(e) =>
                    setDonorForm({ ...donorForm, email: e.target.value })
                  }
                />
              </div>

              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Age")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  type="number"
                  placeholder={t("Age")}
                  className="form-input-custom"
                  value={donorForm.age}
                  onChange={(e) =>
                    setDonorForm({ ...donorForm, age: e.target.value })
                  }
                />
              </div>

              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Gender")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <div style={{ display: "flex", gap: "20px", marginTop: "8px" }}>
                  <label
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "8px",
                      cursor: "pointer",
                      color: "#cbd5e1",
                    }}
                  >
                    <input
                      type="radio"
                      name="donorGender"
                      checked={donorForm.gender === "Male"}
                      onChange={() =>
                        setDonorForm({ ...donorForm, gender: "Male" })
                      }
                    />
                    {t("Male")}
                  </label>
                  <label
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "8px",
                      cursor: "pointer",
                      color: "#cbd5e1",
                    }}
                  >
                    <input
                      type="radio"
                      name="donorGender"
                      checked={donorForm.gender === "Female"}
                      onChange={() =>
                        setDonorForm({ ...donorForm, gender: "Female" })
                      }
                    />
                    {t("Female")}
                  </label>
                </div>
              </div>

              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Blood Group")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select
                  className="form-select-custom"
                  value={donorForm.bloodGroup}
                  onChange={(e) =>
                    setDonorForm({ ...donorForm, bloodGroup: e.target.value })
                  }
                >
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

              <div
                className="form-group-custom"
                style={{ marginBottom: "24px" }}
              >
                <label className="form-label-custom">
                  {t("Last Donation Date")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  type="date"
                  className="form-input-custom"
                  value={donorForm.lastDonationDate}
                  onChange={(e) =>
                    setDonorForm({
                      ...donorForm,
                      lastDonationDate: e.target.value,
                    })
                  }
                />
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button type="submit" className="btn-action-blue">
                  {t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowAddDonor(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: Add Blood Donor Report */}
      {showAddReport && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "680px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("Add Blood Donor Report")}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowAddReport(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveReport} className="modal-body-custom">
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Donor Name")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select
                  required
                  className="form-select-custom"
                  value={reportForm.donorName}
                  onChange={(e) => handleSelectDonorForReport(e.target.value)}
                >
                  <option value="">-- {t("Select Donor")} --</option>
                  {donors.map((d) => (
                    <option key={d.id} value={d.name}>
                      {d.name} ({d.bloodGroup})
                    </option>
                  ))}
                </select>
              </div>

              {/* 4 autofill boxes */}
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "repeat(4, 1fr)",
                  gap: "12px",
                  marginBottom: "18px",
                }}
              >
                <div>
                  <label className="form-label-custom">{t("Age")}:</label>
                  <input
                    disabled
                    readOnly
                    className="form-input-custom"
                    placeholder="Age"
                    value={reportForm.age}
                  />
                </div>
                <div>
                  <label className="form-label-custom">{t("Gender")}:</label>
                  <input
                    disabled
                    readOnly
                    className="form-input-custom"
                    placeholder="Gender"
                    value={reportForm.gender}
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Blood Group")}:
                  </label>
                  <input
                    disabled
                    readOnly
                    className="form-input-custom"
                    placeholder="Blood"
                    value={reportForm.bloodGroup}
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Last Donation Date")}:
                  </label>
                  <input
                    disabled
                    readOnly
                    className="form-input-custom"
                    placeholder="N/A"
                    value={reportForm.lastDonationDate}
                  />
                </div>
              </div>

              {/* Diseases section */}
              <div style={{ marginBottom: "16px" }}>
                <label className="form-label-custom">
                  {t("Diseases")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>

                {reportForm.diseases.map((disease, idx) => (
                  <div
                    key={idx}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "10px",
                      marginBottom: "10px",
                    }}
                  >
                    <input
                      placeholder="Disease Name"
                      className="form-input-custom"
                      style={{ flex: 1 }}
                      value={disease.name}
                      onChange={(e) => {
                        const updated = [...reportForm.diseases];
                        updated[idx].name = e.target.value;
                        setReportForm({ ...reportForm, diseases: updated });
                      }}
                    />
                    <select
                      className="form-select-custom"
                      style={{ width: "200px" }}
                      value={disease.result}
                      onChange={(e) => {
                        const updated = [...reportForm.diseases];
                        updated[idx].result = e.target.value as
                          "Negative" | "Positive";
                        setReportForm({ ...reportForm, diseases: updated });
                      }}
                    >
                      <option value="Negative">Negative</option>
                      <option value="Positive">Positive</option>
                    </select>
                    {reportForm.diseases.length > 1 && (
                      <button
                        type="button"
                        className="action-btn-delete"
                        style={{
                          background: "#dc2626",
                          color: "#fff",
                          width: "36px",
                          height: "36px",
                          borderRadius: "6px",
                          display: "inline-flex",
                          alignItems: "center",
                          justifyContent: "center",
                        }}
                        onClick={() => removeDiseaseRow(idx)}
                      >
                        ✕
                      </button>
                    )}
                  </div>
                ))}

                <button
                  type="button"
                  style={{
                    background: "transparent",
                    border: "1px solid #5b73e8",
                    color: "#5b73e8",
                    padding: "6px 14px",
                    borderRadius: "6px",
                    fontSize: "13px",
                    cursor: "pointer",
                    marginTop: "6px",
                  }}
                  onClick={addDiseaseRow}
                >
                  + {t("Add Diseases")}
                </button>
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                  marginTop: "20px",
                }}
              >
                <button type="submit" className="btn-action-blue">
                  {t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowAddReport(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: New Blood Issue */}
      {showAddIssue && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "560px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("New Blood Issue")}</h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowAddIssue(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                setShowAddIssue(false);
              }}
              className="modal-body-custom"
            >
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Patient")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder="Patient Name"
                  className="form-input-custom"
                />
              </div>
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Doctor")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder="Doctor Name"
                  className="form-input-custom"
                />
              </div>
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Donor Name")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select className="form-select-custom">
                  {donors.map((d) => (
                    <option key={d.id} value={d.name}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </div>
              <div
                className="form-group-custom"
                style={{ marginBottom: "16px" }}
              >
                <label className="form-label-custom">
                  {t("Blood Group")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select className="form-select-custom">
                  <option value="A+">A+</option>
                  <option value="B+">B+</option>
                  <option value="O+">O+</option>
                  <option value="AB+">AB+</option>
                </select>
              </div>
              <div
                className="form-group-custom"
                style={{ marginBottom: "20px" }}
              >
                <label className="form-label-custom">{t("Amount")}:</label>
                <input defaultValue="$150.00" className="form-input-custom" />
              </div>
              <div
                style={{
                  display: "flex",
                  justifyContent: "flex-end",
                  gap: "12px",
                }}
              >
                <button type="submit" className="btn-action-blue">
                  {t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setShowAddIssue(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
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
                    {t("This action cannot be undone.")}
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
                  onClick={confirmDelete}
                >
                  {t("Delete")}
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
