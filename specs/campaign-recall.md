# Campaign Recall — The Bone

## Overview

AI-powered campaign memory recall, accessed via the `/bone` Discord command. Players ask questions about the campaign and receive contextual answers drawn from session transcripts, correspondence, broadsheets, findings, and their character's personal notes.

The system uses RAG (retrieval-augmented generation) to find relevant campaign data and Gemini to synthesize an answer, scoped to what the querying player's character would know.

## Diegetic Framing

The AI is presented as an in-universe Remnant artifact: a Seljuk memory resonator shaped like a curved piece of warm bone-stone. It works through memory resonance, not display — holding it sharpens recall, and with focus, surfaces relevant memories.

`/bone` is the Discord command. The name is a frontier nickname — it may change if the players come up with their own.

### Concept

- A hand-sized bone-like object, warm to the touch, that records and indexes the holder's experiences
- Queries surface memories — recent experiences clearly, historical layers as shapeless impressions
- It's ancient and partially degraded — incomplete responses, gaps, and occasionally unsettling foreign memories are features not bugs
- It doesn't understand modern concepts, doesn't speculate, doesn't interpret
- It has no voice of its own — it gives you memories, not words (until the Lompen connection)

### Passive Recording

The device automatically absorbs and indexes the holder's experiences — this is the diegetic wrapper for session transcript ingestion. Players can query recent events immediately. Historical layers (pre-player data) are locked behind narrative progression.

## Commands

| Command | Description |
|---|---|
| `/bone <question>` | AI-powered memory recall scoped to querying character |
| `/bone search <term>` | Keyword search across session transcripts and notes |

### Discord Integration

- Could be its own webhook identity with a Remnant-themed name and avatar
- Responses formatted with a Remnant aesthetic (monospace, fragmented)
- Player identity resolved via: Discord user → hx-web user (discord_id) → active character
- The system prompt shifts based on which data tiers are unlocked for the querying player

## Per-Character Knowledge Scoping

Each query is scoped to the querying player's character. The device resonates with *their* recorded experiences.

### Context fed to the LLM includes:

**Included:**
- Session transcripts (shared — all players were present)
- Public events (broadsheets, official mail, roster data)
- That character's private correspondence
- That character's private notes
- That character's finding notes
- Documents their character has read/seen

**Excluded:**
- Other players' private mail
- Other players' private notes
- GM-only notes and data
- Unlocked tiers the character hasn't reached

The hx-web data model already supports this — mail is per-character, notes have private/public visibility.

### Implementation

When building the LLM prompt, filter content by character ownership and visibility before stuffing context. The filtering query runs against Turso (for structured data) and ChromaDB (for vector search), both filtered by character_slug and visibility.

## Data Access Tiers

| Tier | In-world unlock | System equivalent |
|---|---|---|
| Holder's recent experiences | Automatic (always recording) | Session transcripts, filed reports, correspondence |
| Shared common knowledge | No barrier | Broadsheets, official mail, roster data |
| Recent historical layers | Scholarship / Antiquarian talent | GM-curated lore snippets |
| Ancient experiential data | Dwaler + Lomp bridge | Deep lore, cosmology hints |
| Suspect/entity-touched impressions | No unlock — mixed in | GM-planted misdirection or foreshadowing |

Tiers are GM control levers. Expanding what the LLM has access to (adding curated lore to the context) is how new tiers "unlock" in play. Restricting access is how the device "degrades" or "goes dormant."

## RAG Architecture

```
Player: /bone "What did the governor say about the mining rights?"
  │
  ▼
Query embedding (Gemini embedding API or local model)
  │
  ▼
Vector search (ChromaDB)
  - Filter: character's accessible data (visibility + tier)
  - Return top N relevant chunks
  │
  ▼
Context assembly
  - Relevant transcript chunks
  - Related character notes, mail, findings
  - System prompt with Bone personality
  │
  ▼
Gemini Flash
  - Synthesize answer from provided context only
  - Enforce personality and limitations
  - Return formatted response
  │
  ▼
Discord: formatted response in Remnant aesthetic
```

### Context Sources

| Source | How it's indexed | Notes |
|---|---|---|
| Session transcripts | Chunked per scene or time window, embedded in ChromaDB | Primary source for "what happened" queries |
| Broadsheets | Per-article, embedded | In-world news, common knowledge |
| Character sheets | Per-character summary | Who's who, relationships |
| Mail/correspondence | Per-letter, filtered by recipient | Plot threads, personal knowledge |
| Finding notes | Per-finding, per-character | Observations on artifacts/locations |
| Handbook data | Per-section | World rules, kin, professions |
| GM-curated lore | Per-snippet, tier-gated | Historical layers, unlocked progressively |

## System Prompt Direction

The Bone responds not as a device speaking, but as memory surfacing — terse, impressionistic, fragmentary.

### Constraints

- Only reference information from provided context (the holder's recorded experiences)
- Do not speculate, interpret motives, or predict
- Do not understand post-collapse concepts
- Can fail to surface anything ("nothing resonates," "impression too degraded to parse")
- No conversational filler, no apologies, no disclaimers

### Tone

- Terse, fragmentary, present-tense impressions
- Recent memories: clear, specific, almost verbatim
- Older memories: faded, impressionistic, incomplete
- Lomp-bridged historical layers: older, stranger, more vivid — a different quality of memory

### GM Control Levers

- **Data access:** Control what data the LLM has access to per query. Inaccessible memory layers = curated input.
- **Forced failures:** The device can surface confusing/irrelevant impressions, refuse to resonate, or go dormant.
- **Progressive unlock:** New capabilities tied to narrative progression (Dwaler/Lomp connection).
- **Planted content:** GM can inject specific "memories" (lore, misdirection, foreshadowing) into the context.
- **System prompt shifts:** The prompt evolves as tiers unlock, changing the device's personality and capabilities.

## Cost

Per query: ~500 tokens in (question + system prompt), ~2K-5K context from RAG, ~200-500 tokens out.

At ~$0.30/$2.50 per 1M tokens (Flash), each query costs ~$0.002. Even 100 queries per session = $0.20. Free tier (250 req/day) covers this easily.

## Open Questions

- **LLM provider:** Spec assumes Gemini Flash. Could also use Claude or GPT for personality quality. Gemini is cheapest but tone/personality may need more prompt engineering.
- **Chunking strategy:** Per-session, fixed token windows, or semantic splitting? Affects retrieval quality.
- **Cost management:** Cache common questions? Rate limit per player per session?
- **Spoiler protection:** The data filtering should prevent spoilers, but edge cases exist. What if a transcript contains GM narration that reveals something a specific character wasn't present for? Needs careful scoping.
- **Bone personality iteration:** The system prompt needs extensive testing with real queries. May need to maintain a test suite of query/expected-response pairs.
- **Webhook identity:** Should the Bone have its own webhook avatar, or respond as the bot?
