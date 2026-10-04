package domain

import (
	"strings"
	"time"
)

type MedicineInput struct {
	Name              string `json:"name"`
	Category          string `json:"category"`
	Brand             string `json:"brand"`
	Unit              string `json:"unit"`
	Composition       string `json:"composition"`
	SideEffects       string `json:"sideEffects"`
	SellingPriceMinor int64  `json:"sellingPriceMinor"`
}
type Medicine struct {
	MedicineInput
	ID string `json:"id"`
}

func validText(s *string, min, max int) bool {
	*s = strings.TrimSpace(*s)
	n := len([]rune(*s))
	return n >= min && n <= max
}
func (m *MedicineInput) Validate() error {
	if !validText(&m.Name, 2, 150) || !validText(&m.Category, 1, 100) || !validText(&m.Brand, 1, 100) || !validText(&m.Unit, 1, 40) || !validText(&m.Composition, 0, 2000) || !validText(&m.SideEffects, 0, 2000) || m.SellingPriceMinor < 0 || m.SellingPriceMinor > 1000000000 {
		return ErrValidation
	}
	return nil
}

type BatchInput struct {
	MedicineID        string `json:"medicineId"`
	Lot               string `json:"lot"`
	ExpiryDate        string `json:"expiryDate"`
	Supplier          string `json:"supplier"`
	PurchaseReference string `json:"purchaseReference"`
	UnitCostMinor     int64  `json:"unitCostMinor"`
	Quantity          int    `json:"quantity"`
}
type MedicineBatch struct {
	BatchInput
	ID      string `json:"id"`
	Balance int    `json:"balance"`
}

func (b *BatchInput) Validate(now time.Time) error {
	expiry, e := time.ParseInLocation("2006-01-02", b.ExpiryDate, HospitalLocation)
	if e != nil || expiry.Format("2006-01-02") < now.In(HospitalLocation).Format("2006-01-02") || !UUIDPattern.MatchString(b.MedicineID) || !validText(&b.Lot, 1, 100) || !validText(&b.Supplier, 1, 200) || !validText(&b.PurchaseReference, 1, 100) || b.UnitCostMinor < 0 || b.UnitCostMinor > 1000000000 || b.Quantity < 1 || b.Quantity > 1000000 {
		return ErrValidation
	}
	return nil
}

type MedicationInput struct {
	EncounterID  string `json:"encounterId"`
	MedicineID   string `json:"medicineId"`
	Quantity     int    `json:"quantity"`
	Dose         string `json:"dose"`
	Route        string `json:"route"`
	Frequency    string `json:"frequency"`
	DurationDays int    `json:"durationDays"`
	Instructions string `json:"instructions"`
}
type MedicationOrder struct {
	MedicationInput
	ID           string    `json:"id"`
	MedicineName string    `json:"medicineName"`
	Unit         string    `json:"unit"`
	SignedBy     string    `json:"signedBy"`
	SignedAt     time.Time `json:"signedAt"`
	Cancelled    bool      `json:"cancelled"`
	Dispensed    int       `json:"dispensed"`
}

func (m *MedicationInput) Validate() error {
	if !UUIDPattern.MatchString(m.EncounterID) || !UUIDPattern.MatchString(m.MedicineID) || m.Quantity < 1 || m.Quantity > 10000 || !validText(&m.Dose, 1, 100) || !validText(&m.Route, 1, 100) || !validText(&m.Frequency, 1, 100) || m.DurationDays < 1 || m.DurationDays > 365 || !validText(&m.Instructions, 0, 2000) {
		return ErrValidation
	}
	return nil
}

type StockInput struct {
	BatchID    string `json:"batchId"`
	OrderID    string `json:"orderId"`
	Kind       string `json:"kind"`
	Quantity   int    `json:"quantity"`
	OriginalID string `json:"originalId"`
	Reason     string `json:"reason"`
}
type StockMovement struct {
	StockInput
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *StockInput) Validate() error {
	if !UUIDPattern.MatchString(s.BatchID) || s.Quantity < 1 || s.Quantity > 1000000 || !validText(&s.Reason, 0, 1000) {
		return ErrValidation
	}
	switch s.Kind {
	case "dispense":
		if !UUIDPattern.MatchString(s.OrderID) || s.OriginalID != "" {
			return ErrValidation
		}
	case "return":
		if !UUIDPattern.MatchString(s.OrderID) || !UUIDPattern.MatchString(s.OriginalID) || s.Reason == "" {
			return ErrValidation
		}
	case "disposal":
		if s.OrderID != "" || s.OriginalID != "" || s.Reason == "" {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}
