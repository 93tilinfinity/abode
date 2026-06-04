# Ticket 4 — Daily batch runner & static sortable page

**Depends on:** Ticket 3 (produces the matching set this ticket renders and schedules).

## Goal

Run the finding pipeline unattended once a day and publish its matches as one
static, sortable HTML page at a stable public address — wake, do the work, write
the page, exit. Fail loudly if anything breaks; never silently produce nothing.

Deliver:

- **A batch entry point.** A simple wake-work-exit job: load config, run the
  Ticket 3 pipeline, render the page, write it to the published location, exit.
  No long-running process, no state held between runs beyond caches.
- **Platform scheduling.** The job is scheduled by the platform (cron / scheduled
  task / CI schedule) on free or low-cost infra. The code stays a plain batch
  task; the schedule lives in platform config, documented in the repo.
- **The page — one static HTML file**, regenerated each run, served at a stable
  public URL. It shows the current matching set as a table: one row per property
  **with a photo**, columns **price, floor area, bedrooms, bathrooms, commute
  time, postcode, link out**. Every column is **clickable to sort ascending or
  descending**, client-side. A sensible default order, but the system imposes no
  opinion beyond it — the couple reorder to suit. Public is acceptable because it
  shows only already-public listing data.
- **Fail-loudly behaviour.** Any failure (pipeline error, empty/garbage feed,
  write failure) surfaces visibly — a non-zero exit and a surfaced error — rather
  than overwriting yesterday's good page with a blank one. Distinguish "ran, zero
  matches today" (legitimate, page says so) from "the run broke" (loud failure,
  last good page left intact).

## Out of scope

- The nudge email pointing to the page (v2).
- Favourites / any interactivity beyond column sorting (v2+).
- The shortlist comparison (Ticket 5).
- Maps, change-tracking between days (deferred).

## How to validate

1. **Wake-work-exit.** Invoke the batch entry point against a mock pipeline and
   assert it loads config, runs the pipeline, writes the HTML file, and exits 0 —
   holding no process open afterward.
2. **Page contents.** Render a known matching set and assert the HTML has one row
   per property, each with a photo and all required columns (price, floor area,
   beds, baths, commute time, postcode, link out).
3. **Sortable columns.** Load the rendered page and assert clicking each column
   header reorders the rows ascending then descending (numeric columns sort
   numerically, not lexically), with a defined default order on load.
4. **Stable address & overwrite.** Run twice and assert the page is written to
   the same stable path and reflects the latest run.
5. **Zero matches vs failure.** Run with an empty-but-valid result and assert the
   page renders cleanly saying zero matches (exit 0). Then force a pipeline error
   and assert the job exits non-zero, surfaces the error, and does **not** clobber
   the previous good page.
6. **Scheduling documented.** Assert the repo contains the schedule definition and
   a runbook note so the daily run is reproducible on the chosen infra.
