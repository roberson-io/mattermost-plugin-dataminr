package main

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

func (p *Plugin) runJob() {
	// Include job logic here
	p.API.LogInfo("Job is currently running")
}

// runDataminrJob polls alerts for all connected users based on their configured intervals.
// This job runs on a fixed tick (e.g., every 30 seconds) and checks each user/subscription
// to determine if enough time has passed since their last poll.
func (p *Plugin) runDataminrJob() error {
	if !p.shouldStartDataminrJob() {
		return nil
	}

	// Get all connected users
	users, err := p.getAllConnectedDataminrUsers()
	if err != nil {
		return errors.Wrap(err, "failed to get connected users")
	}

	// Get all subscriptions for interval-based channel polling
	subs, err := p.getSubscriptions()
	if err != nil {
		return errors.Wrap(err, "failed to get subscriptions")
	}

	now := time.Now().Unix()

	// Poll each user's DM if due
	for _, userID := range users {
		if err := p.pollUserDMIfDue(userID, now); err != nil {
			// Log error but continue with other users
			p.API.LogWarn("Failed to poll DM alerts for user", "user_id", userID, "error", err.Error())
		}
	}

	// Poll each subscription if due
	subsUpdated := false
	for _, userSubs := range subs.Users {
		for _, sub := range userSubs {
			if p.isSubscriptionDueForPoll(sub, now) {
				if err := p.pollSubscriptionAlerts(sub); err != nil {
					p.API.LogWarn("Failed to poll subscription alerts",
						"user_id", sub.DataminrUser,
						"channel_id", sub.ChannelID,
						"error", err.Error())
				} else {
					// Update last poll time for this subscription
					sub.LastPollAt = now
					subsUpdated = true
				}
			}
		}
	}

	// Save updated subscriptions if any were polled
	if subsUpdated {
		if err := p.storeSubscriptions(subs); err != nil {
			p.API.LogWarn("Failed to store updated subscriptions", "error", err.Error())
		}
	}

	return nil
}

// isUserDMDueForPoll checks if a user's DM notifications are due for polling
func (p *Plugin) isUserDMDueForPoll(userInfo *dataminr.UserInfo, now int64) bool {
	// Check if DM notifications are enabled
	if userInfo.Settings == nil || !userInfo.Settings.DMNotifications {
		return false
	}

	// Check if interval is set (0 = manual only)
	interval := userInfo.Settings.DMPollInterval
	if interval <= 0 {
		return false
	}

	// Check if enough time has passed since last poll
	return now-userInfo.LastPollAt >= int64(interval)
}

// isSubscriptionDueForPoll checks if a subscription is due for polling
func (p *Plugin) isSubscriptionDueForPoll(sub *dataminr.Subscription, now int64) bool {
	// Check if subscription is enabled
	if !sub.Enabled {
		return false
	}

	// Check if interval is set (0 = manual only)
	if sub.PollInterval <= 0 {
		return false
	}

	// Check if enough time has passed since last poll
	return now-sub.LastPollAt >= int64(sub.PollInterval)
}

// pollUserDMIfDue checks if a user's DM is due for polling and polls if so
func (p *Plugin) pollUserDMIfDue(userID string, now int64) error {
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user info")
	}

	if userInfo == nil {
		return nil // User not connected
	}

	if !p.isUserDMDueForPoll(userInfo, now) {
		return nil // Not due for polling
	}

	// Poll alerts for this user's DM
	if err := p.pollUserAlerts(userID); err != nil {
		return err
	}

	// Update last poll time
	userInfo.LastPollAt = now
	if err := p.storeUserInfo(userInfo); err != nil {
		return errors.Wrap(err, "failed to update last poll time")
	}

	return nil
}

// pollSubscriptionAlerts polls alerts for a specific subscription
func (p *Plugin) pollSubscriptionAlerts(sub *dataminr.Subscription) error {
	// Get user info to access credentials
	userInfo, err := p.getUserInfo(sub.DataminrUser)
	if err != nil {
		return errors.Wrap(err, "failed to get user info")
	}

	if userInfo == nil {
		return errors.New("user not connected to Dataminr")
	}

	// In a real implementation, we would:
	// 1. Get credentials and create Dataminr client
	// 2. Fetch alerts using the user's cursor
	// 3. Process alerts specifically for this channel
	// 4. Update cursor

	// For now, this is a placeholder
	p.API.LogDebug("Polling subscription",
		"user_id", sub.DataminrUser,
		"channel_id", sub.ChannelID,
		"interval", sub.PollInterval)

	return nil
}

// getAllConnectedDataminrUsers returns a list of all user IDs with Dataminr connections
func (p *Plugin) getAllConnectedDataminrUsers() ([]string, error) {
	// List all keys in KV store
	keys, appErr := p.API.KVList(0, 100)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to list KV keys")
	}

	// Filter for user info keys and extract user IDs
	var users []string
	for _, key := range keys {
		if userID, found := strings.CutSuffix(key, userInfoKeyPrefix); found {
			users = append(users, userID)
		}
	}

	return users, nil
}

// pollUserAlerts fetches and processes alerts for a single user
func (p *Plugin) pollUserAlerts(userID string) error {
	// Get user info
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user info")
	}

	if userInfo == nil {
		return errors.New("user not connected to Dataminr")
	}

	// Get cursor for pagination
	cursor, err := p.getDataminrCursor(userID)
	if err != nil {
		// Log but continue - will start from beginning
		p.API.LogWarn("Failed to get cursor", "user_id", userID, "error", err.Error())
	}

	// In a real implementation, we would:
	// 1. Get credentials and create Dataminr client
	// 2. Fetch alerts using the cursor
	// 3. Process alerts
	// 4. Update cursor
	// 5. Update last poll time

	// For now, just mark the cursor as used (placeholder)
	_ = cursor

	return nil
}

// processAlerts deduplicates and routes alerts for a user
func (p *Plugin) processAlerts(userID string, alerts []dataminr.Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	// Deduplicate alerts
	newAlerts := p.deduplicateAlerts(alerts)

	// Route each new alert
	for _, alert := range newAlerts {
		if err := p.routeAlert(userID, &alert); err != nil {
			// Log error but continue with other alerts
			p.API.LogWarn("Failed to route alert", "alert_id", alert.AlertID, "error", err.Error())
		}
	}

	return nil
}

// updateLastPollTime updates the last poll timestamp for a user
func (p *Plugin) updateLastPollTime(userID string) error {
	userInfo, err := p.getUserInfo(userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user info")
	}

	if userInfo == nil {
		return errors.New("user info not found")
	}

	// Update last poll time
	userInfo.LastPollAt = time.Now().Unix()

	// Store updated user info
	if err := p.storeUserInfo(userInfo); err != nil {
		return errors.Wrap(err, "failed to store user info")
	}

	return nil
}
