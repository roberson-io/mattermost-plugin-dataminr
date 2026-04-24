package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/alerts"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

const (
	// BotUsername is the username for the Dataminr bot
	BotUsername = "dataminr"
	// BotDisplayName is the display name for the Dataminr bot
	BotDisplayName = "Dataminr"
	// BotDescription is the description for the Dataminr bot
	BotDescription = "A bot account created by the Dataminr plugin for delivering real-time alerts."
)

// GetBotUserID returns the bot user ID
func (p *Plugin) GetBotUserID() string {
	return p.botUserID
}

// CreateBotDMPost creates a DM post from the bot to a user
func (p *Plugin) CreateBotDMPost(userID, message, postType string) error {
	// Get or create direct channel between user and bot
	channel, appErr := p.API.GetDirectChannel(userID, p.botUserID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get DM channel with bot")
	}

	// Create post from bot
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channel.Id,
		Message:   message,
		Type:      postType,
	}

	_, appErr = p.API.CreatePost(post)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to create DM post")
	}

	return nil
}

// SendAlertDMFromBot sends an alert to a user via DM from the bot
func (p *Plugin) SendAlertDMFromBot(userID string, alert *dataminr.Alert) error {
	// Get or create direct channel between user and bot
	channel, appErr := p.API.GetDirectChannel(userID, p.botUserID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get DM channel with bot")
	}

	// Format the alert as an enhanced post with attachments
	post := alerts.FormatAlertPostEnhanced(alert, userID)
	post.UserId = p.botUserID
	post.ChannelId = channel.Id
	post.Type = "custom_dataminr_alert"

	// Create the post
	_, appErr = p.API.CreatePost(post)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to create alert DM post")
	}

	return nil
}

// SendAlertToChannel sends an alert to a channel from the bot
func (p *Plugin) SendAlertToChannel(channelID string, alert *dataminr.Alert, dataminrUserID string) error {
	// Format the alert as an enhanced post with attachments
	post := alerts.FormatAlertPostEnhanced(alert, dataminrUserID)
	post.UserId = p.botUserID
	post.ChannelId = channelID
	post.Type = "custom_dataminr_alert"

	// Create the post
	_, appErr := p.API.CreatePost(post)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to create alert channel post")
	}

	return nil
}
