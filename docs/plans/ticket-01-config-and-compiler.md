# Plan — Ticket 1: Config & requirement compiler

Implementation plan for [Ticket 1](../tickets/01-config-and-requirement-compiler.md),
worked against the couple's real must-haves. It pins down the config format, the
requirement model, the compiler's behaviour, and the exact rule each requirement
compiles to. Companion artifact: [`abode-config.example.yaml`](../examples/abode-config.example.yaml).

## 1. What this ticket owns (and doesn't)

Ticket 1 owns **the rules, not the data**. It reads one config file, turns each
must-have into a callable rule that returns `pass` / `fail` / `fail (no data)`,
and records each rule's kind (`deterministic` vs `judgement`). It does **not**
fetch anything — every rule names a *source* it will later ask for a field. The
real data behind those source names is Ticket 3's pipeline.

## 2. The config file

- **Format:** YAML — human-editable, supports comments so each requirement can
  explain itself.
- **Location (runtime):** a single `abode.config.yaml` at the project root. The
  worked example lives at `docs/examples/abode-config.example.yaml`.
- **Top-level shape:**
  - `work_anchor` — label, address, postcode the commute is measured to.
  - `commute_window` — `day` + `depart` time the journey is measured at.
  - `budget_reference_gbp` — the couple's true budget, for reference/reporting.
  - `requirements[]` — the must-haves (the gates).

## 3. The requirement model

Each requirement compiles to a rule carrying:

| Attribute | Meaning |
|-----------|---------|
| `id` | stable identifier (used in reporting, page columns, tests) |
| `description` | the plain-language "why", verbatim from the couple |
| `kind` | `deterministic` (v1 gate) or `judgement` (v2, recorded only) |
| `field` | the property attribute the rule reads |
| `source` | which pipeline lookup (Ticket 3) supplies the field |
| `op` | comparison: `<=`, `>=`, `<`, `>`, `==`, `within` |
| `value` | the threshold |
| `on_missing` | `fail` (default) — fail closed if the field can't be obtained |
| `any_of` / `all_of` | composite branches for OR / AND requirements |
| `defer_to` | for judgement requirements: the version that will implement it |

## 4. The compiler pipeline

1. **Load** the YAML.
2. **Validate** structure: every requirement has `id`, `description`, `kind`.
   A `deterministic` requirement must have either (`field` + `source` + `op` +
   `value`) or a composite (`any_of`/`all_of`); a `judgement` requirement must
   declare `defer_to`. Unknown `op`, duplicate `id`, or a `source` not registered
   by Ticket 3's pipeline is a loud compile error — config never half-compiles.
3. **Bind** each `field`/`source` pair to the data-source component that
   produces it (binding only — no fetch).
4. **Compile** each requirement into a callable rule `(property) -> Verdict`,
   where `Verdict ∈ {pass, fail, fail_no_data}` and carries the reason.
5. **Expose** the compiled set, each rule reporting its `kind` so the pipeline,
   page, and tests can see which are gating in v1.

### Evaluation semantics

- **Three verdicts only**, never a score or a maybe: `pass`, `fail`,
  `fail (no data)`.
- **Fail closed.** A field that is absent and unrecoverable yields
  `fail (no data)` with reason `"checkable by no data"` — never `pass`.
- **Composite.** `any_of` passes if any branch passes (used by `outdoor_space`);
  `all_of` passes only if all branches pass. A composite where every branch is
  `fail (no data)` is itself `fail (no data)`.

## 5. Kinds and the v1 judgement guard

- `deterministic` requirements compile and gate normally in v1.
- `judgement` requirements are **recorded but not built** in v1. The compiler
  emits them as non-gating, clearly labelled, so they show up everywhere as
  "deferred to v2 — judged by eye", and so v2 can pick them up with no config
  rewrite.
- **Guard:** if a `judgement` requirement is configured *as if it were gating*
  in v1 (e.g. given an `op`/`value` with no `defer_to`), the compiler **refuses
  loudly** rather than silently passing or failing it. A requirement the system
  cannot honestly confirm is never satisfied-by-assumption.

## 6. The couple's requirements → compiled rules

Eight deterministic v1 gates; two judgement requirements recorded for v2.

