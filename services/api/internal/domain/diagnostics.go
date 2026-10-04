package domain

import (
	"math"
	"sort"
	"strconv"
	"time"
)

type DiagnosticParameter struct {
	Position       int    `json:"position"`
	Name           string `json:"name"`
	Unit           string `json:"unit"`
	ReferenceRange string `json:"referenceRange"`
	ValueType      string `json:"valueType"`
}
type DiagnosticTestInput struct {
	Kind        string                `json:"kind"`
	Name        string                `json:"name"`
	ShortName   string                `json:"shortName"`
	Category    string                `json:"category"`
	Method      string                `json:"method"`
	ReportDays  int                   `json:"reportDays"`
	ChargeMinor int64                 `json:"chargeMinor"`
	Parameters  []DiagnosticParameter `json:"parameters"`
}
type DiagnosticTest struct {
	Active     bool   `json:"active"`
	Version    int    `json:"version"`
	Revision   int    `json:"revision"`
	RootID     string `json:"rootId"`
	Supersedes string `json:"supersedes"`
	DiagnosticTestInput
	ID string `json:"id"`
}

func (i *DiagnosticTestInput) Validate() error {
	if (i.Kind != "pathology" && i.Kind != "radiology") || !validText(&i.Name, 1, 150) || !validText(&i.ShortName, 1, 50) || !validText(&i.Category, 1, 100) || !validText(&i.Method, 0, 200) || i.ReportDays < 0 || i.ReportDays > 365 || i.ChargeMinor < 0 || i.ChargeMinor > 1000000000 || len(i.Parameters) < 1 || len(i.Parameters) > 50 {
		return ErrValidation
	}
	names := map[string]bool{}
	for n := range i.Parameters {
		p := &i.Parameters[n]
		p.Position = n + 1
		if !validText(&p.Name, 1, 150) || !validText(&p.Unit, 0, 40) || !validText(&p.ReferenceRange, 0, 500) || (p.ValueType != "number" && p.ValueType != "text") || names[p.Name] {
			return ErrValidation
		}
		names[p.Name] = true
	}
	return nil
}

type DiagnosticOrderInput struct {
	EncounterID string `json:"encounterId"`
	TestID      string `json:"testId"`
	Indication  string `json:"indication"`
}

func (i *DiagnosticOrderInput) Validate() error {
	if !UUIDPattern.MatchString(i.EncounterID) || !UUIDPattern.MatchString(i.TestID) || !validText(&i.Indication, 1, 2000) {
		return ErrValidation
	}
	return nil
}

type DiagnosticOrder struct {
	DiagnosticOrderInput
	ID              string    `json:"id"`
	Accession       int64     `json:"accession"`
	Status          string    `json:"status"`
	SampleReference string    `json:"sampleReference"`
	Version         int       `json:"version"`
	TestName        string    `json:"testName"`
	Kind            string    `json:"kind"`
	CreatedAt       time.Time `json:"createdAt"`
}
type DiagnosticAction struct {
	Version         int    `json:"version"`
	Action          string `json:"action"`
	SampleReference string `json:"sampleReference"`
	Reason          string `json:"reason"`
}

func (i *DiagnosticAction) Validate() error {
	if i.Version < 1 || !validText(&i.SampleReference, 0, 100) || !validText(&i.Reason, 0, 1000) {
		return ErrValidation
	}
	switch i.Action {
	case "collect":
		if i.SampleReference == "" {
			return ErrValidation
		}
	case "process", "sign", "release":
	case "cancel", "reject":
		if i.Reason == "" {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	if i.Action != "collect" && i.SampleReference != "" {
		return ErrValidation
	}
	return nil
}

type DiagnosticValue struct {
	Position int    `json:"position"`
	Value    string `json:"value"`
}
type DiagnosticResultInput struct {
	Version         int               `json:"version"`
	Summary         string            `json:"summary"`
	AmendmentReason string            `json:"amendmentReason"`
	Values          []DiagnosticValue `json:"values"`
}

func (i *DiagnosticResultInput) Validate() error {
	if i.Version < 1 || !validText(&i.Summary, 1, 5000) || !validText(&i.AmendmentReason, 0, 1000) || len(i.Values) < 1 || len(i.Values) > 50 {
		return ErrValidation
	}
	sort.Slice(i.Values, func(a, b int) bool { return i.Values[a].Position < i.Values[b].Position })
	for n := range i.Values {
		v := &i.Values[n]
		if v.Position != n+1 || !validText(&v.Value, 1, 1000) {
			return ErrValidation
		}
	}
	return nil
}
func ValidateDiagnosticNumber(value string) bool {
	v, e := strconv.ParseFloat(value, 64)
	return e == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
}

type DiagnosticResult struct {
	Parameters      []DiagnosticParameter `json:"parameters"`
	ID              string                `json:"id"`
	OrderID         string                `json:"orderId"`
	Revision        int                   `json:"revision"`
	Summary         string                `json:"summary"`
	AmendmentReason string                `json:"amendmentReason"`
	Values          []DiagnosticValue     `json:"values"`
	Signed          bool                  `json:"signed"`
	Released        bool                  `json:"released"`
	CreatedAt       time.Time             `json:"createdAt"`
}

type DiagnosticRevisionInput struct {
	DiagnosticTestInput
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
type DiagnosticArchiveInput struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
