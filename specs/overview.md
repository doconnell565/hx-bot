# hx-bot — Project Overview

## What This Is

hx-bot is the Discord bot for the Forbidden West Discovery Company campaign. It runs as a standalone process alongside hx-web on a shared DigitalOcean droplet, connecting the campaign's web portal, voice sessions, and AI systems through Discord.

It serves four roles:

1. **GM tooling** — log access, server status, admin commands without SSH
2. **Player engagement** — in-universe notifications and correspondence via themed webhook identities
3. **Voice intelligence** — real-time transcription, reference surfacing, and session note generation
4. **Campaign recall** — AI-powered memory search scoped per character (the Bone)

## Architecture

```
┌─── Discord ──────────────────────────────────────┐
│  Voice Channels          Text Channels            │
│  ┌──────────┐           ┌───────────────────┐    │
│  │ Per-user  │           │ #gm-private       │    │
│  │ audio     │           │ #mail-<player>    │    │
│  │ streams   │           │ #campaign         │    │
│  └─────┬─────┘           └────────┬──────────┘    │
└────────┼──────────────────────────┼───────────────┘
         │                          │
         ▼                          ▼
┌─── $12/mo DO Droplet (2GB) ──────────────────────┐
│                                                    │
│  hx-bot (pm2)                                      │
│  ├── Discord.js gateway (commands, events)          │
│  ├── Webhook manager (themed identities)            │
│  ├── ChromaDB (campaign knowledge base)             │
│  ├── Reference sidebar API (future)                 │
│  └── Voice listener → AssemblyAI STT (future)       │
│                                                    │
│  hx-web (pm2)                                      │
│  └── SvelteKit app (portal, API)                    │
│                                                    │
└──────────┬────────────────────────┬───────────────┘
           │                        │
           ▼                        ▼
    ┌──────────┐           ┌────────────────┐
    │  Turso   │           │  External APIs  │
    │  (DB)    │           │  - Gemini Flash  │
    │          │           │  - AssemblyAI   │
    └──────────┘           └────────────────┘
           │
           ▼
    ┌──────────────────┐
    │  Foundry VTT     │
    │  (on The Forge)  │
    │  └── Module calls│
    │     sidebar API  │
    └──────────────────┘
```

## Technology Choices

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Runtime | Node.js (Discord.js) | Standard for Discord bots, same runtime as hx-web |
| Database | Turso (shared with hx-web) | Already in use, SQLite-compatible, works when app is down |
| LLM | Gemini Flash / Flash-Lite | Free tier covers most usage, cheapest paid tier for overages |
| STT | AssemblyAI Universal-3 (streaming) | Best accuracy on fantasy nouns in testing, 1,000 word keyterm budget, cheaper than alternatives |
| Vector DB | ChromaDB | File-based, minimal RAM, Python-native but has JS client |
| Hosting | Same DO droplet as hx-web | No additional infra cost, pm2-managed |
| Process manager | pm2 | Already managing hx-web |

## Spec Documents

| Document | Covers |
|----------|--------|
| [discord-bot.md](discord-bot.md) | Core bot commands, webhook identities, channels, logging, event integration |
| [voice-transcription.md](voice-transcription.md) | Discord voice → Deepgram STT pipeline, custom vocabulary, speaker attribution |
| [realtime-sidebar.md](realtime-sidebar.md) | Live reference panel during sessions — entity extraction, ChromaDB, UX |
| [session-notes.md](session-notes.md) | Post-session note generation — attribution, summaries, wiki updates |
| [campaign-recall.md](campaign-recall.md) | The Bone — AI-powered per-character memory recall |
| [tablet-voice.md](tablet-voice.md) | Stretch goal — TTS via Lompen in Foundry VTT |
| [infrastructure.md](infrastructure.md) | Consolidated costs, hosting, API budgets |

## Relationship to hx-web

hx-bot and hx-web share the Turso database and run on the same droplet, but are separate processes. The bot is designed to function when hx-web is down (direct DB access for log queries, status checks).

**Bot → hx-web:** API calls for campaign data (characters, mail, findings, broadsheets). The bot is a client of hx-web's API.

**hx-web → Bot:** Event notifications. When hx-web actions occur (mail sent, broadsheet published, character updated), the bot is notified and posts to Discord via webhooks. This could be direct webhook calls from SvelteKit routes, or writes to a Turso events queue that the bot polls.

## Implementation Phases

| Phase | Scope | Depends on |
|-------|-------|------------|
| 1 | Logging & GM tooling (commands, log queries, status) | Nothing |
| 2 | Player notifications (webhooks, private channels, mail delivery) | Phase 1 |
| 3 | Interactive features (player commands, threaded replies, `/sheet`) | Phase 2 |
| 4 | Session notes (local: Craig → AssemblyAI → Gemini → notes) | Nothing (runs locally) |
| 5 | Campaign recall — the Bone (`/bone` command, RAG, per-character scoping) | Phase 4 |
| 6 | Live transcription (bot joins voice, streams to AssemblyAI) | Phase 1 |
| 7 | Real-time sidebar (entity extraction, ChromaDB, sidebar API) | Phase 6 |
| 8 | Tablet voice via TTS (stretch goal, mid-to-late campaign) | Phase 5 |

Phases 1-3 (core bot) and Phase 4 (session notes) can be developed in parallel. Phase 4 doesn't require the bot at all — it's a local script pipeline.

## Open Questions

- **Same Turso database or separate?** Sharing is simpler and allows cross-queries, but isolation protects against bot bugs corrupting app data.
- **ChromaDB JS client vs Python sidecar?** ChromaDB is Python-native. The JS client exists but is less mature. Could run a thin Python service for vector operations, or use an alternative like LanceDB.
- **Bot framework?** Plain Discord.js or something like Sapphire for command handling? Depends on complexity once all commands are mapped out.
