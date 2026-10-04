package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Tests use a fresh schema, never patient tables in the public development schema.
func TestClinicalTransactions(t *testing.T) {
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
	defer adminDB.Close()
	random := make([]byte, 12)
	if _, err = rand.Read(random); err != nil {
		t.Fatal(err)
	}
	schema := "hms_test_" + hex.EncodeToString(random)
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = adminDB.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, e := adminDB.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); e != nil {
			t.Error(e)
		}
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
	migrations, err := filepath.Glob("../../../../../db/migrations/*.sql")
	if err != nil || len(migrations) < 3 {
		t.Fatal("migration files unavailable")
	}
	for _, file := range migrations {
		sql, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = db.Exec(ctx, string(sql)); e != nil {
			t.Fatal(file, e)
		}
	}
	actors := []domain.Actor{{ID: "admin", Role: "admin"}, {ID: "doctor", Role: "doctor", Name: "Doctor"}, {ID: "other-doctor", Role: "doctor"}, {ID: "patient", Role: "patient"}, {ID: "reception", Role: "receptionist"}}
	for _, a := range actors {
		if _, err = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test')`, a.ID); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,$2)`, a.ID, a.Role); err != nil {
			t.Fatal(err)
		}
	}
	store := Store{DB: db}
	clinic := application.Clinical{Store: store, Now: time.Now}
	hospital := application.Hospital{Store: store, Now: time.Now}
	patients := []domain.Patient{}
	cases := []domain.Case{}
	for n := 0; n < 2; n++ {
		p, e := hospital.Register(ctx, actors[0], domain.PatientInput{GivenName: "Synthetic", FamilyName: "Test", DateOfBirth: "2000-01-01"})
		if e != nil {
			t.Fatal(e)
		}
		patients = append(patients, p)
		c, e := clinic.CreateCase(ctx, actors[0], domain.CaseInput{PatientID: p.ID, DoctorID: "doctor"})
		if e != nil {
			t.Fatal(e)
		}
		cases = append(cases, c)
	}
	if err = hospital.LinkPatient(ctx, actors[0], patients[0].ID, domain.PatientAccess{UserID: "patient", ClinicianID: "doctor"}); err != nil {
		t.Fatal(err)
	}
	bed, err := clinic.CreateBed(ctx, actors[0], domain.BedInput{Name: "A1", Type: "General", ChargeMinor: 12345})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clinic.CreateBed(ctx, actors[3], domain.BedInput{Name: "Blocked", Type: "General"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("patient created bed", err)
	}
	inputs := []domain.EncounterInput{{Kind: "ipd", CaseID: cases[0].ID, BedID: bed.ID, AdmittedAt: time.Now().Add(-time.Minute)}, {Kind: "ipd", CaseID: cases[1].ID, BedID: bed.ID, AdmittedAt: time.Now().Add(-time.Minute)}}
	results := make([]domain.Encounter, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results[n], errs[n] = clinic.Admit(ctx, actors[0], inputs[n], "admission-key-000"+string(rune('0'+n)))
		}(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	loser := 1 - winner
	if errs[winner] != nil || !errors.Is(errs[loser], domain.ErrStale) {
		t.Fatalf("bed race: %v", errs)
	}
	enc := results[winner]
	retry, e := clinic.Admit(ctx, actors[0], inputs[winner], "admission-key-000"+string(rune('0'+winner)))
	if e != nil || retry.ID != enc.ID {
		t.Fatal("retry did not return same admission", e)
	}
	beds, e := clinic.Beds(ctx, actors[0], 1)
	if e != nil || beds[0].Available {
		t.Fatal("bed occupancy not reflected", e)
	}
	if _, e = clinic.SignNote(ctx, actors[2], enc.ID, domain.NoteInput{Body: "Unauthorized"}, "note-key-00000001"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("unassigned doctor wrote note", e)
	}
	if _, e = clinic.Notes(ctx, actors[4], enc.ID); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("reception read clinical notes", e)
	}
	note, e := clinic.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: "Signed synthetic clinical assessment"}, "note-key-00000002")
	if e != nil {
		t.Fatal(e)
	}
	duplicate, e := clinic.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: note.Body}, "note-key-00000002")
	if e != nil || duplicate.ID != note.ID {
		t.Fatal("note retry duplicated", e)
	}
	if _, e = db.Exec(ctx, `UPDATE clinical_note SET body='changed' WHERE id=$1`, note.ID); e == nil {
		t.Fatal("signed note was changed")
	}
	if _, e = db.Exec(ctx, `DELETE FROM clinical_note WHERE id=$1`, note.ID); e == nil {
		t.Fatal("signed note was deleted")
	}
	own, e := clinic.Encounters(ctx, actors[3], "ipd", 1)
	if e != nil {
		t.Fatal(e)
	}
	for _, row := range own {
		if row.PatientID != patients[0].ID {
			t.Fatal("patient saw someone else's admission")
		}
	}
	if _, e = clinic.Discharge(ctx, actors[0], enc.ID, domain.Discharge{Version: 1, Summary: "Admin cannot sign discharge"}); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("admin signed discharge", e)
	}
	out, e := clinic.Discharge(ctx, actors[1], enc.ID, domain.Discharge{Version: 1, Summary: "Synthetic discharge summary"})
	if e != nil || out.Status != "discharged" {
		t.Fatal(e)
	}
	if _, e = clinic.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: "Late note"}, "note-key-00000003"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("closed encounter accepted note", e)
	}
	beds, e = clinic.Beds(ctx, actors[0], 1)
	if e != nil || !beds[0].Available {
		t.Fatal("discharge did not free bed", e)
	}
	if _, e = clinic.Admit(ctx, actors[0], inputs[loser], "admission-key-rebook"); e != nil {
		t.Fatal("released bed could not be allocated", e)
	}
	billing := application.Billing{Store: store, Now: time.Now}
	account, e := billing.CreateAccount(ctx, actors[0], "Consultation")
	if e != nil {
		t.Fatal(e)
	}
	invoiceInput := domain.InvoiceInput{PatientID: patients[0].ID, InvoiceDate: time.Now().In(domain.HospitalLocation).Format("2006-01-02"), DiscountBasisPoints: 1000, Lines: []domain.InvoiceLine{{AccountID: account.ID, Quantity: 3, UnitPriceMinor: 333}}}
	invoice, e := billing.CreateInvoice(ctx, actors[0], invoiceInput, "invoice-key-00000001")
	if e != nil || invoice.TotalMinor != 899 {
		t.Fatal("invoice totals", e, invoice.TotalMinor)
	}
	for _, statement := range []string{
		`UPDATE invoice SET total_minor=1 WHERE id=$1`,
		`UPDATE invoice SET patient_id='00000000-0000-0000-0000-000000000001' WHERE id=$1`,
		`UPDATE invoice SET sealed=false WHERE id=$1`,
		`DELETE FROM invoice WHERE id=$1`,
		`UPDATE invoice_line SET quantity=1 WHERE invoice_id=$1`,
		`DELETE FROM invoice_line WHERE invoice_id=$1`,
		`INSERT INTO invoice_line SELECT invoice_id,2,account_id,account_name,description,quantity,unit_price_minor FROM invoice_line WHERE invoice_id=$1`,
	} {
		if _, err := db.Exec(ctx, statement, invoice.ID); err == nil {
			t.Fatalf("issued invoice mutation succeeded: %s", statement)
		}
	}
	// An incomplete invoice must not survive a commit, even if an adapter omits sealing.
	if _, err := db.Exec(ctx, `INSERT INTO invoice(patient_id,invoice_date,subtotal_minor,discount_basis_points,total_minor,created_by,request_key,request_hash) VALUES($1,CURRENT_DATE,10,0,10,'admin','incomplete-invoice','hash')`, patients[0].ID); err == nil {
		t.Fatal("incomplete invoice committed")
	}
	again, e := billing.CreateInvoice(ctx, actors[0], invoiceInput, "invoice-key-00000001")
	if e != nil || again.ID != invoice.ID {
		t.Fatal("invoice retry", e)
	}
	if _, e = billing.Invoice(ctx, actors[3], invoice.ID); e != nil {
		t.Fatal("patient cannot view own invoice", e)
	}
	if _, e = billing.Invoice(ctx, actors[1], invoice.ID); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("doctor read finance", e)
	}
	var payments [2]domain.Payment
	var payErrors [2]error
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			payments[n], payErrors[n] = billing.RecordPayment(ctx, actors[0], invoice.ID, domain.PaymentInput{AmountMinor: 899, Method: "cash"}, "payment-key-000000"+string(rune('0'+n)))
		}(n)
	}
	wg.Wait()
	payWinner := 0
	if payErrors[0] != nil {
		payWinner = 1
	}
	if payErrors[payWinner] != nil || !errors.Is(payErrors[1-payWinner], domain.ErrStale) {
		t.Fatal("concurrent overpayment", payErrors)
	}
	paid := payments[payWinner]
	retryPayment, e := billing.RecordPayment(ctx, actors[0], invoice.ID, domain.PaymentInput{AmountMinor: 899, Method: "cash"}, "payment-key-000000"+string(rune('0'+payWinner)))
	if e != nil || retryPayment.ID != paid.ID {
		t.Fatal("duplicate payment", e)
	}
	if _, e = db.Exec(ctx, `UPDATE invoice_payment SET amount_minor=1 WHERE id=$1`, paid.ID); e == nil {
		t.Fatal("payment ledger changed")
	}
	refund := domain.PaymentInput{AmountMinor: 200, Method: "cash", OriginalPaymentID: paid.ID, Reason: "Synthetic correction"}
	if _, e = billing.RecordPayment(ctx, actors[0], invoice.ID, refund, "refund-key-00000001"); e != nil {
		t.Fatal(e)
	}
	balance, e := billing.Invoice(ctx, actors[0], invoice.ID)
	if e != nil || balance.PaidMinor != 699 {
		t.Fatal("refund balance", e, balance.PaidMinor)
	}
	refund.AmountMinor = 700
	if _, e = billing.RecordPayment(ctx, actors[0], invoice.ID, refund, "refund-key-00000002"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("over-refund", e)
	}
	if _, e = billing.RecordPayment(ctx, actors[3], invoice.ID, domain.PaymentInput{AmountMinor: 1, Method: "cash"}, "patient-payment-key"); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("patient forged payment", e)
	}

	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Billing: billing, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		actor, body string
		want        int
	}{{"patient", `{"amountMinor":200,"method":"cash"}`, 403}, {"admin", `{"amountMinor":200,"method":"cash","role":"admin"}`, 400}, {"admin", `{"amountMinor":200,"method":"cash"}`, 201}} {
		req := httptest.NewRequest("POST", "/v1/invoices/"+invoice.ID+"/payments", strings.NewReader(tc.body))
		req.Header.Set("Origin", "http://hospital.test")
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Idempotency-Key", "http-payment-key-0001")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("payment HTTP wanted %d got %d: %s", tc.want, w.Code, w.Body.String())
		}
	}

}
