package main

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleChannelInterval(t *testing.T) {
	t.Run("user not connected returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"
		channelID := "channel456"

		// User is not connected (no user info)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(nil, nil)

		response, err := plugin.HandleChannelInterval(userID, channelID, 60)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("no subscription in channel returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No subscriptions exist
		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)

		response, err := plugin.HandleChannelInterval(userID, channelID, 60)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "No subscription found")
	})

	t.Run("interval below minimum returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Note: Interval validation happens before fetching subscriptions,
		// so we don't expect a KVGet for subscriptions

		response, err := plugin.HandleChannelInterval(userID, channelID, 15) // Below minimum

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "at least 30 seconds")
	})

	t.Run("interval of zero is valid (manual only mode)", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Subscription exists
		subs := dataminr.NewSubscriptions()
		subs.Add(dataminr.NewSubscription(channelID, userID, userID))
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", "dataminr_subscriptions").Return(subsJSON, nil)

		// Expect updated subscriptions to be stored
		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(nil)

		response, err := plugin.HandleChannelInterval(userID, channelID, 0)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "manual only")
	})

	t.Run("successful interval update", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Subscription exists
		subs := dataminr.NewSubscriptions()
		subs.Add(dataminr.NewSubscription(channelID, userID, userID))
		subsJSON, _ := json.Marshal(subs)
		api.On("KVGet", "dataminr_subscriptions").Return(subsJSON, nil)

		// Expect updated subscriptions to be stored
		api.On("KVSet", "dataminr_subscriptions", mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.Subscriptions
			if err := json.Unmarshal(data, &updated); err != nil {
				return false
			}
			userSubs := updated.GetByDataminrUser(userID)
			if len(userSubs) != 1 {
				return false
			}
			return userSubs[0].PollInterval == 60
		})).Return(nil)

		response, err := plugin.HandleChannelInterval(userID, channelID, 60)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "60 seconds")
	})
}

func TestHandleDMInterval(t *testing.T) {
	t.Run("user not connected returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"

		// User is not connected (no user info)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(nil, nil)

		response, err := plugin.HandleDMInterval(userID, 60)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("interval below minimum returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"

		// User is connected
		userInfo := dataminr.NewUserInfo(userID)
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		response, err := plugin.HandleDMInterval(userID, 15) // Below minimum

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "at least 30 seconds")
	})

	t.Run("interval of zero is valid (manual only mode)", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"

		// User is connected
		userInfo := dataminr.NewUserInfo(userID)
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Expect updated user info to be stored
		api.On("KVSet", userID+"_dataminr_userinfo", mock.AnythingOfType("[]uint8")).Return(nil)

		response, err := plugin.HandleDMInterval(userID, 0)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "manual only")
	})

	t.Run("successful interval update", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:   "test-encryption-key-32-bytes!!!",
			DataminrMinPollInterval: 30,
		})

		userID := "user123"

		// User is connected
		userInfo := dataminr.NewUserInfo(userID)
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Expect updated user info to be stored with new interval
		api.On("KVSet", userID+"_dataminr_userinfo", mock.MatchedBy(func(data []byte) bool {
			var updated dataminr.UserInfo
			if err := json.Unmarshal(data, &updated); err != nil {
				return false
			}
			return updated.Settings != nil && updated.Settings.DMPollInterval == 120
		})).Return(nil)

		response, err := plugin.HandleDMInterval(userID, 120)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "120 seconds")
	})
}
