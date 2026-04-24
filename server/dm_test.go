package main

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleDM(t *testing.T) {
	t.Run("enables DM notifications for connected user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		// User is connected
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    false,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Expect settings to be updated with DM enabled
		api.On("KVSet", userID+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.Settings.DMNotifications == true
		})).Return(nil)

		response, err := plugin.HandleDM(userID, true)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "enabled")
	})

	t.Run("disables DM notifications for connected user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		// User is connected with DM enabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Expect settings to be updated with DM disabled
		api.On("KVSet", userID+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.Settings.DMNotifications == false
		})).Return(nil)

		response, err := plugin.HandleDM(userID, false)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "disabled")
	})

	t.Run("returns error when user not connected", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		// User info not found
		api.On("KVGet", userID+userInfoKeyPrefix).Return(nil, nil)

		response, err := plugin.HandleDM(userID, true)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("returns error when KV get fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		api.On("KVGet", userID+userInfoKeyPrefix).Return(nil, &model.AppError{Message: "db error"})

		response, err := plugin.HandleDM(userID, true)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "error")
	})
}

func TestHandleFilter(t *testing.T) {
	t.Run("sets filter to 'all' for connected user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		// User is connected
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "flash",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Expect settings to be updated with new filter
		api.On("KVSet", userID+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.Settings.NotificationFilter == "all"
		})).Return(nil)

		response, err := plugin.HandleFilter(userID, "all")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "all")
	})

	t.Run("sets filter to 'flash' for connected user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		api.On("KVSet", userID+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.Settings.NotificationFilter == "flash"
		})).Return(nil)

		response, err := plugin.HandleFilter(userID, "flash")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "flash")
	})

	t.Run("sets filter to 'urgent' for connected user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		api.On("KVSet", userID+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.Settings.NotificationFilter == "urgent"
		})).Return(nil)

		response, err := plugin.HandleFilter(userID, "urgent")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "urgent")
	})

	t.Run("sets filter to 'flash_urgent' for connected user", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		api.On("KVSet", userID+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.Settings.NotificationFilter == "flash_urgent"
		})).Return(nil)

		response, err := plugin.HandleFilter(userID, "flash_urgent")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "flash_urgent")
	})

	t.Run("returns error when user not connected", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		api.On("KVGet", userID+userInfoKeyPrefix).Return(nil, nil)

		response, err := plugin.HandleFilter(userID, "all")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("returns error when KV get fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		api.On("KVGet", userID+userInfoKeyPrefix).Return(nil, &model.AppError{Message: "db error"})

		response, err := plugin.HandleFilter(userID, "all")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "error")
	})
}
