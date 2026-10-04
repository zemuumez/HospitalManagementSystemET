package domain

import (
	"testing"
	"time"
)

func TestPatientValidation(t *testing.T) {
	now := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	good := PatientInput{GivenName: " Jane ", FamilyName: "Doe", DateOfBirth: "2000-01-01", Phone: "+254700000001"}
	if good.Validate(now) != nil || good.GivenName != "Jane" {
		t.Fatal("valid record rejected")
	}
	for _, dob := range []string{"2027-01-01", "1800-01-01", "2000-99-99"} {
		p := good
		p.DateOfBirth = dob
		if p.Validate(now) == nil {
			t.Fatalf("accepted invalid DOB %s", dob)
		}
	}
	p := good
	p.Phone = "0700000001"
	if p.Validate(now) == nil {
		t.Fatal("accepted non-E164 phone")
	}
}
func TestPermissionsDenyByDefault(t *testing.T) {
	for _, role := range []string{"patient", "doctor", "nurse", "accountant", "pharmacist", "lab_technician", "case_manager", "unknown", ""} {
		a := Actor{Role: role}
		if a.Can("patients.create") || a.Can("messages.manage") || a.Can("arbitrary") {
			t.Fatalf("overprivileged %s", role)
		}
	}
	if !(Actor{Role: "receptionist"}).Can("patients.create") {
		t.Fatal("receptionist must register patients")
	}
}
func TestMessagesRejectHeaderInjectionAndInvalidNumbers(t *testing.T) {
	for _, m := range []MessageInput{{Channel: "email", Recipient: "a@example.com", Subject: "hello\r\nBcc: x@example.com", Body: "test"}, {Channel: "sms", Recipient: "0700", Body: "test"}, {Channel: "other", Recipient: "a@example.com", Body: "test"}} {
		if m.Validate() == nil {
			t.Fatal("unsafe message accepted")
		}
	}
}

func TestPatientUnknownDOBAndHospitalDateBoundary(t *testing.T) {
	now := time.Date(2026, 10, 4, 22, 30, 0, 0, time.UTC) // October 5 in Addis Ababa.
	for _, dob := range []string{"", "2026-10-05", "1876-10-05"} {
		p := PatientInput{GivenName: "Test", FamilyName: "Patient", DateOfBirth: dob}
		if e := p.Validate(now); e != nil {
			t.Fatal(dob, e)
		}
	}
	for _, dob := range []string{"2026-10-06", "1876-10-04", "2025-02-29"} {
		p := PatientInput{GivenName: "Test", FamilyName: "Patient", DateOfBirth: dob}
		if p.Validate(now) == nil {
			t.Fatal("invalid date", dob)
		}
	}
}
