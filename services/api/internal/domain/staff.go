package domain

import "time"

type StaffDetails struct {
	GivenName     string `json:"givenName"`
	FamilyName    string `json:"familyName"`
	Phone         string `json:"phone"`
	Gender        string `json:"gender"`
	DateOfBirth   string `json:"dateOfBirth"`
	BloodGroup    string `json:"bloodGroup,omitempty"`
	Designation   string `json:"designation"`
	Qualification string `json:"qualification"`
	Specialty     string `json:"specialty"`
	Address1      string `json:"address1"`
	Address2      string `json:"address2"`
	City          string `json:"city"`
	Region        string `json:"region"`
	Country       string `json:"country"`
	PostalCode    string `json:"postalCode"`
}
type StaffProfile struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Email   string       `json:"email"`
	Role    string       `json:"role"`
	Active  bool         `json:"active"`
	Version int          `json:"version"`
	Details StaffDetails `json:"details"`
}
type StaffProfileInput struct {
	Details StaffDetails `json:"details"`
	Version int          `json:"version"`
	Reason  string       `json:"reason"`
}
type StaffRoleInput struct {
	PreviousRole string `json:"previousRole"`
	Role         string `json:"role"`
	Reason       string `json:"reason"`
}

func (i *StaffProfileInput) Validate(now time.Time) error {
	d := &i.Details
	if !validText(&d.GivenName, 1, 100) || !validText(&d.FamilyName, 1, 100) || i.Version < 1 || i.Version > 1000000000 || !validText(&i.Reason, 1, 500) {
		return ErrValidation
	}
	for _, f := range []*string{&d.Designation, &d.Qualification, &d.Specialty, &d.Address1, &d.Address2, &d.City, &d.Region, &d.Country, &d.PostalCode} {
		if !validText(f, 0, 250) {
			return ErrValidation
		}
	}
	if d.Phone != "" && !PhonePattern.MatchString(d.Phone) {
		return ErrValidation
	}
	switch d.Gender {
	case "unknown", "female", "male", "other":
	default:
		return ErrValidation
	}
	if d.DateOfBirth != "" {
		dob, e := time.Parse("2006-01-02", d.DateOfBirth)
		if e != nil || dob.After(now) || dob.Year() < 1850 {
			return ErrValidation
		}
	}
	return nil
}
func StaffRole(role string) bool {
	switch role {
	case "admin", "doctor", "patient", "nurse", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician":
		return true
	}
	return false
}
func (i *StaffRoleInput) Validate() error {
	if !StaffRole(i.Role) || !StaffRole(i.PreviousRole) || i.Role == i.PreviousRole || !validText(&i.Reason, 1, 500) {
		return ErrValidation
	}
	return nil
}
