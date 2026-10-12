package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
)

// ─── Bed Occupancy & Assignments ──────────────────────────────────────────────

func (s Store) BedOccupancyReport(ctx context.Context) (domain.BedOccupancyReport, error) {
	var rep domain.BedOccupancyReport
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM hospital_bed WHERE active = true`).Scan(&rep.TotalBeds)
	if err != nil {
		return rep, err
	}
	err = s.DB.QueryRow(ctx, `
		SELECT COUNT(DISTINCT b.id) 
		FROM hospital_bed b
		WHERE b.active = true 
		  AND (
		    EXISTS (SELECT 1 FROM encounter e WHERE e.bed_id = b.id AND e.status = 'active')
		    OR EXISTS (SELECT 1 FROM bed_assignment ba WHERE ba.bed_id = b.id AND ba.assigned_to IS NULL)
		  )
	`).Scan(&rep.OccupiedBeds)
	if err != nil {
		return rep, err
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM hospital_bed WHERE active = true AND state = 'maintenance'`).Scan(&rep.MaintenanceBeds)
	if err != nil {
		return rep, err
	}
	err = s.DB.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM hospital_bed b 
		WHERE b.active = true 
		  AND b.state = 'ready' 
		  AND NOT EXISTS (SELECT 1 FROM encounter e WHERE e.bed_id = b.id AND e.status = 'active')
		  AND NOT EXISTS (SELECT 1 FROM bed_assignment ba WHERE ba.bed_id = b.id AND ba.assigned_to IS NULL)
	`).Scan(&rep.AvailableBeds)
	if err != nil {
		return rep, err
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM encounter WHERE status = 'active'`).Scan(&rep.ActiveAdmissions)
	if err != nil {
		return rep, err
	}
	if rep.TotalBeds > 0 {
		rep.OccupancyRate = float64(rep.OccupiedBeds) / float64(rep.TotalBeds) * 100.0
	}
	return rep, nil
}

func (s Store) AssignBed(ctx context.Context, a domain.Actor, input domain.AssignBedInput) (domain.BedAssignment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.BedAssignment{}, err
	}
	defer tx.Rollback(ctx)

	// 1. Verify patient exists
	var patientName, patientEmail, patientPhone, patientGender string
	err = tx.QueryRow(ctx, `
		SELECT (p.given_name || ' ' || p.family_name), COALESCE(pp.email, ''), COALESCE(p.phone, ''), COALESCE(pp.gender, '') 
		FROM patient p
		LEFT JOIN patient_profile pp ON pp.patient_id = p.id
		WHERE p.id = $1
	`, input.PatientID).Scan(&patientName, &patientEmail, &patientPhone, &patientGender)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}

	// 2. Verify encounter exists and validate ownership server-side!
	var encPatientID string
	var encStatus string
	var encKind string
	var encCurrentBedID *string
	var encVersion int
	var encNumber int64
	err = tx.QueryRow(ctx, `
		SELECT patient_id, status, kind, bed_id, version, number 
		FROM encounter 
		WHERE id = $1 FOR UPDATE OF encounter
	`, input.EncounterID).Scan(&encPatientID, &encStatus, &encKind, &encCurrentBedID, &encVersion, &encNumber)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}

	// Server-side ownership validation:
	if encPatientID != input.PatientID {
		return domain.BedAssignment{}, domain.ErrValidation
	}
	if encStatus != "active" {
		return domain.BedAssignment{}, domain.ErrConflict
	}
	if input.Version <= 0 || encVersion != input.Version {
		return domain.BedAssignment{}, domain.ErrStale
	}

	// 3. Verify target bed exists, is active, is ready, and lock for update
	var bedName, bedType, bedWard, bedState string
	var bedCharge int64
	err = tx.QueryRow(ctx, `
		SELECT b.name, t.name, COALESCE(b.ward_name, 'General Ward'), b.charge_minor, b.state
		FROM hospital_bed b
		JOIN bed_type t ON t.id = b.type_id
		WHERE b.id = $1 AND b.active = true
		FOR UPDATE OF b
	`, input.BedID).Scan(&bedName, &bedType, &bedWard, &bedCharge, &bedState)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}
	if bedState != "ready" {
		return domain.BedAssignment{}, domain.ErrConflict
	}

	// Check if bed is already occupied by ANY other active encounter
	var occupiedByOther bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM encounter 
			WHERE bed_id = $1 AND status = 'active' AND id != $2
		)
	`, input.BedID, input.EncounterID).Scan(&occupiedByOther)
	if err != nil {
		return domain.BedAssignment{}, err
	}
	if occupiedByOther {
		return domain.BedAssignment{}, domain.ErrConflict
	}

	// Check if bed has an active assignment in bed_assignment
	var hasActiveAssignment bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM bed_assignment
			WHERE bed_id = $1 AND assigned_to IS NULL AND encounter_id != $2
		)
	`, input.BedID, input.EncounterID).Scan(&hasActiveAssignment)
	if err != nil {
		return domain.BedAssignment{}, err
	}
	if hasActiveAssignment {
		return domain.BedAssignment{}, domain.ErrConflict
	}

	// 4. Handle transfer or initial assignment
	ipdNo := fmt.Sprintf("IPD-%d", encNumber)
	if encCurrentBedID != nil && *encCurrentBedID != "" {
		if *encCurrentBedID == input.BedID {
			// Bed is already actively assigned to this encounter - reject duplicate assignment
			return domain.BedAssignment{}, domain.ErrConflict
		}
		// Transfer: close old assignment
			_, err = tx.Exec(ctx, `
				UPDATE bed_assignment 
				SET assigned_to = clock_timestamp() 
				WHERE encounter_id = $1 AND assigned_to IS NULL
			`, input.EncounterID)
			if err != nil {
				return domain.BedAssignment{}, err
			}
			// Record transfer bed event
			_, err = tx.Exec(ctx, `
				INSERT INTO bed_event (encounter_id, kind, from_bed_id, to_bed_id, charge_minor, actor_id, reason, encounter_version)
				VALUES ($1, 'transfer', $2, $3, $4, $5, $6, $7)
			`, input.EncounterID, *encCurrentBedID, input.BedID, bedCharge, a.ID, input.Notes, encVersion+1)
			if err != nil {
				return domain.BedAssignment{}, err
			}
	} else {
		// First admission to a bed
		_, err = tx.Exec(ctx, `
			INSERT INTO bed_event (encounter_id, kind, to_bed_id, charge_minor, actor_id, reason, encounter_version)
			VALUES ($1, 'admission', $2, $3, $4, $5, $6)
		`, input.EncounterID, input.BedID, bedCharge, a.ID, input.Notes, encVersion+1)
		if err != nil {
			return domain.BedAssignment{}, err
		}
	}

	// 5. Update encounter
	_, err = tx.Exec(ctx, `
		UPDATE encounter 
		SET bed_id = $2, bed_charge_minor = $3, version = version + 1 
		WHERE id = $1
	`, input.EncounterID, input.BedID, bedCharge)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}

	// 6. Insert new bed_assignment
	var assign domain.BedAssignment
	assign.BedID = input.BedID
	assign.BedName = bedName
	assign.BedType = bedType
	assign.WardName = bedWard
	assign.EncounterID = input.EncounterID
	assign.IPDNo = ipdNo
	assign.PatientID = input.PatientID
	assign.PatientName = patientName
	assign.PatientEmail = patientEmail
	assign.PatientPhone = patientPhone
	assign.PatientGender = patientGender
	assign.Notes = input.Notes
	assign.AssignedBy = a.ID
	assign.Status = true

	err = tx.QueryRow(ctx, `
		INSERT INTO bed_assignment (
			bed_id, encounter_id, patient_id, notes, assigned_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, assigned_from, created_at
	`, input.BedID, input.EncounterID, input.PatientID, input.Notes, a.ID).Scan(
		&assign.ID, &assign.AssignedFrom, &assign.CreatedAt,
	)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}

	if err = pharmacyAudit(ctx, tx, a, "bed.assigned", assign.ID); err != nil {
		return domain.BedAssignment{}, err
	}

	assign.EncounterVersion = encVersion + 1

	if err := tx.Commit(ctx); err != nil {
		return domain.BedAssignment{}, err
	}
	return assign, nil
}

