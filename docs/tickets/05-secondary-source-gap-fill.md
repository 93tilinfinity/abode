# Ticket 05 — Secondary-source gap-fill (Step 3)

## Goal

For fields the primary feed didn't provide, fill them from secondary API sources — light field-reads and lookups applied across the candidate set, each cached. A recovered value is as authoritative as a feed value; a field that stays missing-and-unrecoverable is marked so its gate fails closed.

## Scope

- One or more source components (per Ticket 02) covering the v1 secondary fields: **fibre / internet availability**, **outdoor-space / park proximity**, and similar light lookups. Each declares its needs/produces/cost so the orchestrator slots it after the paid search and before area data.
- Each runs only for properties whose field is **missing** from the primary feed (don't re-pay for fields already populated), and each result is **cached**.
- A recovered value is written into the property record tagged with its **source** and **sanity-checked**; downstream it is treated as authoritative, flagged only by provenance.
- Fields that cannot be recovered are explicitly **marked unrecoverable**, so when their gate is evaluated the property **fails closed** (never satisfied-by-assumption).
- Honest-about-what-it-knows: any field here that rests on a heuristic rather than a clean dataset is labelled as such.

## Out of scope

- OCR recovery of image-only fields such as bathroom count / square footage — **v1 fails these closed**; OCR is v2.
- Area-level data such as crime (Ticket 06).

## How to validate

- A property with fibre absent from the feed gets it populated by the secondary source, tagged with that source; a property that already had it from the feed is **not** re-queried (assert no call).
- Recovered values pass a sanity check; an implausible value is rejected and the field marked unrecoverable.
- A field that no source can supply is marked unrecoverable, and its gate evaluates to `fail` (fail-closed proven end-to-end).
- Identical lookups are cached (second call hits cache).
- An image-only field (bathroom count missing from feed) is left to fail closed in v1, not OCR-recovered.
