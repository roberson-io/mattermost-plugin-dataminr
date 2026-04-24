package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken_Success(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/v1/token", r.URL.Path)

		// Verify Content-Type
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Parse form data
		err := r.ParseForm()
		require.NoError(t, err)

		// Verify credentials
		assert.Equal(t, "test-client-id", r.Form.Get("client_id"))
		assert.Equal(t, "test-client-secret", r.Form.Get("client_secret"))
		assert.Equal(t, "api_key", r.Form.Get("grant_type"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.TokenResponse{
			DMAToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token",
			ExpiresIn: dataminr.TokenExpirySeconds,
		}
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Create client with test server URL
	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// Get token
	token, err := client.GetToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token", token)
}

func TestGetToken_Caching(t *testing.T) {
	callCount := 0

	// Create mock HTTP server that tracks calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.TokenResponse{
			DMAToken:  "cached-token",
			ExpiresIn: dataminr.TokenExpirySeconds,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// First call should hit the server
	token1, err := client.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "cached-token", token1)
	assert.Equal(t, 1, callCount)

	// Second call should use cached token
	token2, err := client.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "cached-token", token2)
	assert.Equal(t, 1, callCount, "Should not make second API call")
}

func TestGetToken_RefreshWhenExpired(t *testing.T) {
	callCount := 0
	tokenValue := "token-v1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 2 {
			tokenValue = "token-v2"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.TokenResponse{
			DMAToken:  tokenValue,
			ExpiresIn: dataminr.TokenExpirySeconds,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// First call
	token1, err := client.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "token-v1", token1)
	assert.Equal(t, 1, callCount)

	// Simulate token expiry by setting issued time to 4+ hours ago
	client.tokenIssuedAt = time.Now().Add(-4*time.Hour - 1*time.Minute).Unix()

	// Second call should refresh token
	token2, err := client.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "token-v2", token2)
	assert.Equal(t, 2, callCount, "Should refresh expired token")
}

func TestGetToken_AuthenticationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"Invalid credentials"}`))
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "invalid-client-id",
		ClientSecret: "invalid-client-secret",
	}
	client := NewClient(credentials, server.URL)

	token, err := client.GetToken()
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "401")
}

func TestGetToken_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, "http://invalid-host-that-does-not-exist-12345.com")

	token, err := client.GetToken()
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestGetToken_MalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	token, err := client.GetToken()
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to decode")
}

func TestGetToken_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.TokenResponse{
			DMAToken:  "",
			ExpiresIn: dataminr.TokenExpirySeconds,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	token, err := client.GetToken()
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "empty token")
}

func TestGetToken_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	token, err := client.GetToken()
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "500")
}

func TestNewClient(t *testing.T) {
	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, "https://test.dataminr.com")

	assert.NotNil(t, client)
	assert.Equal(t, credentials, client.credentials)
	assert.Equal(t, "https://test.dataminr.com", client.baseURL)
	assert.NotNil(t, client.httpClient)
}
