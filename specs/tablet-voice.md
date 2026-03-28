# Tablet Voice via TTS (Stretch Goal)

## Overview

Stretch goal: instead of text responses in Discord, the Remnant tablet speaks aloud during sessions via text-to-speech, integrated into Foundry VTT. This is Phase 8 — late-campaign spectacle, not a near-term build target.

## Narrative Concept

The tablet has no speaker. It speaks *through a Lompen* — using their flexible vocal cords and psychic connection as an output device. The Lomp doesn't choose to speak; the tablet uses it as a vessel.

This ties the voice capability directly to the Lomp awakening progression: no awakened Lomp = no voice. The voice feature is a reward for engaging the hardest plot threads.

## Technical Pieces

| Component | Purpose |
|-----------|---------|
| Foundry module | UI trigger and audio playback in VTT |
| Gemini | Answer generation (same as `/bone`, see [campaign-recall.md](campaign-recall.md)) |
| TTS service | ElevenLabs, Google Cloud TTS, or similar |
| WebSocket | Push audio from bot to Foundry module |

### Voice Design

An alien synthetic voice — slightly distorted, tonal, not quite right. The Lomp's vocal apparatus is flexible but not human. The voice should feel unsettling and ancient, not like a polished assistant.

ElevenLabs offers voice cloning and custom voice design. A custom voice trained on specific tonal qualities could sell the alien-artifact feel.

### Flow

```
Player asks /bone question (or Foundry UI trigger)
  → Gemini generates text response (same as existing recall)
  → Text sent to TTS service
  → Audio returned
  → WebSocket pushes audio to Foundry module
  → Module plays audio in the VTT session
  → Text response also posted to Discord (fallback / log)
```

## Prerequisites

- `/bone` recall system working and personality proven (Phase 7)
- Lomp awakening reached in narrative (mid-to-late campaign)
- Foundry VTT still in use (see sidebar spec for Foundry vs standalone discussion)

## Cost

- ElevenLabs: $5/mo starter tier, or ~$0.30 per 1K characters. At ~200 characters per response and ~50 responses per session, that's ~$3/session. More expensive than the entire rest of the stack.
- Google Cloud TTS: cheaper but less voice customization.
- This is the one component where cost might actually matter.

## Open Questions

- **TTS provider:** ElevenLabs for quality/customization vs Google Cloud TTS for cost? The alien voice design may require ElevenLabs' capabilities.
- **Latency:** TTS adds 1-3 seconds on top of the LLM response. Is that acceptable for the spectacle value?
- **Foundry module:** Does this require a different module than the sidebar, or an extension of it?
- **Fallback:** If Foundry isn't running or the Foundry module fails, the text response in Discord is the fallback. Should the voice feature be Foundry-only or also work via Discord voice?
- **Player trigger:** Do players invoke this via Discord `/bone` and hear it in Foundry, or via a Foundry UI element?
