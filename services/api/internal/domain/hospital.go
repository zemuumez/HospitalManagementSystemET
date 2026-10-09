package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var ErrForbidden = errors.New("forbidden")
var ErrValidation = errors.New("validation failed")
var ErrConflict = errors.New("idempotency key conflict")
var ErrInUse = errors.New("record is in use and cannot be deleted")
var ErrExportLimitExceeded = errors.New("export exceeds maximum limit of 5000 records; please apply search filters")

type Actor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func (a Actor) Can(permission string) bool {
	switch permission {
	case "diagnostics.catalog":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "lab_technician"
	case "diagnostics.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "lab_technician" || a.Role == "patient"

	case "pharmacy.catalog":
		return a.Role == "admin" || a.Role == "pharmacist" || a.Role == "doctor"
	case "pharmacy.manage":
		return a.Role == "admin" || a.Role == "pharmacist"
	case "medication.read":
		return a.Role == "admin" || a.Role == "pharmacist" || a.Role == "doctor" || a.Role == "patient"

	case "billing.read":
		return a.Role == "admin" || a.Role == "accountant" || a.Role == "patient"
	case "billing.manage":
		return a.Role == "admin" || a.Role == "accountant"
	case "finance.read", "finance.manage":
		return a.Role == "admin" || a.Role == "accountant"
	case "payroll.read", "payroll.manage":
		return a.Role == "admin" || a.Role == "accountant"
	case "payroll.read_own":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "receptionist" || a.Role == "pharmacist" || a.Role == "accountant" || a.Role == "case_manager" || a.Role == "lab_technician"
	case "beds.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse"
	case "clinical.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "patient"
	case "clinical.admit":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor"

	case "attendance.clock", "attendance.read_own":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "receptionist" || a.Role == "pharmacist" || a.Role == "accountant" || a.Role == "case_manager" || a.Role == "lab_technician"
	case "attendance.manage":
		return a.Role == "admin"
	case "ambulance.manage", "ambulance_call.manage":
		return a.Role == "admin" || a.Role == "case_manager" || a.Role == "receptionist"
	case "ambulance.read":
		return a.Role == "admin" || a.Role == "case_manager" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse"
	case "ambulance_call.read":
		return a.Role == "admin" || a.Role == "case_manager" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "accountant" || a.Role == "patient"
	case "packages.manage":
		return a.Role == "admin" || a.Role == "receptionist"
	case "packages.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "case_manager" || a.Role == "patient"
	case "insurances.manage":
		return a.Role == "admin" || a.Role == "receptionist"
	case "insurances.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "case_manager" || a.Role == "patient"
	case "services.manage":
		return a.Role == "admin" || a.Role == "accountant"
	case "services.read":
		return a.Role == "admin" || a.Role == "accountant" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "receptionist"
	case "operations.manage":
		return a.Role == "admin" || a.Role == "doctor"
	case "operations.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "receptionist"
	case "settings.manage", "settings.read":
		return a.Role == "admin"
	case "cms.manage":
		return a.Role == "admin"
	case "cms.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "pharmacist" || a.Role == "accountant" || a.Role == "case_manager" || a.Role == "lab_technician" || a.Role == "patient"
	case "front_office.manage":
		return a.Role == "admin" || a.Role == "receptionist"
	case "front_office.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "case_manager" || a.Role == "accountant"
	case "complaints.manage":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "case_manager"
	case "complaints.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "case_manager" || a.Role == "patient"
	case "complaints.create":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "patient"
	case "notices.manage":
		return a.Role == "admin"
	case "notices.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "pharmacist" || a.Role == "accountant" || a.Role == "case_manager" || a.Role == "lab_technician" || a.Role == "patient"
	case "live_consultations.manage":
		return a.Role == "admin" || a.Role == "doctor"
	case "live_consultations.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "patient"
	case "live_meetings.manage":
		return a.Role == "admin" || a.Role == "doctor"
	case "live_meetings.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "pharmacist" || a.Role == "receptionist" || a.Role == "accountant" || a.Role == "case_manager" || a.Role == "lab_technician"
	case "blood_bank.manage":
		return a.Role == "admin" || a.Role == "lab_technician" || a.Role == "doctor"
	case "blood_bank.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "lab_technician" || a.Role == "pharmacist" || a.Role == "receptionist" || a.Role == "patient"
	case "prescriptions.manage":
		return a.Role == "admin" || a.Role == "doctor"
	case "prescriptions.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "pharmacist" || a.Role == "patient"
	case "appointments.read", "appointments.book":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "patient"
	case "staff.manage":
		return a.Role == "admin"
	case "patients.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "patient"
	case "patients.create":
		return a.Role == "admin" || a.Role == "receptionist"
	case "messages.manage":
		return a.Role == "admin" || a.Role == "receptionist"
	case "enquiries.manage":
		return a.Role == "admin" || a.Role == "receptionist"
	case "enquiries.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse" || a.Role == "case_manager"
	case "scheduling.manage":
		return a.Role == "admin" || a.Role == "doctor"
	case "doctor_dept.manage":
		return a.Role == "admin"
	}
	return false
}
func (a Actor) Permissions() []string {
	result := []string{}
	for _, p := range []string{
		"diagnostics.catalog", "diagnostics.read", "pharmacy.catalog", "pharmacy.manage", "medication.read",
		"patients.read", "patients.create", "messages.manage", "appointments.read", "appointments.book",
		"staff.manage", "beds.read", "clinical.read", "clinical.admit", "billing.read", "billing.manage",
		"attendance.clock", "attendance.read_own", "attendance.manage", "ambulance.manage", "ambulance.read",
		"ambulance_call.manage", "ambulance_call.read", "packages.manage", "packages.read", "insurances.manage", "insurances.read", "services.manage", "services.read", "operations.manage",
		"operations.read", "settings.manage", "settings.read", "cms.read", "cms.manage", "enquiries.read", "enquiries.manage",
		"complaints.read", "complaints.manage", "complaints.create", "notices.read", "notices.manage",
		"front_office.manage", "front_office.read", "live_consultations.manage", "live_consultations.read",
		"live_meetings.manage", "live_meetings.read", "blood_bank.manage", "blood_bank.read",
		"prescriptions.manage", "prescriptions.read", "finance.read", "finance.manage",
		"payroll.read", "payroll.read_own", "payroll.manage",
		"scheduling.manage", "doctor_dept.manage",
	} {
		if a.Can(p) {
			result = append(result, p)
		}
	}
	return result
}

