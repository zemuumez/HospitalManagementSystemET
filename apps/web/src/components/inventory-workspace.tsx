"use client";

import { useEffect, useRef, useState } from "react";
import {
  Boxes,
  Layers,
  ArrowDownToLine,
  ArrowUpFromLine,
  Plus,
  RefreshCw,
  Search,
  Package,
  AlertTriangle,
  Building,
  CheckCircle2,
  X,
  Tag,
  DollarSign,
  TrendingDown,
} from "lucide-react";
import { useLanguage } from "./language";
import { Modal } from "./modal";
import { api } from "@/lib/api";

export type InventoryTab =
  "items" | "item-categories" | "item-stocks" | "issued-items";

interface InventoryCategory {
  id: string;
  name: string;
  description: string;
  active: boolean;
  version: number;
}

interface InventoryItem {
  id: string;
  categoryId: string;
  name: string;
  unit: string;
  description: string;
  reorderMilli: number;
  balanceMilli: number;
  active: boolean;
  version: number;
}

interface InventoryMovement {
  id: string;
  itemId: string;
  kind: string; // receive, issue, return, adjust, waste
  quantityMilli: number;
  deltaMilli: number;
  recipientId?: string;
  supplier?: string;
  storeName?: string;
  reference?: string;
  costMinor?: number;
  reason?: string;
  createdAt: string;
}

