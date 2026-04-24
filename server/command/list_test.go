package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteListCommand(t *testing.T) {
	t.Run("lists subscriptions in channel", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr list",
		}

		mockAPI.handleListFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			assert.Equal(t, "user123", userID)
			assert.Equal(t, "channel123", channelID)
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Subscriptions in this channel:\n- @user1",
			}, nil
		}

		response := handler.executeListCommand(args)

		require.NotNil(t, response)
		assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
		assert.Contains(t, response.Text, "Subscriptions")
	})

	t.Run("no subscriptions found", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr list",
		}

		mockAPI.handleListFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "No subscriptions found in this channel",
			}, nil
		}

		response := handler.executeListCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "No subscriptions")
	})

	t.Run("handles error from plugin", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr list",
		}

		mockAPI.handleListFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			return nil, assert.AnError
		}

		response := handler.executeListCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "error")
	})
}

func TestListCommandRouting(t *testing.T) {
	t.Run("routes list command correctly", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		listCalled := false
		mockAPI.handleListFunc = func(userID, channelID string) (*model.CommandResponse, error) {
			listCalled = true
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Subscriptions list",
			}, nil
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr list",
		}

		response, err := handler.Handle(args)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.True(t, listCalled, "HandleList should have been called")
	})
}
