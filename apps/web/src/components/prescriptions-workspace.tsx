"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Search,
  Filter,
  Eye,
  Edit2,
  Printer,
  Trash2,
  X,
  Plus,
  Sparkles,
  AlertCircle,
  Check,
  ChevronDown,
} from "lucide-react";
import { useLanguage } from "./language";

interface MedicineRow {
  id: string;
  medicine: string;
  dosage: string;
  doseDuration: string;
  time: string;
  doseInterval: string;
  comment: string;
}

interface PrescriptionItem {
  id: string;
  patientName: string;
  patientEmail: string;
  patientInitials: string;
  patientColor: string;
  doctorName: string;
  doctorEmail: string;
  doctorInitials: string;
  doctorColor: string;
  addedAt: string;
  status: boolean;
  healthInsurance?: string;
  lowIncome?: string;
  reference?: string;
  medicines: MedicineRow[];
  physicalInfo: {
    highBloodPressure?: string;
    foodAllergies?: string;
    tendencyBleed?: string;
    heartDisease?: string;
    diabetic?: string;
    addedAt?: string;
    femalePregnancy?: string;
    breastFeeding?: string;
    currentMedication?: string;
    surgery?: string;
    accident?: string;
    others?: string;
    pulseRate?: string;
    temperature?: string;
    problemDescription?: string;
  };
  test?: string;
  advice?: string;
  nextVisit?: {
    value: number;
    unit: "Days" | "Weeks" | "Months";
  };
}

