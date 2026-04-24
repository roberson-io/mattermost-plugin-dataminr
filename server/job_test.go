package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRunDataminrJob(t *testing.T) {
	t.Run("skips when disabled", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEnabled: false,
		})

		// Should not call KVList or any other method when disabled
		err := plugin.runDataminrJob()

		require.NoError(t, err)
	})

	t.Run("polls users with DM interval due", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEnabled: true,
		})

		// User with DM polling enabled and due (LastPollAt = 0, interval = 60)
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user1",
			LastPollAt:       0, // Never polled, so it's due
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
				DMPollInterval:     60, // Poll every 60 seconds
			},
		}

		// List all connected users via KV list
		api.On("KVList", 0, 100).Return([]string{
			"user1" + userInfoKeyPrefix,
		}, nil)

		// Get subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// User1 info (called twice: once for DM check, once inside pollUserAlerts)
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", "user1"+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Get cursor
		api.On("KVGet", "user1"+cursorKeyPrefix).Return(nil, nil)

		// Store updated user info with new LastPollAt
		api.On("KVSet", "user1"+userInfoKeyPrefix, mock.AnythingOfType("[]uint8")).Return(nil)

		err := plugin.runDataminrJob()

		require.NoError(t, err)
	})

	t.Run("skips users with DM interval not due", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEnabled: true,
		})

		// User with DM polling enabled but polled very recently (use real time)
		now := time.Now().Unix()
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user1",
			LastPollAt:       now - 5, // Polled just 5 seconds ago
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
				DMPollInterval:     60, // Poll every 60 seconds (not due yet - 55 seconds remaining)
			},
		}

		// List all connected users via KV list
		api.On("KVList", 0, 100).Return([]string{
			"user1" + userInfoKeyPrefix,
		}, nil)

		// Get subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// User1 info
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", "user1"+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Should NOT call pollUserAlerts since not due

		err := plugin.runDataminrJob()

		require.NoError(t, err)
	})

	t.Run("skips users with manual-only DM interval", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEnabled: true,
		})

		// User with DM polling set to manual only (0)
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user1",
			LastPollAt:       0,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
				DMPollInterval:     0, // Manual only
			},
		}

		// List all connected users via KV list
		api.On("KVList", 0, 100).Return([]string{
			"user1" + userInfoKeyPrefix,
		}, nil)

		// Get subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// User1 info
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", "user1"+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Should NOT call pollUserAlerts since interval is 0

		err := plugin.runDataminrJob()

		require.NoError(t, err)
	})

	t.Run("handles no connected users gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEnabled: true,
		})

		// No users
		api.On("KVList", 0, 100).Return([]string{}, nil)

		// Get subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		err := plugin.runDataminrJob()

		require.NoError(t, err)
	})

	t.Run("handles KVList error gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEnabled: true,
		})

		api.On("KVList", 0, 100).Return(nil, &model.AppError{Message: "db error"})

		err := plugin.runDataminrJob()

		require.Error(t, err)
	})
}

func TestGetAllConnectedDataminrUsers(t *testing.T) {
	t.Run("returns list of connected user IDs", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVList", 0, 100).Return([]string{
			"user1" + userInfoKeyPrefix,
			"user2" + userInfoKeyPrefix,
			"random_key",
			"user3" + userInfoKeyPrefix,
		}, nil)

		users, err := plugin.getAllConnectedDataminrUsers()

		require.NoError(t, err)
		require.Len(t, users, 3)
		require.Contains(t, users, "user1")
		require.Contains(t, users, "user2")
		require.Contains(t, users, "user3")
	})

	t.Run("returns empty slice when no users connected", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVList", 0, 100).Return([]string{}, nil)

		users, err := plugin.getAllConnectedDataminrUsers()

		require.NoError(t, err)
		require.Len(t, users, 0)
	})

	t.Run("returns error when KVList fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVList", 0, 100).Return(nil, &model.AppError{Message: "db error"})

		users, err := plugin.getAllConnectedDataminrUsers()

		require.Error(t, err)
		require.Nil(t, users)
	})
}

