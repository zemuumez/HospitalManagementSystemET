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
