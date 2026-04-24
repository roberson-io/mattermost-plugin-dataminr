package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	plugin.router = plugin.initRouter()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
	r.Header.Set("Mattermost-User-ID", "test-user-id")

	plugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer func() { _ = result.Body.Close() }()
	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)
	bodyString := string(bodyBytes)

	assert.Equal("Hello, world!", bodyString)
}

func TestStoreDataminrCredentials(t *testing.T) {
	t.Run("successfully encrypts and stores credentials", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		}

		userID := "user123"
		credentials := &dataminr.Credentials{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		// Expect the KV store to be called with encrypted data
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return key == userID+"_dataminr_credentials"
		}), mock.AnythingOfType("[]uint8")).Return(nil)

		err := plugin.storeDataminrCredentials(userID, credentials)
		require.NoError(t, err)
	})

	t.Run("returns error when encryption fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "", // Empty key will cause encryption to fail
		}

		userID := "user123"
		credentials := &dataminr.Credentials{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		err := plugin.storeDataminrCredentials(userID, credentials)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption key cannot be empty")
	})

	t.Run("returns error when KV store fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		}

		userID := "user123"
		credentials := &dataminr.Credentials{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		// Simulate KV store failure
		api.On("KVSet", mock.Anything, mock.Anything).Return(&model.AppError{Message: "store error"})

		err := plugin.storeDataminrCredentials(userID, credentials)
		assert.Error(t, err)
	})
}

func TestGetDataminrCredentials(t *testing.T) {
	t.Run("successfully retrieves and decrypts credentials", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		encryptionKey := "test-encryption-key-32-bytes!!!"
		plugin.configuration = &configuration{
			DataminrEncryptionKey: encryptionKey,
		}

		userID := "user123"
		originalCredentials := &dataminr.Credentials{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		// Manually encrypt the credentials to simulate what's in the KV store
		jsonData, err := json.Marshal(originalCredentials)
		require.NoError(t, err)

		encrypted, err := encrypt([]byte(encryptionKey), string(jsonData))
		require.NoError(t, err)

		// Mock the KV store to return encrypted data
		api.On("KVGet", userID+"_dataminr_credentials").Return([]byte(encrypted), nil)

		credentials, err := plugin.getDataminrCredentials(userID)
		require.NoError(t, err)
		require.NotNil(t, credentials)
		assert.Equal(t, originalCredentials.ClientID, credentials.ClientID)
		assert.Equal(t, originalCredentials.ClientSecret, credentials.ClientSecret)
	})

	t.Run("returns error when credentials not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		}

		userID := "user123"

		// Mock KV store returning nil (not found)
		api.On("KVGet", userID+"_dataminr_credentials").Return(nil, nil)

		credentials, err := plugin.getDataminrCredentials(userID)
		assert.Error(t, err)
		assert.Nil(t, credentials)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error when decryption fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		}

		userID := "user123"

		// Return corrupted/invalid encrypted data
		api.On("KVGet", userID+"_dataminr_credentials").Return([]byte("corrupted-data"), nil)

		credentials, err := plugin.getDataminrCredentials(userID)
		assert.Error(t, err)
		assert.Nil(t, credentials)
	})
}

func TestStoreDataminrToken(t *testing.T) {
	t.Run("successfully encrypts and stores token", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		}

		userID := "user123"
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." //nolint:gosec // Test token, not a real credential

		api.On("KVSet", userID+"_dataminr_token", mock.AnythingOfType("[]uint8")).Return(nil)

		err := plugin.storeDataminrToken(userID, token)
		require.NoError(t, err)
	})
}

