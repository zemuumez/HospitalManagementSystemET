package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) CreateAmbulance(ctx context.Context, a domain.Actor, in domain.AmbulanceInput) (domain.Ambulance, error) {
	out := domain.Ambulance{
		VehicleNumber: in.VehicleNumber,
		VehicleModel:  in.VehicleModel,
		YearMade:      in.YearMade,
		DriverName:    in.DriverName,
		DriverLicense: in.DriverLicense,
		DriverContact: in.DriverContact,
		VehicleType:   in.VehicleType,
		IsAvailable:   true,
		Note:          in.Note,
		Version:       1,
	}
	if in.IsAvailable != nil {
		out.IsAvailable = *in.IsAvailable
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO ambulance (vehicle_number, vehicle_model, year_made, driver_name, driver_license, driver_contact, vehicle_type, is_available, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`, out.VehicleNumber, out.VehicleModel, out.YearMade, out.DriverName, out.DriverLicense, out.DriverContact, out.VehicleType, out.IsAvailable, out.Note).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'ambulance.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateAmbulance(ctx context.Context, a domain.Actor, id string, in domain.AmbulanceInput, version int) (domain.Ambulance, error) {
	var out domain.Ambulance
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	var currentVersion int
	var currentAvailable bool
	err = tx.QueryRow(ctx, `SELECT version, is_available FROM ambulance WHERE id = $1 FOR UPDATE`, id).Scan(&currentVersion, &currentAvailable)
	if err != nil {
		return out, clinicalError(err)
	}
	if currentVersion != version {
		return out, domain.ErrStale
	}

	isAvailable := currentAvailable
	if in.IsAvailable != nil {
		isAvailable = *in.IsAvailable
	}

	err = tx.QueryRow(ctx, `
		UPDATE ambulance
		SET vehicle_number = $1, vehicle_model = $2, year_made = $3, driver_name = $4,
		    driver_license = $5, driver_contact = $6, vehicle_type = $7, is_available = $8,
		    note = $9, version = version + 1, updated_at = clock_timestamp()
		WHERE id = $10 AND version = $11
		RETURNING id, vehicle_number, vehicle_model, year_made, driver_name, driver_license, driver_contact, vehicle_type, is_available, note, version, created_at, updated_at
	`, in.VehicleNumber, in.VehicleModel, in.YearMade, in.DriverName, in.DriverLicense, in.DriverContact, in.VehicleType, isAvailable, in.Note, id, version).
		Scan(&out.ID, &out.VehicleNumber, &out.VehicleModel, &out.YearMade, &out.DriverName, &out.DriverLicense, &out.DriverContact, &out.VehicleType, &out.IsAvailable, &out.Note, &out.Version, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'ambulance.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) Ambulances(ctx context.Context, a domain.Actor, page int, availableOnly *bool) ([]domain.Ambulance, int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)

	whereClause := "WHERE 1=1"
	args := []any{}
	argIdx := 1
	if availableOnly != nil {
		whereClause += fmt.Sprintf(" AND is_available = $%d", argIdx)
		args = append(args, *availableOnly)
		argIdx++
	}

	var total int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM ambulance `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := 25
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, vehicle_number, vehicle_model, year_made, driver_name, driver_license, driver_contact, vehicle_type, is_available, note, version, created_at, updated_at
		FROM ambulance
		%s
		ORDER BY vehicle_number ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Ambulance{}
	for rows.Next() {
		var amb domain.Ambulance
		if err = rows.Scan(&amb.ID, &amb.VehicleNumber, &amb.VehicleModel, &amb.YearMade, &amb.DriverName, &amb.DriverLicense, &amb.DriverContact, &amb.VehicleType, &amb.IsAvailable, &amb.Note, &amb.Version, &amb.CreatedAt, &amb.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, amb)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action) VALUES ($1, 'ambulance_list.viewed')`, a.ID); err != nil {
		return nil, 0, err
	}
	return out, total, tx.Commit(ctx)
}

func (s Store) Ambulance(ctx context.Context, a domain.Actor, id string) (domain.Ambulance, error) {
	var amb domain.Ambulance
	err := s.DB.QueryRow(ctx, `
		SELECT id, vehicle_number, vehicle_model, year_made, driver_name, driver_license, driver_contact, vehicle_type, is_available, note, version, created_at, updated_at
		FROM ambulance
		WHERE id = $1
	`, id).Scan(&amb.ID, &amb.VehicleNumber, &amb.VehicleModel, &amb.YearMade, &amb.DriverName, &amb.DriverLicense, &amb.DriverContact, &amb.VehicleType, &amb.IsAvailable, &amb.Note, &amb.Version, &amb.CreatedAt, &amb.UpdatedAt)
	if err != nil {
		return amb, clinicalError(err)
	}
	return amb, nil
}

func (s Store) CreateAmbulanceCall(ctx context.Context, a domain.Actor, in domain.AmbulanceCallInput, key string) (domain.AmbulanceCall, error) {
	var out domain.AmbulanceCall
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	// Check patient exists
	var patientExists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient WHERE id = $1)`, in.PatientID).Scan(&patientExists); err != nil {
		return out, err
	}
	if !patientExists {
		return out, domain.ErrNotFound
	}

	// Lock ambulance to prevent concurrent double-dispatch
	var ambAvailable bool
	var defaultDriver, vehicleModel, vehicleNumber string
	err = tx.QueryRow(ctx, `
		SELECT is_available, driver_name, vehicle_model, vehicle_number 
		FROM ambulance 
		WHERE id = $1 
		FOR UPDATE
	`, in.AmbulanceID).Scan(&ambAvailable, &defaultDriver, &vehicleModel, &vehicleNumber)
	if err != nil {
		return out, clinicalError(err)
	}
	if !ambAvailable {
		return out, domain.ErrConflict
	}

	driverName := in.DriverName
	if driverName == "" {
		driverName = defaultDriver
	}

	out = domain.AmbulanceCall{
		AmbulanceID:    in.AmbulanceID,
		VehicleModel:   vehicleModel,
		VehicleNumber:  vehicleNumber,
		PatientID:      in.PatientID,
		DriverName:     driverName,
		CallDate:       in.CallDate,
		AmountMinor:    in.AmountMinor,
		Status:         "dispatched",
		PickupLocation: in.PickupLocation,
		Destination:    in.Destination,
		Notes:          in.Notes,
		Version:        1,
		CreatedBy:      a.ID,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO ambulance_call (ambulance_id, patient_id, driver_name, call_date, amount_minor, status, pickup_location, destination, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, out.AmbulanceID, out.PatientID, out.DriverName, out.CallDate, out.AmountMinor, out.Status, out.PickupLocation, out.Destination, out.Notes, a.ID).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	// Mark ambulance as in-use / unavailable
	if _, err = tx.Exec(ctx, `UPDATE ambulance SET is_available = false, updated_at = clock_timestamp() WHERE id = $1`, in.AmbulanceID); err != nil {
		return out, err
	}

	// Fetch patient name for return object
	_ = tx.QueryRow(ctx, `SELECT given_name || ' ' || family_name FROM patient WHERE id = $1`, in.PatientID).Scan(&out.PatientName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'ambulance_call.dispatched', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateAmbulanceCall(ctx context.Context, a domain.Actor, id string, in domain.AmbulanceCallUpdateInput) (domain.AmbulanceCall, error) {
	var out domain.AmbulanceCall
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	var currentVersion int
	var currentStatus, currentAmbulanceID string
	err = tx.QueryRow(ctx, `SELECT version, status, ambulance_id FROM ambulance_call WHERE id = $1 FOR UPDATE`, id).
		Scan(&currentVersion, &currentStatus, &currentAmbulanceID)
	if err != nil {
		return out, clinicalError(err)
	}
	if currentVersion != in.Version {
		return out, domain.ErrStale
	}

	// If vehicle changed while dispatched
	if in.AmbulanceID != currentAmbulanceID && currentStatus == "dispatched" {
		// Free old ambulance
		if _, err = tx.Exec(ctx, `UPDATE ambulance SET is_available = true, updated_at = clock_timestamp() WHERE id = $1`, currentAmbulanceID); err != nil {
			return out, err
		}
		// Claim new ambulance if call is still dispatched
		if in.Status == "dispatched" {
			var newAvail bool
			err = tx.QueryRow(ctx, `SELECT is_available FROM ambulance WHERE id = $1 FOR UPDATE`, in.AmbulanceID).Scan(&newAvail)
			if err != nil {
				return out, clinicalError(err)
			}
			if !newAvail {
				return out, domain.ErrConflict
			}
			if _, err = tx.Exec(ctx, `UPDATE ambulance SET is_available = false, updated_at = clock_timestamp() WHERE id = $1`, in.AmbulanceID); err != nil {
				return out, err
			}
		}
	} else if currentStatus == "dispatched" && (in.Status == "completed" || in.Status == "cancelled") {
		// Release current vehicle
		if _, err = tx.Exec(ctx, `UPDATE ambulance SET is_available = true, updated_at = clock_timestamp() WHERE id = $1`, currentAmbulanceID); err != nil {
			return out, err
		}
	}

	err = tx.QueryRow(ctx, `
		UPDATE ambulance_call
		SET ambulance_id = $1, driver_name = $2, call_date = $3, amount_minor = $4,
		    status = $5, pickup_location = $6, destination = $7, notes = $8,
		    version = version + 1, updated_at = clock_timestamp()
		WHERE id = $9 AND version = $10
		RETURNING id, ambulance_id, patient_id, driver_name, call_date, amount_minor, status, pickup_location, destination, notes, version, created_by, created_at, updated_at
	`, in.AmbulanceID, in.DriverName, in.CallDate, in.AmountMinor, in.Status, in.PickupLocation, in.Destination, in.Notes, id, in.Version).
		Scan(&out.ID, &out.AmbulanceID, &out.PatientID, &out.DriverName, &out.CallDate, &out.AmountMinor, &out.Status, &out.PickupLocation, &out.Destination, &out.Notes, &out.Version, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	// Fetch invoice linkage if exists
	var invoiceID string
	if err = tx.QueryRow(ctx, `SELECT invoice_id FROM ambulance_call_invoice WHERE call_id = $1`, id).Scan(&invoiceID); err == nil {
		out.InvoiceID = &invoiceID
	}

	_ = tx.QueryRow(ctx, `SELECT vehicle_model, vehicle_number FROM ambulance WHERE id = $1`, out.AmbulanceID).Scan(&out.VehicleModel, &out.VehicleNumber)
	_ = tx.QueryRow(ctx, `SELECT given_name || ' ' || family_name FROM patient WHERE id = $1`, out.PatientID).Scan(&out.PatientName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'ambulance_call.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) AmbulanceCalls(ctx context.Context, a domain.Actor, page int, status, patientID, ambulanceID string) ([]domain.AmbulanceCall, int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)

	whereClause := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if a.Role == "patient" {
		whereClause += fmt.Sprintf(" AND patient_portal_owner(p.id) = $%d", argIdx)
		args = append(args, a.ID)
		argIdx++
	} else if patientID != "" {
		whereClause += fmt.Sprintf(" AND c.patient_id = $%d", argIdx)
		args = append(args, patientID)
		argIdx++
	}

	if status != "" {
		whereClause += fmt.Sprintf(" AND c.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if ambulanceID != "" {
		whereClause += fmt.Sprintf(" AND c.ambulance_id = $%d", argIdx)
		args = append(args, ambulanceID)
		argIdx++
	}

	var total int
	err = tx.QueryRow(ctx, `
		SELECT count(*) 
		FROM ambulance_call c
		JOIN patient p ON p.id = c.patient_id
		`+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := 25
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT c.id, c.ambulance_id, a.vehicle_model, a.vehicle_number,
		       c.patient_id, p.given_name || ' ' || p.family_name,
		       c.driver_name, c.call_date, c.amount_minor, c.status,
		       c.pickup_location, c.destination, c.notes,
		       i.invoice_id, c.version, c.created_by, c.created_at, c.updated_at
		FROM ambulance_call c
		JOIN patient p ON p.id = c.patient_id
		JOIN ambulance a ON a.id = c.ambulance_id
		LEFT JOIN ambulance_call_invoice i ON i.call_id = c.id
		%s
		ORDER BY c.call_date DESC, c.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.AmbulanceCall{}
	for rows.Next() {
		var call domain.AmbulanceCall
		var inv *string
		if err = rows.Scan(
			&call.ID, &call.AmbulanceID, &call.VehicleModel, &call.VehicleNumber,
			&call.PatientID, &call.PatientName,
			&call.DriverName, &call.CallDate, &call.AmountMinor, &call.Status,
			&call.PickupLocation, &call.Destination, &call.Notes,
			&inv, &call.Version, &call.CreatedBy, &call.CreatedAt, &call.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		call.InvoiceID = inv
		out = append(out, call)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action) VALUES ($1, 'ambulance_call_list.viewed')`, a.ID); err != nil {
		return nil, 0, err
	}
	return out, total, tx.Commit(ctx)
}

func (s Store) AmbulanceCall(ctx context.Context, a domain.Actor, id string) (domain.AmbulanceCall, error) {
	var call domain.AmbulanceCall
	var inv *string
	var patientUserID string
	err := s.DB.QueryRow(ctx, `
		SELECT c.id, c.ambulance_id, a.vehicle_model, a.vehicle_number,
		       c.patient_id, p.given_name || ' ' || p.family_name, patient_portal_owner(p.id),
		       c.driver_name, c.call_date, c.amount_minor, c.status,
		       c.pickup_location, c.destination, c.notes,
		       i.invoice_id, c.version, c.created_by, c.created_at, c.updated_at
		FROM ambulance_call c
		JOIN patient p ON p.id = c.patient_id
		JOIN ambulance a ON a.id = c.ambulance_id
		LEFT JOIN ambulance_call_invoice i ON i.call_id = c.id
		WHERE c.id = $1
	`, id).Scan(
		&call.ID, &call.AmbulanceID, &call.VehicleModel, &call.VehicleNumber,
		&call.PatientID, &call.PatientName, &patientUserID,
		&call.DriverName, &call.CallDate, &call.AmountMinor, &call.Status,
		&call.PickupLocation, &call.Destination, &call.Notes,
		&inv, &call.Version, &call.CreatedBy, &call.CreatedAt, &call.UpdatedAt,
	)
	if err != nil {
		return call, clinicalError(err)
	}
	if a.Role == "patient" && patientUserID != a.ID {
		return call, domain.ErrForbidden
	}
	call.InvoiceID = inv
	return call, nil
}

func (s Store) BillAmbulanceCall(ctx context.Context, a domain.Actor, callID string, i domain.SourceInvoiceInput, key, date string) (domain.Invoice, error) {
	var out domain.Invoice
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	// Pharmacy/diagnostic lock helper
	if err = pharmacyLock(ctx, tx, a, key, "invoice"); err != nil {
		return out, err
	}

	var status, patientID, vehicleModel, vehicleNumber string
	var amountMinor int64
	err = tx.QueryRow(ctx, `
		SELECT c.status, c.patient_id, c.amount_minor, a.vehicle_model, a.vehicle_number
		FROM ambulance_call c
		JOIN ambulance a ON a.id = c.ambulance_id
		WHERE c.id = $1
		FOR UPDATE OF c
	`, callID).Scan(&status, &patientID, &amountMinor, &vehicleModel, &vehicleNumber)
	if err != nil {
		return out, clinicalError(err)
	}

	var alreadyLinked bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM service_invoice_link WHERE source_type='ambulance' AND source_id=$1)`, callID).Scan(&alreadyLinked); err != nil {
		return out, err
	}
	if alreadyLinked {
		return out, domain.ErrConflict
	}
	var existingInvoice, accountID string
	var discount int64
	err = tx.QueryRow(ctx, `
		SELECT invoice_id, account_id, discount_basis_points 
		FROM ambulance_call_invoice 
		WHERE call_id = $1
	`, callID).Scan(&existingInvoice, &accountID, &discount)
	if err == nil {
		if accountID != i.AccountID || discount != i.DiscountBasisPoints {
			return out, domain.ErrConflict
		}
		if err = tx.Commit(ctx); err != nil {
			return out, err
		}
		return s.Invoice(ctx, a, existingInvoice)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}

	if status == "cancelled" {
		return out, domain.ErrConflict
	}
	if amountMinor <= 0 {
		return out, domain.ErrValidation
	}

	var used bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM invoice WHERE created_by = $1 AND request_key = $2)`, a.ID, key).Scan(&used); err != nil {
		return out, err
	}
	if used {
		return out, domain.ErrConflict
	}

	hash := requestHash(struct {
		AmbulanceCall string
		Input         domain.SourceInvoiceInput
	}{callID, i})

	lineDescription := fmt.Sprintf("Ambulance Emergency Transport: %s (%s)", vehicleModel, vehicleNumber)
	invoiceID, err := insertSourceInvoice(ctx, tx, a, patientID, lineDescription, 1, amountMinor, i, key, date, hash)
	if err != nil {
		return out, err
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO ambulance_call_invoice (call_id, invoice_id, account_id, discount_basis_points, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`, callID, invoiceID, i.AccountID, i.DiscountBasisPoints, a.ID); err != nil {
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'ambulance_call.billed', $2)`, a.ID, callID); err != nil {
		return out, err
	}

	if err = tx.Commit(ctx); err != nil {
		return out, err
	}
	return s.Invoice(ctx, a, invoiceID)
}
