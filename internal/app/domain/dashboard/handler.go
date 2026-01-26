package dashboard

import (
	"log"
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	dashboardpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/dashboard"
)

// Handler handles dashboard HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new dashboard handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Show renders the dashboard page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionDataFromContext(r.Context())

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	// Default data
	data := dashboardpages.DashboardData{
		UserName:   sessionData.UserName,
		UserAvatar: sessionData.UserAvatar,
		SuccessMsg: successMsg,
		Stats: dashboardpages.DashboardStats{
			Points:   0,
			Tier:     "Bronze",
			Matches:  0,
			Sessions: 0,
		},
		ProfileStrength: dashboardpages.ProfileStrength{
			Percentage: 0,
		},
	}

	if sessionData.UserID != "" && h.repo != nil {
		// Fetch dashboard stats
		if stats, err := h.repo.GetDashboardStats(r.Context(), sessionData.UserID); err == nil {
			data.Stats.Points = stats.Points
			data.Stats.Tier = stats.Tier
			data.Stats.Matches = stats.Matches
			data.Stats.Sessions = stats.Sessions
		} else {
			log.Printf("failed to load dashboard stats for %s: %v", sessionData.UserID, err)
		}

		// Fetch top matches
		if matches, err := h.repo.GetTopMatches(r.Context(), sessionData.UserID, 3); err == nil {
			for _, m := range matches {
				data.TopMatches = append(data.TopMatches, dashboardpages.TopMatch{
					UserID:      m.UserID,
					DisplayName: m.DisplayName,
					Username:    m.Username,
					AvatarURL:   m.AvatarURL,
					Skills:      m.Skills,
					MatchScore:  m.MatchScore,
				})
			}
		} else {
			log.Printf("failed to load top matches for %s: %v", sessionData.UserID, err)
		}

		// Fetch recent activity
		if activities, err := h.repo.GetRecentActivity(r.Context(), sessionData.UserID, 4); err == nil {
			for _, a := range activities {
				data.Activities = append(data.Activities, dashboardpages.ActivityItem{
					Type:     a.Type,
					UserName: a.UserName,
					Action:   a.Action,
					TimeAgo:  a.TimeAgo,
				})
			}
		} else {
			log.Printf("failed to load recent activity for %s: %v", sessionData.UserID, err)
		}

		// Fetch profile strength
		if ps, err := h.repo.GetProfileStrength(r.Context(), sessionData.UserID); err == nil {
			data.ProfileStrength = dashboardpages.ProfileStrength{
				Percentage:       ps.Percentage,
				HasAvatar:        ps.HasAvatar,
				HasBio:           ps.HasBio,
				HasOfferedSkills: ps.HasOfferedSkills,
				HasWantedSkills:  ps.HasWantedSkills,
				HasSession:       ps.HasSession,
				HasSocialLinks:   ps.HasSocialLinks,
			}
		} else {
			log.Printf("failed to load profile strength for %s: %v", sessionData.UserID, err)
		}
	}

	component := dashboardpages.Dashboard(data)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
