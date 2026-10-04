package postgres

import (
	"context"
	"hms.local/api/internal/domain"
)

func (s Store) ReviseDiagnosticTest(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticRevisionInput) (domain.DiagnosticTest, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.DiagnosticTest{}, e
	}
	defer tx.Rollback(ctx)
	old, e := scanDiagnosticTest(tx.QueryRow(ctx, `SELECT `+diagnosticTestFields+` FROM diagnostic_test WHERE id=$1 FOR UPDATE`, id))
	if e != nil {
		return old, e
	}
	if !old.Active || old.Version != i.Version || old.Kind != i.Kind {
		return old, domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `UPDATE diagnostic_test SET active=false,version=version+1 WHERE id=$1`, id); e != nil {
		return old, e
	}
	out, e := scanDiagnosticTest(tx.QueryRow(ctx, `INSERT INTO diagnostic_test(kind,name,short_name,category,method,report_days,charge_minor,created_by,root_id,supersedes,revision) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING `+diagnosticTestFields, i.Kind, i.Name, i.ShortName, i.Category, i.Method, i.ReportDays, i.ChargeMinor, a.ID, old.RootID, id, old.Revision+1))
	if e != nil {
		return out, clinicalError(e)
	}
	for _, p := range i.Parameters {
		if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_parameter(test_id,position,name,unit,reference_range,value_type) VALUES($1,$2,$3,$4,$5,$6)`, out.ID, p.Position, p.Name, p.Unit, p.ReferenceRange, p.ValueType); e != nil {
			return out, e
		}
	}
	out.Parameters = i.Parameters
	if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_catalog_event(test_id,replacement_id,actor_id,reason) VALUES($1,$2,$3,$4)`, id, out.ID, a.ID, i.Reason); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_test.revised", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) ArchiveDiagnosticTest(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticArchiveInput) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	old, e := scanDiagnosticTest(tx.QueryRow(ctx, `SELECT `+diagnosticTestFields+` FROM diagnostic_test WHERE id=$1 FOR UPDATE`, id))
	if e != nil {
		return e
	}
	if !old.Active || old.Version != i.Version {
		return domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `UPDATE diagnostic_test SET active=false,version=version+1 WHERE id=$1`, id); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_catalog_event(test_id,actor_id,reason) VALUES($1,$2,$3)`, id, a.ID, i.Reason); e != nil {
		return e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_test.archived", id); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s Store) DiagnosticRevisions(ctx context.Context, id string, page int) ([]domain.DiagnosticTest, error) {
	var root string
	if e := s.DB.QueryRow(ctx, `SELECT COALESCE(root_id,id)::text FROM diagnostic_test WHERE id=$1`, id).Scan(&root); e != nil {
		return nil, clinicalError(e)
	}
	rows, e := s.DB.Query(ctx, `SELECT `+diagnosticTestFields+` FROM diagnostic_test WHERE id=$1 OR root_id=$1 ORDER BY revision DESC LIMIT 25 OFFSET $2`, root, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.DiagnosticTest{}
	for rows.Next() {
		v, err := scanDiagnosticTest(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	for n := range out {
		out[n].Parameters, e = diagnosticParameters(ctx, s.DB, out[n].ID)
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
