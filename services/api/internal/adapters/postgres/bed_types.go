package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) BedTypes(ctx context.Context, page int) ([]domain.BedType, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,name,description,active,version FROM bed_type WHERE active = true ORDER BY name,id LIMIT 25 OFFSET $1`, (page-1)*25)
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

func (s Store) SearchBedTypes(ctx context.Context, search string, page, pageSize int) ([]domain.BedType, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	baseQuery := `FROM bed_type bt WHERE bt.active = true`
	args := []any{}
	argIdx := 1

	if search = strings.TrimSpace(search); search != "" {
		baseQuery += fmt.Sprintf(` AND (bt.name ILIKE $%d OR bt.description ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	if err := s.DB.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT bt.id, bt.name, bt.description, bt.active, bt.version,
		       COALESCE((SELECT COUNT(*) FROM hospital_bed b WHERE b.type_id = bt.id AND b.active = true), 0) AS beds_count
		%s
		ORDER BY bt.name, bt.id
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.DB.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.BedType{}
	for rows.Next() {
		var v domain.BedType
		if err := rows.Scan(&v.ID, &v.Name, &v.Description, &v.Active, &v.Version, &v.BedsCount); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
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

func (s Store) DeleteBedType(ctx context.Context, a domain.Actor, id string, version int) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentVersion int
	err = tx.QueryRow(ctx, `SELECT version FROM bed_type WHERE id = $1 AND active = true FOR UPDATE`, id).Scan(&currentVersion)
	if err != nil {
		return clinicalError(err)
	}
	if version <= 0 || currentVersion != version {
		return domain.ErrStale
	}

	// In-use protection: prevent deletion if active beds exist with this type
	var bedCount int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM hospital_bed WHERE type_id = $1 AND active = true`, id).Scan(&bedCount)
	if err != nil {
		return err
	}
	if bedCount > 0 {
		return domain.ErrInUse
	}

	// Check if any inactive beds reference this type
	var anyBeds int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM hospital_bed WHERE type_id = $1`, id).Scan(&anyBeds)
	if err != nil {
		return err
	}
	if anyBeds > 0 {
		// Soft delete to avoid FK integrity violation
		if _, err = tx.Exec(ctx, `UPDATE bed_type SET active = false, version = version + 1 WHERE id = $1`, id); err != nil {
			return err
		}
	} else {
		if _, err = tx.Exec(ctx, `DELETE FROM bed_type WHERE id = $1`, id); err != nil {
			return err
		}
	}

	if err = pharmacyAudit(ctx, tx, a, "bed_type.deleted", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
