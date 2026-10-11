package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func setupSmartCardsTestDB(t *testing.T) (*pgxpool.Pool, Store, func()) {
	t.Helper()
	dsn := os.Getenv("HMS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set HMS_TEST_DATABASE_URL to run isolated-schema PostgreSQL tests")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		t.Fatal("test database must be loopback")
	}
	ctx := context.Background()
	adminDB, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	random := make([]byte, 12)
	if _, err = rand.Read(random); err != nil {
		adminDB.Close()
		t.Fatal(err)
	}
	schema := "hms_test_sc_" + hex.EncodeToString(random)
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = adminDB.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		adminDB.Close()
		t.Fatal(err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		adminDB.Close()
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		adminDB.Close()
		t.Fatal(err)
	}

	migrations, err := filepath.Glob("../../../../../db/migrations/*.sql")
	if err != nil || len(migrations) < 3 {
		db.Close()
		adminDB.Close()
		t.Fatal("migration files unavailable")
	}
	for _, file := range migrations {
		sql, e := os.ReadFile(file)
		if e != nil {
			db.Close()
			adminDB.Close()
			t.Fatal(e)
		}
		if _, e = db.Exec(ctx, string(sql)); e != nil {
			db.Close()
			adminDB.Close()
			t.Fatalf("failed migration %s: %v", file, e)
		}
	}

	// Seed actor users
	actors := []domain.Actor{
		{ID: "admin-user", Role: "admin", Name: "Admin User"},
		{ID: "reception-user", Role: "receptionist", Name: "Receptionist"},
		{ID: "doctor-user", Role: "doctor", Name: "Doctor"},
	}
	for _, a := range actors {
		if _, err = db.Exec(ctx, `INSERT INTO "user"(id, name, email) VALUES($1, $2, $1||'@example.test')`, a.ID, a.Name); err != nil {
			db.Close()
			adminDB.Close()
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO staff_access(user_id, role) VALUES($1, $2)`, a.ID, a.Role); err != nil {
			db.Close()
			adminDB.Close()
			t.Fatal(err)
		}
	}

	store := Store{DB: db}
	cleanup := func() {
		db.Close()
		_, _ = adminDB.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
		adminDB.Close()
	}

	return db, store, cleanup
}

func TestSmartCards_TemplateLifecycleAndIntegrity(t *testing.T) {
	_, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	app := application.PatientExtensionsService{Store: store}

	// 1. List pre-seeded templates (Standard and VIP from migration 057)
	templates, err := app.ListTemplates(ctx, admin)
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	if len(templates) < 2 {
		t.Fatalf("expected at least 2 seeded templates, got %d", len(templates))
	}

	// 2. Create new template with custom toggles
	emailFalse := false
	dobFalse := false
	t1, err := app.CreateTemplate(ctx, admin, domain.CreatePatientIDCardTemplateInput{
		Name:  "Pediatric-Card",
		Color: "#10b981",
		Email: &emailFalse,
		DOB:   &dobFalse,
	})
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	if t1.Name != "Pediatric-Card" || t1.Color != "#10b981" || t1.Email != false || t1.DOB != false || t1.Phone != true {
		t.Fatalf("unexpected template values: %+v", t1)
	}

	// 3. Duplicate template name rejection (case-insensitive)
	_, err = app.CreateTemplate(ctx, admin, domain.CreatePatientIDCardTemplateInput{
		Name:  "pediatric-card",
		Color: "#3b82f6",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate template name, got %v", err)
	}

	// 4. Update template name and color
	newName := "Pediatric-Gold"
	newColor := "#059669"
	updated, err := app.UpdateTemplate(ctx, admin, t1.ID, domain.UpdatePatientIDCardTemplateInput{
		Name:  &newName,
		Color: &newColor,
	})
	if err != nil {
		t.Fatalf("failed to update template: %v", err)
	}
	if updated.Name != "Pediatric-Gold" || updated.Color != "#059669" {
		t.Fatalf("updated template mismatch: %+v", updated)
	}

	// 5. Update template duplicate name collision
	existingName := "Standard"
	_, err = app.UpdateTemplate(ctx, admin, t1.ID, domain.UpdatePatientIDCardTemplateInput{
		Name: &existingName,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when renaming to existing template name, got %v", err)
	}

	// 6. Delete unreferenced template succeeds
	if err := app.DeleteTemplate(ctx, admin, t1.ID); err != nil {
		t.Fatalf("failed to delete unreferenced template: %v", err)
	}
	_, err = app.GetTemplate(ctx, admin, t1.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for deleted template, got %v", err)
	}
}

func TestSmartCards_GenerationModesReassignmentAndDeletion(t *testing.T) {
	db, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	// Create 3 patients
	p1, err := hospital.Register(ctx, admin, domain.PatientInput{GivenName: "Abebe", FamilyName: "Kebede", DateOfBirth: "1990-01-01", Phone: "+251911111111"})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := hospital.Register(ctx, admin, domain.PatientInput{GivenName: "Almaz", FamilyName: "Ayana", DateOfBirth: "1995-05-05", Phone: "+251922222222"})
	if err != nil {
		t.Fatal(err)
	}
	p3, err := hospital.Register(ctx, admin, domain.PatientInput{GivenName: "Bekele", FamilyName: "Desta", DateOfBirth: "1988-08-08", Phone: "+251933333333"})
	if err != nil {
		t.Fatal(err)
	}
	_ = p2
	_ = p3


	// Retrieve Standard template
	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) < 2 {
		t.Fatal("templates not available")
	}
	standardTmpl := templates[0]
	vipTmpl := templates[1]

	// ─── Mode 1: "single" patient generation ──────────────────────────────────
	genSingle, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: standardTmpl.ID,
		Mode:       "single",
		PatientID:  p1.ID,
	})
	if err != nil {
		t.Fatalf("failed single generation: %v", err)
	}
	if genSingle.GeneratedCount != 1 || len(genSingle.Cards) != 1 {
		t.Fatalf("expected 1 card generated, got %d", genSingle.GeneratedCount)
	}
	cardP1 := genSingle.Cards[0]
	if cardP1.PatientID != p1.ID || cardP1.PatientUniqueID == "" {
		t.Fatalf("invalid generated card: %+v", cardP1)
	}
	origP1UniqueID := cardP1.PatientUniqueID

	// ─── Mode 2: "unassigned" generation ─────────────────────────────────────
	// p2 and p3 do not have cards yet, p1 already has an active card.
	genUnassigned, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: standardTmpl.ID,
		Mode:       "unassigned",
	})
	if err != nil {
		t.Fatalf("failed unassigned generation: %v", err)
	}
	// Should only generate for p2 and p3 (2 cards)
	if genUnassigned.GeneratedCount != 2 {
		t.Fatalf("expected 2 unassigned cards generated, got %d", genUnassigned.GeneratedCount)
	}

	// ─── Idempotency & Persistence: Verify patient_unique_id never changes ───
	genRepeatSingle, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: standardTmpl.ID,
		Mode:       "single",
		PatientID:  p1.ID,
	})
	if err != nil {
		t.Fatalf("failed repeat single generation: %v", err)
	}
	if genRepeatSingle.Cards[0].PatientUniqueID != origP1UniqueID {
		t.Fatalf("patient unique ID mutated on re-generation! was %s, now %s", origP1UniqueID, genRepeatSingle.Cards[0].PatientUniqueID)
	}

	// ─── Mode 3: "all" generation & Reassignment ──────────────────────────────
	// Switch all patients to VIP template
	genAll, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: vipTmpl.ID,
		Mode:       "all",
	})
	if err != nil {
		t.Fatalf("failed all generation: %v", err)
	}
	if genAll.GeneratedCount != 3 {
		t.Fatalf("expected 3 cards generated in all mode, got %d", genAll.GeneratedCount)
	}

	// Verify that previous cards were revoked upon reassignment
	var activeCardsCount, revokedCardsCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM patient_smart_card WHERE status='active'`).Scan(&activeCardsCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM patient_smart_card WHERE status='revoked'`).Scan(&revokedCardsCount); err != nil {
		t.Fatal(err)
	}
	if activeCardsCount != 3 {
		t.Fatalf("expected exactly 3 active cards after reassignment, got %d", activeCardsCount)
	}
	if revokedCardsCount < 3 {
		t.Fatalf("expected previously issued cards to be revoked, got %d revoked", revokedCardsCount)
	}

	// ─── Referenced-Template Deletion: Clears Assignments, Preserves Patients, Revokes Cards ─────
	// VIP template is currently in use by patients p1, p2, p3
	err = app.DeleteTemplate(ctx, admin, vipTmpl.ID)
	if err != nil {
		t.Fatalf("expected successful deletion of referenced template clearing assignments, got %v", err)
	}

	// Verify template is gone
	_, err = app.GetTemplate(ctx, admin, vipTmpl.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for deleted template, got %v", err)
	}

	// Verify patients still exist, unique IDs preserved, and smart_card_template_id is NULL
	for _, pid := range []string{p1.ID, p2.ID, p3.ID} {
		var pExists bool
		var pUID string
		var pTmplID *string
		err = db.QueryRow(ctx, `SELECT true, patient_unique_id, smart_card_template_id::text FROM patient WHERE id = $1`, pid).Scan(&pExists, &pUID, &pTmplID)
		if err != nil || !pExists {
			t.Fatalf("patient %s was deleted or missing after template deletion!", pid)
		}
		if pUID == "" {
			t.Fatalf("patient %s unique ID was wiped after template deletion!", pid)
		}
		if pTmplID != nil {
			t.Fatalf("expected patient %s smart_card_template_id to be NULL after template deletion, got %v", pid, *pTmplID)
		}
	}

	// ─── Card Assignment Deletion: Preserves Patient & Revokes QR ─────────────
	// Issue a new card for p1 under standardTmpl to test isolated card assignment deletion
	_, err = app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: standardTmpl.ID,
		Mode:       "single",
		PatientID:  p1.ID,
	})
	if err != nil {
		t.Fatalf("failed generating card for p1: %v", err)
	}

	issuedList, err := app.ListIssuedCards(ctx, admin, domain.ListIssuedCardsFilter{Page: 1, PageSize: 10})
	if err != nil || len(issuedList.Cards) == 0 {
		t.Fatal("failed listing issued cards")
	}
	cardToDelete := issuedList.Cards[0]

	err = app.DeleteCardAssignment(ctx, admin, cardToDelete.ID)
	if err != nil {
		t.Fatalf("failed deleting card assignment: %v", err)
	}

	// Check patient record and unique ID are preserved
	var patientExists bool
	var preservedUniqueID string
	var currentTemplateID *string
	err = db.QueryRow(ctx, `SELECT true, patient_unique_id, smart_card_template_id::text FROM patient WHERE id = $1`, cardToDelete.PatientID).Scan(
		&patientExists, &preservedUniqueID, &currentTemplateID,
	)
	if err != nil || !patientExists {
		t.Fatalf("patient record was deleted or not found: %v", err)
	}
	if preservedUniqueID != cardToDelete.PatientUniqueID {
		t.Fatalf("patient unique ID lost or changed after card unassignment: %s vs %s", preservedUniqueID, cardToDelete.PatientUniqueID)
	}
	if currentTemplateID != nil {
		t.Fatalf("expected smart_card_template_id to be NULL after card unassignment, got %v", *currentTemplateID)
	}

	// Verification of unassigned/deleted card must return revoked!
	verifyUnassigned, err := app.VerifySmartCardQR(ctx, cardToDelete.QRToken)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if verifyUnassigned.Valid {
		t.Fatalf("unassigned/deleted card must NOT be valid! got valid: true")
	}
	if verifyUnassigned.Status != "revoked" {
		t.Fatalf("expected status 'revoked', got %s", verifyUnassigned.Status)
	}
}

func TestSmartCards_AuthorizedQRVerificationAndAccessSeparation(t *testing.T) {
	_, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	// 1. Create a template where email, phone, and dob are FALSE, but blood_group and unique_id are TRUE
	falseVal := false
	trueVal := true
	privacyTmpl, err := app.CreateTemplate(ctx, admin, domain.CreatePatientIDCardTemplateInput{
		Name:            "Privacy-Enforced",
		Color:           "#475569",
		Email:           &falseVal,
		Phone:           &falseVal,
		DOB:             &falseVal,
		BloodGroup:      &trueVal,
		Address:         &falseVal,
		PatientUniqueID: &trueVal,
	})
	if err != nil {
		t.Fatalf("failed creating privacy template: %v", err)
	}

	// 2. Register patient with full demographics
	p, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "Chala",
		FamilyName:  "Tolossa",
		DateOfBirth: "1985-04-12",
		Phone:       "+251944556677",
	})
	if err != nil {
		t.Fatal(err)
	}

	// 3. Issue card under Privacy-Enforced template
	genRes, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: privacyTmpl.ID,
		Mode:       "single",
		PatientID:  p.ID,
	})
	if err != nil || len(genRes.Cards) == 0 {
		t.Fatalf("failed generating card: %v", err)
	}
	card := genRes.Cards[0]

	// 4. Verify QR code
	verifyRes, err := app.VerifySmartCardQR(ctx, card.QRToken)
	if err != nil {
		t.Fatalf("failed verifying valid card: %v", err)
	}
	if !verifyRes.Valid {
		t.Fatalf("expected card to be valid")
	}
	if verifyRes.PatientName != "Chala Tolossa" {
		t.Fatalf("expected PatientName 'Chala Tolossa', got %q", verifyRes.PatientName)
	}
	if verifyRes.PatientUniqueID != card.PatientUniqueID {
		t.Fatalf("expected PatientUniqueID %s, got %s", card.PatientUniqueID, verifyRes.PatientUniqueID)
	}
	// Verify that unpermitted fields are stripped/empty
	if verifyRes.Email != "" {
		t.Fatalf("email should be empty per template toggle, got %q", verifyRes.Email)
	}
	if verifyRes.Phone != "" {
		t.Fatalf("phone should be empty per template toggle, got %q", verifyRes.Phone)
	}
	if verifyRes.DOB != "" {
		t.Fatalf("dob should be empty per template toggle, got %q", verifyRes.DOB)
	}
	if verifyRes.Address != "" {
		t.Fatalf("address should be empty per template toggle, got %q", verifyRes.Address)
	}

	// 5. Test forged / non-existent QR token returns ErrNotFound
	_, err = app.VerifySmartCardQR(ctx, "nonexistent-token-1234567890abcdef")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existent token, got %v", err)
	}

	// 6. Revoke card and verify revoked status
	err = app.RevokeSmartCard(ctx, admin, card.ID, "Patient lost card")
	if err != nil {
		t.Fatalf("failed revoking card: %v", err)
	}
	revokedRes, err := app.VerifySmartCardQR(ctx, card.QRToken)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if revokedRes.Valid {
		t.Fatalf("revoked card must NOT be valid")
	}
	if revokedRes.Status != "revoked" || revokedRes.RevokedReason != "Patient lost card" {
		t.Fatalf("unexpected revoked result: %+v", revokedRes)
	}
}

func TestSmartCards_ConcurrentGenerationSafety(t *testing.T) {
	db, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	// Register 5 patients
	var patientIDs []string
	for i := 1; i <= 5; i++ {
		p, err := hospital.Register(ctx, admin, domain.PatientInput{
			GivenName:   fmt.Sprintf("Concurrent%d", i),
			FamilyName:  "Tester",
			DateOfBirth: "1992-02-02",
		})
		if err != nil {
			t.Fatal(err)
		}
		patientIDs = append(patientIDs, p.ID)
	}

	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) == 0 {
		t.Fatal("templates unavailable")
	}
	tmpl := templates[0]

	// Run 8 concurrent generation requests across these patients
	concurrency := 8
	var wg sync.WaitGroup
	wg.Add(concurrency)
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			mode := "all"
			if idx%2 == 0 {
				mode = "unassigned"
			}
			_, e := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
				TemplateID: tmpl.ID,
				Mode:       mode,
			})
			if e != nil {
				errCh <- e
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		if e != nil {
			t.Fatalf("concurrent generation failed: %v", e)
		}
	}

	// Assert invariants:
	// 1. Each patient has EXACTLY 1 active card
	for _, pid := range patientIDs {
		var activeCount int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM patient_smart_card WHERE patient_id=$1 AND status='active'`, pid).Scan(&activeCount); err != nil {
			t.Fatal(err)
		}
		if activeCount != 1 {
			t.Fatalf("patient %s has %d active cards, expected exactly 1", pid, activeCount)
		}
	}

	// 2. All active card numbers are unique
	var uniqueCardNumbers, totalActiveCards int
	if err := db.QueryRow(ctx, `SELECT count(DISTINCT card_number), count(*) FROM patient_smart_card WHERE status='active'`).Scan(&uniqueCardNumbers, &totalActiveCards); err != nil {
		t.Fatal(err)
	}
	if uniqueCardNumbers != totalActiveCards {
		t.Fatalf("card numbers collided! %d unique vs %d total", uniqueCardNumbers, totalActiveCards)
	}

	// 3. All patient_unique_ids are unique and non-empty
	var uniquePatientIDs, totalPatients int
	if err := db.QueryRow(ctx, `SELECT count(DISTINCT patient_unique_id), count(*) FROM patient WHERE patient_unique_id IS NOT NULL`).Scan(&uniquePatientIDs, &totalPatients); err != nil {
		t.Fatal(err)
	}
	if uniquePatientIDs != totalPatients {
		t.Fatalf("patient_unique_id collided! %d unique vs %d total", uniquePatientIDs, totalPatients)
	}
}