func (s Store) ListAllBedAssignments(ctx context.Context, a domain.Actor, search, status string, page, pageSize int) ([]domain.BedAssignment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	baseQuery := `
		FROM bed_assignment ba
		JOIN hospital_bed b ON b.id = ba.bed_id
		JOIN bed_type bt ON bt.id = b.type_id
		JOIN patient p ON p.id = ba.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = p.id
		JOIN encounter e ON e.id = ba.encounter_id
		WHERE (
			$1 = 'admin' OR $1 = 'receptionist'
			OR ($1 = 'doctor' AND (
				e.doctor_id = $2
				OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)
			))
			OR ($1 = 'nurse' AND (
				EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)
				OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)
			))
			OR ($1 = 'patient' AND patient_portal_owner(p.id) = $2)
		)
	`
	args := []any{a.Role, a.ID}
	argIdx := 3

	if status == "active" {
		baseQuery += ` AND ba.assigned_to IS NULL`
	} else if status == "discharged" {
		baseQuery += ` AND ba.assigned_to IS NOT NULL`
	}

	if search = strings.TrimSpace(search); search != "" {
		baseQuery += fmt.Sprintf(` AND (
			(p.given_name || ' ' || p.family_name) ILIKE $%d
			OR b.name ILIKE $%d
			OR b.ward_name ILIKE $%d
			OR ba.notes ILIKE $%d
			OR p.phone ILIKE $%d
			OR e.number::text ILIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	if err := s.DB.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT ba.id, ba.bed_id, b.name AS bed_name, bt.name AS bed_type, COALESCE(b.ward_name, 'General Ward') AS ward_name,
		       ba.encounter_id, COALESCE('IPD-' || e.number::text, e.id::text) AS ipd_no,
		       ba.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       COALESCE(pp.email, '') AS patient_email, COALESCE(p.phone, '') AS patient_phone,
		       COALESCE(pp.gender, '') AS patient_gender,
		       ba.assigned_from, ba.assigned_to, ba.notes, ba.assigned_by,
		       (ba.assigned_to IS NULL) AS status, ba.created_at, e.version
		%s
		ORDER BY ba.assigned_from DESC, ba.id
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.DB.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []domain.BedAssignment{}
	for rows.Next() {
		var a domain.BedAssignment
		if err := rows.Scan(
			&a.ID, &a.BedID, &a.BedName, &a.BedType, &a.WardName,
			&a.EncounterID, &a.IPDNo,
			&a.PatientID, &a.PatientName,
			&a.PatientEmail, &a.PatientPhone,
			&a.PatientGender,
			&a.AssignedFrom, &a.AssignedTo, &a.Notes, &a.AssignedBy,
			&a.Status, &a.CreatedAt, &a.EncounterVersion,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}

func (s Store) ListBedAssignments(ctx context.Context, encounterID string) ([]domain.BedAssignment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT ba.id, ba.bed_id, b.name AS bed_name, bt.name AS bed_type, COALESCE(b.ward_name, 'General Ward') AS ward_name,
		       ba.encounter_id, COALESCE('IPD-' || e.number::text, e.id::text) AS ipd_no,
		       ba.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       COALESCE(pp.email, '') AS patient_email, COALESCE(p.phone, '') AS patient_phone,
		       COALESCE(pp.gender, '') AS patient_gender,
		       ba.assigned_from, ba.assigned_to, ba.notes, ba.assigned_by,
		       (ba.assigned_to IS NULL) AS status, ba.created_at
		FROM bed_assignment ba
		JOIN hospital_bed b ON b.id = ba.bed_id
		JOIN bed_type bt ON bt.id = b.type_id
		JOIN patient p ON p.id = ba.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = p.id
		JOIN encounter e ON e.id = ba.encounter_id
		WHERE ba.encounter_id = $1
		ORDER BY ba.assigned_from DESC
	`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.BedAssignment
	for rows.Next() {
		var a domain.BedAssignment
		if err := rows.Scan(
			&a.ID, &a.BedID, &a.BedName, &a.BedType, &a.WardName,
			&a.EncounterID, &a.IPDNo,
			&a.PatientID, &a.PatientName,
			&a.PatientEmail, &a.PatientPhone,
			&a.PatientGender,
			&a.AssignedFrom, &a.AssignedTo, &a.Notes, &a.AssignedBy,
			&a.Status, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	if list == nil {
		list = []domain.BedAssignment{}
	}
	return list, nil
}

func (s Store) DischargeBedAssignment(ctx context.Context, a domain.Actor, id string, input domain.DischargeBedAssignmentInput) (domain.BedAssignment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.BedAssignment{}, err
	}
	defer tx.Rollback(ctx)

	var assign domain.BedAssignment
	var bedID, encID string
	var assignedTo *time.Time
	err = tx.QueryRow(ctx, `
		SELECT bed_id, encounter_id, patient_id, assigned_to
		FROM bed_assignment
		WHERE id = $1
	`, id).Scan(&bedID, &encID, &assign.PatientID, &assignedTo)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}
	if assignedTo != nil {
		return domain.BedAssignment{}, domain.ErrConflict
	}

	// Canonical clinical record authorization (unrelated staff forbidden)
	if err := s.AuthorizeClinicalRecord(ctx, a, assign.PatientID, encID, "admission"); err != nil {
		return domain.BedAssignment{}, err
	}

	// Enforce deadlock-free lock order: encounter -> hospital_bed -> bed_assignment
	// 1. Lock encounter FOR UPDATE
	var encVersion int
	var encStatus string
	var encNumber int64
	var encBedID *string
	err = tx.QueryRow(ctx, `
		SELECT version, status, number, bed_id
		FROM encounter
		WHERE id = $1 FOR UPDATE OF encounter
	`, encID).Scan(&encVersion, &encStatus, &encNumber, &encBedID)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}
	if input.Version <= 0 || encVersion != input.Version {
		return domain.BedAssignment{}, domain.ErrStale
	}

	// 2. Lock hospital_bed FOR UPDATE
	var bedName, bedType, bedWard string
	err = tx.QueryRow(ctx, `
		SELECT b.name, bt.name, COALESCE(b.ward_name, 'General Ward')
		FROM hospital_bed b
		JOIN bed_type bt ON bt.id = b.type_id
		WHERE b.id = $1 FOR UPDATE OF b
	`, bedID).Scan(&bedName, &bedType, &bedWard)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}

	// 3. Lock bed_assignment FOR UPDATE
	err = tx.QueryRow(ctx, `
		SELECT ba.notes, (p.given_name || ' ' || p.family_name), COALESCE(pp.email, ''), COALESCE(p.phone, ''),
		       COALESCE(pp.gender, ''), ba.assigned_from, ba.created_at, ba.assigned_to
		FROM bed_assignment ba
		JOIN patient p ON p.id = ba.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = p.id
		WHERE ba.id = $1 FOR UPDATE OF ba
	`, id).Scan(
		&assign.Notes, &assign.PatientName, &assign.PatientEmail, &assign.PatientPhone,
		&assign.PatientGender, &assign.AssignedFrom, &assign.CreatedAt, &assignedTo,
	)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}
	if assignedTo != nil {
		return domain.BedAssignment{}, domain.ErrConflict
	}

	now := time.Now().UTC()
	appendNote := assign.Notes
	if notes := strings.TrimSpace(input.DischargeNotes); notes != "" {
		if appendNote != "" {
			appendNote += " | " + notes
		} else {
			appendNote = notes
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE bed_assignment 
		SET assigned_to = $2, notes = $3 
		WHERE id = $1
	`, id, now, appendNote)
	if err != nil {
		return domain.BedAssignment{}, err
	}

	// Update encounter to release bed and increment version
	_, err = tx.Exec(ctx, `
		UPDATE encounter 
		SET bed_id = NULL, version = version + 1 
		WHERE id = $1 AND bed_id = $2
	`, encID, bedID)
	if err != nil {
		return domain.BedAssignment{}, clinicalError(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO bed_event (encounter_id, kind, from_bed_id, charge_minor, actor_id, reason, encounter_version)
		VALUES ($1, 'discharge', $2, 0, $3, $4, $5)
	`, encID, bedID, a.ID, input.DischargeNotes, encVersion+1)
	if err != nil {
		return domain.BedAssignment{}, err
	}

	if err = pharmacyAudit(ctx, tx, a, "bed.discharged", id); err != nil {
		return domain.BedAssignment{}, err
	}

	assign.ID = id
	assign.BedID = bedID
	assign.BedName = bedName
	assign.BedType = bedType
	assign.WardName = bedWard
	assign.EncounterID = encID
	assign.IPDNo = fmt.Sprintf("IPD-%d", encNumber)
	assign.AssignedTo = &now
	assign.Notes = appendNote
	assign.Status = false
	assign.EncounterVersion = encVersion + 1

	return assign, tx.Commit(ctx)
}

