# Ticket 07 — Daily batch entrypoint & static sortable page

## Goal

Wire the pipeline into a once-a-day wake-work-exit batch job that the platform schedules, and have each run regenerate one static HTML page — the day's matching set as a sortable table — served at a stable public address. Runs unattended; fails loudly.

## Scope

### Batch entrypoint
- A single entrypoint that, per run: loads config (Ticket 01), builds the catchment (03), runs the paid search (04), gap-fill (05) and area-data (06) via the orchestrator (02), and writes the page. Then exits. **No daemon, no internal scheduler** — the platform's scheduler invokes it.
- **Fails loudly:** any step failure surfaces visibly (non-zero exit + visible error) rather than silently producing nothing or an empty page. A run that cannot complete must not overwrite a good page with a blank one without a clear signal.
- Runs on free / low-cost infrastructure; once set up, needs no attention.

### The page
- One **static HTML file**, regenerated each run, served at a **stable public address**. Public is acceptable — it shows only already-public listing data.
- The matching set as a **table**, one row per property **with a photo**, columns: price, floor area, bedrooms, bathrooms, commute time, postcode, link out.
- **Every meaningful column is clickable to sort** ascending/descending — the user sorts, the system doesn't. The system imposes only a sensible **default order**, no "best match" / ranking.
- **Glanceable:** plain presentation of the matching set, no scores, no ratings, no "almost" rows.
- Fields resting on a heuristic/recovered source are labelled honestly where shown.

## Out of scope

- The nudge email (v2), favourites (v2), maps / change-tracking (deferred).
- The shortlist comparison (Ticket 08) — separate surface.

## How to validate

- Invoking the entrypoint once produces a fresh HTML file containing exactly the survivors, then exits 0; invoking it again regenerates it.
- Each row shows a photo and the seven columns; clicking each column header reorders the visible set asc/desc (sortable verified for every column).
- Default order is applied on load and is a neutral/sensible one, not a quality ranking.
- A forced failure in any pipeline step makes the run exit non-zero with a visible error and does **not** silently publish an empty page.
- The output is a single self-contained static file servable at a fixed path/URL.
