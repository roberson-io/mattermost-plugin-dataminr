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

func TestSendAlertToChannel(t *testing.T) {
	t.Run("posts alert to channel as bot", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		channelID := "channel456"
		dataminrUserID := "user123"

		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test Alert Headline",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
		}

		// Expect post creation
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == channelID &&
				post.UserId == "bot123" &&
				post.Props["from_dataminr"] == true &&
				post.Props["alert_id"] == "alert-123"
		})).Return(&model.Post{}, nil)

		err := plugin.SendAlertToChannel(channelID, alert, dataminrUserID)

		require.NoError(t, err)
	})

	t.Run("returns error when CreatePost fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		channelID := "channel456"
		dataminrUserID := "user123"

		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test Alert Headline",
		}

		// CreatePost fails
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "failed to create post"})

		err := plugin.SendAlertToChannel(channelID, alert, dataminrUserID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create")
	})
}

func TestPollAndPostToChannel(t *testing.T) {
	t.Run("returns error when credentials not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// Get cursor first
		api.On("KVGet", userID+cursorKeyPrefix).Return(nil, nil)

		// Token caching: check for cached token timestamp (none found)
		api.On("KVGet", userID+"_dataminr_token_issued_at").Return(nil, nil)

		// No credentials stored - getOrRefreshToken will fail
		api.On("KVGet", userID+"_dataminr_credentials").Return(nil, nil)

		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

		response, err := plugin.pollAndPostToChannel(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "authenticate")
	})
}

func TestPollAndSendDMs(t *testing.T) {
	t.Run("returns error when credentials not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"

		// Get cursor first
		api.On("KVGet", userID+cursorKeyPrefix).Return(nil, nil)

		// Token caching: check for cached token timestamp (none found)
		api.On("KVGet", userID+"_dataminr_token_issued_at").Return(nil, nil)

		// No credentials stored - getOrRefreshToken will fail
		api.On("KVGet", userID+"_dataminr_credentials").Return(nil, nil)

		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

		response, err := plugin.pollAndSendDMs(userID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "authenticate")
	})
}

func TestHandlePoll(t *testing.T) {
	t.Run("user not connected returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is not connected (no user info)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(nil, nil)

		response, err := plugin.HandlePoll(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("returns loading message when subscription exists (async polling)", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := dataminr.NewUserInfo(userID)
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Subscription exists for this channel
		subs := dataminr.NewSubscriptions()
		subs.Add(dataminr.NewSubscription(channelID, userID, userID))
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Allow any additional calls from the async goroutine
		api.On("KVGet", mock.Anything).Return(nil, nil).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil).Maybe()

		response, err := plugin.HandlePoll(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		// Should return loading message immediately (async polling)
		assert.Contains(t, response.Text, "Polling Dataminr")
	})

	t.Run("returns loading message when DM enabled and no subscription (async polling)", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected with DM notifications enabled
		userInfo := dataminr.NewUserInfo(userID)
		userInfo.Settings.DMNotifications = true
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscription for this channel
		subs := dataminr.NewSubscriptions()
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Allow any additional calls from the async goroutine
		api.On("KVGet", mock.Anything).Return(nil, nil).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil).Maybe()

		response, err := plugin.HandlePoll(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		// Should return loading message immediately (async polling)
		assert.Contains(t, response.Text, "Polling Dataminr")
	})

	t.Run("no subscription and DM disabled returns error", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected but DM notifications disabled
		userInfo := dataminr.NewUserInfo(userID)
		userInfo.Settings.DMNotifications = false
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscription for this channel
		subs := dataminr.NewSubscriptions()
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		response, err := plugin.HandlePoll(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		// Should indicate no valid target for polling
		assert.Contains(t, response.Text, "No subscription")
	})

	t.Run("any channel without subscription falls back to DM if enabled (async)", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		anyChannelID := "any_channel"

		// User is connected with DM notifications enabled
		userInfo := dataminr.NewUserInfo(userID)
		userInfo.Settings.DMNotifications = true
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscription for this channel
		subs := dataminr.NewSubscriptions()
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Allow any additional calls from the async goroutine
		api.On("KVGet", mock.Anything).Return(nil, nil).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil).Maybe()

		response, err := plugin.HandlePoll(userID, anyChannelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		// Should return loading message immediately (async polling)
		assert.Contains(t, response.Text, "Polling Dataminr")
	})

	t.Run("returns loading message when subscription exists (async channel polling)", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := dataminr.NewUserInfo(userID)
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Subscription exists for this channel
		subs := dataminr.NewSubscriptions()
		subs.Add(dataminr.NewSubscription(channelID, userID, userID))
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", subscriptionsKey).Return(subsJSON, nil)

		// Allow any additional calls from the async goroutine
		api.On("KVGet", mock.Anything).Return(nil, nil).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil).Maybe()

		response, err := plugin.HandlePoll(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		// Should return loading message immediately (async polling)
		assert.Contains(t, response.Text, "Polling Dataminr")
	})
}