func TestGetDataminrToken(t *testing.T) {
	t.Run("successfully retrieves and decrypts token", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		encryptionKey := "test-encryption-key-32-bytes!!!"
		plugin.configuration = &configuration{
			DataminrEncryptionKey: encryptionKey,
		}

		userID := "user123"
		originalToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." //nolint:gosec // Test token, not a real credential

		encrypted, err := encrypt([]byte(encryptionKey), originalToken)
		require.NoError(t, err)

		api.On("KVGet", userID+"_dataminr_token").Return([]byte(encrypted), nil)

		token, err := plugin.getDataminrToken(userID)
		require.NoError(t, err)
		assert.Equal(t, originalToken, token)
	})

	t.Run("returns empty string when token not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.configuration = &configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		}

		userID := "user123"

		api.On("KVGet", userID+"_dataminr_token").Return(nil, nil)

		token, err := plugin.getDataminrToken(userID)
		require.NoError(t, err) // Token not found is not an error, just empty
		assert.Equal(t, "", token)
	})
}

func TestStoreDataminrCursor(t *testing.T) {
	t.Run("successfully stores cursor", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		cursor := "cursor-abc123"

		api.On("KVSet", userID+"_dataminr_cursor", []byte(cursor)).Return(nil)

		err := plugin.storeDataminrCursor(userID, cursor)
		require.NoError(t, err)
	})
}

func TestGetDataminrCursor(t *testing.T) {
	t.Run("successfully retrieves cursor", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"
		expectedCursor := "cursor-abc123"

		api.On("KVGet", userID+"_dataminr_cursor").Return([]byte(expectedCursor), nil)

		cursor, err := plugin.getDataminrCursor(userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCursor, cursor)
	})

	t.Run("returns empty string when cursor not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userID := "user123"

		api.On("KVGet", userID+"_dataminr_cursor").Return(nil, nil)

		cursor, err := plugin.getDataminrCursor(userID)
		require.NoError(t, err)
		assert.Equal(t, "", cursor)
	})
}

func TestStoreUserInfo(t *testing.T) {
	t.Run("successfully stores user info", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		userInfo := dataminr.NewUserInfo("user123")

		api.On("KVSet", "user123_dataminr_userinfo", mock.AnythingOfType("[]uint8")).Return(nil)

		err := plugin.storeUserInfo(userInfo)
		require.NoError(t, err)
	})
}

func TestGetUserInfo(t *testing.T) {
	t.Run("successfully retrieves user info", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		originalUserInfo := dataminr.NewUserInfo("user123")
		jsonData, err := json.Marshal(originalUserInfo)
		require.NoError(t, err)

		api.On("KVGet", "user123_dataminr_userinfo").Return(jsonData, nil)

		userInfo, err := plugin.getUserInfo("user123")
		require.NoError(t, err)
		require.NotNil(t, userInfo)
		assert.Equal(t, originalUserInfo.MattermostUserID, userInfo.MattermostUserID)
	})

	t.Run("returns nil when user info not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVGet", "user123_dataminr_userinfo").Return(nil, nil)

		userInfo, err := plugin.getUserInfo("user123")
		require.NoError(t, err)
		assert.Nil(t, userInfo)
	})
}

func TestExecuteCommand_Connect_Success(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	encryptionKey := "test-encryption-key-32-bytes!!!"
	plugin := &Plugin{}
	plugin.SetAPI(api)
	plugin.configuration = &configuration{
		DataminrEncryptionKey: encryptionKey,
	}

	// Mock: Check if user already has credentials (not found)
	api.On("KVGet", "user123_dataminr_userinfo").Return(nil, nil)

	// Mock: Store credentials (encrypted)
	api.On("KVSet", "user123_dataminr_credentials", mock.AnythingOfType("[]uint8")).Return(nil)

	// Mock: Store user info
	api.On("KVSet", "user123_dataminr_userinfo", mock.AnythingOfType("[]uint8")).Return(nil)

	response, err := plugin.HandleConnect("user123", "test-client-id", "test-client-secret")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Successfully connected")
}

