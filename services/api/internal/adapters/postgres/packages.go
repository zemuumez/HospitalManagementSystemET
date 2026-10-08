package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
)

func (s Store) CreatePackage(ctx context.Context, a domain.Actor, in domain.PackageInput) (domain.Package, error) {
	out := domain.Package{
		Name:           in.Name,
		Description:    in.Description,
		Discount:       in.Discount,
		CurrencySymbol: "ETB",
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	// Validate each service in input against active services
	for i := range in.Services {
		var sName string
		var sRate int64
		var sStatus int16
		err = tx.QueryRow(ctx, `SELECT name, rate_minor, status FROM hospital_service WHERE id = $1`, in.Services[i].ServiceID).
			Scan(&sName, &sRate, &sStatus)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return out, domain.ErrValidation
			}
			return out, err
		}
		if sStatus != 1 {
			return out, domain.ErrValidation
		}
		if !in.Services[i].HasRate && sRate > 0 {
			in.Services[i].RateMinor = sRate
			in.Services[i].HasRate = true
		}
	}

	// Recalculate authoritative totals
	subtotalMinor, _, totalMinor, lineAmounts, err := domain.CalculatePackageTotals(in.Discount, in.Services)
	if err != nil {
		return out, domain.ErrValidation
	}
	out.TotalAmountMinor = totalMinor

	err = tx.QueryRow(ctx, `
		INSERT INTO package (name, description, discount, total_amount_minor, currency_symbol)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, out.Name, out.Description, out.Discount, out.TotalAmountMinor, out.CurrencySymbol).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	// Insert service lines
	out.Services = make([]domain.PackageServiceLine, len(in.Services))
	for i, line := range in.Services {
		var lineOut domain.PackageServiceLine
		lineOut.PackageID = out.ID
		lineOut.ServiceID = line.ServiceID
		lineOut.Quantity = line.Quantity
		lineOut.RateMinor = line.RateMinor
		lineOut.AmountMinor = lineAmounts[i]

		err = tx.QueryRow(ctx, `
			INSERT INTO package_service (package_id, service_id, quantity, rate_minor, amount_minor)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`, out.ID, line.ServiceID, line.Quantity, line.RateMinor, lineAmounts[i]).
			Scan(&lineOut.ID, &lineOut.CreatedAt, &lineOut.UpdatedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503") {
				return out, domain.ErrValidation
			}
			return out, err
		}

		_ = tx.QueryRow(ctx, `SELECT name FROM hospital_service WHERE id = $1`, line.ServiceID).Scan(&lineOut.ServiceName)
		out.Services[i] = lineOut
	}

	// Reconcile persisted line totals
	var persistedLineSum int64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount_minor), 0) FROM package_service WHERE package_id = $1`, out.ID).Scan(&persistedLineSum)
	if err != nil {
		return out, err
	}
	if persistedLineSum != subtotalMinor {
		return out, domain.ErrValidation
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'package.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}

	return out, tx.Commit(ctx)
}

