package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPluginAPI implements PluginAPI for testing
type mockPluginAPI struct {
	handleConnectResponse     *model.CommandResponse
	handleConnectError        error
	handleDisconnectResponse  *model.CommandResponse
	handleDisconnectError     error
	handleStatusResponse      *model.CommandResponse
	handleStatusError         error
	handleLatestResponse      *model.CommandResponse
	handleLatestError         error
	handleLatestFunc          func(userID, channelID string, count int) (*model.CommandResponse, error)
	handleSubscribeFunc       func(userID, channelID string) (*model.CommandResponse, error)
	handleUnsubscribeFunc     func(userID, channelID string) (*model.CommandResponse, error)
	handleListFunc            func(userID, channelID string) (*model.CommandResponse, error)
	handleDMFunc              func(userID string, enabled bool) (*model.CommandResponse, error)
	handleFilterFunc          func(userID string, filter string) (*model.CommandResponse, error)
	handleChannelIntervalFunc func(userID, channelID string, interval int) (*model.CommandResponse, error)
	handleDMIntervalFunc      func(userID string, interval int) (*model.CommandResponse, error)
	handlePollFunc            func(userID, channelID string) (*model.CommandResponse, error)
}

func (m *mockPluginAPI) HandleConnect(userID, clientID, clientSecret string) (*model.CommandResponse, error) {
	return m.handleConnectResponse, m.handleConnectError
}

func (m *mockPluginAPI) HandleDisconnect(userID string) (*model.CommandResponse, error) {
	return m.handleDisconnectResponse, m.handleDisconnectError
}

func (m *mockPluginAPI) HandleStatus(userID string) (*model.CommandResponse, error) {
	return m.handleStatusResponse, m.handleStatusError
}

func (m *mockPluginAPI) HandleLatest(userID, channelID string, count int) (*model.CommandResponse, error) {
	if m.handleLatestFunc != nil {
		return m.handleLatestFunc(userID, channelID, count)
	}
	return m.handleLatestResponse, m.handleLatestError
}

func (m *mockPluginAPI) HandleSubscribe(userID, channelID string) (*model.CommandResponse, error) {
	if m.handleSubscribeFunc != nil {
		return m.handleSubscribeFunc(userID, channelID)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandleUnsubscribe(userID, channelID string) (*model.CommandResponse, error) {
	if m.handleUnsubscribeFunc != nil {
		return m.handleUnsubscribeFunc(userID, channelID)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandleList(userID, channelID string) (*model.CommandResponse, error) {
	if m.handleListFunc != nil {
		return m.handleListFunc(userID, channelID)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandleDM(userID string, enabled bool) (*model.CommandResponse, error) {
	if m.handleDMFunc != nil {
		return m.handleDMFunc(userID, enabled)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandleFilter(userID string, filter string) (*model.CommandResponse, error) {
	if m.handleFilterFunc != nil {
		return m.handleFilterFunc(userID, filter)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandleChannelInterval(userID, channelID string, interval int) (*model.CommandResponse, error) {
	if m.handleChannelIntervalFunc != nil {
		return m.handleChannelIntervalFunc(userID, channelID, interval)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandleDMInterval(userID string, interval int) (*model.CommandResponse, error) {
	if m.handleDMIntervalFunc != nil {
		return m.handleDMIntervalFunc(userID, interval)
	}
	return nil, nil
}

func (m *mockPluginAPI) HandlePoll(userID, channelID string) (*model.CommandResponse, error) {
	if m.handlePollFunc != nil {
		return m.handlePollFunc(userID, channelID)
	}
	return nil, nil
}

type env struct {
	client    *pluginapi.Client
	api       *plugintest.API
	pluginAPI *mockPluginAPI
}

func setupTest() *env {
	api := &plugintest.API{}
	driver := &plugintest.Driver{}
	client := pluginapi.NewClient(api, driver)

	return &env{
		client:    client,
		api:       api,
		pluginAPI: &mockPluginAPI{},
	}
}

func TestNewCommandHandler(t *testing.T) {
	env := setupTest()

	// NewCommandHandler no longer registers the command (that's done in OnConfigurationChange)
	// It just returns a handler
	cmdHandler := NewCommandHandler(env.client, env.pluginAPI)
	require.NotNil(t, cmdHandler)
}

func TestHandle_NoSubcommand(t *testing.T) {
	env := setupTest()
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Available commands")
}

func TestHandle_HelpSubcommand(t *testing.T) {
	env := setupTest()
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr help",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Available commands")
}

func TestHandle_UnknownSubcommand(t *testing.T) {
	env := setupTest()
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr unknown",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Unknown command")
}

func TestHandle_ConnectSubcommand_MissingArgs(t *testing.T) {
	env := setupTest()
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr connect",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Usage")
}

func TestHandle_ConnectSubcommand_Success(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleConnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully connected to Dataminr!",
	}
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr connect test-client-id test-client-secret",
		UserId:  "user123",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Successfully connected")
}

func TestHandle_DisconnectSubcommand(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleDisconnectResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ Successfully disconnected from Dataminr.",
	}
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr disconnect",
		UserId:  "user123",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Successfully disconnected")
}

func TestHandle_StatusSubcommand(t *testing.T) {
	env := setupTest()
	env.pluginAPI.handleStatusResponse = &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "✅ **Connected to Dataminr**",
	}
	cmdHandler := &Handler{client: env.client, pluginAPI: env.pluginAPI}

	args := &model.CommandArgs{
		Command: "/dataminr status",
		UserId:  "user123",
	}

	response, err := cmdHandler.Handle(args)
	require.NoError(t, err)
	assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
	assert.Contains(t, response.Text, "Connected to Dataminr")
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		expectedTrigger string
		expectedSubcmd  string
		expectedArgs    []string
	}{
		{
			name:            "no subcommand",
			command:         "/dataminr",
			expectedTrigger: "dataminr",
			expectedSubcmd:  "",
			expectedArgs:    []string{},
		},
		{
			name:            "help subcommand",
			command:         "/dataminr help",
			expectedTrigger: "dataminr",
			expectedSubcmd:  "help",
			expectedArgs:    []string{},
		},
		{
			name:            "connect with args",
			command:         "/dataminr connect client_id client_secret",
			expectedTrigger: "dataminr",
			expectedSubcmd:  "connect",
			expectedArgs:    []string{"client_id", "client_secret"},
		},
		{
			name:            "extra whitespace",
			command:         "/dataminr   status  ",
			expectedTrigger: "dataminr",
			expectedSubcmd:  "status",
			expectedArgs:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger, subcommand, args := parseCommand(tt.command)
			assert.Equal(t, tt.expectedTrigger, trigger)
			assert.Equal(t, tt.expectedSubcmd, subcommand)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}
