# TalentSynapse

![TalentSynapse Logo](https://via.placeholder.com/150?text=TalentSynapse) <!-- Replace with actual logo URL if available -->

## Overview

TalentSynapse is a peer-to-peer (P2P) skill exchange platform that connects users to trade skills in real-time. Whether you're teaching Python programming or learning guitar, TalentSynapse facilitates 1:1 exchanges through profiles, matching algorithms, and interactive sessions. Built as a lightweight, scalable web app with modern RPC architecture, it emphasizes simplicity for solo developers while supporting growth into a monetized service.

The platform uses skill matching algorithms (e.g., cosine similarity on skill vectors or embedding-based semantic matching via AI) to recommend partners. It's designed for the gig economy, where users can discover, exchange, and even certify skills in a freemium model.

## Architecture & Current Focus
TalentSynapse is built as a **Progressive Web App (PWA)** with server-side rendering and minimal JavaScript. The architecture emphasizes simplicity, performance, and modern web standards without the complexity of API layers or mobile-first design.

### Key Features
- **User Profiles**: Create detailed profiles with bio, offered/wanted skills (e.g., programming, languages, arts, cooking, fitness), proficiency levels (1-10 scale), location, and availability.
- **Skill Matching**: Search and get recommendations using algorithms like Euclidean distance, cosine similarity, or AI embeddings for semantic matches (e.g., "ML" matches "machine learning").
- **Discovery**: Users find new skills via a search bar (keyword or category-based), personalized recommendations, trending skills feed, or browsing user listings. Categories include Tech, Languages, Creative Arts, Professional Development, Hobbies, and more.
- **Sessions and Chat**: Schedule 1:1 exchanges with real-time chat via WebSockets. Basic exchanges are free; premium users get priority scheduling and unlimited chats.
- **Monetization Features**: Freemium—free basic access; premium subscriptions for advanced matching, certifications (e.g., badges for completed exchanges), and ad-free experience. Users pay for access to high-demand 1:1 chats with verified experts.
- **Admin Tools**: Moderation for profiles, dispute resolution, and analytics.

### Target Audience
- Learners seeking affordable, personalized skill-building (e.g., students, career changers).
- Experts monetizing niche skills in the gig economy.
- Focus on global users, with potential localization (e.g., for Brazil's growing edtech market).

## Technology Stack

TalentSynapse is built as a **Progressive Web App (PWA)** with a focus on simplicity, performance, and modern web standards.

- **Backend**: Go with standard `net/http` for routing and handlers. Authentication via JWT/OAuth (e.g., integrate with Google/Auth0) using middleware.
- **Database**:
  - **PostgreSQL** for relational data, user profiles, and transactional consistency
- **Frontend**:
  - **Templ** for type-safe, server-rendered HTML templates
  - **HTMX** for asynchronous updates and dynamic interactions without writing JavaScript
  - **Alpine.js** (optional) for lightweight client-side interactivity (e.g., modals, toggles)
- **Real-Time Features**: WebSocket for chat and session notifications.
- **AI Integration**: Optional—use Google Gemini SDK (Go client) to generate skill embeddings for semantic similarity matching. Store embeddings in PostgreSQL with pgvector extension.
- **Deployment**: Fly.io for easy, global deployment with low-cost tiers. Alternatives: Hetzner for budget VPS or Google Cloud Platform (GCP).
- **Other Tools**: Stripe for payments, Prometheus for metrics, Docker for containerization.

### Why This Stack?
- **Go + HTMX + Templ**: Delivers fast server-side rendering with minimal JavaScript. HTMX provides modern SPA-like interactions without the complexity of frontend frameworks. PostgreSQL provides reliable relational data storage with excellent transactional support.
- **Pros**: Fast development (2-4 weeks MVP), low overhead, highly scalable. Go handles concurrency excellently for real-time features. PostgreSQL is battle-tested and reliable. Server-side focus means less client-side complexity and better SEO.
- **Cons**: For complex UIs (e.g., drag-and-drop), you may need additional JavaScript. The PWA approach works great for mobile web; native apps would require a separate build.
- **AI Decision**: Yes, for better matching—Gemini provides cost-effective embeddings without heavy ML training. Skip for MVP if focusing on basic algorithms.

### Application Architecture
TalentSynapse uses standard HTTP handlers and middleware for a clean, maintainable architecture. Key components:

- **User Management**: Registration, authentication, profile management
- **Skill Matching**: Search and recommendation algorithms (run server-side)
- **Session Management**: Schedule and track skill exchange sessions
- **Chat System**: Real-time messaging via WebSocket
- **Payment Integration**: Stripe for subscriptions and payments

Middleware handles authentication (JWT validation), logging, rate limiting, and error handling. The frontend uses HTMX to make requests and update the DOM without full page reloads.

## Installation and Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 15+ (with pgvector extension for AI features: `CREATE EXTENSION vector;`)
- Templ CLI (`go install github.com/a-h/templ/cmd/templ@latest`)

### Steps
1. Clone the repo:
   ```bash
   git clone https://github.com/yourusername/talentsynapse.git
   cd talentsynapse
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Generate Templ templates:
   ```bash
   templ generate
   ```

4. Start database:
   ```bash
   # PostgreSQL (if not already running)
   # macOS: brew services start postgresql
   # Create database: createdb talentsynapse
   ```

5. Set up environment variables (`.env` file):
   ```env
   DATABASE_URL=postgres://user:pass@localhost:5432/talentsynapse
   JWT_SECRET=your-secret-key
   GEMINI_API_KEY=your-gemini-key
   STRIPE_KEY=sk_test_...
   SERVER_PORT=8080
   ```

6. Initialize PostgreSQL database:
   ```bash
   # Run migrations
   go run cmd/migrate/main.go
   ```

7. Run the server:
   ```bash
   air  # For hot-reloading (install via go install github.com/air-verse/air@latest)
   # Or: go run cmd/server/main.go
   ```
   Access at `http://localhost:8080`.

8. Deploy to Fly.io:
   ```bash
   # Install Fly CLI
   curl -L https://fly.io/install.sh | sh

   # Launch app (follow prompts for Postgres add-on)
   fly launch

   # Add secrets
   fly secrets set DATABASE_URL=... JWT_SECRET=... GEMINI_API_KEY=... STRIPE_KEY=...

   # Deploy
   fly deploy
   ```

### Project Structure
```
talentsynapse/
├── cmd/
│   ├── server/            # Main server entry point
│   └── migrate/           # Database migrations
├── internal/
│   ├── handlers/          # HTTP handlers
│   ├── service/           # Business logic
│   ├── matching/          # Matching algorithms
│   ├── db/                # PostgreSQL models and queries
│   └── middleware/        # HTTP middleware
├── web/
│   ├── templates/         # Templ files
│   └── static/            # CSS, JS (HTMX, Alpine)
├── go.mod
└── README.md
```

## Skill Taxonomy (Optional Pre-MVP)
TalentSynapse may eventually need a formal skill ontology to keep skill names, proficiency ranges, and category relations consistent as the platform grows. This becomes important when AI matching, analytics, and external partners need a shared vocabulary.

- **What it is**: A structured taxonomy of skills, categories, and relationships (e.g., "Python" → "Programming" → "Tech")
- **When to implement**: When you notice inconsistent skill naming, need explainable AI matches, or begin syncing data with third parties
- **How to start**: Define a simple JSON schema for skills and categories in the database, then gradually evolve it as needed
- **Future considerations**: Could emit RDF/JSON-LD events for downstream analytics or partner integrations

## Usage

### Frontend Integration with HTMX
HTMX makes HTTP requests and updates the DOM without page reloads:
```html
<!-- In Templ template -->
<button hx-post="/matches/find"
        hx-include="#search-form"
        hx-target="#results"
        hx-swap="innerHTML">
  Find Matches
</button>
<div id="results"></div>
```
Server handlers return HTML fragments (generated via Templ) that HTMX swaps into the page.

### API Endpoints (Internal)
Standard HTTP endpoints for the web interface:
```bash
# User routes
POST   /auth/register
POST   /auth/login
GET    /profile/:id
POST   /profile/update

# Matching routes
POST   /matches/find
GET    /matches/recommendations

# Session routes
POST   /sessions/create
GET    /sessions/:id
```

### User Flow
- **Sign Up/Login**: Via email/OAuth. JWT tokens issued via UserService.
- **Create Profile**: Add skills (e.g., "Go Programming - Level 8" offered, "Spanish - Level 3" wanted) via UpdateProfile RPC.
- **Search & Match**: Call FindMatches RPC; get sorted recommendations via algorithms.
- **Exchange**: Schedule sessions via SessionService; chat in real-time (WebSocket or Connect streaming).
- **Premium**: Subscribe via PaymentService (Stripe integration) for perks like exclusive 1:1 access.

For code examples of matching algorithms, see `internal/matching/` directory (e.g., cosine similarity in Go using gonum).

## Code Example

### Templ Template (web/templates/matches.templ)
```templ
package templates

import "github.com/yourusername/talentsynapse/internal/db"

templ MatchResults(matches []db.UserMatch) {
    <div id="results" class="matches-container">
        for _, match := range matches {
            <div class="match-card">
                <h3>{ match.Name }</h3>
                <p>Similarity: { fmt.Sprintf("%.0f%%", match.SimilarityScore * 100) }</p>
                <div class="skills">
                    for _, skill := range match.OfferedSkills {
                        <span class="skill-tag">{ skill.Name } (Level { skill.Proficiency })</span>
                    }
                </div>
            </div>
        }
    </div>
}
```

### HTTP Handler (internal/handlers/matching.go)
```go
package handlers

import (
    "net/http"
    "github.com/yourusername/talentsynapse/internal/service"
    "github.com/yourusername/talentsynapse/web/templates"
)

type MatchingHandler struct {
    svc *service.MatchingService
}

func NewMatchingHandler(svc *service.MatchingService) *MatchingHandler {
    return &MatchingHandler{svc: svc}
}

func (h *MatchingHandler) FindMatches(w http.ResponseWriter, r *http.Request) {
    // Parse form data
    userID := r.FormValue("user_id")
    wantedSkills := r.Form["wanted_skills"]

    // Business logic in service layer
    matches, err := h.svc.FindMatches(r.Context(), userID, wantedSkills, 10)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Render Templ template and return HTML fragment
    templates.MatchResults(matches).Render(r.Context(), w)
}
```

### Main Server Setup (cmd/server/main.go)
```go
package main

import (
    "log"
    "net/http"
    "github.com/yourusername/talentsynapse/internal/handlers"
    "github.com/yourusername/talentsynapse/internal/service"
    "github.com/yourusername/talentsynapse/internal/middleware"
)

func main() {
    // Initialize services
    matchingSvc := service.NewMatchingService()
    matchingHandler := handlers.NewMatchingHandler(matchingSvc)

    // Set up routes
    mux := http.NewServeMux()
    mux.HandleFunc("/matches/find", matchingHandler.FindMatches)
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

    // Add middleware
    handler := middleware.Auth(middleware.Logging(mux))

    // Start server
    addr := ":8080"
    log.Printf("Server listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, handler))
}
```

## Roadmap

### MVP (Weeks 1-4): Core Skill Exchange Platform

#### User Flow
1. **User A & User B set up profiles**: Create accounts, add bio, avatar, and configure skills (offered/wanted with proficiency levels)
2. **Skill Discovery**: Users browse/search for potential matches via discover page or recommendations
3. **Matching**: User A finds User B based on skill overlap (cosine similarity). User A "likes" User B's profile
4. **Mutual Match**: When both users accept each other, they become **connected**
5. **Messaging**: Connected users can now message each other in real-time via WebSocket chat
6. **Learning Profiles**: Users can view each other's complete learning profiles, including skills they teach, want to learn, and their learning history
7. **Skill Exchange**: Users coordinate 1:1 sessions to exchange knowledge

#### Reviews & Verification System
- **Peer Reviews**: After each skill exchange session, users can rate and review each other (1-5 stars + written feedback)
- **Points System**: Good reviews earn points. Points accumulate toward verification thresholds
- **Verification Tiers**:
  - 🥉 **Bronze**: 10+ positive reviews → Basic verified badge
  - 🥈 **Silver**: 50+ points + 90% positive rating → Enhanced visibility in search
  - 🥇 **Gold**: 100+ points + expert endorsements → Featured in recommendations
- **Benefits**: Verified users get priority in matching, appear higher in search results, and can charge for premium sessions

#### Gamification & Badges

| Badge | Criteria | Reward | Priority |
|-------|----------|--------|----------|
| 🎯 **First Match** | Complete first exchange | Profile badge | MVP |
| 🔥 **Hot Streak** | 5 exchanges/week | 2x points | V1 |
| 📚 **Bookworm** | Learn 10 skills | Unlock stats | V1 |
| 🎓 **Mentor** | Teach 25 sessions (4.5+ rating) | Priority listing | V2 |
| 🏆 **Top Teacher** | Top 10% in category | Expert badge | V2 |
| 🌟 **Community Star** | 100+ reviews | Review badge | V1 |
| 🎖️ **Verified Expert** | Pass quiz | Certification | V1 |
| 📈 **Rising Star** | Fastest growth/month | Featured profile | V2 |

**Additional Gamification Ideas**:
- **Daily Challenges**: "Complete 1 session today" → Bonus points
- **Weekly Leaderboards**: Top learners/teachers by category
- **Learning Paths**: Complete curated skill tracks for special badges
- **Referral Rewards**: Invite friends → Both get points
- **Milestone Celebrations**: Pop-up animations for achievements
- **Streak System**: Consecutive days of activity → Multiplier bonuses

#### Moderation & Trust System
- **User Reporting**: Flag users for false content, inappropriate behavior, or spam
- **Automated Flagging**: System detects suspicious patterns (fake reviews, spam messages)
- **Moderation Queue**: Reported users go into review queue
- **Consequences**:
  - ⚠️ **Warning**: First offense, verbal warning
  - 🔇 **Shadow Ban**: Content hidden from others, user unaware
  - 🚫 **Temporary Ban**: 7-30 day suspension
  - ❌ **Permanent Ban**: Account terminated

#### Admin Panel Features
- **Super Admin Powers**:
  - View all users, sessions, and messages
  - Ban/shadow ban/warn users
  - Override verification status
  - View moderation queue and take action
  - Access platform analytics and metrics
  - Manage skill categories and badges
  - Feature/unfeature users and content
- **Audit Logs**: All admin actions logged for accountability

#### Core MVP Features Summary
- ✅ User authentication (email + OAuth)
- ✅ User profiles with skills management
- ✅ Basic matching with cosine similarity
- ✅ Real-time chat via WebSocket
- ✅ Connections (mutual matches)
- ✅ Public profile viewing
- ✅ Dedicated error pages (404, 500, 403)
- ✅ Skill-based search with exchange-aware results
- 🔲 Review system with ratings
- 🔲 Points and verification tiers
- 🔲 Badge system (gamification)
- 🔲 User reporting and moderation
- 🔲 Admin panel

---

### V1 (Months 2-3)
- AI integration (Gemini embeddings for semantic matching)
- Payments (Stripe for premium subscriptions)
- Enhanced PWA features (offline support, push notifications)
- Session scheduling with calendar integration
- Skill verification quizzes

### V2 (Months 4-6)
- Mobile PWA improvements
- Group sessions and workshops
- Advanced matching algorithms (ML-based)
- Blockchain certifications/badges (NFTs)
- B2B/Enterprise features

### Future
- Analytics dashboard for users
- API access for third-party integrations
- Microservices architecture if needed
- AI coaching and personalized learning paths

For questions, open an issue or contact [your.email@example.com].

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## TalentSynapse Business Plan

### Executive Summary
TalentSynapse is a P2P skill exchange platform launching as an MVP in 2-4 weeks, targeting the intersection of the gig economy and online learning markets. With a freemium model, it connects users for skill trades (e.g., tech, languages) using advanced matching algorithms. Built with a modern, lightweight stack (Go, HTMX, Templ) for a simple, scalable architecture. Projected revenue from subscriptions, premium chats, and ads/partnerships.

Market opportunity: The global gig economy is valued at ~$582B in 2025, online learning at ~$353B, and sharing economy at ~$246B. TalentSynapse differentiates with real-time P2P focus, AI-driven discovery, simple GoSHT stack for rapid development, and low-barrier entry.

Goal: Achieve 10K users in Year 1, $500K revenue by Year 2. Solo-founder viable, with scalable tech.

### Market Analysis
- **Size & Growth**: Gig economy: $582.2B by 2025, with 70M+ US workers (36% of workforce). Online learning: $320-353B in 2025, growing 12-14% CAGR to $842B by 2030. P2P/sharing subsets: $194B in 2024 to $246B in 2025.
- **Trends**: Rise in remote work, skill gaps (e.g., AI/tech demand). In emerging markets like Brazil, edtech booms (~$3B by 2025). Users seek personalized, affordable alternatives to courses.
- **Target Segments**: 18-45 year-olds in tech, education, creative fields. Initial focus: Global, with SEO for niches like "P2P coding exchange."
- **Competitive Analysis**: Direct competitors include Preply (language tutoring), Skillshare (class-based, not pure P2P), and alternatives like Udemy, Coursera, LinkedIn Learning, MasterClass. P2P platforms: Teachable/Kajabi for creators, but less exchange-focused. Differentiation: Free basic P2P, AI matching, real-time chat, simple GoSHT stack for rapid development. Weaknesses: Established players have scale; TalentSynapse starts lean.

### Product Description
- **Core Offering**: P2P exchanges of any skills (e.g., programming, foreign languages, music, cooking, business). Profiles include bio, skill lists with proficiencies, ratings.
- **User Flow**: Sign up → Build profile → Search/discover (algorithms recommend based on wanted/offered skills) → Match & chat (pay for premium 1:1 with high-demand users) → Exchange session → Rate/certify.
- **MVP Scope**: Profiles, basic matching (cosine similarity), chat, search. Expand to AI (Gemini for embeddings), categories.
- **Tech & Operations**: Modern stack (Go, HTMX, Templ); Go backend with standard HTTP. Development: Solo, 2-4 weeks MVP. Hosting: Fly.io (~$10-30/month initially including databases). Support: Community forums, email.

### Monetization Strategy
- **Freemium Model**: Basic free (limited matches/chats). Premium: $5-20/month for unlimited chats, priority matching, certifications.
- **Additional Revenue**:
    - Pay-per-chat: Users pay $1-10 for 1:1 access to premium profiles (e.g., verified experts).
    - Ads/Partnerships: Targeted ads from edtech firms; affiliate links (e.g., tools for skills).
    - Certifications: $10-50 for badges post-exchange.
- **Projections**: Year 1: 10K users, 10% premium conversion → $60K revenue. Scale to $500K by Year 2 via marketing.

### Marketing & Growth Strategy
- **Acquisition**: SEO (e.g., "free skill exchange app"), social media (Reddit, LinkedIn, X), content marketing (blogs on skill-building). Partnerships with influencers in niches.
- **Retention**: Email newsletters with recommendations, gamification (badges, streaks).
- **Metrics**: Track sign-ups, retention, match success rate. Budget: $1K/month initial (ads/SEO tools).
- **Launch Plan**: Beta via landing page (built with Templ), invite-only, then public.

### Operations & Team
- **Founder-Led**: Solo initially; outsource design if needed.
- **Risks**: Data privacy (GDPR compliance), moderation (AI flags). Mitigation: JWT auth via interceptors, user reports.
- **Legal**: Incorporate as LLC; terms for exchanges (no liability).

### Financial Projections
- **Startup Costs**: $500-2K (domain, hosting ~$15-30/month, Gemini API credits, Stripe fees).
- **Revenue Model**: Subscriptions (70%), pay-per-chat (20%), ads (10%).
- **Break-Even**: Month 6 at 1K premium users.
- **3-Year Forecast**:
    - Year 1: Revenue $100K, Expenses $50K (Profit $50K).
    - Year 2: $500K revenue.
    - Year 3: $2M (with 100K users).
    - Assumptions: 20% MoM growth, low churn.

This plan is adaptable—validate MVP feedback for pivots. For refinements, consult advisors.

---

## Overview of Skill Matching Algorithms

Skill matching algorithms are computational methods used to pair individuals, jobs, or resources based on skills, proficiencies, or requirements. In TalentSynapse, matching algorithms run server-side, with results returned as HTML fragments via HTMX. They range from simple rule-based systems to advanced AI-driven ones, balancing factors like accuracy, scalability, and computational cost.

Key considerations in design:
- **Input Representation**: Skills as vectors (e.g., proficiency levels from 1-10) or embeddings (semantic vectors from NLP models).
- **Scoring**: Measure similarity or distance between profiles.
- **Additional Factors**: Incorporate location, availability, or user ratings for refined matches.
- **Challenges**: Handling synonyms (e.g., "ML" vs. "machine learning"), sparse data (missing skills), and scalability for large user bases.

Below, I'll detail common algorithms, with examples of how they work and implementations in Go (using libraries like gonum for vector math) within HTTP handlers.

### 1. Distance-Based Algorithms (e.g., Euclidean/Cartesian Distance)
These treat skills as points in multi-dimensional space, where each skill is a dimension and proficiency is a coordinate. The "closeness" of two profiles is the geometric distance—smaller distances indicate better matches.

- **How It Works**:
    - Represent user profiles as vectors: e.g., User A: [Java: 8, SQL: 5, Python: 0] → Vector [8, 5, 0].
    - Compute distance to a query vector (e.g., [7, 6, 3]).
    - Formula: Euclidean Distance = √Σ (user_i - query_i)². For efficiency, use squared distance (skip the square root).
    - Missing skills default to 0.
    - Sort results by ascending distance for top matches.

- **Pros**: Simple, fast for databases; no training data needed.
- **Cons**: Doesn't handle semantic similarities (e.g., "Java" and "C#" as related).
- **Implementation Example**: In Go HTTP handlers, compute distances after fetching profiles from PostgreSQL. Use gonum/floats for vector operations.
  ```go
  package matching

  import (
      "math"
      "gonum.org/v1/gonum/floats"
  )

  // UserProfile represents a skill vector (e.g., [Java, SQL, Python] proficiencies)
  type UserProfile struct {
      ID     string
      Vector []float64
  }

  func euclideanDistance(a, b []float64) float64 {
      if len(a) != len(b) {
          panic("Vectors must be same length")
      }
      diff := make([]float64, len(a))
      floats.SubTo(diff, a, b) // diff = a - b
      floats.Mul(diff, diff)   // square each element
      return math.Sqrt(floats.Sum(diff))
  }

  func FindMatchesByDistance(query []float64, users []UserProfile, limit int) []UserProfile {
      type match struct {
          user UserProfile
          dist float64
      }
      matches := make([]match, 0, len(users))

      for _, u := range users {
          dist := euclideanDistance(u.Vector, query)
          matches = append(matches, match{user: u, dist: dist})
      }

      // Sort by distance and return top N
      sort.Slice(matches, func(i, j int) bool {
          return matches[i].dist < matches[j].dist
      })

      result := make([]UserProfile, 0, limit)
      for i := 0; i < limit && i < len(matches); i++ {
          result = append(result, matches[i].user)
      }
      return result
  }
  ```
- **Use in TalentSynapse**: Ideal for MVP with fixed skill lists; integrate into matching HTTP handlers. Fetch users from PostgreSQL, compute distances, return as HTML via Templ.

### 2. Similarity-Based Algorithms (e.g., Cosine Similarity)
These measure the angle between vectors, focusing on direction rather than magnitude—useful for sparse or varying-length profiles.

- **How It Works**:
    - Vectors as above, but normalize for length.
    - Formula: Cosine Similarity = (A · B) / (||A|| ||B||), where · is dot product (Σ A_i * B_i), and || || is magnitude (√Σ X_i²).
    - Scores range from -1 (opposite) to 1 (identical); threshold e.g., >0.7 for matches.
    - Handles zeros well (e.g., irrelevant skills don't penalize).

- **Pros**: Robust to scale differences; works well with database vector extensions like pgvector.
- **Cons**: Ignores absolute proficiency levels if not weighted.
- **Implementation Example**: In Go with gonum/floats for dot product and norms.
  ```go
  package matching

  import (
      "gonum.org/v1/gonum/floats"
  )

  func cosineSimilarity(a, b []float64) float64 {
      if len(a) != len(b) {
          panic("Vectors must be same length")
      }
      dot := floats.Dot(a, b)
      normA := floats.Norm(a, 2) // L2 norm
      normB := floats.Norm(b, 2)
      if normA == 0 || normB == 0 {
          return 0
      }
      return dot / (normA * normB)
  }

  func FindMatchesBySimilarity(query []float64, users []UserProfile, threshold float64, limit int) []UserProfile {
      type match struct {
          user UserProfile
          sim  float64
      }
      matches := make([]match, 0, len(users))

      for _, u := range users {
          sim := cosineSimilarity(u.Vector, query)
          if sim >= threshold {
              matches = append(matches, match{user: u, sim: sim})
          }
      }

      // Sort by similarity (descending) and return top N
      sort.Slice(matches, func(i, j int) bool {
          return matches[i].sim > matches[j].sim
      })

      result := make([]UserProfile, 0, limit)
      for i := 0; i < limit && i < len(matches); i++ {
          result = append(result, matches[i].user)
      }
      return result
  }
  ```
- **Use in TalentSynapse**: For fuzzy matching; combine with PostgreSQL's full-text search capabilities for skill keywords. Add to matching HTTP handlers with configurable threshold.

### 3. Machine Learning/Embedding-Based Algorithms (e.g., Word2Vec, Gemini Embeddings)
These use NLP to create dense vector representations (embeddings) of skills, capturing semantic relationships (e.g., "Python" close to "programming").

- **How It Works** (Using Gemini API):
    - **Step 1: Skill Extraction**: Parse skill names from user profiles.
    - **Step 2: Generate Embeddings**: Call Gemini API to get 768D (or similar) vectors per skill.
    - **Step 3: Profile Aggregation**: Average embeddings of all skills in a profile to create a profile vector.
    - **Step 4: Matching**: Compute cosine similarity between query profile embedding and candidate embeddings.
    - **Step 5: Threshold & Rank**: Return matches above similarity threshold (e.g., >0.62), sorted by score.

- **Pros**: Semantic understanding; handles variations like abbreviations, synonyms ("ML" = "machine learning").
- **Cons**: Requires API calls (latency, cost); cache embeddings in PostgreSQL with pgvector extension for performance.
- **Implementation Example**: Go with Gemini SDK for embeddings, gonum for similarity.
  ```go
  package matching

  import (
      "context"
      "gonum.org/v1/gonum/floats"
      "github.com/google/generative-ai-go/genai"
      "google.golang.org/api/option"
  )

  type EmbeddingMatcher struct {
      client *genai.Client
      cache  map[string][]float64 // Cache skill embeddings
  }

  func NewEmbeddingMatcher(apiKey string) (*EmbeddingMatcher, error) {
      ctx := context.Background()
      client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
      if err != nil {
          return nil, err
      }
      return &EmbeddingMatcher{
          client: client,
          cache:  make(map[string][]float64),
      }, nil
  }

  func (m *EmbeddingMatcher) GetEmbedding(ctx context.Context, skill string) ([]float64, error) {
      // Check cache first
      if emb, ok := m.cache[skill]; ok {
          return emb, nil
      }

      // Generate embedding via Gemini
      model := m.client.EmbeddingModel("text-embedding-004")
      result, err := model.EmbedContent(ctx, genai.Text(skill))
      if err != nil {
          return nil, err
      }

      // Store in cache
      m.cache[skill] = result.Embedding.Values
      return result.Embedding.Values, nil
  }

  func (m *EmbeddingMatcher) GetProfileEmbedding(ctx context.Context, skills []string) ([]float64, error) {
      embeddings := make([][]float64, 0, len(skills))

      for _, skill := range skills {
          emb, err := m.GetEmbedding(ctx, skill)
          if err != nil {
              return nil, err
          }
          embeddings = append(embeddings, emb)
      }

      // Average embeddings to get profile vector
      if len(embeddings) == 0 {
          return nil, fmt.Errorf("no skills provided")
      }

      dim := len(embeddings[0])
      avg := make([]float64, dim)
      for _, emb := range embeddings {
          floats.Add(avg, emb)
      }
      floats.Scale(1.0/float64(len(embeddings)), avg)

      return avg, nil
  }

  func (m *EmbeddingMatcher) FindSemanticMatches(
      ctx context.Context,
      querySkills []string,
      users []UserProfile, // Assume profiles have pre-computed embeddings
      threshold float64,
      limit int,
  ) ([]UserProfile, error) {
      queryEmb, err := m.GetProfileEmbedding(ctx, querySkills)
      if err != nil {
          return nil, err
      }

      type match struct {
          user UserProfile
          sim  float64
      }
      matches := make([]match, 0, len(users))

      for _, u := range users {
          sim := cosineSimilarity(u.Vector, queryEmb) // Vector is pre-computed embedding
          if sim >= threshold {
              matches = append(matches, match{user: u, sim: sim})
          }
      }

      sort.Slice(matches, func(i, j int) bool {
          return matches[i].sim > matches[j].sim
      })

      result := make([]UserProfile, 0, limit)
      for i := 0; i < limit && i < len(matches); i++ {
          result = append(result, matches[i].user)
      }
      return result, nil
  }
  ```
- **Use in TalentSynapse**: Integrate into matching HTTP handlers. Pre-compute and cache embeddings in PostgreSQL with pgvector extension for efficient vector similarity search. Return semantic matches as HTML via Templ.

### 4. Other Advanced Algorithms
- **Ontology-Based**: Use knowledge graphs (e.g., skill hierarchies like "programming > Python") for semantic matching. Metric: Graph distance or custom similarity. Can be implemented in PostgreSQL with recursive queries or dedicated graph extensions.
- **Clustering (e.g., k-Means)**: Group users into skill clusters, then match queries to nearest clusters. Useful for discovery at scale. Implement with gonum/cluster.
- **Rule-Based/Hybrid**: Simple thresholds (e.g., match if >70% skills overlap) combined with ML for ties. Implement as middleware or service layer logic that applies business rules before calling embedding matcher.
- **AI-Enhanced with LLMs**: Use Gemini to build dynamic skill ontologies, improving accuracy over time. Background job updates ontology weekly. Store in PostgreSQL with appropriate schema design.

### Recommendations for TalentSynapse
1. **MVP**: Start with **Cosine Similarity** on vectorized profiles—it's balanced for your Go backend (use gonum). Implement in matching service layer. Store all user data in PostgreSQL.
2. **V1**: Add **Gemini Embeddings** for semantic matching. Cache embeddings in PostgreSQL with pgvector extension for efficient vector similarity queries.
3. **Scaling**: Use WebSocket for real-time match updates. Leverage PostgreSQL's LISTEN/NOTIFY for real-time data updates. Split matching into separate microservice if needed.
4. **Testing**: Create synthetic data (100-1000 users); measure precision/recall. Use `httptest` for handler testing.

Install gonum: `go get gonum.org/v1/gonum`
Install Templ: `go install github.com/a-h/templ/cmd/templ@latest`

---

## Extending TalentSynapse

To grow beyond MVP, leverage 2025 edtech/gig trends like AI personalization, microlearning, blockchain creds, and hybrid work demands.

Here are targeted extensions, prioritized for monetization and feasibility:

### 1. AI-Enhanced Matching and Coaching
- Integrate LLM coaches (e.g., Gemini bots for session prep/tips) or AI-driven microlearning paths (e.g., generate personalized skill roadmaps).
- Monetize as premium: $5/month for "AI Mentor" access.
- Add ontology-based matching: Build skill graphs (e.g., "Python" → "Data Science") using knowledge bases; compute graph distances for recommendations.
- Implement with WebSocket streaming for real-time AI chat interactions.

### 2. Blockchain Certifications and Badges
- Issue verifiable badges/NFTs for completed exchanges via Polygon/Ethereum (low-fee chains).
- Users link to LinkedIn/resumes. Charge $10-50 per cert; partner with credential platforms.
- Implement certification endpoints: `/certifications/issue`, `/certifications/verify`.

### 3. Gig Economy Integrations
- **Freelance Marketplace**: Allow paid P2P gigs (e.g., "Teach Python for $20/hr") with escrow via Stripe.
- Add job boards for skill-based freelance (e.g., integrate with Upwork APIs).
- Implement gig endpoints: `/gigs/create`, `/gigs/accept`, `/gigs/complete`.
- **Mobile PWA Features**: Enhanced mobile web experience with offline support, push notifications, and installable PWA for iOS/Android.

### 4. Community and Social Extensions
- **Group Sessions**: Scale 1:1 to 1:many workshops; monetize tickets. Use WebSocket for live group interactions.
- **Trending/Global Challenges**: Weekly skill challenges (e.g., "Learn AI basics") with leaderboards; sponsor with partners.
- Implement challenge endpoints: `/challenges/trending`, `/challenges/join`.

### 5. B2B Enterprise Features
- Enterprise mode for companies (e.g., internal skill-sharing); white-label for schools.
- Multi-tenancy via middleware (tenant ID in request headers or subdomains).
- Implement organization endpoints: `/orgs/create`, `/orgs/users`, `/orgs/reports`.

### 6. Tech/Infra Extensions
- **Microservices**: Split matching into a separate service for scalability if needed. Use standard HTTP APIs between services.
- **Enhanced PWA**: Offline matching (cache data in IndexedDB), background sync, native push notifications via Firebase.
- **Analytics**: User heatmaps for trending skills; sell insights to edtech firms. Implement analytics endpoints: `/analytics/trends`, `/analytics/metrics`. Use PostgreSQL for structured analytics with appropriate aggregation queries.

These build on your freemium model, targeting $500K+ revenue by Year 2 via 20% MoM growth. Validate with beta users for pivots. The simple HTTP/HTML architecture makes it easy to iterate and add features quickly.

---

## Best Practices for TalentSynapse

1. **Templating**: Use Templ for type-safe HTML generation. Keep templates focused and composable.
2. **Middleware**: Chain middleware for auth (JWT validation), logging, rate limiting, and error handling.
3. **Error Handling**: Return appropriate HTTP status codes and user-friendly error messages. Use Templ to render error pages.
4. **Real-Time Features**: Use WebSocket for chat and live updates. PostgreSQL's LISTEN/NOTIFY provides real-time data change notifications. Consider Server-Sent Events (SSE) for one-way streaming.
5. **Testing**: Write unit tests for handlers and services. Use `httptest` for HTTP handler testing. Mock dependencies (e.g., PostgreSQL client, Gemini client).
6. **PWA Features**: Add service worker for offline support, manifest.json for installability, and web push for notifications.
7. **Observability**: Add Prometheus metrics via middleware. Track request counts, latency, error rates per endpoint.
8. **Performance**: Use HTMX for partial page updates to reduce bandwidth. Cache database queries where appropriate. Use PostgreSQL indexes for common queries.

For implementations, see the `internal/handlers/` and `internal/service/` directories.
# talentsynapse
