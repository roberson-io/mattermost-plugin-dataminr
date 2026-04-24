package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteLatestCommand_Success(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "### Latest Dataminr Alerts\n\n**Alert 1:** Breaking news event",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Latest Dataminr Alerts")
}

func TestExecuteLatestCommand_NotConnected(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "⚠️ You are not connected to Dataminr.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "not connected")
}

func TestExecuteLatestCommand_NoAlerts(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "No new alerts available.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "No new alerts")
}

func TestExecuteLatestCommand_Error(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "❌ Failed to fetch alerts. Please try again.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Failed to fetch alerts")
}

func TestExecuteLatestCommand_WithCountDefault(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestFunc = func(userID, channelID string, count int) (*model.CommandResponse, error) {
		// Verify default count is 5
		assert.Equal(t, 5, count)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "### Latest Dataminr Alerts\n\n**Alert 1:** Breaking news event",
		}, nil
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command:   "/dataminr latest",
		UserId:    "user123",
		ChannelId: "channel456",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Latest Dataminr Alerts")
}

func TestExecuteLatestCommand_WithCountSpecified(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestFunc = func(userID, channelID string, count int) (*model.CommandResponse, error) {
		// Verify specified count is passed through
		assert.Equal(t, 20, count)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "### Latest Dataminr Alerts\n\nShowing 20 alerts",
		}, nil
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command:   "/dataminr latest 20",
		UserId:    "user123",
		ChannelId: "channel456",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
}

func TestExecuteLatestCommand_CountMaximum(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleLatestFunc = func(userID, channelID string, count int) (*model.CommandResponse, error) {
		// Verify max count of 100 is accepted
		assert.Equal(t, 100, count)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "### Latest Dataminr Alerts",
		}, nil
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command:   "/dataminr latest 100",
		UserId:    "user123",
		ChannelId: "channel456",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
}

func TestExecuteLatestCommand_CountExceedsMaximum(t *testing.T) {
	env := setupTest()
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest 101",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "must be between 1 and 100")
}

func TestExecuteLatestCommand_CountZero(t *testing.T) {
	env := setupTest()
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest 0",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "must be between 1 and 100")
}

func TestExecuteLatestCommand_CountNegative(t *testing.T) {
	env := setupTest()
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest -5",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "must be between 1 and 100")
}

func TestExecuteLatestCommand_CountNotInteger(t *testing.T) {
	env := setupTest()
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest abc",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "must be a valid number")
}

func TestExecuteLatestCommand_CountFloat(t *testing.T) {
	env := setupTest()
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr latest 5.5",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "must be a valid number")
}
