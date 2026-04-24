package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
)

// Tests for /dataminr poll command

func TestExecutePollCommand_Success(t *testing.T) {
	var capturedUserID, capturedChannelID string

	mock := &mockPluginAPI{
		handlePollFunc: func(userID, channelID string) (*model.CommandResponse, error) {
			capturedUserID = userID
			capturedChannelID = channelID
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "✅ Polling complete. 5 new alerts posted.",
			}, nil
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr poll",
		UserId:    "user123",
		ChannelId: "channel456",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "Polling complete")
	assert.Equal(t, "user123", capturedUserID)
	assert.Equal(t, "channel456", capturedChannelID)
}

func TestExecutePollCommand_PluginError(t *testing.T) {
	mock := &mockPluginAPI{
		handlePollFunc: func(userID, channelID string) (*model.CommandResponse, error) {
			return nil, assert.AnError
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr poll",
		UserId:    "user123",
		ChannelId: "channel456",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "error occurred")
}
