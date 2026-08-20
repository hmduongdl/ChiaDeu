package handler

import (
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Handler xử lý GET /api/health hoặc GET /api/backend/health
func Handler(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
