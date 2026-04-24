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
