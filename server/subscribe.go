package main

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

// HandleSubscribe handles the /dataminr subscribe command logic
func (p *Plugin) HandleSubscribe(userID, channelID string) (*model.CommandResponse, error) {
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

	// Check if user has permission to create subscriptions in this channel
	if permErr := p.canCreateSubscription(userID, channelID); permErr != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ You do not have permission to create subscriptions in this channel.",
		}, nil
	}

	// Check if already subscribed
	subs, err := p.getSubscriptions()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to check existing subscriptions. Please try again.",
		}, nil
	}

	// Check if this user already has a subscription for this channel
	existingSubs := subs.GetByDataminrUser(userID)
	for _, sub := range existingSubs {
		if sub.ChannelID == channelID {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "ℹ️ This channel is already subscribed to your Dataminr alerts.",
			}, nil
		}
	}

	// Add the subscription
	if err := p.addSubscription(channelID, userID, userID); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to create subscription. Please try again.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully subscribed this channel to your Dataminr alerts.",
	}, nil
}

// HandleUnsubscribe handles the /dataminr unsubscribe command logic
func (p *Plugin) HandleUnsubscribe(userID, channelID string) (*model.CommandResponse, error) {
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

	// Try to remove the subscription
	removed, err := p.removeSubscription(channelID, userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to remove subscription. Please try again.",
		}, nil
	}

	if !removed {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "ℹ️ No subscription found for this channel from your Dataminr account.",
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully unsubscribed this channel from your Dataminr alerts.",
	}, nil
}

// HandleList handles the /dataminr list command logic
func (p *Plugin) HandleList(userID, channelID string) (*model.CommandResponse, error) {
	// Get all subscriptions
	subs, err := p.getSubscriptions()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch subscriptions. Please try again.",
		}, nil
	}

	// Get subscriptions for this channel
	channelSubs := subs.GetByChannel(channelID)

	if len(channelSubs) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "ℹ️ No subscriptions found in this channel.",
		}, nil
	}

	// Build the list of subscriptions
	var sb strings.Builder
	sb.WriteString("### Dataminr Subscriptions in this channel\n\n")
	for _, sub := range channelSubs {
		// Get user info for display
		user, appErr := p.API.GetUser(sub.DataminrUser)
		username := sub.DataminrUser
		if appErr == nil && user != nil {
			username = "@" + user.Username
		}

		sb.WriteString("- ")
		sb.WriteString(username)
		sb.WriteString("'s Dataminr alerts")
		if sub.CreatorID != sub.DataminrUser {
			creator, _ := p.API.GetUser(sub.CreatorID)
			if creator != nil {
				sb.WriteString(" (created by @")
				sb.WriteString(creator.Username)
				sb.WriteString(")")
			}
		}
		sb.WriteString("\n")
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         sb.String(),
	}, nil
}
