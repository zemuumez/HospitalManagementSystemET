package postgres

import (
	"context"
	"hms.local/api/internal/domain"
)

// New clinical records are staff-only until a separately tested release policy
// is available. A broad role permission never grants access to arbitrary IDs.
func (s Store) AuthorizeClinicalRecord(ctx context.Context, a domain.Actor, patientID, encounterID, purpose string) error {
	var allowed bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient p
 WHERE ($3='' OR p.id=NULLIF($3,'')::uuid)
 AND ($4='' OR EXISTS(SELECT 1 FROM encounter e WHERE e.id=NULLIF($4,'')::uuid AND e.patient_id=p.id))
 AND ($1='admin'
 OR ($5='billing' AND $1='accountant')
 OR ($5='admission' AND $1='receptionist')
 OR ($1='doctor' AND (($4='' AND p.clinician_user_id=$2) OR EXISTS(SELECT 1 FROM encounter e WHERE e.patient_id=p.id AND e.doctor_id=$2 AND ($4='' OR e.id=NULLIF($4,'')::uuid))))
 OR ($1='nurse' AND $5<>'billing' AND EXISTS(SELECT 1 FROM encounter e JOIN encounter_nurse n ON n.encounter_id=e.id WHERE e.patient_id=p.id AND n.nurse_id=$2 AND n.active AND ($4='' OR e.id=NULLIF($4,'')::uuid)))))`, a.Role, a.ID, patientID, encounterID, purpose).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}