func TestExecuteCommand_Connect_AlreadyConnected(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	encryptionKey := "test-encryption-key-32-bytes!!!"
	plugin := &Plugin{}
	plugin.SetAPI(api)
	plugin.configuration = &configuration{
		DataminrEncryptionKey: encryptionKey,
	}

	// Mock: User already has user info (already connected)
	existingUserInfo := dataminr.NewUserInfo("user123")
	userInfoJSON, _ := json.Marshal(existingUserInfo)
	api.On("KVGet", "user123_dataminr_userinfo").Return(userInfoJSON, nil)

	response, err := plugin.HandleConnect("user123", "test-client-id", "test-client-secret")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "already connected")
}

func TestExecuteCommand_Connect_StorageError(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	encryptionKey := "test-encryption-key-32-bytes!!!"
	plugin := &Plugin{}
	plugin.SetAPI(api)
	plugin.configuration = &configuration{
		DataminrEncryptionKey: encryptionKey,
	}

	// Mock: Check if user already connected (not found)
	api.On("KVGet", "user123_dataminr_userinfo").Return(nil, nil)

	// Mock: Store credentials fails
	api.On("KVSet", "user123_dataminr_credentials", mock.AnythingOfType("[]uint8")).Return(&model.AppError{Message: "storage error"})

	response, err := plugin.HandleConnect("user123", "test-client-id", "test-client-secret")
	require.NoError(t, err) // Command doesn't return error, just error message
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Failed to store credentials")
}

func TestExecuteCommand_Disconnect_Success(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: User is connected (has user info)
	existingUserInfo := dataminr.NewUserInfo("user123")
	userInfoJSON, _ := json.Marshal(existingUserInfo)
	api.On("KVGet", "user123_dataminr_userinfo").Return(userInfoJSON, nil)

	// Mock: Delete all user data
	api.On("KVDelete", "user123_dataminr_credentials").Return(nil)
	api.On("KVDelete", "user123_dataminr_token").Return(nil)
	api.On("KVDelete", "user123_dataminr_cursor").Return(nil)
	api.On("KVDelete", "user123_dataminr_userinfo").Return(nil)

	response, err := plugin.HandleDisconnect("user123")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Successfully disconnected")
}

func TestExecuteCommand_Disconnect_NotConnected(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: User is not connected (no user info)
	api.On("KVGet", "user123_dataminr_userinfo").Return(nil, nil)

	response, err := plugin.HandleDisconnect("user123")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "not connected")
}

func TestExecuteCommand_Disconnect_DeleteError(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: User is connected
	existingUserInfo := dataminr.NewUserInfo("user123")
	userInfoJSON, _ := json.Marshal(existingUserInfo)
	api.On("KVGet", "user123_dataminr_userinfo").Return(userInfoJSON, nil)

	// Mock: Delete credentials fails
	api.On("KVDelete", "user123_dataminr_credentials").Return(&model.AppError{Message: "delete error"})

	response, err := plugin.HandleDisconnect("user123")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Failed to disconnect")
}

func TestExecuteCommand_Status_Connected(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: User is connected (has user info)
	existingUserInfo := dataminr.NewUserInfo("user123")
	userInfoJSON, _ := json.Marshal(existingUserInfo)
	api.On("KVGet", "user123_dataminr_userinfo").Return(userInfoJSON, nil)

	response, err := plugin.HandleStatus("user123")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Connected to Dataminr")
}

func TestExecuteCommand_Status_NotConnected(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: User is not connected (no user info)
	api.On("KVGet", "user123_dataminr_userinfo").Return(nil, nil)

	response, err := plugin.HandleStatus("user123")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Not connected")
}

func TestExecuteCommand_Status_Error(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: Error checking user info
	api.On("KVGet", "user123_dataminr_userinfo").Return(nil, &model.AppError{Message: "kv error"})

	response, err := plugin.HandleStatus("user123")
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Failed to check status")
}

func TestExecuteCommand_Latest_NotConnected(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	plugin := &Plugin{}
	plugin.SetAPI(api)

	// Mock: User is not connected (no user info)
	api.On("KVGet", "user123_dataminr_userinfo").Return(nil, nil)

	response, err := plugin.HandleLatest("user123", 5)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "not connected")
}

