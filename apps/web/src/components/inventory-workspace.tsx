"use client";

import { useEffect, useRef, useState, useMemo } from "react";
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
  RotateCcw,
  Pencil,
  Trash2,
  Eye,
  Info,
  Check,
  Clock,
  Filter,
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
  kind: string; // receive, issue, return, writeoff
  quantityMilli: number;
  deltaMilli: number;
  returnedMilli?: number;
  recipientId?: string;
  originalId?: string;
  supplier?: string;
  storeName?: string;
  reference?: string;
  costMinor?: number;
  restock?: boolean;
  reason?: string;
  createdAt: string;
}

interface StaffUser {
  id: string;
  name: string;
  role: string;
  active: boolean;
}

export function InventoryWorkspace({ id = "items" }: { id?: string }) {
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<InventoryTab>(
    (id as InventoryTab) || "items",
  );

  const [searchTerm, setSearchTerm] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [lowStockOnly, setLowStockOnly] = useState(false);
  const [issueStatusFilter, setIssueStatusFilter] = useState<
    "all" | "returnable" | "returned"
  >("all");

  const [loading, setLoading] = useState(false);
  const [isLive, setIsLive] = useState(false);

  const [error, setError] = useState("");
  const [successMsg, setSuccessMsg] = useState("");
  const [saving, setSaving] = useState(false);
  const submitting = useRef(false);
  const pending = useRef<{ signature: string; key: string } | null>(null);

  // Data states
  const [categories, setCategories] = useState<InventoryCategory[]>([]);
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [movements, setMovements] = useState<InventoryMovement[]>([]);

  // Modal open states
  const [showAddItem, setShowAddItem] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);
  const [viewingItem, setViewingItem] = useState<InventoryItem | null>(null);

  const [showAddCategory, setShowAddCategory] = useState(false);
  const [editingCategory, setEditingCategory] =
    useState<InventoryCategory | null>(null);

  const [showAddStock, setShowAddStock] = useState(false);
  const [viewingMovement, setViewingMovement] =
    useState<InventoryMovement | null>(null);

  const [showIssueItem, setShowIssueItem] = useState(false);
  const [viewingIssue, setViewingIssue] = useState<InventoryMovement | null>(
    null,
  );
  const [returningIssue, setReturningIssue] =
    useState<InventoryMovement | null>(null);

  const [deleteConfirm, setDeleteConfirm] = useState<{
    type: "category" | "item";
    id: string;
    name: string;
  } | null>(null);

  // Staff recipients
  const [recipientSearch, setRecipientSearch] = useState("");
  const [recipients, setRecipients] = useState<StaffUser[]>([]);

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
    categoryId: "",
    itemId: "",
    quantity: 50,
    supplier: "",
    storeName: "Central Hospital Store",
    reference: "",
    unitCost: 15.0,
    reason: "Routine replenishment",
  });

  const [issueForm, setIssueForm] = useState({
    categoryId: "",
    itemId: "",
    quantity: 5,
    recipientId: "",
    reason: "Departmental clinical supply",
  });

  const [returnForm, setReturnForm] = useState({
    quantity: 1,
    restock: true,
    reason: "Unused clinical item returned",
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

  async function save(
    path: string,
    payload: object,
    method: "POST" | "PATCH" = "POST",
    onSaved?: () => void,
  ) {
    if (submitting.current) return;
    submitting.current = true;
    setSaving(true);
    setError("");
    setSuccessMsg("");
    const signature = JSON.stringify({ path, method, payload });
    if (pending.current?.signature !== signature) {
      pending.current = { signature, key: crypto.randomUUID() };
    }
    try {
      await api(path, {
        method,
        headers: { "Idempotency-Key": pending.current.key },
        body: JSON.stringify(payload),
      });
      pending.current = null;
      if (onSaved) onSaved();
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

  async function handleDelete(type: "category" | "item", id: string) {
    if (submitting.current) return;
    submitting.current = true;
    setSaving(true);
    setError("");
    try {
      const endpoint =
        type === "category"
          ? `inventory/categories/${id}`
          : `inventory/items/${id}`;
      await api(endpoint, { method: "DELETE" });
      setDeleteConfirm(null);
      await fetchInventoryData();
    } catch (cause) {
      setError(
        cause instanceof Error ? cause.message : "Unable to delete record",
      );
    } finally {
      submitting.current = false;
      setSaving(false);
    }
  }

  // --- Category Handlers ---
  const handleSaveCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (editingCategory) {
      await save(
        `inventory/categories/${editingCategory.id}`,
        {
          name: categoryForm.name,
          description: categoryForm.description,
          active: editingCategory.active,
          version: editingCategory.version,
        },
        "PATCH",
        () => {
          setEditingCategory(null);
          setCategoryForm({ name: "", description: "" });
        },
      );
    } else {
      await save(
        "inventory/categories",
        { ...categoryForm, active: true, version: 1 },
        "POST",
        () => {
          setShowAddCategory(false);
          setCategoryForm({ name: "", description: "" });
        },
      );
    }
  };

  const openEditCategory = (cat: InventoryCategory) => {
    setEditingCategory(cat);
    setCategoryForm({ name: cat.name, description: cat.description || "" });
  };

  // --- Item Handlers ---
  const handleSaveItem = async (e: React.FormEvent) => {
    e.preventDefault();
    if (editingItem) {
      await save(
        `inventory/items/${editingItem.id}`,
        {
          categoryId: itemForm.categoryId || editingItem.categoryId,
          name: itemForm.name,
          unit: itemForm.unit,
          description: itemForm.description,
          reorderMilli: Number(itemForm.reorderLevel) * 1000,
          active: editingItem.active,
          version: editingItem.version,
        },
        "PATCH",
        () => {
          setEditingItem(null);
          setItemForm({
            name: "",
            categoryId: "",
            unit: "Piece",
            reorderLevel: 10,
            description: "",
          });
        },
      );
    } else {
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
        "POST",
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
    }
  };

  const openEditItem = (item: InventoryItem) => {
    setEditingItem(item);
    setItemForm({
      name: item.name,
      categoryId: item.categoryId,
      unit: item.unit,
      reorderLevel: Math.round(item.reorderMilli / 1000),
      description: item.description || "",
    });
  };

  // --- Stock Handlers ---
  const handleSaveStockReceive = async (e: React.FormEvent) => {
    e.preventDefault();
    const targetItemId =
      stockForm.itemId || filteredStockItems[0]?.id || items[0]?.id || "";
    await save(
      "inventory/movements",
      {
        itemId: targetItemId,
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
      "POST",
      () => setShowAddStock(false),
    );
  };

  // --- Issue Handlers ---
  const handleSaveIssue = async (e: React.FormEvent) => {
    e.preventDefault();
    const targetItemId =
      issueForm.itemId || filteredIssueItems[0]?.id || items[0]?.id || "";
    await save(
      "inventory/movements",
      {
        itemId: targetItemId,
        kind: "issue",
        quantityMilli: Number(issueForm.quantity) * 1000,
        recipientId: issueForm.recipientId,
        reason: issueForm.reason,
      },
      "POST",
      () => setShowIssueItem(false),
    );
  };

  // --- Return Handlers ---
  const handleSaveReturn = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!returningIssue) return;
    await save(
      "inventory/movements",
      {
        itemId: returningIssue.itemId,
        kind: "return",
        originalId: returningIssue.id,
        quantityMilli: Number(returnForm.quantity) * 1000,
        restock: returnForm.restock,
        reason: returnForm.reason,
      },
      "POST",
      () => setReturningIssue(null),
    );
  };

  const openReturnModal = (issue: InventoryMovement) => {
    setReturningIssue(issue);
    const returned = issue.returnedMilli || 0;
    const remaining = Math.max(0, issue.quantityMilli - returned);
    const remainingUnits = remaining / 1000;
    setReturnForm({
      quantity: remainingUnits > 0 ? remainingUnits : 1,
      restock: true,
      reason: "Unused departmental supplies returned",
    });
  };

  // --- Helpers ---
  const getCategoryName = (catId: string) => {
    return categories.find((c) => c.id === catId)?.name || t("General");
  };

  const getItemName = (itemId: string) => {
    return items.find((i) => i.id === itemId)?.name || t("Medical Supply Item");
  };

  const getItemUnit = (itemId: string) => {
    return items.find((i) => i.id === itemId)?.unit || "Units";
  };

  const getItemCategory = (itemId: string) => {
    const item = items.find((i) => i.id === itemId);
    if (!item) return "-";
    return getCategoryName(item.categoryId);
  };

  const getItemBalance = (itemId: string) => {
    const item = items.find((i) => i.id === itemId);
    return item ? item.balanceMilli / 1000 : 0;
  };

  // Filtered item lists for forms
  const filteredStockItems = useMemo(() => {
    if (!stockForm.categoryId) return items;
    return items.filter((i) => i.categoryId === stockForm.categoryId);
  }, [items, stockForm.categoryId]);

  const filteredIssueItems = useMemo(() => {
    if (!issueForm.categoryId) return items;
    return items.filter((i) => i.categoryId === issueForm.categoryId);
  }, [items, issueForm.categoryId]);

  const selectedIssueItem = useMemo(() => {
    const targetId =
      issueForm.itemId || filteredIssueItems[0]?.id || items[0]?.id;
    return items.find((i) => i.id === targetId);
  }, [items, issueForm.itemId, filteredIssueItems]);

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
    <div
      className="space-y-6 legacy-workspace"
      data-ready={isLive ? "true" : "false"}
    >
      {error && (
        <p role="alert" className="error">
          {t(error)}
        </p>
      )}

      {/* Subtabs Nav */}
      <div className="border-b border-border/80 bg-card/50 backdrop-blur rounded-xl p-1.5 shadow-xs">
        <nav className="flex space-x-1 overflow-x-auto">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => {
                  setActiveTab(tab.id as InventoryTab);
                  setSearchTerm("");
                }}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 whitespace-nowrap ${
                  isActive
                    ? "bg-primary text-primary-foreground shadow-xs"
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
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-gradient-to-r from-emerald-500/10 via-emerald-500/5 to-transparent border border-emerald-500/20 rounded-xl p-3.5 text-sm shadow-xs">
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
              onClick={() => {
                setItemForm({
                  name: "",
                  categoryId: categories[0]?.id || "",
                  unit: "Piece",
                  reorderLevel: 10,
                  description: "",
                });
                setShowAddItem(true);
              }}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-xs"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Item")}
            </button>
          )}
          {activeTab === "item-categories" && (
            <button
              onClick={() => {
                setCategoryForm({ name: "", description: "" });
                setShowAddCategory(true);
              }}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-xs"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("New Item Category")}
            </button>
          )}
          {activeTab === "item-stocks" && (
            <button
              onClick={() => {
                const defaultCat = categories[0]?.id || "";
                const itemsInCat = items.filter(
                  (i) => i.categoryId === defaultCat,
                );
                setStockForm({
                  categoryId: defaultCat,
                  itemId: itemsInCat[0]?.id || items[0]?.id || "",
                  quantity: 50,
                  supplier: "",
                  storeName: "Central Hospital Store",
                  reference: "",
                  unitCost: 15.0,
                  reason: "Routine replenishment",
                });
                setShowAddStock(true);
              }}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-xs"
            >
              <Plus className="w-3.5 h-3.5" />
              {t("Receive New Stock")}
            </button>
          )}
          {activeTab === "issued-items" && (
            <button
              onClick={() => {
                const defaultCat = categories[0]?.id || "";
                const itemsInCat = items.filter(
                  (i) => i.categoryId === defaultCat,
                );
                setIssueForm({
                  categoryId: defaultCat,
                  itemId: itemsInCat[0]?.id || items[0]?.id || "",
                  quantity: 5,
                  recipientId: "",
                  reason: "Departmental clinical supply",
                });
                setShowIssueItem(true);
              }}
              className="btn-primary text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg shadow-xs"
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
          <div className="p-4 border-b border-border/80 flex flex-col md:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Medical Inventory Items")}
            </h3>
            <div className="flex flex-wrap items-center gap-2 w-full md:w-auto">
              {/* Category Filter */}
              <select
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
                aria-label={t("Filter by category")}
                className="text-xs bg-background border border-border rounded-lg px-2.5 py-1.5 text-foreground focus:outline-hidden"
              >
                <option value="">{t("All Categories")}</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>

              {/* Low Stock Toggle */}
              <button
                type="button"
                onClick={() => setLowStockOnly(!lowStockOnly)}
                className={`text-xs px-2.5 py-1.5 rounded-lg border transition-colors flex items-center gap-1 ${
                  lowStockOnly
                    ? "bg-red-500/15 border-red-500/30 text-red-600 dark:text-red-400 font-medium"
                    : "bg-background border-border text-muted-foreground hover:text-foreground"
                }`}
              >
                <AlertTriangle className="w-3 h-3" />
                {t("Low Stock Only")}
              </button>

              {/* Search */}
              <div className="relative flex-1 sm:w-64">
                <Search className="w-4 h-4 absolute left-3 top-2 text-muted-foreground" />
                <input
                  type="text"
                  placeholder={t("Search items...")}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-9 pr-3 py-1.5 text-xs bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
                />
              </div>
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
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {items
                  .filter(
                    (i) => !categoryFilter || i.categoryId === categoryFilter,
                  )
                  .filter((i) => {
                    if (!lowStockOnly) return true;
                    return i.balanceMilli <= i.reorderMilli;
                  })
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
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              type="button"
                              onClick={() => setViewingItem(item)}
                              aria-label={t("Item Details")}
                              title={t("Item Details")}
                              className="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted"
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                            <button
                              type="button"
                              onClick={() => openEditItem(item)}
                              aria-label={t("Edit Item")}
                              title={t("Edit Item")}
                              className="p-1 rounded-md text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/50"
                            >
                              <Pencil className="w-4 h-4" />
                            </button>
                            <button
                              type="button"
                              onClick={() =>
                                setDeleteConfirm({
                                  type: "item",
                                  id: item.id,
                                  name: item.name,
                                })
                              }
                              aria-label={t("Delete Item")}
                              title={t("Delete Item")}
                              className="p-1 rounded-md text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/50"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                {items.length === 0 && !loading && (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-4 py-8 text-center text-muted-foreground text-sm"
                    >
                      <Package className="w-8 h-8 mx-auto mb-2 opacity-40" />
                      {t("No inventory items found")}
                    </td>
                  </tr>
                )}
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
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
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
                          {itemCount} {t("items")}
                        </td>
                        <td className="px-4 py-3">
                          <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                            {t("Active")}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              type="button"
                              onClick={() => openEditCategory(cat)}
                              aria-label={t("Edit Category")}
                              title={t("Edit Category")}
                              className="p-1 rounded-md text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/50"
                            >
                              <Pencil className="w-4 h-4" />
                            </button>
                            <button
                              type="button"
                              onClick={() =>
                                setDeleteConfirm({
                                  type: "category",
                                  id: cat.id,
                                  name: cat.name,
                                })
                              }
                              aria-label={t("Delete Category")}
                              title={t("Delete Category")}
                              className="p-1 rounded-md text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/50"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                {categories.length === 0 && !loading && (
                  <tr>
                    <td
                      colSpan={5}
                      className="px-4 py-8 text-center text-muted-foreground text-sm"
                    >
                      <Layers className="w-8 h-8 mx-auto mb-2 opacity-40" />
                      {t("No categories found")}
                    </td>
                  </tr>
                )}
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
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Supplier / Store")}</th>
                  <th className="px-4 py-3">{t("Reference No")}</th>
                  <th className="px-4 py-3">{t("Quantity Received")}</th>
                  <th className="px-4 py-3">{t("Total Cost (ETB)")}</th>
                  <th className="px-4 py-3">{t("Date")}</th>
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
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
                        .includes(searchTerm.toLowerCase()) ||
                      (m.supplier || "")
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
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {getItemCategory(m.itemId)}
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-xs text-foreground">
                          {m.supplier || "-"}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {m.storeName || "-"}
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
                      <td className="px-4 py-3 text-right">
                        <button
                          type="button"
                          onClick={() => setViewingMovement(m)}
                          aria-label={t("Stock Details")}
                          title={t("Stock Details")}
                          className="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                {movements.filter((m) => m.kind === "receive").length === 0 &&
                  !loading && (
                    <tr>
                      <td
                        colSpan={8}
                        className="px-4 py-8 text-center text-muted-foreground text-sm"
                      >
                        <ArrowDownToLine className="w-8 h-8 mx-auto mb-2 opacity-40" />
                        {t("No stock receipts recorded")}
                      </td>
                    </tr>
                  )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 4. ISSUED ITEMS */}
      {activeTab === "issued-items" && (
        <div className="bg-card border border-border/80 rounded-xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border/80 flex flex-col md:flex-row items-center justify-between gap-3">
            <h3 className="font-semibold text-base text-foreground">
              {t("Issued Items & Consumables")}
            </h3>
            <div className="flex flex-wrap items-center gap-2 w-full md:w-auto">
              {/* Status Filter */}
              <select
                value={issueStatusFilter}
                onChange={(e) =>
                  setIssueStatusFilter(
                    e.target.value as "all" | "returnable" | "returned",
                  )
                }
                aria-label={t("Filter by status")}
                className="text-xs bg-background border border-border rounded-lg px-2.5 py-1.5 text-foreground focus:outline-hidden"
              >
                <option value="all">{t("All Statuses")}</option>
                <option value="returnable">{t("Pending Return")}</option>
                <option value="returned">{t("Returned")}</option>
              </select>

              {/* Search */}
              <div className="relative flex-1 sm:w-64">
                <Search className="w-4 h-4 absolute left-3 top-2 text-muted-foreground" />
                <input
                  type="text"
                  placeholder={t("Search issued items...")}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-9 pr-3 py-1.5 text-xs bg-background border border-border rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary/20"
                />
              </div>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs">
                <tr>
                  <th className="px-4 py-3">{t("Item")}</th>
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Issued To / Recipient")}</th>
                  <th className="px-4 py-3">{t("Quantity Issued")}</th>
                  <th className="px-4 py-3">{t("Returned / Remaining")}</th>
                  <th className="px-4 py-3">{t("Status")}</th>
                  <th className="px-4 py-3">{t("Issued Date")}</th>
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {movements
                  .filter((m) => m.kind === "issue")
                  .filter((m) => {
                    const returned = m.returnedMilli || 0;
                    const isFullyReturned = returned >= m.quantityMilli;
                    if (issueStatusFilter === "returnable")
                      return !isFullyReturned;
                    if (issueStatusFilter === "returned")
                      return isFullyReturned;
                    return true;
                  })
                  .filter(
                    (m) =>
                      getItemName(m.itemId)
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()) ||
                      (m.recipientId || "")
                        .toLowerCase()
                        .includes(searchTerm.toLowerCase()),
                  )
                  .map((m) => {
                    const issued = m.quantityMilli / 1000;
                    const returned = (m.returnedMilli || 0) / 1000;
                    const remaining = Math.max(0, issued - returned);
                    const isFullyReturned = remaining === 0;
                    const isPartial = returned > 0 && remaining > 0;

                    return (
                      <tr
                        key={m.id}
                        className="hover:bg-muted/30 transition-colors"
                      >
                        <td className="px-4 py-3 font-semibold text-foreground">
                          {getItemName(m.itemId)}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {getItemCategory(m.itemId)}
                        </td>
                        <td className="px-4 py-3 font-medium text-xs text-foreground">
                          {m.recipientId || "-"}
                        </td>
                        <td className="px-4 py-3 font-bold font-mono text-amber-600 dark:text-amber-400">
                          -{issued.toLocaleString()} {getItemUnit(m.itemId)}
                        </td>
                        <td className="px-4 py-3 font-mono text-xs">
                          <span className="text-emerald-600 dark:text-emerald-400 font-medium">
                            {returned}
                          </span>
                          {" / "}
                          <span className="text-muted-foreground">
                            {remaining} {getItemUnit(m.itemId)}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          {isFullyReturned ? (
                            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                              <Check className="w-3.5 h-3.5" />
                              {t("Returned")}
                            </span>
                          ) : isPartial ? (
                            <button
                              type="button"
                              onClick={() => openReturnModal(m)}
                              aria-label={t("Partial Return")}
                              className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/15 text-amber-700 dark:text-amber-300 hover:bg-amber-500/25 transition-colors cursor-pointer"
                            >
                              <RotateCcw className="w-3 h-3" />
                              {t("Partial Return")} ({returned}/{issued})
                            </button>
                          ) : (
                            <button
                              type="button"
                              onClick={() => openReturnModal(m)}
                              aria-label={t("Return Item")}
                              className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/15 text-blue-700 dark:text-blue-300 hover:bg-blue-500/25 transition-colors cursor-pointer"
                            >
                              <RotateCcw className="w-3 h-3" />
                              {t("Return Item")}
                            </button>
                          )}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {new Date(m.createdAt).toLocaleDateString()}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              type="button"
                              onClick={() => setViewingIssue(m)}
                              aria-label={t("Issue Details")}
                              title={t("Issue Details")}
                              className="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted"
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                            {!isFullyReturned && (
                              <button
                                type="button"
                                onClick={() => openReturnModal(m)}
                                aria-label={t("Return Item")}
                                title={t("Return Item")}
                                className="p-1 rounded-md text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/50"
                              >
                                <RotateCcw className="w-4 h-4" />
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                {movements.filter((m) => m.kind === "issue").length === 0 &&
                  !loading && (
                    <tr>
                      <td
                        colSpan={8}
                        className="px-4 py-8 text-center text-muted-foreground text-sm"
                      >
                        <ArrowUpFromLine className="w-8 h-8 mx-auto mb-2 opacity-40" />
                        {t("No issued items found")}
                      </td>
                    </tr>
                  )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* --- MODALS --- */}

      {/* 1. New / Edit Item Modal */}
      {(showAddItem || editingItem) && (
        <Modal
          onClose={() => {
            setShowAddItem(false);
            setEditingItem(null);
          }}
          titleId="item-modal-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="item-modal-title" className="font-semibold text-lg">
                {editingItem ? t("Edit Item") : t("New Item")}
              </h3>
              <button
                onClick={() => {
                  setShowAddItem(false);
                  setEditingItem(null);
                }}
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
                  {t("Reorder Level")}
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
                  onClick={() => {
                    setShowAddItem(false);
                    setEditingItem(null);
                  }}
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

      {/* 2. New / Edit Category Modal */}
      {(showAddCategory || editingCategory) && (
        <Modal
          onClose={() => {
            setShowAddCategory(false);
            setEditingCategory(null);
          }}
          titleId="category-modal-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="category-modal-title" className="font-semibold text-lg">
                {editingCategory ? t("Edit Category") : t("New Item Category")}
              </h3>
              <button
                onClick={() => {
                  setShowAddCategory(false);
                  setEditingCategory(null);
                }}
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
                  onClick={() => {
                    setShowAddCategory(false);
                    setEditingCategory(null);
                  }}
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
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Category")}
                  </label>
                  <select
                    value={stockForm.categoryId}
                    onChange={(e) => {
                      const newCatId = e.target.value;
                      const catItems = items.filter(
                        (i) => i.categoryId === newCatId,
                      );
                      setStockForm({
                        ...stockForm,
                        categoryId: newCatId,
                        itemId: catItems[0]?.id || "",
                      });
                    }}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    <option value="">{t("All Categories")}</option>
                    {categories.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Item")} *
                  </label>
                  <select
                    aria-label={t("Item")}
                    value={stockForm.itemId || filteredStockItems[0]?.id || ""}
                    onChange={(e) =>
                      setStockForm({ ...stockForm, itemId: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    {filteredStockItems.map((i) => (
                      <option key={i.id} value={i.id}>
                        {i.name} ({i.unit})
                      </option>
                    ))}
                  </select>
                </div>
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
                    placeholder="e.g. Ethiopian Pharmaceuticals"
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
                  {t("Reference No")} *
                </label>
                <input
                  type="text"
                  required
                  aria-label={t("Reference")}
                  value={stockForm.reference}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, reference: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg font-mono"
                  placeholder="PO-2026-001"
                />
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Reason")} *
                </label>
                <input
                  type="text"
                  required
                  aria-label={t("Reason")}
                  value={stockForm.reason}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, reason: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>

              {/* Total Calculation Display */}
              <div className="p-3 rounded-lg bg-muted/60 border border-border flex items-center justify-between text-xs font-mono">
                <span className="text-muted-foreground">
                  {t("Estimated Total")}:
                </span>
                <span className="font-bold text-foreground">
                  {(
                    Number(stockForm.quantity || 0) *
                    Number(stockForm.unitCost || 0)
                  ).toLocaleString(undefined, {
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2,
                  })}{" "}
                  ETB
                </span>
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
                {t("Issue Item")}
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
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Category")}
                  </label>
                  <select
                    value={issueForm.categoryId}
                    onChange={(e) => {
                      const newCatId = e.target.value;
                      const catItems = items.filter(
                        (i) => i.categoryId === newCatId,
                      );
                      setIssueForm({
                        ...issueForm,
                        categoryId: newCatId,
                        itemId: catItems[0]?.id || "",
                      });
                    }}
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    <option value="">{t("All Categories")}</option>
                    {categories.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    {t("Item")} *
                  </label>
                  <select
                    aria-label={t("Item")}
                    value={issueForm.itemId || filteredIssueItems[0]?.id || ""}
                    onChange={(e) =>
                      setIssueForm({ ...issueForm, itemId: e.target.value })
                    }
                    className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  >
                    {filteredIssueItems.map((i) => (
                      <option key={i.id} value={i.id}>
                        {i.name} ({t("Available")}: {i.balanceMilli / 1000}{" "}
                        {i.unit})
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Live stock indicator */}
              {selectedIssueItem && (
                <div className="p-2.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-xs flex items-center justify-between">
                  <span className="text-muted-foreground">
                    {t("Available Quantity")}:
                  </span>
                  <span className="font-bold text-foreground">
                    {(selectedIssueItem.balanceMilli / 1000).toLocaleString()}{" "}
                    {selectedIssueItem.unit}
                  </span>
                </div>
              )}

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
                    placeholder={t("Filter staff...")}
                    value={recipientSearch}
                    onChange={(e) => setRecipientSearch(e.target.value)}
                    className="w-full px-2 py-1 mb-1 text-xs bg-background border border-border rounded-md"
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
                  {t("Reason / Purpose")} *
                </label>
                <textarea
                  rows={2}
                  required
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

      {/* 5. Return Item Modal */}
      {returningIssue && (
        <Modal
          onClose={() => setReturningIssue(null)}
          titleId="return-item-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="return-item-title" className="font-semibold text-lg">
                {t("Return Issued Item")}
              </h3>
              <button
                onClick={() => setReturningIssue(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleSaveReturn} className="space-y-4">
              {error && (
                <p role="alert" className="error">
                  {t(error)}
                </p>
              )}

              {/* Issue Summary Box */}
              <div className="p-3 rounded-lg bg-muted/60 border border-border text-xs space-y-1.5">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t("Item")}:</span>
                  <span className="font-semibold text-foreground">
                    {getItemName(returningIssue.itemId)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">
                    {t("Issued To")}:
                  </span>
                  <span className="font-mono text-foreground">
                    {returningIssue.recipientId || "-"}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">
                    {t("Quantity Issued")}:
                  </span>
                  <span className="font-mono font-medium text-foreground">
                    {(returningIssue.quantityMilli / 1000).toLocaleString()}{" "}
                    {getItemUnit(returningIssue.itemId)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">
                    {t("Already Returned")}:
                  </span>
                  <span className="font-mono text-emerald-600 dark:text-emerald-400 font-medium">
                    {(
                      (returningIssue.returnedMilli || 0) / 1000
                    ).toLocaleString()}{" "}
                    {getItemUnit(returningIssue.itemId)}
                  </span>
                </div>
                <div className="flex justify-between border-t border-border/60 pt-1.5">
                  <span className="text-muted-foreground font-medium">
                    {t("Remaining Returnable")}:
                  </span>
                  <span className="font-mono font-bold text-foreground">
                    {(
                      Math.max(
                        0,
                        returningIssue.quantityMilli -
                          (returningIssue.returnedMilli || 0),
                      ) / 1000
                    ).toLocaleString()}{" "}
                    {getItemUnit(returningIssue.itemId)}
                  </span>
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Quantity to Return")} *
                </label>
                <input
                  type="number"
                  min={1}
                  max={
                    Math.max(
                      0,
                      returningIssue.quantityMilli -
                        (returningIssue.returnedMilli || 0),
                    ) / 1000
                  }
                  required
                  aria-label={t("Quantity to Return")}
                  value={returnForm.quantity}
                  onChange={(e) =>
                    setReturnForm({
                      ...returnForm,
                      quantity: Number(e.target.value),
                    })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="restock-checkbox"
                  aria-label={t("Restock to Inventory")}
                  checked={returnForm.restock}
                  onChange={(e) =>
                    setReturnForm({ ...returnForm, restock: e.target.checked })
                  }
                  className="h-4 w-4 rounded border-border text-primary focus:ring-primary"
                />
                <label
                  htmlFor="restock-checkbox"
                  className="text-xs font-medium text-foreground cursor-pointer"
                >
                  {t("Restock to Inventory")}{" "}
                  <span className="text-muted-foreground font-normal">
                    ({t("increases available balance")})
                  </span>
                </label>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">
                  {t("Return Reason")} *
                </label>
                <textarea
                  rows={2}
                  required
                  aria-label={t("Return Reason")}
                  value={returnForm.reason}
                  onChange={(e) =>
                    setReturnForm({ ...returnForm, reason: e.target.value })
                  }
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg"
                  placeholder="e.g. Unused clinical supply returned"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-border/80">
                <button
                  type="button"
                  onClick={() => setReturningIssue(null)}
                  className="btn-secondary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary px-4 py-2 text-xs rounded-lg"
                >
                  {t("Confirm Return")}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      )}

      {/* 6. Item Details Modal */}
      {viewingItem && (
        <Modal
          onClose={() => setViewingItem(null)}
          titleId="item-details-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="item-details-title" className="font-semibold text-lg">
                {t("Item Details")}
              </h3>
              <button
                onClick={() => setViewingItem(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Item Name")}:</span>
                <span className="font-semibold text-foreground">
                  {viewingItem.name}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Category")}:</span>
                <span className="text-foreground">
                  {getCategoryName(viewingItem.categoryId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Unit")}:</span>
                <span className="font-mono text-foreground">
                  {viewingItem.unit}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Available Quantity")}:
                </span>
                <span className="font-mono font-bold text-foreground">
                  {(viewingItem.balanceMilli / 1000).toLocaleString()}{" "}
                  {viewingItem.unit}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Reorder Level")}:
                </span>
                <span className="font-mono text-muted-foreground">
                  {(viewingItem.reorderMilli / 1000).toLocaleString()}{" "}
                  {viewingItem.unit}
                </span>
              </div>
              <div className="border-b border-border/50 pb-2">
                <span className="text-muted-foreground block mb-1">
                  {t("Description")}:
                </span>
                <p className="text-xs text-foreground bg-muted/40 p-2 rounded-lg">
                  {viewingItem.description || "-"}
                </p>
              </div>
            </div>
            <div className="flex justify-end pt-2">
              <button
                type="button"
                onClick={() => setViewingItem(null)}
                className="btn-secondary px-4 py-2 text-xs rounded-lg"
              >
                {t("Close")}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* 7. Stock Receipt Details Modal */}
      {viewingMovement && (
        <Modal
          onClose={() => setViewingMovement(null)}
          titleId="stock-details-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="stock-details-title" className="font-semibold text-lg">
                {t("Stock Details")}
              </h3>
              <button
                onClick={() => setViewingMovement(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Item")}:</span>
                <span className="font-semibold text-foreground">
                  {getItemName(viewingMovement.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Category")}:</span>
                <span className="text-foreground">
                  {getItemCategory(viewingMovement.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Reference No")}:
                </span>
                <span className="font-mono font-semibold text-foreground">
                  {viewingMovement.reference || "-"}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Supplier")}:</span>
                <span className="text-foreground">
                  {viewingMovement.supplier || "-"}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Store Name")}:
                </span>
                <span className="text-foreground">
                  {viewingMovement.storeName || "-"}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Quantity Received")}:
                </span>
                <span className="font-mono font-bold text-emerald-600 dark:text-emerald-400">
                  +{(viewingMovement.quantityMilli / 1000).toLocaleString()}{" "}
                  {getItemUnit(viewingMovement.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Total Cost")}:
                </span>
                <span className="font-mono text-foreground">
                  {viewingMovement.costMinor
                    ? `${(viewingMovement.costMinor / 100).toLocaleString(
                        undefined,
                        { minimumFractionDigits: 2 },
                      )} ETB`
                    : "-"}
                </span>
              </div>
              <div className="border-b border-border/50 pb-2">
                <span className="text-muted-foreground block mb-1">
                  {t("Reason")}:
                </span>
                <p className="text-xs text-foreground bg-muted/40 p-2 rounded-lg">
                  {viewingMovement.reason || "-"}
                </p>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t("Date")}:</span>
                <span className="text-xs text-muted-foreground">
                  {new Date(viewingMovement.createdAt).toLocaleString()}
                </span>
              </div>
            </div>
            <div className="flex justify-end pt-2">
              <button
                type="button"
                onClick={() => setViewingMovement(null)}
                className="btn-secondary px-4 py-2 text-xs rounded-lg"
              >
                {t("Close")}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* 8. Issue Details Modal */}
      {viewingIssue && (
        <Modal
          onClose={() => setViewingIssue(null)}
          titleId="issue-details-title"
        >
          <div className="p-6 space-y-4 max-w-lg w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center justify-between border-b border-border/80 pb-3">
              <h3 id="issue-details-title" className="font-semibold text-lg">
                {t("Issue Details")}
              </h3>
              <button
                onClick={() => setViewingIssue(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Item")}:</span>
                <span className="font-semibold text-foreground">
                  {getItemName(viewingIssue.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Category")}:</span>
                <span className="text-foreground">
                  {getItemCategory(viewingIssue.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">{t("Issued To")}:</span>
                <span className="font-mono text-foreground">
                  {viewingIssue.recipientId || "-"}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Quantity Issued")}:
                </span>
                <span className="font-mono font-bold text-amber-600 dark:text-amber-400">
                  -{(viewingIssue.quantityMilli / 1000).toLocaleString()}{" "}
                  {getItemUnit(viewingIssue.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Quantity Returned")}:
                </span>
                <span className="font-mono text-emerald-600 dark:text-emerald-400 font-medium">
                  {((viewingIssue.returnedMilli || 0) / 1000).toLocaleString()}{" "}
                  {getItemUnit(viewingIssue.itemId)}
                </span>
              </div>
              <div className="flex justify-between border-b border-border/50 pb-2">
                <span className="text-muted-foreground">
                  {t("Remaining Returnable")}:
                </span>
                <span className="font-mono font-bold text-foreground">
                  {(
                    Math.max(
                      0,
                      viewingIssue.quantityMilli -
                        (viewingIssue.returnedMilli || 0),
                    ) / 1000
                  ).toLocaleString()}{" "}
                  {getItemUnit(viewingIssue.itemId)}
                </span>
              </div>
              <div className="border-b border-border/50 pb-2">
                <span className="text-muted-foreground block mb-1">
                  {t("Reason")}:
                </span>
                <p className="text-xs text-foreground bg-muted/40 p-2 rounded-lg">
                  {viewingIssue.reason || "-"}
                </p>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">
                  {t("Issued Date")}:
                </span>
                <span className="text-xs text-muted-foreground">
                  {new Date(viewingIssue.createdAt).toLocaleString()}
                </span>
              </div>
            </div>
            <div className="flex justify-end pt-2">
              <button
                type="button"
                onClick={() => setViewingIssue(null)}
                className="btn-secondary px-4 py-2 text-xs rounded-lg"
              >
                {t("Close")}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* 9. Delete Confirmation Modal */}
      {deleteConfirm && (
        <Modal
          onClose={() => setDeleteConfirm(null)}
          titleId="delete-confirm-title"
        >
          <div className="p-6 space-y-4 max-w-md w-full bg-card rounded-2xl shadow-xl">
            <div className="flex items-center gap-3 text-red-600 dark:text-red-400">
              <AlertTriangle className="w-6 h-6 flex-shrink-0" />
              <h3 id="delete-confirm-title" className="font-semibold text-lg">
                {deleteConfirm.type === "category"
                  ? t("Delete Category")
                  : t("Delete Item")}
              </h3>
            </div>
            {error && (
              <p role="alert" className="error">
                {t(error)}
              </p>
            )}
            <p className="text-sm text-muted-foreground">
              {deleteConfirm.type === "category"
                ? t("Are you sure you want to delete this category?")
                : t("Are you sure you want to delete this item?")}
            </p>
            <p className="font-semibold text-sm text-foreground bg-muted/40 p-2.5 rounded-lg font-mono">
              {deleteConfirm.name}
            </p>
            <div className="flex justify-end gap-2 pt-3 border-t border-border/80">
              <button
                type="button"
                onClick={() => setDeleteConfirm(null)}
                className="btn-secondary px-4 py-2 text-xs rounded-lg"
              >
                {t("Cancel")}
              </button>
              <button
                type="button"
                disabled={saving}
                onClick={() =>
                  handleDelete(deleteConfirm.type, deleteConfirm.id)
                }
                className="btn-danger px-4 py-2 text-xs rounded-lg bg-red-600 text-white hover:bg-red-700"
              >
                {t("Confirm Delete")}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
