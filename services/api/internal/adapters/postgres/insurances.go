package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
)

func (s Store) CreateInsurance(ctx context.Context, a domain.Actor, in domain.InsuranceInput) (domain.Insurance, error) {
	status := 1
	if in.Status != nil {
		status = *in.Status
	}

	out := domain.Insurance{
		Name:              in.Name,
		ServiceTaxMinor:   in.ServiceTaxMinor,
		ServiceTax:        float64(in.ServiceTaxMinor) / 100.0,
		Discount:          in.Discount,
		Remark:            in.Remark,
		InsuranceNo:       in.InsuranceNo,
		InsuranceCode:     in.InsuranceCode,
		HospitalRateMinor: in.HospitalRateMinor,
		HospitalRate:      float64(in.HospitalRateMinor) / 100.0,
		Status:            status,
		CurrencySymbol:    "ETB",
	}

	_, _, totalMinor := domain.CalculateInsuranceTotals(in.ServiceTaxMinor, in.HospitalRateMinor, in.Discount, in.Diseases)
	out.TotalMinor = totalMinor
	out.Total = float64(totalMinor) / 100.0

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO insurance (name, service_tax_minor, discount, remark, insurance_no, insurance_code, hospital_rate_minor, total_minor, status, currency_symbol)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, out.Name, out.ServiceTaxMinor, out.Discount, out.Remark, out.InsuranceNo, out.InsuranceCode, out.HospitalRateMinor, out.TotalMinor, out.Status, out.CurrencySymbol).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	out.Diseases = make([]domain.InsuranceDiseaseLine, len(in.Diseases))
	for i, d := range in.Diseases {
		var dOut domain.InsuranceDiseaseLine
		dOut.InsuranceID = out.ID
		dOut.DiseaseName = d.DiseaseName
		dOut.DiseaseChargeMinor = d.DiseaseChargeMinor
		dOut.DiseaseCharge = float64(d.DiseaseChargeMinor) / 100.0

		err = tx.QueryRow(ctx, `
			INSERT INTO insurance_disease (insurance_id, disease_name, disease_charge_minor)
			VALUES ($1, $2, $3)
			RETURNING id, created_at, updated_at
		`, out.ID, d.DiseaseName, d.DiseaseChargeMinor).
			Scan(&dOut.ID, &dOut.CreatedAt, &dOut.UpdatedAt)
		if err != nil {
			return out, err
		}
		out.Diseases[i] = dOut
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'insurance.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}

	return out, tx.Commit(ctx)
}

