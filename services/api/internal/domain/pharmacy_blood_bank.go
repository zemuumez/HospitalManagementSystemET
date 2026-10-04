package domain

import (
	"strings"
	"time"
)

// Allowed blood groups
var AllowedBloodGroups = map[string]bool{
	"A+":  true,
	"A-":  true,
	"B+":  true,
	"B-":  true,
	"AB+": true,
	"AB-": true,
	"O+":  true,
	"O-":  true,
}

// --- Medicine Category & Brand Masters ---

type MedicineCategory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MedicineCategoryInput struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (i *MedicineCategoryInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	if i.Name == "" || len(i.Name) > 160 {
		return ErrValidation
	}
	return nil
}

type MedicineBrand struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MedicineBrandInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (i *MedicineBrandInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Email = strings.TrimSpace(i.Email)
	i.Phone = strings.TrimSpace(i.Phone)
	if i.Name == "" || len(i.Name) > 160 {
		return ErrValidation
	}
	return nil
}

// --- Blood Bank Models ---

type BloodBankItem struct {
	ID           string    `json:"id"`
	BloodGroup   string    `json:"blood_group"`
	RemainedBags int       `json:"remained_bags"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BloodDonor struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Age            int       `json:"age"`
	Gender         int       `json:"gender"` // 0: male, 1: female
	BloodGroup     string    `json:"blood_group"`
	LastDonateDate time.Time `json:"last_donate_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BloodDonorInput struct {
	Name           string     `json:"name"`
	Age            int        `json:"age"`
	Gender         int        `json:"gender"`
	BloodGroup     string     `json:"blood_group"`
	LastDonateDate *time.Time `json:"last_donate_date"`
}

func (i *BloodDonorInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.BloodGroup = strings.ToUpper(strings.TrimSpace(i.BloodGroup))
	if i.Name == "" || len(i.Name) > 191 {
		return ErrValidation
	}
	if i.Age < 16 || i.Age > 80 {
		return ErrValidation
	}
	if i.Gender != 0 && i.Gender != 1 {
		return ErrValidation
	}
	if !AllowedBloodGroups[i.BloodGroup] {
		return ErrValidation
	}
	if i.LastDonateDate == nil || i.LastDonateDate.IsZero() {
		return ErrValidation
	}
	return nil
}

type BloodDonation struct {
	ID           string    `json:"id"`
	DonorID      string    `json:"donor_id"`
	DonorName    string    `json:"donor_name,omitempty"`
	BloodGroup   string    `json:"blood_group,omitempty"`
	Bags         int       `json:"bags"`
	DonationDate time.Time `json:"donation_date"`
	RecordedBy   string    `json:"recorded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type BloodDonationInput struct {
	DonorID      string     `json:"donor_id"`
	Bags         int        `json:"bags"`
	DonationDate *time.Time `json:"donation_date"`
}

func (i *BloodDonationInput) Validate() error {
	i.DonorID = strings.TrimSpace(i.DonorID)
	if i.DonorID == "" {
		return ErrValidation
	}
	if i.Bags <= 0 {
		i.Bags = 1
	}
	return nil
}

type BloodIssue struct {
	ID            string    `json:"id"`
	IssueDate     time.Time `json:"issue_date"`
	DoctorID      string    `json:"doctor_id"`
	DonorID       *string   `json:"donor_id,omitempty"`
	PatientID     string    `json:"patient_id"`
	PatientUserID *string   `json:"patient_user_id,omitempty"`
	BloodGroup    string    `json:"blood_group"`
	Bags          int       `json:"bags"`
	AmountMinor   int64     `json:"amount_minor"`
	Remarks       string    `json:"remarks"`
	IssuedBy      string    `json:"issued_by"`
	InvoiceID     *string   `json:"invoice_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BloodIssueInput struct {
	IssueDate   *time.Time `json:"issue_date"`
	DoctorID    string     `json:"doctor_id"`
	DonorID     *string    `json:"donor_id,omitempty"`
	PatientID   string     `json:"patient_id"`
	BloodGroup  string     `json:"blood_group"`
	Bags        int        `json:"bags"`
	AmountMinor int64      `json:"amount_minor"`
	Remarks     string     `json:"remarks"`
}

func (i *BloodIssueInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	i.PatientID = strings.TrimSpace(i.PatientID)
	i.BloodGroup = strings.ToUpper(strings.TrimSpace(i.BloodGroup))
	i.Remarks = strings.TrimSpace(i.Remarks)
	if i.DoctorID == "" || i.PatientID == "" {
		return ErrValidation
	}
	if !AllowedBloodGroups[i.BloodGroup] {
		return ErrValidation
	}
	if i.Bags <= 0 {
		i.Bags = 1
	}
	if i.AmountMinor < 0 {
		return ErrValidation
	}
	if i.IssueDate == nil || i.IssueDate.IsZero() {
		return ErrValidation
	}
	return nil
}

