package domain

import (
	"strings"
	"time"
)

type PackageServiceLine struct {
	ID          string    `json:"id"`
	PackageID   string    `json:"package_id"`
	ServiceID   string    `json:"service_id"`
	ServiceName string    `json:"service_name,omitempty"`
	Quantity    int       `json:"quantity"`
	RateMinor   int64     `json:"rate_minor"`
	AmountMinor int64     `json:"amount_minor"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Package struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	Discount         int                  `json:"discount"`
	TotalAmountMinor int64                `json:"total_amount_minor"`
	CurrencySymbol   string               `json:"currency_symbol"`
	Services         []PackageServiceLine `json:"services,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type PackageServiceLineInput struct {
	ID           string `json:"id,omitempty"`
	ServiceID    string `json:"service_id"`
	AltServiceID string `json:"serviceId,omitempty"`
	Quantity     int    `json:"quantity"`
	RateMinor    int64  `json:"rate_minor"`
	AltRateMinor int64  `json:"rateMinor,omitempty"`
}

type PackageInput struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Discount    int                       `json:"discount"`
	Services    []PackageServiceLineInput `json:"services"`
}

func (in *PackageInput) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	for i := range in.Services {
		if in.Services[i].ServiceID == "" && in.Services[i].AltServiceID != "" {
			in.Services[i].ServiceID = in.Services[i].AltServiceID
		}
		if in.Services[i].RateMinor == 0 && in.Services[i].AltRateMinor != 0 {
			in.Services[i].RateMinor = in.Services[i].AltRateMinor
		}
	}
}

func (in *PackageInput) Validate() error {
	in.Normalize()
	if in.Name == "" || len(in.Name) > 160 {
		return ErrValidation
	}
	if in.Discount < 0 || in.Discount > 100 {
		return ErrValidation
	}
	if len(in.Services) == 0 {
		return ErrValidation
	}
	seenServices := make(map[string]bool)
	for i := range in.Services {
		in.Services[i].ServiceID = strings.TrimSpace(in.Services[i].ServiceID)
		in.Services[i].ID = strings.TrimSpace(in.Services[i].ID)
		if in.Services[i].ServiceID == "" || in.Services[i].Quantity <= 0 || in.Services[i].RateMinor < 0 {
			return ErrValidation
		}
		if seenServices[in.Services[i].ServiceID] {
			return ErrValidation
		}
		seenServices[in.Services[i].ServiceID] = true
	}
	return nil
}

// CalculatePackageTotals computes line amounts, subtotal, and authoritative discount and total.
func CalculatePackageTotals(discount int, services []PackageServiceLineInput) (subtotal int64, discountAmount int64, total int64, lineAmounts []int64) {
	lineAmounts = make([]int64, len(services))
	for i, s := range services {
		amt := int64(s.Quantity) * s.RateMinor
		lineAmounts[i] = amt
		subtotal += amt
	}
	if discount > 0 {
		// Explicit half-up rounding: (subtotal * discount + 50) / 100
		discountAmount = (subtotal*int64(discount) + 50) / 100
	}
	total = subtotal - discountAmount
	if total < 0 {
		total = 0
	}
	return subtotal, discountAmount, total, lineAmounts
}
