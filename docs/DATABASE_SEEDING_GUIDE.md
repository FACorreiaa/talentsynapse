# Database Seeding Guide

This guide explains the difference between **migrations** and **seeders**, and how to use each correctly.

---

## Philosophy

### Migrations = Schema + Critical Data
- Schema changes (tables, columns, indexes)
- Reference data (categories, static lookup tables)
- Data required for app to function
- **Run in ALL environments** (dev, staging, prod)

### Seeders = Development/Test Data Only
- Test users
- Sample data for development
- Demo content
- **NEVER run in production**

---

## Current Problem

Your migration `027_seed_test_data.sql` has:

```sql
-- ❌ PROBLEM: This creates test users in PRODUCTION
INSERT INTO users (username, email, display_name, hashed_password, role, is_active)
VALUES ('alice', 'alice@example.com', 'Alice Coder',
        '$2a$10$...', 'member', true);
```

**This is dangerous because:**
1. Test users exist in production with known passwords
2. Fake data pollutes production database
3. Can't be easily removed

---

## Solution: Split Into Two Approaches

### 1. Production Migration (Reference Data Only)

Create `028_seed_reference_data.sql`:

```sql
-- +goose Up
-- Reference data that MUST exist in all environments
-- This is safe to run in production

-- Skill Categories (required for app to function)
INSERT INTO skill_categories (name, description)
VALUES
    ('Development', 'Software engineering and coding'),
    ('Languages', 'Spoken languages'),
    ('Music', 'Musical instruments and theory'),
    ('Design', 'Graphic and UI/UX design'),
    ('Business', 'Business and entrepreneurship'),
    ('Marketing', 'Marketing and social media'),
    ('Writing', 'Content writing and copywriting'),
    ('Photography', 'Photography and videography'),
    ('Fitness', 'Fitness and personal training'),
    ('Cooking', 'Cooking and culinary arts')
ON CONFLICT (name) DO NOTHING;

-- Common Skills (seed popular skills)
WITH
    dev AS (SELECT id FROM skill_categories WHERE name = 'Development'),
    lang AS (SELECT id FROM skill_categories WHERE name = 'Languages'),
    music AS (SELECT id FROM skill_categories WHERE name = 'Music'),
    design AS (SELECT id FROM skill_categories WHERE name = 'Design')
INSERT INTO skills (name, category_id, description)
VALUES
    -- Development
    ('Go', (SELECT id FROM dev), 'Go programming language'),
    ('Python', (SELECT id FROM dev), 'Python programming'),
    ('JavaScript', (SELECT id FROM dev), 'JavaScript/TypeScript'),
    ('React', (SELECT id FROM dev), 'React framework'),
    ('PostgreSQL', (SELECT id FROM dev), 'Database management'),

    -- Languages
    ('English', (SELECT id FROM lang), 'English language'),
    ('Spanish', (SELECT id FROM lang), 'Spanish language'),
    ('French', (SELECT id FROM lang), 'French language'),
    ('German', (SELECT id FROM lang), 'German language'),
    ('Mandarin', (SELECT id FROM lang), 'Mandarin Chinese'),

    -- Music
    ('Piano', (SELECT id FROM music), 'Piano playing'),
    ('Guitar', (SELECT id FROM music), 'Guitar playing'),
    ('Drums', (SELECT id FROM music), 'Drum playing'),

    -- Design
    ('Figma', (SELECT id FROM design), 'UI/UX design with Figma'),
    ('Photoshop', (SELECT id FROM design), 'Adobe Photoshop'),
    ('Illustrator', (SELECT id FROM design), 'Adobe Illustrator')
ON CONFLICT (category_id, name) DO NOTHING;

-- +goose Down
-- Note: We don't delete reference data on rollback
-- because other data might depend on it
```

### 2. Development Seeder (Go Code)

Create `internal/database/seeds/dev_seeder.go`:

