# Decision 0001 — Data sources & pipeline shape

**Status:** Accepted · **Date:** 2026-06-04 · **Affects:** SPEC ("Data strategy",
principles, the v1 algorithm), Ticket 3. **Removes** the separate data-source
framework (old Ticket 2).

## Context

The spec originally assumed two things we've since reversed, both in the name of
keeping this a *simple personal tool*:

1. a **paid property API** as the licensed "spine" for listings, and
2. a **pluggable source framework** that derives run order by topological sort
   (cost classes, provenance generics, a budget guard, a cache interface).

Research into the 2025–26 UK market killed (1); the goal of "it shouldn't be a
complex repo" killed (2). v1 now **scrapes one portal (Rightmove)** for listings
and runs everything else as a **plain linear pipeline**.

## Decision 1 — Listings come from scraping Rightmove's search API

- **No UK portal sells a self-serve listings API.** Rightmove's ADF/RTDF are
  agent-*upload* feeds; Zoopla closed its public API; OnTheMarket is upload-only;
  Nestoria is dead. The paid aggregators (PropertyData, PaTMa, Homedata) each add
  cost and integration *and* carry gaps — no native price/beds filter, no photos,
  or undocumented coverage — which isn't worth it for a single-user tool.
- **v1 calls Rightmove's own search endpoint** —
  `https://www.rightmove.co.uk/api/_search`, the JSON XHR the site itself uses —
  filtering **server-side** on `locationIdentifier` (one home centre, resolved
  once and hard-coded), `radius`, `minPrice`/`maxPrice`, `minBedrooms`. It returns
  structured `properties[]`: price, beds, **bathrooms**, type, coordinates,
  address, listing URL, thumbnail.
- **Bathrooms:** Rightmove has **no bathrooms *filter*** (the "filter" people see
  on the site is a browser extension). But the `bathrooms` value *is* returned per
  listing, so we **post-filter it in code**, failing closed when it's null — which
  matches the config's `toilet_count … on_missing: fail`.
- **Completeness via pagination + price tiling.** A single query is capped at
  ~1,050 results. To get *all* matches we **paginate** (`index` in page-size
  steps) and, where one radius+price query would exceed the cap, **tile by price
  band** (e.g. £50k steps) and merge, de-duplicating by listing id.
- **This is scraping, not a licensed feed.** A private, undocumented,
  unauthenticated endpoint used against Rightmove's ToU — accepted deliberately
  for a personal tool. Be a polite client: real browser `User-Agent`, low request
  rate, retry with backoff, and **fail loudly** if the response shape changes
  (don't publish a silently-wrong page).

## Decision 2 — A linear pipeline, not a source framework

The old Ticket 2 — "uniform pluggable source, ordering engine, `Value[T]`
provenance, `CostClass`, `BudgetGuard`, `Cache` interface" — is **removed**. For a
fixed, known set of checks the order is obvious and hand-written:

```
scrape Rightmove  (price + beds server-side; bathrooms in code)
  → broadband gate        (postcode lookup)
  → outdoor-space gate    (listing garden flag  OR  park ≤ 800 m)
  → crime gate            (by area, fetched once per area)
  → commute gate          (Google Routes, last — the most expensive step)
  → matches
```

Each step drops failures before the next runs, so the costly per-property journey
lookups see the smallest possible set ("heavy work on the fewest items" — kept;
the machinery to *derive* that order — dropped). What we drop and what replaces
it:

| Dropped abstraction | Replaced by |
|---|---|
| Topological ordering engine | one hand-written sequence (above) |
| `Value[T]` provenance generics | a plain `source string` on recovered fields |
| `CostClass` + cost-based tie-breaks | the order is already cheapest-first by hand |
| `BudgetGuard` (paid-spend cap) | nothing — no paid listings feed to cap |
| `Cache` interface + impl | an in-run `map` for per-area lookups |
| Cross-run persistence | none — recompute each run |

In-run memoisation (fetch each area's crime once, reuse for nearby properties) is
kept, because it's just a map.

## Secondary data — free and authoritative (unchanged)

| Field / need | Source | Cost |
|---|---|---|
| For-sale listings (price, beds, **baths**, type, coords, URL, thumbnail) | Rightmove `/api/_search` | free (scraped) |
| Crime (safety gate) | police.uk Data API | free |
| EPC rating + **floor area (m²)** (page column, £/sqft) | Open Data Communities EPC API | free |
| Parks / greenspace (outdoor-space branch) | OSM Overpass / OS Open Greenspace | free |
| Fibre / broadband speed | Ofcom Connected Nations (postcode-level) | free |
| Sold-price comps (£/sqft, v2 shortlist) | HM Land Registry Price Paid | free |
| Geocoding (postcode → lat/long) | postcodes.io | free |
| Commute time + modes (door-to-door) | Google Routes API (transit) | free tier (~10k/mo), then paid |

## Commute — radius + Google Routes (no isochrone)

The straight-line **radius** is already the Rightmove search filter (the coarse
geographic cut). The real gate is a **Google Routes door-to-door transit journey**
per surviving property (depart Tue 08:00 → work anchor): read total minutes
(≤ 50) and count non-walking leg modes (≤ 2). Chosen over a computed isochrone
(no free hosted transit isochrone exists) and over TfL alone (which weakens for
fast-rail commuter towns). **TfL Unified API is the free fallback** if Google
billing/limits ever bite.

## Consequences

- **Build set shrinks to four:** 0 Go setup · 1 config + gates · **3 the finding
  pipeline** (now also owns what was Ticket 2) · 4 daily batch + page. (5 shortlist
  stays deferred to v2.) The Ticket 2 doc/plan are deleted.
- **No paid provider** → no credit budget, no spend logging, no PropertyData
  trial. One dependency fewer to integrate and reason about.
- **Google still needs a billing account** (free tier) for Routes; journeys run
  only on the final survivor set, so the ~10k/mo free tier should be ample.
- **Two requirements degrade gracefully, not perfectly:** fibre is postcode-level
  (fails closed where unknown); council-tax band has no free source (out of scope).
- **The one real risk is Rightmove changing or blocking the endpoint** — handled
  by failing loudly: the daily job exits non-zero, the deploy is skipped, and
  yesterday's page stays up.

## Key references

- Rightmove has no public listings API: https://api-docs.rightmove.co.uk/apis
- How Rightmove's search/listing JSON is structured (2026): https://scrapfly.io/blog/posts/how-to-scrape-rightmove · https://scrape.do/blog/rightmove-scraping/
- The on-site "bathroom filter" is a browser extension, not Rightmove: https://chromewebstore.google.com/detail/bathroom-filters-for-righ/doimjlepfkdblpncdddjpelbggbkdhmo
- police.uk Data API: https://data.police.uk/docs/
- EPC (rating + floor area): https://epc.opendatacommunities.org/docs/api/domestic
- HM Land Registry Price Paid: https://landregistry.data.gov.uk/
- OS Open Greenspace: https://www.ordnancesurvey.co.uk/products/os-open-greenspace · OSM Overpass: https://overpass-api.de/
- postcodes.io (geocoding): https://postcodes.io/
- Google Routes API (transit + billing): https://developers.google.com/maps/documentation/routes/transit-route · https://developers.google.com/maps/documentation/routes/usage-and-billing
- TfL Unified API (free fallback): https://tfl.gov.uk/info-for/open-data-users/api-documentation
- Ofcom Connected Nations: https://www.ofcom.org.uk/phones-and-broadband/coverage-and-speeds/connected-nations-20252/data-downloads-2025
