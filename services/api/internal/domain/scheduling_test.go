package domain

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"
)

func TestScheduleBoundaries(t *testing.T) {
	d := Doctor{ID: "doctor", Department: "General", SlotMinutes: 30, Hours: []DoctorHours{{Weekday: 1, StartMinute: 540, EndMinute: 720}, {Weekday: 1, StartMinute: 780, EndMinute: 1020}}}
	if d.Validate() != nil {
		t.Fatal("valid split day rejected")
	}
	for _, tc := range []struct {
		clock string
		want  bool
	}{{"09:00", true}, {"11:30", true}, {"11:45", false}, {"12:00", false}, {"13:00", true}, {"16:30", true}, {"17:00", false}, {"09:15", false}} {
		start, _ := time.ParseInLocation("2006-01-02 15:04", "2026-10-05 "+tc.clock, HospitalLocation)
		if d.Allows(start, start.Add(30*time.Minute)) != tc.want {
			t.Errorf("unexpected availability at %s", tc.clock)
		}
	}
	d.Hours = append(d.Hours, DoctorHours{Weekday: 1, StartMinute: 600, EndMinute: 800})
	if d.Validate() == nil {
		t.Fatal("overlap accepted")
	}
}
func TestAppointmentStateAndOwnershipPolicy(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	if CanTransition("patient", "booked", "completed", now, now) || CanTransition("nurse", "booked", "cancelled", future, now) || CanTransition("admin", "cancelled", "booked", future, now) {
		t.Fatal("invalid transition allowed")
	}
	if !CanTransition("patient", "booked", "cancelled", future, now) || !CanTransition("doctor", "arrived", "completed", now, now) {
		t.Fatal("valid transition denied")
	}
	if CanTransition("admin", "arrived", "completed", future, now) {
		t.Fatal("future completion allowed")
	}
	for _, role := range []string{"nurse", "pharmacist", "accountant", "lab_technician", "case_manager", "unknown"} {
		if (Actor{Role: role}).Can("appointments.book") {
			t.Fatal("unexpected booking permission", role)
		}
	}
}

func TestDoctorEmailValidation(t *testing.T) {
	var fixtureBytes []byte
	for _, p := range []string{
		"../../../../docs/module-audit/doctor-email-fixtures.json",
		"../../../../../docs/module-audit/doctor-email-fixtures.json",
		"docs/module-audit/doctor-email-fixtures.json",
	} {
		if data, readErr := os.ReadFile(p); readErr == nil {
			fixtureBytes = data
			break
		}
	}
	if len(fixtureBytes) == 0 {
		t.Fatalf("failed to read shared email fixtures from candidate paths")
	}
	var fixtures struct {
		Valid   []string `json:"valid"`
		Invalid []string `json:"invalid"`
	}
	if err := json.Unmarshal(fixtureBytes, &fixtures); err != nil {
		t.Fatalf("failed to parse fixtures: %v", err)
	}

	for _, v := range fixtures.Valid {
		email := v
		in := UpdateDoctorInput{
			Version: 1,
			Email:   &email,
		}
		if err := in.Validate(); err != nil {
			t.Errorf("expected valid email %q to pass validation, got %v", v, err)
		}
	}

	for _, inv := range fixtures.Invalid {
		email := inv
		in := UpdateDoctorInput{
			Version: 1,
			Email:   &email,
		}
		if err := in.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("expected invalid email %q to be rejected with ErrValidation, got %v", inv, err)
		}
	}
}

