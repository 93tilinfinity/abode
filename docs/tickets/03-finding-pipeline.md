# Ticket 3 — The finding pipeline (the four evaluation steps)

**Depends on:** Ticket 1 (rules) and Ticket 2 (source framework + ordering).

## Goal

Implement the four real data sources of the v1 evaluation algorithm as uniform
components on Ticket 2's framework, so that running the pipeline against the
config's must-haves produces the day's matching set — every UK for-sale property
that passes every deterministic gate. The cost ordering is inherited from the
framework; this ticket supplies the actual sources in the order the principles
demand: lightest work on the most properties, paid breadth exactly once.

Deliver the four sources:

1. **Catchment builder (free, local).** Compute, or reuse a cached, polygon of
   everywhere within the target commute time of the work anchor. It is both the
   search boundary and the free commute check — most of the UK is excluded here
   at no cost. Cache the polygon between runs.
2. **Paid API property search (one feed, broad).** Search properties inside the
   catchment using every constraint the paid API supports directly (price, beds,
   baths) and request every other useful field in the same call (floor area, EPC,
   tenure, type, coordinates, photos, agent). One paid breadth call; most
   attributes populated from this single feed. (Commute time is **not** asked of
   this feed — the listings API does not know the journey to the work anchor; it
   comes from the journey source below.)
3. **Secondary-source gap-fill (light, across candidates).** For fields the feed
   didn't provide — fibre/internet availability, outdoor space / park proximity,
   and similar — do light per-field lookups across the candidate set, each
   cached. Recovered values are authoritative but flagged by source **and
   sanity-checked** (a recovered value outside a plausible range is treated as
   not recovered). Mark any field that stays missing-and-unrecoverable so its
   gate fails closed. This step also includes a **per-property commute-time
   lookup** against the journey/isochrone source: the catchment gate already
   proved each candidate is inside the target time, but the page column and the
   shortlist criterion need the actual minutes, so look them up here — one cached
   call per candidate, light across the shrunk set — and populate `commute_time`
   onto each property.
4. **Area-data population for survivors (light).** For the surviving set, pull
   area-level data **by area** (e.g. crime stats for the safety requirement) and
   populate it onto each property in that area. v1 is a straight pull + threshold
   check — breach the safety threshold and fail — **not** model judgement.

The output is the day's matches: properties passing every step's gates. This
ticket produces the matching **set** (a list of property records); rendering and
scheduling are Ticket 4.

## Out of scope

- The HTML page, the batch runner, scheduling (Ticket 4).
- The shortlist scoring (Ticket 5).
- OCR, model condition/area judgement (all v2).

## How to validate

1. **Step order & cost shape.** Run the pipeline and assert the realised order is
   catchment → paid search → gap-fill → area data, that the paid search is the
   only broad/paid call, and that each later step's input count is ≤ the previous
   (the set only shrinks).
2. **Catchment as a gate.** Place a property outside the polygon and assert it
   never reaches the paid search; place one just inside and assert it does. Assert
   a second run reuses the cached polygon.
3. **Single paid call.** Assert the paid source is invoked once per run for the
   catchment, not per property, and that it returns the requested extra fields.
4. **Gap-fill & fail-closed.** Mock the feed to omit fibre; assert the secondary
   source fills it where available, and that a property whose fibre stays
   unrecoverable **fails** the fibre gate rather than passing.
5. **Provenance.** Assert a gap-filled field records its secondary source and is
   still treated as authoritative by the rule. Assert a recovered value outside a
   plausible range is rejected (treated as not recovered) rather than used.
6. **Per-property commute time.** Assert every candidate gets a `commute_time`
   from the journey source (not the listings feed), that the lookup runs on the
   post-catchment candidate set (not the whole field), and that the populated
   minutes are consumed by the page column and shortlist criterion.
7. **Area threshold, not judgement.** Give two areas crime data either side of
   the safety threshold; assert every property in the breaching area fails and
   none in the safe area fails on that gate, with no model involved.
8. **End-to-end set.** With a fixed mock dataset and a fixed config, assert the
   pipeline returns exactly the properties that pass all gates — and that flipping
   one config threshold changes the matching set as expected (ties back to
   Ticket 1).
