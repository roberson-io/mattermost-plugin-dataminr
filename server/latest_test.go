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

// Note: Alert formatting tests are now in server/alerts/formatter_test.go
// since the formatting logic has been consolidated into FormatAlertAttachment

func TestHandleLatest(t *testing.T) {
	t.Run("returns error when user not connected", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// User not connected
		api.On("KVGet", "user123"+userInfoKeyPrefix).Return(nil, nil)

		response, err := plugin.HandleLatest("user123", "channel456", 5)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("returns loading message when user is connected (async fetch)", func(t *testing.T) {
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
		api.On("KVGet", userID+userInfoKeyPrefix).Return(userInfoJSON, nil)

		// Allow any additional calls from the async goroutine
		api.On("KVGet", mock.Anything).Return(nil, nil).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil).Maybe()

		response, err := plugin.HandleLatest(userID, channelID, 5)

		require.NoError(t, err)
		require.NotNil(t, response)
		// Should return loading message immediately (async fetch)
		assert.Contains(t, response.Text, "Fetching latest")
	})
}
