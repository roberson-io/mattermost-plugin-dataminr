package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
)

// Client is the Dataminr API client
type Client struct {
	credentials   *dataminr.Credentials
	baseURL       string
	httpClient    *http.Client
	cachedToken   string
	tokenIssuedAt int64
}

// NewClient creates a new Dataminr API client
func NewClient(credentials *dataminr.Credentials, baseURL string) *Client {
	return &Client{
		credentials: credentials,
		baseURL:     baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken retrieves a bearer token for API authentication
func (c *Client) GetToken() (string, error) {
	// Check if we have a cached token that hasn't expired
	if c.cachedToken != "" && !c.isTokenExpired() {
		return c.cachedToken, nil
	}

	// Fetch new token from API
	token, err := c.fetchToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to get authentication token")
	}

	// Cache the token
	c.cachedToken = token
	c.tokenIssuedAt = time.Now().Unix()

	return token, nil
}

// isTokenExpired checks if the cached token has expired
func (c *Client) isTokenExpired() bool {
	elapsed := time.Now().Unix() - c.tokenIssuedAt
	return elapsed >= dataminr.TokenExpirySeconds
}

// fetchToken makes the HTTP request to get a new token
func (c *Client) fetchToken() (string, error) {
	// Build form data
	data := url.Values{}
	data.Set("client_id", c.credentials.ClientID)
	data.Set("client_secret", c.credentials.ClientSecret)
	data.Set("grant_type", "api_key")

	// Create request
	authURL := c.baseURL + "/auth/v1/token"
	req, err := http.NewRequest(http.MethodPost, authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to make request")
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp dataminr.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", errors.Wrap(err, "failed to decode response")
	}

	// Validate token
	if tokenResp.DMAToken == "" {
		return "", errors.New("received empty token from API")
	}

	return tokenResp.DMAToken, nil
}

// GetAlerts retrieves alerts from the Dataminr API with default page size (40)
func (c *Client) GetAlerts(token string, cursor string) (*dataminr.AlertResponse, error) {
	return c.GetAlertsWithPageSize(token, cursor, 40)
}

// GetAlertsWithPageSize retrieves alerts with a specific page size
func (c *Client) GetAlertsWithPageSize(token string, cursor string, pageSize int) (*dataminr.AlertResponse, error) {
	// Build request
	req, err := c.buildAlertsRequest(token, cursor, pageSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("alerts request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var alertResp dataminr.AlertResponse
	if err := json.NewDecoder(resp.Body).Decode(&alertResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &alertResp, nil
}

// buildAlertsRequest constructs the HTTP request for fetching alerts
func (c *Client) buildAlertsRequest(token string, cursor string, pageSize int) (*http.Request, error) {
	var alertsURL string

	// If cursor is provided and looks like a nextPage URL (starts with /), use it directly
	// per API spec: "Use the full URL: The subsequent request URL should be constructed
	// using the base API URL + the relative URL value provided in the nextPage field"
	if cursor != "" && len(cursor) > 0 && cursor[0] == '/' {
		// cursor is a nextPage value like "/v1/alerts?from=xyz" - append to base URL
		alertsURL = c.baseURL + "/firstalert" + cursor
	} else {
		// No cursor or legacy cursor format - build URL with parameters
		alertsURL = c.baseURL + "/firstalert/v1/alerts"
	}

	req, err := http.NewRequest(http.MethodGet, alertsURL, nil)
	if err != nil {
		return nil, err
	}

	// Only add query parameters if we're not using a nextPage URL
	// (nextPage URLs already contain the cursor)
	if cursor == "" || (len(cursor) > 0 && cursor[0] != '/') {
		q := req.URL.Query()
		q.Add("pageSize", fmt.Sprintf("%d", pageSize))
		if cursor != "" {
			q.Add("from", cursor)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	return req, nil
}
