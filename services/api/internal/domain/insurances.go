package domain

import (
	"math"
	"strings"
	"time"
)

type InsuranceDiseaseLine struct {
	ID                 string    `json:"id"`
	InsuranceID        string    `json:"insurance_id"`
	DiseaseName        string    `json:"disease_name"`
	DiseaseChargeMinor int64     `json:"disease_charge_minor"`
	DiseaseCharge      float64   `json:"disease_charge"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Insurance struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	ServiceTaxMinor   int64                  `json:"service_tax_minor"`
	ServiceTax        float64                `json:"service_tax"`
	Discount          int                    `json:"discount"`
	Remark            string                 `json:"remark"`
	InsuranceNo       string                 `json:"insurance_no"`
	InsuranceCode     string                 `json:"insurance_code"`
	HospitalRateMinor int64                  `json:"hospital_rate_minor"`
	HospitalRate      float64                `json:"hospital_rate"`
	TotalMinor        int64                  `json:"total_minor"`
	Total             float64                `json:"total"`
	Status            int                    `json:"status"` // 1 = active, 0 = inactive
	CurrencySymbol    string                 `json:"currency_symbol"`
	Diseases          []InsuranceDiseaseLine `json:"diseases,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type InsuranceDiseaseLineInput struct {
	ID                    string   `json:"id,omitempty"`
	DiseaseName           string   `json:"disease_name"`
	AltDiseaseName        string   `json:"diseaseName,omitempty"`
	AltName               string   `json:"name,omitempty"`
	DiseaseChargeMinor    int64    `json:"disease_charge_minor"`
	AltDiseaseChargeMinor int64    `json:"diseaseChargeMinor,omitempty"`
	AltDiseaseCharge      *float64 `json:"diseaseCharge,omitempty"`
	AltCharge             *float64 `json:"charge,omitempty"`
}

type InsuranceInput struct {
	Name                 string                      `json:"name"`
	ServiceTaxMinor      int64                       `json:"service_tax_minor"`
	AltServiceTaxMinor   int64                       `json:"serviceTaxMinor,omitempty"`
	AltServiceTax        *float64                    `json:"serviceTax,omitempty"`
	Discount             int                         `json:"discount"`
	Remark               string                      `json:"remark"`
	InsuranceNo          string                      `json:"insurance_no"`
	AltInsuranceNo       string                      `json:"insuranceNo,omitempty"`
	InsuranceCode        string                      `json:"insurance_code"`
	AltInsuranceCode     string                      `json:"insuranceCode,omitempty"`
	HospitalRateMinor    int64                       `json:"hospital_rate_minor"`
	AltHospitalRateMinor int64                       `json:"hospitalRateMinor,omitempty"`
	AltHospitalRate      *float64                    `json:"hospitalRate,omitempty"`
	Status               *int                        `json:"status,omitempty"`
	Diseases             []InsuranceDiseaseLineInput `json:"diseases"`
}

func (in *InsuranceInput) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.InsuranceNo == "" && in.AltInsuranceNo != "" {
		in.InsuranceNo = in.AltInsuranceNo
	}
	in.InsuranceNo = strings.TrimSpace(in.InsuranceNo)

	if in.InsuranceCode == "" && in.AltInsuranceCode != "" {
		in.InsuranceCode = in.AltInsuranceCode
	}
	in.InsuranceCode = strings.TrimSpace(in.InsuranceCode)

	if in.ServiceTaxMinor == 0 && in.AltServiceTaxMinor != 0 {
		in.ServiceTaxMinor = in.AltServiceTaxMinor
	} else if in.ServiceTaxMinor == 0 && in.AltServiceTax != nil {
		in.ServiceTaxMinor = int64(math.Round(*in.AltServiceTax * 100))
	}

	if in.HospitalRateMinor == 0 && in.AltHospitalRateMinor != 0 {
		in.HospitalRateMinor = in.AltHospitalRateMinor
	} else if in.HospitalRateMinor == 0 && in.AltHospitalRate != nil {
		in.HospitalRateMinor = int64(math.Round(*in.AltHospitalRate * 100))
	}

	for i := range in.Diseases {
		if in.Diseases[i].DiseaseName == "" && in.Diseases[i].AltDiseaseName != "" {
			in.Diseases[i].DiseaseName = in.Diseases[i].AltDiseaseName
		}
		if in.Diseases[i].DiseaseName == "" && in.Diseases[i].AltName != "" {
			in.Diseases[i].DiseaseName = in.Diseases[i].AltName
		}
		in.Diseases[i].DiseaseName = strings.TrimSpace(in.Diseases[i].DiseaseName)

		if in.Diseases[i].DiseaseChargeMinor == 0 && in.Diseases[i].AltDiseaseChargeMinor != 0 {
			in.Diseases[i].DiseaseChargeMinor = in.Diseases[i].AltDiseaseChargeMinor
		} else if in.Diseases[i].DiseaseChargeMinor == 0 && in.Diseases[i].AltDiseaseCharge != nil {
			in.Diseases[i].DiseaseChargeMinor = int64(math.Round(*in.Diseases[i].AltDiseaseCharge * 100))
		} else if in.Diseases[i].DiseaseChargeMinor == 0 && in.Diseases[i].AltCharge != nil {
			in.Diseases[i].DiseaseChargeMinor = int64(math.Round(*in.Diseases[i].AltCharge * 100))
		}
	}
}

