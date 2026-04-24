package dataminr

import (
	"errors"
	"time"
)

// Subscription represents a channel subscription to a user's Dataminr alerts
type Subscription struct {
	ChannelID    string `json:"channel_id"`
	CreatorID    string `json:"creator_id"`
	DataminrUser string `json:"dataminr_user"`
	Enabled      bool   `json:"enabled"`
	Features     string `json:"features"`
	CreatedAt    int64  `json:"created_at"`
	PollInterval int    `json:"poll_interval"`
	LastPollAt   int64  `json:"last_poll_at"` // Unix timestamp of last poll for this subscription
}

// Subscriptions manages all channel subscriptions
type Subscriptions struct {
	// Maps: dataminr_user_id -> []*Subscription
	Users map[string][]*Subscription `json:"users"`
}

// Validate checks if the subscription has all required fields
func (s *Subscription) Validate() error {
	if s.ChannelID == "" {
		return errors.New("channel_id is required")
	}
	if s.CreatorID == "" {
		return errors.New("creator_id is required")
	}
	if s.DataminrUser == "" {
		return errors.New("dataminr_user is required")
	}
	return nil
}

// NewSubscription creates a new subscription with default values
func NewSubscription(channelID, creatorID, dataminrUser string) *Subscription {
	return &Subscription{
		ChannelID:    channelID,
		CreatorID:    creatorID,
		DataminrUser: dataminrUser,
		Enabled:      true,
		Features:     "all",
		CreatedAt:    time.Now().Unix(),
		PollInterval: 0, // Manual only by default
	}
}

// NewSubscriptions creates a new empty Subscriptions container
func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		Users: make(map[string][]*Subscription),
	}
}

// Add adds a subscription for a Dataminr user
func (s *Subscriptions) Add(sub *Subscription) {
	if s.Users == nil {
		s.Users = make(map[string][]*Subscription)
	}
	s.Users[sub.DataminrUser] = append(s.Users[sub.DataminrUser], sub)
}

// GetByDataminrUser returns all subscriptions for a given Dataminr user
func (s *Subscriptions) GetByDataminrUser(dataminrUserID string) []*Subscription {
	if s.Users == nil {
		return []*Subscription{}
	}
	subs, ok := s.Users[dataminrUserID]
	if !ok {
		return []*Subscription{}
	}
	return subs
}

// GetByChannel returns all subscriptions for a given channel
func (s *Subscriptions) GetByChannel(channelID string) []*Subscription {
	var result []*Subscription
	for _, userSubs := range s.Users {
		for _, sub := range userSubs {
			if sub.ChannelID == channelID {
				result = append(result, sub)
			}
		}
	}
	if result == nil {
		return []*Subscription{}
	}
	return result
}

// Remove removes a subscription for a channel and Dataminr user
// Returns true if a subscription was removed, false if not found
func (s *Subscriptions) Remove(channelID, dataminrUserID string) bool {
	subs, ok := s.Users[dataminrUserID]
	if !ok {
		return false
	}

	for i, sub := range subs {
		if sub.ChannelID == channelID {
			// Remove the subscription by slicing
			s.Users[dataminrUserID] = append(subs[:i], subs[i+1:]...)
			return true
		}
	}
	return false
}
