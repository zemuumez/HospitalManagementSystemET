package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) packages(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// Export endpoint
	if r.URL.Path == "/v1/packages-export" && r.Method == "GET" {
		list, total, err := s.Packages.Packages(r.Context(), a, 1, 1000, "")
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"packages": list, "total": total})
		return true
	}

	if !strings.HasPrefix(r.URL.Path, "/v1/packages") {
		return false
	}

	switch {
	case r.URL.Path == "/v1/packages" && r.Method == "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 25
		}
		search := r.URL.Query().Get("search")
		list, total, err := s.Packages.Packages(r.Context(), a, page, limit, search)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"packages": list, "total": total, "page": page, "limit": limit})
		return true

	case r.URL.Path == "/v1/packages" && r.Method == "POST":
		var in domain.PackageInput
		if !decode(w, r, &in) {
			return true
		}
		pkg, err := s.Packages.CreatePackage(r.Context(), a, in)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, pkg)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/packages/") && r.Method == "GET":
		id := strings.TrimPrefix(r.URL.Path, "/v1/packages/")
		pkg, err := s.Packages.Package(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, pkg)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/packages/") && (r.Method == "PUT" || r.Method == "PATCH"):
		id := strings.TrimPrefix(r.URL.Path, "/v1/packages/")
		var in domain.PackageInput
		if !decode(w, r, &in) {
			return true
		}
		pkg, err := s.Packages.UpdatePackage(r.Context(), a, id, in)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, pkg)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/packages/") && r.Method == "DELETE":
		id := strings.TrimPrefix(r.URL.Path, "/v1/packages/")
		if err := s.Packages.DeletePackage(r.Context(), a, id); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"deleted": true})
		return true
	}

	return false
}
