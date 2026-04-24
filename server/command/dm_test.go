package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteDMCommand(t *testing.T) {
	t.Run("enables DM notifications with 'on' argument", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr dm on",
		}

		mockAPI.handleDMFunc = func(userID string, enabled bool) (*model.CommandResponse, error) {
			assert.Equal(t, "user123", userID)
			assert.True(t, enabled)
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "DM notifications enabled",
			}, nil
		}

		response := handler.executeDMCommand(args)

		require.NotNil(t, response)
		assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
		assert.Contains(t, response.Text, "enabled")
	})

	t.Run("disables DM notifications with 'off' argument", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr dm off",
		}

		mockAPI.handleDMFunc = func(userID string, enabled bool) (*model.CommandResponse, error) {
			assert.Equal(t, "user123", userID)
			assert.False(t, enabled)
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "DM notifications disabled",
			}, nil
		}

		response := handler.executeDMCommand(args)

		require.NotNil(t, response)
		assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
		assert.Contains(t, response.Text, "disabled")
	})

	t.Run("returns error for invalid argument", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr dm invalid",
		}

		response := handler.executeDMCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "on")
		assert.Contains(t, response.Text, "off")
	})

	t.Run("returns error when no argument provided", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr dm",
		}

		response := handler.executeDMCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "on")
		assert.Contains(t, response.Text, "off")
	})

	t.Run("handles error from plugin", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr dm on",
		}

		mockAPI.handleDMFunc = func(userID string, enabled bool) (*model.CommandResponse, error) {
			return nil, assert.AnError
		}

		response := handler.executeDMCommand(args)

		require.NotNil(t, response)
		assert.Contains(t, response.Text, "error")
	})
}

func TestDMCommandRouting(t *testing.T) {
	t.Run("routes dm command correctly", func(t *testing.T) {
		mockAPI := &mockPluginAPI{}
		handler := &Handler{
			pluginAPI: mockAPI,
		}

		dmCalled := false
		mockAPI.handleDMFunc = func(userID string, enabled bool) (*model.CommandResponse, error) {
			dmCalled = true
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "DM notifications enabled",
			}, nil
		}

		args := &model.CommandArgs{
			UserId:    "user123",
			ChannelId: "channel123",
			Command:   "/dataminr dm on",
		}

		response, err := handler.Handle(args)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.True(t, dmCalled, "HandleDM should have been called")
	})
}
