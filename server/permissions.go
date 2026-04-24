package main

import (
	"strings"

	"github.com/pkg/errors"
)

// canCreateSubscription checks if a user has permission to create a subscription in a channel
// Returns nil if allowed, error if not allowed
func (p *Plugin) canCreateSubscription(userID, channelID string) error {
	config := p.getConfiguration()
	permission := config.DataminrSubscriptionPermission

	// Default to channel_admin if not set
	if permission == "" {
		permission = "channel_admin"
	}

	switch permission {
	case "anyone":
		return p.checkChannelMembership(userID, channelID)

	case "channel_admin":
		return p.checkChannelAdmin(userID, channelID)

	case "system_admin":
		return p.checkSystemAdmin(userID)

	default:
		// Unknown permission setting, default to channel_admin
		return p.checkChannelAdmin(userID, channelID)
	}
}

// checkChannelMembership verifies the user is a member of the channel
func (p *Plugin) checkChannelMembership(userID, channelID string) error {
	_, appErr := p.API.GetChannelMember(channelID, userID)
	if appErr != nil {
		return errors.Wrap(appErr, "user is not a member of this channel")
	}
	return nil
}

// checkChannelAdmin verifies the user is a channel admin, team admin, or system admin
func (p *Plugin) checkChannelAdmin(userID, channelID string) error {
	// First check if user is a channel member
	member, appErr := p.API.GetChannelMember(channelID, userID)
	if appErr != nil {
		return errors.Wrap(appErr, "user is not a member of this channel")
	}

	// Check if user is channel admin
	if member.SchemeAdmin {
		return nil
	}

	// Get channel to find team ID
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get channel")
	}

	// Check if user is team admin
	teamMember, appErr := p.API.GetTeamMember(channel.TeamId, userID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get team member")
	}

	if teamMember.SchemeAdmin {
		return nil
	}

	return errors.New("only channel admins can create subscriptions")
}

// checkSystemAdmin verifies the user is a system admin
func (p *Plugin) checkSystemAdmin(userID string) error {
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get user")
	}

	if strings.Contains(user.Roles, "system_admin") {
		return nil
	}

	return errors.New("only system admins can create subscriptions")
}
