package main

import (
	"fmt"
	"time"
)

const (
	// DefaultPollInterval is the default polling interval (2 minutes)
	DefaultPollInterval = 2 * time.Minute
	// DefaultMinPollInterval is the default minimum polling interval (30 seconds)
	DefaultMinPollInterval = 30 * time.Second
)

// getPollInterval returns the configured polling interval for the Dataminr job
func (p *Plugin) getPollInterval() time.Duration {
	config := p.getConfiguration()

	// Get minimum interval (default to 30 seconds)
	minInterval := config.DataminrMinPollInterval
	if minInterval <= 0 {
		minInterval = int(DefaultMinPollInterval.Seconds())
	}

	// Get configured interval (default to 2 minutes)
	interval := config.DataminrDefaultPollInterval
	if interval <= 0 {
		interval = int(DefaultPollInterval.Seconds())
	}

	// Enforce minimum
	if interval < minInterval {
		interval = minInterval
	}

	return time.Duration(interval) * time.Second
}

// shouldStartDataminrJob returns true if the Dataminr job should be started
func (p *Plugin) shouldStartDataminrJob() bool {
	config := p.getConfiguration()
	if config == nil {
		return false
	}
	return config.DataminrEnabled
}

// validatePollInterval validates that a poll interval meets the minimum requirements
func (p *Plugin) validatePollInterval(interval int) error {
	// 0 is always valid (manual only mode)
	if interval == 0 {
		return nil
	}

	config := p.getConfiguration()

	// Get minimum interval
	minInterval := config.DataminrMinPollInterval
	if minInterval <= 0 {
		minInterval = int(DefaultMinPollInterval.Seconds())
	}

	// Non-zero must meet minimum
	if interval < minInterval {
		return fmt.Errorf("polling interval must be 0 (manual only) or at least %d seconds", minInterval)
	}

	return nil
}
