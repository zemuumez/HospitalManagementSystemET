"use client";

import { useLanguage } from "@/components/language";
import { useState, useEffect, useRef } from "react";
import Link from "next/link";
import {
  Search,
  Filter,
  Plus,
  Trash2,
  Edit2,
  Eye,
  ArrowLeft,
  X,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from "lucide-react";

export type ServicesWorkspaceProps = {
  id?: string;
};

interface DiseaseItem {
  id?: string;
  name: string;
  charge: number;
}

interface InsuranceItem {
  id: string;
  name: string;
  serviceTax: number;
  discount: number;
  insuranceNo: string;
  insuranceCode: string;
  hospitalRate: number;
  remark: string;
  status: boolean;
  diseases: DiseaseItem[];
  totalAmount: number;
  serviceTaxMinor?: number;
  hospitalRateMinor?: number;
  totalAmountMinor?: number;
}

interface PackageServiceLine {
  id?: string;
  packageId?: string;
  serviceId: string;
  serviceName?: string;
  quantity: number;
  rateMinor: number;
  rate: number;
  amountMinor: number;
  amount: number;
}

interface PackageItem {
  id: string;
  name: string;
  description?: string;
  discount: number;
  totalAmountMinor: number;
  totalAmount: number;
  currencySymbol: string;
  services: PackageServiceLine[];
  createdAt?: string;
  updatedAt?: string;
}

interface ServiceItem {
  id: string;
  name: string;
  quantity: number;
  rate: number;
  status: boolean;
}

interface AmbulanceItem {
  id: string;
  vehicleNumber: string;
  vehicleModel: string;
  yearMade: number;
  driverName: string;
  driverLicense: string;
  driverContact: string;
  vehicleType: string;
  status: boolean;
}

interface AmbulanceCallItem {
  id: string;
  patientName: string;
  patientEmail: string;
  vehicleModel: string;
  driverName: string;
  date: string;
  time: string;
  amount: number;
}

export function ServicesWorkspace({
  id = "insurances",
}: ServicesWorkspaceProps) {
  const { t } = useLanguage();
  const currentTab = id;

  // View mode for Insurance: list or create
  const [insuranceMode, setInsuranceMode] = useState<"list" | "create">("list");

  // Search & Filter
  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);
  const isSearchMounted = useRef(false);

  // Insurances State
  const [insurances, setInsurances] = useState<InsuranceItem[]>([]);
  const [insurancesTotal, setInsurancesTotal] = useState(0);
  const [insurancesLoading, setInsurancesLoading] = useState(false);
  const [insurancesError, setInsurancesError] = useState<string | null>(null);
  const [pendingStatusIds, setPendingStatusIds] = useState<Set<string>>(
    new Set(),
  );
  const [insuranceEditModal, setInsuranceEditModal] = useState(false);
  const [insuranceDetailModal, setInsuranceDetailModal] = useState(false);
  const [selectedInsurance, setSelectedInsurance] =
    useState<InsuranceItem | null>(null);

  // New Insurance Form Fields
  const [newInsuranceName, setNewInsuranceName] = useState("");
  const [newServiceTax, setNewServiceTax] = useState("");
  const [newDiscount, setNewDiscount] = useState("0");
  const [newInsuranceNo, setNewInsuranceNo] = useState("");
  const [newInsuranceCode, setNewInsuranceCode] = useState("");
  const [newHospitalRate, setNewHospitalRate] = useState("");
  const [newRemark, setNewRemark] = useState("");
  const [newStatus, setNewStatus] = useState(true);
  const [newDiseases, setNewDiseases] = useState<
    { id: string; name: string; charge: string }[]
  >([{ id: "1", name: "", charge: "" }]);

  // Edit Insurance Form Fields
  const [editInsuranceId, setEditInsuranceId] = useState("");
  const [editInsuranceName, setEditInsuranceName] = useState("");
  const [editServiceTax, setEditServiceTax] = useState("");
  const [editDiscount, setEditDiscount] = useState("0");
  const [editInsuranceNo, setEditInsuranceNo] = useState("");
  const [editInsuranceCode, setEditInsuranceCode] = useState("");
  const [editHospitalRate, setEditHospitalRate] = useState("");
  const [editRemark, setEditRemark] = useState("");
  const [editStatus, setEditStatus] = useState(true);
  const [editDiseases, setEditDiseases] = useState<
    { id: string; name: string; charge: string }[]
  >([{ id: "1", name: "", charge: "" }]);

  // Packages State
  const [packages, setPackages] = useState<PackageItem[]>([]);
  const [packagesTotal, setPackagesTotal] = useState(0);
  const [packagesLoading, setPackagesLoading] = useState(false);
  const [packagesError, setPackagesError] = useState<string | null>(null);
  const [packageModal, setPackageModal] = useState(false);
  const [packageEditModal, setPackageEditModal] = useState(false);
  const [packageDetailModal, setPackageDetailModal] = useState(false);
  const [selectedPackage, setSelectedPackage] = useState<PackageItem | null>(
    null,
  );
  const [modalSrvFilter, setModalSrvFilter] = useState("");

  // New Package Form State
  const [pkgName, setPkgName] = useState("");
  const [pkgDescription, setPkgDescription] = useState("");
  const [pkgDiscount, setPkgDiscount] = useState("0");
  const [pkgLines, setPkgLines] = useState<
    { id: string; serviceId: string; quantity: string; rate: string }[]
  >([{ id: "1", serviceId: "", quantity: "1", rate: "0" }]);

  // Edit Package Form State
  const [editPkgName, setEditPkgName] = useState("");
  const [editPkgDescription, setEditPkgDescription] = useState("");
  const [editPkgDiscount, setEditPkgDiscount] = useState("0");
  const [editPkgLines, setEditPkgLines] = useState<
    {
      id?: string;
      serviceId: string;
      serviceName?: string;
      quantity: string;
      rate: string;
    }[]
  >([]);

  // Services State
  const [services, setServices] = useState<ServiceItem[]>([
    {
      id: "SRV-01",
      name: "General Consultation",
      quantity: 1,
      rate: 50,
      status: true,
    },
    {
      id: "SRV-02",
      name: "Blood Test Routine",
      quantity: 1,
      rate: 35,
      status: true,
    },
    {
      id: "SRV-03",
      name: "X-Ray Single View",
      quantity: 1,
      rate: 120,
      status: true,
    },
    {
      id: "SRV-04",
      name: "Physical Therapy Session",
      quantity: 1,
      rate: 80,
      status: true,
    },
    {
      id: "SRV-05",
      name: "Emergency Dressing",
      quantity: 2,
      rate: 45,
      status: true,
    },
  ]);
  const [servicesTotal, setServicesTotal] = useState(0);
  const [servicesLoading, setServicesLoading] = useState(false);
  const [servicesError, setServicesError] = useState<string | null>(null);
  const [catalogServices, setCatalogServices] = useState<ServiceItem[]>([]);
  const [catalogServicesLoading, setCatalogServicesLoading] = useState(false);
  const [catalogServicesError, setCatalogServicesError] = useState<
    string | null
  >(null);
  const [serviceModal, setServiceModal] = useState(false);
  const [srvName, setSrvName] = useState("");
  const [srvQuantity, setSrvQuantity] = useState("1");
  const [srvRate, setSrvRate] = useState("");

  // Ambulances State
  const [ambulances, setAmbulances] = useState<AmbulanceItem[]>([
    {
      id: "AMB-01",
      vehicleNumber: "AMB-2024-01",
      vehicleModel: "Toyota HiAce Ambulance",
      yearMade: 2023,
      driverName: "Daniel Mengistu",
      driverLicense: "DL-ETH-89412",
      driverContact: "+251911223344",
      vehicleType: "Owned",
      status: true,
    },
    {
      id: "AMB-02",
      vehicleNumber: "AMB-2023-09",
      vehicleModel: "Mercedes Sprinter ICU",
      yearMade: 2022,
      driverName: "Yonas Abebe",
      driverLicense: "DL-ETH-65719",
      driverContact: "+251922334455",
      vehicleType: "Contractual",
      status: true,
    },
  ]);
  const [ambulanceModal, setAmbulanceModal] = useState(false);
  const [ambNumber, setAmbNumber] = useState("");
  const [ambModel, setAmbModel] = useState("");
  const [ambYear, setAmbYear] = useState(new Date().getFullYear().toString());
  const [ambDriver, setAmbDriver] = useState("");
  const [ambLicense, setAmbLicense] = useState("");
  const [ambContact, setAmbContact] = useState("");
  const [ambType, setAmbType] = useState("Owned");

  // Ambulance Calls State
  const [ambulanceCalls, setAmbulanceCalls] = useState<AmbulanceCallItem[]>([
    {
      id: "CALL-01",
      patientName: "Abebe Kebede",
      patientEmail: "abebe@example.com",
      vehicleModel: "Toyota HiAce Ambulance",
      driverName: "Daniel Mengistu",
      date: "05 Oct, 2026",
      time: "02:15 PM",
      amount: 150,
    },
    {
      id: "CALL-02",
      patientName: "Sara Tesfaye",
      patientEmail: "sara@example.com",
      vehicleModel: "Mercedes Sprinter ICU",
      driverName: "Yonas Abebe",
      date: "04 Oct, 2026",
      time: "09:30 AM",
      amount: 220,
    },
  ]);
  const [callModal, setCallModal] = useState(false);
  const [callPatient, setCallPatient] = useState("");
  const [callVehicle, setCallVehicle] = useState("Toyota HiAce Ambulance");
  const [callDriver, setCallDriver] = useState("Daniel Mengistu");
  const [callAmount, setCallAmount] = useState("");

  // Subtabs list matching screenshot
  const tabs = [
    { id: "insurances", label: "Insurances", href: "/modules/insurances" },
    { id: "packages", label: "Packages", href: "/modules/packages" },
    { id: "services", label: "Services", href: "/modules/services" },
    { id: "ambulances", label: "Ambulances", href: "/modules/ambulances" },
    {
      id: "ambulance-calls",
      label: "Ambulance Calls",
      href: "/modules/ambulance-calls",
    },
  ];

  // Dynamic Disease Details Helpers for Insurance
  function resetNewInsuranceForm() {
    setNewInsuranceName("");
    setNewServiceTax("");
    setNewDiscount("0");
    setNewInsuranceNo("");
    setNewInsuranceCode("");
    setNewHospitalRate("");
    setNewRemark("");
    setNewStatus(true);
    setNewDiseases([{ id: "1", name: "", charge: "" }]);
  }

  function addDiseaseRow() {
    setNewDiseases((prev) => [
      ...prev,
      { id: String(Date.now()), name: "", charge: "" },
    ]);
  }

  function removeDiseaseRow(index: number) {
    if (newDiseases.length <= 1) return;
    setNewDiseases((prev) => prev.filter((_, i) => i !== index));
  }

  function updateDisease(
    index: number,
    field: "name" | "charge",
    value: string,
  ) {
    setNewDiseases((prev) =>
      prev.map((row, i) => (i === index ? { ...row, [field]: value } : row)),
    );
  }

  function startEditInsurance(item: InsuranceItem) {
    setSelectedInsurance(item);
    setEditInsuranceId(item.id);
    setEditInsuranceName(item.name);
    setEditServiceTax(item.serviceTax ? String(item.serviceTax) : "0");
    setEditDiscount(String(item.discount || 0));
    setEditInsuranceNo(item.insuranceNo || "");
    setEditInsuranceCode(item.insuranceCode || "");
    setEditHospitalRate(item.hospitalRate ? String(item.hospitalRate) : "0");
    setEditRemark(item.remark || "");
    setEditStatus(item.status);

    const diseasesToUse =
      Array.isArray(item.diseases) && item.diseases.length > 0
        ? item.diseases
        : [{ id: "1", name: "", charge: "0" }];

    setEditDiseases(
      diseasesToUse.map((d, idx) => ({
        id: d.id || String(idx + 1),
        name: d.name,
        charge: String(d.charge),
      })),
    );
    setInsuranceEditModal(true);
  }

  function addEditDiseaseRow() {
    setEditDiseases((prev) => [
      ...prev,
      { id: String(Date.now()), name: "", charge: "" },
    ]);
  }

  function removeEditDiseaseRow(index: number) {
    if (editDiseases.length <= 1) return;
    setEditDiseases((prev) => prev.filter((_, i) => i !== index));
  }

  function updateEditDisease(
    index: number,
    field: "name" | "charge",
    value: string,
  ) {
    setEditDiseases((prev) =>
      prev.map((row, i) => (i === index ? { ...row, [field]: value } : row)),
    );
  }

  async function openInsuranceDetails(item: InsuranceItem) {
    setSelectedInsurance(item);
    setInsuranceDetailModal(true);
    try {
      const res = await fetch(`/api/hms/insurances/${item.id}`);
      if (res.ok) {
        const full = await res.json();
        if (full && full.id === item.id) {
          setSelectedInsurance({
            id: full.id,
            name: full.name,
            serviceTax:
              full.service_tax ??
              full.serviceTax ??
              (full.service_tax_minor ?? full.serviceTaxMinor ?? 0) / 100,
            discount: full.discount ?? 0,
            insuranceNo: full.insurance_no || full.insuranceNo || "",
            insuranceCode: full.insurance_code || full.insuranceCode || "",
            hospitalRate:
              full.hospital_rate ??
              full.hospitalRate ??
              (full.hospital_rate_minor ?? full.hospitalRateMinor ?? 0) / 100,
            remark: full.remark || "",
            status: full.status === 1 || full.status === true,
            diseases: Array.isArray(full.diseases)
              ? full.diseases.map((d: any) => ({
                  id: d.id,
                  name: d.disease_name || d.diseaseName || d.name || "",
                  charge:
                    d.disease_charge ??
                    d.diseaseCharge ??
                    (d.disease_charge_minor ?? d.diseaseChargeMinor ?? 0) / 100,
                }))
              : [],
            totalAmount:
              full.total ??
              full.totalAmount ??
              (full.total_minor ?? full.totalAmountMinor ?? 0) / 100,
          });
        }
      }
    } catch {
      // Keep existing item
    }
  }

  function calculateInsurancePreviewTotals(
    taxStr: string,
    rateStr: string,
    discountStr: string,
    diseases: { name: string; charge: string }[],
  ) {
    const tax = Math.max(0, parseFloat(taxStr) || 0);
    const rate = Math.max(0, parseFloat(rateStr) || 0);
    const discount = Math.min(100, Math.max(0, parseInt(discountStr, 10) || 0));
    const diseaseSum = diseases.reduce(
      (sum, d) => sum + Math.max(0, parseFloat(d.charge) || 0),
      0,
    );
    const base = tax + rate + diseaseSum;
    const discountAmount = Math.round(base * discount) / 100;
    const total = Math.max(0, base - discountAmount);
    return { base, diseaseSum, discountAmount, total };
  }

  const createCalculatedPreview = calculateInsurancePreviewTotals(
    newServiceTax,
    newHospitalRate,
    newDiscount,
    newDiseases,
  );

  const editCalculatedPreview = calculateInsurancePreviewTotals(
    editServiceTax,
    editHospitalRate,
    editDiscount,
    editDiseases,
  );

  // Dynamic Service Lines Helpers for Packages
  function resetPackageForm() {
    setPkgName("");
    setPkgDescription("");
    setPkgDiscount("0");
    setPkgLines([{ id: "1", serviceId: "", quantity: "1", rate: "0" }]);
    setModalSrvFilter("");
    if (catalogServices.length === 0 || catalogServicesError) {
      loadActiveCatalogServices();
    }
  }

  function addPkgLine() {
    setPkgLines((prev) => [
      ...prev,
      { id: String(Date.now()), serviceId: "", quantity: "1", rate: "0" },
    ]);
  }

  function removePkgLine(index: number) {
    if (pkgLines.length <= 1) return;
    setPkgLines((prev) => prev.filter((_, i) => i !== index));
  }

  function updatePkgLine(
    index: number,
    field: "serviceId" | "quantity" | "rate",
    value: string,
  ) {
    const srvPool = catalogServices.length > 0 ? catalogServices : services;
    setPkgLines((prev) =>
      prev.map((row, i) => {
        if (i !== index) return row;
        if (field === "serviceId") {
          const selectedSrv = srvPool.find((s) => s.id === value);
          const defaultRate = selectedSrv ? String(selectedSrv.rate) : row.rate;
          return { ...row, serviceId: value, rate: defaultRate };
        }
        return { ...row, [field]: value };
      }),
    );
  }

  function addEditPkgLine() {
    setEditPkgLines((prev) => [
      ...prev,
      { serviceId: "", quantity: "1", rate: "0" },
    ]);
  }

  function removeEditPkgLine(index: number) {
    if (editPkgLines.length <= 1) return;
    setEditPkgLines((prev) => prev.filter((_, i) => i !== index));
  }

  function updateEditPkgLine(
    index: number,
    field: "serviceId" | "quantity" | "rate",
    value: string,
  ) {
    const srvPool = catalogServices.length > 0 ? catalogServices : services;
    setEditPkgLines((prev) =>
      prev.map((row, i) => {
        if (i !== index) return row;
        if (field === "serviceId") {
          const selectedSrv = srvPool.find((s) => s.id === value);
          const defaultRate = selectedSrv ? String(selectedSrv.rate) : row.rate;
          return {
            ...row,
            serviceId: value,
            serviceName: selectedSrv ? selectedSrv.name : row.serviceName,
            rate: defaultRate,
          };
        }
        return { ...row, [field]: value };
      }),
    );
  }

  function calculatePreviewTotals(
    lines: { serviceId: string; quantity: string; rate: string }[],
    discountPercent: string,
  ) {
    let subtotalMinor = 0;
    for (const line of lines) {
      const q = Math.max(0, parseInt(line.quantity, 10) || 0);
      const r = Math.max(0, Math.round((parseFloat(line.rate) || 0) * 100));
      subtotalMinor += q * r;
    }
    const disc = Math.max(0, Math.min(100, parseInt(discountPercent, 10) || 0));
    const discountAmountMinor = Math.floor((subtotalMinor * disc + 50) / 100);
    const totalAmountMinor = Math.max(0, subtotalMinor - discountAmountMinor);

    return {
      subtotal: subtotalMinor / 100,
      discountAmount: discountAmountMinor / 100,
      total: totalAmountMinor / 100,
    };
  }

  function openPackageEdit(pkg: PackageItem) {
    setSelectedPackage(pkg);
    setEditPkgName(pkg.name);
    setEditPkgDescription(pkg.description || "");
    setEditPkgDiscount(String(pkg.discount));
    setModalSrvFilter("");
    setEditPkgLines(
      pkg.services && pkg.services.length > 0
        ? pkg.services.map((s) => ({
            id: s.id,
            serviceId: s.serviceId,
            serviceName: s.serviceName,
            quantity: String(s.quantity),
            rate: String(s.rate),
          }))
        : [{ serviceId: "", quantity: "1", rate: "0" }],
    );
    if (catalogServices.length === 0 || catalogServicesError) {
      loadActiveCatalogServices();
    }
    setPackageEditModal(true);
  }

  function openPackageDetail(pkg: PackageItem) {
    setSelectedPackage(pkg);
    setPackageDetailModal(true);
  }

  const createPreview = calculatePreviewTotals(pkgLines, pkgDiscount);
  const editPreview = calculatePreviewTotals(editPkgLines, editPkgDiscount);

  /* -------------------------------------------------------------
     LIVE BACKEND API INTEGRATION (Go / PostgreSQL /v1/services & /v1/ambulances)
     ------------------------------------------------------------- */
  const [apiConnected, setApiConnected] = useState(false);
  const [isLoadingApi, setIsLoadingApi] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");

  async function loadActiveCatalogServices() {
    setCatalogServicesLoading(true);
    setCatalogServicesError(null);
    try {
      let currentPage = 1;
      let all: ServiceItem[] = [];
      const maxPages = 50; // safe bounded contract: up to 5,000 active services
      while (currentPage <= maxPages) {
        const res = await fetch(
          `/api/hms/services?page=${currentPage}&limit=100&status=1`,
        );
        if (!res.ok) {
          const err = await res.json().catch(() => ({}));
          throw new Error(
            err.error ||
              err.message ||
              `Failed to load active services catalog (HTTP ${res.status})`,
          );
        }
        const data = await res.json();
        const items = Array.isArray(data.services) ? data.services : [];
        const total =
          typeof data.total === "number" ? data.total : items.length;
        const mapped: ServiceItem[] = items.map((s: any) => ({
          id: s.id,
          name: s.name,
          quantity: s.quantity || 1,
          rate: (s.rate_minor ?? s.rateMinor ?? 0) / 100,
          status: s.status === 1,
        }));
        all = all.concat(mapped);
        setCatalogServices([...all]);
        if (items.length < 100 || all.length >= total) {
          break;
        }
        currentPage++;
      }
      setCatalogServices(all);
    } catch (err: any) {
      setCatalogServices([]);
      setCatalogServicesError(
        err?.message || "Failed to load active services catalog",
      );
    } finally {
      setCatalogServicesLoading(false);
    }
  }

  async function loadPackagesData(p = page, ps = pageSize, s = search) {
    setPackagesLoading(true);
    setPackagesError(null);
    try {
      const q = new URLSearchParams({
        page: String(p),
        limit: String(ps),
      });
      if (s.trim()) q.set("search", s.trim());

      const [pkgRes] = await Promise.all([
        fetch(`/api/hms/packages?${q.toString()}`),
        loadActiveCatalogServices(),
      ]);

      if (!pkgRes.ok) {
        setPackages([]);
        const err = await pkgRes.json().catch(() => ({}));
        const statusMsg =
          err.error ||
          err.message ||
          (pkgRes.status === 403
            ? "Access denied: insufficient permissions to view medical packages (HTTP 403)"
            : pkgRes.status >= 500
              ? `Server error loading medical packages (HTTP ${pkgRes.status})`
              : `Failed to load medical packages (HTTP ${pkgRes.status})`);
        throw new Error(statusMsg);
      }

      const pkgData = await pkgRes.json();
      const items = Array.isArray(pkgData.packages) ? pkgData.packages : [];
      setPackages(
        items.map((p: any) => ({
          id: p.id,
          name: p.name,
          description: p.description || "",
          discount: p.discount || 0,
          totalAmountMinor: p.total_amount_minor ?? p.totalAmountMinor ?? 0,
          totalAmount: (p.total_amount_minor ?? p.totalAmountMinor ?? 0) / 100,
          currencySymbol: p.currency_symbol || p.currencySymbol || "ETB",
          services: Array.isArray(p.services)
            ? p.services.map((s: any) => ({
                id: s.id,
                packageId: s.package_id || s.packageId,
                serviceId: s.service_id || s.serviceId,
                serviceName: s.service_name || s.serviceName || "",
                quantity: s.quantity || 1,
                rateMinor: s.rate_minor ?? s.rateMinor ?? 0,
                rate: (s.rate_minor ?? s.rateMinor ?? 0) / 100,
                amountMinor: s.amount_minor ?? s.amountMinor ?? 0,
                amount: (s.amount_minor ?? s.amountMinor ?? 0) / 100,
              }))
            : [],
          createdAt: p.created_at || p.createdAt,
          updatedAt: p.updated_at || p.updatedAt,
        })),
      );
      setPackagesTotal(
        typeof pkgData.total === "number" ? pkgData.total : items.length,
      );
      setApiConnected(true);
    } catch (err: any) {
      setPackages([]);
      setPackagesError(err?.message || "Failed to load medical packages");
      setApiConnected(false);
    } finally {
      setPackagesLoading(false);
      setIsLoadingApi(false);
    }
  }

  async function loadInsurancesData(p = page, ps = pageSize, s = search) {
    setInsurancesLoading(true);
    setInsurancesError(null);
    try {
      const q = new URLSearchParams({
        page: String(p),
        limit: String(ps),
      });
      if (s.trim()) q.set("search", s.trim());

      const insRes = await fetch(`/api/hms/insurances?${q.toString()}`);
      if (!insRes.ok) {
        setInsurances([]);
        const err = await insRes.json().catch(() => ({}));
        const statusMsg =
          err.error ||
          err.message ||
          (insRes.status === 403
            ? "Access denied: insufficient permissions to view insurance policies (HTTP 403)"
            : insRes.status >= 500
              ? `Server error loading insurance policies (HTTP ${insRes.status})`
              : `Failed to load insurance policies (HTTP ${insRes.status})`);
        throw new Error(statusMsg);
      }

      const insData = await insRes.json();
      const items = Array.isArray(insData.insurances) ? insData.insurances : [];
      setInsurances(
        items.map((i: any) => ({
          id: i.id,
          name: i.name,
          serviceTax:
            i.service_tax ??
            i.serviceTax ??
            (i.service_tax_minor ?? i.serviceTaxMinor ?? 0) / 100,
          serviceTaxMinor:
            i.service_tax_minor ??
            i.serviceTaxMinor ??
            Math.round((i.service_tax ?? i.serviceTax ?? 0) * 100),
          discount: i.discount ?? 0,
          insuranceNo: i.insurance_no || i.insuranceNo || "",
          insuranceCode: i.insurance_code || i.insuranceCode || "",
          hospitalRate:
            i.hospital_rate ??
            i.hospitalRate ??
            (i.hospital_rate_minor ?? i.hospitalRateMinor ?? 0) / 100,
          hospitalRateMinor:
            i.hospital_rate_minor ??
            i.hospitalRateMinor ??
            Math.round((i.hospital_rate ?? i.hospitalRate ?? 0) * 100),
          remark: i.remark || "",
          status: i.status === 1 || i.status === true,
          diseases: Array.isArray(i.diseases)
            ? i.diseases.map((d: any) => ({
                id: d.id,
                name: d.disease_name || d.diseaseName || d.name || "",
                charge:
                  d.disease_charge ??
                  d.diseaseCharge ??
                  (d.disease_charge_minor ?? d.diseaseChargeMinor ?? 0) / 100,
              }))
            : [],
          totalAmount:
            i.total ??
            i.totalAmount ??
            (i.total_minor ?? i.totalAmountMinor ?? 0) / 100,
          totalAmountMinor:
            i.total_minor ??
            i.totalAmountMinor ??
            Math.round((i.total ?? i.totalAmount ?? 0) * 100),
        })),
      );
      setInsurancesTotal(
        typeof insData.total === "number" ? insData.total : items.length,
      );
      setApiConnected(true);
    } catch (err: any) {
      setInsurances([]);
      setInsurancesError(err?.message || "Failed to load insurance policies");
      setApiConnected(false);
    } finally {
      setInsurancesLoading(false);
      setIsLoadingApi(false);
    }
  }

  async function loadServicesListData(p = page, ps = pageSize, s = search) {
    setServicesLoading(true);
    setServicesError(null);
    try {
      const q = new URLSearchParams({
        page: String(p),
        limit: String(ps),
      });
      if (s.trim()) q.set("search", s.trim());

      const srvRes = await fetch(`/api/hms/services?${q.toString()}`);
      if (!srvRes.ok) {
        setServices([]);
        const err = await srvRes.json().catch(() => ({}));
        const statusMsg =
          err.error ||
          err.message ||
          (srvRes.status === 403
            ? "Access denied: insufficient permissions to view services (HTTP 403)"
            : srvRes.status >= 500
              ? `Server error loading services (HTTP ${srvRes.status})`
              : `Failed to load services (HTTP ${srvRes.status})`);
        throw new Error(statusMsg);
      }

      const srvData = await srvRes.json();
      const items = Array.isArray(srvData.services) ? srvData.services : [];
      setServices(
        items.map((s: any) => ({
          id: s.id,
          name: s.name,
          quantity: s.quantity || 1,
          rate: (s.rate_minor ?? s.rateMinor ?? 0) / 100,
          status: s.status === 1,
        })),
      );
      setServicesTotal(
        typeof srvData.total === "number" ? srvData.total : items.length,
      );
      setApiConnected(true);
    } catch (err: any) {
      setServices([]);
      setServicesError(err?.message || "Failed to load services");
      setApiConnected(false);
    } finally {
      setServicesLoading(false);
      setIsLoadingApi(false);
    }
  }

  async function loadAmbulancesData() {
    setIsLoadingApi(true);
    try {
      const ambRes = await fetch("/api/hms/ambulances");
      if (ambRes.ok) {
        const ambData = await ambRes.json();
        if (Array.isArray(ambData.ambulances)) {
          setAmbulances(
            ambData.ambulances.map((a: any) => ({
              id: a.id,
              vehicleNumber: a.vehicleNumber || a.vehicle_number || "AMB-01",
              vehicleModel: a.vehicleModel || a.vehicle_model || "Ambulance",
              yearMade: a.yearMade || a.year_made || 2023,
              driverName: a.driverName || a.driver_name || "Driver",
              driverLicense: a.driverLicense || a.driver_license || "DL-01",
              driverContact:
                a.driverContact || a.driver_contact || "+251911000000",
              vehicleType: a.vehicleType || a.vehicle_type || "Owned",
              status: a.isAvailable ?? true,
            })),
          );
        }
        setApiConnected(true);
      } else {
        setApiConnected(false);
      }
    } catch {
      setApiConnected(false);
    } finally {
      setIsLoadingApi(false);
    }
  }

  async function loadAmbulanceCallsData() {
    setIsLoadingApi(true);
    try {
      const callRes = await fetch("/api/hms/ambulance-calls");
      if (callRes.ok) {
        const callData = await callRes.json();
        if (Array.isArray(callData.ambulance_calls)) {
          setAmbulanceCalls(
            callData.ambulance_calls.map((c: any) => ({
              id: c.id,
              patientName: c.patientName || "Patient",
              patientEmail: "patient@hospital.et",
              vehicleModel: c.vehicleModel || "Ambulance",
              driverName: c.driverName || "Driver",
              date: "05 Oct, 2026",
              time: "02:00 PM",
              amount: (c.amountMinor || 0) / 100,
            })),
          );
        }
        setApiConnected(true);
      } else {
        setApiConnected(false);
      }
    } catch {
      setApiConnected(false);
    } finally {
      setIsLoadingApi(false);
    }
  }

  async function loadServicesData(p = page, ps = pageSize, s = search) {
    if (currentTab === "packages") {
      await loadPackagesData(p, ps, s);
    } else if (currentTab === "insurances") {
      await loadInsurancesData(p, ps, s);
    } else if (currentTab === "services") {
      await loadServicesListData(p, ps, s);
    } else if (currentTab === "ambulances") {
      await loadAmbulancesData();
    } else if (currentTab === "ambulance-calls") {
      await loadAmbulanceCallsData();
    } else {
      await Promise.all([
        loadPackagesData(p, ps, s),
        loadInsurancesData(p, ps, s),
        loadServicesListData(p, ps, s),
      ]);
    }
  }

  function handlePageChange(newPage: number) {
    setPage(newPage);
    loadServicesData(newPage, pageSize, search);
  }

  function handlePageSizeChange(newPageSize: number) {
    setPageSize(newPageSize);
    setPage(1);
    loadServicesData(1, newPageSize, search);
  }

  function handleRetry() {
    loadServicesData(page, pageSize, search);
  }

  function renderPagination(total: number, currentCount: number) {
    const totalPages = Math.max(1, Math.ceil(total / pageSize));
    const startRecord = total > 0 ? (page - 1) * pageSize + 1 : 0;
    const endRecord = total > 0 ? Math.min(page * pageSize, total) : 0;

    const pageNumbers: number[] = [];
    const maxButtons = 5;
    let startPage = Math.max(1, page - 2);
    let endPage = Math.min(totalPages, startPage + maxButtons - 1);
    if (endPage - startPage < maxButtons - 1) {
      startPage = Math.max(1, endPage - maxButtons + 1);
    }
    for (let i = startPage; i <= endPage; i++) {
      pageNumbers.push(i);
    }

    return (
      <div className="billing-pagination d-flex justify-content-between align-items-center mt-3">
        <div className="d-flex align-items-center gap-2">
          <span>{t("Show")}</span>
          <select
            className="form-select-custom"
            value={pageSize}
            onChange={(e) => handlePageSizeChange(Number(e.target.value))}
          >
            <option value={10}>10</option>
            <option value={25}>25</option>
            <option value={50}>50</option>
            <option value={100}>100</option>
          </select>
          <span>
            {t("Showing")} {startRecord} {t("to")} {endRecord} {t("of")} {total}{" "}
            {t("Results")}
          </span>
        </div>
        <div className="billing-pagination-controls">
          <button
            type="button"
            className="billing-page-nav-btn"
            disabled={page <= 1}
            onClick={() => handlePageChange(page - 1)}
            aria-label="Previous page"
          >
            &laquo;
          </button>
          {pageNumbers.map((p) => (
            <button
              key={p}
              type="button"
              className={`billing-page-btn ${p === page ? "active" : ""}`}
              onClick={() => handlePageChange(p)}
            >
              {p}
            </button>
          ))}
          <button
            type="button"
            className="billing-page-nav-btn"
            disabled={page >= totalPages}
            onClick={() => handlePageChange(page + 1)}
            aria-label="Next page"
          >
            &raquo;
          </button>
        </div>
      </div>
    );
  }

  useEffect(() => {
    loadServicesData(page, pageSize, search);
  }, []);

  useEffect(() => {
    if (!isSearchMounted.current) {
      isSearchMounted.current = true;
      return;
    }
    const timer = setTimeout(() => {
      setPage(1);
      loadServicesData(1, pageSize, search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  async function handleCreateService(e: React.FormEvent) {
    e.preventDefault();
    if (!srvName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");
    const rateVal = parseFloat(srvRate) || 0;
    const qtyVal = parseInt(srvQuantity) || 1;

    try {
      const res = await fetch("/api/hms/services", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: srvName.trim(),
          quantity: qtyVal,
          rateMinor: Math.round(rateVal * 100),
          status: 1,
        }),
      });
      if (res.ok) {
        const saved = await res.json();
        setServices((prev) => [
          {
            id: saved.id || `SRV-${Date.now()}`,
            name: saved.name || srvName.trim(),
            quantity: qtyVal,
            rate: rateVal,
            status: true,
          },
          ...prev,
        ]);
        setApiSuccessBanner(
          t("Hospital service successfully registered in PostgreSQL backend"),
        );
        setServiceModal(false);
        setSrvName("");
        setSrvRate("");
        setSrvQuantity("1");
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }
    setServices((prev) => [
      {
        id: `SRV-${prev.length + 1}`,
        name: srvName.trim(),
        quantity: qtyVal,
        rate: rateVal,
        status: true,
      },
      ...prev,
    ]);
    setServiceModal(false);
    setSrvName("");
    setSrvRate("");
    setSrvQuantity("1");
  }

  async function handleCreateAmbulance(e: React.FormEvent) {
    e.preventDefault();
    if (!ambNumber.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const res = await fetch("/api/hms/ambulances", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          vehicleNumber: ambNumber.trim(),
          vehicleModel: ambModel.trim() || "Standard Ambulance",
          yearMade: parseInt(ambYear) || new Date().getFullYear(),
          driverName: ambDriver.trim() || "Driver",
          driverLicense: ambLicense.trim() || "DL-DEFAULT",
          driverContact: ambContact.trim() || "+251911000000",
          vehicleType: ambType,
          isAvailable: true,
        }),
      });
      if (res.ok) {
        const saved = await res.json();
        setAmbulances((prev) => [
          {
            id: saved.id || `AMB-${Date.now()}`,
            vehicleNumber: ambNumber.trim(),
            vehicleModel: ambModel.trim() || "Standard Ambulance",
            yearMade: parseInt(ambYear) || new Date().getFullYear(),
            driverName: ambDriver.trim() || "Driver",
            driverLicense: ambLicense.trim() || "DL-DEFAULT",
            driverContact: ambContact.trim() || "+251911000000",
            vehicleType: ambType,
            status: true,
          },
          ...prev,
        ]);
        setApiSuccessBanner(
          t("Ambulance successfully registered in emergency dispatch registry"),
        );
        setAmbulanceModal(false);
        setAmbNumber("");
        setAmbModel("");
        setAmbDriver("");
        setAmbLicense("");
        setAmbContact("");
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }
    setAmbulances((prev) => [
      {
        id: `AMB-${prev.length + 1}`,
        vehicleNumber: ambNumber.trim(),
        vehicleModel: ambModel.trim(),
        yearMade: parseInt(ambYear) || new Date().getFullYear(),
        driverName: ambDriver.trim(),
        driverLicense: ambLicense.trim(),
        driverContact: ambContact.trim(),
        vehicleType: ambType,
        status: true,
      },
      ...prev,
    ]);
    setAmbulanceModal(false);
    setAmbNumber("");
    setAmbModel("");
    setAmbDriver("");
    setAmbLicense("");
    setAmbContact("");
  }

  async function handleCreateAmbulanceCall(e: React.FormEvent) {
    e.preventDefault();
    if (!callPatient.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");
    const amt = parseFloat(callAmount) || 0;

    try {
      const res = await fetch("/api/hms/ambulance-calls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          patientName: callPatient.trim(),
          vehicleModel: callVehicle,
          driverName: callDriver,
          amountMinor: Math.round(amt * 100),
          status: "completed",
        }),
      });
      if (res.ok) {
        setApiSuccessBanner(t("Ambulance call record created and logged"));
        loadServicesData();
        setCallModal(false);
        setCallPatient("");
        setCallAmount("");
        setIsSubmitting(false);
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }
    setAmbulanceCalls((prev) => [
      {
        id: `CALL-${prev.length + 1}`,
        patientName: callPatient.trim(),
        patientEmail: "patient@hospital.et",
        vehicleModel: callVehicle,
        driverName: callDriver,
        date: "05 Oct, 2026",
        time: "02:30 PM",
        amount: amt,
      },
      ...prev,
    ]);
    setCallModal(false);
    setCallPatient("");
    setCallAmount("");
  }

  async function handleCreatePackage(e: React.FormEvent) {
    e.preventDefault();
    if (!pkgName.trim()) {
      setApiErrorBanner(t("Package name is required"));
      return;
    }
    const validLines = pkgLines.filter((l) => l.serviceId);
    if (validLines.length === 0) {
      setApiErrorBanner(t("Package must contain at least one service"));
      return;
    }

    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const payload = {
        name: pkgName.trim(),
        description: pkgDescription.trim(),
        discount: Math.max(0, Math.min(100, parseInt(pkgDiscount, 10) || 0)),
        services: validLines.map((l) => ({
          service_id: l.serviceId,
          serviceId: l.serviceId,
          quantity: Math.max(1, parseInt(l.quantity, 10) || 1),
          rate_minor: Math.round((parseFloat(l.rate) || 0) * 100),
          rateMinor: Math.round((parseFloat(l.rate) || 0) * 100),
        })),
      };

      const res = await fetch("/api/hms/packages", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Medical package registered in catalog"));
        setPackageModal(false);
        resetPackageForm();
        await loadServicesData();
        return;
      }

      const err = await res.json().catch(() => ({}));
      setApiErrorBanner(
        err.message || err.error || t("Failed to save package"),
      );
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleUpdatePackage(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedPackage?.id || !editPkgName.trim()) {
      setApiErrorBanner(t("Package name is required"));
      return;
    }
    const validLines = editPkgLines.filter((l) => l.serviceId);
    if (validLines.length === 0) {
      setApiErrorBanner(t("Package must contain at least one service"));
      return;
    }

    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const payload = {
        name: editPkgName.trim(),
        description: editPkgDescription.trim(),
        discount: Math.max(
          0,
          Math.min(100, parseInt(editPkgDiscount, 10) || 0),
        ),
        services: validLines.map((l) => ({
          ...(l.id ? { id: l.id } : {}),
          service_id: l.serviceId,
          serviceId: l.serviceId,
          quantity: Math.max(1, parseInt(l.quantity, 10) || 1),
          rate_minor: Math.round((parseFloat(l.rate) || 0) * 100),
          rateMinor: Math.round((parseFloat(l.rate) || 0) * 100),
        })),
      };

      const res = await fetch(`/api/hms/packages/${selectedPackage.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Medical package updated successfully"));
        setPackageEditModal(false);
        setSelectedPackage(null);
        await loadServicesData();
        return;
      }

      const err = await res.json().catch(() => ({}));
      setApiErrorBanner(
        err.message || err.error || t("Failed to update package"),
      );
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDeletePackage(pkgId: string) {
    if (!window.confirm(t("Are you sure you want to delete this package?"))) {
      return;
    }

    setApiErrorBanner("");
    setApiSuccessBanner("");
    setIsSubmitting(true);

    try {
      const res = await fetch(`/api/hms/packages/${pkgId}`, {
        method: "DELETE",
      });

      if (res.status === 204 || res.ok) {
        setApiSuccessBanner(t("Package removed from catalog"));
        await loadServicesData();
        return;
      }

      const err = await res.json().catch(() => ({}));
      if (res.status === 409 || err.error === "RECORD_IN_USE") {
        setApiErrorBanner(
          err.message ||
            t("Package is in use by patient admissions and cannot be deleted"),
        );
      } else {
        setApiErrorBanner(
          err.message || err.error || t("Failed to delete package"),
        );
      }
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleSaveInsurance(e: React.FormEvent) {
    e.preventDefault();
    if (!newInsuranceName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const payload = {
        name: newInsuranceName.trim(),
        service_tax_minor: Math.round((parseFloat(newServiceTax) || 0) * 100),
        serviceTaxMinor: Math.round((parseFloat(newServiceTax) || 0) * 100),
        discount: Math.max(0, Math.min(100, parseInt(newDiscount, 10) || 0)),
        insurance_no: newInsuranceNo.trim(),
        insuranceNo: newInsuranceNo.trim(),
        insurance_code: newInsuranceCode.trim(),
        insuranceCode: newInsuranceCode.trim(),
        hospital_rate_minor: Math.round(
          (parseFloat(newHospitalRate) || 0) * 100,
        ),
        hospitalRateMinor: Math.round((parseFloat(newHospitalRate) || 0) * 100),
        remark: newRemark.trim(),
        status: newStatus ? 1 : 0,
        diseases: newDiseases
          .filter((d) => d.name.trim() !== "")
          .map((d) => ({
            disease_name: d.name.trim(),
            diseaseName: d.name.trim(),
            name: d.name.trim(),
            disease_charge_minor: Math.round((parseFloat(d.charge) || 0) * 100),
            diseaseChargeMinor: Math.round((parseFloat(d.charge) || 0) * 100),
          })),
      };

      const res = await fetch("/api/hms/insurances", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Insurance policy successfully registered"));
        resetNewInsuranceForm();
        setInsuranceMode("list");
        await loadServicesData();
        return;
      }

      const err = await res.json().catch(() => ({}));
      setApiErrorBanner(
        err.message || err.error || t("Failed to save insurance"),
      );
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleUpdateInsurance(e?: React.FormEvent | React.MouseEvent) {
    if (e) {
      e.preventDefault();
    }
    if (isSubmitting) return;
    const id = editInsuranceId || selectedInsurance?.id;
    if (!id || !editInsuranceName.trim()) {
      setApiErrorBanner(t("Insurance name is required"));
      return;
    }

    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const payload = {
        name: editInsuranceName.trim(),
        service_tax_minor: Math.round((parseFloat(editServiceTax) || 0) * 100),
        serviceTaxMinor: Math.round((parseFloat(editServiceTax) || 0) * 100),
        discount: Math.max(0, Math.min(100, parseInt(editDiscount, 10) || 0)),
        insurance_no: editInsuranceNo.trim(),
        insuranceNo: editInsuranceNo.trim(),
        insurance_code: editInsuranceCode.trim(),
        insuranceCode: editInsuranceCode.trim(),
        hospital_rate_minor: Math.round(
          (parseFloat(editHospitalRate) || 0) * 100,
        ),
        hospitalRateMinor: Math.round(
          (parseFloat(editHospitalRate) || 0) * 100,
        ),
        remark: editRemark.trim(),
        status: editStatus ? 1 : 0,
        diseases: editDiseases
          .filter((d) => d.name.trim() !== "")
          .map((d) => ({
            ...(d.id && !d.id.startsWith("new") && !/^\d{13}$/.test(d.id)
              ? { id: d.id }
              : {}),
            disease_name: d.name.trim(),
            diseaseName: d.name.trim(),
            name: d.name.trim(),
            disease_charge_minor: Math.round((parseFloat(d.charge) || 0) * 100),
            diseaseChargeMinor: Math.round((parseFloat(d.charge) || 0) * 100),
          })),
      };

      const res = await fetch(`/api/hms/insurances/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setApiSuccessBanner(t("Insurance policy updated successfully"));
        setInsuranceEditModal(false);
        setSelectedInsurance(null);
        await loadServicesData();
        return;
      }

      const err = await res.json().catch(() => ({}));
      setApiErrorBanner(
        err.message || err.error || t("Failed to update insurance"),
      );
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleToggleInsuranceStatus(item: InsuranceItem) {
    if (pendingStatusIds.has(item.id)) return;
    setApiErrorBanner("");
    setApiSuccessBanner("");
    const targetStatus = item.status ? 0 : 1;

    setPendingStatusIds((prev) => new Set(prev).add(item.id));
    try {
      const res = await fetch(`/api/hms/insurances/${item.id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: targetStatus }),
      });
      if (res.ok) {
        await loadServicesData(page, pageSize, search);
        return;
      }
      const err = await res.json().catch(() => ({}));
      setApiErrorBanner(
        err.message || err.error || t("Failed to update status"),
      );
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setPendingStatusIds((prev) => {
        const next = new Set(prev);
        next.delete(item.id);
        return next;
      });
    }
  }

  async function handleDeleteInsurance(id: string) {
    if (!window.confirm(t("Are you sure you want to delete this insurance?"))) {
      return;
    }

    setApiErrorBanner("");
    setApiSuccessBanner("");
    setIsSubmitting(true);

    try {
      const res = await fetch(`/api/hms/insurances/${id}`, {
        method: "DELETE",
      });

      if (res.status === 204 || res.ok) {
        setApiSuccessBanner(t("Insurance removed from catalog"));
        await loadServicesData();
        return;
      }

      const err = await res.json().catch(() => ({}));
      if (res.status === 409 || err.error === "RECORD_IN_USE") {
        setApiErrorBanner(
          err.message ||
            t(
              "Insurance is in use by patient admissions and cannot be deleted",
            ),
        );
      } else {
        setApiErrorBanner(
          err.message || err.error || t("Failed to delete insurance"),
        );
      }
    } catch {
      setApiErrorBanner(
        t("The hospital service is unavailable. Please try again shortly."),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  /* -------------------------------------------------------------
     RENDER: NEW INSURANCE FULL-PAGE VIEW (MATCHING SCREENSHOT 185115)
     ------------------------------------------------------------- */
  if (currentTab === "insurances" && insuranceMode === "create") {
    return (
      <div
        className="legacy-workspace"
        data-ready={!isLoadingApi ? "true" : "false"}
      >
        {/* Top subtabs */}
        <div className="module-subtabs-nav">
          {tabs.map((tab) => (
            <Link
              key={tab.id}
              href={tab.href}
              className={`module-subtab-link ${currentTab === tab.id ? "active" : ""}`}
            >
              {t(tab.label)}
            </Link>
          ))}
        </div>

        <div className="form-card-container">
          {apiErrorBanner && (
            <div
              className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded mb-3"
              style={{
                borderColor: "#ef4444",
                backgroundColor: "rgba(239, 68, 68, 0.1)",
                color: "#ef4444",
              }}
            >
              <span>{apiErrorBanner}</span>
              <button
                type="button"
                className="btn-icon-link"
                onClick={() => setApiErrorBanner("")}
              >
                <X size={14} />
              </button>
            </div>
          )}

          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2 className="workspace-heading m-0">{t("New Insurance")}</h2>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => {
                resetNewInsuranceForm();
                setInsuranceMode("list");
              }}
            >
              {t("Back")}
            </button>
          </div>

          <form onSubmit={handleSaveInsurance}>
            <div className="form-grid-2">
              <div className="form-group-custom">
                <label>
                  {t("Insurance")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Insurance")}
                  value={newInsuranceName}
                  onChange={(e) => setNewInsuranceName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Service Tax")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Service Tax")}
                  value={newServiceTax}
                  onChange={(e) => setNewServiceTax(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Discount: (In Percentage(%))")}</label>
                <input
                  type="number"
                  min="0"
                  max="100"
                  value={newDiscount}
                  onChange={(e) => setNewDiscount(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Insurance No")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Insurance No")}
                  value={newInsuranceNo}
                  onChange={(e) => setNewInsuranceNo(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Insurance Code")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Insurance Code")}
                  value={newInsuranceCode}
                  onChange={(e) => setNewInsuranceCode(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Hospital Rate")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Hospital Rate")}
                  value={newHospitalRate}
                  onChange={(e) => setNewHospitalRate(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Remark")}:</label>
                <textarea
                  rows={4}
                  placeholder={t("Remark")}
                  value={newRemark}
                  onChange={(e) => setNewRemark(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Status")}:</label>
                <div className="mt-2">
                  <label className="switch-toggle">
                    <input
                      type="checkbox"
                      checked={newStatus}
                      onChange={(e) => setNewStatus(e.target.checked)}
                    />
                    <span className="slider-toggle"></span>
                  </label>
                </div>
              </div>
            </div>

            {/* Disease Details Table Section */}
            <div className="mt-4 pt-3 border-top border-secondary-subtle">
              <div className="d-flex justify-content-between align-items-center mb-3">
                <h4 className="m-0 fs-5 fw-semibold">{t("Disease Details")}</h4>
                <button
                  type="button"
                  className="btn-action-blue px-3 py-1 fs-6"
                  onClick={addDiseaseRow}
                >
                  {t("Add")}
                </button>
              </div>

              <table className="billing-table w-100">
                <thead>
                  <tr>
                    <th style={{ width: "60px" }}>#</th>
                    <th>
                      {t("DISEASES NAME")}{" "}
                      <span className="text-danger">*</span>
                    </th>
                    <th>
                      {t("DISEASES CHARGE")}{" "}
                      <span className="text-danger">*</span>
                    </th>
                    <th style={{ width: "80px" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {newDiseases.map((row, idx) => (
                    <tr key={row.id}>
                      <td>{idx + 1}</td>
                      <td>
                        <input
                          type="text"
                          required
                          className="form-control-custom w-100"
                          placeholder={t("Diseases Name")}
                          value={row.name}
                          onChange={(e) =>
                            updateDisease(idx, "name", e.target.value)
                          }
                        />
                      </td>
                      <td>
                        <input
                          type="number"
                          step="0.01"
                          required
                          className="form-control-custom w-100"
                          placeholder={t("Diseases charge")}
                          value={row.charge}
                          onChange={(e) =>
                            updateDisease(idx, "charge", e.target.value)
                          }
                        />
                      </td>
                      <td className="text-center">
                        <button
                          type="button"
                          className="btn-icon-danger"
                          disabled={newDiseases.length <= 1}
                          onClick={() => removeDiseaseRow(idx)}
                        >
                          <Trash2 size={16} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              <div className="d-flex flex-column align-items-end mt-3 gap-1 fs-7">
                <div>
                  {t("Service Tax")}: ETB{" "}
                  {(parseFloat(newServiceTax) || 0).toFixed(2)}
                </div>
                <div>
                  {t("Hospital Rate")}: ETB{" "}
                  {(parseFloat(newHospitalRate) || 0).toFixed(2)}
                </div>
                <div>
                  {t("Disease Details")}: ETB{" "}
                  {createCalculatedPreview.diseaseSum.toFixed(2)}
                </div>
                <div>
                  {t("Subtotal")}: ETB {createCalculatedPreview.base.toFixed(2)}
                </div>
                <div>
                  {t("Discount (%)")}: -ETB{" "}
                  {createCalculatedPreview.discountAmount.toFixed(2)} (
                  {newDiscount}%)
                </div>
                <div className="fs-6 fw-bold text-primary mt-1">
                  {t("Total Amount")}: ETB{" "}
                  {createCalculatedPreview.total.toFixed(2)}
                </div>
              </div>
            </div>

            <div className="d-flex justify-content-end gap-2 mt-4">
              <button
                type="submit"
                className="btn-action-blue px-4 py-2"
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
              <button
                type="button"
                className="btn-action-grey px-4 py-2"
                disabled={isSubmitting}
                onClick={() => {
                  resetNewInsuranceForm();
                  setInsuranceMode("list");
                }}
              >
                {t("Cancel")}
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  /* -------------------------------------------------------------
     RENDER: STANDARD TABBED LISTING VIEW
     ------------------------------------------------------------- */
  return (
    <div
      className="legacy-workspace"
      data-ready={!isLoadingApi ? "true" : "false"}
    >
      {/* Top subtabs */}
      <div className="module-subtabs-nav">
        {tabs.map((tab) => (
          <Link
            key={tab.id}
            href={tab.href}
            className={`module-subtab-link ${currentTab === tab.id ? "active" : ""}`}
          >
            {t(tab.label)}
          </Link>
        ))}
      </div>

      {/* Backend API Connection Banner */}
      <div className="d-flex flex-column gap-2 mb-3">
        {(() => {
          const tabLoading =
            currentTab === "packages"
              ? packagesLoading
              : currentTab === "insurances"
                ? insurancesLoading
                : currentTab === "services"
                  ? servicesLoading
                  : isLoadingApi;

          const tabError =
            currentTab === "packages"
              ? packagesError
              : currentTab === "insurances"
                ? insurancesError
                : currentTab === "services"
                  ? servicesError
                  : null;

          const isConnected = apiConnected && !tabError;

          return (
            <div
              className="d-flex align-items-center justify-content-between p-2 px-3 rounded border"
              style={{
                backgroundColor: tabError
                  ? "rgba(239, 68, 68, 0.08)"
                  : isConnected
                    ? "rgba(16, 185, 129, 0.08)"
                    : "rgba(59, 130, 246, 0.08)",
                borderColor: tabError
                  ? "rgba(239, 68, 68, 0.3)"
                  : isConnected
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
                    backgroundColor: tabError
                      ? "#ef4444"
                      : isConnected
                        ? "#10b981"
                        : "#3b82f6",
                  }}
                />
                <span className="fs-7 fw-semibold">
                  {tabError
                    ? `${t("Catalog sync error")}: ${tabError}`
                    : isConnected
                      ? t(
                          "Connected to Go/PostgreSQL Services & Ambulances (/v1/services, /v1/ambulances, /v1/packages, /v1/insurances)",
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
                  {currentTab === "packages"
                    ? `${t("Packages")}: ${packagesTotal}`
                    : currentTab === "insurances"
                      ? `${t("Insurances")}: ${insurancesTotal}`
                      : currentTab === "services"
                        ? `${t("Services")}: ${servicesTotal}`
                        : `${t("Ambulances")}: ${ambulances.length} | ${t("Calls")}: ${ambulanceCalls.length}`}
                </span>
              </div>
              <button
                type="button"
                className="btn-icon-link fs-7 d-flex align-items-center gap-1"
                onClick={() => loadServicesData(page, pageSize, search)}
                disabled={isLoadingApi || tabLoading}
                title={t("Refresh data from Go API")}
              >
                <RefreshCw
                  size={13}
                  className={isLoadingApi || tabLoading ? "animate-spin" : ""}
                />
                <span>
                  {isLoadingApi || tabLoading
                    ? t("Syncing...")
                    : t("Sync Backend")}
                </span>
              </button>
            </div>
          );
        })()}

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

      {/* 1. INSURANCES TAB */}
      {currentTab === "insurances" && (
        <div className="billing-card">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={16} />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="d-flex gap-2">
              <button className="btn-icon-blue" title={t("Filter")}>
                <Filter size={18} />
              </button>
              <button
                className="btn-action-blue"
                onClick={() => setInsuranceMode("create")}
              >
                {t("New Insurance")}
              </button>
            </div>
          </div>

          {insurancesError && (
            <div
              className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded mb-3"
              style={{
                borderColor: "#ef4444",
                backgroundColor: "rgba(239, 68, 68, 0.1)",
                color: "#ef4444",
              }}
            >
              <div className="d-flex align-items-center gap-2">
                <AlertCircle size={16} />
                <span className="fs-7">{insurancesError}</span>
              </div>
              <button
                type="button"
                className="btn-action-blue btn-sm py-1 px-2 fs-8"
                onClick={handleRetry}
              >
                <RefreshCw size={12} className="me-1 d-inline" />
                {t("Retry")}
              </button>
            </div>
          )}

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("INSURANCE")} ↕</th>
                  <th>{t("SERVICE TAX")} ↕</th>
                  <th>{t("DISCOUNT")} ↕</th>
                  <th>{t("INSURANCE NO")} ↕</th>
                  <th>{t("INSURANCE CODE")} ↕</th>
                  <th>{t("HOSPITAL RATE")} ↕</th>
                  <th>{t("TOTAL")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {insurancesLoading ? (
                  <tr>
                    <td colSpan={9} className="text-center py-4">
                      <Loader2
                        size={18}
                        className="animate-spin d-inline me-2"
                      />
                      {t("Loading insurance policies...")}
                    </td>
                  </tr>
                ) : insurancesError ? (
                  <tr>
                    <td colSpan={9} className="text-center py-4 text-danger">
                      <AlertCircle size={18} className="d-inline me-2" />
                      <span>{insurancesError}</span>
                      <button
                        type="button"
                        className="btn btn-sm btn-outline-danger ms-3"
                        onClick={handleRetry}
                      >
                        {t("Retry")}
                      </button>
                    </td>
                  </tr>
                ) : insurances.length === 0 ? (
                  <tr>
                    <td colSpan={9} className="text-center py-4 text-muted">
                      {t("No insurance policies found in catalog")}
                    </td>
                  </tr>
                ) : (
                  insurances.map((item) => (
                    <tr key={item.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {item.name}
                        </span>
                      </td>
                      <td>
                        ETB{" "}
                        {item.serviceTax.toLocaleString(undefined, {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </td>
                      <td>{item.discount}%</td>
                      <td>
                        <span className="badge-blue-pill">
                          {item.insuranceNo}
                        </span>
                      </td>
                      <td>{item.insuranceCode}</td>
                      <td>
                        ETB{" "}
                        {item.hospitalRate.toLocaleString(undefined, {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </td>
                      <td>
                        ETB{" "}
                        {item.totalAmount.toLocaleString(undefined, {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </td>
                      <td>
                        <label
                          className={`switch-toggle ${pendingStatusIds.has(item.id) ? "opacity-50 pointer-events-none" : ""}`}
                          onClick={(e) => {
                            e.preventDefault();
                            if (!pendingStatusIds.has(item.id)) {
                              handleToggleInsuranceStatus(item);
                            }
                          }}
                        >
                          <input
                            type="checkbox"
                            checked={item.status}
                            disabled={pendingStatusIds.has(item.id)}
                            readOnly
                          />
                          <span className="slider-toggle"></span>
                        </label>
                      </td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            type="button"
                            className="btn-icon-blue-link"
                            title={t("View")}
                            onClick={() => openInsuranceDetails(item)}
                          >
                            <Eye size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-icon-blue-link"
                            title={t("Edit")}
                            onClick={() => startEditInsurance(item)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-icon-danger"
                            title={t("Delete")}
                            disabled={isSubmitting}
                            onClick={() => handleDeleteInsurance(item.id)}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {renderPagination(insurancesTotal, insurances.length)}
        </div>
      )}

      {/* 2. PACKAGES TAB */}
      {currentTab === "packages" && (
        <div className="billing-card">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={16} />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="d-flex gap-2">
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => {
                  resetPackageForm();
                  setPackageModal(true);
                }}
              >
                {t("New Package")}
              </button>
            </div>
          </div>

          {packagesError && (
            <div
              className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded mb-3"
              style={{
                borderColor: "#ef4444",
                backgroundColor: "rgba(239, 68, 68, 0.1)",
                color: "#ef4444",
              }}
            >
              <div className="d-flex align-items-center gap-2">
                <AlertCircle size={16} />
                <span className="fs-7">{packagesError}</span>
              </div>
              <button
                type="button"
                className="btn-action-blue btn-sm py-1 px-2 fs-8"
                onClick={handleRetry}
              >
                <RefreshCw size={12} className="me-1 d-inline" />
                {t("Retry")}
              </button>
            </div>
          )}

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("PACKAGE NAME")} ↕</th>
                  <th>{t("DISCOUNT")} ↕</th>
                  <th>{t("TOTAL AMOUNT")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {packagesLoading ? (
                  <tr>
                    <td colSpan={4} className="text-center py-4">
                      <Loader2
                        size={18}
                        className="animate-spin d-inline me-2"
                      />
                      {t("Loading medical packages...")}
                    </td>
                  </tr>
                ) : packagesError ? (
                  <tr>
                    <td colSpan={4} className="text-center py-4 text-danger">
                      <AlertCircle size={18} className="d-inline me-2" />
                      <span>{packagesError}</span>
                      <button
                        type="button"
                        className="btn btn-sm btn-outline-danger ms-3"
                        onClick={handleRetry}
                      >
                        {t("Retry")}
                      </button>
                    </td>
                  </tr>
                ) : packages.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="text-center py-4 text-muted">
                      {t("No medical packages found in catalog")}
                    </td>
                  </tr>
                ) : (
                  packages.map((pkg) => (
                    <tr key={pkg.id}>
                      <td>
                        <div>
                          <span className="fw-semibold text-primary">
                            {pkg.name}
                          </span>
                          {pkg.description && (
                            <div className="fs-8 text-muted mt-0.5">
                              {pkg.description}
                            </div>
                          )}
                        </div>
                      </td>
                      <td>{pkg.discount}%</td>
                      <td>
                        ETB{" "}
                        {pkg.totalAmount.toLocaleString(undefined, {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            type="button"
                            className="btn-icon-blue-link"
                            title={t("View")}
                            onClick={() => openPackageDetail(pkg)}
                          >
                            <Eye size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-icon-blue-link"
                            title={t("Edit")}
                            onClick={() => openPackageEdit(pkg)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-icon-danger"
                            title={t("Delete")}
                            onClick={() => handleDeletePackage(pkg.id)}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {renderPagination(packagesTotal, packages.length)}
        </div>
      )}

      {/* 3. SERVICES TAB */}
      {currentTab === "services" && (
        <div className="billing-card">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={16} />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="d-flex gap-2">
              <button
                className="btn-action-blue"
                onClick={() => setServiceModal(true)}
              >
                {t("New Service")}
              </button>
            </div>
          </div>

          {servicesError && (
            <div
              className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded mb-3"
              style={{
                borderColor: "#ef4444",
                backgroundColor: "rgba(239, 68, 68, 0.1)",
                color: "#ef4444",
              }}
            >
              <div className="d-flex align-items-center gap-2">
                <AlertCircle size={16} />
                <span className="fs-7">{servicesError}</span>
              </div>
              <button
                type="button"
                className="btn-action-blue btn-sm py-1 px-2 fs-8"
                onClick={handleRetry}
              >
                <RefreshCw size={12} className="me-1 d-inline" />
                {t("Retry")}
              </button>
            </div>
          )}

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("SERVICE NAME")} ↕</th>
                  <th>{t("QUANTITY")} ↕</th>
                  <th>{t("RATE")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {servicesLoading ? (
                  <tr>
                    <td colSpan={5} className="text-center py-4">
                      <Loader2
                        size={18}
                        className="animate-spin d-inline me-2"
                      />
                      {t("Loading services...")}
                    </td>
                  </tr>
                ) : servicesError ? (
                  <tr>
                    <td colSpan={5} className="text-center py-4 text-danger">
                      <AlertCircle size={18} className="d-inline me-2" />
                      <span>{servicesError}</span>
                      <button
                        type="button"
                        className="btn btn-sm btn-outline-danger ms-3"
                        onClick={handleRetry}
                      >
                        {t("Retry")}
                      </button>
                    </td>
                  </tr>
                ) : services.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="text-center py-4 text-muted">
                      {t("No services found in catalog")}
                    </td>
                  </tr>
                ) : (
                  services.map((srv) => (
                    <tr key={srv.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {srv.name}
                        </span>
                      </td>
                      <td>{srv.quantity}</td>
                      <td>${srv.rate.toLocaleString()}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={srv.status}
                            onChange={() =>
                              setServices((prev) =>
                                prev.map((x) =>
                                  x.id === srv.id
                                    ? { ...x, status: !x.status }
                                    : x,
                                ),
                              )
                            }
                          />
                          <span className="slider-toggle"></span>
                        </label>
                      </td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            className="btn-icon-blue-link"
                            title={t("Edit")}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="btn-icon-danger"
                            title={t("Delete")}
                            onClick={() =>
                              setServices((prev) =>
                                prev.filter((x) => x.id !== srv.id),
                              )
                            }
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {renderPagination(servicesTotal, services.length)}
        </div>
      )}

      {/* 4. AMBULANCES TAB */}
      {currentTab === "ambulances" && (
        <div className="billing-card">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={16} />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="d-flex gap-2">
              <button
                className="btn-action-blue"
                onClick={() => setAmbulanceModal(true)}
              >
                {t("New Ambulance")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("VEHICLE NUMBER")} ↕</th>
                  <th>{t("VEHICLE MODEL")} ↕</th>
                  <th>{t("YEAR MADE")} ↕</th>
                  <th>{t("DRIVER NAME")} ↕</th>
                  <th>{t("DRIVER CONTACT")} ↕</th>
                  <th>{t("VEHICLE TYPE")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {ambulances
                  .filter((a) =>
                    (
                      a.vehicleNumber +
                      " " +
                      a.vehicleModel +
                      " " +
                      a.driverName
                    )
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((amb) => (
                    <tr key={amb.id}>
                      <td>
                        <span className="badge-blue-pill">
                          {amb.vehicleNumber}
                        </span>
                      </td>
                      <td>{amb.vehicleModel}</td>
                      <td>{amb.yearMade}</td>
                      <td>{amb.driverName}</td>
                      <td>{amb.driverContact}</td>
                      <td>
                        <span className="badge-green">{amb.vehicleType}</span>
                      </td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={amb.status}
                            onChange={() =>
                              setAmbulances((prev) =>
                                prev.map((x) =>
                                  x.id === amb.id
                                    ? { ...x, status: !x.status }
                                    : x,
                                ),
                              )
                            }
                          />
                          <span className="slider-toggle"></span>
                        </label>
                      </td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            className="btn-icon-blue-link"
                            title={t("Edit")}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="btn-icon-danger"
                            title={t("Delete")}
                            onClick={() =>
                              setAmbulances((prev) =>
                                prev.filter((x) => x.id !== amb.id),
                              )
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
        </div>
      )}

      {/* 5. AMBULANCE CALLS TAB */}
      {currentTab === "ambulance-calls" && (
        <div className="billing-card">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={16} />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="d-flex gap-2">
              <button
                className="btn-action-blue"
                onClick={() => setCallModal(true)}
              >
                {t("New Ambulance Call")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("PATIENT")} ↕</th>
                  <th>{t("VEHICLE MODEL")} ↕</th>
                  <th>{t("DRIVER NAME")} ↕</th>
                  <th>{t("DATE")} ↕</th>
                  <th>{t("AMOUNT")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {ambulanceCalls
                  .filter((c) =>
                    (c.patientName + " " + c.vehicleModel + " " + c.driverName)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((call) => (
                    <tr key={call.id}>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <div className="patient-avatar-circle">
                            {call.patientName
                              .split(" ")
                              .map((n) => n[0])
                              .join("")
                              .slice(0, 2)
                              .toUpperCase()}
                          </div>
                          <div>
                            <div className="fw-semibold text-primary">
                              {call.patientName}
                            </div>
                            <div className="text-secondary small">
                              {call.patientEmail}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>{call.vehicleModel}</td>
                      <td>{call.driverName}</td>
                      <td>
                        <span className="tx-date-badge">
                          <span>{call.time}</span>
                          <span>{call.date}</span>
                        </span>
                      </td>
                      <td>${call.amount.toLocaleString()}</td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            className="btn-icon-danger"
                            title={t("Delete")}
                            onClick={() =>
                              setAmbulanceCalls((prev) =>
                                prev.filter((x) => x.id !== call.id),
                              )
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
        </div>
      )}

      {/* MODAL: NEW PACKAGE */}
      {packageModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "680px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Package")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setPackageModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreatePackage}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Package Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Package Name")}
                    value={pkgName}
                    onChange={(e) => setPkgName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={2}
                    className="form-control-custom w-100"
                    placeholder={t("Description")}
                    value={pkgDescription}
                    onChange={(e) => setPkgDescription(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Discount (%)")}:</label>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    placeholder="0"
                    value={pkgDiscount}
                    onChange={(e) => setPkgDiscount(e.target.value)}
                  />
                </div>

                {catalogServicesError && (
                  <div
                    className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded mb-3"
                    style={{
                      borderColor: "#ef4444",
                      backgroundColor: "rgba(239, 68, 68, 0.1)",
                      color: "#ef4444",
                    }}
                    data-testid="catalog-services-error"
                  >
                    <div className="d-flex align-items-center gap-2">
                      <AlertCircle size={16} />
                      <span className="fs-7">{catalogServicesError}</span>
                    </div>
                    <button
                      type="button"
                      className="btn-action-blue btn-sm py-1 px-2 fs-8"
                      onClick={() => loadActiveCatalogServices()}
                    >
                      <RefreshCw size={12} className="me-1 d-inline" />
                      {t("Retry")}
                    </button>
                  </div>
                )}
                {catalogServicesLoading && (
                  <div className="text-muted fs-8 mb-2 d-flex align-items-center gap-1">
                    <RefreshCw size={12} className="spinner-border-sm" />
                    <span>{t("Loading service catalog...")}</span>
                  </div>
                )}

                <div className="mt-3">
                  <div className="d-flex justify-content-between align-items-center mb-2">
                    <h4 className="fs-6 fw-bold mb-0">{t("Services")}</h4>
                    <div className="d-flex align-items-center gap-2">
                      <input
                        type="text"
                        className="form-control-custom"
                        style={{
                          maxWidth: "200px",
                          height: "30px",
                          fontSize: "12px",
                        }}
                        placeholder={t("Filter services...")}
                        value={modalSrvFilter}
                        onChange={(e) => setModalSrvFilter(e.target.value)}
                      />
                      <button
                        type="button"
                        className="btn-action-blue btn-sm py-1 px-2"
                        onClick={addPkgLine}
                      >
                        <Plus size={14} className="me-1" />
                        {t("Add Service")}
                      </button>
                    </div>
                  </div>

                  <div className="table-responsive">
                    <table className="billing-table w-100">
                      <thead>
                        <tr>
                          <th style={{ width: "40px" }}>#</th>
                          <th>
                            {t("Service")}{" "}
                            <span className="text-danger">*</span>
                          </th>
                          <th style={{ width: "100px" }}>
                            {t("Quantity")}{" "}
                            <span className="text-danger">*</span>
                          </th>
                          <th style={{ width: "130px" }}>
                            {t("Rate")} (ETB){" "}
                            <span className="text-danger">*</span>
                          </th>
                          <th style={{ width: "120px" }}>
                            {t("Amount")} (ETB)
                          </th>
                          <th style={{ width: "60px" }}></th>
                        </tr>
                      </thead>
                      <tbody>
                        {pkgLines.map((row, idx) => {
                          const lineAmount =
                            Math.max(0, parseInt(row.quantity, 10) || 0) *
                            Math.max(0, parseFloat(row.rate) || 0);
                          return (
                            <tr key={row.id}>
                              <td>{idx + 1}</td>
                              <td>
                                <select
                                  required
                                  className="form-control-custom w-100"
                                  value={row.serviceId}
                                  onChange={(e) =>
                                    updatePkgLine(
                                      idx,
                                      "serviceId",
                                      e.target.value,
                                    )
                                  }
                                >
                                  <option value="">
                                    -- {t("Select Service")} --
                                  </option>
                                  {(catalogServices.length > 0
                                    ? catalogServices
                                    : services
                                  )
                                    .filter(
                                      (s) =>
                                        !modalSrvFilter.trim() ||
                                        s.name
                                          .toLowerCase()
                                          .includes(
                                            modalSrvFilter.toLowerCase(),
                                          ) ||
                                        s.id === row.serviceId,
                                    )
                                    .map((s) => (
                                      <option key={s.id} value={s.id}>
                                        {s.name} ({s.rate} ETB)
                                      </option>
                                    ))}
                                </select>
                              </td>
                              <td>
                                <input
                                  type="number"
                                  min="1"
                                  required
                                  className="form-control-custom w-100"
                                  value={row.quantity}
                                  onChange={(e) =>
                                    updatePkgLine(
                                      idx,
                                      "quantity",
                                      e.target.value,
                                    )
                                  }
                                />
                              </td>
                              <td>
                                <input
                                  type="number"
                                  step="0.01"
                                  min="0"
                                  required
                                  className="form-control-custom w-100"
                                  value={row.rate}
                                  onChange={(e) =>
                                    updatePkgLine(idx, "rate", e.target.value)
                                  }
                                />
                              </td>
                              <td>ETB {lineAmount.toFixed(2)}</td>
                              <td>
                                <button
                                  type="button"
                                  className="btn-icon-danger"
                                  disabled={pkgLines.length <= 1}
                                  onClick={() => removePkgLine(idx)}
                                >
                                  <Trash2 size={16} />
                                </button>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>

                  <div className="d-flex flex-column align-items-end mt-3 gap-1 fs-7">
                    <div>
                      {t("Subtotal")}: ETB {createPreview.subtotal.toFixed(2)}
                    </div>
                    <div>
                      {t("Discount (%)")}: -ETB{" "}
                      {createPreview.discountAmount.toFixed(2)} ({pkgDiscount}%)
                    </div>
                    <div className="fs-6 fw-bold text-primary mt-1">
                      {t("Estimated Total")}: ETB{" "}
                      {createPreview.total.toFixed(2)}
                    </div>
                  </div>
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
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
                <button
                  type="button"
                  className="btn-action-grey"
                  disabled={isSubmitting}
                  onClick={() => setPackageModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: EDIT PACKAGE */}
      {packageEditModal && selectedPackage && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "680px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("Edit Package")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setPackageEditModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleUpdatePackage}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Package Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Package Name")}
                    value={editPkgName}
                    onChange={(e) => setEditPkgName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={2}
                    className="form-control-custom w-100"
                    placeholder={t("Description")}
                    value={editPkgDescription}
                    onChange={(e) => setEditPkgDescription(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Discount (%)")}:</label>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    placeholder="0"
                    value={editPkgDiscount}
                    onChange={(e) => setEditPkgDiscount(e.target.value)}
                  />
                </div>

                {catalogServicesError && (
                  <div
                    className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded mb-3"
                    style={{
                      borderColor: "#ef4444",
                      backgroundColor: "rgba(239, 68, 68, 0.1)",
                      color: "#ef4444",
                    }}
                    data-testid="catalog-services-error"
                  >
                    <div className="d-flex align-items-center gap-2">
                      <AlertCircle size={16} />
                      <span className="fs-7">{catalogServicesError}</span>
                    </div>
                    <button
                      type="button"
                      className="btn-action-blue btn-sm py-1 px-2 fs-8"
                      onClick={() => loadActiveCatalogServices()}
                    >
                      <RefreshCw size={12} className="me-1 d-inline" />
                      {t("Retry")}
                    </button>
                  </div>
                )}
                {catalogServicesLoading && (
                  <div className="text-muted fs-8 mb-2 d-flex align-items-center gap-1">
                    <RefreshCw size={12} className="spinner-border-sm" />
                    <span>{t("Loading service catalog...")}</span>
                  </div>
                )}

                <div className="mt-3">
                  <div className="d-flex justify-content-between align-items-center mb-2">
                    <h4 className="fs-6 fw-bold mb-0">{t("Services")}</h4>
                    <div className="d-flex align-items-center gap-2">
                      <input
                        type="text"
                        className="form-control-custom"
                        style={{
                          maxWidth: "200px",
                          height: "30px",
                          fontSize: "12px",
                        }}
                        placeholder={t("Filter services...")}
                        value={modalSrvFilter}
                        onChange={(e) => setModalSrvFilter(e.target.value)}
                      />
                      <button
                        type="button"
                        className="btn-action-blue btn-sm py-1 px-2"
                        onClick={addEditPkgLine}
                      >
                        <Plus size={14} className="me-1" />
                        {t("Add Service")}
                      </button>
                    </div>
                  </div>

                  <div className="table-responsive">
                    <table className="billing-table w-100">
                      <thead>
                        <tr>
                          <th style={{ width: "40px" }}>#</th>
                          <th>
                            {t("Service")}{" "}
                            <span className="text-danger">*</span>
                          </th>
                          <th style={{ width: "100px" }}>
                            {t("Quantity")}{" "}
                            <span className="text-danger">*</span>
                          </th>
                          <th style={{ width: "130px" }}>
                            {t("Rate")} (ETB){" "}
                            <span className="text-danger">*</span>
                          </th>
                          <th style={{ width: "120px" }}>
                            {t("Amount")} (ETB)
                          </th>
                          <th style={{ width: "60px" }}></th>
                        </tr>
                      </thead>
                      <tbody>
                        {editPkgLines.map((row, idx) => {
                          const lineAmount =
                            Math.max(0, parseInt(row.quantity, 10) || 0) *
                            Math.max(0, parseFloat(row.rate) || 0);
                          return (
                            <tr key={row.id || idx}>
                              <td>{idx + 1}</td>
                              <td>
                                <select
                                  required
                                  className="form-control-custom w-100"
                                  value={row.serviceId}
                                  onChange={(e) =>
                                    updateEditPkgLine(
                                      idx,
                                      "serviceId",
                                      e.target.value,
                                    )
                                  }
                                >
                                  <option value="">
                                    -- {t("Select Service")} --
                                  </option>
                                  {/* Preserve existing selected archived reference for display without permitting new invalid selections */}
                                  {row.serviceId &&
                                    !catalogServices.some(
                                      (s) => s.id === row.serviceId,
                                    ) && (
                                      <option
                                        key={row.serviceId}
                                        value={row.serviceId}
                                      >
                                        {row.serviceName
                                          ? `${row.serviceName} (${t("Archived")})`
                                          : `${row.serviceId} (${t("Archived")})`}
                                      </option>
                                    )}
                                  {(catalogServices.length > 0
                                    ? catalogServices
                                    : services
                                  )
                                    .filter(
                                      (s) =>
                                        !modalSrvFilter.trim() ||
                                        s.name
                                          .toLowerCase()
                                          .includes(
                                            modalSrvFilter.toLowerCase(),
                                          ) ||
                                        s.id === row.serviceId,
                                    )
                                    .map((s) => (
                                      <option key={s.id} value={s.id}>
                                        {s.name} ({s.rate} ETB)
                                      </option>
                                    ))}
                                </select>
                              </td>
                              <td>
                                <input
                                  type="number"
                                  min="1"
                                  required
                                  className="form-control-custom w-100"
                                  value={row.quantity}
                                  onChange={(e) =>
                                    updateEditPkgLine(
                                      idx,
                                      "quantity",
                                      e.target.value,
                                    )
                                  }
                                />
                              </td>
                              <td>
                                <input
                                  type="number"
                                  step="0.01"
                                  min="0"
                                  required
                                  className="form-control-custom w-100"
                                  value={row.rate}
                                  onChange={(e) =>
                                    updateEditPkgLine(
                                      idx,
                                      "rate",
                                      e.target.value,
                                    )
                                  }
                                />
                              </td>
                              <td>ETB {lineAmount.toFixed(2)}</td>
                              <td>
                                <button
                                  type="button"
                                  className="btn-icon-danger"
                                  disabled={editPkgLines.length <= 1}
                                  onClick={() => removeEditPkgLine(idx)}
                                >
                                  <Trash2 size={16} />
                                </button>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>

                  <div className="d-flex flex-column align-items-end mt-3 gap-1 fs-7">
                    <div>
                      {t("Subtotal")}: ETB {editPreview.subtotal.toFixed(2)}
                    </div>
                    <div>
                      {t("Discount (%)")}: -ETB{" "}
                      {editPreview.discountAmount.toFixed(2)} ({editPkgDiscount}
                      %)
                    </div>
                    <div className="fs-6 fw-bold text-primary mt-1">
                      {t("Estimated Total")}: ETB {editPreview.total.toFixed(2)}
                    </div>
                  </div>
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
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
                <button
                  type="button"
                  className="btn-action-grey"
                  disabled={isSubmitting}
                  onClick={() => setPackageEditModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: PACKAGE DETAILS */}
      {packageDetailModal && selectedPackage && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "640px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("Package Details")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setPackageDetailModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <div className="mb-3">
                <div className="fs-5 fw-bold text-primary mb-1">
                  {selectedPackage.name}
                </div>
                {selectedPackage.description && (
                  <p className="text-muted fs-7 mb-2">
                    {selectedPackage.description}
                  </p>
                )}
                <div className="d-flex gap-4 fs-7 bg-light p-2 rounded">
                  <div>
                    <span className="text-muted">{t("Discount (%)")}: </span>
                    <span className="fw-semibold">
                      {selectedPackage.discount}%
                    </span>
                  </div>
                  <div>
                    <span className="text-muted">{t("Total Amount")}: </span>
                    <span className="fw-semibold text-primary">
                      ETB {selectedPackage.totalAmount.toFixed(2)}
                    </span>
                  </div>
                </div>
              </div>

              <h4 className="fs-6 fw-bold mb-2">{t("Services")}</h4>
              <div className="table-responsive">
                <table className="billing-table w-100">
                  <thead>
                    <tr>
                      <th style={{ width: "40px" }}>#</th>
                      <th>{t("Service")}</th>
                      <th>{t("Quantity")}</th>
                      <th>{t("Rate")} (ETB)</th>
                      <th>{t("Amount")} (ETB)</th>
                    </tr>
                  </thead>
                  <tbody>
                    {selectedPackage.services.map((s, idx) => (
                      <tr key={s.id || idx}>
                        <td>{idx + 1}</td>
                        <td className="fw-semibold">
                          {s.serviceName ||
                            services.find((srv) => srv.id === s.serviceId)
                              ?.name ||
                            s.serviceId}
                        </td>
                        <td>{s.quantity}</td>
                        <td>ETB {s.rate.toFixed(2)}</td>
                        <td className="fw-semibold">
                          ETB {s.amount.toFixed(2)}
                        </td>
                      </tr>
                    ))}
                    {selectedPackage.services.length === 0 && (
                      <tr>
                        <td colSpan={5} className="text-center py-2 text-muted">
                          {t("No services in this package")}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="modal-footer-custom d-flex justify-content-end">
              <button
                type="button"
                className="btn-action-grey"
                onClick={() => setPackageDetailModal(false)}
              >
                {t("Cancel")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL: EDIT INSURANCE */}
      {insuranceEditModal && (
        <div
          className="modal-backdrop-custom"
          style={{ overflowY: "auto", padding: "16px" }}
        >
          <div
            className="modal-card-custom"
            style={{
              maxWidth: "780px",
              maxHeight: "90vh",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div className="modal-header-custom d-flex justify-content-between align-items-center flex-shrink-0">
              <h3>{t("Edit Insurance")}</h3>
              <button
                type="button"
                className="btn-close-custom"
                onClick={() => setInsuranceEditModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={handleUpdateInsurance}
              style={{
                display: "flex",
                flexDirection: "column",
                overflow: "hidden",
                flex: 1,
              }}
            >
              <div
                className="modal-body-custom"
                style={{
                  overflowY: "auto",
                  flex: 1,
                  maxHeight: "calc(90vh - 140px)",
                }}
              >
                <div className="form-grid-2">
                  <div className="form-group-custom">
                    <label>
                      {t("Insurance")}: <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder={t("Insurance")}
                      value={editInsuranceName}
                      onChange={(e) => setEditInsuranceName(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>
                      {t("Service Tax")}: <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder={t("Service Tax")}
                      value={editServiceTax}
                      onChange={(e) => setEditServiceTax(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>{t("Discount: (In Percentage(%))")}</label>
                    <input
                      type="number"
                      min="0"
                      max="100"
                      value={editDiscount}
                      onChange={(e) => setEditDiscount(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>
                      {t("Insurance No")}:{" "}
                      <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder={t("Insurance No")}
                      value={editInsuranceNo}
                      onChange={(e) => setEditInsuranceNo(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>
                      {t("Insurance Code")}:{" "}
                      <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder={t("Insurance Code")}
                      value={editInsuranceCode}
                      onChange={(e) => setEditInsuranceCode(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>
                      {t("Hospital Rate")}:{" "}
                      <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder={t("Hospital Rate")}
                      value={editHospitalRate}
                      onChange={(e) => setEditHospitalRate(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>{t("Remark")}:</label>
                    <textarea
                      rows={3}
                      placeholder={t("Remark")}
                      value={editRemark}
                      onChange={(e) => setEditRemark(e.target.value)}
                    />
                  </div>

                  <div className="form-group-custom">
                    <label>{t("Status")}:</label>
                    <div className="mt-2">
                      <label className="switch-toggle">
                        <input
                          type="checkbox"
                          checked={editStatus}
                          onChange={(e) => setEditStatus(e.target.checked)}
                        />
                        <span className="slider-toggle"></span>
                      </label>
                    </div>
                  </div>
                </div>

                {/* Disease Details Table Section */}
                <div className="mt-4 pt-3 border-top border-secondary-subtle">
                  <div className="d-flex justify-content-between align-items-center mb-3">
                    <h4 className="m-0 fs-5 fw-semibold">
                      {t("Disease Details")}
                    </h4>
                    <button
                      type="button"
                      className="btn-action-blue px-3 py-1 fs-6"
                      onClick={addEditDiseaseRow}
                    >
                      {t("Add")}
                    </button>
                  </div>

                  <table className="billing-table w-100">
                    <thead>
                      <tr>
                        <th style={{ width: "60px" }}>#</th>
                        <th>
                          {t("DISEASES NAME")}{" "}
                          <span className="text-danger">*</span>
                        </th>
                        <th>
                          {t("DISEASES CHARGE")}{" "}
                          <span className="text-danger">*</span>
                        </th>
                        <th style={{ width: "80px" }}>{t("ACTION")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {editDiseases.map((row, idx) => (
                        <tr key={row.id || idx}>
                          <td>{idx + 1}</td>
                          <td>
                            <input
                              type="text"
                              required
                              className="form-control-custom w-100"
                              placeholder={t("Diseases Name")}
                              value={row.name}
                              onChange={(e) =>
                                updateEditDisease(idx, "name", e.target.value)
                              }
                            />
                          </td>
                          <td>
                            <input
                              type="number"
                              step="0.01"
                              required
                              className="form-control-custom w-100"
                              placeholder={t("Diseases charge")}
                              value={row.charge}
                              onChange={(e) =>
                                updateEditDisease(idx, "charge", e.target.value)
                              }
                            />
                          </td>
                          <td className="text-center">
                            <button
                              type="button"
                              className="btn-icon-danger"
                              disabled={editDiseases.length <= 1}
                              onClick={() => removeEditDiseaseRow(idx)}
                            >
                              <Trash2 size={16} />
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>

                  <div className="d-flex flex-column align-items-end mt-3 gap-1 fs-7">
                    <div>
                      {t("Service Tax")}: ETB{" "}
                      {(parseFloat(editServiceTax) || 0).toFixed(2)}
                    </div>
                    <div>
                      {t("Hospital Rate")}: ETB{" "}
                      {(parseFloat(editHospitalRate) || 0).toFixed(2)}
                    </div>
                    <div>
                      {t("Disease Details")}: ETB{" "}
                      {editCalculatedPreview.diseaseSum.toFixed(2)}
                    </div>
                    <div>
                      {t("Subtotal")}: ETB{" "}
                      {editCalculatedPreview.base.toFixed(2)}
                    </div>
                    <div>
                      {t("Discount (%)")}: -ETB{" "}
                      {editCalculatedPreview.discountAmount.toFixed(2)} (
                      {editDiscount}%)
                    </div>
                    <div className="fs-6 fw-bold text-primary mt-1">
                      {t("Estimated Total")}: ETB{" "}
                      {editCalculatedPreview.total.toFixed(2)}
                    </div>
                  </div>
                </div>
              </div>

              <div className="modal-footer-custom d-flex justify-content-end gap-2 flex-shrink-0">
                <button
                  type="submit"
                  className="btn-action-blue"
                  disabled={isSubmitting}
                  onClick={(e) => {
                    e.preventDefault();
                    handleUpdateInsurance(e);
                  }}
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
                <button
                  type="button"
                  className="btn-action-grey"
                  disabled={isSubmitting}
                  onClick={() => {
                    setInsuranceEditModal(false);
                    setSelectedInsurance(null);
                  }}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: INSURANCE DETAILS */}
      {insuranceDetailModal && selectedInsurance && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "640px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("Insurance Details")}</h3>
              <button
                type="button"
                className="btn-close-custom"
                onClick={() => setInsuranceDetailModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <div className="mb-3">
                <div className="d-flex justify-content-between align-items-start mb-2">
                  <div>
                    <h4 className="fs-5 fw-bold text-primary m-0">
                      {selectedInsurance.name}
                    </h4>
                    <div className="fs-7 text-muted mt-1">
                      {t("Insurance No")}:{" "}
                      <span className="fw-semibold">
                        {selectedInsurance.insuranceNo}
                      </span>{" "}
                      | {t("Insurance Code")}:{" "}
                      <span className="fw-semibold">
                        {selectedInsurance.insuranceCode}
                      </span>
                    </div>
                  </div>
                  <span
                    className={
                      selectedInsurance.status
                        ? "badge-available-stock"
                        : "badge-blue-pill"
                    }
                  >
                    {selectedInsurance.status ? t("Active") : t("Inactive")}
                  </span>
                </div>

                {selectedInsurance.remark && (
                  <p className="text-muted fs-7 mb-2 bg-light p-2 rounded">
                    <span className="fw-semibold">{t("Remark")}: </span>
                    {selectedInsurance.remark}
                  </p>
                )}

                <div className="d-flex flex-wrap gap-3 fs-7 bg-light p-2 rounded">
                  <div>
                    <span className="text-muted">{t("Service Tax")}: </span>
                    <span className="fw-semibold">
                      ETB {selectedInsurance.serviceTax.toFixed(2)}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted">{t("Hospital Rate")}: </span>
                    <span className="fw-semibold">
                      ETB {selectedInsurance.hospitalRate.toFixed(2)}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted">{t("Discount (%)")}: </span>
                    <span className="fw-semibold">
                      {selectedInsurance.discount}%
                    </span>
                  </div>
                  <div>
                    <span className="text-muted">{t("Total Amount")}: </span>
                    <span className="fw-semibold text-primary">
                      ETB {selectedInsurance.totalAmount.toFixed(2)}
                    </span>
                  </div>
                </div>
              </div>

              <h4 className="fs-6 fw-bold mb-2">{t("Disease Details")}</h4>
              <div className="table-responsive">
                <table className="billing-table w-100">
                  <thead>
                    <tr>
                      <th style={{ width: "40px" }}>#</th>
                      <th>{t("DISEASES NAME")}</th>
                      <th>{t("DISEASES CHARGE")} (ETB)</th>
                    </tr>
                  </thead>
                  <tbody>
                    {selectedInsurance.diseases.map((d, idx) => (
                      <tr key={d.id || idx}>
                        <td>{idx + 1}</td>
                        <td className="fw-semibold">{d.name}</td>
                        <td>ETB {d.charge.toFixed(2)}</td>
                      </tr>
                    ))}
                    {selectedInsurance.diseases.length === 0 && (
                      <tr>
                        <td colSpan={3} className="text-center py-2 text-muted">
                          {t("No diseases registered for this policy")}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="modal-footer-custom d-flex justify-content-end">
              <button
                type="button"
                className="btn-action-grey"
                onClick={() => setInsuranceDetailModal(false)}
              >
                {t("Cancel")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL: NEW SERVICE */}
      {serviceModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "500px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Service")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setServiceModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreateService}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Service Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Service Name")}
                    value={srvName}
                    onChange={(e) => setSrvName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Quantity")}:</label>
                  <input
                    type="number"
                    min="1"
                    value={srvQuantity}
                    onChange={(e) => setSrvQuantity(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Rate")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0.00"
                    value={srvRate}
                    onChange={(e) => setSrvRate(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
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
                <button
                  type="button"
                  className="btn-action-grey"
                  disabled={isSubmitting}
                  onClick={() => setServiceModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW AMBULANCE */}
      {ambulanceModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "540px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Ambulance")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setAmbulanceModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreateAmbulance}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Vehicle Number")}:{" "}
                    <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Vehicle Number")}
                    value={ambNumber}
                    onChange={(e) => setAmbNumber(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Vehicle Model")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Vehicle Model")}
                    value={ambModel}
                    onChange={(e) => setAmbModel(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Year Made")}:</label>
                  <input
                    type="number"
                    value={ambYear}
                    onChange={(e) => setAmbYear(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Driver Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Driver Name")}
                    value={ambDriver}
                    onChange={(e) => setAmbDriver(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Driver Contact")}:</label>
                  <input
                    type="text"
                    placeholder="+251..."
                    value={ambContact}
                    onChange={(e) => setAmbContact(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Vehicle Type")}:</label>
                  <select
                    className="form-select-custom w-100"
                    value={ambType}
                    onChange={(e) => setAmbType(e.target.value)}
                  >
                    <option value="Owned">Owned</option>
                    <option value="Contractual">Contractual</option>
                  </select>
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
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
                <button
                  type="button"
                  className="btn-action-grey"
                  disabled={isSubmitting}
                  onClick={() => setAmbulanceModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW AMBULANCE CALL */}
      {callModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "500px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Ambulance Call")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setCallModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreateAmbulanceCall}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Patient Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Patient Name")}
                    value={callPatient}
                    onChange={(e) => setCallPatient(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Ambulance")}:</label>
                  <select
                    className="form-select-custom w-100"
                    value={callVehicle}
                    onChange={(e) => setCallVehicle(e.target.value)}
                  >
                    {ambulances.map((a) => (
                      <option key={a.id} value={a.vehicleModel}>
                        {a.vehicleModel} ({a.vehicleNumber})
                      </option>
                    ))}
                  </select>
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Amount")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0.00"
                    value={callAmount}
                    onChange={(e) => setCallAmount(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
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
                <button
                  type="button"
                  className="btn-action-grey"
                  disabled={isSubmitting}
                  onClick={() => setCallModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
