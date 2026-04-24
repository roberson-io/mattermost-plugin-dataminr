package alerts

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// alertEmojis maps alert types to their display emojis
var alertEmojis = map[string]string{
	"Flash":  "🔴",
	"Urgent": "🟠",
	"Alert":  "🟡",
}

// alertColors maps alert types to their attachment sidebar colors
var alertColors = map[string]string{
	"Flash":  "#FF0000", // Red
	"Urgent": "#FF9900", // Orange
	"Alert":  "#FFFF00", // Yellow
}

// countryMap maps common country codes to full names for hashtags
var countryMap = map[string]string{
	"USA": "UnitedStates",
	"US":  "UnitedStates",
	"UK":  "UnitedKingdom",
	"CA":  "Canada",
	"MX":  "Mexico",
	"BR":  "Brazil",
	"FR":  "France",
	"DE":  "Germany",
	"IT":  "Italy",
	"ES":  "Spain",
	"NL":  "Netherlands",
	"BE":  "Belgium",
	"CH":  "Switzerland",
	"AT":  "Austria",
	"PL":  "Poland",
	"RU":  "Russia",
	"UA":  "Ukraine",
	"TR":  "Turkey",
	"GR":  "Greece",
	"CN":  "China",
	"JP":  "Japan",
	"KR":  "SouthKorea",
	"IN":  "India",
	"PK":  "Pakistan",
	"TH":  "Thailand",
	"VN":  "Vietnam",
	"PH":  "Philippines",
	"ID":  "Indonesia",
	"MY":  "Malaysia",
	"SG":  "Singapore",
	"AU":  "Australia",
	"NZ":  "NewZealand",
	"ZA":  "SouthAfrica",
	"EG":  "Egypt",
	"NG":  "Nigeria",
	"IL":  "Israel",
	"SA":  "SaudiArabia",
	"AE":  "UnitedArabEmirates",
	"IQ":  "Iraq",
	"IR":  "Iran",
	"AF":  "Afghanistan",
	"SY":  "Syria",
}

// fullCountryNames for matching in location strings
var fullCountryNames = []string{
	"Afghanistan", "Argentina", "Australia", "Austria", "Bangladesh", "Belgium",
	"Brazil", "Canada", "Chile", "China", "Colombia", "Egypt", "France",
	"Germany", "Greece", "India", "Indonesia", "Iran", "Iraq", "Ireland",
	"Israel", "Italy", "Japan", "Kenya", "Malaysia", "Mexico", "Netherlands",
	"Nigeria", "Pakistan", "Peru", "Philippines", "Poland", "Portugal",
	"Russia", "Singapore", "Spain", "Sweden", "Switzerland", "Syria",
	"Thailand", "Turkey", "Ukraine", "Venezuela", "Vietnam",
}

// Priority represents a Mattermost post priority
type Priority struct {
	Priority string `json:"priority"`
}

// GetAlertEmoji returns the appropriate emoji for an alert type
func GetAlertEmoji(alertType string) string {
	if emoji, ok := alertEmojis[alertType]; ok {
		return emoji
	}
	return "🟡" // Default to yellow for unknown types
}

// GetAlertColor returns the appropriate color hex code for an alert type
func GetAlertColor(alertType string) string {
	if color, ok := alertColors[alertType]; ok {
		return color
	}
	return "#808080" // Default to gray for unknown types
}

// GetAlertPriority returns the Mattermost priority for an alert type
// Returns nil for regular alerts that don't need priority
func GetAlertPriority(alertType string) *Priority {
	switch alertType {
	case "Flash":
		return &Priority{Priority: "urgent"}
	case "Urgent":
		return &Priority{Priority: "important"}
	default:
		return nil
	}
}

