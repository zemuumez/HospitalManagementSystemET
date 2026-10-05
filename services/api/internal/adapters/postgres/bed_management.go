package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) SetBedState(ctx context.Context, a domain.Actor, id string, i domain.BedStateInput) (domain.Bed, error) {
	var b domain.Bed
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return b, e
	}
	defer tx.Rollback(ctx)
	e = tx.QueryRow(ctx, `SELECT id,name,bed_type,type_id,charge_minor,state,version FROM hospital_bed WHERE id=$1 AND active FOR UPDATE`, id).Scan(&b.ID, &b.Name, &b.Type, &b.TypeID, &b.ChargeMinor, &b.State, &b.Version)
	if e != nil {
		return b, clinicalError(e)
	}
	if b.Version != i.Version {
		return b, domain.ErrStale
	}
	var occupied bool
	if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM encounter WHERE bed_id=$1 AND status='active')`, id).Scan(&occupied); e != nil {
		return b, e
	}
	if occupied && i.State != "ready" {
		return b, domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `UPDATE hospital_bed SET state=$2,version=version+1 WHERE id=$1`, id, i.State); e != nil {
		return b, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO bed_state_event(bed_id,from_state,to_state,actor_id,reason,version) VALUES($1,$2,$3,$4,$5,$6)`, id, b.State, i.State, a.ID, i.Reason, b.Version+1); e != nil {
		return b, e
	}
	if e = pharmacyAudit(ctx, tx, a, "bed.state_changed", id); e != nil {
		return b, e
	}
	b.State = i.State
	b.Version++
	b.Available = !occupied && i.State == "ready"
	return b, tx.Commit(ctx)
}
func (s Store) TransferBed(ctx context.Context, a domain.Actor, id string, i domain.BedTransfer) (domain.Encounter, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Encounter{}, err
	}
	defer tx.Rollback(ctx)
	out, err := transferBedTx(ctx, tx, a, id, i)
	if err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
func transferBedTx(ctx context.Context, tx pgx.Tx, a domain.Actor, id string, i domain.BedTransfer) (domain.Encounter, error) {
	out, e := scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+encounterScope+` AND e.id=$3 FOR UPDATE OF e`, a.Role, a.ID, id))
	if e != nil {
		return out, e
	}
	if out.Kind != "ipd" || out.Status != "active" || out.Version != i.Version || out.BedID == i.BedID {
		return out, domain.ErrStale
	}
	var charge int64
	var name string
	e = tx.QueryRow(ctx, `SELECT name,charge_minor FROM hospital_bed WHERE id=$1 AND active AND state='ready' FOR UPDATE`, i.BedID).Scan(&name, &charge)
	if e != nil {
		return out, clinicalError(e)
	}
	var occupied bool
	if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM encounter WHERE bed_id=$1 AND status='active')`, i.BedID).Scan(&occupied); e != nil {
		return out, e
	}
	if occupied {
		return out, domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `UPDATE encounter SET bed_id=$2,version=version+1 WHERE id=$1`, id, i.BedID); e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `INSERT INTO bed_event(encounter_id,kind,from_bed_id,to_bed_id,charge_minor,actor_id,reason,encounter_version) VALUES($1,'transfer',$2,$3,$4,$5,$6,$7)`, id, out.BedID, i.BedID, charge, a.ID, i.Reason, out.Version+1); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "encounter.bed_transferred", id); e != nil {
		return out, e
	}
	out.BedID = i.BedID
	out.BedName = name
	out.Version++
	return out, nil
}
func (s Store) BedHistory(ctx context.Context, a domain.Actor, id string, page int) ([]domain.BedEvent, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if _, e = scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+encounterScope+` AND e.id=$3`, a.Role, a.ID, id)); e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT id,kind,COALESCE(from_bed_id::text,''),COALESCE(to_bed_id::text,''),charge_minor,actor_id,reason,recorded_at,encounter_version FROM bed_event WHERE encounter_id=$1 ORDER BY encounter_version DESC LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.BedEvent{}
	for rows.Next() {
		var v domain.BedEvent
		if e = rows.Scan(&v.ID, &v.Kind, &v.FromBedID, &v.ToBedID, &v.ChargeMinor, &v.ActorID, &v.Reason, &v.RecordedAt, &v.Version); e != nil {
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
	if e = pharmacyAudit(ctx, tx, a, "encounter.bed_history_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
