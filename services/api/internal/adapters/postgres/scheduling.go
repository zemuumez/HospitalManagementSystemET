package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func doctor(ctx context.Context, q querier, id string) (domain.Doctor, error) {
	d := domain.Doctor{Hours: []domain.DoctorHours{}}
	var deptID *string
	var desc, photo, spec, desig, qual string
	var deptTitle string
	var opdCharge, apptCharge float64
	var active bool
	var detailsRaw []byte

	err := q.QueryRow(ctx, `
		SELECT d.user_id,
		       u.name,
		       COALESCE(u.email, ''),
		       COALESCE(dd.title, d.department),
		       d.department_id::text,
		       COALESCE(d.specialist, ''),
		       COALESCE(d.designation, ''),
		       COALESCE(d.qualification, ''),
		       COALESCE(d.description, ''),
		       COALESCE(d.photo_url, ''),
		       COALESCE(d.opd_charge::float8, 0),
		       COALESCE(d.appointment_charge::float8, 0),
		       d.slot_minutes,
		       d.version,
		       a.active,
		       COALESCE(sp.details, '{}'::jsonb)
		FROM doctor_profile d
		JOIN "user" u ON u.id = d.user_id
		JOIN staff_access a ON a.user_id = d.user_id
		LEFT JOIN doctor_department dd ON dd.id = d.department_id
		LEFT JOIN staff_profile sp ON sp.user_id = d.user_id
		WHERE d.user_id = $1 AND a.role = 'doctor'`, id).Scan(
		&d.ID, &d.Name, &d.Email, &deptTitle, &deptID,
		&spec, &desig, &qual, &desc, &photo,
		&opdCharge, &apptCharge, &d.SlotMinutes, &d.Version,
		&active, &detailsRaw,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	if err != nil {
		return d, err
	}
	d.Department = deptTitle
	d.DepartmentID = deptID
	d.Specialist = spec
	d.Designation = desig
	d.Qualification = qual
	d.Description = desc
	d.PhotoURL = photo
	d.OpdCharge = opdCharge
	d.AppointmentCharge = apptCharge
	d.Active = active

	if len(detailsRaw) > 0 {
		var details domain.StaffDetails
		if err := json.Unmarshal(detailsRaw, &details); err == nil {
			d.Phone = details.Phone
			d.Gender = details.Gender
			d.DateOfBirth = details.DateOfBirth
			d.BloodGroup = details.BloodGroup
			d.Address1 = details.Address1
			d.Address2 = details.Address2
			d.City = details.City
			d.Zip = details.PostalCode
			if d.Specialist == "" && details.Specialty != "" {
				d.Specialist = details.Specialty
			}
			if d.Designation == "" && details.Designation != "" {
				d.Designation = details.Designation
			}
			if d.Qualification == "" && details.Qualification != "" {
				d.Qualification = details.Qualification
			}
		}
	}
	if d.Specialist == "" {
		if desc != "" {
			d.Specialist = desc
		} else {
			d.Specialist = d.Department
		}
	}

	rows, err := q.Query(ctx, `SELECT weekday, start_minute, end_minute FROM doctor_hours WHERE doctor_id=$1 ORDER BY weekday, start_minute`, id)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var h domain.DoctorHours
		if err = rows.Scan(&h.Weekday, &h.StartMinute, &h.EndMinute); err != nil {
			return d, err
		}
		d.Hours = append(d.Hours, h)
	}
	return d, rows.Err()
}

func (s Store) Doctor(ctx context.Context, id string) (domain.Doctor, error) {
	return doctor(ctx, s.DB, id)
}

func (s Store) Doctors(ctx context.Context) ([]domain.Doctor, error) {
	rows, err := s.DB.Query(ctx, `SELECT d.user_id FROM doctor_profile d JOIN staff_access a ON a.user_id=d.user_id WHERE a.active AND a.role='doctor' ORDER BY d.user_id LIMIT 500`)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := []domain.Doctor{}
	for _, id := range ids {
		d, e := doctor(ctx, s.DB, id)
		if e != nil {
			return nil, e
		}
		result = append(result, d)
	}
	return result, nil
}

func (s Store) ListDoctors(ctx context.Context, statusFilter string, deptFilter string, search string) ([]domain.Doctor, error) {
	query := `SELECT d.user_id
		FROM doctor_profile d
		JOIN "user" u ON u.id = d.user_id
		JOIN staff_access a ON a.user_id = d.user_id
		LEFT JOIN doctor_department dd ON dd.id = d.department_id
		WHERE a.role = 'doctor'`
	args := []any{}
	idx := 1

	if statusFilter == "active" {
		query += ` AND a.active = true`
	} else if statusFilter == "inactive" {
		query += ` AND a.active = false`
	}

	if deptFilter != "" {
		query += fmt.Sprintf(` AND d.department_id = $%d`, idx)
		args = append(args, deptFilter)
		idx++
	}

	if search != "" {
		query += fmt.Sprintf(` AND (strpos(lower(u.name), $%d) > 0 OR strpos(lower(u.email), $%d) > 0 OR strpos(lower(COALESCE(dd.title, d.department)), $%d) > 0 OR strpos(lower(COALESCE(d.specialist, '')), $%d) > 0)`, idx, idx, idx, idx)
		args = append(args, strings.ToLower(search))
		idx++
	}

	query += ` ORDER BY u.name, d.user_id LIMIT 500`

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	out := []domain.Doctor{}
	for _, id := range ids {
		doc, err := doctor(ctx, s.DB, id)
		if err != nil {
			return nil, err
		}
		out = append(out, doc)
	}
	return out, nil
}

