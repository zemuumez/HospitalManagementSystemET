package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type staticActorStore struct {
	actors map[string]domain.Actor
}

func (s staticActorStore) Actor(ctx context.Context, id string) (domain.Actor, error) {
	if a, ok := s.actors[id]; ok {
		return a, nil
	}
	return domain.Actor{}, domain.ErrForbidden
}

func testAttendance(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()

	// Provision test staff actors
	for _, id := range []string{"att-doctor", "att-nurse", "att-reception"} {
		if _, e := db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test') ON CONFLICT(id) DO NOTHING`, id); e != nil {
			t.Fatal(e)
		}
		role := "doctor"
		if id == "att-nurse" {
			role = "nurse"
		} else if id == "att-reception" {
			role = "receptionist"
		}
		if _, e := db.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,$2) ON CONFLICT(user_id) DO NOTHING`, id, role); e != nil {
			t.Fatal(e)
		}
	}

	admin := actors[0]   // role: admin
	patient := actors[3] // role: patient
	doc := domain.Actor{ID: "att-doctor", Role: "doctor", Name: "att-doctor"}
	nurse := domain.Actor{ID: "att-nurse", Role: "nurse", Name: "att-nurse"}

	// 1. Shift Definitions
	app := application.Attendance{Store: store, Now: time.Now}

	shifts, err := app.ListShifts(ctx, doc)
	if err != nil || len(shifts) < 2 {
		t.Fatalf("expected at least 2 default seeded shifts, got %d, err=%v", len(shifts), err)
	}

	// Create a custom shift as admin
	customShiftInput := domain.ShiftInput{
		Name:                 "ICU Special Shift",
		StartTime:            "07:00",
		EndTime:              "19:00",
		GracePeriodMinutes:   10,
		BreakDurationMinutes: 60,
		HalfDayMinutes:       300,
		FullDayMinutes:       660,
		Active:               true,
	}
	customShift, err := app.CreateShift(ctx, admin, customShiftInput)
	if err != nil || customShift.Name != "ICU Special Shift" {
		t.Fatalf("admin failed to create shift: %v", err)
	}

	// Unauthorized shift creation by doctor
	if _, err = app.CreateShift(ctx, doc, customShiftInput); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doctor should be forbidden from creating shifts: %v", err)
	}

	// Patient unauthorized
	if _, err = app.ListShifts(ctx, patient); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("patient should be forbidden from reading shifts: %v", err)
	}

	// Update shift with version check
	customShiftInput.GracePeriodMinutes = 15
	updatedShift, err := app.UpdateShift(ctx, admin, customShift.ID, customShiftInput, customShift.Version)
	if err != nil || updatedShift.GracePeriodMinutes != 15 || updatedShift.Version != customShift.Version+1 {
		t.Fatalf("failed to update shift: %v", err)
	}

	// Stale update rejection
	if _, err = app.UpdateShift(ctx, admin, customShift.ID, customShiftInput, customShift.Version); !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale for outdated version update: %v", err)
	}

	// 2. Shift Assignment
	assignInput := domain.ShiftAssignmentInput{
		StaffID:       doc.ID,
		ShiftID:       customShift.ID,
		EffectiveFrom: "2026-10-01",
	}
	assignment, err := app.AssignShift(ctx, admin, assignInput)
	if err != nil || assignment.StaffID != doc.ID {
		t.Fatalf("admin failed to assign shift: %v", err)
	}

	// Non-admin assigning shift must fail
	if _, err = app.AssignShift(ctx, doc, assignInput); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doctor cannot assign shifts: %v", err)
	}

	// Query assignments
	docAssignments, err := app.ListShiftAssignments(ctx, doc, doc.ID)
	if err != nil || len(docAssignments) == 0 {
		t.Fatalf("doctor should view own assignments: %v", err)
	}

	// Doctor cannot query nurse's assignments
	if _, err = app.ListShiftAssignments(ctx, doc, nurse.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doctor cannot query other staff assignments: %v", err)
	}

	// 3. Clock In / Out & Breaks (Simulating day in EAT)
	simulatedTime := time.Date(2026, 10, 5, 8, 5, 0, 0, domain.HospitalLocation)
	appTime := application.Attendance{
		Store: store,
		Now:   func() time.Time { return simulatedTime },
	}

	// Patient clock in forbidden
	if _, err = appTime.ClockIn(ctx, patient, "", "key-1"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("patient cannot clock in: %v", err)
	}

	// Doctor Clock In
	clockInRec, err := appTime.ClockIn(ctx, doc, "", "key-doc-in")
	if err != nil || clockInRec.Status != "checked_in" {
		t.Fatalf("doctor clock in failed: %v", err)
	}

	// Duplicate clock in on same work date must fail
	if _, err = appTime.ClockIn(ctx, doc, "", "key-doc-in-2"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate clock in on same date must fail with ErrConflict: %v", err)
	}

	// Concurrent Clock In test for nurse
	var wg sync.WaitGroup
	results := make(chan error, 3)
	for n := 0; n < 3; n++ {
		wg.Add(1)
		k := fmt.Sprintf("nurse-in-%d", n)
		go func(key string) {
			defer wg.Done()
			_, e := appTime.ClockIn(ctx, nurse, "", key)
			results <- e
		}(k)
	}
	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0
	for e := range results {
		if e == nil {
			successCount++
		} else if errors.Is(e, domain.ErrConflict) {
			conflictCount++
		} else {
			t.Fatalf("unexpected concurrent error: %v", e)
		}
	}
	if successCount != 1 || conflictCount != 2 {
		t.Fatalf("expected 1 success and 2 conflicts, got %d success, %d conflicts", successCount, conflictCount)
	}

	// 4. Breaks Workflow
	simulatedTime = simulatedTime.Add(4 * time.Hour) // 12:05
	brk, err := appTime.StartBreak(ctx, doc, "Lunch break", "brk-key-1")
	if err != nil || brk.EndAt != nil {
		t.Fatalf("start break failed: %v", err)
	}

	// Overlapping break must fail
	if _, err = appTime.StartBreak(ctx, doc, "Second break", "brk-key-2"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("starting break while already on break must fail: %v", err)
	}

	// End break
	simulatedTime = simulatedTime.Add(45 * time.Minute) // 12:50
	brkEnded, err := appTime.EndBreak(ctx, doc, "brk-end-1")
	if err != nil || brkEnded.EndAt == nil || *brkEnded.DurationMinutes != 45 {
		t.Fatalf("end break failed: dur=%v, err=%v", brkEnded.DurationMinutes, err)
	}

	// Ending break when not on break
	if _, err = appTime.EndBreak(ctx, doc, "brk-end-2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ending break when none active must fail: %v", err)
	}

	// 5. Clock Out
	simulatedTime = time.Date(2026, 10, 5, 17, 5, 0, 0, domain.HospitalLocation)
	outRec, err := appTime.ClockOut(ctx, doc, "doc-out-key")
	if err != nil || outRec.CheckOutAt == nil {
		t.Fatalf("clock out failed: %v", err)
	}
	if outRec.TotalBreakMinutes != 45 {
		t.Fatalf("expected 45m break, got: %d", outRec.TotalBreakMinutes)
	}
	// Worked = (17:05 - 08:05) = 540m - 45m = 495m
	if outRec.WorkedMinutes != 495 {
		t.Fatalf("expected 495m worked, got: %d", outRec.WorkedMinutes)
	}

	// Clock out again when not clocked in
	if _, err = appTime.ClockOut(ctx, doc, "doc-out-key-2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("clocking out with no active check-in must fail: %v", err)
	}

	// 6. Overnight Shift Workflow
	if _, err = appTime.ClockOut(ctx, nurse, "nurse-day-out"); err != nil {
		t.Fatalf("close nurse day shift before night shift: %v", err)
	}
	nightShift, err := store.resolveShift(ctx, nil, "any", "2026-10-06", "")
	for _, s := range shifts {
		if s.IsOvernight {
			nightShift = s
			break
		}
	}
	if !nightShift.IsOvernight {
		t.Fatalf("expected seeded overnight night shift")
	}

	// Assign nurse to night shift
	_, err = app.AssignShift(ctx, admin, domain.ShiftAssignmentInput{
		StaffID:       nurse.ID,
		ShiftID:       nightShift.ID,
		EffectiveFrom: "2026-10-06",
	})
	if err != nil {
		t.Fatalf("failed assigning night shift: %v", err)
	}

	// Nurse clocks in on Oct 6 at 20:00 (Night shift)
	simulatedNight := time.Date(2026, 10, 6, 20, 0, 0, 0, domain.HospitalLocation)
	appNight := application.Attendance{
		Store: store,
		Now:   func() time.Time { return simulatedNight },
	}
	nurseRec, err := appNight.ClockIn(ctx, nurse, nightShift.ID, "night-in-1")
	if err != nil || nurseRec.WorkDate != "2026-10-06" {
		t.Fatalf("night clock in failed: %v", err)
	}

	// Nurse clocks out next morning at 05:00 on Oct 7
	simulatedMorning := time.Date(2026, 10, 7, 5, 0, 0, 0, domain.HospitalLocation)
	appMorning := application.Attendance{
		Store: store,
		Now:   func() time.Time { return simulatedMorning },
	}
	nurseOut, err := appMorning.ClockOut(ctx, nurse, "night-out-1")
	if err != nil || nurseOut.WorkDate != "2026-10-06" {
		t.Fatalf("night clock out workDate must stay scheduled date (2026-10-06), got: %v", nurseOut.WorkDate)
	}
	if nurseOut.Status != "present" || nurseOut.WorkedMinutes != 540 {
		t.Fatalf("night worked calculation mismatch: status=%s, worked=%d", nurseOut.Status, nurseOut.WorkedMinutes)
	}

	// 7. Scoped List & Cross-Staff ID Substitution
	filter := domain.AttendanceFilter{Page: 1, PageSize: 10}
	docList, total, err := app.ListRecords(ctx, doc, filter)
	if err != nil || total < 1 {
		t.Fatalf("doc list failed: %v, total=%d", err, total)
	}
	for _, r := range docList {
		if r.StaffID != doc.ID {
			t.Fatalf("doc saw other staff attendance: %s", r.StaffID)
		}
	}

	// Doctor substituting nurse's ID in filter
	filterSub := domain.AttendanceFilter{StaffID: nurse.ID, Page: 1, PageSize: 10}
	subList, _, err := app.ListRecords(ctx, doc, filterSub)
	if err != nil {
		t.Fatalf("sub list failed: %v", err)
	}
	// Application layer forcibly sets StaffID = actor.ID for non-admin
	for _, r := range subList {
		if r.StaffID != doc.ID {
			t.Fatalf("cross-staff ID substitution allowed: %s", r.StaffID)
		}
	}

	// Cross-staff direct ID retrieval
	if _, err = app.GetRecord(ctx, doc, nurseOut.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doc should be forbidden from getting nurse record directly: %v", err)
	}
	if _, err = app.GetHistory(ctx, doc, nurseOut.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doc should be forbidden from getting nurse history: %v", err)
	}

	// Admin can see all records
	adminList, adminTotal, err := app.ListRecords(ctx, admin, domain.AttendanceFilter{Page: 1, PageSize: 10})
	if err != nil || adminTotal < 2 || len(adminList) < 2 {
		t.Fatalf("admin list failed: %v, total=%d", err, adminTotal)
	}

	// 8. Admin Record Creation, Reasoned Corrections & Approval Lifecycle
	adminCreatedIn := time.Date(2026, 10, 4, 8, 30, 0, 0, domain.HospitalLocation)
	adminCreatedOut := time.Date(2026, 10, 4, 17, 30, 0, 0, domain.HospitalLocation)
	manualRec, err := app.AdminCreate(ctx, admin, domain.AdminAttendanceInput{
		StaffID:           doc.ID,
		WorkDate:          "2026-10-04",
		ShiftID:           shifts[0].ID,
		CheckInAt:         adminCreatedIn,
		CheckOutAt:        &adminCreatedOut,
		TotalBreakMinutes: 60,
		AdminNotes:        "Added retroactively after system maintenance",
		Reason:            "Manual entry authorized",
	}, "key-admin-manual")
	if err != nil || manualRec.Source != "administrator" {
		t.Fatalf("admin manual create failed: %v", err)
	}

	// Approve record
	approvedRec, err := app.AdminUpdateApproval(ctx, admin, manualRec.ID, domain.ApprovalInput{
		Status: "approved",
		Reason: "Verified with department supervisor",
	})
	if err != nil || approvedRec.ApprovalStatus != "approved" {
		t.Fatalf("approval update failed: %v", err)
	}

	// Non-admin approving record must fail
	if _, err = app.AdminUpdateApproval(ctx, doc, manualRec.ID, domain.ApprovalInput{Status: "approved"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("non-admin approving attendance must fail: %v", err)
	}

	// Stale correction check
	correctionInput := domain.AdminCorrectionInput{
		Version:           approvedRec.Version,
		ShiftID:           shifts[0].ID,
		CheckInAt:         adminCreatedIn.Add(-15 * time.Minute),
		CheckOutAt:        &adminCreatedOut,
		TotalBreakMinutes: 60,
		AdminNotes:        "Corrected check-in time",
		Reason:            "Corrected based on physical log sheet",
	}

	// Concurrent correction test
	corrResults := make(chan error, 2)
	var corrWg sync.WaitGroup
	for n := 0; n < 2; n++ {
		corrWg.Add(1)
		go func() {
			defer corrWg.Done()
			_, e := app.AdminCorrect(ctx, admin, manualRec.ID, correctionInput)
			corrResults <- e
		}()
	}
	corrWg.Wait()
	close(corrResults)

	corrSuccess := 0
	corrStale := 0
	for e := range corrResults {
		if e == nil {
			corrSuccess++
		} else if errors.Is(e, domain.ErrStale) {
			corrStale++
		} else {
			t.Fatalf("unexpected correction error: %v", e)
		}
	}
	if corrSuccess != 1 || corrStale != 1 {
		t.Fatalf("expected 1 success and 1 stale, got: %d success, %d stale", corrSuccess, corrStale)
	}

	// Fetch history to verify correction provenance and approval status reversion
	history, err := app.GetHistory(ctx, admin, manualRec.ID)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history.Corrections) != 1 {
		t.Fatalf("expected 1 correction event, got %d", len(history.Corrections))
	}
	if history.Corrections[0].Reason != "Corrected based on physical log sheet" {
		t.Fatalf("unexpected correction reason: %s", history.Corrections[0].Reason)
	}
	// Modifying approved attendance reset approval to submitted!
	if history.Record.ApprovalStatus != "submitted" {
		t.Fatalf("approved record modified by correction should revert to submitted, got: %s", history.Record.ApprovalStatus)
	}
	if len(history.ApprovalEvents) < 2 {
		t.Fatalf("expected approval events (approved + reset to submitted), got: %d", len(history.ApprovalEvents))
	}

	// 9. Database-level Immutability Trigger Protection
	// Trying to UPDATE or DELETE correction history directly in Postgres must fail
	if _, err = db.Exec(ctx, `UPDATE attendance_correction SET reason='tampered' WHERE id=$1`, history.Corrections[0].ID); err == nil {
		t.Fatalf("database must block UPDATE on attendance_correction table")
	}
	if _, err = db.Exec(ctx, `DELETE FROM attendance_correction WHERE id=$1`, history.Corrections[0].ID); err == nil {
		t.Fatalf("database must block DELETE on attendance_correction table")
	}
	if _, err = db.Exec(ctx, `UPDATE attendance_approval_history SET reason='tampered' WHERE id=$1`, history.ApprovalEvents[0].ID); err == nil {
		t.Fatalf("database must block UPDATE on attendance_approval_history table")
	}
	if _, err = db.Exec(ctx, `DELETE FROM attendance_approval_history WHERE id=$1`, history.ApprovalEvents[0].ID); err == nil {
		t.Fatalf("database must block DELETE on attendance_approval_history table")
	}

	// 10. Summary Statistics
	summary, err := app.Summary(ctx, admin, "2026-10-05")
	if err != nil || summary.TotalCount < 1 {
		t.Fatalf("summary failed: %v, total=%d", err, summary.TotalCount)
	}

	// 11. HTTP API Testing
	actorsMap := map[string]domain.Actor{
		admin.ID:   admin,
		doc.ID:     doc,
		nurse.ID:   nurse,
		patient.ID: patient,
	}
	server := httpapi.Server{
		Attendance: app,
		Actors:     staticActorStore{actors: actorsMap},
		AuthURL:    "http://auth.local",
		Origin:     "http://127.0.0.1:3000",
		Client: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				cookie := req.Header.Get("Cookie")
				var uid string
				switch cookie {
				case "session=admin":
					uid = admin.ID
				case "session=doc":
					uid = doc.ID
				case "session=nurse":
					uid = nurse.ID
				case "session=patient":
					uid = patient.ID
				default:
					return &http.Response{StatusCode: 401, Body: http.NoBody}, nil
				}
				body := fmt.Sprintf(`{"user":{"id":"%s"},"session":{"userId":"%s","expiresAt":"%s"}}`, uid, uid, time.Now().Add(time.Hour).Format(time.RFC3339))
				return &http.Response{
					StatusCode: 200,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       ioNopCloser(bytes.NewBufferString(body)),
				}, nil
			}),
		},
	}
	handler := server.Handler()

	// Unauthenticated test (401)
	req := httptest.NewRequest("GET", "/v1/attendance/today", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}

	// Forbidden origin test (403)
	req = httptest.NewRequest("POST", "/v1/attendance/clock-in", bytes.NewBufferString(`{}`))
	req.Header.Set("Cookie", "session=doc")
	req.Header.Set("Origin", "http://malicious.origin")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 origin denied, got %d", rec.Code)
	}

	// Disallowed unknown fields test (400)
	req = httptest.NewRequest("POST", "/v1/attendance/breaks/start", bytes.NewBufferString(`{"reason":"lunch","unknown_field":true}`))
	req.Header.Set("Cookie", "session=doc")
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown fields, got %d", rec.Code)
	}

	// Patient forbidden on records list (403)
	req = httptest.NewRequest("GET", "/v1/attendance/records", nil)
	req.Header.Set("Cookie", "session=patient")
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("patient should get 403 on attendance list, got %d", rec.Code)
	}

	// Doctor reading own records (200)
	req = httptest.NewRequest("GET", "/v1/attendance/records", nil)
	req.Header.Set("Cookie", "session=doc")
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("doc reading own records expected 200, got %d", rec.Code)
	}

	// Admin reading summary (200)
	req = httptest.NewRequest("GET", "/v1/attendance/summary?date=2026-10-05", nil)
	req.Header.Set("Cookie", "session=admin")
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin summary expected 200, got %d", rec.Code)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type nopCloser struct {
	*bytes.Buffer
}

func (nopCloser) Close() error { return nil }

func ioNopCloser(b *bytes.Buffer) nopCloser {
	return nopCloser{Buffer: b}
}
