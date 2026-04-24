package alerts

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatAlertPost(t *testing.T) {
	t.Run("formats Flash alert with red emoji", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major earthquake reported",
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "🔴")
		assert.Contains(t, post.Message, "Flash")
		assert.Contains(t, post.Message, "Major earthquake reported")
	})

	t.Run("formats Urgent alert with orange emoji", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Urgent"},
			Headline:       "Severe weather warning",
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "🟠")
		assert.Contains(t, post.Message, "Urgent")
		assert.Contains(t, post.Message, "Severe weather warning")
	})

	t.Run("formats Alert type with yellow emoji", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Alert"},
			Headline:       "Traffic incident reported",
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "🟡")
		assert.Contains(t, post.Message, "Alert")
	})

	t.Run("handles missing AlertType gracefully", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      nil,
			Headline:       "Unknown alert type",
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "Unknown alert type")
		// Should use default emoji
		assert.Contains(t, post.Message, "🟡")
	})

	t.Run("includes subheadline when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			SubHeadline: &dataminr.SubHeadline{
				Title:   "Details",
				Content: []string{"First responders on scene", "Road closures expected"},
			},
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "Details")
		assert.Contains(t, post.Message, "First responders on scene")
	})

	t.Run("includes location when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Name:              "San Francisco, CA",
				Coordinates:       []float64{37.7749, -122.4194},
				ProbabilityRadius: 1.5,
			},
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "San Francisco, CA")
		assert.Contains(t, post.Message, "Location")
	})

	t.Run("includes topics when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			AlertTopics: []dataminr.AlertTopic{
				{ID: "1", Name: "Natural Disasters"},
				{ID: "2", Name: "Emergency Response"},
			},
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "Natural Disasters")
		assert.Contains(t, post.Message, "Emergency Response")
	})

	t.Run("includes Dataminr URL as link", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:          "alert123",
			AlertTimestamp:   "2026-04-23T10:00:00Z",
			AlertType:        &dataminr.AlertType{Name: "Flash"},
			Headline:         "Major incident",
			DataminrAlertURL: "https://app.dataminr.com/alerts/123",
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "[View in Dataminr]")
		assert.Contains(t, post.Message, "https://app.dataminr.com/alerts/123")
	})

	t.Run("includes source post link when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			PublicPost: &dataminr.PublicPost{
				Href: "https://twitter.com/user/status/123",
			},
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "[Source Post]")
		assert.Contains(t, post.Message, "https://twitter.com/user/status/123")
	})

	t.Run("includes live brief summary when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			LiveBrief: []dataminr.LiveBrief{
				{
					Summary:   "AI-generated summary of the event",
					Timestamp: "2026-04-23T10:05:00Z",
					Version:   "current",
				},
			},
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "AI-generated summary of the event")
	})

	t.Run("sets post props correctly", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
		}

		post := FormatAlertPost(alert, "user456")

		require.NotNil(t, post)
		require.NotNil(t, post.Props)
		assert.Equal(t, true, post.Props["from_dataminr"])
		assert.Equal(t, "alert123", post.Props["alert_id"])
		assert.Equal(t, "Flash", post.Props["alert_type"])
		assert.Equal(t, "user456", post.Props["dataminr_user"])
	})

	t.Run("handles minimal alert with only required fields", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "2026-04-23T10:00:00Z",
			Headline:       "Minimal alert",
		}

		post := FormatAlertPost(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "Minimal alert")
		// Should not panic or error
	})
}

func TestGetAlertEmoji(t *testing.T) {
	tests := []struct {
		alertType string
		expected  string
	}{
		{"Flash", "🔴"},
		{"Urgent", "🟠"},
		{"Alert", "🟡"},
		{"Unknown", "🟡"},
		{"", "🟡"},
	}

	for _, tt := range tests {
		t.Run(tt.alertType, func(t *testing.T) {
			emoji := GetAlertEmoji(tt.alertType)
			assert.Equal(t, tt.expected, emoji)
		})
	}
}

// Ensure the Post type is correctly used
var _ *model.Post = (*model.Post)(nil)

