"use client";

import { useEffect, useState, useTransition } from "react";
import Link from "next/link";
import {
  PhoneCall,
  Users,
  Mail,
  HelpCircle,
  AlertTriangle,
  Plus,
  RefreshCw,
  Search,
  CheckCircle2,
  Clock,
  ArrowDownLeft,
  ArrowUpRight,
  ShieldCheck,
  Building,
  Check,
  FileText,
  X,
} from "lucide-react";
import { useLanguage } from "./language";
import { Modal } from "./modal";

export type FrontOfficeTab =
  | "call-logs"
  | "visitors"
  | "postals"
  | "enquiries"
  | "complaints";

interface CallLogItem {
  id: string;
  name: string;
  phone: string;
  date: string;
  follow_up_date?: string;
  note: string;
  call_type: number; // 1: Incoming, 2: Outgoing
}

interface VisitorItem {
  id: string;
  purpose: number; // 1: Visit, 2: Enquiry, 3: Seminar/Vendor
  name: string;
  phone: string;
  id_card: string;
  no_of_person: number;
  date: string;
  in_time: string;
  out_time: string;
  note: string;
}

interface PostalItem {
  id: string;
  from_title: string;
  to_title: string;
  reference_no: string;
  date: string;
  address: string;
  type: number; // 1: Receive, 2: Dispatch
}

interface EnquiryItem {
  id: string;
  full_name: string;
  email: string;
  contact_no: string;
  type: number; // 1: General, 2: Admission, 3: Billing, 4: Feedback
  message: string;
  status: number; // 0: Unread, 1: Read
  created_at?: string;
}

interface ComplaintItem {
  id: string;
  patient_id?: string;
  title: string;
  description: string;
  status: number; // 0: Pending, 1: In Progress, 2: Resolved, 3: Rejected
  response: string;
  created_at?: string;
}