func (s Store) UpdatePackage(ctx context.Context, a domain.Actor, id string, in domain.PackageInput) (domain.Package, error) {
	var out domain.Package

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	// Lock package for update
	var currentID string
	err = tx.QueryRow(ctx, `SELECT id FROM package WHERE id = $1 FOR UPDATE`, id).Scan(&currentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, domain.ErrNotFound
		}
		return out, err
	}

	// Validate child line ownership: any supplied line ID MUST belong to this package and cannot be duplicated
	seenLineIDs := make(map[string]bool)
	for _, line := range in.Services {
		lineID := strings.TrimSpace(line.ID)
		if lineID != "" {
			if seenLineIDs[lineID] {
				return out, domain.ErrValidation
			}
			seenLineIDs[lineID] = true
			var ownerPkgID string
			err = tx.QueryRow(ctx, `SELECT package_id FROM package_service WHERE id = $1`, lineID).Scan(&ownerPkgID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return out, domain.ErrValidation // Line ID does not exist
				}
				return out, err
			}
			if ownerPkgID != id {
				return out, domain.ErrConflict // Cross-parent child-ID rejection!
			}
		}
	}

	// Validate services against active services catalog
	for i := range in.Services {
		var sName string
		var sRate int64
		var sStatus int16
		err = tx.QueryRow(ctx, `SELECT name, rate_minor, status FROM hospital_service WHERE id = $1`, in.Services[i].ServiceID).
			Scan(&sName, &sRate, &sStatus)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return out, domain.ErrValidation
			}
			return out, err
		}
		if sStatus != 1 {
			// Allow existing service lines in this package to remain if archived after creation,
			// but reject selecting or adding new archived services.
			var alreadyInPkg int
			_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM package_service WHERE package_id = $1 AND service_id = $2`, id, in.Services[i].ServiceID).Scan(&alreadyInPkg)
			if alreadyInPkg == 0 {
				return out, domain.ErrValidation
			}
		}
		if !in.Services[i].HasRate && sRate > 0 {
			in.Services[i].RateMinor = sRate
			in.Services[i].HasRate = true
		}
	}

	// Recalculate totals
	subtotalMinor, _, totalMinor, lineAmounts, err := domain.CalculatePackageTotals(in.Discount, in.Services)
	if err != nil {
		return out, domain.ErrValidation
	}

	err = tx.QueryRow(ctx, `
		UPDATE package
		SET name = $1, description = $2, discount = $3, total_amount_minor = $4, updated_at = clock_timestamp()
		WHERE id = $5
		RETURNING id, name, description, discount, total_amount_minor, currency_symbol, created_at, updated_at
	`, in.Name, in.Description, in.Discount, totalMinor, id).
		Scan(&out.ID, &out.Name, &out.Description, &out.Discount, &out.TotalAmountMinor, &out.CurrencySymbol, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	// Existing line IDs in DB
	rows, err := tx.Query(ctx, `SELECT id FROM package_service WHERE package_id = $1`, id)
	if err != nil {
		return out, err
	}
	existingDBLines := make(map[string]bool)
	for rows.Next() {
		var lineID string
		if err = rows.Scan(&lineID); err == nil {
			existingDBLines[lineID] = true
		}
	}
	rows.Close()

	retainedLineIDs := make(map[string]bool)
	out.Services = make([]domain.PackageServiceLine, len(in.Services))

	for i, line := range in.Services {
		var lineOut domain.PackageServiceLine
		lineOut.PackageID = id
		lineOut.ServiceID = line.ServiceID
		lineOut.Quantity = line.Quantity
		lineOut.RateMinor = line.RateMinor
		lineOut.AmountMinor = lineAmounts[i]

		if line.ID != "" && existingDBLines[line.ID] {
			// Update existing line
			retainedLineIDs[line.ID] = true
			err = tx.QueryRow(ctx, `
				UPDATE package_service
				SET service_id = $1, quantity = $2, rate_minor = $3, amount_minor = $4, updated_at = clock_timestamp()
				WHERE id = $5 AND package_id = $6
				RETURNING id, created_at, updated_at
			`, line.ServiceID, line.Quantity, line.RateMinor, lineAmounts[i], line.ID, id).
				Scan(&lineOut.ID, &lineOut.CreatedAt, &lineOut.UpdatedAt)
			if err != nil {
				return out, err
			}
		} else {
			// Insert new line
			err = tx.QueryRow(ctx, `
				INSERT INTO package_service (package_id, service_id, quantity, rate_minor, amount_minor)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, created_at, updated_at
			`, id, line.ServiceID, line.Quantity, line.RateMinor, lineAmounts[i]).
				Scan(&lineOut.ID, &lineOut.CreatedAt, &lineOut.UpdatedAt)
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503") {
					return out, domain.ErrValidation
				}
				return out, err
			}
			retainedLineIDs[lineOut.ID] = true
		}

		_ = tx.QueryRow(ctx, `SELECT name FROM hospital_service WHERE id = $1`, line.ServiceID).Scan(&lineOut.ServiceName)
		out.Services[i] = lineOut
	}

	// Delete omitted lines
	for oldLineID := range existingDBLines {
		if !retainedLineIDs[oldLineID] {
			if _, err = tx.Exec(ctx, `DELETE FROM package_service WHERE id = $1 AND package_id = $2`, oldLineID, id); err != nil {
				return out, err
			}
		}
	}

	// Reconcile persisted line totals
	var persistedLineSum int64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount_minor), 0) FROM package_service WHERE package_id = $1`, id).Scan(&persistedLineSum)
	if err != nil {
		return out, err
	}
	if persistedLineSum != subtotalMinor {
		return out, domain.ErrValidation
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'package.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}

	return out, tx.Commit(ctx)
}

