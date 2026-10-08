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
// EmailPattern defines the practical email acceptance contract matching the installed Next.js validator.
var EmailPattern = regexp.MustCompile(`^(?:[a-zA-Z0-9_'+\-]+\.)*[a-zA-Z0-9_'+\-]*[a-zA-Z0-9_+-]@(?:[a-zA-Z0-9][a-zA-Z0-9\-]*\.)+[a-zA-Z]{2,}$`)
var HospitalLocation, _ = time.LoadLocation("Africa/Addis_Ababa")

type DoctorHours struct {
	Weekday     int `json:"weekday"`
	StartMinute int `json:"startMinute"`
	EndMinute   int `json:"endMinute"`
}
type Doctor struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Email             string        `json:"email,omitempty"`
	Phone             string        `json:"phone,omitempty"`
	Department        string        `json:"department"`
	DepartmentID      *string       `json:"departmentId,omitempty"`
	Specialist        string        `json:"specialist,omitempty"`
	Designation       string        `json:"designation,omitempty"`
	Qualification     string        `json:"qualification,omitempty"`
	Gender            string        `json:"gender,omitempty"`
	DateOfBirth       string        `json:"dateOfBirth,omitempty"`
	BloodGroup        string        `json:"bloodGroup,omitempty"`
	Address1          string        `json:"address1,omitempty"`
	Address2          string        `json:"address2,omitempty"`
	City              string        `json:"city,omitempty"`
	Zip               string        `json:"zip,omitempty"`
	Description       string        `json:"description,omitempty"`
	PhotoURL          string        `json:"photoUrl,omitempty"`
	OpdCharge         float64       `json:"opdCharge,omitempty"`
	AppointmentCharge float64       `json:"appointmentCharge,omitempty"`
	SlotMinutes       int           `json:"slotMinutes"`
	Active            bool          `json:"active"`
	Version           int           `json:"version"`
	Hours             []DoctorHours `json:"hours"`
}

func (d Doctor) PublicDirectory(actorRole string) Doctor {
	if actorRole == "admin" {
		return d
	}
	sanitized := d
	sanitized.DateOfBirth = ""
	sanitized.BloodGroup = ""
	sanitized.Address1 = ""
	sanitized.Address2 = ""
	sanitized.City = ""
	sanitized.Zip = ""
	if actorRole == "patient" {
		sanitized.Phone = ""
	}
	return sanitized
}

type CreateDoctorInput struct {
	UserID            string        `json:"userId"`
	DepartmentID      string        `json:"departmentId"`
	Specialist        string        `json:"specialist"`
	Designation       string        `json:"designation,omitempty"`
	Qualification     string        `json:"qualification,omitempty"`
	Phone             string        `json:"phone,omitempty"`
	Gender            string        `json:"gender,omitempty"`
	DateOfBirth       string        `json:"dateOfBirth,omitempty"`
	BloodGroup        string        `json:"bloodGroup,omitempty"`
	Address1          string        `json:"address1,omitempty"`
	Address2          string        `json:"address2,omitempty"`
	City              string        `json:"city,omitempty"`
	Zip               string        `json:"zip,omitempty"`
	Description       string        `json:"description,omitempty"`
	PhotoURL          string        `json:"photoUrl,omitempty"`
	AppointmentCharge float64       `json:"appointmentCharge,omitempty"`
	OpdCharge         float64       `json:"opdCharge,omitempty"`
	SlotMinutes       int           `json:"slotMinutes,omitempty"`
	Hours             []DoctorHours `json:"hours,omitempty"`
}

