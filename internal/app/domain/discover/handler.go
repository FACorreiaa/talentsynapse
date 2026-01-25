package discover

import (
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/skills"
	discoverpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/discover"
)

// Handler handles discover page HTTP requests
type Handler struct {
	skillsRepo *skills.Repository
}

// NewHandler creates a new discover handler
func NewHandler(skillsRepo *skills.Repository) *Handler {
	return &Handler{skillsRepo: skillsRepo}
}

// Show renders the discover/browse skills page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	selectedCategory := r.URL.Query().Get("category")

	// Get all skills from catalog
	allSkills, err := h.skillsRepo.GetAllSkills(r.Context())
	if err != nil {
		http.Error(w, "Failed to load skills", http.StatusInternalServerError)
		return
	}

	// Convert to view models
	var skillCards []discoverpages.SkillCard
	for _, s := range allSkills {
		skillCards = append(skillCards, discoverpages.SkillCard{
			ID:            s.ID,
			Name:          s.Name,
			CategoryName:  s.CategoryName,
			UsersOffering: 10, // TODO: Get actual counts from DB
			UsersWanting:  5,
		})
	}

	// Get unique categories for filters
	categoryMap := make(map[string]string)
	for _, s := range allSkills {
		categoryMap[s.CategoryID] = s.CategoryName
	}
	var categories []discoverpages.CategoryFilter
	for id, name := range categoryMap {
		categories = append(categories, discoverpages.CategoryFilter{
			ID:   id,
			Name: name,
		})
	}

	component := discoverpages.Discover(
		skillCards,
		categories,
		selectedCategory,
		sessionData.UserName,
		sessionData.UserAvatar,
		sessionData.IsAuthenticated,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
