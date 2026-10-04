package domain

import (
	"strings"
	"time"
)

type ChargeCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ChargeType  int       `json:"charge_type"` // 1: Investigation, 2: Operation, 3: Bed, 4: Doctor, 5: Other
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChargeCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ChargeType  int    `json:"charge_type"`
}

func (i *ChargeCategoryInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	if i.Name == "" || len(i.Name) > 160 {
		return ErrValidation
	}
	if i.ChargeType < 1 || i.ChargeType > 5 {
		return ErrValidation
	}
	return nil
}

type HospitalCharge struct {
	ID                  string    `json:"id"`
	ChargeType          int       `json:"charge_type"`
	ChargeCategoryID    string    `json:"charge_category_id"`
	ChargeCategoryName  string    `json:"charge_category_name,omitempty"`
	Code                string    `json:"code"`
	StandardChargeMinor int64     `json:"standard_charge_minor"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type HospitalChargeInput struct {
	ChargeType          int    `json:"charge_type"`
	ChargeCategoryID    string `json:"charge_category_id"`
	Code                string `json:"code"`
	StandardChargeMinor int64  `json:"standard_charge_minor"`
	Description         string `json:"description"`
}

func (i *HospitalChargeInput) Validate() error {
	i.ChargeCategoryID = strings.TrimSpace(i.ChargeCategoryID)
	i.Code = strings.ToUpper(strings.TrimSpace(i.Code))
	i.Description = strings.TrimSpace(i.Description)
	if !UUIDPattern.MatchString(i.ChargeCategoryID) {
		return ErrValidation
	}
	if i.Code == "" || len(i.Code) > 160 {
		return ErrValidation
	}
	if i.ChargeType < 1 || i.ChargeType > 5 {
		return ErrValidation
	}
	if i.StandardChargeMinor < 0 {
		return ErrValidation
	}
	return nil
}

type HospitalService struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	RateMinor   int64     `json:"rate_minor"`
	Status      int       `json:"status"` // 1: Active, 0: Inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HospitalServiceInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	RateMinor   int64  `json:"rate_minor"`
	Status      int    `json:"status"`
}

func (i *HospitalServiceInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	if i.Name == "" || len(i.Name) > 160 {
		return ErrValidation
	}
	if i.Quantity < 1 {
		i.Quantity = 1
	}
	if i.RateMinor < 0 {
		return ErrValidation
	}
	if i.Status != 0 && i.Status != 1 {
		i.Status = 1
	}
	return nil
}

type OperationCategory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OperationCategoryInput struct {
	Name string `json:"name"`
}

func (i *OperationCategoryInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	if i.Name == "" || len(i.Name) > 191 {
		return ErrValidation
	}
	return nil
}

type HospitalOperation struct {
	ID                    string    `json:"id"`
	OperationCategoryID   string    `json:"operation_category_id"`
	OperationCategoryName string    `json:"operation_category_name,omitempty"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Status                int       `json:"status"` // 1: Active, 0: Inactive
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type HospitalOperationInput struct {
	OperationCategoryID string `json:"operation_category_id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Status              int    `json:"status"`
}

func (i *HospitalOperationInput) Validate() error {
	i.OperationCategoryID = strings.TrimSpace(i.OperationCategoryID)
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	if !UUIDPattern.MatchString(i.OperationCategoryID) {
		return ErrValidation
	}
	if i.Name == "" || len(i.Name) > 191 {
		return ErrValidation
	}
	if i.Status != 0 && i.Status != 1 {
		i.Status = 1
	}
	return nil
}

type CustomField struct {
	ID         string    `json:"id"`
	ModuleName string    `json:"module_name"`
	FieldType  string    `json:"field_type"` // "text", "number", "select", "date", "boolean", "textarea"
	FieldName  string    `json:"field_name"`
	IsRequired bool      `json:"is_required"`
	Values     string    `json:"values"`
	Grid       int       `json:"grid"` // 1-12
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CustomFieldInput struct {
	ModuleName string `json:"module_name"`
	FieldType  string `json:"field_type"`
	FieldName  string `json:"field_name"`
	IsRequired bool   `json:"is_required"`
	Values     string `json:"values"`
	Grid       int    `json:"grid"`
}

func (i *CustomFieldInput) Validate() error {
	i.ModuleName = strings.ToLower(strings.TrimSpace(i.ModuleName))
	i.FieldType = strings.ToLower(strings.TrimSpace(i.FieldType))
	i.FieldName = strings.TrimSpace(i.FieldName)
	i.Values = strings.TrimSpace(i.Values)
	if i.ModuleName == "" || len(i.ModuleName) > 191 {
		return ErrValidation
	}
	if i.FieldName == "" || len(i.FieldName) > 191 {
		return ErrValidation
	}
	validTypes := map[string]bool{
		"text": true, "number": true, "select": true, "date": true, "boolean": true, "textarea": true,
	}
	if !validTypes[i.FieldType] {
		return ErrValidation
	}
	if i.Grid == 0 {
		i.Grid = 12
	} else if i.Grid < 1 || i.Grid > 12 {
		return ErrValidation
	}
	return nil
}

type HospitalModuleSetting struct {
	ID        string    `json:"id"`
	ModuleKey string    `json:"module_key"`
	Name      string    `json:"name"`
	Route     string    `json:"route"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ModuleSettingUpdateInput struct {
	IsActive bool `json:"is_active"`
}
