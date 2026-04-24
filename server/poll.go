package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr/client"
)

// HandlePoll handles the /dataminr poll command logic
// It manually triggers polling and posts alerts to the appropriate destination:
// - If in a channel with a subscription: posts alerts to that channel
// - If in a DM with the bot or no subscription: sends alerts as DMs (if DM enabled)
func (p *Plugin) HandlePoll(userID, channelID string) (*model.CommandResponse, error) {
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

	// Check if user has a subscription for this channel
	subs, err := p.getSubscriptions()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch subscriptions. Please try again.",
		}, nil
	}

	// Find user's subscription for this channel
	userSubs := subs.GetByDataminrUser(userID)
	var hasChannelSub bool
	for _, sub := range userSubs {
		if sub.ChannelID == channelID && sub.Enabled {
			hasChannelSub = true
			break
		}
	}

	// Determine the polling target
	if hasChannelSub {
		// Poll and post to this channel
		return p.pollAndPostToChannel(userID, channelID)
	}

	// Check if DM notifications are enabled
	if userInfo.Settings != nil && userInfo.Settings.DMNotifications {
		// Poll and send as DMs
		return p.pollAndSendDMs(userID)
	}

	// No valid target for polling
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "❌ No subscription found for this channel and DM notifications are disabled.\n\nEither:\n• Use `/dataminr subscribe` to subscribe this channel\n• Use `/dataminr dm on` to enable DM notifications",
	}, nil
}

// pollAndPostToChannel fetches alerts and posts them to a specific channel
func (p *Plugin) pollAndPostToChannel(userID, channelID string) (*model.CommandResponse, error) {
	// Get cursor for pagination
	cursor, err := p.getDataminrCursor(userID)
	if err != nil {
		// Log but continue - will start from beginning
		p.API.LogWarn("Failed to get cursor for manual poll", "user_id", userID, "error", err.Error())
	}

	// Get user credentials
	credentials, err := p.getDataminrCredentials(userID)
	if err != nil {
		p.API.LogError("Failed to get Dataminr credentials", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to retrieve credentials. Please try reconnecting with `/dataminr connect`.",
		}, nil
	}

	// Create API client and get token
	apiClient := client.NewClient(credentials, dataminrAPIBaseURL)
	token, err := apiClient.GetToken()
	if err != nil {
		p.API.LogError("Failed to get Dataminr token", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to authenticate with Dataminr. Please check your credentials and try reconnecting.",
		}, nil
	}

	// Fetch alerts
	alertResp, err := apiClient.GetAlerts(token, cursor)
	if err != nil {
		p.API.LogError("Failed to fetch Dataminr alerts", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch alerts from Dataminr. Please try again later.",
		}, nil
	}

	// Deduplicate alerts
	newAlerts := p.deduplicateAlerts(alertResp.Alerts)

	if len(newAlerts) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "✅ Poll complete. No new alerts to post.",
		}, nil
	}

	// Post each alert to the channel
	postedCount := 0
	for _, alert := range newAlerts {
		if err := p.SendAlertToChannel(channelID, &alert, userID); err != nil {
			p.API.LogWarn("Failed to post alert to channel", "alert_id", alert.AlertID, "error", err.Error())
			continue
		}
		postedCount++
	}

	// Update cursor for next poll
	if alertResp.NextPage != "" {
		if err := p.storeDataminrCursor(userID, alertResp.NextPage); err != nil {
			p.API.LogWarn("Failed to store cursor", "userID", userID, "error", err.Error())
		}
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Poll complete. Posted %d new alert(s) to this channel.", postedCount),
	}, nil
}

// pollAndSendDMs fetches alerts and sends them as DMs to the user
func (p *Plugin) pollAndSendDMs(userID string) (*model.CommandResponse, error) {
	// Get cursor for pagination
	cursor, err := p.getDataminrCursor(userID)
	if err != nil {
		// Log but continue - will start from beginning
		p.API.LogWarn("Failed to get cursor for manual DM poll", "user_id", userID, "error", err.Error())
	}

	// Get user credentials
	credentials, err := p.getDataminrCredentials(userID)
	if err != nil {
		p.API.LogError("Failed to get Dataminr credentials", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to retrieve credentials. Please try reconnecting with `/dataminr connect`.",
		}, nil
	}

	// Create API client and get token
	apiClient := client.NewClient(credentials, dataminrAPIBaseURL)
	token, err := apiClient.GetToken()
	if err != nil {
		p.API.LogError("Failed to get Dataminr token", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to authenticate with Dataminr. Please check your credentials and try reconnecting.",
		}, nil
	}

	// Fetch alerts
	alertResp, err := apiClient.GetAlerts(token, cursor)
	if err != nil {
		p.API.LogError("Failed to fetch Dataminr alerts", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch alerts from Dataminr. Please try again later.",
		}, nil
	}

	// Deduplicate alerts
	newAlerts := p.deduplicateAlerts(alertResp.Alerts)

	if len(newAlerts) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "✅ Poll complete. No new alerts to send.",
		}, nil
	}

	// Send each alert as a DM
	sentCount := 0
	for _, alert := range newAlerts {
		if err := p.SendAlertDMFromBot(userID, &alert); err != nil {
			p.API.LogWarn("Failed to send alert DM", "alert_id", alert.AlertID, "error", err.Error())
			continue
		}
		sentCount++
	}

	// Update cursor for next poll
	if alertResp.NextPage != "" {
		if err := p.storeDataminrCursor(userID, alertResp.NextPage); err != nil {
			p.API.LogWarn("Failed to store cursor", "userID", userID, "error", err.Error())
		}
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Poll complete. Sent %d new alert(s) to your DM.", sentCount),
	}, nil
}
