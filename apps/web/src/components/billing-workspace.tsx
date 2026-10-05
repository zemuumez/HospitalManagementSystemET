"use client";
import React, { useEffect, useState, useMemo } from "react";
import Link from "next/link";
import {
  Search,
  Filter,
  Pencil,
  Trash2,
  Eye,
  FileSpreadsheet,
  FileDown,
  Plus,
  X,
  ArrowUpDown,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  AlertCircle,
  Loader2,
} from "lucide-react";
import { useLanguage } from "./language";
import { Modal } from "./modal";

export type BillingRow = {
  id: string;
  patient?: string;
  email?: string;
  initial?: string;
  avatarBg?: string;
  receiptNo?: string;
  invoiceId?: string;
  billId?: string;
  payrollId?: string;
  account?: string;
  payTo?: string;
  type?: "Credit" | "Debit";
  date?: string;
  txTime?: string;
  txDate?: string;
  amount?: string;
  status?: string;
  paymentStatus?: string;
  month?: string;
  year?: string;
  netSalary?: string;
  srNo?: number;
  description?: string;
  [key: string]: unknown;
};

const canonicalManualPayments: BillingRow[] = [
  {
    id: "mb-1",
    patient: "Trith Shah",
    email: "tirth@gmail.com",
    initial: "TS",
    avatarBg: "#6f42c1",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "01:18 PM",
    txDate: "21st March 2025",
    amount: "$15,600.00",
  },
  {
    id: "mb-2",
    patient: "Ttt 123",
    email: "ttt123@gmail.com",
    initial: "T1",
    avatarBg: "#fd7e14",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "10:30 AM",
    txDate: "6th January 2025",
    amount: "$3,000.00",
  },
  {
    id: "mb-3",
    patient: "Ttt 123",
    email: "ttt123@gmail.com",
    initial: "T1",
    avatarBg: "#fd7e14",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "10:09 AM",
    txDate: "6th January 2025",
    amount: "$27,944.00",
  },
  {
    id: "mb-4",
    patient: "Sanchita Sinha",
    email: "sanchitasinha@gmail.com",
    initial: "SS",
    avatarBg: "#f59e0b",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "06:15 AM",
    txDate: "6th January 2025",
    amount: "$2,000.00",
  },
  {
    id: "mb-5",
    patient: "Sanchita Sinha",
    email: "sanchitasinha@gmail.com",
    initial: "SS",
    avatarBg: "#f59e0b",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "08:01 AM",
    txDate: "3rd January 2025",
    amount: "$2,000.00",
  },
  {
    id: "mb-6",
    patient: "Ram Singh",
    email: "ramsingh@gmail.com",
    initial: "RS",
    avatarBg: "#6f42c1",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "01:16 PM",
    txDate: "31st December 2024",
    amount: "$2,000.00",
  },
  {
    id: "mb-7",
    patient: "Ram Singh",
    email: "ramsingh@gmail.com",
    initial: "RS",
    avatarBg: "#6f42c1",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "01:16 PM",
    txDate: "31st December 2024",
    amount: "$1,200.00",
  },
  {
    id: "mb-8",
    patient: "Trith Shah",
    email: "tirth@gmail.com",
    initial: "TS",
    avatarBg: "#6f42c1",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "07:54 PM",
    txDate: "26th December 2024",
    amount: "$31,444.00",
  },
  {
    id: "mb-9",
    patient: "Trith Shah",
    email: "tirth@gmail.com",
    initial: "TS",
    avatarBg: "#6f42c1",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "05:34 AM",
    txDate: "19th December 2024",
    amount: "$9,000.00",
  },
  {
    id: "mb-10",
    patient: "11111 1111",
    email: "sulapojim@mailinator.com",
    initial: "11",
    avatarBg: "#20c997",
    paymentStatus: "Approved",
    status: "Paid",
    txTime: "04:08 PM",
    txDate: "23rd October 2024",
    amount: "$26,544.00",
  },
];

const canonicalAdvancePayments: BillingRow[] = [
  {
    id: "ap-1",
    receiptNo: "X4MFUWHZ",
    patient: "Vinay Grover",
    email: "vinaygrover14365@gmail.com",
    initial: "VG",
    avatarBg: "#0284c7",
    date: "5th Aug, 2026",
    amount: "$55,000.00",
  },
  {
    id: "ap-2",
    receiptNo: "U57GUJP3",
    patient: "Vinay Grover",
    email: "vinaygrover14365@gmail.com",
    initial: "VG",
    avatarBg: "#0284c7",
    date: "5th Aug, 2026",
    amount: "$800.00",
  },
  {
    id: "ap-3",
    receiptNo: "WLHFJGGZ",
    patient: "123123 123123",
    email: "thisledome@hotmail.com",
    initial: "11",
    avatarBg: "#6f42c1",
    date: "25th Jul, 2026",
    amount: "$100.00",
  },
  {
    id: "ap-4",
    receiptNo: "IVUJNW75",
    patient: "Irfan Özer",
    email: "ozerr933@gmail.com",
    initial: "IÖ",
    avatarBg: "#38bdf8",
    date: "25th Jul, 2026",
    amount: "$100.00",
  },
  {
    id: "ap-5",
    receiptNo: "HL4NKFLA",
    patient: "123123 123123",
    email: "thisledome@hotmail.com",
    initial: "11",
    avatarBg: "#6f42c1",
    date: "17th Jul, 2026",
    amount: "N/A",
  },
  {
    id: "ap-6",
    receiptNo: "B33WBCQB",
    patient: "Test Test",
    email: "testsse2232545@test.com",
    initial: "TT",
    avatarBg: "#0284c7",
    date: "8th Jun, 2026",
    amount: "$100.00",
  },
  {
    id: "ap-7",
    receiptNo: "D4XESVMM",
    patient: "Abeer Saleem",
    email: "mohammadabeer3@gmail.com",
    initial: "AS",
    avatarBg: "#10b981",
    date: "19th Apr, 2026",
    amount: "$50,000.00",
  },
  {
    id: "ap-8",
    receiptNo: "OBV6LCAL",
    patient: "Vishwjeet Kumar",
    email: "vishwjeet@hms.com",
    initial: "VK",
    avatarBg: "#f43f5e",
    date: "16th Apr, 2026",
    amount: "$5,000.00",
  },
  {
    id: "ap-9",
    receiptNo: "BXWC5SKG",
    patient: "ANDRE SANTOS",
    email: "jfismaeldjfdj@gmail.com",
    initial: "AS",
    avatarBg: "#f43f5e",
    date: "17th Apr, 2026",
    amount: "$100.00",
  },
  {
    id: "ap-10",
    receiptNo: "XTID9WVE",
    patient: "Aamer Idris",
    email: "aamer.idris91@gmail.com",
    initial: "AI",
    avatarBg: "#14b8a6",
    date: "17th Apr, 2026",
    amount: "$500.00",
  },
];

