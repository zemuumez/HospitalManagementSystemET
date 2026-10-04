package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

// --- Expense Heads ---

func (s Store) ExpenseHeads(ctx context.Context) ([]domain.ExpenseHead, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM hospital_expense_head
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ExpenseHead{}
	for rows.Next() {
		var h domain.ExpenseHead
		if err := rows.Scan(&h.ID, &h.Name, &h.Description, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

func (s Store) CreateExpenseHead(ctx context.Context, a domain.Actor, in domain.ExpenseHeadInput) (domain.ExpenseHead, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.ExpenseHead{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.ExpenseHead
	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_expense_head (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, in.Name, in.Description).Scan(&out.ID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'expense_head.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Expenses ---

func (s Store) Expenses(ctx context.Context, headID string, from, to *time.Time, page int) ([]domain.Expense, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	var whereClauses []string
	var args []any
	argIdx := 1

	if headID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.expense_head_id = $%d", argIdx))
		args = append(args, headID)
		argIdx++
	}
	if from != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.date >= $%d", argIdx))
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.date <= $%d", argIdx))
		args = append(args, *to)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_expense e"+whereSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.expense_head_id, h.name, e.name, e.invoice_number, e.date, e.amount_minor,
		       e.description, e.recorded_by, COALESCE(u.name, ''), e.created_at, e.updated_at
		FROM hospital_expense e
		JOIN hospital_expense_head h ON e.expense_head_id = h.id
		JOIN "user" u ON e.recorded_by = u.id
		%s
		ORDER BY e.date DESC, e.created_at DESC
		LIMIT %d OFFSET %d
	`, whereSQL, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Expense{}
	for rows.Next() {
		var e domain.Expense
		if err := rows.Scan(
			&e.ID, &e.ExpenseHeadID, &e.ExpenseHeadName, &e.Name, &e.InvoiceNumber,
			&e.Date, &e.AmountMinor, &e.Description, &e.RecordedBy, &e.RecordedByName,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, nil
}

func (s Store) Expense(ctx context.Context, id string) (domain.Expense, error) {
	var e domain.Expense
	err := s.DB.QueryRow(ctx, `
		SELECT e.id, e.expense_head_id, h.name, e.name, e.invoice_number, e.date, e.amount_minor,
		       e.description, e.recorded_by, COALESCE(u.name, ''), e.created_at, e.updated_at
		FROM hospital_expense e
		JOIN hospital_expense_head h ON e.expense_head_id = h.id
		JOIN "user" u ON e.recorded_by = u.id
		WHERE e.id = $1
	`, id).Scan(
		&e.ID, &e.ExpenseHeadID, &e.ExpenseHeadName, &e.Name, &e.InvoiceNumber,
		&e.Date, &e.AmountMinor, &e.Description, &e.RecordedBy, &e.RecordedByName,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Expense{}, domain.ErrNotFound
	}
	return e, err
}

func (s Store) CreateExpense(ctx context.Context, a domain.Actor, in domain.ExpenseInput, date time.Time) (domain.Expense, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Expense{}, err
	}
	defer tx.Rollback(ctx)

	var headName string
	err = tx.QueryRow(ctx, `SELECT name FROM hospital_expense_head WHERE id = $1`, in.ExpenseHeadID).Scan(&headName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Expense{}, domain.ErrValidation
	}
	if err != nil {
		return domain.Expense{}, err
	}

	var out domain.Expense
	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_expense (
			expense_head_id, name, invoice_number, date, amount_minor, description, recorded_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, expense_head_id, name, invoice_number, date, amount_minor, description, recorded_by, created_at, updated_at
	`, in.ExpenseHeadID, in.Name, in.InvoiceNumber, date, in.AmountMinor, in.Description, a.ID).Scan(
		&out.ID, &out.ExpenseHeadID, &out.Name, &out.InvoiceNumber,
		&out.Date, &out.AmountMinor, &out.Description, &out.RecordedBy,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return out, err
	}
	out.ExpenseHeadName = headName
	out.RecordedByName = a.Name

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'expense.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Income Heads ---

func (s Store) IncomeHeads(ctx context.Context) ([]domain.IncomeHead, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM hospital_income_head
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.IncomeHead{}
	for rows.Next() {
		var h domain.IncomeHead
		if err := rows.Scan(&h.ID, &h.Name, &h.Description, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

func (s Store) CreateIncomeHead(ctx context.Context, a domain.Actor, in domain.IncomeHeadInput) (domain.IncomeHead, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.IncomeHead{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.IncomeHead
	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_income_head (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, in.Name, in.Description).Scan(&out.ID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'income_head.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Incomes ---

func (s Store) Incomes(ctx context.Context, headID string, from, to *time.Time, page int) ([]domain.Income, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	var whereClauses []string
	var args []any
	argIdx := 1

	if headID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("i.income_head_id = $%d", argIdx))
		args = append(args, headID)
		argIdx++
	}
	if from != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("i.date >= $%d", argIdx))
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("i.date <= $%d", argIdx))
		args = append(args, *to)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_income i"+whereSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT i.id, i.income_head_id, h.name, i.name, i.invoice_number, i.date, i.amount_minor,
		       i.description, i.recorded_by, COALESCE(u.name, ''), i.created_at, i.updated_at
		FROM hospital_income i
		JOIN hospital_income_head h ON i.income_head_id = h.id
		JOIN "user" u ON i.recorded_by = u.id
		%s
		ORDER BY i.date DESC, i.created_at DESC
		LIMIT %d OFFSET %d
	`, whereSQL, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Income{}
	for rows.Next() {
		var inc domain.Income
		if err := rows.Scan(
			&inc.ID, &inc.IncomeHeadID, &inc.IncomeHeadName, &inc.Name, &inc.InvoiceNumber,
			&inc.Date, &inc.AmountMinor, &inc.Description, &inc.RecordedBy, &inc.RecordedByName,
			&inc.CreatedAt, &inc.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, inc)
	}
	return out, total, nil
}

func (s Store) Income(ctx context.Context, id string) (domain.Income, error) {
	var inc domain.Income
	err := s.DB.QueryRow(ctx, `
		SELECT i.id, i.income_head_id, h.name, i.name, i.invoice_number, i.date, i.amount_minor,
		       i.description, i.recorded_by, COALESCE(u.name, ''), i.created_at, i.updated_at
		FROM hospital_income i
		JOIN hospital_income_head h ON i.income_head_id = h.id
		JOIN "user" u ON i.recorded_by = u.id
		WHERE i.id = $1
	`, id).Scan(
		&inc.ID, &inc.IncomeHeadID, &inc.IncomeHeadName, &inc.Name, &inc.InvoiceNumber,
		&inc.Date, &inc.AmountMinor, &inc.Description, &inc.RecordedBy, &inc.RecordedByName,
		&inc.CreatedAt, &inc.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Income{}, domain.ErrNotFound
	}
	return inc, err
}

func (s Store) CreateIncome(ctx context.Context, a domain.Actor, in domain.IncomeInput, date time.Time) (domain.Income, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Income{}, err
	}
	defer tx.Rollback(ctx)

	var headName string
	err = tx.QueryRow(ctx, `SELECT name FROM hospital_income_head WHERE id = $1`, in.IncomeHeadID).Scan(&headName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Income{}, domain.ErrValidation
	}
	if err != nil {
		return domain.Income{}, err
	}

	var out domain.Income
	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_income (
			income_head_id, name, invoice_number, date, amount_minor, description, recorded_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, income_head_id, name, invoice_number, date, amount_minor, description, recorded_by, created_at, updated_at
	`, in.IncomeHeadID, in.Name, in.InvoiceNumber, date, in.AmountMinor, in.Description, a.ID).Scan(
		&out.ID, &out.IncomeHeadID, &out.Name, &out.InvoiceNumber,
		&out.Date, &out.AmountMinor, &out.Description, &out.RecordedBy,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return out, err
	}
	out.IncomeHeadName = headName
	out.RecordedByName = a.Name

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'income.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Payroll ---

func (s Store) Payrolls(ctx context.Context, userID string, month string, year int, status *int, page int) ([]domain.EmployeePayroll, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	var whereClauses []string
	var args []any
	argIdx := 1

	if userID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.user_id = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}
	if month != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.month = $%d", argIdx))
		args = append(args, month)
		argIdx++
	}
	if year > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("p.year = $%d", argIdx))
		args = append(args, year)
		argIdx++
	}
	if status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("p.status = $%d", argIdx))
		args = append(args, *status)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM employee_payroll p"+whereSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.payroll_number, p.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''),
		       p.role, p.month, p.year, p.basic_salary_minor, p.allowance_minor, p.deductions_minor,
		       p.net_salary_minor, p.status, p.payment_date, p.created_by, COALESCE(c.name, ''),
		       p.created_at, p.updated_at
		FROM employee_payroll p
		JOIN "user" u ON p.user_id = u.id
		JOIN "user" c ON p.created_by = c.id
		%s
		ORDER BY p.year DESC, p.created_at DESC
		LIMIT %d OFFSET %d
	`, whereSQL, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.EmployeePayroll{}
	for rows.Next() {
		var ep domain.EmployeePayroll
		if err := rows.Scan(
			&ep.ID, &ep.PayrollNumber, &ep.UserID, &ep.UserName, &ep.UserEmail,
			&ep.Role, &ep.Month, &ep.Year, &ep.BasicSalaryMinor, &ep.AllowanceMinor,
			&ep.DeductionsMinor, &ep.NetSalaryMinor, &ep.Status, &ep.PaymentDate,
			&ep.CreatedBy, &ep.CreatedByName, &ep.CreatedAt, &ep.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, ep)
	}
	return out, total, nil
}

func (s Store) Payroll(ctx context.Context, id string) (domain.EmployeePayroll, error) {
	var ep domain.EmployeePayroll
	err := s.DB.QueryRow(ctx, `
		SELECT p.id, p.payroll_number, p.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''),
		       p.role, p.month, p.year, p.basic_salary_minor, p.allowance_minor, p.deductions_minor,
		       p.net_salary_minor, p.status, p.payment_date, p.created_by, COALESCE(c.name, ''),
		       p.created_at, p.updated_at
		FROM employee_payroll p
		JOIN "user" u ON p.user_id = u.id
		JOIN "user" c ON p.created_by = c.id
		WHERE p.id = $1
	`, id).Scan(
		&ep.ID, &ep.PayrollNumber, &ep.UserID, &ep.UserName, &ep.UserEmail,
		&ep.Role, &ep.Month, &ep.Year, &ep.BasicSalaryMinor, &ep.AllowanceMinor,
		&ep.DeductionsMinor, &ep.NetSalaryMinor, &ep.Status, &ep.PaymentDate,
		&ep.CreatedBy, &ep.CreatedByName, &ep.CreatedAt, &ep.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.EmployeePayroll{}, domain.ErrNotFound
	}
	return ep, err
}

func (s Store) CreatePayroll(ctx context.Context, a domain.Actor, in domain.PayrollInput, payrollNumber string) (domain.EmployeePayroll, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.EmployeePayroll{}, err
	}
	defer tx.Rollback(ctx)

	var userName, userEmail, role string
	err = tx.QueryRow(ctx, `
		SELECT u.name, u.email, COALESCE(s.role, 'staff')
		FROM "user" u
		LEFT JOIN staff_access s ON s.user_id = u.id
		WHERE u.id = $1
	`, in.UserID).Scan(&userName, &userEmail, &role)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.EmployeePayroll{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.EmployeePayroll{}, err
	}

	netSalaryMinor := in.BasicSalaryMinor + in.AllowanceMinor - in.DeductionsMinor
	if netSalaryMinor < 0 {
		return domain.EmployeePayroll{}, domain.ErrValidation
	}

	var out domain.EmployeePayroll
	err = tx.QueryRow(ctx, `
		INSERT INTO employee_payroll (
			payroll_number, user_id, role, month, year, basic_salary_minor, allowance_minor,
			deductions_minor, net_salary_minor, status, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, $10)
		RETURNING id, payroll_number, user_id, role, month, year, basic_salary_minor, allowance_minor,
		          deductions_minor, net_salary_minor, status, payment_date, created_by, created_at, updated_at
	`, payrollNumber, in.UserID, role, in.Month, in.Year, in.BasicSalaryMinor, in.AllowanceMinor,
		in.DeductionsMinor, netSalaryMinor, a.ID).Scan(
		&out.ID, &out.PayrollNumber, &out.UserID, &out.Role, &out.Month, &out.Year,
		&out.BasicSalaryMinor, &out.AllowanceMinor, &out.DeductionsMinor, &out.NetSalaryMinor,
		&out.Status, &out.PaymentDate, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}
	out.UserName = userName
	out.UserEmail = userEmail
	out.CreatedByName = a.Name

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'payroll.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) PayPayroll(ctx context.Context, a domain.Actor, id string, paymentDate time.Time) (domain.EmployeePayroll, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.EmployeePayroll{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.EmployeePayroll
	err = tx.QueryRow(ctx, `
		UPDATE employee_payroll
		SET status = 1, payment_date = $1, updated_at = clock_timestamp()
		WHERE id = $2 AND status = 0
		RETURNING id, payroll_number, user_id, role, month, year, basic_salary_minor, allowance_minor,
		          deductions_minor, net_salary_minor, status, payment_date, created_by, created_at, updated_at
	`, paymentDate, id).Scan(
		&out.ID, &out.PayrollNumber, &out.UserID, &out.Role, &out.Month, &out.Year,
		&out.BasicSalaryMinor, &out.AllowanceMinor, &out.DeductionsMinor, &out.NetSalaryMinor,
		&out.Status, &out.PaymentDate, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Check if exists
		var currentStatus int
		checkErr := tx.QueryRow(ctx, `SELECT status FROM employee_payroll WHERE id = $1`, id).Scan(&currentStatus)
		if errors.Is(checkErr, pgx.ErrNoRows) {
			return out, domain.ErrNotFound
		}
		if currentStatus == 1 {
			return out, domain.ErrConflict // already paid
		}
		return out, err
	}
	if err != nil {
		return out, err
	}

	// Fetch user & creator names
	_ = tx.QueryRow(ctx, `SELECT name, email FROM "user" WHERE id = $1`, out.UserID).Scan(&out.UserName, &out.UserEmail)
	_ = tx.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, out.CreatedBy).Scan(&out.CreatedByName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'payroll.paid', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Service Invoice Link ---

func (s Store) CreateServiceInvoiceLink(ctx context.Context, a domain.Actor, in domain.ServiceInvoiceLinkInput) (domain.ServiceInvoiceLink, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.ServiceInvoiceLink{}, err
	}
	defer tx.Rollback(ctx)

	// Validate invoice exists
	var invID string
	err = tx.QueryRow(ctx, `SELECT id FROM invoice WHERE id = $1`, in.InvoiceID).Scan(&invID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ServiceInvoiceLink{}, domain.ErrValidation
	}
	if err != nil {
		return domain.ServiceInvoiceLink{}, err
	}

	var out domain.ServiceInvoiceLink
	err = tx.QueryRow(ctx, `
		INSERT INTO service_invoice_link (invoice_id, source_type, source_id, amount_minor)
		VALUES ($1, $2, $3, $4)
		RETURNING id, invoice_id, source_type, source_id, amount_minor, created_at
	`, in.InvoiceID, in.SourceType, in.SourceID, in.AmountMinor).Scan(
		&out.ID, &out.InvoiceID, &out.SourceType, &out.SourceID, &out.AmountMinor, &out.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict // Anti-double-billing triggered!
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'service_invoice_link.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) ServiceInvoiceLinks(ctx context.Context, invoiceID string) ([]domain.ServiceInvoiceLink, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, invoice_id, source_type, source_id, amount_minor, created_at
		FROM service_invoice_link
		WHERE invoice_id = $1
		ORDER BY created_at ASC
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ServiceInvoiceLink{}
	for rows.Next() {
		var l domain.ServiceInvoiceLink
		if err := rows.Scan(&l.ID, &l.InvoiceID, &l.SourceType, &l.SourceID, &l.AmountMinor, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// --- Finance Summary ---

func (s Store) FinanceSummary(ctx context.Context, from, to time.Time) (domain.FinanceSummary, error) {
	res := domain.FinanceSummary{
		FromDate:       from,
		ToDate:         to,
		ExpensesByHead: []domain.HeadSummary{},
		IncomesByHead:  []domain.HeadSummary{},
	}

	// 1. Total Incomes
	err := s.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0)
		FROM hospital_income
		WHERE date >= $1 AND date <= $2
	`, from, to).Scan(&res.TotalIncomeMinor)
	if err != nil {
		return res, err
	}

	// 2. Total Expenses
	err = s.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0)
		FROM hospital_expense
		WHERE date >= $1 AND date <= $2
	`, from, to).Scan(&res.TotalExpenseMinor)
	if err != nil {
		return res, err
	}

	// 3. Total Payroll Paid
	err = s.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(net_salary_minor), 0)
		FROM employee_payroll
		WHERE status = 1 AND payment_date >= $1 AND payment_date <= $2
	`, from, to).Scan(&res.TotalPayrollMinor)
	if err != nil {
		return res, err
	}

	res.NetBalanceMinor = res.TotalIncomeMinor - (res.TotalExpenseMinor + res.TotalPayrollMinor)

	// 4. Expenses by Head
	eRows, err := s.DB.Query(ctx, `
		SELECT h.id, h.name, COALESCE(SUM(e.amount_minor), 0)
		FROM hospital_expense_head h
		LEFT JOIN hospital_expense e ON e.expense_head_id = h.id AND e.date >= $1 AND e.date <= $2
		GROUP BY h.id, h.name
		ORDER BY 3 DESC, h.name ASC
	`, from, to)
	if err != nil {
		return res, err
	}
	defer eRows.Close()

	for eRows.Next() {
		var hs domain.HeadSummary
		if err := eRows.Scan(&hs.HeadID, &hs.HeadName, &hs.TotalMinor); err != nil {
			return res, err
		}
		res.ExpensesByHead = append(res.ExpensesByHead, hs)
	}

	// 5. Incomes by Head
	iRows, err := s.DB.Query(ctx, `
		SELECT h.id, h.name, COALESCE(SUM(i.amount_minor), 0)
		FROM hospital_income_head h
		LEFT JOIN hospital_income i ON i.income_head_id = h.id AND i.date >= $1 AND i.date <= $2
		GROUP BY h.id, h.name
		ORDER BY 3 DESC, h.name ASC
	`, from, to)
	if err != nil {
		return res, err
	}
	defer iRows.Close()

	for iRows.Next() {
		var hs domain.HeadSummary
		if err := iRows.Scan(&hs.HeadID, &hs.HeadName, &hs.TotalMinor); err != nil {
			return res, err
		}
		res.IncomesByHead = append(res.IncomesByHead, hs)
	}

	return res, nil
}
