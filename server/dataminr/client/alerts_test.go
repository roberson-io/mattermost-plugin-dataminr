package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAlerts_Success(t *testing.T) {
	// Create mock HTTP server for alerts
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/firstalert/v1/alerts", r.URL.Path)

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-token", authHeader)

		// Return successful response with alerts
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.AlertResponse{
			Alerts: []dataminr.Alert{
				{
					AlertID:        "test-alert-1",
					AlertTimestamp: "2025-07-07T19:19:00.397Z",
					Headline:       "Test alert headline",
					AlertType: &dataminr.AlertType{
						Name: "Urgent",
					},
				},
			},
			NextPage: "/v1/alerts?from=cursor123",
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// Get alerts
	alertResp, err := client.GetAlerts("test-token", "")
	require.NoError(t, err)
	require.NotNil(t, alertResp)
	assert.Len(t, alertResp.Alerts, 1)
	assert.Equal(t, "test-alert-1", alertResp.Alerts[0].AlertID)
	assert.Equal(t, "/v1/alerts?from=cursor123", alertResp.NextPage)
}

func TestGetAlerts_WithCursor(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify cursor is passed correctly
		cursor := r.URL.Query().Get("from")
		assert.Equal(t, "cursor-abc123", cursor)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.AlertResponse{
			Alerts:   []dataminr.Alert{},
			NextPage: "/v1/alerts?from=cursor-abc123",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// Get alerts with cursor
	alertResp, err := client.GetAlerts("test-token", "cursor-abc123")
	require.NoError(t, err)
	require.NotNil(t, alertResp)
	assert.Len(t, alertResp.Alerts, 0)
}

func TestGetAlerts_WithNextPageURL(t *testing.T) {
	// Test that nextPage URL (starting with /) is handled correctly
	// per API spec: use base URL + relative nextPage URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the path includes the nextPage path
		assert.Equal(t, "/firstalert/v1/alerts", r.URL.Path)
		// Verify cursor from nextPage URL is in query (URL-decoded by server)
		cursor := r.URL.Query().Get("from")
		assert.Equal(t, "2wVWwq3bBSqy/tkFROaX2wUysoSh", cursor)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.AlertResponse{
			Alerts: []dataminr.Alert{
				{
					AlertID:  "new-alert-1",
					Headline: "New alert after cursor",
				},
			},
			NextPage: "/v1/alerts?from=newcursor123",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// Pass the nextPage URL format (starts with /) as the cursor
	// This simulates what happens when we store alertResp.NextPage and use it
	alertResp, err := client.GetAlerts("test-token", "/v1/alerts?from=2wVWwq3bBSqy%2FtkFROaX2wUysoSh")
	require.NoError(t, err)
	require.NotNil(t, alertResp)
	assert.Len(t, alertResp.Alerts, 1)
	assert.Equal(t, "new-alert-1", alertResp.Alerts[0].AlertID)
}

func TestGetAlerts_WithPageSize(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify pageSize is passed correctly
		pageSize := r.URL.Query().Get("pageSize")
		assert.Equal(t, "100", pageSize)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.AlertResponse{
			Alerts:   []dataminr.Alert{},
			NextPage: "/v1/alerts?from=cursor123",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	// Get alerts with page size
	alertResp, err := client.GetAlertsWithPageSize("test-token", "", 100)
	require.NoError(t, err)
	require.NotNil(t, alertResp)
}

func TestGetAlerts_EmptyAlerts(t *testing.T) {
	// Create mock HTTP server that returns empty alerts (end of stream)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.AlertResponse{
			Alerts:   []dataminr.Alert{},
			NextPage: "/v1/alerts?from=cursor123",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	alertResp, err := client.GetAlerts("test-token", "cursor123")
	require.NoError(t, err)
	require.NotNil(t, alertResp)
	assert.Len(t, alertResp.Alerts, 0)
	assert.Equal(t, "/v1/alerts?from=cursor123", alertResp.NextPage)
}

func TestGetAlerts_UnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	alertResp, err := client.GetAlerts("invalid-token", "")
	assert.Error(t, err)
	assert.Nil(t, alertResp)
	assert.Contains(t, err.Error(), "401")
}

func TestGetAlerts_NetworkError(t *testing.T) {
	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, "http://invalid-host-that-does-not-exist-12345.com")

	alertResp, err := client.GetAlerts("test-token", "")
	assert.Error(t, err)
	assert.Nil(t, alertResp)
}

func TestGetAlerts_MalformedJSONResponse(t *testing.T) {
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

	alertResp, err := client.GetAlerts("test-token", "")
	assert.Error(t, err)
	assert.Nil(t, alertResp)
	assert.Contains(t, err.Error(), "failed to decode")
}

func TestGetAlerts_ServerError(t *testing.T) {
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

	alertResp, err := client.GetAlerts("test-token", "")
	assert.Error(t, err)
	assert.Nil(t, alertResp)
	assert.Contains(t, err.Error(), "500")
}

func TestGetAlerts_ComplexAlert(t *testing.T) {
	// Test with a more complex alert structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := dataminr.AlertResponse{
			Alerts: []dataminr.Alert{
				{
					AlertID:        "16105705780476211705283-1",
					AlertTimestamp: "2025-07-07T19:19:00.397Z",
					AlertType: &dataminr.AlertType{
						Name: "Flash",
					},
					Headline: "Fire department on scene of structure fire",
					SubHeadline: &dataminr.SubHeadline{
						Title:   "According to Emergency responder",
						Content: []string{"Reported on Fire Dispatch scanner feed"},
					},
					PublicPost: &dataminr.PublicPost{
						Href: "https://r.dataminr.com/1TYMJcWb807913668478549",
					},
					EstimatedEventLocation: &dataminr.EstimatedEventLocation{
						Name:              "Avenue C & W 22nd St, Bayonne, NJ 07002, USA",
						Coordinates:       []float64{40.6682, -74.1177},
						ProbabilityRadius: 0.12,
						MGRS:              "18T WL 74278 01882",
					},
					AlertTopics: []dataminr.AlertTopic{
						{
							ID:   "154032",
							Name: "Disasters and Weather - Structure Fires and Collapses",
						},
					},
					DataminrAlertURL: "https://firstalert.dataminr.com/#alertDetail/6/456145358-1761668883367-3",
					AlertReferenceTerms: []dataminr.AlertReferenceTerms{
						{Text: "structure fires"},
					},
					ListsMatched: []dataminr.ListsMatched{
						{
							ID:       "123",
							Name:     "Emergency Alerts",
							TopicIDs: []string{"154032"},
						},
					},
					LinkedAlerts: []dataminr.LinkedAlerts{
						{
							Count:         2,
							ParentAlertID: "881485414401001766138888-1712598032000-1",
						},
					},
					LiveBrief: []dataminr.LiveBrief{
						{
							Summary:   "Fire reported at location",
							Timestamp: "2025-07-07T19:19:00.397Z",
							Version:   "current",
						},
					},
					TermsOfUse: "https://www.dataminr.com/legal/firstalert-tos",
				},
			},
			NextPage: "/v1/alerts?from=next-cursor",
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	credentials := &dataminr.Credentials{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	client := NewClient(credentials, server.URL)

	alertResp, err := client.GetAlerts("test-token", "")
	require.NoError(t, err)
	require.NotNil(t, alertResp)
	require.Len(t, alertResp.Alerts, 1)

	alert := alertResp.Alerts[0]
	assert.Equal(t, "16105705780476211705283-1", alert.AlertID)
	assert.Equal(t, "Flash", alert.AlertType.Name)
	assert.Equal(t, "Fire department on scene of structure fire", alert.Headline)
	assert.NotNil(t, alert.SubHeadline)
	assert.Equal(t, "According to Emergency responder", alert.SubHeadline.Title)
	assert.NotNil(t, alert.EstimatedEventLocation)
	assert.Equal(t, 40.6682, alert.EstimatedEventLocation.Coordinates[0])
	assert.Len(t, alert.AlertTopics, 1)
	assert.Equal(t, "154032", alert.AlertTopics[0].ID)
	assert.Len(t, alert.LinkedAlerts, 1)
	assert.Equal(t, 2, alert.LinkedAlerts[0].Count)
	assert.Len(t, alert.LiveBrief, 1)
	assert.Equal(t, "current", alert.LiveBrief[0].Version)
}
