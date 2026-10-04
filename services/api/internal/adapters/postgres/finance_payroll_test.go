package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testFinancePayroll(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()

	admin := actors[0]
	doctor := actors[1]
	otherDoctor := actors[2]
	patientUser := actors[3]

	srv := application.FinancePayrollService{Store: store, Now: time.Now}

	// 1. Authorization checks
	// Patient cannot read expense heads
	_, err := srv.ExpenseHeads(ctx, patientUser)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient reading expense heads, got %v", err)
	}

	// Doctor cannot create expense head
	_, err = srv.CreateExpenseHead(ctx, doctor, domain.ExpenseHeadInput{Name: "Facility Rent"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor creating expense head, got %v", err)
	}

	// Doctor cannot create expense
	_, err = srv.CreateExpense(ctx, doctor, domain.ExpenseInput{
		ExpenseHeadID: "dummy",
		Name:          "Generator Fuel",
		AmountMinor:   5000,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor creating expense, got %v", err)
	}

	// Patient cannot create income head
	_, err = srv.CreateIncomeHead(ctx, patientUser, domain.IncomeHeadInput{Name: "Donation"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating income head, got %v", err)
	}

	// Doctor cannot create payroll
	_, err = srv.CreatePayroll(ctx, doctor, domain.PayrollInput{
		UserID:           doctor.ID,
		Month:            "October",
		Year:             2026,
		BasicSalaryMinor: 100000,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor creating payroll, got %v", err)
	}

	// Patient cannot create service invoice link
	_, err = srv.CreateServiceInvoiceLink(ctx, patientUser, domain.ServiceInvoiceLinkInput{
		InvoiceID:   "dummy",
		SourceType:  "ambulance",
		SourceID:    "amb-1",
		AmountMinor: 1000,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating service invoice link, got %v", err)
	}

	// 2. Validation checks
	// Empty expense head name
	_, err = srv.CreateExpenseHead(ctx, admin, domain.ExpenseHeadInput{Name: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty expense head name, got %v", err)
	}

	// Expense with zero or negative amount
	_, err = srv.CreateExpense(ctx, admin, domain.ExpenseInput{
		ExpenseHeadID: "dummy",
		Name:          "Test Exp",
		AmountMinor:   0,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for zero expense amount, got %v", err)
	}

	// Payroll with negative allowance or invalid year
	_, err = srv.CreatePayroll(ctx, admin, domain.PayrollInput{
		UserID:           doctor.ID,
		Month:            "October",
		Year:             1990,
		BasicSalaryMinor: 1000,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for year < 2000, got %v", err)
	}

	_, err = srv.CreatePayroll(ctx, admin, domain.PayrollInput{
		UserID:           doctor.ID,
		Month:            "October",
		Year:             2026,
		BasicSalaryMinor: 1000,
		DeductionsMinor:  5000, // net < 0!
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for negative net salary, got %v", err)
	}

	// 3. Expense Heads and Expenses CRUD
	expHead, err := srv.CreateExpenseHead(ctx, admin, domain.ExpenseHeadInput{
		Name:        "Facility Operations & Maintenance",
		Description: "General facility maintenance and repairs",
	})
	if err != nil {
		t.Fatalf("failed to create expense head: %v", err)
	}

	// Duplicate expense head conflict
	_, err = srv.CreateExpenseHead(ctx, admin, domain.ExpenseHeadInput{
		Name: "Facility Operations & Maintenance",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate expense head name, got %v", err)
	}

	heads, err := srv.ExpenseHeads(ctx, admin)
	if err != nil || len(heads) == 0 {
		t.Fatalf("failed to list expense heads: %v", err)
	}

	// Create expense
	expenseDate := "2026-10-01"
	exp, err := srv.CreateExpense(ctx, admin, domain.ExpenseInput{
		ExpenseHeadID: expHead.ID,
		Name:          "Air Conditioning Servicing",
		InvoiceNumber: "INV-EXP-001",
		Date:          expenseDate,
		AmountMinor:   450000, // 4,500 ETB
		Description:   "Comprehensive HVAC unit overhaul",
	})
	if err != nil || exp.ID == "" {
		t.Fatalf("failed to create expense: %v", err)
	}

	expenses, totalExp, err := srv.Expenses(ctx, admin, expHead.ID, nil, nil, 1)
	if err != nil || totalExp == 0 || len(expenses) == 0 {
		t.Fatalf("failed to list expenses: %v, total: %d", err, totalExp)
	}

	singleExp, err := srv.Expense(ctx, admin, exp.ID)
	if err != nil || singleExp.Name != "Air Conditioning Servicing" {
		t.Fatalf("failed to get single expense: %v", err)
	}

	// 4. Income Heads and Incomes CRUD
	incHead, err := srv.CreateIncomeHead(ctx, admin, domain.IncomeHeadInput{
		Name:        "Executive Clinical Consultations",
		Description: "VIP and specialized consult fees",
	})
	if err != nil {
		t.Fatalf("failed to create income head: %v", err)
	}

	// Duplicate income head conflict
	_, err = srv.CreateIncomeHead(ctx, admin, domain.IncomeHeadInput{
		Name: "Executive Clinical Consultations",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate income head name, got %v", err)
	}

	incHeads, err := srv.IncomeHeads(ctx, admin)
	if err != nil || len(incHeads) == 0 {
		t.Fatalf("failed to list income heads: %v", err)
	}

	inc, err := srv.CreateIncome(ctx, admin, domain.IncomeInput{
		IncomeHeadID:  incHead.ID,
		Name:          "Executive Checkup Package",
		InvoiceNumber: "INV-INC-001",
		Date:          "2026-10-02",
		AmountMinor:   950000, // 9,500 ETB
		Description:   "Full health screening package",
	})
	if err != nil || inc.ID == "" {
		t.Fatalf("failed to create income: %v", err)
	}

	incomes, totalInc, err := srv.Incomes(ctx, admin, incHead.ID, nil, nil, 1)
	if err != nil || totalInc == 0 || len(incomes) == 0 {
		t.Fatalf("failed to list incomes: %v, total: %d", err, totalInc)
	}

	singleInc, err := srv.Income(ctx, admin, inc.ID)
	if err != nil || singleInc.Name != "Executive Checkup Package" {
		t.Fatalf("failed to get single income: %v", err)
	}

	// 5. Employee Payroll Lifecycle & Scoped Authorization
	payroll, err := srv.CreatePayroll(ctx, admin, domain.PayrollInput{
		UserID:           doctor.ID,
		Month:            "October",
		Year:             2026,
		BasicSalaryMinor: 3000000, // 30,000 ETB
		AllowanceMinor:   500000,  // 5,000 ETB
		DeductionsMinor:  200000,  // 2,000 ETB
	})
	if err != nil || payroll.ID == "" {
		t.Fatalf("failed to create employee payroll: %v", err)
	}
	if payroll.NetSalaryMinor != 3300000 {
		t.Fatalf("expected net salary 3300000, got %d", payroll.NetSalaryMinor)
	}
	if payroll.Status != 0 {
		t.Fatalf("expected initial status 0 (unpaid), got %d", payroll.Status)
	}

	// Duplicate payroll for same user, month, year must conflict
	_, err = srv.CreatePayroll(ctx, admin, domain.PayrollInput{
		UserID:           doctor.ID,
		Month:            "October",
		Year:             2026,
		BasicSalaryMinor: 3000000,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate payroll period, got %v", err)
	}

	// Doctor can view own payroll
	docPayrolls, dTotal, err := srv.Payrolls(ctx, doctor, doctor.ID, "", 0, nil, 1)
	if err != nil || dTotal == 0 || len(docPayrolls) == 0 {
		t.Fatalf("doctor should view own payrolls: %v, total: %d", err, dTotal)
	}

	// Doctor cannot view other doctor's payroll
	_, _, err = srv.Payrolls(ctx, doctor, otherDoctor.ID, "", 0, nil, 1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor querying other doctor payrolls, got %v", err)
	}

	// Other doctor cannot view doctor's single payroll slip
	_, err = srv.Payroll(ctx, otherDoctor, payroll.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for other doctor reading payroll slip, got %v", err)
	}

	// Doctor can view own single payroll slip
	docSlip, err := srv.Payroll(ctx, doctor, payroll.ID)
	if err != nil || docSlip.ID != payroll.ID {
		t.Fatalf("doctor failed to read own payroll slip: %v", err)
	}

	// Admin pays the payroll
	paidPayroll, err := srv.PayPayroll(ctx, admin, payroll.ID)
	if err != nil || paidPayroll.Status != 1 || paidPayroll.PaymentDate == nil {
		t.Fatalf("failed to pay payroll: %v, status: %d", err, paidPayroll.Status)
	}

	// Duplicate payment attempt on already-paid payroll must return conflict
	_, err = srv.PayPayroll(ctx, admin, payroll.ID)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on paying already paid payroll, got %v", err)
	}

	// 6. Anti-Double-Billing Service Invoice Linkages
	// Create an invoice first using billing service
	billing := application.Billing{Store: store, Now: time.Now}
	bAccount, err := billing.CreateAccount(ctx, admin, "Operational Linkages Account")
	if err != nil {
		t.Fatalf("failed to create billing account: %v", err)
	}
	inv, err := billing.CreateInvoice(ctx, admin, domain.InvoiceInput{
		PatientID:           patients[0].ID,
		InvoiceDate:         "2026-10-05",
		DiscountBasisPoints: 0,
		Lines: []domain.InvoiceLine{
			{AccountID: bAccount.ID, Quantity: 1, UnitPriceMinor: 50000},
		},
	}, "test-finance-invoice-key-01")
	if err != nil {
		t.Fatalf("failed to create invoice for linkage: %v", err)
	}

	// Create service invoice link for ambulance call
	link1, err := srv.CreateServiceInvoiceLink(ctx, admin, domain.ServiceInvoiceLinkInput{
		InvoiceID:   inv.ID,
		SourceType:  "ambulance",
		SourceID:    "amb-call-oct-01",
		AmountMinor: 50000,
	})
	if err != nil || link1.ID == "" {
		t.Fatalf("failed to create service invoice link: %v", err)
	}

	// Duplicate source link (same source_type and source_id) MUST conflict (anti-double-billing)
	_, err = srv.CreateServiceInvoiceLink(ctx, admin, domain.ServiceInvoiceLinkInput{
		InvoiceID:   inv.ID,
		SourceType:  "ambulance",
		SourceID:    "amb-call-oct-01",
		AmountMinor: 50000,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate source charge billing, got %v", err)
	}

	// Query invoice links
	links, err := srv.ServiceInvoiceLinks(ctx, admin, inv.ID)
	if err != nil || len(links) != 1 || links[0].SourceID != "amb-call-oct-01" {
		t.Fatalf("failed to get service invoice links: %v", err)
	}

	// 7. Finance Summary Report
	summary, err := srv.FinanceSummary(ctx, admin, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("failed to generate finance summary: %v", err)
	}
	if summary.TotalIncomeMinor != 950000 {
		t.Fatalf("expected total income 950000, got %d", summary.TotalIncomeMinor)
	}
	if summary.TotalExpenseMinor != 450000 {
		t.Fatalf("expected total expense 450000, got %d", summary.TotalExpenseMinor)
	}
	if summary.TotalPayrollMinor != 3300000 {
		t.Fatalf("expected total payroll 3300000, got %d", summary.TotalPayrollMinor)
	}
	expectedBalance := int64(950000 - (450000 + 3300000)) // -2,800,000 minor ETB
	if summary.NetBalanceMinor != expectedBalance {
		t.Fatalf("expected net balance %d, got %d", expectedBalance, summary.NetBalanceMinor)
	}
	if len(summary.ExpensesByHead) == 0 || len(summary.IncomesByHead) == 0 {
		t.Fatalf("expected expense and income breakdown in summary")
	}

	// 8. HTTP API Handlers
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")
		uid := admin.ID
		if strings.Contains(cookie, patientUser.ID) {
			uid = patientUser.ID
		} else if strings.Contains(cookie, doctor.ID) {
			uid = doctor.ID
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id": uid,
			},
			"session": map[string]any{
				"userId":    uid,
				"expiresAt": time.Now().Add(time.Hour),
			},
		})
	}))
	defer authSrv.Close()

	handler := httpapi.Server{
		FinancePayroll: srv,
		Actors:         store,
		AuthURL:        authSrv.URL,
		Origin:         "http://hospital.test",
		Client:         authSrv.Client(),
	}.Handler()

	// Test GET /v1/expense-heads
	req := httptest.NewRequest("GET", "/v1/expense-heads", nil)
	req.Header.Set("Cookie", "session="+admin.ID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /v1/expense-heads wanted 200, got %d: %s", w.Code, w.Body.String())
	}
	var headsResp struct {
		Heads []domain.ExpenseHead `json:"heads"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &headsResp); err != nil || len(headsResp.Heads) == 0 {
		t.Fatalf("failed to unmarshal expense heads: %v", err)
	}

	// Test GET /v1/payrolls
	req = httptest.NewRequest("GET", "/v1/payrolls?month=October&year=2026", nil)
	req.Header.Set("Cookie", "session="+admin.ID)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /v1/payrolls wanted 200, got %d: %s", w.Code, w.Body.String())
	}

	// Test GET /v1/finance-reports/summary
	req = httptest.NewRequest("GET", "/v1/finance-reports/summary?from=2026-10-01&to=2026-10-31", nil)
	req.Header.Set("Cookie", "session="+admin.ID)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /v1/finance-reports/summary wanted 200, got %d: %s", w.Code, w.Body.String())
	}
	var sumResp domain.FinanceSummary
	if err := json.Unmarshal(w.Body.Bytes(), &sumResp); err != nil || sumResp.TotalIncomeMinor == 0 {
		t.Fatalf("failed to unmarshal finance summary: %v", err)
	}

	// Test 403 on Patient querying summary
	req = httptest.NewRequest("GET", "/v1/finance-reports/summary", nil)
	req.Header.Set("Cookie", "session="+patientUser.ID)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on patient reading finance summary, got %d: %s", w.Code, w.Body.String())
	}

	// Test POST /v1/expenses invalid JSON (400)
	req = httptest.NewRequest("POST", "/v1/expenses", strings.NewReader(`{invalid}`))
	req.Header.Set("Cookie", "session="+admin.ID)
	req.Header.Set("Origin", "http://hospital.test")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on malformed JSON, got %d", w.Code)
	}
}
