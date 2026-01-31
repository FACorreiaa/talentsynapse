package push

import (
	"context"
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

func (s *Service) SendNotification(ctx context.Context, userID uuid.UUID, message string) error {
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

	// 2. Send to all
	for _, sub := range subscriptions {
		// Fire and forget individual errors for now, or log them
		// In a real app, handle 410 Gone (remove subscription)
		resp, err := webpush.SendNotification([]byte(message), &sub, &webpush.Options{
			Subscriber:      "mailto:admin@example.com", // Should be env var
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			TTL:             30,
		})
		if err != nil {
			fmt.Printf("Failed to send push: %v\n", err)
			continue
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Printf("Error closing response body: %v\n", err)
			}
		}()
	}

	return nil
}
