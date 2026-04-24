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

func TestRouteAlert(t *testing.T) {
	t.Run("routes to DM when user has DM notifications enabled", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Alert",
			},
		}

		// User info with DM notifications enabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscriptions for this user
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// Expect DM channel to be created/fetched and post to be created
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(&model.Channel{Id: "dm_channel_123"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "dm_channel_123" && post.Props["from_dataminr"] == true
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("does not route to DM when DM notifications disabled", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Alert",
			},
		}

		// User info with DM notifications disabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    false,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscriptions for this user
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// NO DM should be created - no CreatePost expectation

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("routes to subscribed channels", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Alert",
			},
		}

		// User info with DM notifications disabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    false,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Two channel subscriptions for this user
		subs := &dataminr.Subscriptions{
			Users: map[string][]*dataminr.Subscription{
				userID: {
					{ChannelID: "channel1", DataminrUser: userID, Enabled: true},
					{ChannelID: "channel2", DataminrUser: userID, Enabled: true},
				},
			},
		}
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Expect posts to both channels
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "channel1" && post.Props["from_dataminr"] == true
		})).Return(&model.Post{}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "channel2" && post.Props["from_dataminr"] == true
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("routes to both DM and channels", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
		}

		// User info with DM notifications enabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// One channel subscription
		subs := &dataminr.Subscriptions{
			Users: map[string][]*dataminr.Subscription{
				userID: {
					{ChannelID: "channel1", DataminrUser: userID, Enabled: true},
				},
			},
		}
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Expect DM
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(&model.Channel{Id: "dm_channel_123"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "dm_channel_123"
		})).Return(&model.Post{}, nil)

		// Expect channel post
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "channel1"
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("filters DM by alert type - flash filter", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		// This is a normal "Alert" type, not Flash
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Alert",
			},
		}

		// User wants only Flash alerts via DM
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "flash", // Only Flash alerts
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// NO DM should be sent because alert type doesn't match filter

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("sends Flash alert when filter is flash", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Flash Alert",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
		}

		// User wants only Flash alerts via DM
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "flash",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// Expect DM because Flash matches flash filter
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(&model.Channel{Id: "dm_channel_123"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "dm_channel_123"
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("sends Urgent alert when filter is urgent", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Urgent Alert",
			AlertType: &dataminr.AlertType{
				Name: "Urgent",
			},
		}

		// User wants only Urgent alerts via DM
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "urgent",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// Expect DM
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(&model.Channel{Id: "dm_channel_123"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "dm_channel_123"
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("handles user info not found gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
		}

		// User info not found
		api.On("KVGet", userID+"_dataminr_userinfo").Return(nil, nil)

		// Should still check subscriptions
		api.On("KVGet", subscriptionsKey).Return(nil, nil)

		// No errors, just nothing happens
		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("skips disabled subscriptions", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Alert",
			},
		}

		// User info with DM notifications disabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    false,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// One enabled, one disabled subscription
		subs := &dataminr.Subscriptions{
			Users: map[string][]*dataminr.Subscription{
				userID: {
					{ChannelID: "channel1", DataminrUser: userID, Enabled: true},
					{ChannelID: "channel2", DataminrUser: userID, Enabled: false}, // Disabled
				},
			},
		}
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Only channel1 should receive post
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "channel1"
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		require.NoError(t, err)
	})

	t.Run("continues routing on DM post failure", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Alert",
			},
		}

		// User info with DM notifications enabled
		userInfo := &dataminr.UserInfo{
			MattermostUserID: userID,
			Settings: &dataminr.UserSettings{
				DMNotifications:    true,
				NotificationFilter: "all",
			},
		}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// One channel subscription
		subs := &dataminr.Subscriptions{
			Users: map[string][]*dataminr.Subscription{
				userID: {
					{ChannelID: "channel1", DataminrUser: userID, Enabled: true},
				},
			},
		}
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// DM fails
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(nil, &model.AppError{Message: "DM error"})

		// But channel post should still succeed
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "channel1"
		})).Return(&model.Post{}, nil)

		err := plugin.routeAlert(userID, alert)

		// Should succeed even though DM failed
		require.NoError(t, err)
	})
}

func TestAlertMatchesFilter(t *testing.T) {
	t.Run("all filter matches all alert types", func(t *testing.T) {
		plugin := &Plugin{}

		assert.True(t, plugin.alertMatchesFilter("Flash", "all"))
		assert.True(t, plugin.alertMatchesFilter("Urgent", "all"))
		assert.True(t, plugin.alertMatchesFilter("Alert", "all"))
		assert.True(t, plugin.alertMatchesFilter("Unknown", "all"))
	})

	t.Run("flash filter matches only Flash", func(t *testing.T) {
		plugin := &Plugin{}

		assert.True(t, plugin.alertMatchesFilter("Flash", "flash"))
		assert.False(t, plugin.alertMatchesFilter("Urgent", "flash"))
		assert.False(t, plugin.alertMatchesFilter("Alert", "flash"))
	})

	t.Run("urgent filter matches only Urgent", func(t *testing.T) {
		plugin := &Plugin{}

		assert.False(t, plugin.alertMatchesFilter("Flash", "urgent"))
		assert.True(t, plugin.alertMatchesFilter("Urgent", "urgent"))
		assert.False(t, plugin.alertMatchesFilter("Alert", "urgent"))
	})

	t.Run("flash_urgent filter matches Flash and Urgent", func(t *testing.T) {
		plugin := &Plugin{}

		assert.True(t, plugin.alertMatchesFilter("Flash", "flash_urgent"))
		assert.True(t, plugin.alertMatchesFilter("Urgent", "flash_urgent"))
		assert.False(t, plugin.alertMatchesFilter("Alert", "flash_urgent"))
	})

	t.Run("empty filter defaults to all", func(t *testing.T) {
		plugin := &Plugin{}

		assert.True(t, plugin.alertMatchesFilter("Flash", ""))
		assert.True(t, plugin.alertMatchesFilter("Urgent", ""))
		assert.True(t, plugin.alertMatchesFilter("Alert", ""))
	})
}

func TestSendAlertDM(t *testing.T) {
	t.Run("creates DM and posts alert", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
		}

		// Get bot user ID
		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "dm_channel" &&
				post.Props["from_dataminr"] == true &&
				post.Props["alert_id"] == "alert1"
		})).Return(&model.Post{}, nil)

		err := plugin.sendAlertDM(userID, alert)

		require.NoError(t, err)
	})

	t.Run("returns error when GetDirectChannel fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
		}

		api.On("GetDirectChannel", userID, mock.AnythingOfType("string")).Return(nil, &model.AppError{Message: "channel error"})

		err := plugin.sendAlertDM(userID, alert)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel")
	})
}

func TestPostAlertToChannel(t *testing.T) {
	t.Run("creates post in channel", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		channelID := "channel123"
		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Urgent",
			},
		}

		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == channelID &&
				post.Props["from_dataminr"] == true &&
				post.Props["alert_id"] == "alert1" &&
				post.Props["dataminr_user"] == userID
		})).Return(&model.Post{}, nil)

		err := plugin.postAlertToChannel(channelID, alert, userID)

		require.NoError(t, err)
	})

	t.Run("returns error when CreatePost fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		channelID := "channel123"
		userID := "user123"
		alert := &dataminr.Alert{
			AlertID:  "alert1",
			Headline: "Test Alert",
		}

		api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, &model.AppError{Message: "post error"})

		err := plugin.postAlertToChannel(channelID, alert, userID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "post")
	})
}
