package skills

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	skillspages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/skills"
)

// Handler handles skills HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new skills handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// List renders the skills management page
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's current skills
	userSkills, err := h.repo.GetUserSkills(r.Context(), sessionData.UserID)
	if err != nil {
		http.Error(w, "Failed to load skills", http.StatusInternalServerError)
		return
	}

	// Get all available skills for the catalog
	allSkills, err := h.repo.GetAllSkills(r.Context())
	if err != nil {
		http.Error(w, "Failed to load skill catalog", http.StatusInternalServerError)
		return
	}

	// Convert to view models
	var offeredSkills, wantedSkills []skillspages.SkillItem
	for _, s := range userSkills {
		item := skillspages.SkillItem{
			ID:           s.SkillID,
			Name:         s.SkillName,
			CategoryName: s.CategoryName,
			Proficiency:  s.Proficiency,
		}
		if s.SkillType == SkillTypeOffered {
			offeredSkills = append(offeredSkills, item)
		} else {
			wantedSkills = append(wantedSkills, item)
		}
	}

	var catalogSkills []skillspages.CatalogSkill
	for _, s := range allSkills {
		catalogSkills = append(catalogSkills, skillspages.CatalogSkill{
			ID:           s.ID,
			Name:         s.Name,
			CategoryName: s.CategoryName,
		})
	}

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	component := skillspages.Skills(
		offeredSkills,
		wantedSkills,
		catalogSkills,
		sessionData.UserName,
		sessionData.UserAvatar,
		successMsg,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ShowAdd renders the add skill page
func (h *Handler) ShowAdd(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	skillType := r.URL.Query().Get("type")
	if skillType != "offered" && skillType != "wanted" {
		skillType = "offered" // default
	}

	// Get all available skills for the catalog
	allSkills, err := h.repo.GetAllSkills(r.Context())
	if err != nil {
		http.Error(w, "Failed to load skill catalog", http.StatusInternalServerError)
		return
	}

	var catalogSkills []skillspages.CatalogSkill
	for _, s := range allSkills {
		catalogSkills = append(catalogSkills, skillspages.CatalogSkill{
			ID:           s.ID,
			Name:         s.Name,
			CategoryName: s.CategoryName,
		})
	}

	component := skillspages.AddSkill(
		skillType,
		catalogSkills,
		sessionData.UserName,
		sessionData.UserAvatar,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Add handles adding a skill to a user's profile
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	skillID := r.FormValue("skill_id")
	skillTypeStr := r.FormValue("skill_type")
	proficiencyStr := r.FormValue("proficiency")

	if skillID == "" || skillTypeStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	proficiency := 5 // default
	if p, err := strconv.Atoi(proficiencyStr); err == nil && p >= 1 && p <= 10 {
		proficiency = p
	}

	skillType := SkillType(skillTypeStr)
	if skillType != SkillTypeOffered && skillType != SkillTypeWanted {
		http.Error(w, "Invalid skill type", http.StatusBadRequest)
		return
	}

	err := h.repo.AddUserSkill(r.Context(), sessionData.UserID, skillID, skillType, proficiency)
	if err != nil {
		http.Error(w, "Failed to add skill", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return a partial
	if r.Header.Get("HX-Request") == "true" {
		_ = auth.SetFlash(w, r, "Skill added!", auth.FlashSuccess)
		w.Header().Set("HX-Redirect", "/skills")
		w.WriteHeader(http.StatusOK)
		return
	}

	_ = auth.SetFlash(w, r, "Skill added successfully!", auth.FlashSuccess)
	http.Redirect(w, r, "/skills", http.StatusSeeOther)
}

// Remove handles removing a skill from a user's profile
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	skillID := chi.URLParam(r, "id")
	skillTypeStr := r.URL.Query().Get("type")

	if skillID == "" {
		http.Error(w, "Missing skill ID", http.StatusBadRequest)
		return
	}

	skillType := SkillType(skillTypeStr)
	if skillType != SkillTypeOffered && skillType != SkillTypeWanted {
		http.Error(w, "Invalid skill type", http.StatusBadRequest)
		return
	}

	err := h.repo.RemoveUserSkill(r.Context(), sessionData.UserID, skillID, skillType)
	if err != nil {
		http.Error(w, "Failed to remove skill", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty response
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	_ = auth.SetFlash(w, r, "Skill removed!", auth.FlashSuccess)
	http.Redirect(w, r, "/skills", http.StatusSeeOther)
}
