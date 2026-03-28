# Voice Transcription Pipeline

## Overview

Post-session transcription of recorded Discord sessions, producing speaker-attributed text that feeds into session notes and campaign recall systems. The workflow is manual and runs locally — no always-on transcription pipeline on the droplet.

## Architecture

```
During Session:
  Craig bot records Discord voice channel
  → Per-speaker FLAC files, time-synchronized

After Session (local, manual):
  Download Craig recordings
  → Feed per-speaker audio to AssemblyAI batch API (with keyterms)
  → Post-STT fuzzy correction against campaign glossary
  → Merge into chronological session transcript
  → Commit transcript to git (campaign repo)
  → Run Gemini note generation passes
  → Upload results to Turso (for /bone, portal)
       │
       ├──→ Session notes (see session-notes.md)
       └──→ Indexed for campaign recall (see campaign-recall.md)
```

## Future: Live Transcription

If real-time transcription becomes desirable (for the sidebar — see [realtime-sidebar.md](realtime-sidebar.md)), the bot could be extended to:
- Auto-join the voice channel and receive per-user audio streams
- Stream to AssemblyAI in real-time with keyterms
- Feed live transcript to the sidebar pipeline

This is additive — the batch workflow remains the baseline. Live transcription is an upgrade, not a prerequisite.

## STT Provider: AssemblyAI Universal-3 (Recommended)

### Head-to-Head Test Results

Both providers were tested against the same FWDC campaign audio — a proper noun stress test with kin names, Snuffel names, NPC names, locations, and fauna terms.

| Term (Actual) | Deepgram Nova-3 | AssemblyAI Universal-3 | Winner |
|---------------|-----------------|------------------------|--------|
| Snorredol Wuffeknol | "Snordat, Ulfnol" (shattered) | "Snorredol Wuffeknol" | **AssemblyAI** |
| Snorredol Wuffeknol (with keyterm) | "Snorredol Wuffenel" (still wrong) | "Snorredol Wuffeknol" | **AssemblyAI** |
| Dallas "Patch" Westorvan | "Dallas, Patch, West Orban" | "Dallas, Patch, Westorvan" | **AssemblyAI** |
| Verdaal | "Verdal" | "Verdaal" | **AssemblyAI** |
| Old Piet | "Old Pete" | "Old Piet" | **AssemblyAI** |
| Lompen | "Lumpin" | "Lompen" | **AssemblyAI** |
| Lomp | "lump" | "lomp" | **AssemblyAI** |
| Bounders | "founders" | "bounders" | **AssemblyAI** |
| Varkens | "Farken" | "Varken" | **AssemblyAI** |
| Krekels | "crinkles" | "crickles" | **AssemblyAI** |
| Cornelis Vex | "Cornelius Specks" | "Cornelius Vex" | **AssemblyAI** |
| Dwaler | "Dwaller" | "Dwaller" | Tie |
| Willem Swan | "William Swan" | "William Swan" | Tie |
| Helena Koster | "Elena Koster" | "Elena Koster" | Tie |

**Key finding:** Deepgram failed to apply a provided keyterm correctly ("Wuffeknol" → "Wuffenel" even with the exact keyterm). AssemblyAI produced consistently better results across the board, especially on unusual fantasy nouns.

### Provider Comparison

| Requirement | AssemblyAI Universal-3 | Deepgram Nova-3 | SeaVoice | Whisper |
|------------|----------------------|-----------------|----------|---------|
| Custom vocabulary | 1,000 word keyterms + 1,500 word prompting | 500 tokens (~100 terms) | No | Prompt hints only |
| Fantasy noun accuracy | Strong (see test) | Weak even with keyterms | Unknown | Moderate |
| Streaming | Yes | Yes | Yes (no API) | Batch only |
| Per-user streams | We control input | We control input | Bot controls | We control input |
| Pipeline control | Full | Full | None | Full |
| Cost/min (streaming) | ~$0.0065 | $0.0077 | Free | Varies |

### Decision

