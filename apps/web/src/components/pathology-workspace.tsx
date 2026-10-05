"use client";

import { useLanguage } from "@/components/language";
import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  Search,
  Edit2,
  Trash2,
  X,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from "lucide-react";

export type PathologyWorkspaceProps = {
  id?: string;
};

interface ParameterItem {
  id: string;
  name: string;
  referenceRange: string;
  unit: string;
  description?: string;
}

interface CategoryItem {
  id: string;
  name: string;
  description: string;
}

interface UnitItem {
  id: string;
  name: string;
  description: string;
}

interface TestItem {
  id: string;
  testName: string;
  shortName: string;
  testType: string;
  category: string;
  unit: string;
  subCategory: string;
  method: string;
  reportDays: number;
  charge: number;
}

export function PathologyWorkspace({
  id = "pathology-parameters",
}: PathologyWorkspaceProps) {
  const { t } = useLanguage();
  const currentTab = id;

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);

  // Subtabs
  const tabs = [
    {
      id: "pathology-categories",
      label: "Pathology Categories",
      href: "/modules/pathology-categories",
    },
    {
      id: "pathology-units",
      label: "Pathology Units",
      href: "/modules/pathology-units",
    },
    {
      id: "pathology-parameters",
      label: "Pathology Parameters",
      href: "/modules/pathology-parameters",
    },
    {
      id: "pathology-tests",
      label: "Pathology Tests",
      href: "/modules/pathology-tests",
    },
  ];

  // Parameters State (Matching Screenshot 184427)
  const [parameters, setParameters] = useState<ParameterItem[]>([
    { id: "1", name: "Ahmad", referenceRange: "2233", unit: "HRC" },
    { id: "2", name: "blood sugar", referenceRange: "345", unit: "eco" },
    { id: "3", name: "RBC Count", referenceRange: "300", unit: "ecg" },
    { id: "4", name: "ECG", referenceRange: "66", unit: "g/dl" },
    { id: "5", name: "URIC ACID 30", referenceRange: "15-20", unit: "mg" },
    { id: "6", name: "WBC Count", referenceRange: "Ddd", unit: "mill/cumm" },
    {
      id: "7",
      name: "lipid",
      referenceRange: "Male: 1-10 Female: 20-30",
      unit: "mg %",
    },
    { id: "8", name: "tyyftuy", referenceRange: "gtugu", unit: "mill/cumm" },
    { id: "9", name: "Glucose mo", referenceRange: "70-180", unit: "%" },
    { id: "10", name: "ebc", referenceRange: "2000", unit: "mill/cumm" },
    {
      id: "11",
      name: "Platelet Count",
      referenceRange: "150-450",
      unit: "10^3/uL",
    },
    { id: "12", name: "Hemoglobin", referenceRange: "13.5-17.5", unit: "g/dl" },
  ]);

  // Modal: New Pathology Parameter (Screenshot 184453)
  const [parameterModal, setParameterModal] = useState(false);
  const [paramName, setParamName] = useState("");
  const [paramRange, setParamRange] = useState("");
  const [paramUnit, setParamUnit] = useState("");
  const [paramDesc, setParamDesc] = useState("");

  // Categories State
  const [categories, setCategories] = useState<CategoryItem[]>([
    {
      id: "CAT-1",
      name: "Hematology",
      description: "Blood cell and plasma tests",
    },
    {
      id: "CAT-2",
      name: "Biochemistry",
      description: "Enzyme and hormone panels",
    },
    {
      id: "CAT-3",
      name: "Microbiology",
      description: "Culture and sensitivity analysis",
    },
    {
      id: "CAT-4",
      name: "Immunology",
      description: "Antibody and viral screens",
    },
  ]);
  const [categoryModal, setCategoryModal] = useState(false);
  const [catName, setCatName] = useState("");
  const [catDesc, setCatDesc] = useState("");

  // Units State
  const [units, setUnits] = useState<UnitItem[]>([
    { id: "U-1", name: "mg/dl", description: "Milligrams per deciliter" },
    { id: "U-2", name: "g/dl", description: "Grams per deciliter" },
    {
      id: "U-3",
      name: "mill/cumm",
      description: "Million per cubic millimeter",
    },
    { id: "U-4", name: "mg %", description: "Milligram percentage" },
    { id: "U-5", name: "%", description: "Percentage concentration" },
    { id: "U-6", name: "HRC", description: "Hemoglobin recovery coefficient" },
  ]);
  const [unitModal, setUnitModal] = useState(false);
  const [unitName, setUnitName] = useState("");
  const [unitDesc, setUnitDesc] = useState("");

  // Tests State
  const [tests, setTests] = useState<TestItem[]>([
    {
      id: "TEST-1",
      testName: "Complete Blood Count (CBC)",
      shortName: "CBC",
      testType: "Automated",
      category: "Hematology",
      unit: "mill/cumm",
      subCategory: "Cell Count",
      method: "Flow Cytometry",
      reportDays: 1,
      charge: 45,
    },
    {
      id: "TEST-2",
      testName: "Fasting Blood Glucose",
      shortName: "FBS",
      testType: "Enzymatic",
      category: "Biochemistry",
      unit: "mg/dl",
      subCategory: "Diabetes Panel",
      method: "Hexokinase",
      reportDays: 1,
      charge: 25,
    },
    {
      id: "TEST-3",
      testName: "Lipid Profile Full",
      shortName: "LIPID",
      testType: "Spectrophotometry",
      category: "Biochemistry",
      unit: "mg/dl",
      subCategory: "Cardiac Panel",
      method: "Colorimetric",
      reportDays: 1,
      charge: 60,
    },
  ]);
  const [testModal, setTestModal] = useState(false);
  const [tName, setTName] = useState("");
  const [tShort, setTShort] = useState("");
  const [tType, setTType] = useState("Automated");
  const [tCategory, setTCategory] = useState("Hematology");
  const [tUnit, setTUnit] = useState("mg/dl");
  const [tCharge, setTCharge] = useState("");

  const [apiConnected, setApiConnected] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");

  const loadPathologyData = useCallback(async () => {
    setIsSyncing(true);
    let connected = false;
    try {
      const [catRes, unitRes, testRes] = await Promise.all([
        fetch("/api/hms/diagnostic-categories?kind=pathology").catch(() => null),
        fetch("/api/hms/diagnostic-units").catch(() => null),
        fetch("/api/hms/diagnostic-tests").catch(() => null),
      ]);

      if (catRes && catRes.ok) {
        const catData = await catRes.json();
        if (Array.isArray(catData.categories) && catData.categories.length > 0) {
          setCategories(
            catData.categories.map((c: any) => ({
              id: c.id,
              name: c.name,
              description: c.description || "",
            })),
          );
        }
        connected = true;
      }

      if (unitRes && unitRes.ok) {
        const unitData = await unitRes.json();
        if (Array.isArray(unitData.units) && unitData.units.length > 0) {
          setUnits(
            unitData.units.map((u: any) => ({
              id: u.id,
              name: u.name,
              description: u.description || "",
            })),
          );
        }
        connected = true;
      }

      if (testRes && testRes.ok) {
        const testData = await testRes.json();
        if (Array.isArray(testData.tests) && testData.tests.length > 0) {
          setTests(
            testData.tests.map((t: any) => ({
              id: t.id,
              testName: t.name,
              shortName: t.shortName || t.name.slice(0, 4).toUpperCase(),
              testType: t.method || "Automated",
              category: t.category || "Hematology",
              unit: (t.parameters && t.parameters[0]?.unit) || "mg/dl",
              subCategory: "Diagnostic",
              method: t.method || "Standard",
              reportDays: t.reportDays || 1,
              charge: (t.chargeMinor || 0) / 100,
            })),
          );

          // Extract parameters from tests
          const remoteParams: ParameterItem[] = [];
          for (const t of testData.tests) {
            if (Array.isArray(t.parameters)) {
              for (const p of t.parameters) {
                remoteParams.push({
                  id: `${t.id}-${p.position || remoteParams.length + 1}`,
                  name: p.name,
                  referenceRange: p.referenceRange || "Standard",
                  unit: p.unit || "mg/dl",
                  description: `Parameter for ${t.name}`,
                });
              }
            }
          }
          if (remoteParams.length > 0) {
            setParameters((prev) => {
              const names = new Set(prev.map((x) => x.name.toLowerCase()));
              const newlyAdded = remoteParams.filter(
                (p) => !names.has(p.name.toLowerCase()),
              );
              return [...newlyAdded, ...prev];
            });
          }
        }
        connected = true;
      }
    } catch {
      // dual offline preview fallback
    } finally {
      setIsSyncing(false);
      setApiConnected(connected);
    }
  }, []);

  useEffect(() => {
    loadPathologyData();
  }, [loadPathologyData]);

  // Handle Save Parameter
  async function handleSaveParameter(e: React.FormEvent) {
    e.preventDefault();
    if (!paramName.trim() || !paramRange.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const res = await fetch("/api/hms/diagnostic-tests", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          kind: "pathology",
          name: paramName.trim(),
          shortName: paramName.trim().slice(0, 4).toUpperCase(),
          category: categories[0]?.name || "Hematology",
          method: "Automated",
          reportDays: 1,
          chargeMinor: 0,
          parameters: [
            {
              position: 1,
              name: paramName.trim(),
              unit: paramUnit || "mg/dl",
              referenceRange: paramRange.trim(),
              valueType: "text",
            },
          ],
        }),
      });
      if (res.ok) {
        setApiSuccessBanner(
          t(`Pathology parameter "${paramName}" registered in PostgreSQL`),
        );
        setApiConnected(true);
      }
    } catch {
      // Dual fallback
    } finally {
      const newParam: ParameterItem = {
        id: String(parameters.length + 1),
        name: paramName.trim(),
        referenceRange: paramRange.trim(),
        unit: paramUnit || "mg/dl",
        description: paramDesc,
      };
      setParameters((prev) => [newParam, ...prev]);
      setIsSubmitting(false);
      setParameterModal(false);
      setParamName("");
      setParamRange("");
      setParamUnit("");
      setParamDesc("");
    }
  }

  // Handle Save Category
  async function handleSaveCategory(e: React.FormEvent) {
    e.preventDefault();
    if (!catName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const res = await fetch("/api/hms/diagnostic-categories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: catName.trim(),
          kind: "pathology",
          description: catDesc.trim(),
        }),
      });
      if (res.ok) {
        const saved = await res.json();
        setCategories((prev) => [
          {
            id: saved.id || `CAT-${prev.length + 1}`,
            name: saved.name || catName.trim(),
            description: saved.description || catDesc.trim(),
          },
          ...prev,
        ]);
        setApiSuccessBanner(
          t(`Pathology category "${catName}" registered in PostgreSQL`),
        );
        setApiConnected(true);
        setCategoryModal(false);
        setCatName("");
        setCatDesc("");
        setIsSubmitting(false);
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }

    setCategories((prev) => [
      {
        id: `CAT-${prev.length + 1}`,
        name: catName.trim(),
        description: catDesc.trim(),
      },
      ...prev,
    ]);
    setCategoryModal(false);
    setCatName("");
    setCatDesc("");
  }

  // Handle Save Unit
  async function handleSaveUnit(e: React.FormEvent) {
    e.preventDefault();
    if (!unitName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");

    try {
      const res = await fetch("/api/hms/diagnostic-units", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: unitName.trim(),
          description: unitDesc.trim(),
        }),
      });
      if (res.ok) {
        const saved = await res.json();
        setUnits((prev) => [
          {
            id: saved.id || `U-${prev.length + 1}`,
            name: saved.name || unitName.trim(),
            description: saved.description || unitDesc.trim(),
          },
          ...prev,
        ]);
        setApiSuccessBanner(
          t(`Pathology unit "${unitName}" registered in PostgreSQL`),
        );
        setApiConnected(true);
        setUnitModal(false);
        setUnitName("");
        setUnitDesc("");
        setIsSubmitting(false);
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }

    setUnits((prev) => [
      {
        id: `U-${prev.length + 1}`,
        name: unitName.trim(),
        description: unitDesc.trim(),
      },
      ...prev,
    ]);
    setUnitModal(false);
    setUnitName("");
    setUnitDesc("");
  }

  // Handle Save Test
  async function handleSaveTest(e: React.FormEvent) {
    e.preventDefault();
    if (!tName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");
    setApiSuccessBanner("");
    const chargeVal = parseFloat(tCharge) || 0;
    const short = tShort.trim() || tName.slice(0, 4).toUpperCase();

    try {
      const res = await fetch("/api/hms/diagnostic-tests", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          kind: "pathology",
          name: tName.trim(),
          shortName: short,
          category: tCategory || "Hematology",
          method: tType || "Automated",
          reportDays: 1,
          chargeMinor: Math.round(chargeVal * 100),
          parameters: [
            {
              position: 1,
              name: `${tName.trim()} Test Parameter`,
              unit: tUnit || "mg/dl",
              referenceRange: "Standard",
              valueType: "text",
            },
          ],
        }),
      });
      if (res.ok) {
        const saved = await res.json();
        setTests((prev) => [
          {
            id: saved.id || `TEST-${prev.length + 1}`,
            testName: saved.name || tName.trim(),
            shortName: saved.shortName || short,
            testType: saved.method || tType || "Automated",
            category: saved.category || tCategory,
            unit: tUnit,
            subCategory: "Diagnostic",
            method: saved.method || tType || "Standard",
            reportDays: saved.reportDays || 1,
            charge: chargeVal,
          },
          ...prev,
        ]);
        setApiSuccessBanner(
          t(`Pathology test "${tName}" registered in PostgreSQL`),
        );
        setApiConnected(true);
        setTestModal(false);
        setTName("");
        setTShort("");
        setTCharge("");
        setIsSubmitting(false);
        return;
      }
    } catch {
      // Fallback
    } finally {
      setIsSubmitting(false);
    }

    setTests((prev) => [
      {
        id: `TEST-${prev.length + 1}`,
        testName: tName.trim(),
        shortName: short,
        testType: tType,
        category: tCategory,
        unit: tUnit,
        subCategory: "General",
        method: "Standard",
        reportDays: 1,
        charge: chargeVal,
      },
      ...prev,
    ]);
    setTestModal(false);
    setTName("");
    setTShort("");
    setTCharge("");
  }

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
                    "Connected to Go/PostgreSQL Pathology & Diagnostic Catalog (/v1/diagnostic-categories, /v1/diagnostic-units, /v1/diagnostic-tests)",
                  )
                : t("Local preview mode · Syncing locally")}
            </span>
            <span
              className="badge-available-stock fs-8 py-0 px-2"
              style={{ fontSize: "11px" }}
            >
              {categories.length} {t("Categories")} · {units.length}{" "}
              {t("Units")} · {tests.length} {t("Tests")} · {parameters.length}{" "}
              {t("Parameters")}
            </span>
          </div>

          <div className="d-flex align-items-center gap-2">
            <button
              type="button"
              className="btn btn-sm btn-outline-primary d-flex align-items-center gap-1"
              style={{ fontSize: "12px", padding: "2px 8px" }}
              disabled={isSyncing}
              onClick={loadPathologyData}
            >
              <RefreshCw
                size={12}
                className={isSyncing ? "spinner-border spinner-border-sm" : ""}
              />
              {isSyncing ? t("Syncing...") : t("Sync Backend")}
            </button>
          </div>
        </div>

        {apiSuccessBanner && (
          <div
            className="d-flex align-items-center gap-2 p-2 px-3 rounded border text-success"
            style={{
              backgroundColor: "rgba(16, 185, 129, 0.1)",
              borderColor: "rgba(16, 185, 129, 0.3)",
              fontSize: "13px",
            }}
          >
            <CheckCircle2 size={16} />
            <span>{apiSuccessBanner}</span>
          </div>
        )}

        {apiErrorBanner && (
          <div
            className="d-flex align-items-center gap-2 p-2 px-3 rounded border text-danger"
            style={{
              backgroundColor: "rgba(239, 68, 68, 0.1)",
              borderColor: "rgba(239, 68, 68, 0.3)",
              fontSize: "13px",
            }}
          >
            <AlertCircle size={16} />
            <span>{apiErrorBanner}</span>
          </div>
        )}
      </div>

      {/* 1. PATHOLOGY PARAMETERS TAB (SCREENSHOT 184427) */}
      {currentTab === "pathology-parameters" && (
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
                onClick={() => setParameterModal(true)}
              >
                {t("New Pathology Parameter")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("NAME")} ↕</th>
                  <th>{t("REFERENCE RANGE")} ↕</th>
                  <th>{t("UNIT")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {parameters
                  .filter((p) =>
                    (p.name + " " + p.referenceRange + " " + p.unit)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((param) => (
                    <tr key={param.id}>
                      <td>
                        <span className="fw-normal">{param.name}</span>
                      </td>
                      <td>{param.referenceRange}</td>
                      <td>{param.unit}</td>
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
                              setParameters((prev) =>
                                prev.filter((x) => x.id !== param.id),
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
                {t("Showing")} 1 {t("to")}{" "}
                {Math.min(pageSize, parameters.length)} {t("of")}{" "}
                {parameters.length} {t("Results")}
              </span>
            </div>
            <div className="pagination-numbers">
              <button className="page-btn active">1</button>
              <button className="page-btn">2</button>
              <button className="page-btn">3</button>
              <button className="page-btn">4</button>
              <button className="page-btn">&gt;</button>
            </div>
          </div>
        </div>
      )}

      {/* 2. PATHOLOGY CATEGORIES TAB */}
      {currentTab === "pathology-categories" && (
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
                onClick={() => setCategoryModal(true)}
              >
                {t("New Pathology Category")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("CATEGORY NAME")} ↕</th>
                  <th>{t("DESCRIPTION")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {categories
                  .filter((c) =>
                    c.name.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((cat) => (
                    <tr key={cat.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {cat.name}
                        </span>
                      </td>
                      <td>{cat.description}</td>
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
                              setCategories((prev) =>
                                prev.filter((x) => x.id !== cat.id),
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

      {/* 3. PATHOLOGY UNITS TAB */}
      {currentTab === "pathology-units" && (
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
                onClick={() => setUnitModal(true)}
              >
                {t("New Pathology Unit")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("UNIT NAME")} ↕</th>
                  <th>{t("DESCRIPTION")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {units
                  .filter((u) =>
                    u.name.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((unit) => (
                    <tr key={unit.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {unit.name}
                        </span>
                      </td>
                      <td>{unit.description}</td>
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
                              setUnits((prev) =>
                                prev.filter((x) => x.id !== unit.id),
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

      {/* 4. PATHOLOGY TESTS TAB */}
      {currentTab === "pathology-tests" && (
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
                onClick={() => setTestModal(true)}
              >
                {t("New Pathology Test")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("TEST NAME")} ↕</th>
                  <th>{t("SHORT NAME")} ↕</th>
                  <th>{t("TEST TYPE")} ↕</th>
                  <th>{t("CATEGORY")} ↕</th>
                  <th>{t("UNIT")} ↕</th>
                  <th>{t("CHARGE")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {tests
                  .filter((t) =>
                    (t.testName + " " + t.shortName + " " + t.category)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((test) => (
                    <tr key={test.id}>
                      <td>
                        <span className="fw-semibold text-primary">
                          {test.testName}
                        </span>
                      </td>
                      <td>
                        <span className="badge-blue-pill">
                          {test.shortName}
                        </span>
                      </td>
                      <td>{test.testType}</td>
                      <td>{test.category}</td>
                      <td>{test.unit}</td>
                      <td>${test.charge.toLocaleString()}</td>
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
                              setTests((prev) =>
                                prev.filter((x) => x.id !== test.id),
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

      {/* MODAL: NEW PATHOLOGY PARAMETER (SCREENSHOT 184453) */}
      {parameterModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "480px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Pathology Parameter")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setParameterModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveParameter}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Name")}
                    value={paramName}
                    onChange={(e) => setParamName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Reference Range")}:{" "}
                    <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Reference Range")}
                    value={paramRange}
                    onChange={(e) => setParamRange(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Unit")}: <span className="text-danger">*</span>
                  </label>
                  <select
                    className="form-select-custom w-100"
                    required
                    value={paramUnit}
                    onChange={(e) => setParamUnit(e.target.value)}
                  >
                    <option value="">{t("Select Unit")}</option>
                    <option value="HRC">HRC</option>
                    <option value="eco">eco</option>
                    <option value="ecg">ecg</option>
                    <option value="g/dl">g/dl</option>
                    <option value="mg">mg</option>
                    <option value="mill/cumm">mill/cumm</option>
                    <option value="mg %">mg %</option>
                    <option value="%">%</option>
                    <option value="10^3/uL">10^3/uL</option>
                  </select>
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={4}
                    placeholder={t("Description")}
                    value={paramDesc}
                    onChange={(e) => setParamDesc(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
                <button
                  type="submit"
                  className="btn-action-blue d-flex align-items-center gap-1"
                  disabled={isSubmitting}
                >
                  {isSubmitting && (
                    <Loader2 size={14} className="spinner-border spinner-border-sm" />
                  )}
                  <span>{t("Save")}</span>
                </button>
                <button
                  type="button"
                  className="btn-action-grey"
                  onClick={() => setParameterModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW CATEGORY */}
      {categoryModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "480px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Pathology Category")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setCategoryModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveCategory}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Category Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Category Name")}
                    value={catName}
                    onChange={(e) => setCatName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={3}
                    placeholder={t("Description")}
                    value={catDesc}
                    onChange={(e) => setCatDesc(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
                <button
                  type="submit"
                  className="btn-action-blue d-flex align-items-center gap-1"
                  disabled={isSubmitting}
                >
                  {isSubmitting && (
                    <Loader2 size={14} className="spinner-border spinner-border-sm" />
                  )}
                  <span>{t("Save")}</span>
                </button>
                <button
                  type="button"
                  className="btn-action-grey"
                  onClick={() => setCategoryModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW UNIT */}
      {unitModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "480px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Pathology Unit")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setUnitModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveUnit}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Unit Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Unit Name")}
                    value={unitName}
                    onChange={(e) => setUnitName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={3}
                    placeholder={t("Description")}
                    value={unitDesc}
                    onChange={(e) => setUnitDesc(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
                <button
                  type="submit"
                  className="btn-action-blue d-flex align-items-center gap-1"
                  disabled={isSubmitting}
                >
                  {isSubmitting && (
                    <Loader2 size={14} className="spinner-border spinner-border-sm" />
                  )}
                  <span>{t("Save")}</span>
                </button>
                <button
                  type="button"
                  className="btn-action-grey"
                  onClick={() => setUnitModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW TEST */}
      {testModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "540px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Pathology Test")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setTestModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveTest}>
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Test Name")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Test Name")}
                    value={tName}
                    onChange={(e) => setTName(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Short Name")}:</label>
                  <input
                    type="text"
                    placeholder={t("Short Name")}
                    value={tShort}
                    onChange={(e) => setTShort(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Category")}:</label>
                  <select
                    className="form-select-custom w-100"
                    value={tCategory}
                    onChange={(e) => setTCategory(e.target.value)}
                  >
                    {categories.map((c) => (
                      <option key={c.id} value={c.name}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Charge")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0.00"
                    value={tCharge}
                    onChange={(e) => setTCharge(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
                <button
                  type="submit"
                  className="btn-action-blue d-flex align-items-center gap-1"
                  disabled={isSubmitting}
                >
                  {isSubmitting && (
                    <Loader2 size={14} className="spinner-border spinner-border-sm" />
                  )}
                  <span>{t("Save")}</span>
                </button>
                <button
                  type="button"
                  className="btn-action-grey"
                  onClick={() => setTestModal(false)}
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