func (s Store) CreateDoctorProfile(ctx context.Context, a domain.Actor, input domain.CreateDoctorInput) (domain.Doctor, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Doctor{}, err
	}
	defer tx.Rollback(ctx)

	if err = staffAdmin(ctx, tx, a); err != nil {
		return domain.Doctor{}, err
	}

	// 1. Verify user exists in staff_access with role='doctor'
	var userActive bool
	var userRole string
	err = tx.QueryRow(ctx, `SELECT role, active FROM staff_access WHERE user_id=$1 FOR UPDATE`, input.UserID).Scan(&userRole, &userActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Doctor{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Doctor{}, err
	}
	if userRole != "doctor" {
		return domain.Doctor{}, domain.ErrValidation
	}

	// 2. Check if doctor_profile already exists
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM doctor_profile WHERE user_id=$1)`, input.UserID).Scan(&exists)
	if err != nil {
		return domain.Doctor{}, err
	}
	if exists {
		return domain.Doctor{}, domain.ErrConflict
	}

	// 3. Verify department exists and is NOT archived
	var deptTitle string
	var deptArchived bool
	err = tx.QueryRow(ctx, `SELECT title, archived FROM doctor_department WHERE id=$1`, input.DepartmentID).Scan(&deptTitle, &deptArchived)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Doctor{}, domain.ErrValidation
	}
	if err != nil {
		return domain.Doctor{}, err
	}
	if deptArchived {
		return domain.Doctor{}, domain.ErrValidation
	}

	slotMinutes := input.SlotMinutes
	if slotMinutes <= 0 {
		slotMinutes = 60
	}

	// 4. Insert doctor_profile
	_, err = tx.Exec(ctx, `
		INSERT INTO doctor_profile(
			user_id, department, department_id, specialist, designation, qualification,
			description, photo_url, opd_charge, appointment_charge, slot_minutes, version
		) VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8, $9, $10, $11, 1)
	`, input.UserID, deptTitle, input.DepartmentID, input.Specialist, input.Designation, input.Qualification,
		input.Description, input.PhotoURL, input.OpdCharge, input.AppointmentCharge, slotMinutes)
	if err != nil {
		return domain.Doctor{}, err
	}

	// 5. Update staff_profile details
	var curDetailsRaw []byte
	_ = tx.QueryRow(ctx, `SELECT details FROM staff_profile WHERE user_id=$1 FOR UPDATE`, input.UserID).Scan(&curDetailsRaw)
	var details domain.StaffDetails
	if len(curDetailsRaw) > 0 {
		_ = json.Unmarshal(curDetailsRaw, &details)
	}
	if input.Phone != "" {
		details.Phone = input.Phone
	}
	if input.Gender != "" {
		details.Gender = input.Gender
	}
	if input.DateOfBirth != "" {
		details.DateOfBirth = input.DateOfBirth
	}
	if input.BloodGroup != "" {
		details.BloodGroup = input.BloodGroup
	}
	if input.Designation != "" {
		details.Designation = input.Designation
	}
	if input.Qualification != "" {
		details.Qualification = input.Qualification
	}
	if input.Specialist != "" {
		details.Specialty = input.Specialist
	}
	if input.Address1 != "" {
		details.Address1 = input.Address1
	}
	if input.Address2 != "" {
		details.Address2 = input.Address2
	}
	if input.City != "" {
		details.City = input.City
	}
	if input.Zip != "" {
		details.PostalCode = input.Zip
	}
	newDetailsRaw, err := json.Marshal(details)
	if err != nil {
		return domain.Doctor{}, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO staff_profile(user_id, details, version)
		VALUES($1, $2, 1)
		ON CONFLICT(user_id) DO UPDATE SET details = EXCLUDED.details
	`, input.UserID, newDetailsRaw)
	if err != nil {
		return domain.Doctor{}, err
	}

	// 6. Insert working hours: default or supplied
	hours := input.Hours
	if len(hours) == 0 {
		// D2: Original Laravel default is all 7 days (0..6), 10:00:00 (minute 600) to 19:30:00 (minute 1170)
		for w := 0; w <= 6; w++ {
			hours = append(hours, domain.DoctorHours{Weekday: w, StartMinute: 600, EndMinute: 1170})
		}
	}
	for _, h := range hours {
		if _, err = tx.Exec(ctx, `INSERT INTO doctor_hours(doctor_id, weekday, start_minute, end_minute) VALUES($1, $2, $3, $4)`, input.UserID, h.Weekday, h.StartMinute, h.EndMinute); err != nil {
			return domain.Doctor{}, err
		}
	}

	// 7. Audit event
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id, action, resource_id) VALUES($1, 'doctor.created', $2)`, a.ID, input.UserID); err != nil {
		return domain.Doctor{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.Doctor{}, err
	}
	return doctor(ctx, s.DB, input.UserID)
}

func (s Store) UpdateDoctorProfile(ctx context.Context, a domain.Actor, id string, input domain.UpdateDoctorInput) (domain.Doctor, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Doctor{}, err
	}
	defer tx.Rollback(ctx)

	// Lock doctor_profile
	var curVersion int
	var curDeptID *string
	var curDeptTitle string
	err = tx.QueryRow(ctx, `
		SELECT version, department_id::text, department
		FROM doctor_profile
		WHERE user_id = $1
		FOR UPDATE`, id).Scan(&curVersion, &curDeptID, &curDeptTitle)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Doctor{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Doctor{}, err
	}
	if curVersion != input.Version {
		return domain.Doctor{}, domain.ErrStale
	}

	var newDeptID *string
	var newDeptTitle *string
	if input.DepartmentID != nil {
		var deptArchived bool
		var title string
		err = tx.QueryRow(ctx, `SELECT title, archived FROM doctor_department WHERE id=$1`, *input.DepartmentID).Scan(&title, &deptArchived)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Doctor{}, domain.ErrValidation
		}
		if err != nil {
			return domain.Doctor{}, err
		}
		if deptArchived {
			return domain.Doctor{}, domain.ErrValidation
		}
		newDeptID = input.DepartmentID
		newDeptTitle = &title
	}

	// Update doctor_profile columns
	_, err = tx.Exec(ctx, `
		UPDATE doctor_profile
		SET department_id = COALESCE($2::uuid, department_id),
		    department = COALESCE($3, department),
		    specialist = COALESCE($4, specialist),
		    designation = COALESCE($5, designation),
		    qualification = COALESCE($6, qualification),
		    description = COALESCE($7, description),
		    photo_url = COALESCE($8, photo_url),
		    appointment_charge = COALESCE($9, appointment_charge),
		    opd_charge = COALESCE($10, opd_charge),
		    slot_minutes = COALESCE($11, slot_minutes),
		    version = version + 1
		WHERE user_id = $1
	`, id, newDeptID, newDeptTitle, input.Specialist, input.Designation, input.Qualification,
		input.Description, input.PhotoURL, input.AppointmentCharge, input.OpdCharge, input.SlotMinutes)
	if err != nil {
		return domain.Doctor{}, err
	}

	if input.Name != nil {
		if _, err = tx.Exec(ctx, `UPDATE "user" SET name = $1 WHERE id = $2`, *input.Name, id); err != nil {
			return domain.Doctor{}, err
		}
	}
	if input.Email != nil {
		if _, err = tx.Exec(ctx, `UPDATE "user" SET email = $1 WHERE id = $2`, *input.Email, id); err != nil {
			if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {
				return domain.Doctor{}, domain.ErrConflict
			}
			return domain.Doctor{}, err
		}
	}

	// Update staff_profile details
	var curDetailsRaw []byte
	_ = tx.QueryRow(ctx, `SELECT details FROM staff_profile WHERE user_id=$1 FOR UPDATE`, id).Scan(&curDetailsRaw)
	var details domain.StaffDetails
	if len(curDetailsRaw) > 0 {
		_ = json.Unmarshal(curDetailsRaw, &details)
	}
	if input.Phone != nil {
		details.Phone = *input.Phone
	}
	if input.Gender != nil {
		details.Gender = *input.Gender
	}
	if input.DateOfBirth != nil {
		details.DateOfBirth = *input.DateOfBirth
	}
	if input.BloodGroup != nil {
		details.BloodGroup = *input.BloodGroup
	}
	if input.Designation != nil {
		details.Designation = *input.Designation
	}
	if input.Qualification != nil {
		details.Qualification = *input.Qualification
	}
	if input.Specialist != nil {
		details.Specialty = *input.Specialist
	}
	if input.Address1 != nil {
		details.Address1 = *input.Address1
	}
	if input.Address2 != nil {
		details.Address2 = *input.Address2
	}
	if input.City != nil {
		details.City = *input.City
	}
	if input.Zip != nil {
		details.PostalCode = *input.Zip
	}
	newDetailsRaw, err := json.Marshal(details)
	if err != nil {
		return domain.Doctor{}, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO staff_profile(user_id, details, version)
		VALUES($1, $2, 1)
		ON CONFLICT(user_id) DO UPDATE SET details = EXCLUDED.details, version = staff_profile.version + 1
	`, id, newDetailsRaw)
	if err != nil {
		return domain.Doctor{}, err
	}

	if input.Hours != nil {
		if _, err = tx.Exec(ctx, `DELETE FROM doctor_hours WHERE doctor_id=$1`, id); err != nil {
			return domain.Doctor{}, err
		}
		for _, h := range input.Hours {
			if _, err = tx.Exec(ctx, `INSERT INTO doctor_hours(doctor_id, weekday, start_minute, end_minute) VALUES($1, $2, $3, $4)`, id, h.Weekday, h.StartMinute, h.EndMinute); err != nil {
				return domain.Doctor{}, err
			}
		}
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id, action, resource_id) VALUES($1, 'doctor.updated', $2)`, a.ID, id); err != nil {
		return domain.Doctor{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.Doctor{}, err
	}
	return doctor(ctx, s.DB, id)
}

func (s Store) SetDoctorStatus(ctx context.Context, a domain.Actor, id string, active bool) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err = staffAdmin(ctx, tx, a); err != nil {
		return err
	}

	var curActive bool
	err = tx.QueryRow(ctx, `SELECT active FROM staff_access WHERE user_id=$1 AND role='doctor' FOR UPDATE`, id).Scan(&curActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if curActive == active {
		return nil
	}

	_, err = tx.Exec(ctx, `UPDATE staff_access SET active=$2 WHERE user_id=$1`, id, active)
	if err != nil {
		return err
	}

	if !active {
		_, err = tx.Exec(ctx, `DELETE FROM session WHERE "userId"=$1`, id)
		if err != nil {
			return err
		}
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id, action, resource_id) VALUES($1, 'doctor.status_changed', $2)`, a.ID, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) DeleteDoctor(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err = staffAdmin(ctx, tx, a); err != nil {
		return err
	}

	if a.ID == id {
		return domain.ErrConflict
	}

	var role string
	err = tx.QueryRow(ctx, `SELECT role FROM staff_access WHERE user_id=$1 FOR UPDATE`, id).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if role != "doctor" {
		return domain.ErrNotFound
	}

	var hasDocProfile bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM doctor_profile WHERE user_id=$1)`, id).Scan(&hasDocProfile)
	if err != nil {
		return err
	}
	if !hasDocProfile {
		return domain.ErrNotFound
	}

	var inUse bool
	err = tx.QueryRow(ctx, `
		SELECT
			EXISTS(SELECT 1 FROM patient_case WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM encounter WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM appointment WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM birth_report WHERE delivered_by=$1)
			OR EXISTS(SELECT 1 FROM death_report WHERE certified_by=$1)
			OR EXISTS(SELECT 1 FROM investigation_report WHERE investigated_by=$1)
			OR EXISTS(SELECT 1 FROM operation_report WHERE surgeon_id=$1)
			OR EXISTS(SELECT 1 FROM prescription WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM ipd_admission_details i JOIN encounter e ON e.id=i.encounter_id WHERE e.doctor_id=$1)
			OR EXISTS(SELECT 1 FROM employee_payroll WHERE user_id=$1)
			OR EXISTS(SELECT 1 FROM opd_follow_up WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM patient_queue WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM public_appointment_request WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM live_consultation WHERE doctor_id=$1)
			OR EXISTS(SELECT 1 FROM patient_referral WHERE referred_by=$1)
			OR EXISTS(SELECT 1 FROM patient_odontogram_entry WHERE diagnosed_by=$1)
			OR EXISTS(SELECT 1 FROM clinical_note WHERE author_id=$1)
			OR EXISTS(SELECT 1 FROM patient WHERE user_id=$1 OR clinician_user_id=$1)
			OR EXISTS(SELECT 1 FROM audit_event WHERE actor_id=$1)
			OR EXISTS(SELECT 1 FROM message_outbox WHERE actor_id=$1)
	`, id).Scan(&inUse)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}

	if _, err = tx.Exec(ctx, `DELETE FROM doctor_hours WHERE doctor_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM doctor_absence WHERE doctor_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM doctor_profile WHERE user_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM staff_profile_revision WHERE user_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM staff_role_event WHERE user_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM staff_profile WHERE user_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM staff_access WHERE user_id=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM session WHERE "userId"=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM verification WHERE value=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM account WHERE "userId"=$1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id, action, resource_id) VALUES($1, 'doctor.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM "user" WHERE id=$1`, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
func (s Store) SaveDoctor(ctx context.Context, a domain.Actor, d domain.Doctor) (domain.Doctor, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return d, err
	}
	defer tx.Rollback(ctx)
	// All schedule edits and bookings lock the same staff row before checking availability.
	var id string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, d.ID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	if err != nil {
		return d, err
	}
	var future bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment WHERE doctor_id=$1 AND ends_at>now() AND status IN ('booked','arrived'))`, d.ID).Scan(&future)
	if err != nil {
		return d, err
	}
	if future {
		return d, domain.ErrStale
	}
	var createdVersion int
	err = tx.QueryRow(ctx, `INSERT INTO doctor_profile(user_id,department,slot_minutes) SELECT $1,$2,$3 WHERE $4=0 ON CONFLICT(user_id) DO NOTHING RETURNING version`, d.ID, d.Department, d.SlotMinutes, d.Version).Scan(&createdVersion)
	// UPDATE is optimistic; create uses version zero.
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return d, err
	}
	if d.Version > 0 {
		err = tx.QueryRow(ctx, `UPDATE doctor_profile SET department=$2,slot_minutes=$3,version=version+1 WHERE user_id=$1 AND version=$4 RETURNING version`, d.ID, d.Department, d.SlotMinutes, d.Version).Scan(&d.Version)
		if errors.Is(err, pgx.ErrNoRows) {
			return d, domain.ErrStale
		}
		if err != nil {
			return d, err
		}
	} else {
		if errors.Is(err, pgx.ErrNoRows) {
			return d, domain.ErrStale
		}
		d.Version = 1
	}
	if _, err = tx.Exec(ctx, `DELETE FROM doctor_hours WHERE doctor_id=$1`, d.ID); err != nil {
		return d, err
	}
	for _, h := range d.Hours {
		if _, err = tx.Exec(ctx, `INSERT INTO doctor_hours(doctor_id,weekday,start_minute,end_minute) VALUES($1,$2,$3,$4)`, d.ID, h.Weekday, h.StartMinute, h.EndMinute); err != nil {
			return d, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'doctor.schedule_saved',$2)`, a.ID, d.ID); err != nil {
		return d, err
	}
	if err = tx.Commit(ctx); err != nil {
		return d, err
	}
	return doctor(ctx, s.DB, d.ID)
}
func (s Store) Slots(ctx context.Context, id string, day, now time.Time) ([]time.Time, error) {
	d, err := doctor(ctx, s.DB, id)
	if err != nil {
		return nil, err
	}
	var isHoliday bool
	err = s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM doctor_holiday WHERE doctor_id=$1 AND holiday_date=$2::date)`, id, day).Scan(&isHoliday)
	if err != nil {
		return nil, err
	}
	if isHoliday {
		return []time.Time{}, nil
	}
	rows, err := s.DB.Query(ctx, `SELECT starts_at,ends_at FROM appointment WHERE doctor_id=$1 AND status<>'cancelled' AND starts_at<$3 AND ends_at>$2 UNION ALL SELECT starts_at,ends_at FROM doctor_absence WHERE doctor_id=$1 AND cancelled_at IS NULL AND starts_at<$3 AND ends_at>$2`, id, day, day.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	busy := [][2]time.Time{}
	for rows.Next() {
		var b [2]time.Time
		if err = rows.Scan(&b[0], &b[1]); err != nil {
			return nil, err
		}
		busy = append(busy, b)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	breakRows, err := s.DB.Query(ctx, `SELECT break_from::text, break_to::text FROM doctor_lunch_break WHERE doctor_id=$1 AND (every_day=true OR break_date=$2::date)`, id, day)
	if err != nil {
		return nil, err
	}
	defer breakRows.Close()
	for breakRows.Next() {
		var bFrom, bTo string
		if err = breakRows.Scan(&bFrom, &bTo); err != nil {
			return nil, err
		}
		fromMin, e1 := domain.ParseTimeToMinutes(bFrom)
		toMin, e2 := domain.ParseTimeToMinutes(bTo)
		if e1 == nil && e2 == nil {
			bStart := day.Add(time.Duration(fromMin) * time.Minute)
			bEnd := day.Add(time.Duration(toMin) * time.Minute)
			busy = append(busy, [2]time.Time{bStart, bEnd})
		}
	}
	if err = breakRows.Err(); err != nil {
		return nil, err
	}

	result := []time.Time{}
	for _, h := range d.Hours {
		if h.Weekday != int(day.Weekday()) {
			continue
		}
		for m := h.StartMinute; m+d.SlotMinutes <= h.EndMinute; m += d.SlotMinutes {
			start := day.Add(time.Duration(m) * time.Minute)
			end := start.Add(time.Duration(d.SlotMinutes) * time.Minute)
			available := start.After(now)
			for _, b := range busy {
				if start.Before(b[1]) && b[0].Before(end) {
					available = false
				}
			}
			if available {
				result = append(result, start)
			}
		}
	}
	return result, nil
}

const appointmentColumns = `a.patient_id,a.doctor_id,a.starts_at,a.problem,a.notify_sms,a.id,p.given_name||' '||p.family_name,u.name,a.ends_at,a.status,a.version`
const appointmentJoin = ` FROM appointment a JOIN patient p ON p.id=a.patient_id JOIN "user" u ON u.id=a.doctor_id `
const appointmentScope = `($1 IN ('admin','receptionist') OR ($1='doctor' AND a.doctor_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2))`

func scanAppointment(row pgx.Row) (domain.Appointment, error) {
	var a domain.Appointment
	err := row.Scan(&a.PatientID, &a.DoctorID, &a.StartsAt, &a.Problem, &a.NotifySMS, &a.ID, &a.PatientName, &a.DoctorName, &a.EndsAt, &a.Status, &a.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		err = domain.ErrNotFound
	}
	return a, err
}
func (s Store) Appointments(ctx context.Context, actor domain.Actor, page int) ([]domain.Appointment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT `+appointmentColumns+appointmentJoin+`WHERE `+appointmentScope+` ORDER BY a.starts_at DESC,a.id LIMIT 25 OFFSET $3`, actor.Role, actor.ID, (page-1)*25)
	if err != nil {
		return nil, err
	}
	result := []domain.Appointment{}
	for rows.Next() {
		a, e := scanAppointment(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		result = append(result, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'appointments.list_viewed')`, actor.ID); err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}
func (s Store) Book(ctx context.Context, actor domain.Actor, i domain.AppointmentInput, key string, now time.Time) (domain.Appointment, error) {
	empty := domain.Appointment{}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(ctx)
	// Also serialize retries by actor/key, including retries with different doctors.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, actor.ID+":"+key); err != nil {
		return empty, err
	}
	old, err := scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+`WHERE a.created_by=$1 AND a.request_key=$2`, actor.ID, key))
	if err == nil {
		var original time.Time
		if err = tx.QueryRow(ctx, `SELECT original_starts_at FROM appointment WHERE id=$1`, old.ID).Scan(&original); err != nil {
			return empty, err
		}
		if old.PatientID != i.PatientID || old.DoctorID != i.DoctorID || !original.Equal(i.StartsAt) || old.Problem != i.Problem || old.NotifySMS != i.NotifySMS {
			return empty, domain.ErrConflict
		}
		return old, tx.Commit(ctx)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return empty, err
	}
	var id string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, i.DoctorID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return empty, domain.ErrNotFound
	}
	if err != nil {
		return empty, err
	}
	var phone string
	err = tx.QueryRow(ctx, `SELECT phone FROM patient WHERE id=$1 AND ($2 IN ('admin','receptionist') OR ($2='patient' AND patient_portal_owner(id)=$3) OR ($2='doctor' AND clinician_user_id=$3)) FOR UPDATE`, i.PatientID, actor.Role, actor.ID).Scan(&phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return empty, domain.ErrNotFound
	}
	if err != nil {
		return empty, err
	}
	d, err := doctor(ctx, tx, i.DoctorID)
	if err != nil {
		return empty, err
	}
	end := i.StartsAt.Add(time.Duration(d.SlotMinutes) * time.Minute)
	if !d.Allows(i.StartsAt, end) || !i.StartsAt.After(now) {
		return empty, domain.ErrStale
	}
	var occupied bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment WHERE (doctor_id=$1 OR patient_id=$2) AND status<>'cancelled' AND starts_at<$4 AND ends_at>$3) OR EXISTS(SELECT 1 FROM doctor_absence WHERE doctor_id=$1 AND cancelled_at IS NULL AND starts_at<$4 AND ends_at>$3)`, i.DoctorID, i.PatientID, i.StartsAt, end).Scan(&occupied)
	if err != nil {
		return empty, err
	}
	if occupied {
		return empty, domain.ErrStale
	}
	var holidayOrBreak bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM doctor_holiday
			WHERE doctor_id=$1 AND holiday_date=($2 AT TIME ZONE 'Africa/Addis_Ababa')::date
		) OR EXISTS(
			SELECT 1 FROM doctor_lunch_break
			WHERE doctor_id=$1
			  AND (every_day=true OR break_date=($2 AT TIME ZONE 'Africa/Addis_Ababa')::date)
			  AND ($2 AT TIME ZONE 'Africa/Addis_Ababa')::time < break_to
			  AND ($3 AT TIME ZONE 'Africa/Addis_Ababa')::time > break_from
		)
	`, i.DoctorID, i.StartsAt, end).Scan(&holidayOrBreak)
	if err != nil {
		return empty, err
	}
	if holidayOrBreak {
		return empty, domain.ErrConflict
	}
	if i.NotifySMS && !domain.PhonePattern.MatchString(phone) {
		return empty, domain.ErrValidation
	}
	err = tx.QueryRow(ctx, `INSERT INTO appointment(patient_id,doctor_id,starts_at,ends_at,problem,created_by,request_key,notify_sms) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, i.PatientID, i.DoctorID, i.StartsAt, end, i.Problem, actor.ID, key, i.NotifySMS).Scan(&id)
	if err != nil {
		return empty, err
	}
	if i.NotifySMS {
		_, err = tx.Exec(ctx, `INSERT INTO message_outbox(actor_id,channel,recipient,body,idempotency_key) VALUES($1,'sms',$2,$3,$4)`, actor.ID, phone, fmt.Sprintf("Your hospital appointment is booked for %s (EAT). Contact the hospital if you need to change it.", i.StartsAt.In(domain.HospitalLocation).Format("02 Jan 2006 15:04")), "appointment:"+id)
		if err != nil {
			return empty, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'appointment.booked',$2)`, actor.ID, id); err != nil {
		return empty, err
	}
	out, err := scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+`WHERE a.id=$1`, id))
	if err != nil {
		return empty, err
	}
	return out, tx.Commit(ctx)
}
func (s Store) ChangeAppointment(ctx context.Context, actor domain.Actor, id string, c domain.AppointmentChange, now time.Time) (domain.Appointment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Appointment{}, err
	}
	defer tx.Rollback(ctx)
	a, err := scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+`WHERE `+appointmentScope+` AND a.id=$3 FOR UPDATE OF a`, actor.Role, actor.ID, id))
	if err != nil {
		return a, err
	}
	if a.Version != c.Version {
		return a, domain.ErrStale
	}
	if !domain.CanTransition(actor.Role, a.Status, c.Status, a.StartsAt, now) {
		return a, domain.ErrValidation
	}
	_, err = tx.Exec(ctx, `UPDATE appointment SET status=$2,version=version+1,updated_at=now() WHERE id=$1`, id, c.Status)
	if err != nil {
		return a, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,$2,$3)`, actor.ID, "appointment."+c.Status, id); err != nil {
		return a, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO appointment_status_event(appointment_id,previous_status,next_status,actor_id,reason,version) VALUES($1,$2,$3,$4,$5,$6)`, id, a.Status, c.Status, actor.ID, c.Reason, a.Version+1); err != nil {
		return a, err
	}
	a.Status = c.Status
	a.Version++
	return a, tx.Commit(ctx)
}

func (s Store) DoctorSchedules(ctx context.Context, doctorID string) ([]domain.DoctorSchedule, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT p.user_id, u.name, p.slot_minutes
		FROM doctor_profile p
		JOIN "user" u ON u.id = p.user_id
		JOIN staff_access sa ON sa.user_id = p.user_id
		WHERE sa.active = true AND ($1 = '' OR p.user_id = $1)
		ORDER BY u.name ASC
	`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type docMeta struct {
		id          string
		name        string
		slotMinutes int
	}
	var list []docMeta
	for rows.Next() {
		var m docMeta
		if err := rows.Scan(&m.id, &m.name, &m.slotMinutes); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var results []domain.DoctorSchedule
	for _, m := range list {
		hourRows, err := s.DB.Query(ctx, `
			SELECT weekday, start_minute, end_minute
			FROM doctor_hours
			WHERE doctor_id = $1
			ORDER BY weekday ASC, start_minute ASC
		`, m.id)
		if err != nil {
			return nil, err
		}
		var days []domain.ScheduleDayRow
		for hourRows.Next() {
			var w, sMin, eMin int
			if err := hourRows.Scan(&w, &sMin, &eMin); err != nil {
				hourRows.Close()
				return nil, err
			}
			dayName := "Sunday"
			if w >= 0 && w < 7 {
				dayName = domain.WeekdayNames[w]
			}
			days = append(days, domain.ScheduleDayRow{
				Day:  dayName,
				From: domain.FormatMinutesToTime(sMin),
				To:   domain.FormatMinutesToTime(eMin),
			})
		}
		hourRows.Close()
		if err := hourRows.Err(); err != nil {
			return nil, err
		}

		results = append(results, domain.DoctorSchedule{
			ID:             m.id,
			DoctorID:       m.id,
			DoctorName:     m.name,
			PerPatientTime: domain.FormatMinutesToTime(m.slotMinutes),
			SlotMinutes:    m.slotMinutes,
			Days:           days,
		})
	}
	return results, nil
}

func (s Store) DoctorSchedule(ctx context.Context, doctorID string) (domain.DoctorSchedule, error) {
	list, err := s.DoctorSchedules(ctx, doctorID)
	if err != nil {
		return domain.DoctorSchedule{}, err
	}
	if len(list) == 0 {
		return domain.DoctorSchedule{}, domain.ErrNotFound
	}
	return list[0], nil
}

func (s Store) SaveDoctorSchedule(ctx context.Context, a domain.Actor, input domain.SaveDoctorScheduleInput) (domain.DoctorSchedule, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorSchedule{}, err
	}
	defer tx.Rollback(ctx)

	var lockedID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, input.DoctorID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorSchedule{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorSchedule{}, err
	}

	var docName string
	err = tx.QueryRow(ctx, `
		SELECT u.name FROM doctor_profile p
		JOIN "user" u ON u.id = p.user_id
		WHERE p.user_id = $1
	`, input.DoctorID).Scan(&docName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorSchedule{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorSchedule{}, err
	}

	type apptTime struct {
		startsAt time.Time
		endsAt   time.Time
	}
	apptRows, err := tx.Query(ctx, `
		SELECT starts_at, ends_at FROM appointment
		WHERE doctor_id = $1 AND status IN ('booked', 'arrived') AND starts_at > now()
	`, input.DoctorID)
	if err != nil {
		return domain.DoctorSchedule{}, err
	}
	var futureAppts []apptTime
	for apptRows.Next() {
		var at apptTime
		if err := apptRows.Scan(&at.startsAt, &at.endsAt); err != nil {
			apptRows.Close()
			return domain.DoctorSchedule{}, err
		}
		futureAppts = append(futureAppts, at)
	}
	apptRows.Close()
	if err := apptRows.Err(); err != nil {
		return domain.DoctorSchedule{}, err
	}

	type interval struct {
		start, end int
	}
	newHours := make(map[int][]interval)
	for _, d := range input.Days {
		w := domain.WeekdayNameToIndex(d.Day)
		fromMin, _ := domain.ParseTimeToMinutes(d.From)
		toMin, _ := domain.ParseTimeToMinutes(d.To)
		if fromMin < toMin {
			newHours[w] = append(newHours[w], interval{start: fromMin, end: toMin})
		}
	}

	for _, appt := range futureAppts {
		eatStart := appt.startsAt.In(domain.HospitalLocation)
		eatEnd := appt.endsAt.In(domain.HospitalLocation)
		w := int(eatStart.Weekday())
		startMin := eatStart.Hour()*60 + eatStart.Minute()
		endMin := eatEnd.Hour()*60 + eatEnd.Minute()

		allowed := false
		for _, iv := range newHours[w] {
			if iv.start <= startMin && endMin <= iv.end {
				allowed = true
				break
			}
		}
		if !allowed {
			return domain.DoctorSchedule{}, domain.ErrConflict
		}
	}

	_, err = tx.Exec(ctx, `UPDATE doctor_profile SET slot_minutes = $2, version = version + 1 WHERE user_id = $1`, input.DoctorID, input.SlotMinutes)
	if err != nil {
		return domain.DoctorSchedule{}, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM doctor_hours WHERE doctor_id = $1`, input.DoctorID)
	if err != nil {
		return domain.DoctorSchedule{}, err
	}

	var savedDays []domain.ScheduleDayRow
	for _, d := range input.Days {
		w := domain.WeekdayNameToIndex(d.Day)
		fromMin, _ := domain.ParseTimeToMinutes(d.From)
		toMin, _ := domain.ParseTimeToMinutes(d.To)
		if fromMin < toMin {
			_, err = tx.Exec(ctx, `INSERT INTO doctor_hours (doctor_id, weekday, start_minute, end_minute) VALUES ($1, $2, $3, $4)`, input.DoctorID, w, fromMin, toMin)
			if err != nil {
				return domain.DoctorSchedule{}, err
			}
			savedDays = append(savedDays, domain.ScheduleDayRow{
				Day:  domain.WeekdayNames[w],
				From: domain.FormatMinutesToTime(fromMin),
				To:   domain.FormatMinutesToTime(toMin),
			})
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor.schedule_saved', $2)`, a.ID, input.DoctorID)
	if err != nil {
		return domain.DoctorSchedule{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DoctorSchedule{}, err
	}

	return domain.DoctorSchedule{
		ID:             input.DoctorID,
		DoctorID:       input.DoctorID,
		DoctorName:     docName,
		PerPatientTime: domain.FormatMinutesToTime(input.SlotMinutes),
		SlotMinutes:    input.SlotMinutes,
		Days:           savedDays,
	}, nil
}

func (s Store) DeleteDoctorSchedule(ctx context.Context, a domain.Actor, doctorID string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, doctorID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}

	var hasAppointments bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment WHERE doctor_id = $1 AND status NOT IN ('cancelled'))`, doctorID).Scan(&hasAppointments)
	if err != nil {
		return err
	}
	if hasAppointments {
		return domain.ErrInUse
	}

	_, err = tx.Exec(ctx, `DELETE FROM doctor_hours WHERE doctor_id = $1`, doctorID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor.schedule_deleted', $2)`, a.ID, doctorID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) DoctorHolidays(ctx context.Context, doctorID string) ([]domain.DoctorHoliday, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT h.id::text, h.doctor_id, u.name, h.holiday_date::text, h.name, h.created_at
		FROM doctor_holiday h
		JOIN "user" u ON u.id = h.doctor_id
		WHERE ($1 = '' OR h.doctor_id = $1)
		ORDER BY h.holiday_date DESC, h.created_at DESC
	`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DoctorHoliday
	for rows.Next() {
		var item domain.DoctorHoliday
		if err := rows.Scan(&item.ID, &item.DoctorID, &item.DoctorName, &item.Date, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s Store) CreateDoctorHoliday(ctx context.Context, a domain.Actor, input domain.CreateDoctorHolidayInput) (domain.DoctorHoliday, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorHoliday{}, err
	}
	defer tx.Rollback(ctx)

	var lockedID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, input.DoctorID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorHoliday{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorHoliday{}, err
	}

	var docName string
	err = tx.QueryRow(ctx, `
		SELECT u.name FROM doctor_profile p
		JOIN "user" u ON u.id = p.user_id
		WHERE p.user_id = $1
	`, input.DoctorID).Scan(&docName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorHoliday{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorHoliday{}, err
	}

	var existsHoliday bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM doctor_holiday WHERE doctor_id = $1 AND holiday_date = $2::date)
	`, input.DoctorID, input.Date).Scan(&existsHoliday)
	if err != nil {
		return domain.DoctorHoliday{}, err
	}
	if existsHoliday {
		return domain.DoctorHoliday{}, domain.ErrConflict
	}

	var hasAppt bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM appointment
			WHERE doctor_id = $1 AND status NOT IN ('cancelled')
			AND (starts_at AT TIME ZONE 'Africa/Addis_Ababa')::date = $2::date
		)
	`, input.DoctorID, input.Date).Scan(&hasAppt)
	if err != nil {
		return domain.DoctorHoliday{}, err
	}
	if hasAppt {
		return domain.DoctorHoliday{}, domain.ErrConflict
	}

	var id string
	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO doctor_holiday (doctor_id, holiday_date, name, created_by)
		VALUES ($1, $2::date, $3, $4)
		RETURNING id::text, created_at
	`, input.DoctorID, input.Date, input.Reason, a.ID).Scan(&id, &createdAt)
	if err != nil {
		return domain.DoctorHoliday{}, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_holiday.created', $2)`, a.ID, id)
	if err != nil {
		return domain.DoctorHoliday{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DoctorHoliday{}, err
	}

	return domain.DoctorHoliday{
		ID:         id,
		DoctorID:   input.DoctorID,
		DoctorName: docName,
		Date:       input.Date,
		Reason:     input.Reason,
		CreatedAt:  createdAt,
	}, nil
}

func (s Store) DeleteDoctorHoliday(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var docID string
	err = tx.QueryRow(ctx, `SELECT doctor_id FROM doctor_holiday WHERE id = $1::uuid`, id).Scan(&docID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if a.Role == "doctor" && docID != a.ID {
		return domain.ErrForbidden
	}

	var lockedID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' FOR UPDATE`, docID).Scan(&lockedID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM doctor_holiday WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_holiday.deleted', $2)`, a.ID, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) DoctorBreaks(ctx context.Context, doctorID string) ([]domain.DoctorLunchBreak, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT b.id::text, b.doctor_id, u.name, u.email,
		       b.break_from::text, b.break_to::text, b.every_day,
		       COALESCE(b.break_date::text, ''), b.created_at
		FROM doctor_lunch_break b
		JOIN "user" u ON u.id = b.doctor_id
		WHERE ($1 = '' OR b.doctor_id = $1)
		ORDER BY b.created_at DESC
	`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DoctorLunchBreak
	for rows.Next() {
		var item domain.DoctorLunchBreak
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.DoctorID, &item.DoctorName, &item.DoctorEmail, &item.BreakFrom, &item.BreakTo, &item.EveryDay, &item.Date, &createdAt); err != nil {
			return nil, err
		}
		if item.EveryDay {
			item.DateType = "Every Day"
		} else {
			item.DateType = item.Date
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s Store) CreateDoctorBreak(ctx context.Context, a domain.Actor, input domain.CreateDoctorBreakInput) (domain.DoctorLunchBreak, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorLunchBreak{}, err
	}
	defer tx.Rollback(ctx)

	var lockedID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, input.DoctorID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorLunchBreak{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorLunchBreak{}, err
	}

	var docName, docEmail string
	err = tx.QueryRow(ctx, `
		SELECT u.name, u.email FROM doctor_profile p
		JOIN "user" u ON u.id = p.user_id
		WHERE p.user_id = $1
	`, input.DoctorID).Scan(&docName, &docEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorLunchBreak{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorLunchBreak{}, err
	}

	if input.EveryDay {
		var duplicate bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM doctor_lunch_break
				WHERE doctor_id = $1 AND every_day = true
				AND break_from = $2::time AND break_to = $3::time
			)
		`, input.DoctorID, input.BreakFrom, input.BreakTo).Scan(&duplicate)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}
		if duplicate {
			return domain.DoctorLunchBreak{}, domain.ErrConflict
		}

		var hasAppt bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM appointment
				WHERE doctor_id = $1 AND status NOT IN ('cancelled') AND starts_at > now()
				AND (starts_at AT TIME ZONE 'Africa/Addis_Ababa')::time < $3::time
				AND (ends_at AT TIME ZONE 'Africa/Addis_Ababa')::time > $2::time
			)
		`, input.DoctorID, input.BreakFrom, input.BreakTo).Scan(&hasAppt)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}
		if hasAppt {
			return domain.DoctorLunchBreak{}, domain.ErrConflict
		}

		var id string
		err = tx.QueryRow(ctx, `
			INSERT INTO doctor_lunch_break (doctor_id, break_from, break_to, every_day, break_date, created_by)
			VALUES ($1, $2::time, $3::time, true, NULL, $4)
			RETURNING id::text
		`, input.DoctorID, input.BreakFrom, input.BreakTo, a.ID).Scan(&id)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}

		_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_break.created', $2)`, a.ID, id)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return domain.DoctorLunchBreak{}, err
		}

		return domain.DoctorLunchBreak{
			ID:          id,
			DoctorID:    input.DoctorID,
			DoctorName:  docName,
			DoctorEmail: docEmail,
			BreakFrom:   input.BreakFrom,
			BreakTo:     input.BreakTo,
			EveryDay:    true,
			DateType:    "Every Day",
		}, nil
	} else {
		var duplicate bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM doctor_lunch_break
				WHERE doctor_id = $1 AND every_day = false AND break_date = $4::date
				AND break_from = $2::time AND break_to = $3::time
			)
		`, input.DoctorID, input.BreakFrom, input.BreakTo, input.Date).Scan(&duplicate)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}
		if duplicate {
			return domain.DoctorLunchBreak{}, domain.ErrConflict
		}

		var hasAppt bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM appointment
				WHERE doctor_id = $1 AND status NOT IN ('cancelled')
				AND (starts_at AT TIME ZONE 'Africa/Addis_Ababa')::date = $4::date
				AND (starts_at AT TIME ZONE 'Africa/Addis_Ababa')::time < $3::time
				AND (ends_at AT TIME ZONE 'Africa/Addis_Ababa')::time > $2::time
			)
		`, input.DoctorID, input.BreakFrom, input.BreakTo, input.Date).Scan(&hasAppt)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}
		if hasAppt {
			return domain.DoctorLunchBreak{}, domain.ErrConflict
		}

		var id string
		err = tx.QueryRow(ctx, `
			INSERT INTO doctor_lunch_break (doctor_id, break_from, break_to, every_day, break_date, created_by)
			VALUES ($1, $2::time, $3::time, false, $4::date, $5)
			RETURNING id::text
		`, input.DoctorID, input.BreakFrom, input.BreakTo, input.Date, a.ID).Scan(&id)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}

		_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_break.created', $2)`, a.ID, id)
		if err != nil {
			return domain.DoctorLunchBreak{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return domain.DoctorLunchBreak{}, err
		}

		return domain.DoctorLunchBreak{
			ID:          id,
			DoctorID:    input.DoctorID,
			DoctorName:  docName,
			DoctorEmail: docEmail,
			BreakFrom:   input.BreakFrom,
			BreakTo:     input.BreakTo,
			EveryDay:    false,
			Date:        input.Date,
			DateType:    input.Date,
		}, nil
	}
}

