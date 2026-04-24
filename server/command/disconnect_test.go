package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteDisconnectCommand_Success(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleDisconnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully disconnected from Dataminr.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr disconnect",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Successfully disconnected")
}

func TestExecuteDisconnectCommand_NotConnected(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleDisconnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "⚠️ You are not connected to Dataminr.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr disconnect",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "not connected")
}

func TestExecuteDisconnectCommand_Error(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleDisconnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "❌ Failed to disconnect. Please try again.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr disconnect",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Failed to disconnect")
}
