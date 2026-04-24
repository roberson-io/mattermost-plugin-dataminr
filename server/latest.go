package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/alerts"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr/client"
)

const dataminrAPIBaseURL = "https://api.dataminr.com"

// HandleLatest handles the /dataminr latest command logic
func (p *Plugin) HandleLatest(userID string, count int) (*model.CommandResponse, error) {
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

	// Fetch alerts using the requested count
	alertResp, err := apiClient.GetAlertsWithPageSize(token, "", count)
	if err != nil {
		p.API.LogError("Failed to fetch Dataminr alerts", "userID", userID, "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to fetch alerts from Dataminr. Please try again later.",
		}, nil
	}

	// Format response
	if len(alertResp.Alerts) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "### Latest Dataminr Alerts\n\nNo alerts available at this time.",
		}, nil
	}

	// Build response with formatted alerts
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Latest Dataminr Alerts (%d)\n\n", len(alertResp.Alerts)))

	for i, alert := range alertResp.Alerts {
		sb.WriteString(formatAlertFull(i+1, &alert))
		sb.WriteString("\n---\n\n")
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         sb.String(),
	}, nil
}

// formatAlertFull formats a single alert with all available information
func formatAlertFull(index int, alert *dataminr.Alert) string {
	var sb strings.Builder

	// Get alert type and emoji
	alertTypeName := "Alert"
	if alert.AlertType != nil && alert.AlertType.Name != "" {
		alertTypeName = alert.AlertType.Name
	}
	emoji := alerts.GetAlertEmoji(alertTypeName)

	// Header with number, emoji, type, and headline
	sb.WriteString(fmt.Sprintf("### %d. %s [%s] %s\n\n", index, emoji, alertTypeName, alert.Headline))

	// Timestamp
	if alert.AlertTimestamp != "" {
		if ts, err := strconv.ParseInt(alert.AlertTimestamp, 10, 64); err == nil {
			t := time.UnixMilli(ts)
			sb.WriteString(fmt.Sprintf("**Time:** %s\n", t.Format("Jan 2, 2006 3:04 PM MST")))
		}
	}

	// SubHeadline (additional context)
	if alert.SubHeadline != nil {
		if alert.SubHeadline.Title != "" {
			sb.WriteString(fmt.Sprintf("**%s:** %s\n", alert.SubHeadline.Title, strings.Join(alert.SubHeadline.Content, " ")))
		}
	}

	// Location
	if alert.EstimatedEventLocation != nil {
		if alert.EstimatedEventLocation.Name != "" {
			sb.WriteString(fmt.Sprintf("**📍 Location:** %s\n", alert.EstimatedEventLocation.Name))
		}
		if len(alert.EstimatedEventLocation.Coordinates) >= 2 {
			coords := fmt.Sprintf("%.5f, %.5f", alert.EstimatedEventLocation.Coordinates[0], alert.EstimatedEventLocation.Coordinates[1])
			if alert.EstimatedEventLocation.ProbabilityRadius > 0 {
				coords += fmt.Sprintf(" (±%.2f mi)", alert.EstimatedEventLocation.ProbabilityRadius)
			}
			sb.WriteString(fmt.Sprintf("**Coordinates:** %s\n", coords))
		}
		if alert.EstimatedEventLocation.MGRS != "" {
			sb.WriteString(fmt.Sprintf("**MGRS:** %s\n", alert.EstimatedEventLocation.MGRS))
		}
	}

	// Topics
	if len(alert.AlertTopics) > 0 {
		var topicNames []string
		for _, topic := range alert.AlertTopics {
			topicNames = append(topicNames, topic.Name)
		}
		sb.WriteString(fmt.Sprintf("**Topics:** %s\n", strings.Join(topicNames, ", ")))
	}

	// Reference terms (keywords)
	if len(alert.AlertReferenceTerms) > 0 {
		var terms []string
		for _, term := range alert.AlertReferenceTerms {
			terms = append(terms, term.Text)
		}
		sb.WriteString(fmt.Sprintf("**Keywords:** %s\n", strings.Join(terms, ", ")))
	}

	// Lists matched
	if len(alert.ListsMatched) > 0 {
		var listNames []string
		for _, list := range alert.ListsMatched {
			listNames = append(listNames, list.Name)
		}
		sb.WriteString(fmt.Sprintf("**Matched Lists:** %s\n", strings.Join(listNames, ", ")))
	}

	// Linked alerts
	if len(alert.LinkedAlerts) > 0 {
		// Sum up counts from all linked alert entries
		totalCount := 0
		for _, la := range alert.LinkedAlerts {
			totalCount += la.Count
		}
		if totalCount > 0 {
			sb.WriteString(fmt.Sprintf("**Linked Alerts:** %d related alerts\n", totalCount))
		}
	}

	// Live Brief (AI summary)
	for _, brief := range alert.LiveBrief {
		if brief.Version == "current" && brief.Summary != "" {
			sb.WriteString(fmt.Sprintf("\n**🤖 AI Summary:** %s\n", brief.Summary))
			break
		}
	}

	// Links
	sb.WriteString("\n")
	if alert.DataminrAlertURL != "" {
		sb.WriteString(fmt.Sprintf("[View in Dataminr](%s)", alert.DataminrAlertURL))
	}
	if alert.PublicPost != nil && alert.PublicPost.Href != "" {
		if alert.DataminrAlertURL != "" {
			sb.WriteString(" | ")
		}
		sb.WriteString(fmt.Sprintf("[Source Post](%s)", alert.PublicPost.Href))
	}
	sb.WriteString("\n")

	return sb.String()
}
