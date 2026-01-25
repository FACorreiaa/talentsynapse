package admin

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin ensures that an admin user exists based on environment variables
func SeedAdmin(ctx context.Context, repo *user.Repository) error {
	adminEmail := os.Getenv("is the seed")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		log.Println("ADMIN_EMAIL or ADMIN_PASSWORD not set, skipping admin seeding")
		return nil
	}

	u, err := repo.GetByEmail(ctx, adminEmail)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			// Create new admin user
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			_, err = repo.Create(ctx, adminEmail, "admin", string(hashedPassword), "System Admin")
			if err != nil {
				return err
			}

			// Retrieve the user again to get ID, or the Create method returns it.
			// The Create method returns a pointer to User.
			// We need to set the role to "admin" explicitly as Create defaults to "member".

			// Hmmm, Create method returns the created user.
			// But we returned nil on err. Wait, let's check Create signature.
			// It returns *User, error. It does insert default role.
			// So we created a user with role 'member'. Now update it.
		} else {
			return err
		}
	}

	// Make sure the user has admin role (whether newly created or existing)
	if u == nil { // Should not happen if we just created it, but let's re-fetch if we just created it?
		// Actually, let's refetch to be sure we have the ID if we just created it or if it existed.
		u, err = repo.GetByEmail(ctx, adminEmail)
		if err != nil {
			return err
		}
	}

	if u.Role != "admin" {
		log.Printf("Promoting user %s to admin", adminEmail)
		err = repo.UpdateRole(ctx, u.ID, "admin")
		if err != nil {
			return err
		}
	}

	log.Printf("Admin check complete for %s", adminEmail)
	return nil
}
