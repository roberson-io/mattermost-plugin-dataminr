package dataminr

import "time"

const (
	// TokenExpirySeconds defines how long a Dataminr bearer token is valid (4 hours)
	TokenExpirySeconds = 14400
)

// UserInfo stores per-user Dataminr connection information
type UserInfo struct {
	MattermostUserID string        `json:"mattermost_user_id"`
	ConnectedAt      int64         `json:"connected_at"`
	LastPollAt       int64         `json:"last_poll_at"`
	Settings         *UserSettings `json:"settings,omitempty"`
}

// UserSettings stores per-user notification preferences
type UserSettings struct {
	DMNotifications    bool   `json:"dm_notifications"`    // Enable/disable DM notifications
	NotificationFilter string `json:"notification_filter"` // "all", "urgent", "flash"
	DMPollInterval     int    `json:"dm_poll_interval"`    // 0 = manual only, >0 = auto-poll (seconds)
}

// Credentials stores encrypted per-user credentials
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// TokenResponse represents the response from Dataminr authentication API
type TokenResponse struct {
	DMAToken  string `json:"dmaToken"`
	ExpiresIn int    `json:"expires_in,omitempty"`
}

// NewUserInfo creates a new UserInfo with default settings
func NewUserInfo(mattermostUserID string) *UserInfo {
	return &UserInfo{
		MattermostUserID: mattermostUserID,
		ConnectedAt:      time.Now().Unix(),
		LastPollAt:       0,
		Settings: &UserSettings{
			DMNotifications:    true,  // Default: enabled
			NotificationFilter: "all", // Default: all alerts
			DMPollInterval:     0,     // Default: manual only (no auto-polling)
		},
	}
}

// IsTokenExpired checks if the authentication token has expired
// tokenIssuedAt is the unix timestamp when the token was issued
// elapsedSeconds is how many seconds have passed since the token was issued
func (u *UserInfo) IsTokenExpired(tokenIssuedAt int64, elapsedSeconds int64) bool {
	return elapsedSeconds >= TokenExpirySeconds
}