```go
package seeds

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    "golang.org/x/crypto/bcrypt"
)

type DevSeeder struct {
    db *pgxpool.Pool
}

func NewDevSeeder(db *pgxpool.Pool) *DevSeeder {
    return &DevSeeder{db: db}
}

// ShouldRun checks if seeder should run
func (s *DevSeeder) ShouldRun() bool {
    env := os.Getenv("GO_ENV")

    // Only run in development or test
    if env == "production" {
        log.Println("⚠️  Refusing to seed data in production environment")
        return false
    }

    return true
}

// Seed runs the development seeder
func (s *DevSeeder) Seed(ctx context.Context) error {
    if !s.ShouldRun() {
        return nil
    }

    log.Println("🌱 Seeding development data...")

    // Hash password for test users
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    // Insert test users
    users := []struct {
        username     string
        email        string
        displayName  string
        role         string
    }{
        {"alice", "alice@example.com", "Alice Coder", "member"},
        {"bob", "bob@example.com", "Bob Polyglot", "member"},
        {"charlie", "charlie@example.com", "Charlie Musician", "member"},
    }

    for _, u := range users {
        _, err := s.db.Exec(ctx, `
            INSERT INTO users (username, email, display_name, hashed_password, role, is_active)
            VALUES ($1, $2, $3, $4, $5, true)
            ON CONFLICT (email) DO NOTHING
        `, u.username, u.email, u.displayName, hashedPassword, u.role)

        if err != nil {
            return fmt.Errorf("failed to insert user %s: %w", u.username, err)
        }
    }

    // Insert user skills
    if err := s.seedUserSkills(ctx); err != nil {
        return err
    }

    log.Println("✅ Development data seeded successfully")
    return nil
}

func (s *DevSeeder) seedUserSkills(ctx context.Context) error {
    // Get user IDs
    var aliceID, bobID, charlieID int64

    err := s.db.QueryRow(ctx, "SELECT id FROM users WHERE username = 'alice'").Scan(&aliceID)
    if err != nil {
        return err
    }

    err = s.db.QueryRow(ctx, "SELECT id FROM users WHERE username = 'bob'").Scan(&bobID)
    if err != nil {
        return err
    }

    err = s.db.QueryRow(ctx, "SELECT id FROM users WHERE username = 'charlie'").Scan(&charlieID)
    if err != nil {
        return err
    }

    // Get skill IDs
    var goID, spanishID, pianoID, jsID int64

    s.db.QueryRow(ctx, "SELECT id FROM skills WHERE name = 'Go'").Scan(&goID)
    s.db.QueryRow(ctx, "SELECT id FROM skills WHERE name = 'Spanish'").Scan(&spanishID)
    s.db.QueryRow(ctx, "SELECT id FROM skills WHERE name = 'Piano'").Scan(&pianoID)
    s.db.QueryRow(ctx, "SELECT id FROM skills WHERE name = 'JavaScript'").Scan(&jsID)

    // Alice: Offers Go, Wants Spanish
    _, err = s.db.Exec(ctx, `
        INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency)
        VALUES ($1, $2, 'offered', 9), ($1, $3, 'wanted', 3)
        ON CONFLICT (user_id, skill_id, skill_type) DO NOTHING
    `, aliceID, goID, spanishID)
    if err != nil {
        return err
    }

    // Bob: Offers Spanish, Wants Go
    _, err = s.db.Exec(ctx, `
        INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency)
        VALUES ($1, $2, 'offered', 10), ($1, $3, 'wanted', 5)
        ON CONFLICT (user_id, skill_id, skill_type) DO NOTHING
    `, bobID, spanishID, goID)
    if err != nil {
        return err
    }

    // Charlie: Offers Piano, Wants JavaScript
    _, err = s.db.Exec(ctx, `
        INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency)
        VALUES ($1, $2, 'offered', 8), ($1, $3, 'wanted', 4)
        ON CONFLICT (user_id, skill_id, skill_type) DO NOTHING
    `, charlieID, pianoID, jsID)

    return err
}

// Clean removes all seeded data
func (s *DevSeeder) Clean(ctx context.Context) error {
    if !s.ShouldRun() {
        return nil
    }

    log.Println("🧹 Cleaning development data...")

    // Delete in correct order (foreign keys)
    _, err := s.db.Exec(ctx, `
        DELETE FROM user_skills WHERE user_id IN (
            SELECT id FROM users WHERE email LIKE '%@example.com'
        );
        DELETE FROM users WHERE email LIKE '%@example.com';
    `)

    if err != nil {
        return fmt.Errorf("failed to clean data: %w", err)
    }

    log.Println("✅ Development data cleaned")
    return nil
}
```

