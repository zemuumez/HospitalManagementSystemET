package postgres

import (
	"context"

	"hms.local/api/internal/domain"
)

// ─── Categories & Units ───────────────────────────────────────────────────────

func (s Store) DiagnosticCategories(ctx context.Context, kind string) ([]domain.DiagnosticCategory, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, kind, description, created_at
		FROM diagnostic_category
		WHERE ($1 = '' OR kind = $1)
		ORDER BY name ASC
	`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DiagnosticCategory
	for rows.Next() {
		var c domain.DiagnosticCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Kind, &c.Description, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	if list == nil {
		list = []domain.DiagnosticCategory{}
	}
	return list, nil
}

func (s Store) CreateDiagnosticCategory(ctx context.Context, a domain.Actor, input domain.DiagnosticCategoryInput) (domain.DiagnosticCategory, error) {
	var c domain.DiagnosticCategory
	c.Name = input.Name
	c.Kind = input.Kind
	c.Description = input.Description
	err := s.DB.QueryRow(ctx, `
		INSERT INTO diagnostic_category (name, kind, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, input.Name, input.Kind, input.Description).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return domain.DiagnosticCategory{}, err
	}
	return c, nil
}

func (s Store) DiagnosticUnits(ctx context.Context) ([]domain.DiagnosticUnit, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, description, created_at
		FROM diagnostic_unit
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DiagnosticUnit
	for rows.Next() {
		var u domain.DiagnosticUnit
		if err := rows.Scan(&u.ID, &u.Name, &u.Description, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	if list == nil {
		list = []domain.DiagnosticUnit{}
	}
	return list, nil
}

func (s Store) CreateDiagnosticUnit(ctx context.Context, a domain.Actor, input domain.DiagnosticUnitInput) (domain.DiagnosticUnit, error) {
	var u domain.DiagnosticUnit
	u.Name = input.Name
	u.Description = input.Description
	err := s.DB.QueryRow(ctx, `
		INSERT INTO diagnostic_unit (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at
	`, input.Name, input.Description).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return domain.DiagnosticUnit{}, err
	}
	return u, nil
}

// ─── Report Files & Portal Release ───────────────────────────────────────────

func (s Store) DiagnosticReportFiles(ctx context.Context, orderID string, patientOnly bool) ([]domain.DiagnosticReportFile, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, order_id, file_name, file_url, file_size_bytes, mime_type,
		       patient_released, released_at, released_by, uploaded_by, uploaded_at
		FROM diagnostic_report_file
		WHERE order_id = $1 AND ($2 = false OR patient_released = true)
		ORDER BY uploaded_at DESC
	`, orderID, patientOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DiagnosticReportFile
	for rows.Next() {
		var f domain.DiagnosticReportFile
		if err := rows.Scan(
			&f.ID, &f.OrderID, &f.FileName, &f.FileURL, &f.FileSizeBytes, &f.MimeType,
			&f.PatientReleased, &f.ReleasedAt, &f.ReleasedBy, &f.UploadedBy, &f.UploadedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	if list == nil {
		list = []domain.DiagnosticReportFile{}
	}
	return list, nil
}

func (s Store) UploadDiagnosticReportFile(ctx context.Context, a domain.Actor, orderID string, input domain.UploadDiagnosticReportFileInput) (domain.DiagnosticReportFile, error) {
	var f domain.DiagnosticReportFile
	f.OrderID = orderID
	f.FileName = input.FileName
	f.FileURL = input.FileURL
	f.FileSizeBytes = input.FileSizeBytes
	f.MimeType = input.MimeType
	f.UploadedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO diagnostic_report_file (
			order_id, file_name, file_url, file_size_bytes, mime_type, uploaded_by
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, patient_released, uploaded_at
	`, orderID, input.FileName, input.FileURL, input.FileSizeBytes, input.MimeType, a.ID,
	).Scan(&f.ID, &f.PatientReleased, &f.UploadedAt)
	if err != nil {
		return domain.DiagnosticReportFile{}, err
	}
	return f, nil
}