// Tests for enhanced attachment-based formatting
func TestFormatAlertAttachment(t *testing.T) {
	t.Run("creates attachment with correct color for Flash alert", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000", // Unix millis
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major earthquake reported",
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		assert.Equal(t, "#FF0000", attachment.Color)
		assert.Contains(t, attachment.Pretext, "🔴")
		assert.Contains(t, attachment.Pretext, "FLASH")
		assert.Equal(t, "Major earthquake reported", attachment.Title)
	})

	t.Run("creates attachment with correct color for Urgent alert", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Urgent"},
			Headline:       "Severe weather warning",
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		assert.Equal(t, "#FF9900", attachment.Color)
		assert.Contains(t, attachment.Pretext, "🟠")
		assert.Contains(t, attachment.Pretext, "URGENT")
	})

	t.Run("creates attachment with correct color for Alert type", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Alert"},
			Headline:       "Traffic incident",
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		assert.Equal(t, "#FFFF00", attachment.Color)
		assert.Contains(t, attachment.Pretext, "🟡")
	})

	t.Run("includes formatted timestamp field", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000", // April 25, 2024
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Test alert",
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		// Find the Event Time field
		var foundTime bool
		for _, field := range attachment.Fields {
			if field.Title == "Event Time" {
				foundTime = true
				assert.NotEmpty(t, field.Value)
				break
			}
		}
		assert.True(t, foundTime, "Event Time field should be present")
	})

	t.Run("includes linked alerts count when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			LinkedAlerts: []dataminr.LinkedAlerts{
				{Count: 5, ParentAlertID: "parent123"},
			},
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		var foundLinked bool
		for _, field := range attachment.Fields {
			if field.Title == "Related Alerts" {
				foundLinked = true
				assert.Contains(t, field.Value, "5")
				break
			}
		}
		assert.True(t, foundLinked, "Related Alerts field should be present")
	})

	t.Run("includes matched lists when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			ListsMatched: []dataminr.ListsMatched{
				{ID: "1", Name: "Emergency Alerts"},
				{ID: "2", Name: "Critical Infrastructure"},
			},
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		var foundLists bool
		for _, field := range attachment.Fields {
			if field.Title == "Alert Lists" {
				foundLists = true
				assert.Contains(t, field.Value, "Emergency Alerts")
				assert.Contains(t, field.Value, "Critical Infrastructure")
				break
			}
		}
		assert.True(t, foundLists, "Alert Lists field should be present")
	})

	t.Run("includes keywords when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			AlertReferenceTerms: []dataminr.AlertReferenceTerms{
				{Text: "earthquake"},
				{Text: "emergency"},
			},
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		var foundKeywords bool
		for _, field := range attachment.Fields {
			if field.Title == "Keywords" {
				foundKeywords = true
				assert.Contains(t, field.Value, "earthquake")
				assert.Contains(t, field.Value, "emergency")
				break
			}
		}
		assert.True(t, foundKeywords, "Keywords field should be present")
	})

	t.Run("includes MGRS when present", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Name:        "San Francisco, CA",
				Coordinates: []float64{37.7749, -122.4194},
				MGRS:        "10S EG 12345 67890",
			},
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		var foundMGRS bool
		for _, field := range attachment.Fields {
			if field.Title == "MGRS" {
				foundMGRS = true
				assert.Equal(t, "10S EG 12345 67890", field.Value)
				break
			}
		}
		assert.True(t, foundMGRS, "MGRS field should be present")
	})

	t.Run("sets title link to Dataminr URL", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:          "alert123",
			AlertTimestamp:   "1714039200000",
			AlertType:        &dataminr.AlertType{Name: "Flash"},
			Headline:         "Major incident",
			DataminrAlertURL: "https://app.dataminr.com/alerts/123",
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		assert.Equal(t, "https://app.dataminr.com/alerts/123", attachment.TitleLink)
	})

	t.Run("includes footer with alert ID", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert-abc-123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major incident",
		}

		attachment := FormatAlertAttachment(alert)

		require.NotNil(t, attachment)
		assert.Contains(t, attachment.Footer, "alert-abc-123")
		assert.Contains(t, attachment.Footer, "Dataminr")
	})
}

func TestGetAlertColor(t *testing.T) {
	tests := []struct {
		alertType string
		expected  string
	}{
		{"Flash", "#FF0000"},
		{"Urgent", "#FF9900"},
		{"Alert", "#FFFF00"},
		{"Unknown", "#808080"},
		{"", "#808080"},
	}

	for _, tt := range tests {
		t.Run(tt.alertType, func(t *testing.T) {
			color := GetAlertColor(tt.alertType)
			assert.Equal(t, tt.expected, color)
		})
	}
}