func (s Store) DeletePackage(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check if package exists
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM package WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}

	// Check if referenced by patient admissions (ipd_admission_details)
	var inUse bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ipd_admission_details WHERE package_id = $1)`, id).Scan(&inUse)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}

	// Execute deletion (cascades to package_service)
	tag, err := tx.Exec(ctx, `DELETE FROM package WHERE id = $1`, id)
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

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'package.deleted', $2)`, a.ID, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) Package(ctx context.Context, a domain.Actor, id string) (domain.Package, error) {
	var out domain.Package
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description, discount, total_amount_minor, currency_symbol, created_at, updated_at
		FROM package
		WHERE id = $1
	`, id).Scan(&out.ID, &out.Name, &out.Description, &out.Discount, &out.TotalAmountMinor, &out.CurrencySymbol, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, domain.ErrNotFound
		}
		return out, err
	}

	// Fetch service lines with service names
	rows, err := s.DB.Query(ctx, `
		SELECT ps.id, ps.package_id, ps.service_id, s.name, ps.quantity, ps.rate_minor, ps.amount_minor, ps.created_at, ps.updated_at
		FROM package_service ps
		JOIN hospital_service s ON s.id = ps.service_id
		WHERE ps.package_id = $1
		ORDER BY ps.created_at ASC, ps.id ASC
	`, id)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	out.Services = []domain.PackageServiceLine{}
	for rows.Next() {
		var line domain.PackageServiceLine
		if err = rows.Scan(&line.ID, &line.PackageID, &line.ServiceID, &line.ServiceName, &line.Quantity, &line.RateMinor, &line.AmountMinor, &line.CreatedAt, &line.UpdatedAt); err != nil {
			return out, err
		}
		out.Services = append(out.Services, line)
	}

	return out, rows.Err()
}

func (s Store) Packages(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.Package, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}
	argIdx := 1
	if strings.TrimSpace(search) != "" {
		whereClause += fmt.Sprintf(" AND p.name ILIKE $%d", argIdx)
		args = append(args, "%"+strings.TrimSpace(search)+"%")
		argIdx++
	}

	var total int
	err := s.DB.QueryRow(ctx, `SELECT count(*) FROM package p `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.discount, p.total_amount_minor, p.currency_symbol, p.created_at, p.updated_at
		FROM package p
		%s
		ORDER BY p.name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	pkgs := []domain.Package{}
	pkgIDs := []string{}
	for rows.Next() {
		var p domain.Package
		if err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Discount, &p.TotalAmountMinor, &p.CurrencySymbol, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		p.Services = []domain.PackageServiceLine{}
		pkgs = append(pkgs, p)
		pkgIDs = append(pkgIDs, p.ID)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	if len(pkgIDs) > 0 {
		// Batch load service lines
		lineRows, err := s.DB.Query(ctx, `
			SELECT ps.id, ps.package_id, ps.service_id, s.name, ps.quantity, ps.rate_minor, ps.amount_minor, ps.created_at, ps.updated_at
			FROM package_service ps
			JOIN hospital_service s ON s.id = ps.service_id
			WHERE ps.package_id = ANY($1)
			ORDER BY ps.created_at ASC, ps.id ASC
		`, pkgIDs)
		if err != nil {
			return nil, 0, err
		}
		defer lineRows.Close()

		linesByPkg := make(map[string][]domain.PackageServiceLine)
		for lineRows.Next() {
			var line domain.PackageServiceLine
			if err = lineRows.Scan(&line.ID, &line.PackageID, &line.ServiceID, &line.ServiceName, &line.Quantity, &line.RateMinor, &line.AmountMinor, &line.CreatedAt, &line.UpdatedAt); err != nil {
				return nil, 0, err
			}
			linesByPkg[line.PackageID] = append(linesByPkg[line.PackageID], line)
		}
		for i := range pkgs {
			if lines, ok := linesByPkg[pkgs[i].ID]; ok {
				pkgs[i].Services = lines
			}
		}
	}

	return pkgs, total, nil
}
