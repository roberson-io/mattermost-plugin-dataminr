package main

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// getAutocompleteData returns the autocomplete data for the /dataminr command
func (p *Plugin) getAutocompleteData() *model.AutocompleteData {
	dataminr := model.NewAutocompleteData("dataminr", "[command]",
		"Available commands: connect, disconnect, status, latest, subscribe, unsubscribe, list, channel-interval, dm, dm-interval, poll, filter, help")

	// connect command
	connect := model.NewAutocompleteData("connect", "<client_id> <client_secret>",
		"Connect your Mattermost account to Dataminr")
	connect.AddTextArgument("Dataminr Client ID", "<client_id>", "")
	connect.AddTextArgument("Dataminr Client Secret", "<client_secret>", "")
	dataminr.AddCommand(connect)

	// disconnect command
	disconnect := model.NewAutocompleteData("disconnect", "",
		"Disconnect your Mattermost account from Dataminr")
	dataminr.AddCommand(disconnect)

	// status command
	status := model.NewAutocompleteData("status", "",
		"Check your Dataminr connection status")
	dataminr.AddCommand(status)

	// latest command
	latest := model.NewAutocompleteData("latest", "[count]",
		"Get the latest Dataminr alerts as an ephemeral message (default: 5, max: 100)")
	latest.AddTextArgument("Number of alerts to fetch (1-100)", "[count]", "")
	dataminr.AddCommand(latest)

	// subscribe command
	subscribe := model.NewAutocompleteData("subscribe", "",
		"Subscribe this channel to receive Dataminr alerts")
	dataminr.AddCommand(subscribe)

	// unsubscribe command
	unsubscribe := model.NewAutocompleteData("unsubscribe", "",
		"Unsubscribe this channel from Dataminr alerts")
	dataminr.AddCommand(unsubscribe)

	// list command
	list := model.NewAutocompleteData("list", "",
		"List all channel subscriptions")
	dataminr.AddCommand(list)

	// channel-interval command
	channelInterval := model.NewAutocompleteData("channel-interval", "<seconds>",
		"Set polling interval for this channel's subscription (0 = manual only)")
	channelInterval.AddTextArgument("Polling interval in seconds (0 for manual only)", "<seconds>", "")
	dataminr.AddCommand(channelInterval)

	// dm command with on/off subcommands
	dm := model.NewAutocompleteData("dm", "<on|off>",
		"Enable or disable direct message alerts")
	dmOn := model.NewAutocompleteData("on", "", "Enable direct message alerts")
	dmOff := model.NewAutocompleteData("off", "", "Disable direct message alerts")
	dm.AddCommand(dmOn)
	dm.AddCommand(dmOff)
	dataminr.AddCommand(dm)

	// dm-interval command
	dmInterval := model.NewAutocompleteData("dm-interval", "<seconds>",
		"Set polling interval for DM notifications (0 = manual only)")
	dmInterval.AddTextArgument("Polling interval in seconds (0 for manual only)", "<seconds>", "")
	dataminr.AddCommand(dmInterval)

	// poll command
	poll := model.NewAutocompleteData("poll", "",
		"Manually poll and post alerts (to channel if subscribed, otherwise to DM)")
	dataminr.AddCommand(poll)

	// filter command with set/clear/show subcommands
	filter := model.NewAutocompleteData("filter", "[set|clear|show]",
		"Manage alert filters for this channel")
	filterSet := model.NewAutocompleteData("set", "<filter_expression>",
		"Set a filter for alerts in this channel")
	filterSet.AddTextArgument("Filter expression (e.g., type:Flash or topic:Weather)", "<filter>", "")
	filterClear := model.NewAutocompleteData("clear", "",
		"Clear the current filter for this channel")
	filterShow := model.NewAutocompleteData("show", "",
		"Show the current filter for this channel")
	filter.AddCommand(filterSet)
	filter.AddCommand(filterClear)
	filter.AddCommand(filterShow)
	dataminr.AddCommand(filter)

	// help command
	help := model.NewAutocompleteData("help", "",
		"Display help information for Dataminr commands")
	dataminr.AddCommand(help)

	return dataminr
}
