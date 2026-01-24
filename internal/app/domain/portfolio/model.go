package portfolio

import (
	"time"
)

// PortfolioItem represents a project in a user's portfolio
type PortfolioItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	LinkURL     string    `json:"link_url"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
}
