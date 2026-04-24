package main

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// HandleDisconnect handles the /dataminr disconnect command logic
func (p *Plugin) HandleDisconnect(userID string) (*model.CommandResponse, error) {
	// Check if user is connected
	existingUserInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to check connection status. Please try again.",
		}, nil
	}

	if existingUserInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are not connected to Dataminr.",
		}, nil
	}

	// Delete all user data
	if err := p.deleteDataminrCredentials(userID); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to disconnect. Please try again.",
		}, nil
	}

	if err := p.deleteDataminrToken(userID); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to disconnect. Please try again.",
		}, nil
	}

	if err := p.deleteDataminrCursor(userID); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to disconnect. Please try again.",
		}, nil
	}

	if err := p.deleteUserInfo(userID); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to disconnect. Please try again.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully disconnected from Dataminr.\n\nYour credentials and all stored data have been removed.",
	}, nil
}
