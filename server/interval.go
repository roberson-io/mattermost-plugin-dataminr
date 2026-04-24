package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// HandleChannelInterval handles the /dataminr channel-interval command logic
func (p *Plugin) HandleChannelInterval(userID, channelID string, interval int) (*model.CommandResponse, error) {
	// Check if user is connected to Dataminr
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to check connection status. Please try again.",
		}, nil
	}

	if userInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are not connected to Dataminr. Use `/dataminr connect <client_id> <client_secret>` to connect your account.",
		}, nil
	}

	// Validate interval
	if validationErr := p.validatePollInterval(interval); validationErr != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("❌ %s", validationErr.Error()),
		}, nil
	}

	// Get subscriptions and find user's subscription for this channel
	subs, err := p.getSubscriptions()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch subscriptions. Please try again.",
		}, nil
	}

	// Find the user's subscription for this channel
	userSubs := subs.GetByDataminrUser(userID)
	var found bool
	for _, sub := range userSubs {
		if sub.ChannelID == channelID {
			sub.PollInterval = interval
			found = true
			break
		}
	}

	if !found {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ No subscription found for this channel. Use `/dataminr subscribe` first.",
		}, nil
	}

	// Store updated subscriptions
	if err := p.storeSubscriptions(subs); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to update subscription. Please try again.",
		}, nil
	}

	// Return success message
	if interval == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "✅ Polling disabled for this channel (manual only mode). Use `/dataminr latest` to fetch alerts manually.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Polling interval set to %d seconds for this channel.", interval),
	}, nil
}

// HandleDMInterval handles the /dataminr dm-interval command logic
func (p *Plugin) HandleDMInterval(userID string, interval int) (*model.CommandResponse, error) {
	// Check if user is connected to Dataminr
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to check connection status. Please try again.",
		}, nil
	}

	if userInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are not connected to Dataminr. Use `/dataminr connect <client_id> <client_secret>` to connect your account.",
		}, nil
	}

	// Validate interval
	if validationErr := p.validatePollInterval(interval); validationErr != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("❌ %s", validationErr.Error()),
		}, nil
	}

	// Update user settings
	if userInfo.Settings == nil {
		userInfo.Settings = &dataminr.UserSettings{
			DMNotifications:    true,
			NotificationFilter: "all",
		}
	}
	userInfo.Settings.DMPollInterval = interval

	// Store updated user info
	if err := p.storeUserInfo(userInfo); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to update settings. Please try again.",
		}, nil
	}

	// Return success message
	if interval == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "✅ DM polling disabled (manual only mode). Use `/dataminr latest` to fetch alerts manually.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ DM polling interval set to %d seconds.", interval),
	}, nil
}
