package domain

import (
	"testing"
	"time"
)

func TestPatientProfileValidation(t *testing.T) {
	base := PatientProfileInput{PatientInput: PatientInput{GivenName: "Test", FamilyName: "Person", DateOfBirth: "2000-01-01"}, Gender: "unknown", Version: 1, Reason: "Correction"}
	if e := base.Validate(time.Now()); e != nil {
		t.Fatal(e)
	}
	for _, change := range []func(*PatientProfileInput){
		func(p *PatientProfileInput) { p.Email = "Name <test@example.test>" },
		func(p *PatientProfileInput) { p.Gender = "invalid" },
		func(p *PatientProfileInput) { p.BloodGroup = "invalid" },
		func(p *PatientProfileInput) { p.EmergencyPhone = "911" },
		func(p *PatientProfileInput) { p.Version = 0 },
		func(p *PatientProfileInput) { p.Reason = "   " },
	} {
		input := base
		change(&input)
		if input.Validate(time.Now()) == nil {
			t.Fatal("invalid profile accepted", input)
		}
	}
}
