package domain

import (
	"strings"
	"time"
)

// --- Complaints ---

type HospitalComplaint struct {
	ID          string     `json:"id"`
	PatientID   string     `json:"patient_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      int        `json:"status"` // 0: Pending, 1: In Progress, 2: Resolved, 3: Rejected
	Response    string     `json:"response"`
	ResolvedBy  *string    `json:"resolved_by,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ComplaintCreateInput struct {
	PatientID   string `json:"patient_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (i *ComplaintCreateInput) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.Description = strings.TrimSpace(i.Description)
	if i.Title == "" || len(i.Title) > 191 {
		return ErrValidation
	}
	if i.Description == "" {
		return ErrValidation
	}
	return nil
}

type ComplaintResolveInput struct {
	Status   int    `json:"status"` // 1: In Progress, 2: Resolved, 3: Rejected
	Response string `json:"response"`
}

func (i *ComplaintResolveInput) Validate() error {
	if i.Status < 1 || i.Status > 3 {
		return ErrValidation
	}
	i.Response = strings.TrimSpace(i.Response)
	return nil
}

// --- Notice Board ---

type HospitalNoticeBoard struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NoticeBoardInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (i *NoticeBoardInput) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.Description = strings.TrimSpace(i.Description)
	if i.Title == "" || len(i.Title) > 191 {
		return ErrValidation
	}
	return nil
}

// --- Enquiries ---

type HospitalEnquiry struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	ContactNo string    `json:"contact_no"`
	Type      int       `json:"type"` // 1: General, 2: Admission, 3: Billing, 4: Feedback
	Message   string    `json:"message"`
	ViewedBy  *string   `json:"viewed_by,omitempty"`
	Status    int       `json:"status"` // 0: Unread, 1: Read
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EnquiryInput struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	ContactNo string `json:"contact_no"`
	Type      int    `json:"type"`
	Message   string `json:"message"`
}

func (i *EnquiryInput) Validate() error {
	i.FullName = strings.TrimSpace(i.FullName)
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))
	i.ContactNo = strings.TrimSpace(i.ContactNo)
	i.Message = strings.TrimSpace(i.Message)
	if i.FullName == "" || len(i.FullName) > 191 {
		return ErrValidation
	}
	if i.Email == "" || !strings.Contains(i.Email, "@") {
		return ErrValidation
	}
	if i.Message == "" {
		return ErrValidation
	}
	if i.Type < 1 || i.Type > 4 {
		i.Type = 1
	}
	return nil
}

// --- Visitors ---

type HospitalVisitor struct {
	ID         string    `json:"id"`
	Purpose    int       `json:"purpose"` // 1: Visit, 2: Enquiry, 3: Seminar/Vendor
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	IDCard     string    `json:"id_card"`
	NoOfPerson int       `json:"no_of_person"`
	Date       string    `json:"date"` // YYYY-MM-DD
	InTime     string    `json:"in_time"`
	OutTime    string    `json:"out_time"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type VisitorInput struct {
	Purpose    int    `json:"purpose"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	IDCard     string `json:"id_card"`
	NoOfPerson int    `json:"no_of_person"`
	Date       string `json:"date"`
	InTime     string `json:"in_time"`
	OutTime    string `json:"out_time"`
	Note       string `json:"note"`
}

func (i *VisitorInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Phone = strings.TrimSpace(i.Phone)
	i.Date = strings.TrimSpace(i.Date)
	if i.Name == "" || len(i.Name) > 191 {
		return ErrValidation
	}
	if i.Date == "" {
		return ErrValidation
	}
	if i.NoOfPerson < 1 {
		i.NoOfPerson = 1
	}
	if i.Purpose < 1 || i.Purpose > 3 {
		i.Purpose = 1
	}
	return nil
}

// --- Call Logs ---

type HospitalCallLog struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Date         string    `json:"date"`
	FollowUpDate *string   `json:"follow_up_date,omitempty"`
	Note         string    `json:"note"`
	CallType     int       `json:"call_type"` // 1: Incoming, 2: Outgoing
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CallLogInput struct {
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	Date         string  `json:"date"`
	FollowUpDate *string `json:"follow_up_date,omitempty"`
	Note         string  `json:"note"`
	CallType     int     `json:"call_type"`
}

func (i *CallLogInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Phone = strings.TrimSpace(i.Phone)
	i.Date = strings.TrimSpace(i.Date)
	if i.Name == "" || len(i.Name) > 191 {
		return ErrValidation
	}
	if i.Date == "" {
		return ErrValidation
	}
	if i.CallType != 1 && i.CallType != 2 {
		return ErrValidation
	}
	return nil
}

// --- Postals ---

type HospitalPostal struct {
	ID          string    `json:"id"`
	FromTitle   string    `json:"from_title"`
	ToTitle     string    `json:"to_title"`
	ReferenceNo string    `json:"reference_no"`
	Date        string    `json:"date"`
	Address     string    `json:"address"`
	Type        int       `json:"type"` // 1: Receive, 2: Dispatch
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PostalInput struct {
	FromTitle   string `json:"from_title"`
	ToTitle     string `json:"to_title"`
	ReferenceNo string `json:"reference_no"`
	Date        string `json:"date"`
	Address     string `json:"address"`
	Type        int    `json:"type"`
}

func (i *PostalInput) Validate() error {
	i.FromTitle = strings.TrimSpace(i.FromTitle)
	i.ToTitle = strings.TrimSpace(i.ToTitle)
	i.ReferenceNo = strings.TrimSpace(i.ReferenceNo)
	i.Date = strings.TrimSpace(i.Date)
	if i.Date == "" {
		return ErrValidation
	}
	if i.Type != 1 && i.Type != 2 {
		return ErrValidation
	}
	return nil
}
