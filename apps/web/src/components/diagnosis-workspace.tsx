"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  Printer,
  X,
  AlertCircle,
  ChevronDown,
} from "lucide-react";
import { useLanguage } from "./language";

export type DiagnosisTab = "diagnosis-categories" | "patient-diagnosis-test";

interface CategoryItem {
  id: string;
  name: string;
  description?: string;
}

interface DiagnosisProperty {
  id: string;
  name: string;
  value: string;
}

interface DiagnosisTestItem {
  id: string;
  reportNumber: string;
  patientName: string;
  patientEmail: string;
  patientInitials: string;
  patientColor: string;
  doctorName: string;
  doctorEmail: string;
  doctorInitials: string;
  doctorColor: string;
  category: string;
  createdOn: string;
  age?: string;
  height?: string;
  weight?: string;
  averageGlucose?: string;
  fastingBloodSugar?: string;
  urineSugar?: string;
  bloodPressure?: string;
  diabetes?: string;
  cholesterol?: string;
  properties?: DiagnosisProperty[];
}

const INITIAL_CATEGORIES: CategoryItem[] = [
  { id: "cat-1", name: "X-Ray", description: "Radiological imaging" },
  {
    id: "cat-2",
    name: "Reproductive Health & Family Planning",
    description: "Women & family health tests",
  },
  { id: "cat-3", name: "DC002 - Neoplasms", description: "Oncology screening" },
  {
    id: "cat-4",
    name: "DC003 - Endocrine, Nutritional & Metabolic",
    description: "Metabolic panel",
  },
  {
    id: "cat-5",
    name: "Rheumatological Disorders",
    description: "Joint and autoimmune screening",
  },
  {
    id: "cat-6",
    name: "Psychiatric and Psychological Disorders",
    description: "Mental health assessment",
  },
  { id: "cat-7", name: "Cancer Diagnosis", description: "Biopsy and markers" },
  {
    id: "cat-8",
    name: "Endocrine Disorders",
    description: "Thyroid & hormone tests",
  },
  {
    id: "cat-9",
    name: "Dermatological Disorders",
    description: "Skin allergy and scraping",
  },
  {
    id: "cat-10",
    name: "Musculoskeletal Disorders",
    description: "Bone density and joint scans",
  },
  { id: "cat-11", name: "cbc", description: "Complete Blood Count" },
  { id: "cat-12", name: "MRI", description: "Magnetic Resonance Imaging" },
  { id: "cat-13", name: "ENT", description: "Ear Nose Throat testing" },
  { id: "cat-14", name: "Blood Test", description: "General blood work" },
  { id: "cat-15", name: "Diabetes Test", description: "HbA1c and glucose" },
];

