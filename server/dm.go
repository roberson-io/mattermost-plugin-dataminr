package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// HandleDM handles the /dataminr dm command to enable/disable DM notifications
func (p *Plugin) HandleDM(userID string, enabled bool) (*model.CommandResponse, error) {
	// Get user info
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred while checking your connection status. Please try again.",
		}, nil
	}

	if userInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are not connected to Dataminr. Use `/dataminr connect <client_id> <client_secret>` to connect your account.",
		}, nil
	}

	// Ensure settings exists
	if userInfo.Settings == nil {
		userInfo.Settings = &dataminr.UserSettings{
			DMNotifications:    true,
			NotificationFilter: "all",
		}
	}

	// Update DM notifications setting
	userInfo.Settings.DMNotifications = enabled

	// Save updated user info
	if err := p.storeUserInfo(userInfo); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred while saving your settings. Please try again.",
		}, nil
	}

	var statusText string
	if enabled {
		statusText = "✅ DM notifications **enabled**. You will receive alerts via direct message."
	} else {
		statusText = "✅ DM notifications **disabled**. You will no longer receive alerts via direct message."
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         statusText,
	}, nil
}

// HandleFilter handles the /dataminr filter command to set notification filter
func (p *Plugin) HandleFilter(userID string, filter string) (*model.CommandResponse, error) {
	// Get user info
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred while checking your connection status. Please try again.",
		}, nil
	}

	if userInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are not connected to Dataminr. Use `/dataminr connect <client_id> <client_secret>` to connect your account.",
		}, nil
	}

	// Ensure settings exists
	if userInfo.Settings == nil {
		userInfo.Settings = &dataminr.UserSettings{
			DMNotifications:    true,
			NotificationFilter: "all",
		}
	}

	// Update notification filter
	userInfo.Settings.NotificationFilter = filter

	// Save updated user info
	if err := p.storeUserInfo(userInfo); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred while saving your settings. Please try again.",
		}, nil
	}

	filterDescription := getFilterDescription(filter)
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Notification filter set to **%s**. %s", filter, filterDescription),
	}, nil
}

// getFilterDescription returns a human-readable description of the filter
func getFilterDescription(filter string) string {
	switch filter {
	case "all":
		return "You will receive all alert types (Flash, Urgent, and Alert)."
	case "flash":
		return "You will only receive Flash alerts (🔴 highest priority)."
	case "urgent":
		return "You will only receive Urgent alerts (🟠 high priority)."
	case "flash_urgent":
		return "You will receive Flash (🔴) and Urgent (🟠) alerts."
	default:
		return ""
	}
}
