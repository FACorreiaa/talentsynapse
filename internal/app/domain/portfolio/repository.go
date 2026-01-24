package portfolio

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateItem adds a new portfolio item
func (r *Repository) CreateItem(ctx context.Context, userID, title, desc, link, img string) (*PortfolioItem, error) {
	query := `
		INSERT INTO portfolio_items (user_id, title, description, link_url, image_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, description, link_url, image_url, created_at
	`
	var item PortfolioItem
	err := r.pool.QueryRow(ctx, query, userID, title, desc, link, img).Scan(
		&item.ID, &item.UserID, &item.Title, &item.Description, &item.LinkURL, &item.ImageURL, &item.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio item: %w", err)
	}
	return &item, nil
}

// GetUserItems fetches all portfolio items for a user
func (r *Repository) GetUserItems(ctx context.Context, userID string) ([]PortfolioItem, error) {
	query := `
		SELECT id, user_id, title, description, link_url, image_url, created_at
		FROM portfolio_items
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PortfolioItem
	for rows.Next() {
		var i PortfolioItem
		if err := rows.Scan(&i.ID, &i.UserID, &i.Title, &i.Description, &i.LinkURL, &i.ImageURL, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

// DeleteItem removes a portfolio item
func (r *Repository) DeleteItem(ctx context.Context, itemID, userID string) error {
	query := `DELETE FROM portfolio_items WHERE id = $1 AND user_id = $2`
	ct, err := r.pool.Exec(ctx, query, itemID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("item not found or unauthorized")
	}
	return nil
}