const INITIAL_TESTS: DiagnosisTestItem[] = [
  {
    id: "dt-1",
    reportNumber: "X1SAVE2Y",
    patientName: "Ahmed Noor",
    patientEmail: "noor2@gmail.com",
    patientInitials: "AN",
    patientColor: "#f59e0b",
    doctorName: "Dr noor Noor",
    doctorEmail: "noor1@gmail.com",
    doctorInitials: "DN",
    doctorColor: "#ef4444",
    category: "cbc",
    createdOn: "22nd Aug,2026",
    age: "34",
    fastingBloodSugar: "92 mg/dL",
    bloodPressure: "120/80",
    cholesterol: "185 mg/dL",
  },
  {
    id: "dt-2",
    reportNumber: "GJDFERXG",
    patientName: "Abdullah Abdullah",
    patientEmail: "abdullah123@gmail.com",
    patientInitials: "AA",
    patientColor: "#8b5cf6",
    doctorName: "AARAV Singh",
    doctorEmail: "aarav@gmail.com",
    doctorInitials: "AS",
    doctorColor: "#3b82f6",
    category: "55",
    createdOn: "12th Aug,2026",
  },
  {
    id: "dt-3",
    reportNumber: "NSKHMJVF",
    patientName: "ASX ASX",
    patientEmail: "asx@wd.com",
    patientInitials: "AA",
    patientColor: "#8b5cf6",
    doctorName: "AAAAA BBBBB",
    doctorEmail: "abab8764350@gmail.com",
    doctorInitials: "AB",
    doctorColor: "#10b981",
    category: "MRI",
    createdOn: "25th Jul,2026",
  },
  {
    id: "dt-4",
    reportNumber: "VVMMYD1Q",
    patientName: "Aljun Cardona",
    patientEmail: "cardona.aljun@gmail.com",
    patientInitials: "AC",
    patientColor: "#f59e0b",
    doctorName: "Abdiqafar Haaaa",
    doctorEmail: "abdi@gmail.com",
    doctorInitials: "AH",
    doctorColor: "#3b82f6",
    category: "Analyse médical",
    createdOn: "23rd Jul,2026",
  },
  {
    id: "dt-5",
    reportNumber: "OKEADLS7",
    patientName: "Alo Kdk",
    patientEmail: "hola@gmail.com",
    patientInitials: "AK",
    patientColor: "#10b981",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    doctorInitials: "HM",
    doctorColor: "#3b82f6",
    category: "Alcohol/Drug Use or Induced Mental Disorderss",
    createdOn: "16th Jun,2026",
  },
  {
    id: "dt-6",
    reportNumber: "FLMX3D1T",
    patientName: "Ahmed Mohamed",
    patientEmail: "ahmed00@gmail.com",
    patientInitials: "AM",
    patientColor: "#6366f1",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    doctorInitials: "HM",
    doctorColor: "#3b82f6",
    category: "X-Ray",
    createdOn: "1st Jun,2026",
  },
  {
    id: "dt-7",
    reportNumber: "K4CJQBZI",
    patientName: "Liz Cecilia Zawadi",
    patientEmail: "zawadizandre@gmail.com",
    patientInitials: "LZ",
    patientColor: "#8b5cf6",
    doctorName: "Ali Sahil",
    doctorEmail: "alisahil@gmail.com",
    doctorInitials: "AS",
    doctorColor: "#8b5cf6",
    category: "ENT",
    createdOn: "26th May,2026",
  },
  {
    id: "dt-8",
    reportNumber: "XFXNCXQY",
    patientName: "Aadil Bandi",
    patientEmail: "aadil@gmail.com",
    patientInitials: "AB",
    patientColor: "#6366f1",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    doctorInitials: "HM",
    doctorColor: "#3b82f6",
    category: "Blood Test",
    createdOn: "6th May,2026",
  },
  {
    id: "dt-9",
    reportNumber: "FRBGYYOC",
    patientName: "Rose Miller",
    patientEmail: "rose@gmail.com",
    patientInitials: "RM",
    patientColor: "#10b981",
    doctorName: "Nero Patrick",
    doctorEmail: "neropatrick@gmail.com",
    doctorInitials: "NP",
    doctorColor: "#f59e0b",
    category: "Diabetes Test",
    createdOn: "16th Apr,2026",
  },
  {
    id: "dt-10",
    reportNumber: "SI2QYZMR",
    patientName: "ANDRE SANTOS",
    patientEmail: "jfismaeldjfdj@gmail.com",
    patientInitials: "AS",
    patientColor: "#ef4444",
    doctorName: "Aya Hasan",
    doctorEmail: "ayaalsoreky@gmail.com",
    doctorInitials: "AH",
    doctorColor: "#ef4444",
    category: "Cancer Diagnosis",
    createdOn: "9th Apr,2026",
  },
];

