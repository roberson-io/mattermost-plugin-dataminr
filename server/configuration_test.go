package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationClone(t *testing.T) {
	t.Run("creates independent copy", func(t *testing.T) {
		original := &configuration{
			DataminrEnabled:                true,
			DataminrEncryptionKey:          "test-key-123",
			DataminrSubscriptionPermission: "all",
			DataminrDefaultPollInterval:    60,
			DataminrMinPollInterval:        30,
		}

		clone := original.Clone()

		// Verify values match
		assert.Equal(t, original.DataminrEnabled, clone.DataminrEnabled)
		assert.Equal(t, original.DataminrEncryptionKey, clone.DataminrEncryptionKey)
		assert.Equal(t, original.DataminrSubscriptionPermission, clone.DataminrSubscriptionPermission)
		assert.Equal(t, original.DataminrDefaultPollInterval, clone.DataminrDefaultPollInterval)
		assert.Equal(t, original.DataminrMinPollInterval, clone.DataminrMinPollInterval)

		// Verify they are different memory addresses
		assert.NotSame(t, original, clone)
	})

	t.Run("modifying clone does not affect original", func(t *testing.T) {
		original := &configuration{
			DataminrEnabled:             true,
			DataminrEncryptionKey:       "original-key",
			DataminrDefaultPollInterval: 60,
		}

		clone := original.Clone()
		clone.DataminrEnabled = false
		clone.DataminrEncryptionKey = "modified-key"
		clone.DataminrDefaultPollInterval = 120

		// Original should remain unchanged
		assert.True(t, original.DataminrEnabled)
		assert.Equal(t, "original-key", original.DataminrEncryptionKey)
		assert.Equal(t, 60, original.DataminrDefaultPollInterval)
	})

	t.Run("clones empty configuration", func(t *testing.T) {
		original := &configuration{}
		clone := original.Clone()

		assert.False(t, clone.DataminrEnabled)
		assert.Empty(t, clone.DataminrEncryptionKey)
		assert.Empty(t, clone.DataminrSubscriptionPermission)
		assert.Zero(t, clone.DataminrDefaultPollInterval)
		assert.Zero(t, clone.DataminrMinPollInterval)
		assert.NotSame(t, original, clone)
	})
}

func TestGetConfiguration(t *testing.T) {
	t.Run("returns empty configuration when nil", func(t *testing.T) {
		plugin := &Plugin{
			configurationLock: sync.RWMutex{},
			configuration:     nil,
		}

		config := plugin.getConfiguration()

		require.NotNil(t, config)
		assert.False(t, config.DataminrEnabled)
		assert.Empty(t, config.DataminrEncryptionKey)
	})

	t.Run("returns stored configuration", func(t *testing.T) {
		plugin := &Plugin{
			configurationLock: sync.RWMutex{},
			configuration: &configuration{
				DataminrEnabled:             true,
				DataminrEncryptionKey:       "test-key",
				DataminrDefaultPollInterval: 60,
			},
		}

		config := plugin.getConfiguration()

		require.NotNil(t, config)
		assert.True(t, config.DataminrEnabled)
		assert.Equal(t, "test-key", config.DataminrEncryptionKey)
		assert.Equal(t, 60, config.DataminrDefaultPollInterval)
	})
}

func TestSetConfiguration(t *testing.T) {
	t.Run("stores new configuration", func(t *testing.T) {
		plugin := &Plugin{
			configurationLock: sync.RWMutex{},
			configuration:     nil,
		}

		newConfig := &configuration{
			DataminrEnabled:       true,
			DataminrEncryptionKey: "new-key",
		}

		plugin.setConfiguration(newConfig)

		assert.Equal(t, newConfig, plugin.configuration)
	})

	t.Run("replaces existing configuration", func(t *testing.T) {
		oldConfig := &configuration{
			DataminrEnabled:       false,
			DataminrEncryptionKey: "old-key",
		}

		plugin := &Plugin{
			configurationLock: sync.RWMutex{},
			configuration:     oldConfig,
		}

		newConfig := &configuration{
			DataminrEnabled:       true,
			DataminrEncryptionKey: "new-key",
		}

		plugin.setConfiguration(newConfig)

		assert.Equal(t, newConfig, plugin.configuration)
		assert.True(t, plugin.configuration.DataminrEnabled)
		assert.Equal(t, "new-key", plugin.configuration.DataminrEncryptionKey)
	})

	t.Run("accepts nil configuration", func(t *testing.T) {
		plugin := &Plugin{
			configurationLock: sync.RWMutex{},
			configuration: &configuration{
				DataminrEnabled: true,
			},
		}

		plugin.setConfiguration(nil)

		assert.Nil(t, plugin.configuration)
	})

	t.Run("panics when setting same configuration pointer", func(t *testing.T) {
		config := &configuration{
			DataminrEnabled:       true,
			DataminrEncryptionKey: "test",
		}

		plugin := &Plugin{
			configurationLock: sync.RWMutex{},
			configuration:     config,
		}

		// Should panic when setting the same pointer
		assert.Panics(t, func() {
			plugin.setConfiguration(config)
		})
	})
}

func TestGenerateEncryptionKey(t *testing.T) {
	t.Run("generates valid base64 key", func(t *testing.T) {
		key, err := generateEncryptionKey()

		require.NoError(t, err)
		require.NotEmpty(t, key)
		// Base64 encoded 32 bytes should be 44 characters
		assert.Len(t, key, 44)
	})

	t.Run("generates unique keys on each call", func(t *testing.T) {
		key1, err1 := generateEncryptionKey()
		key2, err2 := generateEncryptionKey()

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, key1, key2)
	})
}
