# Abode — System Spec

## What it is

A personal tool for a couple buying their first home. Once a day it finds every UK for-sale property matching their hard must-haves and publishes the matches to a simple sortable web page. Separately, on demand, it helps compare a hand-picked shortlist of properties they genuinely like, using an explainable weighted score. It runs itself on free or low-cost infrastructure and, once set up, needs no attention.

## Principles

These are the commitments the design answers to; when a choice is unclear, they break the tie.

**Glanceable beats complete.** The page is for a quick scan, not a deep report. It shows the matching set plainly and gets out of the way.

**Pass or fail, nothing in between.** The must-haves are gates. A property either passes every one and is shown, or fails one and never appears. The finding stage makes no attempt to rank or rate survivors — it presents the matching set plainly. No "best match," no "almost."

**The user sorts; the system doesn't.** Because the finding stage holds no opinion on which survivor is better, the page hands that control over: every meaningful column is sortable so the couple can order the set however they're thinking that day.

**Heavy processing on the fewest items; light processing on the most.** The defining cost principle. Cheap checks run across the whole field and shrink it hard; each successively more expensive step runs only on what survived the cheaper ones. By the time anything heavy happens, the set is as small as it will ever be. Enforced by ordering, not by hoping.

**Honest about what it knows.** Where a requirement has no real dataset, the system uses a clearly-labelled heuristic and says so. Nothing implies more certainty than the data supports. A requirement the system cannot confirm is treated as unmet, never satisfied-by-assumption.

**Explainability over precision (in the shortlist stage).** Any score the system computes must trace back to the evidence that produced it, and the couple set the weights themselves. A weighted score that can't be interrogated is worse than none. Where a ranking is finer than the data can justify, the system says so rather than presenting false precision.

**Config is the only source of truth.** The must-haves and settings live in one editable file. Changing what you want means editing config, not code.

**Uniform parts.** Every data source is the same kind of pluggable component, declaring what it needs, what it produces, and how costly it is. Adding a source is a one-file change; the system orders everything automatically.

**Buy simplicity where it's worth it.** A paid property API is used deliberately to collapse many fragmented data sources into single field-reads, and to keep the listings feed licensed rather than scraped. It is never expected to answer things that aren't property data.

**Runs unattended, fails loudly.** The daily job is a batch task the platform schedules; the code stays a simple wake-work-exit. Any failure surfaces visibly rather than silently producing nothing.

## How requirements become checks

Each must-have is declared in plain language in the config, compiled into a checkable rule bound to a data source, and evaluated per property into pass or fail. Requirements are of two kinds, and which kind a requirement is is a recorded, visible decision.

**Deterministic** — a hard comparison against real structured data: price, bedrooms, bathrooms, EPC rating, fibre availability, inside the commute boundary. Computed, identical every time, and the system pushes every requirement here where data exists. EPC is its own deterministic data point, used directly, not as a stand-in for anything. Where a field is missing from the primary feed, it may be recovered from another source; a recovered value is as authoritative as a feed value, flagged only by source and sanity-checked. Anything checkable by no data fails closed.

**Judgement** — reserved for the irreducibly qualitative, where no dataset gives a clean answer: "interesting area," "feels safe walking home," and the condition of the property. These require a model to form a view from real evidence the way a person would. **In v1 there is no model judgement at all** (see Versioning): area questions are answered by pulling the relevant area data and populating it for survivors, and condition is left to the couple's own eye on the listing. Judgement proper — a model reasoning over evidence, always labelled as judgement and always sitting beneath a deterministic guardrail where one exists — is a v2 capability.

## The evaluation algorithm (v1)

Ordered so that the lightest work touches the most properties and the heaviest touches the fewest.

**Step 1 — Build the catchment (free, local).**
Compute, or reuse a cached, polygon of everywhere within the target commute time of the work anchor. It is both the geographic boundary of the search and a free local commute check. Most of the UK is excluded here at no cost.

**Step 2 — Paid API property search (one feed, broad).**
Search all properties inside the catchment using every constraint the paid API supports directly — price, bedrooms, bathrooms — and request every other useful field the API can return in the same results (floor area, EPC, tenure, type, coordinates, photos, agent details, and so on). This returns the candidate set already filtered on the cheap structured constraints, with most attributes populated from the single feed.

**Step 3 — Fill gaps from other sources (light, across the candidates).**
For fields the primary feed didn't provide, capture them from secondary API sources: fibre/internet availability, outdoor-space and park proximity, and similar — light field-reads and lookups, applied across the candidate set, each cached. Missing-and-unrecoverable fields are marked, to fail closed if their gate can't be met.

**Step 4 — Populate area data for survivors (light).**
For the surviving set, pull the relevant area-level data by area (e.g. crime statistics for the safety requirement) and populate it onto the properties. In v1 this is a straight data pull and threshold check — anywhere breaching the safety threshold fails — not a model judgement. Pulling by area and applying to each property in it keeps the work small.

The properties passing every step are the day's matches. By construction the only paid breadth is the single API search; everything after is light lookups on a shrinking set.

## Versioning

