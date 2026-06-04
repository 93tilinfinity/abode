# Plan — Ticket 3: The finding pipeline

Implementation plan for [Ticket 3](../tickets/03-finding-pipeline.md). It turns
the abstract four steps into concrete Go `Source` components (on the [Ticket 2](ticket-02-data-source-framework.md)
framework) bound to the providers chosen in [Decision 0001](../decisions/0001-data-providers.md),
evaluating the [Ticket 1](ticket-01-config-and-compiler.md) rules to produce the
day's matches.

## 1. What this ticket owns

The six real data sources and the order they run in (derived by the Ticket 2
engine from their declared `Needs`/`Produces`/`Cost`). It does **not** own the
framework, the rules, the page, or the shortlist.

| Package (`internal/source/...`) | Provider | Produces | Cost class |
|---|---|---|---|
| `catchment` | postcodes.io (geocode) + arithmetic | anchor coords, in-radius flag | free-local |
| `propertyapi` | PropertyData | price, beds, toilets?, type, coords, photo, agent, garden? | paid-broad |
| `epc` | Open Data Communities EPC | floor area (m²), EPC rating | light-lookup (free) |
| `broadband` | Ofcom postcode data / PropertyData `/internet-speed` | max download Mbps | light-lookup |
| `greenspace` | OSM Overpass / OS Open Greenspace | nearest-park metres | light-lookup (free) |
| `journey` | Google Routes (transit) | commute minutes, non-walking mode count | light-lookup (paid tier) |
| `crime` | police.uk + ONS population | LSOA violent/sexual-crime national percentile | area-data (free) |

## 2. The pipeline, step by step

### Step 1 — Catchment = radius pre-filter (free, local)
No isochrone. Geocode the work anchor (WC2A 1DD → lat/long via postcodes.io,
once, cached in-run) and define a **30-mile** radius around it. This is just the
search boundary handed to Step 2, plus a cheap point-in-radius sanity check. Most
of the UK is excluded here at no cost.

### Step 2 — PropertyData search (paid, one broad call per run)
One `/sourced-properties` call: anchor location, `radius=30` (miles),
`standardised_type` = all residential types (no type gate in this couple's
config), **`exclude_sstc=0`** (include under-offer / Sold-STC, per the decision),
requesting as many results as the matching set needs. Then `/sourced-property`
for full per-listing detail where the list rows are thin — used sparingly to
conserve credits. Populate price, beds, toilets (if present), type, coords,
photo, agent, and a garden/outdoor-space signal.

> **Strategy-list coverage is deferred to the free trial.** PropertyData has no
> single "everything for sale" query — only ~39 investor-oriented lists. Before
> locking the approach we'll prototype against the **500-credit trial**, measure
> coverage of one or more broad lists against a known Rightmove search for the
> area, and only then decide whether one broad list suffices or several must be
> merged + de-duplicated. This is a planned validation task, not a guess.

**Post-filter in code:** PropertyData has no native price/beds filter, so the
`price ≤ £625k` and `bedrooms ≥ 2` gates are applied client-side immediately
after this step — shrinking the set before any per-property lookup runs.

### Step 3 — Gap-fill (light, across the shrunk candidates)
Ordered so the **free** pruning gates run before the costed journey lookup, so
Google Routes only ever runs on properties that already passed everything else:

1. **`epc`** — match each listing to its EPC certificate (by address/postcode)
   for **floor area (m²)** (page column + shortlist £/sqft) and EPC rating. *Not
   a gate for this couple* — a missing EPC just leaves floor area blank; it never
   fails a property.
2. **`broadband`** — `fibre ≥ 300 Mbps` gate. Postcode-level availability/max
   speed (Ofcom or PropertyData `/internet-speed`). Unknown → **fails closed**.
3. **`greenspace`** — the park branch of `outdoor_space`. Because the rule is
   *garden OR park ≤ 800 m*, this only needs to run when the garden signal from
   Step 2 was false/unknown (short-circuit). Unknown both ways → fails closed.
4. **`journey`** (Google Routes, last) — door-to-door transit, **depart Tuesday
   08:00** → anchor. Read total minutes for `commute_time ≤ 50`, and count
   distinct non-walking leg modes for `commute_modes ≤ 2`. Runs only on survivors
   of all the above, minimising paid journeys.