func TestStoreSubscriptions(t *testing.T) {
	t.Run("successfully stores subscriptions", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		subs := dataminr.NewSubscriptions()
		subs.Add(dataminr.NewSubscription("channel123", "creator456", "user789"))

		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(nil)

		err := plugin.storeSubscriptions(subs)
		require.NoError(t, err)
	})

	t.Run("returns error when KV store fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		subs := dataminr.NewSubscriptions()

		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(&model.AppError{Message: "store error"})

		err := plugin.storeSubscriptions(subs)
		assert.Error(t, err)
	})
}

func TestGetSubscriptions(t *testing.T) {
	t.Run("successfully retrieves subscriptions", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// Create subscriptions and marshal to JSON
		original := dataminr.NewSubscriptions()
		original.Add(dataminr.NewSubscription("channel123", "creator456", "user789"))
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		api.On("KVGet", "dataminr_subscriptions").Return(jsonData, nil)

		subs, err := plugin.getSubscriptions()
		require.NoError(t, err)
		require.NotNil(t, subs)

		userSubs := subs.GetByDataminrUser("user789")
		require.Len(t, userSubs, 1)
		assert.Equal(t, "channel123", userSubs[0].ChannelID)
	})

	t.Run("returns empty subscriptions when not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)

		subs, err := plugin.getSubscriptions()
		require.NoError(t, err)
		require.NotNil(t, subs)

		// Should return empty subscriptions
		userSubs := subs.GetByDataminrUser("anyuser")
		assert.Len(t, userSubs, 0)
	})

	t.Run("returns error when KV get fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		api.On("KVGet", "dataminr_subscriptions").Return(nil, &model.AppError{Message: "get error"})

		subs, err := plugin.getSubscriptions()
		assert.Error(t, err)
		assert.Nil(t, subs)
	})
}

func TestAddSubscription(t *testing.T) {
	t.Run("successfully adds subscription", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// First call gets existing (empty) subscriptions
		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)
		// Then stores updated subscriptions
		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(nil)

		err := plugin.addSubscription("channel123", "creator456", "user789")
		require.NoError(t, err)
	})
}

func TestRemoveSubscription(t *testing.T) {
	t.Run("successfully removes subscription", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// Create existing subscriptions
		existing := dataminr.NewSubscriptions()
		existing.Add(dataminr.NewSubscription("channel123", "creator456", "user789"))
		jsonData, _ := json.Marshal(existing)

		api.On("KVGet", "dataminr_subscriptions").Return(jsonData, nil)
		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(nil)

		removed, err := plugin.removeSubscription("channel123", "user789")
		require.NoError(t, err)
		assert.True(t, removed)
	})

	t.Run("returns false when subscription not found", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// Empty subscriptions
		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)

		removed, err := plugin.removeSubscription("channel123", "user789")
		require.NoError(t, err)
		assert.False(t, removed)
	})
}

func TestHandleSubscribe(t *testing.T) {
	t.Run("user not connected returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:          "test-encryption-key-32-bytes!!!",
			DataminrSubscriptionPermission: "anyone",
		})

		userID := "user123"
		channelID := "channel456"

		// User is not connected (no user info)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(nil, nil)

		response, err := plugin.HandleSubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("permission denied returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:          "test-encryption-key-32-bytes!!!",
			DataminrSubscriptionPermission: "channel_admin",
		})

		userID := "user123"
		channelID := "channel456"
		teamID := "team789"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// User is a channel member but not admin
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

		response, err := plugin.HandleSubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "permission")
	})

	t.Run("successful subscription", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:          "test-encryption-key-32-bytes!!!",
			DataminrSubscriptionPermission: "anyone",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// User is a channel member (anyone mode)
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId: channelID,
			UserId:    userID,
		}, nil)

		// Get existing subscriptions (empty)
		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)

		// Store new subscription
		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(nil)

		response, err := plugin.HandleSubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "subscribed")
	})

	t.Run("already subscribed returns friendly message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey:          "test-encryption-key-32-bytes!!!",
			DataminrSubscriptionPermission: "anyone",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// User is a channel member (anyone mode)
		api.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
			ChannelId: channelID,
			UserId:    userID,
		}, nil)

		// Subscription already exists
		existing := dataminr.NewSubscriptions()
		existing.Add(dataminr.NewSubscription(channelID, userID, userID))
		jsonData, _ := json.Marshal(existing)
		api.On("KVGet", "dataminr_subscriptions").Return(jsonData, nil)

		response, err := plugin.HandleSubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "already subscribed")
	})
}