export function FrontOfficeWorkspace({ id = "call-logs" }: { id?: string }) {
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<FrontOfficeTab>(
    (id as FrontOfficeTab) || "call-logs",
  );
  const [isPending, startTransition] = useTransition();

  // Search filter
  const [searchTerm, setSearchTerm] = useState("");

  // Live data states
  const [callLogs, setCallLogs] = useState<CallLogItem[]>([]);
  const [visitors, setVisitors] = useState<VisitorItem[]>([]);
  const [postals, setPostals] = useState<PostalItem[]>([]);
  const [enquiries, setEnquiries] = useState<EnquiryItem[]>([]);
  const [complaints, setComplaints] = useState<ComplaintItem[]>([]);

  const [loading, setLoading] = useState(false);
  const [isLive, setIsLive] = useState(false);

  // Modal open states
  const [showAddCallLog, setShowAddCallLog] = useState(false);
  const [showAddVisitor, setShowAddVisitor] = useState(false);
  const [showAddPostal, setShowAddPostal] = useState(false);
  const [showAddEnquiry, setShowAddEnquiry] = useState(false);
  const [showAddComplaint, setShowAddComplaint] = useState(false);
  const [resolveTarget, setResolveTarget] = useState<ComplaintItem | null>(null);

  // Form states
  const [callLogForm, setCallLogForm] = useState({
    name: "",
    phone: "",
    date: new Date().toISOString().split("T")[0],
    follow_up_date: "",
    note: "",
    call_type: 1,
  });

  const [visitorForm, setVisitorForm] = useState({
    name: "",
    phone: "",
    purpose: 1,
    id_card: "",
    no_of_person: 1,
    date: new Date().toISOString().split("T")[0],
    in_time: "09:00",
    out_time: "10:00",
    note: "",
  });

  const [postalForm, setPostalForm] = useState({
    from_title: "",
    to_title: "",
    reference_no: `POST-${Math.floor(1000 + Math.random() * 9000)}`,
    date: new Date().toISOString().split("T")[0],
    address: "",
    type: 1,
  });

  const [enquiryForm, setEnquiryForm] = useState({
    full_name: "",
    email: "",
    contact_no: "",
    type: 1,
    message: "",
  });

  const [complaintForm, setComplaintForm] = useState({
    title: "",
    description: "",
  });

  const [resolveForm, setResolveForm] = useState({
    status: 2,
    response: "",
  });

  // Seed default fallback data
  const seedFallback = () => {
    setCallLogs([
      {
        id: "call-1",
        name: "Abebe Bekele",
        phone: "+251911223344",
        date: "2026-10-05",
        follow_up_date: "2026-10-08",
        note: "Inquired about pediatric cardiology OPD slots.",
        call_type: 1,
      },
      {
        id: "call-2",
        name: "Tigist Haile",
        phone: "+251922334455",
        date: "2026-10-05",
        follow_up_date: "",
        note: "Followed up on lab pathology test results release.",
        call_type: 2,
      },
    ]);

    setVisitors([
      {
        id: "vis-1",
        purpose: 1,
        name: "Kassahun Tadesse",
        phone: "+251911445566",
        id_card: "KEBELE-08-9921",
        no_of_person: 2,
        date: "2026-10-05",
        in_time: "14:00",
        out_time: "15:30",
        note: "Visiting patient in Ward 2 Bed 104",
      },
      {
        id: "vis-2",
        purpose: 3,
        name: "Sara Mengistu (Pharma Rep)",
        phone: "+251933445566",
        id_card: "EMP-MED-441",
        no_of_person: 1,
        date: "2026-10-05",
        in_time: "11:00",
        out_time: "12:00",
        note: "Meeting with Chief Pharmacist",
      },
    ]);

    setPostals([
      {
        id: "post-1",
        from_title: "Ministry of Health (Ethiopia)",
        to_title: "Hospital Medical Director",
        reference_no: "MOH-REF-8842",
        date: "2026-10-04",
        address: "Sudan St, Addis Ababa",
        type: 1,
      },
      {
        id: "post-2",
        from_title: "Hospital Administration",
        to_title: "Ethiopian Blood Bank Service",
        reference_no: "EBBS-REQ-1029",
        date: "2026-10-05",
        address: "Ras Desta Damtew St, Addis Ababa",
        type: 2,
      },
    ]);

    setEnquiries([
      {
        id: "enq-1",
        full_name: "Dawit Wolde",
        email: "dawit.w@example.com",
        contact_no: "+251911778899",
        type: 2,
        message: "What are the admission requirements for surgical ward?",
        status: 1,
        created_at: "2026-10-05T09:30:00Z",
      },
      {
        id: "enq-2",
        full_name: "Meron Assefa",
        email: "meron.a@example.com",
        contact_no: "+251922889900",
        type: 1,
        message: "Do you offer MRI diagnostics on weekends?",
        status: 0,
        created_at: "2026-10-05T14:15:00Z",
      },
    ]);

    setComplaints([
      {
        id: "comp-1",
        title: "Wait time at pharmacy billing counter",
        description: "Waited 40 minutes to clear medicine bill.",
        status: 2,
        response: "Added a second dedicated counter for cash and mobile payments.",
        created_at: "2026-10-04T11:00:00Z",
      },
      {
        id: "comp-2",
        title: "Air conditioning in Waiting Area B",
        description: "Waiting area was uncomfortably hot during morning rounds.",
        status: 0,
        response: "",
        created_at: "2026-10-05T08:45:00Z",
      },
    ]);
  };

  const fetchFrontOfficeData = async () => {
    setLoading(true);
    try {
      const [resCalls, resVis, resPost, resEnq, resComp] = await Promise.allSettled([
        fetch("/api/hms/call-logs"),
        fetch("/api/hms/visitors"),
        fetch("/api/hms/postals"),
        fetch("/api/hms/enquiries"),
        fetch("/api/hms/complaints"),
      ]);

      let loadedCount = 0;

      if (resCalls.status === "fulfilled" && resCalls.value.ok) {
        const data = await resCalls.value.json();
        if (data.call_logs && Array.isArray(data.call_logs) && data.call_logs.length > 0) {
          setCallLogs(data.call_logs);
          loadedCount += data.call_logs.length;
        }
      }
      if (resVis.status === "fulfilled" && resVis.value.ok) {
        const data = await resVis.value.json();
        if (data.visitors && Array.isArray(data.visitors) && data.visitors.length > 0) {
          setVisitors(data.visitors);
          loadedCount += data.visitors.length;
        }
      }
      if (resPost.status === "fulfilled" && resPost.value.ok) {
        const data = await resPost.value.json();
        if (data.postals && Array.isArray(data.postals) && data.postals.length > 0) {
          setPostals(data.postals);
          loadedCount += data.postals.length;
        }
      }
      if (resEnq.status === "fulfilled" && resEnq.value.ok) {
        const data = await resEnq.value.json();
        if (data.enquiries && Array.isArray(data.enquiries) && data.enquiries.length > 0) {
          setEnquiries(data.enquiries);
          loadedCount += data.enquiries.length;
        }
      }
      if (resComp.status === "fulfilled" && resComp.value.ok) {
        const data = await resComp.value.json();
        if (data.complaints && Array.isArray(data.complaints) && data.complaints.length > 0) {
          setComplaints(data.complaints);
          loadedCount += data.complaints.length;
        }
      }

      setIsLive(true);
    } catch {
      setIsLive(false);
      seedFallback();
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    seedFallback();
    fetchFrontOfficeData();
  }, []);

  // Save Handlers
  const handleSaveCallLog = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name: callLogForm.name,
      phone: callLogForm.phone,
      date: callLogForm.date,
      follow_up_date: callLogForm.follow_up_date ? callLogForm.follow_up_date : undefined,
      note: callLogForm.note,
      call_type: Number(callLogForm.call_type),
    };

    try {
      const res = await fetch("/api/hms/call-logs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const created = await res.json();
        setCallLogs((prev) => [created, ...prev]);
      } else {
        const mock: CallLogItem = {
          id: `call-${Date.now()}`,
          ...payload,
        };
        setCallLogs((prev) => [mock, ...prev]);
      }
    } catch {
      const mock: CallLogItem = {
        id: `call-${Date.now()}`,
        ...payload,
      };
      setCallLogs((prev) => [mock, ...prev]);
    }
    setShowAddCallLog(false);
    setCallLogForm({
      name: "",
      phone: "",
      date: new Date().toISOString().split("T")[0],
      follow_up_date: "",
      note: "",
      call_type: 1,
    });
  };

  const handleSaveVisitor = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name: visitorForm.name,
      phone: visitorForm.phone,
      purpose: Number(visitorForm.purpose),
      id_card: visitorForm.id_card,
      no_of_person: Number(visitorForm.no_of_person) || 1,
      date: visitorForm.date,
      in_time: visitorForm.in_time,
      out_time: visitorForm.out_time,
      note: visitorForm.note,
    };

    try {
      const res = await fetch("/api/hms/visitors", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const created = await res.json();
        setVisitors((prev) => [created, ...prev]);
      } else {
        const mock: VisitorItem = {
          id: `vis-${Date.now()}`,
          ...payload,
        };
        setVisitors((prev) => [mock, ...prev]);
      }
    } catch {
      const mock: VisitorItem = {
        id: `vis-${Date.now()}`,
        ...payload,
      };
      setVisitors((prev) => [mock, ...prev]);
    }
    setShowAddVisitor(false);
    setVisitorForm({
      name: "",
      phone: "",
      purpose: 1,
      id_card: "",
      no_of_person: 1,
      date: new Date().toISOString().split("T")[0],
      in_time: "09:00",
      out_time: "10:00",
      note: "",
    });
  };

  const handleSavePostal = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      from_title: postalForm.from_title,
      to_title: postalForm.to_title,
      reference_no: postalForm.reference_no,
      date: postalForm.date,
      address: postalForm.address,
      type: Number(postalForm.type),
    };

    try {
      const res = await fetch("/api/hms/postals", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const created = await res.json();
        setPostals((prev) => [created, ...prev]);
      } else {
        const mock: PostalItem = {
          id: `post-${Date.now()}`,
          ...payload,
        };
        setPostals((prev) => [mock, ...prev]);
      }
    } catch {
      const mock: PostalItem = {
        id: `post-${Date.now()}`,
        ...payload,
      };
      setPostals((prev) => [mock, ...prev]);
    }
    setShowAddPostal(false);
    setPostalForm({
      from_title: "",
      to_title: "",
      reference_no: `POST-${Math.floor(1000 + Math.random() * 9000)}`,
      date: new Date().toISOString().split("T")[0],
      address: "",
      type: 1,
    });
  };

  const handleSaveEnquiry = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      full_name: enquiryForm.full_name,
      email: enquiryForm.email,
      contact_no: enquiryForm.contact_no,
      type: Number(enquiryForm.type),
      message: enquiryForm.message,
    };

    try {
      const res = await fetch("/api/hms/enquiries", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const created = await res.json();
        setEnquiries((prev) => [created, ...prev]);
      } else {
        const mock: EnquiryItem = {
          id: `enq-${Date.now()}`,
          ...payload,
          status: 0,
        };
        setEnquiries((prev) => [mock, ...prev]);
      }
    } catch {
      const mock: EnquiryItem = {
        id: `enq-${Date.now()}`,
        ...payload,
        status: 0,
      };
      setEnquiries((prev) => [mock, ...prev]);
    }
    setShowAddEnquiry(false);
    setEnquiryForm({
      full_name: "",
      email: "",
      contact_no: "",
      type: 1,
      message: "",
    });
  };

  const handleMarkEnquiryRead = async (id: string) => {
    try {
      await fetch(`/api/hms/enquiries/${id}/read`, { method: "PUT" });
    } catch {}
    setEnquiries((prev) =>
      prev.map((item) => (item.id === id ? { ...item, status: 1 } : item)),
    );
  };

  const handleSaveComplaint = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      title: complaintForm.title,
      description: complaintForm.description,
    };

    try {
      const res = await fetch("/api/hms/complaints", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        const created = await res.json();
        setComplaints((prev) => [created, ...prev]);
      } else {
        const mock: ComplaintItem = {
          id: `comp-${Date.now()}`,
          ...payload,
          status: 0,
          response: "",
        };
        setComplaints((prev) => [mock, ...prev]);
      }
    } catch {
      const mock: ComplaintItem = {
        id: `comp-${Date.now()}`,
        ...payload,
        status: 0,
        response: "",
      };
      setComplaints((prev) => [mock, ...prev]);
    }
    setShowAddComplaint(false);
    setComplaintForm({ title: "", description: "" });
  };

  const handleResolveComplaint = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!resolveTarget) return;

    const payload = {
      status: Number(resolveForm.status),
      response: resolveForm.response,
    };

    try {
      await fetch(`/api/hms/complaints/${resolveTarget.id}/resolve`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
    } catch {}

    setComplaints((prev) =>
      prev.map((c) =>
        c.id === resolveTarget.id
          ? { ...c, status: payload.status, response: payload.response }
          : c,
      ),
    );
    setResolveTarget(null);
    setResolveForm({ status: 2, response: "" });
  };

  const tabs = [
    { id: "call-logs", label: t("Call Logs"), icon: PhoneCall, count: callLogs.length },
    { id: "visitors", label: t("Visitors"), icon: Users, count: visitors.length },
    { id: "postals", label: t("Postal"), icon: Mail, count: postals.length },
    { id: "enquiries", label: t("Enquiries"), icon: HelpCircle, count: enquiries.length },
    { id: "complaints", label: t("Complaints"), icon: AlertTriangle, count: complaints.length },
  ];

  return (
    <div className="space-y-6">
      {/* Subtabs Navigation */}
      <div className="border-b border-border/80 bg-card/50 backdrop-blur rounded-xl p-1.5 shadow-sm">
        <nav className="flex space-x-1 overflow-x-auto">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as FrontOfficeTab)}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 whitespace-nowrap ${
                  isActive
                    ? "bg-primary text-primary-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{tab.label}</span>
                <span
                  className={`text-xs px-1.5 py-0.5 rounded-full ${
                    isActive
                      ? "bg-primary-foreground/20 text-primary-foreground font-semibold"
                      : "bg-muted text-muted-foreground"
                  }`}
                >
                  {tab.count}
                </span>
              </button>
            );
          })}
        </nav>
      </div>

      {/* Live Status Banner */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-gradient-to-r from-emerald-500/10 via-emerald-500/5 to-transparent border border-emerald-500/20 rounded-xl p-3.5 text-sm shadow-sm">
        <div className="flex items-center gap-2.5">
          <span className="relative flex h-2.5 w-2.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
          </span>
          <span className="font-semibold text-emerald-950 dark:text-emerald-200">
            {t("Front Office Active")}
          </span>
          <span className="text-muted-foreground hidden sm:inline">•</span>
          <span className="text-muted-foreground text-xs sm:text-sm">
            {isLive ? t("Live Go / PostgreSQL Connected") : t("Dual-mode local preview")}
          </span>
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto justify-end">
          <button
            onClick={() => fetchFrontOfficeData()}
            disabled={loading}
            className="btn-secondary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg border border-border shadow-2xs hover:bg-muted"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
            {t("Sync")}
          </button>
          {activeTab === "call-logs" && (
            <button
              onClick={() => setShowAddCallLog(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Call Log")}
            </button>
          )}
          {activeTab === "visitors" && (
            <button
              onClick={() => setShowAddVisitor(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Visitor")}
            </button>
          )}
          {activeTab === "postals" && (
            <button
              onClick={() => setShowAddPostal(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Postal")}
            </button>
          )}
          {activeTab === "enquiries" && (
            <button
              onClick={() => setShowAddEnquiry(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Enquiry")}
            </button>
          )}
          {activeTab === "complaints" && (
            <button
              onClick={() => setShowAddComplaint(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Complaint")}
            </button>
          )}
        </div>
      </div>

      {/* Main Content Areas */}

      {/* 1. CALL LOGS */}
      {activeTab === "call-logs" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">{t("Telephonic Call Logs")}</h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search caller or phone...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Caller Name")}</th>
                  <th className="px-4 py-3">{t("Phone")}</th>
                  <th className="px-4 py-3">{t("Call Date")}</th>
                  <th className="px-4 py-3">{t("Follow Up Date")}</th>
                  <th className="px-4 py-3">{t("Call Type")}</th>
                  <th className="px-4 py-3">{t("Note")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {callLogs
                  .filter(
                    (c) =>
                      c.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      c.phone.includes(searchTerm),
                  )
                  .map((c) => (
                    <tr key={c.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-medium text-foreground">{c.name}</td>
                      <td className="px-4 py-3 font-mono text-xs">{c.phone}</td>
                      <td className="px-4 py-3 text-muted-foreground">{c.date}</td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {c.follow_up_date || "-"}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            c.call_type === 1
                              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                              : "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                          }`}
                        >
                          {c.call_type === 1 ? (
                            <>
                              <ArrowDownLeft className="w-3 h-3" /> {t("Incoming")}
                            </>
                          ) : (
                            <>
                              <ArrowUpRight className="w-3 h-3" /> {t("Outgoing")}
                            </>
                          )}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground text-xs max-w-xs truncate">
                        {c.note || "-"}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 2. VISITORS */}
      {activeTab === "visitors" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">{t("Visitor Register")}</h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search visitor or ID...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Visitor Name")}</th>
                  <th className="px-4 py-3">{t("Purpose")}</th>
                  <th className="px-4 py-3">{t("Phone")}</th>
                  <th className="px-4 py-3">{t("ID Card")}</th>
                  <th className="px-4 py-3">{t("Persons")}</th>
                  <th className="px-4 py-3">{t("Date")}</th>
                  <th className="px-4 py-3">{t("Timing")}</th>
                  <th className="px-4 py-3">{t("Note")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {visitors
                  .filter(
                    (v) =>
                      v.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      v.id_card.toLowerCase().includes(searchTerm.toLowerCase()),
                  )
                  .map((v) => (
                    <tr key={v.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-medium text-foreground">{v.name}</td>
                      <td className="px-4 py-3">
                        <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">
                          {v.purpose === 1
                            ? t("Visit")
                            : v.purpose === 2
                              ? t("Enquiry")
                              : t("Seminar/Vendor")}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs">{v.phone}</td>
                      <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                        {v.id_card || "-"}
                      </td>
                      <td className="px-4 py-3">{v.no_of_person}</td>
                      <td className="px-4 py-3 text-muted-foreground">{v.date}</td>
                      <td className="px-4 py-3 font-mono text-xs">
                        {v.in_time} - {v.out_time}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground text-xs max-w-xs truncate">
                        {v.note || "-"}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 3. POSTALS */}
      {activeTab === "postals" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Postal Dispatch & Receive")}
            </h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search reference or title...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Type")}</th>
                  <th className="px-4 py-3">{t("Reference No")}</th>
                  <th className="px-4 py-3">{t("From")}</th>
                  <th className="px-4 py-3">{t("To")}</th>
                  <th className="px-4 py-3">{t("Date")}</th>
                  <th className="px-4 py-3">{t("Address / Courier")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {postals
                  .filter(
                    (p) =>
                      p.reference_no.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      p.from_title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      p.to_title.toLowerCase().includes(searchTerm.toLowerCase()),
                  )
                  .map((p) => (
                    <tr key={p.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            p.type === 1
                              ? "bg-purple-500/10 text-purple-600 dark:text-purple-400"
                              : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                          }`}
                        >
                          {p.type === 1 ? t("Receive") : t("Dispatch")}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono font-medium text-xs">
                        {p.reference_no}
                      </td>
                      <td className="px-4 py-3 font-medium text-foreground">{p.from_title}</td>
                      <td className="px-4 py-3 font-medium text-foreground">{p.to_title}</td>
                      <td className="px-4 py-3 text-muted-foreground">{p.date}</td>
                      <td className="px-4 py-3 text-muted-foreground text-xs max-w-xs truncate">
                        {p.address || "-"}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 4. ENQUIRIES */}
      {activeTab === "enquiries" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">{t("Public & Patient Enquiries")}</h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search enquiry or name...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Full Name")}</th>
                  <th className="px-4 py-3">{t("Contact Details")}</th>
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Message")}</th>
                  <th className="px-4 py-3">{t("Status")}</th>
                  <th className="px-4 py-3 text-right">{t("Action")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {enquiries
                  .filter(
                    (e) =>
                      e.full_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      e.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      e.message.toLowerCase().includes(searchTerm.toLowerCase()),
                  )
                  .map((e) => (
                    <tr key={e.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-medium text-foreground">{e.full_name}</td>
                      <td className="px-4 py-3">
                        <div className="font-mono text-xs">{e.contact_no}</div>
                        <div className="text-xs text-muted-foreground">{e.email}</div>
                      </td>
                      <td className="px-4 py-3">
                        <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">
                          {e.type === 1
                            ? t("General")
                            : e.type === 2
                              ? t("Admission")
                              : e.type === 3
                                ? t("Billing")
                                : t("Feedback")}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-foreground text-xs max-w-sm">
                        {e.message}
                      </td>
                      <td className="px-4 py-3">
                        {e.status === 1 ? (
                          <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-600 dark:text-emerald-400">
                            <CheckCircle2 className="w-3.5 h-3.5" /> {t("Read")}
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-600 dark:text-amber-400">
                            <Clock className="w-3.5 h-3.5" /> {t("Unread")}
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-right">
                        {e.status === 0 && (
                          <button
                            onClick={() => handleMarkEnquiryRead(e.id)}
                            className="px-2.5 py-1 text-xs font-medium bg-muted hover:bg-primary hover:text-primary-foreground rounded-md transition-colors"
                          >
                            {t("Mark Read")}
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 5. COMPLAINTS */}
      {activeTab === "complaints" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">{t("Complaints & Grievances")}</h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search complaints...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Title")}</th>
                  <th className="px-4 py-3">{t("Description")}</th>
                  <th className="px-4 py-3">{t("Status")}</th>
                  <th className="px-4 py-3">{t("Resolution Response")}</th>
                  <th className="px-4 py-3 text-right">{t("Action")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {complaints
                  .filter(
                    (c) =>
                      c.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      c.description.toLowerCase().includes(searchTerm.toLowerCase()),
                  )
                  .map((c) => (
                    <tr key={c.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-semibold text-foreground">{c.title}</td>
                      <td className="px-4 py-3 text-xs text-muted-foreground max-w-sm">
                        {c.description}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                            c.status === 2
                              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                              : c.status === 1
                                ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                                : c.status === 3
                                  ? "bg-red-500/10 text-red-600 dark:text-red-400"
                                  : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                          }`}
                        >
                          {c.status === 2
                            ? t("Resolved")
                            : c.status === 1
                              ? t("In Progress")
                              : c.status === 3
                                ? t("Rejected")
                                : t("Pending")}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground max-w-xs">
                        {c.response || t("Awaiting review")}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button
                          onClick={() => {
                            setResolveTarget(c);
                            setResolveForm({
                              status: c.status === 0 ? 2 : c.status,
                              response: c.response || "",
                            });
                          }}
                          className="px-2.5 py-1 text-xs font-medium bg-muted hover:bg-primary hover:text-primary-foreground rounded-md transition-colors"
                        >
                          {t("Resolve / Update")}
                        </button>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* --- MODALS --- */}

      {/* 1. Add Call Log Modal */}
      {showAddCallLog && (
        <Modal onClose={() => setShowAddCallLog(false)} titleId="add-call-log-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-call-log-title" className="font-semibold text-lg">{t("New Call Log")}</h3>
              <button
                onClick={() => setShowAddCallLog(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveCallLog} className="space-y-4">
              <div>
                <label className="block text-xs font-medium mb-1">{t("Caller Name")} *</label>
                <input
                  type="text"
                  required
                  value={callLogForm.name}
                  onChange={(e) => setCallLogForm({ ...callLogForm, name: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Phone")} *</label>
                  <input
                    type="text"
                    required
                    value={callLogForm.phone}
                    onChange={(e) => setCallLogForm({ ...callLogForm, phone: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Call Type")}</label>
                  <select
                    value={callLogForm.call_type}
                    onChange={(e) =>
                      setCallLogForm({ ...callLogForm, call_type: Number(e.target.value) })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    <option value={1}>{t("Incoming")}</option>
                    <option value={2}>{t("Outgoing")}</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Date")} *</label>
                  <input
                    type="date"
                    required
                    value={callLogForm.date}
                    onChange={(e) => setCallLogForm({ ...callLogForm, date: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Follow Up Date")}</label>
                  <input
                    type="date"
                    value={callLogForm.follow_up_date}
                    onChange={(e) =>
                      setCallLogForm({ ...callLogForm, follow_up_date: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Note")}</label>
                <textarea
                  rows={3}
                  value={callLogForm.note}
                  onChange={(e) => setCallLogForm({ ...callLogForm, note: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddCallLog(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button type="submit" className="btn-primary px-4 py-2 text-xs rounded-lg">
                  {t("Save Call Log")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 2. Add Visitor Modal */}
      {showAddVisitor && (
        <Modal onClose={() => setShowAddVisitor(false)} titleId="add-visitor-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-visitor-title" className="font-semibold text-lg">{t("New Visitor Entry")}</h3>
              <button
                onClick={() => setShowAddVisitor(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveVisitor} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Visitor Name")} *</label>
                  <input
                    type="text"
                    required
                    value={visitorForm.name}
                    onChange={(e) => setVisitorForm({ ...visitorForm, name: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Purpose")}</label>
                  <select
                    value={visitorForm.purpose}
                    onChange={(e) =>
                      setVisitorForm({ ...visitorForm, purpose: Number(e.target.value) })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    <option value={1}>{t("Visit")}</option>
                    <option value={2}>{t("Enquiry")}</option>
                    <option value={3}>{t("Seminar/Vendor")}</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Phone")} *</label>
                  <input
                    type="text"
                    required
                    value={visitorForm.phone}
                    onChange={(e) => setVisitorForm({ ...visitorForm, phone: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("ID Card / Kebele No")}</label>
                  <input
                    type="text"
                    value={visitorForm.id_card}
                    onChange={(e) => setVisitorForm({ ...visitorForm, id_card: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("No of Persons")}</label>
                  <input
                    type="number"
                    min={1}
                    value={visitorForm.no_of_person}
                    onChange={(e) =>
                      setVisitorForm({ ...visitorForm, no_of_person: Number(e.target.value) })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("In Time")}</label>
                  <input
                    type="text"
                    value={visitorForm.in_time}
                    onChange={(e) => setVisitorForm({ ...visitorForm, in_time: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Out Time")}</label>
                  <input
                    type="text"
                    value={visitorForm.out_time}
                    onChange={(e) => setVisitorForm({ ...visitorForm, out_time: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg font-mono"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Note / Visiting Patient")}</label>
                <textarea
                  rows={2}
                  value={visitorForm.note}
                  onChange={(e) => setVisitorForm({ ...visitorForm, note: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddVisitor(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button type="submit" className="btn-primary px-4 py-2 text-xs rounded-lg">
                  {t("Save Visitor")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 3. Add Postal Modal */}
      {showAddPostal && (
        <Modal onClose={() => setShowAddPostal(false)} titleId="add-postal-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-postal-title" className="font-semibold text-lg">{t("New Postal Dispatch / Receive")}</h3>
              <button
                onClick={() => setShowAddPostal(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSavePostal} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Postal Type")}</label>
                  <select
                    value={postalForm.type}
                    onChange={(e) =>
                      setPostalForm({ ...postalForm, type: Number(e.target.value) })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    <option value={1}>{t("Receive")}</option>
                    <option value={2}>{t("Dispatch")}</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Reference No")} *</label>
                  <input
                    type="text"
                    required
                    value={postalForm.reference_no}
                    onChange={(e) =>
                      setPostalForm({ ...postalForm, reference_no: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg font-mono"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("From Title")} *</label>
                  <input
                    type="text"
                    required
                    value={postalForm.from_title}
                    onChange={(e) =>
                      setPostalForm({ ...postalForm, from_title: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("To Title")} *</label>
                  <input
                    type="text"
                    required
                    value={postalForm.to_title}
                    onChange={(e) => setPostalForm({ ...postalForm, to_title: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Date")} *</label>
                <input
                  type="date"
                  required
                  value={postalForm.date}
                  onChange={(e) => setPostalForm({ ...postalForm, date: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Address / Courier Notes")}</label>
                <textarea
                  rows={2}
                  value={postalForm.address}
                  onChange={(e) => setPostalForm({ ...postalForm, address: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddPostal(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button type="submit" className="btn-primary px-4 py-2 text-xs rounded-lg">
                  {t("Save Postal")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 4. Add Enquiry Modal */}
      {showAddEnquiry && (
        <Modal onClose={() => setShowAddEnquiry(false)} titleId="add-enquiry-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-enquiry-title" className="font-semibold text-lg">{t("New Patient Enquiry")}</h3>
              <button
                onClick={() => setShowAddEnquiry(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveEnquiry} className="space-y-4">
              <div>
                <label className="block text-xs font-medium mb-1">{t("Full Name")} *</label>
                <input
                  type="text"
                  required
                  value={enquiryForm.full_name}
                  onChange={(e) =>
                    setEnquiryForm({ ...enquiryForm, full_name: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Email")} *</label>
                  <input
                    type="email"
                    required
                    value={enquiryForm.email}
                    onChange={(e) =>
                      setEnquiryForm({ ...enquiryForm, email: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{t("Phone / Mobile")} *</label>
                  <input
                    type="text"
                    required
                    value={enquiryForm.contact_no}
                    onChange={(e) =>
                      setEnquiryForm({ ...enquiryForm, contact_no: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Category")}</label>
                <select
                  value={enquiryForm.type}
                  onChange={(e) =>
                    setEnquiryForm({ ...enquiryForm, type: Number(e.target.value) })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                >
                  <option value={1}>{t("General Enquiry")}</option>
                  <option value={2}>{t("Admission")}</option>
                  <option value={3}>{t("Billing")}</option>
                  <option value={4}>{t("Feedback")}</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Message")} *</label>
                <textarea
                  rows={3}
                  required
                  value={enquiryForm.message}
                  onChange={(e) =>
                    setEnquiryForm({ ...enquiryForm, message: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddEnquiry(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button type="submit" className="btn-primary px-4 py-2 text-xs rounded-lg">
                  {t("Submit Enquiry")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 5. Add Complaint Modal */}
      {showAddComplaint && (
        <Modal onClose={() => setShowAddComplaint(false)} titleId="add-complaint-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-complaint-title" className="font-semibold text-lg">{t("File Patient Complaint")}</h3>
              <button
                onClick={() => setShowAddComplaint(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveComplaint} className="space-y-4">
              <div>
                <label className="block text-xs font-medium mb-1">{t("Title")} *</label>
                <input
                  type="text"
                  required
                  value={complaintForm.title}
                  onChange={(e) =>
                    setComplaintForm({ ...complaintForm, title: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  placeholder={t("Brief subject of the complaint")}
                />
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">{t("Detailed Description")} *</label>
                <textarea
                  rows={4}
                  required
                  value={complaintForm.description}
                  onChange={(e) =>
                    setComplaintForm({ ...complaintForm, description: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  placeholder={t("Explain the occurrence, department, and relevant context...")}
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddComplaint(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button type="submit" className="btn-primary px-4 py-2 text-xs rounded-lg">
                  {t("Submit Complaint")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 6. Resolve Complaint Modal */}
      {resolveTarget && (
        <Modal onClose={() => setResolveTarget(null)} titleId="resolve-complaint-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="resolve-complaint-title" className="font-semibold text-lg">{t("Resolve Complaint")}</h3>
              <button
                onClick={() => setResolveTarget(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="bg-muted/50 p-3 rounded-lg text-xs space-y-1">
              <div className="font-semibold text-foreground">{resolveTarget.title}</div>
              <div className="text-muted-foreground">{resolveTarget.description}</div>
            </div>
            <form onSubmit={handleResolveComplaint} className="space-y-4">
              <div>
                <label className="block text-xs font-medium mb-1">{t("Update Status")}</label>
                <select
                  value={resolveForm.status}
                  onChange={(e) =>
                    setResolveForm({ ...resolveForm, status: Number(e.target.value) })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                >
                  <option value={1}>{t("In Progress")}</option>
                  <option value={2}>{t("Resolved")}</option>
                  <option value={3}>{t("Rejected")}</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Response / Resolution Actions")}
                </label>
                <textarea
                  rows={3}
                  value={resolveForm.response}
                  onChange={(e) =>
                    setResolveForm({ ...resolveForm, response: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  placeholder={t("Detail resolution steps or findings...")}
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setResolveTarget(null)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button type="submit" className="btn-primary px-4 py-2 text-xs rounded-lg">
                  {t("Save Resolution")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}
    </div>
  );
}
