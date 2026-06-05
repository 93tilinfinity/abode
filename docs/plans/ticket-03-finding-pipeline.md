# Plan — Ticket 3: The finding pipeline

Implementation plan for [Ticket 3](../tickets/03-finding-pipeline.md). One linear
Go pipeline that scrapes Rightmove, runs the [Ticket 1](ticket-01-config-and-compiler.md)
gates cheapest-first, and emits the day's matches. Source choices and the
"no framework" decision are in [Decision 0001](../decisions/0001-data-providers.md).

## 1. What this ticket owns

The scrape and the five gate steps, in a fixed hand-written order — **not** a
plugin framework, the rules, the page, or the shortlist. Suggested layout under
`internal/`:

| Package | Responsibility |
|---|---|
| `rightmove` | build the `/api/_search` URL, fetch, paginate, price-band tile, parse `properties[]` |
| `pipeline` | run the gates in order over the candidate set, dropping failures early |
| `broadband` | Ofcom postcode → max download Mbps |
| `greenspace` | OSM Overpass / OS Open Greenspace → nearest-park metres |
| `crime` | police.uk → violent/sexual count within ~1 mi, last 12 mo (per area) |
| `journey` | Google Routes (transit) → minutes + non-walking mode count |
| `epc` | Open Data Communities → floor area (m²) + EPC rating (page only, not a gate) |
| `geocode` | postcodes.io → lat/long where needed |

These are just packages with ordinary functions; `pipeline` calls them in
sequence. No `Source` interface, no ordering engine, no provenance generics.

## 2. The scrape (step 1, the only broad fetch)

- **Endpoint:** `https://www.rightmove.co.uk/api/_search` (JSON), `channel=BUY`.
- **Server-side filters from config:** `locationIdentifier` (one home centre,
  resolved once and stored in config), `radius` (miles), `minPrice`/`maxPrice`,
  `minBedrooms`, `numberOfPropertiesPerPage` (page size), `index` (offset).
- **Pagination:** loop `index` in page-size steps until the returned page is short
  or the result total is reached.
- **Beat the ~1,050 cap by price-band tiling:** when a single radius query's
  `resultCount` would exceed the cap, split the price range into bands (e.g. £50k
  steps between `minPrice` and `maxPrice`), fetch each band fully, and **merge,
  de-duplicating by listing id**. This guarantees completeness.
- **Parse** each `properties[]` entry → `{ id, price, bedrooms, bathrooms,
  propertySubType, latitude, longitude, displayAddress, propertyUrl, thumbnail }`.
- **Bathrooms in code:** Rightmove has no `minBathrooms` filter, so apply
  `toilet_count ≥ 2` here. `bathrooms == null` → **fail closed** (matches config
  `on_missing: fail`). Note in the page that "bathrooms" is Rightmove's count and
  may differ from a strict toilet/cloakroom count (OCR recovery is v2).
- **Be a polite client:** real browser `User-Agent`, low request rate, retry with
  backoff; if the JSON shape changes, return an error → the run **fails loudly**.

> **Validation task:** confirm the live `/api/_search` field names against one real
> response **locally** (not from CI), and lock a small saved fixture for parser
> tests. Do not hammer Rightmove from the shared CI runner.

## 3. The gates (steps 2–5, cheapest first)

Applied in code immediately after each field lands, so the set shrinks before the
next (costlier) step:

1. **price ≤ £625k, beds ≥ 2** — already filtered server-side in step 1; re-assert
   in code as a guard.
2. **toilets ≥ 2** — from the scraped `bathrooms` (step 1), fail closed if null.
3. **broadband** — `max_download_mbps ≥ 300`, Ofcom postcode-level. Unknown →
   fail closed. One lookup per unique postcode, reused.
4. **outdoor_space** — `garden OR park ≤ 800 m`. Use the listing garden signal
   first; only call `greenspace` when it's false/unknown (short-circuit the OR).
5. **safe_area** — police.uk count within ~1 mi over 12 months (12 monthly
   snapshots summed) vs the config threshold. Fetched **once per area**, reused
   for nearby properties (in-run `map`). Absolute, tunable cutoff — **calibrate
   against known areas** during the trial; the placeholder is a starting point.
6. **commute_time ≤ 50 and commute_modes ≤ 2** — Google Routes door-to-door
   transit, depart **Tue 08:00** → work anchor. **Runs last**, only on survivors,
   minimising paid journeys. Read total minutes and count distinct non-walking leg
   modes. Populate `commute_time` for the page column / v2 £-per-min sorting.

`epc` (floor area + rating) is pulled across survivors for the page / £/sqft and
is **not** a gate — a missing EPC leaves floor area blank.

## 4. The matching set

Properties passing every gate are the day's matches, emitted as records for
Ticket 4 to render. Recovered fields carry a `source` string and are
sanity-checked (out-of-range → not recovered). The only broad fetch is step 1;
everything after is light lookups on a shrinking set, journey deliberately last.

## 5. Decisions captured

| Decision | Choice | Consequence |
|----------|--------|-------------|
| Listings source | **Scrape Rightmove `/api/_search`** (Decision 0001) | No paid provider; completeness via pagination + price-band tiling; fails loudly if the shape changes. |
| Bathrooms | **Post-filter the returned field in code** | No server filter exists; null → fail closed. |
| Search radius | **~40 miles around the home centre** | Covers fast-rail commuter towns; tune against journey volume. |
| Availability | **Include under-offer / Sold-STC** | Bigger set → more journeys (watch the Routes free tier); page may show under-offer homes. |
| Safety measure | **Raw police.uk count within ~1 mi over 12 mo, absolute tunable threshold** | Recomputes each run; not population-adjusted — calibrate in the trial. |

## 6. Hand-offs

- **Ticket 1** rules are evaluated by the pipeline as each field lands.
- **Ticket 4** renders the matches (floor area shown *as advertised*) and owns
  scheduling; it leaves yesterday's page up if the run fails.
- **Ticket 5 (v2)** reuses `commute_time`, floor area + Land Registry comps for
  £/sqft.

## 7. Validation

See the acceptance checklist in [Ticket 3](../tickets/03-finding-pipeline.md):
complete scrape (pagination + tiling), server-vs-code filters, shrink-early with
journey last, fail-closed gates, the outdoor OR short-circuit, crime-by-area
threshold, commute read, EPC-not-a-gate, and the end-to-end set.