func (i *CreateDoctorInput) Validate() error {
	i.UserID = strings.TrimSpace(i.UserID)
	i.Specialist = strings.TrimSpace(i.Specialist)
	i.DepartmentID = strings.TrimSpace(i.DepartmentID)
	i.Designation = strings.TrimSpace(i.Designation)
	i.Qualification = strings.TrimSpace(i.Qualification)
	i.Gender = strings.TrimSpace(strings.ToLower(i.Gender))

	if i.UserID == "" || len(i.UserID) > 128 {
		return ErrValidation
	}
	if i.DepartmentID == "" || !UUIDPattern.MatchString(i.DepartmentID) {
		return ErrValidation
	}
	if i.Specialist == "" || len([]rune(i.Specialist)) > 191 {
		return ErrValidation
	}
	// D3: Designation and Qualification are required by original rules
	if i.Designation == "" || len([]rune(i.Designation)) > 191 {
		return ErrValidation
	}
	if i.Qualification == "" || len([]rune(i.Qualification)) > 191 {
		return ErrValidation
	}
	// D3: Gender is required by original rules
	if i.Gender == "" || (i.Gender != "male" && i.Gender != "female" && i.Gender != "other" && i.Gender != "0" && i.Gender != "1") {
		return ErrValidation
	}
	if len([]rune(i.Description)) > 2000 {
		return ErrValidation
	}
	if i.AppointmentCharge < 0 || i.OpdCharge < 0 {
		return ErrValidation
	}
	// D2: Default SlotMinutes to 60 (01:00:00 per_patient_time in DoctorRepository::store)
	if i.SlotMinutes == 0 {
		i.SlotMinutes = 60
	}
	if i.SlotMinutes < 5 || i.SlotMinutes > 120 {
		return ErrValidation
	}
	if len(i.Hours) > 21 {
		return ErrValidation
	}
	for idx, h := range i.Hours {
		if h.Weekday < 0 || h.Weekday > 6 || h.StartMinute < 0 || h.EndMinute > 1440 || h.EndMinute-h.StartMinute < i.SlotMinutes {
			return ErrValidation
		}
		for _, other := range i.Hours[:idx] {
			if h.Weekday == other.Weekday && h.StartMinute < other.EndMinute && other.StartMinute < h.EndMinute {
				return ErrValidation
			}
		}
	}
	if i.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", i.DateOfBirth)
		if err != nil || dob.Year() < 1850 || dob.After(time.Now()) {
			return ErrValidation
		}
	}
	return nil
}

type UpdateDoctorInput struct {
	Name              *string       `json:"name,omitempty"`
	Email             *string       `json:"email,omitempty"`
	DepartmentID      *string       `json:"departmentId,omitempty"`
	Specialist        *string       `json:"specialist,omitempty"`
	Designation       *string       `json:"designation,omitempty"`
	Qualification     *string       `json:"qualification,omitempty"`
	Phone             *string       `json:"phone,omitempty"`
	Gender            *string       `json:"gender,omitempty"`
	DateOfBirth       *string       `json:"dateOfBirth,omitempty"`
	BloodGroup        *string       `json:"bloodGroup,omitempty"`
	Address1          *string       `json:"address1,omitempty"`
	Address2          *string       `json:"address2,omitempty"`
	City              *string       `json:"city,omitempty"`
	Zip               *string       `json:"zip,omitempty"`
	Description       *string       `json:"description,omitempty"`
	PhotoURL          *string       `json:"photoUrl,omitempty"`
	AppointmentCharge *float64      `json:"appointmentCharge,omitempty"`
	OpdCharge         *float64      `json:"opdCharge,omitempty"`
	SlotMinutes       *int          `json:"slotMinutes,omitempty"`
	Hours             []DoctorHours `json:"hours,omitempty"`
	Version           int           `json:"version"`
}