// ─── Care Team Delegation ─────────────────────────────────────────────────────

func (s Store) CareTeam(ctx context.Context, encounterID string) ([]domain.CareTeamMember, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT ct.id, ct.encounter_id, ct.staff_id, u.name AS staff_name,
		       ct.role_title, ct.assigned_at, ct.revoked_at, ct.assigned_by, ct.notes
		FROM encounter_care_team ct
		JOIN "user" u ON u.id = ct.staff_id
		WHERE ct.encounter_id = $1
		ORDER BY ct.assigned_at ASC
	`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.CareTeamMember
	for rows.Next() {
		var m domain.CareTeamMember
		if err := rows.Scan(
			&m.ID, &m.EncounterID, &m.StaffID, &m.StaffName,
			&m.RoleTitle, &m.AssignedAt, &m.RevokedAt, &m.AssignedBy, &m.Notes,
		); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	if list == nil {
		list = []domain.CareTeamMember{}
	}
	return list, nil
}

func (s Store) AddCareTeamMember(ctx context.Context, a domain.Actor, encounterID string, input domain.AddCareTeamMemberInput) (domain.CareTeamMember, error) {
	var m domain.CareTeamMember
	m.EncounterID = encounterID
	m.StaffID = input.StaffID
	m.RoleTitle = input.RoleTitle
	m.Notes = input.Notes
	m.AssignedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO encounter_care_team (
			encounter_id, staff_id, role_title, assigned_by, notes
		) SELECT $1,$2,$3,$4,$5 WHERE EXISTS(SELECT 1 FROM staff_access WHERE user_id=$2 AND active AND role IN ('doctor','nurse')) AND EXISTS(SELECT 1 FROM encounter WHERE id=$1 AND status='active')
		RETURNING id, assigned_at
	`, encounterID, input.StaffID, input.RoleTitle, a.ID, input.Notes,
	).Scan(&m.ID, &m.AssignedAt)
	if err != nil {
		return domain.CareTeamMember{}, clinicalError(err)
	}

	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.StaffID).Scan(&m.StaffName)
	return m, nil
}

func (s Store) RevokeCareTeamMember(ctx context.Context, a domain.Actor, memberID string) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE encounter_care_team
		SET revoked_at = clock_timestamp()
		WHERE id = $1 AND revoked_at IS NULL AND ($2='admin' OR EXISTS(SELECT 1 FROM encounter e WHERE e.id=encounter_care_team.encounter_id AND e.doctor_id=$3))
	`, memberID, a.Role, a.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ─── Diagnoses, Procedures, Attachments ───────────────────────────────────────

func (s Store) Diagnoses(ctx context.Context, encounterID string) ([]domain.EncounterDiagnosis, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, encounter_id, icd10_code, description, category, status, diagnosed_by, diagnosed_at, created_at
		FROM encounter_diagnosis
		WHERE encounter_id = $1
		ORDER BY diagnosed_at DESC
	`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.EncounterDiagnosis
	for rows.Next() {
		var d domain.EncounterDiagnosis
		if err := rows.Scan(
			&d.ID, &d.EncounterID, &d.ICD10Code, &d.Description, &d.Category, &d.Status,
			&d.DiagnosedBy, &d.DiagnosedAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []domain.EncounterDiagnosis{}
	}
	return list, nil
}

func (s Store) AddDiagnosis(ctx context.Context, a domain.Actor, encounterID string, input domain.AddDiagnosisInput) (domain.EncounterDiagnosis, error) {
	var d domain.EncounterDiagnosis
	d.EncounterID = encounterID
	d.ICD10Code = input.ICD10Code
	d.Description = input.Description
	d.Category = input.Category
	d.Status = input.Status
	d.DiagnosedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO encounter_diagnosis (
			encounter_id, icd10_code, description, category, status, diagnosed_by
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, diagnosed_at, created_at
	`, encounterID, input.ICD10Code, input.Description, input.Category, input.Status, a.ID,
	).Scan(&d.ID, &d.DiagnosedAt, &d.CreatedAt)
	if err != nil {
		return domain.EncounterDiagnosis{}, err
	}
	return d, nil
}

func (s Store) Procedures(ctx context.Context, encounterID string) ([]domain.EncounterProcedure, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT p.id, p.encounter_id, p.name, p.description, p.performed_by, u.name AS performed_by_name,
		       p.performed_at, p.anesthesia_type, p.findings, p.complications, p.created_at
		FROM encounter_procedure p
		JOIN "user" u ON u.id = p.performed_by
		WHERE p.encounter_id = $1
		ORDER BY p.performed_at DESC
	`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.EncounterProcedure
	for rows.Next() {
		var proc domain.EncounterProcedure
		if err := rows.Scan(
			&proc.ID, &proc.EncounterID, &proc.Name, &proc.Description, &proc.PerformedBy,
			&proc.PerformedByName, &proc.PerformedAt, &proc.AnesthesiaType, &proc.Findings,
			&proc.Complications, &proc.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, proc)
	}
	if list == nil {
		list = []domain.EncounterProcedure{}
	}
	return list, nil
}

func (s Store) AddProcedure(ctx context.Context, a domain.Actor, encounterID string, input domain.AddProcedureInput) (domain.EncounterProcedure, error) {
	var proc domain.EncounterProcedure
	proc.EncounterID = encounterID
	proc.Name = input.Name
	proc.Description = input.Description
	proc.PerformedBy = a.ID
	proc.AnesthesiaType = input.AnesthesiaType
	proc.Findings = input.Findings
	proc.Complications = input.Complications

	err := s.DB.QueryRow(ctx, `
		INSERT INTO encounter_procedure (
			encounter_id, name, description, performed_by, anesthesia_type, findings, complications
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, performed_at, created_at
	`, encounterID, input.Name, input.Description, a.ID, input.AnesthesiaType, input.Findings, input.Complications,
	).Scan(&proc.ID, &proc.PerformedAt, &proc.CreatedAt)
	if err != nil {
		return domain.EncounterProcedure{}, err
	}
	proc.PerformedByName = a.Name
	return proc, nil
}

func (s Store) Attachments(ctx context.Context, encounterID string) ([]domain.EncounterAttachment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, encounter_id, title, file_url, file_type, file_size_bytes, uploaded_by, uploaded_at
		FROM encounter_attachment
		WHERE encounter_id = $1
		ORDER BY uploaded_at DESC
	`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.EncounterAttachment
	for rows.Next() {
		var att domain.EncounterAttachment
		if err := rows.Scan(
			&att.ID, &att.EncounterID, &att.Title, &att.FileURL, &att.FileType,
			&att.FileSizeBytes, &att.UploadedBy, &att.UploadedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, att)
	}
	if list == nil {
		list = []domain.EncounterAttachment{}
	}
	return list, nil
}

func (s Store) AddAttachment(ctx context.Context, a domain.Actor, encounterID string, input domain.AddAttachmentInput) (domain.EncounterAttachment, error) {
	var att domain.EncounterAttachment
	att.EncounterID = encounterID
	att.Title = input.Title
	att.FileURL = input.FileURL
	att.FileType = input.FileType
	att.FileSizeBytes = input.FileSizeBytes
	att.UploadedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO encounter_attachment (
			encounter_id, title, file_url, file_type, file_size_bytes, uploaded_by
		) SELECT $1,$2,$3,sa.mime_type,sa.file_size_bytes,$4 FROM secure_attachment sa JOIN encounter e ON e.id=$1 WHERE sa.encounter_id=e.id AND sa.patient_id=e.patient_id AND '/v1/attachments/'||sa.token||'/content'=$3
		RETURNING id, uploaded_at, file_type, file_size_bytes
	`, encounterID, input.Title, input.FileURL, a.ID,
	).Scan(&att.ID, &att.UploadedAt, &att.FileType, &att.FileSizeBytes)
	if err != nil {
		return domain.EncounterAttachment{}, clinicalError(err)
	}
	return att, nil
}