type PatientInput struct {
	GivenName   string `json:"givenName"`
	FamilyName  string `json:"familyName"`
	DateOfBirth string `json:"dateOfBirth"`
	Phone       string `json:"phone"`
}

var PhonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func (p *PatientInput) Validate(now time.Time) error {
	p.GivenName = strings.TrimSpace(p.GivenName)
	p.FamilyName = strings.TrimSpace(p.FamilyName)
	p.Phone = strings.TrimSpace(p.Phone)
	p.DateOfBirth = strings.TrimSpace(p.DateOfBirth)
	if p.DateOfBirth != "" {
		local := now.In(HospitalLocation)
		today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, HospitalLocation)
		dob, err := time.ParseInLocation("2006-01-02", p.DateOfBirth, HospitalLocation)
		if err != nil || dob.After(today) || dob.Before(today.AddDate(-150, 0, 0)) {
			return ErrValidation
		}
	}
	if len([]rune(p.GivenName)) < 1 || len([]rune(p.GivenName)) > 80 || len([]rune(p.FamilyName)) < 1 || len([]rune(p.FamilyName)) > 80 {
		return ErrValidation
	}
	if p.Phone != "" && !PhonePattern.MatchString(p.Phone) {
		return ErrValidation
	}
	return nil
}

type Patient struct {
	UserID      string `json:"userId"`
	ClinicianID string `json:"clinicianId"`
	ID          string `json:"id"`
	MRN         string `json:"mrn"`
	PatientInput
	Email      string    `json:"email,omitempty"`
	Gender     string    `json:"gender,omitempty"`
	BloodGroup string    `json:"bloodGroup,omitempty"`
	FatherName string    `json:"fatherName,omitempty"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
}
type PatientAccess struct {
	UserID      string `json:"userId"`
	ClinicianID string `json:"clinicianId"`
}
type Overview struct {
	PatientCount         int   `json:"patientCount"`
	RegisteredToday      int   `json:"registeredToday"`
	InvoicesMinor        int64 `json:"invoicesMinor"`
	BillsMinor           int64 `json:"billsMinor"`
	PaymentsMinor        int64 `json:"paymentsMinor"`
	AdvancePaymentsMinor int64 `json:"advancePaymentsMinor"`
	TotalBeds            int   `json:"totalBeds"`
	AvailableBeds        int   `json:"availableBeds"`
	OccupiedBeds         int   `json:"occupiedBeds"`
	Doctors              int   `json:"doctors"`
	Patients             int   `json:"patients"`
	Nurses               int   `json:"nurses"`
	Admins               int   `json:"admins"`
	Accountants          int   `json:"accountants"`
	LabTechnicians       int   `json:"labTechnicians"`
	Pharmacists          int   `json:"pharmacists"`
	Receptionists        int   `json:"receptionists"`
}
type MessageInput struct {
	Channel   string `json:"channel"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

func (m *MessageInput) Validate() error {
	m.Recipient = strings.TrimSpace(m.Recipient)
	m.Subject = strings.TrimSpace(m.Subject)
	m.Body = strings.TrimSpace(m.Body)
	if len(m.Body) == 0 || len(m.Body) > 4000 || len(m.Subject) > 200 || strings.ContainsAny(m.Subject, "\r\n") {
		return ErrValidation
	}
	switch m.Channel {
	case "sms":
		if !PhonePattern.MatchString(m.Recipient) || len([]rune(m.Body)) > 480 {
			return ErrValidation
		}
	case "email":
		a, err := mail.ParseAddress(m.Recipient)
		if err != nil || a.Address != m.Recipient || m.Subject == "" {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}

type Message struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}
