package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPollInterval(t *testing.T) {
	t.Run("returns default interval when not configured", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{}

		interval := plugin.getPollInterval()

		// Default should be 2 minutes (120 seconds)
		assert.Equal(t, 2*time.Minute, interval)
	})

	t.Run("returns configured interval", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrDefaultPollInterval: 60, // 1 minute
		}

		interval := plugin.getPollInterval()

		assert.Equal(t, 60*time.Second, interval)
	})

	t.Run("enforces minimum interval", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrDefaultPollInterval: 10, // Too short
			DataminrMinPollInterval:     30, // Minimum 30 seconds
		}

		interval := plugin.getPollInterval()

		// Should return minimum, not configured
		assert.Equal(t, 30*time.Second, interval)
	})

	t.Run("uses default minimum when not configured", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrDefaultPollInterval: 10, // Too short
			DataminrMinPollInterval:     0,  // Not set, should use default 30
		}

		interval := plugin.getPollInterval()

		// Should return default minimum (30 seconds)
		assert.Equal(t, 30*time.Second, interval)
	})
}

func TestShouldStartDataminrJob(t *testing.T) {
	t.Run("returns true when plugin is enabled", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrEnabled: true,
		}

		assert.True(t, plugin.shouldStartDataminrJob())
	})

	t.Run("returns false when plugin is disabled", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrEnabled: false,
		}

		assert.False(t, plugin.shouldStartDataminrJob())
	})

	t.Run("returns false when configuration is nil", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = nil

		assert.False(t, plugin.shouldStartDataminrJob())
	})
}

func TestValidatePollInterval(t *testing.T) {
	t.Run("allows zero interval (manual only)", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrMinPollInterval: 30,
		}

		err := plugin.validatePollInterval(0)

		require.NoError(t, err)
	})

	t.Run("rejects interval below minimum", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrMinPollInterval: 30,
		}

		err := plugin.validatePollInterval(15)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "30 seconds")
	})

	t.Run("allows interval at minimum", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrMinPollInterval: 30,
		}

		err := plugin.validatePollInterval(30)

		require.NoError(t, err)
	})

	t.Run("allows interval above minimum", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrMinPollInterval: 30,
		}

		err := plugin.validatePollInterval(120)

		require.NoError(t, err)
	})

	t.Run("uses default minimum when not configured", func(t *testing.T) {
		plugin := &Plugin{}
		plugin.configuration = &configuration{
			DataminrMinPollInterval: 0, // Not set
		}

		// 15 is below default minimum of 30
		err := plugin.validatePollInterval(15)

		require.Error(t, err)
	})
}