// ─── IPD Admission Details ────────────────────────────────────────────────────

func (s Store) IPDAdmissionDetails(ctx context.Context, encounterID string) (domain.IPDAdmissionDetails, error) {
	var d domain.IPDAdmissionDetails
	d.EncounterID = encounterID
	err := s.DB.QueryRow(ctx, `
		SELECT package_name, package_charge_minor, insurance_policy_number, insurance_provider,
		       guardian_name, guardian_relation, guardian_phone, guardian_address, created_at, updated_at
		FROM ipd_admission_details
		WHERE encounter_id = $1
	`, encounterID).Scan(
		&d.PackageName, &d.PackageChargeMinor, &d.InsurancePolicyNumber, &d.InsuranceProvider,
		&d.GuardianName, &d.GuardianRelation, &d.GuardianPhone, &d.GuardianAddress,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		d.CreatedAt = time.Now().UTC()
		d.UpdatedAt = time.Now().UTC()
		return d, nil
	}
	if err != nil {
		return domain.IPDAdmissionDetails{}, err
	}
	return d, nil
}

func (s Store) SaveIPDAdmissionDetails(ctx context.Context, a domain.Actor, details domain.IPDAdmissionDetails) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO ipd_admission_details (
			encounter_id, package_name, package_charge_minor, insurance_policy_number, insurance_provider,
			guardian_name, guardian_relation, guardian_phone, guardian_address, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, clock_timestamp(), clock_timestamp())
		ON CONFLICT (encounter_id) DO UPDATE SET
			package_name = EXCLUDED.package_name,
			package_charge_minor = EXCLUDED.package_charge_minor,
			insurance_policy_number = EXCLUDED.insurance_policy_number,
			insurance_provider = EXCLUDED.insurance_provider,
			guardian_name = EXCLUDED.guardian_name,
			guardian_relation = EXCLUDED.guardian_relation,
			guardian_phone = EXCLUDED.guardian_phone,
			guardian_address = EXCLUDED.guardian_address,
			updated_at = clock_timestamp()
	`, details.EncounterID, details.PackageName, details.PackageChargeMinor, details.InsurancePolicyNumber,
		details.InsuranceProvider, details.GuardianName, details.GuardianRelation, details.GuardianPhone,
		details.GuardianAddress)
	return err
}

// ─── Encounter Billing & Clearance ────────────────────────────────────────────

func (s Store) EncounterBilling(ctx context.Context, encounterID string) (domain.EncounterBilling, error) {
	var b domain.EncounterBilling
	var invoiceID *string
	err := s.DB.QueryRow(ctx, `
		SELECT encounter_id::text, bed_days, bed_total_minor, doctor_fee_minor, procedure_fee_minor,
		       other_charges_minor, total_minor, invoice_id::text, financial_clearance,
		       cleared_by, cleared_at, waiver_reason, created_at, updated_at
		FROM encounter_billing
		WHERE encounter_id = $1
	`, encounterID).Scan(
		&b.EncounterID, &b.BedDays, &b.BedTotalMinor, &b.DoctorFeeMinor, &b.ProcedureFeeMinor,
		&b.OtherChargesMinor, &b.TotalMinor, &invoiceID, &b.FinancialClearance,
		&b.ClearedBy, &b.ClearedAt, &b.WaiverReason, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Calculate default bed charge from encounter bed_charge_minor
		var bedCharge int64
		_ = s.DB.QueryRow(ctx, `SELECT COALESCE(bed_charge_minor, 0) FROM encounter WHERE id = $1`, encounterID).Scan(&bedCharge)
		return domain.EncounterBilling{
			EncounterID:   encounterID,
			BedDays:       1,
			BedTotalMinor: bedCharge,
			TotalMinor:    bedCharge,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}, nil
	}
	if err != nil {
		return domain.EncounterBilling{}, err
	}
	b.InvoiceID = invoiceID
	return b, nil
}

func (s Store) UpdateEncounterBilling(ctx context.Context, a domain.Actor, encounterID string, input domain.UpdateEncounterBillingInput) (domain.EncounterBilling, error) {
	totalMinor := input.BedTotalMinor + input.DoctorFeeMinor + input.ProcedureFeeMinor + input.OtherChargesMinor
	var b domain.EncounterBilling
	var invoiceID *string
	err := s.DB.QueryRow(ctx, `
		INSERT INTO encounter_billing (
			encounter_id, bed_days, bed_total_minor, doctor_fee_minor, procedure_fee_minor,
			other_charges_minor, total_minor, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, clock_timestamp(), clock_timestamp())
		ON CONFLICT (encounter_id) DO UPDATE SET
			bed_days = EXCLUDED.bed_days,
			bed_total_minor = EXCLUDED.bed_total_minor,
			doctor_fee_minor = EXCLUDED.doctor_fee_minor,
			procedure_fee_minor = EXCLUDED.procedure_fee_minor,
			other_charges_minor = EXCLUDED.other_charges_minor,
			total_minor = EXCLUDED.total_minor,
			updated_at = clock_timestamp()
 WHERE encounter_billing.invoice_id IS NULL AND NOT encounter_billing.financial_clearance
		RETURNING encounter_id::text, bed_days, bed_total_minor, doctor_fee_minor, procedure_fee_minor,
		          other_charges_minor, total_minor, invoice_id::text, financial_clearance,
		          cleared_by, cleared_at, waiver_reason, created_at, updated_at
	`, encounterID, input.BedDays, input.BedTotalMinor, input.DoctorFeeMinor, input.ProcedureFeeMinor,
		input.OtherChargesMinor, totalMinor,
	).Scan(
		&b.EncounterID, &b.BedDays, &b.BedTotalMinor, &b.DoctorFeeMinor, &b.ProcedureFeeMinor,
		&b.OtherChargesMinor, &b.TotalMinor, &invoiceID, &b.FinancialClearance,
		&b.ClearedBy, &b.ClearedAt, &b.WaiverReason, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return domain.EncounterBilling{}, clinicalError(err)
	}
	b.InvoiceID = invoiceID
	return b, nil
}

func (s Store) GrantFinancialClearance(ctx context.Context, a domain.Actor, encounterID, reason string) (domain.EncounterBilling, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.EncounterBilling{}, err
	}
	defer tx.Rollback(ctx)
	var cleared bool
	var invoiceID *string
	var total int64
	err = tx.QueryRow(ctx, `SELECT financial_clearance,invoice_id,total_minor FROM encounter_billing WHERE encounter_id=$1 FOR UPDATE`, encounterID).Scan(&cleared, &invoiceID, &total)
	if err != nil {
		return domain.EncounterBilling{}, clinicalError(err)
	}
	if !cleared {
		paid := total == 0
		if invoiceID != nil {
			if err = tx.QueryRow(ctx, `SELECT paid_minor>=total_minor FROM invoice WHERE id=$1 FOR UPDATE`, *invoiceID).Scan(&paid); err != nil {
				return domain.EncounterBilling{}, err
			}
		}
		if !paid && len(strings.TrimSpace(reason)) < 10 {
			return domain.EncounterBilling{}, domain.ErrStale
		}
		if _, err = tx.Exec(ctx, `UPDATE encounter_billing SET financial_clearance=true,cleared_by=$2,cleared_at=clock_timestamp(),waiver_reason=$3,updated_at=clock_timestamp() WHERE encounter_id=$1`, encounterID, a.ID, strings.TrimSpace(reason)); err != nil {
			return domain.EncounterBilling{}, err
		}
		if err = pharmacyAudit(ctx, tx, a, "encounter.financial_clearance", encounterID); err != nil {
			return domain.EncounterBilling{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.EncounterBilling{}, err
	}
	return s.EncounterBilling(ctx, encounterID)
}

func (s Store) LinkEncounterInvoice(ctx context.Context, a domain.Actor, encounterID, invoiceID string) (domain.EncounterBilling, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.EncounterBilling{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = linkOperationalInvoice(ctx, tx, a, "encounter", encounterID, invoiceID, nil); err != nil {
		return domain.EncounterBilling{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE encounter_billing SET invoice_id=$2,updated_at=clock_timestamp() WHERE encounter_id=$1`, encounterID, invoiceID); err != nil {
		return domain.EncounterBilling{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.EncounterBilling{}, err
	}
	return s.EncounterBilling(ctx, encounterID)
}