// --- Prescriptions ---

type PrescriptionMedicine struct {
	ID             string `json:"id"`
	PrescriptionID string `json:"prescription_id"`
	MedicineID     *string `json:"medicine_id,omitempty"`
	MedicineName   string `json:"medicine_name"`
	Dosage         string `json:"dosage"`
	Day            string `json:"day"`
	Time           string `json:"time"`
	Comment        string `json:"comment"`
}

type PrescriptionMedicineInput struct {
	MedicineID   *string `json:"medicine_id,omitempty"`
	MedicineName string  `json:"medicine_name"`
	Dosage       string  `json:"dosage"`
	Day          string  `json:"day"`
	Time         string  `json:"time"`
	Comment      string  `json:"comment"`
}

func (i *PrescriptionMedicineInput) Validate() error {
	i.MedicineName = strings.TrimSpace(i.MedicineName)
	i.Dosage = strings.TrimSpace(i.Dosage)
	if i.MedicineName == "" {
		return ErrValidation
	}
	return nil
}

type Prescription struct {
	ID                 string                 `json:"id"`
	PatientID          string                 `json:"patient_id"`
	PatientUserID      *string                `json:"patient_user_id,omitempty"`
	DoctorID           string                 `json:"doctor_id"`
	EncounterID        *string                `json:"encounter_id,omitempty"`
	FoodAllergies      string                 `json:"food_allergies"`
	TendencyBleed      string                 `json:"tendency_bleed"`
	HeartDisease       string                 `json:"heart_disease"`
	HighBloodPressure  string                 `json:"high_blood_pressure"`
	Diabetic           string                 `json:"diabetic"`
	Surgery            string                 `json:"surgery"`
	Accident           string                 `json:"accident"`
	Others             string                 `json:"others"`
	MedicalHistory     string                 `json:"medical_history"`
	CurrentMedication  string                 `json:"current_medication"`
	FemalePregnancy    string                 `json:"female_pregnancy"`
	BreastFeeding      string                 `json:"breast_feeding"`
	HealthInsurance    string                 `json:"health_insurance"`
	LowIncome          string                 `json:"low_income"`
	Reference          string                 `json:"reference"`
	Status             int                    `json:"status"` // 0: pending, 1: dispensed
	PlusRate           string                 `json:"plus_rate"`
	Temperature        string                 `json:"temperature"`
	ProblemDescription string                 `json:"problem_description"`
	Test               string                 `json:"test"`
	Advice             string                 `json:"advice"`
	NextVisitQty       string                 `json:"next_visit_qty"`
	NextVisitTime      string                 `json:"next_visit_time"`
	Medicines          []PrescriptionMedicine `json:"medicines,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type PrescriptionInput struct {
	PatientID          string                      `json:"patient_id"`
	DoctorID           string                      `json:"doctor_id"`
	EncounterID        *string                     `json:"encounter_id,omitempty"`
	FoodAllergies      string                      `json:"food_allergies"`
	TendencyBleed      string                      `json:"tendency_bleed"`
	HeartDisease       string                      `json:"heart_disease"`
	HighBloodPressure  string                      `json:"high_blood_pressure"`
	Diabetic           string                      `json:"diabetic"`
	Surgery            string                      `json:"surgery"`
	Accident           string                      `json:"accident"`
	Others             string                      `json:"others"`
	MedicalHistory     string                      `json:"medical_history"`
	CurrentMedication  string                      `json:"current_medication"`
	FemalePregnancy    string                      `json:"female_pregnancy"`
	BreastFeeding      string                      `json:"breast_feeding"`
	HealthInsurance    string                      `json:"health_insurance"`
	LowIncome          string                      `json:"low_income"`
	Reference          string                      `json:"reference"`
	PlusRate           string                      `json:"plus_rate"`
	Temperature        string                      `json:"temperature"`
	ProblemDescription string                      `json:"problem_description"`
	Test               string                      `json:"test"`
	Advice             string                      `json:"advice"`
	NextVisitQty       string                      `json:"next_visit_qty"`
	NextVisitTime      string                      `json:"next_visit_time"`
	Medicines          []PrescriptionMedicineInput `json:"medicines"`
}

func (i *PrescriptionInput) Validate() error {
	i.PatientID = strings.TrimSpace(i.PatientID)
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	if i.PatientID == "" || i.DoctorID == "" {
		return ErrValidation
	}
	if len(i.Medicines) == 0 {
		return ErrValidation
	}
	for idx := range i.Medicines {
		if err := i.Medicines[idx].Validate(); err != nil {
			return err
		}
	}
	return nil
}
