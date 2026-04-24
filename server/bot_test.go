package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/roberson-io/mattermost-plugin-dataminr/server/dataminr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBotUserID(t *testing.T) {
	t.Run("GetBotUserID returns stored bot ID", func(t *testing.T) {
		plugin := &Plugin{
			botUserID: "bot123",
		}

		botID := plugin.GetBotUserID()

		assert.Equal(t, "bot123", botID)
	})

	t.Run("GetBotUserID returns empty when not set", func(t *testing.T) {
		plugin := &Plugin{}

		botID := plugin.GetBotUserID()

		assert.Empty(t, botID)
	})
}

func TestCreateBotDMPost(t *testing.T) {
	t.Run("creates DM post successfully", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		userID := "user456"
		message := "Test message"
		postType := "custom_type"

		// Mock GetDirectChannel
		dmChannel := &model.Channel{Id: "dm_channel_id"}
		api.On("GetDirectChannel", userID, "bot123").Return(dmChannel, nil)

		// Mock CreatePost
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" &&
				post.ChannelId == "dm_channel_id" &&
				post.Message == message &&
				post.Type == postType
		})).Return(&model.Post{}, nil)

		err := plugin.CreateBotDMPost(userID, message, postType)

		require.NoError(t, err)
	})

	t.Run("returns error when GetDirectChannel fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		userID := "user456"

		// Mock GetDirectChannel failure
		api.On("GetDirectChannel", userID, "bot123").Return(nil, &model.AppError{Message: "channel error"})

		err := plugin.CreateBotDMPost(userID, "message", "type")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel")
	})

	t.Run("returns error when CreatePost fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		userID := "user456"

		// Mock GetDirectChannel success
		dmChannel := &model.Channel{Id: "dm_channel_id"}
		api.On("GetDirectChannel", userID, "bot123").Return(dmChannel, nil)

		// Mock CreatePost failure
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "post error"})

		err := plugin.CreateBotDMPost(userID, "message", "type")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create DM post")
	})
}

func TestSendAlertDMFromBot(t *testing.T) {
	t.Run("sends alert DM successfully", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		userID := "user456"
		alert := &dataminr.Alert{
			AlertID:  "alert-789",
			Headline: "Test Alert",
			AlertType: &dataminr.AlertType{
				Name: "Flash",
			},
		}

		// Mock GetDirectChannel
		dmChannel := &model.Channel{Id: "dm_channel_id"}
		api.On("GetDirectChannel", userID, "bot123").Return(dmChannel, nil)

		// Mock CreatePost
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" &&
				post.ChannelId == "dm_channel_id" &&
				post.Type == "custom_dataminr_alert" &&
				post.Props["from_dataminr"] == true &&
				post.Props["alert_id"] == "alert-789"
		})).Return(&model.Post{}, nil)

		err := plugin.SendAlertDMFromBot(userID, alert)

		require.NoError(t, err)
	})

	t.Run("returns error when GetDirectChannel fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		userID := "user456"
		alert := &dataminr.Alert{
			AlertID:  "alert-789",
			Headline: "Test Alert",
		}

		// Mock GetDirectChannel failure
		api.On("GetDirectChannel", userID, "bot123").Return(nil, &model.AppError{Message: "channel error"})

		err := plugin.SendAlertDMFromBot(userID, alert)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel")
	})

	t.Run("returns error when CreatePost fails", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123"

		userID := "user456"
		alert := &dataminr.Alert{
			AlertID:  "alert-789",
			Headline: "Test Alert",
		}

		// Mock GetDirectChannel success
		dmChannel := &model.Channel{Id: "dm_channel_id"}
		api.On("GetDirectChannel", userID, "bot123").Return(dmChannel, nil)

		// Mock CreatePost failure
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "post error"})

		err := plugin.SendAlertDMFromBot(userID, alert)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create alert DM post")
	})
}

func TestBotConstants(t *testing.T) {
	t.Run("bot username is set correctly", func(t *testing.T) {
		require.Equal(t, "dataminr", BotUsername)
	})

	t.Run("bot display name is set correctly", func(t *testing.T) {
		require.Equal(t, "Dataminr", BotDisplayName)
	})

	t.Run("bot description is set correctly", func(t *testing.T) {
		require.Contains(t, BotDescription, "Dataminr")
	})
}
