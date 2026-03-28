# Session Note Generation

## Overview

Automated post-session processing that turns a raw transcript into structured session notes: summaries, NPC attribution, entity updates, and scene breakdowns. Runs after each session using the full transcript as input.

## Pipeline

```
Full session transcript (from voice-transcription pipeline or Craig reprocessing)
  │
  ▼
Pass 1: NPC/Character Attribution
  GM speech classified as: narration | NPC dialogue (which NPC) | OOC
  │
  ▼
Pass 2: Session Summary
  1-2 paragraph "what happened this session"
  │
  ▼
Pass 3: Entity Extraction & Wiki Updates
  New/updated entries for NPCs, locations, items, plot threads
  │
  ▼
Pass 4: Per-Scene Breakdowns (optional)
  Scene-by-scene detailed notes
  │
  ▼
Storage (Turso) + ChromaDB re-index
  │
  ├──→ Web portal: session archive, per-session pages
  └──→ Campaign recall: indexed for /bone queries
```

## Processing Passes

### Pass 1: NPC/Character Attribution

The hardest pass. The GM is one audio stream voicing multiple NPCs, narrating, and speaking out-of-character. Discord gives us one "GM" stream — no acoustic separation.

**Approach:** LLM post-processing on text. The transcript text contains contextual cues that a human reader would use:

- "The blacksmith looks up. 'Aye, I can forge that.'" — obviously the blacksmith speaking
- "You see the duke approaching from the east." — narration
- "Hold on, let me check the rules for that." — OOC

**Prompt structure:**

```
Here is a TTRPG session transcript. The GM is [name]. When the GM speaks,
classify each segment as one of:
- NARRATION: scene description, world-building, mechanical rulings
- NPC:[name]: dialogue spoken as a specific NPC
- OOC: out-of-character remarks, rules discussion, table talk

Known NPCs in this campaign: [list from campaign data]
Known locations: [list]

If a new NPC appears that is not in the known list, flag them as NPC:[new name].

Return the annotated transcript with attribution tags.
```