func (i *UpdateDoctorInput) Validate() error {
	if i.Version <= 0 {
		return ErrValidation
	}
	if i.Name != nil {
		*i.Name = strings.TrimSpace(*i.Name)
		if *i.Name == "" || len([]rune(*i.Name)) > 120 {
			return ErrValidation
		}
	}
	if i.Email != nil {
		email := strings.ToLower(*i.Email)
		if len(email) > 254 || !EmailPattern.MatchString(email) {
			return ErrValidation
		}
		*i.Email = email
	}
	if i.DepartmentID != nil {
		*i.DepartmentID = strings.TrimSpace(*i.DepartmentID)
		if *i.DepartmentID == "" || !UUIDPattern.MatchString(*i.DepartmentID) {
			return ErrValidation
		}
	}
	if i.Specialist != nil {
		*i.Specialist = strings.TrimSpace(*i.Specialist)
		if *i.Specialist == "" || len([]rune(*i.Specialist)) > 191 {
			return ErrValidation
		}
	}
	if i.Designation != nil {
		*i.Designation = strings.TrimSpace(*i.Designation)
		if *i.Designation == "" || len([]rune(*i.Designation)) > 191 {
			return ErrValidation
		}
	}
	if i.Qualification != nil {
		*i.Qualification = strings.TrimSpace(*i.Qualification)
		if *i.Qualification == "" || len([]rune(*i.Qualification)) > 191 {
			return ErrValidation
		}
	}
	if i.Gender != nil {
		*i.Gender = strings.TrimSpace(strings.ToLower(*i.Gender))
		if *i.Gender == "" || (*i.Gender != "male" && *i.Gender != "female" && *i.Gender != "other" && *i.Gender != "0" && *i.Gender != "1") {
			return ErrValidation
		}
	}
	if i.Description != nil && len([]rune(*i.Description)) > 2000 {
		return ErrValidation
	}
	if i.AppointmentCharge != nil && *i.AppointmentCharge < 0 {
		return ErrValidation
	}
	if i.OpdCharge != nil && *i.OpdCharge < 0 {
		return ErrValidation
	}
	if i.SlotMinutes != nil && (*i.SlotMinutes < 5 || *i.SlotMinutes > 120) {
		return ErrValidation
	}
	if len(i.Hours) > 21 {
		return ErrValidation
	}
	slot := 60
	if i.SlotMinutes != nil {
		slot = *i.SlotMinutes
	}
	for idx, h := range i.Hours {
		if h.Weekday < 0 || h.Weekday > 6 || h.StartMinute < 0 || h.EndMinute > 1440 || h.EndMinute-h.StartMinute < slot {
			return ErrValidation
		}
		for _, other := range i.Hours[:idx] {
			if h.Weekday == other.Weekday && h.StartMinute < other.EndMinute && other.StartMinute < h.EndMinute {
				return ErrValidation
			}
		}
	}
	if i.DateOfBirth != nil && *i.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", *i.DateOfBirth)
		if err != nil || dob.Year() < 1850 || dob.After(time.Now()) {
			return ErrValidation
		}
	}
	return nil
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
	Reason  string `json:"reason"`
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

type AppointmentStatusEvent struct {
	ID             string    `json:"id"`
	PreviousStatus string    `json:"previousStatus"`
	NextStatus     string    `json:"nextStatus"`
	ActorID        string    `json:"actorId"`
	Reason         string    `json:"reason"`
	Version        int       `json:"version"`
	RecordedAt     time.Time `json:"recordedAt"`
}

type ScheduleDayRow struct {
	Day  string `json:"day"`
	From string `json:"from"`
	To   string `json:"to"`
}

type DoctorSchedule struct {
	ID             string           `json:"id"`
	DoctorID       string           `json:"doctorId"`
	DoctorName     string           `json:"doctorName"`
	PerPatientTime string           `json:"perPatientTime"`
	SlotMinutes    int              `json:"slotMinutes"`
	Days           []ScheduleDayRow `json:"days"`
}

type SaveDoctorScheduleInput struct {
	DoctorID       string           `json:"doctorId"`
	PerPatientTime string           `json:"perPatientTime,omitempty"`
	SlotMinutes    int              `json:"slotMinutes,omitempty"`
	Days           []ScheduleDayRow `json:"days"`
}

func ParseTimeToMinutes(s string) (int, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, errors.New("invalid time format")
	}
	var h, m int
	for idx, part := range parts[:2] {
		val := 0
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return 0, errors.New("invalid digit")
			}
			val = val*10 + int(ch-'0')
		}
		if idx == 0 {
			if val < 0 || val > 23 {
				return 0, errors.New("invalid hour")
			}
			h = val
		} else {
			if val < 0 || val > 59 {
				return 0, errors.New("invalid minute")
			}
			m = val
		}
	}
	return h*60 + m, nil
}

