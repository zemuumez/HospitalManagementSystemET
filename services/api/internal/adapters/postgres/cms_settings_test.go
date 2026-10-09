package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	t.Run("S2: hospital schedule batch atomicity and duplicate rejection", func(t *testing.T) {
		// Set baseline for day 1 and 2
		if err := srv.UpdateHospitalSchedules(ctx, admin, []domain.HospitalScheduleDayInput{
			{DayOfWeek: 1, StartTime: "08:00", EndTime: "17:00", IsClosed: false},
			{DayOfWeek: 2, StartTime: "08:00", EndTime: "17:00", IsClosed: false},
		}); err != nil {
			t.Fatal(err)
		}

		// 1. Invalid trailing item: item 1 is valid, item 2 has invalid time (EndTime <= StartTime)
		invalidTrailing := []domain.HospitalScheduleDayInput{
			{DayOfWeek: 1, StartTime: "09:00", EndTime: "18:00", IsClosed: false},
			{DayOfWeek: 2, StartTime: "18:00", EndTime: "09:00", IsClosed: false}, // invalid!
		}
		if err := srv.UpdateHospitalSchedules(ctx, admin, invalidTrailing); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for invalid trailing schedule item, got %v", err)
		}
		// Assert day 1 was NOT mutated
		scheds, err := srv.HospitalSchedules(ctx, admin)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range scheds {
			if s.DayOfWeek == 1 && s.StartTime != "08:00" {
				t.Fatalf("expected day 1 to remain 08:00 after trailing validation failure, got %s", s.StartTime)
			}
		}

		// 2. Duplicate day in batch
		dupBatch := []domain.HospitalScheduleDayInput{
			{DayOfWeek: 3, StartTime: "08:00", EndTime: "17:00", IsClosed: false},
			{DayOfWeek: 3, StartTime: "09:00", EndTime: "18:00", IsClosed: false},
		}
		if err := srv.UpdateHospitalSchedules(ctx, admin, dupBatch); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for duplicate day entry, got %v", err)
		}

		// 3. Injected late database failure in aggregate transaction
		if _, err := db.Exec(ctx, `ALTER TABLE hospital_schedule_day ADD CONSTRAINT schedule_test_failure CHECK(start_time <> '09:45')`); err != nil {
			t.Fatal(err)
		}
		err = srv.UpdateHospitalSchedules(ctx, admin, []domain.HospitalScheduleDayInput{
			{DayOfWeek: 1, StartTime: "09:15", EndTime: "17:00", IsClosed: false},
			{DayOfWeek: 2, StartTime: "09:45", EndTime: "17:00", IsClosed: false}, // violates constraint!
		})
		_, _ = db.Exec(ctx, `ALTER TABLE hospital_schedule_day DROP CONSTRAINT IF EXISTS schedule_test_failure`)
		if err == nil {
			t.Fatal("expected DB error on schedule save")
		}
		// Verify day 1 was completely rolled back
		scheds, err = srv.HospitalSchedules(ctx, admin)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range scheds {
			if s.DayOfWeek == 1 && s.StartTime != "08:00" {
				t.Fatalf("expected day 1 to remain 08:00 after late DB rollback, got %s", s.StartTime)
			}
		}
	})

	t.Run("S2: front CMS batch atomicity and duplicate rejection", func(t *testing.T) {
		// Set baseline
		if err := srv.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
			{Key: "home_atomic_test", Value: "Initial Value", Type: "home"},
		}); err != nil {
			t.Fatal(err)
		}

		// 1. Invalid trailing item (empty key)
		invalidTrailing := []domain.FrontCMSSettingInput{
			{Key: "home_atomic_test", Value: "New Mutated Value", Type: "home"},
			{Key: "", Value: "Blank Key", Type: "home"},
		}
		if err := srv.UpdateFrontCMSSettings(ctx, admin, invalidTrailing); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for invalid trailing CMS item, got %v", err)
		}
		// Verify baseline unchanged
		cmsList, err := srv.FrontCMSSettings(ctx, admin, "home")
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range cmsList {
			if c.Key == "home_atomic_test" && c.Value != "Initial Value" {
				t.Fatalf("expected home_atomic_test to remain 'Initial Value', got '%s'", c.Value)
			}
		}

		// 2. Duplicate key rejection
		dupBatch := []domain.FrontCMSSettingInput{
			{Key: "home_dup_key", Value: "Val 1", Type: "home"},
			{Key: "home_dup_key", Value: "Val 2", Type: "home"},
		}
		if err := srv.UpdateFrontCMSSettings(ctx, admin, dupBatch); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for duplicate key in CMS batch, got %v", err)
		}
	})

	t.Run("S3: general settings and schedule server-side validation", func(t *testing.T) {
		// Direct blank required key must return ErrValidation
		for _, reqKey := range []string{"app_name", "company_name", "hospital_email", "hospital_phone", "hospital_address", "current_currency", "about_us"} {
			err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
				reqKey: "",
			})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for blank required key %q, got %v", reqKey, err)
			}
		}

		// Invalid email format
		err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"hospital_email": "not-an-email-address",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for malformed email, got %v", err)
		}

		// Invalid currency format (must be 3 alpha chars)
		err = srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"current_currency": "INVALID_LONG",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for malformed currency, got %v", err)
		}

		// Invalid schedule day input time format (not HH:MM)
		invDay := domain.HospitalScheduleDayInput{
			DayOfWeek: 1,
			StartTime: "25:00", // invalid!
			EndTime:   "17:00",
			IsClosed:  false,
		}
		if err := invDay.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for invalid schedule time, got %v", err)
		}

		// Schedule day with StartTime >= EndTime
		invDay2 := domain.HospitalScheduleDayInput{
			DayOfWeek: 1,
			StartTime: "17:00",
			EndTime:   "08:00",
			IsClosed:  false,
		}
		if err := invDay2.Validate(); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for StartTime >= EndTime, got %v", err)
		}

		// Closed day with valid optional times or blank times must pass validation
		closedDay := domain.HospitalScheduleDayInput{
			DayOfWeek: 7,
			StartTime: "",
			EndTime:   "",
			IsClosed:  true,
		}
		if err := closedDay.Validate(); err != nil {
			t.Fatalf("expected valid closed day, got %v", err)
		}

		// HTTP API endpoint direct request test: malformed request returns 422
		authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": admin.ID},
				"session": map[string]any{"userId": admin.ID, "expiresAt": time.Now().Add(time.Hour)},
			})
		}))
		defer authSrv.Close()

		handler := httpapi.Server{
			CMSSettings: srv,
			Actors:      store,
			AuthURL:     authSrv.URL,
			Origin:      "http://hospital.test",
			Client:      authSrv.Client(),
		}.Handler()

		malformedBody, _ := json.Marshal(map[string]string{
			"app_name": "", // blank required field!
		})
		req := httptest.NewRequest("POST", "/v1/general-settings", bytes.NewReader(malformedBody))
		req.Header.Set("Cookie", "session=admin-token")
		req.Header.Set("Origin", "http://hospital.test")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected HTTP 422 Unprocessable Entity for direct malformed request, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("S4: attachment retirement on settings update & 404 on retired token", func(t *testing.T) {
		files, err := privatefiles.New(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer files.Close()
		attSrv := application.AttachmentsService{Store: store, Files: files}
		cmsWithFiles := application.CMSSettingsService{Store: store, Files: files, Now: time.Now}

		// 1. Upload initial logo A
		pngDataA := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82")
		attA, err := attSrv.UploadPublic(ctx, admin, "logo_a.png", pngDataA)
		if err != nil {
			t.Fatal("failed to upload logo A", err)
		}
		urlA := "/api/hms/attachments/" + attA.Token + "/content"

		// Save settings referencing urlA
		if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
			"app_logo": urlA,
		}); err != nil {
			t.Fatal("failed to set logo A", err)
		}

		// Verify token A is downloadable
		_, readerA, err := attSrv.Download(ctx, domain.Actor{}, attA.Token)
		if err != nil {
			t.Fatal("expected token A to be downloadable while referenced", err)
		}
		readerA.Close()

		// 2. Upload replacement logo B
		attB, err := attSrv.UploadPublic(ctx, admin, "logo_b.png", pngDataA)
		if err != nil {
			t.Fatal("failed to upload logo B", err)
		}
		urlB := "/api/hms/attachments/" + attB.Token + "/content"

		// Save settings referencing urlB (replacing urlA)
		if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
			"app_logo": urlB,
		}); err != nil {
			t.Fatal("failed to set logo B", err)
		}

		// 3. Verify unreferenced token A has been retired (returns ErrNotFound)
		_, _, err = attSrv.Download(ctx, domain.Actor{}, attA.Token)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for retired token A, got %v", err)
		}

		// Verify newly referenced token B remains downloadable
		_, readerB, err := attSrv.Download(ctx, domain.Actor{}, attB.Token)
		if err != nil {
			t.Fatalf("expected token B to be downloadable, got %v", err)
		}
		readerB.Close()

		// 4. Pending upload surviving unrelated settings save
		attC, err := attSrv.UploadPublic(ctx, admin, "pending_logo.png", pngDataA)
		if err != nil {
			t.Fatal("failed to upload pending logo C", err)
		}
		// Unrelated settings update: updating company_name without referencing attC
		if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
			"company_name": "Hospital Care Ltd",
		}); err != nil {
			t.Fatal("failed unrelated settings update", err)
		}
		// Assert pending logo C was NOT deleted
		_, readerC, err := attSrv.Download(ctx, domain.Actor{}, attC.Token)
		if err != nil {
			t.Fatalf("pending upload C should survive unrelated save, but got err: %v", err)
		}
		readerC.Close()

		// 5. Direct retirement rejected while referenced (domain.ErrInUse / HTTP 409)
		// attB is currently referenced as app_logo
		_, err = store.DeleteAttachment(ctx, admin, attB.Token)
		if !errors.Is(err, domain.ErrInUse) {
			t.Fatalf("expected ErrInUse on direct deletion of in-use attachment B, got %v", err)
		}

		// Also test HTTP endpoint rejection
		authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user":    map[string]any{"id": admin.ID},
				"session": map[string]any{"userId": admin.ID, "expiresAt": time.Now().Add(time.Hour)},
			})
		}))
		defer authSrv.Close()

		httpHandler := httpapi.Server{
			CMSSettings: cmsWithFiles,
			Attachments: attSrv,
			Actors:      store,
			AuthURL:     authSrv.URL,
			Origin:      "http://hospital.test",
			Client:      authSrv.Client(),
		}.Handler()

		delReq := httptest.NewRequest("DELETE", "/v1/attachments/"+attB.Token, nil)
		delReq.Header.Set("Cookie", "session=admin-token")
		delReq.Header.Set("Origin", "http://hospital.test")
		delRec := httptest.NewRecorder()
		httpHandler.ServeHTTP(delRec, delReq)
		if delRec.Code != http.StatusConflict {
			t.Fatalf("expected HTTP 409 Conflict for DELETE in-use attachment, got %d: %s", delRec.Code, delRec.Body.String())
		}

		// 6. Shared reference preserved across general & CMS
		attD, err := attSrv.UploadPublic(ctx, admin, "shared_logo.png", pngDataA)
		if err != nil {
			t.Fatal("failed to upload shared logo D", err)
		}
		urlD := "/api/hms/attachments/" + attD.Token + "/content"
		// Reference in general settings
		if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
			"app_logo": urlD,
		}); err != nil {
			t.Fatal("failed to set app_logo D", err)
		}
		// Also reference in front CMS settings
		if err := cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
			{Key: "home_banner", Value: urlD, Type: "home"},
		}); err != nil {
			t.Fatal("failed to set home_banner D", err)
		}

		// Now update front CMS replacing home_banner with a new valid banner E
		attE, err := attSrv.UploadPublic(ctx, admin, "replacement_banner.png", pngDataA)
		if err != nil {
			t.Fatal("failed to upload replacement banner E", err)
		}
		urlE := "/api/hms/attachments/" + attE.Token + "/content"
		if err := cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
			{Key: "home_banner", Value: urlE, Type: "home"},
		}); err != nil {
			t.Fatal("failed to displace home_banner D", err)
		}

		// Token D was displaced from CMS, BUT it is still referenced in general settings (app_logo)
		// Therefore it MUST NOT be retired!
		_, readerD, err := attSrv.Download(ctx, domain.Actor{}, attD.Token)
		if err != nil {
			t.Fatalf("shared reference D must be preserved because general setting still references it, got err: %v", err)
		}
		readerD.Close()

		// 7. Bounded abandoned cleanup with olderThan threshold
		attF, err := attSrv.UploadPublic(ctx, admin, "abandoned.png", pngDataA)
		if err != nil {
			t.Fatal("failed to upload abandoned attachment F", err)
		}
		// Immediate cleanup: attF is brand new (<24h), so it must NOT be cleaned up
		if err := cmsWithFiles.CleanupAbandonedAttachments(ctx, admin, 24*time.Hour); err != nil {
			t.Fatal("cleanup error", err)
		}
		_, readerF, err := attSrv.Download(ctx, domain.Actor{}, attF.Token)
		if err != nil {
			t.Fatalf("new upload F (<24h) should NOT be cleaned up: %v", err)
		}
		readerF.Close()

		// Artificially age attF by 48 hours in DB
		if _, err := db.Exec(ctx, `UPDATE secure_attachment SET created_at = clock_timestamp() - interval '48 hours' WHERE token = $1`, attF.Token); err != nil {
			t.Fatal("failed to age attachment F", err)
		}
		// Also age referenced attD by 48 hours to ensure referenced old files are NOT cleaned up
		if _, err := db.Exec(ctx, `UPDATE secure_attachment SET created_at = clock_timestamp() - interval '48 hours' WHERE token = $1`, attD.Token); err != nil {
			t.Fatal("failed to age attachment D", err)
		}

		// Run bounded cleanup with 24h threshold
		if err := cmsWithFiles.CleanupAbandonedAttachments(ctx, admin, 24*time.Hour); err != nil {
			t.Fatal("cleanup error", err)
		}
		// Assert: attF is now deleted (ErrNotFound)
		_, _, err = attSrv.Download(ctx, domain.Actor{}, attF.Token)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("aged unreferenced attachment F should be cleaned up, got %v", err)
		}
		// Assert: aged referenced attD remains untouched
		_, readerD2, err := attSrv.Download(ctx, domain.Actor{}, attD.Token)
		if err != nil {
			t.Fatalf("aged referenced attachment D should NOT be cleaned up: %v", err)
		}
		readerD2.Close()

		// 8. R1a deterministic concurrency probe holding lock on T
		{
			attProbe, err := attSrv.UploadPublic(ctx, admin, "probe.png", pngDataA)
			if err != nil {
				t.Fatal("failed to upload probe attachment", err)
			}
			urlProbe := "/api/hms/attachments/" + attProbe.Token + "/content"
			// Save initial logo referencing attProbe
			if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": urlProbe,
			}); err != nil {
				t.Fatal("failed to set probe logo", err)
			}

			// In a separate test connection, hold row lock on attProbe
			holdConn, err := db.Acquire(ctx)
			if err != nil {
				t.Fatal("failed to acquire test connection", err)
			}
			holdTx, err := holdConn.Begin(ctx)
			if err != nil {
				holdConn.Release()
				t.Fatal("failed to begin hold tx", err)
			}
			var probeRowID string
			if err := holdTx.QueryRow(ctx, `SELECT id FROM secure_attachment WHERE token = $1 FOR UPDATE`, attProbe.Token).Scan(&probeRowID); err != nil {
				_ = holdTx.Rollback(ctx)
				holdConn.Release()
				t.Fatal("failed to lock probe attachment", err)
			}

			// Goroutine A: UpdateGeneralSettings displaces attProbe (app_logo: "")
			errChA := make(chan error, 1)
			go func() {
				errChA <- cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
					"app_logo": "",
				})
			}()

			// Poll until Goroutine A is blocked on the held attachment row lock
			for i := 0; i < 50; i++ {
				var count int
				_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM pg_locks WHERE NOT granted`).Scan(&count)
				if count > 0 {
					break
				}
				time.Sleep(20 * time.Millisecond)
			}

			// Goroutine B: UpdateFrontCMSSettings binds attProbe (review_banner: urlProbe)
			// Because binding participates in the lock protocol, it also blocks on attProbe!
			errChB := make(chan error, 1)
			go func() {
				errChB <- cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
					{Key: "review_banner", Value: urlProbe, Type: "review"},
				})
			}()

			// Give Goroutine B a moment to queue behind the lock
			time.Sleep(50 * time.Millisecond)

			// Release hold lock
			_ = holdTx.Rollback(ctx)
			holdConn.Release()

			errA := <-errChA
			errB := <-errChB

			if errA != nil {
				t.Fatalf("UpdateGeneralSettings failed: %v", errA)
			}

			// Validate invariant: if CMS contains attProbe, attProbe MUST be downloadable (never retired/deleted)
			cmsItems, err := store.FrontCMSSettings(ctx, "review")
			if err != nil {
				t.Fatal("failed to query CMS settings", err)
			}
			var cmsHasProbe bool
			for _, item := range cmsItems {
				if item.Key == "review_banner" && strings.Contains(item.Value, attProbe.Token) {
					cmsHasProbe = true
				}
			}
			if cmsHasProbe {
				// Binding won: attProbe must be downloadable!
				_, reader, err := attSrv.Download(ctx, domain.Actor{}, attProbe.Token)
				if err != nil {
					t.Fatalf("successfully bound CMS image was retired: %v", err)
				}
				reader.Close()
			} else {
				// Retirement won: binding must have been rejected
				if errB == nil {
					t.Fatal("expected binding to fail when retirement won")
				}
				if !errors.Is(errB, domain.ErrNotFound) {
					t.Fatalf("expected ErrNotFound when binding retired token, got %v", errB)
				}
			}
		}

		// 9. Deterministic Ordering Outcome A: Binding wins -> retirement preserves reference
		{
			attWin, err := attSrv.UploadPublic(ctx, admin, "win.png", pngDataA)
			if err != nil {
				t.Fatal("failed to upload win attachment", err)
			}
			urlWin := "/api/hms/attachments/" + attWin.Token + "/content"
			// Save in general settings
			if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": urlWin,
			}); err != nil {
				t.Fatal(err)
			}
			// Binding wins: bind in CMS first
			if err := cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
				{Key: "home_banner", Value: urlWin, Type: "home"},
			}); err != nil {
				t.Fatal(err)
			}
			// Now displace from general settings
			if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": "",
			}); err != nil {
				t.Fatal(err)
			}
			// Assert: attWin is preserved and downloadable because CMS references it
			_, readerWin, err := attSrv.Download(ctx, domain.Actor{}, attWin.Token)
			if err != nil {
				t.Fatalf("expected win token to be preserved, got error: %v", err)
			}
			readerWin.Close()
		}

		// 10. Deterministic Ordering Outcome B: Retirement wins -> binding rejects missing token
		{
			attRet, err := attSrv.UploadPublic(ctx, admin, "ret.png", pngDataA)
			if err != nil {
				t.Fatal("failed to upload ret attachment", err)
			}
			urlRet := "/api/hms/attachments/" + attRet.Token + "/content"
			// Save in general settings
			if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": urlRet,
			}); err != nil {
				t.Fatal(err)
			}
			// Retirement wins: displace and retire from general settings
			if err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": "",
			}); err != nil {
				t.Fatal(err)
			}
			// Verify token was deleted
			_, _, err = attSrv.Download(ctx, domain.Actor{}, attRet.Token)
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("expected ErrNotFound for retired token, got %v", err)
			}
			// Now attempt to bind retired token in CMS
			err = cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
				{Key: "home_banner", Value: urlRet, Type: "home"},
			})
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("expected ErrNotFound when binding already-retired token, got %v", err)
			}
			// Verify broken URL was NOT persisted in CMS
			cmsItems, _ := store.FrontCMSSettings(ctx, "home")
			for _, item := range cmsItems {
				if item.Key == "home_banner" && strings.Contains(item.Value, attRet.Token) {
					t.Fatalf("broken URL with retired token was persisted in CMS: %s", item.Value)
				}
			}
		}

		// 11. Binding clinical attachment is rejected (ErrValidation)
		{
			clinicalID := "11111111-1111-1111-1111-111111111111"
			clinicalToken := "c1111111111111111111111111111111"
			_, err := db.Exec(ctx, `
				INSERT INTO secure_attachment (id, token, file_name, mime_type, file_size_bytes, storage_path, sha256_hash, uploader_id, is_public)
				VALUES ($1, $2, 'clinical.pdf', 'application/pdf', 1024, 'clinical.pdf', $3, $4, false)
				ON CONFLICT (id) DO NOTHING
			`, clinicalID, clinicalToken, strings.Repeat("a", 64), admin.ID)
			if err != nil {
				t.Fatalf("failed to insert clinical fixture: %v", err)
			}

			// Attempt to bind clinical token to app_logo
			err = cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": "/v1/attachments/" + clinicalToken + "/content",
			})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation when binding private/clinical attachment, got %v", err)
			}
		}

		// 12. R1b: Operational cleanup entry point with aged, pending, referenced, and clinical fixtures
		{
			// Fixture 1: Aged unreferenced public asset (>24h)
			attAgedUnref, err := attSrv.UploadPublic(ctx, admin, "aged_unref.png", pngDataA)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = db.Exec(ctx, `UPDATE secure_attachment SET created_at = clock_timestamp() - interval '48 hours' WHERE token = $1`, attAgedUnref.Token)

			// Fixture 2: Recent pending public asset (<24h)
			attRecentPending, err := attSrv.UploadPublic(ctx, admin, "recent_pending.png", pngDataA)
			if err != nil {
				t.Fatal(err)
			}

			// Fixture 3: Aged referenced public asset (>24h, in CMS home_banner)
			attAgedRef, err := attSrv.UploadPublic(ctx, admin, "aged_ref.png", pngDataA)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = db.Exec(ctx, `UPDATE secure_attachment SET created_at = clock_timestamp() - interval '48 hours' WHERE token = $1`, attAgedRef.Token)
			if err := cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
				{Key: "home_banner", Value: "/v1/attachments/" + attAgedRef.Token + "/content", Type: "home"},
			}); err != nil {
				t.Fatal(err)
			}

			// Fixture 4: Aged clinical private asset (>24h, is_public=false)
			clinToken := "c4444444444444444444444444444444"
			clinID := "44444444-4444-4444-4444-444444444444"
			_, err = db.Exec(ctx, `
				INSERT INTO secure_attachment (id, token, file_name, mime_type, file_size_bytes, storage_path, sha256_hash, uploader_id, is_public, created_at)
				VALUES ($1, $2, 'clinical_lab.pdf', 'application/pdf', 512, 'clinical_lab.pdf', $3, $4, false, clock_timestamp() - interval '48 hours')
				ON CONFLICT (id) DO NOTHING
			`, clinID, clinToken, strings.Repeat("b", 64), admin.ID)
			if err != nil {
				t.Fatalf("failed to insert clinical fixture 4: %v", err)
			}

			// Test operational entry point: RunOperationalAttachmentCleanup
			cleanedCount, err := cmsWithFiles.RunOperationalAttachmentCleanup(ctx, admin, 24*time.Hour)
			if err != nil {
				t.Fatalf("operational cleanup failed: %v", err)
			}
			if cleanedCount < 1 {
				t.Fatalf("expected at least 1 cleaned attachment, got %d", cleanedCount)
			}

			// Assert Fixture 1 (aged unreferenced) is deleted
			_, _, err = attSrv.Download(ctx, domain.Actor{}, attAgedUnref.Token)
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("expected Fixture 1 to be deleted, got %v", err)
			}

			// Assert Fixture 2 (recent pending) is preserved
			_, readerRecent, err := attSrv.Download(ctx, domain.Actor{}, attRecentPending.Token)
			if err != nil {
				t.Fatalf("Fixture 2 (recent pending) must be preserved: %v", err)
			}
			readerRecent.Close()

			// Assert Fixture 3 (aged referenced) is preserved
			_, readerRef, err := attSrv.Download(ctx, domain.Actor{}, attAgedRef.Token)
			if err != nil {
				t.Fatalf("Fixture 3 (aged referenced) must be preserved: %v", err)
			}
			readerRef.Close()

			// Assert Fixture 4 (aged clinical) is preserved
			var clinExists bool
			_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token = $1)`, clinToken).Scan(&clinExists)
			if !clinExists {
				t.Fatal("Fixture 4 (clinical) must NEVER be touched by abandoned cleanup")
			}

			// Test HTTP operational maintenance endpoint POST /v1/attachments/cleanup
			authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"user":    map[string]any{"id": admin.ID},
					"session": map[string]any{"userId": admin.ID, "expiresAt": time.Now().Add(time.Hour)},
				})
			}))
			defer authSrv.Close()

			httpHandler := httpapi.Server{
				CMSSettings: cmsWithFiles,
				Attachments: attSrv,
				Actors:      store,
				AuthURL:     authSrv.URL,
				Origin:      "http://hospital.test",
				Client:      authSrv.Client(),
			}.Handler()

			// Admin call to POST /v1/attachments/cleanup
			cleanReq := httptest.NewRequest("POST", "/v1/attachments/cleanup", strings.NewReader(`{"olderThanSeconds":86400}`))
			cleanReq.Header.Set("Cookie", "session=admin-token")
			cleanReq.Header.Set("Origin", "http://hospital.test")
			cleanReq.Header.Set("Content-Type", "application/json")
			cleanRec := httptest.NewRecorder()
			httpHandler.ServeHTTP(cleanRec, cleanReq)
			if cleanRec.Code != http.StatusOK {
				t.Fatalf("expected HTTP 200 from POST /v1/attachments/cleanup, got %d: %s", cleanRec.Code, cleanRec.Body.String())
			}

			// Doctor (non-admin) call to POST /v1/attachments/cleanup -> 403 Forbidden
			docAuthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"user":    map[string]any{"id": doctor.ID},
					"session": map[string]any{"userId": doctor.ID, "expiresAt": time.Now().Add(time.Hour)},
				})
			}))
			defer docAuthSrv.Close()

			docHttpHandler := httpapi.Server{
				CMSSettings: cmsWithFiles,
				Attachments: attSrv,
				Actors:      store,
				AuthURL:     docAuthSrv.URL,
				Origin:      "http://hospital.test",
				Client:      docAuthSrv.Client(),
			}.Handler()

			docReq := httptest.NewRequest("POST", "/v1/attachments/cleanup", nil)
			docReq.Header.Set("Cookie", "session=doctor-token")
			docReq.Header.Set("Origin", "http://hospital.test")
			docRec := httptest.NewRecorder()
			docHttpHandler.ServeHTTP(docRec, docReq)
			if docRec.Code != http.StatusForbidden {
				t.Fatalf("expected HTTP 403 Forbidden for non-admin cleanup, got %d: %s", docRec.Code, docRec.Body.String())
			}
		}

		// 13. L1: Ordinary setting content, provider keys, and remote URLs containing 32-hex strings are NOT mistaken for attachments
		{
			hex32 := "0123456789abcdef0123456789abcdef"
			// A. UpdateGeneralSettings with ordinary text containing 32-hex
			ordinaryText := "Reference " + hex32 + " for hospital information"
			err := cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"about_us":            ordinaryText,
				"hospital_address":    "Central Street " + hex32,
				"stripe_key":          "pk_test_" + hex32,
				"phonepe_merchant_id": hex32,
				"facebook_url":        "https://cdn.example.com/assets/logo_" + hex32 + ".png",
				"twitter_url":         "https://external.example.com/attachments/" + hex32 + "/content",
			})
			if err != nil {
				t.Fatalf("expected successful save of ordinary text, provider keys, and external URLs containing 32-hex, got: %v", err)
			}

			// Verify values were persisted unchanged
			rawSettings, err := store.GeneralSettings(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if rawSettings["about_us"] != ordinaryText {
				t.Fatalf("expected about_us to match, got %s", rawSettings["about_us"])
			}
			if rawSettings["stripe_key"] != "pk_test_"+hex32 {
				t.Fatalf("expected stripe_key to match, got %s", rawSettings["stripe_key"])
			}
			if rawSettings["phonepe_merchant_id"] != hex32 {
				t.Fatalf("expected phonepe_merchant_id to match, got %s", rawSettings["phonepe_merchant_id"])
			}
			if rawSettings["facebook_url"] != "https://cdn.example.com/assets/logo_"+hex32+".png" {
				t.Fatalf("expected facebook_url to match, got %s", rawSettings["facebook_url"])
			}
			if rawSettings["twitter_url"] != "https://external.example.com/attachments/"+hex32+"/content" {
				t.Fatalf("expected twitter_url to match, got %s", rawSettings["twitter_url"])
			}

			// B. Front CMS with ordinary text containing 32-hex
			err = cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, []domain.FrontCMSSettingInput{
				{Key: "home_page_title", Value: "Welcome to Hospital (" + hex32 + ")", Type: "home"},
				{Key: "home_page_description", Value: "https://photos.external.com/banner_" + hex32 + ".jpg", Type: "home"},
			})
			if err != nil {
				t.Fatalf("expected front CMS to save ordinary text and remote URLs containing 32-hex, got: %v", err)
			}

			// C. Dedicated attachment field (app_logo) with genuine attachment URL binds and validates properly
			validAtt, err := attSrv.UploadPublic(ctx, admin, "valid_logo.png", pngDataA)
			if err != nil {
				t.Fatal(err)
			}
			validURL := "/v1/attachments/" + validAtt.Token + "/content"
			err = cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": validURL,
			})
			if err != nil {
				t.Fatalf("expected valid attachment URL to bind successfully, got: %v", err)
			}

			// D. Dedicated attachment field (app_logo) with nonexistent attachment URL fails with ErrNotFound
			nonexistentURL := "/v1/attachments/ffffffffffffffffffffffffffffffff/content"
			err = cmsWithFiles.UpdateGeneralSettings(ctx, admin, map[string]string{
				"app_logo": nonexistentURL,
			})
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("expected ErrNotFound for nonexistent attachment URL, got: %v", err)
			}
			// Verify nonexistent URL was NOT persisted
			rawSettings, _ = store.GeneralSettings(ctx)
			if rawSettings["app_logo"] == nonexistentURL {
				t.Fatal("broken attachment URL was improperly persisted")
			}
		}

		// 14. L3: More-than-100 fixtures progress across cleanup runs without starvation
		{
			// Seed 105 aged public non-clinical attachments in ascending order
			var bulkTokens []string
			for i := 0; i < 105; i++ {
				tok := fmt.Sprintf("bulk_%04d_%s", i, strings.Repeat("0", 22))
				id := fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1000)
				storagePath, err := files.Put(ctx, pngDataA)
				if err != nil {
					t.Fatal(err)
				}

				_, err = db.Exec(ctx, `
					INSERT INTO secure_attachment (id, token, file_name, mime_type, file_size_bytes, storage_path, sha256_hash, uploader_id, is_public, created_at)
					VALUES ($1, $2, $3, 'image/png', 68, $4, $5, $6, true, clock_timestamp() - interval '48 hours')
					ON CONFLICT (id) DO UPDATE SET token = EXCLUDED.token, storage_path = EXCLUDED.storage_path, created_at = EXCLUDED.created_at
				`, id, tok, fmt.Sprintf("bulk_%04d.png", i), storagePath, strings.Repeat("f", 64), admin.ID)
				if err != nil {
					t.Fatalf("failed to insert bulk fixture %d: %v", i, err)
				}
				bulkTokens = append(bulkTokens, tok)
			}

			// Reference the first 100 attachments in front_cms_setting
			cmsInputs := make([]domain.FrontCMSSettingInput, 100)
			for i := 0; i < 100; i++ {
				cmsInputs[i] = domain.FrontCMSSettingInput{
					Key:   fmt.Sprintf("bulk_cms_ref_%04d", i),
					Value: "/v1/attachments/" + bulkTokens[i] + "/content",
					Type:  "bulk",
				}
			}
			if err := cmsWithFiles.UpdateFrontCMSSettings(ctx, admin, cmsInputs); err != nil {
				t.Fatalf("failed to reference first 100 bulk attachments: %v", err)
			}

			// Tokens bulkTokens[100..104] (5 attachments) are unreferenced orphans.
			// Run operational cleanup: with candidate pre-filtering, the referenced 100 tokens
			// do NOT monopolize the batch. The 5 unreferenced orphans must be cleaned up!
			cleanedCount, err := cmsWithFiles.RunOperationalAttachmentCleanup(ctx, admin, 24*time.Hour)
			if err != nil {
				t.Fatalf("operational cleanup failed on >100 fixture set: %v", err)
			}
			if cleanedCount < 5 {
				t.Fatalf("expected at least 5 unreferenced orphans to be cleaned, got %d", cleanedCount)
			}

			// Verify the 5 unreferenced orphans were deleted
			for i := 100; i < 105; i++ {
				var exists bool
				_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token = $1)`, bulkTokens[i]).Scan(&exists)
				if exists {
					t.Fatalf("orphan %s was NOT deleted; referenced assets blocked cleanup", bulkTokens[i])
				}
			}

			// Verify all 100 referenced assets remain intact in secure_attachment
			for i := 0; i < 100; i++ {
				var exists bool
				_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token = $1)`, bulkTokens[i]).Scan(&exists)
				if !exists {
					t.Fatalf("referenced asset %s was mistakenly deleted", bulkTokens[i])
				}
			}
		}

		// 15. L2: Real Worker CLI execution against isolated schema & missing-actor failure
		{
			workerBinPath, err := filepath.Abs("../../../bin/worker.exe")
			if err != nil || !func() bool { _, e := os.Stat(workerBinPath); return e == nil }() {
				// Build worker binary if needed
				buildCmd := exec.Command("go", "build", "-o", "../../../bin/worker.exe", "../../../cmd/worker")
				if out, err := buildCmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to build worker binary for CLI test: %v, out: %s", err, string(out))
				}
				workerBinPath, _ = filepath.Abs("../../../bin/worker.exe")
			}

			dbURL := os.Getenv("DATABASE_URL")
			if dbURL == "" {
				t.Skip("DATABASE_URL not set; skipping worker CLI exec test")
			}

			// Case A: Successful worker execution with active admin in staff_access
			orphanTok := "cli_orphan_0123456789abcdef012345"
			orphanID := "99999999-9999-9999-9999-999999999999"
			orphanPath, err := files.Put(ctx, pngDataA)
			if err != nil {
				t.Fatal(err)
			}

			_, err = db.Exec(ctx, `
				INSERT INTO secure_attachment (id, token, file_name, mime_type, file_size_bytes, storage_path, sha256_hash, uploader_id, is_public, created_at)
				VALUES ($1, $2, 'orphan.png', 'image/png', 68, $3, $4, $5, true, clock_timestamp() - interval '48 hours')
				ON CONFLICT (id) DO UPDATE SET token = EXCLUDED.token, storage_path = EXCLUDED.storage_path, created_at = EXCLUDED.created_at
			`, orphanID, orphanTok, orphanPath, strings.Repeat("c", 64), admin.ID)
			if err != nil {
				t.Fatalf("failed to seed worker CLI orphan fixture: %v", err)
			}

			// Run real worker binary with -cleanup-attachments
			cmd := exec.Command(workerBinPath, "-cleanup-attachments", "-older-than=24h")
			cmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("worker -cleanup-attachments failed with error: %v, output: %s", err, string(out))
			}

			// Verify orphan was deleted from secure_attachment
			var orphanExists bool
			_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM secure_attachment WHERE token = $1)`, orphanTok).Scan(&orphanExists)
			if orphanExists {
				t.Fatal("worker CLI failed to delete aged orphan attachment")
			}

			// Verify audit_event has valid admin actor_id (not empty string!)
			var auditActorID string
			err = db.QueryRow(ctx, `
				SELECT actor_id FROM audit_event
				WHERE action = 'attachment.retired' AND resource_id = $1
				ORDER BY created_at DESC LIMIT 1
			`, orphanTok).Scan(&auditActorID)
			if err != nil {
				t.Fatalf("failed to find audit event for worker deletion: %v", err)
			}
			if auditActorID == "" || auditActorID != admin.ID {
				t.Fatalf("expected audit_event actor_id to be valid admin ID %q, got %q", admin.ID, auditActorID)
			}

			// Case B: No-eligible-actor failure (when no active admin exists in staff_access)
			noAdminSchema := "test_no_admin_" + strings.ReplaceAll(admin.ID, "-", "_")
			_, err = db.Exec(ctx, fmt.Sprintf(`
				CREATE SCHEMA IF NOT EXISTS %s;
				CREATE TABLE IF NOT EXISTS %s."user" (id text PRIMARY KEY, name text, email text);
				CREATE TABLE IF NOT EXISTS %s.staff_access (user_id text, role text, active boolean);
				INSERT INTO %s."user"(id, name, email) VALUES('doc-only', 'Doctor NonAdmin', 'doc@test.local') ON CONFLICT DO NOTHING;
				INSERT INTO %s.staff_access(user_id, role, active) VALUES('doc-only', 'doctor', true);
			`, noAdminSchema, noAdminSchema, noAdminSchema, noAdminSchema, noAdminSchema))
			if err != nil {
				t.Fatalf("failed to setup isolated no-admin schema: %v", err)
			}
			defer func() {
				_, _ = db.Exec(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, noAdminSchema))
			}()

			// Run worker binary pointing to the no-admin schema search path
			sep := "?"
			if strings.Contains(dbURL, "?") {
				sep = "&"
			}
			noAdminURL := dbURL + sep + "search_path=" + noAdminSchema
			cmdNoAdmin := exec.Command(workerBinPath, "-cleanup-attachments")
			cmdNoAdmin.Env = append(os.Environ(), "DATABASE_URL="+noAdminURL)
			outNoAdmin, errNoAdmin := cmdNoAdmin.CombinedOutput()
			if errNoAdmin == nil {
				t.Fatalf("expected worker CLI to exit with error when no active admin exists, got success. Output: %s", string(outNoAdmin))
			}
			if !strings.Contains(string(outNoAdmin), "missing administrative audit actor") && !strings.Contains(string(outNoAdmin), "no active administrator found") {
				t.Fatalf("expected output to mention missing administrator, got: %s", string(outNoAdmin))
			}
		}
	})

	t.Run("S5: concurrent secret preservation under unrelated-field saves regression", func(t *testing.T) {
		// Set initial secret
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"stripe_secret": "stripe-initial-secret-v1",
		}); err != nil {
			t.Fatal(err)
		}

		// Client A reads settings (receives redacted placeholder [CONFIGURED])
		clientARead, err := srv.GeneralSettings(ctx, admin)
		if err != nil {
			t.Fatal(err)
		}
		if clientARead["stripe_secret"] != domain.SecretConfiguredPlaceholder {
			t.Fatalf("expected redacted placeholder, got %s", clientARead["stripe_secret"])
		}

		// Client B concurrently replaces the secret with a fresh value
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"stripe_secret": "stripe-concurrent-updated-secret-v2",
		}); err != nil {
			t.Fatal(err)
		}

		// Client A now submits their form without changing the secret (sending placeholder [CONFIGURED] alongside company_name update)
		if err := srv.UpdateGeneralSettings(ctx, admin, map[string]string{
			"stripe_secret": domain.SecretConfiguredPlaceholder,
			"company_name":  "Client A Updated Company Name",
		}); err != nil {
			t.Fatal("Client A save failed", err)
		}

		// Assert: Client B's fresh secret was NOT overwritten with Client A's stale initial value!
		rawDB, err := store.GeneralSettings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rawDB["stripe_secret"] != "stripe-concurrent-updated-secret-v2" {
			t.Fatalf("concurrency violation: expected stripe_secret to remain Client B's update 'stripe-concurrent-updated-secret-v2', but got %q", rawDB["stripe_secret"])
		}
		if rawDB["company_name"] != "Client A Updated Company Name" {
			t.Fatalf("expected company_name updated, got %q", rawDB["company_name"])
		}

		// True overlapping concurrent goroutine test:
		// Launch 10 concurrent saves where some preserve secrets and others update company name
		errCh := make(chan error, 10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				if idx%2 == 0 {
					errCh <- srv.UpdateGeneralSettings(ctx, admin, map[string]string{
						"stripe_secret": domain.SecretConfiguredPlaceholder,
						"hospital_city": "Addis Ababa",
					})
				} else {
					errCh <- srv.UpdateGeneralSettings(ctx, admin, map[string]string{
						"hospital_phone": "+251911000111",
					})
				}
			}(i)
		}
		for i := 0; i < 10; i++ {
			if gErr := <-errCh; gErr != nil {
				t.Fatalf("concurrent save error: %v", gErr)
			}
		}
		rawDBAfter, err := store.GeneralSettings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if rawDBAfter["stripe_secret"] != "stripe-concurrent-updated-secret-v2" {
			t.Fatalf("stripe_secret corrupted after concurrent saves: got %q", rawDBAfter["stripe_secret"])
		}
	})
}
