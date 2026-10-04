"use client";
import { useEffect, useState, useCallback } from "react";
import { Modal } from "@/components/modal";
import { Plus, Search, Users, X } from "lucide-react";
import { api, type Patient } from "@/lib/api";
import { useIdentity } from "@/components/workspace";
export default function Patients() {
  const identity = useIdentity();
  const [rows, setRows] = useState<Patient[]>([]);
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [modal, setModal] = useState(false);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState("");
  const [notice, setNotice] = useState("");
  const load = useCallback(() => {
    setLoading(true);
    setError("");
    api<{ patients: Patient[] }>(
      `patients?search=${encodeURIComponent(query)}&page=${page}`,
    )
      .then((r) => setRows(r.patients))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [query, page]);
  useEffect(load, [load]);
  useEffect(() => {
    if (new URLSearchParams(window.location.search).get("register") === "1")
      setModal(true);
  }, []);
  async function register(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setFormError("");
    const body = Object.fromEntries(new FormData(e.currentTarget));
    try {
      const result = await api<Patient>("patients", {
        method: "POST",
        body: JSON.stringify(body),
      });
      setNotice(`${result.givenName} registered as ${result.mrn}.`);
      setModal(false);
      load();
    } catch (err) {
      setFormError(
        err instanceof Error ? err.message : "Unable to register patient",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div>
          <p className="eyebrow text-brand">Patient management</p>
          <h1 className="mt-2 text-[28px] font-semibold tracking-tight">
            Patient directory
          </h1>
          <p className="mt-2 text-sm text-muted">
            The people at the heart of your hospital.
          </p>
        </div>
        {identity?.permissions.includes("patients.create") && (
          <button
            onClick={() => {
              setModal(true);
              setFormError("");
            }}
            className="primary self-center"
          >
            <Plus size={16} />
            Register patient
          </button>
        )}
      </div>
      {notice && (
        <p role="status" className="success">
          {notice}
        </p>
      )}
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <section className="card overflow-hidden">
        <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-100 p-5">
          <form
            className="flex w-full max-w-md gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              setPage(1);
              setQuery(search);
            }}
          >
            <div className="relative flex-1">
              <Search
                size={16}
                className="absolute left-3 top-3.5 text-muted"
              />
              <input
                aria-label="Search patients by name or record number"
                className="field !pl-10"
                placeholder="Search by name or record number…"
                value={search}
                maxLength={80}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <button className="secondary">Search</button>
          </form>
          <span className="text-xs text-muted">25 records per page</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-[#fafcfc] text-[10px] uppercase tracking-wider text-muted">
              <tr>
                {[
                  "Patient",
                  "Record number",
                  "Date of birth",
                  "Phone",
                  "Registered",
                ].map((h) => (
                  <th key={h} className="px-6 py-4 font-medium">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {!loading &&
                rows.map((p) => (
                  <tr
                    key={p.id}
                    className="border-t border-slate-100 hover:bg-slate-50/50"
                  >
                    <td className="px-6 py-5">
                      <span className="mr-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-[#edf4f1] text-[10px] font-bold text-brand">
                        {p.givenName[0]}
                        {p.familyName[0]}
                      </span>
                      <span className="font-semibold">
                        {p.givenName} {p.familyName}
                      </span>
                    </td>
                    <td className="px-6 py-5 font-mono text-muted">{p.mrn}</td>
                    <td className="px-6 py-5">{p.dateOfBirth}</td>
                    <td className="px-6 py-5 text-muted">
                      {p.phone || "Not provided"}
                    </td>
                    <td className="px-6 py-5 text-muted">
                      {new Date(p.createdAt).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
        {loading ? (
          <div className="p-16 text-center text-muted" role="status">
            Loading patient records…
          </div>
        ) : (
          !rows.length && (
            <div className="flex flex-col items-center px-6 py-16 text-center">
              <span className="mb-4 rounded-2xl bg-[#f1f6f4] p-4 text-brand">
                <Users size={28} strokeWidth={1.4} />
              </span>
              <h2 className="font-semibold">
                {query ? "No matching patients" : "No patient records yet"}
              </h2>
              <p className="mt-2 text-xs text-muted">
                {query
                  ? "Try another name or the numeric part of the record number."
                  : "Registered patients within your access scope will appear here."}
              </p>
            </div>
          )
        )}
        <div className="flex items-center justify-between border-t border-slate-100 p-4">
          <span className="text-xs text-muted">Page {page}</span>
          <div className="flex gap-2">
            <button
              className="secondary"
              disabled={page === 1 || loading}
              onClick={() => setPage(page - 1)}
            >
              Previous
            </button>
            <button
              className="secondary"
              disabled={rows.length < 25 || loading}
              onClick={() => setPage(page + 1)}
            >
              Next
            </button>
          </div>
        </div>
      </section>
      {modal && identity?.permissions.includes("patients.create") && (
        <Modal titleId="register-title" onClose={() => setModal(false)}>
          <section className="p-7">
            <div className="mb-6 flex justify-between">
              <div>
                <h2 id="register-title" className="text-xl font-semibold">
                  Register a patient
                </h2>
                <p className="mt-2 text-xs text-muted">
                  Start with their essential details.
                </p>
              </div>
              <button
                aria-label="Close registration"
                onClick={() => setModal(false)}
              >
                <X size={19} />
              </button>
            </div>
            <form className="space-y-5" onSubmit={register}>
              {formError && (
                <p role="alert" className="error">
                  {formError}
                </p>
              )}
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label className="label" htmlFor="givenName">
                    First name
                  </label>
                  <input
                    autoFocus
                    className="field"
                    id="givenName"
                    name="givenName"
                    required
                    maxLength={80}
                    autoComplete="off"
                  />
                </div>
                <div>
                  <label className="label" htmlFor="familyName">
                    Last name
                  </label>
                  <input
                    className="field"
                    id="familyName"
                    name="familyName"
                    required
                    maxLength={80}
                    autoComplete="off"
                  />
                </div>
              </div>
              <div>
                <label className="label" htmlFor="dateOfBirth">
                  Date of birth
                </label>
                <input
                  className="field"
                  id="dateOfBirth"
                  name="dateOfBirth"
                  type="date"
                  required
                  max={new Date().toISOString().slice(0, 10)}
                />
              </div>
              <div>
                <label className="label" htmlFor="patientPhone">
                  Phone number{" "}
                  <span className="font-normal text-muted">· optional</span>
                </label>
                <input
                  className="field"
                  id="patientPhone"
                  name="phone"
                  type="tel"
                  pattern="\+[1-9][0-9]{7,14}"
                  placeholder="+254700000000"
                />
              </div>
              <p className="text-xs leading-5 text-muted">
                Registration creates a clinical record. Portal access and
                care-team assignment will be managed separately.
              </p>
              <div className="flex justify-end gap-3 border-t border-slate-100 pt-5">
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setModal(false)}
                >
                  Cancel
                </button>
                <button className="primary" disabled={busy}>
                  {busy ? "Saving…" : "Register patient"}
                </button>
              </div>
            </form>
          </section>
        </Modal>
      )}
    </div>
  );
}
