package postgres

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const staffSelect = `SELECT u.id,u.name,u.email,a.role,a.active,p.version,p.details FROM "user" u JOIN staff_access a ON a.user_id=u.id JOIN staff_profile p ON p.user_id=u.id `

func scanStaff(row pgx.Row) (domain.StaffProfile, error) {
	var p domain.StaffProfile
	var raw []byte
	e := row.Scan(&p.ID, &p.Name, &p.Email, &p.Role, &p.Active, &p.Version, &raw)
	if e != nil {
		return p, clinicalError(e)
	}
	e = json.Unmarshal(raw, &p.Details)
	return p, e
}
func staffAdmin(ctx context.Context, tx pgx.Tx, a domain.Actor) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if _, e := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(72841022)`); e != nil {
		return e
	}
	var ok bool
	if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM staff_access WHERE user_id=$1 AND role='admin' AND active)`, a.ID).Scan(&ok); e != nil {
		return e
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}
func (s Store) StaffProfile(ctx context.Context, a domain.Actor, id string) (domain.StaffProfile, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.StaffProfile{}, e
	}
	defer tx.Rollback(ctx)
	if e = staffAdmin(ctx, tx, a); e != nil {
		return domain.StaffProfile{}, e
	}
	p, e := scanStaff(tx.QueryRow(ctx, staffSelect+`WHERE u.id=$1`, id))
	if e != nil {
		return p, e
	}
	if e = pharmacyAudit(ctx, tx, a, "staff.profile_viewed", id); e != nil {
		return p, e
	}
	return p, tx.Commit(ctx)
}
func (s Store) UpdateStaffProfile(ctx context.Context, a domain.Actor, id string, i domain.StaffProfileInput) (domain.StaffProfile, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.StaffProfile{}, e
	}
	defer tx.Rollback(ctx)
	if e = staffAdmin(ctx, tx, a); e != nil {
		return domain.StaffProfile{}, e
	}
	before, e := scanStaff(tx.QueryRow(ctx, staffSelect+`WHERE u.id=$1 FOR UPDATE OF p,u`, id))
	if e != nil {
		return before, e
	}
	if before.Version != i.Version {
		return before, domain.ErrStale
	}
	details, e := json.Marshal(i.Details)
	if e != nil {
		return before, e
	}
	if _, e = tx.Exec(ctx, `UPDATE staff_profile SET details=$2,version=version+1 WHERE user_id=$1`, id, details); e != nil {
		return before, e
	}
	if _, e = tx.Exec(ctx, `UPDATE "user" SET name=$2,"updatedAt"=now() WHERE id=$1`, id, i.Details.GivenName+" "+i.Details.FamilyName); e != nil {
		return before, e
	}
	after, e := scanStaff(tx.QueryRow(ctx, staffSelect+`WHERE u.id=$1`, id))
	if e != nil {
		return after, e
	}
	b, e := json.Marshal(before)
	if e != nil {
		return after, e
	}
	n, e := json.Marshal(after)
	if e != nil {
		return after, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO staff_profile_revision(user_id,version,actor_id,reason,before_snapshot,after_snapshot) VALUES($1,$2,$3,$4,$5,$6)`, id, after.Version, a.ID, i.Reason, b, n); e != nil {
		return after, e
	}
	if e = pharmacyAudit(ctx, tx, a, "staff.profile_updated", id); e != nil {
		return after, e
	}
	return after, tx.Commit(ctx)
}
func (s Store) ChangeStaffRole(ctx context.Context, a domain.Actor, id string, i domain.StaffRoleInput) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if e = staffAdmin(ctx, tx, a); e != nil {
		return e
	}
	if a.ID == id {
		return domain.ErrStale
	}
	var role string
	if e = tx.QueryRow(ctx, `SELECT role FROM staff_access WHERE user_id=$1 FOR UPDATE`, id).Scan(&role); e != nil {
		return clinicalError(e)
	}
	if role != i.PreviousRole {
		return domain.ErrStale
	}
	// Assignment and booking writers lock this staff row; portal linking shares the advisory lock.
	var busy bool
	if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM encounter WHERE doctor_id=$1 AND status='active') OR EXISTS(SELECT 1 FROM appointment WHERE doctor_id=$1 AND status IN ('booked','arrived')) OR EXISTS(SELECT 1 FROM encounter_nurse n JOIN encounter e ON e.id=n.encounter_id WHERE n.nurse_id=$1 AND n.active AND e.status='active') OR EXISTS(SELECT 1 FROM patient WHERE user_id=$1 OR clinician_user_id=$1)`, id).Scan(&busy); e != nil {
		return e
	}
	if busy {
		return domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `UPDATE staff_access SET role=$2 WHERE user_id=$1`, id, i.Role); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `DELETE FROM session WHERE "userId"=$1`, id); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `DELETE FROM verification WHERE value=$1 AND (identifier LIKE '2fa-%' OR identifier LIKE 'trust-device-%')`, id); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO staff_role_event(user_id,actor_id,previous_role,next_role,reason) VALUES($1,$2,$3,$4,$5)`, id, a.ID, role, i.Role, i.Reason); e != nil {
		return e
	}
	if e = pharmacyAudit(ctx, tx, a, "staff.role_changed", id); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
