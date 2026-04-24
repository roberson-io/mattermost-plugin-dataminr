package dataminr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserInfoMarshalUnmarshal(t *testing.T) {
	t.Run("successfully marshals and unmarshals complete UserInfo", func(t *testing.T) {
		original := &UserInfo{
			MattermostUserID: "user123",
			ConnectedAt:      1234567890,
			LastPollAt:       1234567900,
			Settings: &UserSettings{
				DMNotifications:    true,
				NotificationFilter: "flash_urgent",
				DMPollInterval:     60,
			},
		}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, jsonData)

		var unmarshaled UserInfo
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.MattermostUserID, unmarshaled.MattermostUserID)
		assert.Equal(t, original.ConnectedAt, unmarshaled.ConnectedAt)
		assert.Equal(t, original.LastPollAt, unmarshaled.LastPollAt)
		assert.NotNil(t, unmarshaled.Settings)
		assert.Equal(t, original.Settings.DMNotifications, unmarshaled.Settings.DMNotifications)
		assert.Equal(t, original.Settings.NotificationFilter, unmarshaled.Settings.NotificationFilter)
		assert.Equal(t, original.Settings.DMPollInterval, unmarshaled.Settings.DMPollInterval)
	})

	t.Run("handles UserInfo with nil Settings", func(t *testing.T) {
		original := &UserInfo{
			MattermostUserID: "user123",
			ConnectedAt:      1234567890,
			LastPollAt:       1234567900,
			Settings:         nil,
		}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled UserInfo
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.MattermostUserID, unmarshaled.MattermostUserID)
		assert.Nil(t, unmarshaled.Settings)
	})

	t.Run("handles empty UserInfo", func(t *testing.T) {
		original := &UserInfo{}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled UserInfo
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, "", unmarshaled.MattermostUserID)
		assert.Equal(t, int64(0), unmarshaled.ConnectedAt)
	})
}

func TestUserSettingsMarshalUnmarshal(t *testing.T) {
	t.Run("successfully marshals and unmarshals UserSettings", func(t *testing.T) {
		original := &UserSettings{
			DMNotifications:    true,
			NotificationFilter: "flash",
			DMPollInterval:     120,
		}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled UserSettings
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.DMNotifications, unmarshaled.DMNotifications)
		assert.Equal(t, original.NotificationFilter, unmarshaled.NotificationFilter)
		assert.Equal(t, original.DMPollInterval, unmarshaled.DMPollInterval)
	})

	t.Run("handles default values", func(t *testing.T) {
		original := &UserSettings{}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled UserSettings
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.False(t, unmarshaled.DMNotifications)
		assert.Equal(t, "", unmarshaled.NotificationFilter)
		assert.Equal(t, 0, unmarshaled.DMPollInterval)
	})
}

func TestCredentialsMarshalUnmarshal(t *testing.T) {
	t.Run("successfully marshals and unmarshals Credentials", func(t *testing.T) {
		original := &Credentials{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled Credentials
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.ClientID, unmarshaled.ClientID)
		assert.Equal(t, original.ClientSecret, unmarshaled.ClientSecret)
	})

	t.Run("handles empty credentials", func(t *testing.T) {
		original := &Credentials{}

		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled Credentials
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, "", unmarshaled.ClientID)
		assert.Equal(t, "", unmarshaled.ClientSecret)
	})
}

func TestTokenResponseMarshalUnmarshal(t *testing.T) {
	t.Run("successfully unmarshals TokenResponse from API", func(t *testing.T) {
		// Simulate actual API response format
		apiResponse := `{
			"dmaToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			"expires_in": 14400
		}`

		var tokenResponse TokenResponse
		err := json.Unmarshal([]byte(apiResponse), &tokenResponse)
		require.NoError(t, err)

		assert.NotEmpty(t, tokenResponse.DMAToken)
		assert.Equal(t, 14400, tokenResponse.ExpiresIn)
	})

	t.Run("handles missing optional fields", func(t *testing.T) {
		apiResponse := `{
			"dmaToken": "token123"
		}`

		var tokenResponse TokenResponse
		err := json.Unmarshal([]byte(apiResponse), &tokenResponse)
		require.NoError(t, err)

		assert.Equal(t, "token123", tokenResponse.DMAToken)
		assert.Equal(t, 0, tokenResponse.ExpiresIn)
	})
}

func TestNewUserInfo(t *testing.T) {
	t.Run("creates UserInfo with default settings", func(t *testing.T) {
		userID := "user123"
		userInfo := NewUserInfo(userID)

		assert.NotNil(t, userInfo)
		assert.Equal(t, userID, userInfo.MattermostUserID)
		assert.NotZero(t, userInfo.ConnectedAt)
		assert.Zero(t, userInfo.LastPollAt)
		assert.NotNil(t, userInfo.Settings)
		assert.True(t, userInfo.Settings.DMNotifications) // Default enabled
		assert.Equal(t, "all", userInfo.Settings.NotificationFilter)
		assert.Equal(t, 0, userInfo.Settings.DMPollInterval) // 0 = use default
	})
}

func TestUserInfoIsTokenExpired(t *testing.T) {
	t.Run("returns true when token is expired", func(t *testing.T) {
		userInfo := &UserInfo{
			MattermostUserID: "user123",
			ConnectedAt:      1234567890,
			LastPollAt:       1234567890,
		}

		// Token from 5 hours ago (> 4 hour expiry)
		tokenIssuedAt := userInfo.LastPollAt
		isExpired := userInfo.IsTokenExpired(tokenIssuedAt, 5*60*60) // 5 hours ago

		assert.True(t, isExpired)
	})

	t.Run("returns false when token is still valid", func(t *testing.T) {
		userInfo := &UserInfo{
			MattermostUserID: "user123",
			ConnectedAt:      1234567890,
			LastPollAt:       1234567890,
		}

		// Token from 1 hour ago (< 4 hour expiry)
		tokenIssuedAt := userInfo.LastPollAt
		isExpired := userInfo.IsTokenExpired(tokenIssuedAt, 1*60*60) // 1 hour ago

		assert.False(t, isExpired)
	})
}
