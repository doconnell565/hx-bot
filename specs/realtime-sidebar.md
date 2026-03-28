# Real-Time Reference Sidebar

## Overview

A GM-facing panel that surfaces campaign reference material relevant to the live conversation during a session. When someone mentions an NPC, location, or plot thread, the sidebar shows the relevant campaign data within seconds.

## Pipeline

```
Live transcript (from voice-transcription pipeline)
  │
  ├── Fast path: pattern match known entity names → instant ChromaDB lookup
  │
  └── Slow path: batch 15-30s sliding window
        → Gemini Flash-Lite (entity extraction)
        → ChromaDB vector search
        → return top matches
  │
  ▼
Sidebar API (JSON response)
  │
  ▼
Frontend (Foundry module or standalone browser tab)
```

### Fast Path (No LLM)

Known entity names from the campaign glossary are pattern-matched directly against the transcript text. When "Duke Ashmore" appears in the transcript, the sidebar can surface his entry in ~4 seconds from speech (STT latency + lookup) without any LLM call.

This covers the common case: players and GM referring to known entities by name.

### Slow Path (LLM-Assisted)

For ambiguous or indirect references ("that merchant we met last session," "the thing from the ruins"), a sliding window of transcript is sent to Gemini for entity extraction. The LLM identifies what's being discussed and returns structured entity references for ChromaDB lookup.

- **Sliding window:** Every 15-30 seconds, send the last 60 seconds of transcript
- **Prompt:** "From this TTRPG conversation transcript, extract all character names, location names, items, spells, and rules concepts mentioned. Match against these known entities where possible: [glossary]. Return as JSON."
- **Model:** Gemini Flash-Lite ($0.10/$0.40 per 1M tokens) — entity extraction doesn't need full Flash reasoning

## Campaign Knowledge Base (ChromaDB)

ChromaDB stores embedded campaign data for vector similarity search.

### Data Sources

| Source | Content | Update frequency |
|--------|---------|-----------------|
| NPCs | Name, description, relationships, last interaction | On change in hx-web |
| Locations | Name, description, connections, hex coordinates | On change in hx-web |
| Findings | Specimen reports, player names, descriptions | On change in hx-web |
| Broadsheets | Published articles (in-world news) | On publish |
| Session summaries | Previous session notes | After each session |
| Mail/correspondence | Plot-relevant letters | On send |
| Handbook data | World rules, kin, professions | Rarely |

### Chunking

One chunk per entity/document. Each chunk includes:
- Entity type and name (for filtering)
- Full text content
- Metadata (last updated, source, campaign_id)

Embedding model: `nomic-embed-text` or similar small model, run via the Gemini embedding API or a local model on the droplet.

### Sync

Campaign data syncs from hx-web to ChromaDB when entities are created or updated. This could be:
- A webhook from hx-web on data change
- A periodic sync job (every N minutes)
- On-demand re-index via bot command

## UX Design

### GM-Only (v1)

The sidebar is a GM tool. Players don't see it. This avoids:
- Spoiler risk (AI surfacing secret plot information)
- Meta-gaming (players acting on sidebar info their characters don't have)
- Complexity (no per-player filtering needed for v1)

The GM can selectively share information with players through existing channels (Foundry handouts, Discord messages, etc.).

### Layout

```
┌─ Active Topics ──────────────────┐
│                                  │
│  ● Thornwall (location)     [pin]│
│  ● Duke Ashmore (NPC)       [pin]│
│  ● The Silver Key (item)         │
│                                  │
│  ▶ Duke Ashmore              ──┐ │
│    Leader of Thornwall. Has     │ │
│    offered 200gp for the        │ │
│    retrieval of...              │ │
│                                 │ │
├─ Fading ─────────────────────────┤
│  ○ Market District               │
│  ○ healing potions               │
└──────────────────────────────────┘
```

### Interaction

- **Collapsed by default** — topic names with type icons, not walls of text
- **Click to expand** — shows the full reference entry
- **Pin** — locks a topic in the Active section regardless of recency
- **Dismiss** — removes a topic from the sidebar
- **Active vs Fading** — topics mentioned recently/frequently stay in Active; one-off mentions fade after a few minutes of no re-mention
- **Deduplication** — same entity mentioned multiple times = one entry with bumped timestamp, not duplicate entries

### Relevance Scoring

- Frequency: mentioned 3+ times in a short window > mentioned once
- Recency: mentioned 30 seconds ago > mentioned 10 minutes ago
- An entity mentioned once in passing shouldn't claim sidebar space; repeated mentions should

### Push to Foundry (Future)

The GM could push a sidebar reference to players' Foundry VTT view — e.g., revealing an NPC portrait or location description as a handout. This requires the Foundry module to support receiving push events from the bot's API. Not v1.

## Frontend Options

### Option A: Foundry VTT Module (on The Forge)

- Native sidebar panel in the Foundry UI
- Module makes `fetch()` calls to the bot's API endpoint on the DO droplet
- Constraint: The Forge is managed hosting. Module can only make external HTTP calls. No custom server processes.
- Bot API must be HTTPS-accessible (Cloudflare Tunnel or Tailscale Funnel)
- CORS headers needed (Forge serves over HTTPS)

### Option B: Standalone Browser Tab

- Simple HTML/JS page served by the bot
- Auto-refreshes or uses WebSocket for live updates
- Runs on a second monitor
- No Foundry dependency — works regardless of VTT choice
- Simplest to build and maintain

**Recommendation:** Start with Option B (standalone tab). It's faster to build, has no external dependencies, and works immediately. If Foundry integration proves valuable, build the module later — it's just a different frontend consuming the same API.

## Cost

| Component | Rate | Monthly (4 sessions, 15s intervals) |
|-----------|------|--------------------------------------|
| Gemini Flash-Lite (entity extraction) | $0.10/$0.40 per 1M tokens | ~$0.50 |
| Gemini spend cap | Set at $5-10/mo | Safety net |
| ChromaDB | File-based, on droplet | $0 (included in droplet) |

The fast path (pattern matching) has zero API cost. Only the slow path (ambiguous references) triggers Gemini calls.

## Implementation

1. Build the ChromaDB knowledge base ingestion from hx-web campaign data
2. Build the entity extraction prompt and test against sample transcripts
3. Build the sidebar API endpoint (returns current active topics as JSON)
4. Build the standalone HTML frontend with auto-refresh
5. Wire into the live transcript stream from the voice pipeline
6. Test with Craig recordings before going live

## Open Questions

- **Embedding model:** Gemini's embedding API vs. a local model (nomic-embed-text via Ollama). Local saves API calls but adds a process to the droplet. Gemini embedding API cost is negligible at this volume.
- **ChromaDB on 2GB droplet:** Memory footprint needs testing. Campaign-sized dataset should be small (~50-100MB), but need to verify it fits alongside the bot and hx-web.
- **WebSocket vs polling for the frontend:** WebSocket gives instant updates. Polling every 5 seconds is simpler and fine for the use case. Start with polling.
- **What happens when the sidebar is wrong?** GM sees an irrelevant reference. Is there a feedback mechanism (thumbs down, dismiss + "not relevant") that improves future results? Probably overkill for v1.
- **Foundry module feasibility:** The Forge's restrictions on external HTTP calls need testing. Can a module freely call an external API, or are there CSP/network restrictions?
