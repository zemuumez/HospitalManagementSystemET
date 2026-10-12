"use client";

import { useState, useMemo, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Search,
  Plus,
  Filter,
  Edit2,
  Trash2,
  X,
  ChevronDown,
  Download,
  AlertCircle,
  Loader2,
  RefreshCw,
  CheckCircle2,
  LogOut,
} from "lucide-react";
import { useLanguage } from "./language";
import { api, type Patient } from "@/lib/api";

export type BedTab = "bed-status" | "bed-assigns" | "beds" | "bed-types";

export interface ApiBed {
  id: string;
  name: string;
  type: string;
  typeId?: string;
  wardName?: string;
  chargeMinor: number;
  available: boolean;
  state?: string;
  description?: string;
  version: number;
  patientName?: string;
  patientPhone?: string;
  patientGender?: string;
  admissionDate?: string;
  encounterId?: string;
  occupiedBy?: {
    patientId: string;
    patientName: string;
    phone: string;
    gender: string;
    assignedFrom: string;
    encounterId: string;
  };
}

export interface ApiBedType {
  id: string;
  name: string;
  description: string;
  active: boolean;
  version: number;
  bedsCount?: number;
}

export interface ApiBedAssignment {
  id: string;
  bedId: string;
  bedName: string;
  encounterId: string;
  encounterVersion?: number;
  ipdNo: string;
  patientId: string;
  patientName: string;
  patientEmail: string;
  patientPhone: string;
  patientGender: string;
  assignedFrom: string;
  assignedTo?: string | null;
  notes: string;
  status?: boolean;
  isActive?: boolean;
}

export interface ApiBedOccupancyReport {
  totalBeds: number;
  occupiedBeds: number;
  availableBeds: number;
  occupancyRate: number;
  activeAdmissions: number;
}

interface BedStatusItem {
  id: string;
  name: string;
  isAvailable: boolean;
  state: "available" | "occupied" | "maintenance";
  patientName?: string;
  phone?: string;
  admissionDate?: string;
  gender?: string;
}

interface WardGroup {
  id: string;
  name: string;
  beds: BedStatusItem[];
}

interface BedAssignRow {
  id: string;
  ipdNo: string;
  patientId: string;
  encounterId: string;
  encounterVersion: number;
  patientName: string;
  patientEmail: string;
  initials: string;
  avatarColor: string;
  bedId: string;
  bedName: string;
  assignDate: string;
  dischargeDate: string;
  status: boolean; // true = active, false = discharged/inactive
  caseId?: string;
  description?: string;
}

interface BedRow {
  id: string;
  bedId: string;
  bedName: string;
  bedType: string;
  typeId?: string;
  wardName?: string;
  charge: number;
  available: boolean;
  state?: string;
  description?: string;
  version: number;
}

interface BedTypeRow {
  id: string;
  title: string;
  description?: string;
  version: number;
  bedsCount?: number;
}

// Initial empty states (Strict Parity: Zero hardcoded mock fixtures)
const initialWards: WardGroup[] = [];
const initialBedAssigns: BedAssignRow[] = [];
const initialBeds: BedRow[] = [];
const initialBedTypes: BedTypeRow[] = [];

