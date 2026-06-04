# Plan — Ticket 4: Daily batch runner & static page

Implementation plan for [Ticket 4](../tickets/04-daily-batch-and-page.md). It runs
the [Ticket 3](ticket-03-finding-pipeline.md) pipeline unattended once a day and
publishes the matches as one static, sortable HTML page — wake, work, write,
exit; fail loudly.

## 1. What this ticket owns

The batch entrypoint, the schedule, the static page, and the publish/fail
behaviour. It does **not** own the pipeline (Ticket 3) or the shortlist (Ticket 5).

**Platform (Decision):** **GitHub Actions + GitHub Pages** — the repo already
lives on GitHub, it's free, it runs the Go binary, holds the API keys as Actions
secrets, and deploys the static page to a stable public URL.

> The scheduled workflow, the Pages publish, and the fail-loudly wiring are
> **already stood up as a heartbeat in [Ticket 0](../tickets/00-project-setup-go.md)**.
> This ticket inherits that working loop and only makes the page *meaningful* —
> replacing the heartbeat with the sortable listings table below.

## 2. The batch entrypoint — `cmd/abode-daily`

A thin `main` that does exactly: load `abode.config.yaml` → run the Ticket 3
pipeline → render the page → write it to the publish location → exit. No
long-running process, no state between runs beyond the in-run cache. Any error
returned up the stack causes a **non-zero exit** (Go idiom: errors wrapped with
`%w` and surfaced at `main`).

## 3. Scheduling — GitHub Actions

A scheduled workflow (`.github/workflows/daily.yml`):

- **`on: schedule: cron`** — runs early-morning UK so the page is fresh by
  breakfast. Cron in Actions is **UTC**, so e.g. `0 5 * * *` ≈ 06:00 BST / 05:00
  GMT; the one-hour DST drift is harmless for a daily page (documented, not
  fought). Also `workflow_dispatch` for manual runs.
- **Steps:** checkout → set up Go (version from `go.mod`) → `go run ./cmd/abode-daily`
  → publish to Pages.
- **Secrets (Actions secrets, never committed):** `PROPERTYDATA_KEY`,
  `GOOGLE_ROUTES_KEY`, `EPC_API_KEY` (Open Data Communities). The config file
  holds settings; secrets hold credentials.

## 4. The page

One static HTML file, regenerated each run, served at a stable public URL
(public is fine — only already-public listing data).

- **One row per property, with a photo** (the listing's primary image, hotlinked
  from the feed's `image_url`).
- **Columns:** price, floor area, bedrooms, bathrooms, commute time, postcode,
  link out. **Floor area is labelled "as advertised"** (honest-about-what-it-
  knows), and **recovered/heuristic values carry a subtle marker/tooltip**
  (provenance, per Ticket 2).
- **Every column is clickable to sort** ascending/descending, client-side, with
  numeric columns sorted numerically. Implemented in **vanilla JS** (no
  framework) so the page is a single self-contained file.
- **Default order: newest listed first** (see dependency note below). The system
  imposes no opinion beyond this initial order — the couple re-sort at will.
- **Honest empty state:** a run with zero matches renders a clean "0 matches
  today" page (this is success, not failure).

> **Dependency note:** "newest listed first" needs a per-listing *listed date*
> from PropertyData. Confirming that field is part of the Ticket 3 free-trial
> check; if it's unavailable, the default falls back to commute-time ascending,
> with listing date dropped from sorting.

## 5. Fail loudly (your decision: GitHub Actions email)

- Any pipeline/source error (after Ticket 2's retries) → the binary exits
  non-zero → the workflow **fails** → **GitHub Actions sends its built-in failure
  email**. No extra alerting to build.
- **The publish step only runs on success** (`if: success()`), so a failed run
  **never overwrites yesterday's good page** — Pages keeps serving the last
  successful deploy.
- **Zero matches ≠ failure:** the binary exits 0 and publishes the honest empty
  page; only an actual error blocks the deploy.

## 6. Decisions captured during the interview

| Decision | Choice |
|----------|--------|
| Platform | **GitHub Actions (cron) + GitHub Pages** |
| Failure alerting | **GitHub Actions' built-in failure email**; publish only on success so yesterday's page survives |
| Page privacy | **Public URL** (only public listing data) |
| Default sort | **Newest listed first** (fallback: commute ascending if listed-date unavailable) |
| Run time | Early-morning UK (~06:00; cron is UTC, DST drift accepted) |

## 7. Hand-offs

- **Ticket 3** supplies the matching set (with provenance flags) this renders.
- **Ticket 2** supplies the fail-loudly exit semantics this relies on.
- **v2** adds the nudge email pointing at this page; the page stays the content.

## 8. Validation (acceptance for Ticket 4)

1. **Wake-work-exit.** `cmd/abode-daily` against a mock pipeline loads config,
   runs it, writes the HTML, exits 0, holds no process open.
2. **Page contents.** A known matching set renders one row per property, each
   with a photo and all required columns; floor area shows an "as advertised"
   label; a recovered value shows a subtle provenance marker.
3. **Sortable + default.** Each column header sorts asc/desc (numeric columns
   numerically); the page loads in newest-listed-first order (or the documented
   fallback).
4. **Stable address.** Two runs publish to the same public URL, reflecting the
   latest successful run.
5. **Zero vs failure.** A zero-match run publishes a clean "0 matches" page and
   exits 0. A forced pipeline error exits non-zero, the workflow fails, the
   failure email fires, and the **previous page is left intact** (publish skipped).
6. **Secrets, not config.** API keys are read from Actions secrets/env, never
   from the committed config or repo.
7. **Schedule present.** The repo contains the scheduled workflow and a short
   runbook so the daily run is reproducible.