func TestPollUserAlerts(t *testing.T) {
	t.Run("fetches user info and cursor", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)

		api.On("KVGet", "user123"+userInfoKeyPrefix).Return(userInfoJSON, nil)
		api.On("KVGet", "user123"+cursorKeyPrefix).Return([]byte("cursor123"), nil)

		// No subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil).Maybe()

		err := plugin.pollUserAlerts("user123")

		require.NoError(t, err)
	})

	t.Run("returns error when user info not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVGet", "user123"+userInfoKeyPrefix).Return(nil, nil)

		err := plugin.pollUserAlerts("user123")

		require.Error(t, err)
	})

	t.Run("handles missing cursor gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)

		api.On("KVGet", "user123"+userInfoKeyPrefix).Return(userInfoJSON, nil)
		api.On("KVGet", "user123"+cursorKeyPrefix).Return(nil, nil) // No cursor

		// No subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil).Maybe()

		err := plugin.pollUserAlerts("user123")

		require.NoError(t, err)
	})
}

func TestProcessAlerts(t *testing.T) {
	t.Run("deduplicates and routes alerts", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alerts := []dataminr.Alert{
			{AlertID: "alert1", Headline: "First alert", AlertType: &dataminr.AlertType{Name: "Alert"}},
			{AlertID: "alert2", Headline: "Second alert", AlertType: &dataminr.AlertType{Name: "Flash"}},
		}

		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)

		// Deduplication checks
		api.On("KVGet", "dataminr_alert_alert1").Return(nil, nil)
		api.On("KVSetWithExpiry", "dataminr_alert_alert1", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)
		api.On("KVGet", "dataminr_alert_alert2").Return(nil, nil)
		api.On("KVSetWithExpiry", "dataminr_alert_alert2", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		// Routing - get user info for each alert
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// No subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// DM channel and posts
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(&model.Post{}, nil)

		err := plugin.processAlerts(userID, alerts)

		require.NoError(t, err)
	})

	t.Run("handles empty alerts slice", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alerts := []dataminr.Alert{}

		err := plugin.processAlerts("user123", alerts)

		require.NoError(t, err)
	})

	t.Run("skips duplicate alerts", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alerts := []dataminr.Alert{
			{AlertID: "alert1", Headline: "Seen alert", AlertType: &dataminr.AlertType{Name: "Alert"}},
		}

		// alert1 already seen
		api.On("KVGet", "dataminr_alert_alert1").Return([]byte("seen"), nil)

		// No routing should happen (no CreatePost)

		err := plugin.processAlerts(userID, alerts)

		require.NoError(t, err)
	})
}

func TestUpdateLastPollTime(t *testing.T) {
	t.Run("updates user info with current time", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			LastPollAt:       0,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)

		api.On("KVGet", "user123"+userInfoKeyPrefix).Return(userInfoJSON, nil)
		api.On("KVSet", "user123"+userInfoKeyPrefix, mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			_ = json.Unmarshal(data, &updated)
			return updated.LastPollAt > 0
		})).Return(nil)

		err := plugin.updateLastPollTime("user123")

		require.NoError(t, err)
	})
}