export function BedManagementWorkspace({ id }: { id: string }) {
  const { t, locale } = useLanguage();
  const router = useRouter();

  // Normalize slug to tab
  const activeTab: BedTab =
    id === "bed-status"
      ? "bed-status"
      : id === "bed-assigns"
        ? "bed-assigns"
        : id === "beds"
          ? "beds"
          : "bed-types";

  // Data states
  const [wards, setWards] = useState<WardGroup[]>(initialWards);
  const [bedAssigns, setBedAssigns] = useState<BedAssignRow[]>(initialBedAssigns);
  const [beds, setBeds] = useState<BedRow[]>(initialBeds);
  const [bedTypes, setBedTypes] = useState<BedTypeRow[]>(initialBedTypes);

  // Pagination & counts per tab
  const [totalRecords, setTotalRecords] = useState(0);
  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);

  // Persistent API states
  const [apiConnected, setApiConnected] = useState(false);
  const [isLoadingApi, setIsLoadingApi] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [occupancyReport, setOccupancyReport] = useState<ApiBedOccupancyReport | null>(null);
  const [remotePatients, setRemotePatients] = useState<Patient[]>([]);
  const [remoteEncounters, setRemoteEncounters] = useState<any[]>([]);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");

  // Modals state
  const [bedTypeModalOpen, setBedTypeModalOpen] = useState(false);
  const [bedModalOpen, setBedModalOpen] = useState(false);
  const [bedAssignModalOpen, setBedAssignModalOpen] = useState(false);
  const [dischargeModalOpen, setDischargeModalOpen] = useState(false);
  const [assignmentToDischarge, setAssignmentToDischarge] = useState<BedAssignRow | null>(null);
  const [dischargeNotes, setDischargeNotes] = useState("");
  const [editingItem, setEditingItem] = useState<any>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<{
    id: string;
    type: "beds" | "bed-types" | "bed-assigns";
    name?: string;
    version?: number;
  } | null>(null);

  const [actionsDropdownOpen, setActionsDropdownOpen] = useState(false);
  const [hoveredBed, setHoveredBed] = useState<BedStatusItem | null>(null);

  // Form inputs for Bed Type Modal
  const [formBedTypeTitle, setFormBedTypeTitle] = useState("");
  const [formBedTypeDescription, setFormBedTypeDescription] = useState("");

  // Form inputs for Bed Modal
  const [formBedName, setFormBedName] = useState("");
  const [formBedTypeId, setFormBedTypeId] = useState("");
  const [formBedWardName, setFormBedWardName] = useState("");
  const [formBedCharge, setFormBedCharge] = useState("");
  const [formBedState, setFormBedState] = useState("available");
  const [formBedDescription, setFormBedDescription] = useState("");

  // Form inputs for Bed Assign Modal
  const [formAssignPatientId, setFormAssignPatientId] = useState("");
  const [formAssignEncounterId, setFormAssignEncounterId] = useState("");
  const [formAssignBedId, setFormAssignBedId] = useState("");
  const [formAssignDescription, setFormAssignDescription] = useState("");
  const [patientSearch, setPatientSearch] = useState("");
  const [encounterSearch, setEncounterSearch] = useState("");

  function formatCurrency(amount: number) {
    const formatted = amount.toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
    return locale === "am" ? `ብር ${formatted}` : `ETB ${formatted}`;
  }

  // Main loader: connects to Go API and loads persistent data for current tab
  const loadBedManagementData = useCallback(async () => {
    setIsLoadingApi(true);
    setApiErrorBanner("");

    try {
      // 1. Always load bed types for dropdown selectors
      const typesRes = await api<{ bedTypes: ApiBedType[]; total?: number }>(
        `bed-types?page=1&pageSize=100`,
      );
      if (typesRes?.bedTypes && Array.isArray(typesRes.bedTypes)) {
        const loadedTypes: BedTypeRow[] = typesRes.bedTypes.map((bt) => ({
          id: bt.id,
          title: bt.name,
          description: bt.description,
          version: bt.version || 1,
          bedsCount: bt.bedsCount || 0,
        }));
        setBedTypes(loadedTypes);
      }

      // 2. Tab-specific loading & server-backed pagination/search
      if (activeTab === "bed-types") {
        const pagedTypes = await api<{ bedTypes: ApiBedType[]; total: number }>(
          `bed-types?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`,
        );
        if (pagedTypes?.bedTypes) {
          setBedTypes(
            pagedTypes.bedTypes.map((bt) => ({
              id: bt.id,
              title: bt.name,
              description: bt.description,
              version: bt.version || 1,
              bedsCount: bt.bedsCount || 0,
            })),
          );
          setTotalRecords(pagedTypes.total || pagedTypes.bedTypes.length);
        } else {
          setBedTypes([]);
          setTotalRecords(0);
        }
      } else if (activeTab === "beds") {
        const pagedBeds = await api<{ beds: ApiBed[]; total: number }>(
          `beds?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`,
        );
        if (pagedBeds?.beds) {
          setBeds(
            pagedBeds.beds.map((b) => ({
              id: b.id,
              bedId: b.id.slice(0, 8).toUpperCase(),
              bedName: b.name,
              bedType: b.type,
              typeId: b.typeId,
              wardName: b.wardName,
              charge: (b.chargeMinor || 0) / 100,
              available: b.available,
              state: b.state || "available",
              description: b.description || "",
              version: b.version || 1,
            })),
          );
          setTotalRecords(pagedBeds.total || pagedBeds.beds.length);
        } else {
          setBeds([]);
          setTotalRecords(0);
        }
      } else if (activeTab === "bed-assigns") {
        const pagedAssigns = await api<{ assignments: ApiBedAssignment[]; total: number }>(
          `bed-assignments?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`,
        );
        if (pagedAssigns?.assignments) {
          const mappedAssigns: BedAssignRow[] = pagedAssigns.assignments.map((ba) => {
            const names = (ba.patientName || "Patient").trim().split(" ");
            const initials =
              names.length > 1
                ? `${names[0][0]}${names[1][0]}`.toUpperCase()
                : names[0].slice(0, 2).toUpperCase();

            let assignDateStr = "N/A";
            if (ba.assignedFrom) {
              const d = new Date(ba.assignedFrom);
              assignDateStr = isNaN(d.getTime())
                ? ba.assignedFrom
                : d.toLocaleDateString("en-GB", {
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  });
            }

            let dischargeDateStr = "N/A";
            if (ba.assignedTo) {
              const d = new Date(ba.assignedTo);
              dischargeDateStr = isNaN(d.getTime())
                ? ba.assignedTo
                : d.toLocaleDateString("en-GB", {
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  });
            }

            return {
              id: ba.id,
              ipdNo: ba.ipdNo || ba.encounterId.slice(0, 8).toUpperCase(),
              patientId: ba.patientId,
              encounterId: ba.encounterId,
              encounterVersion: ba.encounterVersion || 1,
              patientName: ba.patientName,
              patientEmail: ba.patientEmail || "-",
              initials,
              avatarColor: "#3b82f6",
              bedId: ba.bedId,
              bedName: ba.bedName,
              assignDate: assignDateStr,
              dischargeDate: dischargeDateStr,
              status: typeof ba.status === "boolean" ? ba.status : (ba.isActive ?? !ba.assignedTo),
              description: ba.notes || "",
            };
          });
          setBedAssigns(mappedAssigns);
          setTotalRecords(pagedAssigns.total || mappedAssigns.length);
        } else {
          setBedAssigns([]);
          setTotalRecords(0);
        }
      }

      // 3. For Bed Status tab (or dropdowns), fetch ALL live beds across pages
      let allLoadedBeds: ApiBed[] = [];
      let bedPage = 1;
      while (true) {
        const pagedBedsRes = await api<{ beds: ApiBed[]; total: number }>(
          `beds?page=${bedPage}&pageSize=200`,
        );
        if (!pagedBedsRes?.beds || pagedBedsRes.beds.length === 0) break;
        allLoadedBeds.push(...pagedBedsRes.beds);
        if (
          allLoadedBeds.length >= (pagedBedsRes.total || 0) ||
          pagedBedsRes.beds.length < 200
        ) {
          break;
        }
        bedPage++;
      }

      if (allLoadedBeds.length > 0) {
        const loadedBeds: BedRow[] = allLoadedBeds.map((b) => ({
          id: b.id,
          bedId: b.id.slice(0, 8).toUpperCase(),
          bedName: b.name,
          bedType: b.type,
          typeId: b.typeId,
          wardName: b.wardName,
          charge: (b.chargeMinor || 0) / 100,
          available: b.available,
          state: b.state || "available",
          description: b.description || "",
          version: b.version || 1,
        }));
        if (activeTab !== "beds") {
          setBeds(loadedBeds);
        }

        // Group beds by wardName or bedType dynamically
        const wardMap = new Map<string, BedStatusItem[]>();
        for (const b of allLoadedBeds) {
          const ward = (b.wardName || b.type || "General Ward").trim();
          if (!wardMap.has(ward)) wardMap.set(ward, []);

          let admissionDateStr: string | undefined;
          if (b.occupiedBy?.assignedFrom) {
            const d = new Date(b.occupiedBy.assignedFrom);
            admissionDateStr = isNaN(d.getTime())
              ? b.occupiedBy.assignedFrom
              : d.toLocaleDateString("en-GB", {
                  day: "numeric",
                  month: "short",
                  year: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                });
          }

          const rawState = (b.state || "").toLowerCase();
          const bedState: "available" | "occupied" | "maintenance" =
            rawState === "maintenance"
              ? "maintenance"
              : (!b.available || b.occupiedBy || b.patientName ? "occupied" : "available");

          wardMap.get(ward)!.push({
            id: b.id,
            name: b.name,
            isAvailable: bedState === "available",
            state: bedState,
            patientName: b.patientName || b.occupiedBy?.patientName,
            phone: b.patientPhone || b.occupiedBy?.phone || "-",
            admissionDate: b.admissionDate || admissionDateStr || "-",
            gender: b.patientGender || b.occupiedBy?.gender || "-",
          });
        }

        const groupedWards: WardGroup[] = Array.from(wardMap.entries()).map(
          ([wardName, wardBeds], idx) => ({
            id: `ward-${idx}-${wardName.toLowerCase().replace(/\s+/g, "-")}`,
            name: wardName,
            beds: wardBeds,
          }),
        );
        setWards(groupedWards);
      } else {
        setWards([]);
      }

      // 4. Fetch live occupancy report, patients, and encounters without swallowing errors
      const auxErrors: string[] = [];
      try {
        const rep = await api<ApiBedOccupancyReport>("bed-occupancy/report");
        if (rep) setOccupancyReport(rep);
      } catch (err: any) {
        auxErrors.push(`Occupancy report: ${err?.message || "fetch failed"}`);
      }

      try {
        const patRes = await api<{ patients: Patient[] }>("patients?page=1&pageSize=200");
        if (patRes?.patients) setRemotePatients(patRes.patients);
      } catch (err: any) {
        auxErrors.push(`Patients: ${err?.message || "fetch failed"}`);
      }

      try {
        const encRes = await api<{ encounters: any[] }>("encounters?kind=ipd&page=1&pageSize=200");
        if (encRes?.encounters) setRemoteEncounters(encRes.encounters);
      } catch (err: any) {
        auxErrors.push(`Encounters: ${err?.message || "fetch failed"}`);
      }

      if (auxErrors.length > 0) {
        setApiErrorBanner(auxErrors.join("; "));
      }

      setApiConnected(true);
    } catch (err: any) {
      setApiConnected(false);
      setApiErrorBanner(err?.message || t("Could not connect to Go bed management service."));
    } finally {
      setIsLoadingApi(false);
    }
  }, [activeTab, page, pageSize, search, t]);

  useEffect(() => {
    loadBedManagementData();
  }, [loadBedManagementData]);

  // Tab switching
  function navigateToTab(tab: BedTab) {
    setSearch("");
    setPage(1);
    router.push(`/modules/${tab}`);
  }

  // Open Bed Type Modal
  function handleOpenNewBedType(existing?: BedTypeRow) {
    setApiErrorBanner("");
    setApiSuccessBanner("");
    if (existing) {
      setEditingItem(existing);
      setFormBedTypeTitle(existing.title);
      setFormBedTypeDescription(existing.description || "");
    } else {
      setEditingItem(null);
      setFormBedTypeTitle("");
      setFormBedTypeDescription("");
    }
    setBedTypeModalOpen(true);
  }

  // Save Bed Type (Create or Update with Concurrency Protection)
  async function handleSaveBedType(e: React.FormEvent) {
    e.preventDefault();
    if (!formBedTypeTitle.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");

    try {
      if (editingItem) {
        await api<ApiBedType>(`bed-types/${editingItem.id}`, {
          method: "PATCH",
          body: JSON.stringify({
            name: formBedTypeTitle.trim(),
            description: formBedTypeDescription.trim(),
            active: true,
            version: editingItem.version || 1,
          }),
        });
        setApiSuccessBanner(t("Bed Type updated and persisted successfully."));
      } else {
        await api<ApiBedType>("bed-types", {
          method: "POST",
          body: JSON.stringify({
            name: formBedTypeTitle.trim(),
            description: formBedTypeDescription.trim(),
            active: true,
          }),
        });
        setApiSuccessBanner(t("Bed Type created and registered in Go database successfully."));
      }
      setBedTypeModalOpen(false);
      await loadBedManagementData();
    } catch (err: any) {
      // PRESERVE FORM: do not close modal, do not fake local mutations
      setApiErrorBanner(err?.message || t("Failed to save Bed Type."));
    } finally {
      setIsSubmitting(false);
    }
  }

  // Open Bed Modal
  function handleOpenNewBed(existing?: BedRow) {
    setApiErrorBanner("");
    setApiSuccessBanner("");
    if (existing) {
      setEditingItem(existing);
      setFormBedName(existing.bedName);
      setFormBedTypeId(existing.typeId || existing.bedType);
      setFormBedWardName(existing.wardName || "");
      setFormBedCharge(String(existing.charge));
      setFormBedState(existing.state || "available");
      setFormBedDescription(existing.description || "");
    } else {
      setEditingItem(null);
      setFormBedName("");
      setFormBedTypeId(bedTypes[0]?.id || bedTypes[0]?.title || "");
      setFormBedWardName("");
      setFormBedCharge("500");
      setFormBedState("available");
      setFormBedDescription("");
    }
    setBedModalOpen(true);
    setActionsDropdownOpen(false);
  }

  // Save Bed (Create or Update with Maintenance/State & Optimistic Locking)
  async function handleSaveBed(e: React.FormEvent) {
    e.preventDefault();
    if (!formBedName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");

    const chargeNum = parseFloat(formBedCharge) || 0;
    const chargeMinor = Math.round(chargeNum * 100);

    // Resolve bed type name from ID or text
    const matchedType = bedTypes.find(
      (bt) => bt.id === formBedTypeId || bt.title.toLowerCase() === formBedTypeId.toLowerCase(),
    );
    const resolvedTypeName = matchedType?.title || formBedTypeId || "General Ward";
    const resolvedTypeId = matchedType?.id;

    try {
      if (editingItem) {
        await api<ApiBed>(`beds/${editingItem.id}`, {
          method: "PATCH",
          body: JSON.stringify({
            name: formBedName.trim(),
            type: resolvedTypeName,
            typeId: resolvedTypeId,
            wardName: formBedWardName.trim(),
            chargeMinor,
            state: formBedState,
            description: formBedDescription.trim(),
            version: editingItem.version || 1,
          }),
        });
        setApiSuccessBanner(t("Bed updated and state persisted successfully."));
      } else {
        await api<ApiBed>("beds", {
          method: "POST",
          body: JSON.stringify({
            name: formBedName.trim(),
            type: resolvedTypeName,
            typeId: resolvedTypeId,
            wardName: formBedWardName.trim(),
            chargeMinor,
            description: formBedDescription.trim(),
          }),
        });
        setApiSuccessBanner(t("Bed created and registered in Go database successfully."));
      }
      setBedModalOpen(false);
      await loadBedManagementData();
    } catch (err: any) {
      // PRESERVE FORM: do not close modal, do not fake local mutations
      setApiErrorBanner(err?.message || t("Failed to save Bed."));
    } finally {
      setIsSubmitting(false);
    }
  }

  // Open Bed Assign Modal
  function handleOpenNewBedAssign(preselectedBedId?: string) {
    setApiErrorBanner("");
    setApiSuccessBanner("");
    setEditingItem(null);
    setFormAssignPatientId(remotePatients[0]?.id || "");
    setFormAssignEncounterId("");
    setFormAssignBedId(preselectedBedId || beds.find((b) => b.available)?.id || beds[0]?.id || "");
    setFormAssignDescription("");
    setBedAssignModalOpen(true);
  }

  // Auto-select encounter matching selected patient
  useEffect(() => {
    if (formAssignPatientId && remoteEncounters.length > 0) {
      const patientEncounters = remoteEncounters.filter(
        (e) => e.patientId === formAssignPatientId,
      );
      if (patientEncounters.length > 0) {
        setFormAssignEncounterId(patientEncounters[0].id);
      } else {
        setFormAssignEncounterId("");
      }
    }
  }, [formAssignPatientId, remoteEncounters]);

  // Save Bed Assign (Persistent UUID validation & positive version contract)
  async function handleSaveBedAssign(e: React.FormEvent) {
    e.preventDefault();
    if (!formAssignPatientId || !formAssignEncounterId || !formAssignBedId) {
      setApiErrorBanner(
        t("Please select a Patient, an active IPD Encounter, and an available Bed."),
      );
      return;
    }
    const selectedEnc = remoteEncounters.find((enc) => enc.id === formAssignEncounterId);
    const encVersion = selectedEnc?.version || 1;
    if (encVersion <= 0) {
      setApiErrorBanner(t("A positive encounter version is required"));
      return;
    }

    setIsSubmitting(true);
    setApiErrorBanner("");

    try {
      await api("bed-assignments", {
        method: "POST",
        body: JSON.stringify({
          patientId: formAssignPatientId,
          encounterId: formAssignEncounterId,
          bedId: formAssignBedId,
          notes: formAssignDescription.trim(),
          version: encVersion,
        }),
      });

      setApiSuccessBanner(t("Bed assignment stored in persistent clinical records."));
      setBedAssignModalOpen(false);
      await loadBedManagementData();
    } catch (err: any) {
      // PRESERVE FORM: do not fake local state, display exact error
      setApiErrorBanner(err?.message || t("Could not assign bed."));
    } finally {
      setIsSubmitting(false);
    }
  }

  // Open Discharge Modal
  function handleOpenDischarge(row: BedAssignRow) {
    setApiErrorBanner("");
    setApiSuccessBanner("");
    setAssignmentToDischarge(row);
    setDischargeNotes("");
    setDischargeModalOpen(true);
  }

  // Submit Discharge (Atomic status update, occupancy release & positive version contract)
  async function handleConfirmDischarge(e: React.FormEvent) {
    e.preventDefault();
    if (!assignmentToDischarge) return;
    const encVersion = assignmentToDischarge.encounterVersion || 1;
    if (encVersion <= 0) {
      setApiErrorBanner(t("A positive encounter version is required"));
      return;
    }

    setIsSubmitting(true);
    setApiErrorBanner("");

    try {
      await api(`bed-assignments/${assignmentToDischarge.id}/discharge`, {
        method: "POST",
        body: JSON.stringify({
          version: encVersion,
          dischargeNotes: dischargeNotes.trim() || t("Discharged from Bed Management Workspace"),
        }),
      });
      setApiSuccessBanner(
        t("Patient discharged and bed occupancy released successfully."),
      );
      setDischargeModalOpen(false);
      setAssignmentToDischarge(null);
      await loadBedManagementData();
    } catch (err: any) {
      setApiErrorBanner(err?.message || t("Failed to discharge patient from bed."));
    } finally {
      setIsSubmitting(false);
    }
  }

  // Confirm Delete with In-Use & Occupancy Protection
  async function handleDeleteConfirm() {
    if (!itemToDelete) return;
    setIsSubmitting(true);
    setApiErrorBanner("");

    try {
      if (itemToDelete.type === "bed-types") {
        const btVersion = itemToDelete.version || 1;
        if (btVersion <= 0) {
          throw new Error(t("A positive bed type version is required"));
        }
        await api(`bed-types/${itemToDelete.id}?version=${btVersion}`, {
          method: "DELETE",
        });
        setApiSuccessBanner(t("Bed type deleted successfully."));
      } else if (itemToDelete.type === "beds") {
        await api(`beds/${itemToDelete.id}`, {
          method: "DELETE",
        });
        setApiSuccessBanner(t("Bed deleted successfully."));
      }
      setDeleteConfirmOpen(false);
      setItemToDelete(null);
      await loadBedManagementData();
    } catch (err: any) {
      // In-use or conflict error
      setApiErrorBanner(err?.message || t("Could not delete record. Record is protected or in use."));
      setDeleteConfirmOpen(false);
    } finally {
      setIsSubmitting(false);
    }
  }

  // Export to Excel / CSV
  function handleExportCsv(type: "beds" | "bed-assigns" | "bed-types") {
    let headers: string[] = [];
    let rows: string[][] = [];

    if (type === "beds") {
      headers = ["BED ID", "BED", "BED TYPE", "CHARGE", "AVAILABLE", "STATE"];
      rows = beds.map((b) => [
        b.bedId,
        b.bedName,
        b.bedType,
        formatCurrency(b.charge),
        b.available ? "Yes" : "No",
        b.state || "available",
      ]);
    } else if (type === "bed-assigns") {
      headers = ["IPD NO", "PATIENT", "BED", "ASSIGN DATE", "DISCHARGE DATE", "STATUS"];
      rows = bedAssigns.map((ba) => [
        ba.ipdNo,
        ba.patientName,
        ba.bedName,
        ba.assignDate,
        ba.dischargeDate,
        ba.status ? "Active" : "Discharged",
      ]);
    } else if (type === "bed-types") {
      headers = ["BED TYPE", "DESCRIPTION", "BEDS COUNT"];
      rows = bedTypes.map((bt) => [bt.title, bt.description || "", String(bt.bedsCount || 0)]);
    }

    const csvContent =
      "data:text/csv;charset=utf-8," +
      [
        headers.join(","),
        ...rows.map((r) => r.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(",")),
      ].join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", `${type}-export.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setActionsDropdownOpen(false);
  }

  // Active encounters for selected patient in modal
  const eligibleEncounters = useMemo(() => {
    if (!formAssignPatientId) return remoteEncounters;
    return remoteEncounters.filter((e) => e.patientId === formAssignPatientId);
  }, [formAssignPatientId, remoteEncounters]);

  const filteredPatients = useMemo(() => {
    if (!patientSearch.trim()) return remotePatients;
    const q = patientSearch.toLowerCase().trim();
    return remotePatients.filter(
      (p) =>
        p.givenName?.toLowerCase().includes(q) ||
        p.familyName?.toLowerCase().includes(q) ||
        p.mrn?.toLowerCase().includes(q) ||
        p.id.toLowerCase().includes(q),
    );
  }, [remotePatients, patientSearch]);

  const filteredEncounters = useMemo(() => {
    const list = eligibleEncounters;
    if (!encounterSearch.trim()) return list;
    const q = encounterSearch.toLowerCase().trim();
    return list.filter(
      (e) =>
        e.id.toLowerCase().includes(q) ||
        (e.patientName && e.patientName.toLowerCase().includes(q)) ||
        (e.status && e.status.toLowerCase().includes(q)),
    );
  }, [eligibleEncounters, encounterSearch]);

  const totalPages = Math.max(1, Math.ceil(totalRecords / pageSize));

  return (
    <div className="bed-workspace-container" data-ready="true">
      {/* Persistent Connection Status & Live Occupancy Report */}
      <div className="mb-3 d-flex flex-column gap-2" style={{ padding: "0 4px" }}>
        <div
          className="d-flex justify-content-between align-items-center flex-wrap gap-2 px-3 py-2 rounded border"
          style={{
            backgroundColor: apiConnected ? "rgba(16, 185, 129, 0.08)" : "rgba(245, 158, 11, 0.08)",
            borderColor: apiConnected ? "#10b981" : "#f59e0b",
          }}
        >
          <div className="d-flex align-items-center flex-wrap gap-2">
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                backgroundColor: apiConnected ? "#10b981" : "#f59e0b",
                display: "inline-block",
              }}
            />
            <span className="fs-7 fw-semibold">
              {apiConnected
                ? t("Connected to Go/PostgreSQL Bed Service (/v1/beds, /v1/bed-types, /v1/bed-occupancy)")
                : t("Connecting to backend bed service...")}
            </span>
            {occupancyReport && (
              <span
                className="badge-available-stock fs-8 py-0 px-2"
                style={{
                  backgroundColor: "#3b82f622",
                  color: "#3b82f6",
                  borderColor: "#3b82f6",
                }}
              >
                {t("Total")}: {occupancyReport.totalBeds} | {t("Occupied")}:{" "}
                {occupancyReport.occupiedBeds} | {t("Available")}:{" "}
                {occupancyReport.availableBeds} ({occupancyReport.occupancyRate.toFixed(1)}%)
              </span>
            )}
          </div>
          <button
            type="button"
            className="btn-icon-link fs-7 d-flex align-items-center gap-1"
            onClick={loadBedManagementData}
            disabled={isLoadingApi}
            title={t("Refresh bed records from Go API")}
          >
            <RefreshCw size={13} className={isLoadingApi ? "animate-spin" : ""} />
            <span>{isLoadingApi ? t("Syncing...") : t("Sync Backend")}</span>
          </button>
        </div>

        {apiSuccessBanner && (
          <div
            className="alert-notice alert-notice-success d-flex align-items-center justify-content-between py-2 px-3 border rounded"
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
              aria-label={t("Close")}
            >
              <X size={14} />
            </button>
          </div>
        )}

        {apiErrorBanner && (
          <div
            className="alert-notice alert-notice-error d-flex align-items-center justify-content-between py-2 px-3 border rounded"
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
              aria-label={t("Close")}
            >
              <X size={14} />
            </button>
          </div>
        )}
      </div>

      {/* Tab: Bed Status */}
      {activeTab === "bed-status" && (
        <div className="bed-status-view">
          <div className="bed-status-top-bar">
            <h1 className="bed-status-title">{t("Bed Status")}</h1>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => navigateToTab("bed-assigns")}
            >
              {t("Back")}
            </button>
          </div>

          <div className="bed-status-grid">
            {wards.length === 0 ? (
              <div className="ward-section-card">
                <div className="ward-content-box justify-content-center text-center py-5">
                  <span className="no-bed-message">{t("No bed records available.")}</span>
                </div>
              </div>
            ) : (
              wards.map((ward) => (
                <div key={ward.id} className="ward-section-card">
                  <h2 className="ward-title">{ward.name}</h2>
                  <div className="ward-content-box">
                    {ward.beds.length === 0 ? (
                      <div className="no-bed-message">
                        <span>{t("No Bed Available")}</span>
                      </div>
                    ) : (
                      <div className="ward-beds-container">
                        {ward.beds.map((bed) => {
                          const isOccupied = bed.state === "occupied";
                          const isMaintenance = bed.state === "maintenance";
                          const isAvailable = bed.state === "available";

                          const stateClass = isMaintenance
                            ? "maintenance"
                            : isOccupied
                              ? "occupied"
                              : "available";

                          const labelColorClass = isMaintenance
                            ? "text-warning"
                            : isOccupied
                              ? "text-danger"
                              : "text-success";

                          return (
                            <div
                              key={bed.id}
                              className={`bed-status-item ${stateClass}`}
                              onMouseEnter={() => isOccupied && setHoveredBed(bed)}
                              onMouseLeave={() => setHoveredBed(null)}
                              onClick={() => {
                                if (isAvailable) {
                                  handleOpenNewBedAssign(bed.id);
                                }
                              }}
                            >
                              <div className="bed-icon-wrapper">
                                {isMaintenance ? (
                                  <svg
                                    className="bed-maintenance-svg text-warning"
                                    viewBox="0 0 64 64"
                                    width="48"
                                    height="48"
                                    fill="none"
                                    stroke="#fbbf24"
                                    strokeWidth="2.5"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                  >
                                    <path d="M6 48V28" />
                                    <path d="M58 48V36" />
                                    <path d="M6 42h52" />
                                    <circle cx="18" cy="26" r="6" fill="#fbbf24" opacity="0.3" stroke="#fbbf24" />
                                    <path
                                      d="M26 32h26a6 6 0 0 1 6 6v4H16v-4a6 6 0 0 1 10-6z"
                                      fill="#fbbf24"
                                      opacity="0.3"
                                    />
                                    <path d="M36 14l6 6-4 4-6-6 4-4z" stroke="#fbbf24" strokeWidth="2.5" />
                                    <path d="M38 16l-10 10" stroke="#f59e0b" strokeWidth="2.5" />
                                  </svg>
                                ) : isOccupied ? (
                                  <svg
                                    className="bed-pulse-svg text-danger"
                                    viewBox="0 0 64 64"
                                    width="48"
                                    height="48"
                                    fill="none"
                                    stroke="#f87171"
                                    strokeWidth="2.5"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                  >
                                    <path d="M4 48V26" />
                                    <path d="M60 48V36" />
                                    <path d="M4 42h56" />
                                    <circle cx="16" cy="24" r="5" fill="#f87171" />
                                    <path
                                      d="M22 28h16a8 8 0 0 1 8 8v6H14v-6a8 8 0 0 1 8-8z"
                                      fill="#f87171"
                                      opacity="0.3"
                                    />
                                    <path
                                      d="M26 18h6l3-6 4 12 3-8 2 4h6"
                                      stroke="#ef4444"
                                      strokeWidth="3"
                                    />
                                  </svg>
                                ) : (
                                  <svg
                                    className="bed-available-svg text-success"
                                    viewBox="0 0 64 64"
                                    width="48"
                                    height="48"
                                    fill="none"
                                    stroke="#34d399"
                                    strokeWidth="2.5"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                  >
                                    <path d="M6 48V28" />
                                    <path d="M58 48V36" />
                                    <path d="M6 42h52" />
                                    <circle cx="18" cy="26" r="6" fill="#10b981" />
                                    <path
                                      d="M26 32h26a6 6 0 0 1 6 6v4H16v-4a6 6 0 0 1 10-6z"
                                      fill="#10b981"
                                      opacity="0.3"
                                    />
                                  </svg>
                                )}
                              </div>

                              <span className={`bed-status-label ${labelColorClass}`}>
                                {isMaintenance
                                  ? `${bed.name} (${t("Maintenance")})`
                                  : isOccupied
                                    ? (bed.patientName || bed.name)
                                    : bed.name}
                              </span>

                              {/* Hover tooltip for occupied bed */}
                              {hoveredBed?.id === bed.id && isOccupied && (
                                <div className="bed-hover-popover">
                                  <div className="popover-row">
                                    <span className="popover-label">{t("Bed Name")} :</span>
                                    <span className="popover-val">{bed.name}</span>
                                  </div>
                                  <div className="popover-row">
                                    <span className="popover-label">{t("Patient")} :</span>
                                    <span className="popover-val">{bed.patientName || "-"}</span>
                                  </div>
                                  <div className="popover-row">
                                    <span className="popover-label">{t("Phone")} :</span>
                                    <span className="popover-val">{bed.phone || "-"}</span>
                                  </div>
                                  <div className="popover-row">
                                    <span className="popover-label">{t("Admission Date")} :</span>
                                    <span className="popover-val">{bed.admissionDate || "-"}</span>
                                  </div>
                                  <div className="popover-row">
                                    <span className="popover-label">{t("Gender")} :</span>
                                    <span className="popover-val">{bed.gender || "-"}</span>
                                  </div>
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {/* Tab: Bed Assigns */}
      {activeTab === "bed-assigns" && (
        <div className="bed-assigns-view">
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
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setActionsDropdownOpen(!actionsDropdownOpen)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => handleOpenNewBedAssign()}
              >
                {t("New Bed Assign")}
              </button>
            </div>
          </div>

          <div className="billing-card">
            <div className="table-responsive">
              <table className="billing-table">
                <thead>
                  <tr>
                    <th>
                      <div className="th-sort">
                        <span>{t("IPD NO")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("PATIENT")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("ASSIGN DATE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("DISCHARGE DATE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>{t("STATUS")}</th>
                    <th style={{ textAlign: "right" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {bedAssigns.map((row) => (
                    <tr key={row.id}>
                      {/* IPD NO pill */}
                      <td>
                        <Link
                          href={`/modules/ipd-patient-departments?id=${row.ipdNo}`}
                          className="badge-blue-link"
                        >
                          {row.ipdNo}
                        </Link>
                      </td>
                      {/* Patient info */}
                      <td>
                        <div className="patient-cell">
                          <div
                            className="patient-avatar"
                            style={{ backgroundColor: row.avatarColor }}
                          >
                            {row.initials}
                          </div>
                          <div className="patient-info">
                            <span className="patient-name">{row.patientName}</span>
                            <span className="patient-email">{row.patientEmail}</span>
                          </div>
                        </div>
                      </td>
                      {/* Bed name */}
                      <td>
                        <span className="bed-name-link">{row.bedName}</span>
                      </td>
                      {/* Assign Date */}
                      <td>
                        <span className="date-link">{row.assignDate}</span>
                      </td>
                      {/* Discharge Date */}
                      <td>
                        {row.dischargeDate === "N/A" ? (
                          <span className="text-muted">N/A</span>
                        ) : (
                          <span className="date-link">{row.dischargeDate}</span>
                        )}
                      </td>
                      {/* Status Toggle Switch */}
                      <td>
                        <label
                          className="switch-toggle"
                          aria-label={`Toggle status for ${row.patientName}`}
                        >
                          <input
                            type="checkbox"
                            checked={row.status}
                            onChange={() => {
                              if (row.status) {
                                handleOpenDischarge(row);
                              }
                            }}
                            disabled={!row.status}
                          />
                          <span className="slider round"></span>
                        </label>
                      </td>
                      {/* Actions */}
                      <td style={{ textAlign: "right" }}>
                        <div className="action-buttons-group">
                          {row.status ? (
                            <button
                              type="button"
                              className="btn-edit-blue"
                              title={t("Discharge Patient")}
                              aria-label={`Discharge ${row.patientName}`}
                              onClick={() => handleOpenDischarge(row)}
                            >
                              <LogOut size={16} />
                            </button>
                          ) : (
                            <span className="badge-available-no fs-8 py-1 px-2">
                              {t("Discharged")}
                            </span>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                  {bedAssigns.length === 0 && (
                    <tr>
                      <td colSpan={7} className="no-data-cell">
                        {t("No records found")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination footer */}
            <div className="billing-footer">
              <div className="footer-left">
                <span className="footer-text">{t("Show")}</span>
                <select
                  className="page-size-select"
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  aria-label={t("Page size")}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="footer-text">
                  {t("Showing")} {totalRecords ? (page - 1) * pageSize + 1 : 0}{" "}
                  {t("to")} {Math.min(page * pageSize, totalRecords)} {t("of")}{" "}
                  {totalRecords} {t("Results")}
                </span>
              </div>

              <div className="footer-pagination">
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                  aria-label={t("Previous")}
                >
                  ‹
                </button>
                <button
                  type="button"
                  className={`pagination-btn ${page === 1 ? "active" : ""}`}
                  onClick={() => setPage(1)}
                >
                  1
                </button>
                {totalPages > 1 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 2 ? "active" : ""}`}
                    onClick={() => setPage(2)}
                  >
                    2
                  </button>
                )}
                {totalPages > 2 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 3 ? "active" : ""}`}
                    onClick={() => setPage(3)}
                  >
                    3
                  </button>
                )}
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page >= totalPages}
                  onClick={() => setPage(page + 1)}
                  aria-label={t("Next")}
                >
                  ›
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab: Beds */}
      {activeTab === "beds" && (
        <div className="beds-view">
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
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
              >
                <Filter size={18} />
              </button>

              <button
                type="button"
                className="btn-action-blue"
                onClick={() => handleOpenNewBed()}
              >
                <Plus size={16} className="me-1 inline" />
                {t("New Bed")}
              </button>

              <div className="relative-actions-dropdown">
                <button
                  type="button"
                  className="btn-action-blue dropdown-trigger"
                  onClick={() => setActionsDropdownOpen(!actionsDropdownOpen)}
                >
                  {t("Actions")}
                  <ChevronDown size={16} className="ml-1" />
                </button>

                {actionsDropdownOpen && (
                  <div className="dropdown-action-menu">
                    <button
                      type="button"
                      className="dropdown-menu-item"
                      onClick={() => handleOpenNewBed()}
                    >
                      <Plus size={16} className="me-2" />
                      {t("New Bed")}
                    </button>
                    <button
                      type="button"
                      className="dropdown-menu-item"
                      onClick={() => handleExportCsv("beds")}
                    >
                      <Download size={16} className="me-2" />
                      {t("Export to Excel")}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="billing-card">
            <div className="table-responsive">
              <table className="billing-table">
                <thead>
                  <tr>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED ID")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED TYPE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("CHARGE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>{t("AVAILABLE")}</th>
                    <th style={{ textAlign: "right" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {beds.map((row) => (
                    <tr key={row.id}>
                      {/* Bed ID pill */}
                      <td>
                        <span className="badge-blue-link">{row.bedId}</span>
                      </td>
                      {/* Bed Name */}
                      <td>
                        <span className="bed-name-text">{row.bedName}</span>
                      </td>
                      {/* Bed Type */}
                      <td>
                        <Link
                          href={`/modules/bed-types?title=${encodeURIComponent(row.bedType)}`}
                          className="bed-type-link"
                        >
                          {row.bedType}
                        </Link>
                      </td>
                      {/* Charge */}
                      <td>
                        <span className="currency-amount">
                          {formatCurrency(row.charge)}
                        </span>
                      </td>
                      {/* Available status badge */}
                      <td>
                        {row.available ? (
                          <span className="badge-available-yes">{t("Available")}</span>
                        ) : row.state === "maintenance" ? (
                          <span className="badge-available-maintenance">{t("Maintenance")}</span>
                        ) : (
                          <span className="badge-available-no">{t("Occupied")}</span>
                        )}
                      </td>
                      {/* Actions */}
                      <td style={{ textAlign: "right" }}>
                        <div className="action-buttons-group">
                          <button
                            type="button"
                            className="btn-edit-blue"
                            title={t("Edit")}
                            aria-label={`Edit ${row.bedName}`}
                            onClick={() => handleOpenNewBed(row)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-delete-red"
                            title={t("Delete")}
                            aria-label={`Delete ${row.bedName}`}
                            onClick={() => {
                              setItemToDelete({
                                id: row.id,
                                type: "beds",
                                name: row.bedName,
                              });
                              setDeleteConfirmOpen(true);
                            }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {beds.length === 0 && (
                    <tr>
                      <td colSpan={6} className="no-data-cell">
                        {t("No records found")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination footer */}
            <div className="billing-footer">
              <div className="footer-left">
                <span className="footer-text">{t("Show")}</span>
                <select
                  className="page-size-select"
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  aria-label={t("Page size")}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="footer-text">
                  {t("Showing")} {totalRecords ? (page - 1) * pageSize + 1 : 0}{" "}
                  {t("to")} {Math.min(page * pageSize, totalRecords)} {t("of")}{" "}
                  {totalRecords} {t("Results")}
                </span>
              </div>

              <div className="footer-pagination">
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                  aria-label={t("Previous")}
                >
                  ‹
                </button>
                <button
                  type="button"
                  className={`pagination-btn ${page === 1 ? "active" : ""}`}
                  onClick={() => setPage(1)}
                >
                  1
                </button>
                {totalPages > 1 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 2 ? "active" : ""}`}
                    onClick={() => setPage(2)}
                  >
                    2
                  </button>
                )}
                {totalPages > 2 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 3 ? "active" : ""}`}
                    onClick={() => setPage(3)}
                  >
                    3
                  </button>
                )}
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page >= totalPages}
                  onClick={() => setPage(page + 1)}
                  aria-label={t("Next")}
                >
                  ›
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab: Bed Types */}
      {activeTab === "bed-types" && (
        <div className="bed-types-view">
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
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => handleOpenNewBedType()}
              >
                {t("New Bed Type")}
              </button>
            </div>
          </div>

          <div className="billing-card">
            <div className="table-responsive">
              <table className="billing-table">
                <thead>
                  <tr>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED TYPE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>{t("DESCRIPTION")}</th>
                    <th style={{ textAlign: "right" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {bedTypes.map((row) => (
                    <tr key={row.id}>
                      {/* Bed Type Name in blue link style */}
                      <td>
                        <Link
                          href={`/modules/beds?type=${encodeURIComponent(row.title)}`}
                          className="bed-type-title-link"
                        >
                          {row.title}
                        </Link>
                      </td>
                      {/* Description */}
                      <td>
                        <span className="text-muted">{row.description || "-"}</span>
                      </td>
                      {/* Actions */}
                      <td style={{ textAlign: "right" }}>
                        <div className="action-buttons-group">
                          <button
                            type="button"
                            className="btn-edit-blue"
                            title={t("Edit")}
                            aria-label={`Edit ${row.title}`}
                            onClick={() => handleOpenNewBedType(row)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-delete-red"
                            title={t("Delete")}
                            aria-label={`Delete ${row.title}`}
                            onClick={() => {
                              setItemToDelete({
                                id: row.id,
                                type: "bed-types",
                                name: row.title,
                                version: row.version,
                              });
                              setDeleteConfirmOpen(true);
                            }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {bedTypes.length === 0 && (
                    <tr>
                      <td colSpan={3} className="no-data-cell">
                        {t("No records found")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination footer */}
            <div className="billing-footer">
              <div className="footer-left">
                <span className="footer-text">{t("Show")}</span>
                <select
                  className="page-size-select"
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  aria-label={t("Page size")}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="footer-text">
                  {t("Showing")} {totalRecords ? (page - 1) * pageSize + 1 : 0}{" "}
                  {t("to")} {Math.min(page * pageSize, totalRecords)} {t("of")}{" "}
                  {totalRecords} {t("Results")}
                </span>
              </div>

              <div className="footer-pagination">
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                  aria-label={t("Previous")}
                >
                  ‹
                </button>
                <button
                  type="button"
                  className={`pagination-btn ${page === 1 ? "active" : ""}`}
                  onClick={() => setPage(1)}
                >
                  1
                </button>
                {totalPages > 1 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 2 ? "active" : ""}`}
                    onClick={() => setPage(2)}
                  >
                    2
                  </button>
                )}
                {totalPages > 2 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 3 ? "active" : ""}`}
                    onClick={() => setPage(3)}
                  >
                    3
                  </button>
                )}
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page >= totalPages}
                  onClick={() => setPage(page + 1)}
                  aria-label={t("Next")}
                >
                  ›
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Modal: New / Edit Bed Type */}
      {bedTypeModalOpen && (
        <div className="modal-backdrop-custom" onClick={() => setBedTypeModalOpen(false)}>
          <div className="modal-card-custom" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header-custom">
              <h3 className="modal-title-custom" id="bed-type-modal-title">
                {editingItem ? t("Edit Bed Type") : t("New Bed Type")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setBedTypeModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveBedType} className="modal-form-custom">
              <div className="form-group-custom">
                <label className="form-label-custom">
                  {t("Bed Type")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Bed Type")}
                  className="form-input-custom"
                  value={formBedTypeTitle}
                  onChange={(e) => setFormBedTypeTitle(e.target.value)}
                  autoFocus
                />
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Description")}
                  className="form-textarea-custom"
                  value={formBedTypeDescription}
                  onChange={(e) => setFormBedTypeDescription(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button type="submit" className="btn-modal-save" disabled={isSubmitting}>
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setBedTypeModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: New / Edit Bed */}
      {bedModalOpen && (
        <div
          className="modal-backdrop-custom"
          role="dialog"
          aria-modal="true"
          aria-labelledby="bed-modal-title"
          onClick={() => setBedModalOpen(false)}
        >
          <div className="modal-card-custom" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header-custom">
              <h3 className="modal-title-custom" id="bed-modal-title">
                {editingItem ? t("Edit Bed") : t("New Bed")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setBedModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveBed} className="modal-form-custom">
              <div className="form-group-custom">
                <label className="form-label-custom" htmlFor="bed-name-input">
                  {t("Bed")}: <span className="text-danger">*</span>
                </label>
                <input
                  id="bed-name-input"
                  aria-label="Name"
                  type="text"
                  required
                  placeholder={t("Bed")}
                  className="form-input-custom"
                  value={formBedName}
                  onChange={(e) => setFormBedName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom" htmlFor="bed-type-input">
                  {t("Bed Type")}: <span className="text-danger">*</span>
                </label>
                <select
                  id="bed-type-input"
                  aria-label="Bed Type"
                  required
                  className="form-select-custom"
                  value={formBedTypeId}
                  onChange={(e) => setFormBedTypeId(e.target.value)}
                >
                  <option value="">{t("Select Bed Type")}</option>
                  {bedTypes.map((bt) => (
                    <option key={bt.id} value={bt.id}>
                      {bt.title}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom" htmlFor="bed-ward-input">
                  {t("Ward")}:
                </label>
                <input
                  id="bed-ward-input"
                  type="text"
                  placeholder="e.g. Ward A, ICU1, General Ward"
                  className="form-input-custom"
                  value={formBedWardName}
                  onChange={(e) => setFormBedWardName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom" htmlFor="bed-charge-input">
                  {t("Charge")}: <span className="text-danger">*</span>
                </label>
                <input
                  id="bed-charge-input"
                  aria-label="Charge (ETB)"
                  type="text"
                  inputMode="decimal"
                  required
                  placeholder={t("Charge")}
                  className="form-input-custom"
                  value={formBedCharge}
                  onChange={(e) => setFormBedCharge(e.target.value)}
                />
              </div>

              {editingItem && (
                <div className="form-group-custom">
                  <label className="form-label-custom" htmlFor="bed-state-input">
                    {t("State / Status")}:
                  </label>
                  <select
                    id="bed-state-input"
                    className="form-select-custom"
                    value={formBedState}
                    onChange={(e) => setFormBedState(e.target.value)}
                  >
                    <option value="available">{t("Available")}</option>
                    <option value="maintenance">{t("Maintenance")}</option>
                    <option value="cleaning">{t("Cleaning")}</option>
                    <option value="reserved">{t("Reserved")}</option>
                  </select>
                </div>
              )}

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Description")}
                  className="form-textarea-custom"
                  value={formBedDescription}
                  onChange={(e) => setFormBedDescription(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button type="submit" className="btn-modal-save" disabled={isSubmitting}>
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setBedModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: New Bed Assign (Persistent UUID-backed) */}
      {bedAssignModalOpen && (
        <div className="modal-backdrop-custom" onClick={() => setBedAssignModalOpen(false)}>
          <div
            className="modal-card-custom modal-wide"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom" id="bed-assign-modal-title">
                {t("New Bed Assign")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setBedAssignModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveBedAssign} className="modal-form-custom">
              <div className="form-grid-2">
                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("Patient")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    placeholder={t("Search patients")}
                    className="form-input-custom mb-1"
                    value={patientSearch}
                    onChange={(e) => setPatientSearch(e.target.value)}
                    aria-label={t("Search patients")}
                  />
                  <select
                    required
                    className="form-select-custom"
                    value={formAssignPatientId}
                    onChange={(e) => setFormAssignPatientId(e.target.value)}
                  >
                    <option value="">{t("Select Patient")}</option>
                    {filteredPatients.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.givenName} {p.familyName} ({p.mrn})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("IPD Encounter")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    placeholder={t("Search")}
                    className="form-input-custom mb-1"
                    value={encounterSearch}
                    onChange={(e) => setEncounterSearch(e.target.value)}
                    aria-label={t("Search")}
                  />
                  <select
                    required
                    className="form-select-custom"
                    value={formAssignEncounterId}
                    onChange={(e) => setFormAssignEncounterId(e.target.value)}
                  >
                    <option value="">{t("Select IPD Encounter")}</option>
                    {filteredEncounters.map((enc) => (
                      <option key={enc.id} value={enc.id}>
                        {enc.id.slice(0, 8).toUpperCase()} - {enc.patientName || "IPD Patient"} (v{enc.version || 1}) [{enc.status}]
                      </option>
                    ))}
                  </select>
                  {formAssignPatientId && eligibleEncounters.length === 0 && (
                    <span className="fs-8 text-danger mt-1">
                      {t("No active IPD encounters found for this patient.")}
                    </span>
                  )}
                </div>

                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("Bed")}: <span className="text-danger">*</span>
                  </label>
                  <select
                    required
                    className="form-select-custom"
                    value={formAssignBedId}
                    onChange={(e) => setFormAssignBedId(e.target.value)}
                  >
                    <option value="">{t("Select Bed")}</option>
                    {beds.map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.bedName} ({b.bedType}) - {b.available ? t("Available") : t("Occupied")}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Description / Notes")}
                  className="form-textarea-custom"
                  value={formAssignDescription}
                  onChange={(e) => setFormAssignDescription(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button type="submit" className="btn-modal-save" disabled={isSubmitting}>
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setBedAssignModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Discharge Patient */}
      {dischargeModalOpen && assignmentToDischarge && (
        <div className="modal-backdrop-custom" onClick={() => setDischargeModalOpen(false)}>
          <div className="modal-card-custom" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("Discharge Patient")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setDischargeModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleConfirmDischarge} className="modal-form-custom">
              <div className="modal-body-custom mb-3">
                <p className="fs-7 text-muted">
                  {t("Discharge patient")} <strong>{assignmentToDischarge.patientName}</strong> {t("from bed")}{" "}
                  <strong>{assignmentToDischarge.bedName}</strong> ({assignmentToDischarge.ipdNo})?
                </p>
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Discharge Notes")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Clinical notes or reason for discharge...")}
                  className="form-textarea-custom"
                  value={dischargeNotes}
                  onChange={(e) => setDischargeNotes(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button
                  type="submit"
                  className="btn-delete-confirm-red"
                  style={{ backgroundColor: "#3b82f6", borderColor: "#3b82f6" }}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? t("Discharging...") : t("Confirm Discharge")}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setDischargeModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Confirm Delete */}
      {deleteConfirmOpen && itemToDelete && (
        <div className="modal-backdrop-custom" onClick={() => setDeleteConfirmOpen(false)}>
          <div className="modal-card-custom modal-sm" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("Confirm Delete")}</h3>
              <button
                type="button"
                className="modal-close-btn"
                onClick={() => setDeleteConfirmOpen(false)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <p>
                {t("Are you sure you want to delete")} {itemToDelete.name || t("this record")}?
              </p>
            </div>
            <div className="modal-footer-custom">
              <button
                type="button"
                className="btn-delete-confirm-red"
                onClick={handleDeleteConfirm}
                disabled={isSubmitting}
              >
                {isSubmitting ? t("Deleting...") : t("Delete")}
              </button>
              <button
                type="button"
                className="btn-modal-cancel"
                onClick={() => setDeleteConfirmOpen(false)}
                disabled={isSubmitting}
              >
                {t("Cancel")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