**Expected accuracy:** ~85-90% from context alone. Ambiguous cases (GM doesn't name the NPC before speaking as them, pronoun-heavy dialogue) get flagged for GM review.

**GM review step:** After processing, the GM sees a list of flagged ambiguities: "At timestamp 1:23:45, the GM said 'I'll have it ready by morning.' Attributed to: Blacksmith (70% confidence). Correct?" Quick yes/no review.

### Pass 2: Session Summary

A 1-2 paragraph summary of the session: what the party did, key decisions, major revelations, cliffhanger ending.

**Input:** The attributed transcript from Pass 1 + prior session summaries for continuity context.

**Output format:**
```
## Session 12 — "The Governor's Gambit"

The party arrived in Thornwall seeking an audience with Duke Ashmore regarding
the mining rights dispute. After a tense negotiation...

**Key events:**
- Met Duke Ashmore; he offered 200gp for retrieval of [item]
- Discovered the underground passage beneath the market
- Kael'theron's private meeting with the informant (Elise)

**Cliffhanger:** The party heard marching boots as they entered the tunnels.
```

### Pass 3: Entity Extraction & Wiki Updates

Identify new or updated entities mentioned in the session:

- **New NPCs:** Name, description, first impression, relationship to party
- **Updated NPCs:** New information learned, status changes
- **New locations:** Name, description, how the party found it
- **Items:** Acquired, lost, or learned about
- **Plot threads:** New threads opened, existing threads advanced or resolved

Output is structured data that can be:
- Added to the campaign knowledge base (ChromaDB re-index)
- Surfaced as suggested wiki updates for the GM to approve
- Stored in Turso for `/bone` queries

### Pass 4: Per-Scene Breakdowns (Optional)

For GMs who want detailed notes, break the session into scenes with:
- Scene title and location
- NPCs present
- Key dialogue and decisions
- Mechanical events (combat, skill checks)
- Scene transitions

This is the most token-heavy pass and can be skipped if the summary is sufficient.

## IC vs OOC Detection

In-character vs out-of-character classification applies to all speakers, not just the GM:

- "I draw my sword and approach the gate" → IC action
- "Wait, can I add my proficiency bonus?" → OOC rules question
- "Hold on, my pizza's here" → OOC life
- "Kael'theron turns to the duke and says..." → IC dialogue

The LLM handles this as part of the attribution pass. OOC content is excluded from the session summary and entity extraction (no one needs "pizza arrived" in the campaign wiki).

## Cost

Estimated for a 4-hour session (~23,400 words, ~31,000 tokens of transcript):

| Pass | Input tokens | Output tokens | Flash cost | Flash-Lite cost |
|------|-------------|---------------|------------|-----------------|
| Attribution | ~33K | ~35K | $0.10 | $0.01 |
| Summary | ~35K | ~2K | $0.02 | $0.004 |
| Entity extraction | ~35K | ~3K | $0.02 | $0.005 |
| Scene breakdowns | ~35K | ~5K | $0.02 | $0.006 |
| **Total** | **~138K** | **~45K** | **~$0.15** | **~$0.03** |

Monthly (4 sessions): **~$0.62** on Flash, **~$0.13** on Flash-Lite.

## Testing Strategy

### Real Session Data (Primary)

Craig recordings from actual sessions provide the best test data:
- Real speech patterns, crosstalk, fantasy nouns
- Per-speaker audio files → Deepgram transcription → notes pipeline
- Can be reprocessed as prompts improve
- Start banking recordings immediately, even before the pipeline is built

### Noun-Swapped Critical Role Transcripts

Critical Role episodes are in every LLM's training data. To prevent the model from regurgitating memorized knowledge:

1. Take a CR episode transcript (from the [CR Wiki](https://criticalrole.fandom.com/wiki/Transcripts) or [CRD3 dataset](https://aclanthology.org/2022.emnlp-main.637.pdf))
2. Find-replace all proper nouns (characters, locations, items) with invented equivalents
3. Provide the replacement glossary as the "campaign knowledge base"
4. Run the pipeline and evaluate accuracy

This preserves the natural speech patterns and GM multi-voicing while preventing training data leakage.

### Obscure Actual Plays

Small/unknown actual play podcasts are unlikely to be in training data. Transcripts from these provide realistic quality without the training data concern. However:
- Single mixed audio (not per-speaker streams) makes STT testing less representative
- Text transcripts (if available) still work for testing the Gemini notes passes

### CRD3 Dataset

159 Critical Role episodes with dialogue turns AND matching wiki summaries. Ideal for benchmarking: run your pipeline, compare output against the wiki summary. Use with noun replacement.

## Integration

### Storage

Session notes stored in Turso:
- Per-session record: summary, attributed transcript, scene breakdowns
- Entity updates linked to session for provenance
- Accessible via hx-web portal (session archive pages) and `/bone` queries

### Portal (hx-web)

- Searchable session archive (similar to broadsheets page)
- Per-session pages: transcript, key events, NPCs mentioned
- Cross-references to characters, locations, broadsheets
- Timeline view across sessions
- Character notes displayed on sheet (private/public tabs)

### Backfill

Craig recordings banked from early sessions can be reprocessed as the pipeline matures. Improving the prompts or switching models retroactively improves all historical notes.

## Open Questions

- **Attribution review UX:** Where does the GM review flagged ambiguities? A web page? Discord thread? Bot DM?
- **Automatic vs manual trigger:** Does note generation run automatically when a session ends, or does the GM trigger it with a command?
- **Partial sessions:** What if the bot joins late or Craig wasn't recording? Handle gracefully with whatever transcript is available.
- **Wiki update approval:** Should entity updates be auto-applied to the knowledge base, or queued for GM approval? Auto-apply is convenient but risks hallucinated entities polluting the data.
- **Transcript editing:** Should the GM be able to edit the raw transcript before note generation (fix misheard words, remove garbage)? Useful but adds friction.
- **Multi-session context:** How much prior session context should be included in prompts? Full prior summary? Last 3 sessions? Token budget vs accuracy tradeoff.
