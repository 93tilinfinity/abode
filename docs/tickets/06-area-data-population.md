# Ticket 06 — Area-data population source (Step 4)

## Goal

For the surviving set, pull the relevant area-level data **by area** and populate it onto the properties, then apply a straight threshold check. In v1 the safety requirement is exactly this: pull crime statistics by area and fail anywhere breaching the configured safety threshold. No model judgement.

## Scope

- A source component (per Ticket 02) that runs **last** (heaviest of the v1 light steps, on the fewest items), keyed **by area** not by property: it fetches each distinct area's data once and applies it to every survivor in that area, keeping the work small.
- v1 area datum: **crime / safety statistics**. The data is populated onto each property and compared to the **safety threshold from config** (Ticket 01); a property in an area breaching the threshold **fails**.
- This is a deterministic data pull + threshold check, **not** a model judgement — explicitly the v1 handling of the otherwise-qualitative "feels safe" requirement, labelled as a thresholded-area-data heuristic.
- Results cached per area; areas with no resolvable safety data **fail closed** (cannot confirm → unmet).
- Structured so v2 can layer model judgement of area questions on top of this same populated data without reworking it.

## Out of scope

- Model judgement of "interesting area" / "feels safe" (v2).
- "Interesting area" as a v1 gate beyond available thresholded area data — area feel proper is left to the couple's eye in v1.

## How to validate

- For a survivor set spanning N distinct areas, the source makes at most one data fetch per area (assert fetch count == distinct-area count), each applied to all properties in that area.
- A property in an area above the safety threshold fails; one below passes; both carry the populated, labelled area datum.
- An area whose safety data can't be resolved causes its properties to fail closed.
- Per-area results are cached across the run and across runs where appropriate.
- Because it runs after Tickets 03–05, it operates on the already-shrunk survivor set (assert input size ≤ the post-gap-fill set).
