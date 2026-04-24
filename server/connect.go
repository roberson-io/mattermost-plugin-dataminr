package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// HandleConnect handles the /dataminr connect command logic
func (p *Plugin) HandleConnect(userID, clientID, clientSecret string) (*model.CommandResponse, error) {
	// Check if user is already connected
	existingUserInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to check connection status. Please try again.",
		}, nil
	}

	if existingUserInfo != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are already connected to Dataminr. Use `/dataminr disconnect` first if you want to reconnect with different credentials.",
		}, nil
	}

	// Store credentials
	credentials := &dataminr.Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	if err := p.storeDataminrCredentials(userID, credentials); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to store credentials. Please try again.",
		}, nil
	}

	// Create and store user info
	userInfo := dataminr.NewUserInfo(userID)
	if err := p.storeUserInfo(userInfo); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to store user info. Please try again.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully connected to Dataminr!\n\nYou can now use `/dataminr status` to check your connection or `/dataminr latest` to fetch alerts.",
	}, nil
}
