package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/adapters/privatefiles"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testCMSSettings(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
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

	srv := application.CMSSettingsService{Store: store, Now: time.Now}
	t.Run("settings form is atomic and admin only", func(t *testing.T) {
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{"audit_name": "Before", "audit_email": "before@example.test"}); err != nil {
			t.Fatal(err)
		}
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{"audit_name": "Partial", "": "invalid"}); !errors.Is(err, domain.ErrValidation) {
			t.Fatal("invalid form", err)
		}
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{"audit_name": "One", " audit_name ": "Two"}); !errors.Is(err, domain.ErrValidation) {
			t.Fatal("normalized duplicate", err)
		}
		if _, err := db.Exec(ctx, `ALTER TABLE hospital_general_setting ADD CONSTRAINT audit_failure CHECK(value <> 'synthetic-db-rejection')`); err != nil {
			t.Fatal(err)
		}
		err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{"audit_name": "Partial", "z_audit_failure": "synthetic-db-rejection"})
		if _, dropErr := db.Exec(ctx, `ALTER TABLE hospital_general_setting DROP CONSTRAINT audit_failure`); dropErr != nil {
			t.Fatal(dropErr)
		}
		if err == nil {
			t.Fatal("expected database failure")
		}
		got, err := srv.GeneralSettings(ctx, admin)
		if err != nil || got["audit_name"] != "Before" {
			t.Fatal("partial settings persisted", got["audit_name"], err)
		}
		for _, role := range []string{"doctor", "patient", "nurse", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
			if err := srv.UpdateGeneralSettings(ctx, domain.Actor{ID: doctor.ID, Role: role}, map[string]string{"audit_name": "Denied"}); !errors.Is(err, domain.ErrForbidden) {
				t.Fatal("settings write exposed", role, err)
			}
		}
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{"audit_name": "After", "audit_email": "after@example.test"}); err != nil {
			t.Fatal(err)
		}
		got, err = srv.GeneralSettings(ctx, admin)
		if err != nil || got["audit_name"] != "After" || got["audit_email"] != "after@example.test" {
			t.Fatal("settings reload", err)
		}
	})

	// 1. Authorization checks
	// Patient cannot update general settings
	_, err := srv.UpdateGeneralSetting(ctx, patientUser, domain.GeneralSettingInput{
		Key:   "app_name",
		Value: "Hacked HMS",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient updating general settings, got %v", err)
	}

	// Doctor cannot update hospital schedule
	_, err = srv.UpdateHospitalSchedule(ctx, doctor, domain.HospitalScheduleDayInput{
		DayOfWeek: 1,
		StartTime: "07:00",
		EndTime:   "18:00",
		IsClosed:  false,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor updating schedule, got %v", err)
	}

	// Receptionist cannot update CMS content
	_, err = srv.UpdateFrontCMSSetting(ctx, receptionist, domain.FrontCMSSettingInput{
		Key:   "home_title",
		Value: "New Title",
		Type:  "home",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for receptionist updating CMS, got %v", err)
	}

	// 2. Validation checks
	// Validation error: schedule start >= end when not closed
	_, err = srv.UpdateHospitalSchedule(ctx, admin, domain.HospitalScheduleDayInput{
		DayOfWeek: 1,
		StartTime: "18:00",
		EndTime:   "08:00",
		IsClosed:  false,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for start >= end time, got %v", err)
	}

	// Validation error: invalid day of week
	_, err = srv.UpdateHospitalSchedule(ctx, admin, domain.HospitalScheduleDayInput{
		DayOfWeek: 9,
		StartTime: "08:00",
		EndTime:   "17:00",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for day_of_week 9, got %v", err)
	}

	// Validation error: empty general setting key
	_, err = srv.UpdateGeneralSetting(ctx, admin, domain.GeneralSettingInput{
		Key:   "",
		Value: "Value",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty setting key, got %v", err)
	}

	// Validation error: empty testimonial name
	_, err = srv.CreateTestimonial(ctx, admin, domain.CMSTestimonialInput{
		Name:        "",
		Description: "Great care",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty testimonial name, got %v", err)
	}

	// 3. General Settings CRUD
	// Read existing seeded settings
	genSettings, err := srv.GeneralSettings(ctx, admin)
	if err != nil {
		t.Fatalf("failed to read general settings: %v", err)
	}
	if genSettings["hospital_name"] != "Addis Ababa Central Hospital" {
		t.Fatalf("expected Addis Ababa Central Hospital, got %s", genSettings["hospital_name"])
	}

	// Update single setting
	updatedSetting, err := srv.UpdateGeneralSetting(ctx, admin, domain.GeneralSettingInput{
		Key:   "hospital_name",
		Value: "St. Paul Specialized Hospital",
	})
	if err != nil {
		t.Fatalf("failed to update general setting: %v", err)
	}
	if updatedSetting.Value != "St. Paul Specialized Hospital" {
		t.Fatalf("expected updated hospital name, got %s", updatedSetting.Value)
	}

	// Bulk update settings
	err = srv.UpdateGeneralSettings(ctx, admin, map[string]string{
		"hospital_phone": "+251911999888",
		"queue_theme":    "dark-emerald",
	})
	if err != nil {
		t.Fatalf("failed to bulk update settings: %v", err)
	}

	// Re-verify
	genSettings, err = srv.GeneralSettings(ctx, admin)
	if err != nil || genSettings["queue_theme"] != "dark-emerald" {
		t.Fatalf("expected queue_theme dark-emerald, got %v", genSettings["queue_theme"])
	}

	// 4. Hospital Schedules
	scheds, err := srv.HospitalSchedules(ctx, admin)
	if err != nil || len(scheds) != 7 {
		t.Fatalf("expected 7 schedule days, got %d (err: %v)", len(scheds), err)
	}

	// Update Sunday schedule to open
	sun, err := srv.UpdateHospitalSchedule(ctx, admin, domain.HospitalScheduleDayInput{
		DayOfWeek: 7,
		StartTime: "09:00",
		EndTime:   "13:00",
		IsClosed:  false,
	})
	if err != nil {
		t.Fatalf("failed to update sunday schedule: %v", err)
	}
	if sun.IsClosed || sun.StartTime != "09:00" {
		t.Fatalf("expected open sunday from 09:00, got %+v", sun)
	}

	// 5. Front CMS Settings
	cmsSettings, err := srv.FrontCMSSettings(ctx, admin, "home")
	if err != nil || len(cmsSettings) == 0 {
		t.Fatalf("failed to read home CMS settings: %v", err)
	}

	// Update CMS setting
	newCMS, err := srv.UpdateFrontCMSSetting(ctx, admin, domain.FrontCMSSettingInput{
		Key:   "home_title",
		Value: "Leading Clinical Excellence in East Africa",
		Type:  "home",
	})
	if err != nil {
		t.Fatalf("failed to update CMS setting: %v", err)
	}
	if newCMS.Value != "Leading Clinical Excellence in East Africa" {
		t.Fatalf("unexpected home title: %s", newCMS.Value)
	}

	// 6. Testimonials Lifecycle
	// Admin creates draft testimonial
	draftTestimonial, err := srv.CreateTestimonial(ctx, admin, domain.CMSTestimonialInput{
		Name:        "Tigist Alemu",
		Description: "The intensive care unit doctors were very attentive.",
		Position:    "Patient",
		Rating:      5,
		Status:      0, // Draft
	})
	if err != nil {
		t.Fatalf("failed to create draft testimonial: %v", err)
	}

	// Patient should NOT see draft testimonial
	patientView, err := srv.Testimonials(ctx, patientUser, nil)
	if err != nil {
		t.Fatalf("failed to list testimonials for patient: %v", err)
	}
	for _, tItem := range patientView {
		if tItem.ID == draftTestimonial.ID {
			t.Fatalf("patient should not see draft testimonial")
		}
	}

	// Admin publishes testimonial
	pubTestimonial, err := srv.UpdateTestimonial(ctx, admin, draftTestimonial.ID, domain.CMSTestimonialInput{
		Name:        "Tigist Alemu",
		Description: "The intensive care unit doctors were very attentive and kind.",
		Position:    "Patient",
		Rating:      5,
		Status:      1, // Published
	})
	if err != nil {
		t.Fatalf("failed to publish testimonial: %v", err)
	}
	if pubTestimonial.Status != 1 {
		t.Fatalf("expected published status 1, got %d", pubTestimonial.Status)
	}

	// Patient now sees it
	patientViewAfter, err := srv.Testimonials(ctx, patientUser, nil)
	if err != nil {
		t.Fatalf("failed to list testimonials for patient: %v", err)
	}
	found := false
	for _, tItem := range patientViewAfter {
		if tItem.ID == pubTestimonial.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected published testimonial to be visible to patient")
	}

	// Delete testimonial
	if err := srv.DeleteTestimonial(ctx, admin, pubTestimonial.ID); err != nil {
		t.Fatalf("failed to delete testimonial: %v", err)
	}

	// 7. HTTP API Integration
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id": admin.ID,
			},
			"session": map[string]any{
				"userId":    admin.ID,
				"expiresAt": time.Now().Add(time.Hour),
			},
		})
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		CMSSettings: srv,
		Actors:      store,
		AuthURL:     authSrv.URL,
		Origin:      "http://hospital.test",
		Client:      authSrv.Client(),
	}.Handler()

	// GET /v1/general-settings
	req := httptest.NewRequest("GET", "/v1/general-settings", nil)
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec := httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/general-settings, got %d: %s", rec.Code, rec.Body.String())
	}

	// POST /v1/general-settings
	bodyJSON, _ := json.Marshal(map[string]string{
		"hospital_name": "HTTP Test Hospital",
	})
	req = httptest.NewRequest("POST", "/v1/general-settings", bytes.NewReader(bodyJSON))
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for POST /v1/general-settings, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/hospital-schedules
	req = httptest.NewRequest("GET", "/v1/hospital-schedules", nil)
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/hospital-schedules, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/testimonials
	req = httptest.NewRequest("GET", "/v1/testimonials", nil)
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/testimonials, got %d: %s", rec.Code, rec.Body.String())
	}

	t.Run("provider secrets protection and lifecycle", func(t *testing.T) {
		// Set a secret
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"open_ai_key":   "sk-test-secret-value-12345",
			"stripe_secret": "sk_live_stripe_secret_67890",
		}); err != nil {
			t.Fatal("failed to set secrets", err)
		}

		// Read via GeneralSettings: secrets must be redacted
		readSettings, err := srv.GeneralSettings(ctx, admin)
		if err != nil {
			t.Fatal("failed to read settings", err)
		}
		if readSettings["open_ai_key"] != domain.SecretConfiguredPlaceholder {
			t.Fatalf("expected open_ai_key redacted to %s, got %s", domain.SecretConfiguredPlaceholder, readSettings["open_ai_key"])
		}
		if readSettings["stripe_secret"] != domain.SecretConfiguredPlaceholder {
			t.Fatalf("expected stripe_secret redacted, got %s", readSettings["stripe_secret"])
		}

		// Verify database stores the actual plaintext secret
		rawDB, err := store.GeneralSettings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rawDB["open_ai_key"] != "sk-test-secret-value-12345" {
			t.Fatalf("expected DB to store actual secret, got %s", rawDB["open_ai_key"])
		}

		// Save form with placeholder unchanged: secret must NOT be overwritten
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"open_ai_key":   domain.SecretConfiguredPlaceholder,
			"hospital_name": "Hospital Name With Secret Retained",
		}); err != nil {
			t.Fatal("failed to update with placeholder", err)
		}
		rawDB, err = store.GeneralSettings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rawDB["open_ai_key"] != "sk-test-secret-value-12345" {
			t.Fatalf("expected secret to be preserved when placeholder sent, got %s", rawDB["open_ai_key"])
		}

		// Replace secret with new string
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"open_ai_key": "sk-new-replaced-key-99999",
		}); err != nil {
			t.Fatal("failed to update secret", err)
		}
		rawDB, err = store.GeneralSettings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rawDB["open_ai_key"] != "sk-new-replaced-key-99999" {
			t.Fatalf("expected replaced secret in DB, got %s", rawDB["open_ai_key"])
		}

		// Explicit clearing with empty string
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"open_ai_key": "",
		}); err != nil {
			t.Fatal("failed to clear secret", err)
		}
		rawDB, err = store.GeneralSettings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rawDB["open_ai_key"] != "" {
			t.Fatalf("expected cleared secret in DB, got %s", rawDB["open_ai_key"])
		}
		readSettings, err = srv.GeneralSettings(ctx, admin)
		if err != nil {
			t.Fatal(err)
		}
		if readSettings["open_ai_key"] != "" {
			t.Fatalf("expected cleared secret to read empty string, got %s", readSettings["open_ai_key"])
		}
	})

	t.Run("batch hospital schedules update", func(t *testing.T) {
		batch := []domain.HospitalScheduleDayInput{
			{DayOfWeek: 1, StartTime: "08:30", EndTime: "17:30", IsClosed: false},
			{DayOfWeek: 2, StartTime: "08:30", EndTime: "17:30", IsClosed: false},
		}
		if err := srv.UpdateHospitalSchedules(ctx, admin, batch); err != nil {
			t.Fatal("failed batch update schedules", err)
		}
		scheds, err := srv.HospitalSchedules(ctx, admin)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range scheds {
			if s.DayOfWeek == 1 && (s.StartTime != "08:30" || s.EndTime != "17:30") {
				t.Fatalf("expected Monday schedule updated, got %+v", s)
			}
		}
	})

	t.Run("public attachment upload and unauthenticated content download", func(t *testing.T) {
		files, err := privatefiles.New(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer files.Close()
		attSrv := application.AttachmentsService{Store: store, Files: files}
		// Admin uploads public logo
		pngData := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82")
		att, err := attSrv.UploadPublic(ctx, admin, "hospital_logo.png", pngData)
		if err != nil {
			t.Fatal("failed to upload public logo", err)
		}
		if !att.IsPublic {
			t.Fatal("expected IsPublic to be true")
		}

		// Non-admin role cannot upload public attachment
		_, err = attSrv.UploadPublic(ctx, doctor, "fake_logo.png", pngData)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden for non-admin UploadPublic, got %v", err)
		}

		// Download with anonymous Actor{} succeeds because it is public
		meta, reader, err := attSrv.Download(ctx, domain.Actor{}, att.Token)
		if err != nil {
			t.Fatal("failed to download public attachment", err)
		}
		reader.Close()
		if meta.FileName != "hospital_logo.png" {
			t.Fatalf("expected hospital_logo.png, got %s", meta.FileName)
		}
	})
}