// Tests for hashtag generation
func TestGenerateHashtags(t *testing.T) {
	t.Run("generates alert type hashtag", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Test alert",
		}

		hashtags := GenerateHashtags(alert)

		assert.Contains(t, hashtags, "#Flash")
	})

	t.Run("generates country hashtag from location", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Test alert",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Name: "San Francisco, CA, USA",
			},
		}

		hashtags := GenerateHashtags(alert)

		assert.Contains(t, hashtags, "#UnitedStates")
	})

	t.Run("generates topic hashtags", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Test alert",
			AlertTopics: []dataminr.AlertTopic{
				{ID: "1", Name: "Disasters and Weather - Structure Fires"},
			},
		}

		hashtags := GenerateHashtags(alert)

		assert.Contains(t, hashtags, "#Disasters")
		assert.Contains(t, hashtags, "#Weather")
		assert.Contains(t, hashtags, "#StructureFires")
	})

	t.Run("returns formatted hashtag string with emoji prefix", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Urgent"},
			Headline:       "Test alert",
		}

		hashtags := GenerateHashtags(alert)

		assert.True(t, len(hashtags) > 0)
		assert.Contains(t, hashtags, "🏷️")
	})

	t.Run("handles empty alert gracefully", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert123",
			Headline: "Test alert",
		}

		hashtags := GenerateHashtags(alert)

		// Should at least have default alert type
		assert.Contains(t, hashtags, "#Alert")
	})
}

// Tests for priority flag determination
func TestGetAlertPriority(t *testing.T) {
	t.Run("returns urgent priority for Flash alerts", func(t *testing.T) {
		priority := GetAlertPriority("Flash")

		require.NotNil(t, priority)
		assert.Equal(t, "urgent", priority.Priority)
	})

	t.Run("returns important priority for Urgent alerts", func(t *testing.T) {
		priority := GetAlertPriority("Urgent")

		require.NotNil(t, priority)
		assert.Equal(t, "important", priority.Priority)
	})

	t.Run("returns nil for Alert type", func(t *testing.T) {
		priority := GetAlertPriority("Alert")

		assert.Nil(t, priority)
	})

	t.Run("returns nil for unknown types", func(t *testing.T) {
		priority := GetAlertPriority("Unknown")

		assert.Nil(t, priority)
	})
}

// Tests for the enhanced post creation
func TestFormatAlertPostEnhanced(t *testing.T) {
	t.Run("creates post with attachment", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major earthquake reported",
		}

		post := FormatAlertPostEnhanced(alert, "user123")

		require.NotNil(t, post)
		// Check that attachments are set in Props
		attachments, ok := post.Props["attachments"]
		assert.True(t, ok, "Post should have attachments")
		assert.NotNil(t, attachments)
	})

	t.Run("includes hashtags in message text", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Major earthquake",
			AlertTopics: []dataminr.AlertTopic{
				{ID: "1", Name: "Natural Disasters"},
			},
		}

		post := FormatAlertPostEnhanced(alert, "user123")

		require.NotNil(t, post)
		assert.Contains(t, post.Message, "🏷️")
		assert.Contains(t, post.Message, "#Flash")
	})

	t.Run("sets priority for Flash alerts", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Critical alert",
		}

		post := FormatAlertPostEnhanced(alert, "user123")

		require.NotNil(t, post)
		priority, ok := post.Props["priority"]
		assert.True(t, ok, "Flash alert should have priority")
		priorityMap, ok := priority.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "urgent", priorityMap["priority"])
	})

	t.Run("sets priority for Urgent alerts", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Urgent"},
			Headline:       "Important alert",
		}

		post := FormatAlertPostEnhanced(alert, "user123")

		require.NotNil(t, post)
		priority, ok := post.Props["priority"]
		assert.True(t, ok, "Urgent alert should have priority")
		priorityMap, ok := priority.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "important", priorityMap["priority"])
	})

	t.Run("no priority for regular Alert type", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Alert"},
			Headline:       "Regular alert",
		}

		post := FormatAlertPostEnhanced(alert, "user123")

		require.NotNil(t, post)
		_, ok := post.Props["priority"]
		assert.False(t, ok, "Alert type should not have priority")
	})

	t.Run("includes all metadata in props", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert123",
			AlertTimestamp: "1714039200000",
			AlertType:      &dataminr.AlertType{Name: "Flash"},
			Headline:       "Test alert",
		}

		post := FormatAlertPostEnhanced(alert, "user456")

		require.NotNil(t, post)
		assert.Equal(t, true, post.Props["from_dataminr"])
		assert.Equal(t, "alert123", post.Props["alert_id"])
		assert.Equal(t, "Flash", post.Props["alert_type"])
		assert.Equal(t, "user456", post.Props["dataminr_user"])
	})
}

// Test country extraction helper
func TestExtractCountryFromLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected string
	}{
		{"USA full", "San Francisco, CA, USA", "UnitedStates"},
		{"US code", "New York, US", "UnitedStates"},
		{"UK code", "London, UK", "UnitedKingdom"},
		{"Full country name", "Paris, France", "France"},
		{"Germany", "Berlin, Germany", "Germany"},
		{"Japan", "Tokyo, Japan", "Japan"},
		{"No country", "Some Location", ""},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			country := ExtractCountryFromLocation(tt.location)
			assert.Equal(t, tt.expected, country)
		})
	}
}
