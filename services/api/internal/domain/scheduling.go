package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"
)

var ErrNotFound = errors.New("record not found")
var ErrStale = errors.New("record changed or time unavailable")
var UUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var HospitalLocation, _ = time.LoadLocation("Africa/Addis_Ababa")

type DoctorHours struct {
	Weekday     int `json:"weekday"`
	StartMinute int `json:"startMinute"`
	EndMinute   int `json:"endMinute"`
}
type Doctor struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Department  string        `json:"department"`
	SlotMinutes int           `json:"slotMinutes"`
	Version     int           `json:"version"`
	Hours       []DoctorHours `json:"hours"`
}

func (d *Doctor) Validate() error {
	d.Department = strings.TrimSpace(d.Department)
	if d.ID == "" || len(d.ID) > 128 || len([]rune(d.Department)) < 1 || len([]rune(d.Department)) > 100 || d.SlotMinutes < 5 || d.SlotMinutes > 120 || d.Version < 0 || len(d.Hours) > 21 {
		return ErrValidation
	}
	for i, h := range d.Hours {
		if h.Weekday < 0 || h.Weekday > 6 || h.StartMinute < 0 || h.EndMinute > 1440 || h.EndMinute-h.StartMinute < d.SlotMinutes {
			return ErrValidation
		}
		for _, other := range d.Hours[:i] {
			if h.Weekday == other.Weekday && h.StartMinute < other.EndMinute && other.StartMinute < h.EndMinute {
				return ErrValidation
			}
		}
	}
	return nil
}
func (d Doctor) Allows(start, end time.Time) bool {
	local := start.In(HospitalLocation)
	minute := local.Hour()*60 + local.Minute()
	if local.Second() != 0 || local.Nanosecond() != 0 || end.Sub(start) != time.Duration(d.SlotMinutes)*time.Minute {
		return false
	}
	for _, h := range d.Hours {
		if int(local.Weekday()) == h.Weekday && minute >= h.StartMinute && minute+d.SlotMinutes <= h.EndMinute && (minute-h.StartMinute)%d.SlotMinutes == 0 {
			return true
		}
	}
	return false
}

type AppointmentInput struct {
	PatientID string    `json:"patientId"`
	DoctorID  string    `json:"doctorId"`
	StartsAt  time.Time `json:"startsAt"`
	Problem   string    `json:"problem"`
	NotifySMS bool      `json:"notifySms"`
}

func (i *AppointmentInput) Validate(now time.Time) error {
	i.StartsAt = i.StartsAt.UTC().Truncate(time.Microsecond)
	i.Problem = strings.TrimSpace(i.Problem)
	if !UUIDPattern.MatchString(i.PatientID) || i.DoctorID == "" || len(i.DoctorID) > 128 || len([]rune(i.Problem)) > 2000 || !i.StartsAt.After(now) || i.StartsAt.After(now.AddDate(1, 0, 0)) {
		return ErrValidation
	}
	return nil
}

type Appointment struct {
	AppointmentInput
	ID          string    `json:"id"`
	PatientName string    `json:"patientName"`
	DoctorName  string    `json:"doctorName"`
	EndsAt      time.Time `json:"endsAt"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
}
type AppointmentChange struct {
	Status  string `json:"status"`
	Version int    `json:"version"`
}

func CanTransition(role, from, to string, start, now time.Time) bool {
	if from == "cancelled" || from == "completed" || from == "no_show" {
		return false
	}
	if role == "patient" {
		return from == "booked" && to == "cancelled" && start.After(now)
	}
	if role != "admin" && role != "receptionist" && role != "doctor" {
		return false
	}
	switch to {
	case "cancelled":
		return from == "booked" || from == "arrived"
	case "arrived":
		return from == "booked" && !start.After(now.Add(24*time.Hour))
	case "completed":
		return from == "arrived" && !start.After(now)
	case "no_show":
		return from == "booked" && !start.After(now)
	}
	return false
}

type AbsenceInput struct {
	DoctorID string    `json:"doctorId"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
	Reason   string    `json:"reason"`
}
type DoctorAbsence struct {
	AbsenceInput
	ID           string     `json:"id"`
	Version      int        `json:"version"`
	CancelledAt  *time.Time `json:"cancelledAt"`
	CancelReason string     `json:"cancelReason"`
}
type AbsenceCancel struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
type RescheduleInput struct {
	StartsAt time.Time `json:"startsAt"`
	Version  int       `json:"version"`
	Reason   string    `json:"reason"`
}
type RescheduleEvent struct {
	ID         string    `json:"id"`
	FromStart  time.Time `json:"fromStart"`
	FromEnd    time.Time `json:"fromEnd"`
	ToStart    time.Time `json:"toStart"`
	ToEnd      time.Time `json:"toEnd"`
	ActorID    string    `json:"actorId"`
	Reason     string    `json:"reason"`
	Version    int       `json:"version"`
	RecordedAt time.Time `json:"recordedAt"`
}
