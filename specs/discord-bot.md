# Core Bot & Webhook System

## Overview

The core Discord bot handles two concerns: GM tooling (log access, server status, admin commands) and player engagement (in-universe notifications via themed webhook identities, per-player private channels).

This covers implementation phases 1-3. Voice, transcription, and AI features are in separate specs.

## Standalone Bot Process

- Runs on the droplet alongside hx-web, managed by pm2
- Connects to Discord via Discord.js (WebSocket gateway)
- Queries Turso directly — works even when hx-web is down
- Handles slash commands and interactive responses

## Webhook Identities

Webhooks post to server channels with custom names and avatars, giving the feel of multiple in-world entities. Each identity has a distinct voice and purpose.

| Identity | Avatar | Purpose |
|---|---|---|
| FWDC Postal Service | Wax seal / mail bag | Mail delivery to players |
| The Verdaal Courant | Newspaper masthead | Broadsheet publication announcements |
| Company Clerk | Quill & ledger | Character changes, roster updates |
| Board of Directors | Company seal | System announcements, session reminders |

## Per-Player Private Channels

- A private channel per player (e.g., `#mail-alice`, `#mail-bob`)
- Only the player and the bot can see the channel
- Webhooks post here for personal correspondence
- Players can respond in-thread
- Created automatically when a player joins / is invited

## Bot Commands

### GM Commands (private GM channel)

| Command | Description |
|---|---|
| `/logs` | Show recent error logs (last 25) |
| `/logs search <term>` | Search logs by keyword |
| `/logs tail` | Stream recent logs |
| `/status` | Check if sites are up, pm2 status |
| `/who` | List players with active sessions |
| `/mail <player> <message>` | Quick-send correspondence |
| `/announce <message>` | Post as Board of Directors to campaign channel |

### Player Commands (their private channel or public)

| Command | Description |
|---|---|
| `/sheet` | Quick summary of active character |
| `/mail` | Check unread correspondence count |
| `/bone <question>` | AI-powered memory recall scoped to your character (see [campaign-recall.md](campaign-recall.md)) |
| `/bone search <term>` | Keyword search across session transcripts and notes |
| `/private <text>` | Add a private note to active character (only player + GM can see) |
| `/public <text>` | Add a public note to active character (visible to all players) |
| `/reply <text>` | Reply to most recent correspondence in-character |

## App-Side Event Integration

hx-web fires events to Discord when key actions occur. This could be:
- Direct webhook calls from SvelteKit API routes
- Or writes to a Turso `events` queue that the bot polls

### Events to fire

| Event | Webhook identity | Target channel |
|---|---|---|
| Mail sent to a player | Postal Service | Player's private channel |
| Broadsheet published | The Verdaal Courant | Public campaign channel |
| Character created/updated | Company Clerk | GM private channel |
| Character marked deceased | Company Clerk | Campaign channel |

## Log Storage (Turso)

### Schema

```sql
CREATE TABLE logs (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
  campaign_id TEXT,
  level TEXT NOT NULL CHECK(level IN ('error', 'warn', 'info')),
  source TEXT NOT NULL DEFAULT 'app',
  message TEXT NOT NULL,
  metadata TEXT,
  created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_logs_created ON logs(created_at);
CREATE INDEX idx_logs_level ON logs(level, created_at);
```

### Retention

- 5-day retention
- Cleanup on write (every 100th write, delete rows older than 5 days)
- Or a simple cron via the bot: daily purge

### What Gets Logged

- Unhandled exceptions (from SvelteKit hooks)
- 500 responses
- App start/stop events
- Auth failures (repeated)
- Key user actions (character created, mail sent) at info level

### What Doesn't Get Logged

- Normal requests
- 4xx client errors (except repeated auth failures)
- Static asset serving

## Per-Character Notes

Players can take notes scoped to their active character, with visibility levels:

| Visibility | Who sees it | Use case |
|---|---|---|
| **Private** | Player + GM | Suspicions, plans, personal observations |
| **Public** | All players | Shared discoveries, announcements, in-character statements |
| **GM-only** | GM only | Meta notes, rules questions, out-of-character flags |

- Notes are tied to the active character, not the player — switching characters switches context
- Notes sync to the character sheet's existing notes section on the portal
- `/private` and `/public` commands from Discord, or editable on the web
- Timestamped and tagged with session if taken during a session
- Searchable alongside session transcripts via `/bone search`

## Correspondence Replies

Players can reply to in-character mail directly from Discord:

- `/reply` responds to the most recent letter in their private channel
- Reply is stored as a mail item in hx-web, attributed to their character
- GM receives the reply in their channel (posted by Company Clerk webhook)
- Threaded Discord replies on a mail message could also trigger this

## Discord Permissions

Bot needs: `SEND_MESSAGES`, `MANAGE_CHANNELS` (for creating player channels), `MANAGE_WEBHOOKS`, `USE_SLASH_COMMANDS`, `CONNECT` (for voice features), `SPEAK` (for future TTS)

## User Mapping

Discord users are mapped to hx-web players via the `discord_id` field in the users table. Discord OAuth is already supported in hx-web.

## Limitations

- **Not captured:** nginx errors, OS-level issues, pm2 crashes. These still require SSH.
- **Webhook rate limit:** 30 messages/minute per webhook. Fine for notifications, not for log streaming.
- **DMs:** Bot can only DM as itself, not as webhook identities. Per-player private channels solve this.
- **Bot downtime:** If the bot process is down, commands don't work. Notifications queue in Turso until it restarts.

## Implementation Phases

### Phase 1 — Logging & GM tooling
- Turso logs table
- SvelteKit error hook writing to logs
- Bot with `/logs` and `/status` commands
- GM-only private channel

### Phase 2 — Player notifications
- Per-player private channels
- Webhook identities (Postal Service, Courant)
- Mail delivery notifications
- Broadsheet announcements

### Phase 3 — Interactive features
- Player `/sheet` command
- Threaded replies to mail
- GM `/mail` quick-send
- Character change notifications

## Open Questions

- **Event delivery mechanism:** Direct webhook calls from hx-web API routes vs. Turso events queue that the bot polls? Queue is more resilient (survives bot restarts) but adds latency.
- **Public campaign channel:** Do we want one for broadsheets/announcements, or keep everything in private channels?
- **Same Turso database or separate for logs?** Isolation vs. convenience.