// ExtractCountryFromLocation extracts a country name from a location string
func ExtractCountryFromLocation(location string) string {
	if location == "" {
		return ""
	}

	parts := strings.Split(location, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		partUpper := strings.ToUpper(part)

		// Check for country codes
		if country, ok := countryMap[partUpper]; ok {
			// Special case: CA could be California if USA is also present
			if partUpper == "CA" {
				for _, p := range parts {
					pUpper := strings.ToUpper(strings.TrimSpace(p))
					if pUpper == "USA" || pUpper == "US" {
						continue // Skip CA, it's California
					}
				}
				// Check if USA/US is in the location
				if strings.Contains(strings.ToUpper(location), "USA") || strings.Contains(strings.ToUpper(location), ", US") {
					continue
				}
			}
			return country
		}

		// Check for full country names
		for _, countryName := range fullCountryNames {
			if strings.EqualFold(part, countryName) {
				return strings.ReplaceAll(countryName, " ", "")
			}
		}
	}

	return ""
}

// GenerateHashtags generates searchable hashtags from an alert
func GenerateHashtags(alert *dataminr.Alert) string {
	var tags []string
	seen := make(map[string]bool)

	addTag := func(tag string) {
		tagLower := strings.ToLower(tag)
		if !seen[tagLower] {
			tags = append(tags, tag)
			seen[tagLower] = true
		}
	}

	// 1. Alert level (always first)
	alertType := "Alert"
	if alert.AlertType != nil && alert.AlertType.Name != "" {
		alertType = alert.AlertType.Name
	}
	addTag("#" + alertType)

	// 2. Country from location
	if alert.EstimatedEventLocation != nil && alert.EstimatedEventLocation.Name != "" {
		country := ExtractCountryFromLocation(alert.EstimatedEventLocation.Name)
		if country != "" {
			addTag("#" + country)
		}
	}

	// 3. Topics - split on " - " and " and "
	for _, topic := range alert.AlertTopics {
		// Split on " - " (category separator)
		for segment := range strings.SplitSeq(topic.Name, " - ") {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}

			// Split on " and " for most phrases
			for part := range strings.SplitSeq(segment, " and ") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				// CamelCase multi-word phrases
				tag := toCamelCase(part)
				addTag("#" + tag)
			}
		}
	}

	if len(tags) == 0 {
		return ""
	}

	return "🏷️ " + strings.Join(tags, ", ")
}

