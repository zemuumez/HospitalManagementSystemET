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

func testLiveConsultations(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	t.Helper()
	ctx := context.Background()

	admin := actors[0]
	doctor := actors[1]
	otherDoctor := actors[2]
	patientUser := actors[3]
	receptionist := actors[4]

	srv := application.LiveConsultationService{Store: store, Now: time.Now}

	// 1. Authorization checks
	// Patient cannot create live consultation
	conDate := time.Now().Add(24 * time.Hour)
	_, err := srv.CreateLiveConsultation(ctx, patientUser, domain.LiveConsultationInput{
		DoctorID:          doctor.ID,
		PatientID:         patients[0].ID,
		ConsultationTitle: "Unauthorized Video Consultation",
		ConsultationDate:  &conDate,
		DurationMinutes:   30,
		PlatformType:      "zoom",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating live consultation, got %v", err)
	}

	// Patient cannot create live meeting
	meetDate := time.Now().Add(48 * time.Hour)
	_, err = srv.CreateLiveMeeting(ctx, patientUser, domain.LiveMeetingInput{
		Title:            "Unauthorized Staff Meeting",
		MeetingDate:      &meetDate,
		DurationMinutes:  45,
		PlatformType:     "meet",
		CandidateUserIDs: []string{doctor.ID},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating live meeting, got %v", err)
	}

	// Patient cannot read or update provider settings
	_, err = srv.ProviderSetting(ctx, patientUser)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient reading provider settings, got %v", err)
	}
	_, err = srv.UpdateProviderSetting(ctx, patientUser, domain.LiveProviderSettingInput{
		PlatformType: "zoom",
		APIKey:       "dummy-key",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient updating provider settings, got %v", err)
	}

	// 2. Validation checks
	// Validation error: missing title
	_, err = srv.CreateLiveConsultation(ctx, doctor, domain.LiveConsultationInput{
		PatientID:        patients[0].ID,
		ConsultationDate: &conDate,
		DurationMinutes:  30,
		PlatformType:     "zoom",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty consultation title, got %v", err)
	}

	// Validation error: invalid platform type
	_, err = srv.CreateLiveConsultation(ctx, doctor, domain.LiveConsultationInput{
		PatientID:         patients[0].ID,
		ConsultationTitle: "Checkup",
		ConsultationDate:  &conDate,
		DurationMinutes:   30,
		PlatformType:      "unsupported_vendor",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid platform type, got %v", err)
	}

	// 3. Live consultation lifecycle & Access Isolation
	consultation, err := srv.CreateLiveConsultation(ctx, doctor, domain.LiveConsultationInput{
		PatientID:         patients[0].ID,
		ConsultationTitle: "Cardiology Teleconsultation",
		ConsultationDate:  &conDate,
		DurationMinutes:   30,
		HostVideo:         true,
		ParticipantVideo:  true,
		Type:              "OPD",
		PlatformType:      "zoom",
		Description:       "Routine post-admission check",
	})
	if err != nil {
		t.Fatalf("failed to create live consultation: %v", err)
	}
	if consultation.Status != 0 || consultation.MeetingID == "" || consultation.Password == "" {
		t.Fatalf("unexpected consultation fields: %+v", consultation)
	}

	// Patient (whose user is linked to patients[0]) lists live consultations and finds it
	patientConsultations, count, err := srv.LiveConsultations(ctx, patientUser, 1, nil)
	if err != nil || count == 0 || len(patientConsultations) == 0 {
		t.Fatalf("failed to list live consultations for patient: %v, count: %d", err, count)
	}

	// Other doctor (not host or creator) fetching details gets ErrForbidden
	_, err = srv.LiveConsultation(ctx, otherDoctor, consultation.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for other doctor reading consultation, got %v", err)
	}

	// Doctor (host) fetches consultation details successfully
	conDoc, err := srv.LiveConsultation(ctx, doctor, consultation.ID)
	if err != nil || conDoc.ID != consultation.ID {
		t.Fatalf("doctor failed to read live consultation: %v", err)
	}

	// Doctor updates status to 1 (Finished)
	updatedCon, err := srv.UpdateLiveConsultationStatus(ctx, doctor, consultation.ID, 1)
	if err != nil || updatedCon.Status != 1 {
		t.Fatalf("failed to update consultation status: %v", err)
	}

	// 4. Live Meetings lifecycle
	meeting, err := srv.CreateLiveMeeting(ctx, admin, domain.LiveMeetingInput{
		Title:            "Clinical Governance Committee",
		MeetingDate:      &meetDate,
		DurationMinutes:  60,
		HostVideo:        true,
		ParticipantVideo: true,
		PlatformType:     "meet",
		Description:      "Quarterly antibiotic stewardship and safety review",
		CandidateUserIDs: []string{doctor.ID, receptionist.ID},
	})
	if err != nil {
		t.Fatalf("failed to create live meeting: %v", err)
	}
	if meeting.Status != 0 || meeting.MeetingID == "" {
		t.Fatalf("unexpected meeting fields: %+v", meeting)
	}

	// Doctor lists live meetings
	meetings, mCount, err := srv.LiveMeetings(ctx, doctor, 1, nil)
	if err != nil || mCount == 0 || len(meetings) == 0 {
		t.Fatalf("failed to list live meetings: %v, count: %d", err, mCount)
	}

	// Read meeting details and verify candidates
	meetDetails, err := srv.LiveMeeting(ctx, doctor, meeting.ID)
	if err != nil || len(meetDetails.CandidateUserIDs) != 2 {
		t.Fatalf("failed to read meeting details with candidates: %v, candidates: %v", err, meetDetails.CandidateUserIDs)
	}

	// Admin updates meeting status to 1 (Finished)
	updatedMeet, err := srv.UpdateLiveMeetingStatus(ctx, admin, meeting.ID, 1)
	if err != nil || updatedMeet.Status != 1 {
		t.Fatalf("failed to update meeting status: %v", err)
	}

	// 5. Provider Settings
	savedSetting, err := srv.UpdateProviderSetting(ctx, doctor, domain.LiveProviderSettingInput{
		PlatformType: "zoom",
		APIKey:       "test-zoom-api-key",
		APISecret:    "test-zoom-api-secret",
	})
	if err != nil {
		t.Fatalf("failed to save provider setting: %v", err)
	}
	if savedSetting.PlatformType != "zoom" || !savedSetting.HasAPIKey || !savedSetting.HasAPISecret {
		t.Fatalf("unexpected saved provider setting: %+v", savedSetting)
	}

	readSetting, err := srv.ProviderSetting(ctx, doctor)
	if err != nil || !readSetting.HasAPIKey || !readSetting.HasAPISecret {
		t.Fatalf("failed to read provider setting: %v, got: %+v", err, readSetting)
	}

	// 6. HTTP API Integration
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id": doctor.ID,
			},
			"session": map[string]any{
				"userId":    doctor.ID,
				"expiresAt": time.Now().Add(time.Hour),
			},
		})
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		LiveConsultation: srv,
		Actors:           store,
		AuthURL:          authSrv.URL,
		Origin:           "http://hospital.test",
		Client:           authSrv.Client(),
	}.Handler()

	// GET /v1/live-consultations
	req := httptest.NewRequest("GET", "/v1/live-consultations", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec := httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/live-consultations, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/live-meetings
	req = httptest.NewRequest("GET", "/v1/live-meetings", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/live-meetings, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/live-consultations/provider-settings
	req = httptest.NewRequest("GET", "/v1/live-consultations/provider-settings", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/live-consultations/provider-settings, got %d: %s", rec.Code, rec.Body.String())
	}
}
