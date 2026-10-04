package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
	"time"
)

type querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func doctor(ctx context.Context, q querier, id string) (domain.Doctor, error) {
	d := domain.Doctor{Hours: []domain.DoctorHours{}}
	err := q.QueryRow(ctx, `SELECT d.user_id,u.name,d.department,d.slot_minutes,d.version FROM doctor_profile d JOIN "user" u ON u.id=d.user_id JOIN staff_access a ON a.user_id=d.user_id WHERE d.user_id=$1 AND a.active AND a.role='doctor'`, id).Scan(&d.ID, &d.Name, &d.Department, &d.SlotMinutes, &d.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	if err != nil {
		return d, err
	}
	rows, err := q.Query(ctx, `SELECT weekday,start_minute,end_minute FROM doctor_hours WHERE doctor_id=$1 ORDER BY weekday,start_minute`, id)
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
	rows, err := s.DB.Query(ctx, `SELECT starts_at,ends_at FROM appointment WHERE doctor_id=$1 AND status<>'cancelled' AND starts_at<$3 AND ends_at>$2 UNION ALL SELECT starts_at,ends_at FROM doctor_absence WHERE doctor_id=$1 AND starts_at<$3 AND ends_at>$2`, id, day, day.AddDate(0, 0, 1))
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
const appointmentScope = `($1 IN ('admin','receptionist') OR ($1='doctor' AND a.doctor_id=$2) OR ($1='patient' AND p.user_id=$2))`

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
		if old.PatientID != i.PatientID || old.DoctorID != i.DoctorID || !old.StartsAt.Equal(i.StartsAt) || old.Problem != i.Problem || old.NotifySMS != i.NotifySMS {
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
	err = tx.QueryRow(ctx, `SELECT phone FROM patient WHERE id=$1 AND ($2 IN ('admin','receptionist') OR ($2='patient' AND user_id=$3) OR ($2='doctor' AND clinician_user_id=$3)) FOR UPDATE`, i.PatientID, actor.Role, actor.ID).Scan(&phone)
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
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment WHERE (doctor_id=$1 OR patient_id=$2) AND status<>'cancelled' AND starts_at<$4 AND ends_at>$3) OR EXISTS(SELECT 1 FROM doctor_absence WHERE doctor_id=$1 AND starts_at<$4 AND ends_at>$3)`, i.DoctorID, i.PatientID, i.StartsAt, end).Scan(&occupied)
	if err != nil {
		return empty, err
	}
	if occupied {
		return empty, domain.ErrStale
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
	a.Status = c.Status
	a.Version++
	return a, tx.Commit(ctx)
}
