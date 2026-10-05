package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/adapters/privatefiles"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testAttachmentPrivacy(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	ctx := context.Background()
	ops := application.AppointmentOpsService{Store: store}
	for _, actor := range []domain.Actor{actors[3], actors[2]} {
		if _, err := ops.DoctorQueue(ctx, actor, actors[1].ID, "2026-10-06"); !errors.Is(err, domain.ErrForbidden) {
			t.Fatal("queue scope", err)
		}
		if _, err := ops.ListPublicRequests(ctx, actor, "", 1); !errors.Is(err, domain.ErrForbidden) {
			t.Fatal("public request scope", err)
		}
	}
	if err := store.ValidateClinicalAttribution(ctx, actors[1], patients[0].ID, actors[2].ID, nil); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("doctor attribution spoof", err)
	}
	if err := store.ValidateClinicalAttribution(ctx, actors[2], patients[0].ID, actors[2].ID, nil); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("unassigned attribution", err)
	}
	if err := store.ValidateClinicalAttribution(ctx, actors[1], patients[0].ID, actors[1].ID, nil); err != nil {
		t.Fatal("assigned attribution rejected", err)
	}

	app := application.AttachmentsService{Store: store}
	input := domain.CreateSecureAttachmentInput{FileName: "private.txt", MimeType: "text/plain", FileSizeBytes: 5, StoragePath: "test/private", Sha256Hash: strings.Repeat("a", 64), PatientID: &patients[1].ID}
	att, err := app.CreateAttachment(ctx, actors[0], input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.GetAttachmentByToken(ctx, actors[3], att.Token); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other patient's token disclosed metadata: %v", err)
	}
	if _, err = app.ListPatientAttachments(ctx, actors[2], patients[1].ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned doctor listed attachments: %v", err)
	}
	if _, err = app.CreateAttachment(ctx, actors[2], input); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned doctor uploaded attachment: %v", err)
	}
	own := input
	own.PatientID = &patients[0].ID
	own.StoragePath = "test/own"
	ownAtt, err := app.CreateAttachment(ctx, actors[0], own)
	if err != nil {
		t.Fatal(err)
	}
	if err = app.Release(ctx, actors[0], ownAtt.Token); err != nil {
		t.Fatal(err)
	}
	if _, err = app.GetAttachmentByToken(ctx, actors[3], ownAtt.Token); err != nil {
		t.Fatalf("patient cannot read own attachment: %v", err)
	}
	own.IsPublic = true
	if _, err = app.CreateAttachment(ctx, actors[0], own); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("clinical public flag accepted: %v", err)
	}
	files, e := privatefiles.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer files.Close()
	app.Files = files
	data := []byte("Synthetic confidential attachment")
	upload, e := app.Upload(ctx, actors[0], "result.txt", &patients[0].ID, nil, data)
	if e != nil {
		t.Fatal(e)
	}
	hash := sha256.Sum256(data)
	if upload.Sha256Hash != hex.EncodeToString(hash[:]) || upload.FileSizeBytes != int64(len(data)) {
		t.Fatal("server did not calculate metadata")
	}
	if _, _, e = app.Download(ctx, actors[3], upload.Token); !errors.Is(e, domain.ErrForbidden) {
		t.Fatalf("unreleased download allowed: %v", e)
	}
	if e = app.Release(ctx, actors[0], upload.Token); e != nil {
		t.Fatal(e)
	}
	_, reader, e := app.Download(ctx, actors[3], upload.Token)
	if e != nil {
		t.Fatal(e)
	}
	got, e := io.ReadAll(reader)
	reader.Close()
	if e != nil || string(got) != string(data) {
		t.Fatalf("download mismatch: %v", e)
	}
	if _, _, e = app.Download(ctx, actors[2], upload.Token); !errors.Is(e, domain.ErrForbidden) {
		t.Fatalf("unassigned doctor downloaded: %v", e)
	}
	for _, name := range []string{"../outside.txt", "bad\r\nheader.txt"} {
		if _, e = app.Upload(ctx, actors[0], name, &patients[0].ID, nil, data); !errors.Is(e, domain.ErrValidation) {
			t.Fatalf("unsafe filename accepted: %v", e)
		}
	}
	if _, e = app.Upload(ctx, actors[0], "fake.png", &patients[0].ID, nil, []byte("<html><script>alert(1)</script></html>")); !errors.Is(e, domain.ErrValidation) {
		t.Fatalf("HTML upload accepted: %v", e)
	}
	if _, e = app.Upload(ctx, actors[0], "large.txt", &patients[0].ID, nil, make([]byte, application.MaxAttachmentBytes+1)); !errors.Is(e, domain.ErrValidation) {
		t.Fatalf("oversized upload accepted: %v", e)
	}
	if _, e = files.Open(ctx, "../outside.txt"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatalf("storage traversal accepted: %v", e)
	}
	server := httpapi.Server{Attachments: app, Actors: store, AuthURL: "http://auth.local", Origin: "http://127.0.0.1:3000", Client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		uid := req.Header.Get("Cookie")
		body := fmt.Sprintf(`{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, uid, uid, time.Now().Add(time.Hour).Format(time.RFC3339))
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	handler := server.Handler()
	var body bytes.Buffer
	multi := multipart.NewWriter(&body)
	part, e := multi.CreateFormFile("file", "result.txt")
	if e != nil {
		t.Fatal(e)
	}
	part.Write(data)
	multi.WriteField("patientId", patients[0].ID)
	multi.Close()
	req := httptest.NewRequest("POST", "/v1/attachments", &body)
	req.Header.Set("Content-Type", multi.FormDataContentType())
	req.Header.Set("Cookie", actors[0].ID)
	req.Header.Set("Origin", server.Origin)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("multipart upload: %d %s", rec.Code, rec.Body.String())
	}
	var saved domain.SecureAttachment
	if e = json.Unmarshal(rec.Body.Bytes(), &saved); e != nil {
		t.Fatal(e)
	}
	if e = app.Release(ctx, actors[0], saved.Token); e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		actor  string
		status int
	}{{actors[3].ID, 200}, {actors[2].ID, 403}} {
		req = httptest.NewRequest("GET", "/v1/attachments/"+saved.Token+"/content", nil)
		req.Header.Set("Cookie", tc.actor)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("download status %d expected %d", rec.Code, tc.status)
		}
		if tc.status == 200 && (rec.Body.String() != string(data) || rec.Header().Get("X-Content-Type-Options") != "nosniff") {
			t.Fatal("download content/headers mismatch")
		}
	}
}