func (s Store) ReleaseReportFileToPortal(ctx context.Context, a domain.Actor, fileID string) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE diagnostic_report_file
		SET patient_released = true, released_at = clock_timestamp(), released_by = $1
		WHERE id = $2
	`, a.ID, fileID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ─── Diagnosis Templates ──────────────────────────────────────────────────────

func (s Store) DiagnosisTemplates(ctx context.Context) ([]domain.DiagnosisTemplate, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, title, category, template_content, created_by, created_at, updated_at
		FROM diagnosis_template
		ORDER BY title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DiagnosisTemplate
	for rows.Next() {
		var t domain.DiagnosisTemplate
		if err := rows.Scan(&t.ID, &t.Title, &t.Category, &t.TemplateContent, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	if list == nil {
		list = []domain.DiagnosisTemplate{}
	}
	return list, nil
}

func (s Store) CreateDiagnosisTemplate(ctx context.Context, a domain.Actor, input domain.DiagnosisTemplateInput) (domain.DiagnosisTemplate, error) {
	var t domain.DiagnosisTemplate
	t.Title = input.Title
	t.Category = input.Category
	t.TemplateContent = input.TemplateContent
	t.CreatedBy = a.ID
	err := s.DB.QueryRow(ctx, `
		INSERT INTO diagnosis_template (title, category, template_content, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, input.Title, input.Category, input.TemplateContent, a.ID).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.DiagnosisTemplate{}, err
	}
	return t, nil
}

// ─── Vaccines & Vaccinations ──────────────────────────────────────────────────

