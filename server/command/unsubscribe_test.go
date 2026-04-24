package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteUnsubscribeCommand(t *testing.T) {
	t.Run("successful unsubscription", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr unsubscribe",
		}

		mockAPI.handleUnsubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			assert.Equal(t, "user123", userID)
			assert.Equal(t, "channel123", channelID)
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Successfully unsubscribed",
			}, nil
		}

		response := handler.executeUnsubscribeCommand(args)

		require.NotNil(t, response)
		assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
		assert.Contains(t, response.Text, "unsubscribed")
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr unsubscribe",
		}

		mockAPI.handleUnsubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "No subscription found for this channel",
			}, nil
		}

		response := handler.executeUnsubscribeCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "No subscription")
	})

	t.Run("handles error from plugin", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr unsubscribe",
		}

		mockAPI.handleUnsubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return nil, assert.AnError
		}

		response := handler.executeUnsubscribeCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "error")
	})
}

func TestUnsubscribeCommandRouting(t *testing.T) {
	t.Run("routes unsubscribe command correctly", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		unsubscribeCalled := false
		mockAPI.handleUnsubscribeFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			unsubscribeCalled = true
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Unsubscribed",
			}, nil
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr unsubscribe",
		}

		response, err := handler.Handle(args)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.True(t, unsubscribeCalled, "HandleUnsubscribe should have been called")
	})
}
