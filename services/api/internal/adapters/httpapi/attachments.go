package httpapi

import (
	"hms.local/api/internal/application"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) attachments(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	case strings.HasPrefix(r.URL.Path, "/v1/attachments/") && strings.HasSuffix(r.URL.Path, "/release") && r.Method == "POST":
		token := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/attachments/"), "/release")
		if err := s.Attachments.Release(r.Context(), a, token); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"released": true})
		return true

	case r.URL.Path == "/v1/attachments" && r.Method == "POST":

		r.Body = http.MaxBytesReader(w, r.Body, application.MaxAttachmentBytes+(1<<20))
		err := r.ParseMultipartForm(1 << 20)
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
		if err != nil || r.MultipartForm == nil || len(r.MultipartForm.File["file"]) != 1 {
			write(w, 400, map[string]string{"error": "Provide one multipart file"})
			return true
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			fail(w, domain.ErrValidation)
			return true
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, application.MaxAttachmentBytes+1))
		if err != nil {
			fail(w, domain.ErrValidation)
			return true
		}
		patientID := strings.TrimSpace(r.FormValue("patientId"))
		isPublic := r.FormValue("isPublic") == "true"
		if isPublic || patientID == "" {
			att, err := s.Attachments.UploadPublic(r.Context(), a, header.Filename, data)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, http.StatusCreated, att)
			return true
		}
		var encounterID *string
		if value := r.FormValue("encounterId"); value != "" {
			encounterID = &value
		}
		att, err := s.Attachments.Upload(r.Context(), a, header.Filename, &patientID, encounterID, data)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusCreated, att)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attachments/") && strings.HasSuffix(r.URL.Path, "/content") && r.Method == "GET":
		token := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/attachments/"), "/content")
		att, reader, err := s.Attachments.Download(r.Context(), a, token)
		if err != nil {
			fail(w, err)
			return true
		}
		defer reader.Close()
		w.Header().Set("Content-Type", att.MimeType)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": att.FileName}))
		w.Header().Set("Content-Length", strconv.FormatInt(att.FileSizeBytes, 10))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = io.Copy(w, io.LimitReader(reader, att.FileSizeBytes))
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attachments/") && r.Method == "GET":
		token := strings.TrimPrefix(r.URL.Path, "/v1/attachments/")
		att, err := s.Attachments.GetAttachmentByToken(r.Context(), a, token)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusOK, att)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attachments/") && r.Method == "DELETE":
		token := strings.TrimPrefix(r.URL.Path, "/v1/attachments/")
		if err := s.Attachments.RetireAttachment(r.Context(), a, token); err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusNoContent, nil)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/patients/") && strings.HasSuffix(r.URL.Path, "/attachments") && r.Method == "GET":
		patientID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/patients/"), "/attachments")
		list, err := s.Attachments.ListPatientAttachments(r.Context(), a, patientID)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusOK, list)
		return true

	default:
		return false
	}
}