const canonicalPaymentReports: BillingRow[] = [
  {
    id: "pr-1",
    date: "21st Sep, 2026",
    account: "2",
    payTo: "Pharmacy",
    type: "Credit",
    amount: "$5,000.00",
  },
  {
    id: "pr-2",
    date: "5th Aug, 2026",
    account: "Victor Stephen",
    payTo: "Vinay",
    type: "Debit",
    amount: "$880.00",
  },
  {
    id: "pr-3",
    date: "3rd Aug, 2026",
    account: "Johny",
    payTo: "kazi",
    type: "Credit",
    amount: "$500.00",
  },
  {
    id: "pr-4",
    date: "8th Jun, 2026",
    account: "Anisur Rahman",
    payTo: "IIIII",
    type: "Credit",
    amount: "$800.00",
  },
  {
    id: "pr-5",
    date: "29th Dec, 2025",
    account: "lab test",
    payTo: "Muhammad Haseeb",
    type: "Credit",
    amount: "$2,353.00",
  },
  {
    id: "pr-6",
    date: "13th Nov, 2025",
    account: "x-ray",
    payTo: "Hospital",
    type: "Credit",
    amount: "$35,000.00",
  },
  {
    id: "pr-7",
    date: "4th Sep, 2025",
    account: "HDFC",
    payTo: "cash",
    type: "Debit",
    amount: "$100.00",
  },
  {
    id: "pr-8",
    date: "21st May, 2025",
    account: "Mobile Bill",
    payTo: "mnk",
    type: "Credit",
    amount: "$25.00",
  },
  {
    id: "pr-9",
    date: "6th May, 2025",
    account: "MILK",
    payTo: "Trith Shah",
    type: "Credit",
    amount: "$500.00",
  },
  {
    id: "pr-10",
    date: "24th Apr, 2025",
    account: "H2Soft",
    payTo: "abba adamu",
    type: "Credit",
    amount: "$500.00",
  },
];

const canonicalPayments: BillingRow[] = [
  {
    id: "pm-1",
    account: "2",
    date: "21st Sep, 2026",
    payTo: "Pharmacy",
    amount: "$5,000.00",
  },
  {
    id: "pm-2",
    account: "Victor Stephen",
    date: "5th Aug, 2026",
    payTo: "Vinay",
    amount: "$880.00",
  },
  {
    id: "pm-3",
    account: "Johny",
    date: "3rd Aug, 2026",
    payTo: "kazi",
    amount: "$500.00",
  },
  {
    id: "pm-4",
    account: "Anisur Rahman",
    date: "8th Jun, 2026",
    payTo: "IIIII",
    amount: "$800.00",
  },
  {
    id: "pm-5",
    account: "lab test",
    date: "29th Dec, 2025",
    payTo: "Muhammad Haseeb",
    amount: "$2,353.00",
  },
  {
    id: "pm-6",
    account: "x-ray",
    date: "13th Nov, 2025",
    payTo: "Hospital",
    amount: "$35,000.00",
  },
  {
    id: "pm-7",
    account: "HDFC",
    date: "4th Sep, 2025",
    payTo: "cash",
    amount: "$100.00",
  },
  {
    id: "pm-8",
    account: "Mobile Bill",
    date: "21st May, 2025",
    payTo: "mnk",
    amount: "$25.00",
  },
  {
    id: "pm-9",
    account: "MILK",
    date: "6th May, 2025",
    payTo: "Trith Shah",
    amount: "$500.00",
  },
  {
    id: "pm-10",
    account: "H2Soft",
    date: "24th Apr, 2025",
    payTo: "abba adamu",
    amount: "$500.00",
  },
];

const canonicalInvoices: BillingRow[] = [
  {
    id: "inv-1",
    invoiceId: "HMS13",
    patient: "11111 1111",
    email: "sulapojim@mailinator.com",
    initial: "11",
    avatarBg: "#14b8a6",
    date: "26th Sep, 2026",
    amount: "$500.00",
    status: "Paid",
  },
  {
    id: "inv-2",
    invoiceId: "HMS12",
    patient: "Aa Aa",
    email: "avmedia04@gmail.com",
    initial: "AA",
    avatarBg: "#f59e0b",
    date: "20th Sep, 2026",
    amount: "$10,000.00",
    status: "Paid",
  },
  {
    id: "inv-3",
    invoiceId: "HMS11",
    patient: "Vinay Grover",
    email: "vinaygrover14365@gmail.com",
    initial: "VG",
    avatarBg: "#0284c7",
    date: "5th Aug, 2026",
    amount: "$750.00",
    status: "Paid",
  },
  {
    id: "inv-4",
    invoiceId: "HMS10",
    patient: "Minhaj Ali",
    email: "minhajali@gmail.com",
    initial: "MA",
    avatarBg: "#0284c7",
    date: "27th Jul, 2026",
    amount: "$800.00",
    status: "Paid",
  },
  {
    id: "inv-5",
    invoiceId: "HMS09",
    patient: "Irfan Özer",
    email: "ozerr933@gmail.com",
    initial: "IÖ",
    avatarBg: "#38bdf8",
    date: "25th Jul, 2026",
    amount: "$1.00",
    status: "Paid",
  },
  {
    id: "inv-6",
    invoiceId: "HMS08",
    patient: "Aaa Bbb",
    email: "abc1@joymali.com",
    initial: "AB",
    avatarBg: "#0284c7",
    date: "9th Jul, 2026",
    amount: "$5,460.00",
    status: "Paid",
  },
  {
    id: "inv-7",
    invoiceId: "HMS07",
    patient: "ABBA ADAMU",
    email: "adamuabba9@gmail.com",
    initial: "AA",
    avatarBg: "#0284c7",
    date: "8th Jun, 2026",
    amount: "$800.00",
    status: "Paid",
  },
  {
    id: "inv-8",
    invoiceId: "HMS06",
    patient: "ABBA ADAMU",
    email: "adamuabba9@gmail.com",
    initial: "AA",
    avatarBg: "#0284c7",
    date: "8th Jun, 2026",
    amount: "$550.00",
    status: "Pending",
  },
  {
    id: "inv-9",
    invoiceId: "HMS05",
    patient: "Malayali America",
    email: "sajascj@gmail.com",
    initial: "MA",
    avatarBg: "#0284c7",
    date: "6th Jun, 2026",
    amount: "$25,155.20",
    status: "Paid",
  },
  {
    id: "inv-10",
    invoiceId: "HMS04",
    patient: "Vishwjeet Kumar",
    email: "vishwjeet@hms.com",
    initial: "VK",
    avatarBg: "#f43f5e",
    date: "17th Apr, 2026",
    amount: "$1,800.00",
    status: "Pending",
  },
];

const canonicalAccounts: BillingRow[] = [
  {
    id: "acc-1",
    account: "Emergency Medicine",
    type: "Credit",
    status: "Active",
  },
  {
    id: "acc-2",
    account: "Cardiology Consumables",
    type: "Debit",
    status: "Active",
  },
  {
    id: "acc-3",
    account: "Pharmacy General",
    type: "Credit",
    status: "Active",
  },
  {
    id: "acc-4",
    account: "Laboratory Diagnostics",
    type: "Credit",
    status: "Active",
  },
  {
    id: "acc-5",
    account: "Hospital Utilities & Power",
    type: "Debit",
    status: "Active",
  },
  {
    id: "acc-6",
    account: "Medical Equipment Maintenance",
    type: "Debit",
    status: "Active",
  },
  {
    id: "acc-7",
    account: "Surgical Theatre Services",
    type: "Credit",
    status: "Active",
  },
  {
    id: "acc-8",
    account: "Radiology & Imaging",
    type: "Credit",
    status: "Active",
  },
  {
    id: "acc-9",
    account: "General Administration",
    type: "Debit",
    status: "Active",
  },
  {
    id: "acc-10",
    account: "Ambulance Transportation",
    type: "Credit",
    status: "Active",
  },
];

