package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
)

// Tests for /dataminr channel-interval command

func TestExecuteChannelIntervalCommand_MissingArgument(t *testing.T) {
	handler := &Handler{
		pluginAPI: &mockPluginAPI{},
	}

	args := &model.CommandArgs{
		Command:   "/dataminr channel-interval",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "Usage:")
	assert.Contains(t, response.Text, "/dataminr channel-interval")
}

func TestExecuteChannelIntervalCommand_InvalidNumber(t *testing.T) {
	handler := &Handler{
		pluginAPI: &mockPluginAPI{},
	}

	args := &model.CommandArgs{
		Command:   "/dataminr channel-interval abc",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "must be a valid number")
}

func TestExecuteChannelIntervalCommand_NegativeNumber(t *testing.T) {
	handler := &Handler{
		pluginAPI: &mockPluginAPI{},
	}

	args := &model.CommandArgs{
		Command:   "/dataminr channel-interval -10",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "cannot be negative")
}

func TestExecuteChannelIntervalCommand_ValidZero(t *testing.T) {
	mock := &mockPluginAPI{
		handleChannelIntervalFunc: func(userID, channelID string, interval int) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "✅ Polling disabled for this channel (manual only mode).",
			}, nil
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr channel-interval 0",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "manual only")
}

func TestExecuteChannelIntervalCommand_ValidInterval(t *testing.T) {
	var capturedUserID, capturedChannelID string
	var capturedInterval int

	mock := &mockPluginAPI{
		handleChannelIntervalFunc: func(userID, channelID string, interval int) (*model.CommandResponse, error) {
			capturedUserID = userID
			capturedChannelID = channelID
			capturedInterval = interval
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "✅ Polling interval set to 60 seconds for this channel.",
			}, nil
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr channel-interval 60",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "60 seconds")
	assert.Equal(t, "user123", capturedUserID)
	assert.Equal(t, "channel123", capturedChannelID)
	assert.Equal(t, 60, capturedInterval)
}

func TestExecuteChannelIntervalCommand_PluginError(t *testing.T) {
	mock := &mockPluginAPI{
		handleChannelIntervalFunc: func(userID, channelID string, interval int) (*model.CommandResponse, error) {
			return nil, assert.AnError
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr channel-interval 60",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "error occurred")
}

// Tests for /dataminr dm-interval command

func TestExecuteDMIntervalCommand_MissingArgument(t *testing.T) {
	handler := &Handler{
		pluginAPI: &mockPluginAPI{},
	}

	args := &model.CommandArgs{
		Command:   "/dataminr dm-interval",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "Usage:")
	assert.Contains(t, response.Text, "/dataminr dm-interval")
}

func TestExecuteDMIntervalCommand_InvalidNumber(t *testing.T) {
	handler := &Handler{
		pluginAPI: &mockPluginAPI{},
	}

	args := &model.CommandArgs{
		Command:   "/dataminr dm-interval xyz",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "must be a valid number")
}

func TestExecuteDMIntervalCommand_NegativeNumber(t *testing.T) {
	handler := &Handler{
		pluginAPI: &mockPluginAPI{},
	}

	args := &model.CommandArgs{
		Command:   "/dataminr dm-interval -5",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "cannot be negative")
}

func TestExecuteDMIntervalCommand_ValidZero(t *testing.T) {
	mock := &mockPluginAPI{
		handleDMIntervalFunc: func(userID string, interval int) (*model.CommandResponse, error) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "✅ DM polling disabled (manual only mode).",
			}, nil
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr dm-interval 0",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "manual only")
}

func TestExecuteDMIntervalCommand_ValidInterval(t *testing.T) {
	var capturedUserID string
	var capturedInterval int

	mock := &mockPluginAPI{
		handleDMIntervalFunc: func(userID string, interval int) (*model.CommandResponse, error) {
			capturedUserID = userID
			capturedInterval = interval
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "✅ DM polling interval set to 120 seconds.",
			}, nil
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr dm-interval 120",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "120 seconds")
	assert.Equal(t, "user123", capturedUserID)
	assert.Equal(t, 120, capturedInterval)
}

func TestExecuteDMIntervalCommand_PluginError(t *testing.T) {
	mock := &mockPluginAPI{
		handleDMIntervalFunc: func(userID string, interval int) (*model.CommandResponse, error) {
			return nil, assert.AnError
		},
	}
	handler := &Handler{
		pluginAPI: mock,
	}

	args := &model.CommandArgs{
		Command:   "/dataminr dm-interval 60",
		UserId:    "user123",
		ChannelId: "channel123",
	}

	response, err := handler.Handle(args)

	assert.NoError(t, err)
	assert.Contains(t, response.Text, "error occurred")
}
