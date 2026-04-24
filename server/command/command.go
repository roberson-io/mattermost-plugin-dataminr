package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

// PluginAPI defines the interface for plugin operations needed by commands
type PluginAPI interface {
	HandleConnect(userID, clientID, clientSecret string) (*model.CommandResponse, error)
	HandleDisconnect(userID string) (*model.CommandResponse, error)
	HandleStatus(userID string) (*model.CommandResponse, error)
	HandleLatest(userID, channelID string, count int) (*model.CommandResponse, error)
	HandleSubscribe(userID, channelID string) (*model.CommandResponse, error)
	HandleUnsubscribe(userID, channelID string) (*model.CommandResponse, error)
	HandleList(userID, channelID string) (*model.CommandResponse, error)
	HandleDM(userID string, enabled bool) (*model.CommandResponse, error)
	HandleFilter(userID string, filter string) (*model.CommandResponse, error)
	HandleChannelInterval(userID, channelID string, interval int) (*model.CommandResponse, error)
	HandleDMInterval(userID string, interval int) (*model.CommandResponse, error)
	HandlePoll(userID, channelID string) (*model.CommandResponse, error)
}

type Handler struct {
	client    *pluginapi.Client
	pluginAPI PluginAPI
}

type Command interface {
	Handle(args *model.CommandArgs) (*model.CommandResponse, error)
}

const dataminrCommandTrigger = "dataminr"

// NewCommandHandler returns a handler for /dataminr commands
// Note: The slash command is registered in OnConfigurationChange, not here
func NewCommandHandler(client *pluginapi.Client, pluginAPI PluginAPI) Command {
	return &Handler{
		client:    client,
		pluginAPI: pluginAPI,
	}
}

// Handle routes the command to the appropriate subcommand handler
func (h *Handler) Handle(args *model.CommandArgs) (*model.CommandResponse, error) {
	trigger, subcommand, cmdArgs := parseCommand(args.Command)

	if trigger != dataminrCommandTrigger {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: %s", args.Command),
		}, nil
	}

	switch subcommand {
	case "":
		return h.executeHelpCommand(), nil
	case "help":
		return h.executeHelpCommand(), nil
	case "connect":
		return h.executeConnectCommand(args, cmdArgs), nil
	case "disconnect":
		return h.executeDisconnectCommand(args), nil
	case "status":
		return h.executeStatusCommand(args), nil
	case "latest":
		return h.executeLatestCommand(args, cmdArgs), nil
	case "subscribe":
		return h.executeSubscribeCommand(args), nil
	case "unsubscribe":
		return h.executeUnsubscribeCommand(args), nil
	case "list":
		return h.executeListCommand(args), nil
	case "dm":
		return h.executeDMCommand(args), nil
	case "filter":
		return h.executeFilterCommand(args), nil
	case "channel-interval":
		return h.executeChannelIntervalCommand(args), nil
	case "dm-interval":
		return h.executeDMIntervalCommand(args), nil
	case "poll":
		return h.executePollCommand(args), nil
	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: %s", subcommand),
		}, nil
	}
}

// parseCommand extracts trigger, subcommand, and args from a command string
func parseCommand(command string) (trigger string, subcommand string, args []string) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "", "", []string{}
	}

	trigger = strings.TrimPrefix(fields[0], "/")
	if len(fields) > 1 {
		subcommand = fields[1]
	}
	if len(fields) > 2 {
		args = fields[2:]
	} else {
		args = []string{}
	}

	return trigger, subcommand, args
}

// executeHelpCommand shows available commands
func (h *Handler) executeHelpCommand() *model.CommandResponse {
	helpText := "### Dataminr First Alert Commands\n\n" +
		"**Available commands:**\n" +
		"* `/dataminr connect <client_id> <client_secret>` - Connect your Dataminr account\n" +
		"* `/dataminr disconnect` - Disconnect your Dataminr account\n" +
		"* `/dataminr status` - Check your connection status\n" +
		"* `/dataminr latest [count]` - Get latest alerts as an ephemeral message (default: 5, max: 100)\n" +
		"* `/dataminr subscribe` - Subscribe this channel to alerts\n" +
		"* `/dataminr unsubscribe` - Unsubscribe this channel from alerts\n" +
		"* `/dataminr list` - List channel subscriptions\n" +
		"* `/dataminr channel-interval <seconds>` - Set polling interval for this channel (0 = manual only)\n" +
		"* `/dataminr dm <on|off>` - Enable or disable DM notifications\n" +
		"* `/dataminr dm-interval <seconds>` - Set polling interval for DM notifications (0 = manual only)\n" +
		"* `/dataminr poll` - Manually poll and post alerts (to channel or DM based on context)\n" +
		"* `/dataminr filter <all|flash|urgent|flash_urgent>` - Set alert type filter for DMs\n" +
		"* `/dataminr help` - Show this help message"

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         helpText,
	}
}

