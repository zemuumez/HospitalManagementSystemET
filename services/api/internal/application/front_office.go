package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type FrontOfficeStore interface {
	Complaints(context.Context, string, int) ([]domain.HospitalComplaint, int, error)
	Complaint(context.Context, string) (domain.HospitalComplaint, error)
	CreateComplaint(context.Context, domain.Actor, domain.ComplaintCreateInput) (domain.HospitalComplaint, error)
	ResolveComplaint(context.Context, domain.Actor, string, domain.ComplaintResolveInput) (domain.HospitalComplaint, error)

	Notices(context.Context) ([]domain.HospitalNoticeBoard, error)
	CreateNotice(context.Context, domain.Actor, domain.NoticeBoardInput) (domain.HospitalNoticeBoard, error)
	DeleteNotice(context.Context, domain.Actor, string) error

	Enquiries(context.Context, int) ([]domain.HospitalEnquiry, int, error)
	SubmitEnquiry(context.Context, domain.EnquiryInput) (domain.HospitalEnquiry, error)
	MarkEnquiryRead(context.Context, domain.Actor, string) (domain.HospitalEnquiry, error)

	Visitors(context.Context, int, string) ([]domain.HospitalVisitor, int, error)
	CreateVisitor(context.Context, domain.Actor, domain.VisitorInput) (domain.HospitalVisitor, error)
	UpdateVisitor(context.Context, domain.Actor, string, domain.VisitorInput) (domain.HospitalVisitor, error)

	CallLogs(context.Context, int, int) ([]domain.HospitalCallLog, int, error)
	CreateCallLog(context.Context, domain.Actor, domain.CallLogInput) (domain.HospitalCallLog, error)
	UpdateCallLog(context.Context, domain.Actor, string, domain.CallLogInput) (domain.HospitalCallLog, error)

	Postals(context.Context, int, int) ([]domain.HospitalPostal, int, error)
	CreatePostal(context.Context, domain.Actor, domain.PostalInput) (domain.HospitalPostal, error)
	UpdatePostal(context.Context, domain.Actor, string, domain.PostalInput) (domain.HospitalPostal, error)
}

type FrontOfficeService struct {
	Store FrontOfficeStore
	Now   func() time.Time
}

// --- Complaints ---

func (s FrontOfficeService) Complaints(ctx context.Context, a domain.Actor, patientID string, page int) ([]domain.HospitalComplaint, int, error) {
	if !a.Can("complaints.read") {
		return nil, 0, domain.ErrForbidden
	}
	// Patients can ONLY view their own complaints
	if a.Role == "patient" {
		patientID = a.ID
	}
	return s.Store.Complaints(ctx, patientID, page)
}

func (s FrontOfficeService) Complaint(ctx context.Context, a domain.Actor, id string) (domain.HospitalComplaint, error) {
	if !a.Can("complaints.read") {
		return domain.HospitalComplaint{}, domain.ErrForbidden
	}
	c, err := s.Store.Complaint(ctx, id)
	if err != nil {
		return domain.HospitalComplaint{}, err
	}
	if a.Role == "patient" && c.PatientID != a.ID {
		return domain.HospitalComplaint{}, domain.ErrForbidden
	}
	return c, nil
}

