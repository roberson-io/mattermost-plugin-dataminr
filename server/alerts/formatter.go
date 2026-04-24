package alerts

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// alertEmojis maps alert types to their display emojis
var alertEmojis = map[string]string{
	"Flash":  "🔴",
	"Urgent": "🟠",
	"Alert":  "🟡",
}

// GetAlertEmoji returns the appropriate emoji for an alert type
func GetAlertEmoji(alertType string) string {
	if emoji, ok := alertEmojis[alertType]; ok {
		return emoji
	}
	return "🟡" // Default to yellow for unknown types
}

// FormatAlertPost formats a Dataminr alert into a Mattermost post
func FormatAlertPost(alert *dataminr.Alert, dataminrUserID string) *model.Post {
	// Get alert type name and emoji
	alertTypeName := "Alert"
	if alert.AlertType != nil && alert.AlertType.Name != "" {
		alertTypeName = alert.AlertType.Name
	}
	emoji := GetAlertEmoji(alertTypeName)

	// Build the message
	var sb strings.Builder

	// Header with emoji, type, and headline
	sb.WriteString(fmt.Sprintf("### %s [%s] %s\n\n", emoji, alertTypeName, alert.Headline))

	// SubHeadline if present
	if alert.SubHeadline != nil {
		sb.WriteString(fmt.Sprintf("**%s**: %s\n\n", alert.SubHeadline.Title, strings.Join(alert.SubHeadline.Content, " ")))
	}

	// Location if present
	if alert.EstimatedEventLocation != nil {
		sb.WriteString(fmt.Sprintf("**Location**: %s\n", alert.EstimatedEventLocation.Name))
		if len(alert.EstimatedEventLocation.Coordinates) >= 2 {
			sb.WriteString(fmt.Sprintf("**Coordinates**: %.5f, %.5f",
				alert.EstimatedEventLocation.Coordinates[0],
				alert.EstimatedEventLocation.Coordinates[1]))
			if alert.EstimatedEventLocation.ProbabilityRadius > 0 {
				sb.WriteString(fmt.Sprintf(" (±%.2f mi)", alert.EstimatedEventLocation.ProbabilityRadius))
			}
			sb.WriteString("\n")
		}
	}

	// Topics if present
	if len(alert.AlertTopics) > 0 {
		var topics []string
		for _, topic := range alert.AlertTopics {
			topics = append(topics, topic.Name)
		}
		sb.WriteString(fmt.Sprintf("**Topics**: %s\n", strings.Join(topics, ", ")))
	}

	// Links
	var links []string
	if alert.DataminrAlertURL != "" {
		links = append(links, fmt.Sprintf("[View in Dataminr](%s)", alert.DataminrAlertURL))
	}
	if alert.PublicPost != nil && alert.PublicPost.Href != "" {
		links = append(links, fmt.Sprintf("[Source Post](%s)", alert.PublicPost.Href))
	}
	if len(links) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s\n", strings.Join(links, " | ")))
	}

	// Live Brief (AI summary) if present and current
	if len(alert.LiveBrief) > 0 {
		for _, brief := range alert.LiveBrief {
			if brief.Version == "current" && brief.Summary != "" {
				sb.WriteString(fmt.Sprintf("\n**Summary**: %s\n", brief.Summary))
				break
			}
		}
	}

	return &model.Post{
		Message: sb.String(),
		Props: map[string]any{
			"from_dataminr": true,
			"alert_id":      alert.AlertID,
			"alert_type":    alertTypeName,
			"dataminr_user": dataminrUserID,
		},
	}
}