func (s Store) DischargeSummary(ctx context.Context, encounterID string) (domain.DischargeSummary, error) {
	var d domain.DischargeSummary
	var followUpDate *time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT ds.encounter_id, ds.admission_diagnosis, ds.discharge_diagnosis, ds.condition_at_discharge,
		       ds.hospital_course, ds.surgical_procedures, ds.discharge_medications, ds.follow_up_advice,
		       ds.follow_up_date, ds.signed_by, u.name AS signed_by_name, ds.signed_at, ds.created_at, ds.updated_at
		FROM discharge_summary ds
		JOIN "user" u ON u.id = ds.signed_by
		WHERE ds.encounter_id = $1
	`, encounterID).Scan(
		&d.EncounterID, &d.AdmissionDiagnosis, &d.DischargeDiagnosis, &d.ConditionAtDischarge,
		&d.HospitalCourse, &d.SurgicalProcedures, &d.DischargeMedications, &d.FollowUpAdvice,
		&followUpDate, &d.SignedBy, &d.SignedByName, &d.SignedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DischargeSummary{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DischargeSummary{}, err
	}
	if followUpDate != nil {
		str := followUpDate.Format("2006-01-02")
		d.FollowUpDate = &str
	}
	return d, nil
}

func (s Store) SaveDischargeSummary(ctx context.Context, a domain.Actor, summary domain.DischargeSummary) error {
	var followUpDate *string
	if summary.FollowUpDate != nil && *summary.FollowUpDate != "" {
		followUpDate = summary.FollowUpDate
	}
	_, err := s.DB.Exec(ctx, `
		INSERT INTO discharge_summary (
			encounter_id, admission_diagnosis, discharge_diagnosis, condition_at_discharge,
			hospital_course, surgical_procedures, discharge_medications, follow_up_advice,
			follow_up_date, signed_by, signed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date, $10, clock_timestamp(), clock_timestamp(), clock_timestamp())
	`, summary.EncounterID, summary.AdmissionDiagnosis, summary.DischargeDiagnosis,
		summary.ConditionAtDischarge, summary.HospitalCourse, summary.SurgicalProcedures,
		summary.DischargeMedications, summary.FollowUpAdvice, followUpDate, a.ID)
	return clinicalError(err)
}

// ─── OPD Follow-ups & Referrals ───────────────────────────────────────────────

func (s Store) OPDFollowUps(ctx context.Context, patientID string) ([]domain.OPDFollowUp, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT f.id, f.encounter_id, f.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       f.doctor_id, u.name AS doctor_name, f.follow_up_date::text, f.notes, f.status, f.created_at
		FROM opd_follow_up f
		JOIN patient p ON p.id = f.patient_id
		JOIN "user" u ON u.id = f.doctor_id
		WHERE canonical_patient_id(f.patient_id) = canonical_patient_id($1::uuid)
		ORDER BY f.follow_up_date DESC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.OPDFollowUp
	for rows.Next() {
		var f domain.OPDFollowUp
		if err := rows.Scan(
			&f.ID, &f.EncounterID, &f.PatientID, &f.PatientName,
			&f.DoctorID, &f.DoctorName, &f.FollowUpDate, &f.Notes, &f.Status, &f.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	if list == nil {
		list = []domain.OPDFollowUp{}
	}
	return list, nil
}

func (s Store) CreateOPDFollowUp(ctx context.Context, a domain.Actor, input domain.CreateOPDFollowUpInput) (domain.OPDFollowUp, error) {
	var f domain.OPDFollowUp
	f.EncounterID = input.EncounterID
	f.PatientID = input.PatientID
	f.DoctorID = input.DoctorID
	f.FollowUpDate = input.FollowUpDate
	f.Notes = input.Notes
	f.Status = "scheduled"

	err := s.DB.QueryRow(ctx, `
		INSERT INTO opd_follow_up (
			encounter_id, patient_id, doctor_id, follow_up_date, notes, status
		) VALUES ($1, $2, $3, $4::date, $5, $6)
		RETURNING id, created_at
	`, input.EncounterID, input.PatientID, input.DoctorID, input.FollowUpDate, input.Notes, f.Status,
	).Scan(&f.ID, &f.CreatedAt)
	if err != nil {
		return domain.OPDFollowUp{}, err
	}

	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&f.PatientName)
	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.DoctorID).Scan(&f.DoctorName)
	return f, nil
}

func (s Store) PatientReferrals(ctx context.Context, patientID string) ([]domain.PatientReferral, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT r.id, r.encounter_id, r.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       r.referral_type, r.external_facility, r.department, r.reason, r.referred_by, r.referred_at
		FROM patient_referral r
		JOIN patient p ON p.id = r.patient_id
		WHERE canonical_patient_id(r.patient_id) = canonical_patient_id($1::uuid)
		ORDER BY r.referred_at DESC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PatientReferral
	for rows.Next() {
		var ref domain.PatientReferral
		if err := rows.Scan(
			&ref.ID, &ref.EncounterID, &ref.PatientID, &ref.PatientName,
			&ref.ReferralType, &ref.ExternalFacility, &ref.Department, &ref.Reason,
			&ref.ReferredBy, &ref.ReferredAt,
		); err != nil {
			return nil, err
		}
		list = append(list, ref)
	}
	if list == nil {
		list = []domain.PatientReferral{}
	}
	return list, nil
}

func (s Store) CreatePatientReferral(ctx context.Context, a domain.Actor, input domain.CreatePatientReferralInput) (domain.PatientReferral, error) {
	var ref domain.PatientReferral
	ref.EncounterID = input.EncounterID
	ref.PatientID = input.PatientID
	ref.ReferralType = input.ReferralType
	ref.ExternalFacility = input.ExternalFacility
	ref.Department = input.Department
	ref.Reason = input.Reason
	ref.ReferredBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO patient_referral (
			encounter_id, patient_id, referral_type, external_facility, department, reason, referred_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, referred_at
	`, input.EncounterID, input.PatientID, input.ReferralType, input.ExternalFacility, input.Department, input.Reason, a.ID,
	).Scan(&ref.ID, &ref.ReferredAt)
	if err != nil {
		return domain.PatientReferral{}, err
	}

	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&ref.PatientName)
	return ref, nil
}

// ─── Odontogram (Dental Chart) ────────────────────────────────────────────────

func (s Store) ValidatePatientEncounter(ctx context.Context, patientID, encounterID string) error {
	if encounterID == "" {
		return nil
	}
	var exists bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM encounter
			WHERE id = $1::uuid AND canonical_patient_id(patient_id) = canonical_patient_id($2::uuid)
		)
	`, encounterID, patientID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrValidation
	}
	return nil
}

func (s Store) ValidateDoctor(ctx context.Context, doctorID string) error {
	var exists bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "user" u
			WHERE u.id = $1
		)
	`, doctorID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrValidation
	}
	return nil
}