const generateReportNumber = () => {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  let result = "";
  for (let i = 0; i < 8; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
};

export function DiagnosisWorkspace({ id }: { id: string }) {
  const router = useRouter();
  const { t } = useLanguage();

  const activeTab: DiagnosisTab = (
    id === "diagnosis-categories"
      ? "diagnosis-categories"
      : "patient-diagnosis-test"
  ) as DiagnosisTab;

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);

  // Data states
  const [categories, setCategories] =
    useState<CategoryItem[]>(INITIAL_CATEGORIES);
  const [tests, setTests] = useState<DiagnosisTestItem[]>(INITIAL_TESTS);

  // Sub-view: list or create test
  const [isCreatingTest, setIsCreatingTest] = useState(false);

  // Modals
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    type: "category" | "test";
    name: string;
  } | null>(null);

  // Category Modal Form State (Screenshot 175907)
  const [categoryName, setCategoryName] = useState("");
  const [categoryDesc, setCategoryDesc] = useState("");

  // New Patient Diagnosis Test Form State (Screenshot 175655)
  const [formPatient, setFormPatient] = useState("");
  const [formDoctor, setFormDoctor] = useState("");
  const [formCategory, setFormCategory] = useState("");
  const [reportNumber, setReportNumber] = useState(generateReportNumber());
  const [age, setAge] = useState("");
  const [height, setHeight] = useState("");
  const [weight, setWeight] = useState("");
  const [avgGlucose, setAvgGlucose] = useState("");
  const [fastingSugar, setFastingSugar] = useState("");
  const [urineSugar, setUrineSugar] = useState("");
  const [bloodPressure, setBloodPressure] = useState("");
  const [diabetes, setDiabetes] = useState("");
  const [cholesterol, setCholesterol] = useState("");

  // Custom properties
  const [customProperties, setCustomProperties] = useState<DiagnosisProperty[]>(
    [],
  );

  const addCustomPropertyRow = () => {
    setCustomProperties([
      ...customProperties,
      { id: `prop-${Date.now()}`, name: "", value: "" },
    ]);
  };

  const removeCustomPropertyRow = (id: string) => {
    setCustomProperties(customProperties.filter((p) => p.id !== id));
  };

  const handleSaveCategory = (e: React.FormEvent) => {
    e.preventDefault();
    if (!categoryName) return;
    setCategories([
      {
        id: `cat-${Date.now()}`,
        name: categoryName,
        description: categoryDesc,
      },
      ...categories,
    ]);
    setShowCategoryModal(false);
    setCategoryName("");
    setCategoryDesc("");
  };

  const handleSaveTest = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formPatient || !formDoctor || !formCategory) {
      alert("Please select Patient, Doctor, and Diagnosis Category.");
      return;
    }

    const newTest: DiagnosisTestItem = {
      id: `dt-${Date.now()}`,
      reportNumber,
      patientName: formPatient,
      patientEmail: `${formPatient.toLowerCase().replace(/\s+/g, "")}@gmail.com`,
      patientInitials: formPatient.slice(0, 2).toUpperCase(),
      patientColor: "#3b82f6",
      doctorName: formDoctor,
      doctorEmail: `${formDoctor.toLowerCase().replace(/\s+/g, "")}@hospital.local`,
      doctorInitials: formDoctor.slice(0, 2).toUpperCase(),
      doctorColor: "#10b981",
      category: formCategory,
      createdOn: new Date().toLocaleDateString("en-GB", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      }),
      age,
      height,
      weight,
      averageGlucose: avgGlucose,
      fastingBloodSugar: fastingSugar,
      urineSugar,
      bloodPressure,
      diabetes,
      cholesterol,
      properties: customProperties.filter((p) => p.name.trim()),
    };

    setTests([newTest, ...tests]);
    setIsCreatingTest(false);
    // Reset form
    setReportNumber(generateReportNumber());
    setFormPatient("");
    setFormDoctor("");
    setFormCategory("");
    setAge("");
    setHeight("");
    setWeight("");
    setAvgGlucose("");
    setFastingSugar("");
    setUrineSugar("");
    setBloodPressure("");
    setDiabetes("");
    setCholesterol("");
    setCustomProperties([]);
  };

  const confirmDelete = () => {
    if (!deleteTarget) return;
    if (deleteTarget.type === "category") {
      setCategories(categories.filter((c) => c.id !== deleteTarget.id));
    } else {
      setTests(tests.filter((t) => t.id !== deleteTarget.id));
    }
    setDeleteTarget(null);
  };

  // SCREEN: New Patient Diagnosis Test (Screenshot 175655)
  if (isCreatingTest) {
    return (
      <div className="legacy-page-container" style={{ padding: "24px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "24px",
          }}
        >
          <h2 style={{ fontSize: "20px", fontWeight: 600, color: "#f1f5f9" }}>
            {t("New Patient Diagnosis Test")}
          </h2>
          <button
            type="button"
            className="btn-action-secondary"
            onClick={() => setIsCreatingTest(false)}
          >
            {t("Back")}
          </button>
        </div>

        <form onSubmit={handleSaveTest} className="form-card-container">
          {/* Row 1: Patient, Doctor, Category, Report Number */}
          <div className="form-grid-4" style={{ marginBottom: "20px" }}>
            <div>
              <label className="form-label-custom">
                {t("Patient")}: <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <select
                required
                className="form-select-custom"
                value={formPatient}
                onChange={(e) => setFormPatient(e.target.value)}
              >
                <option value="">{t("Select Patient")}</option>
                <option value="Ahmed Noor">Ahmed Noor</option>
                <option value="Abdullah Abdullah">Abdullah Abdullah</option>
                <option value="Aljun Cardona">Aljun Cardona</option>
                <option value="Alo Kdk">Alo Kdk</option>
                <option value="Rose Miller">Rose Miller</option>
                <option value="ANDRE SANTOS">ANDRE SANTOS</option>
              </select>
            </div>

            <div>
              <label className="form-label-custom">
                {t("Doctor")}: <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <select
                required
                className="form-select-custom"
                value={formDoctor}
                onChange={(e) => setFormDoctor(e.target.value)}
              >
                <option value="">{t("Select Doctor")}</option>
                <option value="Dr noor Noor">Dr noor Noor</option>
                <option value="AARAV Singh">AARAV Singh</option>
                <option value="Harish Mohan">Harish Mohan</option>
                <option value="Nero Patrick">Nero Patrick</option>
                <option value="Aya Hasan">Aya Hasan</option>
              </select>
            </div>

            <div>
              <label className="form-label-custom">
                {t("Diagnosis Category")}:{" "}
                <span style={{ color: "#ef4444" }}>*</span>
              </label>
              <select
                required
                className="form-select-custom"
                value={formCategory}
                onChange={(e) => setFormCategory(e.target.value)}
              >
                <option value="">{t("Choose Diagnosis Category")}</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.name}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="form-label-custom">{t("Report number")}:</label>
              <input
                readOnly
                className="form-input-custom"
                value={reportNumber}
              />
            </div>
          </div>

          {/* Row 2: Age, Height, Weight, Average glucose */}
          <div className="form-grid-4" style={{ marginBottom: "20px" }}>
            <div>
              <label className="form-label-custom">{t("Age")}:</label>
              <input
                placeholder={t("Age")}
                className="form-input-custom"
                value={age}
                onChange={(e) => setAge(e.target.value)}
              />
            </div>
            <div>
              <label className="form-label-custom">{t("Height")}:</label>
              <input
                placeholder={t("Height")}
                className="form-input-custom"
                value={height}
                onChange={(e) => setHeight(e.target.value)}
              />
            </div>
            <div>
              <label className="form-label-custom">{t("Weight")}:</label>
              <input
                placeholder={t("Weight")}
                className="form-input-custom"
                value={weight}
                onChange={(e) => setWeight(e.target.value)}
              />
            </div>
            <div>
              <label className="form-label-custom">
                {t("Average glucose")}:
              </label>
              <input
                placeholder={t("Average glucose")}
                className="form-input-custom"
                value={avgGlucose}
                onChange={(e) => setAvgGlucose(e.target.value)}
              />
            </div>
          </div>

          {/* Row 3: Fasting Blood Sugar, Urine Sugar, Blood Pressure, Diabetes */}
          <div className="form-grid-4" style={{ marginBottom: "20px" }}>
            <div>
              <label className="form-label-custom">
                {t("Fasting Blood Sugar")}:
              </label>
              <input
                placeholder={t("Fasting Blood Sugar")}
                className="form-input-custom"
                value={fastingSugar}
                onChange={(e) => setFastingSugar(e.target.value)}
              />
            </div>
            <div>
              <label className="form-label-custom">{t("Urine Sugar")}:</label>
              <input
                placeholder={t("Urine Sugar")}
                className="form-input-custom"
                value={urineSugar}
                onChange={(e) => setUrineSugar(e.target.value)}
              />
            </div>
            <div>
              <label className="form-label-custom">
                {t("Blood Pressure")}:
              </label>
              <input
                placeholder={t("Blood Pressure")}
                className="form-input-custom"
                value={bloodPressure}
                onChange={(e) => setBloodPressure(e.target.value)}
              />
            </div>
            <div>
              <label className="form-label-custom">{t("Diabetes")}:</label>
              <input
                placeholder={t("Diabetes")}
                className="form-input-custom"
                value={diabetes}
                onChange={(e) => setDiabetes(e.target.value)}
              />
            </div>
          </div>

          {/* Row 4: Cholesterol */}
          <div style={{ maxWidth: "24%", marginBottom: "28px" }}>
            <label className="form-label-custom">{t("Cholesterol")}:</label>
            <input
              placeholder={t("Cholesterol")}
              className="form-input-custom"
              value={cholesterol}
              onChange={(e) => setCholesterol(e.target.value)}
            />
          </div>

          {/* Add other diagnosis property section */}
          <div style={{ marginBottom: "28px" }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "14px",
              }}
            >
              <h4
                style={{
                  fontSize: "14px",
                  fontWeight: 600,
                  color: "#cbd5e1",
                  margin: 0,
                }}
              >
                {t("Add other diagnosis property")}
              </h4>
              <button
                type="button"
                className="btn-action-blue"
                style={{ height: "34px", padding: "0 14px" }}
                onClick={addCustomPropertyRow}
              >
                {t("Add")}
              </button>
            </div>

            {customProperties.length > 0 && (
              <div className="table-responsive">
                <table className="billing-table w-100">
                  <thead>
                    <tr>
                      <th style={{ width: "5%" }}>#</th>
                      <th style={{ width: "45%" }}>
                        {t("DIAGNOSIS PROPERTY NAME")}
                      </th>
                      <th style={{ width: "45%" }}>
                        {t("DIAGNOSIS PROPERTY VALUE")}
                      </th>
                      <th style={{ width: "5%" }} className="text-end">
                        {t("ACTION")}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {customProperties.map((prop, idx) => (
                      <tr key={prop.id}>
                        <td>{idx + 1}</td>
                        <td>
                          <input
                            placeholder="Property Name"
                            className="form-input-custom"
                            value={prop.name}
                            onChange={(e) => {
                              const updated = [...customProperties];
                              updated[idx].name = e.target.value;
                              setCustomProperties(updated);
                            }}
                          />
                        </td>
                        <td>
                          <input
                            placeholder="Property Value"
                            className="form-input-custom"
                            value={prop.value}
                            onChange={(e) => {
                              const updated = [...customProperties];
                              updated[idx].value = e.target.value;
                              setCustomProperties(updated);
                            }}
                          />
                        </td>
                        <td className="text-end">
                          <button
                            type="button"
                            className="action-btn-delete"
                            onClick={() => removeCustomPropertyRow(prop.id)}
                          >
                            <Trash2 size={16} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
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
              onClick={() => setIsCreatingTest(false)}
            >
              {t("Cancel")}
            </button>
          </div>
        </form>
      </div>
    );
  }

  // DEFAULT SCREEN: Sub-Tabs List
  return (
    <div className="legacy-page-container" style={{ padding: "24px" }}>
      {/* Toolbar */}
      <div className="billing-toolbar">
        <div className="billing-search-box">
          <Search size={16} className="search-icon" />
          <input
            placeholder={t("Search")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        <div className="billing-actions">
          {activeTab === "diagnosis-categories" ? (
            <button
              className="btn-action-blue"
              onClick={() => setShowCategoryModal(true)}
            >
              {t("New Diagnosis Category")}
            </button>
          ) : (
            <button
              className="btn-action-blue"
              onClick={() => setIsCreatingTest(true)}
            >
              {t("New Patient Diagnosis Test")}
            </button>
          )}
        </div>
      </div>

      {/* Main Card */}
      <div className="billing-card">
        {/* TAB 1: Diagnosis Categories (Screenshot 175837) */}
        {activeTab === "diagnosis-categories" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("DIAGNOSIS CATEGORY")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {categories
                  .filter((c) =>
                    c.name.toLowerCase().includes(search.toLowerCase()),
                  )
                  .map((cat) => (
                    <tr key={cat.id}>
                      <td style={{ fontWeight: 500, color: "#f1f5f9" }}>
                        {cat.name}
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
                              const newName = prompt(
                                "Enter Category Name:",
                                cat.name,
                              );
                              if (newName) {
                                setCategories(
                                  categories.map((c) =>
                                    c.id === cat.id
                                      ? { ...c, name: newName }
                                      : c,
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
                                id: cat.id,
                                type: "category",
                                name: cat.name,
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

        {/* TAB 2: Diagnosis Tests (Screenshot 175323) */}
        {activeTab === "patient-diagnosis-test" && (
          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>
                    <span className="th-sort">{t("REPORT NUMBER")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("PATIENT")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("DOCTOR")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("DIAGNOSIS CATEGORY")} ↕</span>
                  </th>
                  <th>
                    <span className="th-sort">{t("CREATED ON")} ↕</span>
                  </th>
                  <th className="text-end">{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {tests
                  .filter(
                    (test) =>
                      test.reportNumber
                        .toLowerCase()
                        .includes(search.toLowerCase()) ||
                      test.patientName
                        .toLowerCase()
                        .includes(search.toLowerCase()) ||
                      test.doctorName
                        .toLowerCase()
                        .includes(search.toLowerCase()) ||
                      test.category
                        .toLowerCase()
                        .includes(search.toLowerCase()),
                  )
                  .map((test) => (
                    <tr key={test.id}>
                      <td>
                        <span className="badge-blue-pill">
                          {test.reportNumber}
                        </span>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <div
                            className="avatar-circle"
                            style={{ background: test.patientColor }}
                          >
                            {test.patientInitials}
                          </div>
                          <div className="patient-info">
                            <span className="link-cyan">
                              {test.patientName}
                            </span>
                            <span className="patient-email">
                              {test.patientEmail}
                            </span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="patient-cell">
                          <div
                            className="avatar-circle"
                            style={{ background: test.doctorColor }}
                          >
                            {test.doctorInitials}
                          </div>
                          <div className="patient-info">
                            <span className="link-cyan">{test.doctorName}</span>
                            <span className="patient-email">
                              {test.doctorEmail}
                            </span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="link-cyan">{test.category}</span>
                      </td>
                      <td>
                        <span className="badge-blue-pill">
                          {test.createdOn}
                        </span>
                      </td>
                      <td className="text-end">
                        <div
                          className="action-buttons"
                          style={{ justifyContent: "flex-end" }}
                        >
                          <button
                            className="action-btn-print"
                            title="Print"
                            aria-label="Print"
                            onClick={() => window.print()}
                          >
                            <Printer size={16} />
                          </button>
                          <button
                            className="action-btn-edit"
                            title="Edit"
                            aria-label="Edit"
                            onClick={() => setIsCreatingTest(true)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            className="action-btn-delete"
                            title="Delete"
                            aria-label="Delete"
                            onClick={() =>
                              setDeleteTarget({
                                id: test.id,
                                type: "test",
                                name: `Report ${test.reportNumber}`,
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

        {/* Footer */}
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
              {activeTab === "diagnosis-categories"
                ? categories.length
                : tests.length}{" "}
              {t("Results")}
            </span>
          </div>

          <div className="billing-pagination-controls">
            <button className="billing-page-btn" disabled>
              ‹
            </button>
            <button className="billing-page-btn is-active">1</button>
            <button className="billing-page-btn">2</button>
            <button className="billing-page-btn">›</button>
          </div>
        </div>
      </div>

      {/* MODAL: New Diagnosis Category (Screenshot 175907) */}
      {showCategoryModal && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "520px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("New Diagnosis Category")}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowCategoryModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleSaveCategory} className="modal-body-custom">
              <div
                className="form-group-custom"
                style={{ marginBottom: "18px" }}
              >
                <label className="form-label-custom">
                  {t("Diagnosis Category")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder={t("Diagnosis Category")}
                  className="form-input-custom"
                  value={categoryName}
                  onChange={(e) => setCategoryName(e.target.value)}
                />
              </div>

              <div
                className="form-group-custom"
                style={{ marginBottom: "24px" }}
              >
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={4}
                  placeholder={t("Description")}
                  className="form-textarea-custom"
                  value={categoryDesc}
                  onChange={(e) => setCategoryDesc(e.target.value)}
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
                  onClick={() => setShowCategoryModal(false)}
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
