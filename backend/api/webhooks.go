package handler

import (
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Webhooks xử lý POST /api/webhooks/:provider (sepay | payos | momo)
// phân nhánh theo query param "provider"
func Webhooks(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	provider := r.URL.Query().Get("provider")
	switch provider {
	case "sepay":
		handleWebhookSepay(w, r)
	case "payos":
		handleWebhookPayos(w, r)
	case "momo":
		handleWebhookMomo(w, r)
	default:
		vercel.SendError(w, http.StatusNotFound, "provider không hợp lệ")
	}
}

// handleWebhookSepay xử lý webhook từ SePay
func handleWebhookSepay(w http.ResponseWriter, r *http.Request) {
	// TODO: implement SePay webhook logic
	vercel.SendJSON(w, http.StatusOK, map[string]string{"message": "not implemented"})
}

// handleWebhookPayos xử lý webhook từ PayOS
func handleWebhookPayos(w http.ResponseWriter, r *http.Request) {
	// TODO: implement PayOS webhook logic
	vercel.SendJSON(w, http.StatusOK, map[string]string{"message": "not implemented"})
}

// handleWebhookMomo xử lý webhook từ MoMo
func handleWebhookMomo(w http.ResponseWriter, r *http.Request) {
	// TODO: implement MoMo webhook logic
	vercel.SendJSON(w, http.StatusOK, map[string]string{"message": "not implemented"})
}
