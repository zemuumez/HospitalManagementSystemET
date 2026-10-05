package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// ─── Bed Occupancy & Assignments ──────────────────────────────────────────────

func (s Store) BedOccupancyReport(ctx context.Context) (domain.BedOccupancyReport, error) {
	var rep domain.BedOccupancyReport
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM hospital_bed WHERE active = true`).Scan(&rep.TotalBeds)
	if err != nil {
		return rep, err
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(DISTINCT bed_id) FROM encounter WHERE status = 'active' AND bed_id IS NOT NULL`).Scan(&rep.OccupiedBeds)
	if err != nil {
		return rep, err
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM encounter WHERE status = 'active'`).Scan(&rep.ActiveAdmissions)
	if err != nil {
		return rep, err
	}
	rep.AvailableBeds = rep.TotalBeds - rep.OccupiedBeds
	if rep.AvailableBeds < 0 {
		rep.AvailableBeds = 0
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

	out, err := transferBedTx(ctx, tx, a, input.EncounterID, domain.BedTransfer{BedID: input.BedID, Version: input.Version, Reason: input.Notes})
	if err != nil {
		return domain.BedAssignment{}, err
	}
	if out.PatientID != input.PatientID {
		return domain.BedAssignment{}, domain.ErrValidation
	}
	if _, err = tx.Exec(ctx, `UPDATE bed_assignment SET assigned_to=clock_timestamp() WHERE encounter_id=$1 AND assigned_to IS NULL`, input.EncounterID); err != nil {
		return domain.BedAssignment{}, err
	}

	var assign domain.BedAssignment
	assign.BedID = input.BedID
	assign.EncounterID = input.EncounterID
	assign.PatientID = input.PatientID
	assign.Notes = input.Notes
	assign.AssignedBy = a.ID

	err = tx.QueryRow(ctx, `
		INSERT INTO bed_assignment (
			bed_id, encounter_id, patient_id, notes, assigned_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, assigned_from, created_at
	`, input.BedID, input.EncounterID, input.PatientID, input.Notes, a.ID,
	).Scan(&assign.ID, &assign.AssignedFrom, &assign.CreatedAt)
	if err != nil {
		return domain.BedAssignment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.BedAssignment{}, err
	}
	return assign, nil
}

func (s Store) ListBedAssignments(ctx context.Context, encounterID string) ([]domain.BedAssignment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT ba.id, ba.bed_id, b.name AS bed_name, b.bed_type, ba.encounter_id,
		       ba.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       ba.assigned_from, ba.assigned_to, ba.notes, ba.assigned_by, ba.created_at
		FROM bed_assignment ba
		JOIN hospital_bed b ON b.id = ba.bed_id
		JOIN patient p ON p.id = ba.patient_id
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
			&a.ID, &a.BedID, &a.BedName, &a.BedType, &a.EncounterID,
			&a.PatientID, &a.PatientName, &a.AssignedFrom, &a.AssignedTo,
			&a.Notes, &a.AssignedBy, &a.CreatedAt,
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
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, assigned_at
	`, encounterID, input.StaffID, input.RoleTitle, a.ID, input.Notes,
	).Scan(&m.ID, &m.AssignedAt)
	if err != nil {
		return domain.CareTeamMember{}, err
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
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, uploaded_at
	`, encounterID, input.Title, input.FileURL, input.FileType, input.FileSizeBytes, a.ID,
	).Scan(&att.ID, &att.UploadedAt)
	if err != nil {
		return domain.EncounterAttachment{}, err
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

func (s Store) Odontogram(ctx context.Context, patientID string) ([]domain.OdontogramEntry, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, patient_id, encounter_id::text, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
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
			&e.ID, &e.PatientID, &encID, &e.ToothNumber, &e.Condition, &e.ProcedureNotes,
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
	var e domain.OdontogramEntry
	var encID *string
	err := s.DB.QueryRow(ctx, `
		INSERT INTO patient_odontogram_entry (
			patient_id, encounter_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, clock_timestamp(), clock_timestamp())
		ON CONFLICT (patient_id, tooth_number) DO UPDATE SET
			encounter_id = COALESCE(EXCLUDED.encounter_id, patient_odontogram_entry.encounter_id),
			condition = EXCLUDED.condition,
			procedure_notes = EXCLUDED.procedure_notes,
			diagnosed_by = EXCLUDED.diagnosed_by,
			updated_at = clock_timestamp()
		RETURNING id, patient_id, encounter_id::text, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
	`, patientID, input.EncounterID, input.ToothNumber, input.Condition, input.ProcedureNotes, a.ID).Scan(
		&e.ID, &e.PatientID, &encID, &e.ToothNumber, &e.Condition, &e.ProcedureNotes,
		&e.DiagnosedBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return domain.OdontogramEntry{}, err
	}
	e.EncounterID = encID
	return e, nil
}