### 3. CLI Command to Run Seeder

Create `cmd/seed/main.go`:

```go
package main

import (
    "context"
    "flag"
    "log"
    "os"

    "skillsphere/internal/database"
    "skillsphere/internal/database/seeds"
)

func main() {
    clean := flag.Bool("clean", false, "Clean seeded data instead of seeding")
    flag.Parse()

    // Prevent accidental production seeding
    if os.Getenv("GO_ENV") == "production" {
        log.Fatal("❌ Cannot run seeder in production environment")
    }

    // Connect to database
    ctx := context.Background()
    db, err := database.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create seeder
    seeder := seeds.NewDevSeeder(db)

    if *clean {
        if err := seeder.Clean(ctx); err != nil {
            log.Fatal(err)
        }
    } else {
        if err := seeder.Seed(ctx); err != nil {
            log.Fatal(err)
        }
    }
}
```

### 4. Add to Makefile

```makefile
# =========================================================================
# Database Seeding
# =========================================================================

seed-dev: ## Seed development data (NOT for production)
	@if [ "$$GO_ENV" = "production" ]; then \
		echo "❌ Cannot seed in production!"; \
		exit 1; \
	fi
	@echo "🌱 Seeding development data..."
	@go run ./cmd/seed

seed-clean: ## Clean development data
	@echo "🧹 Cleaning development data..."
	@go run ./cmd/seed -clean
```

---

## Usage

### Production Deployment

```bash
# 1. Run migrations (includes reference data)
make db-up

# 2. Never run seeders
# ❌ DON'T: make seed-dev
```

### Local Development

```bash
# 1. Run migrations (schema + reference data)
make db-up

# 2. Seed test data
make seed-dev

# 3. Clean when needed
make seed-clean

# 4. Re-seed
make seed-dev
```

### CI/CD Testing

```yaml
# GitHub Actions
- name: Setup test database
  run: |
    make db-up
    make seed-dev
    make test
```

---

## Migration Strategy

### Remove 027_seed_test_data.sql

**Option 1: Create rollback migration**

```bash
# Create new migration
goose -dir ./internal/database/migrations create remove_test_data sql

# In the new migration:
# +goose Up
DELETE FROM user_skills WHERE user_id IN (
    SELECT id FROM users WHERE email LIKE '%@example.com'
);
DELETE FROM users WHERE email LIKE '%@example.com';

# +goose Down
-- Can't restore deleted data
```

**Option 2: Manual cleanup in production**

```bash
# Connect to production DB
docker exec -it skillsphere_postgres psql -U skillsphere -d skillsphere_prod

-- Delete test users
DELETE FROM user_skills WHERE user_id IN (
    SELECT id FROM users WHERE email LIKE '%@example.com'
);
DELETE FROM users WHERE email LIKE '%@example.com';
```

---

## Best Practices

### ✅ DO:
- Use migrations for reference data
- Use Go seeders for test data
- Check `GO_ENV` before seeding
- Use `ON CONFLICT DO NOTHING` in migrations
- Document what's reference vs test data

### ❌ DON'T:
- Put test users in migrations
- Run seeders in production
- Hard-code production passwords
- Delete reference data on rollback
- Mix schema changes with data seeding

---

## Summary

| Type | Method | Environments | Example |
|------|--------|--------------|---------|
| Schema | Migration | All | `CREATE TABLE users` |
| Reference Data | Migration | All | Skill categories, countries |
| Test Data | Go Seeder | Dev/Test | Test users, sample data |
| Production Data | Admin UI/API | Production | Real users via registration |

---

## Security Note

**NEVER** commit:
- Production passwords
- Real user data
- API keys
- Secrets

Always use:
- Environment variables for secrets
- `.env` files (gitignored)
- Vault/Secrets Manager for production
