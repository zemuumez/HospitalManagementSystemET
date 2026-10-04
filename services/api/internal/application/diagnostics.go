package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
)

type DiagnosticsRepository interface {
	ReviseDiagnosticTest(context.Context, domain.Actor, string, domain.DiagnosticRevisionInput) (domain.DiagnosticTest, error)
	ArchiveDiagnosticTest(context.Context, domain.Actor, string, domain.DiagnosticArchiveInput) error
	DiagnosticRevisions(context.Context, string, int) ([]domain.DiagnosticTest, error)
	DiagnosticTests(context.Context, string, int) ([]domain.DiagnosticTest, error)
	CreateDiagnosticTest(context.Context, domain.Actor, domain.DiagnosticTestInput) (domain.DiagnosticTest, error)
	DiagnosticOrders(context.Context, domain.Actor, string, int) ([]domain.DiagnosticOrder, error)
	CreateDiagnosticOrder(context.Context, domain.Actor, domain.DiagnosticOrderInput, string) (domain.DiagnosticOrder, error)
	DiagnosticTransition(context.Context, domain.Actor, string, domain.DiagnosticAction) (domain.DiagnosticOrder, error)
	SubmitDiagnosticResult(context.Context, domain.Actor, string, domain.DiagnosticResultInput) (domain.DiagnosticResult, error)
	DiagnosticResults(context.Context, domain.Actor, string, int) ([]domain.DiagnosticResult, error)
}
type Diagnostics struct{ Store DiagnosticsRepository }

func (d Diagnostics) Tests(ctx context.Context, a domain.Actor, search string, page int) ([]domain.DiagnosticTest, error) {
	if !a.Can("diagnostics.catalog") {
		return nil, domain.ErrForbidden
	}
	if len(search) > 100 || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return d.Store.DiagnosticTests(ctx, search, page)
}
func (d Diagnostics) CreateTest(ctx context.Context, a domain.Actor, i domain.DiagnosticTestInput) (domain.DiagnosticTest, error) {
	if a.Role != "admin" && a.Role != "lab_technician" {
		return domain.DiagnosticTest{}, domain.ErrForbidden
	}
	i.Parameters = append([]domain.DiagnosticParameter(nil), i.Parameters...)
	if e := i.Validate(); e != nil {
		return domain.DiagnosticTest{}, e
	}
	return d.Store.CreateDiagnosticTest(ctx, a, i)
}
func (d Diagnostics) Orders(ctx context.Context, a domain.Actor, encounter string, page int) ([]domain.DiagnosticOrder, error) {
	if !a.Can("diagnostics.read") {
		return nil, domain.ErrForbidden
	}
	if (encounter != "" && !domain.UUIDPattern.MatchString(encounter)) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return d.Store.DiagnosticOrders(ctx, a, encounter, page)
}
func (d Diagnostics) Order(ctx context.Context, a domain.Actor, i domain.DiagnosticOrderInput, key string) (domain.DiagnosticOrder, error) {
	if a.Role != "doctor" {
		return domain.DiagnosticOrder{}, domain.ErrForbidden
	}
	if !keyOK(key) {
		return domain.DiagnosticOrder{}, domain.ErrValidation
	}
	if e := i.Validate(); e != nil {
		return domain.DiagnosticOrder{}, e
	}
	return d.Store.CreateDiagnosticOrder(ctx, a, i, key)
}
func (d Diagnostics) Transition(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticAction) (domain.DiagnosticOrder, error) {
	if a.Role != "doctor" && a.Role != "lab_technician" {
		return domain.DiagnosticOrder{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.DiagnosticOrder{}, domain.ErrValidation
	}
	if e := i.Validate(); e != nil {
		return domain.DiagnosticOrder{}, e
	}
	if (i.Action == "collect" || i.Action == "process") != (a.Role == "lab_technician") {
		return domain.DiagnosticOrder{}, domain.ErrForbidden
	}
	return d.Store.DiagnosticTransition(ctx, a, id, i)
}
func (d Diagnostics) Submit(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticResultInput) (domain.DiagnosticResult, error) {
	if a.Role != "lab_technician" {
		return domain.DiagnosticResult{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.DiagnosticResult{}, domain.ErrValidation
	}
	i.Values = append([]domain.DiagnosticValue(nil), i.Values...)
	if e := i.Validate(); e != nil {
		return domain.DiagnosticResult{}, e
	}
	return d.Store.SubmitDiagnosticResult(ctx, a, id, i)
}
func (d Diagnostics) Results(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DiagnosticResult, error) {
	if !a.Can("diagnostics.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return d.Store.DiagnosticResults(ctx, a, id, page)
}

func (d Diagnostics) ReviseTest(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticRevisionInput) (domain.DiagnosticTest, error) {
	if a.Role != "admin" && a.Role != "lab_technician" {
		return domain.DiagnosticTest{}, domain.ErrForbidden
	}
	i.Reason = strings.TrimSpace(i.Reason)
	if !domain.UUIDPattern.MatchString(id) || strings.ContainsRune(i.Reason, 0) || i.Version < 1 || i.Version > 1000000000 || len(i.Reason) < 1 || len([]rune(i.Reason)) > 1000 {
		return domain.DiagnosticTest{}, domain.ErrValidation
	}
	i.Parameters = append([]domain.DiagnosticParameter(nil), i.Parameters...)
	if e := i.DiagnosticTestInput.Validate(); e != nil {
		return domain.DiagnosticTest{}, e
	}
	return d.Store.ReviseDiagnosticTest(ctx, a, id, i)
}
func (d Diagnostics) ArchiveTest(ctx context.Context, a domain.Actor, id string, i domain.DiagnosticArchiveInput) error {
	if a.Role != "admin" && a.Role != "lab_technician" {
		return domain.ErrForbidden
	}
	i.Reason = strings.TrimSpace(i.Reason)
	if !domain.UUIDPattern.MatchString(id) || strings.ContainsRune(i.Reason, 0) || i.Version < 1 || i.Version > 1000000000 || len(i.Reason) < 1 || len([]rune(i.Reason)) > 1000 {
		return domain.ErrValidation
	}
	return d.Store.ArchiveDiagnosticTest(ctx, a, id, i)
}
func (d Diagnostics) TestRevisions(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DiagnosticTest, error) {
	if !a.Can("diagnostics.catalog") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return d.Store.DiagnosticRevisions(ctx, id, page)
}
