package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
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
	low, _, e := inv.Items(ctx, a, "", "", true, 1, 25)
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
	all, totalItems, e := inv.Items(ctx, a, "Synthetic", "", false, 1, 25)
	if e != nil || len(all) != 1 || totalItems != 1 || all[0].BalanceMilli != 1500 {
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
	all, _, e = inv.Items(ctx, a, "Synthetic", "", false, 1, 25)
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
	history, totalMovs, e := inv.Movements(ctx, a, item.ID, "", 1, 25, "", "")
	if e != nil || len(history) != 5 || totalMovs != 5 {
		t.Fatal("history count", e, len(history), totalMovs)
	}
	t.Run("admin movement register and parent filter", func(t *testing.T) {
		second, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: category.ID, Name: "Second supply", Unit: "box", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		// Create attachment for REGISTER-1
		_, err = db.Exec(ctx, `INSERT INTO secure_attachment(token, file_name, mime_type, file_size_bytes, storage_path, sha256_hash, uploader_id, is_public)
			VALUES('11112222333344445555666677778888', 'reg-invoice.pdf', 'application/pdf', 1024, '/storage/attachments/reg-invoice.pdf', 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', $1, false)
			ON CONFLICT (token) DO NOTHING`, a.ID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID:        second.ID,
			Kind:          "receive",
			QuantityMilli: 1000,
			Reference:     "REGISTER-1",
			Reason:        "Register fixture",
			AttachmentURL: "/v1/attachments/11112222333344445555666677778888/content",
		}, "inventory-register-01")
		if err != nil {
			t.Fatal(err)
		}
		all, totalAll, err := inv.Movements(ctx, a, "", "", 1, 25, "", "")
		if err != nil || len(all) != 6 || totalAll != 6 {
			t.Fatal("aggregate register", len(all), totalAll, err)
		}
		filtered, totalFiltered, err := inv.Movements(ctx, a, second.ID, "", 1, 25, "", "")
		if err != nil || len(filtered) != 1 || totalFiltered != 1 || filtered[0].ItemID != second.ID {
			t.Fatal("parent scope", filtered, err)
		}
		if filtered[0].AttachmentURL != "/v1/attachments/11112222333344445555666677778888/content" {
			t.Fatalf("expected attachment URL preserved, got %q", filtered[0].AttachmentURL)
		}
		if filtered[0].ItemName != "Second supply" {
			t.Fatalf("expected ItemName populated, got %q", filtered[0].ItemName)
		}
		searched, totalSearched, err := inv.Movements(ctx, a, "", "", 1, 25, "Register fixture", "")
		if err != nil || len(searched) != 1 || totalSearched != 1 || searched[0].Reference != "REGISTER-1" {
			t.Fatal("search movement filter", len(searched), totalSearched, err)
		}
		page, _, err := inv.Movements(ctx, a, "", "", 2, 25, "", "")
		if err != nil || len(page) != 0 {
			t.Fatal("pagination", page, err)
		}
		if _, _, err = inv.Movements(ctx, a, "invalid", "", 1, 25, "", ""); !errors.Is(err, domain.ErrValidation) {
			t.Fatal("invalid parent", err)
		}
		for _, role := range []string{"doctor", "nurse", "patient", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
			actor := domain.Actor{ID: actors[1].ID, Role: role}
			if _, _, err = inv.Movements(ctx, actor, "", "", 1, 25, "", ""); !errors.Is(err, domain.ErrForbidden) {
				t.Fatal("aggregate exposed", role, err)
			}
			if _, _, err = store.InventoryMovements(ctx, actor, second.ID, "", 1, 25, "", ""); !errors.Is(err, domain.ErrForbidden) {
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

	t.Run("R1: receipt attachment lifecycle and reference protection", func(t *testing.T) {
		genToken := func() string {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				t.Fatal(err)
			}
			return hex.EncodeToString(b)
		}
		genHash := func() string {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				t.Fatal(err)
			}
			return hex.EncodeToString(b)
		}

		attItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: category.ID, Name: "Attachment item", Unit: "vial", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		// 1. Create a genuine non-clinical receipt attachment
		tokRec := genToken()
		pathRec := "/storage/" + tokRec + "-receipt.pdf"
		attReceipt, err := store.SaveAttachment(ctx, a, tokRec, domain.CreateSecureAttachmentInput{
			FileName: "receipt-real.pdf", MimeType: "application/pdf", FileSizeBytes: 2048, StoragePath: pathRec,
			Sha256Hash: genHash(), IsPublic: false,
		})
		if err != nil {
			t.Fatal(err)
		}
		// 2. Reject binding clinical attachment
		var patientUUID string
		err = db.QueryRow(ctx, `SELECT id FROM patient LIMIT 1`).Scan(&patientUUID)
		if err != nil {
			t.Fatalf("failed to query patient: %v", err)
		}
		tokClin := genToken()
		pathClin := "/storage/" + tokClin + "-patient-xray.png"
		attClinical, err := store.SaveAttachment(ctx, a, tokClin, domain.CreateSecureAttachmentInput{
			FileName: "patient-xray.png", MimeType: "image/png", FileSizeBytes: 2048, StoragePath: pathClin,
			Sha256Hash: genHash(), IsPublic: false, PatientID: &patientUUID,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: attItem.ID, Kind: "receive", QuantityMilli: 1000, Reference: "REC-CLIN-FAIL", Reason: "Clinical attach test",
			AttachmentURL: "/v1/attachments/" + attClinical.Token + "/content",
		}, "key-clinical-fail-001")
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation binding clinical attachment, got %v", err)
		}
		// 3. Reject binding nonexistent/retired attachment
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: attItem.ID, Kind: "receive", QuantityMilli: 1000, Reference: "REC-NONEXIST-FAIL", Reason: "Nonexist test",
			AttachmentURL: "/v1/attachments/00000000000000000000000000000000/content",
		}, "key-nonexist-fail")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound binding nonexistent attachment, got %v", err)
		}
		// 4. Bind valid attachment to receipt
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: attItem.ID, Kind: "receive", QuantityMilli: 1000, Reference: "REC-BIND-OK", Reason: "Valid receipt attachment",
			AttachmentURL: "/v1/attachments/" + attReceipt.Token + "/content",
		}, "key-receipt-bind-ok-01")
		if err != nil {
			t.Fatalf("valid receipt bind failed: %v", err)
		}
		// 5. Direct retirement conflict: DeleteAttachment must return ErrInUse (409)
		_, err = store.DeleteAttachment(ctx, a, attReceipt.Token)
		if !errors.Is(err, domain.ErrInUse) {
			t.Fatalf("expected ErrInUse for receipt-referenced attachment, got %v", err)
		}
		// 6. F3: Aged private upload lifecycle, reference preservation, and clinical protection
		// Aged unreferenced private upload (created 48h ago)
		tokAged := genToken()
		pathAged := "/storage/" + tokAged + "-abandoned.pdf"
		attAgedUnref, err := store.SaveAttachment(ctx, a, tokAged, domain.CreateSecureAttachmentInput{
			FileName: "abandoned-receipt.pdf", MimeType: "application/pdf", FileSizeBytes: 1024,
			StoragePath: pathAged, Sha256Hash: genHash(),
			IsPublic: false,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ctx, `UPDATE secure_attachment SET created_at = clock_timestamp() - interval '48 hours' WHERE token = $1`, attAgedUnref.Token); err != nil {
			t.Fatal(err)
		}

		// Recent pending private upload (created recently)
		tokRecent := genToken()
		pathRecent := "/storage/" + tokRecent + "-recent.pdf"
		attRecent, err := store.SaveAttachment(ctx, a, tokRecent, domain.CreateSecureAttachmentInput{
			FileName: "recent-receipt.pdf", MimeType: "application/pdf", FileSizeBytes: 1024,
			StoragePath: pathRecent, Sha256Hash: genHash(),
			IsPublic: false,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Aged referenced private receipt (created 48h ago, referenced by receipt)
		if _, err := db.Exec(ctx, `UPDATE secure_attachment SET created_at = clock_timestamp() - interval '48 hours' WHERE token = $1`, attReceipt.Token); err != nil {
			t.Fatal(err)
		}

		// Aged clinical private attachment (created 48h ago, patient attached)
		var testPatID string
		err = db.QueryRow(ctx, `SELECT id FROM patient LIMIT 1`).Scan(&testPatID)
		if err != nil {
			_ = db.QueryRow(ctx, `INSERT INTO patient(given_name, family_name, gender, birth_date) VALUES('Test', 'Patient', 'male', '1990-01-01') RETURNING id`).Scan(&testPatID)
		}
		tokClinRec := genToken()
		pathClinRec := "/storage/" + tokClinRec + "-clinical-rec.pdf"
		_, err = db.Exec(ctx, `
			INSERT INTO secure_attachment(token, file_name, mime_type, file_size_bytes, storage_path, sha256_hash, uploader_id, is_public, patient_id, created_at)
			VALUES($1, 'clinical-record.pdf', 'application/pdf', 1024, $2, $3, $4, false, $5::uuid, clock_timestamp() - interval '48 hours')
		`, tokClinRec, pathClinRec, genHash(), a.ID, testPatID)
		if err != nil {
			t.Fatal(err)
		}

		// Execute cleanup with 24 hours threshold
		cleaned, err := store.CleanupAbandonedAttachments(ctx, a, 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		// 1. Aged unreferenced private upload MUST be removed
		var agedUnrefExists, recentExists, agedRefExists, clinExists bool
		_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token=$1)`, attAgedUnref.Token).Scan(&agedUnrefExists)
		if agedUnrefExists {
			t.Fatalf("expected aged unreferenced private upload %s to be deleted by cleanup", attAgedUnref.Token)
		}
		foundRemoved := false
		for _, p := range cleaned {
			if p == attAgedUnref.StoragePath {
				foundRemoved = true
				break
			}
		}
		if !foundRemoved {
			t.Fatalf("expected storage path %s in removed paths, got %v", attAgedUnref.StoragePath, cleaned)
		}

		// 2. Recent pending private upload MUST be preserved
		_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token=$1)`, attRecent.Token).Scan(&recentExists)
		if !recentExists {
			t.Fatalf("expected recent pending private upload %s to be retained", attRecent.Token)
		}

		// 3. Aged referenced receipt attachment MUST be preserved
		_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token=$1)`, attReceipt.Token).Scan(&agedRefExists)
		if !agedRefExists {
			t.Fatalf("expected aged referenced receipt attachment %s to be retained", attReceipt.Token)
		}

		// 4. Clinical attachment MUST be preserved
		_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token=$1)`, tokClinRec).Scan(&clinExists)
		if !clinExists {
			t.Fatalf("expected clinical attachment %s to be retained", tokClinRec)
		}

		// 7. Shared settings reference: if displaced from general settings, still preserved by receipt
		tokShared := genToken()
		pathShared := "/storage/" + tokShared + "-seal.png"
		attShared, err := store.SaveAttachment(ctx, a, tokShared, domain.CreateSecureAttachmentInput{
			FileName: "hospital-seal.png", MimeType: "image/png", FileSizeBytes: 2048, StoragePath: pathShared,
			Sha256Hash: genHash(), IsPublic: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: attItem.ID, Kind: "receive", QuantityMilli: 1000, Reference: "REC-SHARED-OK", Reason: "Shared seal receipt",
			AttachmentURL: "/v1/attachments/" + attShared.Token + "/content",
		}, "key-shared-receipt-01")
		if err != nil {
			t.Fatal(err)
		}
		_, err = store.UpdateGeneralSetting(ctx, a, domain.GeneralSettingInput{Key: "app_logo", Value: "/v1/attachments/" + attShared.Token + "/content"})
		if err != nil {
			t.Fatal(err)
		}
		// Displace it from general settings
		_, err = store.UpdateGeneralSetting(ctx, a, domain.GeneralSettingInput{Key: "app_logo", Value: "https://example.com/logo.png"})
		if err != nil {
			t.Fatal(err)
		}
		// Must still exist in secure_attachment
		var stillExists bool
		_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token=$1)`, attShared.Token).Scan(&stillExists)
		if !stillExists {
			t.Fatal("shared attachment was deleted when displaced from settings despite receipt reference")
		}
	})

	t.Run("R2: durable linked void receipt and consumption protection", func(t *testing.T) {
		voidItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: category.ID, Name: "Void test item", Unit: "ampoule", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		// 1. Receive Receipt A (100 units = 100000 milli) and Receipt B (100 units)
		recA, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "receive", QuantityMilli: 100000, Reference: "REC-A", Supplier: "Vendor A", Reason: "Batch A",
		}, "key-receipt-a-00001")
		if err != nil {
			t.Fatal(err)
		}
		recB, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "receive", QuantityMilli: 100000, Reference: "REC-B", Supplier: "Vendor B", Reason: "Batch B",
		}, "key-receipt-b-00001")
		if err != nil {
			t.Fatal(err)
		}
		// Balance is now 200 units (200000 milli)

		// 2. Void Receipt A successfully
		voidA, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "void_receipt", OriginalID: recA.ID, QuantityMilli: 100000, Reason: "Vendor A cancelled batch",
		}, "key-void-receipt-a-1")
		if err != nil {
			t.Fatalf("expected void of Receipt A to succeed, got %v", err)
		}
		if voidA.DeltaMilli != -100000 || voidA.OriginalID != recA.ID {
			t.Fatalf("invalid void movement: %+v", voidA)
		}
		// Verify balance dropped to 100 units
		var curBal int64
		_ = db.QueryRow(ctx, `SELECT balance_milli FROM inventory_item WHERE id=$1`, voidItem.ID).Scan(&curBal)
		if curBal != 100000 {
			t.Fatalf("expected balance 100000 after void A, got %d", curBal)
		}

		// Verify Receipt A is flagged as isVoided in register
		movsA, _, err := inv.Movements(ctx, a, voidItem.ID, "receive", 1, 10, "", "")
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range movsA {
			if m.ID == recA.ID && !m.IsVoided {
				t.Fatalf("expected Receipt A to have IsVoided=true in Movements query")
			}
			if m.ID == recB.ID && m.IsVoided {
				t.Fatalf("expected Receipt B to have IsVoided=false in Movements query")
			}
		}

		// 3. Repeated void of Receipt A with a DIFFERENT request key must fail with ErrConflict (409)
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "void_receipt", OriginalID: recA.ID, QuantityMilli: 100000, Reason: "Vendor A cancelled batch again",
		}, "key-void-receipt-a-2")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict repeating void of Receipt A with fresh key, got %v", err)
		}

		// 4. Intervening consumption protection:
		// Issue 80 units against Receipt B. Remaining balance = 20 units.
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "issue", QuantityMilli: 80000, RecipientID: actors[1].ID, Reason: "Ward consumption of batch B",
		}, "key-issue-receipt-b-1")
		if err != nil {
			t.Fatal(err)
		}
		// Attempt to void Receipt B (100 units): must be rejected because balance (20) < 100 units
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "void_receipt", OriginalID: recB.ID, QuantityMilli: 100000, Reason: "Void B while consumed",
		}, "key-void-b-fail-0001")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict voiding consumed receipt B with insufficient stock, got %v", err)
		}

		// 5. Subsequent replenishment does NOT un-consume B:
		// Replenish 100 units via Receipt C. Balance is now 120 units (>= 100 units).
		recC, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "receive", QuantityMilli: 100000, Reference: "REC-C", Supplier: "Vendor C", Reason: "Batch C replenishment",
		}, "key-receipt-c-00001")
		if err != nil {
			t.Fatal(err)
		}
		// Attempting to void Receipt B must STILL fail with ErrConflict because B was consumed (min running balance was 20 < 100)
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "void_receipt", OriginalID: recB.ID, QuantityMilli: 100000, Reason: "Void B after replenishment",
		}, "key-void-b-fail-0002")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict voiding receipt B whose stock was consumed even after replenishment, got %v", err)
		}

		// But Receipt C (unconsumed) CAN be voided:
		voidC, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: voidItem.ID, Kind: "void_receipt", OriginalID: recC.ID, QuantityMilli: 100000, Reason: "Void unconsumed receipt C",
		}, "key-void-c-ok-000001")
		if err != nil {
			t.Fatalf("expected void of unconsumed Receipt C to succeed, got %v", err)
		}
		if voidC.DeltaMilli != -100000 {
			t.Fatalf("invalid delta for void C: %d", voidC.DeltaMilli)
		}
	})

	t.Run("F2: transaction-start vs lock-acquisition ordering and timestamp ties", func(t *testing.T) {
		orderItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{
			CategoryID: category.ID, Name: "Order witness item", Unit: "vial", Active: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// 1. Receipt A adds 100 units (100000 milli).
		recA, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: orderItem.ID, Kind: "receive", QuantityMilli: 100000, Reference: "REC-A-ORDER", Reason: "Batch A witness",
		}, "key-order-rec-a-0001")
		if err != nil {
			t.Fatal(err)
		}

		// 2. Controlled concurrent witness:
		// Transaction B begins at t1, but does NOT obtain item lock yet.
		// Transaction C begins at t2 (t2 > t1), acquires item lock, and issues 50 units (reducing balance to 50 units).
		// Then Transaction B acquires item lock, and replenishes 100 units (increasing balance to 150 units).
		txB, err := db.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer txB.Rollback(ctx)

		time.Sleep(10 * time.Millisecond)

		// Execute Issue C in a separate transaction (which acquires item lock first and commits)
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: orderItem.ID, Kind: "issue", QuantityMilli: 50000, RecipientID: actors[1].ID, Reason: "Issue C intermediate consumption",
		}, "key-order-issue-c")
		if err != nil {
			t.Fatal(err)
		}

		// Now txB completes replenishment of 100 units under its earlier started transaction
		var curBalB, nextSeqB int64
		_ = txB.QueryRow(ctx, `SELECT balance_milli FROM inventory_item WHERE id=$1 FOR UPDATE`, orderItem.ID).Scan(&curBalB)
		_ = txB.QueryRow(ctx, `SELECT COALESCE(MAX(ledger_seq), 0) + 1 FROM inventory_movement WHERE item_id=$1`, orderItem.ID).Scan(&nextSeqB)
		var bID string
		err = txB.QueryRow(ctx, `
			INSERT INTO inventory_movement(item_id, kind, quantity_milli, delta_milli, supplier, store_name, reference, cost_minor, restock, reason, actor_id, request_key, request_hash, attachment_url, issued_date, return_due_date, issued_by, department, ledger_seq, balance_after_milli)
			VALUES($1, 'receive', 100000, 100000, '', '', 'REC-B-REPLENISH', 0, false, 'Replenishment B', $2, 'key-order-rec-b-0001', 'hash-b', '', '', '', '', '', $3, $4)
			RETURNING id
		`, orderItem.ID, a.ID, nextSeqB, curBalB+100000).Scan(&bID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = txB.Exec(ctx, `UPDATE inventory_item SET balance_milli = balance_milli + 100000, version = version + 1 WHERE id = $1`, orderItem.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := txB.Commit(ctx); err != nil {
			t.Fatal(err)
		}

		// Current item balance is now 150000 milli (150 units >= 100 units).
		// However, because Issue C intervened, the running balance dropped to 50000 milli.
		// Attempting to void Receipt A (100 units) MUST FAIL with ErrConflict!
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: orderItem.ID, Kind: "void_receipt", OriginalID: recA.ID, QuantityMilli: 100000, Reason: "Void A despite intervening consumption",
		}, "key-order-void-a-attempt")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict voiding Receipt A after intervening consumption (witness test), got %v", err)
		}

		// 3. Timestamp ties regression:
		tieItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{
			CategoryID: category.ID, Name: "Timestamp ties item", Unit: "ampoule", Active: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		recTie, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: tieItem.ID, Kind: "receive", QuantityMilli: 100000, Reference: "REC-TIE-1", Reason: "Tie test receipt",
		}, "key-tie-rec-00000001")
		if err != nil {
			t.Fatal(err)
		}
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: tieItem.ID, Kind: "issue", QuantityMilli: 60000, RecipientID: actors[1].ID, Reason: "Tie test issue",
		}, "key-tie-issue-0000001")
		if err != nil {
			t.Fatal(err)
		}
		// Force identical created_at on all movements of tieItem
		_, _ = db.Exec(ctx, `ALTER TABLE inventory_movement DISABLE TRIGGER immutable_inventory_movement`)
		_, err = db.Exec(ctx, `UPDATE inventory_movement SET created_at = '2026-10-10 12:00:00+00' WHERE item_id = $1`, tieItem.ID)
		_, _ = db.Exec(ctx, `ALTER TABLE inventory_movement ENABLE TRIGGER immutable_inventory_movement`)
		if err != nil {
			t.Fatal(err)
		}
		// Attempting to void recTie (100 units) must fail because 60 units were consumed (balance is 40).
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: tieItem.ID, Kind: "void_receipt", OriginalID: recTie.ID, QuantityMilli: 100000, Reason: "Void tie receipt",
		}, "key-tie-void-attempt")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict voiding receipt with timestamp ties, got %v", err)
		}
	})

	t.Run("F1: database upgrade schema validation and void operation", func(t *testing.T) {
		var hasVoidKind, hasUniqueIndex bool
		err := db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM pg_constraint
				WHERE conrelid = 'inventory_movement'::regclass
				  AND conname = 'inventory_movement_kind_check'
				  AND pg_get_constraintdef(oid) LIKE '%void_receipt%'
			)
		`).Scan(&hasVoidKind)
		if err != nil || !hasVoidKind {
			t.Fatalf("expected inventory_movement_kind_check to allow void_receipt, err=%v", err)
		}

		err = db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM pg_indexes
				WHERE tablename = 'inventory_movement'
				  AND indexname = 'inventory_movement_void_receipt'
			)
		`).Scan(&hasUniqueIndex)
		if err != nil || !hasUniqueIndex {
			t.Fatalf("expected inventory_movement_void_receipt index to exist, err=%v", err)
		}

		upgItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{
			CategoryID: category.ID, Name: "Upgrade test item", Unit: "pack", Active: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		upgRec, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: upgItem.ID, Kind: "receive", QuantityMilli: 25000, Reference: "REC-UPG-01", Reason: "Upgrade receipt test",
		}, "key-upg-rec-00000001")
		if err != nil {
			t.Fatal(err)
		}
		if upgRec.LedgerSeq <= 0 {
			t.Fatalf("expected positive ledger_seq on upgRec, got %d", upgRec.LedgerSeq)
		}
		upgVoid, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: upgItem.ID, Kind: "void_receipt", OriginalID: upgRec.ID, QuantityMilli: 25000, Reason: "Upgrade void test",
		}, "key-upg-void-00000001")
		if err != nil {
			t.Fatalf("expected void on upgraded schema to succeed, got %v", err)
		}
		if upgVoid.DeltaMilli != -25000 || upgVoid.LedgerSeq <= upgRec.LedgerSeq {
			t.Fatalf("unexpected upgVoid result: %+v", upgVoid)
		}

		// Populated prior-054 upgrade regression:
		// Create a separate isolated schema to verify upgrade transition with historical data
		upgRand := make([]byte, 8)
		if _, err := rand.Read(upgRand); err != nil {
			t.Fatal(err)
		}
		upgSchema := fmt.Sprintf("hms_upg_%x", upgRand)
		quotedSchema := pgx.Identifier{upgSchema}.Sanitize()
		if _, err := db.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
			t.Fatalf("failed to create upgrade test schema: %v", err)
		}
		defer func() {
			_, _ = db.Exec(ctx, "DROP SCHEMA "+quotedSchema+" CASCADE")
		}()

		cfg, err := pgxpool.ParseConfig(db.Config().ConnString())
		if err != nil {
			t.Fatal(err)
		}
		cfg.ConnConfig.RuntimeParams["search_path"] = upgSchema
		upgPool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer upgPool.Close()

		// Run migrations 001 through 054 (strictly excluding 055 and later)
		migrationFiles, err := filepath.Glob("../../../../../db/migrations/*.sql")
		if err != nil || len(migrationFiles) < 3 {
			t.Fatal("migrations unavailable")
		}
		for _, file := range migrationFiles {
			if filepath.Base(file) >= "055" {
				continue
			}
			sql, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := upgPool.Exec(ctx, string(sql)); err != nil {
				t.Fatalf("failed applying pre-055 migration %s: %v", file, err)
			}
		}

		// Seed actors and permissions in the pre-055 schema
		for _, act := range actors {
			if _, err := upgPool.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test') ON CONFLICT DO NOTHING`, act.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := upgPool.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,$2) ON CONFLICT DO NOTHING`, act.ID, act.Role); err != nil {
				t.Fatal(err)
			}
		}

		upgStore := Store{DB: upgPool}
		upgInv := application.Inventory{Store: upgStore}

		// Create category and item under 054 schema
		preCat, err := upgInv.SaveCategory(ctx, a, "", domain.InventoryCategoryInput{Name: "Pre-Upgrade Category", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		preItem, err := upgInv.SaveItem(ctx, a, "", domain.InventoryItemInput{
			CategoryID: preCat.ID, Name: "Pre-Upgrade Item", Unit: "vial", Active: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Insert valid ledger movements in actual order:
		// A (+100 = 100000 milli, balance=100000)
		// C (-50 = 50000 milli issue, balance=50000)
		// B (+100 = 100000 milli, balance=150000)
		// In pre-055 schema (migration 054), inventory_movement does not have ledger_seq or balance_after_milli yet.
		baseTime := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
		insertLegacyMovement := func(kind string, qtyMilli, deltaMilli int64, recipientID, ref, key string, createdAt time.Time) string {
			tx, err := upgPool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			var mID string
			err = tx.QueryRow(ctx, `
				INSERT INTO inventory_movement(
					item_id, kind, quantity_milli, delta_milli, recipient_id, supplier,
					store_name, reference, cost_minor, restock, reason, actor_id,
					request_key, request_hash, attachment_url, issued_date, return_due_date,
					issued_by, department, created_at
				) VALUES (
					$1, $2, $3, $4, NULLIF($5, ''), '',
					'Main store', $6, 0, false, 'Legacy test', $7,
					$8, 'hash', '', '', '',
					'', '', $9
				) RETURNING id
			`, preItem.ID, kind, qtyMilli, deltaMilli, recipientID, ref, a.ID, key, createdAt).Scan(&mID)
			if err != nil {
				t.Fatalf("failed inserting legacy movement: %v", err)
			}
			_, err = tx.Exec(ctx, `UPDATE inventory_item SET balance_milli = balance_milli + $2, version = version + 1 WHERE id = $1`, preItem.ID, deltaMilli)
			if err != nil {
				t.Fatalf("failed updating item balance for legacy movement: %v", err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatalf("failed committing legacy movement: %v", err)
			}
			return mID
		}

		recLegacyAID := insertLegacyMovement("receive", 100000, 100000, "", "REC-LEGACY-A", "key-leg-rec-a-0000001", baseTime)
		_ = insertLegacyMovement("issue", 50000, -50000, actors[1].ID, "", "key-leg-iss-c-0000001", baseTime.Add(20*time.Second))
		_ = insertLegacyMovement("receive", 100000, 100000, "", "REC-LEGACY-B", "key-leg-rec-b-0000001", baseTime.Add(10*time.Second))

		// Verify balance before upgrade is 150000 milli
		var balBefore int64
		_ = upgPool.QueryRow(ctx, `SELECT balance_milli FROM inventory_item WHERE id=$1`, preItem.ID).Scan(&balBefore)
		if balBefore != 150000 {
			t.Fatalf("expected balance 150000 before upgrade, got %d", balBefore)
		}

		// Apply Migration 055 to upgrade the schema
		sql055, err := os.ReadFile("../../../../../db/migrations/055_inventory_void_and_ledger_order.sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := upgPool.Exec(ctx, string(sql055)); err != nil {
			t.Fatalf("failed applying migration 055 on populated database: %v", err)
		}

		// Apply Migration 056 to run the legacy ledger cutover
		sql056, err := os.ReadFile("../../../../../db/migrations/056_inventory_legacy_ledger_cutover.sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := upgPool.Exec(ctx, string(sql056)); err != nil {
			t.Fatalf("failed applying migration 056 on populated database: %v", err)
		}

		// Verify existing balance is preserved
		var balPreserved int64
		_ = upgPool.QueryRow(ctx, `SELECT balance_milli FROM inventory_item WHERE id=$1`, preItem.ID).Scan(&balPreserved)
		if balPreserved != 150000 {
			t.Fatalf("expected balance 150000 preserved after upgrade, got %d", balPreserved)
		}

		// Verify legacy rows have ledger_seq = 0 and legacy_unverified = true
		var legSeqA int64
		var legUnverifiedA bool
		_ = upgPool.QueryRow(ctx, `SELECT ledger_seq, legacy_unverified FROM inventory_movement WHERE id=$1`, recLegacyAID).Scan(&legSeqA, &legUnverifiedA)
		if legSeqA != 0 || !legUnverifiedA {
			t.Fatalf("expected legacy row to have ledger_seq=0 and legacy_unverified=true, got seq=%d, unverified=%v", legSeqA, legUnverifiedA)
		}

		// 1. Attempting to void legacy Receipt A MUST BE REJECTED with ErrConflict (409)
		// Even though current item balance (150) >= receipt quantity (100), Receipt A was consumed,
		// and legacy unverified history cannot be trusted to reconstruct serialized order.
		_, err = upgInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: preItem.ID, Kind: "void_receipt", OriginalID: recLegacyAID, QuantityMilli: 100000, Reason: "Attempt void legacy receipt A",
		}, "key-upg-void-legacy-001")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict (409) voiding legacy receipt A, got %v", err)
		}

		// 2. Fresh post-upgrade receipt CAN be recorded and voided correctly
		recFresh, err := upgInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: preItem.ID, Kind: "receive", QuantityMilli: 40000, Reference: "REC-POST-UPG", Reason: "Post-upgrade fresh batch",
		}, "key-upg-rec-fresh-0001")
		if err != nil {
			t.Fatalf("failed to record post-upgrade receipt: %v", err)
		}
		if recFresh.LedgerSeq <= 0 {
			t.Fatalf("expected post-upgrade receipt to have positive ledger_seq, got %d", recFresh.LedgerSeq)
		}

		var freshUnverified bool
		_ = upgPool.QueryRow(ctx, `SELECT legacy_unverified FROM inventory_movement WHERE id=$1`, recFresh.ID).Scan(&freshUnverified)
		if freshUnverified {
			t.Fatalf("expected post-upgrade receipt to have legacy_unverified=false")
		}

		// Voiding fresh unconsumed post-upgrade receipt succeeds:
		voidFresh, err := upgInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: preItem.ID, Kind: "void_receipt", OriginalID: recFresh.ID, QuantityMilli: 40000, Reason: "Void fresh post-upgrade receipt",
		}, "key-upg-void-fresh-0001")
		if err != nil {
			t.Fatalf("expected void of fresh post-upgrade receipt to succeed, got %v", err)
		}
		if voidFresh.DeltaMilli != -40000 || voidFresh.LedgerSeq <= recFresh.LedgerSeq {
			t.Fatalf("invalid voidFresh result: %+v", voidFresh)
		}

		// =========================================================================
		// Populated already-applied original 055 upgrade regression (Finding H1):
		// Verifies that databases that previously ran migration 055 from commit 8cc47e7
		// (which backfilled flawed sequence numbers and running balances derived from timestamps)
		// are safely repaired by migration 056:
		// - Migration runner skips 055 (because it is recorded in schema_migration)
		// - Migration 056 runs, establishes cutover provenance in inventory_ledger_cutover,
		//   resets ledger_seq and balance_after_milli to 0, and marks legacy_unverified = true
		// - Existing stock totals, movement rows, and audit history are strictly preserved
		// - Legacy receipts are rejected for automated voiding (domain.ErrConflict / 409)
		// - Fresh post-repair receipts record positive sequence numbers and void cleanly
		// =========================================================================
		appliedRand := make([]byte, 8)
		if _, err := rand.Read(appliedRand); err != nil {
			t.Fatal(err)
		}
		appliedSchema := fmt.Sprintf("hms_upg_applied_%x", appliedRand)
		quotedAppliedSchema := pgx.Identifier{appliedSchema}.Sanitize()
		if _, err := db.Exec(ctx, "CREATE SCHEMA "+quotedAppliedSchema); err != nil {
			t.Fatalf("failed to create applied-055 upgrade test schema: %v", err)
		}
		defer func() {
			_, _ = db.Exec(ctx, "DROP SCHEMA "+quotedAppliedSchema+" CASCADE")
		}()

		cfgApp, err := pgxpool.ParseConfig(db.Config().ConnString())
		if err != nil {
			t.Fatal(err)
		}
		cfgApp.ConnConfig.RuntimeParams["search_path"] = appliedSchema
		appPool, err := pgxpool.NewWithConfig(ctx, cfgApp)
		if err != nil {
			t.Fatal(err)
		}
		defer appPool.Close()

		// Create schema_migration tracking table
		if _, err := appPool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migration(name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
			t.Fatal(err)
		}

		// Run migrations 001 through 054 and record them in schema_migration
		for _, file := range migrationFiles {
			base := filepath.Base(file)
			if base >= "055" {
				continue
			}
			sql, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := appPool.Exec(ctx, string(sql)); err != nil {
				t.Fatalf("failed applying pre-055 migration %s: %v", file, err)
			}
			if _, err := appPool.Exec(ctx, `INSERT INTO schema_migration(name) VALUES($1)`, base); err != nil {
				t.Fatal(err)
			}
		}

		// Seed actors and permissions in appPool schema
		for _, act := range actors {
			if _, err := appPool.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test') ON CONFLICT DO NOTHING`, act.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := appPool.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,$2) ON CONFLICT DO NOTHING`, act.ID, act.Role); err != nil {
				t.Fatal(err)
			}
		}

		appStore := Store{DB: appPool}
		appInv := application.Inventory{Store: appStore}

		// Create category and item under 054 schema
		catApp, err := appInv.SaveCategory(ctx, a, "", domain.InventoryCategoryInput{Name: "Applied-055 Category", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		itemApp, err := appInv.SaveItem(ctx, a, "", domain.InventoryItemInput{
			CategoryID: catApp.ID, Name: "Applied-055 Item", Unit: "ampoule", Active: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Insert valid ledger movements in actual order:
		// A (+100 = 100000 milli, balance=100000) at baseTime
		// C (-50 = 50000 milli issue, balance=50000) at baseTime + 20s
		// B (+100 = 100000 milli, balance=150000) at baseTime + 10s
		insertAppMovement := func(kind string, qtyMilli, deltaMilli int64, recipientID, ref, key string, createdAt time.Time) string {
			tx, err := appPool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			var mID string
			err = tx.QueryRow(ctx, `
				INSERT INTO inventory_movement(
					item_id, kind, quantity_milli, delta_milli, recipient_id, supplier,
					store_name, reference, cost_minor, restock, reason, actor_id,
					request_key, request_hash, attachment_url, issued_date, return_due_date,
					issued_by, department, created_at
				) VALUES (
					$1, $2, $3, $4, NULLIF($5, ''), '',
					'Main store', $6, 0, false, 'Legacy test', $7,
					$8, 'hash', '', '', '',
					'', '', $9
				) RETURNING id
			`, itemApp.ID, kind, qtyMilli, deltaMilli, recipientID, ref, a.ID, key, createdAt).Scan(&mID)
			if err != nil {
				t.Fatalf("failed inserting legacy movement: %v", err)
			}
			_, err = tx.Exec(ctx, `UPDATE inventory_item SET balance_milli = balance_milli + $2, version = version + 1 WHERE id = $1`, itemApp.ID, deltaMilli)
			if err != nil {
				t.Fatalf("failed updating item balance: %v", err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatalf("failed committing legacy movement: %v", err)
			}
			return mID
		}

		recAID := insertAppMovement("receive", 100000, 100000, "", "REC-APPLIED-A", "key-app-rec-a-0000001", baseTime)
		_ = insertAppMovement("issue", 50000, -50000, actors[1].ID, "", "key-app-iss-c-0000001", baseTime.Add(20*time.Second))
		recBID := insertAppMovement("receive", 100000, 100000, "", "REC-APPLIED-B", "key-app-rec-b-0000001", baseTime.Add(10*time.Second))

		// Apply the ORIGINAL migration 055 from commit 8cc47e7 (with flawed timestamp-based backfill)
		// and record 055_inventory_void_and_ledger_order.sql in schema_migration:
		original055SQL := `
