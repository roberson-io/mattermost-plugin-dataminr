package dataminr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscription_JSONMarshalUnmarshal(t *testing.T) {
	original := &Subscription{
		ChannelID:    "channel123",
		CreatorID:    "creator456",
		DataminrUser: "dataminr789",
		Enabled:      true,
		Features:     "all",
		CreatedAt:    1234567890,
		PollInterval: 60,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "channel123")
	assert.Contains(t, string(jsonData), "creator456")

	// Unmarshal back
	var restored Subscription
	err = json.Unmarshal(jsonData, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ChannelID, restored.ChannelID)
	assert.Equal(t, original.CreatorID, restored.CreatorID)
	assert.Equal(t, original.DataminrUser, restored.DataminrUser)
	assert.Equal(t, original.Enabled, restored.Enabled)
	assert.Equal(t, original.Features, restored.Features)
	assert.Equal(t, original.CreatedAt, restored.CreatedAt)
	assert.Equal(t, original.PollInterval, restored.PollInterval)
}

func TestSubscription_Validate(t *testing.T) {
	tests := []struct {
		name        string
		sub         *Subscription
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid subscription",
			sub: &Subscription{
				ChannelID:    "channel123",
				CreatorID:    "creator456",
				DataminrUser: "dataminr789",
				Enabled:      true,
				Features:     "all",
			},
			expectError: false,
		},
		{
			name: "missing channel ID",
			sub: &Subscription{
				ChannelID:    "",
				CreatorID:    "creator456",
				DataminrUser: "dataminr789",
			},
			expectError: true,
			errorMsg:    "channel_id",
		},
		{
			name: "missing creator ID",
			sub: &Subscription{
				ChannelID:    "channel123",
				CreatorID:    "",
				DataminrUser: "dataminr789",
			},
			expectError: true,
			errorMsg:    "creator_id",
		},
		{
			name: "missing dataminr user",
			sub: &Subscription{
				ChannelID:    "channel123",
				CreatorID:    "creator456",
				DataminrUser: "",
			},
			expectError: true,
			errorMsg:    "dataminr_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubscriptions_AddSubscription(t *testing.T) {
	subs := NewSubscriptions()

	sub := &Subscription{
		ChannelID:    "channel123",
		CreatorID:    "creator456",
		DataminrUser: "user789",
		Enabled:      true,
		Features:     "all",
		PollInterval: 60,
	}

	// Add subscription
	subs.Add(sub)

	// Verify it was added
	userSubs := subs.GetByDataminrUser("user789")
	require.Len(t, userSubs, 1)
	assert.Equal(t, "channel123", userSubs[0].ChannelID)
}

func TestSubscriptions_GetByChannel(t *testing.T) {
	subs := NewSubscriptions()

	// Add subscriptions from different users to the same channel
	subs.Add(&Subscription{
		ChannelID:    "channel123",
		CreatorID:    "creator1",
		DataminrUser: "user1",
		Enabled:      true,
	})
	subs.Add(&Subscription{
		ChannelID:    "channel123",
		CreatorID:    "creator2",
		DataminrUser: "user2",
		Enabled:      true,
	})
	subs.Add(&Subscription{
		ChannelID:    "channel456",
		CreatorID:    "creator3",
		DataminrUser: "user3",
		Enabled:      true,
	})

	// Get subscriptions for channel123
	channelSubs := subs.GetByChannel("channel123")
	assert.Len(t, channelSubs, 2)

	// Get subscriptions for channel456
	channelSubs = subs.GetByChannel("channel456")
	assert.Len(t, channelSubs, 1)

	// Get subscriptions for non-existent channel
	channelSubs = subs.GetByChannel("nonexistent")
	assert.Len(t, channelSubs, 0)
}

func TestSubscriptions_Remove(t *testing.T) {
	subs := NewSubscriptions()

	// Add subscriptions
	subs.Add(&Subscription{
		ChannelID:    "channel123",
		CreatorID:    "creator1",
		DataminrUser: "user1",
		Enabled:      true,
	})
	subs.Add(&Subscription{
		ChannelID:    "channel456",
		CreatorID:    "creator1",
		DataminrUser: "user1",
		Enabled:      true,
	})

	// Verify both exist
	userSubs := subs.GetByDataminrUser("user1")
	require.Len(t, userSubs, 2)

	// Remove one subscription
	removed := subs.Remove("channel123", "user1")
	assert.True(t, removed)

	// Verify only one remains
	userSubs = subs.GetByDataminrUser("user1")
	assert.Len(t, userSubs, 1)
	assert.Equal(t, "channel456", userSubs[0].ChannelID)

	// Try to remove non-existent subscription
	removed = subs.Remove("channel123", "user1")
	assert.False(t, removed)
}

func TestSubscriptions_JSONMarshalUnmarshal(t *testing.T) {
	original := NewSubscriptions()
	original.Add(&Subscription{
		ChannelID:    "channel123",
		CreatorID:    "creator1",
		DataminrUser: "user1",
		Enabled:      true,
		Features:     "all",
	})
	original.Add(&Subscription{
		ChannelID:    "channel456",
		CreatorID:    "creator2",
		DataminrUser: "user2",
		Enabled:      false,
		Features:     "flash",
	})

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var restored Subscriptions
	err = json.Unmarshal(jsonData, &restored)
	require.NoError(t, err)

	// Verify data
	user1Subs := restored.GetByDataminrUser("user1")
	require.Len(t, user1Subs, 1)
	assert.Equal(t, "channel123", user1Subs[0].ChannelID)

	user2Subs := restored.GetByDataminrUser("user2")
	require.Len(t, user2Subs, 1)
	assert.Equal(t, "channel456", user2Subs[0].ChannelID)
}

func TestNewSubscription(t *testing.T) {
	sub := NewSubscription("channel123", "creator456", "dataminr789")

	assert.Equal(t, "channel123", sub.ChannelID)
	assert.Equal(t, "creator456", sub.CreatorID)
	assert.Equal(t, "dataminr789", sub.DataminrUser)
	assert.True(t, sub.Enabled)
	assert.Equal(t, "all", sub.Features)
	assert.Equal(t, 0, sub.PollInterval)
	assert.NotZero(t, sub.CreatedAt)
}
