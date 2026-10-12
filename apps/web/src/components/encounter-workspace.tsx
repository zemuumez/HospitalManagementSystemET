"use client";

import { useCallback, useEffect, useState } from "react";
import {
  Search,
  Eye,
  Plus,
  RefreshCw,
  Bed as BedIcon,
  UserCheck,
  Stethoscope,
  Clock,
  DollarSign,
  FileText,
  ShieldCheck,
  LogOut,
  X,
  AlertCircle,
  Calendar,
  Layers,
  ArrowRightLeft,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { api } from "@/lib/api";
import { useLanguage } from "./language";
import { useIdentity } from "./workspace";
import { errorText } from "./connected-scheduling";

export type Encounter = {
  id: string;
  number: number;
  kind: "ipd" | "opd";
  caseId: string;
  bedId: string;
  bedName: string;
  admittedAt: string;
  symptoms: string;
  patientId: string;
  patientName: string;
  patientEmail?: string;
  doctorId: string;
  doctorName: string;
  doctorEmail?: string;
  status: "active" | "discharged";
  version: number;
  dischargedAt?: string;
  dischargeSummary?: string;
  billStatus: string;
  totalVisits: number;
  intake?: {
    telephone?: string;
    taxReference?: string;
    height?: number;
    weight?: number;
    bloodPressure?: string;
    notes?: string;
    identificationNumber?: string;
    reference?: string;
    oldPatient?: boolean;
    standardChargeMinor?: number;
    paymentMode?: string;
  };
};

type Case = {
  id: string;
  number: number;
  patientId: string;
  patientName: string;
  doctorId: string;
  doctorName: string;
  description: string;
};

type BedOption = {
  id: string;
  name: string;
  type: string;
  wardName?: string;
  chargeMinor: number;
  available: boolean;
};

type CareTeamMember = {
  id: string;
  encounterId: string;
  staffId: string;
  staffName: string;
  roleTitle: string;
  assignedAt: string;
  revokedAt?: string;
  assignedBy: string;
  notes?: string;
};

type Diagnosis = {
  id: string;
  encounterId: string;
  icd10Code: string;
  description: string;
  category: string;
  status: string;
  diagnosedBy: string;
  diagnosedAt: string;
};

type Procedure = {
  id: string;
  encounterId: string;
  procedureCode: string;
  name: string;
  description?: string;
  performedBy?: string;
  performedAt: string;
};

type ClinicalNote = {
  id: string;
  authorName: string;
  body: string;
  signedAt: string;
};

type BedEvent = {
  id: string;
  kind: string;
  fromBedId?: string;
  toBedId?: string;
  chargeMinor: number;
  actorId: string;
  reason?: string;
  recordedAt: string;
};

type BedAssignment = {
  id: string;
  bedId: string;
  bedName: string;
  wardName: string;
  assignedFrom: string;
  assignedTo?: string;
  notes?: string;
};

type FollowUp = {
  id: string;
  patientId: string;
  encounterId?: string;
  followUpDate: string;
  notes: string;
  status: string;
};

interface EncounterWorkspaceProps {
  slug: string;
  kind?: "ipd" | "opd";
}

export function EncounterWorkspace({
  slug,
  kind: defaultKind,
}: EncounterWorkspaceProps) {
  const { t } = useLanguage();
  const identity = useIdentity();

  const isOPD = defaultKind === "opd" || slug.startsWith("opd-");
  const kind = isOPD ? "opd" : "ipd";

  // Map slug to initial detail tab if user opens an encounter
  const defaultTabFromSlug = (() => {
    if (slug.includes("diagnosis")) return "diagnoses";
    if (slug.includes("consultant")) return "care-team";
    if (
      slug.includes("charge") ||
      slug.includes("bill") ||
      slug.includes("payment")
    )
      return "billing";
    if (slug.includes("timeline")) return "notes";
    return "overview";
  })();

  const [page, setPage] = useState(1);
  const [pageSize] = useState(25);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    "All" | "active" | "discharged"
  >("All");
  const [encounters, setEncounters] = useState<Encounter[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedEncounterId, setSelectedEncounterId] = useState<string | null>(
    null,
  );
  const [admitModalOpen, setAdmitModalOpen] = useState(false);

  const fetchEncounters = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await api<{ encounters: Encounter[]; total?: number }>(
        `encounters?kind=${kind}&page=${page}&pageSize=${pageSize}`,
      );
      setEncounters(res.encounters || []);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setLoading(false);
    }
  }, [kind, page, pageSize]);

  useEffect(() => {
    fetchEncounters();
  }, [fetchEncounters]);

  const filteredEncounters = encounters.filter((e) => {
    const matchesSearch =
      search === "" ||
      `${e.number} ${e.patientName} ${e.doctorName} ${e.bedName || ""} ${e.symptoms}`
        .toLowerCase()
        .includes(search.toLowerCase());
    const matchesStatus = statusFilter === "All" || e.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getPageTitle = () => {
    switch (slug) {
      case "ipd-diagnosis":
        return t("IPD Diagnoses");
      case "ipd-consultant-registers":
        return t("IPD Consultant Registers");
      case "ipd-charges":
        return t("IPD Charges");
      case "ipd-payments":
        return t("IPD Payments");
      case "ipd-bills":
        return t("IPD Bills");
      case "ipd-timelines":
        return t("IPD Timelines");
      case "opd-diagnosis":
        return t("OPD Diagnoses");
      case "opd-timelines":
        return t("OPD Timelines");
      default:
        return kind === "ipd"
          ? t("IPD Patient Departments")
          : t("OPD Patient Departments");
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-card p-6 rounded-xl border border-border shadow-sm">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {getPageTitle()}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {kind === "ipd"
              ? t(
                  "Manage inpatient admissions, bed tracking, consultations, and discharge clearance.",
                )
              : t(
                  "Manage outpatient registrations, doctor consultations, diagnoses, and follow-ups.",
                )}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            id="btn-refresh-encounters"
            onClick={fetchEncounters}
            disabled={loading}
            className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-lg border border-border bg-background hover:bg-muted transition text-foreground"
            aria-label={t("Refresh")}
          >
            <RefreshCw size={16} className={loading ? "animate-spin" : ""} />
            <span className="hidden sm:inline">{t("Refresh")}</span>
          </button>
          {identity?.permissions.includes("clinical.admit") && (
            <button
              id="btn-new-admission"
              onClick={() => setAdmitModalOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-lg bg-primary text-primary-foreground hover:opacity-90 shadow-sm transition"
            >
              <Plus size={16} />
              <span>
                {kind === "ipd" ? t("New IPD Patient") : t("New OPD Patient")}
              </span>
            </button>
          )}
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-4 bg-card p-4 rounded-xl border border-border shadow-sm">
        <div className="relative flex-1">
          <Search
            className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted-foreground"
            size={18}
          />
          <input
            id="encounter-search-input"
            type="text"
            placeholder={t("Search by patient, doctor, bed, or symptoms...")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 text-sm rounded-lg border border-input bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary/30"
          />
        </div>
        <div className="flex items-center gap-3">
          <label className="text-xs font-semibold uppercase text-muted-foreground tracking-wider">
            {t("Status")}:
          </label>
          <div className="inline-flex p-1 bg-muted rounded-lg border border-border">
            {(["All", "active", "discharged"] as const).map((st) => (
              <button
                key={st}
                id={`filter-status-${st}`}
                onClick={() => setStatusFilter(st)}
                className={`px-3 py-1 text-xs font-medium rounded-md transition ${
                  statusFilter === st
                    ? "bg-background text-foreground shadow-sm font-semibold"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {st === "All"
                  ? t("All")
                  : st === "active"
                    ? t("Active")
                    : t("Discharged")}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Main Table */}
      <div className="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
        {error && (
          <div className="p-4 bg-destructive/10 text-destructive text-sm border-b border-border flex items-center gap-2">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm" id="encounters-table">
            <thead className="bg-muted/50 border-b border-border text-xs uppercase text-muted-foreground font-semibold">
              <tr>
                <th className="px-4 py-3">
                  {kind === "ipd" ? "IPD No." : "OPD No."}
                </th>
                <th className="px-4 py-3">{t("Patient")}</th>
                <th className="px-4 py-3">{t("Consultant Doctor")}</th>
                {kind === "ipd" && (
                  <th className="px-4 py-3">{t("Bed / Ward")}</th>
                )}
                <th className="px-4 py-3">{t("Admission Date")}</th>
                <th className="px-4 py-3">{t("Symptoms / Diagnosis")}</th>
                <th className="px-4 py-3">{t("Bill Status")}</th>
                <th className="px-4 py-3">{t("Status")}</th>
                <th className="px-4 py-3 text-right">{t("Action")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading && encounters.length === 0 ? (
                <tr>
                  <td
                    colSpan={kind === "ipd" ? 9 : 8}
                    className="px-4 py-12 text-center text-muted-foreground"
                  >
                    <div className="flex flex-col items-center justify-center gap-2">
                      <RefreshCw
                        className="animate-spin text-primary"
                        size={24}
                      />
                      <span>{t("Loading encounters...")}</span>
                    </div>
                  </td>
                </tr>
              ) : filteredEncounters.length === 0 ? (
                <tr>
                  <td
                    colSpan={kind === "ipd" ? 9 : 8}
                    className="px-4 py-12 text-center text-muted-foreground"
                  >
                    <div className="flex flex-col items-center justify-center gap-2">
                      <Layers size={32} className="text-muted-foreground/50" />
                      <p className="font-medium text-foreground">
                        {t("No encounters found")}
                      </p>
                      <p className="text-xs">
                        {t("Try adjusting your search query or status filter.")}
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                filteredEncounters.map((enc) => (
                  <tr
                    key={enc.id}
                    id={`encounter-row-${enc.id}`}
                    className="hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-semibold text-primary">
                      {kind.toUpperCase()}-{String(enc.number).padStart(6, "0")}
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium text-foreground">
                        {enc.patientName}
                      </div>
                      {enc.patientEmail && (
                        <div className="text-xs text-muted-foreground">
                          {enc.patientEmail}
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium text-foreground">
                        {enc.doctorName}
                      </div>
                      {enc.doctorEmail && (
                        <div className="text-xs text-muted-foreground">
                          {enc.doctorEmail}
                        </div>
                      )}
                    </td>
                    {kind === "ipd" && (
                      <td className="px-4 py-3">
                        {enc.bedName ? (
                          <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-secondary text-secondary-foreground border border-border">
                            <BedIcon size={14} />
                            <span>{enc.bedName}</span>
                          </div>
                        ) : (
                          <span className="text-muted-foreground text-xs italic">
                            {t("No bed assigned")}
                          </span>
                        )}
                      </td>
                    )}
                    <td className="px-4 py-3 text-muted-foreground text-xs">
                      {new Date(enc.admittedAt).toLocaleDateString("en-GB", {
                        day: "2-digit",
                        month: "short",
                        year: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </td>
                    <td
                      className="px-4 py-3 max-w-[200px] truncate text-foreground"
                      title={enc.symptoms}
                    >
                      {enc.symptoms || "—"}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex px-2 py-0.5 rounded-full text-xs font-semibold ${
                          enc.billStatus === "Paid"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20"
                            : enc.billStatus === "Unpaid"
                              ? "bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20"
                              : "bg-muted text-muted-foreground border border-border"
                        }`}
                      >
                        {enc.billStatus}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                          enc.status === "active"
                            ? "bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20"
                            : "bg-zinc-500/10 text-zinc-600 dark:text-zinc-400 border border-zinc-500/20"
                        }`}
                      >
                        {enc.status === "active"
                          ? t("Active")
                          : t("Discharged")}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        id={`btn-view-encounter-${enc.id}`}
                        onClick={() => setSelectedEncounterId(enc.id)}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border border-border bg-background hover:bg-muted text-foreground transition"
                        title={t("View Detail Register")}
                      >
                        <Eye size={14} />
                        <span>{t("Details")}</span>
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-border bg-muted/20 text-xs text-muted-foreground">
          <span>
            {t("Showing")} {filteredEncounters.length} {t("records")}
          </span>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1 || loading}
              className="p-1.5 rounded border border-border bg-background hover:bg-muted disabled:opacity-50 transition"
              aria-label={t("Previous Page")}
            >
              <ChevronLeft size={16} />
            </button>
            <span className="font-semibold text-foreground px-2">
              {t("Page")} {page}
            </span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={filteredEncounters.length < pageSize || loading}
              className="p-1.5 rounded border border-border bg-background hover:bg-muted disabled:opacity-50 transition"
              aria-label={t("Next Page")}
            >
              <ChevronRight size={16} />
            </button>
          </div>
        </div>
      </div>

      {/* Encounter Detail Modal */}
      {selectedEncounterId && (
        <EncounterDetailModal
          encounterId={selectedEncounterId}
          initialTab={defaultTabFromSlug}
          onClose={() => setSelectedEncounterId(null)}
          onUpdated={fetchEncounters}
        />
      )}

      {/* Admission Modal */}
      {admitModalOpen && (
        <AdmissionModal
          kind={kind}
          onClose={() => setAdmitModalOpen(false)}
          onSuccess={() => {
            setAdmitModalOpen(false);
            fetchEncounters();
          }}
        />
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Detailed Encounter Register Modal with 9 Child Register Tabs
// ─────────────────────────────────────────────────────────────────────────────

interface EncounterDetailModalProps {
  encounterId: string;
  initialTab: string;
  onClose: () => void;
  onUpdated: () => void;
}

function EncounterDetailModal({
  encounterId,
  initialTab,
  onClose,
  onUpdated,
}: EncounterDetailModalProps) {
  const { t } = useLanguage();
  const identity = useIdentity();

  const [encounter, setEncounter] = useState<Encounter | null>(null);
  const [activeTab, setActiveTab] = useState(initialTab);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionSuccess, setActionSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  // Subresource states
  const [diagnoses, setDiagnoses] = useState<Diagnosis[]>([]);
  const [careTeam, setCareTeam] = useState<CareTeamMember[]>([]);
  const [procedures, setProcedures] = useState<Procedure[]>([]);
  const [notes, setNotes] = useState<ClinicalNote[]>([]);
  const [bedHistory, setBedHistory] = useState<BedEvent[]>([]);
  const [bedAssignments, setBedAssignments] = useState<BedAssignment[]>([]);
  const [followUps, setFollowUps] = useState<FollowUp[]>([]);
  const [billingDetails, setBillingDetails] = useState<any>(null);

  // Form states
  const [symptomsInput, setSymptomsInput] = useState("");
  const [bpInput, setBpInput] = useState("");
  const [heightInput, setHeightInput] = useState<number | "">("");
  const [weightInput, setWeightInput] = useState<number | "">("");
  const [noteBody, setNoteBody] = useState("");
  const [dischargeSummaryInput, setDischargeSummaryInput] = useState("");

  // Transfer Bed states
  const [availableBeds, setAvailableBeds] = useState<BedOption[]>([]);
  const [transferBedId, setTransferBedId] = useState("");
  const [transferReason, setTransferReason] = useState("");

  // Add Diagnosis form states
  const [icd10Code, setIcd10Code] = useState("");
  const [diagDescription, setDiagDescription] = useState("");
  const [diagCategory, setDiagCategory] = useState("provisional");

  // Add Care Team form states
  const [teamStaffId, setTeamStaffId] = useState("");
  const [teamRoleTitle, setTeamRoleTitle] = useState("");
  const [teamNotes, setTeamNotes] = useState("");
  const [doctorsList, setDoctorsList] = useState<any[]>([]);

  // Add Procedure form states
  const [procCode, setProcCode] = useState("");
  const [procName, setProcName] = useState("");
  const [procDesc, setProcDesc] = useState("");

  // Add Follow Up form states
  const [followUpDate, setFollowUpDate] = useState("");
  const [followUpNotes, setFollowUpNotes] = useState("");

  const loadEncounter = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await api<Encounter>(`encounters/${encounterId}`);
      setEncounter(data);
      setSymptomsInput(data.symptoms || "");
      setBpInput(data.intake?.bloodPressure || "");
      setHeightInput(data.intake?.height || "");
      setWeightInput(data.intake?.weight || "");
      setDischargeSummaryInput(data.dischargeSummary || "");
    } catch (err) {
      setError(errorText(err));
    } finally {
      setLoading(false);
    }
  }, [encounterId]);

  useEffect(() => {
    loadEncounter();
  }, [loadEncounter]);

  // Load subresources based on active tab
  useEffect(() => {
    if (!encounter) return;

    if (activeTab === "diagnoses") {
      api<{ diagnoses: Diagnosis[] }>(`encounters/${encounter.id}/diagnoses`)
        .then((r) => setDiagnoses(r.diagnoses || []))
        .catch(() => {});
    } else if (activeTab === "care-team") {
      api<{ careTeam: CareTeamMember[] }>(
        `encounters/${encounter.id}/care-team`,
      )
        .then((r) => setCareTeam(r.careTeam || []))
        .catch(() => {});
      api<{ doctors: any[] }>(`doctors?page=1&pageSize=100`)
        .then((r) => setDoctorsList(r.doctors || []))
        .catch(() => {});
    } else if (activeTab === "bed" && encounter.kind === "ipd") {
      api<{ events: BedEvent[] }>(`encounters/${encounter.id}/bed-history`)
        .then((r) => setBedHistory(r.events || []))
        .catch(() => {});
      api<{ assignments: BedAssignment[] }>(
        `encounters/${encounter.id}/bed-assignments`,
      )
        .then((r) => setBedAssignments(r.assignments || []))
        .catch(() => {});
      api<{ beds: BedOption[] }>(`beds?page=1`)
        .then((r) => setAvailableBeds(r.beds?.filter((b) => b.available) || []))
        .catch(() => {});
    } else if (activeTab === "procedures") {
      api<{ procedures: Procedure[] }>(`encounters/${encounter.id}/procedures`)
        .then((r) => setProcedures(r.procedures || []))
        .catch(() => {});
    } else if (activeTab === "notes") {
      api<{ notes: ClinicalNote[] }>(`encounters/${encounter.id}/notes`)
        .then((r) => setNotes(r.notes || []))
        .catch(() => {});
    } else if (activeTab === "billing") {
      api(`encounters/${encounter.id}/billing`)
        .then((r) => setBillingDetails(r))
        .catch(() => {});
    } else if (activeTab === "follow-ups") {
      api<{ followUps: FollowUp[] }>(
        `patients/${encounter.patientId}/follow-ups`,
      )
        .then((r) => setFollowUps(r.followUps || []))
        .catch(() => {});
    }
  }, [activeTab, encounter]);

  // 1. Update Symptoms and Vitals with Optimistic Concurrency
  async function handleUpdateSymptoms(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter) return;
    setBusy(true);
    setActionError("");
    setActionSuccess("");

    try {
      const updated = await api<Encounter>(`encounters/${encounter.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          symptoms: symptomsInput,
          version: encounter.version,
          intake: {
            ...encounter.intake,
            bloodPressure: bpInput,
            height: heightInput === "" ? 0 : Number(heightInput),
            weight: weightInput === "" ? 0 : Number(weightInput),
          },
        }),
      });
      setEncounter(updated);
      setActionSuccess(t("Clinical records updated successfully"));
      onUpdated();
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 2. Add Diagnosis
  async function handleAddDiagnosis(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !icd10Code || !diagDescription) return;
    setBusy(true);
    setActionError("");

    try {
      const diag = await api<Diagnosis>(
        `encounters/${encounter.id}/diagnoses`,
        {
          method: "POST",
          body: JSON.stringify({
            icd10Code,
            description: diagDescription,
            category: diagCategory,
          }),
        },
      );
      setDiagnoses((prev) => [diag, ...prev]);
      setIcd10Code("");
      setDiagDescription("");
      setActionSuccess(t("Diagnosis recorded successfully"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 3. Care Team Member Delegation & Revocation
  async function handleAddCareMember(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !teamStaffId || !teamRoleTitle) return;
    setBusy(true);
    setActionError("");

    try {
      const member = await api<CareTeamMember>(
        `encounters/${encounter.id}/care-team`,
        {
          method: "POST",
          body: JSON.stringify({
            staffId: teamStaffId,
            roleTitle: teamRoleTitle,
            notes: teamNotes,
          }),
        },
      );
      setCareTeam((prev) => [...prev, member]);
      setTeamStaffId("");
      setTeamRoleTitle("");
      setTeamNotes("");
      setActionSuccess(t("Consultant delegated successfully"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  async function handleRevokeCareMember(memberId: string) {
    if (!encounter) return;
    setBusy(true);
    setActionError("");

    try {
      await api(`encounters/${encounter.id}/care-team/${memberId}`, {
        method: "DELETE",
      });
      setCareTeam((prev) =>
        prev.map((m) =>
          m.id === memberId ? { ...m, revokedAt: new Date().toISOString() } : m,
        ),
      );
      setActionSuccess(t("Care team delegation revoked"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 4. Transfer Bed (IPD)
  async function handleTransferBed(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !transferBedId) return;
    setBusy(true);
    setActionError("");

    try {
      await api(`encounters/${encounter.id}/transfer`, {
        method: "POST",
        body: JSON.stringify({
          bedId: transferBedId,
          version: encounter.version,
          reason: transferReason || "Bed transfer requested",
        }),
      });
      setActionSuccess(t("Bed transferred successfully"));
      setTransferBedId("");
      setTransferReason("");
      loadEncounter();
      onUpdated();
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 5. Add Procedure
  async function handleAddProcedure(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !procCode || !procName) return;
    setBusy(true);
    setActionError("");

    try {
      const proc = await api<Procedure>(
        `encounters/${encounter.id}/procedures`,
        {
          method: "POST",
          body: JSON.stringify({
            procedureCode: procCode,
            name: procName,
            description: procDesc,
          }),
        },
      );
      setProcedures((prev) => [proc, ...prev]);
      setProcCode("");
      setProcName("");
      setProcDesc("");
      setActionSuccess(t("Procedure recorded successfully"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 6. Sign Note
  async function handleSignNote(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !noteBody.trim()) return;
    setBusy(true);
    setActionError("");

    try {
      const note = await api<ClinicalNote>(`encounters/${encounter.id}/notes`, {
        method: "POST",
        headers: { "Idempotency-Key": crypto.randomUUID() },
        body: JSON.stringify({ body: noteBody.trim() }),
      });
      setNotes((prev) => [note, ...prev]);
      setNoteBody("");
      setActionSuccess(t("Clinical note signed successfully"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 7. Financial Clearance
  async function handleGrantClearance() {
    if (!encounter) return;
    setBusy(true);
    setActionError("");

    try {
      const res = await api(`encounters/${encounter.id}/billing/clearance`, {
        method: "POST",
        body: JSON.stringify({ waiverReason: "Clinical clearance authorized" }),
      });
      setBillingDetails(res);
      setActionSuccess(t("Financial clearance granted"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 8. Discharge with Atomic Bed Release
  async function handleDischarge(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !dischargeSummaryInput.trim()) return;
    setBusy(true);
    setActionError("");

    try {
      const discharged = await api<Encounter>(
        `encounters/${encounter.id}/discharge`,
        {
          method: "POST",
          body: JSON.stringify({
            version: encounter.version,
            summary: dischargeSummaryInput.trim(),
          }),
        },
      );
      setEncounter(discharged);
      setActionSuccess(
        t("Patient discharged and bed released to ready status successfully"),
      );
      onUpdated();
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  // 9. Add Follow-Up
  async function handleAddFollowUp(e: React.FormEvent) {
    e.preventDefault();
    if (!encounter || !followUpDate) return;
    setBusy(true);
    setActionError("");

    try {
      const fu = await api<FollowUp>(
        `patients/${encounter.patientId}/follow-ups`,
        {
          method: "POST",
          body: JSON.stringify({
            encounterId: encounter.id,
            followUpDate,
            notes: followUpNotes,
          }),
        },
      );
      setFollowUps((prev) => [fu, ...prev]);
      setFollowUpDate("");
      setFollowUpNotes("");
      setActionSuccess(t("Follow-up appointment scheduled"));
    } catch (err) {
      setActionError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  const tabs = [
    { id: "overview", label: t("Overview & Vitals"), icon: Stethoscope },
    { id: "diagnoses", label: t("Diagnoses (ICD-10)"), icon: Layers },
    { id: "care-team", label: t("Care Team"), icon: UserCheck },
    ...(encounter?.kind === "ipd"
      ? [{ id: "bed", label: t("Bed & Transfers"), icon: BedIcon }]
      : []),
    { id: "procedures", label: t("Procedures"), icon: FileText },
    { id: "notes", label: t("Notes & Timelines"), icon: Clock },
    { id: "billing", label: t("Billing & Clearance"), icon: DollarSign },
    { id: "discharge", label: t("Discharge Summary"), icon: LogOut },
    ...(encounter?.kind === "opd"
      ? [{ id: "follow-ups", label: t("Follow-ups"), icon: Calendar }]
      : []),
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 overflow-y-auto">
      <div className="bg-card text-foreground rounded-2xl border border-border shadow-2xl w-full max-w-5xl max-h-[90vh] flex flex-col overflow-hidden animate-in fade-in-50 zoom-in-95">
        {/* Modal Header */}
        <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-muted/40">
          <div className="flex items-center gap-3">
            <span className="p-2 rounded-lg bg-primary/10 text-primary">
              <Stethoscope size={20} />
            </span>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold">
                  {encounter
                    ? `${encounter.patientName} (${encounter.kind.toUpperCase()}-${String(encounter.number).padStart(6, "0")})`
                    : t("Loading encounter...")}
                </h2>
                {encounter && (
                  <span
                    className={`inline-flex px-2 py-0.5 rounded-full text-xs font-semibold ${
                      encounter.status === "active"
                        ? "bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20"
                        : "bg-zinc-500/10 text-zinc-600 dark:text-zinc-400 border border-zinc-500/20"
                    }`}
                  >
                    {encounter.status === "active"
                      ? t("Active")
                      : t("Discharged")}
                  </span>
                )}
                {encounter && (
                  <span className="text-xs text-muted-foreground border border-border px-2 py-0.5 rounded">
                    v{encounter.version}
                  </span>
                )}
              </div>
              {encounter && (
                <p className="text-xs text-muted-foreground mt-0.5">
                  {t("Doctor")}: {encounter.doctorName} • {t("Admitted")}:{" "}
                  {new Date(encounter.admittedAt).toLocaleDateString()}
                  {encounter.bedName && ` • ${t("Bed")}: ${encounter.bedName}`}
                </p>
              )}
            </div>
          </div>
          <button
            id="btn-close-encounter-modal"
            onClick={onClose}
            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition"
            aria-label={t("Close")}
          >
            <X size={20} />
          </button>
        </div>

        {/* Action feedback banners */}
        {actionError && (
          <div className="px-6 py-2.5 bg-destructive/10 text-destructive text-sm flex items-center gap-2 border-b border-border">
            <AlertCircle size={16} />
            <span>{actionError}</span>
          </div>
        )}
        {actionSuccess && (
          <div className="px-6 py-2.5 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-sm flex items-center gap-2 border-b border-border">
            <ShieldCheck size={16} />
            <span>{actionSuccess}</span>
          </div>
        )}

        {/* Modal Navigation Tabs */}
        <div className="flex border-b border-border px-6 overflow-x-auto bg-background shrink-0 scrollbar-none">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                id={`tab-${tab.id}`}
                onClick={() => {
                  setActiveTab(tab.id);
                  setActionError("");
                  setActionSuccess("");
                }}
                className={`flex items-center gap-2 py-3 px-4 text-xs font-semibold border-b-2 transition whitespace-nowrap ${
                  activeTab === tab.id
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
                }`}
              >
                <Icon size={16} />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>

        {/* Modal Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {loading && !encounter ? (
            <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
              <RefreshCw className="animate-spin text-primary" size={28} />
              <span>{t("Loading clinical register...")}</span>
            </div>
          ) : !encounter ? (
            <div className="text-center py-16 text-destructive">
              {error || t("Encounter not found")}
            </div>
          ) : (
            <>
              {/* TAB 1: OVERVIEW & VITALS */}
              {activeTab === "overview" && (
                <div className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="p-4 rounded-xl border border-border bg-card">
                      <span className="text-xs text-muted-foreground">
                        {t("Blood Pressure")}
                      </span>
                      <p className="text-xl font-bold mt-1 text-foreground">
                        {encounter.intake?.bloodPressure || "120/80"}
                      </p>
                    </div>
                    <div className="p-4 rounded-xl border border-border bg-card">
                      <span className="text-xs text-muted-foreground">
                        {t("Height / Weight")}
                      </span>
                      <p className="text-xl font-bold mt-1 text-foreground">
                        {encounter.intake?.height
                          ? `${encounter.intake.height} cm`
                          : "—"}{" "}
                        /{" "}
                        {encounter.intake?.weight
                          ? `${encounter.intake.weight} kg`
                          : "—"}
                      </p>
                    </div>
                    <div className="p-4 rounded-xl border border-border bg-card">
                      <span className="text-xs text-muted-foreground">
                        {t("Standard Charge")}
                      </span>
                      <p className="text-xl font-bold mt-1 text-primary">
                        {(
                          (encounter.intake?.standardChargeMinor || 0) / 100
                        ).toFixed(2)}{" "}
                        ETB
                      </p>
                    </div>
                  </div>

                  {/* Editable Clinical Records Form */}
                  <form
                    onSubmit={handleUpdateSymptoms}
                    className="p-5 rounded-xl border border-border bg-card space-y-4"
                  >
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-bold text-foreground">
                        {t("Symptoms & Intake Records")}
                      </h3>
                      <span className="text-xs text-muted-foreground">
                        {t("Protected by optimistic locking")} (v
                        {encounter.version})
                      </span>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                      <div>
                        <label className="text-xs font-semibold text-muted-foreground">
                          {t("Blood Pressure")}
                        </label>
                        <input
                          id="input-intake-bp"
                          type="text"
                          value={bpInput}
                          onChange={(e) => setBpInput(e.target.value)}
                          placeholder="e.g. 120/80"
                          className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                        />
                      </div>
                      <div>
                        <label className="text-xs font-semibold text-muted-foreground">
                          {t("Height (cm)")}
                        </label>
                        <input
                          id="input-intake-height"
                          type="number"
                          step="0.1"
                          value={heightInput}
                          onChange={(e) =>
                            setHeightInput(
                              e.target.value === ""
                                ? ""
                                : Number(e.target.value),
                            )
                          }
                          placeholder="cm"
                          className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                        />
                      </div>
                      <div>
                        <label className="text-xs font-semibold text-muted-foreground">
                          {t("Weight (kg)")}
                        </label>
                        <input
                          id="input-intake-weight"
                          type="number"
                          step="0.1"
                          value={weightInput}
                          onChange={(e) =>
                            setWeightInput(
                              e.target.value === ""
                                ? ""
                                : Number(e.target.value),
                            )
                          }
                          placeholder="kg"
                          className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                        />
                      </div>
                    </div>

                    <div>
                      <label className="text-xs font-semibold text-muted-foreground">
                        {t("Presenting Symptoms")}
                      </label>
                      <textarea
                        id="input-encounter-symptoms"
                        rows={3}
                        value={symptomsInput}
                        onChange={(e) => setSymptomsInput(e.target.value)}
                        placeholder={t(
                          "Detailed patient symptoms, chief complaint...",
                        )}
                        className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                      />
                    </div>

                    <div className="flex justify-end">
                      <button
                        id="btn-save-symptoms"
                        type="submit"
                        disabled={busy || encounter.status !== "active"}
                        className="px-4 py-2 text-sm font-semibold rounded-lg bg-primary text-primary-foreground hover:opacity-90 disabled:opacity-50 transition"
                      >
                        {busy ? t("Saving…") : t("Save Changes")}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {/* TAB 2: DIAGNOSES (ICD-10) */}
              {activeTab === "diagnoses" && (
                <div className="space-y-6">
                  {/* Add Diagnosis Form */}
                  {encounter.status === "active" && (
                    <form
                      onSubmit={handleAddDiagnosis}
                      className="p-4 rounded-xl border border-border bg-card space-y-3"
                    >
                      <h4 className="text-sm font-bold">
                        {t("Add Diagnosis (ICD-10)")}
                      </h4>
                      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("ICD-10 Code")}
                          </label>
                          <input
                            id="input-diag-code"
                            type="text"
                            placeholder="e.g. I10, J18.9"
                            value={icd10Code}
                            onChange={(e) => setIcd10Code(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Category")}
                          </label>
                          <select
                            id="select-diag-category"
                            value={diagCategory}
                            onChange={(e) => setDiagCategory(e.target.value)}
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          >
                            <option value="provisional">
                              {t("Provisional")}
                            </option>
                            <option value="final">{t("Final")}</option>
                            <option value="differential">
                              {t("Differential")}
                            </option>
                          </select>
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Description")}
                          </label>
                          <input
                            id="input-diag-desc"
                            type="text"
                            placeholder="e.g. Essential hypertension"
                            value={diagDescription}
                            onChange={(e) => setDiagDescription(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <button
                          id="btn-add-diagnosis"
                          type="submit"
                          disabled={busy}
                          className="px-4 py-1.5 text-xs font-semibold rounded bg-primary text-primary-foreground hover:opacity-90 transition"
                        >
                          {busy ? t("Adding…") : t("Add Diagnosis")}
                        </button>
                      </div>
                    </form>
                  )}

                  {/* Diagnoses List */}
                  <div className="rounded-xl border border-border overflow-hidden">
                    <table
                      className="w-full text-left text-sm"
                      id="diagnoses-table"
                    >
                      <thead className="bg-muted/50 border-b border-border text-xs uppercase text-muted-foreground">
                        <tr>
                          <th className="px-4 py-2.5">{t("ICD-10 Code")}</th>
                          <th className="px-4 py-2.5">{t("Description")}</th>
                          <th className="px-4 py-2.5">{t("Category")}</th>
                          <th className="px-4 py-2.5">{t("Date Diagnosed")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {diagnoses.length === 0 ? (
                          <tr>
                            <td
                              colSpan={4}
                              className="px-4 py-8 text-center text-muted-foreground text-xs"
                            >
                              {t(
                                "No diagnoses recorded for this encounter yet.",
                              )}
                            </td>
                          </tr>
                        ) : (
                          diagnoses.map((d) => (
                            <tr key={d.id} className="hover:bg-muted/20">
                              <td className="px-4 py-2.5 font-bold text-primary">
                                {d.icd10Code}
                              </td>
                              <td className="px-4 py-2.5">{d.description}</td>
                              <td className="px-4 py-2.5">
                                <span className="inline-flex px-2 py-0.5 rounded text-xs bg-muted font-medium">
                                  {d.category}
                                </span>
                              </td>
                              <td className="px-4 py-2.5 text-xs text-muted-foreground">
                                {new Date(d.diagnosedAt).toLocaleDateString()}
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* TAB 3: CARE TEAM DELEGATION */}
              {activeTab === "care-team" && (
                <div className="space-y-6">
                  {encounter.status === "active" && (
                    <form
                      onSubmit={handleAddCareMember}
                      className="p-4 rounded-xl border border-border bg-card space-y-3"
                    >
                      <h4 className="text-sm font-bold">
                        {t("Delegate Care Team Member / Consultant")}
                      </h4>
                      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Doctor / Staff")}
                          </label>
                          <select
                            id="select-care-staff"
                            value={teamStaffId}
                            onChange={(e) => setTeamStaffId(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          >
                            <option value="">{t("Select doctor...")}</option>
                            {doctorsList.map((doc) => (
                              <option key={doc.id} value={doc.id}>
                                {doc.name || doc.email}
                              </option>
                            ))}
                          </select>
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Role Title")}
                          </label>
                          <input
                            id="input-care-role"
                            type="text"
                            placeholder="e.g. Consultant Cardiologist, Primary Nurse"
                            value={teamRoleTitle}
                            onChange={(e) => setTeamRoleTitle(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Notes")}
                          </label>
                          <input
                            id="input-care-notes"
                            type="text"
                            placeholder="e.g. Consult on ECG findings"
                            value={teamNotes}
                            onChange={(e) => setTeamNotes(e.target.value)}
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <button
                          id="btn-add-care-member"
                          type="submit"
                          disabled={busy}
                          className="px-4 py-1.5 text-xs font-semibold rounded bg-primary text-primary-foreground hover:opacity-90 transition"
                        >
                          {busy ? t("Delegating…") : t("Delegate Member")}
                        </button>
                      </div>
                    </form>
                  )}

                  <div className="rounded-xl border border-border overflow-hidden">
                    <table
                      className="w-full text-left text-sm"
                      id="care-team-table"
                    >
                      <thead className="bg-muted/50 border-b border-border text-xs uppercase text-muted-foreground">
                        <tr>
                          <th className="px-4 py-2.5">{t("Staff Name")}</th>
                          <th className="px-4 py-2.5">{t("Role Title")}</th>
                          <th className="px-4 py-2.5">{t("Assigned Date")}</th>
                          <th className="px-4 py-2.5">{t("Status")}</th>
                          <th className="px-4 py-2.5 text-right">
                            {t("Action")}
                          </th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {careTeam.length === 0 ? (
                          <tr>
                            <td
                              colSpan={5}
                              className="px-4 py-8 text-center text-muted-foreground text-xs"
                            >
                              {t("No delegated care team members recorded.")}
                            </td>
                          </tr>
                        ) : (
                          careTeam.map((m) => (
                            <tr key={m.id} className="hover:bg-muted/20">
                              <td className="px-4 py-2.5 font-medium">
                                {m.staffName}
                              </td>
                              <td className="px-4 py-2.5">{m.roleTitle}</td>
                              <td className="px-4 py-2.5 text-xs text-muted-foreground">
                                {new Date(m.assignedAt).toLocaleDateString()}
                              </td>
                              <td className="px-4 py-2.5">
                                <span
                                  className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${
                                    m.revokedAt
                                      ? "bg-zinc-500/10 text-zinc-500 border border-zinc-500/20"
                                      : "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20"
                                  }`}
                                >
                                  {m.revokedAt ? t("Revoked") : t("Active")}
                                </span>
                              </td>
                              <td className="px-4 py-2.5 text-right">
                                {!m.revokedAt &&
                                  encounter.status === "active" && (
                                    <button
                                      id={`btn-revoke-member-${m.id}`}
                                      onClick={() =>
                                        handleRevokeCareMember(m.id)
                                      }
                                      disabled={busy}
                                      className="px-2.5 py-1 text-xs font-medium rounded bg-destructive/10 text-destructive hover:bg-destructive/20 transition"
                                    >
                                      {t("Revoke")}
                                    </button>
                                  )}
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* TAB 4: BED & TRANSFERS (IPD) */}
              {activeTab === "bed" && encounter.kind === "ipd" && (
                <div className="space-y-6">
                  {/* Current Bed Widget */}
                  <div className="p-4 rounded-xl border border-border bg-card flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <div className="flex items-center gap-3">
                      <div className="p-3 rounded-xl bg-primary/10 text-primary">
                        <BedIcon size={24} />
                      </div>
                      <div>
                        <span className="text-xs text-muted-foreground">
                          {t("Current Bed Assignment")}
                        </span>
                        <h3 className="text-lg font-bold text-foreground">
                          {encounter.bedName || t("None Assigned")}
                        </h3>
                      </div>
                    </div>
                  </div>

                  {/* Transfer Bed Form */}
                  {encounter.status === "active" && (
                    <form
                      onSubmit={handleTransferBed}
                      className="p-4 rounded-xl border border-border bg-card space-y-3"
                    >
                      <h4 className="text-sm font-bold flex items-center gap-2">
                        <ArrowRightLeft size={16} />
                        <span>{t("Transfer Bed")}</span>
                      </h4>
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Target Available Bed")}
                          </label>
                          <select
                            id="select-transfer-bed"
                            value={transferBedId}
                            onChange={(e) => setTransferBedId(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          >
                            <option value="">{t("Select ready bed...")}</option>
                            {availableBeds.map((b) => (
                              <option key={b.id} value={b.id}>
                                {b.name} ({b.wardName || "General Ward"}) —{" "}
                                {(b.chargeMinor / 100).toFixed(2)} ETB/day
                              </option>
                            ))}
                          </select>
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Transfer Reason")}
                          </label>
                          <input
                            id="input-transfer-reason"
                            type="text"
                            placeholder="e.g. Stepped down to general ward"
                            value={transferReason}
                            onChange={(e) => setTransferReason(e.target.value)}
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <button
                          id="btn-confirm-transfer"
                          type="submit"
                          disabled={busy}
                          className="px-4 py-1.5 text-xs font-semibold rounded bg-primary text-primary-foreground hover:opacity-90 transition"
                        >
                          {busy ? t("Transferring…") : t("Confirm Transfer")}
                        </button>
                      </div>
                    </form>
                  )}

                  {/* Bed Assignment Timeline */}
                  <div className="rounded-xl border border-border overflow-hidden">
                    <div className="px-4 py-2.5 bg-muted/50 border-b border-border text-xs font-bold uppercase text-muted-foreground">
                      {t("Bed Assignment History")}
                    </div>
                    <table
                      className="w-full text-left text-sm"
                      id="bed-assignments-table"
                    >
                      <thead className="bg-muted/30 border-b border-border text-xs text-muted-foreground">
                        <tr>
                          <th className="px-4 py-2">{t("Bed Name")}</th>
                          <th className="px-4 py-2">{t("Assigned From")}</th>
                          <th className="px-4 py-2">{t("Assigned To")}</th>
                          <th className="px-4 py-2">{t("Notes")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {bedAssignments.length === 0 ? (
                          <tr>
                            <td
                              colSpan={4}
                              className="px-4 py-6 text-center text-muted-foreground text-xs"
                            >
                              {t("No bed assignment history records found.")}
                            </td>
                          </tr>
                        ) : (
                          bedAssignments.map((ba) => (
                            <tr key={ba.id} className="hover:bg-muted/20">
                              <td className="px-4 py-2.5 font-bold">
                                {ba.bedName}
                              </td>
                              <td className="px-4 py-2.5 text-xs text-muted-foreground">
                                {new Date(ba.assignedFrom).toLocaleDateString()}
                              </td>
                              <td className="px-4 py-2.5 text-xs text-muted-foreground">
                                {ba.assignedTo
                                  ? new Date(ba.assignedTo).toLocaleDateString()
                                  : t("Current")}
                              </td>
                              <td className="px-4 py-2.5 text-xs">
                                {ba.notes || "—"}
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* TAB 5: PROCEDURES */}
              {activeTab === "procedures" && (
                <div className="space-y-6">
                  {encounter.status === "active" && (
                    <form
                      onSubmit={handleAddProcedure}
                      className="p-4 rounded-xl border border-border bg-card space-y-3"
                    >
                      <h4 className="text-sm font-bold">
                        {t("Record Clinical Procedure")}
                      </h4>
                      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Procedure Code")}
                          </label>
                          <input
                            id="input-proc-code"
                            type="text"
                            placeholder="e.g. CPT-99214, SURG-01"
                            value={procCode}
                            onChange={(e) => setProcCode(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Procedure Name")}
                          </label>
                          <input
                            id="input-proc-name"
                            type="text"
                            placeholder="e.g. Wound Debridement"
                            value={procName}
                            onChange={(e) => setProcName(e.target.value)}
                            required
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-muted-foreground">
                            {t("Notes / Details")}
                          </label>
                          <input
                            id="input-proc-desc"
                            type="text"
                            placeholder="e.g. Performed under local anesthesia"
                            value={procDesc}
                            onChange={(e) => setProcDesc(e.target.value)}
                            className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                          />
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <button
                          id="btn-add-procedure"
                          type="submit"
                          disabled={busy}
                          className="px-4 py-1.5 text-xs font-semibold rounded bg-primary text-primary-foreground hover:opacity-90 transition"
                        >
                          {busy ? t("Recording…") : t("Record Procedure")}
                        </button>
                      </div>
                    </form>
                  )}

                  <div className="rounded-xl border border-border overflow-hidden">
                    <table
                      className="w-full text-left text-sm"
                      id="procedures-table"
                    >
                      <thead className="bg-muted/50 border-b border-border text-xs uppercase text-muted-foreground">
                        <tr>
                          <th className="px-4 py-2.5">{t("Code")}</th>
                          <th className="px-4 py-2.5">{t("Procedure Name")}</th>
                          <th className="px-4 py-2.5">{t("Details")}</th>
                          <th className="px-4 py-2.5">{t("Performed At")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {procedures.length === 0 ? (
                          <tr>
                            <td
                              colSpan={4}
                              className="px-4 py-8 text-center text-muted-foreground text-xs"
                            >
                              {t(
                                "No procedures recorded for this encounter yet.",
                              )}
                            </td>
                          </tr>
                        ) : (
                          procedures.map((p) => (
                            <tr key={p.id} className="hover:bg-muted/20">
                              <td className="px-4 py-2.5 font-bold text-primary">
                                {p.procedureCode}
                              </td>
                              <td className="px-4 py-2.5 font-medium">
                                {p.name}
                              </td>
                              <td className="px-4 py-2.5">
                                {p.description || "—"}
                              </td>
                              <td className="px-4 py-2.5 text-xs text-muted-foreground">
                                {new Date(p.performedAt).toLocaleDateString()}
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* TAB 6: NOTES & TIMELINES */}
              {activeTab === "notes" && (
                <div className="space-y-6">
                  {encounter.status === "active" && (
                    <form
                      onSubmit={handleSignNote}
                      className="p-4 rounded-xl border border-border bg-card space-y-3"
                    >
                      <h4 className="text-sm font-bold">
                        {t("Sign Clinical Progress Note")}
                      </h4>
                      <textarea
                        id="input-note-body"
                        rows={3}
                        value={noteBody}
                        onChange={(e) => setNoteBody(e.target.value)}
                        placeholder={t(
                          "Enter physician observations, plan, examination findings...",
                        )}
                        required
                        className="w-full px-3 py-2 text-sm rounded-lg border border-input bg-background"
                      />
                      <div className="flex justify-end">
                        <button
                          id="btn-sign-note"
                          type="submit"
                          disabled={busy || !noteBody.trim()}
                          className="px-4 py-1.5 text-xs font-semibold rounded bg-primary text-primary-foreground hover:opacity-90 transition"
                        >
                          {busy ? t("Signing…") : t("Sign Note")}
                        </button>
                      </div>
                    </form>
                  )}

                  <div className="space-y-3">
                    <h4 className="text-xs font-semibold uppercase text-muted-foreground tracking-wider">
                      {t("Signed Progress Notes")}
                    </h4>
                    {notes.length === 0 ? (
                      <p className="text-xs text-muted-foreground py-4 text-center">
                        {t("No signed notes found.")}
                      </p>
                    ) : (
                      notes.map((n) => (
                        <div
                          key={n.id}
                          className="p-4 rounded-xl border border-border bg-card space-y-1"
                        >
                          <div className="flex items-center justify-between text-xs text-muted-foreground">
                            <span className="font-semibold text-foreground">
                              {n.authorName}
                            </span>
                            <span>{new Date(n.signedAt).toLocaleString()}</span>
                          </div>
                          <p className="text-sm text-foreground whitespace-pre-wrap mt-2">
                            {n.body}
                          </p>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              )}

              {/* TAB 7: BILLING & FINANCIAL CLEARANCE */}
              {activeTab === "billing" && (
                <div className="space-y-6">
                  <div className="p-5 rounded-xl border border-border bg-card flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <div>
                      <span className="text-xs text-muted-foreground">
                        {t("Financial Clearance Status")}
                      </span>
                      <div className="flex items-center gap-2 mt-1">
                        <span
                          className={`inline-flex px-3 py-1 rounded-full text-xs font-bold ${
                            billingDetails?.financialClearance
                              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20"
                              : "bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20"
                          }`}
                        >
                          {billingDetails?.financialClearance
                            ? t("Cleared for Discharge")
                            : t("Pending Clearance")}
                        </span>
                        {billingDetails?.financialClearance && (
                          <span className="text-xs text-muted-foreground">
                            by {billingDetails.clearedBy}
                          </span>
                        )}
                      </div>
                    </div>
                    {!billingDetails?.financialClearance &&
                      encounter.status === "active" && (
                        <button
                          id="btn-grant-clearance"
                          onClick={handleGrantClearance}
                          disabled={busy}
                          className="px-4 py-2 text-xs font-semibold rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 shadow-sm transition"
                        >
                          {t("Grant Clearance")}
                        </button>
                      )}
                  </div>

                  <div className="p-4 rounded-xl border border-border bg-card space-y-3">
                    <h4 className="text-sm font-bold">
                      {t("Charges & Invoices Summary")}
                    </h4>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs">
                      <div>
                        <span className="text-muted-foreground">
                          {t("Standard Registration Charge")}:
                        </span>
                        <p className="font-bold text-foreground mt-1">
                          {(
                            (encounter.intake?.standardChargeMinor || 0) / 100
                          ).toFixed(2)}{" "}
                          ETB
                        </p>
                      </div>
                      <div>
                        <span className="text-muted-foreground">
                          {t("Payment Mode")}:
                        </span>
                        <p className="font-bold text-foreground mt-1 capitalize">
                          {encounter.intake?.paymentMode || "Cash"}
                        </p>
                      </div>
                      <div>
                        <span className="text-muted-foreground">
                          {t("Invoice Status")}:
                        </span>
                        <p className="font-bold text-foreground mt-1">
                          {encounter.billStatus}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* TAB 8: DISCHARGE SUMMARY */}
              {activeTab === "discharge" && (
                <div className="space-y-6">
                  {encounter.status === "discharged" ? (
                    <div className="p-6 rounded-xl border border-border bg-card space-y-4">
                      <div className="flex items-center gap-2 text-emerald-600 dark:text-emerald-400 font-bold">
                        <ShieldCheck size={20} />
                        <span>{t("Patient Discharged Successfully")}</span>
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {t("Discharged at")}:{" "}
                        {encounter.dischargedAt
                          ? new Date(encounter.dischargedAt).toLocaleString()
                          : "—"}
                      </div>
                      <div className="p-4 rounded-lg bg-muted/40 border border-border">
                        <span className="text-xs font-semibold text-muted-foreground">
                          {t("Discharge Summary")}:
                        </span>
                        <p className="text-sm text-foreground mt-1 whitespace-pre-wrap">
                          {encounter.dischargeSummary}
                        </p>
                      </div>
                    </div>
                  ) : (
                    <form
                      onSubmit={handleDischarge}
                      className="p-6 rounded-xl border border-border bg-card space-y-4"
                    >
                      <div>
                        <h4 className="text-base font-bold text-foreground">
                          {t("Patient Discharge Workflow")}
                        </h4>
                        <p className="text-xs text-muted-foreground mt-1">
                          {t(
                            "Discharging the patient marks the encounter complete, closes any open bed assignments, and atomically releases the assigned bed back to 'ready' status.",
                          )}
                        </p>
                      </div>

                      <div>
                        <label className="text-xs font-semibold text-muted-foreground">
                          {t("Comprehensive Discharge Summary")} *
                        </label>
                        <textarea
                          id="input-discharge-summary"
                          rows={4}
                          value={dischargeSummaryInput}
                          onChange={(e) =>
                            setDischargeSummaryInput(e.target.value)
                          }
                          placeholder={t(
                            "Discharge condition, instructions, follow-up plan, medication summary...",
                          )}
                          required
                          className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                        />
                      </div>

                      <div className="flex items-center justify-between pt-2 border-t border-border">
                        <span className="text-xs text-muted-foreground">
                          {t("Encounter Concurrency Version")}: v
                          {encounter.version}
                        </span>
                        <button
                          id="btn-confirm-discharge"
                          type="submit"
                          disabled={busy || !dischargeSummaryInput.trim()}
                          className="px-5 py-2 text-sm font-semibold rounded-lg bg-destructive text-destructive-foreground hover:opacity-90 shadow-sm transition"
                        >
                          {busy
                            ? t("Discharging…")
                            : t("Confirm Discharge & Release Bed")}
                        </button>
                      </div>
                    </form>
                  )}
                </div>
              )}

              {/* TAB 9: FOLLOW-UPS (OPD) */}
              {activeTab === "follow-ups" && encounter.kind === "opd" && (
                <div className="space-y-6">
                  <form
                    onSubmit={handleAddFollowUp}
                    className="p-4 rounded-xl border border-border bg-card space-y-3"
                  >
                    <h4 className="text-sm font-bold">
                      {t("Schedule OPD Follow-up Visit")}
                    </h4>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <label className="text-xs font-medium text-muted-foreground">
                          {t("Follow-up Date")}
                        </label>
                        <input
                          id="input-followup-date"
                          type="date"
                          value={followUpDate}
                          onChange={(e) => setFollowUpDate(e.target.value)}
                          required
                          className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                        />
                      </div>
                      <div>
                        <label className="text-xs font-medium text-muted-foreground">
                          {t("Clinical Instructions")}
                        </label>
                        <input
                          id="input-followup-notes"
                          type="text"
                          placeholder="e.g. Check blood pressure and repeat HbA1c"
                          value={followUpNotes}
                          onChange={(e) => setFollowUpNotes(e.target.value)}
                          className="w-full mt-1 px-3 py-1.5 text-sm rounded border border-input bg-background"
                        />
                      </div>
                    </div>
                    <div className="flex justify-end">
                      <button
                        id="btn-add-followup"
                        type="submit"
                        disabled={busy}
                        className="px-4 py-1.5 text-xs font-semibold rounded bg-primary text-primary-foreground hover:opacity-90 transition"
                      >
                        {busy ? t("Scheduling…") : t("Schedule Follow-up")}
                      </button>
                    </div>
                  </form>

                  <div className="rounded-xl border border-border overflow-hidden">
                    <table
                      className="w-full text-left text-sm"
                      id="followups-table"
                    >
                      <thead className="bg-muted/50 border-b border-border text-xs uppercase text-muted-foreground">
                        <tr>
                          <th className="px-4 py-2.5">{t("Scheduled Date")}</th>
                          <th className="px-4 py-2.5">
                            {t("Notes / Purpose")}
                          </th>
                          <th className="px-4 py-2.5">{t("Status")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {followUps.length === 0 ? (
                          <tr>
                            <td
                              colSpan={3}
                              className="px-4 py-6 text-center text-muted-foreground text-xs"
                            >
                              {t("No follow-ups scheduled.")}
                            </td>
                          </tr>
                        ) : (
                          followUps.map((fu) => (
                            <tr key={fu.id} className="hover:bg-muted/20">
                              <td className="px-4 py-2.5 font-bold">
                                {new Date(fu.followUpDate).toLocaleDateString()}
                              </td>
                              <td className="px-4 py-2.5">{fu.notes || "—"}</td>
                              <td className="px-4 py-2.5 capitalize">
                                {fu.status}
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Patient Admission Modal for IPD and OPD
// ─────────────────────────────────────────────────────────────────────────────

interface AdmissionModalProps {
  kind: "ipd" | "opd";
  onClose: () => void;
  onSuccess: () => void;
}

function AdmissionModal({ kind, onClose, onSuccess }: AdmissionModalProps) {
  const { t } = useLanguage();

  const [cases, setCases] = useState<Case[]>([]);
  const [beds, setBeds] = useState<BedOption[]>([]);
  const [selectedCaseId, setSelectedCaseId] = useState("");
  const [selectedBedId, setSelectedBedId] = useState("");
  const [symptoms, setSymptoms] = useState("");
  const [bloodPressure, setBloodPressure] = useState("120/80");
  const [height, setHeight] = useState<number | "">("");
  const [weight, setWeight] = useState<number | "">("");
  const [phone, setPhone] = useState("+251911000000");
  const [paymentMode, setPaymentMode] = useState("cash");
  const [standardCharge, setStandardCharge] = useState("500.00");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [loadingInitial, setLoadingInitial] = useState(true);

  useEffect(() => {
    async function loadData() {
      setLoadingInitial(true);
      try {
        const casesRes = await api<{ cases: Case[] }>("cases?page=1");
        setCases(casesRes.cases || []);
        if (casesRes.cases && casesRes.cases.length > 0) {
          setSelectedCaseId(casesRes.cases[0].id);
        }

        if (kind === "ipd") {
          const bedsRes = await api<{ beds: BedOption[] }>("beds?page=1");
          const readyBeds = (bedsRes.beds || []).filter((b) => b.available);
          setBeds(readyBeds);
          if (readyBeds.length > 0) {
            setSelectedBedId(readyBeds[0].id);
          }
        }
      } catch (err) {
        setError(errorText(err));
      } finally {
        setLoadingInitial(false);
      }
    }
    loadData();
  }, [kind]);

  async function handleAdmit(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedCaseId) {
      setError(t("Please select a patient case."));
      return;
    }
    if (kind === "ipd" && !selectedBedId) {
      setError(t("Please select an available bed for IPD admission."));
      return;
    }

    setBusy(true);
    setError("");

    try {
      const chargeMinor = Math.round(Number(standardCharge || "0") * 100);
      const idempotencyKey = crypto.randomUUID();

      await api("encounters", {
        method: "POST",
        headers: { "Idempotency-Key": idempotencyKey },
        body: JSON.stringify({
          kind,
          caseId: selectedCaseId,
          bedId: kind === "ipd" ? selectedBedId : "",
          admittedAt: new Date().toISOString(),
          symptoms: symptoms.trim(),
          intake: {
            telephone: phone,
            bloodPressure,
            height: height === "" ? 0 : Number(height),
            weight: weight === "" ? 0 : Number(weight),
            paymentMode,
            standardChargeMinor: chargeMinor,
          },
        }),
      });

      onSuccess();
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 overflow-y-auto">
      <div className="bg-card text-foreground rounded-2xl border border-border shadow-2xl w-full max-w-xl max-h-[90vh] flex flex-col overflow-hidden animate-in fade-in-50 zoom-in-95">
        <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-muted/40">
          <h2 className="text-lg font-bold">
            {kind === "ipd"
              ? t("New IPD Patient Admission")
              : t("New OPD Patient Registration")}
          </h2>
          <button
            id="btn-close-admit-modal"
            onClick={onClose}
            className="p-1 rounded text-muted-foreground hover:text-foreground"
          >
            <X size={20} />
          </button>
        </div>

        {error && (
          <div className="px-6 py-2.5 bg-destructive/10 text-destructive text-sm flex items-center gap-2 border-b border-border">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleAdmit} className="p-6 space-y-4 overflow-y-auto">
          {loadingInitial ? (
            <div className="py-8 text-center text-muted-foreground text-sm">
              <RefreshCw className="animate-spin inline mr-2" size={16} />
              {t("Loading cases and available beds...")}
            </div>
          ) : (
            <>
              <div>
                <label className="text-xs font-semibold text-muted-foreground">
                  {t("Select Patient Case")} *
                </label>
                <select
                  id="select-admit-case"
                  value={selectedCaseId}
                  onChange={(e) => setSelectedCaseId(e.target.value)}
                  required
                  className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                >
                  {cases.map((c) => (
                    <option key={c.id} value={c.id}>
                      CASE-{String(c.number).padStart(6, "0")} — {c.patientName}{" "}
                      (Dr. {c.doctorName})
                    </option>
                  ))}
                </select>
              </div>

              {kind === "ipd" && (
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Select Available Bed")} *
                  </label>
                  <select
                    id="select-admit-bed"
                    value={selectedBedId}
                    onChange={(e) => setSelectedBedId(e.target.value)}
                    required
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  >
                    {beds.length === 0 ? (
                      <option value="">{t("No available beds")}</option>
                    ) : (
                      beds.map((b) => (
                        <option key={b.id} value={b.id}>
                          {b.name} ({b.wardName || "General Ward"}) —{" "}
                          {(b.chargeMinor / 100).toFixed(2)} ETB
                        </option>
                      ))
                    )}
                  </select>
                </div>
              )}

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Blood Pressure")}
                  </label>
                  <input
                    id="input-admit-bp"
                    type="text"
                    value={bloodPressure}
                    onChange={(e) => setBloodPressure(e.target.value)}
                    placeholder="120/80"
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Telephone")}
                  </label>
                  <input
                    id="input-admit-phone"
                    type="text"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="+251..."
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Height (cm)")}
                  </label>
                  <input
                    id="input-admit-height"
                    type="number"
                    step="0.1"
                    value={height}
                    onChange={(e) =>
                      setHeight(
                        e.target.value === "" ? "" : Number(e.target.value),
                      )
                    }
                    placeholder="cm"
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Weight (kg)")}
                  </label>
                  <input
                    id="input-admit-weight"
                    type="number"
                    step="0.1"
                    value={weight}
                    onChange={(e) =>
                      setWeight(
                        e.target.value === "" ? "" : Number(e.target.value),
                      )
                    }
                    placeholder="kg"
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Payment Mode")}
                  </label>
                  <select
                    id="select-admit-payment"
                    value={paymentMode}
                    onChange={(e) => setPaymentMode(e.target.value)}
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  >
                    <option value="cash">{t("Cash")}</option>
                    <option value="card">{t("Card")}</option>
                    <option value="bank_transfer">{t("Bank Transfer")}</option>
                    <option value="other">{t("Other")}</option>
                  </select>
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground">
                    {t("Standard Charge (ETB)")}
                  </label>
                  <input
                    id="input-admit-charge"
                    type="text"
                    value={standardCharge}
                    onChange={(e) => setStandardCharge(e.target.value)}
                    className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-muted-foreground">
                  {t("Symptoms / Chief Complaint")}
                </label>
                <textarea
                  id="input-admit-symptoms"
                  rows={2}
                  value={symptoms}
                  onChange={(e) => setSymptoms(e.target.value)}
                  placeholder={t("Presenting complaints...")}
                  className="w-full mt-1 px-3 py-2 text-sm rounded-lg border border-input bg-background"
                />
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={onClose}
                  className="px-4 py-2 text-sm rounded-lg border border-border bg-background hover:bg-muted"
                >
                  {t("Cancel")}
                </button>
                <button
                  id="btn-submit-admission"
                  type="submit"
                  disabled={busy || loadingInitial}
                  className="px-5 py-2 text-sm font-semibold rounded-lg bg-primary text-primary-foreground hover:opacity-90 disabled:opacity-50 transition"
                >
                  {busy
                    ? t("Admitting…")
                    : kind === "ipd"
                      ? t("Admit Patient")
                      : t("Register Patient")}
                </button>
              </div>
            </>
          )}
        </form>
      </div>
    </div>
  );
}
