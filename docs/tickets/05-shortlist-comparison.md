# Ticket 5 — Shortlist comparison

**Depends on:** Ticket 1 (config holds the weights), Ticket 2 (data-source
framework, used to pull sold-price comps), and the property record shape from
Ticket 3 (incl. `commute_time`). Otherwise a separate stage — not part of the
daily finding loop.

## Goal

Build the on-demand decision aid: for the 10–15 properties the couple actually
like, score the measurable criteria by transparent formula, normalised across the
shortlist, with the couple's own weights — and present a single sortable metric
plus a full, interrogable breakdown. Explainability over precision: every number
traces back to the evidence that produced it.

Deliver:

- **A shortlist input.** The couple mark which properties form the shortlist
  (10–15). The system scores only this hand-picked group — weighted scoring is
  meaningful only on a small, genuinely-considered set, never the full survivor
  set.
- **Measurable-criteria scoring.** Score **price, floor area, commute, and
  price-per-square-foot vs comps**, each by a transparent formula, **normalised
  relative to the shortlist** (best-in-shortlist anchors the scale). Each criterion
  score carries a **stated reason** in evidence terms (e.g. "8.5/10 — 11% below
  local median £/sqft").
- **Comp data for £/sqft.** The £/sqft-vs-comps criterion needs sold-price comps,
  which the SPEC's data strategy assigns to the paid API. Pull them as a light
  per-property lookup **for the shortlisted set only** (10–15 items, so cheap),
  through the Ticket 2 data-source framework, cached. A property whose comps are
  unrecoverable scores that one criterion as unavailable (and says so) rather than
  guessing — it does not fail the property, since the shortlist is a comparison,
  not a gate.
- **Couple-set weights, out of 100.** Weights live in config (Ticket 1), set by
  the couple exactly as in a manual spreadsheet. The total is the weighted sum.
- **Output.** A single **sortable** total metric per property, **plus** a full
  breakdown showing which criteria drove each total, **plus** an explicit
  **"too close to call"** flag on any pair whose totals are within a small margin
  — the tool refuses false precision the data can't justify.
- **v1 boundary, enforced.** Only measurable criteria are scored. Judgement
  criteria (condition, area feel) are **not** scored in v1, and the model is
  **never** used to score a criterion that has a real number. Leave a clean seam
  for v2 judgement criteria (each later carrying a written justification, weighted
  cautiously) without building them now.

This stage is an input to a human decision, not the decision. Its most valuable
output is the disagreement it surfaces — why a loved property scored below a
lukewarm one — so the breakdown must make that legible.

## Out of scope

- Any model judgement / condition / area-feel scoring (v2).
- The daily page and finding loop (Tickets 3–4).
- Making the decision for the couple — it only informs.

## How to validate

1. **Normalisation.** Score a fixed shortlist and assert each criterion is scaled
   relative to the shortlist (the best value on a criterion anchors the top of
   that criterion's scale), recomputing correctly when a property is added or
   removed.
2. **Weights drive totals.** Change the weights in config and assert the totals
   and resulting order change accordingly — with no code edit.
3. **Stated reasons.** Assert every criterion score exposes an evidence-based
   reason string (value + comparison to the shortlist/comps), not a bare number.
4. **Traceability.** Assert each total decomposes exactly into its weighted
   criterion contributions — the breakdown sums to the total, so any score can be
   interrogated.
5. **Too-close-to-call.** Construct two properties with near-equal totals and
   assert the pair is flagged rather than presented as a confident ranking.
6. **Sortable output.** Assert the result can be sorted by total and by any
   individual criterion.
7. **v1 boundary.** Assert no judgement criterion is scored, and that a criterion
   backed by a real number is never routed through a model.
8. **Disagreement is legible.** Given a "loved" property scoring below a
   "lukewarm" one, assert the breakdown makes the driving criteria explicit so the
   couple can see why.