**AssemblyAI Universal-3** is the primary recommendation based on:
- Superior accuracy on fantasy/unusual proper nouns, tested against actual campaign content
- 2x the keyterm budget (1,000 words vs 500 tokens)
- Additional 1,500 word "prompting" context for guiding transcription style
- Cheaper per minute ($0.0065 vs $0.0077)
- Actually applies keyterm hints when provided

SeaVoice was rejected: no custom vocabulary, no pipeline control, no programmatic API access.

Deepgram remains a viable fallback if AssemblyAI has issues with Discord audio format compatibility or streaming latency.

## Custom Vocabulary (Fantasy Noun Problem)

Standard STT models don't know campaign-specific proper nouns. "Kael'theron" becomes "kale thereon." This is the biggest practical accuracy issue for TTRPG transcription.

### Solution: Keyterm Glossary

The campaign knowledge base (NPCs, locations, items, factions) provides a glossary of proper nouns. This glossary is included in every Deepgram request as keyterm prompts.

- Deepgram Nova-3 supports up to 100 keyterms per request (more on Enterprise)
- Keyterms are passed per-request, so the list can be updated dynamically
- When a new NPC or location is added in hx-web, it's automatically available for the next transcription call
- Source of truth: hx-web campaign data (characters, findings, locations, broadsheet entities)

### Fallback: Post-STT Fuzzy Matching

Even with keyterms, some words will be misheard. A post-transcription correction pass:

1. Tokenize transcript into words/phrases
2. Compare against known entity list using phonetic distance (Soundex, Metaphone) and Levenshtein distance
3. Replace close matches above a confidence threshold
4. Flag uncertain matches for GM review

This is a simple string-matching pass, no LLM needed.

## Speaker Attribution

Discord's voice API provides **separate audio streams per user**. This means:

- No speaker diarization needed — each stream is already one person
- Simultaneous speakers produce separate, correctly attributed transcripts
- Crosstalk (in the mixed-microphone sense) is a non-issue

### Cast Mapping

Each Discord user maps to a player and active character:

```
Discord user "Mike"  → Player: Mike  → Character: Kael'theron
Discord user "Sarah" → Player: Sarah → Character: Elara Nightbloom
Discord user "Dan"   → Player: Dan   → GM (voices all NPCs)
```

This mapping already exists in hx-web (Discord OAuth provides `discord_id` in the users table, users have active characters). The bot resolves it automatically.

For the transcript, player lines are attributed to their character. GM lines are attributed to "GM" — NPC attribution is a post-processing problem handled in [session-notes.md](session-notes.md).

## Craig Bot (Primary Recording)

