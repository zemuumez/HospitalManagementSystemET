package domain

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var SettingsSecretKeys = map[string]bool{
	"open_ai_key":            true,
	"stripe_secret":          true,
	"paypal_secret":          true,
	"razorpay_secret":        true,
	"flutterwave_secret_key": true,
	"phonepe_salt_key":       true,
	"paystack_secret_key":    true,
}

var GeneralSettingRequiredKeys = map[string]bool{
	"app_name":            true,
	"company_name":        true,
	"hospital_name":       true,
	"hospital_email":      true,
	"hospital_phone":      true,
	"hospital_from_day":   true,
	"hospital_start_day":  true,
	"hospital_from_time":  true,
	"hospital_start_time": true,
	"hospital_address":    true,
	"current_currency":    true,
	"currency":            true,
	"about_us":            true,
}

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	timeHHMMRegex = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)
	currencyRegex = regexp.MustCompile(`^[A-Za-z]{3}$`)
)

const SecretConfiguredPlaceholder = "[CONFIGURED]"

type HospitalGeneralSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GeneralSettingInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (i *GeneralSettingInput) Validate() error {
	i.Key = strings.TrimSpace(i.Key)
	i.Value = strings.TrimSpace(i.Value)
	if i.Key == "" || len(i.Key) > 128 {
		return ErrValidation
	}
	// Direct requests attempting to blank out or set whitespace for required fields must be rejected
	if GeneralSettingRequiredKeys[i.Key] && i.Value == "" {
		return ErrValidation
	}
	switch i.Key {
	case "hospital_email":
		if i.Value != "" && !emailRegex.MatchString(i.Value) {
			return ErrValidation
		}
	case "current_currency", "currency":
		if i.Value != "" && !currencyRegex.MatchString(i.Value) {
			return ErrValidation
		}
	case "hospital_from_time", "hospital_start_time":
		if i.Value != "" && !timeHHMMRegex.MatchString(i.Value) {
			return ErrValidation
		}
	case "hospital_from_day", "hospital_start_day":
		if i.Value != "" {
			dayNum, err := strconv.Atoi(i.Value)
			if err == nil {
				if dayNum < 1 || dayNum > 7 {
					return ErrValidation
				}
			} else {
				validDays := map[string]bool{
					"monday": true, "tuesday": true, "wednesday": true,
					"thursday": true, "friday": true, "saturday": true, "sunday": true,
				}
				if !validDays[strings.ToLower(i.Value)] {
					return ErrValidation
				}
			}
		}
	case "hospital_phone":
		if i.Value != "" && len(i.Value) < 7 {
			return ErrValidation
		}
	}
	return nil
}

type HospitalScheduleDay struct {
	ID        string    `json:"id"`
	DayOfWeek int       `json:"day_of_week"` // 1 (Mon) - 7 (Sun)
	StartTime string    `json:"start_time"`  // "HH:MM"
	EndTime   string    `json:"end_time"`    // "HH:MM"
	IsClosed  bool      `json:"is_closed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HospitalScheduleDayInput struct {
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	IsClosed  bool   `json:"is_closed"`
}

func (i *HospitalScheduleDayInput) Validate() error {
	if i.DayOfWeek < 1 || i.DayOfWeek > 7 {
		return ErrValidation
	}
	i.StartTime = strings.TrimSpace(i.StartTime)
	i.EndTime = strings.TrimSpace(i.EndTime)
	if i.IsClosed {
		if i.StartTime != "" && !timeHHMMRegex.MatchString(i.StartTime) {
			return ErrValidation
		}
		if i.EndTime != "" && !timeHHMMRegex.MatchString(i.EndTime) {
			return ErrValidation
		}
	} else {
		if !timeHHMMRegex.MatchString(i.StartTime) || !timeHHMMRegex.MatchString(i.EndTime) {
			return ErrValidation
		}
		st, err1 := time.Parse("15:04", i.StartTime)
		et, err2 := time.Parse("15:04", i.EndTime)
		if err1 != nil || err2 != nil || !st.Before(et) {
			return ErrValidation
		}
	}
	return nil
}

type FrontCMSSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Type      string    `json:"type"` // "home", "about", "services", "contact", "terms", "privacy", "general"
	UpdatedAt time.Time `json:"updated_at"`
}

type FrontCMSSettingInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

func (i *FrontCMSSettingInput) Validate() error {
	i.Key = strings.TrimSpace(i.Key)
	i.Value = strings.TrimSpace(i.Value)
	i.Type = strings.TrimSpace(i.Type)
	if i.Key == "" || len(i.Key) > 128 {
		return ErrValidation
	}
	if i.Type == "" {
		i.Type = "general"
	}
	return nil
}

type CMSTestimonial struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Position    string    `json:"position"`
	Rating      int       `json:"rating"` // 1 - 5
	Status      int       `json:"status"` // 1: Published, 0: Draft
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CMSTestimonialInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    string `json:"position"`
	Rating      int    `json:"rating"`
	Status      int    `json:"status"`
}

func (i *CMSTestimonialInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	i.Position = strings.TrimSpace(i.Position)
	if i.Name == "" || len(i.Name) > 191 {
		return ErrValidation
	}
	if i.Description == "" {
		return ErrValidation
	}
	if i.Position == "" {
		i.Position = "Patient"
	}
	if i.Rating < 1 || i.Rating > 5 {
		i.Rating = 5
	}
	if i.Status != 0 && i.Status != 1 {
		i.Status = 1
	}
	return nil
}
