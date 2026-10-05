"use client";

import { useLanguage } from "@/components/language";
import { useState } from "react";
import Link from "next/link";
import {
  Search,
  Filter,
  Video,
  Trash2,
  ChevronDown,
  X,
  Radio,
  CheckCircle2,
} from "lucide-react";

export type LiveConsultationWorkspaceProps = {
  id?: string;
};

interface ConsultationRow {
  id: string;
  title: string;
  date: string;
  time: string;
  createdBy: string;
  createdFor: string;
  patient: string;
  status: "Finished" | "Cancelled" | "Awaited";
  meetingId: string;
}

interface MeetingRow {
  id: string;
  title: string;
  date: string;
  time: string;
  createdBy: string;
  status: "Finished" | "Cancelled" | "Awaited";
  password: string;
}

export function LiveConsultationWorkspace({
  id = "live-consultations",
}: LiveConsultationWorkspaceProps) {
  const { t } = useLanguage();
  const currentTab =
    id === "live-consultations-live-meetings"
      ? "live-meetings"
      : "live-consultations";

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);

  // Dropdown for Actions button in Live Consultations
  const [actionsDropdownOpen, setActionsDropdownOpen] = useState(false);

  // Subtabs
  const tabs = [
    {
      id: "live-consultations",
      label: "Live Consultations",
      href: "/modules/live-consultations",
    },
    {
      id: "live-meetings",
      label: "Live Meetings",
      href: "/modules/live-consultations-live-meetings",
    },
  ];

  /* -------------------------------------------------------------
     1. LIVE CONSULTATIONS STATE (SCREENSHOT 182630)
     ------------------------------------------------------------- */
  const [consultations, setConsultations] = useState<ConsultationRow[]>([
    {
      id: "LC-1",
      title: "test demo",
      date: "19th Apr,2024",
      time: "12:00 AM",
      createdBy: "Bhautik Bhalala Doctor",
      createdFor: "Bhautik Bhalala Doctor",
      patient: "Pat. Bhautik Bhalala",
      status: "Finished",
      meetingId: "123456",
    },
    {
      id: "LC-2",
      title: "Skin Care",
      date: "1st Feb,2024",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      createdFor: "Adi Kusuma",
      patient: "Aamir 01hdbnzaeg1144gl1fx5j50re4b@mail4u.run",
      status: "Finished",
      meetingId: "123456",
    },
    {
      id: "LC-3",
      title: "Health Care",
      date: "26th Jan,2024",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      createdFor: "Abdoctor Lanstnam",
      patient: "Aamir 01hdbnzaeg1144gl1fx5j50re4b@mail4u.run",
      status: "Cancelled",
      meetingId: "123456",
    },
    {
      id: "LC-4",
      title: "dffsfsddsff",
      date: "16th Oct,2023",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      createdFor: "Bhautik Bhalala Doctor",
      patient: "Pat. Bhautik Bhalala",
      status: "Cancelled",
      meetingId: "123456",
    },
    {
      id: "LC-5",
      title: "Architecto",
      date: "20th Sep,2023",
      time: "12:00 AM",
      createdBy: "Lewis Kim",
      createdFor: "Lewis Kim",
      patient: "Pat. Bhautik Bhalala",
      status: "Cancelled",
      meetingId: "123456",
    },
    {
      id: "LC-6",
      title: "okuyts",
      date: "21st Sep,2023",
      time: "12:00 AM",
      createdBy: "Bhautik Bhalala Doctor",
      createdFor: "Bhautik Bhalala Doctor",
      patient: "Pat. Bhautik Bhalala",
      status: "Cancelled",
      meetingId: "123456",
    },
    {
      id: "LC-7",
      title: "test demo",
      date: "19th Sep,2023",
      time: "12:00 AM",
      createdBy: "Bhautik Bhalala Doctor",
      createdFor: "Bhautik Bhalala Doctor",
      patient: "Pat. Bhautik Bhalala",
      status: "Cancelled",
      meetingId: "123456",
    },
    {
      id: "LC-8",
      title: "BHahd",
      date: "20th Sep,2023",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      createdFor: "Bhautik Bhalala Doctor",
      patient: "Pat. Bhautik Bhalala",
      status: "Finished",
      meetingId: "123456",
    },
    {
      id: "LC-9",
      title: "qwerty",
      date: "19th Sep,2023",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      createdFor: "Bhautik Bhalala Doctor",
      patient: "Pat. Bhautik Bhalala",
      status: "Finished",
      meetingId: "123456",
    },
    {
      id: "LC-10",
      title: "Test",
      date: "2nd Sep,2023",
      time: "12:00 AM",
      createdBy: "Harish Mohan",
      createdFor: "Ali Sahil",
      patient: "Ashish Chaudhary",
      status: "Finished",
      meetingId: "hM53wA",
    },
  ]);

  // Modal: New Live Consultation
  const [consultModal, setConsultModal] = useState(false);
  const [cTitle, setCTitle] = useState("");
  const [cDoctor, setCDoctor] = useState("");
  const [cPatient, setCPatient] = useState("");
  const [cDate, setCDate] = useState("");
  const [cDuration, setCDuration] = useState("30");
  const [cDesc, setCDesc] = useState("");

  // Modal: Add Credential
  const [credentialModal, setCredentialModal] = useState(false);
  const [zoomKey, setZoomKey] = useState("");
  const [zoomSecret, setZoomSecret] = useState("");

  /* -------------------------------------------------------------
     2. LIVE MEETINGS STATE (SCREENSHOT 183109)
     ------------------------------------------------------------- */
  const [meetings, setMeetings] = useState<MeetingRow[]>([
    {
      id: "M-1",
      title: "ghjghjg",
      date: "16th Oct,2023",
      time: "12:00 AM",
      createdBy: "Bhautik Bhalala Doctor",
      status: "Finished",
      password: "123456",
    },
    {
      id: "M-2",
      title: "test demo",
      date: "20th Sep,2023",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      status: "Cancelled",
      password: "123456",
    },
    {
      id: "M-3",
      title: "as",
      date: "5th Apr,2023",
      time: "12:00 AM",
      createdBy: "Harish Mohan",
      status: "Cancelled",
      password: "MKrq1K",
    },
    {
      id: "M-4",
      title: "ff",
      date: "24th Mar,2023",
      time: "12:00 AM",
      createdBy: "Harish Mohan",
      status: "Finished",
      password: "6T7Pah",
    },
    {
      id: "M-5",
      title: "cc",
      date: "16th Mar,2023",
      time: "12:00 AM",
      createdBy: "Harish Mohan",
      status: "Finished",
      password: "Dwna2M",
    },
    {
      id: "M-6",
      title: "dcd",
      date: "22nd Dec,2022",
      time: "12:00 AM",
      createdBy: "Harish Mohan",
      status: "Cancelled",
      password: "9kU7Qa",
    },
    {
      id: "M-7",
      title: "testing",
      date: "19th Jan,2022",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      status: "Finished",
      password: "0GqY55",
    },
    {
      id: "M-8",
      title: "Voluptate magna modi",
      date: "19th Oct,2021",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      status: "Cancelled",
      password: "ktE2d8",
    },
    {
      id: "M-9",
      title: "Id qui est non iure",
      date: "28th Oct,2021",
      time: "12:00 AM",
      createdBy: "Infy HMS",
      status: "Finished",
      password: "fajQ9Z",
    },
  ]);

  // Modal: New Live Meeting
  const [meetingModal, setMeetingModal] = useState(false);
  const [mTitle, setMTitle] = useState("");
  const [mDate, setMDate] = useState("");
  const [mDuration, setMDuration] = useState("45");
  const [mHost, setMHost] = useState("");
  const [mDesc, setMDesc] = useState("");

  // Start Consultation
  function handleStart(row: ConsultationRow) {
    alert(
      `Starting Live Consultation: "${row.title}" (Meeting ID: ${row.meetingId})`,
    );
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

      {/* 1. LIVE CONSULTATIONS TAB (SCREENSHOT 182630) */}
      {currentTab === "live-consultations" && (
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
            <div className="d-flex gap-2 align-items-center">
              <button className="btn-icon-blue" title={t("Filter")}>
                <Filter size={18} />
              </button>

              <div className="dropdown-action-wrapper">
                <button
                  className="btn-action-blue d-flex align-items-center gap-1"
                  onClick={() => setActionsDropdownOpen(!actionsDropdownOpen)}
                >
                  <span>{t("Actions")}</span>
                  <ChevronDown size={14} />
                </button>
                {actionsDropdownOpen && (
                  <div className="dropdown-action-menu">
                    <button
                      className="dropdown-action-item"
                      onClick={() => {
                        setActionsDropdownOpen(false);
                        setConsultModal(true);
                      }}
                    >
                      {t("New Live Consultation")}
                    </button>
                    <button
                      className="dropdown-action-item"
                      onClick={() => {
                        setActionsDropdownOpen(false);
                        setCredentialModal(true);
                      }}
                    >
                      {t("Add Credential")}
                    </button>
                  </div>
                )}
              </div>

              <button
                className="btn-connect-zoom"
                onClick={() =>
                  alert(t("Zoom Integration connected successfully."))
                }
              >
                {t("Connect With Zoom")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("CONSULTATION TITLE")} ↕</th>
                  <th>{t("DATE")} ↕</th>
                  <th>{t("CREATED BY")} ↕</th>
                  <th>{t("CREATED FOR")} ↕</th>
                  <th>{t("PATIENT")} ↕</th>
                  <th>{t("STATUS")} ↕</th>
                  <th>{t("PASSWORD / MEETING ID")}</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {consultations
                  .filter((c) =>
                    (c.title + " " + c.createdBy + " " + c.patient)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((c) => (
                    <tr key={c.id}>
                      <td>
                        <span className="text-primary fw-semibold cursor-pointer">
                          {c.title}
                        </span>
                      </td>
                      <td>
                        <span className="tx-date-badge">
                          <span>{c.time}</span>
                          <span>{c.date}</span>
                        </span>
                      </td>
                      <td>{c.createdBy}</td>
                      <td>{c.createdFor}</td>
                      <td>{c.patient}</td>
                      <td>
                        <span
                          className={
                            c.status === "Finished"
                              ? "badge-green"
                              : "badge-red"
                          }
                        >
                          {c.status}
                        </span>
                      </td>
                      <td>
                        <span className="text-secondary font-monospace">
                          {c.meetingId}
                        </span>
                      </td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            className="btn-icon-blue-link"
                            title={t("Start Live Consultation")}
                            onClick={() => handleStart(c)}
                          >
                            <Video size={16} />
                          </button>
                          <button
                            className="btn-icon-danger"
                            title={t("Delete")}
                            onClick={() =>
                              setConsultations((prev) =>
                                prev.filter((x) => x.id !== c.id),
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
                {t("Showing")} 1 {t("to")} 10 {t("of")} 39 {t("Results")}
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

      {/* 2. LIVE MEETINGS TAB (SCREENSHOT 183109) */}
      {currentTab === "live-meetings" && (
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
                onClick={() => setMeetingModal(true)}
              >
                {t("New Live Meeting")}
              </button>
            </div>
          </div>

          <div className="table-responsive">
            <table className="billing-table w-100">
              <thead>
                <tr>
                  <th>{t("CONSULTATION TITLE")} ↕</th>
                  <th>{t("DATE")} ↕</th>
                  <th>{t("CREATED BY")} ↕</th>
                  <th>{t("STATUS")} ↕</th>
                  <th>{t("PASSWORD")} ↕</th>
                  <th>{t("ACTION")}</th>
                </tr>
              </thead>
              <tbody>
                {meetings
                  .filter((m) =>
                    (m.title + " " + m.createdBy)
                      .toLowerCase()
                      .includes(search.toLowerCase()),
                  )
                  .map((m) => (
                    <tr key={m.id}>
                      <td>
                        <span className="text-primary fw-semibold cursor-pointer">
                          {m.title}
                        </span>
                      </td>
                      <td>
                        <span className="tx-date-badge">
                          <span>{m.time}</span>
                          <span>{m.date}</span>
                        </span>
                      </td>
                      <td>{m.createdBy}</td>
                      <td>
                        <span
                          className={
                            m.status === "Finished"
                              ? "badge-green"
                              : "badge-red"
                          }
                        >
                          {m.status}
                        </span>
                      </td>
                      <td>
                        <span className="text-secondary font-monospace">
                          {m.password}
                        </span>
                      </td>
                      <td>
                        <div className="d-flex gap-2">
                          <button
                            className="btn-icon-danger"
                            title={t("Delete")}
                            onClick={() =>
                              setMeetings((prev) =>
                                prev.filter((x) => x.id !== m.id),
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
              </select>
              <span>
                {t("Showing")} {meetings.length} {t("Results")}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* MODAL: NEW LIVE CONSULTATION */}
      {consultModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "560px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Live Consultation")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setConsultModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                if (!cTitle) return;
                setConsultations([
                  {
                    id: `LC-${consultations.length + 1}`,
                    title: cTitle,
                    date: cDate || "05th Oct,2026",
                    time: "12:00 PM",
                    createdBy: "Admin",
                    createdFor: cDoctor || "Harish Mohan",
                    patient: cPatient || "Patient Demo",
                    status: "Awaited",
                    meetingId: String(
                      Math.floor(100000 + Math.random() * 900000),
                    ),
                  },
                  ...consultations,
                ]);
                setConsultModal(false);
                setCTitle("");
              }}
            >
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Consultation Title")}:{" "}
                    <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Consultation Title")}
                    value={cTitle}
                    onChange={(e) => setCTitle(e.target.value)}
                  />
                </div>
                <div className="form-grid-2">
                  <div className="form-group-custom mb-3">
                    <label>
                      {t("Doctor")}: <span className="text-danger">*</span>
                    </label>
                    <select
                      className="form-select-custom w-100"
                      required
                      value={cDoctor}
                      onChange={(e) => setCDoctor(e.target.value)}
                    >
                      <option value="">{t("Select Doctor")}</option>
                      <option value="Bhautik Bhalala Doctor">
                        Bhautik Bhalala Doctor
                      </option>
                      <option value="Harish Mohan">Harish Mohan</option>
                      <option value="Ali Sahil">Ali Sahil</option>
                    </select>
                  </div>
                  <div className="form-group-custom mb-3">
                    <label>
                      {t("Patient")}: <span className="text-danger">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder={t("Patient Name")}
                      value={cPatient}
                      onChange={(e) => setCPatient(e.target.value)}
                    />
                  </div>
                </div>
                <div className="form-grid-2">
                  <div className="form-group-custom mb-3">
                    <label>{t("Consultation Date")}:</label>
                    <input
                      type="date"
                      value={cDate}
                      onChange={(e) => setCDate(e.target.value)}
                    />
                  </div>
                  <div className="form-group-custom mb-3">
                    <label>{t("Duration (Minutes)")}:</label>
                    <input
                      type="number"
                      value={cDuration}
                      onChange={(e) => setCDuration(e.target.value)}
                    />
                  </div>
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={3}
                    placeholder={t("Description")}
                    value={cDesc}
                    onChange={(e) => setCDesc(e.target.value)}
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
                  onClick={() => setConsultModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: ADD CREDENTIAL */}
      {credentialModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "480px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("Add Credential")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setCredentialModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                alert(t("Zoom Credentials saved successfully."));
                setCredentialModal(false);
                setZoomKey("");
                setZoomSecret("");
              }}
            >
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Zoom API Key")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Zoom API Key")}
                    value={zoomKey}
                    onChange={(e) => setZoomKey(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Zoom API Secret")}:{" "}
                    <span className="text-danger">*</span>
                  </label>
                  <input
                    type="password"
                    required
                    placeholder={t("Zoom API Secret")}
                    value={zoomSecret}
                    onChange={(e) => setZoomSecret(e.target.value)}
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
                  onClick={() => setCredentialModal(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW LIVE MEETING */}
      {meetingModal && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "520px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("New Live Meeting")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setMeetingModal(false)}
              >
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                if (!mTitle) return;
                setMeetings([
                  {
                    id: `M-${meetings.length + 1}`,
                    title: mTitle,
                    date: mDate || "05th Oct,2026",
                    time: "12:00 PM",
                    createdBy: mHost || "Admin",
                    status: "Awaited",
                    password: String(
                      Math.floor(100000 + Math.random() * 900000),
                    ),
                  },
                  ...meetings,
                ]);
                setMeetingModal(false);
                setMTitle("");
              }}
            >
              <div className="modal-body-custom">
                <div className="form-group-custom mb-3">
                  <label>
                    {t("Meeting Title")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder={t("Meeting Title")}
                    value={mTitle}
                    onChange={(e) => setMTitle(e.target.value)}
                  />
                </div>
                <div className="form-grid-2">
                  <div className="form-group-custom mb-3">
                    <label>{t("Meeting Date")}:</label>
                    <input
                      type="date"
                      value={mDate}
                      onChange={(e) => setMDate(e.target.value)}
                    />
                  </div>
                  <div className="form-group-custom mb-3">
                    <label>{t("Duration (Minutes)")}:</label>
                    <input
                      type="number"
                      value={mDuration}
                      onChange={(e) => setMDuration(e.target.value)}
                    />
                  </div>
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Host / Created By")}:</label>
                  <input
                    type="text"
                    placeholder="Host Name"
                    value={mHost}
                    onChange={(e) => setMHost(e.target.value)}
                  />
                </div>
                <div className="form-group-custom mb-3">
                  <label>{t("Description")}:</label>
                  <textarea
                    rows={3}
                    placeholder={t("Description")}
                    value={mDesc}
                    onChange={(e) => setMDesc(e.target.value)}
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
                  onClick={() => setMeetingModal(false)}
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