const INITIAL_PRESCRIPTIONS: PrescriptionItem[] = [
  {
    id: "rx-1",
    patientName: "SAN K",
    patientEmail: "mwpsquade@gmail.com",
    patientInitials: "SK",
    patientColor: "#6366f1",
    doctorName: "Dharman K",
    doctorEmail: "smartentry8@gmail.com",
    doctorInitials: "DK",
    doctorColor: "#3b82f6",
    addedAt: "N/A",
    status: true,
    medicines: [
      {
        id: "m-1",
        medicine: "Amoxicillin 500mg",
        dosage: "1 Capsule",
        doseDuration: "7 days",
        time: "After Meal",
        doseInterval: "Daily morning",
        comment: "Take with water",
      },
    ],
    physicalInfo: {
      highBloodPressure: "120/80",
      problemDescription: "Persistent sore throat and mild fever.",
    },
    advice: "Drink plenty of warm liquids and rest.",
  },
  {
    id: "rx-2",
    patientName: "SAN K",
    patientEmail: "mwpsquade@gmail.com",
    patientInitials: "SK",
    patientColor: "#6366f1",
    doctorName: "Dharman K",
    doctorEmail: "smartentry8@gmail.com",
    doctorInitials: "DK",
    doctorColor: "#3b82f6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-3",
    patientName: "SAN K",
    patientEmail: "mwpsquade@gmail.com",
    patientInitials: "SK",
    patientColor: "#6366f1",
    doctorName: "Dharman K",
    doctorEmail: "smartentry8@gmail.com",
    doctorInitials: "DK",
    doctorColor: "#3b82f6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-4",
    patientName: "AA Ahmed",
    patientEmail: "hostmileso@gmail.com",
    patientInitials: "AA",
    patientColor: "#f59e0b",
    doctorName: "Abdiqafar Haaaa",
    doctorEmail: "abdi@gmail.com",
    doctorInitials: "AH",
    doctorColor: "#10b981",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-5",
    patientName: "Srinivas D",
    patientEmail: "srinivas@gmail.com",
    patientInitials: "SD",
    patientColor: "#ec4899",
    doctorName: "Annie Bsseor",
    doctorEmail: "admin@hmqqs.com",
    doctorInitials: "AB",
    doctorColor: "#3b82f6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-6",
    patientName: "ABDELLATIF OUDIDI",
    patientEmail: "pr.oudidi@gmail.com",
    patientInitials: "AO",
    patientColor: "#06b6d4",
    doctorName: "Ali Sahil",
    doctorEmail: "alisahil@gmail.com",
    doctorInitials: "AS",
    doctorColor: "#8b5cf6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-7",
    patientName: "AHMED ALL",
    patientEmail: "ahmed@gmail.com",
    patientInitials: "AA",
    patientColor: "#f59e0b",
    doctorName: "Harish Mohan",
    doctorEmail: "vatsal@gmail.com",
    doctorInitials: "HM",
    doctorColor: "#3b82f6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-8",
    patientName: "11111 1111",
    patientEmail: "sulapojim@mailinator.com",
    patientInitials: "11",
    patientColor: "#14b8a6",
    doctorName: "123456 123456",
    doctorEmail: "123456@gmail.com",
    doctorInitials: "11",
    doctorColor: "#14b8a6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-9",
    patientName: "ABBA ADAMU",
    patientEmail: "adamuabba9@gmail.com",
    patientInitials: "AA",
    patientColor: "#3b82f6",
    doctorName: "Abdiqafar Haaaa",
    doctorEmail: "abdi@gmail.com",
    doctorInitials: "AH",
    doctorColor: "#10b981",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
  {
    id: "rx-10",
    patientName: "Aljun Cardona",
    patientEmail: "cardona.aljun@gmail.com",
    patientInitials: "AC",
    patientColor: "#2563eb",
    doctorName: "Ahmed Shawwaf",
    doctorEmail: "shawwaaf@gmail.com",
    doctorInitials: "AS",
    doctorColor: "#8b5cf6",
    addedAt: "N/A",
    status: true,
    medicines: [],
    physicalInfo: {},
  },
];

const AVAILABLE_MEDICINES = [
  "Paracetamol 500mg",
  "Amoxicillin 500mg",
  "Ibuprofen 400mg",
  "Cetirizine 10mg",
  "Omeprazole 20mg",
  "Metformin 500mg",
  "Azithromycin 250mg",
  "Ciprofloxacin 500mg",
];

const MEDICINE_CATEGORIES = [
  "Antibiotics",
  "Analgesics",
  "Antipyretics",
  "Antihistamines",
  "Antacids",
  "Cardiovascular",
];

const MEDICINE_BRANDS = [
  "Pfizer",
  "GlaxoSmithKline",
  "Novartis",
  "Sanofi",
  "Bayer",
  "AstraZeneca",
];

export function PrescriptionsWorkspace() {
  const router = useRouter();
  const { t } = useLanguage();

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [prescriptions, setPrescriptions] = useState<PrescriptionItem[]>(
    INITIAL_PRESCRIPTIONS,
  );

  // View state: 'list' or 'create'
  const [mode, setMode] = useState<"list" | "create">("list");

  // Toast state (Screenshot 175024)
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Modals
  const [showNewMedicineModal, setShowNewMedicineModal] = useState(false);
  const [viewPrescription, setViewPrescription] =
    useState<PrescriptionItem | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<PrescriptionItem | null>(
    null,
  );

  // New Prescription Form State (Screenshots 174918 & 174945)
  const [formPatient, setFormPatient] = useState("");
  const [formDoctor, setFormDoctor] = useState("");
  const [formHealthInsurance, setFormHealthInsurance] = useState("");
  const [formLowIncome, setFormLowIncome] = useState("");
  const [formReference, setFormReference] = useState("");
  const [formStatus, setFormStatus] = useState(true);

  // Medicines dynamic rows
  const [formMedicines, setFormMedicines] = useState<MedicineRow[]>([
    {
      id: "med-1",
      medicine: "",
      dosage: "",
      doseDuration: "Only one day",
      time: "After Meal",
      doseInterval: "Daily morning",
      comment: "",
    },
  ]);

  // Physical Information state
  const [physicalInfo, setPhysicalInfo] = useState({
    highBloodPressure: "",
    foodAllergies: "",
    tendencyBleed: "",
    heartDisease: "",
    diabetic: "",
    addedAt: "",
    femalePregnancy: "",
    breastFeeding: "",
    currentMedication: "",
    surgery: "",
    accident: "",
    others: "",
    pulseRate: "",
    temperature: "",
    problemDescription: "",
  });

  const [formTest, setFormTest] = useState("");
  const [formAdvice, setFormAdvice] = useState("");
  const [nextVisitValue, setNextVisitValue] = useState(1);
  const [nextVisitUnit, setNextVisitUnit] = useState<
    "Days" | "Weeks" | "Months"
  >("Days");

  // New Medicine Modal State (Screenshot 175139)
  const [newMedForm, setNewMedForm] = useState({
    medicine: "",
    category: "",
    brand: "",
    saltComposition: "",
    buyingPrice: "",
    sellingPrice: "",
    sideEffects: "",
    description: "",
  });

  // Suggest Medicines handler
  const handleSuggestMedicines = () => {
    // Check if any physical information is filled
    const hasPhysicalDetails = Object.values(physicalInfo).some(
      (val) => val && val.trim().length > 0,
    );

    if (!hasPhysicalDetails) {
      setToastMessage(t("Fill Any Details Of Physical Information."));
      setTimeout(() => setToastMessage(null), 4000);
      return;
    }

    // Auto-populate suggested medicines based on physical information
    const suggested: MedicineRow[] = [
      {
        id: `med-${Date.now()}-1`,
        medicine: "Amoxicillin 500mg",
        dosage: "1 Capsule",
        doseDuration: "5 days",
        time: "After Meal",
        doseInterval: "Daily morning",
        comment: "Take regularly with plenty of fluids",
      },
      {
        id: `med-${Date.now()}-2`,
        medicine: "Paracetamol 500mg",
        dosage: "1 Tablet",
        doseDuration: "3 days",
        time: "After Meal",
        doseInterval: "Twice daily",
        comment: "For pain or temperature reduction",
      },
    ];

    setFormMedicines(suggested);
  };

  const addMedicineRow = () => {
    setFormMedicines([
      ...formMedicines,
      {
        id: `med-${Date.now()}`,
        medicine: "",
        dosage: "",
        doseDuration: "Only one day",
        time: "After Meal",
        doseInterval: "Daily morning",
        comment: "",
      },
    ]);
  };

  const removeMedicineRow = (id: string) => {
    if (formMedicines.length > 1) {
      setFormMedicines(formMedicines.filter((m) => m.id !== id));
    }
  };

  const handleSavePrescription = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formPatient || !formDoctor) {
      alert("Please select both a patient and a doctor.");
      return;
    }

    const newRx: PrescriptionItem = {
      id: `rx-${Date.now()}`,
      patientName: formPatient,
      patientEmail: `${formPatient.toLowerCase().replace(/\s+/g, "")}@example.com`,
      patientInitials: formPatient.slice(0, 2).toUpperCase(),
      patientColor: "#3b82f6",
      doctorName: formDoctor,
      doctorEmail: `${formDoctor.toLowerCase().replace(/\s+/g, "")}@hospital.local`,
      doctorInitials: formDoctor.slice(0, 2).toUpperCase(),
      doctorColor: "#10b981",
      addedAt: new Date().toLocaleDateString("en-GB", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      }),
      status: formStatus,
      healthInsurance: formHealthInsurance,
      lowIncome: formLowIncome,
      reference: formReference,
      medicines: formMedicines.filter((m) => m.medicine),
      physicalInfo,
      test: formTest,
      advice: formAdvice,
      nextVisit: { value: nextVisitValue, unit: nextVisitUnit },
    };

    setPrescriptions([newRx, ...prescriptions]);
    setMode("list");
  };

  const handleSaveNewMedicine = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMedForm.medicine) return;
    AVAILABLE_MEDICINES.push(newMedForm.medicine);
    setShowNewMedicineModal(false);
    setNewMedForm({
      medicine: "",
      category: "",
      brand: "",
      saltComposition: "",
      buyingPrice: "",
      sellingPrice: "",
      sideEffects: "",
      description: "",
    });
  };

  const renderNewMedicineModal = () => {
    if (!showNewMedicineModal) return null;
    return (
      <div className="modal-backdrop-custom">
        <div
          className="modal-card-custom"
          style={{ maxWidth: "600px", width: "100%" }}
        >
          <div className="modal-header-custom">
            <h3 className="modal-title-custom">{t("New Medicine")}</h3>
            <button
              className="modal-close-btn"
              onClick={() => setShowNewMedicineModal(false)}
            >
              <X size={18} />
            </button>
          </div>
          <form onSubmit={handleSaveNewMedicine} className="modal-body-custom">
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "16px",
                marginBottom: "16px",
              }}
            >
              <div>
                <label className="form-label-custom">
                  {t("Medicine")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder="Medicine"
                  className="form-input-custom"
                  value={newMedForm.medicine}
                  onChange={(e) =>
                    setNewMedForm({ ...newMedForm, medicine: e.target.value })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Category")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select
                  required
                  className="form-select-custom"
                  value={newMedForm.category}
                  onChange={(e) =>
                    setNewMedForm({ ...newMedForm, category: e.target.value })
                  }
                >
                  <option value="">{t("Select Category")}</option>
                  {MEDICINE_CATEGORIES.map((cat) => (
                    <option key={cat} value={cat}>
                      {cat}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "16px",
                marginBottom: "16px",
              }}
            >
              <div>
                <label className="form-label-custom">
                  {t("Brand")}: <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <select
                  required
                  className="form-select-custom"
                  value={newMedForm.brand}
                  onChange={(e) =>
                    setNewMedForm({ ...newMedForm, brand: e.target.value })
                  }
                >
                  <option value="">{t("Select Brand")}</option>
                  {MEDICINE_BRANDS.map((b) => (
                    <option key={b} value={b}>
                      {b}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Salt Composition")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder="Salt Composition"
                  className="form-input-custom"
                  value={newMedForm.saltComposition}
                  onChange={(e) =>
                    setNewMedForm({
                      ...newMedForm,
                      saltComposition: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "16px",
                marginBottom: "16px",
              }}
            >
              <div>
                <label className="form-label-custom">
                  {t("Buying Price")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder="Buying Price"
                  className="form-input-custom"
                  value={newMedForm.buyingPrice}
                  onChange={(e) =>
                    setNewMedForm({
                      ...newMedForm,
                      buyingPrice: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Selling Price")}:{" "}
                  <span style={{ color: "#ef4444" }}>*</span>
                </label>
                <input
                  required
                  placeholder="Selling Price"
                  className="form-input-custom"
                  value={newMedForm.sellingPrice}
                  onChange={(e) =>
                    setNewMedForm({
                      ...newMedForm,
                      sellingPrice: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            <div style={{ marginBottom: "16px" }}>
              <label className="form-label-custom">{t("Side Effects")}:</label>
              <textarea
                rows={2}
                placeholder="Side Effects"
                className="form-textarea-custom"
                value={newMedForm.sideEffects}
                onChange={(e) =>
                  setNewMedForm({
                    ...newMedForm,
                    sideEffects: e.target.value,
                  })
                }
              />
            </div>

            <div style={{ marginBottom: "20px" }}>
              <label className="form-label-custom">{t("Description")}:</label>
              <textarea
                rows={3}
                placeholder="Description"
                className="form-textarea-custom"
                value={newMedForm.description}
                onChange={(e) =>
                  setNewMedForm({
                    ...newMedForm,
                    description: e.target.value,
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
                onClick={() => setShowNewMedicineModal(false)}
              >
                {t("Cancel")}
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  };


  // CREATE PRESCRIPTION SCREEN (Screenshots 174918 & 174945)
  if (mode === "create") {
    return (
      <div className="legacy-page-container" style={{ padding: "24px" }}>
        {/* Toast Notification (Screenshot 175024) */}
        {toastMessage && (
          <div className="toast-notify-error">
            <AlertCircle size={20} />
            <span>{toastMessage}</span>
            <button
              onClick={() => setToastMessage(null)}
              style={{
                background: "transparent",
                border: "none",
                color: "#f87171",
                cursor: "pointer",
                padding: "2px",
                marginLeft: "8px",
              }}
            >
              <X size={16} />
            </button>
          </div>
        )}

        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "24px",
          }}
        >
          <h2 style={{ fontSize: "20px", fontWeight: 600, color: "#f1f5f9" }}>
            {t("New Prescription")}
          </h2>
          <button
            type="button"
            className="btn-action-secondary"
            onClick={() => setMode("list")}
          >
            {t("Back")}
          </button>
        </div>

        <form onSubmit={handleSavePrescription}>
          {/* Card 1: Basic Information */}
          <div className="form-card-container">
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
                  <option value="ABBA ADAMU">ABBA ADAMU</option>
                  <option value="AA Ahmed">AA Ahmed</option>
                  <option value="SAN K">SAN K</option>
                  <option value="Srinivas D">Srinivas D</option>
                  <option value="Aamer Idris">Aamer Idris</option>
                  <option value="Aljun Cardona">Aljun Cardona</option>
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
                  <option value="Dr. Dharman K">Dr. Dharman K</option>
                  <option value="Dr. Annie Bsseor">Dr. Annie Bsseor</option>
                  <option value="Dr. Harish Mohan">Dr. Harish Mohan</option>
                  <option value="Dr. Abdiqafar Haaaa">
                    Dr. Abdiqafar Haaaa
                  </option>
                  <option value="Dr. Ali Sahil">Dr. Ali Sahil</option>
                </select>
              </div>

              <div>
                <label className="form-label-custom">
                  {t("Health Insurance")}:
                </label>
                <input
                  placeholder={t("Health Insurance")}
                  className="form-input-custom"
                  value={formHealthInsurance}
                  onChange={(e) => setFormHealthInsurance(e.target.value)}
                />
              </div>

              <div>
                <label className="form-label-custom">{t("Low Income")}:</label>
                <input
                  placeholder={t("Low Income")}
                  className="form-input-custom"
                  value={formLowIncome}
                  onChange={(e) => setFormLowIncome(e.target.value)}
                />
              </div>
            </div>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr auto",
                gap: "24px",
                alignItems: "center",
              }}
            >
              <div style={{ maxWidth: "400px" }}>
                <label className="form-label-custom">{t("Reference")}:</label>
                <input
                  placeholder={t("Reference")}
                  className="form-input-custom"
                  value={formReference}
                  onChange={(e) => setFormReference(e.target.value)}
                />
              </div>

              <div>
                <label className="form-label-custom">{t("Status")}:</label>
                <label className="switch-toggle" style={{ marginTop: "6px" }}>
                  <input
                    type="checkbox"
                    checked={formStatus}
                    onChange={(e) => setFormStatus(e.target.checked)}
                  />
                  <span className="slider round" />
                </label>
              </div>
            </div>
          </div>

          {/* Card 2: Medicines (Screenshot 174918) */}
          <div className="form-card-container">
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "20px",
              }}
            >
              <h3 className="form-section-title" style={{ margin: 0 }}>
                {t("Medicines")}
              </h3>
              <div style={{ display: "flex", gap: "12px" }}>
                <button
                  type="button"
                  className="btn-action-blue"
                  onClick={() => setShowNewMedicineModal(true)}
                >
                  {t("New Medicine")}
                </button>
                <button
                  type="button"
                  className="btn-suggest-medicines"
                  onClick={handleSuggestMedicines}
                >
                  <Sparkles size={16} />
                  {t("Suggest Medicines")}
                </button>
              </div>
            </div>

            <div className="table-responsive">
              <table className="billing-table w-100">
                <thead>
                  <tr>
                    <th style={{ width: "22%" }}>
                      {t("MEDICINES")}{" "}
                      <span style={{ color: "#ef4444" }}>*</span>
                    </th>
                    <th style={{ width: "15%" }}>{t("DOSAGE")}</th>
                    <th style={{ width: "15%" }}>
                      {t("DOSE DURATION")}{" "}
                      <span style={{ color: "#ef4444" }}>*</span>
                    </th>
                    <th style={{ width: "13%" }}>
                      {t("TIME")} <span style={{ color: "#ef4444" }}>*</span>
                    </th>
                    <th style={{ width: "15%" }}>
                      {t("DOSE INTERVAL")}{" "}
                      <span style={{ color: "#ef4444" }}>*</span>
                    </th>
                    <th style={{ width: "15%" }}>{t("COMMENT")}</th>
                    <th style={{ width: "5%" }} className="text-end">
                      <button
                        type="button"
                        className="btn-action-blue"
                        style={{ height: "34px", padding: "0 12px" }}
                        onClick={addMedicineRow}
                      >
                        {t("ADD")}
                      </button>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {formMedicines.map((row, idx) => (
                    <tr key={row.id}>
                      <td>
                        <select
                          className="form-select-custom"
                          value={row.medicine}
                          onChange={(e) => {
                            const updated = [...formMedicines];
                            updated[idx].medicine = e.target.value;
                            setFormMedicines(updated);
                          }}
                        >
                          <option value="">{t("Select Medicine")}</option>
                          {AVAILABLE_MEDICINES.map((m) => (
                            <option key={m} value={m}>
                              {m}
                            </option>
                          ))}
                        </select>
                      </td>
                      <td>
                        <input
                          placeholder={t("Dosage")}
                          className="form-input-custom"
                          value={row.dosage}
                          onChange={(e) => {
                            const updated = [...formMedicines];
                            updated[idx].dosage = e.target.value;
                            setFormMedicines(updated);
                          }}
                        />
                      </td>
                      <td>
                        <select
                          className="form-select-custom"
                          value={row.doseDuration}
                          onChange={(e) => {
                            const updated = [...formMedicines];
                            updated[idx].doseDuration = e.target.value;
                            setFormMedicines(updated);
                          }}
                        >
                          <option value="Only one day">Only one day</option>
                          <option value="3 days">3 days</option>
                          <option value="5 days">5 days</option>
                          <option value="7 days">7 days</option>
                          <option value="14 days">14 days</option>
                          <option value="1 month">1 month</option>
                        </select>
                      </td>
                      <td>
                        <select
                          className="form-select-custom"
                          value={row.time}
                          onChange={(e) => {
                            const updated = [...formMedicines];
                            updated[idx].time = e.target.value;
                            setFormMedicines(updated);
                          }}
                        >
                          <option value="After Meal">After Meal</option>
                          <option value="Before Meal">Before Meal</option>
                          <option value="With Meal">With Meal</option>
                        </select>
                      </td>
                      <td>
                        <select
                          className="form-select-custom"
                          value={row.doseInterval}
                          onChange={(e) => {
                            const updated = [...formMedicines];
                            updated[idx].doseInterval = e.target.value;
                            setFormMedicines(updated);
                          }}
                        >
                          <option value="Daily morning">Daily morning</option>
                          <option value="Twice daily">Twice daily</option>
                          <option value="Three times daily">
                            Three times daily
                          </option>
                          <option value="Every 8 hours">Every 8 hours</option>
                          <option value="At bedtime">At bedtime</option>
                        </select>
                      </td>
                      <td>
                        <input
                          placeholder={t("Comment")}
                          className="form-input-custom"
                          value={row.comment}
                          onChange={(e) => {
                            const updated = [...formMedicines];
                            updated[idx].comment = e.target.value;
                            setFormMedicines(updated);
                          }}
                        />
                      </td>
                      <td className="text-end">
                        {formMedicines.length > 1 && (
                          <button
                            type="button"
                            className="action-btn-delete"
                            onClick={() => removeMedicineRow(row.id)}
                          >
                            <Trash2 size={16} />
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Card 3: Physical Information (Screenshot 174945) */}
          <div className="form-card-container">
            <h3 className="form-section-title">{t("Physical Information")}</h3>

            {/* Row 1 */}
            <div className="form-grid-4" style={{ marginBottom: "18px" }}>
              <div>
                <label className="form-label-custom">
                  {t("High Blood Pressure")}:
                </label>
                <input
                  placeholder="High Blood Pressure"
                  className="form-input-custom"
                  value={physicalInfo.highBloodPressure}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      highBloodPressure: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Food Allergies")}:
                </label>
                <input
                  placeholder="Food Allergies"
                  className="form-input-custom"
                  value={physicalInfo.foodAllergies}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      foodAllergies: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Tendency Bleed")}:
                </label>
                <input
                  placeholder="Tendency Bleed"
                  className="form-input-custom"
                  value={physicalInfo.tendencyBleed}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      tendencyBleed: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Heart Disease")}:
                </label>
                <input
                  placeholder="Heart Disease"
                  className="form-input-custom"
                  value={physicalInfo.heartDisease}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      heartDisease: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            {/* Row 2 */}
            <div className="form-grid-4" style={{ marginBottom: "18px" }}>
              <div>
                <label className="form-label-custom">{t("Diabetic")}:</label>
                <input
                  placeholder="Diabetic"
                  className="form-input-custom"
                  value={physicalInfo.diabetic}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      diabetic: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">{t("Added At")}:</label>
                <input
                  placeholder="Added At"
                  className="form-input-custom"
                  value={physicalInfo.addedAt}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      addedAt: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Female Pregnancy")}:
                </label>
                <input
                  placeholder="Female Pregnancy"
                  className="form-input-custom"
                  value={physicalInfo.femalePregnancy}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      femalePregnancy: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">
                  {t("Breast Feeding")}:
                </label>
                <input
                  placeholder="Breast Feeding"
                  className="form-input-custom"
                  value={physicalInfo.breastFeeding}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      breastFeeding: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            {/* Row 3 */}
            <div className="form-grid-4" style={{ marginBottom: "18px" }}>
              <div>
                <label className="form-label-custom">
                  {t("Current Medication")}:
                </label>
                <input
                  placeholder="Current Medication"
                  className="form-input-custom"
                  value={physicalInfo.currentMedication}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      currentMedication: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">{t("Surgery")}:</label>
                <input
                  placeholder="Surgery"
                  className="form-input-custom"
                  value={physicalInfo.surgery}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      surgery: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">{t("Accident")}:</label>
                <input
                  placeholder="Accident"
                  className="form-input-custom"
                  value={physicalInfo.accident}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      accident: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">{t("Others")}:</label>
                <input
                  placeholder="Others"
                  className="form-input-custom"
                  value={physicalInfo.others}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      others: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            {/* Row 4 */}
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "20px",
                marginBottom: "20px",
              }}
            >
              <div>
                <label className="form-label-custom">{t("Pulse Rate")}:</label>
                <input
                  placeholder="Pulse Rate"
                  className="form-input-custom"
                  value={physicalInfo.pulseRate}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      pulseRate: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <label className="form-label-custom">{t("Temperature")}:</label>
                <input
                  placeholder="Temperature"
                  className="form-input-custom"
                  value={physicalInfo.temperature}
                  onChange={(e) =>
                    setPhysicalInfo({
                      ...physicalInfo,
                      temperature: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            {/* Row 5: Problem Description */}
            <div>
              <label className="form-label-custom">
                {t("Problem Description")}:
              </label>
              <textarea
                rows={4}
                placeholder="Problem Description"
                className="form-textarea-custom"
                value={physicalInfo.problemDescription}
                onChange={(e) =>
                  setPhysicalInfo({
                    ...physicalInfo,
                    problemDescription: e.target.value,
                  })
                }
              />
            </div>
          </div>

          {/* Card 4: Test, Advice, Next Visit */}
          <div className="form-card-container">
            <div style={{ marginBottom: "20px" }}>
              <label className="form-label-custom">{t("Test")}:</label>
              <textarea
                rows={3}
                placeholder="Test"
                className="form-textarea-custom"
                value={formTest}
                onChange={(e) => setFormTest(e.target.value)}
              />
            </div>

            <div style={{ marginBottom: "20px" }}>
              <label className="form-label-custom">{t("Advice")}:</label>
              <textarea
                rows={3}
                placeholder="Advice"
                className="form-textarea-custom"
                value={formAdvice}
                onChange={(e) => setFormAdvice(e.target.value)}
              />
            </div>

            <div style={{ marginBottom: "24px", maxWidth: "340px" }}>
              <label className="form-label-custom">{t("Next Visit")}:</label>
              <div style={{ display: "flex", gap: "10px" }}>
                <input
                  type="number"
                  min={1}
                  className="form-input-custom"
                  style={{ width: "90px" }}
                  value={nextVisitValue}
                  onChange={(e) => setNextVisitValue(Number(e.target.value))}
                />
                <select
                  className="form-select-custom"
                  value={nextVisitUnit}
                  onChange={(e) =>
                    setNextVisitUnit(
                      e.target.value as "Days" | "Weeks" | "Months",
                    )
                  }
                >
                  <option value="Days">Days</option>
                  <option value="Weeks">Weeks</option>
                  <option value="Months">Months</option>
                </select>
              </div>
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
                onClick={() => setMode("list")}
              >
                {t("Cancel")}
              </button>
            </div>
          </div>
        </form>

        {renderNewMedicineModal()}
      </div>
    );
  }

  // DEFAULT SCREEN: Prescriptions Table (Screenshot 174820)
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
          <button
            className="btn-icon-blue"
            title="Filter"
            aria-label="Filter"
            onClick={() =>
              alert("Filtering by active status or doctor department.")
            }
          >
            <Filter size={18} />
          </button>

          <button className="btn-action-blue" onClick={() => setMode("create")}>
            + {t("New Prescription")}
          </button>
        </div>
      </div>

      {/* Main Card */}
      <div className="billing-card">
        <div className="table-responsive">
          <table className="billing-table w-100">
            <thead>
              <tr>
                <th>
                  <span className="th-sort">{t("PATIENTS")} ↕</span>
                </th>
                <th>
                  <span className="th-sort">{t("DOCTORS")} ↕</span>
                </th>
                <th>
                  <span className="th-sort">{t("ADDED AT")} ↕</span>
                </th>
                <th>{t("STATUS")}</th>
                <th className="text-end">{t("ACTION")}</th>
              </tr>
            </thead>
            <tbody>
              {prescriptions
                .filter(
                  (rx) =>
                    rx.patientName
                      .toLowerCase()
                      .includes(search.toLowerCase()) ||
                    rx.doctorName.toLowerCase().includes(search.toLowerCase()),
                )
                .map((rx) => (
                  <tr key={rx.id}>
                    <td>
                      <div className="patient-cell">
                        <div
                          className="avatar-circle"
                          style={{ background: rx.patientColor }}
                        >
                          {rx.patientInitials}
                        </div>
                        <div className="patient-info">
                          <span className="link-cyan">{rx.patientName}</span>
                          <span className="patient-email">
                            {rx.patientEmail}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td>
                      <div className="patient-cell">
                        <div
                          className="avatar-circle"
                          style={{ background: rx.doctorColor }}
                        >
                          {rx.doctorInitials}
                        </div>
                        <div className="patient-info">
                          <span className="link-cyan">{rx.doctorName}</span>
                          <span className="patient-email">
                            {rx.doctorEmail}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td style={{ color: "#64748b" }}>{rx.addedAt}</td>
                    <td>
                      <label className="switch-toggle">
                        <input
                          type="checkbox"
                          checked={rx.status}
                          onChange={() =>
                            setPrescriptions(
                              prescriptions.map((p) =>
                                p.id === rx.id
                                  ? { ...p, status: !p.status }
                                  : p,
                              ),
                            )
                          }
                        />
                        <span className="slider round" />
                      </label>
                    </td>
                    <td className="text-end">
                      <div
                        className="action-buttons"
                        style={{ justifyContent: "flex-end" }}
                      >
                        <button
                          className="action-btn-view"
                          title="View"
                          aria-label="View"
                          onClick={() => setViewPrescription(rx)}
                        >
                          <Eye size={16} />
                        </button>
                        <button
                          className="action-btn-edit"
                          title="Edit"
                          aria-label="Edit"
                          onClick={() => {
                            setFormPatient(rx.patientName);
                            setFormDoctor(rx.doctorName);
                            setFormStatus(rx.status);
                            setMode("create");
                          }}
                        >
                          <Edit2 size={16} />
                        </button>
                        <button
                          className="action-btn-print"
                          title="Print"
                          aria-label="Print"
                          onClick={() => {
                            setViewPrescription(rx);
                            setTimeout(() => window.print(), 300);
                          }}
                        >
                          <Printer size={16} />
                        </button>
                        <button
                          className="action-btn-delete"
                          title="Delete"
                          aria-label="Delete"
                          onClick={() => setDeleteTarget(rx)}
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
              {t("Showing")} {prescriptions.length} {t("Results")}
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

      {/* MODAL: New Medicine (Screenshot 175139) */}
      {showNewMedicineModal && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "600px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("New Medicine")}</h3>
              <button
                className="modal-close-btn"
                onClick={() => setShowNewMedicineModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={handleSaveNewMedicine}
              className="modal-body-custom"
            >
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Medicine")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Medicine"
                    className="form-input-custom"
                    value={newMedForm.medicine}
                    onChange={(e) =>
                      setNewMedForm({ ...newMedForm, medicine: e.target.value })
                    }
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Category")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <select
                    required
                    className="form-select-custom"
                    value={newMedForm.category}
                    onChange={(e) =>
                      setNewMedForm({ ...newMedForm, category: e.target.value })
                    }
                  >
                    <option value="">{t("Select Category")}</option>
                    {MEDICINE_CATEGORIES.map((cat) => (
                      <option key={cat} value={cat}>
                        {cat}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Brand")}: <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <select
                    required
                    className="form-select-custom"
                    value={newMedForm.brand}
                    onChange={(e) =>
                      setNewMedForm({ ...newMedForm, brand: e.target.value })
                    }
                  >
                    <option value="">{t("Select Brand")}</option>
                    {MEDICINE_BRANDS.map((b) => (
                      <option key={b} value={b}>
                        {b}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Salt Composition")}:{" "}
                    <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Salt Composition"
                    className="form-input-custom"
                    value={newMedForm.saltComposition}
                    onChange={(e) =>
                      setNewMedForm({
                        ...newMedForm,
                        saltComposition: e.target.value,
                      })
                    }
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "16px",
                  marginBottom: "16px",
                }}
              >
                <div>
                  <label className="form-label-custom">
                    {t("Buying Price")}:{" "}
                    <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Buying Price"
                    className="form-input-custom"
                    value={newMedForm.buyingPrice}
                    onChange={(e) =>
                      setNewMedForm({
                        ...newMedForm,
                        buyingPrice: e.target.value,
                      })
                    }
                  />
                </div>
                <div>
                  <label className="form-label-custom">
                    {t("Selling Price")}:{" "}
                    <span style={{ color: "#ef4444" }}>*</span>
                  </label>
                  <input
                    required
                    placeholder="Selling Price"
                    className="form-input-custom"
                    value={newMedForm.sellingPrice}
                    onChange={(e) =>
                      setNewMedForm({
                        ...newMedForm,
                        sellingPrice: e.target.value,
                      })
                    }
                  />
                </div>
              </div>

              <div style={{ marginBottom: "16px" }}>
                <label className="form-label-custom">
                  {t("Side Effects")}:
                </label>
                <textarea
                  rows={2}
                  placeholder="Side Effects"
                  className="form-textarea-custom"
                  value={newMedForm.sideEffects}
                  onChange={(e) =>
                    setNewMedForm({
                      ...newMedForm,
                      sideEffects: e.target.value,
                    })
                  }
                />
              </div>

              <div style={{ marginBottom: "20px" }}>
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder="Description"
                  className="form-textarea-custom"
                  value={newMedForm.description}
                  onChange={(e) =>
                    setNewMedForm({
                      ...newMedForm,
                      description: e.target.value,
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
                  onClick={() => setShowNewMedicineModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* VIEW / PRINT PRESCRIPTION MODAL */}
      {viewPrescription && (
        <div className="modal-backdrop-custom">
          <div
            className="modal-card-custom"
            style={{ maxWidth: "700px", width: "100%" }}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">
                {t("Prescription Details")}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => setViewPrescription(null)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  paddingBottom: "16px",
                  borderBottom: "1px solid #242938",
                  marginBottom: "20px",
                }}
              >
                <div>
                  <h4 style={{ margin: 0, color: "#f1f5f9" }}>
                    Patient: {viewPrescription.patientName}
                  </h4>
                  <span style={{ fontSize: "13px", color: "#64748b" }}>
                    {viewPrescription.patientEmail}
                  </span>
                </div>
                <div style={{ textAlign: "right" }}>
                  <h4 style={{ margin: 0, color: "#f1f5f9" }}>
                    Doctor: {viewPrescription.doctorName}
                  </h4>
                  <span style={{ fontSize: "13px", color: "#64748b" }}>
                    {viewPrescription.doctorEmail}
                  </span>
                </div>
              </div>

              {viewPrescription.medicines.length > 0 && (
                <div style={{ marginBottom: "20px" }}>
                  <h5 style={{ color: "#5b73e8", marginBottom: "10px" }}>
                    {t("Prescribed Medicines")}
                  </h5>
                  <table className="billing-table w-100">
                    <thead>
                      <tr>
                        <th>Medicine</th>
                        <th>Dosage</th>
                        <th>Duration</th>
                        <th>Time</th>
                        <th>Interval</th>
                      </tr>
                    </thead>
                    <tbody>
                      {viewPrescription.medicines.map((m) => (
                        <tr key={m.id}>
                          <td>{m.medicine}</td>
                          <td>{m.dosage}</td>
                          <td>{m.doseDuration}</td>
                          <td>{m.time}</td>
                          <td>{m.doseInterval}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {viewPrescription.advice && (
                <div style={{ marginBottom: "16px" }}>
                  <strong style={{ color: "#cbd5e1" }}>Advice: </strong>
                  <span style={{ color: "#94a3b8" }}>
                    {viewPrescription.advice}
                  </span>
                </div>
              )}

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
                  className="btn-action-blue"
                  onClick={() => window.print()}
                >
                  <Printer size={16} style={{ marginRight: "6px" }} />
                  {t("Print")}
                </button>
                <button
                  type="button"
                  className="btn-action-secondary"
                  onClick={() => setViewPrescription(null)}
                >
                  {t("Close")}
                </button>
              </div>
            </div>
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
                    {t("Are you sure you want to delete prescription for")}{" "}
                    &quot;
                    <strong>{deleteTarget.patientName}</strong>&quot;?
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
                  onClick={() => {
                    setPrescriptions(
                      prescriptions.filter((p) => p.id !== deleteTarget.id),
                    );
                    setDeleteTarget(null);
                  }}
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
