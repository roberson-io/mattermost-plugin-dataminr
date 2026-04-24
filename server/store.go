package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr/client"
)

const (
	credentialsKeyPrefix  = "_dataminr_credentials"     //nolint:gosec // Not a credential, just a KV store key prefix
	tokenKeyPrefix        = "_dataminr_token"           //nolint:gosec // Not a credential, just a KV store key prefix
	tokenIssuedAtPrefix   = "_dataminr_token_issued_at" //nolint:gosec // Not a credential, just a KV store key prefix
	cursorKeyPrefix       = "_dataminr_cursor"
	userInfoKeyPrefix     = "_dataminr_userinfo"
	tokenExpiryBufferSecs = 300 // Refresh token 5 minutes before expiry
)

// storeDataminrCredentials encrypts and stores user credentials in the KV store
func (p *Plugin) storeDataminrCredentials(userID string, credentials *dataminr.Credentials) error {
	config := p.getConfiguration()

	// Marshal credentials to JSON
	jsonData, err := json.Marshal(credentials)
	if err != nil {
		return errors.Wrap(err, "failed to marshal credentials")
	}

	// Encrypt the JSON data
	encrypted, err := encrypt([]byte(config.DataminrEncryptionKey), string(jsonData))
	if err != nil {
		return errors.Wrap(err, "failed to encrypt credentials")
	}

	// Store in KV
	key := userID + credentialsKeyPrefix
	if appErr := p.API.KVSet(key, []byte(encrypted)); appErr != nil {
		return errors.Wrap(appErr, "failed to store credentials")
	}

	return nil
}

// getDataminrCredentials retrieves and decrypts user credentials from the KV store
func (p *Plugin) getDataminrCredentials(userID string) (*dataminr.Credentials, error) {
	config := p.getConfiguration()

	// Retrieve from KV
	key := userID + credentialsKeyPrefix
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get credentials")
	}

	if data == nil {
		return nil, errors.New("credentials not found")
	}

	// Decrypt the data
	decrypted, err := decrypt([]byte(config.DataminrEncryptionKey), string(data))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt credentials")
	}

	// Unmarshal JSON
	var credentials dataminr.Credentials
	if err := json.Unmarshal([]byte(decrypted), &credentials); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal credentials")
	}

	return &credentials, nil
}

// storeDataminrToken encrypts and stores the bearer token in the KV store along with timestamp
func (p *Plugin) storeDataminrToken(userID string, token string) error {
	config := p.getConfiguration()

	// Encrypt the token
	encrypted, err := encrypt([]byte(config.DataminrEncryptionKey), token)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt token")
	}

	// Store token in KV
	key := userID + tokenKeyPrefix
	if appErr := p.API.KVSet(key, []byte(encrypted)); appErr != nil {
		return errors.Wrap(appErr, "failed to store token")
	}

	// Store timestamp
	timestampKey := userID + tokenIssuedAtPrefix
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	if appErr := p.API.KVSet(timestampKey, []byte(timestamp)); appErr != nil {
		return errors.Wrap(appErr, "failed to store token timestamp")
	}

	return nil
}

// getDataminrToken retrieves and decrypts the bearer token from the KV store
func (p *Plugin) getDataminrToken(userID string) (string, error) {
	config := p.getConfiguration()

	// Retrieve from KV
	key := userID + tokenKeyPrefix
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return "", errors.Wrap(appErr, "failed to get token")
	}

	if data == nil {
		return "", nil // Token not found is not an error, just empty
	}

	// Decrypt the data
	decrypted, err := decrypt([]byte(config.DataminrEncryptionKey), string(data))
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt token")
	}

	return decrypted, nil
}

// getDataminrTokenIssuedAt retrieves the timestamp when the token was issued
func (p *Plugin) getDataminrTokenIssuedAt(userID string) (int64, error) {
	key := userID + tokenIssuedAtPrefix
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return 0, errors.Wrap(appErr, "failed to get token timestamp")
	}

	if data == nil {
		return 0, nil // No timestamp found
	}

	timestamp, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse token timestamp")
	}

	return timestamp, nil
}

// isTokenValid checks if the cached token is still valid (not expired)
func (p *Plugin) isTokenValid(userID string) bool {
	issuedAt, err := p.getDataminrTokenIssuedAt(userID)
	if err != nil || issuedAt == 0 {
		return false
	}

	elapsed := time.Now().Unix() - issuedAt
	// Token is valid if it hasn't exceeded expiry time minus buffer
	return elapsed < (dataminr.TokenExpirySeconds - tokenExpiryBufferSecs)
}

// getOrRefreshToken returns a valid token, either from cache or by fetching a new one
func (p *Plugin) getOrRefreshToken(userID string) (string, error) {
	// Check if we have a valid cached token
	if p.isTokenValid(userID) {
		token, err := p.getDataminrToken(userID)
		if err == nil && token != "" {
			return token, nil
		}
	}

	// Need to fetch a new token
	credentials, err := p.getDataminrCredentials(userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get credentials")
	}

	apiClient := client.NewClient(credentials, dataminrAPIBaseURL)
	token, err := apiClient.GetToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to get token from Dataminr")
	}

	// Cache the new token
	if err := p.storeDataminrToken(userID, token); err != nil {
		// Log but don't fail - we still have a valid token
		p.API.LogWarn("Failed to cache token", "userID", userID, "error", err.Error())
	}

	return token, nil
}