### Step 4 — Area crime, by LSOA (light, on survivors)
Map each survivor to its **LSOA** (postcodes.io returns it), then evaluate the
`safe_area` gate as the **violent/sexual-crime rate per head over the last 12
months, ranked nationally** — fail if the LSOA is in the worst 25% (above the
75th percentile). Crime fetched per LSOA once and reused for all properties in it.

> **Wrinkle to resolve (national reference distribution).** Ranking an LSOA
> *nationally* needs the national distribution of LSOA crime rates. Rebuilding
> that every run from police.uk's per-area API is impractical (it's not a bulk
> national feed), which conflicts with Ticket 2's "no cross-run cache" decision.
> **Recommendation:** treat the national crime-rate distribution as a small
> **reference table refreshed periodically** (police.uk data is monthly, so
> monthly is ample) — a deliberate, narrow exception to "recompute all", since
> it's reference data, not per-run state. Flagged for your confirmation; the
> alternative is a simpler non-national measure (e.g. raw counts within a radius).

## 3. The matching set

Properties passing every gate are the day's matches, emitted as records (with
provenance flags from Ticket 2) for Ticket 4 to render. The only paid breadth is
the single Step 2 search; everything after is light lookups on a shrinking set,
with the paid Google journey deliberately last.

## 4. Decisions captured during the interview

| Decision | Choice | Consequence |
|----------|--------|-------------|
| Search radius | **30 miles / 48 km** | Covers fast-rail commuter towns; moderate fetch/journey volume. |
| Availability | **Include under-offer / Sold-STC** (`exclude_sstc=0`) | Bigger candidate set → more EPC/broadband/greenspace lookups and **more Google journeys** (watch the free tier); page shows homes that may already be under offer. |
| List coverage | **Decide in the free trial** | A measurement task precedes locking the list strategy. |
| Safety measure | **LSOA, per-capita, 12 months, national percentile** | Fine-grained + population-adjusted; needs ONS populations and a national reference distribution (see wrinkle). |

## 5. Hand-offs

- **Ticket 1** rules are evaluated by the framework as each field lands here.
- **Ticket 2** provides ordering, the shrinking set, in-run caching, provenance,
  spend logging, and fail-loudly.
- **Ticket 4** renders the matches (floor area shown *as advertised*; provenance
  markers subtle) and owns scheduling.
- **Ticket 5** reuses `commute_minutes`, floor area + Land Registry comps for
  £/sqft.

## 6. Validation (acceptance for Ticket 3)

1. **Radius pre-filter.** A property > 30 mi from the anchor never reaches Step 2;
   one just inside does. Anchor geocoded once per run.
2. **One paid broad call.** Step 2 invokes PropertyData once per run for the
   catchment, never per-property; `exclude_sstc=0` is sent; SSTC/under-offer
   listings appear in results.
3. **Client-side pre-filter.** `price ≤ £625k` and `beds ≥ 2` are applied right
   after Step 2 and shrink the set before any per-property lookup.
4. **Journey runs last and least.** Assert Google Routes is called only for
   properties that passed price/beds/toilets/fibre/outdoor — never the whole
   candidate set — and that it reads door-to-door minutes and counts non-walking
   modes correctly (≤ 50, ≤ 2) for a known journey, departing Tue 08:00.
5. **EPC is not a gate.** A property with no EPC match keeps blank floor area and
   is **not** failed; one with an EPC exposes floor area in m².
6. **Fibre fail-closed.** A postcode with unknown broadband fails the fibre gate.
7. **Outdoor OR short-circuit.** A garden property passes without a greenspace
   call; a garden-less one passes iff a park is ≤ 800 m; neither → fail closed.
8. **Crime by LSOA.** Two LSOAs either side of the 75th national percentile: every
   property in the worse one fails `safe_area`, none in the better one fails on
   it; crime is fetched once per LSOA. (Plus: the national reference approach
   agreed per §2 is implemented as decided.)
9. **Coverage trial.** A documented trial run measuring PropertyData list coverage
   vs a reference Rightmove search, with the chosen list strategy recorded.
10. **End-to-end.** With fixed mock providers and the example config, the pipeline
    returns exactly the properties passing all gates; flipping a config threshold
    changes the set as expected.
