# Notification Service - Tuskira

Notification service supporting email, slack, and in-app channels with scheduling.

## Dev Setup

Requires Docker.

```
docker compose up
```

App runs on `http://localhost:8080` with hot reload. Postgres starts automatically.

## Notification Channels

All channels are configured via `PUT /api/v1/channels` (requires JWT auth).

### Email

Sends via SMTP.

```json
{
  "channel": "email",
  "enabled": true,
  "config": {
    "host": "smtp.example.com",
    "port": "587",
    "username": "user@example.com",
    "password": "password",
    "from": "noreply@example.com",
    "tls": true
  }
}
```

Required: `host`, `from`.

### Slack

Sends messages using a bot token.

```json
{
  "channel": "slack",
  "enabled": true,
  "config": {
    "bot_token": "xoxb-your-bot-token"
  }
}
```

Required: `bot_token`. Set `recipient` to the Slack channel ID when sending.

**Getting a bot token:**

1. Go to https://api.slack.com/apps and click **Create New App** → **From scratch**
2. Give it a name and select your workspace
3. Go to **OAuth & Permissions** in the sidebar
4. Under **Bot Token Scopes**, add at minimum:
   - `chat:write` — to post messages
   - `chat:write.public` — to post to channels the bot hasn't joined (optional)
5. Click **Install to Workspace** at the top and authorize
6. Copy the **Bot User OAuth Token** (`xoxb-...`) from the OAuth & Permissions page

### In-App (SSE)

Delivers notifications in real time via Server-Sent Events. A `connection_id` is auto-generated when you configure the channel.

```json
{
  "channel": "inapp",
  "enabled": true,
  "config": {}
}
```

Connect to the stream:

```
GET /api/v1/notifications/stream?connection_id=<connection_id>
```

Get the `connection_id` from `GET /api/v1/channels/inapp`. Missed notifications are replayed on reconnect.
