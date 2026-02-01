package push

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db              *pgxpool.Pool
	vapidPublicKey  string
	vapidPrivateKey string
}

// NotificationType represents different types of push notifications
type NotificationType string

const (
	NotificationTypeNewMatch        NotificationType = "new_match"
	NotificationTypeSessionReminder NotificationType = "session_reminder"
	NotificationTypeNewMessage      NotificationType = "new_message"
	NotificationTypeNewReview       NotificationType = "new_review"
)

// PushNotification represents the structure of a push notification
type PushNotification struct {
	Title string           `json:"title"`
	Body  string           `json:"body"`
	URL   string           `json:"url"`
	Type  NotificationType `json:"type"`
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:              db,
		vapidPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		vapidPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
	}
}

func (s *Service) GetVAPIDPublicKey() string {
	return s.vapidPublicKey
}

func (s *Service) StoreSubscription(ctx context.Context, userID uuid.UUID, sub webpush.Subscription) error {
	query := `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, endpoint) DO NOTHING
	`
	_, err := s.db.Exec(ctx, query, userID, sub.Endpoint, sub.Keys.P256dh, sub.Keys.Auth)
	if err != nil {
		return fmt.Errorf("failed to store subscription: %w", err)
	}
	return nil
}

func (s *Service) SendNotification(ctx context.Context, userID uuid.UUID, notification PushNotification) error {
	// 1. Get all subscriptions for user
	rows, err := s.db.Query(ctx, "SELECT endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to fetch subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []webpush.Subscription
	for rows.Next() {
		var sub webpush.Subscription
		sub.Keys = webpush.Keys{}
		if err := rows.Scan(&sub.Endpoint, &sub.Keys.P256dh, &sub.Keys.Auth); err != nil {
			continue
		}
		subscriptions = append(subscriptions, sub)
	}

	// 2. Marshal notification to JSON
	messageBytes, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// 3. Send to all subscriptions
	subscriber := os.Getenv("VAPID_SUBSCRIBER_EMAIL")
	if subscriber == "" {
		subscriber = "mailto:admin@talentsynapse.org"
	}

	for _, sub := range subscriptions {
		resp, err := webpush.SendNotification(messageBytes, &sub, &webpush.Options{
			Subscriber:      subscriber,
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			TTL:             30,
		})
		if err != nil {
			fmt.Printf("Failed to send push to endpoint %s: %v\n", sub.Endpoint, err)
			continue
		}

		// Handle HTTP status codes
		if resp.StatusCode == 410 || resp.StatusCode == 404 {
			// Subscription is no longer valid, remove it
			s.removeSubscription(ctx, userID, sub.Endpoint)
		}

		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}

	return nil
}

// removeSubscription removes an invalid subscription from the database
func (s *Service) removeSubscription(ctx context.Context, userID uuid.UUID, endpoint string) {
	query := `DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`
	_, err := s.db.Exec(ctx, query, userID, endpoint)
	if err != nil {
		fmt.Printf("Failed to remove invalid subscription: %v\n", err)
	}
}

// SendNewMatchNotification sends a notification for a new match
func (s *Service) SendNewMatchNotification(ctx context.Context, userID uuid.UUID, matchName string) error {
	notification := PushNotification{
		Title: "New Match Request",
		Body:  fmt.Sprintf("%s wants to connect with you!", matchName),
		URL:   "/matches",
		Type:  NotificationTypeNewMatch,
	}
	return s.SendNotification(ctx, userID, notification)
}

// SendSessionReminderNotification sends a reminder for an upcoming session
func (s *Service) SendSessionReminderNotification(ctx context.Context, userID uuid.UUID, sessionTitle string, minutesUntil int) error {
	notification := PushNotification{
		Title: "Session Reminder",
		Body:  fmt.Sprintf("Your session '%s' starts in %d minutes", sessionTitle, minutesUntil),
		URL:   "/sessions",
		Type:  NotificationTypeSessionReminder,
	}
	return s.SendNotification(ctx, userID, notification)
}

// SendNewMessageNotification sends a notification for a new message
func (s *Service) SendNewMessageNotification(ctx context.Context, userID uuid.UUID, senderName string) error {
	notification := PushNotification{
		Title: "New Message",
		Body:  fmt.Sprintf("%s sent you a message", senderName),
		URL:   "/chat",
		Type:  NotificationTypeNewMessage,
	}
	return s.SendNotification(ctx, userID, notification)
}

// SendNewReviewNotification sends a notification for a new review
func (s *Service) SendNewReviewNotification(ctx context.Context, userID uuid.UUID, reviewerName string, rating int) error {
	notification := PushNotification{
		Title: "New Review",
		Body:  fmt.Sprintf("%s left you a %d-star review", reviewerName, rating),
		URL:   "/profile",
		Type:  NotificationTypeNewReview,
	}
	return s.SendNotification(ctx, userID, notification)
}
