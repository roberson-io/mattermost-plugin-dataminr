package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAutocompleteData(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	require.NotNil(t, autocomplete)
	assert.Equal(t, "dataminr", autocomplete.Trigger)
	assert.Equal(t, "[command]", autocomplete.Hint)
	assert.Contains(t, autocomplete.HelpText, "Available commands")
}

func TestGetAutocompleteData_HasExpectedSubcommands(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	require.NotNil(t, autocomplete)
	require.NotNil(t, autocomplete.SubCommands)

	// Build a map of subcommand triggers for easy lookup
	subcommands := make(map[string]*model.AutocompleteData)
	for _, cmd := range autocomplete.SubCommands {
		subcommands[cmd.Trigger] = cmd
	}

	// Verify all expected subcommands exist
	expectedSubcommands := []string{
		"connect",
		"disconnect",
		"status",
		"latest",
		"subscribe",
		"unsubscribe",
		"list",
		"dm",
		"filter",
		"help",
	}

	for _, expected := range expectedSubcommands {
		_, exists := subcommands[expected]
		assert.True(t, exists, "Expected subcommand %q to exist", expected)
	}
}

func TestGetAutocompleteData_ConnectCommand(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	// Find connect command
	var connect *model.AutocompleteData
	for _, cmd := range autocomplete.SubCommands {
		if cmd.Trigger == "connect" {
			connect = cmd
			break
		}
	}

	require.NotNil(t, connect, "connect subcommand should exist")
	assert.Contains(t, connect.HelpText, "Connect")

	// Connect should have arguments for client_id and client_secret
	require.NotNil(t, connect.Arguments)
	assert.GreaterOrEqual(t, len(connect.Arguments), 2, "connect should have at least 2 arguments")
}

func TestGetAutocompleteData_SubscribeCommand(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	// Find subscribe command
	var subscribe *model.AutocompleteData
	for _, cmd := range autocomplete.SubCommands {
		if cmd.Trigger == "subscribe" {
			subscribe = cmd
			break
		}
	}

	require.NotNil(t, subscribe, "subscribe subcommand should exist")
	assert.Contains(t, subscribe.HelpText, "Subscribe")
}

func TestGetAutocompleteData_DMCommand(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	// Find dm command
	var dm *model.AutocompleteData
	for _, cmd := range autocomplete.SubCommands {
		if cmd.Trigger == "dm" {
			dm = cmd
			break
		}
	}

	require.NotNil(t, dm, "dm subcommand should exist")

	// DM should have subcommands: on, off
	subcommands := make(map[string]*model.AutocompleteData)
	for _, cmd := range dm.SubCommands {
		subcommands[cmd.Trigger] = cmd
	}

	// Check for on/off subcommands or static list argument
	hasOnOff := false
	if _, exists := subcommands["on"]; exists {
		hasOnOff = true
	}
	if _, exists := subcommands["off"]; exists {
		hasOnOff = true
	}
	// Or it might use a static list argument
	if len(dm.Arguments) > 0 {
		hasOnOff = true
	}

	assert.True(t, hasOnOff, "dm should have on/off options")
}

func TestGetAutocompleteData_FilterCommand(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	// Find filter command
	var filter *model.AutocompleteData
	for _, cmd := range autocomplete.SubCommands {
		if cmd.Trigger == "filter" {
			filter = cmd
			break
		}
	}

	require.NotNil(t, filter, "filter subcommand should exist")

	// Filter should have subcommands: set, clear, show
	subcommands := make(map[string]*model.AutocompleteData)
	for _, cmd := range filter.SubCommands {
		subcommands[cmd.Trigger] = cmd
	}

	expectedFilterSubcommands := []string{"set", "clear", "show"}
	for _, expected := range expectedFilterSubcommands {
		_, exists := subcommands[expected]
		assert.True(t, exists, "filter should have %q subcommand", expected)
	}
}

func TestGetAutocompleteData_SimpleCommands(t *testing.T) {
	p := &Plugin{}

	autocomplete := p.getAutocompleteData()

	// These commands should exist and have descriptions but no arguments
	simpleCommands := []string{"disconnect", "status", "latest", "list", "help"}

	subcommands := make(map[string]*model.AutocompleteData)
	for _, cmd := range autocomplete.SubCommands {
		subcommands[cmd.Trigger] = cmd
	}

	for _, cmdName := range simpleCommands {
		cmd, exists := subcommands[cmdName]
		require.True(t, exists, "%s subcommand should exist", cmdName)
		assert.NotEmpty(t, cmd.HelpText, "%s should have help text", cmdName)
	}
}
