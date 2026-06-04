# Ticket 04 — Paid API property search source (Step 2)

## Goal

Run the single broad paid API search inside the catchment, applying every constraint the API supports directly, and requesting every other useful field the same results can return. This is the only paid breadth in the system; everything after it is light lookups on a shrinking set.

## Scope

- A source component (per Ticket 02) that searches the paid property API across the catchment polygon (Ticket 03), pushing down every constraint the API supports natively — price, bedrooms, bathrooms — so the candidate set comes back already filtered on cheap structured constraints.
- In the same results, request every other useful field the API returns as single field-reads: floor area, EPC, tenure, type, coordinates, photos, agent details, and any others available.
- Normalise each returned listing into the property record the pipeline carries, tagging each field with **source = primary feed**.
- "Buy simplicity": the source is asked only for property data — never for things that aren't property data.
- Floor area and similar advertised figures are carried as **"as advertised"**, not ground truth, per the data strategy.
- Robust to the API's paging/limits; **fails loudly** on auth/quota/transport errors rather than returning a silent empty set.

## Out of scope

- Filling fields the feed didn't return (Ticket 05).
- Area-level data (Ticket 06).

## How to validate

- A search over a known catchment returns candidates all inside the polygon and all satisfying the pushed-down price/bed/bath constraints.
- Returned records carry the extra fields (floor area, EPC, tenure, type, coords, photos, agent) where the API supplied them, each tagged source = primary feed.
- Fields the API omits are left explicitly absent (so they fail closed downstream), not fabricated.
- An injected API error (bad key / quota) raises a loud, visible failure; the run does not proceed with an empty set masquerading as "no matches".
- The search issues exactly one broad paid query per run for the set (cost discipline), verified against a mock.
