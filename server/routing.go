package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/alerts"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// routeAlert routes an alert to the appropriate destinations (DM and/or channels)
func (p *Plugin) routeAlert(dataminrUserID string, alert *dataminr.Alert) error {
	// Get user info to check DM preferences
	userInfo, err := p.getUserInfo(dataminrUserID)
	if err != nil {
		return errors.Wrap(err, "failed to get user info")
	}

	// Get alert type for filtering
	alertType := "Alert"
	if alert.AlertType != nil && alert.AlertType.Name != "" {
		alertType = alert.AlertType.Name
	}

	// Route to DM if enabled and alert matches filter
	if userInfo != nil && userInfo.Settings != nil && userInfo.Settings.DMNotifications {
		filter := userInfo.Settings.NotificationFilter
		if p.alertMatchesFilter(alertType, filter) {
			// Send DM but don't fail the whole routing if DM fails
			_ = p.sendAlertDM(dataminrUserID, alert)
		}
	}

	// Route to subscribed channels
	subs, err := p.getSubscriptions()
	if err != nil {
		return errors.Wrap(err, "failed to get subscriptions")
	}

	userSubs := subs.GetByDataminrUser(dataminrUserID)
	for _, sub := range userSubs {
		if !sub.Enabled {
			continue
		}
		// Post to channel (ignore individual channel errors)
		_ = p.postAlertToChannel(sub.ChannelID, alert, dataminrUserID)
	}

	return nil
}

// alertMatchesFilter checks if an alert type matches the user's notification filter
func (p *Plugin) alertMatchesFilter(alertType string, filter string) bool {
	// Empty filter or "all" matches everything
	if filter == "" || filter == "all" {
		return true
	}

	// Normalize to lowercase for comparison
	alertTypeLower := strings.ToLower(alertType)
	filterLower := strings.ToLower(filter)

	switch filterLower {
	case "flash":
		return alertTypeLower == "flash"
	case "urgent":
		return alertTypeLower == "urgent"
	case "flash_urgent":
		return alertTypeLower == "flash" || alertTypeLower == "urgent"
	default:
		// Unknown filter, default to matching all
		return true
	}
}

// sendAlertDM sends an alert to a user via DM from the bot
func (p *Plugin) sendAlertDM(userID string, alert *dataminr.Alert) error {
	// Get or create DM channel between user and bot
	dmChannel, appErr := p.API.GetDirectChannel(userID, p.botUserID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get DM channel with bot")
	}

	// Format the alert as an enhanced post with attachments
	post := alerts.FormatAlertPostEnhanced(alert, userID)
	post.UserId = p.botUserID
	post.ChannelId = dmChannel.Id
	post.Type = "custom_dataminr_alert"

	// Create the post
	_, appErr = p.API.CreatePost(post)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to create DM post")
	}

	return nil
}

// postAlertToChannel posts an alert to a channel from the bot
func (p *Plugin) postAlertToChannel(channelID string, alert *dataminr.Alert, dataminrUserID string) error {
	// Format the alert as an enhanced post with attachments
	post := alerts.FormatAlertPostEnhanced(alert, dataminrUserID)
	post.UserId = p.botUserID
	post.ChannelId = channelID
	post.Type = "custom_dataminr_alert"

	// Create the post
	_, appErr := p.API.CreatePost(post)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to create channel post")
	}

	return nil
}
