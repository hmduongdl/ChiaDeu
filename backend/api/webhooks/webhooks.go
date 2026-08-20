package handler

import (
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Handler xử lý POST /api/webhooks/:provider
func Handler(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]string{"message": "not implemented"})
}