// ─── R1: Enforce patient access scope on card list ───────────────────────────
func TestSmartCards_PatientScopeEnforcement(t *testing.T) {
	db, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	doc1 := domain.Actor{ID: "doc-1", Role: "doctor"}
	doc2 := domain.Actor{ID: "doc-2", Role: "doctor"}
	patientUser1 := domain.Actor{ID: "patient-user-1", Role: "patient"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	_, err := db.Exec(ctx, `INSERT INTO "user"(id, name, email) VALUES('doc-1', 'Doctor One', 'doc1@example.com'), ('doc-2', 'Doctor Two', 'doc2@example.com'), ('patient-user-1', 'Patient One', 'p1@example.com') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `INSERT INTO staff_access(user_id, role, active) VALUES('doc-1', 'doctor', true), ('doc-2', 'doctor', true) ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatal(err)
	}

	// Create 2 patients
	p1, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "PatientOne",
		FamilyName:  "AssignedToDoc1",
		DateOfBirth: "1990-01-01",
		Phone:       "+251911000001",
	})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "PatientTwo",
		FamilyName:  "AssignedToDoc2",
		DateOfBirth: "1995-05-05",
		Phone:       "+251911000002",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Assign p1 to doc1 and patientUser1; assign p2 to doc2
	_, err = db.Exec(ctx, `UPDATE patient SET clinician_user_id = $1, user_id = $2 WHERE id = $3`, doc1.ID, patientUser1.ID, p1.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `UPDATE patient SET clinician_user_id = $1 WHERE id = $2`, doc2.ID, p2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Fetch standard template
	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) == 0 {
		t.Fatal(err)
	}
	tmpl := templates[0]

	// Generate cards for both patients
	genRes1, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: tmpl.ID,
		Mode:       "single",
		PatientID:  p1.ID,
	})
	if err != nil || len(genRes1.Cards) == 0 {
		t.Fatalf("failed generating card for p1: %v", err)
	}
	card1 := genRes1.Cards[0]

	genRes2, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: tmpl.ID,
		Mode:       "single",
		PatientID:  p2.ID,
	})
	if err != nil || len(genRes2.Cards) == 0 {
		t.Fatalf("failed generating card for p2: %v", err)
	}
	card2 := genRes2.Cards[0]

	// 1. Doc1 lists cards: must see ONLY card1
	doc1Cards, err := app.ListIssuedCards(ctx, doc1, domain.ListIssuedCardsFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("doc1 ListIssuedCards failed: %v", err)
	}
	if doc1Cards.TotalCount != 1 || len(doc1Cards.Cards) != 1 {
		t.Fatalf("doc1 should see exactly 1 card, got total=%d, len=%d", doc1Cards.TotalCount, len(doc1Cards.Cards))
	}
	if doc1Cards.Cards[0].ID != card1.ID {
		t.Fatalf("doc1 should see card1 (%s), got %s", card1.ID, doc1Cards.Cards[0].ID)
	}
	// Doc1 searching for PatientTwo must return 0 results
	doc1Search, err := app.ListIssuedCards(ctx, doc1, domain.ListIssuedCardsFilter{Page: 1, PageSize: 10, Search: "PatientTwo"})
	if err != nil {
		t.Fatal(err)
	}
	if doc1Search.TotalCount != 0 || len(doc1Search.Cards) != 0 {
		t.Fatalf("doc1 search should return 0 results for unassigned patient, got %d", doc1Search.TotalCount)
	}

	// 2. Doc2 lists cards: must see ONLY card2
	doc2Cards, err := app.ListIssuedCards(ctx, doc2, domain.ListIssuedCardsFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("doc2 ListIssuedCards failed: %v", err)
	}
	if doc2Cards.TotalCount != 1 || len(doc2Cards.Cards) != 1 {
		t.Fatalf("doc2 should see exactly 1 card, got total=%d, len=%d", doc2Cards.TotalCount, len(doc2Cards.Cards))
	}
	if doc2Cards.Cards[0].ID != card2.ID {
		t.Fatalf("doc2 should see card2 (%s), got %s", card2.ID, doc2Cards.Cards[0].ID)
	}
	// Verify doc2 never receives card1's QR bearer token
	for _, c := range doc2Cards.Cards {
		if c.QRToken == card1.QRToken {
			t.Fatalf("doc2 was disclosed card1's QR bearer token!")
		}
	}

	// 3. Patient user lists cards: sees only their own card
	pUserCards, err := app.ListIssuedCards(ctx, patientUser1, domain.ListIssuedCardsFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("patient user ListIssuedCards failed: %v", err)
	}
	if pUserCards.TotalCount != 1 || len(pUserCards.Cards) != 1 || pUserCards.Cards[0].ID != card1.ID {
		t.Fatalf("patient user should see only their own card, got total=%d", pUserCards.TotalCount)
	}

	// 4. Admin lists cards: sees both cards
	adminCards, err := app.ListIssuedCards(ctx, admin, domain.ListIssuedCardsFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if adminCards.TotalCount != 2 || len(adminCards.Cards) != 2 {
		t.Fatalf("admin should see both cards, got %d", adminCards.TotalCount)
	}
}

