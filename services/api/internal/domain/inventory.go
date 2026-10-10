package domain

import "time"

type InventoryCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Version     int    `json:"version"`
}
type InventoryCategory struct {
	InventoryCategoryInput
	ID string `json:"id"`
}

func (i *InventoryCategoryInput) Validate() error {
	if !validText(&i.Name, 1, 100) || !validText(&i.Description, 0, 1000) || i.Version < 0 {
		return ErrValidation
	}
	return nil
}

type InventoryItemInput struct {
	CategoryID   string `json:"categoryId"`
	Name         string `json:"name"`
	Unit         string `json:"unit"`
	Description  string `json:"description"`
	ReorderMilli int64  `json:"reorderMilli"`
	Active       bool   `json:"active"`
	Version      int    `json:"version"`
}
type InventoryItem struct {
	InventoryItemInput
	ID           string `json:"id"`
	BalanceMilli int64  `json:"balanceMilli"`
}

func (i *InventoryItemInput) Validate() error {
	if !UUIDPattern.MatchString(i.CategoryID) || !validText(&i.Name, 1, 150) || !validText(&i.Unit, 1, 40) || !validText(&i.Description, 0, 2000) || i.ReorderMilli < 0 || i.ReorderMilli > 1000000000000 || i.Version < 0 {
		return ErrValidation
	}
	return nil
}

type InventoryMovementInput struct {
	ItemID        string `json:"itemId"`
	Kind          string `json:"kind"`
	QuantityMilli int64  `json:"quantityMilli"`
	RecipientID   string `json:"recipientId"`
	OriginalID    string `json:"originalId"`
	Supplier      string `json:"supplier"`
	StoreName     string `json:"storeName"`
	Reference     string `json:"reference"`
	CostMinor     int64  `json:"costMinor"`
	Restock       bool   `json:"restock"`
	Reason        string `json:"reason"`
	AttachmentURL string `json:"attachmentUrl"`
	IssuedDate    string `json:"issuedDate"`
	ReturnDueDate string `json:"returnDueDate"`
	IssuedBy      string `json:"issuedBy"`
	Department    string `json:"department"`
}
type InventoryMovement struct {
	InventoryMovementInput
	ID            string    `json:"id"`
	ItemName      string    `json:"itemName,omitempty"`
	DeltaMilli    int64     `json:"deltaMilli"`
	ReturnedMilli int64     `json:"returnedMilli"`
	IsVoided      bool      `json:"isVoided,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (i *InventoryMovementInput) Validate() error {
	if !UUIDPattern.MatchString(i.ItemID) || i.QuantityMilli < 1 || i.QuantityMilli > 1000000000000 ||
		!validText(&i.RecipientID, 0, 128) || !validText(&i.Supplier, 0, 200) ||
		!validText(&i.StoreName, 0, 100) || !validText(&i.Reference, 0, 100) ||
		!validText(&i.Reason, 1, 1000) || i.CostMinor < 0 || i.CostMinor > 1000000000000 ||
		!validText(&i.AttachmentURL, 0, 255) || !validText(&i.IssuedDate, 0, 30) ||
		!validText(&i.ReturnDueDate, 0, 30) || !validText(&i.IssuedBy, 0, 128) ||
		!validText(&i.Department, 0, 100) {
		return ErrValidation
	}
	switch i.Kind {
	case "receive":
		if i.Reference == "" || i.RecipientID != "" || i.OriginalID != "" || i.Restock {
			return ErrValidation
		}
	case "issue":
		if i.RecipientID == "" || i.OriginalID != "" || i.Restock {
			return ErrValidation
		}
	case "return":
		if !UUIDPattern.MatchString(i.OriginalID) || i.RecipientID != "" {
			return ErrValidation
		}
	case "writeoff":
		if i.RecipientID != "" || i.OriginalID != "" || i.Restock {
			return ErrValidation
		}
	case "void_receipt":
		if !UUIDPattern.MatchString(i.OriginalID) || i.RecipientID != "" || i.Restock {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	if i.Kind != "receive" && i.Kind != "void_receipt" && (i.Supplier != "" || i.Reference != "" || i.CostMinor != 0 || i.AttachmentURL != "") {
		return ErrValidation
	}
	if i.Kind != "issue" && (i.IssuedDate != "" || i.ReturnDueDate != "" || i.IssuedBy != "" || i.Department != "") {
		return ErrValidation
	}
	return nil
}
