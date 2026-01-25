package chat

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches"
	chatpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/chat"
)

// Handler handles chat HTTP requests
type Handler struct {
	repo        *Repository
	matchesRepo *matches.Repository
	hub         *Hub
}

// NewHandler creates a new chat handler
func NewHandler(repo *Repository, matchesRepo *matches.Repository, hub *Hub) *Handler {
	return &Handler{
		repo:        repo,
		matchesRepo: matchesRepo,
		hub:         hub,
	}
}

// ListConversations renders the conversation list page
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	conversations, err := h.repo.GetConversations(r.Context(), sessionData.UserID)
	if err != nil {
		http.Error(w, "Failed to load conversations", http.StatusInternalServerError)
		return
	}

	// Convert to view models
	var viewConversations []chatpages.ConversationItem
	for _, c := range conversations {
		viewConversations = append(viewConversations, chatpages.ConversationItem{
			ID:            c.ID,
			PartnerID:     c.PartnerID,
			PartnerName:   c.PartnerName,
			PartnerAvatar: c.PartnerAvatar,
			LastMessage:   truncateMessage(c.LastMessage, 50),
			LastMessageAt: c.LastMessageAt,
			UnreadCount:   c.UnreadCount,
		})
	}

	component := chatpages.Conversations(
		viewConversations,
		sessionData.UserName,
		sessionData.UserAvatar,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ShowChat renders the chat room for a specific conversation
func (h *Handler) ShowChat(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		http.Error(w, "Missing conversation ID", http.StatusBadRequest)
		return
	}

	// Verify user is part of this conversation
	isParticipant, err := h.repo.IsParticipant(r.Context(), conversationID, sessionData.UserID)
	if err != nil || !isParticipant {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Get conversation details
	conversation, err := h.repo.GetConversation(r.Context(), conversationID, sessionData.UserID)
	if err != nil {
		http.Error(w, "Failed to load conversation", http.StatusInternalServerError)
		return
	}

	// Get messages
	messages, err := h.repo.GetMessages(r.Context(), conversationID, sessionData.UserID, 50, "")
	if err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}

	// Mark as read
	_ = h.repo.MarkAsRead(r.Context(), conversationID, sessionData.UserID)

	// Convert to view models
	var viewMessages []chatpages.MessageItem
	for _, m := range messages {
		viewMessages = append(viewMessages, chatpages.MessageItem{
			ID:           m.ID,
			SenderID:     m.SenderID,
			SenderName:   m.SenderName,
			SenderAvatar: m.SenderAvatar,
			Content:      m.Content,
			SentAt:       m.SentAt,
			IsOwn:        m.IsOwn,
		})
	}

	component := chatpages.ChatRoom(
		chatpages.ChatRoomData{
			ConversationID: conversation.ID,
			PartnerID:      conversation.PartnerID,
			PartnerName:    conversation.PartnerName,
			PartnerAvatar:  conversation.PartnerAvatar,
			Messages:       viewMessages,
		},
		sessionData.UserName,
		sessionData.UserAvatar,
		sessionData.UserID,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SendMessage handles sending a new message via HTMX
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		http.Error(w, "Missing conversation ID", http.StatusBadRequest)
		return
	}

	// Verify participant
	isParticipant, err := h.repo.IsParticipant(r.Context(), conversationID, sessionData.UserID)
	if err != nil || !isParticipant {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Parse message content
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	content := strings.TrimSpace(r.FormValue("message"))
	if content == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// Create message
	msg, err := h.repo.CreateMessage(r.Context(), conversationID, sessionData.UserID, content, "text")
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// 1. Prepare View Model
	viewMsg := chatpages.MessageItem{
		ID:           msg.ID,
		SenderID:     msg.SenderID,
		SenderName:   msg.SenderName,
		SenderAvatar: msg.SenderAvatar,
		Content:      msg.Content,
		SentAt:       msg.SentAt,
		IsOwn:        false, // For receiver, it's NOT own
	}

	// 2. Broadcast via WebSocket (OOB Swap for receiver)
	if h.hub != nil {
		var buf bytes.Buffer
		// We render with IsOwn=false for the receiver
		// We wrap in OOB swap div
		// <div id="messages-list" hx-swap-oob="beforeend">...bubble...</div>
		// Note: templ doesn't natively support easy string concatenation for this without a component
		// Let's create a wrapper component or just write string manually.
		// Manual string wrapper is risky if ids change.
		// Let's render the bubble first.
		if err := chatpages.MessageBubble(viewMsg).Render(r.Context(), &buf); err == nil {
			oobHTML := `<div id="messages-list" hx-swap-oob="beforeend">` + buf.String() + `</div>`
			h.hub.Broadcast(conversationID, []byte(oobHTML), sessionData.UserID)
		}
	}

	// 3. Return response to Sender (IsOwn=true)
	// Sender gets the bubble directly appended via hx-target logic on the form
	senderViewMsg := viewMsg
	senderViewMsg.IsOwn = true
	component := chatpages.MessageBubble(senderViewMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StartConversation creates or opens a conversation with another user
func (h *Handler) StartConversation(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	partnerID := chi.URLParam(r, "userID")
	if partnerID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Check if connected
	connected, err := h.matchesRepo.AreConnected(r.Context(), sessionData.UserID, partnerID)
	if err != nil {
		http.Error(w, "Failed to check connection", http.StatusInternalServerError)
		return
	}
	if !connected {
		_ = auth.SetFlash(w, r, "You must be connected with this user to chat.", auth.FlashError)
		http.Redirect(w, r, "/users/"+partnerID, http.StatusSeeOther)
		return
	}

	// Get or create conversation
	conversationID, err := h.repo.GetOrCreateConversation(r.Context(), sessionData.UserID, partnerID)
	if err != nil {
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chat/"+conversationID, http.StatusSeeOther)
}

// truncateMessage limits message preview length
func truncateMessage(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
