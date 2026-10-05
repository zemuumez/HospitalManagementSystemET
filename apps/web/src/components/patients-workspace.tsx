"use client";

import { useLanguage } from "@/components/language";
import { useState } from "react";
import Link from "next/link";
import {
  Search,
  Filter,
  FileSpreadsheet,
  Plus,
  Key,
  Edit2,
  Trash2,
  ArrowLeft,
  X,
  Camera,
} from "lucide-react";

export type PatientsWorkspaceProps = {
  id?: string;
};

interface PatientRow {
  id: string;
  name: string;
  email: string;
  phone: string;
  bloodGroup: string;
  status: boolean;
}

interface CaseRow {
  id: string;
  caseId: string;
  patientName: string;
  patientEmail: string;
  doctorName: string;
  doctorEmail: string;
  caseDate: string;
  caseTime: string;
  fee: number;
  status: boolean;
}

interface CaseHandlerRow {
  id: string;
  name: string;
  email: string;
  phone: string;
  qualification: string;
  birthDate: string;
  status: boolean;
}

interface PatientAdmissionRow {
  id: string;
  admissionId: string;
  patientName: string;
  patientEmail: string;
  doctorName: string;
  doctorEmail: string;
  admissionDate: string;
  admissionTime: string;
  dischargeDate: string;
  packageName: string;
  insuranceName: string;
  policyNo: string;
  status: boolean;
}