ALTER TABLE inventory_movement DROP CONSTRAINT IF EXISTS inventory_movement_kind_check;
ALTER TABLE inventory_movement ADD CONSTRAINT inventory_movement_kind_check
  CHECK (kind IN ('receive', 'issue', 'return', 'writeoff', 'void_receipt'));

ALTER TABLE inventory_movement DROP CONSTRAINT IF EXISTS inventory_movement_check;
ALTER TABLE inventory_movement ADD CONSTRAINT inventory_movement_check
  CHECK (
    (kind = 'receive' AND delta_milli = quantity_milli AND length(reference) > 0 AND recipient_id IS NULL AND original_id IS NULL) OR
    (kind = 'issue' AND delta_milli = -quantity_milli AND recipient_id IS NOT NULL AND original_id IS NULL) OR
    (kind = 'return' AND original_id IS NOT NULL AND recipient_id IS NOT NULL AND delta_milli = CASE WHEN restock THEN quantity_milli ELSE 0 END) OR
    (kind = 'writeoff' AND delta_milli = -quantity_milli AND recipient_id IS NULL AND original_id IS NULL) OR
    (kind = 'void_receipt' AND delta_milli = -quantity_milli AND recipient_id IS NULL AND original_id IS NOT NULL)
  );

CREATE UNIQUE INDEX IF NOT EXISTS inventory_movement_void_receipt
  ON inventory_movement(original_id)
  WHERE kind = 'void_receipt';

