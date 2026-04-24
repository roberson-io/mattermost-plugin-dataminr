package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCanCreateSubscription_AnyoneMode(t *testing.T) {
	t.Run("allows channel member", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "anyone",
		})

		channelID := "channel123"
		userID := "user123"

		// User is a channel member
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId: channelID,
			UserId:    userID,
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("denies non-member", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "anyone",
		})

		channelID := "channel123"
		userID := "user123"

		// User is NOT a channel member
		api.On("GetChannelMember", channelID, userID).Return(nil, &model.AppError{
			Message: "channel member not found",
		})

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "member")
		api.AssertExpectations(t)
	})
}

func TestCanCreateSubscription_ChannelAdminMode(t *testing.T) {
	t.Run("allows channel admin", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "channel_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// User is a channel admin - no need to check team admin
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId:   channelID,
			UserId:      userID,
			SchemeAdmin: true,
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("allows team admin", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "channel_admin",
		})

		channelID := "channel123"
		userID := "user123"
		teamID := "team123"

		// User is NOT a channel admin, but IS a team admin
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId:   channelID,
			UserId:      userID,
			SchemeAdmin: false,
		}, nil)
		api.On("GetChannel", channelID).Return(&model.Channel{
			Id:     channelID,
			TeamId: teamID,
		}, nil)
		api.On("GetTeamMember", teamID, userID).Return(&model.TeamMember{
			TeamId:      teamID,
			UserId:      userID,
			SchemeAdmin: true,
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("denies regular channel member", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "channel_admin",
		})

		channelID := "channel123"
		userID := "user123"
		teamID := "team123"

		// User is NOT a channel admin and NOT a team admin
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId:   channelID,
			UserId:      userID,
			SchemeAdmin: false,
		}, nil)
		api.On("GetChannel", channelID).Return(&model.Channel{
			Id:     channelID,
			TeamId: teamID,
		}, nil)
		api.On("GetTeamMember", teamID, userID).Return(&model.TeamMember{
			TeamId:      teamID,
			UserId:      userID,
			SchemeAdmin: false,
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel admin")
		api.AssertExpectations(t)
	})

	t.Run("denies non-member", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "channel_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// User is NOT a channel member at all
		api.On("GetChannelMember", channelID, userID).Return(nil, &model.AppError{
			Message: "channel member not found",
		})

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		api.AssertExpectations(t)
	})
}

func TestCanCreateSubscription_SystemAdminMode(t *testing.T) {
	t.Run("allows system admin", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "system_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// User is a system admin
		api.On("GetUser", userID).Return(&model.User{
			Id:    userID,
			Roles: "system_admin system_user",
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("denies non-system-admin", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "system_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// User is NOT a system admin
		api.On("GetUser", userID).Return(&model.User{
			Id:    userID,
			Roles: "system_user",
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "system admin")
		api.AssertExpectations(t)
	})

	t.Run("denies channel admin who is not system admin", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "system_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// User is a channel admin but not system admin
		api.On("GetUser", userID).Return(&model.User{
			Id:    userID,
			Roles: "system_user", // No system_admin role
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "system admin")
		api.AssertExpectations(t)
	})
}

func TestCanCreateSubscription_DefaultBehavior(t *testing.T) {
	t.Run("empty permission setting defaults to channel_admin behavior", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "", // Empty/not set
		})

		channelID := "channel123"
		userID := "user123"

		// User is a channel admin - should be allowed (no GetChannel call needed)
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId:   channelID,
			UserId:      userID,
			SchemeAdmin: true,
		}, nil)

		err := p.canCreateSubscription(userID, channelID)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})
}

func TestCanCreateSubscription_ErrorHandling(t *testing.T) {
	t.Run("handles API error when getting user", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "system_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// API returns error
		api.On("GetUser", userID).Return(nil, &model.AppError{
			Message: "database error",
		})

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		api.AssertExpectations(t)
	})

	t.Run("handles API error when getting channel", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{
			DataminrSubscriptionPermission: "channel_admin",
		})

		channelID := "channel123"
		userID := "user123"

		// ChannelMember exists but Channel lookup fails
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId:   channelID,
			UserId:      userID,
			SchemeAdmin: false,
		}, nil)
		api.On("GetChannel", channelID).Return(nil, &model.AppError{
			Message: "channel not found",
		})

		err := p.canCreateSubscription(userID, channelID)

		require.Error(t, err)
		api.AssertExpectations(t)
	})
}

// Ignore unused mock warning - we're just setting up expectations
var _ = mock.Anything
