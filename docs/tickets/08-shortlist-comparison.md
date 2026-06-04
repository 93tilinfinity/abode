# Ticket 08 — Shortlist comparison stage

## Goal

Provide the separate, on-demand decision aid: for the 10–15 properties the couple genuinely like, score the **measurable** criteria by transparent formula, normalised relative to the shortlist, weighted by the couple's own weights, and present a single sortable metric plus a full, interrogable breakdown. Explainability over precision.

## Scope

- A distinct stage from the daily finding (Ticket 07) — invoked **on demand**, over a couple-marked shortlist of ~10–15 properties, not the full survivor set.
- v1 measurable criteria scored: **price, floor area, commute, price-per-square-foot vs comps**. Each scored by a transparent formula, **normalised relative to the shortlist**, and emitted with a **stated reason** (e.g. `"8.5/10 — 11% below local median £/sqft"`).
- **Weights are config / couple-set**, out of 100, exactly as a manual spreadsheet — the couple set them, not the system (ties to Ticket 01 config).
- Output: a **single sortable metric** per property **plus a full breakdown** of which criteria drove each total. Every number **traces back to its evidence**.
- An explicit **"too close to call" flag** on any pair whose totals are nearer than the data can justify — say so rather than present false precision.
- The model is **never** used to score a criterion that has a real number. **No judgement criteria in v1** — condition and area feel join in v2 when model judgement exists.
- Framed as an **input to a human decision**, not the decision; surfacing disagreement (a loved property scoring below a lukewarm one) is a feature, not a bug.

## Out of scope

- Judgement criteria (condition, area feel) and their written justifications — **v2**.
- Any auto-decision or recommendation beyond the scored, explained ordering.

## How to validate

- Given a shortlist and weights, each property gets a total and a per-criterion breakdown; changing a weight in config changes totals predictably and with no code edit.
- Each criterion score carries a human-readable reason tracing to its evidence (e.g. its £/sqft vs the shortlist/comps median).
- Scores are normalised across the shortlist (best/worst anchor sensibly within the marked set).
- Two near-tied properties trigger the "too close to call" flag; clearly separated ones do not.
- No criterion with a real number is scored by a model; attempting to add a judgement criterion in v1 is rejected/absent.
- The breakdown lets a user see exactly why a loved property ranked below a lukewarm one.
