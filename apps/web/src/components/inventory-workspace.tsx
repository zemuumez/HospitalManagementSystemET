"use client";

import React, { useState, useEffect, useRef, useMemo } from "react";
import {
  Package,
  Layers,
  ArrowDownToLine,
  ArrowUpFromLine,
  Plus,
  RefreshCw,
  Search,
  Filter,
  CheckCircle,
  AlertTriangle,
  RotateCcw,
  Edit,
  Trash2,
  Eye,
  FileText,
  Upload,
  Ban,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { useLanguage } from "./language";

type InventoryTab =
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
  itemName?: string;
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
  attachmentUrl?: string;
  issuedDate?: string;
  returnDueDate?: string;
  issuedBy?: string;
  department?: string;
  createdAt: string;
}

interface StaffUser {
  id: string;
  name: string;
  role: string;
  active: boolean;
}

const DEPARTMENTS = [
  "General OPD",
  "Dental",
  "Emergency",
  "Inpatient Ward",
  "Surgery",
  "Laboratory",
  "Pharmacy",
  "Maternity",
  "Pediatrics Ward",
];

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

  // Pagination & Totals
  const [page, setPage] = useState(1);
  const [pageSize] = useState(25);
  const [totalItems, setTotalItems] = useState(0);
  const [totalCategories, setTotalCategories] = useState(0);
  const [totalMovements, setTotalMovements] = useState(0);

  // Data states
  const [categories, setCategories] = useState<InventoryCategory[]>([]);
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [movements, setMovements] = useState<InventoryMovement[]>([]);

  // Complete catalogs for modal selectors (unbounded by current page)
  const [allCategoriesCatalog, setAllCategoriesCatalog] = useState<
    InventoryCategory[]
  >([]);
  const [allItemsCatalog, setAllItemsCatalog] = useState<InventoryItem[]>([]);

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

  // Void confirmation modals
  const [voidingReceipt, setVoidingReceipt] =
    useState<InventoryMovement | null>(null);
  const [voidingIssue, setVoidingIssue] = useState<InventoryMovement | null>(
    null,
  );
  const [voidReason, setVoidReason] = useState("");

  const [deleteConfirm, setDeleteConfirm] = useState<{
    type: "category" | "item";
    id: string;
    name: string;
  } | null>(null);

  // Staff recipients
  const [recipientSearch, setRecipientSearch] = useState("");
  const [recipients, setRecipients] = useState<StaffUser[]>([]);
  const [uploadingReceipt, setUploadingReceipt] = useState(false);

  // Form states
  const [itemForm, setItemForm] = useState({
    name: "",
    categoryId: "",
    unit: "Piece",
    reorderLevel: "10",
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
    attachmentUrl: "",
  });

  const [issueForm, setIssueForm] = useState({
    categoryId: "",
    itemId: "",
    quantity: 5,
    recipientId: "",
    reason: "Departmental clinical supply",
    department: "General OPD",
    issuedBy: "Pharmacy Admin",
    issuedDate: new Date().toISOString().split("T")[0],
    returnDueDate: "",
  });

  const [returnForm, setReturnForm] = useState({
    quantity: 1,
    restock: true,
    reason: "Unused clinical item returned",
  });

  // Load staff on mount
  useEffect(() => {
    fetch("/api/staff")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data?.users) {
          setRecipients(
            data.users.filter(
              (u: { active: boolean; role: string }) =>
                u.active && u.role !== "patient",
            ),
          );
        }
      })
      .catch(() => {});
  }, []);

  // Fetch staff with search in issue modal
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
        if (!controller.signal.aborted) {
          setError(
            cause instanceof Error ? cause.message : "Unable to load staff",
          );
        }
      }
    }, 250);
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [showIssueItem, recipientSearch]);

  const fetchInventoryData = async (targetPage?: number) => {
    setLoading(true);
    setError("");
    const currentPage = targetPage ?? page;
    try {
      // 1. Fetch catalogs for selectors (limit=500) so dropdowns are never truncated
      const [catsCatalogRes, itemsCatalogRes, staffCatalogRes] =
        await Promise.all([
          api<{ categories: InventoryCategory[]; total: number }>(
            "inventory/categories?page=1&limit=500",
          ),
          api<{ items: InventoryItem[]; total: number }>(
            "inventory/items?page=1&limit=500",
          ),
          fetch("/api/staff?page=1&limit=500")
            .then((r) => (r.ok ? r.json() : null))
            .catch(() => null),
        ]);
      setAllCategoriesCatalog(catsCatalogRes.categories);
      setAllItemsCatalog(itemsCatalogRes.items);
      if (staffCatalogRes?.users) {
        setRecipients(
          staffCatalogRes.users.filter(
            (u: { active: boolean; role: string }) =>
              u.active && u.role !== "patient",
          ),
        );
      }

      // 2. Tab-specific paginated request
      if (activeTab === "item-categories") {
        const catsRes = await api<{
          categories: InventoryCategory[];
          total: number;
        }>(
          `inventory/categories?page=${currentPage}&limit=${pageSize}&search=${encodeURIComponent(searchTerm)}`,
        );
        setCategories(catsRes.categories);
        setTotalCategories(catsRes.total ?? catsRes.categories.length);
      } else {
        setCategories(catsCatalogRes.categories);
        setTotalCategories(
          catsCatalogRes.total ?? catsCatalogRes.categories.length,
        );
      }

      if (activeTab === "items") {
        const queryParams = new URLSearchParams({
          page: String(currentPage),
          limit: String(pageSize),
          search: searchTerm,
        });
        if (categoryFilter) queryParams.set("categoryId", categoryFilter);
        if (lowStockOnly) queryParams.set("lowStock", "true");
        const itemsRes = await api<{ items: InventoryItem[]; total: number }>(
          `inventory/items?${queryParams.toString()}`,
        );
        setItems(itemsRes.items);
        setTotalItems(itemsRes.total ?? itemsRes.items.length);
      } else {
        setItems(itemsCatalogRes.items);
        setTotalItems(itemsCatalogRes.total ?? itemsCatalogRes.items.length);
      }

      let movEndpoint = `inventory/movements?page=${currentPage}&limit=${pageSize}`;
      if (activeTab === "item-stocks") {
        movEndpoint += "&kind=receive";
      } else if (activeTab === "issued-items") {
        movEndpoint += "&kind=issue";
      }
      if (searchTerm) {
        movEndpoint += `&search=${encodeURIComponent(searchTerm)}`;
      }
      const movRes = await api<{
        movements: InventoryMovement[];
        total: number;
      }>(movEndpoint);
      setMovements(movRes.movements);
      setTotalMovements(movRes.total ?? movRes.movements.length);

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
    setPage(1);
    void fetchInventoryData(1);
  }, [activeTab, searchTerm, categoryFilter, lowStockOnly]);

  useEffect(() => {
    void fetchInventoryData(page);
  }, [page]);

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
      throw cause;
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
    setError("");
    setEditingCategory(cat);
    setCategoryForm({ name: cat.name, description: cat.description || "" });
  };

  // --- Item Handlers ---
  const handleSaveItem = async (e: React.FormEvent) => {
    e.preventDefault();
    const reorderVal = Math.round(Number(itemForm.reorderLevel) * 1000);
    if (editingItem) {
      await save(
        `inventory/items/${editingItem.id}`,
        {
          categoryId: itemForm.categoryId || editingItem.categoryId,
          name: itemForm.name,
          unit: itemForm.unit,
          description: itemForm.description,
          reorderMilli: reorderVal,
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
            reorderLevel: "10",
            description: "",
          });
        },
      );
    } else {
      await save(
        "inventory/items",
        {
          categoryId: itemForm.categoryId,
          name: itemForm.name,
          unit: itemForm.unit,
          description: itemForm.description,
          reorderMilli: reorderVal,
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
            reorderLevel: "10",
            description: "",
          });
        },
      );
    }
  };

  const openEditItem = (item: InventoryItem) => {
    setError("");
    setEditingItem(item);
    setItemForm({
      name: item.name,
      categoryId: item.categoryId,
      unit: item.unit,
      reorderLevel: (item.reorderMilli / 1000).toString(),
      description: item.description || "",
    });
  };

  // --- Receipt Attachment Upload ---
  const handleReceiptUpload = async (file: File) => {
    setUploadingReceipt(true);
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("isPublic", "true");
      const res = await fetch("/api/hms/attachments", {
        method: "POST",
        body: formData,
      });
      if (!res.ok) throw new Error("Receipt upload failed");
      const data = await res.json();
      const fileUrl = data.fileUrl || `/v1/attachments/${data.token}/content`;
      setStockForm((prev) => ({ ...prev, attachmentUrl: fileUrl }));
    } catch (cause) {
      setError(
        cause instanceof Error ? cause.message : "Receipt upload failed",
      );
    } finally {
      setUploadingReceipt(false);
    }
  };

  // --- Stock Handlers ---
  const handleSaveStockReceive = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stockForm.itemId) {
      setError(t("Please select an eligible item"));
      return;
    }
    const eligible = (
      allItemsCatalog.length > 0 ? allItemsCatalog : items
    ).some(
      (i) => i.id === stockForm.itemId && i.categoryId === stockForm.categoryId,
    );
    if (!eligible) {
      setError(t("Please select an eligible item"));
      return;
    }
    await save(
      "inventory/movements",
      {
        itemId: stockForm.itemId,
        kind: "receive",
        quantityMilli: Number(stockForm.quantity) * 1000,
        costMinor: Math.round(
          Number(stockForm.unitCost) * Number(stockForm.quantity) * 100,
        ),
        supplier: stockForm.supplier,
        storeName: stockForm.storeName,
        reference: stockForm.reference,
        reason: stockForm.reason,
        attachmentUrl: stockForm.attachmentUrl,
      },
      "POST",
      () => {
        setShowAddStock(false);
        setStockForm({
          categoryId: "",
          itemId: "",
          quantity: 50,
          supplier: "",
          storeName: "Central Hospital Store",
          reference: "",
          unitCost: 15.0,
          reason: "Routine replenishment",
          attachmentUrl: "",
        });
      },
    );
  };

  // --- Issue Handlers ---
  const handleSaveIssue = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!issueForm.itemId) {
      setError(t("Please select an eligible item"));
      return;
    }
    const eligible = (
      allItemsCatalog.length > 0 ? allItemsCatalog : items
    ).some(
      (i) => i.id === issueForm.itemId && i.categoryId === issueForm.categoryId,
    );
    if (!eligible) {
      setError(t("Please select an eligible item"));
      return;
    }
    await save(
      "inventory/movements",
      {
        itemId: issueForm.itemId,
        kind: "issue",
        quantityMilli: Number(issueForm.quantity) * 1000,
        recipientId: issueForm.recipientId,
        reason: issueForm.reason,
        department: issueForm.department,
        issuedBy: issueForm.issuedBy,
        issuedDate: issueForm.issuedDate,
        returnDueDate: issueForm.returnDueDate,
      },
      "POST",
      () => {
        setShowIssueItem(false);
        setIssueForm({
          categoryId: "",
          itemId: "",
          quantity: 5,
          recipientId: "",
          reason: "Departmental clinical supply",
          department: "General OPD",
          issuedBy: "Pharmacy Admin",
          issuedDate: new Date().toISOString().split("T")[0],
          returnDueDate: "",
        });
      },
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
    setError("");
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

  // --- Void Handlers ---
  const handleConfirmVoidReceipt = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!voidingReceipt) return;
    try {
      await save(
        "inventory/movements",
        {
          itemId: voidingReceipt.itemId,
          kind: "writeoff",
          quantityMilli: voidingReceipt.quantityMilli,
          storeName: voidingReceipt.storeName || "",
          reason:
            `Void receipt ${voidingReceipt.reference || ""}: ${voidReason}`.trim(),
        },
        "POST",
        () => {
          setVoidingReceipt(null);
          setVoidReason("");
          setSuccessMsg(t("Stock receipt voided successfully"));
        },
      );
    } catch {
      setError(t("Cannot void receipt: stock has already been consumed"));
    }
  };

  const handleConfirmVoidIssue = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!voidingIssue) return;
    const returned = voidingIssue.returnedMilli || 0;
    const remaining = Math.max(0, voidingIssue.quantityMilli - returned);
    if (remaining <= 0) {
      setError("This issue has already been completely returned");
      return;
    }
    await save(
      "inventory/movements",
      {
        itemId: voidingIssue.itemId,
        kind: "return",
        originalId: voidingIssue.id,
        quantityMilli: remaining,
        restock: true,
        reason: `Void issue: ${voidReason}`.trim(),
      },
      "POST",
      () => {
        setVoidingIssue(null);
        setVoidReason("");
        setSuccessMsg(t("Issued item voided successfully"));
      },
    );
  };

  // --- Helpers ---
  const activeCategoriesList =
    allCategoriesCatalog.length > 0 ? allCategoriesCatalog : categories;
  const activeItemsList = allItemsCatalog.length > 0 ? allItemsCatalog : items;

  const getCategoryName = (catId: string) => {
    return (
      activeCategoriesList.find((c) => c.id === catId)?.name || t("General")
    );
  };

  const getItemName = (itemId: string, movementItemName?: string) => {
    if (movementItemName) return movementItemName;
    return (
      activeItemsList.find((i) => i.id === itemId)?.name ||
      t("Medical Supply Item")
    );
  };

  const getItemUnit = (itemId: string) => {
    return activeItemsList.find((i) => i.id === itemId)?.unit || "Units";
  };

  const getItemCategory = (itemId: string) => {
    const item = activeItemsList.find((i) => i.id === itemId);
    if (!item) return "-";
    return getCategoryName(item.categoryId);
  };

  const getItemBalance = (itemId: string) => {
    const item = activeItemsList.find((i) => i.id === itemId);
    return item ? item.balanceMilli / 1000 : 0;
  };

  const getRecipientDisplayName = (id?: string) => {
    if (!id) return "-";
    const found = recipients.find((r) => r.id === id);
    return found ? `${found.name} (${id})` : id;
  };

  // Filtered items strictly by selected category in modals
  const eligibleStockItems = useMemo(() => {
    if (!stockForm.categoryId) return [];
    return activeItemsList.filter((i) => i.categoryId === stockForm.categoryId);
  }, [activeItemsList, stockForm.categoryId]);

  const eligibleIssueItems = useMemo(() => {
    if (!issueForm.categoryId) return [];
    return activeItemsList.filter((i) => i.categoryId === issueForm.categoryId);
  }, [activeItemsList, issueForm.categoryId]);

  const selectedIssueItem = useMemo(() => {
    if (!issueForm.itemId) return null;
    return activeItemsList.find((i) => i.id === issueForm.itemId);
  }, [activeItemsList, issueForm.itemId]);

  const tabs = [
    { id: "items", label: t("Items"), icon: Package, count: totalItems },
    {
      id: "item-categories",
      label: t("Item Categories"),
      icon: Layers,
      count: totalCategories,
    },
    {
      id: "item-stocks",
      label: t("Item Stocks"),
      icon: ArrowDownToLine,
      count: totalMovements,
    },
    {
      id: "issued-items",
      label: t("Issued Items"),
      icon: ArrowUpFromLine,
      count: totalMovements,
    },
  ];

  const currentTotal =
    activeTab === "items"
      ? totalItems
      : activeTab === "item-categories"
        ? totalCategories
        : totalMovements;
  const totalPages = Math.max(1, Math.ceil(currentTotal / pageSize));

  return (
    <div
      className="space-y-6 legacy-workspace"
      data-workspace="inventory"
      data-ready={isLive ? "true" : "false"}
    >
      {error && (
        <p role="alert" className="error p-3 rounded-lg text-xs font-medium">
          {t(error)}
        </p>
      )}
      {successMsg && (
        <p
          role="status"
          className="p-3 bg-emerald-500/10 border border-emerald-500/30 text-emerald-800 dark:text-emerald-300 rounded-lg text-xs font-medium"
        >
          {t(successMsg)}
        </p>
      )}

      {/* Subtabs Nav */}
      <div className="border-b border-border/80 bg-card/50 backdrop-blur rounded-xl p-1.5 shadow-xs inventory-card-bg">
        <nav className="flex space-x-1 overflow-x-auto">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                data-tab={tab.id}
                onClick={() => {
                  setActiveTab(tab.id as InventoryTab);
                  setSearchTerm("");
                  setPage(1);
                }}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 whitespace-nowrap ${
                  isActive
                    ? "bg-primary text-primary-foreground shadow-xs"
                    : "text-slate-600 hover:text-slate-900 hover:bg-slate-100 dark:text-slate-200 dark:hover:text-white dark:hover:bg-slate-800/60"
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{tab.label}</span>
                <span
                  className={`text-xs px-1.5 py-0.5 rounded-full ${
                    isActive
                      ? "bg-primary-foreground/20 text-primary-foreground font-semibold"
                      : "bg-slate-200 text-slate-700 dark:bg-slate-700/60 dark:text-slate-200"
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
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-gradient-to-r from-emerald-500/10 via-emerald-500/5 to-transparent border border-emerald-500/20 rounded-xl p-3.5 text-sm shadow-xs inventory-card-bg">
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
                setError("");
                setItemForm({
                  name: "",
                  categoryId: activeCategoriesList[0]?.id || "",
                  unit: "Piece",
                  reorderLevel: "10",
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
                setError("");
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
                setError("");
                const defaultCat = activeCategoriesList[0]?.id || "";
                setStockForm({
                  categoryId: defaultCat,
                  itemId: "",
                  quantity: 50,
                  supplier: "",
                  storeName: "Central Hospital Store",
                  reference: "",
                  unitCost: 15.0,
                  reason: "Routine replenishment",
                  attachmentUrl: "",
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
                setError("");
                const defaultCat = activeCategoriesList[0]?.id || "";
                setIssueForm({
                  categoryId: defaultCat,
                  itemId: "",
                  quantity: 5,
                  recipientId: "",
                  reason: "Departmental clinical supply",
                  department: "General OPD",
                  issuedBy: "Pharmacy Admin",
                  issuedDate: new Date().toISOString().split("T")[0],
                  returnDueDate: "",
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

      {/* TAB 1: ITEMS */}
      {activeTab === "items" && (
        <div className="bg-card border border-border/80 rounded-2xl shadow-xs overflow-hidden inventory-card-bg">
          <div className="p-4 border-b border-border/60 flex flex-col sm:flex-row items-center justify-between gap-3">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {t("Medical Inventory Items")}
              </h2>
              <p className="text-xs text-muted-foreground">
                {t(
                  "Track medical supply stock balances and reorder thresholds",
                )}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2.5 w-full sm:w-auto">
              <select
                aria-label={t("Filter by category")}
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
                className="text-xs py-1.5 px-3 bg-background border border-border rounded-lg focus:outline-hidden inventory-input"
              >
                <option value="">{t("All Categories")}</option>
                {activeCategoriesList.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>

              <button
                type="button"
                onClick={() => setLowStockOnly(!lowStockOnly)}
                className={`text-xs flex items-center gap-1.5 py-1.5 px-3 rounded-lg border transition-all ${
                  lowStockOnly
                    ? "bg-amber-500/15 border-amber-500 text-amber-900 dark:text-amber-200 font-semibold"
                    : "border-border text-muted-foreground hover:bg-muted"
                }`}
              >
                <AlertTriangle className="w-3.5 h-3.5" />
                {t("Low Stock Only")}
              </button>

              <div className="relative flex-1 sm:w-64">
                <Search className="w-4 h-4 absolute left-3 top-2 text-muted-foreground" />
                <input
                  type="text"
                  aria-label={t("Search Items")}
                  placeholder={t("Search items...")}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-9 pr-3 py-1.5 text-xs bg-background border border-border rounded-lg focus:outline-hidden inventory-input"
                />
              </div>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs inventory-table-header">
                <tr>
                  <th className="px-4 py-3">{t("Item Name")}</th>
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Unit")}</th>
                  <th className="px-4 py-3">{t("Available Stock")}</th>
                  <th className="px-4 py-3">{t("Reorder Level")}</th>
                  <th className="px-4 py-3">{t("Status")}</th>
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {items.length === 0 ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-4 py-8 text-center text-muted-foreground text-xs"
                    >
                      {t("No items found")}
                    </td>
                  </tr>
                ) : (
                  items.map((item) => {
                    const balance = item.balanceMilli / 1000;
                    const reorder = item.reorderMilli / 1000;
                    const isLow = balance <= reorder && balance > 0;
                    const isOut = balance <= 0;

                    return (
                      <tr
                        key={item.id}
                        className="hover:bg-muted/30 transition-colors inventory-table-row"
                      >
                        <td className="px-4 py-3 font-semibold text-foreground">
                          {item.name}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {getCategoryName(item.categoryId)}
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                          {item.unit}
                        </td>
                        <td className="px-4 py-3 font-mono font-bold">
                          <span
                            className={
                              isOut
                                ? "text-rose-600 dark:text-rose-400"
                                : isLow
                                  ? "text-amber-600 dark:text-amber-400"
                                  : "text-emerald-600 dark:text-emerald-400"
                            }
                          >
                            {balance.toLocaleString()} {item.unit}
                          </span>
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                          {reorder.toLocaleString()} {item.unit}
                        </td>
                        <td className="px-4 py-3">
                          {isOut ? (
                            <span
                              data-status-badge="out-of-stock"
                              className="inline-flex items-center gap-1 text-xs px-2.5 py-0.5 rounded-full bg-rose-500/15 text-rose-800 dark:text-rose-300 font-medium"
                            >
                              <span className="w-1.5 h-1.5 rounded-full bg-rose-500" />
                              {t("Out of Stock")}
                            </span>
                          ) : isLow ? (
                            <span
                              data-status-badge="low-stock"
                              className="inline-flex items-center gap-1 text-xs px-2.5 py-0.5 rounded-full bg-amber-500/15 text-amber-800 dark:text-amber-300 font-medium"
                            >
                              <span className="w-1.5 h-1.5 rounded-full bg-amber-500" />
                              {t("Low Stock")}
                            </span>
                          ) : (
                            <span
                              data-status-badge="in-stock"
                              className="inline-flex items-center gap-1 text-xs px-2.5 py-0.5 rounded-full bg-emerald-500/15 text-emerald-800 dark:text-emerald-300 font-medium"
                            >
                              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                              {t("In Stock")}
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              type="button"
                              onClick={() => setViewingItem(item)}
                              aria-label={t("Item Details")}
                              title={t("Item Details")}
                              className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-muted"
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                            <button
                              type="button"
                              onClick={() => openEditItem(item)}
                              aria-label={t("Edit Item")}
                              title={t("Edit Item")}
                              className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-muted"
                            >
                              <Edit className="w-4 h-4" />
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
                              className="p-1 text-muted-foreground hover:text-rose-600 rounded-md hover:bg-muted"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination Controls */}
          <div className="p-3 border-t border-border/60 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {t("Page")} {page} {t("of")} {totalPages} ({totalItems}{" "}
              {t("Items")})
            </span>
            <div className="flex items-center gap-2">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                <ChevronLeft className="w-3.5 h-3.5" /> {t("Previous")}
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                {t("Next")} <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: CATEGORIES */}
      {activeTab === "item-categories" && (
        <div className="bg-card border border-border/80 rounded-2xl shadow-xs overflow-hidden inventory-card-bg">
          <div className="p-4 border-b border-border/60 flex flex-col sm:flex-row items-center justify-between gap-3">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {t("Inventory Categories")}
              </h2>
              <p className="text-xs text-muted-foreground">
                {t(
                  "Classification taxonomy for hospital equipment and consumables",
                )}
              </p>
            </div>
            <div className="relative flex-1 sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2 text-muted-foreground" />
              <input
                type="text"
                aria-label={t("Search Categories")}
                placeholder={t("Search categories...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-xs bg-background border border-border rounded-lg focus:outline-hidden inventory-input"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs inventory-table-header">
                <tr>
                  <th className="px-4 py-3">{t("Category Name")}</th>
                  <th className="px-4 py-3">{t("Description")}</th>
                  <th className="px-4 py-3">{t("Items Assigned")}</th>
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {categories.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-4 py-8 text-center text-muted-foreground text-xs"
                    >
                      {t("No categories found")}
                    </td>
                  </tr>
                ) : (
                  categories.map((cat) => {
                    const assigned = allItemsCatalog.filter(
                      (i) => i.categoryId === cat.id,
                    ).length;
                    return (
                      <tr
                        key={cat.id}
                        className="hover:bg-muted/30 transition-colors inventory-table-row"
                      >
                        <td className="px-4 py-3 font-semibold text-foreground">
                          {cat.name}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground max-w-xs truncate">
                          {cat.description || "-"}
                        </td>
                        <td className="px-4 py-3">
                          <span className="font-mono text-xs px-2 py-0.5 rounded-md bg-muted text-muted-foreground">
                            {assigned} {t("Items")}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              type="button"
                              onClick={() => openEditCategory(cat)}
                              aria-label={t("Edit Category")}
                              title={t("Edit Category")}
                              className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-muted"
                            >
                              <Edit className="w-4 h-4" />
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
                              className="p-1 text-muted-foreground hover:text-rose-600 rounded-md hover:bg-muted"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
          {/* Pagination Controls */}
          <div className="p-3 border-t border-border/60 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {t("Page")} {page} {t("of")} {totalPages} ({totalCategories}{" "}
              {t("Item Categories")})
            </span>
            <div className="flex items-center gap-2">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                <ChevronLeft className="w-3.5 h-3.5" /> {t("Previous")}
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                {t("Next")} <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* TAB 3: ITEM STOCKS (RECEIPTS) */}
      {activeTab === "item-stocks" && (
        <div className="bg-card border border-border/80 rounded-2xl shadow-xs overflow-hidden inventory-card-bg">
          <div className="p-4 border-b border-border/60 flex flex-col sm:flex-row items-center justify-between gap-3">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {t("Item Stocks Received")}
              </h2>
              <p className="text-xs text-muted-foreground">
                {t("Stock intake logs, procurement receipts, and lot tracking")}
              </p>
            </div>
            <div className="relative flex-1 sm:w-64">
              <Search className="w-4 h-4 absolute left-3 top-2 text-muted-foreground" />
              <input
                type="text"
                aria-label={t("Search Movements")}
                placeholder={t("Search stock receipts...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-xs bg-background border border-border rounded-lg focus:outline-hidden inventory-input"
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs inventory-table-header">
                <tr>
                  <th className="px-4 py-3">{t("Item")}</th>
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Supplier / Store")}</th>
                  <th className="px-4 py-3">{t("Reference No")}</th>
                  <th className="px-4 py-3">{t("Quantity Received")}</th>
                  <th className="px-4 py-3">{t("Total Cost (ETB)")}</th>
                  <th className="px-4 py-3">{t("Date")}</th>
                  <th className="px-4 py-3">{t("Receipt Attachment")}</th>
                  <th className="px-4 py-3 text-right">{t("Actions")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {movements
                  .filter((m) => m.kind === "receive")
                  .map((m) => (
                    <tr
                      key={m.id}
                      className="hover:bg-muted/30 transition-colors inventory-table-row"
                    >
                      <td className="px-4 py-3 font-semibold text-foreground">
                        {getItemName(m.itemId, m.itemName)}
                        {m.reason && (
                          <div className="text-[11px] font-normal text-muted-foreground line-clamp-1">
                            {m.reason}
                          </div>
                        )}
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
                      <td className="px-4 py-3 text-xs">
                        {m.attachmentUrl ? (
                          <a
                            href={m.attachmentUrl}
                            target="_blank"
                            rel="noreferrer"
                            className="inline-flex items-center gap-1 text-primary hover:underline font-medium"
                          >
                            <FileText className="w-3.5 h-3.5" />
                            {t("View Receipt")}
                          </a>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => setViewingMovement(m)}
                            aria-label={t("Stock Details")}
                            title={t("Stock Details")}
                            className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-muted"
                          >
                            <Eye className="w-4 h-4" />
                          </button>
                          <button
                            type="button"
                            onClick={() => {
                              setVoidingReceipt(m);
                              setVoidReason("");
                            }}
                            aria-label={t("Void Receipt")}
                            title={t("Void Receipt")}
                            className="p-1 text-muted-foreground hover:text-rose-600 rounded-md hover:bg-muted"
                          >
                            <Ban className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
          {/* Pagination Controls */}
          <div className="p-3 border-t border-border/60 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {t("Page")} {page} {t("of")} {totalPages} ({totalMovements}{" "}
              {t("Item Stocks")})
            </span>
            <div className="flex items-center gap-2">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                <ChevronLeft className="w-3.5 h-3.5" /> {t("Previous")}
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                {t("Next")} <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* TAB 4: ISSUED ITEMS */}
      {activeTab === "issued-items" && (
        <div className="bg-card border border-border/80 rounded-2xl shadow-xs overflow-hidden inventory-card-bg">
          <div className="p-4 border-b border-border/60 flex flex-col sm:flex-row items-center justify-between gap-3">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {t("Issued Items & Consumables")}
              </h2>
              <p className="text-xs text-muted-foreground">
                {t(
                  "Items issued to hospital departments, staff members, and return status",
                )}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2.5 w-full sm:w-auto">
              <select
                aria-label={t("Filter by status")}
                value={issueStatusFilter}
                onChange={(e) =>
                  setIssueStatusFilter(
                    e.target.value as "all" | "returnable" | "returned",
                  )
                }
                className="text-xs py-1.5 px-3 bg-background border border-border rounded-lg focus:outline-hidden inventory-input"
              >
                <option value="all">{t("All Statuses")}</option>
                <option value="returnable">{t("Pending Return")}</option>
                <option value="returned">{t("Returned")}</option>
              </select>

              <div className="relative flex-1 sm:w-64">
                <Search className="w-4 h-4 absolute left-3 top-2 text-muted-foreground" />
                <input
                  type="text"
                  aria-label={t("Search Movements")}
                  placeholder={t("Search issued items...")}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-9 pr-3 py-1.5 text-xs bg-background border border-border rounded-lg focus:outline-hidden inventory-input"
                />
              </div>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground uppercase text-xs inventory-table-header">
                <tr>
                  <th className="px-4 py-3">{t("Item")}</th>
                  <th className="px-4 py-3">{t("Category")}</th>
                  <th className="px-4 py-3">{t("Issued To")}</th>
                  <th className="px-4 py-3">{t("Department")}</th>
                  <th className="px-4 py-3">{t("Quantity Issued")}</th>
                  <th className="px-4 py-3">{t("Returned / Remaining")}</th>
                  <th className="px-4 py-3">{t("Status")}</th>
                  <th className="px-4 py-3">{t("Issue Date")}</th>
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
                  .map((m) => {
                    const issued = m.quantityMilli / 1000;
                    const returned = (m.returnedMilli || 0) / 1000;
                    const remaining = Math.max(0, issued - returned);
                    const isFullyReturned = remaining === 0;
                    const isPartial = returned > 0 && remaining > 0;

                    return (
                      <tr
                        key={m.id}
                        className="hover:bg-muted/30 transition-colors inventory-table-row"
                      >
                        <td className="px-4 py-3 font-semibold text-foreground">
                          {getItemName(m.itemId, m.itemName)}
                          {m.reason && (
                            <div className="text-[11px] font-normal text-muted-foreground line-clamp-1">
                              {m.reason}
                            </div>
                          )}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {getItemCategory(m.itemId)}
                        </td>
                        <td className="px-4 py-3 font-medium text-xs text-foreground">
                          {getRecipientDisplayName(m.recipientId)}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {m.department || "-"}
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
                            <span
                              data-status-badge="returned"
                              className="inline-flex items-center gap-1 text-xs px-2.5 py-0.5 rounded-full bg-emerald-500/15 text-emerald-800 dark:text-emerald-300 font-medium"
                            >
                              <CheckCircle className="w-3.5 h-3.5" />
                              {t("Returned")}
                            </span>
                          ) : isPartial ? (
                            <button
                              type="button"
                              onClick={() => openReturnModal(m)}
                              aria-label={t("Partial Return")}
                              data-status-badge="partial"
                              className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/15 text-amber-700 dark:text-amber-300 hover:bg-amber-500/25 transition-colors cursor-pointer"
                            >
                              <RotateCcw className="w-3 h-3" />
                              {t("Partial Return")} ({returned}/{issued})
                            </button>
                          ) : (
                            <span
                              data-status-badge="issued"
                              className="inline-flex items-center gap-1 text-xs px-2.5 py-0.5 rounded-full bg-amber-500/15 text-amber-800 dark:text-amber-300 font-medium"
                            >
                              <ArrowUpFromLine className="w-3.5 h-3.5" />
                              {t("Issued")}
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          <div>
                            {m.issuedDate ||
                              new Date(m.createdAt).toLocaleDateString()}
                          </div>
                          {m.returnDueDate && (
                            <div className="text-[10px] text-amber-600 dark:text-amber-400 font-mono">
                              Due: {m.returnDueDate}
                            </div>
                          )}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              type="button"
                              onClick={() => setViewingIssue(m)}
                              aria-label={t("View issue record")}
                              className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-muted"
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                            {!isFullyReturned && (
                              <>
                                <button
                                  type="button"
                                  onClick={() => openReturnModal(m)}
                                  className="text-xs px-2.5 py-1 bg-primary text-primary-foreground font-medium rounded-md hover:bg-primary/90 transition-colors shadow-2xs"
                                >
                                  {t("Return Item")}
                                </button>
                                <button
                                  type="button"
                                  onClick={() => {
                                    setVoidingIssue(m);
                                    setVoidReason("");
                                  }}
                                  aria-label={t("Void Issue")}
                                  title={t("Void Issue")}
                                  className="p-1 text-muted-foreground hover:text-rose-600 rounded-md hover:bg-muted"
                                >
                                  <Ban className="w-4 h-4" />
                                </button>
                              </>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
          {/* Pagination Controls */}
          <div className="p-3 border-t border-border/60 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {t("Page")} {page} {t("of")} {totalPages} ({totalMovements}{" "}
              {t("Issued Items")})
            </span>
            <div className="flex items-center gap-2">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                <ChevronLeft className="w-3.5 h-3.5" /> {t("Previous")}
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="px-2.5 py-1 border border-border rounded-md hover:bg-muted disabled:opacity-40 flex items-center gap-1"
              >
                {t("Next")} <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 1: ADD / EDIT ITEM */}
      {(showAddItem || editingItem) && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {editingItem ? t("Edit Item") : t("New Item Registration")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <form onSubmit={handleSaveItem} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Category")}
                </label>
                <select
                  aria-label={t("Category")}
                  value={
                    itemForm.categoryId ||
                    editingItem?.categoryId ||
                    activeCategoriesList[0]?.id ||
                    ""
                  }
                  onChange={(e) =>
                    setItemForm({ ...itemForm, categoryId: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                >
                  <option value="">{t("Select Category")}</option>
                  {activeCategoriesList.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Item Name")}
                </label>
                <input
                  type="text"
                  aria-label={t("Item Name")}
                  placeholder={t("e.g. Surgical Gloves (Latex)")}
                  value={itemForm.name}
                  onChange={(e) =>
                    setItemForm({ ...itemForm, name: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Unit")}
                  </label>
                  <input
                    type="text"
                    aria-label={t("Unit")}
                    placeholder="Piece, Box, Pair"
                    value={itemForm.unit}
                    onChange={(e) =>
                      setItemForm({ ...itemForm, unit: e.target.value })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Reorder Level")}
                  </label>
                  <input
                    type="number"
                    step="any"
                    min={0}
                    aria-label={t("Reorder Level")}
                    value={itemForm.reorderLevel}
                    onChange={(e) =>
                      setItemForm({
                        ...itemForm,
                        reorderLevel: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Description")}
                </label>
                <textarea
                  rows={2}
                  aria-label="Description"
                  value={itemForm.description}
                  onChange={(e) =>
                    setItemForm({ ...itemForm, description: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddItem(false);
                    setEditingItem(null);
                  }}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary text-xs px-4 py-2 rounded-lg font-semibold"
                >
                  {saving
                    ? t("Saving...")
                    : editingItem
                      ? t("Update Item")
                      : t("Save Item")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 2: ADD / EDIT CATEGORY */}
      {(showAddCategory || editingCategory) && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {editingCategory ? t("Edit Category") : t("New Item Category")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <form onSubmit={handleSaveCategory} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Category Name")}
                </label>
                <input
                  type="text"
                  aria-label={t("Category Name")}
                  placeholder={t("e.g. Surgical Equipment")}
                  value={categoryForm.name}
                  onChange={(e) =>
                    setCategoryForm({ ...categoryForm, name: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Description")}
                </label>
                <textarea
                  rows={3}
                  aria-label="Description"
                  value={categoryForm.description}
                  onChange={(e) =>
                    setCategoryForm({
                      ...categoryForm,
                      description: e.target.value,
                    })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddCategory(false);
                    setEditingCategory(null);
                  }}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary text-xs px-4 py-2 rounded-lg font-semibold"
                >
                  {saving
                    ? t("Saving...")
                    : editingCategory
                      ? t("Update Category")
                      : t("Save Category")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 3: RECEIVE NEW STOCK */}
      {showAddStock && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 max-h-[90vh] overflow-y-auto inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {t("Receive New Stock")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <form onSubmit={handleSaveStockReceive} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Category")}
                </label>
                <select
                  aria-label={t("Category")}
                  value={stockForm.categoryId}
                  onChange={(e) => {
                    const nextCat = e.target.value;
                    setStockForm((prev) => ({
                      ...prev,
                      categoryId: nextCat,
                      itemId: "",
                    }));
                  }}
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                >
                  <option value="">{t("Select Category")}</option>
                  {activeCategoriesList.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Item")}
                </label>
                <select
                  aria-label={t("Item")}
                  value={stockForm.itemId}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, itemId: e.target.value })
                  }
                  disabled={eligibleStockItems.length === 0}
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input disabled:opacity-50"
                  required
                >
                  {eligibleStockItems.length === 0 ? (
                    <option value="">{t("No items in this category")}</option>
                  ) : (
                    <>
                      <option value="">{t("Select Item")}</option>
                      {eligibleStockItems.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name} ({t("Current:")}{" "}
                          {item.balanceMilli / 1000} {item.unit})
                        </option>
                      ))}
                    </>
                  )}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Quantity Received")}
                  </label>
                  <input
                    type="number"
                    min={1}
                    aria-label={t("Quantity")}
                    value={stockForm.quantity}
                    onChange={(e) =>
                      setStockForm({
                        ...stockForm,
                        quantity: Number(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Unit Cost")} (ETB)
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
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Supplier")}
                  </label>
                  <input
                    type="text"
                    aria-label={t("Supplier")}
                    placeholder="e.g. Ethiopian Pharma Supply"
                    value={stockForm.supplier}
                    onChange={(e) =>
                      setStockForm({ ...stockForm, supplier: e.target.value })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Store")}
                  </label>
                  <input
                    type="text"
                    aria-label={t("Store Name")}
                    value={stockForm.storeName}
                    onChange={(e) =>
                      setStockForm({
                        ...stockForm,
                        storeName: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Reference No")}
                </label>
                <input
                  type="text"
                  aria-label={t("Reference")}
                  placeholder="PO-2026-001"
                  value={stockForm.reference}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, reference: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>

              {/* Receipt File Attachment Upload */}
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Receipt Attachment")}
                </label>
                <div className="flex items-center gap-2">
                  <label className="btn-secondary text-xs px-3 py-1.5 rounded-lg border border-border cursor-pointer flex items-center gap-1.5 hover:bg-muted">
                    <Upload className="w-3.5 h-3.5" />
                    {uploadingReceipt ? t("Uploading...") : t("Upload Receipt")}
                    <input
                      type="file"
                      accept="image/*,application/pdf"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) void handleReceiptUpload(file);
                      }}
                    />
                  </label>
                  {stockForm.attachmentUrl && (
                    <span className="text-xs text-emerald-600 dark:text-emerald-400 font-medium flex items-center gap-1">
                      <CheckCircle className="w-3.5 h-3.5" />
                      {t("Attached")}
                    </span>
                  )}
                </div>
                <input
                  type="text"
                  placeholder="https://... (or upload above)"
                  aria-label={t("Attachment URL")}
                  value={stockForm.attachmentUrl}
                  onChange={(e) =>
                    setStockForm({
                      ...stockForm,
                      attachmentUrl: e.target.value,
                    })
                  }
                  className="w-full px-3 py-1.5 text-xs border border-border rounded-lg bg-background inventory-input mt-1.5 font-mono"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Reason / Notes")}
                </label>
                <textarea
                  rows={2}
                  aria-label="Reason"
                  value={stockForm.reason}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, reason: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowAddStock(false)}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={
                    saving ||
                    !stockForm.itemId ||
                    eligibleStockItems.length === 0
                  }
                  className="btn-primary text-xs px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
                >
                  {saving ? t("Saving...") : t("Confirm Receive")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 4: ISSUE ITEM */}
      {showIssueItem && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 max-h-[90vh] overflow-y-auto inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {t("Issue Item to Staff")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <form onSubmit={handleSaveIssue} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Category")}
                </label>
                <select
                  aria-label={t("Category")}
                  value={issueForm.categoryId}
                  onChange={(e) => {
                    const nextCat = e.target.value;
                    setIssueForm((prev) => ({
                      ...prev,
                      categoryId: nextCat,
                      itemId: "",
                    }));
                  }}
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                >
                  <option value="">{t("Select Category")}</option>
                  {activeCategoriesList.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Item")}
                </label>
                <select
                  aria-label={t("Item")}
                  value={issueForm.itemId}
                  onChange={(e) =>
                    setIssueForm({ ...issueForm, itemId: e.target.value })
                  }
                  disabled={eligibleIssueItems.length === 0}
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input disabled:opacity-50"
                  required
                >
                  {eligibleIssueItems.length === 0 ? (
                    <option value="">{t("No items in this category")}</option>
                  ) : (
                    <>
                      <option value="">{t("Select Item")}</option>
                      {eligibleIssueItems.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name} ({t("Available:")}{" "}
                          {item.balanceMilli / 1000} {item.unit})
                        </option>
                      ))}
                    </>
                  )}
                </select>
                {selectedIssueItem && (
                  <p className="text-[11px] text-muted-foreground mt-1">
                    {t("Live balance:")}{" "}
                    <strong className="text-foreground">
                      {selectedIssueItem.balanceMilli / 1000}{" "}
                      {selectedIssueItem.unit}
                    </strong>
                  </p>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Department")}
                  </label>
                  <select
                    aria-label={t("Department")}
                    value={issueForm.department}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        department: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  >
                    {DEPARTMENTS.map((dept) => (
                      <option key={dept} value={dept}>
                        {t(dept)}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Quantity Issued")}
                  </label>
                  <input
                    type="number"
                    min={1}
                    aria-label={t("Quantity to Issue")}
                    value={issueForm.quantity}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        quantity: Number(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Recipient Staff Member")}
                </label>
                <div className="space-y-1.5">
                  <input
                    type="text"
                    value={recipientSearch}
                    onChange={(e) => setRecipientSearch(e.target.value)}
                    aria-label={t("Search staff")}
                    placeholder={t("Filter staff...")}
                    className="w-full px-3 py-1.5 text-xs border border-border rounded-lg bg-background inventory-input"
                  />
                  <select
                    aria-label={t("Recipient")}
                    value={issueForm.recipientId}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        recipientId: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  >
                    <option value="">{t("Select staff")}</option>
                    {recipients.map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.name} ({user.role}) - {user.id}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Issue Date")}
                  </label>
                  <input
                    type="date"
                    aria-label={t("Issued Date")}
                    value={issueForm.issuedDate}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        issuedDate: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {t("Return Due Date")}
                  </label>
                  <input
                    type="date"
                    aria-label={t("Return Due Date")}
                    value={issueForm.returnDueDate}
                    onChange={(e) =>
                      setIssueForm({
                        ...issueForm,
                        returnDueDate: e.target.value,
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Reason / Clinical Notes")}
                </label>
                <textarea
                  rows={2}
                  aria-label="Reason"
                  value={issueForm.reason}
                  onChange={(e) =>
                    setIssueForm({ ...issueForm, reason: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowIssueItem(false)}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={
                    saving ||
                    !issueForm.itemId ||
                    eligibleIssueItems.length === 0 ||
                    !issueForm.recipientId
                  }
                  className="btn-primary text-xs px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
                >
                  {saving ? t("Saving...") : t("Confirm Issue")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 5: RETURN ISSUED ITEM */}
      {returningIssue && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {t("Return Issued Item")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <div className="p-3 bg-muted/40 rounded-xl space-y-1 text-xs">
              <p>
                {t("Item:")}{" "}
                <strong className="text-foreground">
                  {getItemName(returningIssue.itemId, returningIssue.itemName)}
                </strong>
              </p>
              <p>
                {t("Issued To:")}{" "}
                <span className="text-foreground">
                  {getRecipientDisplayName(returningIssue.recipientId)}
                </span>
              </p>
              <p>
                {t("Issued Quantity:")}{" "}
                <span className="font-mono">
                  {returningIssue.quantityMilli / 1000}{" "}
                  {getItemUnit(returningIssue.itemId)}
                </span>
              </p>
              <p>
                {t("Already Returned:")}{" "}
                <span className="font-mono text-emerald-600 dark:text-emerald-400">
                  {(returningIssue.returnedMilli || 0) / 1000}{" "}
                  {getItemUnit(returningIssue.itemId)}
                </span>
              </p>
              <p>
                {t("Remaining Due:")}{" "}
                <strong className="font-mono text-amber-600 dark:text-amber-400">
                  {(returningIssue.quantityMilli -
                    (returningIssue.returnedMilli || 0)) /
                    1000}{" "}
                  {getItemUnit(returningIssue.itemId)}
                </strong>
              </p>
            </div>

            <form onSubmit={handleSaveReturn} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Quantity to Return")}
                </label>
                <input
                  type="number"
                  min={1}
                  aria-label={t("Quantity to Return")}
                  max={
                    (returningIssue.quantityMilli -
                      (returningIssue.returnedMilli || 0)) /
                    1000
                  }
                  value={returnForm.quantity}
                  onChange={(e) =>
                    setReturnForm({
                      ...returnForm,
                      quantity: Number(e.target.value),
                    })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="restock-checkbox"
                  checked={returnForm.restock}
                  onChange={(e) =>
                    setReturnForm({
                      ...returnForm,
                      restock: e.target.checked,
                    })
                  }
                  className="rounded-sm border-border text-primary"
                />
                <label
                  htmlFor="restock-checkbox"
                  className="text-xs font-medium text-foreground cursor-pointer"
                >
                  {t("Restock to Inventory")} (
                  {t("Check if item is clean & reusable")})
                </label>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Return Reason")}
                </label>
                <textarea
                  rows={2}
                  aria-label="Return Reason"
                  value={returnForm.reason}
                  onChange={(e) =>
                    setReturnForm({ ...returnForm, reason: e.target.value })
                  }
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => setReturningIssue(null)}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="btn-primary text-xs px-4 py-2 rounded-lg font-semibold"
                >
                  {saving ? t("Saving...") : t("Confirm Return")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 6: VOID STOCK RECEIPT */}
      {voidingReceipt && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {t("Void Stock Receipt")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              {t("Are you sure you want to void this stock receipt?")}{" "}
              {t(
                "This will atomically deduct the received quantity from available stock and record an audited writeoff.",
              )}
            </p>
            <div className="p-3 bg-muted/40 rounded-xl space-y-1 text-xs">
              <p>
                {t("Item:")}{" "}
                <strong className="text-foreground">
                  {getItemName(voidingReceipt.itemId, voidingReceipt.itemName)}
                </strong>
              </p>
              <p>
                {t("Reference No:")}{" "}
                <span className="font-mono font-semibold">
                  {voidingReceipt.reference || "-"}
                </span>
              </p>
              <p>
                {t("Quantity Received:")}{" "}
                <span className="font-mono">
                  {voidingReceipt.quantityMilli / 1000}{" "}
                  {getItemUnit(voidingReceipt.itemId)}
                </span>
              </p>
            </div>
            <form onSubmit={handleConfirmVoidReceipt} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Reason for voiding")}
                </label>
                <textarea
                  rows={2}
                  aria-label="Void Reason"
                  placeholder={t("e.g. Data entry error / returned to vendor")}
                  value={voidReason}
                  onChange={(e) => setVoidReason(e.target.value)}
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => setVoidingReceipt(null)}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving || !voidReason.trim()}
                  className="btn-primary bg-rose-600 hover:bg-rose-700 text-xs px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
                >
                  {saving ? t("Saving...") : t("Confirm Void")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 7: VOID ISSUED ITEM */}
      {voidingIssue && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {t("Void Issued Item")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              {t("Are you sure you want to void this issued item?")}{" "}
              {t(
                "This will atomically restock the remaining issued quantity and record an audited void reversal.",
              )}
            </p>
            <div className="p-3 bg-muted/40 rounded-xl space-y-1 text-xs">
              <p>
                {t("Item:")}{" "}
                <strong className="text-foreground">
                  {getItemName(voidingIssue.itemId, voidingIssue.itemName)}
                </strong>
              </p>
              <p>
                {t("Issued To:")}{" "}
                <span className="text-foreground">
                  {getRecipientDisplayName(voidingIssue.recipientId)}
                </span>
              </p>
              <p>
                {t("Quantity Issued:")}{" "}
                <span className="font-mono">
                  {voidingIssue.quantityMilli / 1000}{" "}
                  {getItemUnit(voidingIssue.itemId)}
                </span>
              </p>
            </div>
            <form onSubmit={handleConfirmVoidIssue} className="space-y-3.5">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {t("Reason for voiding")}
                </label>
                <textarea
                  rows={2}
                  aria-label="Void Reason"
                  placeholder={t("e.g. Issue cancelled by doctor")}
                  value={voidReason}
                  onChange={(e) => setVoidReason(e.target.value)}
                  className="w-full px-3 py-2 text-xs border border-border rounded-lg bg-background inventory-input"
                  required
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-border">
                <button
                  type="button"
                  onClick={() => setVoidingIssue(null)}
                  className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                >
                  {t("Cancel")}
                </button>
                <button
                  type="submit"
                  disabled={saving || !voidReason.trim()}
                  className="btn-primary bg-rose-600 hover:bg-rose-700 text-xs px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
                >
                  {saving ? t("Saving...") : t("Confirm Void")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 8: VIEW ITEM DETAILS */}
      {viewingItem && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-lg w-full shadow-xl space-y-4 max-h-[90vh] overflow-y-auto inventory-modal-bg">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-base font-bold text-foreground">
                  {viewingItem.name}
                </h3>
                <p className="text-xs text-muted-foreground">
                  {getCategoryName(viewingItem.categoryId)}
                </p>
              </div>
              <span className="text-xs px-2.5 py-0.5 rounded-full bg-primary/10 text-primary font-mono font-medium">
                {viewingItem.unit}
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3 p-3 bg-muted/40 rounded-xl text-xs">
              <div>
                <span className="text-muted-foreground block">
                  {t("Available Stock:")}
                </span>
                <strong className="text-base text-foreground font-mono">
                  {(viewingItem.balanceMilli / 1000).toLocaleString()}{" "}
                  {viewingItem.unit}
                </strong>
              </div>
              <div>
                <span className="text-muted-foreground block">
                  {t("Reorder Level:")}
                </span>
                <strong className="text-base text-muted-foreground font-mono">
                  {(viewingItem.reorderMilli / 1000).toLocaleString()}{" "}
                  {viewingItem.unit}
                </strong>
              </div>
            </div>

            <div>
              <h4 className="text-xs font-semibold text-foreground mb-1">
                {t("Description")}
              </h4>
              <p className="text-xs text-muted-foreground">
                {viewingItem.description || t("No description provided.")}
              </p>
            </div>

            <div className="flex justify-end pt-2 border-t border-border">
              <button
                type="button"
                onClick={() => setViewingItem(null)}
                className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
              >
                {t("Close")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 9: VIEW MOVEMENT DETAILS */}
      {(viewingMovement || viewingIssue) && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-xl space-y-4 inventory-modal-bg">
            {(() => {
              const m = viewingMovement || viewingIssue!;
              return (
                <>
                  <div className="flex items-start justify-between">
                    <div>
                      <h3 className="text-base font-bold text-foreground">
                        {getItemName(m.itemId, m.itemName)}
                      </h3>
                      <p className="text-xs text-muted-foreground">
                        {getItemCategory(m.itemId)}
                      </p>
                    </div>
                    <span className="uppercase text-xs font-mono font-bold px-2 py-0.5 rounded-md bg-muted text-foreground">
                      {m.kind}
                    </span>
                  </div>

                  <div className="space-y-2 p-3 bg-muted/40 rounded-xl text-xs">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">
                        {t("Quantity:")}
                      </span>
                      <strong className="font-mono">
                        {(m.quantityMilli / 1000).toLocaleString()}{" "}
                        {getItemUnit(m.itemId)}
                      </strong>
                    </div>
                    {m.supplier && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">
                          {t("Supplier:")}
                        </span>
                        <span>{m.supplier}</span>
                      </div>
                    )}
                    {m.storeName && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">
                          {t("Store:")}
                        </span>
                        <span>{m.storeName}</span>
                      </div>
                    )}
                    {m.recipientId && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">
                          {t("Recipient:")}
                        </span>
                        <span>{getRecipientDisplayName(m.recipientId)}</span>
                      </div>
                    )}
                    {m.department && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">
                          {t("Department:")}
                        </span>
                        <span>{t(m.department)}</span>
                      </div>
                    )}
                    {m.reference && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">
                          {t("Reference No:")}
                        </span>
                        <span className="font-mono">{m.reference}</span>
                      </div>
                    )}
                    {m.costMinor ? (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">
                          {t("Cost:")}
                        </span>
                        <span className="font-mono">
                          {(m.costMinor / 100).toLocaleString(undefined, {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })}{" "}
                          ETB
                        </span>
                      </div>
                    ) : null}
                    {m.attachmentUrl && (
                      <div className="flex justify-between items-center">
                        <span className="text-muted-foreground">
                          {t("Receipt Attachment:")}
                        </span>
                        <a
                          href={m.attachmentUrl}
                          target="_blank"
                          rel="noreferrer"
                          className="text-primary hover:underline font-medium inline-flex items-center gap-1"
                        >
                          <FileText className="w-3.5 h-3.5" />
                          {t("View Receipt")}
                        </a>
                      </div>
                    )}
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">
                        {t("Date:")}
                      </span>
                      <span>{new Date(m.createdAt).toLocaleString()}</span>
                    </div>
                  </div>

                  <div>
                    <h4 className="text-xs font-semibold text-foreground mb-1">
                      {t("Reason / Notes")}
                    </h4>
                    <p className="text-xs text-muted-foreground">{m.reason}</p>
                  </div>

                  <div className="flex justify-end pt-2 border-t border-border">
                    <button
                      type="button"
                      onClick={() => {
                        setViewingMovement(null);
                        setViewingIssue(null);
                      }}
                      className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
                    >
                      {t("Close")}
                    </button>
                  </div>
                </>
              );
            })()}
          </div>
        </div>
      )}

      {/* MODAL 10: DELETE CONFIRMATION */}
      {deleteConfirm && (
        <div
          role="dialog"
          aria-modal="true"
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs"
        >
          <div className="bg-card border border-border rounded-2xl p-6 max-w-sm w-full shadow-xl space-y-4 inventory-modal-bg">
            <h3 className="text-base font-bold text-foreground">
              {deleteConfirm.type === "category"
                ? t("Delete Category")
                : t("Delete Item")}
            </h3>
            {error && (
              <p
                role="alert"
                className="text-xs text-rose-500 bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-lg error"
              >
                {t(error)}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              {deleteConfirm.type === "category"
                ? t("Are you sure you want to delete this category?")
                : t("Are you sure you want to delete this item?")}{" "}
              <strong className="text-foreground">{deleteConfirm.name}</strong>
            </p>
            <div className="flex justify-end gap-2 pt-2 border-t border-border">
              <button
                type="button"
                onClick={() => setDeleteConfirm(null)}
                className="btn-secondary text-xs px-4 py-2 rounded-lg border border-border"
              >
                {t("Cancel")}
              </button>
              <button
                type="button"
                disabled={saving}
                onClick={() =>
                  handleDelete(deleteConfirm.type, deleteConfirm.id)
                }
                className="btn-primary bg-rose-600 hover:bg-rose-700 text-xs px-4 py-2 rounded-lg font-semibold"
              >
                {saving ? t("Deleting...") : t("Confirm Delete")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = path.startsWith("http")
    ? path
    : `/api/hms/${path.replace(/^\/+/, "")}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    try {
      const body = await res.json();
      msg = body.error || body.message || msg;
    } catch {}
    throw new Error(msg);
  }
  return res.json();
}
