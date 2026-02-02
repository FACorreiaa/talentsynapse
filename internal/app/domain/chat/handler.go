package chat

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/push"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/storage"
	chatpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/chat"
)

// Handler handles chat HTTP requests
type Handler struct {
	repo           *Repository
	matchesRepo    *matches.Repository
	hub            *Hub
	pushService    *push.Service
	storageService *storage.Service
}

// NewHandler creates a new chat handler
func NewHandler(repo *Repository, matchesRepo *matches.Repository, hub *Hub, pushService *push.Service, storageService *storage.Service) *Handler {
	return &Handler{
		repo:           repo,
		matchesRepo:    matchesRepo,
		hub:            hub,
		pushService:    pushService,
		storageService: storageService,
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
			ID:              m.ID,
			SenderID:        m.SenderID,
			SenderName:      m.SenderName,
			SenderAvatar:    m.SenderAvatar,
			Content:         m.Content,
			SentAt:          m.SentAt,
			IsOwn:           m.IsOwn,
			Type:            m.Type,
			FileURL:         m.FileURL,
			FileName:        m.FileName,
			FileSize:        m.FileSize,
			FileMimeType:    m.FileMimeType,
			DurationSeconds: m.DurationSeconds,
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

	// 2.5. Send push notification to the recipient
	// Get the other participant in the conversation
	participants, err := h.repo.GetConversationParticipants(r.Context(), conversationID)
	if err == nil {
		for _, participantID := range participants {
			// Don't send notification to the sender
			if participantID != sessionData.UserID {
				recipientUUID, err := uuid.Parse(participantID)
				if err == nil {
					// Send push notification
					if err := h.pushService.SendNewMessageNotification(r.Context(), recipientUUID, msg.SenderName); err != nil {
						log.Printf("Failed to send message notification to user %s: %v", participantID, err)
					}
				}
			}
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

// SendVoiceMessage handles voice message uploads
func (h *Handler) SendVoiceMessage(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form (max 10MB for voice)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Failed to close uploaded file: %v", cerr)
		}
	}()

	// Validate mime type
	mimeType := header.Header.Get("Content-Type")
	if !isAllowedAudioMimeType(mimeType) {
		http.Error(w, "Invalid audio format", http.StatusBadRequest)
		return
	}

	// Get duration from form
	durationStr := r.FormValue("duration")
	duration, _ := strconv.Atoi(durationStr)
	if duration <= 0 {
		duration = 1 // Default to 1 second if not provided
	}
	if duration > 300 { // Max 5 minutes
		duration = 300
	}

	// Generate unique filename
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Printf("Failed to generate random filename: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	randomName := hex.EncodeToString(randomBytes)
	ext := getAudioExtension(mimeType)
	filename := fmt.Sprintf("voice_%s%s", randomName, ext)

	// Create uploads directory if not exists
	uploadsDir := "./assets/uploads/voice"
	if err := os.MkdirAll(uploadsDir, 0o750); err != nil {
		log.Printf("Failed to create uploads directory: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Save file - filepath.Clean prevents path traversal
	filePath := filepath.Join(uploadsDir, filepath.Clean(filename))
	dst, err := os.Create(filePath) // #nosec G304 - filename is generated internally, not from user input
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil {
			log.Printf("Failed to close destination file: %v", cerr)
		}
	}()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("Failed to write file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create the file URL
	fileURL := "/assets/uploads/voice/" + filename

	// Create voice message in database
	msg, err := h.repo.CreateVoiceMessage(
		r.Context(),
		conversationID,
		sessionData.UserID,
		fileURL,
		fileSize,
		mimeType,
		duration,
	)
	if err != nil {
		log.Printf("Failed to create voice message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Prepare View Model
	viewMsg := chatpages.MessageItem{
		ID:              msg.ID,
		SenderID:        msg.SenderID,
		SenderName:      msg.SenderName,
		SenderAvatar:    msg.SenderAvatar,
		Content:         msg.Content,
		SentAt:          msg.SentAt,
		IsOwn:           false,
		Type:            msg.Type,
		FileURL:         msg.FileURL,
		DurationSeconds: msg.DurationSeconds,
	}

	// Broadcast via WebSocket (OOB Swap for receiver)
	if h.hub != nil {
		var buf bytes.Buffer
		if err := chatpages.MessageBubble(viewMsg).Render(r.Context(), &buf); err == nil {
			oobHTML := `<div id="messages-list" hx-swap-oob="beforeend">` + buf.String() + `</div>`
			h.hub.Broadcast(conversationID, []byte(oobHTML), sessionData.UserID)
		}
	}

	// Send push notification to the recipient
	participants, err := h.repo.GetConversationParticipants(r.Context(), conversationID)
	if err == nil {
		for _, participantID := range participants {
			if participantID != sessionData.UserID {
				recipientUUID, err := uuid.Parse(participantID)
				if err == nil {
					if err := h.pushService.SendNewMessageNotification(r.Context(), recipientUUID, msg.SenderName+" sent a voice message"); err != nil {
						log.Printf("Failed to send voice message notification to user %s: %v", participantID, err)
					}
				}
			}
		}
	}

	// Return response to Sender (IsOwn=true)
	senderViewMsg := viewMsg
	senderViewMsg.IsOwn = true
	component := chatpages.MessageBubble(senderViewMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// isAllowedAudioMimeType checks if the mime type is allowed for voice messages
func isAllowedAudioMimeType(mimeType string) bool {
	allowed := []string{
		"audio/webm",
		"audio/ogg",
		"audio/mp4",
		"audio/mpeg",
		"audio/wav",
		"audio/x-m4a",
	}
	for _, a := range allowed {
		if strings.HasPrefix(mimeType, a) {
			return true
		}
	}
	return false
}

// getAudioExtension returns the file extension for an audio mime type
func getAudioExtension(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "webm"):
		return ".webm"
	case strings.Contains(mimeType, "ogg"):
		return ".ogg"
	case strings.Contains(mimeType, "mp4"), strings.Contains(mimeType, "m4a"):
		return ".m4a"
	case strings.Contains(mimeType, "mpeg"):
		return ".mp3"
	case strings.Contains(mimeType, "wav"):
		return ".wav"
	default:
		return ".webm"
	}
}

// truncateMessage limits message preview length
func truncateMessage(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// UploadFile handles file and image uploads for chat
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form (max 10MB)
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close uploaded file: %v", err)
		}
	}(file)

	// Validate file type
	mimeType := header.Header.Get("Content-Type")
	if !storage.IsAllowedMimeType(mimeType) {
		http.Error(w, "File type not allowed", http.StatusBadRequest)
		return
	}

	// Upload to storage
	uploaded, err := h.storageService.UploadChatFile(
		r.Context(),
		file,
		header.Filename,
		mimeType,
		header.Size,
		sessionData.UserID,
		conversationID,
	)
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Determine message type
	msgType := "file"
	if isImage(mimeType) {
		msgType = "image"
	}

	// Create message with file metadata
	msg, err := h.repo.CreateFileMessage(
		r.Context(),
		conversationID,
		sessionData.UserID,
		msgType,
		uploaded.URL,
		uploaded.FileName,
		uploaded.Size,
		uploaded.MimeType,
		uploaded.ThumbnailURL,
	)
	if err != nil {
		log.Printf("Failed to create file message: %v", err)
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Broadcast via WebSocket if available
	if h.hub != nil {
		// Render message bubble for broadcast
		viewMsg := chatpages.MessageItem{
			ID:           msg.ID,
			SenderID:     msg.SenderID,
			SenderName:   msg.SenderName,
			SenderAvatar: msg.SenderAvatar,
			Content:      msg.Content,
			Type:         msg.Type,
			FileURL:      msg.FileURL,
			FileName:     msg.FileName,
			FileSize:     msg.FileSize,
			FileMimeType: msg.FileMimeType,
			SentAt:       msg.SentAt,
			IsOwn:        false,
		}

		var buf bytes.Buffer
		component := chatpages.MessageBubble(viewMsg)
		if err := component.Render(r.Context(), &buf); err == nil {
			h.hub.Broadcast(conversationID, buf.Bytes(), sessionData.UserID)
		}
	}

	// Send push notification to partner
	participants, err := h.repo.GetConversationParticipants(r.Context(), conversationID)
	if err == nil {
		for _, participantID := range participants {
			if participantID != sessionData.UserID && h.pushService != nil {
				notifMsg := "📎 Sent a file"
				if msgType == "image" {
					notifMsg = "📷 Sent an image"
				}
				participantUUID, _ := uuid.Parse(participantID)
				_ = h.pushService.SendNotification(r.Context(), participantUUID, push.PushNotification{
					Title: "New Message",
					Body:  notifMsg,
					URL:   "/chat/" + conversationID,
					Type:  push.NotificationTypeNewMessage,
				})
			}
		}
	}

	// Return the message bubble HTML for the sender
	senderViewMsg := chatpages.MessageItem{
		ID:           msg.ID,
		SenderID:     msg.SenderID,
		SenderName:   msg.SenderName,
		SenderAvatar: msg.SenderAvatar,
		Content:      msg.Content,
		Type:         msg.Type,
		FileURL:      msg.FileURL,
		FileName:     msg.FileName,
		FileSize:     msg.FileSize,
		FileMimeType: msg.FileMimeType,
		SentAt:       msg.SentAt,
		IsOwn:        true,
	}
	component := chatpages.MessageBubble(senderViewMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// isImage checks if the MIME type is an image
func isImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}