func TestIsSubscriptionDueForPoll(t *testing.T) {
	plugin := &Plugin{}

	t.Run("returns false when subscription is disabled", func(t *testing.T) {
		sub := &dataminr.Subscription{
			Enabled:      false,
			PollInterval: 60,
			LastPollAt:   0,
		}

		result := plugin.isSubscriptionDueForPoll(sub, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns false when poll interval is zero (manual only)", func(t *testing.T) {
		sub := &dataminr.Subscription{
			Enabled:      true,
			PollInterval: 0,
			LastPollAt:   0,
		}

		result := plugin.isSubscriptionDueForPoll(sub, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns false when poll interval is negative", func(t *testing.T) {
		sub := &dataminr.Subscription{
			Enabled:      true,
			PollInterval: -1,
			LastPollAt:   0,
		}

		result := plugin.isSubscriptionDueForPoll(sub, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns true when enough time has passed", func(t *testing.T) {
		now := time.Now().Unix()
		sub := &dataminr.Subscription{
			Enabled:      true,
			PollInterval: 60,        // 60 seconds
			LastPollAt:   now - 120, // 2 minutes ago
		}

		result := plugin.isSubscriptionDueForPoll(sub, now)

		require.True(t, result)
	})

	t.Run("returns false when not enough time has passed", func(t *testing.T) {
		now := time.Now().Unix()
		sub := &dataminr.Subscription{
			Enabled:      true,
			PollInterval: 60,       // 60 seconds
			LastPollAt:   now - 30, // Only 30 seconds ago
		}

		result := plugin.isSubscriptionDueForPoll(sub, now)

		require.False(t, result)
	})

	t.Run("returns true when never polled before", func(t *testing.T) {
		now := time.Now().Unix()
		sub := &dataminr.Subscription{
			Enabled:      true,
			PollInterval: 60,
			LastPollAt:   0, // Never polled
		}

		result := plugin.isSubscriptionDueForPoll(sub, now)

		require.True(t, result)
	})

	t.Run("returns true exactly at interval boundary", func(t *testing.T) {
		now := time.Now().Unix()
		sub := &dataminr.Subscription{
			Enabled:      true,
			PollInterval: 60,
			LastPollAt:   now - 60, // Exactly 60 seconds ago
		}

		result := plugin.isSubscriptionDueForPoll(sub, now)

		require.True(t, result)
	})
}

func TestIsUserDMDueForPoll(t *testing.T) {
	plugin := &Plugin{}

	t.Run("returns false when settings is nil", func(t *testing.T) {
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			Settings:         nil,
		}

		result := plugin.isUserDMDueForPoll(userInfo, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns false when DM notifications disabled", func(t *testing.T) {
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			Settings: &dataminr.UserSettings{
				DMNotifications: false,
				DMPollInterval:  60,
			},
		}

		result := plugin.isUserDMDueForPoll(userInfo, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns false when interval is zero (manual only)", func(t *testing.T) {
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			Settings: &dataminr.UserSettings{
				DMNotifications: true,
				DMPollInterval:  0,
			},
		}

		result := plugin.isUserDMDueForPoll(userInfo, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns false when interval is negative", func(t *testing.T) {
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			Settings: &dataminr.UserSettings{
				DMNotifications: true,
				DMPollInterval:  -1,
			},
		}

		result := plugin.isUserDMDueForPoll(userInfo, time.Now().Unix())

		require.False(t, result)
	})

	t.Run("returns true when enough time has passed", func(t *testing.T) {
		now := time.Now().Unix()
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			LastPollAt:       now - 120, // 2 minutes ago
			Settings: &dataminr.UserSettings{
				DMNotifications: true,
				DMPollInterval:  60, // Every 60 seconds
			},
		}

		result := plugin.isUserDMDueForPoll(userInfo, now)

		require.True(t, result)
	})

	t.Run("returns false when not enough time has passed", func(t *testing.T) {
		now := time.Now().Unix()
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			LastPollAt:       now - 30, // Only 30 seconds ago
			Settings: &dataminr.UserSettings{
				DMNotifications: true,
				DMPollInterval:  60,
			},
		}

		result := plugin.isUserDMDueForPoll(userInfo, now)

		require.False(t, result)
	})

	t.Run("returns true when never polled before", func(t *testing.T) {
		now := time.Now().Unix()
		userInfo := &dataminr.UserInfo{
			MattermostUserID: "user123",
			LastPollAt:       0, // Never polled
			Settings: &dataminr.UserSettings{
				DMNotifications: true,
				DMPollInterval:  60,
			},
		}

		result := plugin.isUserDMDueForPoll(userInfo, now)

		require.True(t, result)
	})
}