ALTER TABLE inventory_movement
  ADD COLUMN IF NOT EXISTS ledger_seq bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS balance_after_milli bigint NOT NULL DEFAULT 0;

ALTER TABLE inventory_movement DISABLE TRIGGER immutable_inventory_movement;

WITH ordered AS (
  SELECT id,
         row_number() OVER (
           PARTITION BY item_id
           ORDER BY created_at ASC,
                    CASE WHEN delta_milli > 0 THEN 0 ELSE 1 END,
                    id ASC
         ) AS seq,
         sum(delta_milli) OVER (
           PARTITION BY item_id
           ORDER BY created_at ASC,
                    CASE WHEN delta_milli > 0 THEN 0 ELSE 1 END,
                    id ASC
         ) AS bal
  FROM inventory_movement
)
UPDATE inventory_movement m
SET ledger_seq = o.seq,
    balance_after_milli = o.bal
FROM ordered o
WHERE m.id = o.id AND (m.ledger_seq = 0 OR m.balance_after_milli = 0);

ALTER TABLE inventory_movement ENABLE TRIGGER immutable_inventory_movement;

CREATE INDEX IF NOT EXISTS inventory_movement_item_ledger_seq
  ON inventory_movement(item_id, ledger_seq);
`
		if _, err := appPool.Exec(ctx, original055SQL); err != nil {
			t.Fatalf("failed applying original migration 055: %v", err)
		}
		if _, err := appPool.Exec(ctx, `INSERT INTO schema_migration(name) VALUES('055_inventory_void_and_ledger_order.sql')`); err != nil {
			t.Fatal(err)
		}

		// Witness the flawed state under original 055:
		// Receipt A was given ledger_seq=1, balance_after=100000
		// Receipt B was given ledger_seq=2, balance_after=200000 (because tB < tC!)
		// Issue C was given ledger_seq=3, balance_after=150000
		var origSeqA, origBalA int64
		_ = appPool.QueryRow(ctx, `SELECT ledger_seq, balance_after_milli FROM inventory_movement WHERE id=$1`, recAID).Scan(&origSeqA, &origBalA)
		if origSeqA != 1 || origBalA != 100000 {
			t.Fatalf("expected flawed 055 backfill to give seq=1, bal=100000, got seq=%d, bal=%d", origSeqA, origBalA)
		}

		// Now simulate runner migrate.ts:
		// It checks schema_migration: '055_inventory_void_and_ledger_order.sql' is ALREADY present,
		// so edited 055 file is SKIPPED!
		var count055 int
		_ = appPool.QueryRow(ctx, `SELECT count(*) FROM schema_migration WHERE name='055_inventory_void_and_ledger_order.sql'`).Scan(&count055)
		if count055 != 1 {
			t.Fatalf("expected 055 already recorded in schema_migration")
		}

		// The runner then finds 056_inventory_legacy_ledger_cutover.sql which is NOT applied yet:
		sql056App, err := os.ReadFile("../../../../../db/migrations/056_inventory_legacy_ledger_cutover.sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := appPool.Exec(ctx, string(sql056App)); err != nil {
			t.Fatalf("failed applying corrective migration 056: %v", err)
		}
		if _, err := appPool.Exec(ctx, `INSERT INTO schema_migration(name) VALUES('056_inventory_legacy_ledger_cutover.sql')`); err != nil {
			t.Fatal(err)
		}

		// Verify provenance metadata recorded in inventory_ledger_cutover
		var cutoverCount int
		_ = appPool.QueryRow(ctx, `SELECT count(*) FROM inventory_ledger_cutover`).Scan(&cutoverCount)
		if cutoverCount != 1 {
			t.Fatalf("expected 1 cutover record in inventory_ledger_cutover, got %d", cutoverCount)
		}

		// Verify existing balance is preserved
		var balAfterCutover int64
		_ = appPool.QueryRow(ctx, `SELECT balance_milli FROM inventory_item WHERE id=$1`, itemApp.ID).Scan(&balAfterCutover)
		if balAfterCutover != 150000 {
			t.Fatalf("expected balance 150000 preserved after migration 056, got %d", balAfterCutover)
		}

		// Verify movement records are preserved (row count = 3)
		var movCount int
		_ = appPool.QueryRow(ctx, `SELECT count(*) FROM inventory_movement WHERE item_id=$1`, itemApp.ID).Scan(&movCount)
		if movCount != 3 {
			t.Fatalf("expected 3 movements preserved, got %d", movCount)
		}

		// Verify fabricated positive sequences are reset and marked legacy_unverified
		var repSeqA, repBalA int64
		var repUnverifiedA bool
		_ = appPool.QueryRow(ctx, `SELECT ledger_seq, balance_after_milli, legacy_unverified FROM inventory_movement WHERE id=$1`, recAID).Scan(&repSeqA, &repBalA, &repUnverifiedA)
		if repSeqA != 0 || repBalA != 0 || !repUnverifiedA {
			t.Fatalf("expected repaired receipt A to have seq=0, bal=0, legacy_unverified=true, got seq=%d, bal=%d, unverified=%v", repSeqA, repBalA, repUnverifiedA)
		}

		var repSeqB, repBalB int64
		var repUnverifiedB bool
		_ = appPool.QueryRow(ctx, `SELECT ledger_seq, balance_after_milli, legacy_unverified FROM inventory_movement WHERE id=$1`, recBID).Scan(&repSeqB, &repBalB, &repUnverifiedB)
		if repSeqB != 0 || repBalB != 0 || !repUnverifiedB {
			t.Fatalf("expected repaired receipt B to have seq=0, bal=0, legacy_unverified=true, got seq=%d, bal=%d, unverified=%v", repSeqB, repBalB, repUnverifiedB)
		}

		// 1. Proves legacy receipt A cannot be voided (409 Conflict)
		_, err = appInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: itemApp.ID, Kind: "void_receipt", OriginalID: recAID, QuantityMilli: 100000, Reason: "Attempt void legacy receipt A after 056 repair",
		}, "key-app-void-leg-a-001")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict (409) attempting to void legacy receipt A, got %v", err)
		}

		// Legacy receipt B also cannot be voided (409 Conflict)
		_, err = appInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: itemApp.ID, Kind: "void_receipt", OriginalID: recBID, QuantityMilli: 100000, Reason: "Attempt void legacy receipt B after 056 repair",
		}, "key-app-void-leg-b-001")
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict (409) attempting to void legacy receipt B, got %v", err)
		}

		// 2. Proves fresh post-repair receipt gets positive sequence, legacy_unverified=false,
		// and voids cleanly when unconsumed
		recPostRepair, err := appInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: itemApp.ID, Kind: "receive", QuantityMilli: 25000, Reference: "REC-POST-REPAIR-01", Reason: "Post-repair trusted receipt",
		}, "key-app-rec-repair-0001")
		if err != nil {
			t.Fatalf("failed recording post-repair receipt: %v", err)
		}
		if recPostRepair.LedgerSeq <= 0 {
			t.Fatalf("expected post-repair receipt to have positive ledger_seq, got %d", recPostRepair.LedgerSeq)
		}

		var postRepairUnverified bool
		_ = appPool.QueryRow(ctx, `SELECT legacy_unverified FROM inventory_movement WHERE id=$1`, recPostRepair.ID).Scan(&postRepairUnverified)
		if postRepairUnverified {
			t.Fatalf("expected post-repair receipt to have legacy_unverified=false")
		}

		// Voiding fresh unconsumed receipt succeeds (201 / nil error)
		voidPostRepair, err := appInv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: itemApp.ID, Kind: "void_receipt", OriginalID: recPostRepair.ID, QuantityMilli: 25000, Reason: "Void trusted post-repair receipt",
		}, "key-app-void-repair-0001")
		if err != nil {
			t.Fatalf("expected voiding trusted post-repair receipt to succeed, got %v", err)
		}
		if voidPostRepair.DeltaMilli != -25000 || voidPostRepair.LedgerSeq <= recPostRepair.LedgerSeq {
			t.Fatalf("unexpected voidPostRepair: %+v", voidPostRepair)
		}
	})

	t.Run("R3: server-side issue return status filtering", func(t *testing.T) {
		retItem, err := inv.SaveItem(ctx, a, "", domain.InventoryItemInput{CategoryID: category.ID, Name: "Return filter item", Unit: "set", Active: true})
		if err != nil {
			t.Fatal(err)
		}
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: retItem.ID, Kind: "receive", QuantityMilli: 50000, Reference: "REC-RET-01", Reason: "Stock for return tests",
		}, "key-return-filter-rec-01")
		if err != nil {
			t.Fatal(err)
		}
		// Create Issue 1 (10 units) and Issue 2 (10 units)
		iss1, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: retItem.ID, Kind: "issue", QuantityMilli: 10000, RecipientID: actors[1].ID, Reason: "Issue 1 to be fully returned",
		}, "key-return-filter-iss-01")
		if err != nil {
			t.Fatal(err)
		}
		iss2, err := inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: retItem.ID, Kind: "issue", QuantityMilli: 10000, RecipientID: actors[1].ID, Reason: "Issue 2 to stay returnable",
		}, "key-return-filter-iss-02")
		if err != nil {
			t.Fatal(err)
		}
		// Return Issue 1 fully (10 units)
		_, err = inv.Move(ctx, a, domain.InventoryMovementInput{
			ItemID: retItem.ID, Kind: "return", OriginalID: iss1.ID, QuantityMilli: 10000, Restock: true, Reason: "Complete return of issue 1",
		}, "key-return-filter-ret-01")
		if err != nil {
			t.Fatal(err)
		}

		// Query returnStatus = "returnable": must return only Issue 2
		returnable, totalReturnable, err := inv.Movements(ctx, a, retItem.ID, "issue", 1, 10, "", "returnable")
		if err != nil {
			t.Fatal(err)
		}
		if totalReturnable != 1 || len(returnable) != 1 || returnable[0].ID != iss2.ID {
			t.Fatalf("expected 1 returnable issue (Issue 2), got total=%d, len=%d", totalReturnable, len(returnable))
		}

		// Query returnStatus = "returned": must return only Issue 1
		returned, totalReturned, err := inv.Movements(ctx, a, retItem.ID, "issue", 1, 10, "", "returned")
		if err != nil {
			t.Fatal(err)
		}
		if totalReturned != 1 || len(returned) != 1 || returned[0].ID != iss1.ID {
			t.Fatalf("expected 1 returned issue (Issue 1), got total=%d, len=%d", totalReturned, len(returned))
		}

		// Query returnStatus = "" (all): must return both
		allIssues, totalAllIssues, err := inv.Movements(ctx, a, retItem.ID, "issue", 1, 10, "", "")
		if err != nil {
			t.Fatal(err)
		}
		if totalAllIssues != 2 || len(allIssues) != 2 {
			t.Fatalf("expected 2 total issues, got total=%d, len=%d", totalAllIssues, len(allIssues))
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
