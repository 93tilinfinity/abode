# Ticket 5 — Shortlist comparison

> **Deferred to v2** — built alongside the v2 favourites layer that feeds it. Design
> captured: the shortlist is marked on the page; weights default **equal across the
> four criteria**; a too-close-to-call flag avoids false precision on near-ties.

**Depends on:** Ticket 1 (weights in config) + the Ticket 3 property record (incl.
`commute_time`); comps from the free HM Land Registry. A separate stage, not the
daily finding loop.

## Goal

An on-demand decision aid: for the 10–15 properties the couple actually like, score
the measurable criteria by transparent formula, normalised across the shortlist,
with their own weights — a single sortable metric plus a fully interrogable
breakdown. Explainability over precision: every number traces to its evidence.

Deliver:

- **Shortlist input.** The couple mark 10–15 properties; only this hand-picked set
  is scored (weighting is meaningful only on a small considered set, never the full
  survivor set).
- **Measurable-criteria scoring** — **price, floor area, commute, £/sqft vs comps**
  — each by transparent formula, **normalised to the shortlist** (the best value
  anchors the scale), each carrying a **stated reason** ("8.5/10 — 11% below local
  median £/sqft").
- **Comp data** for £/sqft from **HM Land Registry Price Paid** — a light
  per-property lookup for the 10–15 only, cached. Unrecoverable comps → that
  criterion scores "unavailable" (and says so), never guessed; the property is not
  failed (this is a comparison, not a gate).
- **Couple-set weights out of 100** (config, Ticket 1); the total is the weighted sum.
- **Output** — one sortable total per property, a breakdown of which criteria drove
  it, and a **"too close to call"** flag on near-ties (refusing precision the data
  can't justify).
- **v1 boundary.** Only measurable criteria are scored; judgement criteria
  (condition, area feel) are not; the model is **never** used where a real number
  exists. Leave a clean seam for v2 judgement criteria (each later carrying a written
  justification, weighted cautiously).

An input to a human decision, not the decision — its most valuable output is the
disagreement it surfaces (why a loved property scored below a lukewarm one), so the
breakdown must make that legible.

## Out of scope

Model / condition / area-feel scoring (v2); the daily page and finding loop
(Tickets 3–4); making the decision for the couple.

## Validate

1. **Normalisation** scales each criterion to the shortlist; recomputes on
   add/remove.
2. **Weights drive totals** — changing config weights reorders, with no code edit.
3. **Stated reasons** — every criterion score exposes an evidence string, not a bare
   number.
4. **Traceability** — each total decomposes exactly into its weighted contributions.
5. **Too-close-to-call** — near-equal totals are flagged, not ranked confidently.
6. **Sortable** by total and by any individual criterion.
7. **v1 boundary** — no judgement criterion is scored; a numeric criterion is never
   routed through a model.
8. **Disagreement legible** — a loved-below-lukewarm case shows the driving criteria.

## Hygiene gate

`make check` green with tests for new behaviour before every commit — see
[Ticket 0](00-project-setup-go.md).
