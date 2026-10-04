package domain

import (
	"strings"
	"time"
)

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
	if !i.IsClosed {
		if len(i.StartTime) < 4 || len(i.EndTime) < 4 {
			return ErrValidation
		}
		if i.StartTime >= i.EndTime {
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
