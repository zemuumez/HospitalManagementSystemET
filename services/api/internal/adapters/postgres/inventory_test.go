package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testInventory(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()
	a := actors[0]
	inv := application.Inventory{Store: store}
	category, e := inv.SaveCategory(ctx, a, "", domain.InventoryCategoryInput{Name: "General supplies", Active: true})
	if e != nil {
		t.Fatal(e)
	}
	item, e := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: category.ID, Name: "Synthetic consumable", Unit: "kg", Active: true, ReorderMilli: 1000})
	if e != nil {
		t.Fatal(e)
	}
	receive := domain.InventoryMovementInput{ItemID: item.ID, Kind: "receive", QuantityMilli: 2500, Reference: "PO-1", Supplier: "Synthetic supplier", StoreName: "Main store", CostMinor: 1234, Reason: "Received purchase"}
	receipt, e := inv.Move(ctx, a, receive, "inventory-receipt-0001")
	if e != nil || receipt.DeltaMilli != 2500 {
		t.Fatal("inventory receive", e)
	}
	retry, e := inv.Move(ctx, a, receive, "inventory-receipt-0001")
	if e != nil || retry.ID != receipt.ID {
		t.Fatal("receipt retry", e)
	}
	changed := receive
	changed.QuantityMilli = 3500
	if _, e = inv.Move(ctx, a, changed, "inventory-receipt-0001"); !errors.Is(e, domain.ErrConflict) {
		t.Fatal("changed receipt replay", e)
	}
	issue := domain.InventoryMovementInput{ItemID: item.ID, Kind: "issue", QuantityMilli: 2000, RecipientID: actors[1].ID, Reason: "Ward supply"}
	var movements [2]domain.InventoryMovement
	var errs [2]error
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			movements[n], errs[n] = inv.Move(ctx, a, issue, fmt.Sprintf("inventory-issue-%08d", n))
		}(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	if errs[winner] != nil || !errors.Is(errs[1-winner], domain.ErrStale) {
		t.Fatal("inventory issue race", errs)
	}
	issued := movements[winner]
	low, e := inv.Items(ctx, a, "", true, 1)
	if e != nil || len(low) != 1 || low[0].BalanceMilli != 500 {
		t.Fatal("low stock", e, low)
	}
	returned := domain.InventoryMovementInput{ItemID: item.ID, Kind: "return", QuantityMilli: 1000, OriginalID: issued.ID, Restock: true, Reason: "Inspected and reusable"}
	back, e := inv.Move(ctx, a, returned, "inventory-return-0001")
	if e != nil || back.RecipientID != actors[1].ID || back.DeltaMilli != 1000 {
		t.Fatal("return", e, back)
	}
	retry, e = inv.Move(ctx, a, returned, "inventory-return-0001")
	if e != nil || retry.ID != back.ID {
		t.Fatal("return retry", e)
	}
	returned.QuantityMilli = 1001
	if _, e = inv.Move(ctx, a, returned, "inventory-return-0002"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("over-return", e)
	}
	returned.QuantityMilli = 1000
	returned.Restock = false
	returned.Reason = "Damaged return; not reusable"
	if _, e = inv.Move(ctx, a, returned, "inventory-return-0003"); e != nil {
		t.Fatal(e)
	}
	all, e := inv.Items(ctx, a, "Synthetic", false, 1)
	if e != nil || len(all) != 1 || all[0].BalanceMilli != 1500 {
		t.Fatal("non-restocked return changed availability", e, all)
	}
	edit := all[0].InventoryItemInput
	edit.Unit = "box"
	if _, e = inv.SaveItem(ctx, a, item.ID, edit); !errors.Is(e, domain.ErrStale) {
		t.Fatal("historical stock unit changed", e)
	}
	edit.Unit = "kg"
	edit.Active = false
	if _, e = inv.SaveItem(ctx, a, item.ID, edit); !errors.Is(e, domain.ErrStale) {
		t.Fatal("nonempty stock archived", e)
	}
	edit.Active = true
	edit.Description = "Updated description"
	saved, e := inv.SaveItem(ctx, a, item.ID, edit)
	if e != nil || saved.Description != edit.Description {
		t.Fatal("item edit", e)
	}
	if _, e = inv.SaveItem(ctx, a, item.ID, edit); !errors.Is(e, domain.ErrStale) {
		t.Fatal("stale edit accepted", e)
	}
	if _, e = db.Exec(ctx, `UPDATE inventory_item SET balance_milli=9999 WHERE id=$1`, item.ID); e == nil {
		t.Fatal("unreconciled inventory balance")
	}
	for _, query := range []string{`UPDATE inventory_movement SET reason='changed' WHERE id=$1`, `DELETE FROM inventory_movement WHERE id=$1`} {
		if _, e = db.Exec(ctx, query, receipt.ID); e == nil {
			t.Fatal("movement changed")
		}
	}
	for _, actor := range actors[1:] {
		if _, e = inv.Move(ctx, actor, issue, "unauthorized-inventory"); !errors.Is(e, domain.ErrForbidden) {
			t.Fatal("unauthorized inventory movement", actor.Role, e)
		}
	}
	if _, e = inv.Move(ctx, a, domain.InventoryMovementInput{ItemID: item.ID, Kind: "writeoff", QuantityMilli: 1500, Reason: "Synthetic disposal"}, "inventory-writeoff-01"); e != nil {
		t.Fatal(e)
	}
	all, e = inv.Items(ctx, a, "Synthetic", false, 1)
	if e != nil {
		t.Fatal(e)
	}
	edit = all[0].InventoryItemInput
	edit.Active = false
	if _, e = inv.SaveItem(ctx, a, item.ID, edit); e != nil {
		t.Fatal("empty stock archive", e)
	}
	if _, e = inv.Move(ctx, a, receive, "archived-inventory-01"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("archived stock received", e)
	}
	history, e := inv.Movements(ctx, a, item.ID, 1)
	if e != nil || len(history) != 5 {
		t.Fatal("history count", e, len(history))
	}
	t.Run("admin movement register and parent filter", func(t *testing.T) {
		second, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: category.ID, Name: "Second supply", Unit: "box", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{ItemID: second.ID, Kind: "receive", QuantityMilli: 1000, Reference: "REGISTER-1", Reason: "Register fixture"}, "inventory-register-01")
		if err != nil {
			t.Fatal(err)
		}
		all, err := inv.Movements(ctx, a, "", 1)
		if err != nil || len(all) != 6 {
			t.Fatal("aggregate register", len(all), err)
		}
		filtered, err := inv.Movements(ctx, a, second.ID, 1)
		if err != nil || len(filtered) != 1 || filtered[0].ItemID != second.ID {
			t.Fatal("parent scope", filtered, err)
		}
		page, err := inv.Movements(ctx, a, "", 2)
		if err != nil || len(page) != 0 {
			t.Fatal("pagination", page, err)
		}
		if _, err = inv.Movements(ctx, a, "invalid", 1); !errors.Is(err, domain.ErrValidation) {
			t.Fatal("invalid parent", err)
		}
		for _, role := range []string{"doctor", "nurse", "patient", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
			actor := domain.Actor{ID: actors[1].ID, Role: role}
			if _, err = inv.Movements(ctx, actor, "", 1); !errors.Is(err, domain.ErrForbidden) {
				t.Fatal("aggregate exposed", role, err)
			}
			if _, err = store.InventoryMovements(ctx, actor, second.ID, 1); !errors.Is(err, domain.ErrForbidden) {
				t.Fatal("direct store exposed", role, err)
			}
		}
	})
	t.Run("deletion protection and permissions", func(t *testing.T) {
		// Category in use cannot be deleted
		if err := inv.DeleteCategory(ctx, a, category.ID); !errors.Is(err, domain.ErrInUse) {
			t.Fatalf("expected ErrInUse deleting referenced category, got %v", err)
		}
		// Item in use cannot be deleted
		if err := inv.DeleteItem(ctx, a, item.ID); !errors.Is(err, domain.ErrInUse) {
			t.Fatalf("expected ErrInUse deleting referenced item, got %v", err)
		}
		// Non-admin cannot delete
		for _, actor := range actors[1:] {
			if err := inv.DeleteCategory(ctx, actor, category.ID); !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("expected ErrForbidden for %s deleting category, got %v", actor.Role, err)
			}
			if err := inv.DeleteItem(ctx, actor, item.ID); !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("expected ErrForbidden for %s deleting item, got %v", actor.Role, err)
			}
		}
		// Unreferenced category can be deleted
		tempCat, err := inv.SaveCategory(ctx, a, "", domain.InventoryCategoryInput{Name: "Disposable Category", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		// Unreferenced item can be deleted
		tempItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: tempCat.ID, Name: "Disposable Item", Unit: "pc", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		// Now delete item cleanly
		if err := inv.DeleteItem(ctx, a, tempItem.ID); err != nil {
			t.Fatalf("expected clean deletion of unreferenced item, got %v", err)
		}
		// Now delete category cleanly
		if err := inv.DeleteCategory(ctx, a, tempCat.ID); err != nil {
			t.Fatalf("expected clean deletion of unreferenced category, got %v", err)
		}
	})
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Inventory: inv, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		actor, method, path, body, origin string
		want                              int
	}{
		{"admin", "GET", "/v1/inventory/items", "", "", 200}, {"patient", "GET", "/v1/inventory/items", "", "", 403},
		{"admin", "GET", "/v1/inventory/movements", "", "", 200},
		{"doctor", "GET", "/v1/inventory/movements", "", "", 403},
		{"admin", "GET", "/v1/inventory/movements?itemId=invalid", "", "", 422},
		{"admin", "POST", "/v1/inventory/movements", `{"unknown":true}`, "http://hospital.test", 400},
		{"admin", "POST", "/v1/inventory/movements", `{}`, "http://evil.test", 403},
		{"admin", "DELETE", "/v1/inventory/categories/" + category.ID, "", "http://hospital.test", 409},
		{"admin", "DELETE", "/v1/inventory/items/" + item.ID, "", "http://hospital.test", 409},
		{"doctor", "DELETE", "/v1/inventory/categories/" + category.ID, "", "http://hospital.test", 403},
		{"doctor", "DELETE", "/v1/inventory/items/" + item.ID, "", "http://hospital.test", 403},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatal(tc.path, w.Code, w.Body.String())
		}
	}
}
