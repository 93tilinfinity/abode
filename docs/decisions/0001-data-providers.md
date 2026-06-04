# Decision 0001 — Data providers (the API spine and the gap-fill stack)

**Status:** Accepted · **Date:** 2026-06-04 · **Affects:** SPEC "Data strategy",
Ticket 2, Ticket 3.

## Context

The spec assumes "a paid property API is the spine, returning most attributes as
single field-reads … keeping the feed licensed rather than scraped," with free
secondary sources filling gaps. Before designing Ticket 3 we researched what
actually exists in the UK market (2025–2026) to confirm — or correct — that
assumption. Selection criteria the couple set: **most data for the fewest API
calls/integrations**, and for commute, **true door-to-door travel time** (not a
fancy catchment polygon).

## What the research found

### Live for-sale listings are the only genuinely hard part

- **No UK portal offers a self-serve licensed listings API.** Rightmove has none
  (its ADF/RTDF are agent-*upload* feeds); Zoopla closed its public API and gates
  access behind enterprise contracts; OnTheMarket is agent-upload only; Nestoria's
  old free API is dead. Listings must therefore come from a third party that
  licenses and aggregates portal/agent data.
- Self-serve options that actually expose live for-sale listings:
  - **PropertyData** — aggregates Rightmove/Zoopla/OnTheMarket (~75k), **and**
    bundles crime, broadband, schools, council tax, planning, flood, sold comps,
    EPC and floor areas. £28/mo (2,000 credits) + a 500-credit free trial.
    Caveat: listings are returned via predefined "strategy lists" with **no
    native price/beds filter** (post-filter client-side); full per-listing fields
    come from `/sourced-property` and should be verified in the trial.
  - **PaTMa** — £20/mo + cheap credits, listings with floor area and area search,
    but photo/agent field coverage is undocumented.
  - **Homedata (free tier)** — asking price + listing-status timeline + agent,
    but **no photos, UPRN-centric, no area for-sale search**.

### Secondary data is almost all free and authoritative (Open Government Licence)

- **Crime:** police.uk Data API — free, no key, point/polygon, incl. violent/
  sexual offences.
- **EPC incl. floor area (m²) and energy rating:** Open Data Communities EPC API
  — free (email key). This also supplies the floor area for £/sqft.
- **Sold-price comparables:** HM Land Registry Price Paid Data — free; combine
  with EPC floor area for £/sqft.
- **Parks/greenspace:** OS Open Greenspace (bulk, OGL) or OSM/Overpass (live,
  ODbL share-alike).
- **Schools:** GIAS bulk (free) — but Ofsted ratings removed from the dataset in
  Jan 2025.
- **Geocoding:** postcodes.io (free, no key) for postcode → lat/long.
- **Gaps:** per-address **fibre speed** is only free at *postcode* level (Ofcom
  Connected Nations CSVs); address-precise is paid (thinkbroadband) — PropertyData
  bundles an `/internet-speed` endpoint. **Council-tax band** has no free API.

### Commute

- **Google Routes API** gives door-to-door transit journeys globally (London tube/
  bus **and** National Rail / the commuter belt), returning `legs[].steps[].
  travelMode` and total duration — answering both gates: **door-to-door time ≤ 50
  min** and **non-walking modes ≤ 2**. Free tier ~10k Routes events/mo
  (post-March-2025), then ~$5/1k; requires a Google Cloud billing account.
- **TfL Unified API is free** (500 req/min, no card), London-authoritative, same
  per-leg mode detail — but coverage weakens for fast-rail commuter towns just
  outside Greater London.
- **Citymapper** has the best transit engine but its API is now enterprise-only
  ("Citymapper for Cities" via Via — contact-sales, no free tier), so it is ruled
  out for a personal project.
- **No free hosted transit isochrone** exists (TravelTime has no real free tier;
  self-hosting OTP/r5 is free but real ops). Google/Mapbox/ORS can't do transit
  isochrones at all — which is moot, since we use a radius + per-journey check.

## Decisions

1. **Listings spine: PropertyData.** Best fit for "most data, fewest
   integrations/calls": one licensed provider for the listings plus most secondary
   datasets. Choice confirmed pending a free-trial check of `/sourced-property`
   field coverage.
2. **Conserve credits by preferring free gov sources where they clearly win.**
   Use the free, authoritative OGL sources for **EPC + floor area**, **crime**,
   and **sold-price comps** (these are area- or property-level reads we can cache
   and they save PropertyData credits), and use PropertyData for the listings feed
   and any field not freely available. This keeps within the £28/2,000-credit plan.
