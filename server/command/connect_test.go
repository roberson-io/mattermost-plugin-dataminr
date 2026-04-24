package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteConnectCommand_MissingArguments(t *testing.T) {
	env := setupTest()
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "no arguments",
			command: "/dataminr connect",
		},
		{
			name:    "only client_id",
			command: "/dataminr connect my_client_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := &model.CommandArgs{
				Command: tt.command,
				UserId:  "user123",
			}

			response, err := handler.Handle(args)
			require.NoError(t, err)
			assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
			assert.Contains(t, response.Text, "Usage")
			assert.Contains(t, response.Text, "client_id")
			assert.Contains(t, response.Text, "client_secret")
		})
	}
}

func TestExecuteConnectCommand_AlreadyConnected(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleConnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "⚠️ You are already connected to Dataminr.",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr connect new_client_id new_client_secret",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "already connected")
}

func TestExecuteConnectCommand_SuccessfulConnection(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleConnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully connected to Dataminr!",
	}
	handler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr connect test_client_id test_client_secret",
		UserId:  "user123",
	}

	response, err := handler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Successfully connected")
}

// Note: Mattermost command parsing doesn't preserve shell-style quoting,
// so we can't test for truly empty string arguments. The validation for
// empty/whitespace credentials happens at the API level when authenticating.
