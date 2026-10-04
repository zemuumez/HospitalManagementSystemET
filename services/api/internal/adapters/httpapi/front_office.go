package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) frontOffice(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// Complaints
	if strings.HasPrefix(r.URL.Path, "/v1/complaints") {
		switch {
		case r.URL.Path == "/v1/complaints" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			patientID := r.URL.Query().Get("patient_id")
			list, total, err := s.FrontOffice.Complaints(r.Context(), a, patientID, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"complaints": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/complaints" && r.Method == "POST":
			var in domain.ComplaintCreateInput
			if !decode(w, r, &in) {
				return true
			}
			c, err := s.FrontOffice.CreateComplaint(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, c)
			return true

		case strings.HasSuffix(r.URL.Path, "/resolve") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/complaints/")
			id = strings.TrimSuffix(id, "/resolve")
			var in domain.ComplaintResolveInput
			if !decode(w, r, &in) {
				return true
			}
			c, err := s.FrontOffice.ResolveComplaint(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, c)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/complaints/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/complaints/")
			c, err := s.FrontOffice.Complaint(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, c)
			return true
		}
	}

	// Notices
	if strings.HasPrefix(r.URL.Path, "/v1/notices") || strings.HasPrefix(r.URL.Path, "/v1/notice-boards") {
		switch {
		case (r.URL.Path == "/v1/notices" || r.URL.Path == "/v1/notice-boards") && r.Method == "GET":
			list, err := s.FrontOffice.Notices(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"notices": list, "notice_boards": list})
			return true

		case (r.URL.Path == "/v1/notices" || r.URL.Path == "/v1/notice-boards") && r.Method == "POST":
			var in domain.NoticeBoardInput
			if !decode(w, r, &in) {
				return true
			}
			n, err := s.FrontOffice.CreateNotice(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, n)
			return true

		case (strings.HasPrefix(r.URL.Path, "/v1/notices/") || strings.HasPrefix(r.URL.Path, "/v1/notice-boards/")) && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/notices/")
			id = strings.TrimPrefix(id, "/v1/notice-boards/")
			if err := s.FrontOffice.DeleteNotice(r.Context(), a, id); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"deleted": true})
			return true
		}
	}

	// Enquiries
	if strings.HasPrefix(r.URL.Path, "/v1/enquiries") {
		switch {
		case r.URL.Path == "/v1/enquiries" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			list, total, err := s.FrontOffice.Enquiries(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"enquiries": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/enquiries" && r.Method == "POST":
			var in domain.EnquiryInput
			if !decode(w, r, &in) {
				return true
			}
			enq, err := s.FrontOffice.SubmitEnquiry(r.Context(), in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, enq)
			return true

		case strings.HasSuffix(r.URL.Path, "/read") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/enquiries/")
			id = strings.TrimSuffix(id, "/read")
			enq, err := s.FrontOffice.MarkEnquiryRead(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, enq)
			return true
		}
	}

	// Visitors
	if strings.HasPrefix(r.URL.Path, "/v1/visitors") {
		switch {
		case r.URL.Path == "/v1/visitors" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			date := r.URL.Query().Get("date")
			list, total, err := s.FrontOffice.Visitors(r.Context(), a, page, date)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"visitors": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/visitors" && r.Method == "POST":
			var in domain.VisitorInput
			if !decode(w, r, &in) {
				return true
			}
			v, err := s.FrontOffice.CreateVisitor(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, v)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/visitors/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/visitors/")
			var in domain.VisitorInput
			if !decode(w, r, &in) {
				return true
			}
			v, err := s.FrontOffice.UpdateVisitor(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, v)
			return true
		}
	}

	// Call Logs
	if strings.HasPrefix(r.URL.Path, "/v1/call-logs") {
		switch {
		case r.URL.Path == "/v1/call-logs" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			callType, _ := strconv.Atoi(r.URL.Query().Get("type"))
			list, total, err := s.FrontOffice.CallLogs(r.Context(), a, page, callType)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"call_logs": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/call-logs" && r.Method == "POST":
			var in domain.CallLogInput
			if !decode(w, r, &in) {
				return true
			}
			cl, err := s.FrontOffice.CreateCallLog(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, cl)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/call-logs/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/call-logs/")
			var in domain.CallLogInput
			if !decode(w, r, &in) {
				return true
			}
			cl, err := s.FrontOffice.UpdateCallLog(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, cl)
			return true
		}
	}

	// Postals
	if strings.HasPrefix(r.URL.Path, "/v1/postals") {
		switch {
		case r.URL.Path == "/v1/postals" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			pType, _ := strconv.Atoi(r.URL.Query().Get("type"))
			list, total, err := s.FrontOffice.Postals(r.Context(), a, page, pType)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"postals": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/postals" && r.Method == "POST":
			var in domain.PostalInput
			if !decode(w, r, &in) {
				return true
			}
			p, err := s.FrontOffice.CreatePostal(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, p)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/postals/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/postals/")
			var in domain.PostalInput
			if !decode(w, r, &in) {
				return true
			}
			p, err := s.FrontOffice.UpdatePostal(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, p)
			return true
		}
	}

	return false
}
