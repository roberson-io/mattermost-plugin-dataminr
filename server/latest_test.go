package main

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatAlertFull(t *testing.T) {
	t.Run("formats basic alert with headline and type", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Breaking: Major Event",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "### 1.")
		assert.Contains(t, result, "[Flash]")
		assert.Contains(t, result, "Breaking: Major Event")
		assert.Contains(t, result, "🔴") // Flash emoji (red circle)
	})

	t.Run("formats alert with default type when AlertType is nil", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test Headline",
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "[Alert]")
		assert.Contains(t, result, "Test Headline")
	})

	t.Run("formats alert with timestamp", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert-123",
			Headline:       "Test",
			AlertTimestamp: "1700000000000", // Unix timestamp in milliseconds
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Time:**")
		assert.Contains(t, result, "2023") // Year from the timestamp
	})

	t.Run("handles invalid timestamp gracefully", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert-123",
			Headline:       "Test",
			AlertTimestamp: "invalid",
		}

		result := formatAlertFull(1, alert)

		// Should not contain Time field if parsing fails
		assert.NotContains(t, result, "**Time:**")
	})

	t.Run("formats alert with subheadline", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			SubHeadline: &dataminr.SubHeadline{
				Title:   "Related",
				Content: []string{"First item", "Second item"},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Related:**")
		assert.Contains(t, result, "First item")
		assert.Contains(t, result, "Second item")
	})

	t.Run("formats alert with location name", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Name: "New York, NY",
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**📍 Location:**")
		assert.Contains(t, result, "New York, NY")
	})

	t.Run("formats alert with coordinates", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Coordinates: []float64{40.7128, -74.006},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Coordinates:**")
		assert.Contains(t, result, "40.71280")
		assert.Contains(t, result, "-74.00600")
	})

	t.Run("formats alert with coordinates and probability radius", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Coordinates:       []float64{40.7128, -74.006},
				ProbabilityRadius: 5.5,
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "(±5.50 mi)")
	})

	t.Run("formats alert with MGRS", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				MGRS: "18TWL8513619717",
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**MGRS:**")
		assert.Contains(t, result, "18TWL8513619717")
	})

	t.Run("formats alert with topics", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			AlertTopics: []dataminr.AlertTopic{
				{Name: "Weather"},
				{Name: "Emergency"},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Topics:**")
		assert.Contains(t, result, "Weather")
		assert.Contains(t, result, "Emergency")
	})

	t.Run("formats alert with reference terms", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			AlertReferenceTerms: []dataminr.AlertReferenceTerms{
				{Text: "earthquake"},
				{Text: "damage"},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Keywords:**")
		assert.Contains(t, result, "earthquake")
		assert.Contains(t, result, "damage")
	})

	t.Run("formats alert with matched lists", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			ListsMatched: []dataminr.ListsMatched{
				{Name: "Critical Events"},
				{Name: "High Priority"},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Matched Lists:**")
		assert.Contains(t, result, "Critical Events")
		assert.Contains(t, result, "High Priority")
	})

	t.Run("formats alert with linked alerts count", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			LinkedAlerts: []dataminr.LinkedAlerts{
				{Count: 5},
				{Count: 3},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**Linked Alerts:**")
		assert.Contains(t, result, "8 related alerts")
	})

	t.Run("skips linked alerts when count is zero", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			LinkedAlerts: []dataminr.LinkedAlerts{
				{Count: 0},
			},
		}

		result := formatAlertFull(1, alert)

		assert.NotContains(t, result, "**Linked Alerts:**")
	})

	t.Run("formats alert with AI summary (LiveBrief)", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			LiveBrief: []dataminr.LiveBrief{
				{Version: "old", Summary: "Old summary"},
				{Version: "current", Summary: "This is the AI-generated summary of the event."},
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "**🤖 AI Summary:**")
		assert.Contains(t, result, "This is the AI-generated summary of the event.")
		assert.NotContains(t, result, "Old summary")
	})

	t.Run("formats alert with Dataminr URL", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:          "alert-123",
			Headline:         "Test",
			DataminrAlertURL: "https://app.dataminr.com/alert/123",
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "[View in Dataminr](https://app.dataminr.com/alert/123)")
	})

	t.Run("formats alert with source post link", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
			PublicPost: &dataminr.PublicPost{
				Href: "https://twitter.com/example/status/123",
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "[Source Post](https://twitter.com/example/status/123)")
	})

	t.Run("formats alert with both URLs separated by pipe", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:          "alert-123",
			Headline:         "Test",
			DataminrAlertURL: "https://app.dataminr.com/alert/123",
			PublicPost: &dataminr.PublicPost{
				Href: "https://twitter.com/example/status/123",
			},
		}

		result := formatAlertFull(1, alert)

		assert.Contains(t, result, "[View in Dataminr]")
		assert.Contains(t, result, " | ")
		assert.Contains(t, result, "[Source Post]")
	})

	t.Run("formats complete alert with all fields", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:        "alert-123",
			Headline:       "Major Earthquake Strikes",
			AlertTimestamp: "1700000000000",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
			SubHeadline: &dataminr.SubHeadline{
				Title:   "Updates",
				Content: []string{"Aftershocks expected"},
			},
			EstimatedEventLocation: &dataminr.EstimatedEventLocation{
				Name:              "San Francisco, CA",
				Coordinates:       []float64{37.7749, -122.4194},
				ProbabilityRadius: 10.0,
				MGRS:              "10SEG1234567890",
			},
			AlertTopics: []dataminr.AlertTopic{
				{Name: "Natural Disasters"},
			},
			AlertReferenceTerms: []dataminr.AlertReferenceTerms{
				{Text: "earthquake"},
			},
			ListsMatched: []dataminr.ListsMatched{
				{Name: "Critical Events"},
			},
			LinkedAlerts: []dataminr.LinkedAlerts{
				{Count: 15},
			},
			LiveBrief: []dataminr.LiveBrief{
				{Version: "current", Summary: "A major earthquake has struck the Bay Area."},
			},
			DataminrAlertURL: "https://app.dataminr.com/alert/123",
			PublicPost: &dataminr.PublicPost{
				Href: "https://twitter.com/example",
			},
		}

		result := formatAlertFull(1, alert)

		// Verify all sections are present
		assert.Contains(t, result, "### 1.")
		assert.Contains(t, result, "🔴") // Flash emoji is red circle
		assert.Contains(t, result, "[Flash]")
		assert.Contains(t, result, "Major Earthquake Strikes")
		assert.Contains(t, result, "**Time:**")
		assert.Contains(t, result, "**Updates:**")
		assert.Contains(t, result, "**📍 Location:**")
		assert.Contains(t, result, "San Francisco, CA")
		assert.Contains(t, result, "**Coordinates:**")
		assert.Contains(t, result, "**MGRS:**")
		assert.Contains(t, result, "**Topics:**")
		assert.Contains(t, result, "**Keywords:**")
		assert.Contains(t, result, "**Matched Lists:**")
		assert.Contains(t, result, "**Linked Alerts:**")
		assert.Contains(t, result, "**🤖 AI Summary:**")
		assert.Contains(t, result, "[View in Dataminr]")
		assert.Contains(t, result, "[Source Post]")
	})

	t.Run("formats different index numbers correctly", func(t *testing.T) {
		alert := &dataminr.Alert{
			AlertID:  "alert-123",
			Headline: "Test",
		}

		result1 := formatAlertFull(1, alert)
		result5 := formatAlertFull(5, alert)
		result10 := formatAlertFull(10, alert)

		assert.True(t, strings.HasPrefix(result1, "### 1."))
		assert.True(t, strings.HasPrefix(result5, "### 5."))
		assert.True(t, strings.HasPrefix(result10, "### 10."))
	})
}

func TestHandleLatest(t *testing.T) {
	t.Run("returns error when user not connected", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// User not connected
		api.On("KVGet", "user123"+userInfoKeyPrefix).Return(nil, nil)

		response, err := plugin.HandleLatest("user123", 5)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})
}
