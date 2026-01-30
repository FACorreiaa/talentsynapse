package settings

import "github.com/FACorreiaa/talentsynapse/internal/app/views/components"

// SettingsData holds all settings page data
type SettingsData struct {
	UserName   string
	UserAvatar string
	UserEmail  string
	ActiveTab  string
}

// GetSettingsTabs returns the tab configuration for settings
func GetSettingsTabs(activeTab string) []components.TabItem {
	return []components.TabItem{
		{ID: "account", Label: "Account", Icon: "account", URL: "/settings/tab/account", Selected: activeTab == "account"},
		{ID: "notifications", Label: "Notifications", Icon: "notifications", URL: "/settings/tab/notifications", Selected: activeTab == "notifications"},
		{ID: "privacy", Label: "Privacy", Icon: "privacy", URL: "/settings/tab/privacy", Selected: activeTab == "privacy"},
		{ID: "security", Label: "Security", Icon: "security", URL: "/settings/tab/security", Selected: activeTab == "security"},
		{ID: "appearance", Label: "Appearance", Icon: "appearance", URL: "/settings/tab/appearance", Selected: activeTab == "appearance"},
		{ID: "danger", Label: "Danger Zone", Icon: "danger", URL: "/settings/tab/danger", Selected: activeTab == "danger"},
	}
}