// ─── R2: Stale card deletion does not clear newer assignment ─────────────────
func TestSmartCards_StaleDeletionAndReassignmentPreservation(t *testing.T) {
	db, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	// Register patient
	p, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "StaleTest",
		FamilyName:  "Patient",
		DateOfBirth: "1988-08-08",
		Phone:       "+251911223344",
	})
	if err != nil {
		t.Fatal(err)
	}

	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) < 2 {
		t.Fatal("need at least 2 templates")
	}
	standardTmpl := templates[0]
	vipTmpl := templates[1]

	// 1. Issue Card A under Standard template
	resA, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: standardTmpl.ID,
		Mode:       "single",
		PatientID:  p.ID,
	})
	if err != nil || len(resA.Cards) == 0 {
		t.Fatalf("failed issuing Card A: %v", err)
	}
	cardA := resA.Cards[0]

	// 2. Reassign patient to Card B under VIP template
	resB, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID: vipTmpl.ID,
		Mode:       "single",
		PatientID:  p.ID,
	})
	if err != nil || len(resB.Cards) == 0 {
		t.Fatalf("failed issuing Card B: %v", err)
	}
	cardB := resB.Cards[0]

	// Verify Card A is revoked and Card B is active
	var statusA, statusB string
	if err := db.QueryRow(ctx, `SELECT status FROM patient_smart_card WHERE id=$1`, cardA.ID).Scan(&statusA); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT status FROM patient_smart_card WHERE id=$1`, cardB.ID).Scan(&statusB); err != nil {
		t.Fatal(err)
	}
	if statusA != "revoked" || statusB != "active" {
		t.Fatalf("expected cardA=revoked, cardB=active; got %s, %s", statusA, statusB)
	}

	// Verify patient's current assignment is VIP template
	var tmplID *string
	if err := db.QueryRow(ctx, `SELECT smart_card_template_id::text FROM patient WHERE id=$1`, p.ID).Scan(&tmplID); err != nil {
		t.Fatal(err)
	}
	if tmplID == nil || *tmplID != vipTmpl.ID {
		t.Fatalf("expected patient to have vipTmpl %s, got %v", vipTmpl.ID, tmplID)
	}

	// 3. Stale Deletion: delete old revoked Card A
	err = app.DeleteCardAssignment(ctx, admin, cardA.ID)
	if err != nil {
		t.Fatalf("deleting stale card should be a safe no-op, got error: %v", err)
	}

	// Verify patient's current assignment is STILL VIP template!
	if err := db.QueryRow(ctx, `SELECT smart_card_template_id::text FROM patient WHERE id=$1`, p.ID).Scan(&tmplID); err != nil {
		t.Fatal(err)
	}
	if tmplID == nil || *tmplID != vipTmpl.ID {
		t.Fatalf("stale deletion of card A cleared patient assignment! tmplID is %v", tmplID)
	}

	// Verify Card B is still active and valid!
	verifyB, err := app.VerifySmartCardQR(ctx, cardB.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if !verifyB.Valid || verifyB.Status != "active" {
		t.Fatalf("Card B was invalidated by stale deletion of Card A! valid=%v, status=%s", verifyB.Valid, verifyB.Status)
	}

	// 4. Repeated deletion of Card A remains a safe no-op
	err = app.DeleteCardAssignment(ctx, admin, cardA.ID)
	if err != nil {
		t.Fatalf("repeated deletion should be a safe no-op: %v", err)
	}

	// 5. Delete active Card B: clears assignment and revokes Card B
	err = app.DeleteCardAssignment(ctx, admin, cardB.ID)
	if err != nil {
		t.Fatalf("failed deleting active card B: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT smart_card_template_id::text FROM patient WHERE id=$1`, p.ID).Scan(&tmplID); err != nil {
		t.Fatal(err)
	}
	if tmplID != nil {
		t.Fatalf("expected smart_card_template_id to be NULL after deleting active card B, got %v", *tmplID)
	}
	verifyBRevoked, err := app.VerifySmartCardQR(ctx, cardB.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if verifyBRevoked.Valid || verifyBRevoked.Status != "revoked" {
		t.Fatalf("Card B should be revoked after deletion: valid=%v, status=%s", verifyBRevoked.Valid, verifyBRevoked.Status)
	}
}