**v1 — the whole loop, deterministic only, page-delivered.** Catchment → paid API search → secondary-source gap-fill → area-data population, producing the daily sortable page, plus the shortlist comparison for measurable criteria. No email, no favourites, no OCR, no model judgement anywhere. This is the complete, useful system on its own: it finds and presents the matches on a page the couple open when they want, and they judge condition and area feel with their own eyes, as they would anyway.

**v2 — the additions, once v1 is solid.** None require reworking v1, thanks to the uniform-component structure:

- **The nudge email.** A tiny daily email — count and a link to the page — so the couple don't have to remember to check. The page stays the content; the email is just the signal to look.
- **Favourites.** A silent, unexplained personal pick layer: tap to mark the ones worth booking a viewing for, no justification asked. Its design hinges on whether it's solo or shared, which decides where its state lives; deferred until that's settled.
- **Condition by model.** A model assesses property condition from the listing photos — the heaviest step, run on the fewest items (only properties that passed everything else), labelled as judgement, never fact. Staged but explicitly not built first.
- **OCR recovery of missing measurable fields.** Where the feed lacks bathroom count or square footage, recover them from the floor plan. In v1 a missing such field simply fails closed.
- **Model judgement of area questions.** "Interesting" and "feels safe" upgraded from v1's raw area-data population to a model reasoning over that data, still beneath the deterministic guardrails.
- **Judgement criteria in the shortlist.** Condition and area feel join the shortlist comparison once model judgement exists, each carrying a written justification, weighted cautiously by default.

## Data strategy

A paid property API is the spine, returning most attributes as single field-reads — price, beds, type, tenure, floor area, EPC, council tax, internet speed, sold-price comps, crime — and keeping the feed licensed rather than scraped. Gaps are filled from secondary sources. Known weaknesses are handled by the stated principles: coverage gaps fail closed, floor area is treated as "as advertised" rather than ground truth. In v1, fields that live only in the images (recovered by OCR) and qualitative judgements (condition, area feel) are out of scope and handled respectively by failing closed or by the couple's own eye; both become system capabilities in v2.

Three requirements are not property data and stay external regardless of spend: the commute (an isochrone/journey source, mostly absorbed into the catchment), area safety, and area feel.

## Delivery surfaces

**The page (v1)** is one static HTML file, regenerated each run and served at a stable public address. It shows the current matching set as a table — one row per property with a photo — whose columns (price, floor area, bedrooms, bathrooms, commute time, postcode, link out) are each clickable to sort ascending or descending. The couple reorder the set to suit; the system imposes no order beyond a sensible default. Public is acceptable because it only shows already-public listing data. In v1 the couple open the page themselves; the nudge email that points them to it is a v2 addition.

## The shortlist comparison (separate stage)

Distinct from the daily finding, this is a decision aid for the 10-15 properties the couple actually like — not the full survivor set, because weighted scoring is only meaningful on a small, genuinely-considered group. The couple mark the shortlist; the system scores it.

Measurable criteria (price, floor area, commute, price-per-square-foot vs comps) are scored by transparent formula, normalised relative to the shortlist, each with a stated reason ("8.5/10 — 11% below local median £/sqft"). The couple set the weights out of 100, exactly as in a manual spreadsheet. The output is a single sortable metric plus a full breakdown of which criteria drove each total and an explicit flag on any pair too close to call. In v1 only the measurable criteria are scored; judgement criteria (condition, area feel) join the comparison in v2 when model judgement exists, each then carrying a written justification, and weighted cautiously by default. The model is never used to score a criterion that has a real number.

This stage is explicitly an input to a human decision, not the decision. Its most valuable output is often the disagreement it surfaces — why a loved property scored below a lukewarm one — because that question makes the couple articulate what they actually value. The tool's job is to get them to the right front doors faster; the deciding happens at the doors.

## Scope

In scope (v1): UK for-sale property; the couple's must-haves as pass/fail gates, checked deterministically; a sortable daily page of all survivors; an on-demand shortlist comparison over measurable criteria.

v2: a daily nudge email pointing to the page; a favourites layer for marking viewings; model judgement of property condition from photos; OCR recovery of missing bathroom count and square footage; model judgement of area questions; and these judgement criteria joining the shortlist comparison.

Deferred, not designed out (the uniform-component structure leaves room without rework): change-tracking between days; maps; and messaging-app delivery.

---

## Implementation tickets

The v1 build is a setup ticket plus five feature tickets, ordered by dependency. Each has a clear goal and an explicit way to validate it.

0. [Go project setup](tickets/00-project-setup-go.md) — the module, layout, toolchain, tests, lint, and CI the feature tickets build into.
1. [Config & requirement compiler](tickets/01-config-and-requirement-compiler.md) — the single source of truth and the pass/fail gate engine.
2. [Pluggable data-source framework](tickets/02-data-source-framework.md) — uniform components and cheap-to-expensive orchestration.
3. [The finding pipeline](tickets/03-finding-pipeline.md) — catchment → paid search → gap-fill → area data, producing the daily matches.
4. [Daily batch runner & static page](tickets/04-daily-batch-and-page.md) — unattended wake-work-exit and the sortable HTML page.
5. [Shortlist comparison](tickets/05-shortlist-comparison.md) — explainable weighted scoring over measurable criteria.