// toCamelCase converts a phrase to CamelCase
func toCamelCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// FormatAlertAttachment creates a Mattermost SlackAttachment for an alert
func FormatAlertAttachment(alert *dataminr.Alert) *model.SlackAttachment {
	alertTypeName := "Alert"
	if alert.AlertType != nil && alert.AlertType.Name != "" {
		alertTypeName = alert.AlertType.Name
	}

	emoji := GetAlertEmoji(alertTypeName)
	color := GetAlertColor(alertTypeName)

	attachment := &model.SlackAttachment{
		Color:    color,
		Pretext:  fmt.Sprintf("%s **%s**", emoji, strings.ToUpper(alertTypeName)),
		Title:    alert.Headline,
		Fields:   []*model.SlackAttachmentField{},
		Fallback: fmt.Sprintf("[%s] %s", alertTypeName, alert.Headline),
	}

	// Title link
	if alert.DataminrAlertURL != "" {
		attachment.TitleLink = alert.DataminrAlertURL
	}

	// Event Time field
	if alert.AlertTimestamp != "" {
		formattedTime := formatTimestamp(alert.AlertTimestamp)
		if formattedTime != "" {
			attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
				Title: "Event Time",
				Value: formattedTime,
				Short: true,
			})
		}
	}

	// Alert Type field
	attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
		Title: "Alert Type",
		Value: alertTypeName,
		Short: true,
	})

	// SubHeadline (Additional Context)
	if alert.SubHeadline != nil && (alert.SubHeadline.Title != "" || len(alert.SubHeadline.Content) > 0) {
		var subText string
		if alert.SubHeadline.Title != "" {
			subText = "**" + alert.SubHeadline.Title + "**\n"
		}
		if len(alert.SubHeadline.Content) > 0 {
			subText += strings.Join(alert.SubHeadline.Content, " ")
		}
		attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
			Title: "Additional Context",
			Value: subText,
			Short: false,
		})
	}

	// Location
	if alert.EstimatedEventLocation != nil {
		if alert.EstimatedEventLocation.Name != "" {
			locationText := alert.EstimatedEventLocation.Name
			if len(alert.EstimatedEventLocation.Coordinates) >= 2 {
				locationText += fmt.Sprintf("\n📍 %.5f, %.5f",
					alert.EstimatedEventLocation.Coordinates[0],
					alert.EstimatedEventLocation.Coordinates[1])
				if alert.EstimatedEventLocation.ProbabilityRadius > 0 {
					locationText += fmt.Sprintf(" (±%.2f mi)", alert.EstimatedEventLocation.ProbabilityRadius)
				}
			}
			attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
				Title: "Location",
				Value: locationText,
				Short: false,
			})
		}

		// MGRS
		if alert.EstimatedEventLocation.MGRS != "" {
			attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
				Title: "MGRS",
				Value: alert.EstimatedEventLocation.MGRS,
				Short: true,
			})
		}
	}

	// Topics
	if len(alert.AlertTopics) > 0 {
		var topicNames []string
		for _, topic := range alert.AlertTopics {
			topicNames = append(topicNames, "• "+topic.Name)
		}
		attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
			Title: "Topics",
			Value: strings.Join(topicNames, "\n"),
			Short: true,
		})
	}

	// Alert Lists (matched watchlists)
	if len(alert.ListsMatched) > 0 {
		var listNames []string
		for _, list := range alert.ListsMatched {
			listNames = append(listNames, "• "+list.Name)
		}
		attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
			Title: "Alert Lists",
			Value: strings.Join(listNames, "\n"),
			Short: true,
		})
	}

	// Keywords
	if len(alert.AlertReferenceTerms) > 0 {
		var terms []string
		for _, term := range alert.AlertReferenceTerms {
			terms = append(terms, term.Text)
		}
		attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
			Title: "Keywords",
			Value: strings.Join(terms, ", "),
			Short: true,
		})
	}

	// Linked Alerts
	if len(alert.LinkedAlerts) > 0 {
		totalCount := 0
		for _, la := range alert.LinkedAlerts {
			totalCount += la.Count
		}
		if totalCount > 0 {
			attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
				Title: "Related Alerts",
				Value: fmt.Sprintf("%d linked alerts in this event", totalCount),
				Short: true,
			})
		}
	}

	// Public Source Link
	if alert.PublicPost != nil && alert.PublicPost.Href != "" {
		attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
			Title: "Public Source",
			Value: fmt.Sprintf("[View Source](%s)", alert.PublicPost.Href),
			Short: true,
		})
	}

	// Live Brief (AI Summary)
	for _, brief := range alert.LiveBrief {
		if brief.Version == "current" && brief.Summary != "" {
			attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
				Title: "🤖 AI Summary",
				Value: brief.Summary,
				Short: false,
			})
			break
		}
	}

	// Footer
	attachment.Footer = fmt.Sprintf("Dataminr | Alert ID: %s", alert.AlertID)

	return attachment
}

// formatTimestamp formats a timestamp (Unix millis or ISO string) for display
func formatTimestamp(timestamp string) string {
	// Try parsing as Unix milliseconds first
	if ts, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		t := time.UnixMilli(ts)
		return t.Format("Jan 2, 2006 3:04 PM MST")
	}

	// Try parsing as ISO format
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t.Format("Jan 2, 2006 3:04 PM MST")
	}

	// Return as-is if parsing fails
	return timestamp
}

// FormatAlertPostEnhanced creates a rich Mattermost post with attachments
func FormatAlertPostEnhanced(alert *dataminr.Alert, dataminrUserID string) *model.Post {
	alertTypeName := "Alert"
	if alert.AlertType != nil && alert.AlertType.Name != "" {
		alertTypeName = alert.AlertType.Name
	}

	// Generate hashtags for the message text
	hashtags := GenerateHashtags(alert)

	// Create the attachment
	attachment := FormatAlertAttachment(alert)

	// Build post props
	props := map[string]any{
		"from_dataminr": true,
		"alert_id":      alert.AlertID,
		"alert_type":    alertTypeName,
		"dataminr_user": dataminrUserID,
		"attachments":   []*model.SlackAttachment{attachment},
	}

	// Add priority for Flash/Urgent alerts
	priority := GetAlertPriority(alertTypeName)
	if priority != nil {
		props["priority"] = map[string]any{
			"priority": priority.Priority,
		}
	}

	return &model.Post{
		Message: hashtags,
		Props:   props,
	}
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
