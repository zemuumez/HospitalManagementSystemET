package application

import (
	"context"
	"fmt"
	"hms.local/api/internal/domain"
	"math/rand"
	"time"
)

type FinancePayrollStore interface {
	Invoice(context.Context, domain.Actor, string) (domain.Invoice, error)
	CreatePatientServiceCharge(context.Context, domain.Actor, domain.PatientServiceChargeInput, string) (domain.PatientServiceCharge, error)
	// Expense Heads
	ExpenseHeads(ctx context.Context) ([]domain.ExpenseHead, error)
	CreateExpenseHead(ctx context.Context, a domain.Actor, in domain.ExpenseHeadInput) (domain.ExpenseHead, error)

	// Expenses
	Expenses(ctx context.Context, headID string, from, to *time.Time, page int) ([]domain.Expense, int, error)
	Expense(ctx context.Context, id string) (domain.Expense, error)
	CreateExpense(ctx context.Context, a domain.Actor, in domain.ExpenseInput, date time.Time) (domain.Expense, error)

	// Income Heads
	IncomeHeads(ctx context.Context) ([]domain.IncomeHead, error)
	CreateIncomeHead(ctx context.Context, a domain.Actor, in domain.IncomeHeadInput) (domain.IncomeHead, error)

	// Incomes
	Incomes(ctx context.Context, headID string, from, to *time.Time, page int) ([]domain.Income, int, error)
	Income(ctx context.Context, id string) (domain.Income, error)
	CreateIncome(ctx context.Context, a domain.Actor, in domain.IncomeInput, date time.Time) (domain.Income, error)

	// Payroll
	Payrolls(ctx context.Context, userID string, month string, year int, status *int, page int) ([]domain.EmployeePayroll, int, error)
	Payroll(ctx context.Context, id string) (domain.EmployeePayroll, error)
	CreatePayroll(ctx context.Context, a domain.Actor, in domain.PayrollInput, payrollNumber string) (domain.EmployeePayroll, error)
	PayPayroll(ctx context.Context, a domain.Actor, id string, paymentDate time.Time) (domain.EmployeePayroll, error)

	// Service Invoice Link
	CreateServiceInvoiceLink(ctx context.Context, a domain.Actor, in domain.ServiceInvoiceLinkInput) (domain.ServiceInvoiceLink, error)
	ServiceInvoiceLinks(ctx context.Context, invoiceID string) ([]domain.ServiceInvoiceLink, error)

	// Finance Summary
	FinanceSummary(ctx context.Context, from, to time.Time) (domain.FinanceSummary, error)
}

type FinancePayrollService struct {
	Store FinancePayrollStore
	Now   func() time.Time
}

