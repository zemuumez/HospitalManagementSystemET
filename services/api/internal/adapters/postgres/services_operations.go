package postgres

import (
	"context"
	"fmt"
	"strings"

	"hms.local/api/internal/domain"
)

// --- Charge Categories ---

func (s Store) CreateChargeCategory(ctx context.Context, a domain.Actor, in domain.ChargeCategoryInput) (domain.ChargeCategory, error) {
	out := domain.ChargeCategory{
		Name:        in.Name,
		Description: in.Description,
		ChargeType:  in.ChargeType,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO charge_category (name, description, charge_type)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, out.Name, out.Description, out.ChargeType).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'charge_category.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateChargeCategory(ctx context.Context, a domain.Actor, id string, in domain.ChargeCategoryInput) (domain.ChargeCategory, error) {
	var out domain.ChargeCategory
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE charge_category
		SET name = $1, description = $2, charge_type = $3, updated_at = clock_timestamp()
		WHERE id = $4
		RETURNING id, name, description, charge_type, created_at, updated_at
	`, in.Name, in.Description, in.ChargeType, id).
		Scan(&out.ID, &out.Name, &out.Description, &out.ChargeType, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'charge_category.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) ChargeCategories(ctx context.Context, a domain.Actor, chargeType int) ([]domain.ChargeCategory, error) {
	query := `SELECT id, name, description, charge_type, created_at, updated_at FROM charge_category`
	args := []any{}
	if chargeType > 0 {
		query += ` WHERE charge_type = $1`
		args = append(args, chargeType)
	}
	query += ` ORDER BY name ASC`

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ChargeCategory{}
	for rows.Next() {
		var c domain.ChargeCategory
		if err = rows.Scan(&c.ID, &c.Name, &c.Description, &c.ChargeType, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s Store) ChargeCategory(ctx context.Context, a domain.Actor, id string) (domain.ChargeCategory, error) {
	var c domain.ChargeCategory
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description, charge_type, created_at, updated_at 
		FROM charge_category WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Description, &c.ChargeType, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, clinicalError(err)
	}
	return c, nil
}

// --- Charges ---

func (s Store) CreateCharge(ctx context.Context, a domain.Actor, in domain.HospitalChargeInput) (domain.HospitalCharge, error) {
	out := domain.HospitalCharge{
		ChargeType:          in.ChargeType,
		ChargeCategoryID:    in.ChargeCategoryID,
		Code:                in.Code,
		StandardChargeMinor: in.StandardChargeMinor,
		Description:         in.Description,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_charge (charge_type, charge_category_id, code, standard_charge_minor, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, out.ChargeType, out.ChargeCategoryID, out.Code, out.StandardChargeMinor, out.Description).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	_ = tx.QueryRow(ctx, `SELECT name FROM charge_category WHERE id = $1`, out.ChargeCategoryID).Scan(&out.ChargeCategoryName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'charge.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateCharge(ctx context.Context, a domain.Actor, id string, in domain.HospitalChargeInput) (domain.HospitalCharge, error) {
	var out domain.HospitalCharge
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_charge
		SET charge_type = $1, charge_category_id = $2, code = $3, standard_charge_minor = $4,
		    description = $5, updated_at = clock_timestamp()
		WHERE id = $6
		RETURNING id, charge_type, charge_category_id, code, standard_charge_minor, description, created_at, updated_at
	`, in.ChargeType, in.ChargeCategoryID, in.Code, in.StandardChargeMinor, in.Description, id).
		Scan(&out.ID, &out.ChargeType, &out.ChargeCategoryID, &out.Code, &out.StandardChargeMinor, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	_ = tx.QueryRow(ctx, `SELECT name FROM charge_category WHERE id = $1`, out.ChargeCategoryID).Scan(&out.ChargeCategoryName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'charge.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) Charges(ctx context.Context, a domain.Actor, page int, categoryID string) ([]domain.HospitalCharge, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}
	argIdx := 1
	if categoryID != "" {
		whereClause += fmt.Sprintf(" AND c.charge_category_id = $%d", argIdx)
		args = append(args, categoryID)
		argIdx++
	}

	var total int
	err := s.DB.QueryRow(ctx, `SELECT count(*) FROM hospital_charge c `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := 25
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT c.id, c.charge_type, c.charge_category_id, cat.name, c.code, c.standard_charge_minor, c.description, c.created_at, c.updated_at
		FROM hospital_charge c
		JOIN charge_category cat ON cat.id = c.charge_category_id
		%s
		ORDER BY c.code ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalCharge{}
	for rows.Next() {
		var ch domain.HospitalCharge
		if err = rows.Scan(&ch.ID, &ch.ChargeType, &ch.ChargeCategoryID, &ch.ChargeCategoryName, &ch.Code, &ch.StandardChargeMinor, &ch.Description, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, ch)
	}
	return out, total, rows.Err()
}

func (s Store) Charge(ctx context.Context, a domain.Actor, id string) (domain.HospitalCharge, error) {
	var ch domain.HospitalCharge
	err := s.DB.QueryRow(ctx, `
		SELECT c.id, c.charge_type, c.charge_category_id, cat.name, c.code, c.standard_charge_minor, c.description, c.created_at, c.updated_at
		FROM hospital_charge c
		JOIN charge_category cat ON cat.id = c.charge_category_id
		WHERE c.id = $1
	`, id).Scan(&ch.ID, &ch.ChargeType, &ch.ChargeCategoryID, &ch.ChargeCategoryName, &ch.Code, &ch.StandardChargeMinor, &ch.Description, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return ch, clinicalError(err)
	}
	return ch, nil
}

// --- Services ---

func (s Store) CreateService(ctx context.Context, a domain.Actor, in domain.HospitalServiceInput) (domain.HospitalService, error) {
	out := domain.HospitalService{
		Name:        in.Name,
		Description: in.Description,
		Quantity:    in.Quantity,
		RateMinor:   in.RateMinor,
		Status:      in.Status,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_service (name, description, quantity, rate_minor, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, out.Name, out.Description, out.Quantity, out.RateMinor, out.Status).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'service.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateService(ctx context.Context, a domain.Actor, id string, in domain.HospitalServiceInput) (domain.HospitalService, error) {
	var out domain.HospitalService
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_service
		SET name = $1, description = $2, quantity = $3, rate_minor = $4, status = $5, updated_at = clock_timestamp()
		WHERE id = $6
		RETURNING id, name, description, quantity, rate_minor, status, created_at, updated_at
	`, in.Name, in.Description, in.Quantity, in.RateMinor, in.Status, id).
		Scan(&out.ID, &out.Name, &out.Description, &out.Quantity, &out.RateMinor, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'service.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) Services(ctx context.Context, a domain.Actor, page int, limit int, status *int, search string) ([]domain.HospitalService, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}
	argIdx := 1
	if status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}
	if strings.TrimSpace(search) != "" {
		whereClause += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+strings.TrimSpace(search)+"%")
		argIdx++
	}

	var total int
	err := s.DB.QueryRow(ctx, `SELECT count(*) FROM hospital_service `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, name, description, quantity, rate_minor, status, created_at, updated_at
		FROM hospital_service
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalService{}
	for rows.Next() {
		var srv domain.HospitalService
		if err = rows.Scan(&srv.ID, &srv.Name, &srv.Description, &srv.Quantity, &srv.RateMinor, &srv.Status, &srv.CreatedAt, &srv.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, srv)
	}
	return out, total, rows.Err()
}

func (s Store) Service(ctx context.Context, a domain.Actor, id string) (domain.HospitalService, error) {
	var srv domain.HospitalService
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description, quantity, rate_minor, status, created_at, updated_at
		FROM hospital_service
		WHERE id = $1
	`, id).Scan(&srv.ID, &srv.Name, &srv.Description, &srv.Quantity, &srv.RateMinor, &srv.Status, &srv.CreatedAt, &srv.UpdatedAt)
	if err != nil {
		return srv, clinicalError(err)
	}
	return srv, nil
}

// --- Operation Categories & Operations ---

func (s Store) CreateOperationCategory(ctx context.Context, a domain.Actor, in domain.OperationCategoryInput) (domain.OperationCategory, error) {
	out := domain.OperationCategory{Name: in.Name}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO operation_category (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`, out.Name).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'operation_category.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) OperationCategories(ctx context.Context, a domain.Actor) ([]domain.OperationCategory, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, name, created_at, updated_at FROM operation_category ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.OperationCategory{}
	for rows.Next() {
		var oc domain.OperationCategory
		if err = rows.Scan(&oc.ID, &oc.Name, &oc.CreatedAt, &oc.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, oc)
	}
	return out, rows.Err()
}

func (s Store) DeleteOperationCategory(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM hospital_operation WHERE operation_category_id = $1`, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrInUse
	}

	res, err := tx.Exec(ctx, `DELETE FROM operation_category WHERE id = $1`, id)
	if err != nil {
		return clinicalError(err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'operation_category.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) DeleteOperation(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `DELETE FROM hospital_operation WHERE id = $1`, id)
	if err != nil {
		return clinicalError(err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'operation.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) CreateOperation(ctx context.Context, a domain.Actor, in domain.HospitalOperationInput) (domain.HospitalOperation, error) {
	out := domain.HospitalOperation{
		OperationCategoryID: in.OperationCategoryID,
		Name:                in.Name,
		Description:         in.Description,
		Status:              in.Status,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_operation (operation_category_id, name, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, out.OperationCategoryID, out.Name, out.Description, out.Status).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	_ = tx.QueryRow(ctx, `SELECT name FROM operation_category WHERE id = $1`, out.OperationCategoryID).Scan(&out.OperationCategoryName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'operation.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateOperation(ctx context.Context, a domain.Actor, id string, in domain.HospitalOperationInput) (domain.HospitalOperation, error) {
	var out domain.HospitalOperation
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_operation
		SET operation_category_id = $1, name = $2, description = $3, status = $4, updated_at = clock_timestamp()
		WHERE id = $5
		RETURNING id, operation_category_id, name, description, status, created_at, updated_at
	`, in.OperationCategoryID, in.Name, in.Description, in.Status, id).
		Scan(&out.ID, &out.OperationCategoryID, &out.Name, &out.Description, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	_ = tx.QueryRow(ctx, `SELECT name FROM operation_category WHERE id = $1`, out.OperationCategoryID).Scan(&out.OperationCategoryName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'operation.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) Operations(ctx context.Context, a domain.Actor, page int, categoryID string) ([]domain.HospitalOperation, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}
	argIdx := 1
	if categoryID != "" {
		whereClause += fmt.Sprintf(" AND o.operation_category_id = $%d", argIdx)
		args = append(args, categoryID)
		argIdx++
	}

	var total int
	err := s.DB.QueryRow(ctx, `SELECT count(*) FROM hospital_operation o `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := 25
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT o.id, o.operation_category_id, oc.name, o.name, o.description, o.status, o.created_at, o.updated_at
		FROM hospital_operation o
		JOIN operation_category oc ON oc.id = o.operation_category_id
		%s
		ORDER BY o.name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalOperation{}
	for rows.Next() {
		var op domain.HospitalOperation
		if err = rows.Scan(&op.ID, &op.OperationCategoryID, &op.OperationCategoryName, &op.Name, &op.Description, &op.Status, &op.CreatedAt, &op.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, op)
	}
	return out, total, rows.Err()
}

func (s Store) Operation(ctx context.Context, a domain.Actor, id string) (domain.HospitalOperation, error) {
	var op domain.HospitalOperation
	err := s.DB.QueryRow(ctx, `
		SELECT o.id, o.operation_category_id, oc.name, o.name, o.description, o.status, o.created_at, o.updated_at
		FROM hospital_operation o
		JOIN operation_category oc ON oc.id = o.operation_category_id
		WHERE o.id = $1
	`, id).Scan(&op.ID, &op.OperationCategoryID, &op.OperationCategoryName, &op.Name, &op.Description, &op.Status, &op.CreatedAt, &op.UpdatedAt)
	if err != nil {
		return op, clinicalError(err)
	}
	return op, nil
}

// --- Custom Fields ---

func (s Store) CreateCustomField(ctx context.Context, a domain.Actor, in domain.CustomFieldInput) (domain.CustomField, error) {
	out := domain.CustomField{
		ModuleName: in.ModuleName,
		FieldType:  in.FieldType,
		FieldName:  in.FieldName,
		IsRequired: in.IsRequired,
		Values:     in.Values,
		Grid:       in.Grid,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO custom_field (module_name, field_type, field_name, is_required, values, grid)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, out.ModuleName, out.FieldType, out.FieldName, out.IsRequired, out.Values, out.Grid).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'custom_field.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) DeleteCustomField(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `DELETE FROM custom_field WHERE id = $1`, id)
	if err != nil {
		return clinicalError(err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'custom_field.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) CustomFields(ctx context.Context, a domain.Actor, moduleName string) ([]domain.CustomField, error) {
	query := `SELECT id, module_name, field_type, field_name, is_required, values, grid, created_at, updated_at FROM custom_field`
	args := []any{}
	if moduleName != "" {
		query += ` WHERE module_name = $1`
		args = append(args, moduleName)
	}
	query += ` ORDER BY field_name ASC`

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.CustomField{}
	for rows.Next() {
		var cf domain.CustomField
		if err = rows.Scan(&cf.ID, &cf.ModuleName, &cf.FieldType, &cf.FieldName, &cf.IsRequired, &cf.Values, &cf.Grid, &cf.CreatedAt, &cf.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, cf)
	}
	return out, rows.Err()
}

// --- Module Settings ---

func (s Store) ModuleSettings(ctx context.Context, a domain.Actor) ([]domain.HospitalModuleSetting, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, module_key, name, route, is_active, created_at, updated_at FROM hospital_module_setting ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.HospitalModuleSetting{}
	for rows.Next() {
		var m domain.HospitalModuleSetting
		if err = rows.Scan(&m.ID, &m.ModuleKey, &m.Name, &m.Route, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s Store) UpdateModuleSetting(ctx context.Context, a domain.Actor, key string, isActive bool) (domain.HospitalModuleSetting, error) {
	var out domain.HospitalModuleSetting
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_module_setting
		SET is_active = $1, updated_at = clock_timestamp()
		WHERE module_key = $2
		RETURNING id, module_key, name, route, is_active, created_at, updated_at
	`, isActive, key).Scan(&out.ID, &out.ModuleKey, &out.Name, &out.Route, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'module_setting.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