// ─── R3: Conservative disclosure on missing templates & IssueSmartCard ─────────
func TestSmartCards_ConservativeLegacyDisclosurePolicy(t *testing.T) {
	db, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	// Register patient with full demographics
	p, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "Legacy",
		FamilyName:  "Patient",
		DateOfBirth: "1975-03-15",
		Phone:       "+251911998877",
	})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Issue card without template_id (simulating legacy card)
	rawRand := make([]byte, 16)
	rand.Read(rawRand)
	legacyToken := hex.EncodeToString(rawRand)
	now := time.Now().UTC()
	var legacyCardID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient_smart_card (patient_id, card_number, qr_token, template_name, template_id, status, expires_at, issued_by)
		VALUES ($1, $2, $3, 'legacy', NULL, 'active', $4, $5)
		RETURNING id
	`, p.ID, "HMS-LEGACY-001", legacyToken, now.AddDate(1, 0, 0), admin.ID).Scan(&legacyCardID)
	if err != nil {
		t.Fatalf("failed inserting legacy card: %v", err)
	}

	// 2. Verify legacy card via VerifySmartCardQR:
	// Status must be active, but demographics (email, phone, dob, blood group, address) MUST be empty!
	res, err := app.VerifySmartCardQR(ctx, legacyToken)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !res.Valid || res.Status != "active" {
		t.Fatalf("legacy card must be valid and active, got valid=%v status=%s", res.Valid, res.Status)
	}
	if res.PatientName != "Legacy Patient" {
		t.Fatalf("expected PatientName 'Legacy Patient', got %s", res.PatientName)
	}
	if res.Email != "" {
		t.Fatalf("legacy card without template must NOT disclose email, got %s", res.Email)
	}
	if res.Phone != "" {
		t.Fatalf("legacy card without template must NOT disclose phone, got %s", res.Phone)
	}
	if res.DOB != "" {
		t.Fatalf("legacy card without template must NOT disclose DOB, got %s", res.DOB)
	}
	if res.BloodGroup != "" {
		t.Fatalf("legacy card without template must NOT disclose blood group, got %s", res.BloodGroup)
	}
	if res.Address != "" {
		t.Fatalf("legacy card without template must NOT disclose address, got %s", res.Address)
	}

	// 3. Issue card through app.IssueSmartCard: verify it assigns unique ID and routes properly
	issuedCard, err := app.IssueSmartCard(ctx, admin, domain.IssueSmartCardInput{
		PatientID:    p.ID,
		TemplateName: "Standard",
		ValidityDays: 365,
	})
	if err != nil {
		t.Fatalf("IssueSmartCard failed: %v", err)
	}
	if issuedCard.PatientUniqueID == "" {
		t.Fatal("IssueSmartCard must ensure patient_unique_id is assigned")
	}

	// Verify legacy card was revoked by IssueSmartCard
	var legStatus string
	if err := db.QueryRow(ctx, `SELECT status FROM patient_smart_card WHERE id=$1`, legacyCardID).Scan(&legStatus); err != nil {
		t.Fatal(err)
	}
	if legStatus != "revoked" {
		t.Fatalf("legacy card should be revoked after new card issuance, got %s", legStatus)
	}

	// 4. Verify newly issued Standard card exposes permitted demographics per Standard template
	resStd, err := app.VerifySmartCardQR(ctx, issuedCard.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if !resStd.Valid || resStd.Phone == "" {
		t.Fatalf("Standard template card must expose phone (+251911998877), got %s", resStd.Phone)
	}
}

// ─── R4: Durable generation idempotency with payload matching ────────────────
func TestSmartCards_DurableGenerationIdempotency(t *testing.T) {
	_, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	p, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "IdemTest",
		FamilyName:  "Patient",
		DateOfBirth: "1994-04-04",
	})
	if err != nil {
		t.Fatal(err)
	}

	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) < 2 {
		t.Fatal("need templates")
	}
	tmpl1 := templates[0]
	tmpl2 := templates[1]

	idempotencyKey := "idem-test-key-001"

	// 1. Initial generation request
	gen1, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmpl1.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil || len(gen1.Cards) == 0 {
		t.Fatalf("initial generation failed: %v", err)
	}
	card1 := gen1.Cards[0]

	// 2. Sequential retry with same IdempotencyKey and same payload:
	// MUST return the exact same cards without creating duplicates or revoking card1
	gen2, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmpl1.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil || len(gen2.Cards) == 0 {
		t.Fatalf("retry generation failed: %v", err)
	}
	if gen2.Cards[0].ID != card1.ID {
		t.Fatalf("retry generated new card! expected %s, got %s", card1.ID, gen2.Cards[0].ID)
	}
	if gen2.Cards[0].QRToken != card1.QRToken {
		t.Fatalf("retry generated new QR token! expected %s, got %s", card1.QRToken, gen2.Cards[0].QRToken)
	}

	// Verify original QR token is STILL valid and active!
	verify1, err := app.VerifySmartCardQR(ctx, card1.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if !verify1.Valid || verify1.Status != "active" {
		t.Fatalf("original card was invalidated on retry: valid=%v, status=%s", verify1.Valid, verify1.Status)
	}

	// 3. Retry with same IdempotencyKey but mismatched payload: MUST return ErrConflict
	_, err = app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmpl2.ID, // different template
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: idempotencyKey,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on mismatched payload with same idempotency key, got %v", err)
	}

	// 4. Deliberate new reassignment with a NEW idempotency key:
	// MUST succeed and revoke old card1
	newKey := "idem-test-key-002"
	gen3, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmpl2.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: newKey,
	})
	if err != nil || len(gen3.Cards) == 0 {
		t.Fatalf("new reassignment failed: %v", err)
	}
	card3 := gen3.Cards[0]
	if card3.ID == card1.ID {
		t.Fatal("expected new card on reassignment")
	}

	// Original card1 is now revoked by the deliberate reassignment
	verifyOld, err := app.VerifySmartCardQR(ctx, card1.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if verifyOld.Valid || verifyOld.Status != "revoked" {
		t.Fatalf("old card must be revoked after deliberate reassignment: valid=%v, status=%s", verifyOld.Valid, verifyOld.Status)
	}

	// New card3 is active and valid
	verifyNew, err := app.VerifySmartCardQR(ctx, card3.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if !verifyNew.Valid || verifyNew.Status != "active" {
		t.Fatalf("new card must be active: valid=%v, status=%s", verifyNew.Valid, verifyNew.Status)
	}
}

// ─── F1: Database upgrade from prior 057 to 058 ─────────────────────────────
func TestSmartCards_MigrationUpgrade057To058(t *testing.T) {
	dsn := os.Getenv("HMS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set HMS_TEST_DATABASE_URL to run isolated-schema PostgreSQL tests")
	}
	ctx := context.Background()
	adminDB, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer adminDB.Close()

	random := make([]byte, 12)
	rand.Read(random)
	schema := "hms_test_upg_" + hex.EncodeToString(random)
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = adminDB.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = adminDB.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
	}()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Apply migrations up to 057 only
	migrations, err := filepath.Glob("../../../../../db/migrations/*.sql")
	if err != nil || len(migrations) < 3 {
		t.Fatal("migration files unavailable")
	}
	for _, file := range migrations {
		if filepath.Base(file) > "057_patient_smart_cards_parity.sql" {
			continue // stop after 057
		}
		sql, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = db.Exec(ctx, string(sql)); e != nil {
			t.Fatalf("failed migration %s: %v", file, e)
		}
	}

	// Seed actor
	admin := domain.Actor{ID: "admin-upg", Role: "admin", Name: "Admin Upgrade"}
	if _, err = db.Exec(ctx, `INSERT INTO "user"(id, name, email) VALUES($1, $2, 'upg@example.test')`, admin.ID, admin.Name); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO staff_access(user_id, role) VALUES($1, $2)`, admin.ID, admin.Role); err != nil {
		t.Fatal(err)
	}

	// Now apply migration 058 explicitly
	sql058, err := os.ReadFile("../../../../../db/migrations/058_smart_card_generation_idempotency.sql")
	if err != nil {
		t.Fatalf("failed reading 058: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql058)); err != nil {
		t.Fatalf("failed applying 058 upgrade: %v", err)
	}

	// Verify table exists
	var tableExists bool
	err = db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = $1 AND table_name = 'smart_card_generation_request'
		)
	`, schema).Scan(&tableExists)
	if err != nil || !tableExists {
		t.Fatalf("expected smart_card_generation_request to exist after 058, exists=%v err=%v", tableExists, err)
	}

	// Test store operations against upgraded database
	store := Store{DB: db}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	p, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "Upgrade",
		FamilyName:  "Test",
		DateOfBirth: "1991-01-01",
	})
	if err != nil {
		t.Fatal(err)
	}

	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) == 0 {
		t.Fatal("templates missing")
	}

	// Generation and replay must succeed on upgraded schema
	res1, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     templates[0].ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: "upg-idem-key-1",
	})
	if err != nil || len(res1.Cards) == 0 {
		t.Fatalf("generation failed on upgraded schema: %v", err)
	}

	res2, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     templates[0].ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: "upg-idem-key-1",
	})
	if err != nil || len(res2.Cards) == 0 {
		t.Fatalf("replay failed on upgraded schema: %v", err)
	}
	if res2.Cards[0].ID != res1.Cards[0].ID {
		t.Fatalf("expected identical card on replay after upgrade, got %s vs %s", res1.Cards[0].ID, res2.Cards[0].ID)
	}
}

// ─── F2: Completed requests must never re-issue credentials after revocation ──
func TestSmartCards_ReplayAfterRevocationAndDeletion(t *testing.T) {
	db, store, cleanup := setupSmartCardsTestDB(t)
	defer cleanup()
	ctx := context.Background()
	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	hospital := application.Hospital{Store: store, Now: time.Now}
	app := application.PatientExtensionsService{Store: store}

	p, err := hospital.Register(ctx, admin, domain.PatientInput{
		GivenName:   "RevokeReplay",
		FamilyName:  "Patient",
		DateOfBirth: "1993-03-03",
	})
	if err != nil {
		t.Fatal(err)
	}

	templates, err := app.ListTemplates(ctx, admin)
	if err != nil || len(templates) < 2 {
		t.Fatal("need templates")
	}
	tmplA := templates[0]
	tmplB := templates[1]

	keyA := "replay-test-key-A"

	// 1. Generate Card A with keyA
	genA, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplA.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyA,
	})
	if err != nil || len(genA.Cards) == 0 {
		t.Fatalf("failed generating Card A: %v", err)
	}
	cardA := genA.Cards[0]

	// 2. Delete Card A assignment (deliberate unassignment)
	err = app.DeleteCardAssignment(ctx, admin, cardA.ID)
	if err != nil {
		t.Fatalf("failed deleting Card A assignment: %v", err)
	}

	// Confirm patient has no active cards and no template assignment
	var activeCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM patient_smart_card WHERE patient_id=$1 AND status='active'`, p.ID).Scan(&activeCount); err != nil {
		t.Fatal(err)
	}
	if activeCount != 0 {
		t.Fatalf("expected 0 active cards after assignment deletion, got %d", activeCount)
	}
	var tmplID *string
	if err := db.QueryRow(ctx, `SELECT smart_card_template_id::text FROM patient WHERE id=$1`, p.ID).Scan(&tmplID); err != nil {
		t.Fatal(err)
	}
	if tmplID != nil {
		t.Fatalf("expected smart_card_template_id to be NULL, got %v", *tmplID)
	}

	// 3. Replay identical request with keyA after deletion
	// CRITICAL F2 REQUIREMENT: Replay MUST NOT issue a new active credential or restore the assignment!
	replayA, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplA.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyA,
	})
	if err != nil {
		t.Fatalf("replaying request should return without error: %v", err)
	}
	if len(replayA.Cards) == 0 {
		t.Fatal("expected cached cards returned")
	}

	// Replay must return Card A (with its current status "revoked"), NOT a new active card!
	if replayA.Cards[0].ID != cardA.ID {
		t.Fatalf("replaying completed request after deletion issued a NEW credential! expected %s, got %s", cardA.ID, replayA.Cards[0].ID)
	}
	if replayA.Cards[0].Status != "revoked" {
		t.Fatalf("replaying completed request after deletion returned status '%s', expected 'revoked'", replayA.Cards[0].Status)
	}

	// Patient must STILL have 0 active cards in DB!
	if err := db.QueryRow(ctx, `SELECT count(*) FROM patient_smart_card WHERE patient_id=$1 AND status='active'`, p.ID).Scan(&activeCount); err != nil {
		t.Fatal(err)
	}
	if activeCount != 0 {
		t.Fatalf("replaying completed request after deletion issued a NEW active credential in database! activeCount=%d", activeCount)
	}

	// Patient assignment in DB must STILL be NULL!
	if err := db.QueryRow(ctx, `SELECT smart_card_template_id::text FROM patient WHERE id=$1`, p.ID).Scan(&tmplID); err != nil {
		t.Fatal(err)
	}
	if tmplID != nil {
		t.Fatalf("replaying completed request after deletion restored patient assignment! got %v", *tmplID)
	}

	// QR verification of Card A must STILL be revoked!
	verifyA, err := app.VerifySmartCardQR(ctx, cardA.QRToken)
	if err != nil {
		t.Fatal(err)
	}
	if verifyA.Valid || verifyA.Status != "revoked" {
		t.Fatalf("Card A was re-activated! valid=%v status=%s", verifyA.Valid, verifyA.Status)
	}

	// 4. Replay after deliberate reassignment:
	// Generate Card B with new keyB
	keyB := "replay-test-key-B"
	genB, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplB.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyB,
	})
	if err != nil || len(genB.Cards) == 0 {
		t.Fatalf("failed generating Card B: %v", err)
	}
	cardB := genB.Cards[0]

	// Replaying older keyA must NOT supersede Card B!
	replayOldA, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplA.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyA,
	})
	if err != nil {
		t.Fatal(err)
	}
	if replayOldA.Cards[0].ID != cardA.ID || replayOldA.Cards[0].Status != "revoked" {
		t.Fatalf("replaying old key returned unexpected card: %v", replayOldA.Cards[0])
	}

	// Card B must STILL be active in DB!
	var statusB string
	if err := db.QueryRow(ctx, `SELECT status FROM patient_smart_card WHERE id=$1`, cardB.ID).Scan(&statusB); err != nil {
		t.Fatal(err)
	}
	if statusB != "active" {
		t.Fatalf("replaying older key superseded active card B! statusB=%s", statusB)
	}

	// Patient assignment must STILL be tmplB!
	if err := db.QueryRow(ctx, `SELECT smart_card_template_id::text FROM patient WHERE id=$1`, p.ID).Scan(&tmplID); err != nil {
		t.Fatal(err)
	}
	if tmplID == nil || *tmplID != tmplB.ID {
		t.Fatalf("replaying older key mutated patient assignment: expected %s, got %v", tmplB.ID, tmplID)
	}

	// 5. Replay after direct revocation:
	err = app.RevokeSmartCard(ctx, admin, cardB.ID, "lost")
	if err != nil {
		t.Fatalf("failed direct revocation: %v", err)
	}
	replayB, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplB.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyB,
	})
	if err != nil {
		t.Fatal(err)
	}
	if replayB.Cards[0].ID != cardB.ID || replayB.Cards[0].Status != "revoked" {
		t.Fatalf("replaying after direct revocation must return revoked card B, got %v", replayB.Cards[0])
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM patient_smart_card WHERE patient_id=$1 AND status='active'`, p.ID).Scan(&activeCount); err != nil {
		t.Fatal(err)
	}
	if activeCount != 0 {
		t.Fatalf("replaying after direct revocation resurrected card B! activeCount=%d", activeCount)
	}

	// 6. Replay after template deletion:
	tmplC, err := app.CreateTemplate(ctx, admin, domain.CreatePatientIDCardTemplateInput{
		Name:  "Temp Replay Template",
		Color: "#abcdef",
	})
	if err != nil {
		t.Fatal(err)
	}
	keyC := "replay-test-key-C"
	genC, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplC.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyC,
	})
	if err != nil || len(genC.Cards) == 0 {
		t.Fatal(err)
	}
	cardC := genC.Cards[0]

	// Delete template tmplC
	if err := app.DeleteTemplate(ctx, admin, tmplC.ID); err != nil {
		t.Fatalf("failed deleting template tmplC: %v", err)
	}

	// Replay keyC: must return cardC (revoked), NOT crash or insert invalid records
	replayC, err := app.GenerateSmartCards(ctx, admin, domain.GenerateSmartCardsInput{
		TemplateID:     tmplC.ID,
		Mode:           "single",
		PatientID:      p.ID,
		IdempotencyKey: keyC,
	})
	if err != nil {
		t.Fatalf("replaying after template deletion must succeed: %v", err)
	}
	if replayC.Cards[0].ID != cardC.ID || replayC.Cards[0].Status != "revoked" {
		t.Fatalf("expected card C revoked, got %v", replayC.Cards[0])
	}
}
