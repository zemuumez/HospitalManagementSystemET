package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
	"time"
)

const absenceFields = `id,doctor_id,starts_at,ends_at,reason,version,cancelled_at,cancel_reason`

func scanAbsence(row pgx.Row) (domain.DoctorAbsence, error) {
	var a domain.DoctorAbsence
	e := row.Scan(&a.ID, &a.DoctorID, &a.StartsAt, &a.EndsAt, &a.Reason, &a.Version, &a.CancelledAt, &a.CancelReason)
	return a, clinicalError(e)
}
func (s Store) Absences(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DoctorAbsence, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && id != "" && a.ID == id) {
		return nil, domain.ErrForbidden
	}
	rows, e := s.DB.Query(ctx, `SELECT `+absenceFields+` FROM doctor_absence WHERE ($1='' OR doctor_id=$1) ORDER BY starts_at DESC,id LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.DoctorAbsence{}
	for rows.Next() {
		v, e := scanAbsence(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s Store) CreateAbsence(ctx context.Context, a domain.Actor, i domain.AbsenceInput) (domain.DoctorAbsence, error) {
	var out domain.DoctorAbsence
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	var doctorID string
	if e = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, i.DoctorID).Scan(&doctorID); e != nil {
		return out, clinicalError(e)
	}
	var busy bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment WHERE doctor_id=$1 AND status IN ('booked','arrived') AND starts_at<$3 AND ends_at>$2) OR EXISTS(SELECT 1 FROM doctor_absence WHERE doctor_id=$1 AND cancelled_at IS NULL AND starts_at<$3 AND ends_at>$2)`, i.DoctorID, i.StartsAt, i.EndsAt).Scan(&busy)
	if e != nil {
		return out, e
	}
	if busy {
		return out, domain.ErrStale
	}
	out, e = scanAbsence(tx.QueryRow(ctx, `INSERT INTO doctor_absence(doctor_id,starts_at,ends_at,reason,created_by) VALUES($1,$2,$3,$4,$5) RETURNING `+absenceFields, i.DoctorID, i.StartsAt, i.EndsAt, i.Reason, a.ID))
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "doctor.absence_created", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) CancelAbsence(ctx context.Context, a domain.Actor, id string, i domain.AbsenceCancel) (domain.DoctorAbsence, error) {
	var out domain.DoctorAbsence
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	var doctorID string
	if e = tx.QueryRow(ctx, `SELECT doctor_id FROM doctor_absence WHERE id=$1 AND ($2='admin' OR doctor_id=$3)`, id, a.Role, a.ID).Scan(&doctorID); e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 FOR UPDATE`, doctorID); e != nil {
		return out, e
	}
	out, e = scanAbsence(tx.QueryRow(ctx, `SELECT `+absenceFields+` FROM doctor_absence WHERE id=$1 FOR UPDATE`, id))
	if e != nil {
		return out, e
	}
	if out.Version != i.Version || out.CancelledAt != nil {
		return out, domain.ErrStale
	}
	out, e = scanAbsence(tx.QueryRow(ctx, `UPDATE doctor_absence SET cancelled_at=clock_timestamp(),cancelled_by=$2,cancel_reason=$3,version=version+1 WHERE id=$1 RETURNING `+absenceFields, id, a.ID, i.Reason))
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "doctor.absence_cancelled", id); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) Reschedule(ctx context.Context, a domain.Actor, id string, i domain.RescheduleInput, now time.Time) (domain.Appointment, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.Appointment{}, e
	}
	defer tx.Rollback(ctx)
	// Lock in the same order as booking: doctor, patient, then the appointment.
	original, e := scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+` WHERE `+appointmentScope+` AND a.id=$3`, a.Role, a.ID, id))
	if e != nil {
		return original, e
	}
	var doctorID string
	if e = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND role='doctor' AND active FOR UPDATE`, original.DoctorID).Scan(&doctorID); e != nil {
		return original, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `SELECT id FROM patient WHERE id=$1 FOR UPDATE`, original.PatientID); e != nil {
		return original, e
	}
	out, e := scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+` WHERE `+appointmentScope+` AND a.id=$3 FOR UPDATE OF a`, a.Role, a.ID, id))
	if e != nil {
		return out, e
	}
	if out.Version != i.Version || out.Status != "booked" || !out.StartsAt.After(now) || out.StartsAt.Equal(i.StartsAt) {
		return out, domain.ErrStale
	}
	d, e := doctor(ctx, tx, out.DoctorID)
	if e != nil {
		return out, e
	}
	end := i.StartsAt.Add(time.Duration(d.SlotMinutes) * time.Minute)
	if !d.Allows(i.StartsAt, end) {
		return out, domain.ErrStale
	}
	var busy bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment WHERE id<>$5 AND (doctor_id=$1 OR patient_id=$2) AND status<>'cancelled' AND starts_at<$4 AND ends_at>$3) OR EXISTS(SELECT 1 FROM doctor_absence WHERE doctor_id=$1 AND cancelled_at IS NULL AND starts_at<$4 AND ends_at>$3)`, out.DoctorID, out.PatientID, i.StartsAt, end, id).Scan(&busy)
	if e != nil {
		return out, e
	}
	if busy {
		return out, domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `INSERT INTO appointment_reschedule(appointment_id,from_start,from_end,to_start,to_end,actor_id,reason,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, out.StartsAt, out.EndsAt, i.StartsAt, end, a.ID, i.Reason, out.Version+1); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `UPDATE appointment SET starts_at=$2,ends_at=$3,version=version+1,updated_at=clock_timestamp() WHERE id=$1`, id, i.StartsAt, end); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "appointment.rescheduled", id); e != nil {
		return out, e
	}
	out.StartsAt = i.StartsAt
	out.EndsAt = end
	out.Version++
	return out, tx.Commit(ctx)
}
func (s Store) RescheduleHistory(ctx context.Context, a domain.Actor, id string, page int) ([]domain.RescheduleEvent, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if _, e = scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+` WHERE `+appointmentScope+` AND a.id=$3`, a.Role, a.ID, id)); e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT id,from_start,from_end,to_start,to_end,actor_id,reason,version,recorded_at FROM appointment_reschedule WHERE appointment_id=$1 ORDER BY version DESC LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.RescheduleEvent{}
	for rows.Next() {
		var v domain.RescheduleEvent
		if e = rows.Scan(&v.ID, &v.FromStart, &v.FromEnd, &v.ToStart, &v.ToEnd, &v.ActorID, &v.Reason, &v.Version, &v.RecordedAt); e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "appointment.reschedule_history_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}

func (s Store) AppointmentStatusHistory(ctx context.Context, a domain.Actor, id string, page int) ([]domain.AppointmentStatusEvent, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if _, e = scanAppointment(tx.QueryRow(ctx, `SELECT `+appointmentColumns+appointmentJoin+` WHERE `+appointmentScope+` AND a.id=$3`, a.Role, a.ID, id)); e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT id,previous_status,next_status,actor_id,reason,version,recorded_at FROM appointment_status_event WHERE appointment_id=$1 ORDER BY version DESC LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.AppointmentStatusEvent{}
	for rows.Next() {
		var v domain.AppointmentStatusEvent
		if e = rows.Scan(&v.ID, &v.PreviousStatus, &v.NextStatus, &v.ActorID, &v.Reason, &v.Version, &v.RecordedAt); e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "appointment.status_history_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
