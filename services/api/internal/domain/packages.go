package domain

import (
	"encoding/json"
	"math"
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
	HasRate      bool   `json:"-"`
}

func (p *PackageServiceLineInput) UnmarshalJSON(data []byte) error {
	type Alias PackageServiceLineInput
	var aux struct {
		Alias
		RawRateMinor    *int64   `json:"rate_minor"`
		RawAltRateMinor *int64   `json:"rateMinor"`
		RawRate         *float64 `json:"rate"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*p = PackageServiceLineInput(aux.Alias)
	if aux.RawRateMinor != nil {
		p.RateMinor = *aux.RawRateMinor
		p.HasRate = true
	} else if aux.RawAltRateMinor != nil {
		p.RateMinor = *aux.RawAltRateMinor
		p.HasRate = true
	} else if aux.RawRate != nil {
		p.RateMinor = int64(math.Round(*aux.RawRate * 100))
		p.HasRate = true
	}
	return nil
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
			in.Services[i].HasRate = true
		} else if in.Services[i].RateMinor != 0 {
			in.Services[i].HasRate = true
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
	if len(in.Services) == 0 || len(in.Services) > MaxLines {
		return ErrValidation
	}
	seenServices := make(map[string]bool)
	seenLineIDs := make(map[string]bool)
	for i := range in.Services {
		in.Services[i].ServiceID = strings.TrimSpace(in.Services[i].ServiceID)
		in.Services[i].ID = strings.TrimSpace(in.Services[i].ID)
		if in.Services[i].ServiceID == "" || in.Services[i].Quantity <= 0 || in.Services[i].Quantity > MaxQuantity || in.Services[i].RateMinor < 0 || in.Services[i].RateMinor > MaxMoneyMinor {
			return ErrValidation
		}
		if seenServices[in.Services[i].ServiceID] {
			return ErrValidation
		}
		seenServices[in.Services[i].ServiceID] = true
		if in.Services[i].ID != "" {
			if seenLineIDs[in.Services[i].ID] {
				return ErrValidation
			}
			seenLineIDs[in.Services[i].ID] = true
		}
	}
	if _, _, _, _, err := CalculatePackageTotals(in.Discount, in.Services); err != nil {
		return ErrValidation
	}
	return nil
}

// CalculatePackageTotals computes line amounts, subtotal, and authoritative discount and total.
func CalculatePackageTotals(discount int, services []PackageServiceLineInput) (subtotal int64, discountAmount int64, total int64, lineAmounts []int64, err error) {
	if discount < 0 || discount > 100 {
		return 0, 0, 0, nil, ErrValidation
	}
	lineAmounts = make([]int64, len(services))
	for i, s := range services {
		if s.Quantity <= 0 || s.Quantity > MaxQuantity || s.RateMinor < 0 || s.RateMinor > MaxMoneyMinor {
			return 0, 0, 0, nil, ErrValidation
		}
		amt, err := safeMul(int64(s.Quantity), s.RateMinor)
		if err != nil {
			return 0, 0, 0, nil, err
		}
		lineAmounts[i] = amt
		subtotal, err = safeAdd(subtotal, amt)
		if err != nil {
			return 0, 0, 0, nil, err
		}
	}
	if discount > 0 {
		// Explicit half-up rounding: (subtotal * discount + 50) / 100
		discountAmount = (subtotal*int64(discount) + 50) / 100
		if discountAmount > subtotal {
			discountAmount = subtotal
		}
	}
	total = subtotal - discountAmount
	if total < 0 {
		total = 0
	}
	return subtotal, discountAmount, total, lineAmounts, nil
}
