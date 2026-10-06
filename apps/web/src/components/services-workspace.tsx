"use client";

import { useLanguage } from "@/components/language";
import { useState, useEffect } from "react";
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
  id: string;
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
}

interface PackageItem {
  id: string;
  name: string;
  discount: number;
  totalAmount: number;
  status: boolean;
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

  // Insurances State
  const [insurances, setInsurances] = useState<InsuranceItem[]>([
    {
      id: "INS-001",
      name: "BAJAJ",
      serviceTax: 5,
      discount: 10,
      insuranceNo: "BJ-994821",
      insuranceCode: "BAJAJ-CORP",
      hospitalRate: 1500,
      remark: "Corporate insurance partner",
      status: true,
      diseases: [
        { id: "1", name: "Cardiology Consultation", charge: 250 },
        { id: "2", name: "ECG Scan", charge: 150 },
      ],
      totalAmount: 400,
    },
    {
      id: "INS-002",
      name: "Zelda Walls",
      serviceTax: 7.5,
      discount: 5,
      insuranceNo: "ZW-481923",
      insuranceCode: "ZELDA-HLTH",
      hospitalRate: 2200,
      remark: "Individual premium insurance",
      status: true,
      diseases: [{ id: "1", name: "Full Blood Panel", charge: 180 }],
      totalAmount: 180,
    },
    {
      id: "INS-003",
      name: "Brooke Leblan",
      serviceTax: 4,
      discount: 0,
      insuranceNo: "BL-102938",
      insuranceCode: "BROOKE-STD",
      hospitalRate: 1800,
      remark: "Standard coverage policy",
      status: false,
      diseases: [{ id: "1", name: "X-Ray Chest", charge: 220 }],
      totalAmount: 220,
    },
  ]);

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

  // Packages State
  const [packages, setPackages] = useState<PackageItem[]>([
    {
      id: "PKG-01",
      name: "Checkup",
      discount: 15,
      totalAmount: 350,
      status: true,
    },
    {
      id: "PKG-02",
      name: "All in 1",
      discount: 20,
      totalAmount: 750,
      status: true,
    },
    {
      id: "PKG-03",
      name: "Fever Package",
      discount: 10,
      totalAmount: 180,
      status: true,
    },
    {
      id: "PKG-04",
      name: "Daat Test",
      discount: 5,
      totalAmount: 120,
      status: true,
    },
    {
      id: "PKG-05",
      name: "dental",
      discount: 10,
      totalAmount: 200,
      status: true,
    },
  ]);
  const [packageModal, setPackageModal] = useState(false);
  const [pkgName, setPkgName] = useState("");
  const [pkgDiscount, setPkgDiscount] = useState("");
  const [pkgAmount, setPkgAmount] = useState("");

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

  // Dynamic Disease Details Helpers for New Insurance
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

  const calculatedTotal = newDiseases.reduce((acc, row) => {
    const val = parseFloat(row.charge);
    return acc + (isNaN(val) ? 0 : val);
  }, 0);

  /* -------------------------------------------------------------
     LIVE BACKEND API INTEGRATION (Go / PostgreSQL /v1/services & /v1/ambulances)
     ------------------------------------------------------------- */
  const [apiConnected, setApiConnected] = useState(false);
  const [isLoadingApi, setIsLoadingApi] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");