export function InventoryWorkspace({ id = "items" }: { id?: string }) {
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<InventoryTab>(
    (id as InventoryTab) || "items",
  );

  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(false);
  const [isLive, setIsLive] = useState(false);

  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const submitting = useRef(false);
  const pending = useRef<{ signature: string; key: string } | null>(null);

  // Data states
  const [categories, setCategories] = useState<InventoryCategory[]>([]);
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [movements, setMovements] = useState<InventoryMovement[]>([]);

  // Modal open states
  const [showAddItem, setShowAddItem] = useState(false);
  const [showAddCategory, setShowAddCategory] = useState(false);
  const [showAddStock, setShowAddStock] = useState(false);
  const [showIssueItem, setShowIssueItem] = useState(false);
  const [recipientSearch, setRecipientSearch] = useState("");
  const [recipients, setRecipients] = useState<
    { id: string; name: string; role: string; active: boolean }[]
  >([]);
  useEffect(() => {
    if (!showIssueItem) return;
    const controller = new AbortController();
    const timer = setTimeout(async () => {
      try {
        const response = await fetch(
          `/api/staff?search=${encodeURIComponent(recipientSearch)}`,
          { signal: controller.signal },
        );
        if (!response.ok) throw new Error("Unable to load staff");
        const data = await response.json();
        setRecipients(
          data.users.filter(
            (user: { active: boolean; role: string }) =>
              user.active && user.role !== "patient",
          ),
        );
      } catch (cause) {
        if (!controller.signal.aborted)
          setError(
            cause instanceof Error ? cause.message : "Unable to load staff",
          );
      }
    }, 250);
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [showIssueItem, recipientSearch]);

  // Form states
  const [itemForm, setItemForm] = useState({
    name: "",
    categoryId: "",
    unit: "Piece",
    reorderLevel: 10,
    description: "",
  });

  const [categoryForm, setCategoryForm] = useState({
    name: "",
    description: "",
  });

  const [stockForm, setStockForm] = useState({
    itemId: "",
    quantity: 50,
    supplier: "",
    storeName: "Central Hospital Store",
    reference: "",
    unitCost: 15.0,
    reason: "Routine replenishment",
  });

  const [issueForm, setIssueForm] = useState({
    itemId: "",
    quantity: 5,
    recipientId: "",
    reason: "Departmental clinical supply",
  });

  const fetchInventoryData = async () => {
    setLoading(true);
    setError("");
    try {
      const [cats, stock, history] = await Promise.all([
        api<{ categories: InventoryCategory[] }>("inventory/categories"),
        api<{ items: InventoryItem[] }>("inventory/items"),
        api<{ movements: InventoryMovement[] }>("inventory/movements"),
      ]);
      setCategories(cats.categories);
      setItems(stock.items);
      setMovements(history.movements);
      setIsLive(true);
    } catch (cause) {
      setIsLive(false);
      setError(
        cause instanceof Error ? cause.message : "Unable to load inventory",
      );
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchInventoryData();
  }, []);

  async function save(path: string, payload: object, onSaved: () => void) {
    if (submitting.current) return;
    submitting.current = true;
    setSaving(true);
    setError("");
    const signature = JSON.stringify({ path, payload });
    if (pending.current?.signature !== signature) {
      pending.current = { signature, key: crypto.randomUUID() };
    }
    try {
      await api(path, {
        method: "POST",
        headers: { "Idempotency-Key": pending.current.key },
        body: JSON.stringify(payload),
      });
      pending.current = null;
      onSaved();
      await fetchInventoryData();
    } catch (cause) {
      setError(
        cause instanceof Error ? cause.message : "Unable to save inventory",
      );
    } finally {
      submitting.current = false;
      setSaving(false);
    }
  }

  const handleSaveCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    await save(
      "inventory/categories",
      { ...categoryForm, active: true, version: 1 },
      () => {
        setShowAddCategory(false);
        setCategoryForm({ name: "", description: "" });
      },
    );
  };
  const handleSaveItem = async (e: React.FormEvent) => {
    e.preventDefault();
    await save(
      "inventory/items",
      {
        categoryId: itemForm.categoryId || categories[0]?.id || "",
        name: itemForm.name,
        unit: itemForm.unit,
        description: itemForm.description,
        reorderMilli: Number(itemForm.reorderLevel) * 1000,
        active: true,
        version: 1,
      },
      () => {
        setShowAddItem(false);
        setItemForm({
          name: "",
          categoryId: "",
          unit: "Piece",
          reorderLevel: 10,
          description: "",
        });
      },
    );
  };
  const handleSaveStockReceive = async (e: React.FormEvent) => {
    e.preventDefault();
    await save(
      "inventory/movements",
      {
        itemId: stockForm.itemId || items[0]?.id || "",
        kind: "receive",
        quantityMilli: Number(stockForm.quantity) * 1000,
        costMinor: Math.round(
          Number(stockForm.unitCost) * Number(stockForm.quantity) * 100,
        ),
        supplier: stockForm.supplier,
        storeName: stockForm.storeName,
        reference: stockForm.reference,
        reason: stockForm.reason,
      },
      () => setShowAddStock(false),
    );
  };
  const handleSaveIssue = async (e: React.FormEvent) => {
    e.preventDefault();
    await save(
      "inventory/movements",
      {
        itemId: issueForm.itemId || items[0]?.id || "",
        kind: "issue",
        quantityMilli: Number(issueForm.quantity) * 1000,
        recipientId: issueForm.recipientId,
        reason: issueForm.reason,
      },
      () => setShowIssueItem(false),
    );
  };

  const getCategoryName = (catId: string) => {
    return categories.find((c) => c.id === catId)?.name || t("General");
  };

  const getItemName = (itemId: string) => {
    return items.find((i) => i.id === itemId)?.name || t("Medical Supply Item");
  };

  const getItemUnit = (itemId: string) => {
    return items.find((i) => i.id === itemId)?.unit || "Units";
  };

  const tabs = [
    { id: "items", label: t("Items"), icon: Package, count: items.length },
    {
      id: "item-categories",
      label: t("Item Categories"),
      icon: Layers,
      count: categories.length,
    },
    {
      id: "item-stocks",
      label: t("Item Stocks"),
      icon: ArrowDownToLine,
      count: movements.filter((m) => m.kind === "receive").length,
    },
    {
      id: "issued-items",
      label: t("Issued Items"),
      icon: ArrowUpFromLine,
      count: movements.filter((m) => m.kind === "issue").length,
    },
  ];

  return (
    <div className="space-y-6">
      {error && (
        <p role="alert" className="error">
          {t(error)}
        </p>
      )}
      {/* Subtabs Nav */}
      <div className="border-b border-border/80 bg-card/50 backdrop-blur rounded-xl p-1.5 shadow-sm">
        <nav className="flex space-x-1 overflow-x-auto">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as InventoryTab)}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 whitespace-nowrap ${
                  isActive
                    ? "bg-primary text-primary-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{tab.label}</span>
                <span
                  className={`text-xs px-1.5 py-0.5 rounded-full ${
                    isActive
                      ? "bg-primary-foreground/20 text-primary-foreground font-semibold"
                      : "bg-muted text-muted-foreground"
                  }`}
                >
                  {tab.count}
                </span>
              </button>
            );
          })}
        </nav>
      </div>

      {/* Live Status Banner */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-gradient-to-r from-emerald-500/10 via-emerald-500/5 to-transparent border border-emerald-500/20 rounded-xl p-3.5 text-sm shadow-sm">
        <div className="flex items-center gap-2.5">
          <span className="relative flex h-2.5 w-2.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
          </span>
          <span className="font-semibold text-emerald-950 dark:text-emerald-200">
            {t("Inventory Management")}
          </span>
          <span className="text-muted-foreground hidden sm:inline">•</span>
          <span className="text-muted-foreground text-xs sm:text-sm">
            {isLive
              ? t("Inventory loaded")
              : t("Inventory is not synchronized")}
          </span>
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto justify-end">
          <button
            onClick={() => fetchInventoryData()}
            disabled={loading}
            className="btn-secondary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg border border-border shadow-2xs hover:bg-muted"
          >
            <RefreshCw
              className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`}
            />
            {t("Sync")}
          </button>
          {activeTab === "items" && (
            <button
              onClick={() => setShowAddItem(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Item")}
            </button>
          )}
          {activeTab === "item-categories" && (
            <button
              onClick={() => setShowAddCategory(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Item Category")}
            </button>
          )}
          {activeTab === "item-stocks" && (
            <button
              onClick={() => setShowAddStock(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("Receive New Stock")}
            </button>
          )}
          {activeTab === "issued-items" && (
            <button
              onClick={() => setShowIssueItem(true)}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("Issue Item")}
            </button>
          )}
        </div>
      </div>

      {/* 1. ITEMS VIEW */}
      {activeTab === "items" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Medical Inventory Items")}
            </h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search items...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Item Name")}</th>
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Unit")}</th>
                  <th className="px-4 py-3">{t("Available Quantity")}</th>
                  <th className="px-4 py-3">{t("Reorder Level")}</th>
                  <th className="px-4 py-3">{t("Status")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {items
                  .filter(
                    (i) =>
                      i.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      getCategoryName(i.categoryId)
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()),
                  )
                  .map((item) => {
                    const balance = item.balanceMilli / 1000;
                    const reorder = item.reorderMilli / 1000;
                    const isLow = balance <= reorder;
                    return (
                      <tr
                        key={item.id}
                        className="hover:bg-muted/30 transition-colors"
                      >
                        <td className="px-4 py-3 font-semibold text-foreground">
                          {item.name}
                        </td>
                        <td className="px-4 py-3">
                          <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">
                            {getCategoryName(item.categoryId)}
                          </span>
                        </td>
                        <td className="px-4 py-3 font-mono text-xs">
                          {item.unit}
                        </td>
                        <td className="px-4 py-3 font-bold font-mono text-foreground">
                          {balance.toLocaleString()} {item.unit}
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                          {reorder.toLocaleString()} {item.unit}
                        </td>
                        <td className="px-4 py-3">
                          {isLow ? (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-red-500/10 text-red-600 dark:text-red-400">
                              <AlertTriangle className="w-3.5 h-3.5" />{" "}
                              {t("Low Stock")}
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                              <CheckCircle2 className="w-3.5 h-3.5" />{" "}
                              {t("In Stock")}
                            </span>
                          )}
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 2. ITEM CATEGORIES VIEW */}
      {activeTab === "item-categories" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Inventory Categories")}
            </h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search categories...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Category Name")}</th>
                  <th className="px-4 py-3">{t("Description")}</th>
                  <th className="px-4 py-3">{t("Item Count")}</th>
                  <th className="px-4 py-3">{t("Active")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {categories
                  .filter((c) =>
                    c.name.toLowerCase().includes(searchTerm.toLowerCase()),
                  )
                  .map((cat) => {
                    const itemCount = items.filter(
                      (i) => i.categoryId === cat.id,
                    ).length;
                    return (
                      <tr
                        key={cat.id}
                        className="hover:bg-muted/30 transition-colors"
                      >
                        <td className="px-4 py-3 font-semibold text-foreground">
                          {cat.name}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {cat.description || "-"}
                        </td>
                        <td className="px-4 py-3 font-mono text-xs">
                          {itemCount} items
                        </td>
                        <td className="px-4 py-3">
                          <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                            {t("Active")}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 3. ITEM STOCKS / RECEIVE MOVEMENTS */}
      {activeTab === "item-stocks" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Item Stocks Received")}
            </h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search received stock...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Item")}</th>
                  <th className="px-4 py-3">{t("Supplier / Store")}</th>
                  <th className="px-4 py-3">{t("Reference No")}</th>
                  <th className="px-4 py-3">{t("Quantity Received")}</th>
                  <th className="px-4 py-3">{t("Total Cost (ETB)")}</th>
                  <th className="px-4 py-3">{t("Date")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {movements
                  .filter((m) => m.kind === "receive")
                  .filter(
                    (m) =>
                      getItemName(m.itemId)
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()) ||
                      (m.reference || "")
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()),
                  )
                  .map((m) => (
                    <tr
                      key={m.id}
                      className="hover:bg-muted/30 transition-colors"
                    >
                      <td className="px-4 py-3 font-semibold text-foreground">
                        {getItemName(m.itemId)}
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-xs text-foreground">
                          {m.supplier || "-"}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {m.storeName}
                        </div>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs font-semibold">
                        {m.reference || "-"}
                      </td>
                      <td className="px-4 py-3 font-bold font-mono text-emerald-600 dark:text-emerald-400">
                        +{(m.quantityMilli / 1000).toLocaleString()}{" "}
                        {getItemUnit(m.itemId)}
                      </td>
                      <td className="px-4 py-3 font-mono text-xs">
                        {m.costMinor
                          ? `${(m.costMinor / 100).toLocaleString(undefined, {
                              minimumFractionDigits: 2,
                            })} ETB`
                          : "-"}
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {new Date(m.createdAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 4. ISSUED ITEMS */}
      {activeTab === "issued-items" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col sm:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Issued Items & Consumables")}
            </h3>
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("Search issued items...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-sm bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Item")}</th>
                  <th className="px-4 py-3">{t("Issued To / Recipient")}</th>
                  <th className="px-4 py-3">{t("Quantity Issued")}</th>
                  <th className="px-4 py-3">{t("Reason / Purpose")}</th>
                  <th className="px-4 py-3">{t("Issued Date")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {movements
                  .filter((m) => m.kind === "issue")
                  .filter(
                    (m) =>
                      getItemName(m.itemId)
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()) ||
                      (m.recipientId || "")
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()),
                  )
                  .map((m) => (
                    <tr
                      key={m.id}
                      className="hover:bg-muted/30 transition-colors"
                    >
                      <td className="px-4 py-3 font-semibold text-foreground">
                        {getItemName(m.itemId)}
                      </td>
                      <td className="px-4 py-3 font-medium text-xs text-foreground">
                        {m.recipientId || "-"}
                      </td>
                      <td className="px-4 py-3 font-bold font-mono text-amber-600 dark:text-amber-400">
                        -{(m.quantityMilli / 1000).toLocaleString()}{" "}
                        {getItemUnit(m.itemId)}
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {m.reason || "-"}
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {new Date(m.createdAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* --- MODALS --- */}

      {/* 1. New Item Modal */}
      {showAddItem && (
        <Modal onClose={() => setShowAddItem(false)} titleId="add-item-title">
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-item-title" className="font-semibold text-lg">
                {t("New Inventory Item")}
              </h3>
              <button
                onClick={() => setShowAddItem(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveItem} className="space-y-4">
              {error && (
                <p role="alert" className="error">
                  {t(error)}
                </p>
              )}
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Item Name")} *
                </label>
                <input
                  type="text"
                  required
                  aria-label={t("Item Name")}
                  value={itemForm.name}
                  onChange={(e) =>
                    setItemForm({ ...itemForm, name: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  placeholder="e.g. Sterile Syringes 5ml"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Category")} *
                  </label>
                  <select
                    aria-label={t("Category")}
                    value={itemForm.categoryId}
                    onChange={(e) =>
                      setItemForm({ ...itemForm, categoryId: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    {categories.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Unit")} *
                  </label>
                  <input
                    type="text"
                    required
                    aria-label={t("Unit")}
                    value={itemForm.unit}
                    onChange={(e) =>
                      setItemForm({ ...itemForm, unit: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                    placeholder="Piece, Box, Vial, Kit"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Reorder Alert Level")}
                </label>
                <input
                  type="number"
                  min={1}
                  aria-label={t("Reorder Level")}
                  value={itemForm.reorderLevel}
                  onChange={(e) =>
                    setItemForm({
                      ...itemForm,
                      reorderLevel: Number(e.target.value),
                    })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Description")}
                </label>
                <textarea
                  rows={3}
                  aria-label={t("Description")}
                  value={itemForm.description}
                  onChange={(e) =>
                    setItemForm({ ...itemForm, description: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddItem(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Save Item")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 2. New Category Modal */}
      {showAddCategory && (
        <Modal
          onClose={() => setShowAddCategory(false)}
          titleId="add-category-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="add-category-title" className="font-semibold text-lg">
                {t("New Item Category")}
              </h3>
              <button
                onClick={() => setShowAddCategory(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveCategory} className="space-y-4">
              {error && (
                <p role="alert" className="error">
                  {t(error)}
                </p>
              )}
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Category Name")} *
                </label>
                <input
                  type="text"
                  required
                  aria-label={t("Category Name")}
                  value={categoryForm.name}
                  onChange={(e) =>
                    setCategoryForm({ ...categoryForm, name: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  placeholder="e.g. Diagnostics Consumables"
                />
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Description")}
                </label>
                <textarea
                  rows={3}
                  aria-label={t("Description")}
                  value={categoryForm.description}
                  onChange={(e) =>
                    setCategoryForm({
                      ...categoryForm,
                      description: e.target.value,
                    })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddCategory(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Save Category")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 3. Receive Stock Modal */}
      {showAddStock && (
        <Modal
          onClose={() => setShowAddStock(false)}
          titleId="receive-stock-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="receive-stock-title" className="font-semibold text-lg">
                {t("Receive New Stock")}
              </h3>
              <button
                onClick={() => setShowAddStock(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveStockReceive} className="space-y-4">
              {error && (
                <p role="alert" className="error">
                  {t(error)}
                </p>
              )}
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Item")} *
                </label>
                <select
                  value={stockForm.itemId || (items[0]?.id ?? "")}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, itemId: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                >
                  {items.map((i) => (
                    <option key={i.id} value={i.id}>
                      {i.name} ({i.unit})
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Quantity")} *
                  </label>
                  <input
                    type="number"
                    min={1}
                    required
                    aria-label={t("Quantity")}
                    value={stockForm.quantity}
                    onChange={(e) =>
                      setStockForm({
                        ...stockForm,
                        quantity: Number(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Unit Cost (ETB)")}
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    min={0}
                    aria-label={t("Unit Cost")}
                    value={stockForm.unitCost}
                    onChange={(e) =>
                      setStockForm({
                        ...stockForm,
                        unitCost: Number(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Supplier")}
                  </label>
                  <input
                    type="text"
                    aria-label={t("Supplier")}
                    value={stockForm.supplier}
                    onChange={(e) =>
                      setStockForm({ ...stockForm, supplier: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                    placeholder="Supplier name"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Store Name")}
                  </label>
                  <input
                    type="text"
                    aria-label={t("Store Name")}
                    value={stockForm.storeName}
                    onChange={(e) =>
                      setStockForm({ ...stockForm, storeName: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Reference No")}
                </label>
                <input
                  type="text"
                  aria-label={t("Reference")}
                  value={stockForm.reference}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, reference: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg font-mono"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowAddStock(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Confirm Receive")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 4. Issue Item Modal */}
      {showIssueItem && (
        <Modal
          onClose={() => setShowIssueItem(false)}
          titleId="issue-item-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="issue-item-title" className="font-semibold text-lg">
                {t("Issue Inventory Item")}
              </h3>
              <button
                onClick={() => setShowIssueItem(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveIssue} className="space-y-4">
              {error && (
                <p role="alert" className="error">
                  {t(error)}
                </p>
              )}
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Item")} *
                </label>
                <select
                  value={issueForm.itemId || (items[0]?.id ?? "")}
                  onChange={(e) =>
                    setIssueForm({ ...issueForm, itemId: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                >
                  {items.map((i) => (
                    <option key={i.id} value={i.id}>
                      {i.name} (Available:{" "}
                      {(i.balanceMilli / 1000).toLocaleString()} {i.unit})
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Quantity to Issue")} *
                  </label>
                  <input
                    type="number"
                    min={1}
                    required
                    aria-label={t("Quantity to Issue")}
                    value={issueForm.quantity}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        quantity: Number(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Recipient")} *
                  </label>
                  <input
                    aria-label={t("Search staff")}
                    placeholder={t("Search staff")}
                    value={recipientSearch}
                    onChange={(e) => setRecipientSearch(e.target.value)}
                    className="field"
                  />
                  <select
                    aria-label={t("Recipient")}
                    required
                    value={issueForm.recipientId}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        recipientId: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    <option value="">{t("Select staff")}</option>
                    {recipients.map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.name} ({t(user.role)})
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Reason / Clinical Notes")}
                </label>
                <textarea
                  rows={2}
                  aria-label={t("Reason")}
                  value={issueForm.reason}
                  onChange={(e) =>
                    setIssueForm({ ...issueForm, reason: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setShowIssueItem(false)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Confirm Issue")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}
    </div>
  );
}