func (s Store) DeleteDoctorBreak(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var docID string
	err = tx.QueryRow(ctx, `SELECT doctor_id FROM doctor_lunch_break WHERE id = $1::uuid`, id).Scan(&docID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if a.Role == "doctor" && docID != a.ID {
		return domain.ErrForbidden
	}

	var lockedID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' FOR UPDATE`, docID).Scan(&lockedID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM doctor_lunch_break WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_break.deleted', $2)`, a.ID, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) DoctorOPDCharges(ctx context.Context, search string) ([]domain.DoctorOPDCharge, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT COALESCE(c.id::text, p.user_id), p.user_id, u.name, COALESCE(p.department, ''),
		       COALESCE(c.standard_charge, p.opd_charge, 0)::float8, COALESCE(c.currency_symbol, 'ETB'),
		       COALESCE(c.created_at, now()), COALESCE(c.updated_at, now())
		FROM doctor_profile p
		JOIN "user" u ON u.id = p.user_id
		LEFT JOIN doctor_opd_charge c ON c.doctor_id = p.user_id
		WHERE ($1 = '' OR strpos(lower(u.name), lower($1)) > 0 OR strpos(lower(p.department), lower($1)) > 0)
		ORDER BY u.name ASC
	`, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DoctorOPDCharge
	for rows.Next() {
		var item domain.DoctorOPDCharge
		if err := rows.Scan(&item.ID, &item.DoctorID, &item.DoctorName, &item.DoctorDepartment, &item.StandardCharge, &item.CurrencySymbol, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s Store) DoctorOPDCharge(ctx context.Context, doctorID string) (domain.DoctorOPDCharge, error) {
	var item domain.DoctorOPDCharge
	err := s.DB.QueryRow(ctx, `
		SELECT c.id::text, c.doctor_id, u.name, COALESCE(p.department, ''),
		       c.standard_charge::float8, c.currency_symbol, c.created_at, c.updated_at
		FROM doctor_opd_charge c
		JOIN "user" u ON u.id = c.doctor_id
		JOIN doctor_profile p ON p.user_id = c.doctor_id
		WHERE c.doctor_id = $1
	`, doctorID).Scan(&item.ID, &item.DoctorID, &item.DoctorName, &item.DoctorDepartment, &item.StandardCharge, &item.CurrencySymbol, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		var name, dept string
		var opdCharge float64
		err2 := s.DB.QueryRow(ctx, `
			SELECT u.name, COALESCE(p.department, ''), p.opd_charge::float8
			FROM doctor_profile p
			JOIN "user" u ON u.id = p.user_id
			WHERE p.user_id = $1
		`, doctorID).Scan(&name, &dept, &opdCharge)
		if err2 != nil {
			return domain.DoctorOPDCharge{}, domain.ErrNotFound
		}
		return domain.DoctorOPDCharge{
			ID:               doctorID,
			DoctorID:         doctorID,
			DoctorName:       name,
			DoctorDepartment: dept,
			StandardCharge:   opdCharge,
			CurrencySymbol:   "ETB",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}, nil
	}
	if err != nil {
		return domain.DoctorOPDCharge{}, err
	}
	return item, nil
}

func (s Store) SaveDoctorOPDCharge(ctx context.Context, a domain.Actor, input domain.SaveDoctorOPDChargeInput) (domain.DoctorOPDCharge, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorOPDCharge{}, err
	}
	defer tx.Rollback(ctx)

	var docName, dept string
	err = tx.QueryRow(ctx, `
		SELECT u.name, COALESCE(p.department, '')
		FROM doctor_profile p
		JOIN "user" u ON u.id = p.user_id
		WHERE p.user_id = $1
		FOR UPDATE
	`, input.DoctorID).Scan(&docName, &dept)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorOPDCharge{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorOPDCharge{}, err
	}

	var id string
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO doctor_opd_charge (doctor_id, standard_charge, currency_symbol, created_by, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (doctor_id) DO UPDATE SET
			standard_charge = EXCLUDED.standard_charge,
			currency_symbol = EXCLUDED.currency_symbol,
			updated_at = now()
		RETURNING id::text, created_at, updated_at
	`, input.DoctorID, input.StandardCharge, input.CurrencySymbol, a.ID).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return domain.DoctorOPDCharge{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE doctor_profile SET opd_charge = $2, version = version + 1
		WHERE user_id = $1
	`, input.DoctorID, input.StandardCharge)
	if err != nil {
		return domain.DoctorOPDCharge{}, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_opd_charge.saved', $2)`, a.ID, id)
	if err != nil {
		return domain.DoctorOPDCharge{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DoctorOPDCharge{}, err
	}

	return domain.DoctorOPDCharge{
		ID:               id,
		DoctorID:         input.DoctorID,
		DoctorName:       docName,
		DoctorDepartment: dept,
		StandardCharge:   input.StandardCharge,
		CurrencySymbol:   input.CurrencySymbol,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

func (s Store) DeleteDoctorOPDCharge(ctx context.Context, a domain.Actor, doctorID string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM doctor_opd_charge WHERE doctor_id = $1`, doctorID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE doctor_profile SET opd_charge = 0, version = version + 1 WHERE user_id = $1`, doctorID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_opd_charge.deleted', $2)`, a.ID, doctorID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}