const canonicalPayrolls: BillingRow[] = [
  {
    id: "pay-1",
    srNo: 1,
    payrollId: "#EMP1001",
    patient: "Dr. Sarah Johnson",
    email: "sarah.j@hospital.et",
    initial: "SJ",
    avatarBg: "#0284c7",
    month: "September",
    year: "2026",
    netSalary: "$4,500.00",
    status: "Paid",
  },
  {
    id: "pay-2",
    srNo: 2,
    payrollId: "#EMP1002",
    patient: "Nurse Martha Bekele",
    email: "martha.b@hospital.et",
    initial: "MB",
    avatarBg: "#10b981",
    month: "September",
    year: "2026",
    netSalary: "$2,200.00",
    status: "Paid",
  },
  {
    id: "pay-3",
    srNo: 3,
    payrollId: "#EMP1003",
    patient: "Dr. Kebede Tadesse",
    email: "kebede.t@hospital.et",
    initial: "KT",
    avatarBg: "#6f42c1",
    month: "September",
    year: "2026",
    netSalary: "$5,100.00",
    status: "Paid",
  },
  {
    id: "pay-4",
    srNo: 4,
    payrollId: "#EMP1004",
    patient: "Pharmacist Daniel Zewdu",
    email: "daniel.z@hospital.et",
    initial: "DZ",
    avatarBg: "#f59e0b",
    month: "September",
    year: "2026",
    netSalary: "$2,800.00",
    status: "Paid",
  },
  {
    id: "pay-5",
    srNo: 5,
    payrollId: "#EMP1005",
    patient: "Accountant Helen Girma",
    email: "helen.g@hospital.et",
    initial: "HG",
    avatarBg: "#38bdf8",
    month: "September",
    year: "2026",
    netSalary: "$3,000.00",
    status: "Paid",
  },
];

const canonicalBills: BillingRow[] = [
  {
    id: "bl-1",
    billId: "BL1001",
    patient: "Vinay Grover",
    email: "vinaygrover14365@gmail.com",
    initial: "VG",
    avatarBg: "#0284c7",
    date: "5th Aug, 2026",
    amount: "$1,250.00",
    status: "Paid",
  },
  {
    id: "bl-2",
    billId: "BL1002",
    patient: "123123 123123",
    email: "thisledome@hotmail.com",
    initial: "11",
    avatarBg: "#6f42c1",
    date: "25th Jul, 2026",
    amount: "$420.00",
    status: "Paid",
  },
  {
    id: "bl-3",
    billId: "BL1003",
    patient: "Irfan Özer",
    email: "ozerr933@gmail.com",
    initial: "IÖ",
    avatarBg: "#38bdf8",
    date: "25th Jul, 2026",
    amount: "$850.00",
    status: "Unpaid",
  },
  {
    id: "bl-4",
    billId: "BL1004",
    patient: "Abeer Saleem",
    email: "mohammadabeer3@gmail.com",
    initial: "AS",
    avatarBg: "#10b981",
    date: "19th Apr, 2026",
    amount: "$3,100.00",
    status: "Paid",
  },
];

const initialDatasets: Record<string, BillingRow[]> = {
  "manual-billing-payments": canonicalManualPayments,
  "manual-bill-payments": canonicalManualPayments,
  "advance-payments": canonicalAdvancePayments,
  "advanced-payments": canonicalAdvancePayments,
  "payment-reports": canonicalPaymentReports,
  payments: canonicalPayments,
  invoices: canonicalInvoices,
  accounts: canonicalAccounts,
  "employee-payrolls": canonicalPayrolls,
  bills: canonicalBills,
};

const totalCounts: Record<string, number> = {
  "manual-billing-payments": 34,
  "manual-bill-payments": 34,
  "advance-payments": 150,
  "advanced-payments": 150,
  "payment-reports": 84,
  payments: 84,
  invoices: 323,
  accounts: 10,
  "employee-payrolls": 5,
  bills: 4,
};