func (s FrontOfficeService) CreateComplaint(ctx context.Context, a domain.Actor, in domain.ComplaintCreateInput) (domain.HospitalComplaint, error) {
	if !a.Can("complaints.create") {
		return domain.HospitalComplaint{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalComplaint{}, err
	}
	if a.Role == "patient" || in.PatientID == "" {
		in.PatientID = a.ID
	}
	return s.Store.CreateComplaint(ctx, a, in)
}

func (s FrontOfficeService) ResolveComplaint(ctx context.Context, a domain.Actor, id string, in domain.ComplaintResolveInput) (domain.HospitalComplaint, error) {
	if !a.Can("complaints.manage") {
		return domain.HospitalComplaint{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalComplaint{}, err
	}
	return s.Store.ResolveComplaint(ctx, a, id, in)
}

// --- Notice Board ---

func (s FrontOfficeService) Notices(ctx context.Context, a domain.Actor) ([]domain.HospitalNoticeBoard, error) {
	if a.ID != "" && !a.Can("notices.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.Notices(ctx)
}

func (s FrontOfficeService) CreateNotice(ctx context.Context, a domain.Actor, in domain.NoticeBoardInput) (domain.HospitalNoticeBoard, error) {
	if !a.Can("notices.manage") {
		return domain.HospitalNoticeBoard{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalNoticeBoard{}, err
	}
	return s.Store.CreateNotice(ctx, a, in)
}

func (s FrontOfficeService) DeleteNotice(ctx context.Context, a domain.Actor, id string) error {
	if !a.Can("notices.manage") {
		return domain.ErrForbidden
	}
	return s.Store.DeleteNotice(ctx, a, id)
}

// --- Enquiries ---

func (s FrontOfficeService) Enquiries(ctx context.Context, a domain.Actor, page int) ([]domain.HospitalEnquiry, int, error) {
	if !a.Can("front_office.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.Enquiries(ctx, page)
}

func (s FrontOfficeService) SubmitEnquiry(ctx context.Context, in domain.EnquiryInput) (domain.HospitalEnquiry, error) {
	if err := in.Validate(); err != nil {
		return domain.HospitalEnquiry{}, err
	}
	return s.Store.SubmitEnquiry(ctx, in)
}

func (s FrontOfficeService) MarkEnquiryRead(ctx context.Context, a domain.Actor, id string) (domain.HospitalEnquiry, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalEnquiry{}, domain.ErrForbidden
	}
	return s.Store.MarkEnquiryRead(ctx, a, id)
}

// --- Visitors ---

func (s FrontOfficeService) Visitors(ctx context.Context, a domain.Actor, page int, date string) ([]domain.HospitalVisitor, int, error) {
	if !a.Can("front_office.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.Visitors(ctx, page, date)
}

func (s FrontOfficeService) CreateVisitor(ctx context.Context, a domain.Actor, in domain.VisitorInput) (domain.HospitalVisitor, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalVisitor{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalVisitor{}, err
	}
	return s.Store.CreateVisitor(ctx, a, in)
}

func (s FrontOfficeService) UpdateVisitor(ctx context.Context, a domain.Actor, id string, in domain.VisitorInput) (domain.HospitalVisitor, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalVisitor{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalVisitor{}, err
	}
	return s.Store.UpdateVisitor(ctx, a, id, in)
}

// --- Call Logs ---

func (s FrontOfficeService) CallLogs(ctx context.Context, a domain.Actor, page int, callType int) ([]domain.HospitalCallLog, int, error) {
	if !a.Can("front_office.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.CallLogs(ctx, page, callType)
}

func (s FrontOfficeService) CreateCallLog(ctx context.Context, a domain.Actor, in domain.CallLogInput) (domain.HospitalCallLog, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalCallLog{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalCallLog{}, err
	}
	return s.Store.CreateCallLog(ctx, a, in)
}

func (s FrontOfficeService) UpdateCallLog(ctx context.Context, a domain.Actor, id string, in domain.CallLogInput) (domain.HospitalCallLog, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalCallLog{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalCallLog{}, err
	}
	return s.Store.UpdateCallLog(ctx, a, id, in)
}

// --- Postals ---

func (s FrontOfficeService) Postals(ctx context.Context, a domain.Actor, page int, postalType int) ([]domain.HospitalPostal, int, error) {
	if !a.Can("front_office.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.Postals(ctx, page, postalType)
}

func (s FrontOfficeService) CreatePostal(ctx context.Context, a domain.Actor, in domain.PostalInput) (domain.HospitalPostal, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalPostal{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalPostal{}, err
	}
	return s.Store.CreatePostal(ctx, a, in)
}

func (s FrontOfficeService) UpdatePostal(ctx context.Context, a domain.Actor, id string, in domain.PostalInput) (domain.HospitalPostal, error) {
	if !a.Can("front_office.manage") {
		return domain.HospitalPostal{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalPostal{}, err
	}
	return s.Store.UpdatePostal(ctx, a, id, in)
}
