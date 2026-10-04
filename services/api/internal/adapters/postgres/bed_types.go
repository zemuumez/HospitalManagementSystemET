package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) BedTypes(ctx context.Context, page int) ([]domain.BedType, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,name,description,active,version FROM bed_type ORDER BY name,id LIMIT 25 OFFSET $1`, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.BedType{}
	for rows.Next() {
		var v domain.BedType
		if e = rows.Scan(&v.ID, &v.Name, &v.Description, &v.Active, &v.Version); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s Store) SaveBedType(ctx context.Context, a domain.Actor, id string, i domain.BedTypeInput) (domain.BedType, error) {
	out := domain.BedType{ID: id, BedTypeInput: i}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if id == "" {
		e = tx.QueryRow(ctx, `INSERT INTO bed_type(name,description,active) VALUES($1,$2,$3) RETURNING id,version`, i.Name, i.Description, i.Active).Scan(&out.ID, &out.Version)
	} else {
		e = tx.QueryRow(ctx, `UPDATE bed_type SET name=$2,description=$3,active=$4,version=version+1 WHERE id=$1 AND version=$5 RETURNING version`, id, i.Name, i.Description, i.Active, i.Version).Scan(&out.Version)
	}
	if e != nil {
		if id != "" && errors.Is(e, pgx.ErrNoRows) {
			return out, domain.ErrStale
		}
		return out, clinicalError(e)
	}
	if id != "" {
		if _, e = tx.Exec(ctx, `UPDATE hospital_bed SET bed_type=$2 WHERE type_id=$1`, id, i.Name); e != nil {
			return out, e
		}
	}
	if e = pharmacyAudit(ctx, tx, a, "bed_type.saved", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
