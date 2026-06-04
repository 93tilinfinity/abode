# Ticket 4 — Daily batch runner & static sortable page

**Depends on:** Ticket 3 (the matching set this renders and schedules).

> Plan: [`plans/ticket-04-daily-batch-and-page.md`](../plans/ticket-04-daily-batch-and-page.md) — GitHub Actions + Pages, fail-loudly via the Actions failure email, public page, default newest-first.

## Goal

Run the pipeline unattended once a day and publish its matches as one static,
sortable HTML page at a stable public URL — wake, work, write, exit. Fail loudly;
never silently produce nothing.

Deliver:

- **A batch entry point** — wake-work-exit: load config → run the Ticket 3 pipeline
  → render → write to the published location → exit. No long-running process, no
  state held between runs beyond in-run caches.
- **Platform scheduling** on free/low-cost infra (cron / CI schedule). The code
  stays a plain batch task; the schedule lives in repo'd platform config.
- **The page** — one static HTML file, regenerated each run: one row per property
  **with a photo**; columns **price, floor area, beds, baths, commute time,
  postcode, link out**; every column **click-sortable** (numeric columns sort
  numerically), with a sensible default the couple can override. Floor area shown
  **as advertised** and labelled (honest about what it knows). Public is fine — only
  already-public listing data.
- **Fail loudly.** Any failure (pipeline error, garbage feed, write failure) → a
  non-zero exit and a surfaced error, **not** a blank page over yesterday's good
  one. Distinguish "ran, zero matches today" (legitimate, page says so, exit 0) from
  "the run broke" (loud failure, last good page left intact).

## Out of scope

Nudge email (v2); favourites / interactivity beyond sorting (v2+); shortlist
(Ticket 5); maps, change-tracking between days (deferred).

## Validate

1. **Wake-work-exit.** Against a mock pipeline: loads config, runs it, writes the
   HTML, exits 0, holds no process open.
2. **Page contents.** A known matching set → one row per property with a photo and
   all required columns.
3. **Sortable.** Each header sorts asc/desc (numeric numerically), with a defined
   default order on load.
4. **Stable address.** Two runs write the same path, reflecting the latest run.
5. **Zero vs failure.** An empty-but-valid result → clean "0 matches" page, exit 0.
   A forced pipeline error → non-zero exit, surfaced, previous page **not** clobbered.
6. **Scheduling documented.** The repo holds the schedule definition and a runbook
   note so the daily run is reproducible.

## Hygiene gate

`make check` green with tests for new behaviour before every commit — see
[Ticket 0](00-project-setup-go.md).
