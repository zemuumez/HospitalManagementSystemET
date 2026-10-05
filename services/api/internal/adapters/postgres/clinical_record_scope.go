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
 OR ($5='vaccination_read' AND $1='patient' AND patient_portal_owner(p.id)=$2)
 OR ($5='diagnostic' AND $1='lab_technician' AND EXISTS(SELECT 1 FROM encounter e JOIN diagnostic_order o ON o.encounter_id=e.id WHERE e.patient_id=p.id AND ($4='' OR e.id=NULLIF($4,'')::uuid)))
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

func (s Store) AuthorizeDiagnosticOrder(ctx context.Context, a domain.Actor, id string) error {
	var allowed bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM diagnostic_order o JOIN encounter e ON e.id=o.encounter_id JOIN patient p ON p.id=e.patient_id WHERE o.id=$3 AND ($1 IN ('admin','lab_technician') OR ($1='doctor' AND e.doctor_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2)))`, a.Role, a.ID, id).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

func (s Store) ValidateClinicalAttribution(ctx context.Context, a domain.Actor, patientID, doctorID string, encounterID *string) error {
	if !domain.UUIDPattern.MatchString(patientID) || (encounterID != nil && !domain.UUIDPattern.MatchString(*encounterID)) {
		return domain.ErrValidation
	}
	if a.Role == "doctor" && a.ID != doctorID {
		return domain.ErrForbidden
	}
	var allowed bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient p JOIN staff_access s ON s.user_id=$2 AND s.active AND s.role='doctor'
 WHERE p.id=$1 AND canonical_patient_id(p.id)=p.id AND ($3::uuid IS NULL OR EXISTS(SELECT 1 FROM encounter e WHERE e.id=$3::uuid AND e.patient_id=p.id AND e.doctor_id=$2))
 AND ($4<>'doctor' OR p.clinician_user_id=$5 OR EXISTS(SELECT 1 FROM encounter e WHERE e.patient_id=p.id AND e.doctor_id=$5)))`, patientID, doctorID, encounterID, a.Role, a.ID).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

func (s Store) AuthorizeAppointment(ctx context.Context, a domain.Actor, id string) error {
	var allowed bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM appointment ap JOIN patient p ON p.id=ap.patient_id WHERE ap.id=$3 AND ($1 IN ('admin','accountant','receptionist') OR ($1='doctor' AND ap.doctor_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2)))`, a.Role, a.ID, id).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}
