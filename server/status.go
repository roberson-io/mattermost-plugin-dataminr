package main

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// HandleStatus handles the /dataminr status command logic
func (p *Plugin) HandleStatus(userID string) (*model.CommandResponse, error) {
	// Check if user is connected
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to check status. Please try again.",
		}, nil
	}

	if userInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ **Not connected to Dataminr**\n\nUse `/dataminr connect <client_id> <client_secret>` to connect your account.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ **Connected to Dataminr**\n\nYour account is connected and ready to receive alerts.",
	}, nil
}
