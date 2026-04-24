package main

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeduplicateAlerts(t *testing.T) {
	t.Run("allows new alerts through", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alerts := []dataminr.Alert{
			{AlertID: "alert1", Headline: "First alert"},
			{AlertID: "alert2", Headline: "Second alert"},
		}

		// Both alerts are new (not in KV store)
		api.On("KVGet", "dataminr_alert_alert1").Return(nil, nil)
		api.On("KVGet", "dataminr_alert_alert2").Return(nil, nil)

		// Both should be stored with TTL
		api.On("KVSetWithExpiry", "dataminr_alert_alert1", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)
		api.On("KVSetWithExpiry", "dataminr_alert_alert2", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		result := plugin.deduplicateAlerts(alerts)

		require.Len(t, result, 2)
		assert.Equal(t, "alert1", result[0].AlertID)
		assert.Equal(t, "alert2", result[1].AlertID)
	})

	t.Run("filters out already-seen alerts", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alerts := []dataminr.Alert{
			{AlertID: "alert1", Headline: "First alert"},
			{AlertID: "alert2", Headline: "Second alert"},
			{AlertID: "alert3", Headline: "Third alert"},
		}

		// alert1 is new
		api.On("KVGet", "dataminr_alert_alert1").Return(nil, nil)
		api.On("KVSetWithExpiry", "dataminr_alert_alert1", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		// alert2 already exists (seen before)
		api.On("KVGet", "dataminr_alert_alert2").Return([]byte("seen"), nil)

		// alert3 is new
		api.On("KVGet", "dataminr_alert_alert3").Return(nil, nil)
		api.On("KVSetWithExpiry", "dataminr_alert_alert3", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		result := plugin.deduplicateAlerts(alerts)

		require.Len(t, result, 2)
		assert.Equal(t, "alert1", result[0].AlertID)
		assert.Equal(t, "alert3", result[1].AlertID)
	})

	t.Run("returns empty slice when all alerts are duplicates", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alerts := []dataminr.Alert{
			{AlertID: "alert1", Headline: "First alert"},
			{AlertID: "alert2", Headline: "Second alert"},
		}

		// Both alerts already exist
		api.On("KVGet", "dataminr_alert_alert1").Return([]byte("seen"), nil)
		api.On("KVGet", "dataminr_alert_alert2").Return([]byte("seen"), nil)

		result := plugin.deduplicateAlerts(alerts)

		require.Len(t, result, 0)
	})

	t.Run("handles empty input", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alerts := []dataminr.Alert{}

		result := plugin.deduplicateAlerts(alerts)

		require.Len(t, result, 0)
	})

	t.Run("continues processing on KV store error", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alerts := []dataminr.Alert{
			{AlertID: "alert1", Headline: "First alert"},
			{AlertID: "alert2", Headline: "Second alert"},
		}

		// alert1 KV get fails - should be treated as new (safe default)
		api.On("KVGet", "dataminr_alert_alert1").Return(nil, &model.AppError{Message: "db error"})
		api.On("KVSetWithExpiry", "dataminr_alert_alert1", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		// alert2 is new
		api.On("KVGet", "dataminr_alert_alert2").Return(nil, nil)
		api.On("KVSetWithExpiry", "dataminr_alert_alert2", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		result := plugin.deduplicateAlerts(alerts)

		// Both should be included since we default to allowing on error
		require.Len(t, result, 2)
	})
}

func TestMarkAlertAsSeen(t *testing.T) {
	t.Run("stores alert ID with 1 hour TTL", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		alertID := "alert123"

		api.On("KVSetWithExpiry", "dataminr_alert_alert123", mock.AnythingOfType("[]uint8"), int64(3600)).Return(nil)

		err := plugin.markAlertAsSeen(alertID)

		require.NoError(t, err)
	})
}

func TestIsAlertSeen(t *testing.T) {
	t.Run("returns true when alert exists in KV store", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVGet", "dataminr_alert_alert123").Return([]byte("seen"), nil)

		seen, err := plugin.isAlertSeen("alert123")

		require.NoError(t, err)
		assert.True(t, seen)
	})

	t.Run("returns false when alert does not exist", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVGet", "dataminr_alert_alert123").Return(nil, nil)

		seen, err := plugin.isAlertSeen("alert123")

		require.NoError(t, err)
		assert.False(t, seen)
	})
}

// Ensure time package is used (for TTL documentation)
var _ = time.Hour