export function PatientsWorkspace({ id = "patients" }: PatientsWorkspaceProps) {
  const { t } = useLanguage();
  // Normalize id so /patients maps to "patients" tab
  const currentTab = id === "patients" || id === "" ? "patients" : id;

  // View modes for each tab: "list" or "create"
  const [patientMode, setPatientMode] = useState<"list" | "create">("list");
  const [caseMode, setCaseMode] = useState<"list" | "create">("list");
  const [handlerMode, setHandlerMode] = useState<"list" | "create">("list");
  const [admissionMode, setAdmissionMode] = useState<"list" | "create">("list");

  // Search & Pagination
  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);

  // Subtabs matching screenshots
  const tabs = [
    { id: "patients", label: "Patients", href: "/modules/patients" },
    { id: "patient-cases", label: "Cases", href: "/modules/patient-cases" },
    {
      id: "case-handlers",
      label: "Case Handlers",
      href: "/modules/case-handlers",
    },
    {
      id: "patient-admissions",
      label: "Patient Admissions",
      href: "/modules/patient-admissions",
    },
  ];

  /* -------------------------------------------------------------
     1. PATIENTS STATE (SCREENSHOT 183748)
     ------------------------------------------------------------- */
  const [patients, setPatients] = useState<PatientRow[]>([
    {
      id: "P-1",
      name: "Ifeanyichukwu Okoro",
      email: "stifeanyi112@googlemail.com",
      phone: "N/A",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-2",
      name: "SAN K",
      email: "mwpsquade@gmail.com",
      phone: "9080275121",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-3",
      name: "Designer 1",
      email: "d2454354@gmail.com",
      phone: "9080275121",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-4",
      name: "SANTHOSH S",
      email: "22uca105@muthayammal.in",
      phone: "9080275121",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-5",
      name: "Santhosh Kumar",
      email: "santhoshkumarat2004@gmail.com",
      phone: "9080275121",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-6",
      name: "Mari V",
      email: "vmhospitalnkm@gmail.com",
      phone: "8870342172",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-7",
      name: "Sa Fs",
      email: "alehsaan.it@gmail.com",
      phone: "96265684",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-8",
      name: "Kato Kato",
      email: "kato@kato.com",
      phone: "+96550012367",
      bloodGroup: "A",
      status: true,
    },
    {
      id: "P-9",
      name: "Shdjsjsj Sjsjssjj",
      email: "gagsmnzjskakan123@gmail.com",
      phone: "08037367680",
      bloodGroup: "N/A",
      status: true,
    },
    {
      id: "P-10",
      name: "MM Mmm",
      email: "mm@gmail.com",
      phone: "+3519253417748",
      bloodGroup: "45",
      status: true,
    },
  ]);

  // Modal: Change Password / Credentials (yellow key icon)
  const [keyModalPatient, setKeyModalPatient] = useState<PatientRow | null>(
    null,
  );
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  // New Patient Form State (Screenshots 183834 & 183855)
  const [pFirstName, setPFirstName] = useState("");
  const [pLastName, setPLastName] = useState("");
  const [pEmail, setPEmail] = useState("");
  const [pDob, setPDob] = useState("");
  const [pPhone, setPPhone] = useState("");
  const [pGender, setPGender] = useState<"Male" | "Female">("Male");
  const [pStatus, setPStatus] = useState(true);
  const [pBloodGroup, setPBloodGroup] = useState("");
  const [pPass, setPPass] = useState("");
  const [pPassConfirm, setPPassConfirm] = useState("");
  const [pAge, setPAge] = useState("");
  const [pReferer, setPReferer] = useState("");
  const [pFechaImplante, setPFechaImplante] = useState("");
  const [pAddr1, setPAddr1] = useState("");
  const [pAddr2, setPAddr2] = useState("");
  const [pCity, setPCity] = useState("");
  const [pZip, setPZip] = useState("");
  const [pFacebook, setPFacebook] = useState("");
  const [pTwitter, setPTwitter] = useState("");
  const [pInstagram, setPInstagram] = useState("");
  const [pLinkedIn, setPLinkedIn] = useState("");

  /* -------------------------------------------------------------
     2. CASES STATE (SCREENSHOT 183958)
     ------------------------------------------------------------- */
  const [cases, setCases] = useState<CaseRow[]>([
    {
      id: "C-1",
      caseId: "HMS16",
      patientName: "11111 1111",
      patientEmail: "sulapojim@mailinator.com",
      doctorName: "123456 123456",
      doctorEmail: "123456@gmail.com",
      caseDate: "30th September 2026",
      caseTime: "12:00 PM",
      fee: 344334,
      status: true,
    },
    {
      id: "C-2",
      caseId: "HMS15",
      patientName: "123 123",
      patientEmail: "oshannimesh@gmail.com",
      doctorName: "Ahmed Doctor",
      doctorEmail: "doctorahmed@gmail.com",
      caseDate: "27th September 2026",
      caseTime: "12:00 PM",
      fee: 50,
      status: true,
    },
    {
      id: "C-3",
      caseId: "HMS14",
      patientName: "MM Mmm",
      patientEmail: "mm@gmail.com",
      doctorName: "AAAAA BBBBB",
      doctorEmail: "abab8764350@gmail.com",
      caseDate: "21st September 2026",
      caseTime: "12:00 PM",
      fee: 60000,
      status: true,
    },
    {
      id: "C-4",
      caseId: "HMS13",
      patientName: "Vinay Grover",
      patientEmail: "vinaygrover14365@gmail.com",
      doctorName: "Branly De León",
      doctorEmail: "server@mango.com.gt",
      caseDate: "5th August 2026",
      caseTime: "12:00 PM",
      fee: 2580,
      status: true,
    },
    {
      id: "C-5",
      caseId: "HMS12",
      patientName: "Aljun Cardona",
      patientEmail: "cardona.aljun@gmail.com",
      doctorName: "Aremu Lanre",
      doctorEmail: "aremu@google.com",
      caseDate: "23rd July 2026",
      caseTime: "12:00 PM",
      fee: 500,
      status: true,
    },
    {
      id: "C-6",
      caseId: "HMS11",
      patientName: "123 123",
      patientEmail: "oshannimesh@gmail.com",
      doctorName: "AAAAA BBBBB",
      doctorEmail: "abab8764350@gmail.com",
      caseDate: "14th June 2026",
      caseTime: "12:00 PM",
      fee: 10,
      status: true,
    },
    {
      id: "C-7",
      caseId: "HMS10",
      patientName: "AA SS",
      patientEmail: "ss@ss.com",
      doctorName: "Abdiqafar Haaaa",
      doctorEmail: "abdi@gmail.com",
      caseDate: "13th June 2026",
      caseTime: "5:00 PM",
      fee: 200,
      status: true,
    },
    {
      id: "C-8",
      caseId: "HMS09",
      patientName: "Jyothi Laxmi",
      patientEmail: "jyothi@gmail.com",
      doctorName: "123456 123456",
      doctorEmail: "123456@gmail.com",
      caseDate: "11th June 2026",
      caseTime: "12:00 PM",
      fee: 49,
      status: false,
    },
    {
      id: "C-9",
      caseId: "HMS08",
      patientName: "Sdfsdf Sdfsdf",
      patientEmail: "laxmi@gmail.com",
      doctorName: "Ankit Patra",
      doctorEmail: "ankit@gmail.com",
      caseDate: "9th June 2026",
      caseTime: "12:00 PM",
      fee: 500,
      status: true,
    },
    {
      id: "C-10",
      caseId: "HMS07",
      patientName: "123 123",
      patientEmail: "oshannimesh@gmail.com",
      doctorName: "AARAV Singh",
      doctorEmail: "aarav@gmail.com",
      caseDate: "9th June 2026",
      caseTime: "12:00 PM",
      fee: 2000,
      status: true,
    },
  ]);

  // New Case Form State (Screenshot 184023)
  const [casePatient, setCasePatient] = useState("");
  const [caseDoctor, setCaseDoctor] = useState("");
  const [caseDate, setCaseDate] = useState("");
  const [casePhone, setCasePhone] = useState("");
  const [caseStatus, setCaseStatus] = useState(true);
  const [caseFee, setCaseFee] = useState("");
  const [caseDesc, setCaseDesc] = useState("");

  /* -------------------------------------------------------------
     3. CASE HANDLERS STATE (SCREENSHOT 184049)
     ------------------------------------------------------------- */
  const [caseHandlers, setCaseHandlers] = useState<CaseHandlerRow[]>([
    {
      id: "CH-1",
      name: "Ajay22 Makwana",
      email: "ajay@gmail.com",
      phone: "N/A",
      qualification: "LLB",
      birthDate: "N/A",
      status: true,
    },
    {
      id: "CH-2",
      name: "Shaliesh Ladhumai",
      email: "mailtestingservice12@gmail.com",
      phone: "+9779816571299",
      qualification: "General",
      birthDate: "19th Oct, 2021",
      status: true,
    },
    {
      id: "CH-3",
      name: "Kumar K",
      email: "casetest@hms.com",
      phone: "N/A",
      qualification: "B.SC IT",
      birthDate: "N/A",
      status: true,
    },
    {
      id: "CH-4",
      name: "Antony M",
      email: "jysa@mailinator.com",
      phone: "+441466768878",
      qualification: "In ea error voluptas",
      birthDate: "N/A",
      status: true,
    },
    {
      id: "CH-5",
      name: "Deirdre Ball",
      email: "deirdre.ball@mailtrap.com",
      phone: "+919874563211",
      qualification: "Social Worker",
      birthDate: "N/A",
      status: true,
    },
    {
      id: "CH-6",
      name: "Claire White",
      email: "claire.white@mailtrap.com",
      phone: "+919632587410",
      qualification: "qwe",
      birthDate: "1st Jul, 2024",
      status: true,
    },
    {
      id: "CH-7",
      name: "PRANAV BLABLA",
      email: "p@p.com",
      phone: "+915654852531",
      qualification: "MSW",
      birthDate: "10th Dec, 2024",
      status: true,
    },
    {
      id: "CH-8",
      name: "Karan Singh",
      email: "karn@g.com",
      phone: "+6254568564",
      qualification: "MBS",
      birthDate: "N/A",
      status: true,
    },
  ]);

  // New Case Handler Form State (Screenshot 184123)
  const [chFirst, setChFirst] = useState("");
  const [chLast, setChLast] = useState("");
  const [chEmail, setChEmail] = useState("");
  const [chDesignation, setChDesignation] = useState("");
  const [chPhone, setChPhone] = useState("");
  const [chGender, setChGender] = useState<"Male" | "Female">("Male");
  const [chStatus, setChStatus] = useState(true);
  const [chQualification, setChQualification] = useState("");
  const [chDob, setChDob] = useState("");
  const [chBlood, setChBlood] = useState("");
  const [chPass, setChPass] = useState("");
  const [chPassConfirm, setChPassConfirm] = useState("");
  const [chAddr1, setChAddr1] = useState("");
  const [chAddr2, setChAddr2] = useState("");
  const [chCity, setChCity] = useState("");
  const [chZip, setChZip] = useState("");

  /* -------------------------------------------------------------
     4. PATIENT ADMISSIONS STATE (SCREENSHOT 184152)
     ------------------------------------------------------------- */
  const [admissions, setAdmissions] = useState<PatientAdmissionRow[]>([
    {
      id: "ADM-1",
      admissionId: "HMS21",
      patientName: "SAN K",
      patientEmail: "mwpsquade@gmail.com",
      doctorName: "Dharman K",
      doctorEmail: "smartentry8@gmail.com",
      admissionDate: "5th Oct, 2026",
      admissionTime: "12:00 PM",
      dischargeDate: "N/A",
      packageName: "N/A",
      insuranceName: "N/A",
      policyNo: "N/A",
      status: true,
    },
    {
      id: "ADM-2",
      admissionId: "HMS20",
      patientName: "SAN K",
      patientEmail: "mwpsquade@gmail.com",
      doctorName: "Dharman K",
      doctorEmail: "smartentry8@gmail.com",
      admissionDate: "5th Oct, 2026",
      admissionTime: "12:00 PM",
      dischargeDate: "N/A",
      packageName: "Checkup",
      insuranceName: "BAJAJ",
      policyNo: "1212",
      status: true,
    },
    {
      id: "ADM-3",
      admissionId: "HMS19",
      patientName: "MM Mmm",
      patientEmail: "mm@gmail.com",
      doctorName: "123456 123456",
      doctorEmail: "123456@gmail.com",
      admissionDate: "20th Sep, 2026",
      admissionTime: "12:00 AM",
      dischargeDate: "N/A",
      packageName: "All in 1",
      insuranceName: "BAJAJ",
      policyNo: "N/A",
      status: true,
    },
    {
      id: "ADM-4",
      admissionId: "HMS18",
      patientName: "VIJAY BHAI BHIMANI",
      patientEmail: "vijay@casavirtue.co.uk",
      doctorName: "Ahmed Doctor",
      doctorEmail: "doctorahmed@gmail.com",
      admissionDate: "12th Sep, 2026",
      admissionTime: "12:00 AM",
      dischargeDate: "N/A",
      packageName: "Checkup",
      insuranceName: "BAJAJ",
      policyNo: "N/A",
      status: true,
    },
    {
      id: "ADM-5",
      admissionId: "HMS17",
      patientName: "Srinivas D",
      patientEmail: "srinivas@gmail.com",
      doctorName: "AARAV Singh",
      doctorEmail: "aarav@gmail.com",
      admissionDate: "8th Sep, 2026",
      admissionTime: "12:00 AM",
      dischargeDate: "N/A",
      packageName: "All in 1",
      insuranceName: "N/A",
      policyNo: "N/A",
      status: true,
    },
    {
      id: "ADM-6",
      admissionId: "HMS16",
      patientName: "A B",
      patientEmail: "ab@gmail.com",
      doctorName: "Albert Yano",
      doctorEmail: "yano@gmail.com",
      admissionDate: "22nd Aug, 2026",
      admissionTime: "12:00 PM",
      dischargeDate: "N/A",
      packageName: "dental",
      insuranceName: "Zelda Walls",
      policyNo: "12",
      status: true,
    },
    {
      id: "ADM-7",
      admissionId: "HMS15",
      patientName: "Vinay Grover",
      patientEmail: "vinaygrover14365@gmail.com",
      doctorName: "Ankit Patra",
      doctorEmail: "ankit@gmail.com",
      admissionDate: "5th Aug, 2026",
      admissionTime: "12:00 AM",
      dischargeDate: "N/A",
      packageName: "All in 1",
      insuranceName: "N/A",
      policyNo: "N/A",
      status: true,
    },
    {
      id: "ADM-8",
      admissionId: "HMS14",
      patientName: "Aljun Cardona",
      patientEmail: "cardona.aljun@gmail.com",
      doctorName: "123456 123456",
      doctorEmail: "123456@gmail.com",
      admissionDate: "23rd Jul, 2026",
      admissionTime: "12:00 AM",
      dischargeDate: "N/A",
      packageName: "Fever Package",
      insuranceName: "Brooke Leblan",
      policyNo: "2323",
      status: true,
    },
    {
      id: "ADM-9",
      admissionId: "HMS13",
      patientName: "11111 11111",
      patientEmail: "11111@qwe.com",
      doctorName: "123456 123456",
      doctorEmail: "123456@gmail.com",
      admissionDate: "30th Jun, 2026",
      admissionTime: "12:00 PM",
      dischargeDate: "N/A",
      packageName: "Daat Test",
      insuranceName: "N/A",
      policyNo: "LIU",
      status: true,
    },
    {
      id: "ADM-10",
      admissionId: "HMS12",
      patientName: "11111 1111",
      patientEmail: "sulapojim@mailinator.com",
      doctorName: "123456 123456",
      doctorEmail: "123456@gmail.com",
      admissionDate: "24th Jun, 2026",
      admissionTime: "12:00 PM",
      dischargeDate: "N/A",
      packageName: "All in 1",
      insuranceName: "BAJAJ",
      policyNo: "N/A",
      status: true,
    },
  ]);

  // New Admission Form State (Screenshot 184244)
  const [admPatient, setAdmPatient] = useState("");
  const [admDoctor, setAdmDoctor] = useState("");
  const [admDate, setAdmDate] = useState("");
  const [admPackage, setAdmPackage] = useState("");
  const [admInsurance, setAdmInsurance] = useState("");
  const [admBed, setAdmBed] = useState("");
  const [admPolicyNo, setAdmPolicyNo] = useState("");
  const [admAgent, setAdmAgent] = useState("");
  const [admGuardianName, setAdmGuardianName] = useState("");
  const [admGuardianRelation, setAdmGuardianRelation] = useState("");
  const [admGuardianContact, setAdmGuardianContact] = useState("");
  const [admGuardianAddress, setAdmGuardianAddress] = useState("");
  const [admStatus, setAdmStatus] = useState(true);

  /* -------------------------------------------------------------
     HANDLERS FOR SAVING FORMS
     ------------------------------------------------------------- */
  function handleSavePatient(e: React.FormEvent) {
    e.preventDefault();
    if (!pFirstName || !pLastName) return;
    const fullName = `${pFirstName} ${pLastName}`;
    const newP: PatientRow = {
      id: `P-${patients.length + 1}`,
      name: fullName,
      email: pEmail || `${pFirstName.toLowerCase()}@example.com`,
      phone: pPhone || "N/A",
      bloodGroup: pBloodGroup || "N/A",
      status: pStatus,
    };
    setPatients([newP, ...patients]);
    setPatientMode("list");
  }

  function handleSaveCase(e: React.FormEvent) {
    e.preventDefault();
    if (!casePatient || !caseDoctor) return;
    const newC: CaseRow = {
      id: `C-${cases.length + 1}`,
      caseId: `HMS${String(cases.length + 17)}`,
      patientName: casePatient,
      patientEmail: "patient@example.com",
      doctorName: caseDoctor,
      doctorEmail: "doctor@example.com",
      caseDate: caseDate || "05th Oct 2026",
      caseTime: "12:00 PM",
      fee: parseFloat(caseFee) || 100,
      status: caseStatus,
    };
    setCases([newC, ...cases]);
    setCaseMode("list");
  }

  function handleSaveCaseHandler(e: React.FormEvent) {
    e.preventDefault();
    if (!chFirst || !chLast) return;
    const fullName = `${chFirst} ${chLast}`;
    const newCH: CaseHandlerRow = {
      id: `CH-${caseHandlers.length + 1}`,
      name: fullName,
      email: chEmail || `${chFirst.toLowerCase()}@example.com`,
      phone: chPhone || "N/A",
      qualification: chQualification || "General",
      birthDate: chDob || "N/A",
      status: chStatus,
    };
    setCaseHandlers([newCH, ...caseHandlers]);
    setHandlerMode("list");
  }

  function handleSaveAdmission(e: React.FormEvent) {
    e.preventDefault();
    if (!admPatient || !admDoctor) return;
    const newAdm: PatientAdmissionRow = {
      id: `ADM-${admissions.length + 1}`,
      admissionId: `HMS${String(admissions.length + 22)}`,
      patientName: admPatient,
      patientEmail: "patient@example.com",
      doctorName: admDoctor,
      doctorEmail: "doctor@example.com",
      admissionDate: admDate || "05th Oct, 2026",
      admissionTime: "12:00 PM",
      dischargeDate: "N/A",
      packageName: admPackage || "N/A",
      insuranceName: admInsurance || "N/A",
      policyNo: admPolicyNo || "N/A",
      status: admStatus,
    };
    setAdmissions([newAdm, ...admissions]);
    setAdmissionMode("list");
  }

  /* -------------------------------------------------------------
     FULL-PAGE FORM: NEW PATIENT (SCREENSHOTS 183834 & 183855)
     ------------------------------------------------------------- */
  if (currentTab === "patients" && patientMode === "create") {
    return (
      <div className="legacy-workspace">
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
            <h2 className="workspace-heading m-0">{t("New Patient")}</h2>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => setPatientMode("list")}
            >
              {t("Back")}
            </button>
          </div>

          <form onSubmit={handleSavePatient}>
            <div className="form-grid-2">
              <div className="form-group-custom">
                <label>
                  {t("First Name")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("First Name")}
                  value={pFirstName}
                  onChange={(e) => setPFirstName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Last Name")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Last Name")}
                  value={pLastName}
                  onChange={(e) => setPLastName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Email")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="email"
                  required
                  placeholder={t("Email")}
                  value={pEmail}
                  onChange={(e) => setPEmail(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Date Of Birth")}:</label>
                <input
                  type="date"
                  value={pDob}
                  onChange={(e) => setPDob(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Phone")}: <span className="text-danger">*</span>
                </label>
                <div className="d-flex gap-2">
                  <span className="phone-prefix-flag">🇰🇼 +965</span>
                  <input
                    type="text"
                    required
                    placeholder="500 12345"
                    value={pPhone}
                    onChange={(e) => setPPhone(e.target.value)}
                  />
                </div>
              </div>

              <div className="form-group-custom d-flex gap-4">
                <div>
                  <label>
                    {t("Gender")}: <span className="text-danger">*</span>
                  </label>
                  <div className="d-flex gap-3 mt-2">
                    <label className="d-flex align-items-center gap-1 cursor-pointer">
                      <input
                        type="radio"
                        name="gender"
                        checked={pGender === "Male"}
                        onChange={() => setPGender("Male")}
                      />
                      <span>{t("Male")}</span>
                    </label>
                    <label className="d-flex align-items-center gap-1 cursor-pointer">
                      <input
                        type="radio"
                        name="gender"
                        checked={pGender === "Female"}
                        onChange={() => setPGender("Female")}
                      />
                      <span>{t("Female")}</span>
                    </label>
                  </div>
                </div>

                <div>
                  <label>{t("Status")}:</label>
                  <div className="mt-2">
                    <label className="switch-toggle">
                      <input
                        type="checkbox"
                        checked={pStatus}
                        onChange={(e) => setPStatus(e.target.checked)}
                      />
                      <span className="slider-toggle"></span>
                    </label>
                  </div>
                </div>
              </div>

              <div className="form-group-custom">
                <label>{t("Blood Group")}:</label>
                <select
                  className="form-select-custom w-100"
                  value={pBloodGroup}
                  onChange={(e) => setPBloodGroup(e.target.value)}
                >
                  <option value="">{t("Select Blood Group")}</option>
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

              <div className="form-group-custom">
                <label>
                  {t("Password")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="password"
                  required
                  placeholder={t("Password")}
                  value={pPass}
                  onChange={(e) => setPPass(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Confirm Password")}:{" "}
                  <span className="text-danger">*</span>
                </label>
                <input
                  type="password"
                  required
                  placeholder={t("Confirm Password")}
                  value={pPassConfirm}
                  onChange={(e) => setPPassConfirm(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Age-Age")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder="age-age"
                  value={pAge}
                  onChange={(e) => setPAge(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Referer")}:</label>
                <input
                  type="text"
                  placeholder={t("Referer")}
                  value={pReferer}
                  onChange={(e) => setPReferer(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Fecha De Implante")}:{" "}
                  <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder="fecha de implante"
                  value={pFechaImplante}
                  onChange={(e) => setPFechaImplante(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Profile")}:</label>
                <div className="avatar-uploader-container mt-2">
                  <div className="avatar-preview-box">
                    <img
                      src="/brand-icons/avatar-default.svg"
                      alt="Avatar"
                      onError={(e) => {
                        e.currentTarget.style.display = "none";
                      }}
                    />
                    <div className="avatar-edit-badge">
                      <Edit2 size={14} />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Address Details */}
            <div className="mt-4 pt-3 border-top border-secondary-subtle">
              <h4 className="fs-5 fw-semibold mb-3">{t("Address Details")}</h4>
              <div className="form-grid-2">
                <div className="form-group-custom">
                  <label>{t("Address 1")}:</label>
                  <input
                    type="text"
                    placeholder={t("Address 1")}
                    value={pAddr1}
                    onChange={(e) => setPAddr1(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("Address 2")}:</label>
                  <input
                    type="text"
                    placeholder={t("Address 2")}
                    value={pAddr2}
                    onChange={(e) => setPAddr2(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("City")}:</label>
                  <input
                    type="text"
                    placeholder={t("City")}
                    value={pCity}
                    onChange={(e) => setPCity(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("Zipcode")}:</label>
                  <input
                    type="text"
                    placeholder={t("Zipcode")}
                    value={pZip}
                    onChange={(e) => setPZip(e.target.value)}
                  />
                </div>
              </div>
            </div>

            {/* Social Details */}
            <div className="mt-4 pt-3 border-top border-secondary-subtle">
              <h4 className="fs-5 fw-semibold mb-3">{t("Social Details")}</h4>
              <div className="form-grid-2">
                <div className="form-group-custom">
                  <label>{t("Facebook URL")}:</label>
                  <input
                    type="text"
                    placeholder={t("Facebook URL")}
                    value={pFacebook}
                    onChange={(e) => setPFacebook(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("Twitter URL")}:</label>
                  <input
                    type="text"
                    placeholder={t("Twitter URL")}
                    value={pTwitter}
                    onChange={(e) => setPTwitter(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("Instagram URL")}:</label>
                  <input
                    type="text"
                    placeholder={t("Instagram URL")}
                    value={pInstagram}
                    onChange={(e) => setPInstagram(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("LinkedIn URL")}:</label>
                  <input
                    type="text"
                    placeholder={t("LinkedIn URL")}
                    value={pLinkedIn}
                    onChange={(e) => setPLinkedIn(e.target.value)}
                  />
                </div>
              </div>
            </div>

            <div className="d-flex justify-content-end gap-2 mt-4">
              <button type="submit" className="btn-action-blue px-4 py-2">
                {t("Save")}
              </button>
              <button
                type="button"
                className="btn-action-grey px-4 py-2"
                onClick={() => setPatientMode("list")}
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
     FULL-PAGE FORM: NEW CASE (SCREENSHOT 184023)
     ------------------------------------------------------------- */
  if (currentTab === "patient-cases" && caseMode === "create") {
    return (
      <div className="legacy-workspace">
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
            <h2 className="workspace-heading m-0">{t("New Case")}</h2>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => setCaseMode("list")}
            >
              {t("Back")}
            </button>
          </div>

          <form onSubmit={handleSaveCase}>
            <div className="form-grid-2">
              <div className="form-group-custom">
                <label>
                  {t("Patient")}: <span className="text-danger">*</span>
                </label>
                <select
                  className="form-select-custom w-100"
                  required
                  value={casePatient}
                  onChange={(e) => setCasePatient(e.target.value)}
                >
                  <option value="">{t("Select Patient")}</option>
                  {patients.map((p) => (
                    <option key={p.id} value={p.name}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Doctor")}: <span className="text-danger">*</span>
                </label>
                <select
                  className="form-select-custom w-100"
                  required
                  value={caseDoctor}
                  onChange={(e) => setCaseDoctor(e.target.value)}
                >
                  <option value="">{t("Select Doctor")}</option>
                  <option value="Ahmed Doctor">Ahmed Doctor</option>
                  <option value="AARAV Singh">AARAV Singh</option>
                  <option value="Harish Mohan">Harish Mohan</option>
                  <option value="Albert Yano">Albert Yano</option>
                  <option value="Ankit Patra">Ankit Patra</option>
                </select>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Case Date")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="date"
                  required
                  value={caseDate}
                  onChange={(e) => setCaseDate(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Phone")}:</label>
                <div className="d-flex gap-2">
                  <span className="phone-prefix-flag">🇰🇼 +965</span>
                  <input
                    type="text"
                    placeholder="500 12345"
                    value={casePhone}
                    onChange={(e) => setCasePhone(e.target.value)}
                  />
                </div>
              </div>

              <div className="form-group-custom">
                <label>{t("Status")}:</label>
                <div className="mt-2">
                  <label className="switch-toggle">
                    <input
                      type="checkbox"
                      checked={caseStatus}
                      onChange={(e) => setCaseStatus(e.target.checked)}
                    />
                    <span className="slider-toggle"></span>
                  </label>
                </div>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Fee")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="number"
                  step="0.01"
                  required
                  placeholder={t("Fee")}
                  value={caseFee}
                  onChange={(e) => setCaseFee(e.target.value)}
                />
              </div>
            </div>

            <div className="form-group-custom mt-3">
              <label>{t("Description")}:</label>
              <textarea
                rows={4}
                placeholder={t("Description")}
                value={caseDesc}
                onChange={(e) => setCaseDesc(e.target.value)}
              />
            </div>

            <div className="d-flex justify-content-end gap-2 mt-4">
              <button type="submit" className="btn-action-blue px-4 py-2">
                {t("Save")}
              </button>
              <button
                type="button"
                className="btn-action-grey px-4 py-2"
                onClick={() => setCaseMode("list")}
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
     FULL-PAGE FORM: NEW CASE HANDLER (SCREENSHOT 184123)
     ------------------------------------------------------------- */
  if (currentTab === "case-handlers" && handlerMode === "create") {
    return (
      <div className="legacy-workspace">
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
            <h2 className="workspace-heading m-0">{t("New Case Handler")}</h2>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => setHandlerMode("list")}
            >
              {t("Back")}
            </button>
          </div>

          <form onSubmit={handleSaveCaseHandler}>
            <div className="form-grid-2">
              <div className="form-group-custom">
                <label>
                  {t("First Name")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("First Name")}
                  value={chFirst}
                  onChange={(e) => setChFirst(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Last Name")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Last Name")}
                  value={chLast}
                  onChange={(e) => setChLast(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Email")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="email"
                  required
                  placeholder={t("Email")}
                  value={chEmail}
                  onChange={(e) => setChEmail(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Designation")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Designation")}
                  value={chDesignation}
                  onChange={(e) => setChDesignation(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Phone")}:</label>
                <div className="d-flex gap-2">
                  <span className="phone-prefix-flag">🇰🇼 +965</span>
                  <input
                    type="text"
                    placeholder="500 12345"
                    value={chPhone}
                    onChange={(e) => setChPhone(e.target.value)}
                  />
                </div>
              </div>

              <div className="form-group-custom d-flex gap-4">
                <div>
                  <label>
                    {t("Gender")}: <span className="text-danger">*</span>
                  </label>
                  <div className="d-flex gap-3 mt-2">
                    <label className="d-flex align-items-center gap-1 cursor-pointer">
                      <input
                        type="radio"
                        name="chGender"
                        checked={chGender === "Male"}
                        onChange={() => setChGender("Male")}
                      />
                      <span>{t("Male")}</span>
                    </label>
                    <label className="d-flex align-items-center gap-1 cursor-pointer">
                      <input
                        type="radio"
                        name="chGender"
                        checked={chGender === "Female"}
                        onChange={() => setChGender("Female")}
                      />
                      <span>{t("Female")}</span>
                    </label>
                  </div>
                </div>

                <div>
                  <label>{t("Status")}:</label>
                  <div className="mt-2">
                    <label className="switch-toggle">
                      <input
                        type="checkbox"
                        checked={chStatus}
                        onChange={(e) => setChStatus(e.target.checked)}
                      />
                      <span className="slider-toggle"></span>
                    </label>
                  </div>
                </div>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Qualification")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Qualification")}
                  value={chQualification}
                  onChange={(e) => setChQualification(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Date Of Birth")}:</label>
                <input
                  type="date"
                  value={chDob}
                  onChange={(e) => setChDob(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Blood Group")}:</label>
                <select
                  className="form-select-custom w-100"
                  value={chBlood}
                  onChange={(e) => setChBlood(e.target.value)}
                >
                  <option value="">{t("Select Blood Group")}</option>
                  <option value="A+">A+</option>
                  <option value="B+">B+</option>
                  <option value="AB+">AB+</option>
                  <option value="O+">O+</option>
                </select>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Password")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="password"
                  required
                  placeholder={t("Password")}
                  value={chPass}
                  onChange={(e) => setChPass(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Confirm Password")}:{" "}
                  <span className="text-danger">*</span>
                </label>
                <input
                  type="password"
                  required
                  placeholder={t("Confirm Password")}
                  value={chPassConfirm}
                  onChange={(e) => setChPassConfirm(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Profile")}:</label>
                <div className="avatar-uploader-container mt-2">
                  <div className="avatar-preview-box">
                    <img
                      src="/brand-icons/avatar-default.svg"
                      alt="Avatar"
                      onError={(e) => {
                        e.currentTarget.style.display = "none";
                      }}
                    />
                    <div className="avatar-edit-badge">
                      <Edit2 size={14} />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Address Details */}
            <div className="mt-4 pt-3 border-top border-secondary-subtle">
              <h4 className="fs-5 fw-semibold mb-3">{t("Address Details")}</h4>
              <div className="form-grid-2">
                <div className="form-group-custom">
                  <label>{t("Address 1")}:</label>
                  <input
                    type="text"
                    placeholder={t("Address 1")}
                    value={chAddr1}
                    onChange={(e) => setChAddr1(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("Address 2")}:</label>
                  <input
                    type="text"
                    placeholder={t("Address 2")}
                    value={chAddr2}
                    onChange={(e) => setChAddr2(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("City")}:</label>
                  <input
                    type="text"
                    placeholder={t("City")}
                    value={chCity}
                    onChange={(e) => setChCity(e.target.value)}
                  />
                </div>
                <div className="form-group-custom">
                  <label>{t("Zipcode")}:</label>
                  <input
                    type="text"
                    placeholder={t("Zipcode")}
                    value={chZip}
                    onChange={(e) => setChZip(e.target.value)}
                  />
                </div>
              </div>
            </div>

            <div className="d-flex justify-content-end gap-2 mt-4">
              <button type="submit" className="btn-action-blue px-4 py-2">
                {t("Save")}
              </button>
              <button
                type="button"
                className="btn-action-grey px-4 py-2"
                onClick={() => setHandlerMode("list")}
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
     FULL-PAGE FORM: NEW PATIENT ADMISSION (SCREENSHOT 184244)
     ------------------------------------------------------------- */
  if (currentTab === "patient-admissions" && admissionMode === "create") {
    return (
      <div className="legacy-workspace">
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
            <h2 className="workspace-heading m-0">
              {t("New Patient Admission")}
            </h2>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => setAdmissionMode("list")}
            >
              {t("Back")}
            </button>
          </div>

          <form onSubmit={handleSaveAdmission}>
            <div className="form-grid-2">
              <div className="form-group-custom">
                <label>
                  {t("Patient")}: <span className="text-danger">*</span>
                </label>
                <select
                  className="form-select-custom w-100"
                  required
                  value={admPatient}
                  onChange={(e) => setAdmPatient(e.target.value)}
                >
                  <option value="">{t("Select Patient")}</option>
                  {patients.map((p) => (
                    <option key={p.id} value={p.name}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Doctor")}: <span className="text-danger">*</span>
                </label>
                <select
                  className="form-select-custom w-100"
                  required
                  value={admDoctor}
                  onChange={(e) => setAdmDoctor(e.target.value)}
                >
                  <option value="">{t("Select Doctor")}</option>
                  <option value="Dharman K">Dharman K</option>
                  <option value="Ahmed Doctor">Ahmed Doctor</option>
                  <option value="AARAV Singh">AARAV Singh</option>
                  <option value="Albert Yano">Albert Yano</option>
                  <option value="Ankit Patra">Ankit Patra</option>
                </select>
              </div>

              <div className="form-group-custom">
                <label>
                  {t("Admission Date")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="date"
                  required
                  value={admDate}
                  onChange={(e) => setAdmDate(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Package")}:</label>
                <select
                  className="form-select-custom w-100"
                  value={admPackage}
                  onChange={(e) => setAdmPackage(e.target.value)}
                >
                  <option value="">{t("Choose Package")}</option>
                  <option value="Checkup">Checkup</option>
                  <option value="All in 1">All in 1</option>
                  <option value="dental">dental</option>
                  <option value="Fever Package">Fever Package</option>
                  <option value="Daat Test">Daat Test</option>
                </select>
              </div>

              <div className="form-group-custom">
                <label>{t("Insurance")}:</label>
                <select
                  className="form-select-custom w-100"
                  value={admInsurance}
                  onChange={(e) => setAdmInsurance(e.target.value)}
                >
                  <option value="">{t("Choose Insurance")}</option>
                  <option value="BAJAJ">BAJAJ</option>
                  <option value="Zelda Walls">Zelda Walls</option>
                  <option value="Brooke Leblan">Brooke Leblan</option>
                </select>
              </div>

              <div className="form-group-custom">
                <label>{t("Bed")}:</label>
                <select
                  className="form-select-custom w-100"
                  value={admBed}
                  onChange={(e) => setAdmBed(e.target.value)}
                >
                  <option value="">{t("Choose Bed")}</option>
                  <option value="ICU Bed 1">ICU Bed 1</option>
                  <option value="NICU Bed 2">NICU Bed 2</option>
                  <option value="VIP Suite 101">VIP Suite 101</option>
                  <option value="Private 204">Private 204</option>
                  <option value="General Ward 05">General Ward 05</option>
                </select>
              </div>

              <div className="form-group-custom">
                <label>{t("Policy No")}:</label>
                <input
                  type="text"
                  placeholder={t("Policy No")}
                  value={admPolicyNo}
                  onChange={(e) => setAdmPolicyNo(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Agent Name")}:</label>
                <input
                  type="text"
                  placeholder={t("Agent Name")}
                  value={admAgent}
                  onChange={(e) => setAdmAgent(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Guardian Name")}:</label>
                <input
                  type="text"
                  placeholder={t("Guardian Name")}
                  value={admGuardianName}
                  onChange={(e) => setAdmGuardianName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Guardian Relation")}:</label>
                <input
                  type="text"
                  placeholder={t("Guardian Relation")}
                  value={admGuardianRelation}
                  onChange={(e) => setAdmGuardianRelation(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Guardian Contact")}:</label>
                <div className="d-flex gap-2">
                  <span className="phone-prefix-flag">🇰🇼 +965</span>
                  <input
                    type="text"
                    placeholder="500 12345"
                    value={admGuardianContact}
                    onChange={(e) => setAdmGuardianContact(e.target.value)}
                  />
                </div>
              </div>

              <div className="form-group-custom">
                <label>{t("Guardian Address")}:</label>
                <input
                  type="text"
                  placeholder={t("Guardian Address")}
                  value={admGuardianAddress}
                  onChange={(e) => setAdmGuardianAddress(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label>{t("Status")}:</label>
                <div className="mt-2">
                  <label className="switch-toggle">
                    <input
                      type="checkbox"
                      checked={admStatus}
                      onChange={(e) => setAdmStatus(e.target.checked)}
                    />
                    <span className="slider-toggle"></span>
                  </label>
                </div>
              </div>
            </div>

            <div className="d-flex justify-content-end gap-2 mt-4">
              <button type="submit" className="btn-action-blue px-4 py-2">
                {t("Save")}
              </button>
              <button
                type="button"
                className="btn-action-grey px-4 py-2"
                onClick={() => setAdmissionMode("list")}
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
     RENDER: TABBED LIST VIEW (SCREENSHOTS 183748, 183958, 184049, 184152)
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

      {/* 1. PATIENTS TAB (SCREENSHOT 183748) */}
      {currentTab === "patients" && (
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
              <button className="btn-icon-blue" title={t("Export to Excel")}>
                <FileSpreadsheet size={18} />
              </button>
              <button
                className="btn-action-blue"
                onClick={() => setPatientMode("create")}
              >
                {t("New Patient")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("PATIENTS")} ↕</th>
                  <th>{t("PHONE")} ↕</th>
                  <th>{t("BLOOD GROUP")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {patients
                  .filter((p) =>
                    (p.name + " " + p.email + " " + p.phone)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((p) => {
                    const initials = p.name
                      .split(" ")
                      .map((n) => n[0])
                      .join("")
                      .slice(0, 2)
                      .toUpperCase();
                    return (
                      <tr key={p.id}>
                        <td>
                          <div className="d-flex align-items-center gap-2">
                            <div className="patient-avatar-circle">
                              {initials}
                            </div>
                            <div>
                              <div className="fw-semibold text-primary">
                                {p.name}
                              </div>
                              <div className="text-secondary small">
                                {p.email}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td>{p.phone}</td>
                        <td>
                          {p.bloodGroup !== "N/A" ? (
                            <span className="badge-green">{p.bloodGroup}</span>
                          ) : (
                            <span className="text-secondary">N/A</span>
                          )}
                        </td>
                        <td>
                          <label className="switch-toggle">
                            <input
                              type="checkbox"
                              checked={p.status}
                              onChange={() =>
                                setPatients((prev) =>
                                  prev.map((x) =>
                                    x.id === p.id
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
                              className="btn-yellow-key"
                              title={t("Change Password")}
                              onClick={() => setKeyModalPatient(p)}
                            >
                              <Key size={16} />
                            </button>
                            <button
                              className="btn-icon-blue-link"
                              title={t("Edit")}
                              onClick={() => setPatientMode("create")}
                            >
                              <Edit2 size={16} />
                            </button>
                            <button
                              className="btn-icon-danger"
                              title={t("Delete")}
                              onClick={() =>
                                setPatients((prev) =>
                                  prev.filter((x) => x.id !== p.id),
                                )
                              }
                            >
                              <Trash2 size={16} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
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
                {t("Showing")} 1 {t("to")} 10 {t("of")} 2787 {t("Results")}
              </span>
            </div>
            <div className="pagination-numbers">
              <button className="page-btn active">1</button>
              <button className="page-btn">2</button>
              <button className="page-btn">3</button>
              <button className="page-btn">4</button>
              <button className="page-btn">...</button>
              <button className="page-btn">278</button>
              <button className="page-btn">279</button>
              <button className="page-btn">&gt;</button>
            </div>
          </div>
        </div>
      )}

      {/* 2. CASES TAB (SCREENSHOT 183958) */}
      {currentTab === "patient-cases" && (
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
                onClick={() => setCaseMode("create")}
              >
                {t("New Case")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("CASE ID")} ↕</th>
                  <th>{t("PATIENT")} ↕</th>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("CASE DATE")} ↕</th>
                  <th>{t("FEE")} ↕</th>
                  <th>{t("STATUS")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {cases
                  .filter((c) =>
                    (c.caseId + " " + c.patientName + " " + c.doctorName)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((cs) => (
                    <tr key={cs.id}>
                      <td>
                        <span className="badge-blue-pill">{cs.caseId}</span>
                      </td>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <div className="patient-avatar-circle">
                            {cs.patientName.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="fw-semibold text-primary">
                              {cs.patientName}
                            </div>
                            <div className="text-secondary small">
                              {cs.patientEmail}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <div className="doctor-avatar-circle">
                            {cs.doctorName.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="fw-semibold text-primary">
                              {cs.doctorName}
                            </div>
                            <div className="text-secondary small">
                              {cs.doctorEmail}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="tx-date-badge">
                          <span>{cs.caseTime}</span>
                          <span>{cs.caseDate}</span>
                        </span>
                      </td>
                      <td>${cs.fee.toLocaleString()}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={cs.status}
                            onChange={() =>
                              setCases((prev) =>
                                prev.map((x) =>
                                  x.id === cs.id
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
                              setCases((prev) =>
                                prev.filter((x) => x.id !== cs.id),
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
                {t("Showing")} 1 {t("to")} 10 {t("of")} 386 {t("Results")}
              </span>
            </div>
            <div className="pagination-numbers">
              <button className="page-btn active">1</button>
              <button className="page-btn">2</button>
              <button className="page-btn">3</button>
              <button className="page-btn">4</button>
              <button className="page-btn">...</button>
              <button className="page-btn">38</button>
              <button className="page-btn">39</button>
              <button className="page-btn">&gt;</button>
            </div>
          </div>
        </div>
      )}

      {/* 3. CASE HANDLERS TAB (SCREENSHOT 184049) */}
      {currentTab === "case-handlers" && (
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
              <button className="btn-icon-blue" title={t("Export to Excel")}>
                <FileSpreadsheet size={18} />
              </button>
              <button
                className="btn-action-blue"
                onClick={() => setHandlerMode("create")}
              >
                {t("New Case Handler")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("USERS")} ↕</th>
                  <th>{t("PHONE")} ↕</th>
                  <th>{t("QUALIFICATION")} ↕</th>
                  <th>{t("BIRTH DATE")} ↕</th>
                  <th>{t("STATUS")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {caseHandlers
                  .filter((ch) =>
                    (ch.name + " " + ch.email + " " + ch.qualification)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((ch) => (
                    <tr key={ch.id}>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <div className="patient-avatar-circle">
                            {ch.name.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="fw-semibold text-primary">
                              {ch.name}
                            </div>
                            <div className="text-secondary small">
                              {ch.email}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>{ch.phone}</td>
                      <td>{ch.qualification}</td>
                      <td>{ch.birthDate}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={ch.status}
                            onChange={() =>
                              setCaseHandlers((prev) =>
                                prev.map((x) =>
                                  x.id === ch.id
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
                              setCaseHandlers((prev) =>
                                prev.filter((x) => x.id !== ch.id),
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
              </select>
              <span>
                {t("Showing")} {caseHandlers.length} {t("Results")}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* 4. PATIENT ADMISSIONS TAB (SCREENSHOT 184152) */}
      {currentTab === "patient-admissions" && (
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
              <button className="btn-icon-blue" title={t("Export to Excel")}>
                <FileSpreadsheet size={18} />
              </button>
              <button
                className="btn-action-blue"
                onClick={() => setAdmissionMode("create")}
              >
                {t("New Patient Admission")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("ADMISSION ID")} ↕</th>
                  <th>{t("PATIENT")} ↕</th>
                  <th>{t("DOCTOR")} ↕</th>
                  <th>{t("ADMISSION DATE")} ↕</th>
                  <th>{t("DISCHARGE DATE")} ↕</th>
                  <th>{t("PACKAGE")} ↕</th>
                  <th>{t("INSURANCE")} ↕</th>
                  <th>{t("POLICY NO")} ↕</th>
                  <th>{t("STATUS")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {admissions
                  .filter((a) =>
                    (
                      a.admissionId +
                      " " +
                      a.patientName +
                      " " +
                      a.doctorName +
                      " " +
                      a.packageName +
                      " " +
                      a.insuranceName
                    )
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((adm) => (
                    <tr key={adm.id}>
                      <td>
                        <span className="badge-blue-pill">
                          {adm.admissionId}
                        </span>
                      </td>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <div className="patient-avatar-circle">
                            {adm.patientName.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="fw-semibold text-primary">
                              {adm.patientName}
                            </div>
                            <div className="text-secondary small">
                              {adm.patientEmail}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="d-flex align-items-center gap-2">
                          <div className="doctor-avatar-circle">
                            {adm.doctorName.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="fw-semibold text-primary">
                              {adm.doctorName}
                            </div>
                            <div className="text-secondary small">
                              {adm.doctorEmail}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="tx-date-badge">
                          <span>{adm.admissionTime}</span>
                          <span>{adm.admissionDate}</span>
                        </span>
                      </td>
                      <td>{adm.dischargeDate}</td>
                      <td>
                        {adm.packageName !== "N/A" ? (
                          <span className="text-primary fw-semibold cursor-pointer">
                            {adm.packageName}
                          </span>
                        ) : (
                          "N/A"
                        )}
                      </td>
                      <td>
                        {adm.insuranceName !== "N/A" ? (
                          <span className="text-primary fw-semibold cursor-pointer">
                            {adm.insuranceName}
                          </span>
                        ) : (
                          "N/A"
                        )}
                      </td>
                      <td>{adm.policyNo}</td>
                      <td>
                        <label className="switch-toggle">
                          <input
                            type="checkbox"
                            checked={adm.status}
                            onChange={() =>
                              setAdmissions((prev) =>
                                prev.map((x) =>
                                  x.id === adm.id
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
                              setAdmissions((prev) =>
                                prev.filter((x) => x.id !== adm.id),
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
                {t("Showing")} 1 {t("to")} 10 {t("of")} 376 {t("Results")}
              </span>
            </div>
            <div className="pagination-numbers">
              <button className="page-btn active">1</button>
              <button className="page-btn">2</button>
              <button className="page-btn">3</button>
              <button className="page-btn">4</button>
              <button className="page-btn">...</button>
              <button className="page-btn">37</button>
              <button className="page-btn">38</button>
              <button className="page-btn">&gt;</button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL: CHANGE PATIENT PASSWORD (KEY ICON) */}
      {keyModalPatient && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "480px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("Change Password")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setKeyModalPatient(null)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                alert(`Password updated for ${keyModalPatient.name}`);
                setKeyModalPatient(null);
                setNewPassword("");
                setConfirmPassword("");
              }}
            >
              <div className="modal-body-custom">
                <div className="mb-3 text-secondary">
                  {t("User")}:{" "}
                  <strong className="text-light">{keyModalPatient.name}</strong>{" "}
                  ({keyModalPatient.email})
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("New Password")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="password"
                    required
                    placeholder={t("New Password")}
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Confirm Password")}:{" "}
                    <span className="text-danger">*</span>
                  </label>
                  <input
                    type="password"
                    required
                    placeholder={t("Confirm Password")}
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer-custom d-flex justify-content-end gap-2">
                <button type="submit" className="btn-action-blue">
                  {t("Save")}
                </button>
                <button
                  type="button"
                  className="btn-action-grey"
                  onClick={() => setKeyModalPatient(null)}
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