func (s Store) Odontogram(ctx context.Context, patientID string) ([]domain.OdontogramEntry, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, patient_id, chart_id, encounter_id::text, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
		FROM patient_odontogram_entry
		WHERE canonical_patient_id(patient_id) = canonical_patient_id($1::uuid)
		ORDER BY tooth_number ASC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.OdontogramEntry
	for rows.Next() {
		var e domain.OdontogramEntry
		var encID *string
		if err := rows.Scan(
			&e.ID, &e.PatientID, &e.ChartID, &encID, &e.ToothNumber, &e.Condition, &e.ProcedureNotes,
			&e.DiagnosedBy, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		e.EncounterID = encID
		list = append(list, e)
	}
	if list == nil {
		list = []domain.OdontogramEntry{}
	}
	return list, nil
}

func (s Store) SetToothCondition(ctx context.Context, a domain.Actor, patientID string, input domain.SetToothConditionInput) (domain.OdontogramEntry, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.OdontogramEntry{}, err
	}
	defer tx.Rollback(ctx)

	// Shared lock on legend row to coordinate with concurrent legend mutations (G2)
	if err := lockAndValidateLegendCodes(ctx, tx, []string{input.Condition}); err != nil {
		return domain.OdontogramEntry{}, err
	}

	var chartID string
	if input.ChartID != nil && *input.ChartID != "" {
		// Parent-first lock: lock odontogram_chart before touching any child rows
		err = tx.QueryRow(ctx, `
			SELECT id FROM odontogram_chart
			WHERE id = $1 AND canonical_patient_id(patient_id) = canonical_patient_id($2::uuid)
			FOR UPDATE
		`, *input.ChartID, patientID).Scan(&chartID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.OdontogramEntry{}, domain.ErrNotFound
			}
			return domain.OdontogramEntry{}, err
		}
	} else {
		// Target ONLY an explicitly designated legacy chart (is_legacy = true)
		err = tx.QueryRow(ctx, `
			SELECT id FROM odontogram_chart
			WHERE canonical_patient_id(patient_id) = canonical_patient_id($1::uuid) AND is_legacy = true
			FOR UPDATE
		`, patientID).Scan(&chartID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Safely create designated legacy chart serialized with partial unique index
				err = tx.QueryRow(ctx, `
					INSERT INTO odontogram_chart (
						patient_id, doctor_id, description, odontogram,
						version, created_by, updated_by, created_at, updated_at, is_legacy
					) VALUES ($1, $2, 'Legacy clinical examination', '{}'::jsonb, 1, $2, $2, clock_timestamp(), clock_timestamp(), true)
					ON CONFLICT (patient_id) WHERE is_legacy = true DO NOTHING
					RETURNING id
				`, patientID, a.ID).Scan(&chartID)
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						// Concurrent transaction inserted the legacy chart; select and lock it
						err = tx.QueryRow(ctx, `
							SELECT id FROM odontogram_chart
							WHERE canonical_patient_id(patient_id) = canonical_patient_id($1::uuid) AND is_legacy = true
							FOR UPDATE
						`, patientID).Scan(&chartID)
						if err != nil {
							return domain.OdontogramEntry{}, err
						}
					} else {
						return domain.OdontogramEntry{}, err
					}
				}
			} else {
				return domain.OdontogramEntry{}, err
			}
		}
	}

	var e domain.OdontogramEntry
	var encID *string
	err = tx.QueryRow(ctx, `
		INSERT INTO patient_odontogram_entry (
			patient_id, chart_id, encounter_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, clock_timestamp(), clock_timestamp())
		ON CONFLICT (chart_id, tooth_number) DO UPDATE SET
			encounter_id = COALESCE(EXCLUDED.encounter_id, patient_odontogram_entry.encounter_id),
			condition = EXCLUDED.condition,
			procedure_notes = EXCLUDED.procedure_notes,
			diagnosed_by = EXCLUDED.diagnosed_by,
			updated_at = clock_timestamp()
		RETURNING id, patient_id, chart_id, encounter_id::text, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
	`, patientID, chartID, input.EncounterID, input.ToothNumber, input.Condition, input.ProcedureNotes, a.ID).Scan(
		&e.ID, &e.PatientID, &e.ChartID, &encID, &e.ToothNumber, &e.Condition, &e.ProcedureNotes,
		&e.DiagnosedBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return domain.OdontogramEntry{}, err
	}
	e.EncounterID = encID

	// Atomically update chart JSON and version to prevent divergence
	jsonKey := strconv.Itoa(input.ToothNumber)
	_, err = tx.Exec(ctx, `
		UPDATE odontogram_chart
		SET odontogram = odontogram || jsonb_build_object($1::text, $2::text),
			version = version + 1,
			updated_by = $3,
			updated_at = clock_timestamp()
		WHERE id = $4
	`, jsonKey, input.Condition, a.ID, chartID)
	if err != nil {
		return domain.OdontogramEntry{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OdontogramEntry{}, err
	}
	return e, nil
}