  async function loadServicesData() {
    setIsLoadingApi(true);
    let connected = false;
    try {
      const [srvRes, ambRes, callRes, pkgRes, insRes] = await Promise.all([
        fetch("/api/hms/services"),
        fetch("/api/hms/ambulances"),
        fetch("/api/hms/ambulance-calls"),
        fetch("/api/hms/packages"),
        fetch("/api/hms/insurances"),
      ]);

      if (srvRes.ok) {
        const srvData = await srvRes.json();
        if (Array.isArray(srvData.services) && srvData.services.length > 0) {
          setServices(
            srvData.services.map((s: any) => ({
              id: s.id,
              name: s.name,
              quantity: s.quantity || 1,
              rate: (s.rateMinor || 0) / 100,
              status: s.status === 1,
            })),
          );
        }
        connected = true;
      }

      if (ambRes.ok) {
        const ambData = await ambRes.json();
        if (
          Array.isArray(ambData.ambulances) &&
          ambData.ambulances.length > 0
        ) {
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
        connected = true;
      }

      if (callRes.ok) {
        const callData = await callRes.json();
        if (
          Array.isArray(callData.ambulance_calls) &&
          callData.ambulance_calls.length > 0
        ) {
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
        connected = true;
      }

      if (pkgRes.ok) {
        const pkgData = await pkgRes.json();
        if (Array.isArray(pkgData.packages) && pkgData.packages.length > 0) {
          setPackages(
            pkgData.packages.map((p: any) => ({
              id: p.id,
              name: p.name,
              discount: p.discount || 0,
              totalAmount: (p.totalAmountMinor || 0) / 100,
              status: p.status === 1,
            })),
          );
        }
        connected = true;
      }

      if (insRes.ok) {
        const insData = await insRes.json();
        if (
          Array.isArray(insData.insurances) &&
          insData.insurances.length > 0
        ) {
          setInsurances(
            insData.insurances.map((i: any) => ({
              id: i.id,
              name: i.name,
              serviceTax: i.serviceTax || 0,
              discount: i.discount || 0,
              insuranceNo: i.insuranceNo || "INS-001",
              insuranceCode: i.insuranceCode || "CODE",
              hospitalRate: (i.hospitalRateMinor || 0) / 100,
              remark: i.remark || "",
              status: i.status === 1,
              diseases: [],
              totalAmount: (i.totalAmountMinor || 0) / 100,
            })),
          );
        }
        connected = true;
      }

      setApiConnected(connected);
    } catch {
      setApiConnected(false);
    } finally {
      setIsLoadingApi(false);
    }
  }

  useEffect(() => {
    loadServicesData();
  }, []);

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
    if (!pkgName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");
    const amt = parseFloat(pkgAmount) || 0;
    const disc = parseFloat(pkgDiscount) || 0;

    try {
      const res = await fetch("/api/hms/packages", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: pkgName.trim(),
          discount: disc,
          totalAmountMinor: Math.round(amt * 100),
          status: 1,
        }),
      });
      if (res.ok) {
        setApiSuccessBanner(t("Medical package registered in catalog"));
        loadServicesData();
        setPackageModal(false);
        setPkgName("");
        setPkgDiscount("");
        setPkgAmount("");
        setIsSubmitting(false);
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }
    setPackages((prev) => [
      {
        id: `PKG-${prev.length + 1}`,
        name: pkgName.trim(),
        discount: disc,
        totalAmount: amt,
        status: true,
      },
      ...prev,
    ]);
    setPackageModal(false);
    setPkgName("");
    setPkgDiscount("");
    setPkgAmount("");
  }