func (s FinancePayrollService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

// --- Expense Heads & Expenses ---

func (s FinancePayrollService) ExpenseHeads(ctx context.Context, a domain.Actor) ([]domain.ExpenseHead, error) {
	if !a.Can("finance.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.ExpenseHeads(ctx)
}

func (s FinancePayrollService) CreateExpenseHead(ctx context.Context, a domain.Actor, in domain.ExpenseHeadInput) (domain.ExpenseHead, error) {
	if !a.Can("finance.manage") {
		return domain.ExpenseHead{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.ExpenseHead{}, err
	}
	return s.Store.CreateExpenseHead(ctx, a, in)
}

func (s FinancePayrollService) Expenses(ctx context.Context, a domain.Actor, headID string, from, to *time.Time, page int) ([]domain.Expense, int, error) {
	if !a.Can("finance.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.Expenses(ctx, headID, from, to, page)
}

func (s FinancePayrollService) Expense(ctx context.Context, a domain.Actor, id string) (domain.Expense, error) {
	if !a.Can("finance.read") {
		return domain.Expense{}, domain.ErrForbidden
	}
	return s.Store.Expense(ctx, id)
}

func (s FinancePayrollService) CreateExpense(ctx context.Context, a domain.Actor, in domain.ExpenseInput) (domain.Expense, error) {
	if !a.Can("finance.manage") {
		return domain.Expense{}, domain.ErrForbidden
	}
	d, err := in.Validate()
	if err != nil {
		return domain.Expense{}, err
	}
	return s.Store.CreateExpense(ctx, a, in, d)
}

// --- Income Heads & Incomes ---

func (s FinancePayrollService) IncomeHeads(ctx context.Context, a domain.Actor) ([]domain.IncomeHead, error) {
	if !a.Can("finance.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.IncomeHeads(ctx)
}

func (s FinancePayrollService) CreateIncomeHead(ctx context.Context, a domain.Actor, in domain.IncomeHeadInput) (domain.IncomeHead, error) {
	if !a.Can("finance.manage") {
		return domain.IncomeHead{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.IncomeHead{}, err
	}
	return s.Store.CreateIncomeHead(ctx, a, in)
}

func (s FinancePayrollService) Incomes(ctx context.Context, a domain.Actor, headID string, from, to *time.Time, page int) ([]domain.Income, int, error) {
	if !a.Can("finance.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.Incomes(ctx, headID, from, to, page)
}

func (s FinancePayrollService) Income(ctx context.Context, a domain.Actor, id string) (domain.Income, error) {
	if !a.Can("finance.read") {
		return domain.Income{}, domain.ErrForbidden
	}
	return s.Store.Income(ctx, id)
}

func (s FinancePayrollService) CreateIncome(ctx context.Context, a domain.Actor, in domain.IncomeInput) (domain.Income, error) {
	if !a.Can("finance.manage") {
		return domain.Income{}, domain.ErrForbidden
	}
	d, err := in.Validate()
	if err != nil {
		return domain.Income{}, err
	}
	return s.Store.CreateIncome(ctx, a, in, d)
}

// --- Payroll ---

func (s FinancePayrollService) Payrolls(ctx context.Context, a domain.Actor, userID string, month string, year int, status *int, page int) ([]domain.EmployeePayroll, int, error) {
	if a.Can("payroll.read") {
		return s.Store.Payrolls(ctx, userID, month, year, status, page)
	}
	if a.Can("payroll.read_own") {
		if userID != "" && userID != a.ID {
			return nil, 0, domain.ErrForbidden
		}
		return s.Store.Payrolls(ctx, a.ID, month, year, status, page)
	}
	return nil, 0, domain.ErrForbidden
}

func (s FinancePayrollService) Payroll(ctx context.Context, a domain.Actor, id string) (domain.EmployeePayroll, error) {
	p, err := s.Store.Payroll(ctx, id)
	if err != nil {
		return domain.EmployeePayroll{}, err
	}
	if a.Can("payroll.read") {
		return p, nil
	}
	if a.Can("payroll.read_own") && p.UserID == a.ID {
		return p, nil
	}
	return domain.EmployeePayroll{}, domain.ErrForbidden
}

func (s FinancePayrollService) CreatePayroll(ctx context.Context, a domain.Actor, in domain.PayrollInput) (domain.EmployeePayroll, error) {
	if !a.Can("payroll.manage") {
		return domain.EmployeePayroll{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.EmployeePayroll{}, err
	}
	now := s.now()
	payrollNumber := fmt.Sprintf("PAY-%d-%04d", now.Unix()%1000000, rand.Intn(10000))
	return s.Store.CreatePayroll(ctx, a, in, payrollNumber)
}

func (s FinancePayrollService) PayPayroll(ctx context.Context, a domain.Actor, id string) (domain.EmployeePayroll, error) {
	if !a.Can("payroll.manage") {
		return domain.EmployeePayroll{}, domain.ErrForbidden
	}
	return s.Store.PayPayroll(ctx, a, id, s.now())
}

// --- Service Invoice Link ---

func (s FinancePayrollService) CreateServiceInvoiceLink(ctx context.Context, a domain.Actor, in domain.ServiceInvoiceLinkInput) (domain.ServiceInvoiceLink, error) {
	if !a.Can("billing.manage") {
		return domain.ServiceInvoiceLink{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.ServiceInvoiceLink{}, err
	}
	return s.Store.CreateServiceInvoiceLink(ctx, a, in)
}

func (s FinancePayrollService) ServiceInvoiceLinks(ctx context.Context, a domain.Actor, invoiceID string) ([]domain.ServiceInvoiceLink, error) {
	if !a.Can("billing.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(invoiceID) {
		return nil, domain.ErrValidation
	}
	if _, err := s.Store.Invoice(ctx, a, invoiceID); err != nil {
		return nil, err
	}
	return s.Store.ServiceInvoiceLinks(ctx, invoiceID)
}

// --- Finance Summary ---

func (s FinancePayrollService) FinanceSummary(ctx context.Context, a domain.Actor, fromStr, toStr string) (domain.FinanceSummary, error) {
	if !a.Can("finance.read") {
		return domain.FinanceSummary{}, domain.ErrForbidden
	}
	from, err := domain.ParseDate(fromStr)
	if err != nil {
		return domain.FinanceSummary{}, domain.ErrValidation
	}
	to, err := domain.ParseDate(toStr)
	if err != nil {
		return domain.FinanceSummary{}, domain.ErrValidation
	}
	if to.Before(from) {
		return domain.FinanceSummary{}, domain.ErrValidation
	}
	return s.Store.FinanceSummary(ctx, from, to)
}

func (s FinancePayrollService) CreatePatientServiceCharge(ctx context.Context, a domain.Actor, in domain.PatientServiceChargeInput, key string) (domain.PatientServiceCharge, error) {
	if !a.Can("billing.manage") {
		return domain.PatientServiceCharge{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(in.PatientID) || !domain.UUIDPattern.MatchString(in.CatalogID) || (in.Kind != "service" && in.Kind != "operation") || in.Quantity < 1 || in.Quantity > 10000 || in.AmountMinor < 0 || in.AmountMinor > 100000000000 || len(key) < 8 || len(key) > 160 {
		return domain.PatientServiceCharge{}, domain.ErrValidation
	}
	return s.Store.CreatePatientServiceCharge(ctx, a, in, key)
}
