package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testFrontOffice(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()

	actorMap := make(map[string]domain.Actor)
	for _, a := range actors {
		actorMap[a.Role] = a
	}

	admin := actorMap["admin"]
	doctor := actorMap["doctor"]
	receptionist := actorMap["receptionist"]
	patientUser := actorMap["patient"]

	srv := application.FrontOfficeService{Store: store, Now: time.Now}

	// 1. Authorization checks
	// Patient cannot create visitor record
	_, err := srv.CreateVisitor(ctx, patientUser, domain.VisitorInput{
		Purpose:    1,
		Name:       "Unauthorized Visitor",
		Date:       "2026-10-05",
		NoOfPerson: 1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating visitor record, got %v", err)
	}

	// Doctor cannot resolve complaint (requires complaints.manage: admin, receptionist, case_manager)
	_, err = srv.ResolveComplaint(ctx, doctor, "dummy-id", domain.ComplaintResolveInput{
		Status:   2,
		Response: "Doctor resolution note",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor resolving complaint, got %v", err)
	}

	// 2. Validation checks
	// Validation error: empty complaint title
	_, err = srv.CreateComplaint(ctx, patientUser, domain.ComplaintCreateInput{
		Title:       "",
		Description: "Nurse delayed medication",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty complaint title, got %v", err)
	}

	// Validation error: invalid enquiry email
	_, err = srv.SubmitEnquiry(ctx, domain.EnquiryInput{
		FullName: "Kidus Daniel",
		Email:    "invalid-email-address",
		Message:  "General pricing enquiry",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid enquiry email, got %v", err)
	}

	// Validation error: invalid postal type
	_, err = srv.CreatePostal(ctx, receptionist, domain.PostalInput{
		Date: "2026-10-05",
		Type: 9,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid postal type 9, got %v", err)
	}

	// 3. Complaints lifecycle & Patient isolation
	complaint, err := srv.CreateComplaint(ctx, patientUser, domain.ComplaintCreateInput{
		Title:       "Pharmacy wait time too long",
		Description: "Waited 45 minutes for prescribed amoxicillin dispensation.",
	})
	if err != nil {
		t.Fatalf("failed to create complaint: %v", err)
	}
	if complaint.Status != 0 || complaint.PatientID != patientUser.ID {
		t.Fatalf("unexpected complaint: %+v", complaint)
	}

	// Patient lists complaints - should see own complaint
	patientComplaints, total, err := srv.Complaints(ctx, patientUser, "", 1)
	if err != nil || total == 0 || len(patientComplaints) == 0 {
		t.Fatalf("failed to list complaints for patient: %v, total: %d", err, total)
	}

	// Admin resolves complaint
	resolved, err := srv.ResolveComplaint(ctx, admin, complaint.ID, domain.ComplaintResolveInput{
		Status:   2,
		Response: "Pharmacy workflow reviewed and additional pharmacist assigned to peak hours.",
	})
	if err != nil {
		t.Fatalf("failed to resolve complaint: %v", err)
	}
	if resolved.Status != 2 || resolved.ResolvedBy == nil || *resolved.ResolvedBy != admin.ID {
		t.Fatalf("unexpected resolved complaint: %+v", resolved)
	}

	// 4. Notice Board CRUD
	notice, err := srv.CreateNotice(ctx, admin, domain.NoticeBoardInput{
		Title:       "Staff General Assembly on Friday",
		Description: "All clinical and administrative departments meet at Main Auditorium 16:00.",
	})
	if err != nil {
		t.Fatalf("failed to create notice: %v", err)
	}

	// Patient/staff reads notices
	notices, err := srv.Notices(ctx, patientUser)
	if err != nil || len(notices) == 0 {
		t.Fatalf("failed to read notices: %v", err)
	}

	// Admin deletes notice
	if err := srv.DeleteNotice(ctx, admin, notice.ID); err != nil {
		t.Fatalf("failed to delete notice: %v", err)
	}

	// 5. Enquiries lifecycle
	enq, err := srv.SubmitEnquiry(ctx, domain.EnquiryInput{
		FullName:  "Helen Mekonnen",
		Email:     "helen.m@example.com",
		ContactNo: "+251911445566",
		Type:      1,
		Message:   "Do you have pediatric cardiology consultations on weekends?",
	})
	if err != nil {
		t.Fatalf("failed to submit enquiry: %v", err)
	}
	if enq.Status != 0 {
		t.Fatalf("expected unread status 0, got %d", enq.Status)
	}

	// Receptionist marks enquiry read
	enqRead, err := srv.MarkEnquiryRead(ctx, receptionist, enq.ID)
	if err != nil {
		t.Fatalf("failed to mark enquiry read: %v", err)
	}
	if enqRead.Status != 1 || enqRead.ViewedBy == nil || *enqRead.ViewedBy != receptionist.ID {
		t.Fatalf("unexpected enquiry after mark read: %+v", enqRead)
	}

	// 6. Visitors CRUD
	vis, err := srv.CreateVisitor(ctx, receptionist, domain.VisitorInput{
		Purpose:    1,
		Name:       "Yonas Berhe",
		Phone:      "+251911778899",
		IDCard:     "KEBELE-09-8877",
		NoOfPerson: 2,
		Date:       "2026-10-05",
		InTime:     "14:00",
		OutTime:    "15:30",
		Note:       "Visiting Ward 3 Bed 102",
	})
	if err != nil || vis.ID == "" {
		t.Fatalf("failed to create visitor: %v", err)
	}

	visList, _, err := srv.Visitors(ctx, receptionist, 1, "2026-10-05")
	if err != nil || len(visList) == 0 {
		t.Fatalf("failed to list visitors: %v", err)
	}

	// 7. Call Logs CRUD
	call, err := srv.CreateCallLog(ctx, receptionist, domain.CallLogInput{
		Name:     "Tariku Lemma",
		Phone:    "+251922334455",
		Date:     "2026-10-05",
		Note:     "Inquired regarding MRI appointment scheduling",
		CallType: 1, // Incoming
	})
	if err != nil || call.ID == "" {
		t.Fatalf("failed to create call log: %v", err)
	}

	callList, _, err := srv.CallLogs(ctx, receptionist, 1, 1)
	if err != nil || len(callList) == 0 {
		t.Fatalf("failed to list call logs: %v", err)
	}

	// 8. Postals CRUD
	postal, err := srv.CreatePostal(ctx, receptionist, domain.PostalInput{
		FromTitle:   "Ethiopian Red Cross Society",
		ToTitle:     "Addis Ababa Central Hospital Blood Bank",
		ReferenceNo: "ERCS-POST-2026-09",
		Date:        "2026-10-05",
		Address:     "Addis Ababa, Ethiopia",
		Type:        1, // Receive
	})
	if err != nil || postal.ID == "" {
		t.Fatalf("failed to create postal record: %v", err)
	}

	postalList, _, err := srv.Postals(ctx, receptionist, 1, 1)
	if err != nil || len(postalList) == 0 {
		t.Fatalf("failed to list postals: %v", err)
	}

	// 9. HTTP API Integration
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id": receptionist.ID,
			},
			"session": map[string]any{
				"userId":    receptionist.ID,
				"expiresAt": time.Now().Add(time.Hour),
			},
		})
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		FrontOffice: srv,
		Actors:      store,
		AuthURL:     authSrv.URL,
		Origin:      "http://hospital.test",
		Client:      authSrv.Client(),
	}.Handler()

	// GET /v1/complaints
	req := httptest.NewRequest("GET", "/v1/complaints", nil)
	req.Header.Set("Cookie", "session=reception-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec := httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/complaints, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/notices
	req = httptest.NewRequest("GET", "/v1/notices", nil)
	req.Header.Set("Cookie", "session=reception-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/notices, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/enquiries
	req = httptest.NewRequest("GET", "/v1/enquiries", nil)
	req.Header.Set("Cookie", "session=reception-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/enquiries, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/visitors
	req = httptest.NewRequest("GET", "/v1/visitors", nil)
	req.Header.Set("Cookie", "session=reception-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/visitors, got %d: %s", rec.Code, rec.Body.String())
	}
}