[Craig](https://craig.chat/) is the recording method. It records Discord voice channels with **per-speaker, time-synchronized audio files** (FLAC, AAC).

### Behavior

- Per-speaker FLAC files, time-synchronized regardless of when users join/leave
- Late joins: silence padding at the start of their track, maintains sync
- Disconnects/reconnects: silence gap in their track, sync preserved
- Actively corrects for clock drift across tracks
- Up to 6 hours per recording, no speaker limit
- Free tier: manual `/join`, files stored 7 days
- $4/mo Premium: auto-record when N users join, 14-day storage

Manual `/join` each session. The Discord bot could post a reminder when it sees players gathering in voice.

### Audio Storage

- Craig free tier stores files for 7 days — download after each session
- Audio files kept locally (or R2 if cloud backup desired, ~$5/mo for a year of sessions)
- Audio is ephemeral tooling — the transcript in git is the durable artifact
- Reprocessable: improved models or prompts can be re-run against old audio if retained

## Cost

| Component | Rate | Per session (4hr) | Monthly (4 sessions) |
|-----------|------|-------------------|---------------------|
| AssemblyAI batch | ~$0.0065/min | ~$1.56 | ~$6.24 |
| Craig | Free (manual `/join`) | — | $0 |
| Audio storage (R2, optional) | $0.015/GB | ~$0.10 | ~$0.42 |
| **Total** | | **~$1.56** | **~$6.24** |

$49.98 in AssemblyAI free credits covers ~8 months.

## Implementation

1. Record sessions with Craig (manual `/join`)
2. Download per-speaker FLAC files after session
3. Local script: feed each file to AssemblyAI batch API with campaign keyterms
4. Local script: post-STT fuzzy correction against campaign glossary
5. Local script: merge transcripts by timestamp into chronological session transcript
6. Commit transcript to git (campaign repo)
7. Run Gemini note generation passes (see [session-notes.md](session-notes.md))
8. Upload structured notes to Turso for portal and `/bone` access

## Keyterm Budget

AssemblyAI Universal-3 supports **1,000 keyterm words** plus an additional **1,500-word prompting context**. This is significantly more headroom than Deepgram's 500 tokens, but a rich campaign world can still push limits — especially with multi-word Snuffel names, nicknames, and shortenings all consuming budget.

### Mitigation: Three-Layer Vocabulary Correction

No single layer needs to be perfect. Each catches what the previous missed:

**Layer 1 — STT Keyterms (limited, high accuracy)**

Rotate keyterms by session relevance instead of loading the full glossary:
- NPCs active in the current arc/location
- Current location and nearby locations
- Recently mentioned items and plot threads
- Always-include core terms (party names, key factions, world terms)

A "session prep" step selects 60-80 most relevant terms. The bot could auto-select based on recent session context and current arc data.

**Layer 2 — Post-STT Fuzzy Correction (unlimited, medium accuracy)**

After transcription, run a phonetic/string matching pass against the **full campaign glossary** (no token limit — this is local string matching):
- Soundex / Metaphone phonetic encoding
- Levenshtein distance for near-matches
- Configurable confidence threshold
- Uncertain matches flagged, not auto-corrected

This catches terms that weren't in the keyterm list. "Ver doll" → "Verdaal" if "Verdaal" is in the glossary and phonetically close enough.

**Layer 3 — LLM Entity Resolution (unlimited, high accuracy, slower)**

The Gemini entity extraction step (in the sidebar and notes pipelines) gets the full campaign glossary in its prompt context. It can resolve ambiguous transcription with semantic understanding: "that merchant from the coast" → the LLM knows which merchant that is from campaign context.

### Alternative STT Providers (Fallback)

If AssemblyAI proves insufficient, other options:

| Provider | Custom vocab limit | Streaming | Cost/min |
|----------|-------------------|-----------|----------|
| **AssemblyAI Universal-3** | **1,000 words + 1,500 word prompting** | **Yes** | **~$0.0065** |
| Deepgram Nova-3 | 500 tokens (~100 terms) | Yes | $0.0077 |
| Google Cloud STT | PhraseSet with boost (no hard cap documented) | Yes | $0.024 |

Deepgram tested poorly on fantasy proper nouns even with keyterms provided (see test results above). Google Cloud is 3-4x more expensive.

## Open Questions

- **STT provider:** AssemblyAI is the current recommendation based on head-to-head testing. Deepgram is the fallback. Need to verify AssemblyAI's streaming API compatibility with Discord.js audio format.
- **Keyterm rotation strategy:** Manual session prep (GM picks the arc) vs automatic (bot infers from recent session context)? Automatic is better UX but harder to build.
- **Audio format:** Deepgram supports multiple input formats. What does Discord.js provide natively? PCM 48kHz stereo is typical — confirm compatibility.
- **Transcript storage format:** Raw text with timestamps per utterance? Structured JSON? Needs to support both real-time streaming and post-session batch processing.
- **Session boundary detection:** How does the bot know when a "session" starts and ends? Manual `/session start` and `/session stop`? Or inferred from voice channel activity?
- **Background noise handling:** Open mics with TV, keyboard, etc. will produce garbage transcription. Should we implement voice activity detection (VAD) client-side, or rely on Deepgram's built-in endpointing?
- **AssemblyAI free credits:** $49.98 confirmed (~8 months at current estimates). Verify credits apply to streaming, not just batch.