func (s Store) Vaccines(ctx context.Context) ([]domain.Vaccine, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, target_disease, recommended_doses, min_age_months, instructions, created_at
		FROM vaccine_catalog
		ORDER BY min_age_months ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Vaccine
	for rows.Next() {
		var v domain.Vaccine
		if err := rows.Scan(&v.ID, &v.Name, &v.TargetDisease, &v.RecommendedDoses, &v.MinAgeMonths, &v.Instructions, &v.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	if list == nil {
		list = []domain.Vaccine{}
	}
	return list, nil
}

func (s Store) CreateVaccine(ctx context.Context, a domain.Actor, input domain.VaccineInput) (domain.Vaccine, error) {
	var v domain.Vaccine
	v.Name = input.Name
	v.TargetDisease = input.TargetDisease
	v.RecommendedDoses = input.RecommendedDoses
	v.MinAgeMonths = input.MinAgeMonths
	v.Instructions = input.Instructions
	err := s.DB.QueryRow(ctx, `
		INSERT INTO vaccine_catalog (name, target_disease, recommended_doses, min_age_months, instructions)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, input.Name, input.TargetDisease, input.RecommendedDoses, input.MinAgeMonths, input.Instructions).Scan(&v.ID, &v.CreatedAt)
	if err != nil {
		return domain.Vaccine{}, err
	}
	return v, nil
}

func (s Store) PatientVaccinations(ctx context.Context, patientID string) ([]domain.PatientVaccination, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT pv.id, pv.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       pv.vaccine_id, vc.name AS vaccine_name, pv.dose_number, pv.lot_number,
		       pv.expiry_date::text, pv.administered_at, pv.administered_by,
		       pv.next_due_date::text, pv.notes, pv.created_at
		FROM patient_vaccination pv
		JOIN patient p ON p.id = pv.patient_id
		JOIN vaccine_catalog vc ON vc.id = pv.vaccine_id
		WHERE pv.patient_id = $1
		ORDER BY pv.administered_at DESC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PatientVaccination
	for rows.Next() {
		var pv domain.PatientVaccination
		if err := rows.Scan(
			&pv.ID, &pv.PatientID, &pv.PatientName, &pv.VaccineID, &pv.VaccineName,
			&pv.DoseNumber, &pv.LotNumber, &pv.ExpiryDate, &pv.AdministeredAt,
			&pv.AdministeredBy, &pv.NextDueDate, &pv.Notes, &pv.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, pv)
	}
	if list == nil {
		list = []domain.PatientVaccination{}
	}
	return list, nil
}

func (s Store) AdministerVaccine(ctx context.Context, a domain.Actor, input domain.AdministerVaccineInput) (domain.PatientVaccination, error) {
	var pv domain.PatientVaccination
	pv.PatientID = input.PatientID
	pv.VaccineID = input.VaccineID
	pv.DoseNumber = input.DoseNumber
	pv.LotNumber = input.LotNumber
	pv.ExpiryDate = input.ExpiryDate
	pv.NextDueDate = input.NextDueDate
	pv.Notes = input.Notes
	pv.AdministeredBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO patient_vaccination (
			patient_id, vaccine_id, dose_number, lot_number, expiry_date,
			administered_by, next_due_date, notes
		) VALUES ($1, $2, $3, $4, $5::date, $6, $7::date, $8)
		RETURNING id, administered_at, created_at
	`, input.PatientID, input.VaccineID, input.DoseNumber, input.LotNumber,
		input.ExpiryDate, a.ID, input.NextDueDate, input.Notes,
	).Scan(&pv.ID, &pv.AdministeredAt, &pv.CreatedAt)
	if err != nil {
		return domain.PatientVaccination{}, err
	}

	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&pv.PatientName)
	_ = s.DB.QueryRow(ctx, `SELECT name FROM vaccine_catalog WHERE id = $1`, input.VaccineID).Scan(&pv.VaccineName)
	return pv, nil
}

// ─── Vital Reports ────────────────────────────────────────────────────────────

func (s Store) BirthReports(ctx context.Context, page int) ([]domain.BirthReport, error) {
	offset := (page - 1) * 25
	rows, err := s.DB.Query(ctx, `
		SELECT br.id, br.report_number, br.child_name, br.gender, br.birth_date,
		       br.weight_kg::float8, br.mother_patient_id::text, br.mother_name, br.father_name,
		       br.delivered_by, u.name AS delivered_by_name, br.notes, br.created_by,
		       br.created_at, br.updated_at
		FROM birth_report br
		JOIN "user" u ON u.id = br.delivered_by
		ORDER BY br.birth_date DESC
		LIMIT 25 OFFSET $1
	`, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.BirthReport
	for rows.Next() {
		var br domain.BirthReport
		var motherID *string
		if err := rows.Scan(
			&br.ID, &br.ReportNumber, &br.ChildName, &br.Gender, &br.BirthDate,
			&br.WeightKg, &motherID, &br.MotherName, &br.FatherName,
			&br.DeliveredBy, &br.DeliveredByName, &br.Notes, &br.CreatedBy,
			&br.CreatedAt, &br.UpdatedAt,
		); err != nil {
			return nil, err
		}
		br.MotherPatientID = motherID
		list = append(list, br)
	}
	if list == nil {
		list = []domain.BirthReport{}
	}
	return list, nil
}

func (s Store) CreateBirthReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateBirthReportInput) (domain.BirthReport, error) {
	var br domain.BirthReport
	br.ReportNumber = reportNumber
	br.ChildName = input.ChildName
	br.Gender = input.Gender
	br.BirthDate = input.BirthDate
	br.WeightKg = input.WeightKg
	br.MotherPatientID = input.MotherPatientID
	br.MotherName = input.MotherName
	br.FatherName = input.FatherName
	br.DeliveredBy = input.DeliveredBy
	br.Notes = input.Notes
	br.CreatedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO birth_report (
			report_number, child_name, gender, birth_date, weight_kg,
			mother_patient_id, mother_name, father_name, delivered_by, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`, reportNumber, input.ChildName, input.Gender, input.BirthDate, input.WeightKg,
		input.MotherPatientID, input.MotherName, input.FatherName, input.DeliveredBy, input.Notes, a.ID,
	).Scan(&br.ID, &br.CreatedAt, &br.UpdatedAt)
	if err != nil {
		return domain.BirthReport{}, err
	}
	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.DeliveredBy).Scan(&br.DeliveredByName)
	return br, nil
}

func (s Store) DeathReports(ctx context.Context, page int) ([]domain.DeathReport, error) {
	offset := (page - 1) * 25
	rows, err := s.DB.Query(ctx, `
		SELECT dr.id, dr.report_number, dr.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       dr.death_date, dr.cause_of_death, dr.certified_by, u.name AS certified_by_name,
		       dr.guardian_acknowledged, dr.notes, dr.created_by, dr.created_at, dr.updated_at
		FROM death_report dr
		JOIN patient p ON p.id = dr.patient_id
		JOIN "user" u ON u.id = dr.certified_by
		ORDER BY dr.death_date DESC
		LIMIT 25 OFFSET $1
	`, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DeathReport
	for rows.Next() {
		var dr domain.DeathReport
		if err := rows.Scan(
			&dr.ID, &dr.ReportNumber, &dr.PatientID, &dr.PatientName,
			&dr.DeathDate, &dr.CauseOfDeath, &dr.CertifiedBy, &dr.CertifiedByName,
			&dr.GuardianAcknowledged, &dr.Notes, &dr.CreatedBy, &dr.CreatedAt, &dr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, dr)
	}
	if list == nil {
		list = []domain.DeathReport{}
	}
	return list, nil
}

func (s Store) CreateDeathReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateDeathReportInput) (domain.DeathReport, error) {
	var dr domain.DeathReport
	dr.ReportNumber = reportNumber
	dr.PatientID = input.PatientID
	dr.DeathDate = input.DeathDate
	dr.CauseOfDeath = input.CauseOfDeath
	dr.CertifiedBy = input.CertifiedBy
	dr.GuardianAcknowledged = input.GuardianAcknowledged
	dr.Notes = input.Notes
	dr.CreatedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO death_report (
			report_number, patient_id, death_date, cause_of_death, certified_by,
			guardian_acknowledged, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, reportNumber, input.PatientID, input.DeathDate, input.CauseOfDeath, input.CertifiedBy,
		input.GuardianAcknowledged, input.Notes, a.ID,
	).Scan(&dr.ID, &dr.CreatedAt, &dr.UpdatedAt)
	if err != nil {
		return domain.DeathReport{}, err
	}

	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&dr.PatientName)
	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.CertifiedBy).Scan(&dr.CertifiedByName)
	return dr, nil
}

func (s Store) OperationReports(ctx context.Context, page int) ([]domain.OperationReport, error) {
	offset := (page - 1) * 25
	rows, err := s.DB.Query(ctx, `
		SELECT opr.id, opr.report_number, opr.encounter_id, opr.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       opr.operation_name, opr.surgeon_id, u.name AS surgeon_name, opr.assistant_surgeon,
		       opr.anesthetist, opr.anesthesia_type, opr.operation_date, opr.pre_operative_diagnosis,
		       opr.post_operative_diagnosis, opr.procedure_technique, opr.findings, opr.complications,
		       opr.created_by, opr.created_at, opr.updated_at
		FROM operation_report opr
		JOIN patient p ON p.id = opr.patient_id
		JOIN "user" u ON u.id = opr.surgeon_id
		ORDER BY opr.operation_date DESC
		LIMIT 25 OFFSET $1
	`, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.OperationReport
	for rows.Next() {
		var opr domain.OperationReport
		if err := rows.Scan(
			&opr.ID, &opr.ReportNumber, &opr.EncounterID, &opr.PatientID, &opr.PatientName,
			&opr.OperationName, &opr.SurgeonID, &opr.SurgeonName, &opr.AssistantSurgeon,
			&opr.Anesthetist, &opr.AnesthesiaType, &opr.OperationDate, &opr.PreOperativeDiagnosis,
			&opr.PostOperativeDiagnosis, &opr.ProcedureTechnique, &opr.Findings, &opr.Complications,
			&opr.CreatedBy, &opr.CreatedAt, &opr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, opr)
	}
	if list == nil {
		list = []domain.OperationReport{}
	}
	return list, nil
}

func (s Store) CreateOperationReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateOperationReportInput) (domain.OperationReport, error) {
	var opr domain.OperationReport
	opr.ReportNumber = reportNumber
	opr.EncounterID = input.EncounterID
	opr.PatientID = input.PatientID
	opr.OperationName = input.OperationName
	opr.SurgeonID = input.SurgeonID
	opr.AssistantSurgeon = input.AssistantSurgeon
	opr.Anesthetist = input.Anesthetist
	opr.AnesthesiaType = input.AnesthesiaType
	opr.OperationDate = input.OperationDate
	opr.PreOperativeDiagnosis = input.PreOperativeDiagnosis
	opr.PostOperativeDiagnosis = input.PostOperativeDiagnosis
	opr.ProcedureTechnique = input.ProcedureTechnique
	opr.Findings = input.Findings
	opr.Complications = input.Complications
	opr.CreatedBy = a.ID

	err := s.DB.QueryRow(ctx, `
		INSERT INTO operation_report (
			report_number, encounter_id, patient_id, operation_name, surgeon_id,
			assistant_surgeon, anesthetist, anesthesia_type, operation_date,
			pre_operative_diagnosis, post_operative_diagnosis, procedure_technique,
			findings, complications, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`, reportNumber, input.EncounterID, input.PatientID, input.OperationName, input.SurgeonID,
		input.AssistantSurgeon, input.Anesthetist, input.AnesthesiaType, input.OperationDate,
		input.PreOperativeDiagnosis, input.PostOperativeDiagnosis, input.ProcedureTechnique,
		input.Findings, input.Complications, a.ID,
	).Scan(&opr.ID, &opr.CreatedAt, &opr.UpdatedAt)
	if err != nil {
		return domain.OperationReport{}, err
	}

	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&opr.PatientName)
	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.SurgeonID).Scan(&opr.SurgeonName)
	return opr, nil
}

func (s Store) InvestigationReports(ctx context.Context, patientID string) ([]domain.InvestigationReport, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT ir.id, ir.report_number, ir.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       ir.encounter_id::text, ir.title, ir.investigation_type, ir.clinical_notes,
		       ir.conclusion, ir.investigated_by, u.name AS investigated_by_name, ir.created_at
		FROM investigation_report ir
		JOIN patient p ON p.id = ir.patient_id
		JOIN "user" u ON u.id = ir.investigated_by
		WHERE ir.patient_id = $1
		ORDER BY ir.created_at DESC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.InvestigationReport
	for rows.Next() {
		var ir domain.InvestigationReport
		var encID *string
		if err := rows.Scan(
			&ir.ID, &ir.ReportNumber, &ir.PatientID, &ir.PatientName,
			&encID, &ir.Title, &ir.InvestigationType, &ir.ClinicalNotes,
			&ir.Conclusion, &ir.InvestigatedBy, &ir.InvestigatedByName, &ir.CreatedAt,
		); err != nil {
			return nil, err
		}
		ir.EncounterID = encID
		list = append(list, ir)
	}
	if list == nil {
		list = []domain.InvestigationReport{}
	}
	return list, nil
}

func (s Store) CreateInvestigationReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateInvestigationReportInput) (domain.InvestigationReport, error) {
	var ir domain.InvestigationReport
	ir.ReportNumber = reportNumber
	ir.PatientID = input.PatientID
	ir.EncounterID = input.EncounterID
	ir.Title = input.Title
	ir.InvestigationType = input.InvestigationType
	ir.ClinicalNotes = input.ClinicalNotes
	ir.Conclusion = input.Conclusion
	ir.InvestigatedBy = input.InvestigatedBy

	err := s.DB.QueryRow(ctx, `
		INSERT INTO investigation_report (
			report_number, patient_id, encounter_id, title, investigation_type,
			clinical_notes, conclusion, investigated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`, reportNumber, input.PatientID, input.EncounterID, input.Title, input.InvestigationType,
		input.ClinicalNotes, input.Conclusion, input.InvestigatedBy,
	).Scan(&ir.ID, &ir.CreatedAt)
	if err != nil {
		return domain.InvestigationReport{}, err
	}

	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&ir.PatientName)
	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.InvestigatedBy).Scan(&ir.InvestigatedByName)
	return ir, nil
}
