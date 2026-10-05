package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"testing"
	"time"
)

func testOperationalBilling(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	ctx := context.Background()
	admin := actors[0]
	billing := application.Billing{Store: store, Now: time.Now}
	ops := application.AppointmentOpsService{Store: store}
	finance := application.FinancePayrollService{Store: store}
	account, err := billing.CreateAccount(ctx, admin, "Verified operational charges")
	if err != nil {
		t.Fatal(err)
	}
	makeInvoice := func(patient string, key string) domain.Invoice {
		t.Helper()
		v, e := billing.CreateInvoice(ctx, admin, domain.InvoiceInput{PatientID: patient, InvoiceDate: time.Now().In(domain.HospitalLocation).Format("2006-01-02"), Lines: []domain.InvoiceLine{{AccountID: account.ID, Quantity: 1, UnitPriceMinor: 10000}}}, key)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	var appt string
	err = db.QueryRow(ctx, `INSERT INTO appointment(patient_id,doctor_id,starts_at,ends_at,created_by,request_key) VALUES($1,$2,'2030-01-01T08:00:00Z','2030-01-01T09:00:00Z',$3,'operational-billing-test') RETURNING id`, patients[0].ID, actors[1].ID, admin.ID).Scan(&appt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ops.SetAppointmentFee(ctx, admin, appt, 10000); err != nil {
		t.Fatal(err)
	}
	wrong := makeInvoice(patients[1].ID, "wrong-operational-invoice")
	if _, err = ops.LinkAppointmentInvoice(ctx, admin, appt, wrong.ID); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("cross patient invoice linked", err)
	}
	inv := makeInvoice(patients[0].ID, "valid-operational-invoice")
	rec, err := ops.LinkAppointmentInvoice(ctx, admin, appt, inv.ID)
	if err != nil || rec.PaymentStatus != "unpaid" {
		t.Fatal("link invented payment", rec, err)
	}
	if _, err = ops.SetAppointmentFee(ctx, admin, appt, 5); err == nil {
		t.Fatal("linked fee changed")
	}
	other := makeInvoice(patients[0].ID, "other-operational-invoice")
	if _, err = ops.LinkAppointmentInvoice(ctx, admin, appt, other.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("source linked twice", err)
	}
	if _, err = ops.LinkAppointmentInvoice(ctx, admin, appt, inv.ID); err != nil {
		t.Fatal("link retry failed", err)
	}
	payment, err := billing.RecordPayment(ctx, admin, inv.ID, domain.PaymentInput{AmountMinor: 10000, Method: "cash"}, "operational-full-payment")
	if err != nil {
		t.Fatal(err)
	}
	rec, err = ops.GetAppointmentBilling(ctx, admin, appt)
	if err != nil || rec.PaymentStatus != "paid" {
		t.Fatal("actual payment not reflected", err)
	}
	if _, err = billing.RecordPayment(ctx, admin, inv.ID, domain.PaymentInput{AmountMinor: 1000, Method: "cash", OriginalPaymentID: payment.ID, Reason: "Synthetic refund"}, "operational-test-refund"); err != nil {
		t.Fatal(err)
	}
	rec, err = ops.GetAppointmentBilling(ctx, admin, appt)
	if err != nil || rec.PaymentStatus != "unpaid" {
		t.Fatal("refund not reflected", err)
	}
	if _, err = finance.ServiceInvoiceLinks(ctx, actors[3], wrong.ID); err == nil {
		t.Fatal("other patient financial sources disclosed")
	}
	var enc string
	if err = db.QueryRow(ctx, `SELECT id FROM encounter WHERE patient_id=$1 LIMIT 1`, patients[0].ID).Scan(&enc); err != nil {
		t.Fatal(err)
	}
	if _, err = store.UpdateEncounterBilling(ctx, admin, enc, domain.UpdateEncounterBillingInput{DoctorFeeMinor: 10000}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.LinkEncounterInvoice(ctx, admin, enc, other.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GrantFinancialClearance(ctx, admin, enc, ""); !errors.Is(err, domain.ErrStale) {
		t.Fatal("unpaid clearance", err)
	}
	cleared, err := store.GrantFinancialClearance(ctx, admin, enc, "Authorized synthetic financial assistance")
	if err != nil || !cleared.FinancialClearance {
		t.Fatal(err)
	}
	if _, err = store.UpdateEncounterBilling(ctx, admin, enc, domain.UpdateEncounterBillingInput{DoctorFeeMinor: 1}); err == nil {
		t.Fatal("cleared charges changed")
	}
	summary := domain.DischargeSummary{EncounterID: enc, AdmissionDiagnosis: "Synthetic", DischargeDiagnosis: "Synthetic", ConditionAtDischarge: "improved"}
	if err = store.SaveDischargeSummary(ctx, actors[1], summary); err != nil {
		t.Fatal(err)
	}
	if err = store.SaveDischargeSummary(ctx, actors[1], summary); err == nil {
		t.Fatal("signed summary overwritten")
	}
	if _, err = db.Exec(ctx, `DELETE FROM discharge_summary WHERE encounter_id=$1`, enc); err == nil {
		t.Fatal("signed summary deleted")
	}
	var catalog string
	if err = db.QueryRow(ctx, `INSERT INTO hospital_service(name,rate_minor) VALUES('Synthetic delivery',5000) RETURNING id`).Scan(&catalog); err != nil {
		t.Fatal(err)
	}
	in := domain.PatientServiceChargeInput{PatientID: patients[0].ID, Kind: "service", CatalogID: catalog, Quantity: 2}
	charge, err := finance.CreatePatientServiceCharge(ctx, admin, in, "delivered-service-01")
	if err != nil || charge.AmountMinor != 10000 {
		t.Fatal("server tariff", err)
	}
	again, err := finance.CreatePatientServiceCharge(ctx, admin, in, "delivered-service-01")
	if err != nil || again.ID != charge.ID {
		t.Fatal("duplicate delivered service", err)
	}
	in.Quantity = 3
	if _, err = finance.CreatePatientServiceCharge(ctx, admin, in, "delivered-service-01"); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("changed retry", err)
	}
	serviceInv := makeInvoice(patients[0].ID, "service-invoice-01")
	if _, err = finance.CreateServiceInvoiceLink(ctx, admin, domain.ServiceInvoiceLinkInput{InvoiceID: serviceInv.ID, SourceType: "service", SourceID: charge.ID, AmountMinor: 10000}); err != nil {
		t.Fatal(err)
	}
	if _, err = finance.CreatePayroll(ctx, admin, domain.PayrollInput{UserID: actors[3].ID, Month: "October", Year: 2026, BasicSalaryMinor: 100}); err == nil {
		t.Fatal("patient payroll accepted")
	}
}
