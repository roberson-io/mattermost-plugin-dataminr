package main

import (
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

const (
	// alertKeyPrefix is the prefix for alert deduplication keys in KV store
	alertKeyPrefix = "dataminr_alert_"
	// alertTTLSeconds is 1 hour TTL for alert deduplication
	alertTTLSeconds = 3600
)

// deduplicateAlerts filters out alerts that have already been seen
// Returns only new alerts that haven't been processed before
func (p *Plugin) deduplicateAlerts(alerts []dataminr.Alert) []dataminr.Alert {
	if len(alerts) == 0 {
		return []dataminr.Alert{}
	}

	var newAlerts []dataminr.Alert

	for _, alert := range alerts {
		seen, err := p.isAlertSeen(alert.AlertID)
		if err != nil {
			// On error, default to treating as new (safe default - don't miss alerts)
			seen = false
		}

		if !seen {
			// Mark as seen and include in results
			_ = p.markAlertAsSeen(alert.AlertID)
			newAlerts = append(newAlerts, alert)
		}
	}

	if newAlerts == nil {
		return []dataminr.Alert{}
	}

	return newAlerts
}

// isAlertSeen checks if an alert has been seen before
func (p *Plugin) isAlertSeen(alertID string) (bool, error) {
	key := alertKeyPrefix + alertID
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return false, appErr
	}
	return data != nil, nil
}

// markAlertAsSeen stores the alert ID in KV store with 1 hour TTL
func (p *Plugin) markAlertAsSeen(alertID string) error {
	key := alertKeyPrefix + alertID
	// Store a simple "seen" marker with TTL
	appErr := p.API.KVSetWithExpiry(key, []byte("seen"), alertTTLSeconds)
	if appErr != nil {
		return appErr
	}
	return nil
}
