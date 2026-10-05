package domain

import (
	"strings"
	"time"
)

type ExpenseHead struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ExpenseHeadInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (in *ExpenseHeadInput) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" || len([]rune(in.Name)) > 160 {
		return ErrValidation
	}
	return nil
}

type Expense struct {
	ID              string    `json:"id"`
	ExpenseHeadID   string    `json:"expenseHeadId"`
	ExpenseHeadName string    `json:"expenseHeadName,omitempty"`
	Name            string    `json:"name"`
	InvoiceNumber   string    `json:"invoiceNumber"`
	Date            time.Time `json:"date"`
	AmountMinor     int64     `json:"amountMinor"`
	Description     string    `json:"description"`
	RecordedBy      string    `json:"recordedBy"`
	RecordedByName  string    `json:"recordedByName,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ExpenseInput struct {
	ExpenseHeadID string `json:"expenseHeadId"`
	Name          string `json:"name"`
	InvoiceNumber string `json:"invoiceNumber"`
	Date          string `json:"date"`
	AmountMinor   int64  `json:"amountMinor"`
	Description   string `json:"description"`
}

func (in *ExpenseInput) Validate() (time.Time, error) {
	in.ExpenseHeadID = strings.TrimSpace(in.ExpenseHeadID)
	in.Name = strings.TrimSpace(in.Name)
	in.InvoiceNumber = strings.TrimSpace(in.InvoiceNumber)
	in.Description = strings.TrimSpace(in.Description)
	if in.ExpenseHeadID == "" || in.Name == "" || in.AmountMinor <= 0 {
		return time.Time{}, ErrValidation
	}
	d, err := ParseDate(in.Date)
	if err != nil {
		return time.Time{}, ErrValidation
	}
	return d, nil
}

type IncomeHead struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type IncomeHeadInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (in *IncomeHeadInput) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" || len([]rune(in.Name)) > 160 {
		return ErrValidation
	}
	return nil
}

type Income struct {
	ID             string    `json:"id"`
	IncomeHeadID   string    `json:"incomeHeadId"`
	IncomeHeadName string    `json:"incomeHeadName,omitempty"`
	Name           string    `json:"name"`
	InvoiceNumber  string    `json:"invoiceNumber"`
	Date           time.Time `json:"date"`
	AmountMinor    int64     `json:"amountMinor"`
	Description    string    `json:"description"`
	RecordedBy     string    `json:"recordedBy"`
	RecordedByName string    `json:"recordedByName,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type IncomeInput struct {
	IncomeHeadID  string `json:"incomeHeadId"`
	Name          string `json:"name"`
	InvoiceNumber string `json:"invoiceNumber"`
	Date          string `json:"date"`
	AmountMinor   int64  `json:"amountMinor"`
	Description   string `json:"description"`
}

func (in *IncomeInput) Validate() (time.Time, error) {
	in.IncomeHeadID = strings.TrimSpace(in.IncomeHeadID)
	in.Name = strings.TrimSpace(in.Name)
	in.InvoiceNumber = strings.TrimSpace(in.InvoiceNumber)
	in.Description = strings.TrimSpace(in.Description)
	if in.IncomeHeadID == "" || in.Name == "" || in.AmountMinor <= 0 {
		return time.Time{}, ErrValidation
	}
	d, err := ParseDate(in.Date)
	if err != nil {
		return time.Time{}, ErrValidation
	}
	return d, nil
}