func FormatMinutesToTime(m int) string {
	h := (m / 60) % 24
	min := m % 60
	return time.Date(0, 1, 1, h, min, 0, 0, time.UTC).Format("15:04:05")
}

var WeekdayNames = [7]string{
	"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
}

func WeekdayNameToIndex(name string) int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "sunday":
		return 0
	case "monday":
		return 1
	case "tuesday":
		return 2
	case "wednesday":
		return 3
	case "thursday":
		return 4
	case "friday":
		return 5
	case "saturday":
		return 6
	default:
		return -1
	}
}

func (i *SaveDoctorScheduleInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	if i.DoctorID == "" || len(i.DoctorID) > 128 {
		return ErrValidation
	}
	if i.SlotMinutes == 0 && i.PerPatientTime != "" {
		m, err := ParseTimeToMinutes(i.PerPatientTime)
		if err == nil && m > 0 {
			i.SlotMinutes = m
		}
	}
	if i.SlotMinutes == 0 {
		i.SlotMinutes = 60
	}
	if i.SlotMinutes < 5 || i.SlotMinutes > 120 {
		return ErrValidation
	}
	if len(i.Days) == 0 {
		return ErrValidation
	}
	for _, d := range i.Days {
		w := WeekdayNameToIndex(d.Day)
		if w < 0 {
			return ErrValidation
		}
		fromMin, err1 := ParseTimeToMinutes(d.From)
		toMin, err2 := ParseTimeToMinutes(d.To)
		if err1 != nil || err2 != nil {
			return ErrValidation
		}
		if fromMin > toMin {
			return ErrValidation
		}
		if fromMin != toMin && toMin-fromMin < i.SlotMinutes {
			return ErrValidation
		}
	}
	return nil
}

type DoctorHoliday struct {
	ID         string    `json:"id"`
	DoctorID   string    `json:"doctorId"`
	DoctorName string    `json:"doctorName"`
	Date       string    `json:"date"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CreateDoctorHolidayInput struct {
	DoctorID string `json:"doctorId"`
	Date     string `json:"date"`
	Reason   string `json:"reason"`
}

func (i *CreateDoctorHolidayInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	i.Date = strings.TrimSpace(i.Date)
	i.Reason = strings.TrimSpace(i.Reason)
	if i.DoctorID == "" || len(i.DoctorID) > 128 {
		return ErrValidation
	}
	if i.Date == "" {
		return ErrValidation
	}
	_, err := time.Parse("2006-01-02", i.Date)
	if err != nil {
		return ErrValidation
	}
	if len([]rune(i.Reason)) > 200 {
		return ErrValidation
	}
	return nil
}

type DoctorLunchBreak struct {
	ID          string `json:"id"`
	DoctorID    string `json:"doctorId"`
	DoctorName  string `json:"doctorName"`
	DoctorEmail string `json:"doctorEmail"`
	BreakFrom   string `json:"breakFrom"`
	BreakTo     string `json:"breakTo"`
	EveryDay    bool   `json:"everyDay"`
	Date        string `json:"date,omitempty"`
	DateType    string `json:"dateType"`
}

type CreateDoctorBreakInput struct {
	DoctorID  string `json:"doctorId"`
	BreakFrom string `json:"breakFrom"`
	BreakTo   string `json:"breakTo"`
	EveryDay  bool   `json:"everyDay"`
	Date      string `json:"date,omitempty"`
}

func (i *CreateDoctorBreakInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	if i.DoctorID == "" || len(i.DoctorID) > 128 {
		return ErrValidation
	}
	fromMin, err1 := ParseTimeToMinutes(i.BreakFrom)
	toMin, err2 := ParseTimeToMinutes(i.BreakTo)
	if err1 != nil || err2 != nil || toMin <= fromMin {
		return ErrValidation
	}
	if !i.EveryDay {
		i.Date = strings.TrimSpace(i.Date)
		if i.Date == "" {
			return ErrValidation
		}
		_, err := time.Parse("2006-01-02", i.Date)
		if err != nil {
			return ErrValidation
		}
	} else {
		i.Date = ""
	}
	return nil
}

