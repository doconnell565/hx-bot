# Infrastructure & Cost Summary

## Hosting

### DigitalOcean Droplet — $12/mo

2GB RAM droplet running:
- **hx-web** — SvelteKit app (player portal)
- **hx-bot** — Discord bot (all features in these specs)
- **ChromaDB** — file-based vector store for campaign knowledge base
- **pm2** — process manager for both services

The bot adds minimal resource overhead to the existing droplet:
- ChromaDB with campaign-sized data: ~100-200MB RAM
- Discord.js gateway: ~50-100MB RAM
- Deepgram streaming connections: negligible (HTTP/WebSocket)
- Gemini API calls: negligible (outbound HTTP)

Total bot footprint: ~200-300MB RAM, fitting within the 2GB droplet alongside hx-web.

### Foundry VTT on The Forge

Existing cost (separate from hx-bot). The bot communicates with the Foundry module via HTTPS. The Forge is managed hosting — no custom server processes, module can only make external HTTP calls.

The bot's sidebar API must be HTTPS-accessible from The Forge. Options:
- Cloudflare Tunnel (free, recommended)
- Tailscale Funnel (free)
- Direct HTTPS on the droplet (requires cert management)

### Local GPU — RTX 5060 Ti (16GB VRAM)

Available as a fallback but **not needed** given API pricing. Could run:
- Llama 3.1 8B Q8 (~9GB VRAM) for local LLM inference
- nomic-embed-text for local embeddings
- ChromaDB locally

Useful if API costs become a concern or for offline development/testing. Not part of the production architecture.

## API Costs

### Gemini (Google AI Studio)

| Use case | Model | Monthly estimate | Notes |
|----------|-------|-----------------|-------|
| Real-time entity extraction | Flash-Lite | ~$0.50 | 15s intervals, 4hr sessions |
| Session note generation | Flash | ~$0.62 | 4 sessions, 4 passes each |
| Campaign recall (/bone) | Flash | ~$0.20 | ~100 queries/session |
| **Total Gemini** | | **~$1.32/mo** | |

**Free tier:** 250 requests/day, 10 requests/min, 250K tokens/min. At current estimates, the free tier covers all usage. Paid tier is the fallback.

**Spending cap:** Set a project-level spend cap of $5-10/month in Google AI Studio. Launched March 16, 2026. Blocks API requests when cap is reached (~10 minute enforcement delay). Safety net against runaway costs from bugs.

### AssemblyAI

| Use case | Rate | Monthly estimate |
|----------|------|-----------------|
| Live session transcription | ~$0.0065/min streaming | ~$6.24 (4x 4hr sessions) |

Chosen over Deepgram based on head-to-head testing with campaign audio. Superior accuracy on fantasy proper nouns, 2x keyterm budget (1,000 words vs 500 tokens), and cheaper per minute. See [voice-transcription.md](voice-transcription.md) for full test results.

### Craig Bot

| Tier | Cost | Features |
|------|------|----------|
| Free | $0 | Manual `/join`, 6hr max, 7-day file storage |
| Premium | $4/mo | Auto-record, 14-day storage |

Start with free tier (manual `/join`). Consider Premium if auto-recording becomes important.

## Total Monthly Costs

| Component | Cost | Status |
|-----------|------|--------|
| DO droplet (existing) | $12.00 | Already paying |
| Foundry/Forge (existing) | varies | Already paying |
| AssemblyAI streaming | ~$6.24 | New |
| Gemini API | ~$1.32 | New (likely free tier) |
| Craig Premium (optional) | $0-4.00 | Optional |
| **New costs** | **~$7.56-11.56** | |

With AssemblyAI's $49.98 free credits (~8 months) and Gemini's free tier, actual out-of-pocket new cost could be **$0-4/mo** for the first 8 months.

## Scaling Considerations

The current architecture is designed for a single campaign with 4-6 players and weekly 4-hour sessions. At this scale:

- API costs are negligible
- 2GB droplet is sufficient
- ChromaDB handles the data volume easily
- Gemini free tier covers the request volume

If usage grows (multiple campaigns, longer sessions, more frequent queries), the first bottleneck would be Deepgram cost and the droplet's RAM. Upgrading to a 4GB droplet ($24/mo) and managing Deepgram costs would handle significant growth.

## Secrets Management

Required API keys and tokens:
- Discord bot token
- Discord webhook URLs (per identity)
- Deepgram API key
- Gemini API key (Google AI Studio)
- Turso database URL and auth token

All stored in `.env` on the droplet, loaded by the bot process.

## Backup Strategy

| Data | Storage | Rationale |
|------|---------|-----------|
| Audio recordings (Craig) | Local, or R2 if cloud backup desired | Large files, don't need to be in git. R2 at ~$0.015/GB if wanted. Audio is ephemeral tooling — let Craig files expire if not needed. |
| Transcripts | Git (campaign repo) | Text files, versioned alongside other campaign content. The durable artifact. |
| Session notes (generated) | Turso + Git | Turso for live access (portal, `/bone`). Git for archival alongside transcripts. |
| Campaign data | Turso (live) + Git (canonical source) | Already the existing pattern in hx-web. |

## Open Questions

- **Droplet RAM pressure:** Need to test actual memory usage with all components running. If 2GB is tight, the $6/mo 1GB bump to a 4GB droplet is cheap insurance.
- **AssemblyAI free credits:** $49.98 in free credits confirmed. At ~$6.24/mo, that's ~8 months of free transcription.
- **Gemini free tier durability:** Google has adjusted free tier limits before (reduced from 1500 to 250 req/day in early 2026). Monitor for changes.