func (s Store) UpdateInsurance(ctx context.Context, a domain.Actor, id string, in domain.InsuranceInput) (domain.Insurance, error) {
	var out domain.Insurance
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	var currentStatus int
	err = tx.QueryRow(ctx, `SELECT status FROM insurance WHERE id = $1 FOR UPDATE`, id).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, domain.ErrNotFound
		}
		return out, err
	}

	targetStatus := currentStatus
	if in.Status != nil {
		targetStatus = *in.Status
	}

	_, _, totalMinor := domain.CalculateInsuranceTotals(in.ServiceTaxMinor, in.HospitalRateMinor, in.Discount, in.Diseases)

	err = tx.QueryRow(ctx, `
		UPDATE insurance
		SET name = $1, service_tax_minor = $2, discount = $3, remark = $4,
		    insurance_no = $5, insurance_code = $6, hospital_rate_minor = $7,
		    total_minor = $8, status = $9, updated_at = clock_timestamp()
		WHERE id = $10
		RETURNING id, name, service_tax_minor, discount, remark, insurance_no, insurance_code, hospital_rate_minor, total_minor, status, currency_symbol, created_at, updated_at
	`, in.Name, in.ServiceTaxMinor, in.Discount, in.Remark, in.InsuranceNo, in.InsuranceCode, in.HospitalRateMinor, totalMinor, targetStatus, id).
		Scan(&out.ID, &out.Name, &out.ServiceTaxMinor, &out.Discount, &out.Remark, &out.InsuranceNo, &out.InsuranceCode, &out.HospitalRateMinor, &out.TotalMinor, &out.Status, &out.CurrencySymbol, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	out.ServiceTax = float64(out.ServiceTaxMinor) / 100.0
	out.HospitalRate = float64(out.HospitalRateMinor) / 100.0
	out.Total = float64(out.TotalMinor) / 100.0

	// Replace disease lines
	if _, err = tx.Exec(ctx, `DELETE FROM insurance_disease WHERE insurance_id = $1`, id); err != nil {
		return out, err
	}

	out.Diseases = make([]domain.InsuranceDiseaseLine, len(in.Diseases))
	for i, d := range in.Diseases {
		var dOut domain.InsuranceDiseaseLine
		dOut.InsuranceID = id
		dOut.DiseaseName = d.DiseaseName
		dOut.DiseaseChargeMinor = d.DiseaseChargeMinor
		dOut.DiseaseCharge = float64(d.DiseaseChargeMinor) / 100.0

		err = tx.QueryRow(ctx, `
			INSERT INTO insurance_disease (insurance_id, disease_name, disease_charge_minor)
			VALUES ($1, $2, $3)
			RETURNING id, created_at, updated_at
		`, id, d.DiseaseName, d.DiseaseChargeMinor).
			Scan(&dOut.ID, &dOut.CreatedAt, &dOut.UpdatedAt)
		if err != nil {
			return out, err
		}
		out.Diseases[i] = dOut
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'insurance.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}

	return out, tx.Commit(ctx)
}

func (s Store) DeleteInsurance(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM insurance WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}

	var inUse bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ipd_admission_details WHERE insurance_id = $1)`, id).Scan(&inUse)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}

	tag, err := tx.Exec(ctx, `DELETE FROM insurance WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.ErrInUse
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'insurance.deleted', $2)`, a.ID, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) Insurance(ctx context.Context, a domain.Actor, id string) (domain.Insurance, error) {
	var out domain.Insurance
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, service_tax_minor, discount, remark, insurance_no, insurance_code, hospital_rate_minor, total_minor, status, currency_symbol, created_at, updated_at
		FROM insurance
		WHERE id = $1
	`, id).Scan(&out.ID, &out.Name, &out.ServiceTaxMinor, &out.Discount, &out.Remark, &out.InsuranceNo, &out.InsuranceCode, &out.HospitalRateMinor, &out.TotalMinor, &out.Status, &out.CurrencySymbol, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, domain.ErrNotFound
		}
		return out, err
	}

	out.ServiceTax = float64(out.ServiceTaxMinor) / 100.0
	out.HospitalRate = float64(out.HospitalRateMinor) / 100.0
	out.Total = float64(out.TotalMinor) / 100.0

	rows, err := s.DB.Query(ctx, `
		SELECT id, insurance_id, disease_name, disease_charge_minor, created_at, updated_at
		FROM insurance_disease
		WHERE insurance_id = $1
		ORDER BY created_at ASC, id ASC
	`, id)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	out.Diseases = []domain.InsuranceDiseaseLine{}
	for rows.Next() {
		var d domain.InsuranceDiseaseLine
		if err := rows.Scan(&d.ID, &d.InsuranceID, &d.DiseaseName, &d.DiseaseChargeMinor, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return out, err
		}
		d.DiseaseCharge = float64(d.DiseaseChargeMinor) / 100.0
		out.Diseases = append(out.Diseases, d)
	}

	return out, rows.Err()
}