type EmployeePayroll struct {
	ID               string     `json:"id"`
	PayrollNumber    string     `json:"payrollNumber"`
	UserID           string     `json:"userId"`
	UserName         string     `json:"userName,omitempty"`
	UserEmail        string     `json:"userEmail,omitempty"`
	Role             string     `json:"role"`
	Month            string     `json:"month"`
	Year             int        `json:"year"`
	BasicSalaryMinor int64      `json:"basicSalaryMinor"`
	AllowanceMinor   int64      `json:"allowanceMinor"`
	DeductionsMinor  int64      `json:"deductionsMinor"`
	NetSalaryMinor   int64      `json:"netSalaryMinor"`
	Status           int        `json:"status"` // 0=unpaid, 1=paid
	PaymentDate      *time.Time `json:"paymentDate,omitempty"`
	CreatedBy        string     `json:"createdBy"`
	CreatedByName    string     `json:"createdByName,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type PayrollInput struct {
	UserID           string `json:"userId"`
	Month            string `json:"month"`
	Year             int    `json:"year"`
	BasicSalaryMinor int64  `json:"basicSalaryMinor"`
	AllowanceMinor   int64  `json:"allowanceMinor"`
	DeductionsMinor  int64  `json:"deductionsMinor"`
}

func (in *PayrollInput) Validate() error {
	in.UserID = strings.TrimSpace(in.UserID)
	in.Month = strings.TrimSpace(in.Month)
	if in.UserID == "" || in.Month == "" {
		return ErrValidation
	}
	if in.Year < 2000 || in.Year > 2100 {
		return ErrValidation
	}
	if in.BasicSalaryMinor > 100000000000 || in.AllowanceMinor > 100000000000 || in.DeductionsMinor > 100000000000 || in.BasicSalaryMinor < 0 || in.AllowanceMinor < 0 || in.DeductionsMinor < 0 {
		return ErrValidation
	}
	net := in.BasicSalaryMinor + in.AllowanceMinor - in.DeductionsMinor
	if net < 0 {
		return ErrValidation
	}
	return nil
}

type ServiceInvoiceLink struct {
	ID          string    `json:"id"`
	InvoiceID   string    `json:"invoiceId"`
	SourceType  string    `json:"sourceType"` // 'service', 'operation', 'ambulance', 'blood_issue'
	SourceID    string    `json:"sourceId"`
	AmountMinor int64     `json:"amountMinor"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ServiceInvoiceLinkInput struct {
	InvoiceID   string `json:"invoiceId"`
	SourceType  string `json:"sourceType"`
	SourceID    string `json:"sourceId"`
	AmountMinor int64  `json:"amountMinor"`
}

func (in *ServiceInvoiceLinkInput) Validate() error {
	in.InvoiceID = strings.TrimSpace(in.InvoiceID)
	in.SourceType = strings.TrimSpace(in.SourceType)
	in.SourceID = strings.TrimSpace(in.SourceID)
	if !UUIDPattern.MatchString(in.InvoiceID) || !UUIDPattern.MatchString(in.SourceID) || in.AmountMinor < 0 {
		return ErrValidation
	}
	switch in.SourceType {
	case "service", "operation", "ambulance", "blood_issue":
		return nil
	default:
		return ErrValidation
	}
}

type HeadSummary struct {
	HeadID     string `json:"headId"`
	HeadName   string `json:"headName"`
	TotalMinor int64  `json:"totalMinor"`
}

type FinanceSummary struct {
	FromDate          time.Time     `json:"fromDate"`
	ToDate            time.Time     `json:"toDate"`
	TotalIncomeMinor  int64         `json:"totalIncomeMinor"`
	TotalExpenseMinor int64         `json:"totalExpenseMinor"`
	TotalPayrollMinor int64         `json:"totalPayrollMinor"`
	NetBalanceMinor   int64         `json:"netBalanceMinor"`
	ExpensesByHead    []HeadSummary `json:"expensesByHead"`
	IncomesByHead     []HeadSummary `json:"incomesByHead"`
}

func ParseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, HospitalLocation); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, ErrValidation
}

// Operations use an explicitly approved amount because their catalogue has no price.
type PatientServiceChargeInput struct {
	PatientID   string `json:"patientId"`
	Kind        string `json:"kind"`
	CatalogID   string `json:"catalogId"`
	Quantity    int    `json:"quantity"`
	AmountMinor int64  `json:"amountMinor"`
}
type PatientServiceCharge struct {
	ID string `json:"id"`
	PatientServiceChargeInput
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}