func TestHandleUnsubscribe(t *testing.T) {
	t.Run("user not connected returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is not connected (no user info)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(nil, nil)

		response, err := plugin.HandleUnsubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "not connected")
	})

	t.Run("subscription not found returns friendly message", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// No existing subscriptions
		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)

		response, err := plugin.HandleUnsubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "No subscription")
	})

	t.Run("successful unsubscription", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// User is connected
		userInfo := &dataminr.UserInfo{MattermostUserID: userID}
		userInfoJSON, _ := json.Marshal(userInfo)
		api.On("KVGet", userID+"_dataminr_userinfo").Return(userInfoJSON, nil)

		// Subscription exists
		existing := dataminr.NewSubscriptions()
		existing.Add(dataminr.NewSubscription(channelID, userID, userID))
		jsonData, _ := json.Marshal(existing)
		api.On("KVGet", "dataminr_subscriptions").Return(jsonData, nil)

		// Store updated subscriptions (after removal)
		api.On("KVSet", "dataminr_subscriptions", mock.AnythingOfType("[]uint8")).Return(nil)

		response, err := plugin.HandleUnsubscribe(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "unsubscribed")
	})
}

func TestHandleList(t *testing.T) {
	t.Run("no subscriptions in channel", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// No subscriptions exist
		api.On("KVGet", "dataminr_subscriptions").Return(nil, nil)

		response, err := plugin.HandleList(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "No subscriptions")
	})

	t.Run("lists subscriptions in channel", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"

		// Create existing subscriptions for this channel
		existing := dataminr.NewSubscriptions()
		existing.Add(dataminr.NewSubscription(channelID, userID, userID))
		jsonData, _ := json.Marshal(existing)
		api.On("KVGet", "dataminr_subscriptions").Return(jsonData, nil)

		// Get user info for display name
		api.On("GetUser", userID).Return(&model.User{
			Id:       userID,
			Username: "testuser",
		}, nil)

		response, err := plugin.HandleList(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "Subscriptions")
		assert.Contains(t, response.Text, "testuser")
	})

	t.Run("lists multiple subscriptions", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.setConfiguration(&configuration{
			DataminrEncryptionKey: "test-encryption-key-32-bytes!!!",
		})

		userID := "user123"
		channelID := "channel456"
		user2ID := "user789"

		// Create existing subscriptions for this channel from two users
		existing := dataminr.NewSubscriptions()
		existing.Add(dataminr.NewSubscription(channelID, userID, userID))
		existing.Add(dataminr.NewSubscription(channelID, user2ID, user2ID))
		jsonData, _ := json.Marshal(existing)
		api.On("KVGet", "dataminr_subscriptions").Return(jsonData, nil)

		// Get user info for both users
		api.On("GetUser", userID).Return(&model.User{
			Id:       userID,
			Username: "user1",
		}, nil)
		api.On("GetUser", user2ID).Return(&model.User{
			Id:       user2ID,
			Username: "user2",
		}, nil)

		response, err := plugin.HandleList(userID, channelID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Contains(t, response.Text, "Subscriptions")
		assert.Contains(t, response.Text, "user1")
		assert.Contains(t, response.Text, "user2")
	})
}
