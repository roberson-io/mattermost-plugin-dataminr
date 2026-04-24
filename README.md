# Dataminr First Alert Plugin for Mattermost

Receive real-time [Dataminr First Alert](https://www.dataminr.com/) notifications directly in Mattermost channels and direct messages.

## Features

- **Real-time alerts**: Automatically poll Dataminr for new alerts and deliver them to Mattermost
- **Channel subscriptions**: Subscribe channels to receive alerts from specific Dataminr accounts
- **DM notifications**: Receive personal alerts via direct message from the Dataminr bot
- **Configurable polling**: Set custom poll intervals per channel or for DM notifications
- **Alert filtering**: Filter alerts by type (Flash, Urgent, or All)
- **Secure credential storage**: User credentials are encrypted with AES-256

## Requirements

- Mattermost Server 7.0.0+
- Dataminr First Alert API credentials (Client ID and Client Secret)

## Installation

1. Download the latest release from the [releases page](https://github.com/roberson-io/mattermost-plugin-dataminr/releases)
2. Upload the plugin to your Mattermost server via **System Console > Plugins > Plugin Management**
3. Enable the plugin

## Configuration

Configure the plugin in **System Console > Plugins > Dataminr First Alert**:

| Setting | Description | Default |
|---------|-------------|---------|
| Enable Dataminr Integration | Enable/disable the plugin | Enabled |
| Default Poll Interval | Seconds between polling for new alerts | 120 |
| Minimum Poll Interval | Minimum allowed poll interval (rate limit protection) | 30 |
| Subscription Permission | Who can subscribe channels (anyone, channel_admin, system_admin) | channel_admin |

## Usage

### Slash Commands

All commands use the `/dataminr` prefix:

#### Account Management
| Command | Description |
|---------|-------------|
| `/dataminr connect <client_id> <client_secret>` | Connect your Dataminr account |
| `/dataminr disconnect` | Disconnect your Dataminr account |
| `/dataminr status` | Check your connection status |

#### Channel Subscriptions
| Command | Description |
|---------|-------------|
| `/dataminr subscribe` | Subscribe this channel to your Dataminr alerts |
| `/dataminr unsubscribe` | Unsubscribe this channel from your alerts |
| `/dataminr list` | List all subscriptions in this channel |

#### DM Notifications
| Command | Description |
|---------|-------------|
| `/dataminr dm on` | Enable DM notifications |
| `/dataminr dm off` | Disable DM notifications |
| `/dataminr filter <all\|flash\|urgent\|flash_urgent>` | Set alert type filter for DMs |

#### Polling Configuration
| Command | Description |
|---------|-------------|
| `/dataminr channel-interval <seconds>` | Set poll interval for this channel (0 = manual only) |
| `/dataminr dm-interval <seconds>` | Set poll interval for DM notifications (0 = manual only) |

#### Manual Operations
| Command | Description |
|---------|-------------|
| `/dataminr latest [count]` | Fetch the latest alerts (default: 5) |
| `/dataminr poll` | Manually trigger alert polling |

### Quick Start

1. Connect your Dataminr account:
   ```
   /dataminr connect your-client-id your-client-secret
   ```

2. Subscribe a channel to receive alerts:
   ```
   /dataminr subscribe
   ```

3. Or enable DM notifications:
   ```
   /dataminr dm on
   ```

## Alert Format

Alerts are displayed with:
- Color-coded emoji by severity (🔴 Flash, 🟠 Urgent, 🟡 Alert)
- Headline and timestamp
- Location information (name, coordinates, MGRS)
- Topics and keywords
- AI-generated summary (when available)
- Links to Dataminr and source posts

## Development

### Building

```bash
make dist
```

### Running Tests

```bash
make test
```

### Linting

```bash
make check-style
```

### Deploying Locally

```bash
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=your-admin-token
make deploy
```

## License

This project is licensed under the Apache 2.0 License.