func (s Store) CreateOdontogramChart(ctx context.Context, a domain.Actor, input domain.SaveOdontogramChartInput) (domain.OdontogramChart, error) {
	if err := s.ValidatePatientEncounter(ctx, input.PatientID, pointerString(input.EncounterID)); err != nil {
		return domain.OdontogramChart{}, err
	}
	if err := s.ValidateDoctor(ctx, input.DoctorID); err != nil {
		return domain.OdontogramChart{}, err
	}

	strMap := make(map[string]string, len(input.Odontogram))
	var condCodes []string
	for k, v := range input.Odontogram {
		strMap[strconv.Itoa(k)] = v
		condCodes = append(condCodes, v)
	}
	jsonBytes, err := json.Marshal(strMap)
	if err != nil {
		return domain.OdontogramChart{}, domain.ErrValidation
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.OdontogramChart{}, err
	}
	defer tx.Rollback(ctx)

	// Acquire FOR SHARE locks on all referenced condition codes to coordinate with concurrent legend operations (G2)
	if err := lockAndValidateLegendCodes(ctx, tx, condCodes); err != nil {
		return domain.OdontogramChart{}, err
	}

	var chartID string
	var encID *string
	if input.EncounterID != nil && *input.EncounterID != "" {
		encID = input.EncounterID
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO odontogram_chart (
			patient_id, doctor_id, encounter_id, description, odontogram,
			version, created_by, updated_by, created_at, updated_at, is_legacy
		) VALUES ($1, $2, $3, $4, $5, 1, $6, $6, clock_timestamp(), clock_timestamp(), false)
		RETURNING id
	`, input.PatientID, input.DoctorID, encID, input.Description, jsonBytes, a.ID).Scan(&chartID)
	if err != nil {
		return domain.OdontogramChart{}, err
	}

	for toothNum, cond := range input.Odontogram {
		_, err = tx.Exec(ctx, `
			INSERT INTO patient_odontogram_entry (
				patient_id, chart_id, encounter_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, clock_timestamp(), clock_timestamp())
			ON CONFLICT (chart_id, tooth_number) DO UPDATE SET
				condition = EXCLUDED.condition,
				procedure_notes = EXCLUDED.procedure_notes,
				diagnosed_by = EXCLUDED.diagnosed_by,
				updated_at = clock_timestamp()
		`, input.PatientID, chartID, encID, toothNum, cond, input.Description, a.ID)
		if err != nil {
			return domain.OdontogramChart{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OdontogramChart{}, err
	}

	return s.GetOdontogramChart(ctx, chartID)
}

func (s Store) UpdateOdontogramChart(ctx context.Context, a domain.Actor, id string, input domain.SaveOdontogramChartInput) (domain.OdontogramChart, error) {
	if input.Version <= 0 {
		return domain.OdontogramChart{}, domain.ErrValidation
	}
	if err := s.ValidatePatientEncounter(ctx, input.PatientID, pointerString(input.EncounterID)); err != nil {
		return domain.OdontogramChart{}, err
	}
	if err := s.ValidateDoctor(ctx, input.DoctorID); err != nil {
		return domain.OdontogramChart{}, err
	}

	strMap := make(map[string]string, len(input.Odontogram))
	var condCodes []string
	for k, v := range input.Odontogram {
		strMap[strconv.Itoa(k)] = v
		condCodes = append(condCodes, v)
	}
	jsonBytes, err := json.Marshal(strMap)
	if err != nil {
		return domain.OdontogramChart{}, domain.ErrValidation
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.OdontogramChart{}, err
	}
	defer tx.Rollback(ctx)

	// Acquire FOR SHARE locks on all referenced condition codes to coordinate with concurrent legend operations (G2)
	if err := lockAndValidateLegendCodes(ctx, tx, condCodes); err != nil {
		return domain.OdontogramChart{}, err
	}

	var currentVersion int
	var patientID string
	err = tx.QueryRow(ctx, `
		SELECT version, patient_id::text
		FROM odontogram_chart
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&currentVersion, &patientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OdontogramChart{}, domain.ErrNotFound
		}
		return domain.OdontogramChart{}, err
	}

	// Stale edit concurrency check: input.Version must match currentVersion
	if input.Version != currentVersion {
		return domain.OdontogramChart{}, domain.ErrConflict
	}

	var encID *string
	if input.EncounterID != nil && *input.EncounterID != "" {
		encID = input.EncounterID
	}

	cmd, err := tx.Exec(ctx, `
		UPDATE odontogram_chart
		SET doctor_id = $2,
			encounter_id = $3,
			description = $4,
			odontogram = $5,
			version = version + 1,
			updated_by = $6,
			updated_at = clock_timestamp()
		WHERE id = $1 AND version = $7
	`, id, input.DoctorID, encID, input.Description, jsonBytes, a.ID, currentVersion)
	if err != nil {
		return domain.OdontogramChart{}, err
	}
	if cmd.RowsAffected() == 0 {
		return domain.OdontogramChart{}, domain.ErrConflict
	}

	// Delete child rows for teeth no longer present in input.Odontogram
	var toothNumbers []int
	for tNum := range input.Odontogram {
		toothNumbers = append(toothNumbers, tNum)
	}
	if len(toothNumbers) == 0 {
		if _, err = tx.Exec(ctx, `DELETE FROM patient_odontogram_entry WHERE chart_id = $1`, id); err != nil {
			return domain.OdontogramChart{}, err
		}
	} else {
		if _, err = tx.Exec(ctx, `DELETE FROM patient_odontogram_entry WHERE chart_id = $1 AND NOT (tooth_number = ANY($2))`, id, toothNumbers); err != nil {
			return domain.OdontogramChart{}, err
		}
	}

	// Upsert child rows while preserving original tooth notes, author, and creation timestamp
	// when condition has not changed!
	for toothNum, cond := range input.Odontogram {
		_, err = tx.Exec(ctx, `
			INSERT INTO patient_odontogram_entry (
				patient_id, chart_id, encounter_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, clock_timestamp(), clock_timestamp())
			ON CONFLICT (chart_id, tooth_number) DO UPDATE SET
				encounter_id = CASE
					WHEN patient_odontogram_entry.condition = EXCLUDED.condition THEN patient_odontogram_entry.encounter_id
					ELSE COALESCE(EXCLUDED.encounter_id, patient_odontogram_entry.encounter_id)
				END,
				condition = EXCLUDED.condition,
				procedure_notes = CASE
					WHEN patient_odontogram_entry.condition = EXCLUDED.condition THEN patient_odontogram_entry.procedure_notes
					ELSE EXCLUDED.procedure_notes
				END,
				diagnosed_by = CASE
					WHEN patient_odontogram_entry.condition = EXCLUDED.condition THEN patient_odontogram_entry.diagnosed_by
					ELSE EXCLUDED.diagnosed_by
				END,
				updated_at = clock_timestamp()
		`, patientID, id, encID, toothNum, cond, input.Description, a.ID)
		if err != nil {
			return domain.OdontogramChart{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OdontogramChart{}, err
	}

	return s.GetOdontogramChart(ctx, id)
}

func (s Store) GetOdontogramChart(ctx context.Context, id string) (domain.OdontogramChart, error) {
	var c domain.OdontogramChart
	var encID *string
	var rawJSON []byte
	err := s.DB.QueryRow(ctx, `
		SELECT
			c.id, c.patient_id,
			TRIM(p.given_name || ' ' || p.family_name) AS patient_name,
			p.medical_record_number::text AS patient_mrn,
			c.doctor_id,
			COALESCE(u.name, 'Doctor') AS doctor_name,
			c.encounter_id::text,
			c.description,
			c.odontogram,
			c.version,
			c.created_by,
			c.updated_by,
			c.created_at,
			c.updated_at,
			c.is_legacy,
			c.legacy_provenance
		FROM odontogram_chart c
		JOIN patient p ON p.id = c.patient_id
		LEFT JOIN "user" u ON u.id = c.doctor_id
		WHERE c.id = $1
	`, id).Scan(
		&c.ID, &c.PatientID, &c.PatientName, &c.PatientMRN,
		&c.DoctorID, &c.DoctorName, &encID, &c.Description,
		&rawJSON, &c.Version, &c.CreatedBy, &c.UpdatedBy,
		&c.CreatedAt, &c.UpdatedAt, &c.IsLegacy, &c.LegacyProvenance,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OdontogramChart{}, domain.ErrNotFound
		}
		return domain.OdontogramChart{}, err
	}
	c.EncounterID = encID

	strMap := make(map[string]string)
	if len(rawJSON) > 0 {
		_ = json.Unmarshal(rawJSON, &strMap)
	}
	c.Odontogram = make(map[int]string, len(strMap))
	for k, v := range strMap {
		if num, err := strconv.Atoi(k); err == nil {
			c.Odontogram[num] = v
		}
	}
	return c, nil
}