// storeDataminrCursor stores the pagination cursor in the KV store
func (p *Plugin) storeDataminrCursor(userID string, cursor string) error {
	key := userID + cursorKeyPrefix
	if appErr := p.API.KVSet(key, []byte(cursor)); appErr != nil {
		return errors.Wrap(appErr, "failed to store cursor")
	}
	return nil
}

// getDataminrCursor retrieves the pagination cursor from the KV store
func (p *Plugin) getDataminrCursor(userID string) (string, error) {
	key := userID + cursorKeyPrefix
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return "", errors.Wrap(appErr, "failed to get cursor")
	}

	if data == nil {
		return "", nil // Cursor not found is not an error, just empty
	}

	return string(data), nil
}

// storeUserInfo stores user information in the KV store
func (p *Plugin) storeUserInfo(userInfo *dataminr.UserInfo) error {
	// Marshal to JSON
	jsonData, err := json.Marshal(userInfo)
	if err != nil {
		return errors.Wrap(err, "failed to marshal user info")
	}

	// Store in KV
	key := userInfo.MattermostUserID + userInfoKeyPrefix
	if appErr := p.API.KVSet(key, jsonData); appErr != nil {
		return errors.Wrap(appErr, "failed to store user info")
	}

	return nil
}

// getUserInfo retrieves user information from the KV store
func (p *Plugin) getUserInfo(userID string) (*dataminr.UserInfo, error) {
	key := userID + userInfoKeyPrefix
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get user info")
	}

	if data == nil {
		return nil, nil // User info not found is not an error, just nil
	}

	// Unmarshal JSON
	var userInfo dataminr.UserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal user info")
	}

	return &userInfo, nil
}

// deleteDataminrCredentials removes user credentials from the KV store
func (p *Plugin) deleteDataminrCredentials(userID string) error {
	key := userID + credentialsKeyPrefix
	if appErr := p.API.KVDelete(key); appErr != nil {
		return errors.Wrap(appErr, "failed to delete credentials")
	}
	return nil
}

// deleteDataminrToken removes the bearer token from the KV store
func (p *Plugin) deleteDataminrToken(userID string) error {
	key := userID + tokenKeyPrefix
	if appErr := p.API.KVDelete(key); appErr != nil {
		return errors.Wrap(appErr, "failed to delete token")
	}
	return nil
}

// deleteDataminrCursor removes the pagination cursor from the KV store
func (p *Plugin) deleteDataminrCursor(userID string) error {
	key := userID + cursorKeyPrefix
	if appErr := p.API.KVDelete(key); appErr != nil {
		return errors.Wrap(appErr, "failed to delete cursor")
	}
	return nil
}

// deleteUserInfo removes user information from the KV store
func (p *Plugin) deleteUserInfo(userID string) error {
	key := userID + userInfoKeyPrefix
	if appErr := p.API.KVDelete(key); appErr != nil {
		return errors.Wrap(appErr, "failed to delete user info")
	}
	return nil
}

const subscriptionsKey = "dataminr_subscriptions"

// storeSubscriptions stores all subscriptions in the KV store
func (p *Plugin) storeSubscriptions(subs *dataminr.Subscriptions) error {
	jsonData, err := json.Marshal(subs)
	if err != nil {
		return errors.Wrap(err, "failed to marshal subscriptions")
	}

	if appErr := p.API.KVSet(subscriptionsKey, jsonData); appErr != nil {
		return errors.Wrap(appErr, "failed to store subscriptions")
	}

	return nil
}

// getSubscriptions retrieves all subscriptions from the KV store
func (p *Plugin) getSubscriptions() (*dataminr.Subscriptions, error) {
	data, appErr := p.API.KVGet(subscriptionsKey)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get subscriptions")
	}

	if data == nil {
		// No subscriptions stored yet, return empty
		return dataminr.NewSubscriptions(), nil
	}

	var subs dataminr.Subscriptions
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal subscriptions")
	}

	// Ensure Users map is initialized
	if subs.Users == nil {
		subs.Users = make(map[string][]*dataminr.Subscription)
	}

	return &subs, nil
}

// addSubscription adds a new subscription for a channel
func (p *Plugin) addSubscription(channelID, creatorID, dataminrUserID string) error {
	subs, err := p.getSubscriptions()
	if err != nil {
		return errors.Wrap(err, "failed to get existing subscriptions")
	}

	sub := dataminr.NewSubscription(channelID, creatorID, dataminrUserID)
	subs.Add(sub)

	if err := p.storeSubscriptions(subs); err != nil {
		return errors.Wrap(err, "failed to store updated subscriptions")
	}

	return nil
}

// removeSubscription removes a subscription for a channel and user
// Returns true if removed, false if not found
func (p *Plugin) removeSubscription(channelID, dataminrUserID string) (bool, error) {
	subs, err := p.getSubscriptions()
	if err != nil {
		return false, errors.Wrap(err, "failed to get existing subscriptions")
	}

	removed := subs.Remove(channelID, dataminrUserID)
	if !removed {
		return false, nil
	}

	if err := p.storeSubscriptions(subs); err != nil {
		return false, errors.Wrap(err, "failed to store updated subscriptions")
	}

	return true, nil
}
