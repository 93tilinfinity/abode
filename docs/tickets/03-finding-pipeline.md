# Ticket 3 — The finding pipeline (scrape → gate → matches)

**Depends on:** Ticket 1 (the compiled pass/fail rules).

> **Worked plan:** [`plans/ticket-03-finding-pipeline.md`](../plans/ticket-03-finding-pipeline.md).
> **Source & shape choices:** [Decision 0001](../decisions/0001-data-providers.md).
> In short: listings are **scraped from Rightmove's own search API** (paginated +
> price-band tiled for completeness, bathrooms post-filtered in code); the rest of
> the gates use **free** sources (broadband, parks, crime) and a per-property
> **Google Routes** commute. It's one hand-written linear pipeline — there is no
> separate source framework.

## Goal

Implement the v1 finding pipeline: run the config's deterministic gates, cheapest
first, over the scraped candidate set and emit the day's matching set — every UK
for-sale property that passes every gate. Rendering and scheduling are Ticket 4.

## The pipeline (one linear sequence, each step shrinks the set)

1. **Scrape Rightmove (the one broad fetch).** Call `…/api/_search` across a fixed
   home-centred **radius**, with `minPrice`/`maxPrice` and `minBedrooms` set from
   config so price and beds are filtered server-side. **Paginate** (`index`) and,
   where the ~1,050-result cap would truncate, **tile by price band** and merge,
   de-duplicating by listing id, so the candidate set is *complete*. Each result
   yields price, beds, **bathrooms**, type, coordinates, address, listing URL and
   a thumbnail. Apply the **bathrooms** gate (`toilet_count ≥ 2`) in code; a
   listing missing the value **fails closed**.
2. **Broadband gate (free, per postcode).** `max_download_mbps ≥ 300` from Ofcom
   postcode data. Unknown → **fails closed**. Fetched once per postcode, reused.
3. **Outdoor-space gate (free).** `garden OR park ≤ 800 m`. Use the listing's
   garden/outdoor signal first; only when it's false/unknown, look up nearest park
   (OSM/OS). Unknown both ways → fails closed.
4. **Crime gate (free, by area).** police.uk violent/sexual-crime count within
   ~1 mile over the last 12 months vs the config threshold. Fetched **once per
   area** and reused for nearby properties within the run.
5. **Commute gate (Google Routes, last — the costliest step).** A door-to-door
   transit journey to the work anchor (depart Tue 08:00): read total minutes
   (`commute_time ≤ 50`) and count non-walking leg modes (`commute_modes ≤ 2`).
   Runs only on properties that survived steps 1–4, so paid journeys are minimal.
   The minutes are also kept for the page column (and the v2 shortlist).

Floor area + EPC rating (free, Open Data Communities) are pulled across survivors
for the page column / £/sqft — **not a gate**: a missing EPC just leaves floor
area blank, never fails a property.

A field that no source can supply is marked unrecoverable and its gate fails
closed (Ticket 1 semantics). A value recovered from a secondary source is treated
as authoritative but tagged with its `source` and **sanity-checked** (out-of-range
→ treated as not recovered).

The output is the day's matches: the records passing every gate, handed to
Ticket 4. The only broad fetch is step 1; everything after is light lookups on a
shrinking set.

## Out of scope

- The HTML page, the batch runner, scheduling (Ticket 4).
- The shortlist scoring (Ticket 5).
- OCR, model condition/area judgement (all v2).

## How to validate

> **The golden validation is end-to-end, once the requirements are finalised** —
> the complete must-have set (all thresholds locked, crime cutoff and radius
> calibrated) run through the full pipeline, judged on the *matching set itself*
> (see SPEC, "What 'correct' means"). The checks below are scaffolding that keep
> the build honest on the way there; they are not the acceptance bar.

1. **Complete scrape.** Against a known Rightmove search for the area, assert the
   pipeline's candidate count matches (pagination works) and that a query which
   would exceed ~1,050 results is tiled by price band and de-duplicated rather
   than truncated.
2. **Server vs code filters.** Assert price and beds are sent as request params
   (filtered server-side); assert bathrooms is filtered in code and that a listing
   with a null bathrooms value **fails** the toilets gate rather than passing.
3. **Shrink early, journey last.** Assert each later step's input count is ≤ the
   previous, and that Google Routes is called only for properties that passed
   price/beds/baths/broadband/outdoor/crime — never the whole candidate set.
4. **Fail-closed gates.** A postcode with unknown broadband fails the fibre gate;
   a garden-less property with no park ≤ 800 m fails outdoor-space.
5. **Outdoor OR short-circuit.** A garden property passes without a park lookup; a
   garden-less one passes iff a park is ≤ 800 m.
6. **Crime threshold, by area.** Two areas either side of the configured count:
   properties in the breaching area fail `safe_area`, the others pass; crime is
   fetched once per area and reused; the 12-month window is summed correctly.
   Changing the threshold in config moves the line with no code edit.
7. **Commute read.** For a known journey departing Tue 08:00, assert minutes (≤ 50)
   and non-walking mode count (≤ 2) are read correctly and populated for the page.
8. **EPC is not a gate.** A property with no EPC match keeps blank floor area and
   is **not** failed; one with an EPC exposes floor area in m².
9. **End-to-end.** With fixed mock sources and the example config, the pipeline
   returns exactly the properties passing all gates; flipping a config threshold
   changes the set as expected (ties back to Ticket 1).

## Hygiene gate (before committing)

Before every commit on this ticket, the project hygiene gate must be green:

- `make check` passes — `gofmt -l .` clean, and `go vet ./...`, `go build ./...`,
  `go test ./...` all succeed;
- the new/changed behaviour is covered by tests;
- nothing is committed red (CI re-runs the same checks on push).

(The gate and the `make check` target are established in
[Ticket 0](00-project-setup-go.md).)