func (in *InsuranceInput) Validate() error {
	in.Normalize()
	if in.Name == "" || len(in.Name) > 160 {
		return ErrValidation
	}
	if in.InsuranceNo == "" || len(in.InsuranceNo) > 191 {
		return ErrValidation
	}
	if in.InsuranceCode == "" || len(in.InsuranceCode) > 191 {
		return ErrValidation
	}
	if in.Discount < 0 || in.Discount > 100 {
		return ErrValidation
	}
	if in.Status != nil && (*in.Status != 0 && *in.Status != 1) {
		return ErrValidation
	}
	if in.ServiceTaxMinor < 0 || in.ServiceTaxMinor > MaxMoneyMinor || in.HospitalRateMinor < 0 || in.HospitalRateMinor > MaxMoneyMinor {
		return ErrValidation
	}
	if len(in.Diseases) == 0 || len(in.Diseases) > MaxLines {
		return ErrValidation
	}
	seenDiseaseIDs := make(map[string]bool)
	for i := range in.Diseases {
		in.Diseases[i].DiseaseName = strings.TrimSpace(in.Diseases[i].DiseaseName)
		in.Diseases[i].ID = strings.TrimSpace(in.Diseases[i].ID)
		if in.Diseases[i].DiseaseName == "" || len(in.Diseases[i].DiseaseName) > 191 {
			return ErrValidation
		}
		if in.Diseases[i].DiseaseChargeMinor < 0 || in.Diseases[i].DiseaseChargeMinor > MaxMoneyMinor {
			return ErrValidation
		}
		if in.Diseases[i].ID != "" {
			if seenDiseaseIDs[in.Diseases[i].ID] {
				return ErrValidation
			}
			seenDiseaseIDs[in.Diseases[i].ID] = true
		}
	}
	if _, _, _, err := CalculateInsuranceTotals(in.ServiceTaxMinor, in.HospitalRateMinor, in.Discount, in.Diseases); err != nil {
		return ErrValidation
	}
	return nil
}

// CalculateInsuranceTotals computes base = service_tax + hospital_rate + sum(disease charges),
// discount amount, and total with half-up rounding.
func CalculateInsuranceTotals(serviceTaxMinor int64, hospitalRateMinor int64, discount int, diseases []InsuranceDiseaseLineInput) (baseMinor int64, discountAmountMinor int64, totalMinor int64, err error) {
	if discount < 0 || discount > 100 {
		return 0, 0, 0, ErrValidation
	}
	if serviceTaxMinor < 0 || serviceTaxMinor > MaxMoneyMinor || hospitalRateMinor < 0 || hospitalRateMinor > MaxMoneyMinor {
		return 0, 0, 0, ErrValidation
	}
	baseMinor, err = safeAdd(serviceTaxMinor, hospitalRateMinor)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, d := range diseases {
		if d.DiseaseChargeMinor < 0 || d.DiseaseChargeMinor > MaxMoneyMinor {
			return 0, 0, 0, ErrValidation
		}
		baseMinor, err = safeAdd(baseMinor, d.DiseaseChargeMinor)
		if err != nil {
			return 0, 0, 0, err
		}
	}
	if discount > 0 {
		discountAmountMinor = (baseMinor*int64(discount) + 50) / 100
		if discountAmountMinor > baseMinor {
			discountAmountMinor = baseMinor
		}
	}
	totalMinor = baseMinor - discountAmountMinor
	if totalMinor < 0 {
		totalMinor = 0
	}
	return baseMinor, discountAmountMinor, totalMinor, nil
}
