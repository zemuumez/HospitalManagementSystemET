package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const diagnosticTestFields = `id,kind,name,short_name,category,method,report_days,charge_minor,active,version,revision,COALESCE(root_id::text,id::text),COALESCE(supersedes::text,'')`

func scanDiagnosticTest(row pgx.Row) (domain.DiagnosticTest, error) {
	var t domain.DiagnosticTest
	t.Parameters = []domain.DiagnosticParameter{}
	e := row.Scan(&t.ID, &t.Kind, &t.Name, &t.ShortName, &t.Category, &t.Method, &t.ReportDays, &t.ChargeMinor, &t.Active, &t.Version, &t.Revision, &t.RootID, &t.Supersedes)
	return t, clinicalError(e)
}
func diagnosticParameters(ctx context.Context, q querier, id string) ([]domain.DiagnosticParameter, error) {
	rows, e := q.Query(ctx, `SELECT position,name,unit,reference_range,value_type FROM diagnostic_parameter WHERE test_id=$1 ORDER BY position`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.DiagnosticParameter{}
	for rows.Next() {
		var p domain.DiagnosticParameter
		if e = rows.Scan(&p.Position, &p.Name, &p.Unit, &p.ReferenceRange, &p.ValueType); e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s Store) DiagnosticTests(ctx context.Context, search string, page int) ([]domain.DiagnosticTest, error) {
	rows, e := s.DB.Query(ctx, `SELECT `+diagnosticTestFields+` FROM diagnostic_test WHERE active AND ($1='' OR strpos(lower(name),lower($1))>0) ORDER BY kind,name,id LIMIT 25 OFFSET $2`, search, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.DiagnosticTest{}
	for rows.Next() {
		t, e := scanDiagnosticTest(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, t)
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
func (s Store) CreateDiagnosticTest(ctx context.Context, a domain.Actor, i domain.DiagnosticTestInput) (domain.DiagnosticTest, error) {
	out := domain.DiagnosticTest{DiagnosticTestInput: i, Active: true, Version: 1, Revision: 1}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	e = tx.QueryRow(ctx, `INSERT INTO diagnostic_test(kind,name,short_name,category,method,report_days,charge_minor,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, i.Kind, i.Name, i.ShortName, i.Category, i.Method, i.ReportDays, i.ChargeMinor, a.ID).Scan(&out.ID)
	if e != nil {
		return out, clinicalError(e)
	}
	out.RootID = out.ID
	for _, p := range i.Parameters {
		if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_parameter(test_id,position,name,unit,reference_range,value_type) VALUES($1,$2,$3,$4,$5,$6)`, out.ID, p.Position, p.Name, p.Unit, p.ReferenceRange, p.ValueType); e != nil {
			return out, e
		}
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_test.created", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}

const diagnosticOrderFields = `o.id,o.accession,o.encounter_id,o.test_id,o.indication,o.status,o.sample_reference,o.version,t.name,t.kind,o.created_at`
const diagnosticOrderFrom = ` FROM diagnostic_order o JOIN diagnostic_test t ON t.id=o.test_id JOIN encounter e ON e.id=o.encounter_id JOIN patient p ON p.id=e.patient_id `
const diagnosticScope = `($1 IN ('admin','lab_technician') OR ($1='doctor' AND e.doctor_id=$2) OR ($1='patient' AND p.user_id=$2 AND EXISTS(SELECT 1 FROM diagnostic_result r JOIN diagnostic_review v ON v.result_id=r.id AND v.action='release' WHERE r.order_id=o.id)))`

func scanDiagnosticOrder(row pgx.Row) (domain.DiagnosticOrder, error) {
	var o domain.DiagnosticOrder
	e := row.Scan(&o.ID, &o.Accession, &o.EncounterID, &o.TestID, &o.Indication, &o.Status, &o.SampleReference, &o.Version, &o.TestName, &o.Kind, &o.CreatedAt)
	return o, clinicalError(e)
}
func (s Store) DiagnosticOrders(ctx context.Context, a domain.Actor, encounter string, page int) ([]domain.DiagnosticOrder, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT `+diagnosticOrderFields+diagnosticOrderFrom+`WHERE `+diagnosticScope+` AND ($3='' OR o.encounter_id=NULLIF($3,'')::uuid) ORDER BY o.created_at DESC,o.id LIMIT 25 OFFSET $4`, a.Role, a.ID, encounter, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.DiagnosticOrder{}
	for rows.Next() {
		o, e := scanDiagnosticOrder(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		if a.Role == "patient" {
			o.Status = "released"
			o.Version = 0
		}
		out = append(out, o)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_orders.viewed", encounter); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) CreateDiagnosticOrder(ctx context.Context, a domain.Actor, i domain.DiagnosticOrderInput, key string) (domain.DiagnosticOrder, error) {
	out := domain.DiagnosticOrder{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "diagnostic-order"); e != nil {
		return out, e
	}
	var old, hash string
	wanted := requestHash(i)
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM diagnostic_order WHERE ordered_by=$1 AND request_key=$2`, a.ID, key).Scan(&old, &hash)
	if e == nil {
		if hash != wanted {
			return out, domain.ErrConflict
		}
		return scanDiagnosticOrder(tx.QueryRow(ctx, `SELECT `+diagnosticOrderFields+diagnosticOrderFrom+`WHERE o.id=$1`, old))
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	var status string
	e = tx.QueryRow(ctx, `SELECT status FROM encounter WHERE id=$1 AND doctor_id=$2 FOR UPDATE`, i.EncounterID, a.ID).Scan(&status)
	if e != nil {
		return out, clinicalError(e)
	}
	if status != "active" {
		return out, domain.ErrStale
	}
	var test string
	e = tx.QueryRow(ctx, `SELECT id FROM diagnostic_test WHERE id=$1 AND active FOR SHARE`, i.TestID).Scan(&test)
	if e != nil {
		return out, clinicalError(e)
	}
	var id string
	e = tx.QueryRow(ctx, `INSERT INTO diagnostic_order(encounter_id,test_id,ordered_by,indication,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, i.EncounterID, i.TestID, a.ID, i.Indication, key, wanted).Scan(&id)
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_order.created", id); e != nil {
		return out, e
	}
	out, e = scanDiagnosticOrder(tx.QueryRow(ctx, `SELECT `+diagnosticOrderFields+diagnosticOrderFrom+`WHERE o.id=$1`, id))
	if e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func diagnosticEvent(ctx context.Context, tx pgx.Tx, a domain.Actor, o domain.DiagnosticOrder, target, reason string) error {
	if _, e := tx.Exec(ctx, `UPDATE diagnostic_order SET status=$2,version=version+1 WHERE id=$1`, o.ID, target); e != nil {
		return e
	}
	_, e := tx.Exec(ctx, `INSERT INTO diagnostic_event(order_id,from_status,to_status,reason,actor_id,version) VALUES($1,$2,$3,$4,$5,$6)`, o.ID, o.Status, target, reason, a.ID, o.Version+1)
	return e
}
func (s Store) DiagnosticTransition(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticAction) (domain.DiagnosticOrder, error) {
	out := domain.DiagnosticOrder{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	o, e := scanDiagnosticOrder(tx.QueryRow(ctx, `SELECT `+diagnosticOrderFields+diagnosticOrderFrom+`WHERE o.id=$3 AND `+diagnosticScope+` FOR UPDATE OF o`, a.Role, a.ID, id))
	if e != nil {
		return out, e
	}
	if o.Version != i.Version {
		return out, domain.ErrStale
	}
	target := ""
	switch i.Action {
	case "collect":
		if o.Status == "ordered" {
			target = "collected"
		}
	case "process":
		if o.Status == "collected" {
			target = "processing"
		}
	case "sign":
		if o.Status == "review" {
			target = "signed"
		}
	case "release":
		if o.Status == "signed" {
			target = "released"
		}
	case "reject":
		if o.Status == "review" {
			target = "processing"
		}
	case "cancel":
		if o.Status == "ordered" || o.Status == "collected" || o.Status == "processing" || o.Status == "review" {
			target = "cancelled"
		}
	}
	if target == "" {
		return out, domain.ErrStale
	}
	if i.Action == "collect" {
		if _, e = tx.Exec(ctx, `UPDATE diagnostic_order SET sample_reference=$2 WHERE id=$1`, id, i.SampleReference); e != nil {
			return out, clinicalError(e)
		}
	}
	if i.Action == "sign" || i.Action == "release" || i.Action == "reject" {
		var result string
		e = tx.QueryRow(ctx, `SELECT id FROM diagnostic_result WHERE order_id=$1 ORDER BY revision DESC LIMIT 1`, id).Scan(&result)
		if e != nil {
			return out, clinicalError(e)
		}
		if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_review(result_id,action,actor_id,reason) VALUES($1,$2,$3,$4)`, result, i.Action, a.ID, i.Reason); e != nil {
			return out, clinicalError(e)
		}
	}
	if e = diagnosticEvent(ctx, tx, a, o, target, i.Reason); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_order."+i.Action, id); e != nil {
		return out, e
	}
	out, e = scanDiagnosticOrder(tx.QueryRow(ctx, `SELECT `+diagnosticOrderFields+diagnosticOrderFrom+`WHERE o.id=$1`, id))
	if e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}

const diagnosticResultFields = `r.id,r.order_id,r.revision,r.summary,r.amendment_reason,r.created_at,EXISTS(SELECT 1 FROM diagnostic_review v WHERE v.result_id=r.id AND v.action='sign'),EXISTS(SELECT 1 FROM diagnostic_review v WHERE v.result_id=r.id AND v.action='release')`

func scanDiagnosticResult(row pgx.Row) (domain.DiagnosticResult, error) {
	var r domain.DiagnosticResult
	r.Values = []domain.DiagnosticValue{}
	e := row.Scan(&r.ID, &r.OrderID, &r.Revision, &r.Summary, &r.AmendmentReason, &r.CreatedAt, &r.Signed, &r.Released)
	return r, clinicalError(e)
}
func diagnosticValues(ctx context.Context, q querier, id string) ([]domain.DiagnosticValue, error) {
	rows, e := q.Query(ctx, `SELECT position,value FROM diagnostic_value WHERE result_id=$1 ORDER BY position`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.DiagnosticValue{}
	for rows.Next() {
		var v domain.DiagnosticValue
		if e = rows.Scan(&v.Position, &v.Value); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s Store) SubmitDiagnosticResult(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticResultInput) (domain.DiagnosticResult, error) {
	out := domain.DiagnosticResult{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	o, e := scanDiagnosticOrder(tx.QueryRow(ctx, `SELECT `+diagnosticOrderFields+diagnosticOrderFrom+`WHERE o.id=$1 FOR UPDATE OF o`, id))
	if e != nil {
		return out, e
	}
	if o.Version != i.Version || (o.Status != "processing" && o.Status != "released") {
		return out, domain.ErrStale
	}
	params, e := diagnosticParameters(ctx, tx, o.TestID)
	if e != nil {
		return out, e
	}
	if len(params) != len(i.Values) {
		return out, domain.ErrValidation
	}
	for n, p := range params {
		if p.Position != i.Values[n].Position || (p.ValueType == "number" && !domain.ValidateDiagnosticNumber(i.Values[n].Value)) {
			return out, domain.ErrValidation
		}
	}
	var revision int
	e = tx.QueryRow(ctx, `SELECT COALESCE(max(revision),0)+1 FROM diagnostic_result WHERE order_id=$1`, id).Scan(&revision)
	if e != nil {
		return out, e
	}
	if revision > 1 && i.AmendmentReason == "" {
		return out, domain.ErrValidation
	}
	out, e = scanDiagnosticResult(tx.QueryRow(ctx, `INSERT INTO diagnostic_result AS r(order_id,revision,summary,amendment_reason,created_by) VALUES($1,$2,$3,$4,$5) RETURNING `+diagnosticResultFields, id, revision, i.Summary, i.AmendmentReason, a.ID))
	if e != nil {
		return out, e
	}
	for _, v := range i.Values {
		if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_value(result_id,position,value) VALUES($1,$2,$3)`, out.ID, v.Position, v.Value); e != nil {
			return out, e
		}
	}
	if _, e = tx.Exec(ctx, `UPDATE diagnostic_result SET sealed=true WHERE id=$1`, out.ID); e != nil {
		return out, e
	}
	out.Parameters = params
	out.Values = i.Values
	if e = diagnosticEvent(ctx, tx, a, o, "review", i.AmendmentReason); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_result.submitted", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) DiagnosticResults(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DiagnosticResult, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	var allowed bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 `+diagnosticOrderFrom+`WHERE o.id=$3 AND `+diagnosticScope+`)`, a.Role, a.ID, id).Scan(&allowed)
	if e != nil {
		return nil, e
	}
	if !allowed {
		return nil, domain.ErrNotFound
	}
	rows, e := tx.Query(ctx, `SELECT `+diagnosticResultFields+` FROM diagnostic_result r WHERE r.order_id=$1 AND ($2<>'patient' OR EXISTS(SELECT 1 FROM diagnostic_review v WHERE v.result_id=r.id AND v.action='release')) ORDER BY r.revision DESC LIMIT 25 OFFSET $3`, id, a.Role, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.DiagnosticResult{}
	for rows.Next() {
		r, e := scanDiagnosticResult(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, r)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	var testID string
	if e = tx.QueryRow(ctx, `SELECT test_id FROM diagnostic_order WHERE id=$1`, id).Scan(&testID); e != nil {
		return nil, e
	}
	parameters, e := diagnosticParameters(ctx, tx, testID)
	if e != nil {
		return nil, e
	}
	for n := range out {
		out[n].Parameters = parameters
		out[n].Values, e = diagnosticValues(ctx, tx, out[n].ID)
		if e != nil {
			return nil, e
		}
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic_results.viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