func (s Store) Insurances(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.Insurance, int, error) {
	search = strings.TrimSpace(search)
	var total int
	err := s.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM insurance
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR insurance_no ILIKE '%' || $1 || '%' OR insurance_code ILIKE '%' || $1 || '%')
	`, search).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	rows, err := s.DB.Query(ctx, `
		SELECT id, name, service_tax_minor, discount, remark, insurance_no, insurance_code, hospital_rate_minor, total_minor, status, currency_symbol, created_at, updated_at
		FROM insurance
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR insurance_no ILIKE '%' || $1 || '%' OR insurance_code ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Insurance{}
	insIDs := []string{}
	for rows.Next() {
		var ins domain.Insurance
		if err := rows.Scan(&ins.ID, &ins.Name, &ins.ServiceTaxMinor, &ins.Discount, &ins.Remark, &ins.InsuranceNo, &ins.InsuranceCode, &ins.HospitalRateMinor, &ins.TotalMinor, &ins.Status, &ins.CurrencySymbol, &ins.CreatedAt, &ins.UpdatedAt); err != nil {
			return nil, 0, err
		}
		ins.ServiceTax = float64(ins.ServiceTaxMinor) / 100.0
		ins.HospitalRate = float64(ins.HospitalRateMinor) / 100.0
		ins.Total = float64(ins.TotalMinor) / 100.0
		ins.Diseases = []domain.InsuranceDiseaseLine{}
		out = append(out, ins)
		insIDs = append(insIDs, ins.ID)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	if len(insIDs) > 0 {
		dRows, err := s.DB.Query(ctx, `
			SELECT id, insurance_id, disease_name, disease_charge_minor, created_at, updated_at
			FROM insurance_disease
			WHERE insurance_id = ANY($1)
			ORDER BY created_at ASC, id ASC
		`, insIDs)
		if err != nil {
			return nil, 0, err
		}
		defer dRows.Close()

		diseasesByIns := make(map[string][]domain.InsuranceDiseaseLine)
		for dRows.Next() {
			var d domain.InsuranceDiseaseLine
			if err := dRows.Scan(&d.ID, &d.InsuranceID, &d.DiseaseName, &d.DiseaseChargeMinor, &d.CreatedAt, &d.UpdatedAt); err != nil {
				return nil, 0, err
			}
			d.DiseaseCharge = float64(d.DiseaseChargeMinor) / 100.0
			diseasesByIns[d.InsuranceID] = append(diseasesByIns[d.InsuranceID], d)
		}
		for i := range out {
			if lines, ok := diseasesByIns[out[i].ID]; ok {
				out[i].Diseases = lines
			}
		}
	}

	return out, total, nil
}

func (s Store) ToggleInsuranceStatus(ctx context.Context, a domain.Actor, id string) (domain.Insurance, error) {
	var out domain.Insurance
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE insurance
		SET status = CASE WHEN status = 1 THEN 0 ELSE 1 END,
		    updated_at = clock_timestamp()
		WHERE id = $1
		RETURNING id, name, service_tax_minor, discount, remark, insurance_no, insurance_code, hospital_rate_minor, total_minor, status, currency_symbol, created_at, updated_at
	`, id).Scan(&out.ID, &out.Name, &out.ServiceTaxMinor, &out.Discount, &out.Remark, &out.InsuranceNo, &out.InsuranceCode, &out.HospitalRateMinor, &out.TotalMinor, &out.Status, &out.CurrencySymbol, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, domain.ErrNotFound
		}
		return out, err
	}

	out.ServiceTax = float64(out.ServiceTaxMinor) / 100.0
	out.HospitalRate = float64(out.HospitalRateMinor) / 100.0
	out.Total = float64(out.TotalMinor) / 100.0

	rows, err := tx.Query(ctx, `
		SELECT id, insurance_id, disease_name, disease_charge_minor, created_at, updated_at
		FROM insurance_disease
		WHERE insurance_id = $1
		ORDER BY created_at ASC, id ASC
	`, id)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	out.Diseases = []domain.InsuranceDiseaseLine{}
	for rows.Next() {
		var d domain.InsuranceDiseaseLine
		if err := rows.Scan(&d.ID, &d.InsuranceID, &d.DiseaseName, &d.DiseaseChargeMinor, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return out, err
		}
		d.DiseaseCharge = float64(d.DiseaseChargeMinor) / 100.0
		out.Diseases = append(out.Diseases, d)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'insurance.status_toggled', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}

	return out, tx.Commit(ctx)
}
