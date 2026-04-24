package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteSubscribeCommand(t *testing.T) {
	t.Run("successful subscription to current channel", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr subscribe",
		}

		mockAPI.handleSubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			assert.Equal(t, "user123", userID)
			assert.Equal(t, "channel123", channelID)
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Successfully subscribed",
			}, nil
		}

		response := handler.executeSubscribeCommand(args)

		require.NotNil(t, response)
		assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
		assert.Contains(t, response.Text, "subscribed")
	})

	t.Run("user not connected error", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr subscribe",
		}

		mockAPI.handleSubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "You are not connected to Dataminr",
			}, nil
		}

		response := handler.executeSubscribeCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("permission denied error", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr subscribe",
		}

		mockAPI.handleSubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "You do not have permission to create subscriptions",
			}, nil
		}

		response := handler.executeSubscribeCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "permission")
	})

	t.Run("already subscribed message", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr subscribe",
		}

		mockAPI.handleSubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "This channel is already subscribed to your alerts",
			}, nil
		}

		response := handler.executeSubscribeCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "already subscribed")
	})

	t.Run("handles error from plugin", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr subscribe",
		}

		mockAPI.handleSubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return nil, assert.AnError
		}

		response := handler.executeSubscribeCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "error")
	})
}

func TestSubscribeCommandRouting(t *testing.T) {
	t.Run("routes subscribe command correctly", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		subscribeCalled := false
		mockAPI.handleSubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			subscribeCalled = true
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Subscribed",
			}, nil
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr subscribe",
		}

		response, err := handler.Handle(args)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.True(t, subscribeCalled, "HandleSubscribe should have been called")
	})
}
