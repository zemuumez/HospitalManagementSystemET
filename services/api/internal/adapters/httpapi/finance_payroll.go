package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s Server) financePayroll(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// 1. Expense Heads
	if strings.HasPrefix(r.URL.Path, "/v1/expense-heads") {
		switch {
		case r.URL.Path == "/v1/expense-heads" && r.Method == "GET":
			heads, err := s.FinancePayroll.ExpenseHeads(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"heads": heads})
			return true

		case r.URL.Path == "/v1/expense-heads" && r.Method == "POST":
			var in domain.ExpenseHeadInput
			if !decode(w, r, &in) {
				return true
			}
			head, err := s.FinancePayroll.CreateExpenseHead(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, head)
			return true
		}
	}

	// 2. Expenses
	if strings.HasPrefix(r.URL.Path, "/v1/expenses") {
		switch {
		case r.URL.Path == "/v1/expenses" && r.Method == "GET":
			headID := r.URL.Query().Get("headId")
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			var fromPtr, toPtr *time.Time
			if fromStr := r.URL.Query().Get("from"); fromStr != "" {
				if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
					fromPtr = &t
				} else if t, err := time.Parse("2006-01-02", fromStr); err == nil {
					fromPtr = &t
				}
			}
			if toStr := r.URL.Query().Get("to"); toStr != "" {
				if t, err := time.Parse(time.RFC3339, toStr); err == nil {
					toPtr = &t
				} else if t, err := time.Parse("2006-01-02", toStr); err == nil {
					toPtr = &t
				}
			}
			list, total, err := s.FinancePayroll.Expenses(r.Context(), a, headID, fromPtr, toPtr, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"expenses": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/expenses" && r.Method == "POST":
			var in domain.ExpenseInput
			if !decode(w, r, &in) {
				return true
			}
			exp, err := s.FinancePayroll.CreateExpense(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, exp)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/expenses/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/expenses/")
			exp, err := s.FinancePayroll.Expense(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, exp)
			return true
		}
	}

	// 3. Income Heads
	if strings.HasPrefix(r.URL.Path, "/v1/income-heads") {
		switch {
		case r.URL.Path == "/v1/income-heads" && r.Method == "GET":
			heads, err := s.FinancePayroll.IncomeHeads(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"heads": heads})
			return true

		case r.URL.Path == "/v1/income-heads" && r.Method == "POST":
			var in domain.IncomeHeadInput
			if !decode(w, r, &in) {
				return true
			}
			head, err := s.FinancePayroll.CreateIncomeHead(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, head)
			return true
		}
	}

	// 4. Incomes
	if strings.HasPrefix(r.URL.Path, "/v1/incomes") {
		switch {
		case r.URL.Path == "/v1/incomes" && r.Method == "GET":
			headID := r.URL.Query().Get("headId")
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			var fromPtr, toPtr *time.Time
			if fromStr := r.URL.Query().Get("from"); fromStr != "" {
				if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
					fromPtr = &t
				} else if t, err := time.Parse("2006-01-02", fromStr); err == nil {
					fromPtr = &t
				}
			}
			if toStr := r.URL.Query().Get("to"); toStr != "" {
				if t, err := time.Parse(time.RFC3339, toStr); err == nil {
					toPtr = &t
				} else if t, err := time.Parse("2006-01-02", toStr); err == nil {
					toPtr = &t
				}
			}
			list, total, err := s.FinancePayroll.Incomes(r.Context(), a, headID, fromPtr, toPtr, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"incomes": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/incomes" && r.Method == "POST":
			var in domain.IncomeInput
			if !decode(w, r, &in) {
				return true
			}
			inc, err := s.FinancePayroll.CreateIncome(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, inc)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/incomes/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/incomes/")
			inc, err := s.FinancePayroll.Income(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, inc)
			return true
		}
	}

	// 5. Payroll
	if strings.HasPrefix(r.URL.Path, "/v1/payrolls") {
		switch {
		case r.URL.Path == "/v1/payrolls" && r.Method == "GET":
			userID := r.URL.Query().Get("userId")
			month := r.URL.Query().Get("month")
			year, _ := strconv.Atoi(r.URL.Query().Get("year"))
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			var statusPtr *int
			if statusStr := r.URL.Query().Get("status"); statusStr != "" {
				if st, err := strconv.Atoi(statusStr); err == nil {
					statusPtr = &st
				}
			}
			list, total, err := s.FinancePayroll.Payrolls(r.Context(), a, userID, month, year, statusPtr, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"payrolls": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/payrolls" && r.Method == "POST":
			var in domain.PayrollInput
			if !decode(w, r, &in) {
				return true
			}
			p, err := s.FinancePayroll.CreatePayroll(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, p)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/payrolls/") && strings.HasSuffix(r.URL.Path, "/pay") && r.Method == "POST":
			trimmed := strings.TrimPrefix(r.URL.Path, "/v1/payrolls/")
			id := strings.TrimSuffix(trimmed, "/pay")
			p, err := s.FinancePayroll.PayPayroll(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, p)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/payrolls/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/payrolls/")
			p, err := s.FinancePayroll.Payroll(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, p)
			return true
		}
	}

	// 6. Service Invoice Links (Anti-double-billing)
	if strings.HasPrefix(r.URL.Path, "/v1/service-invoice-links") {
		switch {
		case r.URL.Path == "/v1/service-invoice-links" && r.Method == "GET":
			invoiceID := r.URL.Query().Get("invoiceId")
			links, err := s.FinancePayroll.ServiceInvoiceLinks(r.Context(), a, invoiceID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"links": links})
			return true

		case r.URL.Path == "/v1/service-invoice-links" && r.Method == "POST":
			var in domain.ServiceInvoiceLinkInput
			if !decode(w, r, &in) {
				return true
			}
			link, err := s.FinancePayroll.CreateServiceInvoiceLink(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, link)
			return true
		}
	}

	// 7. Finance Summary Report
	if r.URL.Path == "/v1/finance-reports/summary" && r.Method == "GET" {
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		sum, err := s.FinancePayroll.FinanceSummary(r.Context(), a, fromStr, toStr)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, sum)
		return true
	}

	return false
}