3. **Commute: radius + Google Routes, no isochrone.** A generous straight-line
   **radius** around the work anchor is the cheap coarse filter (set in the
   PropertyData search); the real gate is a **Google Routes door-to-door transit
   journey** per surviving property, from which we read total minutes (≤ 50) and
   count non-walking modes (≤ 2). Chosen over TfL because the ≤ 50-min catchment
   reaches fast-rail commuter towns outside TfL's strong zone, and Google covers
   the whole catchment in one API. This matches "I care about door-to-door travel
   time" and removes the isochrone (the most expensive recurring piece flagged in
   Ticket 2). TfL remains a free fallback if the Google billing account or free-
   tier limits ever become a problem.

## Consequences

- **Refines SPEC "Data strategy" and Ticket 3 Step 1.** The catchment is no
  longer a computed transit *isochrone polygon*; it's a radius pre-filter plus a
  per-property door-to-door check. The per-property commute lookup added to
  Ticket 3 is now concretely **Google Routes** (transit mode, 08:00 weekday
  departure).
- **Google needs a billing account.** Even to use the free tier you must enable
  billing (card on file) in Google Cloud and keep an `API key` restricted to the
  Routes API. Monitor usage; the daily batch runs journeys only on the shrunk
  survivor set, so the ~10k/mo free tier should be ample.
- **Credit budget is the real constraint.** 2,000 credits/mo (~66/day) means: one
  broad listings search per day; secondary lookups pulled **once per area** and
  cached within the run; free gov sources used wherever they substitute. If usage
  outgrows the plan, the listings search breadth (radius/result count) is the
  first dial to turn.
- **Two requirements degrade gracefully, not perfectly:**
  - *Fibre ≥ 300 Mbps* — free data is postcode-level (Ofcom) or via PropertyData's
    `/internet-speed`; treat as postcode-level availability rather than guaranteed
    per-address. Fails closed where unknown (per Ticket 1).
  - *Council-tax band* — not a current must-have; if ever wanted, it needs a paid
    source. Out of scope for now.
- **Radius tuning matters.** Fast-rail commuter towns can be ≤ 50 min door-to-door
  yet far in straight-line distance, so the radius must be generous (tuned in
  Ticket 3) to avoid excluding them before the TfL check runs.
- **Coverage is no longer a worry.** Google Routes spans the whole catchment
  (London + commuter belt), so there's no London-only blind spot. The trade is
  the billing-account requirement above rather than a coverage gap.

## Provider → field map (planned for Ticket 3)

| Field / need | Source | Cost |
|---|---|---|
| For-sale listings (price, beds, type, coords, photos, agent) | PropertyData `/sourced-properties` + `/sourced-property` | paid (credits) |
| EPC rating + **floor area (m²)** | Open Data Communities EPC API | free |
| Sold-price comps (for £/sqft) | HM Land Registry Price Paid | free |
| Crime (safety gate) | police.uk Data API | free |
| Fibre / broadband speed | Ofcom CSV (postcode) or PropertyData `/internet-speed` | free / paid |
| Parks / greenspace | OS Open Greenspace or OSM Overpass | free |
| Commute time + modes (door-to-door) | Google Routes API (transit) | free tier (~10k/mo), then paid |
| Geocoding (postcode → lat/long) | postcodes.io | free |

## Key references

- PropertyData: https://propertydata.co.uk/api · https://propertydata.co.uk/api/pricing
- PaTMa: https://www.patma.co.uk/property-data-api/pricing/
- Homedata: https://homedata.co.uk/
- Rightmove (no public listings API): https://api-docs.rightmove.co.uk/apis
- Zoopla (API closed): https://developers.zoopla.co.uk/
- police.uk: https://data.police.uk/docs/
- EPC: https://epc.opendatacommunities.org/docs/api/domestic
- HM Land Registry Price Paid: https://landregistry.data.gov.uk/
- OS Open Greenspace: https://www.ordnancesurvey.co.uk/products/os-open-greenspace
- postcodes.io: https://postcodes.io/
- Google Routes API (transit + billing): https://developers.google.com/maps/documentation/routes/transit-route · https://developers.google.com/maps/documentation/routes/usage-and-billing
- TfL Unified API (free fallback): https://tfl.gov.uk/info-for/open-data-users/api-documentation
- Citymapper API (now enterprise via Via): https://ridewithvia.com/solutions/citymapper
- Ofcom Connected Nations: https://www.ofcom.org.uk/phones-and-broadband/coverage-and-speeds/connected-nations-20252/data-downloads-2025