// executeConnectCommand handles the /dataminr connect command
func (h *Handler) executeConnectCommand(args *model.CommandArgs, cmdArgs []string) *model.CommandResponse {
	// Validate arguments
	if len(cmdArgs) < 2 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr connect <client_id> <client_secret>`\n\nConnect your Dataminr account by providing your API credentials.",
		}
	}

	clientID := strings.TrimSpace(cmdArgs[0])
	clientSecret := strings.TrimSpace(cmdArgs[1])

	// Validate non-empty credentials
	if clientID == "" || clientSecret == "" {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Error: `client_id` and `client_secret` cannot be empty.",
		}
	}

	// Delegate to plugin for storage and connection logic
	response, err := h.pluginAPI.HandleConnect(args.UserId, clientID, clientSecret)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}

	return response
}

// executeDisconnectCommand handles the /dataminr disconnect command
func (h *Handler) executeDisconnectCommand(args *model.CommandArgs) *model.CommandResponse {
	response, err := h.pluginAPI.HandleDisconnect(args.UserId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeStatusCommand handles the /dataminr status command
func (h *Handler) executeStatusCommand(args *model.CommandArgs) *model.CommandResponse {
	response, err := h.pluginAPI.HandleStatus(args.UserId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeLatestCommand handles the /dataminr latest command
func (h *Handler) executeLatestCommand(args *model.CommandArgs, cmdArgs []string) *model.CommandResponse {
	const defaultCount = 5
	const maxCount = 100

	count := defaultCount

	// Parse optional count argument
	if len(cmdArgs) > 0 {
		parsedCount, err := strconv.Atoi(cmdArgs[0])
		if err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "❌ Count must be a valid number. Usage: `/dataminr latest [count]`",
			}
		}
		if parsedCount < 1 || parsedCount > maxCount {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("❌ Count must be between 1 and %d.", maxCount),
			}
		}
		count = parsedCount
	}

	response, err := h.pluginAPI.HandleLatest(args.UserId, args.ChannelId, count)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeSubscribeCommand handles the /dataminr subscribe command
func (h *Handler) executeSubscribeCommand(args *model.CommandArgs) *model.CommandResponse {
	response, err := h.pluginAPI.HandleSubscribe(args.UserId, args.ChannelId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeUnsubscribeCommand handles the /dataminr unsubscribe command
func (h *Handler) executeUnsubscribeCommand(args *model.CommandArgs) *model.CommandResponse {
	response, err := h.pluginAPI.HandleUnsubscribe(args.UserId, args.ChannelId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeListCommand handles the /dataminr list command
func (h *Handler) executeListCommand(args *model.CommandArgs) *model.CommandResponse {
	response, err := h.pluginAPI.HandleList(args.UserId, args.ChannelId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeDMCommand handles the /dataminr dm command
func (h *Handler) executeDMCommand(args *model.CommandArgs) *model.CommandResponse {
	_, _, cmdArgs := parseCommand(args.Command)

	if len(cmdArgs) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr dm <on|off>`\n\nEnable or disable DM notifications for alerts.",
		}
	}

	arg := strings.ToLower(strings.TrimSpace(cmdArgs[0]))

	var enabled bool
	switch arg {
	case "on":
		enabled = true
	case "off":
		enabled = false
	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr dm <on|off>`\n\nPlease specify `on` or `off`.",
		}
	}

	response, err := h.pluginAPI.HandleDM(args.UserId, enabled)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeChannelIntervalCommand handles the /dataminr channel-interval command
func (h *Handler) executeChannelIntervalCommand(args *model.CommandArgs) *model.CommandResponse {
	_, _, cmdArgs := parseCommand(args.Command)

	if len(cmdArgs) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr channel-interval <seconds>`\n\nSet the polling interval for your subscription in this channel.\nUse `0` to disable automatic polling (manual only mode).",
		}
	}

	intervalStr := strings.TrimSpace(cmdArgs[0])
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Interval must be a valid number. Usage: `/dataminr channel-interval <seconds>`",
		}
	}

	if interval < 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Interval cannot be negative. Use `0` to disable automatic polling.",
		}
	}

	response, err := h.pluginAPI.HandleChannelInterval(args.UserId, args.ChannelId, interval)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeDMIntervalCommand handles the /dataminr dm-interval command
func (h *Handler) executeDMIntervalCommand(args *model.CommandArgs) *model.CommandResponse {
	_, _, cmdArgs := parseCommand(args.Command)

	if len(cmdArgs) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr dm-interval <seconds>`\n\nSet the polling interval for DM notifications.\nUse `0` to disable automatic polling (manual only mode).",
		}
	}

	intervalStr := strings.TrimSpace(cmdArgs[0])
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Interval must be a valid number. Usage: `/dataminr dm-interval <seconds>`",
		}
	}

	if interval < 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Interval cannot be negative. Use `0` to disable automatic polling.",
		}
	}

	response, err := h.pluginAPI.HandleDMInterval(args.UserId, interval)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executePollCommand handles the /dataminr poll command
func (h *Handler) executePollCommand(args *model.CommandArgs) *model.CommandResponse {
	response, err := h.pluginAPI.HandlePoll(args.UserId, args.ChannelId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}

// executeFilterCommand handles the /dataminr filter command
func (h *Handler) executeFilterCommand(args *model.CommandArgs) *model.CommandResponse {
	_, _, cmdArgs := parseCommand(args.Command)

	if len(cmdArgs) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr filter <all|flash|urgent|flash_urgent>`\n\nSet which alert types you want to receive via DM.",
		}
	}

	filter := strings.ToLower(strings.TrimSpace(cmdArgs[0]))

	// Validate filter value
	validFilters := map[string]bool{
		"all":          true,
		"flash":        true,
		"urgent":       true,
		"flash_urgent": true,
	}

	if !validFilters[filter] {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Usage:** `/dataminr filter <all|flash|urgent|flash_urgent>`\n\nValid filters: `all`, `flash`, `urgent`, `flash_urgent`.",
		}
	}

	response, err := h.pluginAPI.HandleFilter(args.UserId, filter)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ An error occurred. Please try again.",
		}
	}
	return response
}
