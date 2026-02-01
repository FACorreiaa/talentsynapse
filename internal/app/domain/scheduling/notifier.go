package scheduling

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/push"
)

// Notifier handles sending reminders for scheduled sessions
type Notifier struct {
	repo        *Repository
	hub         *Hub
	pushService *push.Service
	ticker      *time.Ticker
	done        chan bool
}

// NewNotifier creates a new reminder notifier
func NewNotifier(repo *Repository, hub *Hub, pushService *push.Service) *Notifier {
	return &Notifier{
		repo:        repo,
		hub:         hub,
		pushService: pushService,
		ticker:      time.NewTicker(1 * time.Minute), // Check every minute
		done:        make(chan bool),
	}
}

// Start begins the background reminder checking loop
func (n *Notifier) Start() {
	log.Println("Scheduling notifier started")
	go func() {
		// Process immediately on start
		n.processReminders()

		// Then process on ticker
		for {
			select {
			case <-n.ticker.C:
				n.processReminders()
			case <-n.done:
				log.Println("Scheduling notifier stopped")
				return
			}
		}
	}()
}

// Stop stops the notifier
func (n *Notifier) Stop() {
	n.ticker.Stop()
	n.done <- true
}

// processReminders fetches and sends pending reminders
func (n *Notifier) processReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reminders, err := n.repo.GetPendingReminders(ctx)
	if err != nil {
		log.Printf("Failed to fetch reminders: %v", err)
		return
	}

	if len(reminders) == 0 {
		return
	}

	log.Printf("Processing %d pending reminders", len(reminders))

	for _, reminder := range reminders {
		go n.sendReminder(reminder)
	}
}

// sendReminder sends a single reminder
func (n *Notifier) sendReminder(reminder Reminder) {
	ctx := context.Background()

	success := false
	errorMsg := ""

	switch reminder.DeliveryMethod {
	case "in_app", "all":
		// Send via WebSocket (in-app notification)
		err := n.sendInAppNotification(reminder)
		if err != nil {
			errorMsg = fmt.Sprintf("in-app failed: %v", err)
			log.Printf("Failed to send in-app reminder %s: %v", reminder.ID, err)
		} else {
			success = true
			log.Printf("Sent in-app reminder %s to user %s", reminder.ID, reminder.UserID)
		}

	case "email":
		// TODO: Implement email sending via SMTP/SendGrid
		// For now, just mark as sent
		success = true
		log.Printf("Email reminder %s (not implemented yet)", reminder.ID)

	case "push":
		// Send push notification
		err := n.sendPushNotification(reminder)
		if err != nil {
			errorMsg = fmt.Sprintf("push failed: %v", err)
			log.Printf("Failed to send push reminder %s: %v", reminder.ID, err)
		} else {
			success = true
			log.Printf("Sent push reminder %s to user %s", reminder.ID, reminder.UserID)
		}
	}

	// Mark as sent/failed
	err := n.repo.MarkReminderSent(ctx, reminder.ID, success, errorMsg)
	if err != nil {
		log.Printf("Failed to mark reminder %s as sent: %v", reminder.ID, err)
	}
}

// sendInAppNotification sends an in-app notification via WebSocket
func (n *Notifier) sendInAppNotification(reminder Reminder) error {
	// Create HTML notification fragment for HTMX
	notification := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
			<div class="alert alert-info shadow-lg mb-2" role="alert">
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
				<div>
					<h3 class="font-bold">Session Reminder</h3>
					<div class="text-xs">%s</div>
				</div>
				<button class="btn btn-sm" onclick="this.parentElement.remove()">Dismiss</button>
			</div>
		</div>
	`, reminder.Message)

	// Broadcast to user via WebSocket
	n.hub.BroadcastToUser(reminder.UserID.String(), []byte(notification))

	return nil
}

// sendPushNotification sends a push notification for a session reminder
func (n *Notifier) sendPushNotification(reminder Reminder) error {
	if n.pushService == nil {
		return fmt.Errorf("push service not available")
	}

	// Get session details to get title and start time
	ctx := context.Background()
	session, err := n.repo.GetSessionByID(ctx, reminder.SessionID, reminder.UserID)
	if err != nil {
		return fmt.Errorf("failed to get session details: %w", err)
	}

	// Calculate minutes until session
	minutesUntil := int(time.Until(session.ScheduledStart).Minutes())

	// Create a session title from offers or use a default
	sessionTitle := "Upcoming Session"
	if len(session.InitiatorOffers) > 0 {
		sessionTitle = session.InitiatorOffers[0]
	}

	return n.pushService.SendSessionReminderNotification(
		ctx,
		reminder.UserID,
		sessionTitle,
		minutesUntil,
	)
}
