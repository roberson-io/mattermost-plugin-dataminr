package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/alerts"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr/client"
)

const dataminrAPIBaseURL = "https://api.dataminr.com"

// HandleLatest handles the /dataminr latest command logic
func (p *Plugin) HandleLatest(userID, channelID string, count int) (*model.CommandResponse, error) {
	// Check if user is connected
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch alerts. Please try again.",
		}, nil
	}

	if userInfo == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "⚠️ You are not connected to Dataminr. Use `/dataminr connect <client_id> <client_secret>` to connect your account.",
		}, nil
	}

	// Run the fetch async to avoid blocking the UI
	go p.fetchLatestAlertsAsync(userID, channelID, count)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("⏳ Fetching latest %d alerts from Dataminr...", count),
	}, nil
}

// fetchLatestAlertsAsync fetches latest alerts and sends them as ephemeral messages (runs in goroutine)
func (p *Plugin) fetchLatestAlertsAsync(userID, channelID string, count int) {
	// Get token (from cache or refresh if needed) - this is the fast path
	token, err := p.getOrRefreshToken(userID)
	if err != nil {
		p.API.LogError("Failed to get Dataminr token", "userID", userID, "error", err.Error())
		p.sendEphemeralMessage(userID, channelID, "❌ Failed to authenticate with Dataminr. Please check your credentials and try reconnecting.")
		return
	}

	// Get credentials to create the API client for fetching alerts
	credentials, err := p.getDataminrCredentials(userID)
	if err != nil {
		p.API.LogError("Failed to get Dataminr credentials", "userID", userID, "error", err.Error())
		p.sendEphemeralMessage(userID, channelID, "❌ Failed to retrieve credentials. Please try reconnecting with `/dataminr connect`.")
		return
	}

	// Create API client for fetching alerts (token already obtained)
	apiClient := client.NewClient(credentials, dataminrAPIBaseURL)

	// Fetch alerts using the requested count
	alertResp, err := apiClient.GetAlertsWithPageSize(token, "", count)
	if err != nil {
		p.API.LogError("Failed to fetch Dataminr alerts", "userID", userID, "error", err.Error())
		p.sendEphemeralMessage(userID, channelID, "❌ Failed to fetch alerts from Dataminr. Please try again later.")
		return
	}

	// Format response
	if len(alertResp.Alerts) == 0 {
		p.sendEphemeralMessage(userID, channelID, "### Latest Dataminr Alerts\n\nNo alerts available at this time.")
		return
	}

	// Build response with formatted alerts using enhanced attachments
	attachments := make([]*model.SlackAttachment, 0, len(alertResp.Alerts))
	for i := range alertResp.Alerts {
		attachment := alerts.FormatAlertAttachment(&alertResp.Alerts[i])
		attachments = append(attachments, attachment)
	}

	// Send ephemeral post with attachments
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelID,
		Message:   fmt.Sprintf("## Latest Dataminr Alerts (%d)", len(alertResp.Alerts)),
	}
	post.AddProp("attachments", attachments)
	p.API.SendEphemeralPost(userID, post)
}
