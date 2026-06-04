# Ticket 03 — Catchment / isochrone builder (Step 1)

## Goal

Build, or reuse a cached, polygon of everywhere within the target commute time of the work anchor. This polygon is both the geographic boundary that bounds the paid search and a free local commute check — most of the UK is excluded here at no cost, before any paid call.

## Scope

- A source component (per Ticket 02) that, given the work anchor and commute budget from config (Ticket 01), produces a catchment polygon via an isochrone/journey source.
- The result is **cached** and reused across runs; it only recomputes when the anchor or budget changes. This is the cheapest, broadest step and must not re-pay needlessly.
- Provides two products used downstream: (a) the polygon as the **search boundary** for Ticket 04, and (b) a free **inside/outside** commute test usable as the commute-boundary gate so a property outside the polygon fails the commute requirement at zero marginal cost.
- Honest-about-what-it-knows: the commute figure derived here is the catchment-time basis; if a finer per-property journey time is shown later it is labelled for what it is.

## Out of scope

- Per-property door-to-door journey timing beyond the catchment membership test (not required for v1's gate).
- The paid property search (Ticket 04).

## How to validate

- Given an anchor + budget, a polygon is produced and a point known to be within the budget tests inside; a point known to be well beyond tests outside.
- Re-running with unchanged anchor/budget serves the cached polygon and makes no fresh isochrone call (assert cache hit).
- Changing the budget in config invalidates the cache and recomputes.
- The commute-boundary gate, fed only the polygon, correctly passes inside points and fails outside ones (fail-closed for points with no resolvable location).
