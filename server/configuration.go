package main

import (
	"crypto/rand"
	"encoding/base64"
	"reflect"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
//
// As plugins are inherently concurrent (hooks being called asynchronously), and the plugin
// configuration can change at any time, access to the configuration must be synchronized. The
// strategy used in this plugin is to guard a pointer to the configuration, and clone the entire
// struct whenever it changes. You may replace this with whatever strategy you choose.
//
// If you add non-reference types to your configuration struct, be sure to rewrite Clone as a deep
// copy appropriate for your types.
type configuration struct {
	DataminrEnabled                bool
	DataminrEncryptionKey          string
	DataminrSubscriptionPermission string
	DataminrDefaultPollInterval    int // Default polling interval in seconds
	DataminrMinPollInterval        int // Minimum allowed polling interval in seconds
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	clone := *c
	return &clone
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	// Initialize client if needed (OnConfigurationChange can be called before OnActivate)
	if p.client == nil {
		p.client = pluginapi.NewClient(p.API, p.Driver)
	}

	configuration := new(configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	// Auto-generate encryption key if not set
	if configuration.DataminrEncryptionKey == "" {
		key, err := generateEncryptionKey()
		if err != nil {
			return errors.Wrap(err, "failed to generate encryption key")
		}
		configuration.DataminrEncryptionKey = key

		// Save the generated key back to the plugin configuration
		if err := p.saveEncryptionKey(key); err != nil {
			return errors.Wrap(err, "failed to save encryption key")
		}
	}

	p.setConfiguration(configuration)

	// Register the slash command
	if err := p.registerCommand(); err != nil {
		return errors.Wrap(err, "failed to register command")
	}

	return nil
}

// generateEncryptionKey generates a random 32-byte base64-encoded encryption key
func generateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", errors.Wrap(err, "failed to generate random bytes")
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// saveEncryptionKey saves the encryption key to the plugin configuration
func (p *Plugin) saveEncryptionKey(key string) error {
	config := p.API.GetConfig()
	if config.PluginSettings.Plugins == nil {
		config.PluginSettings.Plugins = make(map[string]map[string]any)
	}

	pluginConfig := config.PluginSettings.Plugins["mattermost-plugin-dataminr"]
	if pluginConfig == nil {
		pluginConfig = make(map[string]any)
		config.PluginSettings.Plugins["mattermost-plugin-dataminr"] = pluginConfig
	}

	pluginConfig["dataminrencryptionkey"] = key

	if err := p.client.Configuration.SavePluginConfig(pluginConfig); err != nil {
		return errors.Wrap(err, "failed to save plugin config")
	}

	return nil
}

// registerCommand registers the /dataminr slash command
func (p *Plugin) registerCommand() error {
	return p.client.SlashCommand.Register(&model.Command{
		Trigger:          "dataminr",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage Dataminr First Alert integration",
		AutoCompleteHint: "[command]",
		AutocompleteData: p.getAutocompleteData(),
	})
}
