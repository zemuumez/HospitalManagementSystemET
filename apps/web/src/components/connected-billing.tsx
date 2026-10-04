"use client";
import { useCallback, useState } from "react";
import { api, type Patient } from "@/lib/api";
import { useIdentity } from "./workspace";
import { useLanguage } from "./language";
import {
  useResource,
  Table,
  Status,
  Editor,
  Paging,
  Input,
  errorText,
} from "./connected-scheduling";
type Account = { id: string; name: string };
type Line = {
  accountId: string;
  accountName: string;
  description: string;
  quantity: number;
  unitPriceMinor: number;
};
type Invoice = {
  id: string;
  number: number;
  patientName: string;
  invoiceDate: string;
  subtotalMinor: number;
  discountBasisPoints: number;
  totalMinor: number;
  paidMinor: number;
  version: number;
  lines: Line[];
};
type Payment = {
  id: string;
  amountMinor: number;
  method: string;
  reference: string;
  originalPaymentId: string;
  reason: string;
  direction: string;
  createdAt: string;
};
function money(value: number) {
  return `${(value / 100).toFixed(2)} ETB`;
}
function minor(value: string) {
  if (!/^\d{1,10}(\.\d{1,2})?$/.test(value))
    throw Error("Enter an amount with at most two decimal places");
  const [whole, fraction = ""] = value.split(".");
  return Number(whole) * 100 + Number(fraction.padEnd(2, "0"));
}

export function ConnectedAccounts() {
  const { t } = useLanguage();
  const [page, setPage] = useState(1),
    [open, setOpen] = useState(false),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  const load = useCallback(
      () => api<{ accounts: Account[] }>(`charge-accounts?page=${page}`),
      [page],
    ),
    resource = useResource(load);
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api("charge-accounts", {
        method: "POST",
        body: JSON.stringify(Object.fromEntries(new FormData(e.currentTarget))),
      });
      setOpen(false);
      resource.reload();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <div className="page-heading">
        <h1>{t("Accounts")}</h1>
        <button
          className="primary"
          onClick={() => {
            setError("");
            setOpen(true);
          }}
        >
          {t("New Account")}
        </button>
      </div>
      <Status {...resource} empty={!resource.data?.accounts.length} />
      {resource.data && !resource.loading && (
        <Table head={["Name"]}>
          {resource.data.accounts.map((a) => (
            <tr key={a.id}>
              <td>{a.name}</td>
            </tr>
          ))}
        </Table>
      )}
      <Paging
        page={page}
        setPage={setPage}
        count={resource.data?.accounts.length || 0}
        loading={resource.loading}
      />
      {open && (
        <Editor
          title="New Account"
          onClose={() => {
            if (!busy) setOpen(false);
          }}
        >
          <form onSubmit={save} className="space-y-4">
            {error && (
              <p className="error" role="alert">
                {t(error)}
              </p>
            )}
            <Input label="Name" name="name" required maxLength={100} />
            <button disabled={busy} className="primary">
              {t(busy ? "Saving…" : "Save")}
            </button>
          </form>
        </Editor>
      )}
    </>
  );
}

export function ConnectedInvoices() {
  const { t } = useLanguage(),
    identity = useIdentity();
  const [page, setPage] = useState(1),
    [open, setOpen] = useState(false),
    [selected, setSelected] = useState<string | null>(null);
  const load = useCallback(
      () => api<{ invoices: Invoice[] }>(`invoices?page=${page}`),
      [page],
    ),
    resource = useResource(load);
  return (
    <>
      <div className="page-heading">
        <h1>{t("Invoices")}</h1>
        {identity?.permissions.includes("billing.manage") && (
          <button className="primary" onClick={() => setOpen(true)}>
            {t("New Invoice")}
          </button>
        )}
      </div>
      <Status {...resource} empty={!resource.data?.invoices.length} />
      {resource.data && !resource.loading && (
        <Table
          head={[
            "Invoice",
            "Patient",
            "Date",
            "Amount",
            "Paid",
            "Balance",
            "Action",
          ]}
        >
          {resource.data.invoices.map((i) => (
            <tr key={i.id}>
              <td>INV-{String(i.number).padStart(6, "0")}</td>
              <td>{i.patientName}</td>
              <td>{i.invoiceDate}</td>
              <td>{money(i.totalMinor)}</td>
              <td>{money(i.paidMinor)}</td>
              <td>{money(i.totalMinor - i.paidMinor)}</td>
              <td>
                <button className="secondary" onClick={() => setSelected(i.id)}>
                  {t("View")}
                </button>
              </td>
            </tr>
          ))}
        </Table>
      )}
      <Paging
        page={page}
        setPage={setPage}
        count={resource.data?.invoices.length || 0}
        loading={resource.loading}
      />
      {open && (
        <CreateInvoice
          onClose={() => setOpen(false)}
          onSaved={resource.reload}
        />
      )}{" "}
      {selected && (
        <InvoiceDetails
          id={selected}
          onClose={() => setSelected(null)}
          onSaved={resource.reload}
        />
      )}
    </>
  );
}

