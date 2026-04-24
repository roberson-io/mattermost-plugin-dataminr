package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteStatusCommand_Connected(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleStatusResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ **Connected to Dataminr**",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr status",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Connected to Dataminr")
}

func TestExecuteStatusCommand_NotConnected(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleStatusResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "⚠️ **Not connected to Dataminr**",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr status",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Not connected")
}

func TestExecuteStatusCommand_Error(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleStatusResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "❌ Failed to check status. Please try again.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr status",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Failed to check status")
}
