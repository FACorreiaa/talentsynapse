package push

import (
	"encoding/json"
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetVAPIDKey(w http.ResponseWriter, r *http.Request) {
	key := h.service.GetVAPIDPublicKey()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"publicKey": key}); err != nil {
		http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionDataFromContext(r.Context())
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var sub webpush.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid subscription body", http.StatusBadRequest)
		return
	}

	if err := h.service.StoreSubscription(r.Context(), userID, sub); err != nil {
		http.Error(w, "Failed to store subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
