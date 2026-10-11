"use client";

import { useEffect, useState, useTransition } from "react";
import {
  Download,
  Eye,
  Pencil,
  Plus,
  Printer,
  Search,
  Trash2,
  X,
} from "lucide-react";
import { Modal } from "./modal";
import { useLanguage } from "./language";
import { generateQrSvg, drawQrToCanvas } from "@/lib/qr";

export type CardTemplate = {
  id: string;
  name: string;
  color: string;
  email: boolean;
  phone: boolean;
  dob: boolean;
  bloodGroup: boolean;
  address: boolean;
  patientUniqueId: boolean;
  createdAt?: string;
  updatedAt?: string;
};

export type IssuedCard = {
  id: string;
  patientId: string;
  patientName: string;
  patientUniqueId: string;
  cardNumber: string;
  qrToken: string;
  templateId: string;
  templateName: string;
  templateColor: string;
  status: string;
  issuedAt: string;
  expiresAt: string;
  issuedBy: string;
  email?: string;
  phone?: string;
  dob?: string;
  bloodGroup?: string;
  address?: string;
  mrn?: string;
};

export type PatientOption = {
  id: string;
  givenName: string;
  familyName: string;
  email?: string;
  phone?: string;
};

export function SmartCardsWorkspace({
  initialTab = "templates",
}: {
  initialTab?: "templates" | "cards";
}) {
  const { t } = useLanguage();
  const [tab, setTab] = useState<"templates" | "cards">(initialTab);
  const [, startTransition] = useTransition();

  // Templates state
  const [templates, setTemplates] = useState<CardTemplate[]>([]);
  const [templateSearch, setTemplateSearch] = useState("");
  const [templatePage, setTemplatePage] = useState(1);
  const [templatePageSize, setTemplatePageSize] = useState(10);
  const [editingTemplate, setEditingTemplate] = useState<CardTemplate | null>(
    null,
  );
  const [originalTemplateName, setOriginalTemplateName] = useState<
    string | null
  >(null);
  const [templateError, setTemplateError] = useState("");

  // Cards state
  const [cards, setCards] = useState<IssuedCard[]>([]);
  const [totalCards, setTotalCards] = useState(0);
  const [cardSearch, setCardSearch] = useState("");
  const [cardPage, setCardPage] = useState(1);
  const [cardPageSize, setCardPageSize] = useState(10);
  const [cardFilterTemplate, setCardFilterTemplate] = useState("");
  const [selectedCard, setSelectedCard] = useState<IssuedCard | null>(null);

  // Generation Modal state
  const [generateOpen, setGenerateOpen] = useState(false);
  const [generateMode, setGenerateMode] = useState<
    "all" | "single" | "unassigned"
  >("all");
  const [generateTemplateId, setGenerateTemplateId] = useState("");
  const [generatePatientId, setGeneratePatientId] = useState("");
  const [patientsList, setPatientsList] = useState<PatientOption[]>([]);
  const [generateError, setGenerateError] = useState("");
  const [generating, setGenerating] = useState(false);
  const [successNotice, setSuccessNotice] = useState("");

  // Load templates from API
  async function loadTemplates() {
    try {
      const res = await fetch("/api/hms/smart-cards/templates", {
        headers: { "Content-Type": "application/json" },
        cache: "no-store",
      });
      if (res.ok) {
        const data = await res.json();
        setTemplates(data.templates || []);
        if (
          data.templates &&
          data.templates.length > 0 &&
          !generateTemplateId
        ) {
          setGenerateTemplateId(data.templates[0].id);
        }
      }
    } catch {}
  }

  // Load issued cards from API
  async function loadCards() {
    try {
      const params = new URLSearchParams({
        page: String(cardPage),
        pageSize: String(cardPageSize),
      });
      if (cardSearch.trim()) params.set("search", cardSearch.trim());
      if (cardFilterTemplate) params.set("templateId", cardFilterTemplate);

      const res = await fetch(
        `/api/hms/smart-cards/cards?${params.toString()}`,
        {
          headers: { "Content-Type": "application/json" },
          cache: "no-store",
        },
      );
      if (res.ok) {
        const data = await res.json();
        setCards(data.cards || []);
        setTotalCards(data.totalCount || 0);
      }
    } catch {}
  }

  // Load patients list for single generation mode
  async function loadPatients() {
    try {
      const res = await fetch("/api/hms/patients", {
        headers: { "Content-Type": "application/json" },
        cache: "no-store",
      });
      if (res.ok) {
        const data = await res.json();
        const list = Array.isArray(data)
          ? data
          : data.patients || data.items || [];
        setPatientsList(list);
        if (list.length > 0 && !generatePatientId) {
          setGeneratePatientId(list[0].id);
        }
      }
    } catch {}
  }

  useEffect(() => {
    loadTemplates();
  }, []);

  useEffect(() => {
    loadCards();
  }, [cardPage, cardPageSize, cardSearch, cardFilterTemplate]);

  // Update inline template toggle
  async function toggleTemplateField(
    tmpl: CardTemplate,
    field: keyof CardTemplate,
    value: boolean,
  ) {
    const updated = { ...tmpl, [field]: value };
    setTemplates((old) => old.map((t) => (t.id === tmpl.id ? updated : t)));

    try {
      const res = await fetch(`/api/hms/smart-cards/templates/${tmpl.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ [field]: value }),
      });
      if (!res.ok) {
        // Rollback on failure
        setTemplates((old) => old.map((t) => (t.id === tmpl.id ? tmpl : t)));
        const err = await res.json().catch(() => ({}));
        setTemplateError(err.error || "Failed to update toggle");
      }
    } catch {
      setTemplates((old) => old.map((t) => (t.id === tmpl.id ? tmpl : t)));
      setTemplateError("Failed to update toggle");
    }
  }

  // Save template (Create or Update)
  async function handleSaveTemplate(e: React.FormEvent) {
    e.preventDefault();
    if (!editingTemplate) return;
    const name = editingTemplate.name.trim();
    if (!name) {
      setTemplateError("Template name is required.");
      return;
    }

    try {
      if (originalTemplateName) {
        // Update
        const res = await fetch(
          `/api/hms/smart-cards/templates/${editingTemplate.id}`,
          {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              name,
              color: editingTemplate.color,
              email: editingTemplate.email,
              phone: editingTemplate.phone,
              dob: editingTemplate.dob,
              bloodGroup: editingTemplate.bloodGroup,
              address: editingTemplate.address,
              patientUniqueId: editingTemplate.patientUniqueId,
            }),
          },
        );
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          if (res.status === 409) {
            setTemplateError("Template name must be unique.");
          } else {
            setTemplateError(body.error || "Failed to update template.");
          }
          return;
        }
      } else {
        // Create
        const res = await fetch("/api/hms/smart-cards/templates", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            name,
            color: editingTemplate.color,
            email: editingTemplate.email,
            phone: editingTemplate.phone,
            dob: editingTemplate.dob,
            bloodGroup: editingTemplate.bloodGroup,
            address: editingTemplate.address,
            patientUniqueId: editingTemplate.patientUniqueId,
          }),
        });
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          if (res.status === 409) {
            setTemplateError("Template name must be unique.");
          } else {
            setTemplateError(body.error || "Failed to create template.");
          }
          return;
        }
      }

      setEditingTemplate(null);
      setOriginalTemplateName(null);
      setTemplateError("");
      loadTemplates();
      loadCards();
    } catch {
      setTemplateError("Server error while saving template.");
    }
  }

  // Delete Template (clears patient assignments and revokes associated cards)
  async function handleDeleteTemplate(tmpl: CardTemplate) {
    if (!window.confirm(`${t("Delete this template?")} (${tmpl.name})`)) return;
    try {
      const res = await fetch(`/api/hms/smart-cards/templates/${tmpl.id}`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        setTemplateError(body.error || "Failed to delete template.");
        return;
      }
      setTemplateError("");
      loadTemplates();
    } catch {
      setTemplateError("Server error deleting template.");
    }
  }

  // Generate cards
  async function handleGenerateCards(e: React.FormEvent) {
    e.preventDefault();
    if (!generateTemplateId) {
      setGenerateError("Please select a template.");
      return;
    }
    if (generateMode === "single" && !generatePatientId) {
      setGenerateError("Please select a patient.");
      return;
    }

    setGenerating(true);
    setGenerateError("");
    try {
      const res = await fetch("/api/hms/smart-cards/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          templateId: generateTemplateId,
          mode: generateMode,
          patientId: generateMode === "single" ? generatePatientId : undefined,
        }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        setGenerateError(body.error || "Failed to generate smart cards.");
        setGenerating(false);
        return;
      }

      const result = await res.json();
      setGenerating(false);
      setGenerateOpen(false);
      setSuccessNotice(`${result.generatedCount || 0} ${t("Cards Generated")}`);
      loadCards();
    } catch {
      setGenerating(false);
      setGenerateError("Server error during card generation.");
    }
  }

  // Delete Card Assignment
  async function handleDeleteCard(card: IssuedCard) {
    if (!window.confirm(t("Delete this patient card?"))) return;
    try {
      const res = await fetch(`/api/hms/smart-cards/cards/${card.id}`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
      });
      if (res.ok) {
        loadCards();
      }
    } catch {}
  }

  // Filtered Templates
  const filteredTemplates = templates.filter((t) =>
    t.name.toLowerCase().includes(templateSearch.toLowerCase()),
  );
  const templateCount = filteredTemplates.length;
  const templatePages = Math.max(
    1,
    Math.ceil(templateCount / templatePageSize),
  );
  const currentTemplatePage = Math.min(templatePage, templatePages);

  // Selected template for card view
  const activeTemplate = templates.find(
    (t) => t.id === selectedCard?.templateId,
  ) ||
    templates.find((t) => t.name === selectedCard?.templateName) || {
      id: "fallback",
      name: selectedCard?.templateName || "Standard",
      color: selectedCard?.templateColor || "#f4bf16",
      email: true,
      phone: true,
      dob: true,
      bloodGroup: true,
      address: true,
      patientUniqueId: true,
    };

  return (
    <section className="smart-card-screen">
      {/* Subtab switcher */}
      <div className="detail-tabs mb-4 flex gap-2">
        <button
          className={tab === "templates" ? "active" : ""}
          onClick={() => {
            setTab("templates");
            setTemplateError("");
            setSuccessNotice("");
          }}
        >
          {t("Smart Patient Card Templates")}
        </button>
        <button
          className={tab === "cards" ? "active" : ""}
          onClick={() => {
            setTab("cards");
            setTemplateError("");
            setSuccessNotice("");
          }}
        >
          {t("Generate Patient Smart Cards")}
        </button>
      </div>

      {templateError && !editingTemplate && (
        <p role="alert" className="error mb-4" style={{ color: "#ef4444" }}>
          {t(templateError)}
        </p>
      )}

      {successNotice && (
        <p role="status" className="notice mb-4" style={{ color: "#10b981" }}>
          {successNotice}
        </p>
      )}

      {tab === "templates" ? (
        /* ─── TEMPLATES TAB ─────────────────────────────────────────── */
        <>
          <div className="page-heading">
            <h1>{t("Smart Patient Card Templates")}</h1>
            <button
              className="primary"
              onClick={() => {
                setOriginalTemplateName(null);
                setTemplateError("");
                setEditingTemplate({
                  id: "",
                  name: "",
                  color: "#f4bf16",
                  email: true,
                  phone: true,
                  dob: true,
                  bloodGroup: true,
                  address: true,
                  patientUniqueId: true,
                });
              }}
            >
              <Plus size={16} />
              {t("New Patient Smart Card Template")}
            </button>
          </div>

          <div className="table-toolbar">
            <div className="table-search">
              <Search size={18} />
              <input
                aria-label={t("Search cards")}
                placeholder={t("Search")}
                value={templateSearch}
                onChange={(e) => {
                  setTemplateSearch(e.target.value);
                  setTemplatePage(1);
                }}
              />
            </div>
          </div>

          <div className="legacy-table-wrap">
            <table className="legacy-table reference-striped">
              <thead>
                <tr>
                  <th>{t("Name")}</th>
                  <th>{t("Color")}</th>
                  <th>{t("Email")}</th>
                  <th>{t("Phone")}</th>
                  <th>{t("Date of Birth")}</th>
                  <th>{t("Blood Group")}</th>
                  <th>{t("Address")}</th>
                  <th>{t("Patient Unique ID")}</th>
                  <th>{t("Action")}</th>
                </tr>
              </thead>
              <tbody>
                {filteredTemplates
                  .slice(
                    (currentTemplatePage - 1) * templatePageSize,
                    currentTemplatePage * templatePageSize,
                  )
                  .map((tmpl) => (
                    <tr key={tmpl.id || tmpl.name}>
                      <td>
                        <strong>{tmpl.name}</strong>
                      </td>
                      <td>
                        <span
                          className="color-swatch"
                          style={{
                            background: tmpl.color,
                            display: "inline-block",
                            width: 24,
                            height: 24,
                            borderRadius: 4,
                            border: "1px solid rgba(0,0,0,0.1)",
                          }}
                        />
                      </td>
                      {(
                        [
                          ["email", tmpl.email],
                          ["phone", tmpl.phone],
                          ["dob", tmpl.dob],
                          ["bloodGroup", tmpl.bloodGroup],
                          ["address", tmpl.address],
                          ["patientUniqueId", tmpl.patientUniqueId],
                        ] as const
                      ).map(([fieldKey, val]) => (
                        <td key={fieldKey}>
                          <input
                            className="reference-switch"
                            type="checkbox"
                            role="switch"
                            aria-label={`${tmpl.name} ${fieldKey}`}
                            checked={val}
                            onChange={(e) =>
                              toggleTemplateField(
                                tmpl,
                                fieldKey as keyof CardTemplate,
                                e.target.checked,
                              )
                            }
                          />
                        </td>
                      ))}
                      <td>
                        <div className="row-actions flex gap-2">
                          <button
                            className="text-brand"
                            aria-label={`Edit ${tmpl.name}`}
                            onClick={() => {
                              setOriginalTemplateName(tmpl.name);
                              setTemplateError("");
                              setEditingTemplate({ ...tmpl });
                            }}
                          >
                            <Pencil size={17} />
                          </button>
                          <button
                            className="text-red-500"
                            aria-label={`Delete ${tmpl.name}`}
                            onClick={() => handleDeleteTemplate(tmpl)}
                          >
                            <Trash2 size={17} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>

          <div className="reference-pagination">
            <label>
              {t("Show")}
              <select
                className="field"
                aria-label="Rows per page"
                value={templatePageSize}
                onChange={(e) => {
                  setTemplatePageSize(Number(e.target.value));
                  setTemplatePage(1);
                }}
              >
                {[10, 25, 50].map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
            </label>
            <span>
              {t("Showing")}{" "}
              {templateCount
                ? (currentTemplatePage - 1) * templatePageSize + 1
                : 0}{" "}
              -{" "}
              {Math.min(currentTemplatePage * templatePageSize, templateCount)}{" "}
              {t("of")} {templateCount} {t("Results")}
            </span>
            <div>
              {Array.from({ length: templatePages }, (_, i) => (
                <button
                  key={i}
                  className={
                    currentTemplatePage === i + 1 ? "primary" : "secondary"
                  }
                  onClick={() => setTemplatePage(i + 1)}
                >
                  {i + 1}
                </button>
              ))}
            </div>
          </div>
        </>
      ) : (
        /* ─── GENERATE & CARDS TAB ───────────────────────────────────── */
        <>
          <div className="page-heading">
            <h1>{t("Generate Patient Smart Cards")}</h1>
            <button
              className="primary"
              onClick={() => {
                setGenerateError("");
                setGenerateMode("all");
                loadPatients();
                setGenerateOpen(true);
              }}
            >
              <Plus size={16} />
              {t("Generate Patient Smart Card")}
            </button>
          </div>

          <div className="table-toolbar flex gap-3 items-center">
            <div className="table-search">
              <Search size={18} />
              <input
                aria-label={t("Search cards")}
                placeholder={t("Search")}
                value={cardSearch}
                onChange={(e) => {
                  setCardSearch(e.target.value);
                  setCardPage(1);
                }}
              />
            </div>
            {templates.length > 0 && (
              <select
                className="field"
                aria-label={t("Filter by template")}
                value={cardFilterTemplate}
                onChange={(e) => {
                  setCardFilterTemplate(e.target.value);
                  setCardPage(1);
                }}
              >
                <option value="">
                  {t("All Templates") || "All Templates"}
                </option>
                {templates.map((tmpl) => (
                  <option key={tmpl.id} value={tmpl.id}>
                    {tmpl.name}
                  </option>
                ))}
              </select>
            )}
          </div>

          <div className="legacy-table-wrap">
            <table className="legacy-table reference-striped">
              <thead>
                <tr>
                  <th>{t("Patients")}</th>
                  <th>{t("Patient Unique ID")}</th>
                  <th>{t("Template Name")}</th>
                  <th>{t("Action")}</th>
                </tr>
              </thead>
              <tbody>
                {cards.map((card) => (
                  <tr key={card.id}>
                    <td>
                      <button
                        className="record-link flex items-center gap-3"
                        onClick={() => setSelectedCard(card)}
                      >
                        <span className="avatar">
                          {card.patientName
                            ? card.patientName
                                .split(" ")
                                .map((n) => n[0])
                                .join("")
                                .slice(0, 2)
                            : "PT"}
                        </span>
                        <span>
                          <strong>{card.patientName}</strong>
                          {card.email && <small>{card.email}</small>}
                        </span>
                      </button>
                    </td>
                    <td>
                      <span className="badge card-id">
                        {card.patientUniqueId || "—"}
                      </span>
                    </td>
                    <td>
                      <span
                        className="badge"
                        style={{
                          borderLeft: `4px solid ${card.templateColor || "#f4bf16"}`,
                          paddingLeft: 8,
                        }}
                      >
                        {card.templateName}
                      </span>
                    </td>
                    <td>
                      <div className="row-actions flex gap-2">
                        <button
                          aria-label={`View card ${card.patientName}`}
                          onClick={() => setSelectedCard(card)}
                        >
                          <Eye size={18} />
                        </button>
                        <button
                          aria-label={`Download card ${card.patientName}`}
                          onClick={() => downloadCardPng(card, activeTemplate)}
                        >
                          <Download size={18} />
                        </button>
                        <button
                          className="text-red-500"
                          aria-label={`Delete card ${card.patientName}`}
                          onClick={() => handleDeleteCard(card)}
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {cards.length === 0 && (
                  <tr>
                    <td colSpan={4} className="text-center py-6 text-gray-500">
                      {t("No records found")}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className="reference-pagination">
            <label>
              {t("Show")}
              <select
                className="field"
                aria-label="Rows per page"
                value={cardPageSize}
                onChange={(e) => {
                  setCardPageSize(Number(e.target.value));
                  setCardPage(1);
                }}
              >
                {[10, 25, 50].map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
            </label>
            <span>
              {t("Showing")}{" "}
              {totalCards ? (cardPage - 1) * cardPageSize + 1 : 0} -{" "}
              {Math.min(cardPage * cardPageSize, totalCards)} {t("of")}{" "}
              {totalCards} {t("Results")}
            </span>
            <div>
              {Array.from(
                { length: Math.max(1, Math.ceil(totalCards / cardPageSize)) },
                (_, i) => (
                  <button
                    key={i}
                    className={cardPage === i + 1 ? "primary" : "secondary"}
                    onClick={() => setCardPage(i + 1)}
                  >
                    {i + 1}
                  </button>
                ),
              )}
            </div>
          </div>
        </>
      )}

      {/* ─── TEMPLATE CREATE / EDIT MODAL ──────────────────────────── */}
      {editingTemplate && (
        <Modal
          titleId="template-title"
          onClose={() => setEditingTemplate(null)}
        >
          <form onSubmit={handleSaveTemplate}>
            <div className="modal-heading">
              <h2 id="template-title">
                {t(
                  originalTemplateName
                    ? "Edit Patient Smart Card Template"
                    : "New Patient Smart Card Template",
                )}
              </h2>
              <button
                type="button"
                aria-label={t("Close")}
                onClick={() => setEditingTemplate(null)}
              >
                <X />
              </button>
            </div>
            <div className="modal-content">
              {templateError && (
                <p
                  className="error mb-3"
                  role="alert"
                  style={{ color: "#ef4444" }}
                >
                  {t(templateError)}
                </p>
              )}
              <div className="legacy-form mb-4">
                <label>
                  <span className="label">{t("Template Name")}</span>
                  <input
                    className="field"
                    required
                    value={editingTemplate.name}
                    onChange={(e) =>
                      setEditingTemplate({
                        ...editingTemplate,
                        name: e.target.value,
                      })
                    }
                  />
                </label>
                <label>
                  <span className="label">{t("Header Color")}</span>
                  <div className="flex items-center gap-3">
                    <input
                      className="field"
                      type="color"
                      value={editingTemplate.color}
                      onChange={(e) =>
                        setEditingTemplate({
                          ...editingTemplate,
                          color: e.target.value,
                        })
                      }
                      style={{ width: 60, height: 38, padding: 2 }}
                    />
                    <span>{editingTemplate.color}</span>
                  </div>
                </label>

                <div className="grid grid-cols-2 gap-3 mt-2">
                  {(
                    [
                      ["email", "Email"],
                      ["phone", "Phone"],
                      ["dob", "Date of Birth"],
                      ["bloodGroup", "Blood Group"],
                      ["address", "Address"],
                      ["patientUniqueId", "Patient Unique ID"],
                    ] as const
                  ).map(([key, label]) => (
                    <label className="flex items-center gap-2" key={key}>
                      <input
                        type="checkbox"
                        checked={editingTemplate[key]}
                        onChange={(e) =>
                          setEditingTemplate({
                            ...editingTemplate,
                            [key]: e.target.checked,
                          })
                        }
                      />
                      <span>{t(label)}</span>
                    </label>
                  ))}
                </div>
              </div>

              {/* Live Preview Card */}
              <div className="mt-4">
                <span className="label block mb-2 font-medium">
                  {t("Card Preview") || "Card Preview"}
                </span>
                <SmartCardView
                  patientName="Sample Patient"
                  patientUniqueId="P-01001"
                  email="sample@example.invalid"
                  phone="+251911223344"
                  dob="1995-05-15"
                  bloodGroup="AB+"
                  address="Addis Ababa, Ethiopia"
                  qrToken="SAMPLE-QR-TOKEN-PREVIEW-ONLY-12345"
                  template={editingTemplate}
                />
              </div>
            </div>
            <div className="modal-footer flex gap-3 justify-end mt-4">
              <button className="primary" type="submit">
                {t("Save")}
              </button>
              <button
                className="secondary"
                type="button"
                onClick={() => setEditingTemplate(null)}
              >
                {t("Cancel")}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* ─── GENERATE CARDS MODAL ──────────────────────────────────── */}
      {generateOpen && (
        <Modal
          titleId="generate-cards-title"
          onClose={() => setGenerateOpen(false)}
        >
          <form onSubmit={handleGenerateCards}>
            <div className="modal-heading">
              <h2 id="generate-cards-title">
                {t("Generate Patient Smart Card")}
              </h2>
              <button
                type="button"
                aria-label={t("Close")}
                onClick={() => setGenerateOpen(false)}
              >
                <X />
              </button>
            </div>
            <div className="modal-content legacy-form">
              {generateError && (
                <p
                  className="error mb-3"
                  role="alert"
                  style={{ color: "#ef4444" }}
                >
                  {t(generateError)}
                </p>
              )}

              <label>
                <span className="label">{t("Generate Mode")}</span>
                <select
                  className="field"
                  value={generateMode}
                  onChange={(e) =>
                    setGenerateMode(
                      e.target.value as "all" | "single" | "unassigned",
                    )
                  }
                >
                  <option value="all">{t("All Patients")}</option>
                  <option value="single">{t("Selected Patient")}</option>
                  <option value="unassigned">
                    {t("Patients Without Cards")}
                  </option>
                </select>
              </label>

              <label>
                <span className="label">{t("Select Template")}</span>
                <select
                  className="field"
                  required
                  value={generateTemplateId}
                  onChange={(e) => setGenerateTemplateId(e.target.value)}
                >
                  {templates.map((tmpl) => (
                    <option key={tmpl.id} value={tmpl.id}>
                      {tmpl.name}
                    </option>
                  ))}
                </select>
              </label>

              {generateMode === "single" && (
                <label>
                  <span className="label">{t("Select Patient")}</span>
                  <select
                    className="field"
                    required
                    value={generatePatientId}
                    onChange={(e) => setGeneratePatientId(e.target.value)}
                  >
                    {patientsList.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.givenName} {p.familyName}{" "}
                        {p.phone ? `(${p.phone})` : ""}
                      </option>
                    ))}
                  </select>
                </label>
              )}
            </div>
            <div className="modal-footer flex gap-3 justify-end mt-4">
              <button className="primary" disabled={generating}>
                {generating ? t("Generating...") : t("Save")}
              </button>
              <button
                type="button"
                className="secondary"
                onClick={() => setGenerateOpen(false)}
              >
                {t("Cancel")}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* ─── VIEW SMART CARD MODAL ─────────────────────────────────── */}
      {selectedCard && (
        <Modal
          titleId="smart-card-view-title"
          onClose={() => setSelectedCard(null)}
        >
          <div className="smart-card-modal printable-card-area">
            <button
              className="smart-card-close"
              aria-label={t("Close card")}
              onClick={() => setSelectedCard(null)}
            >
              <X size={19} />
            </button>
            <h2 id="smart-card-view-title" className="sr-only">
              {t("Patient Smart Cards") || "Patient Smart Cards"}
            </h2>
            <SmartCardView
              patientName={selectedCard.patientName}
              patientUniqueId={selectedCard.patientUniqueId}
              email={selectedCard.email}
              phone={selectedCard.phone}
              dob={selectedCard.dob}
              bloodGroup={selectedCard.bloodGroup}
              address={selectedCard.address}
              qrToken={selectedCard.qrToken}
              template={activeTemplate}
            />
            <div className="modal-footer flex gap-3 justify-end mt-4">
              <button
                className="secondary flex items-center gap-2"
                onClick={() => downloadCardPng(selectedCard, activeTemplate)}
              >
                <Download size={16} />
                {t("Download")}
              </button>
              <button
                className="secondary flex items-center gap-2"
                onClick={() => window.print()}
              >
                <Printer size={16} />
                {t("Print")}
              </button>
              <button
                className="secondary"
                onClick={() => setSelectedCard(null)}
              >
                {t("Close")}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </section>
  );
}

// ─── SMART CARD VISUAL COMPONENT ──────────────────────────────────────────────

export function SmartCardView({
  patientName,
  patientUniqueId,
  email,
  phone,
  dob,
  bloodGroup,
  address,
  qrToken,
  template,
}: {
  patientName: string;
  patientUniqueId: string;
  email?: string;
  phone?: string;
  dob?: string;
  bloodGroup?: string;
  address?: string;
  qrToken: string;
  template: CardTemplate;
}) {
  const { t } = useLanguage();
  const origin =
    typeof window !== "undefined"
      ? window.location.origin
      : "http://127.0.0.1:3000";
  const verifyUrl = `${origin}/v1/smart-cards/verify?token=${qrToken}`;
  const qrSvg = generateQrSvg(verifyUrl, 110);

  const initials = patientName
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2);

  return (
    <article className="smart-card">
      <header style={{ background: template.color }}>
        <strong>ULSHMS</strong>
        <span>{t("Addis Ababa, Ethiopia · Hospital Management System")}</span>
      </header>
      <div className="smart-card-details">
        <div className="card-portrait">{initials}</div>
        <dl>
          <div className="card-field">
            <dt>{t("Patient Name")}:</dt>
            <dd>
              <strong>{patientName}</strong>
            </dd>
          </div>
          {template.email && email && (
            <div className="card-field">
              <dt>{t("Email")}:</dt>
              <dd>{email}</dd>
            </div>
          )}
          {template.phone && phone && (
            <div className="card-field">
              <dt>{t("Phone")}:</dt>
              <dd>{phone}</dd>
            </div>
          )}
          {template.dob && dob && (
            <div className="card-field">
              <dt>{t("Date of Birth")}:</dt>
              <dd>{dob}</dd>
            </div>
          )}
          {template.bloodGroup && bloodGroup && (
            <div className="card-field">
              <dt>{t("Blood Group")}:</dt>
              <dd>{bloodGroup}</dd>
            </div>
          )}
          {template.address && address && (
            <div className="card-field">
              <dt>{t("Address")}:</dt>
              <dd>{address}</dd>
            </div>
          )}
        </dl>
        <div
          className="card-qr flex flex-col items-center"
          data-qr-token={qrToken}
          data-qr-url={verifyUrl}
        >
          <div
            dangerouslySetInnerHTML={{ __html: qrSvg }}
            style={{ width: 110, height: 110 }}
          />
          {template.patientUniqueId && (
            <strong className="mt-1">{patientUniqueId}</strong>
          )}
        </div>
      </div>
    </article>
  );
}

// ─── HIGH-RES PNG DOWNLOAD CANVAS BUILDER ─────────────────────────────────────

async function downloadCardPng(card: IssuedCard, template: CardTemplate) {
  const canvas = document.createElement("canvas");
  canvas.width = 1000;
  canvas.height = 460;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  // Background
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, 1000, 460);

  // Header Banner
  ctx.fillStyle = template.color || "#f4bf16";
  ctx.fillRect(0, 0, 1000, 100);

  // Header Text
  ctx.fillStyle = "#ffffff";
  ctx.font = "bold 32px Arial, sans-serif";
  ctx.fillText("ULSHMS", 35, 62);
  ctx.font = "18px Arial, sans-serif";
  ctx.fillText(
    "Hospital Management System · Patient Smart Identity Card",
    220,
    62,
  );

  // Avatar Circle
  ctx.fillStyle = "#f3f4f6";
  ctx.beginPath();
  ctx.arc(85, 230, 50, 0, Math.PI * 2);
  ctx.fill();

  ctx.fillStyle = "#6571ff";
  ctx.font = "bold 34px Arial, sans-serif";
  ctx.textAlign = "center";
  const initials = card.patientName
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2);
  ctx.fillText(initials, 85, 242);
  ctx.textAlign = "left";

  // Demographic Rows
  ctx.fillStyle = "#1e293b";
  ctx.font = "20px Arial, sans-serif";

  const rows: [string, string][] = [["Patient Name", card.patientName]];
  if (template.email && card.email) rows.push(["Email", card.email]);
  if (template.phone && card.phone) rows.push(["Phone", card.phone]);
  if (template.dob && card.dob) rows.push(["Date of Birth", card.dob]);
  if (template.bloodGroup && card.bloodGroup)
    rows.push(["Blood Group", card.bloodGroup]);
  if (template.address && card.address) rows.push(["Address", card.address]);

  rows.forEach(([label, value], i) => {
    ctx.fillStyle = "#64748b";
    ctx.fillText(`${label}:`, 170, 150 + i * 40);
    ctx.fillStyle = "#0f172a";
    ctx.fillText(value, 320, 150 + i * 40);
  });

  // QR Code offscreen canvas render
  const origin =
    typeof window !== "undefined"
      ? window.location.origin
      : "http://127.0.0.1:3000";
  const verifyUrl = `${origin}/v1/smart-cards/verify?token=${card.qrToken}`;
  const qrCanvas = document.createElement("canvas");
  drawQrToCanvas(qrCanvas, verifyUrl, 180);
  ctx.drawImage(qrCanvas, 770, 140, 180, 180);

  // Patient Unique ID below QR
  if (template.patientUniqueId && card.patientUniqueId) {
    ctx.fillStyle = "#0f172a";
    ctx.font = "bold 19px Arial, sans-serif";
    ctx.textAlign = "center";
    ctx.fillText(card.patientUniqueId, 860, 355);
  }

  // Trigger download
  const a = document.createElement("a");
  a.href = canvas.toDataURL("image/png");
  a.download = `patient-card-${card.patientUniqueId || card.cardNumber}.png`;
  a.click();
}