  async function handleSaveInsurance(e: React.FormEvent) {
    e.preventDefault();
    if (!newInsuranceName) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const res = await fetch("/api/hms/insurances", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: newInsuranceName.trim(),
          serviceTax: parseFloat(newServiceTax) || 0,
          discount: parseFloat(newDiscount) || 0,
          insuranceNo: newInsuranceNo.trim(),
          insuranceCode: newInsuranceCode.trim(),
          hospitalRateMinor: Math.round(
            (parseFloat(newHospitalRate) || 0) * 100,
          ),
          remark: newRemark.trim(),
          status: newStatus ? 1 : 0,
        }),
      });
      if (res.ok) {
        setApiSuccessBanner(t("Insurance policy successfully registered"));
        loadServicesData();
        setInsuranceMode("list");
        setIsSubmitting(false);
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }

    const newItem: InsuranceItem = {
      id: `INS-${String(insurances.length + 1).padStart(3, "0")}`,
      name: newInsuranceName,
      serviceTax: parseFloat(newServiceTax) || 0,
      discount: parseFloat(newDiscount) || 0,
      insuranceNo: newInsuranceNo,
      insuranceCode: newInsuranceCode,
      hospitalRate: parseFloat(newHospitalRate) || 0,
      remark: newRemark,
      status: newStatus,
      diseases: newDiseases.map((d, i) => ({
        id: String(i + 1),
        name: d.name || "Service",
        charge: parseFloat(d.charge) || 0,
      })),
      totalAmount: calculatedTotal,
    };
    setInsurances([newItem, ...insurances]);
    setInsuranceMode("list");
    setNewInsuranceName("");
    setNewServiceTax("");
    setNewDiscount("0");
    setNewInsuranceNo("");
    setNewInsuranceCode("");
    setNewHospitalRate("");
    setNewRemark("");
    setNewStatus(true);
    setNewDiseases([{ id: "1", name: "", charge: "" }]);
    setIsSubmitting(false);
  }

  /* -------------------------------------------------------------
     RENDER: NEW INSURANCE FULL-PAGE VIEW (MATCHING SCREENSHOT 185115)
     ------------------------------------------------------------- */
  if (currentTab === "insurances" && insuranceMode === "create") {
    return (
      <div className="legacy-workspace">
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
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2 className="workspace-heading m-0">{t("New Insurance")}</h2>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => setInsuranceMode("list")}
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

              <div className="d-flex justify-content-end align-items-center mt-3 fs-5 fw-bold">
                <span>
                  {t("Total Amount")}: ${calculatedTotal.toFixed(2)}
                </span>
              </div>
            </div>

            <div className="d-flex justify-content-end gap-2 mt-4">
              <button type="submit" className="btn-action-blue px-4 py-2">
                {t("Save")}
              </button>
              <button
                type="button"
                className="btn-action-grey px-4 py-2"
                onClick={() => setInsuranceMode("list")}
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
    <div className="legacy-workspace">
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
              {t("Services")}: {services.length} | {t("Ambulances")}:{" "}
              {ambulances.length} | {t("Calls")}: {ambulanceCalls.length} |{" "}
              {t("Packages")}: {packages.length} | {t("Insurances")}:{" "}
              {insurances.length}
            </span>
          </div>
          <button
            type="button"
            className="btn-icon-link fs-7 d-flex align-items-center gap-1"
            onClick={loadServicesData}
            disabled={isLoadingApi}
            title={t("Refresh services & ambulances from Go API")}
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
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {insurances
                  .filter((item) =>
                    (
                      item.name +
                      " " +
                      item.insuranceNo +
                      " " +
                      item.insuranceCode
                    )
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((item) => (
                    <tr key={item.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {item.name}
                        </span>
                      </td>
                      <td>{item.serviceTax}%</td>
                      <td>{item.discount}%</td>
                      <td>
                        <span className="badge-blue-pill">
                          {item.insuranceNo}
                        </span>
                      </td>
                      <td>{item.insuranceCode}</td>
                      <td>${item.hospitalRate.toLocaleString()}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={item.status}
                            onChange={() =>
                              setInsurances((prev) =>
                                prev.map((x) =>
                                  x.id === item.id
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
                              setInsurances((prev) =>
                                prev.filter((x) => x.id !== item.id),
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

          <div className="billing-pagination d-flex justify-content-between align-items-center mt-3">
            <div className="d-flex align-items-center gap-2">
              <span>{t("Show")}</span>
              <select
                className="form-select-custom"
                value={pageSize}
                onChange={(e) => setPageSize(Number(e.target.value))}
              >
                <option value={10}>10</option>
                <option value={25}>25</option>
                <option value={50}>50</option>
              </select>
              <span>
                {t("Showing")} 1 {t("to")} {insurances.length} {t("of")}{" "}
                {insurances.length} {t("Results")}
              </span>
            </div>
            <div className="pagination-numbers">
              <button className="page-btn active">1</button>
            </div>
          </div>
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
                className="btn-action-blue"
                onClick={() => setPackageModal(true)}
              >
                {t("New Package")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("PACKAGE NAME")} ↕</th>
                  <th>{t("DISCOUNT")} ↕</th>
                  <th>{t("TOTAL AMOUNT")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {packages
                  .filter((p) =>
                    p.name.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((pkg) => (
                    <tr key={pkg.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {pkg.name}
                        </span>
                      </td>
                      <td>{pkg.discount}%</td>
                      <td>${pkg.totalAmount.toLocaleString()}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={pkg.status}
                            onChange={() =>
                              setPackages((prev) =>
                                prev.map((x) =>
                                  x.id === pkg.id
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
                              setPackages((prev) =>
                                prev.filter((x) => x.id !== pkg.id),
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
                {services
                  .filter((s) =>
                    s.name.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((srv) => (
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
                  ))}
              </tbody>
            </table>
          </div>
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
          <div className="modal-card-custom" style={{ maxWidth: "500px" }}>
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
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Total Amount")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0.00"
                    value={pkgAmount}
                    onChange={(e) => setPkgAmount(e.target.value)}
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
                  onClick={() => setPackageModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
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