function CreateInvoice({
  onClose,
  onSaved,
}: {
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useLanguage();
  const [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [search, setSearch] = useState(""),
    [patients, setPatients] = useState<Patient[]>([]),
    [accountPage, setAccountPage] = useState(1),
    [accounts, setAccounts] = useState<Account[]>([]),
    [more, setMore] = useState(true);
  const [lines, setLines] = useState([
      { accountId: "", description: "", quantity: "1", price: "" },
    ]),
    [request, setRequest] = useState<{ key: string; body: string } | null>(
      null,
    );
  async function find() {
    setError("");
    try {
      setPatients(
        (
          await api<{ patients: Patient[] }>(
            `billing-patients?search=${encodeURIComponent(search)}`,
          )
        ).patients,
      );
    } catch (e) {
      setError(errorText(e));
    }
  }
  async function loadAccounts() {
    setError("");
    try {
      const out = await api<{ accounts: Account[] }>(
        `charge-accounts?page=${accountPage}`,
      );
      setAccounts((old) => [
        ...old,
        ...out.accounts.filter((a) => !old.some((o) => o.id === a.id)),
      ]);
      setMore(out.accounts.length === 25);
      setAccountPage((p) => p + 1);
    } catch (e) {
      setError(errorText(e));
    }
  }
  function change(index: number, patch: Partial<(typeof lines)[number]>) {
    setLines((old) =>
      old.map((line, n) => (n === index ? { ...line, ...patch } : line)),
    );
  }
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const data = Object.fromEntries(new FormData(e.currentTarget));
      const payload = {
        patientId: data.patientId,
        invoiceDate: data.invoiceDate,
        discountBasisPoints: minor(String(data.discount)),
        lines: lines.map((l) => ({
          accountId: l.accountId,
          accountName: "",
          description: l.description,
          quantity: Number(l.quantity),
          unitPriceMinor: minor(l.price),
        })),
      };
      const body = JSON.stringify(payload),
        key = request?.body === body ? request.key : crypto.randomUUID();
      setRequest({ key, body });
      await api("invoices", {
        method: "POST",
        headers: { "Idempotency-Key": key },
        body,
      });
      onSaved();
      onClose();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Editor
      title="New Invoice"
      onClose={() => {
        if (!busy) onClose();
      }}
    >
      <form onSubmit={save} className="space-y-4">
        {error && (
          <p className="error" role="alert">
            {t(error)}
          </p>
        )}
        <div className="flex gap-2">
          <Input
            label="Search patients"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            maxLength={80}
          />
          <button type="button" className="secondary" onClick={find}>
            {t("Search")}
          </button>
        </div>
        <label>
          <span className="label">{t("Patient")}</span>
          <select
            aria-label={t("Patient")}
            name="patientId"
            className="field"
            required
            defaultValue=""
          >
            <option value="">{t("Select Patient")}</option>
            {patients.map((p) => (
              <option key={p.id} value={p.id}>
                {p.givenName} {p.familyName} — {p.mrn}
              </option>
            ))}
          </select>
        </label>
        <Input
          label="Invoice Date"
          name="invoiceDate"
          type="date"
          required
          defaultValue={new Date().toLocaleDateString("en-CA", {
            timeZone: "Africa/Addis_Ababa",
          })}
        />
        <Input
          label="Discount (%)"
          name="discount"
          type="number"
          required
          min={0}
          max={100}
          step="0.01"
          defaultValue="0"
        />
        <button
          type="button"
          className="secondary"
          disabled={!more}
          onClick={loadAccounts}
        >
          {t(accounts.length ? "Load more accounts" : "Load accounts")}
        </button>
        {lines.map((l, n) => (
          <fieldset className="border rounded p-3 space-y-3" key={n}>
            <legend>
              {t("Item")} {n + 1}
            </legend>
            <select
              aria-label={`${t("Account")} ${n + 1}`}
              className="field"
              required
              value={l.accountId}
              onChange={(e) => change(n, { accountId: e.target.value })}
            >
              <option value="">{t("Select Account")}</option>
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
            <Input
              label="Description"
              value={l.description}
              maxLength={500}
              onChange={(e) => change(n, { description: e.target.value })}
            />
            <div className="grid grid-cols-2 gap-3">
              <Input
                label="Quantity"
                type="number"
                min={1}
                max={10000}
                step={1}
                required
                value={l.quantity}
                onChange={(e) => change(n, { quantity: e.target.value })}
              />
              <Input
                label="Unit Price (ETB)"
                inputMode="decimal"
                required
                value={l.price}
                onChange={(e) => change(n, { price: e.target.value })}
              />
            </div>
            <button
              className="secondary"
              type="button"
              disabled={lines.length === 1}
              onClick={() => setLines((old) => old.filter((_, i) => i !== n))}
            >
              {t("Remove")}
            </button>
          </fieldset>
        ))}
        <button
          type="button"
          className="secondary"
          disabled={lines.length >= 100}
          onClick={() =>
            setLines((old) => [
              ...old,
              { accountId: "", description: "", quantity: "1", price: "" },
            ])
          }
        >
          {t("Add Item")}
        </button>
        <p className="text-sm text-muted">
          {t(
            "The server calculates the final total. Issued invoices cannot be edited in this workflow.",
          )}
        </p>
        <button className="primary" disabled={busy}>
          {t(busy ? "Saving…" : "Issue invoice")}
        </button>
      </form>
    </Editor>
  );
}

function InvoiceDetails({
  id,
  onClose,
  onSaved,
}: {
  id: string;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useLanguage(),
    identity = useIdentity();
  const loader = useCallback(async () => {
      const [invoice, payments] = await Promise.all([
        api<Invoice>(`invoices/${id}`),
        api<{ payments: Payment[] }>(`invoices/${id}/payments`),
      ]);
      return { invoice, ...payments };
    }, [id]),
    resource = useResource(loader);
  const [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [refund, setRefund] = useState(""),
    [request, setRequest] = useState<{ key: string; body: string } | null>(
      null,
    );
  async function post(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = e.currentTarget;
    setError("");
    setBusy(true);
    try {
      const data = Object.fromEntries(new FormData(e.currentTarget)),
        payload = {
          amountMinor: minor(String(data.amount)),
          method: data.method,
          reference: data.reference,
          reason: data.reason,
          originalPaymentId: refund,
        };
      const body = JSON.stringify(payload),
        key = request?.body === body ? request.key : crypto.randomUUID();
      setRequest({ key, body });
      await api(`invoices/${id}/payments`, {
        method: "POST",
        headers: { "Idempotency-Key": key },
        body,
      });
      resource.reload();
      onSaved();
      form.reset();
      setRefund("");
    } catch (e) {
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }
  const i = resource.data?.invoice;
  return (
    <Editor
      title="Invoice"
      wide
      onClose={() => {
        if (!busy) onClose();
      }}
    >
      <div className="space-y-5">
        <Status {...resource} empty={false} />
        {i && (
          <>
            <p>
              INV-{String(i.number).padStart(6, "0")} · {i.patientName}
            </p>
            <Table head={["Account", "Quantity", "Price", "Amount"]}>
              {i.lines.map((l, n) => (
                <tr key={n}>
                  <td>
                    {l.accountName}
                    <small className="block">{l.description}</small>
                  </td>
                  <td>{l.quantity}</td>
                  <td>{money(l.unitPriceMinor)}</td>
                  <td>{money(l.quantity * l.unitPriceMinor)}</td>
                </tr>
              ))}
            </Table>
            <p>
              {t("Subtotal")}: {money(i.subtotalMinor)} · {t("Discount")}:{" "}
              {(i.discountBasisPoints / 100).toFixed(2)}%
            </p>
            <p>
              {t("Total")}: <strong>{money(i.totalMinor)}</strong> ·{" "}
              {t("Balance")}:{" "}
              <strong>{money(i.totalMinor - i.paidMinor)}</strong>
            </p>
            <h3>{t("Payments")}</h3>
            {resource.data?.payments.length ? (
              <Table head={["Type", "Amount", "Method", "Reference"]}>
                {resource.data.payments.map((p) => (
                  <tr key={p.id}>
                    <td>{t(p.direction)}</td>
                    <td>{money(p.amountMinor)}</td>
                    <td>{t(p.method)}</td>
                    <td>{p.reference || "—"}</td>
                  </tr>
                ))}
              </Table>
            ) : (
              <p>{t("No payments recorded")}</p>
            )}
            {identity?.permissions.includes("billing.manage") && (
              <form onSubmit={post} className="space-y-4 border-t pt-4">
                {error && (
                  <p className="error" role="alert">
                    {t(error)}
                  </p>
                )}
                <label>
                  <span className="label">{t("Transaction")}</span>
                  <select
                    aria-label={t("Transaction")}
                    className="field"
                    value={refund}
                    onChange={(e) => setRefund(e.target.value)}
                  >
                    <option value="">{t("Record payment")}</option>
                    {resource.data?.payments
                      .filter((p) => p.direction === "payment")
                      .map((p) => (
                        <option key={p.id} value={p.id}>
                          {t("Refund")} {money(p.amountMinor)} ·{" "}
                          {new Date(p.createdAt).toLocaleString()}
                        </option>
                      ))}
                  </select>
                </label>
                <Input
                  label="Amount (ETB)"
                  name="amount"
                  inputMode="decimal"
                  required
                />
                <select
                  className="field"
                  aria-label={t("Payment method")}
                  name="method"
                  defaultValue="cash"
                >
                  <option value="cash">{t("Cash")}</option>
                  <option value="bank">{t("Bank transfer")}</option>
                </select>
                <Input label="Reference" name="reference" maxLength={100} />
                <Input
                  label="Reason"
                  name="reason"
                  required={!!refund}
                  maxLength={500}
                />
                <p className="text-sm text-muted">
                  {t(
                    "Record only money already received or refunded. This form does not initiate a bank transfer.",
                  )}
                </p>
                <button className="primary" disabled={busy}>
                  {t(
                    busy
                      ? "Saving…"
                      : refund
                        ? "Record refund"
                        : "Record payment",
                  )}
                </button>
              </form>
            )}
          </>
        )}
      </div>
    </Editor>
  );
}
