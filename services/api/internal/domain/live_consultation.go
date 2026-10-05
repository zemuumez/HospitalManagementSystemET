package domain

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

func randomCode(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

type LiveConsultation struct {
	ID                  string    `json:"id"`
	DoctorID            string    `json:"doctor_id"`
	PatientID           string    `json:"patient_id"`
	PatientUserID       *string   `json:"patient_user_id,omitempty"`
	EncounterID         *string   `json:"encounter_id,omitempty"`
	ConsultationTitle   string    `json:"consultation_title"`
	ConsultationDate    time.Time `json:"consultation_date"`
	DurationMinutes     int       `json:"duration_minutes"`
	HostVideo           bool      `json:"host_video"`
	ParticipantVideo    bool      `json:"participant_video"`
	Type                string    `json:"type"` // "OPD", "IPD"
	TypeNumber          string    `json:"type_number"`
	PlatformType        string    `json:"platform_type"` // "zoom", "meet", "in_house"
	MeetingID           string    `json:"meeting_id"`
	Password            string    `json:"password"`
	TimeZone            string    `json:"time_zone"`
	Status              int       `json:"status"` // 0: Awaited, 1: Finished, 2: Cancelled
	Description         string    `json:"description"`
	CreatedBy           string    `json:"created_by"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type LiveConsultationInput struct {
	DoctorID          string     `json:"doctor_id"`
	PatientID         string     `json:"patient_id"`
	EncounterID       *string    `json:"encounter_id,omitempty"`
	ConsultationTitle string     `json:"consultation_title"`
	ConsultationDate  *time.Time `json:"consultation_date"`
	DurationMinutes   int        `json:"duration_minutes"`
	HostVideo         bool       `json:"host_video"`
	ParticipantVideo  bool       `json:"participant_video"`
	Type              string     `json:"type"`
	TypeNumber        string     `json:"type_number"`
	PlatformType      string     `json:"platform_type"`
	MeetingID         string     `json:"meeting_id"`
	Password          string     `json:"password"`
	Description       string     `json:"description"`
}

func (i *LiveConsultationInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	i.PatientID = strings.TrimSpace(i.PatientID)
	i.ConsultationTitle = strings.TrimSpace(i.ConsultationTitle)
	i.PlatformType = strings.ToLower(strings.TrimSpace(i.PlatformType))
	i.MeetingID = strings.TrimSpace(i.MeetingID)
	i.Type = strings.ToUpper(strings.TrimSpace(i.Type))

	if i.DoctorID == "" || i.PatientID == "" {
		return ErrValidation
	}
	if i.ConsultationTitle == "" || len(i.ConsultationTitle) > 191 {
		return ErrValidation
	}
	if i.ConsultationDate == nil || i.ConsultationDate.IsZero() {
		return ErrValidation
	}
	if i.DurationMinutes <= 0 {
		i.DurationMinutes = 30
	}
	if i.PlatformType == "" {
		i.PlatformType = "zoom"
	}
	switch i.PlatformType {
	case "zoom", "meet", "in_house":
	default:
		return ErrValidation
	}
	if i.MeetingID == "" {
		// Provide default synthesized meeting ID if none supplied
		i.MeetingID = "HMS-" + randomCode(8)
	}
	if i.Password == "" {
		i.Password = randomCode(6)
	}
	if i.Type == "" {
		i.Type = "OPD"
	}
	switch i.Type {
	case "OPD", "IPD":
	default:
		return ErrValidation
	}
	return nil
}

type LiveMeeting struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	MeetingDate      time.Time `json:"meeting_date"`
	DurationMinutes  int       `json:"duration_minutes"`
	HostVideo        bool      `json:"host_video"`
	ParticipantVideo bool      `json:"participant_video"`
	PlatformType     string    `json:"platform_type"`
	MeetingID        string    `json:"meeting_id"`
	Password         string    `json:"password"`
	TimeZone         string    `json:"time_zone"`
	Status           int       `json:"status"` // 0: Awaited, 1: Finished, 2: Cancelled
	Description      string    `json:"description"`
	CreatedBy        string    `json:"created_by"`
	CandidateUserIDs []string  `json:"candidate_user_ids,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type LiveMeetingInput struct {
	Title            string     `json:"title"`
	MeetingDate      *time.Time `json:"meeting_date"`
	DurationMinutes  int        `json:"duration_minutes"`
	HostVideo        bool       `json:"host_video"`
	ParticipantVideo bool       `json:"participant_video"`
	PlatformType     string     `json:"platform_type"`
	MeetingID        string     `json:"meeting_id"`
	Password         string     `json:"password"`
	Description      string     `json:"description"`
	CandidateUserIDs []string   `json:"candidate_user_ids"`
}

func (i *LiveMeetingInput) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.PlatformType = strings.ToLower(strings.TrimSpace(i.PlatformType))
	i.MeetingID = strings.TrimSpace(i.MeetingID)
	if i.Title == "" || len(i.Title) > 191 {
		return ErrValidation
	}
	if i.MeetingDate == nil || i.MeetingDate.IsZero() {
		return ErrValidation
	}
	if i.DurationMinutes <= 0 {
		i.DurationMinutes = 45
	}
	if i.PlatformType == "" {
		i.PlatformType = "zoom"
	}
	switch i.PlatformType {
	case "zoom", "meet", "in_house":
	default:
		return ErrValidation
	}
	if i.MeetingID == "" {
		i.MeetingID = "MTG-" + randomCode(8)
	}
	if i.Password == "" {
		i.Password = randomCode(6)
	}
	return nil
}

type LiveProviderSetting struct {
	UserID       string    `json:"user_id"`
	PlatformType string    `json:"platform_type"`
	HasAPIKey    bool      `json:"has_api_key"`
	HasAPISecret bool      `json:"has_api_secret"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LiveProviderSettingInput struct {
	PlatformType string `json:"platform_type"`
	APIKey       string `json:"api_key"`
	APISecret    string `json:"api_secret"`
}

func (i *LiveProviderSettingInput) Validate() error {
	i.PlatformType = strings.ToLower(strings.TrimSpace(i.PlatformType))
	if i.PlatformType == "" {
		i.PlatformType = "zoom"
	}
	switch i.PlatformType {
	case "zoom", "meet", "in_house":
	default:
		return ErrValidation
	}
	return nil
}