func (s Store) ListOdontogramCharts(ctx context.Context, a domain.Actor, patientID, search string, page, limit int) ([]domain.OdontogramChart, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	offset := (page - 1) * limit

	var whereClauses []string
	var args []any
	idx := 1

	// Role scoping
	if a.Role == "patient" {
		whereClauses = append(whereClauses, `patient_portal_owner(c.patient_id) = $`+strconv.Itoa(idx))
		args = append(args, a.ID)
		idx++
	} else if a.Role == "doctor" {
		whereClauses = append(whereClauses, `(
			p.clinician_user_id = $`+strconv.Itoa(idx)+` OR
			c.doctor_id = $`+strconv.Itoa(idx)+` OR
			c.created_by = $`+strconv.Itoa(idx)+` OR
			EXISTS(SELECT 1 FROM encounter e WHERE e.patient_id = p.id AND (e.doctor_id = $`+strconv.Itoa(idx)+` OR EXISTS(SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $`+strconv.Itoa(idx)+` AND ect.revoked_at IS NULL)))
		)`)
		args = append(args, a.ID)
		idx++
	} else if a.Role != "admin" {
		return nil, 0, domain.ErrForbidden
	}

	if patientID != "" {
		whereClauses = append(whereClauses, `canonical_patient_id(c.patient_id) = canonical_patient_id($`+strconv.Itoa(idx)+`::uuid)`)
		args = append(args, patientID)
		idx++
	}

	if search != "" {
		searchPat := "%" + strings.TrimSpace(search) + "%"
		whereClauses = append(whereClauses, `(
			p.given_name ILIKE $`+strconv.Itoa(idx)+` OR
			p.family_name ILIKE $`+strconv.Itoa(idx)+` OR
			p.medical_record_number::text ILIKE $`+strconv.Itoa(idx)+` OR
			u.name ILIKE $`+strconv.Itoa(idx)+` OR
			c.description ILIKE $`+strconv.Itoa(idx)+`
		)`)
		args = append(args, searchPat)
		idx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := `
		SELECT COUNT(*)
		FROM odontogram_chart c
		JOIN patient p ON p.id = c.patient_id
		LEFT JOIN "user" u ON u.id = c.doctor_id
	` + whereSQL

	var totalCount int
	err := s.DB.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT
			c.id, c.patient_id,
			TRIM(p.given_name || ' ' || p.family_name) AS patient_name,
			p.medical_record_number::text AS patient_mrn,
			c.doctor_id,
			COALESCE(u.name, 'Doctor') AS doctor_name,
			c.encounter_id::text,
			c.description,
			c.odontogram,
			c.version,
			c.created_by,
			c.updated_by,
			c.created_at,
			c.updated_at,
			c.is_legacy,
			c.legacy_provenance
		FROM odontogram_chart c
		JOIN patient p ON p.id = c.patient_id
		LEFT JOIN "user" u ON u.id = c.doctor_id
	` + whereSQL + `
		ORDER BY c.created_at DESC
		LIMIT $` + strconv.Itoa(idx) + ` OFFSET $` + strconv.Itoa(idx+1)

	dataArgs := append(args, limit, offset)
	rows, err := s.DB.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var charts []domain.OdontogramChart
	for rows.Next() {
		var c domain.OdontogramChart
		var encID *string
		var rawJSON []byte
		if err := rows.Scan(
			&c.ID, &c.PatientID, &c.PatientName, &c.PatientMRN,
			&c.DoctorID, &c.DoctorName, &encID, &c.Description,
			&rawJSON, &c.Version, &c.CreatedBy, &c.UpdatedBy,
			&c.CreatedAt, &c.UpdatedAt, &c.IsLegacy, &c.LegacyProvenance,
		); err != nil {
			return nil, 0, err
		}
		c.EncounterID = encID
		strMap := make(map[string]string)
		if len(rawJSON) > 0 {
			_ = json.Unmarshal(rawJSON, &strMap)
		}
		c.Odontogram = make(map[int]string, len(strMap))
		for k, v := range strMap {
			if num, err := strconv.Atoi(k); err == nil {
				c.Odontogram[num] = v
			}
		}
		charts = append(charts, c)
	}
	if charts == nil {
		charts = []domain.OdontogramChart{}
	}
	return charts, totalCount, nil
}

func (s Store) DeleteOdontogramChart(ctx context.Context, id string) error {
	cmd, err := s.DB.Exec(ctx, `DELETE FROM odontogram_chart WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ─── Odontogram Legends ───────────────────────────────────────────────────────

func (s Store) ListOdontogramLegends(ctx context.Context) ([]domain.OdontogramLegend, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, code, name, color, description, is_default, created_at, updated_at
		FROM odontogram_legend
		ORDER BY is_default DESC, code ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var legends []domain.OdontogramLegend
	for rows.Next() {
		var l domain.OdontogramLegend
		if err := rows.Scan(
			&l.ID, &l.Code, &l.Name, &l.Color, &l.Description,
			&l.IsDefault, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		legends = append(legends, l)
	}
	if legends == nil {
		legends = []domain.OdontogramLegend{}
	}
	return legends, nil
}

func (s Store) CreateOdontogramLegend(ctx context.Context, a domain.Actor, input domain.SaveOdontogramLegendInput) (domain.OdontogramLegend, error) {
	var l domain.OdontogramLegend
	err := s.DB.QueryRow(ctx, `
		INSERT INTO odontogram_legend (id, code, name, color, description, is_default, created_at, updated_at)
		VALUES ('leg-' || substr(gen_random_uuid()::text, 1, 8), $1, $2, $3, $4, false, clock_timestamp(), clock_timestamp())
		RETURNING id, code, name, color, description, is_default, created_at, updated_at
	`, input.Code, input.Name, input.Color, input.Description).Scan(
		&l.ID, &l.Code, &l.Name, &l.Color, &l.Description, &l.IsDefault, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.OdontogramLegend{}, domain.ErrConflict
		}
		return domain.OdontogramLegend{}, err
	}
	return l, nil
}

// lockAndValidateLegendCodes acquires a FOR SHARE lock on each referenced legend row
// and validates that every custom condition code exists in odontogram_legend (G2).
func lockAndValidateLegendCodes(ctx context.Context, tx pgx.Tx, codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	// Check if odontogram_legend table exists in current search path
	var tableExists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'odontogram_legend'
		)
	`).Scan(&tableExists)
	if err != nil || !tableExists {
		return nil
	}

	unique := make(map[string]bool, len(codes))
	var sortedCodes []string
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c != "" && !unique[c] {
			unique[c] = true
			sortedCodes = append(sortedCodes, c)
		}
	}
	if len(sortedCodes) == 0 {
		return nil
	}
	sort.Strings(sortedCodes)

	rows, err := tx.Query(ctx, `
		SELECT code FROM odontogram_legend
		WHERE code = ANY($1)
		ORDER BY code
		FOR SHARE
	`, sortedCodes)
	if err != nil {
		return err
	}
	defer rows.Close()

	found := make(map[string]bool, len(sortedCodes))
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return err
		}
		found[code] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	var legacyConditions = map[string]bool{
		"healthy":    true,
		"caries":     true,
		"missing":    true,
		"filled":     true,
		"crown":      true,
		"implant":    true,
		"extracted":  true,
		"root_canal": true,
	}

	for _, c := range sortedCodes {
		if !found[c] {
			if legacyConditions[strings.ToLower(c)] {
				continue
			}
			return fmt.Errorf("%w: unknown or unseeded clinical condition code '%s'", domain.ErrValidation, c)
		}
	}
	return nil
}

// isLegendCodeInUse checks whether a condition code is referenced in active child entries,
// chart odontogram JSON values, or legacy provenance entries (G2).
func isLegendCodeInUse(ctx context.Context, tx pgx.Tx, code string) (bool, error) {
	var inUse bool
	err := tx.QueryRow(ctx, `
		SELECT (
			EXISTS(
				SELECT 1 FROM patient_odontogram_entry
				WHERE condition = $1
			)
			OR EXISTS(
				SELECT 1 FROM odontogram_chart c,
				LATERAL jsonb_each_text(c.odontogram) AS kv
				WHERE kv.value = $1
			)
			OR EXISTS(
				SELECT 1 FROM odontogram_chart c,
				LATERAL jsonb_array_elements(
					CASE
						WHEN c.legacy_provenance IS NOT NULL AND jsonb_typeof(c.legacy_provenance) = 'array'
						THEN c.legacy_provenance
						ELSE '[]'::jsonb
					END
				) AS elem
				WHERE elem->>'condition' = $1
				   OR (
						elem ? 'odontogram'
						AND EXISTS (
							SELECT 1 FROM jsonb_each_text(elem->'odontogram') AS okv
							WHERE okv.value = $1
						)
				   )
			)
		)
	`, code).Scan(&inUse)
	return inUse, err
}

func (s Store) UpdateOdontogramLegend(ctx context.Context, a domain.Actor, id string, input domain.SaveOdontogramLegendInput) (domain.OdontogramLegend, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.OdontogramLegend{}, err
	}
	defer tx.Rollback(ctx)

	var existing domain.OdontogramLegend
	err = tx.QueryRow(ctx, `
		SELECT id, code, name, color, description, is_default, created_at, updated_at
		FROM odontogram_legend
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(
		&existing.ID, &existing.Code, &existing.Name, &existing.Color,
		&existing.Description, &existing.IsDefault, &existing.CreatedAt, &existing.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OdontogramLegend{}, domain.ErrNotFound
		}
		return domain.OdontogramLegend{}, err
	}

	// Default clinical legends cannot have their code or name changed
	if existing.IsDefault {
		if input.Code != existing.Code || input.Name != existing.Name {
			return domain.OdontogramLegend{}, domain.ErrForbidden
		}
	}

	// Check if this legend is referenced in any clinical records or charts (G2 exact value check)
	inUse, err := isLegendCodeInUse(ctx, tx, existing.Code)
	if err != nil {
		return domain.OdontogramLegend{}, err
	}

	if inUse {
		if input.Code != existing.Code {
			return domain.OdontogramLegend{}, fmt.Errorf("%w: cannot change code of legend '%s' because it is referenced in clinical charts", domain.ErrConflict, existing.Code)
		}
		if input.Name != existing.Name {
			return domain.OdontogramLegend{}, fmt.Errorf("%w: cannot rename legend '%s' because it is referenced in clinical charts", domain.ErrConflict, existing.Code)
		}
	}

	var l domain.OdontogramLegend
	err = tx.QueryRow(ctx, `
		UPDATE odontogram_legend
		SET code = $2,
			name = $3,
			color = $4,
			description = $5,
			updated_at = clock_timestamp()
		WHERE id = $1
		RETURNING id, code, name, color, description, is_default, created_at, updated_at
	`, id, input.Code, input.Name, input.Color, input.Description).Scan(
		&l.ID, &l.Code, &l.Name, &l.Color, &l.Description, &l.IsDefault, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.OdontogramLegend{}, domain.ErrConflict
		}
		return domain.OdontogramLegend{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OdontogramLegend{}, err
	}
	return l, nil
}

func (s Store) DeleteOdontogramLegend(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var isDefault bool
	var code string
	err = tx.QueryRow(ctx, `SELECT is_default, code FROM odontogram_legend WHERE id = $1 FOR UPDATE`, id).Scan(&isDefault, &code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if isDefault {
		return domain.ErrForbidden
	}

	inUse, err := isLegendCodeInUse(ctx, tx, code)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}

	cmd, err := tx.Exec(ctx, `DELETE FROM odontogram_legend WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return tx.Commit(ctx)
}

