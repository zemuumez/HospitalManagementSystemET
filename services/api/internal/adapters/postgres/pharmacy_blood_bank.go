package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
	"time"
)

// --- Medicine Categories ---

func (s Store) MedicineCategories(ctx context.Context, page int, search string) ([]domain.MedicineCategory, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if search != "" {
		whereClause = " WHERE name ILIKE $1"
		args = append(args, "%"+search+"%")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM medicine_category"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, name, is_active, created_at, updated_at
		FROM medicine_category
		%s
		ORDER BY name ASC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.MedicineCategory{}
	for rows.Next() {
		var c domain.MedicineCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

func (s Store) CreateMedicineCategory(ctx context.Context, a domain.Actor, in domain.MedicineCategoryInput) (domain.MedicineCategory, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.MedicineCategory{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.MedicineCategory
	err = tx.QueryRow(ctx, `
		INSERT INTO medicine_category (name, is_active)
		VALUES ($1, $2)
		RETURNING id, name, is_active, created_at, updated_at
	`, in.Name, in.IsActive).Scan(&out.ID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'medicine_category.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateMedicineCategory(ctx context.Context, a domain.Actor, id string, in domain.MedicineCategoryInput) (domain.MedicineCategory, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.MedicineCategory{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.MedicineCategory
	err = tx.QueryRow(ctx, `
		UPDATE medicine_category
		SET name = $1, is_active = $2, updated_at = clock_timestamp()
		WHERE id = $3
		RETURNING id, name, is_active, created_at, updated_at
	`, in.Name, in.IsActive, id).Scan(&out.ID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'medicine_category.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Medicine Brands ---

func (s Store) MedicineBrands(ctx context.Context, page int, search string) ([]domain.MedicineBrand, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if search != "" {
		whereClause = " WHERE (name ILIKE $1 OR email ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM medicine_brand"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, name, email, phone, created_at, updated_at
		FROM medicine_brand
		%s
		ORDER BY name ASC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.MedicineBrand{}
	for rows.Next() {
		var b domain.MedicineBrand
		if err := rows.Scan(&b.ID, &b.Name, &b.Email, &b.Phone, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, b)
	}
	return out, total, nil
}

func (s Store) CreateMedicineBrand(ctx context.Context, a domain.Actor, in domain.MedicineBrandInput) (domain.MedicineBrand, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.MedicineBrand{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.MedicineBrand
	err = tx.QueryRow(ctx, `
		INSERT INTO medicine_brand (name, email, phone)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, phone, created_at, updated_at
	`, in.Name, in.Email, in.Phone).Scan(&out.ID, &out.Name, &out.Email, &out.Phone, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'medicine_brand.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateMedicineBrand(ctx context.Context, a domain.Actor, id string, in domain.MedicineBrandInput) (domain.MedicineBrand, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.MedicineBrand{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.MedicineBrand
	err = tx.QueryRow(ctx, `
		UPDATE medicine_brand
		SET name = $1, email = $2, phone = $3, updated_at = clock_timestamp()
		WHERE id = $4
		RETURNING id, name, email, phone, created_at, updated_at
	`, in.Name, in.Email, in.Phone, id).Scan(&out.ID, &out.Name, &out.Email, &out.Phone, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return out, domain.ErrConflict
		}
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'medicine_brand.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Blood Bank ---

func (s Store) BloodBank(ctx context.Context) ([]domain.BloodBankItem, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, blood_group, remained_bags, updated_at
		FROM blood_bank
		ORDER BY blood_group ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.BloodBankItem{}
	for rows.Next() {
		var item domain.BloodBankItem
		if err := rows.Scan(&item.ID, &item.BloodGroup, &item.RemainedBags, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (s Store) BloodDonors(ctx context.Context, page int, search string) ([]domain.BloodDonor, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if search != "" {
		whereClause = " WHERE name ILIKE $1"
		args = append(args, "%"+search+"%")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM blood_donor"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, name, age, gender, blood_group, last_donate_date, created_at, updated_at
		FROM blood_donor
		%s
		ORDER BY name ASC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.BloodDonor{}
	for rows.Next() {
		var d domain.BloodDonor
		if err := rows.Scan(&d.ID, &d.Name, &d.Age, &d.Gender, &d.BloodGroup, &d.LastDonateDate, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, nil
}

func (s Store) CreateBloodDonor(ctx context.Context, a domain.Actor, in domain.BloodDonorInput) (domain.BloodDonor, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.BloodDonor{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.BloodDonor
	err = tx.QueryRow(ctx, `
		INSERT INTO blood_donor (name, age, gender, blood_group, last_donate_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, age, gender, blood_group, last_donate_date, created_at, updated_at
	`, in.Name, in.Age, in.Gender, in.BloodGroup, *in.LastDonateDate).Scan(
		&out.ID, &out.Name, &out.Age, &out.Gender, &out.BloodGroup, &out.LastDonateDate, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'blood_donor.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) BloodDonations(ctx context.Context, page int) ([]domain.BloodDonation, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM blood_donation").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.Query(ctx, fmt.Sprintf(`
		SELECT bd.id, bd.donor_id, d.name, d.blood_group, bd.bags, bd.donation_date, bd.recorded_by, bd.created_at
		FROM blood_donation bd
		JOIN blood_donor d ON d.id = bd.donor_id
		ORDER BY bd.donation_date DESC
		LIMIT %d OFFSET %d
	`, limit, offset))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.BloodDonation{}
	for rows.Next() {
		var item domain.BloodDonation
		if err := rows.Scan(
			&item.ID, &item.DonorID, &item.DonorName, &item.BloodGroup,
			&item.Bags, &item.DonationDate, &item.RecordedBy, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s Store) RecordBloodDonation(ctx context.Context, a domain.Actor, in domain.BloodDonationInput) (domain.BloodDonation, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.BloodDonation{}, err
	}
	defer tx.Rollback(ctx)

	var donorName, bloodGroup string
	err = tx.QueryRow(ctx, `SELECT name, blood_group FROM blood_donor WHERE id = $1`, in.DonorID).Scan(&donorName, &bloodGroup)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BloodDonation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.BloodDonation{}, err
	}

	donDate := time.Now()
	if in.DonationDate != nil && !in.DonationDate.IsZero() {
		donDate = *in.DonationDate
	}

	var out domain.BloodDonation
	err = tx.QueryRow(ctx, `
		INSERT INTO blood_donation (donor_id, bags, donation_date, recorded_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, donor_id, bags, donation_date, recorded_by, created_at
	`, in.DonorID, in.Bags, donDate, a.ID).Scan(
		&out.ID, &out.DonorID, &out.Bags, &out.DonationDate, &out.RecordedBy, &out.CreatedAt,
	)
	if err != nil {
		return out, err
	}
	out.DonorName = donorName
	out.BloodGroup = bloodGroup

	// Update donor's last_donate_date
	_, err = tx.Exec(ctx, `UPDATE blood_donor SET last_donate_date = $1, updated_at = clock_timestamp() WHERE id = $2`, donDate, in.DonorID)
	if err != nil {
		return out, err
	}

	// Increment blood bank inventory for this blood group
	_, err = tx.Exec(ctx, `
		UPDATE blood_bank
		SET remained_bags = remained_bags + $1, updated_at = clock_timestamp()
		WHERE blood_group = $2
	`, in.Bags, bloodGroup)
	if err != nil {
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'blood_donation.recorded', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) BloodIssues(ctx context.Context, patientFilter string, page int) ([]domain.BloodIssue, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if patientFilter != "" {
		whereClause = " WHERE (bi.patient_id = $1 OR bi.patient_id IN (SELECT id::text FROM patient WHERE patient_portal_owner(id) = $1))"
		args = append(args, patientFilter)
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM blood_issue bi"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT bi.id, bi.issue_date, bi.doctor_id, bi.donor_id, bi.patient_id, patient_portal_owner(p.id),
		       bi.blood_group, bi.bags, bi.amount_minor, bi.remarks, bi.issued_by, bi.invoice_id,
		       bi.created_at, bi.updated_at
		FROM blood_issue bi
		LEFT JOIN patient p ON p.id::text = bi.patient_id
		%s
		ORDER BY bi.issue_date DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.BloodIssue{}
	for rows.Next() {
		var bi domain.BloodIssue
		var donorID *string
		var patientUserID *string
		var invoiceID *string
		if err := rows.Scan(
			&bi.ID, &bi.IssueDate, &bi.DoctorID, &donorID, &bi.PatientID, &patientUserID,
			&bi.BloodGroup, &bi.Bags, &bi.AmountMinor, &bi.Remarks, &bi.IssuedBy, &invoiceID,
			&bi.CreatedAt, &bi.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		bi.DonorID = donorID
		bi.PatientUserID = patientUserID
		bi.InvoiceID = invoiceID
		out = append(out, bi)
	}
	return out, total, nil
}

func (s Store) BloodIssue(ctx context.Context, id string) (domain.BloodIssue, error) {
	var bi domain.BloodIssue
	var donorID *string
	var patientUserID *string
	var invoiceID *string
	err := s.DB.QueryRow(ctx, `
		SELECT bi.id, bi.issue_date, bi.doctor_id, bi.donor_id, bi.patient_id, patient_portal_owner(p.id),
		       bi.blood_group, bi.bags, bi.amount_minor, bi.remarks, bi.issued_by, bi.invoice_id,
		       bi.created_at, bi.updated_at
		FROM blood_issue bi
		LEFT JOIN patient p ON p.id::text = bi.patient_id
		WHERE bi.id = $1
	`, id).Scan(
		&bi.ID, &bi.IssueDate, &bi.DoctorID, &donorID, &bi.PatientID, &patientUserID,
		&bi.BloodGroup, &bi.Bags, &bi.AmountMinor, &bi.Remarks, &bi.IssuedBy, &invoiceID,
		&bi.CreatedAt, &bi.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return bi, domain.ErrNotFound
	}
	if err != nil {
		return bi, err
	}
	bi.DonorID = donorID
	bi.PatientUserID = patientUserID
	bi.InvoiceID = invoiceID
	return bi, nil
}

func (s Store) CreateBloodIssue(ctx context.Context, a domain.Actor, in domain.BloodIssueInput) (domain.BloodIssue, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.BloodIssue{}, err
	}
	defer tx.Rollback(ctx)

	// Atomically deduct bags from blood bank with insufficient stock protection
	res, err := tx.Exec(ctx, `
		UPDATE blood_bank
		SET remained_bags = remained_bags - $1, updated_at = clock_timestamp()
		WHERE blood_group = $2 AND remained_bags >= $1
	`, in.Bags, in.BloodGroup)
	if err != nil {
		return domain.BloodIssue{}, err
	}
	if res.RowsAffected() == 0 {
		return domain.BloodIssue{}, domain.ErrValidation
	}

	var out domain.BloodIssue
	var donorID *string
	if in.DonorID != nil && *in.DonorID != "" {
		donorID = in.DonorID
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO blood_issue (issue_date, doctor_id, donor_id, patient_id, blood_group, bags, amount_minor, remarks, issued_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, issue_date, doctor_id, donor_id, patient_id, blood_group, bags, amount_minor, remarks, issued_by, created_at, updated_at
	`, *in.IssueDate, in.DoctorID, donorID, in.PatientID, in.BloodGroup, in.Bags, in.AmountMinor, in.Remarks, a.ID).Scan(
		&out.ID, &out.IssueDate, &out.DoctorID, &out.DonorID, &out.PatientID, &out.BloodGroup, &out.Bags, &out.AmountMinor, &out.Remarks, &out.IssuedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return out, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'blood_issue.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Prescriptions ---

func (s Store) Prescriptions(ctx context.Context, patientFilter string, doctorFilter string, page int) ([]domain.Prescription, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if patientFilter != "" {
		whereClause = " WHERE (pr.patient_id = $1 OR pr.patient_id IN (SELECT id::text FROM patient WHERE patient_portal_owner(id) = $1))"
		args = append(args, patientFilter)
	}

	if doctorFilter != "" {
		if whereClause == "" {
			whereClause = " WHERE "
		} else {
			whereClause += " AND "
		}
		args = append(args, doctorFilter)
		whereClause += fmt.Sprintf("pr.doctor_id = $%d", len(args))
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM prescription pr"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT pr.id, pr.patient_id, patient_portal_owner(p.id), pr.doctor_id, pr.encounter_id, pr.food_allergies,
		       pr.tendency_bleed, pr.heart_disease, pr.high_blood_pressure, pr.diabetic, pr.surgery,
		       pr.accident, pr.others, pr.medical_history, pr.current_medication, pr.female_pregnancy,
		       pr.breast_feeding, pr.health_insurance, pr.low_income, pr.reference, pr.status,
		       pr.plus_rate, pr.temperature, pr.problem_description, pr.test, pr.advice,
		       pr.next_visit_qty, pr.next_visit_time, pr.created_at, pr.updated_at
		FROM prescription pr
		LEFT JOIN patient p ON p.id::text = pr.patient_id
		%s
		ORDER BY pr.created_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Prescription{}
	for rows.Next() {
		var pr domain.Prescription
		var encID *string
		var patientUserID *string
		if err := rows.Scan(
			&pr.ID, &pr.PatientID, &patientUserID, &pr.DoctorID, &encID, &pr.FoodAllergies,
			&pr.TendencyBleed, &pr.HeartDisease, &pr.HighBloodPressure, &pr.Diabetic, &pr.Surgery,
			&pr.Accident, &pr.Others, &pr.MedicalHistory, &pr.CurrentMedication, &pr.FemalePregnancy,
			&pr.BreastFeeding, &pr.HealthInsurance, &pr.LowIncome, &pr.Reference, &pr.Status,
			&pr.PlusRate, &pr.Temperature, &pr.ProblemDescription, &pr.Test, &pr.Advice,
			&pr.NextVisitQty, &pr.NextVisitTime, &pr.CreatedAt, &pr.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		pr.EncounterID = encID
		pr.PatientUserID = patientUserID
		out = append(out, pr)
	}
	return out, total, nil
}

func (s Store) Prescription(ctx context.Context, id string) (domain.Prescription, error) {
	var pr domain.Prescription
	var encID *string
	var patientUserID *string
	err := s.DB.QueryRow(ctx, `
		SELECT pr.id, pr.patient_id, patient_portal_owner(p.id), pr.doctor_id, pr.encounter_id, pr.food_allergies,
		       pr.tendency_bleed, pr.heart_disease, pr.high_blood_pressure, pr.diabetic, pr.surgery,
		       pr.accident, pr.others, pr.medical_history, pr.current_medication, pr.female_pregnancy,
		       pr.breast_feeding, pr.health_insurance, pr.low_income, pr.reference, pr.status,
		       pr.plus_rate, pr.temperature, pr.problem_description, pr.test, pr.advice,
		       pr.next_visit_qty, pr.next_visit_time, pr.created_at, pr.updated_at
		FROM prescription pr
		LEFT JOIN patient p ON p.id::text = pr.patient_id
		WHERE pr.id = $1
	`, id).Scan(
		&pr.ID, &pr.PatientID, &patientUserID, &pr.DoctorID, &encID, &pr.FoodAllergies,
		&pr.TendencyBleed, &pr.HeartDisease, &pr.HighBloodPressure, &pr.Diabetic, &pr.Surgery,
		&pr.Accident, &pr.Others, &pr.MedicalHistory, &pr.CurrentMedication, &pr.FemalePregnancy,
		&pr.BreastFeeding, &pr.HealthInsurance, &pr.LowIncome, &pr.Reference, &pr.Status,
		&pr.PlusRate, &pr.Temperature, &pr.ProblemDescription, &pr.Test, &pr.Advice,
		&pr.NextVisitQty, &pr.NextVisitTime, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return pr, domain.ErrNotFound
	}
	if err != nil {
		return pr, err
	}
	pr.EncounterID = encID
	pr.PatientUserID = patientUserID

	// Fetch prescription medicine items
	mRows, err := s.DB.Query(ctx, `
		SELECT id, prescription_id, medicine_id, medicine_name, dosage, day, time, comment
		FROM prescription_medicine
		WHERE prescription_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		return pr, err
	}
	defer mRows.Close()
	for mRows.Next() {
		var m domain.PrescriptionMedicine
		var medID *string
		if err = mRows.Scan(&m.ID, &m.PrescriptionID, &medID, &m.MedicineName, &m.Dosage, &m.Day, &m.Time, &m.Comment); err != nil {
			return pr, err
		}
		m.MedicineID = medID
		pr.Medicines = append(pr.Medicines, m)
	}
	if err = mRows.Err(); err != nil {
		return pr, err
	}

	return pr, nil
}

func (s Store) CreatePrescription(ctx context.Context, a domain.Actor, in domain.PrescriptionInput) (domain.Prescription, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Prescription{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.Prescription
	var encID *string
	if in.EncounterID != nil && *in.EncounterID != "" {
		encID = in.EncounterID
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO prescription (
			patient_id, doctor_id, encounter_id, food_allergies, tendency_bleed, heart_disease,
			high_blood_pressure, diabetic, surgery, accident, others, medical_history,
			current_medication, female_pregnancy, breast_feeding, health_insurance, low_income,
			reference, status, plus_rate, temperature, problem_description, test, advice,
			next_visit_qty, next_visit_time
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, 0, $19, $20, $21, $22, $23,
			$24, $25
		)
		RETURNING id, patient_id, doctor_id, encounter_id, food_allergies, tendency_bleed, heart_disease,
		          high_blood_pressure, diabetic, surgery, accident, others, medical_history,
		          current_medication, female_pregnancy, breast_feeding, health_insurance, low_income,
		          reference, status, plus_rate, temperature, problem_description, test, advice,
		          next_visit_qty, next_visit_time, created_at, updated_at
	`,
		in.PatientID, in.DoctorID, encID, in.FoodAllergies, in.TendencyBleed, in.HeartDisease,
		in.HighBloodPressure, in.Diabetic, in.Surgery, in.Accident, in.Others, in.MedicalHistory,
		in.CurrentMedication, in.FemalePregnancy, in.BreastFeeding, in.HealthInsurance, in.LowIncome,
		in.Reference, in.PlusRate, in.Temperature, in.ProblemDescription, in.Test, in.Advice,
		in.NextVisitQty, in.NextVisitTime,
	).Scan(
		&out.ID, &out.PatientID, &out.DoctorID, &out.EncounterID, &out.FoodAllergies, &out.TendencyBleed, &out.HeartDisease,
		&out.HighBloodPressure, &out.Diabetic, &out.Surgery, &out.Accident, &out.Others, &out.MedicalHistory,
		&out.CurrentMedication, &out.FemalePregnancy, &out.BreastFeeding, &out.HealthInsurance, &out.LowIncome,
		&out.Reference, &out.Status, &out.PlusRate, &out.Temperature, &out.ProblemDescription, &out.Test, &out.Advice,
		&out.NextVisitQty, &out.NextVisitTime, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return out, err
	}

	// Insert medicine items
	for _, m := range in.Medicines {
		var medID *string
		if m.MedicineID != nil && *m.MedicineID != "" {
			medID = m.MedicineID
		}
		var item domain.PrescriptionMedicine
		err = tx.QueryRow(ctx, `
			INSERT INTO prescription_medicine (prescription_id, medicine_id, medicine_name, dosage, day, time, comment)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, prescription_id, medicine_id, medicine_name, dosage, day, time, comment
		`, out.ID, medID, m.MedicineName, m.Dosage, m.Day, m.Time, m.Comment).Scan(
			&item.ID, &item.PrescriptionID, &medID, &item.MedicineName, &item.Dosage, &item.Day, &item.Time, &item.Comment,
		)
		if err != nil {
			return out, err
		}
		item.MedicineID = medID
		out.Medicines = append(out.Medicines, item)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'prescription.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdatePrescriptionStatus(ctx context.Context, a domain.Actor, id string, status int) (domain.Prescription, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Prescription{}, err
	}
	defer tx.Rollback(ctx)

	var out domain.Prescription
	var encID *string
	err = tx.QueryRow(ctx, `
		UPDATE prescription
		SET status = $1, updated_at = clock_timestamp()
		WHERE id = $2
		RETURNING id, patient_id, doctor_id, encounter_id, food_allergies, tendency_bleed, heart_disease,
		          high_blood_pressure, diabetic, surgery, accident, others, medical_history,
		          current_medication, female_pregnancy, breast_feeding, health_insurance, low_income,
		          reference, status, plus_rate, temperature, problem_description, test, advice,
		          next_visit_qty, next_visit_time, created_at, updated_at
	`, status, id).Scan(
		&out.ID, &out.PatientID, &out.DoctorID, &encID, &out.FoodAllergies, &out.TendencyBleed, &out.HeartDisease,
		&out.HighBloodPressure, &out.Diabetic, &out.Surgery, &out.Accident, &out.Others, &out.MedicalHistory,
		&out.CurrentMedication, &out.FemalePregnancy, &out.BreastFeeding, &out.HealthInsurance, &out.LowIncome,
		&out.Reference, &out.Status, &out.PlusRate, &out.Temperature, &out.ProblemDescription, &out.Test, &out.Advice,
		&out.NextVisitQty, &out.NextVisitTime, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, err
	}
	out.EncounterID = encID

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'prescription.status_updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