| # | Requirement (their words) | `id` | kind | v1 gate? | Check | Source | Missing |
|---|---------------------------|------|------|----------|-------|--------|---------|
| 1 | Under £575k budget | `price` | deterministic | ✅ | `price_gbp` ≤ **£625,000** (true budget £575k recorded separately) | property_api | always present |
| 2 | ≥ 2 bedrooms | `bedrooms` | deterministic | ✅ | `bedrooms` ≥ **2** | property_api | fail |
| 3 | 2 toilets | `toilets` | deterministic | ✅ | `toilet_count` ≥ **2** (a separate WC counts) | property_api | **fail** (OCR recovery is v2) |
| 4 | Great condition | `condition` | judgement | ❌ → v2 | — | — | judged by eye in v1 |
| 5 | ≤ 50 min commute | `commute_time` | deterministic | ✅ | `commute_minutes` ≤ **50**, Tue depart **08:00** → WC2A 1DD | journey | fail |
| 6 | ≤ 2 transport modes | `commute_modes` | deterministic | ✅ | `commute_modes_excluding_walk` ≤ **2** (walking excluded) | journey | fail |
| 7 | Interesting area | `interesting_area` | judgement | ❌ → v2 | — | — | judged by eye in v1 |
| 8 | Safe area | `safe_area` | deterministic | ✅ | violent/sexual-crime count within ~1 mi over 12 mo ≤ **threshold** (tunable) | crime | fail |
| 9 | Outdoor space | `outdoor_space` | deterministic | ✅ | `has_private_outdoor_space == true` **OR** `nearest_park_metres` ≤ **800** | property_api / greenspace | fail |
| 10 | Fibre internet | `fibre` | deterministic | ✅ | `max_download_mbps` ≥ **300** | broadband | fail |

### Decisions captured during the interview

- **Price:** real budget £575k; gate ceiling raised to **£625k** to surface
  negotiable listings. Budget stays recorded for reporting.
- **Toilets:** counted as toilets (a downstairs WC/cloakroom counts toward 2),
  mapped to the best available count field; missing → fail closed in v1.
- **Commute:** measured **Tuesday, depart 08:00**, to **81 Chancery Lane,
  WC2A 1DD**. Two separate gates — ≤ 50 min, and ≤ 2 modes with **walking
  excluded**.
- **Safe area:** *(revised — see [Decision 0001](../decisions/0001-data-providers.md)
  and the Ticket 4 interview)* a **raw police.uk count of violent/sexual crimes
  within ~1 mile over the last 12 months**, failing above an **absolute, tunable
  threshold** (placeholder 1200, calibrated in the trial). Recomputed each run;
  no national ranking. (Originally specified as a national percentile.)
- **Outdoor space:** OR — **any private outdoor space** (garden/yard/terrace/
  patio) satisfies it, *or* a park within **~800m** (≈10 min walk).
- **Fibre:** any line **≥ 300 Mbps** download (FTTP, FTTC or cable).
- **Condition & interesting area:** recorded as `judgement`, **deferred to v2**;
  judged by eye on the listing in v1.

## 7. Hand-offs this plan creates for later tickets

The compiler binds these `source` names; Ticket 3's pipeline implements the real
lookups behind them:

- `property_api` — `price_gbp`, `bedrooms`, `toilet_count`,
  `has_private_outdoor_space` (scraped from Rightmove, Ticket 3 step 1).
- `journey` — `commute_minutes` **and** `commute_modes_excluding_walk`, from a
  per-property **Google Routes** door-to-door lookup (Ticket 3 step 3). The
  catchment (Ticket 3 step 1) is a radius pre-filter, not an isochrone — see
  [Decision 0001](../decisions/0001-data-providers.md).
- `broadband` — `max_download_mbps` (secondary source, Ticket 3 step 3).
- `greenspace` — `nearest_park_metres` (secondary source, Ticket 3 step 3).
- `crime` — `violent_sexual_crime_1mi_12mo` (police.uk, Ticket 3 step 4).

## 8. Validation (acceptance for Ticket 1)

Extends the ticket's generic checklist with cases from this config:

1. **Round-trip.** Load `abode-config.example.yaml`; assert 10 requirements
   compile — 8 deterministic, 2 judgement — each with the field/op/value/source
   from §6.
2. **Boundaries.** `price` passes at £625,000 and fails at £625,001; `bedrooms`
   passes at 2, fails at 1; `commute_time` passes at 50 min, fails at 51;
   `commute_modes` passes at 2, fails at 3; `fibre` passes at 300, fails at 299;
   `safe_area` passes at the threshold count, fails one above it.
3. **Composite OR.** `outdoor_space` passes a property with a garden and no
   nearby park; passes one with no garden but a park at 800m; fails one with
   neither; returns `fail (no data)` if *both* branches are unobtainable.
4. **Fail closed.** A property missing `toilet_count` returns `fail (no data)`
   on `toilets` — never `pass`.
5. **Config is the only knob.** Change the price ceiling to £600k and the park
   distance to 400m in config, recompile, assert behaviour changes with no code
   edit.
6. **Judgement recorded, not gating.** Assert `condition` and `interesting_area`
   compile as `judgement`/`defer_to: v2`, do not gate the v1 result, and are
   labelled as such.
7. **v1 guard fires.** Give `condition` an `op`/`value` and remove `defer_to`;
   assert the compiler refuses loudly rather than gating it.
8. **Unknown source/op.** Point a requirement at an unregistered source or an
   unknown `op`; assert a loud compile error, not a half-built rule set.