export function BillingWorkspace({ id }: { id: string }) {
  const { t } = useLanguage();
  const canonicalSlug = useMemo(() => {
    if (id === "manual-bill-payments") return "manual-billing-payments";
    if (id === "advanced-payments") return "advance-payments";
    return id;
  }, [id]);

  const [data, setData] =
    useState<Record<string, BillingRow[]>>(initialDatasets);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);
  const [filterOpen, setFilterOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [editingRow, setEditingRow] = useState<BillingRow | null>(null);
  const [viewRow, setViewRow] = useState<BillingRow | null>(null);
  const [statusFilter, setStatusFilter] = useState("All");

  // API Integration state
  const [apiConnected, setApiConnected] = useState(false);
  const [isLoadingApi, setIsLoadingApi] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");
  const [remotePatients, setRemotePatients] = useState<
    Array<{ id: string; name: string }>
  >([]);
  const [remoteAccounts, setRemoteAccounts] = useState<
    Array<{ id: string; name: string }>
  >([]);

  async function loadBillingData() {
    setIsLoadingApi(true);
    let connected = false;
    try {
      const [invRes, accRes, payRes, patRes] = await Promise.all([
        fetch("/api/hms/invoices"),
        fetch("/api/hms/charge-accounts"),
        fetch("/api/hms/payrolls"),
        fetch("/api/hms/billing-patients"),
      ]);

      if (invRes.ok) {
        const invData = await invRes.json();
        if (Array.isArray(invData.invoices) && invData.invoices.length > 0) {
          const mappedInvoices: BillingRow[] = invData.invoices.map((inv: any) => ({
            id: inv.id,
            invoiceId: `HMS${inv.number || inv.id.slice(0, 4)}`,
            patient: inv.patientName || "Patient",
            email: "patient@hospital.et",
            initial: (inv.patientName || "PT").slice(0, 2).toUpperCase(),
            avatarBg: "#14b8a6",
            date: inv.invoiceDate || "5th Oct, 2026",
            amount: `$${((inv.totalMinor || 0) / 100).toFixed(2)}`,
            status: (inv.paidMinor || 0) >= (inv.totalMinor || 0) ? "Paid" : "Pending",
          }));
          setData((prev) => ({
            ...prev,
            invoices: mappedInvoices,
            bills: mappedInvoices.map((inv) => ({
              ...inv,
              billId: `BL${inv.invoiceId?.replace("HMS", "") || "1001"}`,
            })),
          }));
        }
        connected = true;
      }

      if (accRes.ok) {
        const accData = await accRes.json();
        if (Array.isArray(accData.accounts) && accData.accounts.length > 0) {
          setRemoteAccounts(accData.accounts);
          const mappedAccounts: BillingRow[] = accData.accounts.map((acc: any) => ({
            id: acc.id,
            account: acc.name,
            type: "Credit",
            status: "Active",
          }));
          setData((prev) => ({ ...prev, accounts: mappedAccounts }));
        }
        connected = true;
      }

      if (payRes.ok) {
        const payData = await payRes.json();
        if (Array.isArray(payData.payrolls) && payData.payrolls.length > 0) {
          const mappedPayrolls: BillingRow[] = payData.payrolls.map((p: any, idx: number) => ({
            id: p.id,
            srNo: idx + 1,
            payrollId: `#EMP${p.id.slice(0, 4)}`,
            patient: p.staffName || "Staff Member",
            email: p.email || "staff@hospital.et",
            initial: (p.staffName || "ST").slice(0, 2).toUpperCase(),
            avatarBg: "#0284c7",
            month: p.month || "October",
            year: String(p.year || 2026),
            netSalary: `$${((p.netSalaryMinor || 0) / 100).toFixed(2)}`,
            status: p.status === 1 ? "Paid" : "Unpaid",
          }));
          setData((prev) => ({ ...prev, "employee-payrolls": mappedPayrolls }));
        }
        connected = true;
      }

      if (patRes.ok) {
        const patData = await patRes.json();
        if (Array.isArray(patData.patients)) {
          setRemotePatients(patData.patients);
        }
      }

      setApiConnected(connected);
    } catch {
      setApiConnected(false);
    } finally {
      setIsLoadingApi(false);
    }
  }

  useEffect(() => {
    loadBillingData();
  }, []);

  // Load persisted modifications from localStorage
  useEffect(() => {
    try {
      const saved = localStorage.getItem("hms-billing-workspace");
      if (saved) {
        const parsed = JSON.parse(saved);
        setData((prev) => ({ ...prev, ...parsed }));
      }
    } catch {}
    setReady(true);
  }, []);

  // Save changes
  useEffect(() => {
    if (!ready) return;
    try {
      localStorage.setItem("hms-billing-workspace", JSON.stringify(data));
    } catch {}
  }, [data, ready]);

  const currentList = useMemo(() => {
    return data[canonicalSlug] || initialDatasets[canonicalSlug] || [];
  }, [data, canonicalSlug]);

  const filteredList = useMemo(() => {
    return currentList.filter((row) => {
      const matchSearch =
        !search ||
        Object.values(row).some(
          (val) =>
            typeof val === "string" &&
            val.toLowerCase().includes(search.toLowerCase()),
        );
      const matchStatus =
        statusFilter === "All" ||
        row.status === statusFilter ||
        row.paymentStatus === statusFilter ||
        row.type === statusFilter;
      return matchSearch && matchStatus;
    });
  }, [currentList, search, statusFilter]);

  const total = totalCounts[canonicalSlug] || filteredList.length;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  function handleDelete(rowId: string) {
    if (window.confirm(t("Delete this record?"))) {
      setData((prev) => {
        const updated = {
          ...prev,
          [canonicalSlug]: (prev[canonicalSlug] || []).filter(
            (r) => r.id !== rowId,
          ),
        };
        return updated;
      });
    }
  }

  async function handleSaveRow(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const rowId = editingRow?.id || `new-${Date.now()}`;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    const patientName =
      (form.get("patient") as string) || editingRow?.patient || "New Patient";
    const amountStr =
      (form.get("amount") as string) || editingRow?.amount || "$0.00";
    const numericAmount = parseFloat(amountStr.replace(/[^0-9.]/g, "")) || 0;
    const amountMinor = Math.round(numericAmount * 100);

    // If connected to Go API:
    if (apiConnected) {
      if (canonicalSlug === "accounts") {
        const accName = (form.get("account") as string) || "General Account";
        try {
          const res = await fetch("/api/hms/charge-accounts", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: accName.trim() }),
          });
          if (res.ok) {
            setApiSuccessBanner(
              t(
                "Charge account successfully created in persistent PostgreSQL ledger",
              ),
            );
            loadBillingData();
            setEditingRow(null);
            setCreateOpen(false);
            setIsSubmitting(false);
            return;
          }
        } catch {}
      } else if (canonicalSlug === "invoices" || canonicalSlug === "bills") {
        const matchedPatient = remotePatients.find(
          (p) =>
            p.name.toLowerCase() === patientName.toLowerCase() ||
            p.id === patientName,
        );
        const patientId =
          matchedPatient?.id ||
          (remotePatients[0]?.id ?? "00000000-0000-0000-0000-000000000001");
        const accountId =
          remoteAccounts[0]?.id ?? "00000000-0000-0000-0000-000000000001";
        const todayStr = new Date().toISOString().slice(0, 10);

        try {
          const res = await fetch("/api/hms/invoices", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              patientId,
              invoiceDate: todayStr,
              discountBasisPoints: 0,
              lines: [
                {
                  accountId,
                  description: "Hospital Medical Service & Care",
                  quantity: 1,
                  unitPriceMinor: amountMinor > 0 ? amountMinor : 50000,
                },
              ],
            }),
          });
          if (res.ok) {
            setApiSuccessBanner(
              t("Invoice officially issued and locked in PostgreSQL ledger"),
            );
            loadBillingData();
            setEditingRow(null);
            setCreateOpen(false);
            setIsSubmitting(false);
            return;
          }
        } catch {}
      }
    }

    // Local fallback for smooth experience
    const newRecord: BillingRow = {
      id: rowId,
      patient: patientName,
      email:
        (form.get("email") as string) ||
        editingRow?.email ||
        "patient@example.com",
      initial: patientName.slice(0, 2).toUpperCase() || "NP",
      avatarBg: editingRow?.avatarBg || "#0284c7",
      receiptNo: (form.get("receiptNo") as string) || editingRow?.receiptNo,
      invoiceId: (form.get("invoiceId") as string) || editingRow?.invoiceId,
      account: (form.get("account") as string) || editingRow?.account,
      payTo: (form.get("payTo") as string) || editingRow?.payTo,
      date: (form.get("date") as string) || editingRow?.date || "5th Oct, 2026",
      amount: amountStr,
      status: (form.get("status") as string) || editingRow?.status || "Paid",
      paymentStatus:
        (form.get("paymentStatus") as string) ||
        editingRow?.paymentStatus ||
        "Approved",
      type:
        (form.get("type") as "Credit" | "Debit") ||
        editingRow?.type ||
        "Credit",
      txTime: editingRow?.txTime || "02:00 PM",
      txDate: editingRow?.txDate || "5th Oct, 2026",
      month: (form.get("month") as string) || editingRow?.month || "October",
      year: (form.get("year") as string) || editingRow?.year || "2026",
      netSalary:
        (form.get("netSalary") as string) || editingRow?.netSalary || "$0.00",
      payrollId:
        editingRow?.payrollId ||
        `#EMP${Math.floor(1000 + Math.random() * 9000)}`,
      billId:
        editingRow?.billId || `BL${Math.floor(1000 + Math.random() * 9000)}`,
    };

    setData((prev) => {
      const list = prev[canonicalSlug] || [];
      const exists = list.some((r) => r.id === rowId);
      const nextList = exists
        ? list.map((r) => (r.id === rowId ? newRecord : r))
        : [newRecord, ...list];
      return { ...prev, [canonicalSlug]: nextList };
    });

    setEditingRow(null);
    setCreateOpen(false);
    setIsSubmitting(false);
  }

  function exportCsv() {
    if (!filteredList.length) return;
    const keys = Object.keys(filteredList[0]).filter((k) => k !== "id");
    const header = keys.join(",");
    const rows = filteredList.map((r) =>
      keys.map((k) => `"${String(r[k] ?? "").replace(/"/g, '""')}"`).join(","),
    );
    const csv = [header, ...rows].join("\n");
    const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${canonicalSlug}-report.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="billing-workspace-container" data-ready="true">
      {/* Backend API Connection Banner */}
      <div className="d-flex flex-column gap-2 mb-3">
        <div
          className="d-flex align-items-center justify-content-between p-2 px-3 rounded border"
          style={{
            backgroundColor: apiConnected
              ? "rgba(16, 185, 129, 0.08)"
              : "rgba(59, 130, 246, 0.08)",
            borderColor: apiConnected
              ? "rgba(16, 185, 129, 0.3)"
              : "rgba(59, 130, 246, 0.3)",
          }}
        >
          <div className="d-flex align-items-center gap-2">
            <span
              style={{
                display: "inline-block",
                width: "8px",
                height: "8px",
                borderRadius: "50%",
                backgroundColor: apiConnected ? "#10b981" : "#3b82f6",
              }}
            />
            <span className="fs-7 fw-semibold">
              {apiConnected
                ? t(
                    "Connected to Go/PostgreSQL Billing Ledger (/v1/invoices, /v1/charge-accounts, /v1/payrolls)",
                  )
                : t("Local preview mode · Syncing locally")}
            </span>
            <span
              className="badge-available-stock fs-8 py-0 px-2"
              style={{
                backgroundColor: "#3b82f622",
                color: "#3b82f6",
                borderColor: "#3b82f6",
              }}
            >
              {t("Invoices")}: {(data.invoices || []).length} | {t("Accounts")}: {(data.accounts || []).length} | {t("Payrolls")}: {(data["employee-payrolls"] || []).length}
            </span>
          </div>
          <button
            type="button"
            className="btn-icon-link fs-7 d-flex align-items-center gap-1"
            onClick={loadBillingData}
            disabled={isLoadingApi}
            title={t("Refresh billing ledger from Go API")}
          >
            <RefreshCw
              size={13}
              className={isLoadingApi ? "animate-spin" : ""}
            />
            <span>{isLoadingApi ? t("Syncing...") : t("Sync Backend")}</span>
          </button>
        </div>

        {apiSuccessBanner && (
          <div
            className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded"
            style={{
              borderColor: "#10b981",
              backgroundColor: "rgba(16, 185, 129, 0.1)",
              color: "#10b981",
            }}
          >
            <div className="d-flex align-items-center gap-2">
              <CheckCircle2 size={16} />
              <span className="fs-7">{apiSuccessBanner}</span>
            </div>
            <button
              type="button"
              onClick={() => setApiSuccessBanner("")}
              className="btn-icon-link"
            >
              <X size={14} />
            </button>
          </div>
        )}

        {apiErrorBanner && (
          <div
            className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded"
            style={{
              borderColor: "#ef4444",
              backgroundColor: "rgba(239, 68, 68, 0.1)",
              color: "#ef4444",
            }}
          >
            <div className="d-flex align-items-center gap-2">
              <AlertCircle size={16} />
              <span className="fs-7">{apiErrorBanner}</span>
            </div>
            <button
              type="button"
              onClick={() => setApiErrorBanner("")}
              className="btn-icon-link"
            >
              <X size={14} />
            </button>
          </div>
        )}
      </div>

      {/* Top action toolbar matching screenshot style */}
      <div className="billing-toolbar">
        <div className="billing-search-box">
          <Search size={18} className="search-icon" />
          <input
            type="text"
            placeholder={t("Search")}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            aria-label={t("Search")}
          />
        </div>

        <div className="billing-actions">
          {/* Payment Reports Actions */}
          {canonicalSlug === "payment-reports" && (
            <>
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setFilterOpen(true)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={exportCsv}
              >
                {t("Export to Excel")}
              </button>
            </>
          )}

          {/* Payments Actions */}
          {canonicalSlug === "payments" && (
            <>
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Export")}
                aria-label={t("Export")}
                onClick={exportCsv}
              >
                <FileDown size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => {
                  setEditingRow(null);
                  setCreateOpen(true);
                }}
              >
                {t("New Payment")}
              </button>
            </>
          )}

          {/* Invoices Actions */}
          {canonicalSlug === "invoices" && (
            <>
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setFilterOpen(true)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => {
                  setEditingRow(null);
                  setCreateOpen(true);
                }}
              >
                {t("New Invoice")}
              </button>
            </>
          )}

          {/* Advance Payments Actions */}
          {canonicalSlug === "advance-payments" && (
            <button
              type="button"
              className="btn-action-blue"
              onClick={() => {
                setEditingRow(null);
                setCreateOpen(true);
              }}
            >
              {t("New Advance Payment")}
            </button>
          )}

          {/* Accounts Actions */}
          {canonicalSlug === "accounts" && (
            <>
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setFilterOpen(true)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => {
                  setEditingRow(null);
                  setCreateOpen(true);
                }}
              >
                {t("New Account")}
              </button>
            </>
          )}

          {/* Employee Payrolls Actions */}
          {canonicalSlug === "employee-payrolls" && (
            <>
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setFilterOpen(true)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={exportCsv}
              >
                {t("Export to Excel")}
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => {
                  setEditingRow(null);
                  setCreateOpen(true);
                }}
              >
                {t("New Employee Payroll")}
              </button>
            </>
          )}

          {/* Bills Actions */}
          {canonicalSlug === "bills" && (
            <>
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setFilterOpen(true)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => {
                  setEditingRow(null);
                  setCreateOpen(true);
                }}
              >
                {t("New Bill")}
              </button>
            </>
          )}
        </div>
      </div>

      {/* Main Table Card */}
      <div className="billing-card">
        <div className="legacy-table-wrap">
          <table className="legacy-table billing-table">
            <thead>
              {/* Manual Billing Payments Columns */}
              {canonicalSlug === "manual-billing-payments" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("PATIENT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PAYMENT STATUS")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("STATUS")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("TRANSACTION DATE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("AMOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                </tr>
              )}

              {/* Advance Payments Columns */}
              {canonicalSlug === "advance-payments" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("RECEIPT NO")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PATIENT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("DATE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("AMOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("ACTION")}</th>
                </tr>
              )}

              {/* Payment Reports Columns */}
              {canonicalSlug === "payment-reports" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("PAYMENT DATE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("ACCOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PAY TO")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("TYPE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("AMOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                </tr>
              )}

              {/* Payments Columns */}
              {canonicalSlug === "payments" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("ACCOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PAYMENT DATE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PAY TO")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("AMOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("ACTION")}</th>
                </tr>
              )}

              {/* Invoices Columns */}
              {canonicalSlug === "invoices" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("INVOICE ID")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PATIENT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("INVOICE DATE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("AMOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              )}

              {/* Accounts Columns */}
              {canonicalSlug === "accounts" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("ACCOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("TYPE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("STATUS")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("ACTION")}</th>
                </tr>
              )}

              {/* Employee Payrolls Columns */}
              {canonicalSlug === "employee-payrolls" && (
                <tr>
                  <th>{t("SR NO")}</th>
                  <th>
                    <div className="th-sort">
                      {t("PAYROLL ID")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("EMPLOYEE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("MONTH")}</th>
                  <th>{t("YEAR")}</th>
                  <th>
                    <div className="th-sort">
                      {t("NET SALARY")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              )}

              {/* Bills Columns */}
              {canonicalSlug === "bills" && (
                <tr>
                  <th>
                    <div className="th-sort">
                      {t("BILL ID")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("PATIENT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("BILL DATE")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>
                    <div className="th-sort">
                      {t("AMOUNT")} <ArrowUpDown size={13} />
                    </div>
                  </th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              )}
            </thead>

            <tbody>
              {filteredList.map((row) => (
                <tr key={row.id}>
                  {/* Manual Billing Payments Row */}
                  {canonicalSlug === "manual-billing-payments" && (
                    <>
                      <td>
                        <div className="patient-cell">
                          <span
                            className="avatar-circle"
                            style={{
                              backgroundColor: row.avatarBg || "#6f42c1",
                            }}
                          >
                            {row.initial ||
                              row.patient?.slice(0, 2).toUpperCase()}
                          </span>
                          <div className="patient-info">
                            <span className="patient-name link-cyan">
                              {row.patient}
                            </span>
                            <span className="patient-email">{row.email}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="badge-pill badge-green">
                          {t(row.paymentStatus || "Approved")}
                        </span>
                      </td>
                      <td>
                        <span className="badge-pill badge-green">
                          {t(row.status || "Paid")}
                        </span>
                      </td>
                      <td>
                        <div className="tx-date-badge">
                          <span className="tx-time link-cyan">
                            {row.txTime}
                          </span>
                          <span className="tx-date link-cyan">
                            {row.txDate}
                          </span>
                        </div>
                      </td>
                      <td>
                        <span className="amount-text">{row.amount}</span>
                      </td>
                    </>
                  )}

                  {/* Advance Payments Row */}
                  {canonicalSlug === "advance-payments" && (
                    <>
                      <td>
                        <span className="badge-pill badge-blue-link">
                          {row.receiptNo}
                        </span>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <span
                            className="avatar-circle"
                            style={{
                              backgroundColor: row.avatarBg || "#0284c7",
                            }}
                          >
                            {row.initial ||
                              row.patient?.slice(0, 2).toUpperCase()}
                          </span>
                          <div className="patient-info">
                            <span className="patient-name link-cyan">
                              {row.patient}
                            </span>
                            <span className="patient-email">{row.email}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="link-cyan">{row.date}</span>
                      </td>
                      <td>
                        <span className="amount-text">{row.amount}</span>
                      </td>
                      <td>
                        <div className="action-buttons">
                          <button
                            type="button"
                            className="action-btn-edit"
                            aria-label={t("Edit")}
                            onClick={() => {
                              setEditingRow(row);
                              setCreateOpen(true);
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-delete"
                            aria-label={t("Delete")}
                            onClick={() => handleDelete(row.id)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}

                  {/* Payment Reports Row */}
                  {canonicalSlug === "payment-reports" && (
                    <>
                      <td>
                        <span className="link-cyan">{row.date}</span>
                      </td>
                      <td>{row.account}</td>
                      <td>{row.payTo}</td>
                      <td>
                        {row.type === "Debit" ? (
                          <span className="badge-pill badge-red">
                            {t("Debit")}
                          </span>
                        ) : (
                          <span className="badge-pill badge-green">
                            {t("Credit")}
                          </span>
                        )}
                      </td>
                      <td>
                        <span className="amount-text">{row.amount}</span>
                      </td>
                    </>
                  )}

                  {/* Payments Row */}
                  {canonicalSlug === "payments" && (
                    <>
                      <td>{row.account}</td>
                      <td>
                        <span className="link-cyan">{row.date}</span>
                      </td>
                      <td>{row.payTo}</td>
                      <td>
                        <span className="amount-text">{row.amount}</span>
                      </td>
                      <td>
                        <div className="action-buttons">
                          <button
                            type="button"
                            className="action-btn-view"
                            aria-label={t("View")}
                            onClick={() => setViewRow(row)}
                          >
                            <Eye size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-edit"
                            aria-label={t("Edit")}
                            onClick={() => {
                              setEditingRow(row);
                              setCreateOpen(true);
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-delete"
                            aria-label={t("Delete")}
                            onClick={() => handleDelete(row.id)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}

                  {/* Invoices Row */}
                  {canonicalSlug === "invoices" && (
                    <>
                      <td>
                        <span className="badge-pill badge-blue-link">
                          {row.invoiceId}
                        </span>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <span
                            className="avatar-circle"
                            style={{
                              backgroundColor: row.avatarBg || "#0284c7",
                            }}
                          >
                            {row.initial ||
                              row.patient?.slice(0, 2).toUpperCase()}
                          </span>
                          <div className="patient-info">
                            <span className="patient-name link-cyan">
                              {row.patient}
                            </span>
                            <span className="patient-email">{row.email}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="link-cyan">{row.date}</span>
                      </td>
                      <td>
                        <span className="amount-text">{row.amount}</span>
                      </td>
                      <td>
                        {row.status === "Pending" ? (
                          <span className="badge-pill badge-amber">
                            {t("Pending")}
                          </span>
                        ) : (
                          <span className="badge-pill badge-green">
                            {t("Paid")}
                          </span>
                        )}
                      </td>
                      <td>
                        <div className="action-buttons">
                          <button
                            type="button"
                            className="action-btn-edit"
                            aria-label={t("Edit")}
                            onClick={() => {
                              setEditingRow(row);
                              setCreateOpen(true);
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-delete"
                            aria-label={t("Delete")}
                            onClick={() => handleDelete(row.id)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}

                  {/* Accounts Row */}
                  {canonicalSlug === "accounts" && (
                    <>
                      <td>{row.account}</td>
                      <td>
                        {row.type === "Debit" ? (
                          <span className="badge-pill badge-red">
                            {t("Debit")}
                          </span>
                        ) : (
                          <span className="badge-pill badge-green">
                            {t("Credit")}
                          </span>
                        )}
                      </td>
                      <td>
                        {row.status === "Inactive" ? (
                          <span className="badge-pill badge-red">
                            {t("Inactive")}
                          </span>
                        ) : (
                          <span className="badge-pill badge-green">
                            {t("Active")}
                          </span>
                        )}
                      </td>
                      <td>
                        <div className="action-buttons">
                          <button
                            type="button"
                            className="action-btn-edit"
                            aria-label={t("Edit")}
                            onClick={() => {
                              setEditingRow(row);
                              setCreateOpen(true);
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-delete"
                            aria-label={t("Delete")}
                            onClick={() => handleDelete(row.id)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}

                  {/* Employee Payrolls Row */}
                  {canonicalSlug === "employee-payrolls" && (
                    <>
                      <td>{row.srNo || 1}</td>
                      <td>
                        <span className="badge-pill badge-blue-link">
                          {row.payrollId}
                        </span>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <span
                            className="avatar-circle"
                            style={{
                              backgroundColor: row.avatarBg || "#0284c7",
                            }}
                          >
                            {row.initial ||
                              row.patient?.slice(0, 2).toUpperCase()}
                          </span>
                          <div className="patient-info">
                            <span className="patient-name link-cyan">
                              {row.patient}
                            </span>
                            <span className="patient-email">{row.email}</span>
                          </div>
                        </div>
                      </td>
                      <td>{t(row.month || "September")}</td>
                      <td>{row.year || "2026"}</td>
                      <td>
                        <span className="amount-text">{row.netSalary}</span>
                      </td>
                      <td>
                        <span className="badge-pill badge-green">
                          {t(row.status || "Paid")}
                        </span>
                      </td>
                      <td>
                        <div className="action-buttons">
                          <button
                            type="button"
                            className="action-btn-view"
                            aria-label={t("View")}
                            onClick={() => setViewRow(row)}
                          >
                            <Eye size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-edit"
                            aria-label={t("Edit")}
                            onClick={() => {
                              setEditingRow(row);
                              setCreateOpen(true);
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-delete"
                            aria-label={t("Delete")}
                            onClick={() => handleDelete(row.id)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}

                  {/* Bills Row */}
                  {canonicalSlug === "bills" && (
                    <>
                      <td>
                        <span className="badge-pill badge-blue-link">
                          {row.billId}
                        </span>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <span
                            className="avatar-circle"
                            style={{
                              backgroundColor: row.avatarBg || "#0284c7",
                            }}
                          >
                            {row.initial ||
                              row.patient?.slice(0, 2).toUpperCase()}
                          </span>
                          <div className="patient-info">
                            <span className="patient-name link-cyan">
                              {row.patient}
                            </span>
                            <span className="patient-email">{row.email}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="link-cyan">{row.date}</span>
                      </td>
                      <td>
                        <span className="amount-text">{row.amount}</span>
                      </td>
                      <td>
                        <span
                          className={`badge-pill ${row.status === "Unpaid" ? "badge-red" : "badge-green"}`}
                        >
                          {t(row.status || "Paid")}
                        </span>
                      </td>
                      <td>
                        <div className="action-buttons">
                          <button
                            type="button"
                            className="action-btn-edit"
                            aria-label={t("Edit")}
                            onClick={() => {
                              setEditingRow(row);
                              setCreateOpen(true);
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            type="button"
                            className="action-btn-delete"
                            aria-label={t("Delete")}
                            onClick={() => handleDelete(row.id)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </>
                  )}
                </tr>
              ))}

              {!filteredList.length && (
                <tr>
                  <td
                    colSpan={7}
                    style={{ textAlign: "center", padding: "3rem" }}
                  >
                    {t("No matching records found")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Bottom Pagination matching screenshot layout */}
        <div className="billing-footer">
          <div className="billing-pagination-info">
            <span>{t("Show")}</span>
            <select
              className="billing-page-size-select"
              value={pageSize}
              onChange={(e) => {
                setPageSize(Number(e.target.value));
                setPage(1);
              }}
              aria-label={t("Show entries")}
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
            <span>
              {t("Showing 1 to 10 of {{count}} Results").replace(
                "{{count}}",
                String(total),
              )}
            </span>
          </div>

          <div className="billing-pagination-controls">
            <button
              type="button"
              className="billing-page-nav-btn"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              aria-label={t("Previous")}
            >
              <ChevronLeft size={16} />
            </button>

            {[1, 2, 3, 4]
              .filter((n) => n <= totalPages)
              .map((n) => (
                <button
                  key={n}
                  type="button"
                  className={`billing-page-btn ${page === n ? "active" : ""}`}
                  onClick={() => setPage(n)}
                >
                  {n}
                </button>
              ))}

            {totalPages > 5 && (
              <>
                <span className="billing-page-ellipsis">…</span>
                <button
                  type="button"
                  className={`billing-page-btn ${page === totalPages - 1 ? "active" : ""}`}
                  onClick={() => setPage(totalPages - 1)}
                >
                  {totalPages - 1}
                </button>
                <button
                  type="button"
                  className={`billing-page-btn ${page === totalPages ? "active" : ""}`}
                  onClick={() => setPage(totalPages)}
                >
                  {totalPages}
                </button>
              </>
            )}

            <button
              type="button"
              className="billing-page-nav-btn"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              aria-label={t("Next")}
            >
              <ChevronRight size={16} />
            </button>
          </div>
        </div>
      </div>

      {/* Filter Modal */}
      {filterOpen && (
        <Modal
          titleId="filter-records-title"
          onClose={() => setFilterOpen(false)}
        >
          <div className="p-6">
            <div className="page-heading flex justify-between items-center mb-4">
              <h2 id="filter-records-title">{t("Filter Records")}</h2>
              <button
                type="button"
                className="btn-icon"
                onClick={() => setFilterOpen(false)}
                aria-label={t("Close")}
              >
                <X size={18} />
              </button>
            </div>
            <div className="space-y-4">
              <div className="form-group">
                <label className="label">{t("Status")}</label>
                <select
                  className="field"
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                >
                  <option value="All">{t("All")}</option>
                  <option value="Paid">{t("Paid")}</option>
                  <option value="Pending">{t("Pending")}</option>
                  <option value="Approved">{t("Approved")}</option>
                  <option value="Credit">{t("Credit")}</option>
                  <option value="Debit">{t("Debit")}</option>
                </select>
              </div>
              <div className="flex justify-end gap-2 mt-4">
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => {
                    setStatusFilter("All");
                    setFilterOpen(false);
                  }}
                >
                  {t("Reset")}
                </button>
                <button
                  type="button"
                  className="btn-action-blue"
                  onClick={() => setFilterOpen(false)}
                >
                  {t("Apply")}
                </button>
              </div>
            </div>
          </div>
        </Modal>
      )}

      {/* Create / Edit Modal */}
      {createOpen && (
        <Modal
          titleId="billing-form-title"
          onClose={() => {
            setCreateOpen(false);
            setEditingRow(null);
          }}
          wide
        >
          <div className="p-6">
            <div className="page-heading flex justify-between items-center mb-4">
              <h2 id="billing-form-title">
                {editingRow
                  ? t(`Edit ${canonicalSlug.replace(/-/g, " ")}`)
                  : canonicalSlug === "advance-payments"
                    ? t("New Advance Payment")
                    : canonicalSlug === "payments"
                      ? t("New Payment")
                      : canonicalSlug === "invoices"
                        ? t("New Invoice")
                        : canonicalSlug === "accounts"
                          ? t("New Account")
                          : canonicalSlug === "employee-payrolls"
                            ? t("New Employee Payroll")
                            : t("New Bill")}
              </h2>
              <button
                type="button"
                className="btn-icon"
                onClick={() => {
                  setCreateOpen(false);
                  setEditingRow(null);
                }}
                aria-label={t("Close")}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveRow} className="space-y-4">
              {/* Advance Payment Form */}
              {canonicalSlug === "advance-payments" && (
                <>
                  <div className="form-group">
                    <label className="label">{t("Patient")}</label>
                    <input
                      name="patient"
                      defaultValue={editingRow?.patient || "Vinay Grover"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Email")}</label>
                    <input
                      name="email"
                      type="email"
                      defaultValue={editingRow?.email || "patient@example.com"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Receipt No")}</label>
                    <input
                      name="receiptNo"
                      defaultValue={
                        editingRow?.receiptNo ||
                        `RCP${Math.random().toString(36).substring(2, 8).toUpperCase()}`
                      }
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Amount")}</label>
                    <input
                      name="amount"
                      defaultValue={editingRow?.amount || "$500.00"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Date")}</label>
                    <input
                      name="date"
                      defaultValue={editingRow?.date || "5th Oct, 2026"}
                      className="field"
                      required
                    />
                  </div>
                </>
              )}

              {/* Payments Form */}
              {canonicalSlug === "payments" && (
                <>
                  <div className="form-group">
                    <label className="label">{t("Account")}</label>
                    <input
                      name="account"
                      defaultValue={editingRow?.account || "Pharmacy"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Pay To")}</label>
                    <input
                      name="payTo"
                      defaultValue={editingRow?.payTo || "Supplier Co"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Payment Date")}</label>
                    <input
                      name="date"
                      defaultValue={editingRow?.date || "5th Oct, 2026"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Amount")}</label>
                    <input
                      name="amount"
                      defaultValue={editingRow?.amount || "$1,000.00"}
                      className="field"
                      required
                    />
                  </div>
                </>
              )}

              {/* Invoices Form */}
              {canonicalSlug === "invoices" && (
                <>
                  <div className="form-group">
                    <label className="label">{t("Invoice ID")}</label>
                    <input
                      name="invoiceId"
                      defaultValue={
                        editingRow?.invoiceId ||
                        `HMS${Math.floor(10 + Math.random() * 90)}`
                      }
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Patient")}</label>
                    <input
                      name="patient"
                      defaultValue={editingRow?.patient || "Aa Aa"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Email")}</label>
                    <input
                      name="email"
                      type="email"
                      defaultValue={editingRow?.email || "avmedia04@gmail.com"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Invoice Date")}</label>
                    <input
                      name="date"
                      defaultValue={editingRow?.date || "5th Oct, 2026"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Amount")}</label>
                    <input
                      name="amount"
                      defaultValue={editingRow?.amount || "$750.00"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Status")}</label>
                    <select
                      name="status"
                      defaultValue={editingRow?.status || "Paid"}
                      className="field"
                    >
                      <option value="Paid">{t("Paid")}</option>
                      <option value="Pending">{t("Pending")}</option>
                    </select>
                  </div>
                </>
              )}

              {/* Accounts Form */}
              {canonicalSlug === "accounts" && (
                <>
                  <div className="form-group">
                    <label className="label">{t("Account Name")}</label>
                    <input
                      name="account"
                      defaultValue={editingRow?.account || ""}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Type")}</label>
                    <select
                      name="type"
                      defaultValue={editingRow?.type || "Credit"}
                      className="field"
                    >
                      <option value="Credit">{t("Credit")}</option>
                      <option value="Debit">{t("Debit")}</option>
                    </select>
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Status")}</label>
                    <select
                      name="status"
                      defaultValue={editingRow?.status || "Active"}
                      className="field"
                    >
                      <option value="Active">{t("Active")}</option>
                      <option value="Inactive">{t("Inactive")}</option>
                    </select>
                  </div>
                </>
              )}

              {/* Employee Payrolls Form */}
              {canonicalSlug === "employee-payrolls" && (
                <>
                  <div className="form-group">
                    <label className="label">{t("Employee Name")}</label>
                    <input
                      name="patient"
                      defaultValue={editingRow?.patient || "Staff Member"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Email")}</label>
                    <input
                      name="email"
                      defaultValue={editingRow?.email || "staff@hospital.et"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Month")}</label>
                    <input
                      name="month"
                      defaultValue={editingRow?.month || "October"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Year")}</label>
                    <input
                      name="year"
                      defaultValue={editingRow?.year || "2026"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Net Salary")}</label>
                    <input
                      name="netSalary"
                      defaultValue={editingRow?.netSalary || "$3,000.00"}
                      className="field"
                      required
                    />
                  </div>
                </>
              )}

              {/* Bills Form */}
              {canonicalSlug === "bills" && (
                <>
                  <div className="form-group">
                    <label className="label">{t("Patient")}</label>
                    <input
                      name="patient"
                      defaultValue={editingRow?.patient || "Patient Name"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Email")}</label>
                    <input
                      name="email"
                      defaultValue={editingRow?.email || "patient@hospital.et"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Bill Date")}</label>
                    <input
                      name="date"
                      defaultValue={editingRow?.date || "5th Oct, 2026"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Amount")}</label>
                    <input
                      name="amount"
                      defaultValue={editingRow?.amount || "$1,000.00"}
                      className="field"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="label">{t("Status")}</label>
                    <select
                      name="status"
                      defaultValue={editingRow?.status || "Paid"}
                      className="field"
                    >
                      <option value="Paid">{t("Paid")}</option>
                      <option value="Unpaid">{t("Unpaid")}</option>
                    </select>
                  </div>
                </>
              )}

              <div className="flex justify-end gap-2 mt-4">
                <button
                  type="button"
                  className="btn-action-secondary"
                  disabled={isSubmitting}
                  onClick={() => {
                    setCreateOpen(false);
                    setEditingRow(null);
                  }}
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* View Modal */}
      {viewRow && (
        <Modal titleId="billing-view-title" onClose={() => setViewRow(null)}>
          <div className="p-6">
            <div className="page-heading flex justify-between items-center mb-4">
              <h2 id="billing-view-title">{t("Record Details")}</h2>
              <button
                type="button"
                className="btn-icon"
                onClick={() => setViewRow(null)}
                aria-label={t("Close")}
              >
                <X size={18} />
              </button>
            </div>
            <div className="space-y-3">
              {Object.entries(viewRow)
                .filter(([k]) => !["id", "avatarBg", "initial"].includes(k))
                .map(([key, val]) => (
                  <div
                    key={key}
                    className="flex justify-between border-b pb-1 text-sm"
                  >
                    <strong className="capitalize">
                      {t(key.replace(/([A-Z])/g, " $1"))}:
                    </strong>
                    <span>{String(val ?? "")}</span>
                  </div>
                ))}
              <div className="flex justify-end mt-4">
                <button
                  type="button"
                  className="btn-action-blue"
                  onClick={() => setViewRow(null)}
                >
                  {t("Close")}
                </button>
              </div>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
