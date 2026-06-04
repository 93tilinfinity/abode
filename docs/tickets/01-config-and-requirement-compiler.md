# Ticket 01 — Config & requirement compiler

## Goal

Make config the only source of truth. Provide one editable config file where the couple declare their must-haves in plain language plus all settings (work anchor, commute budget, price/bed/bath bounds, EPC floor, thresholds, shortlist weights). Provide a compiler that turns each declared must-have into a checkable rule bound to a named data source, and that records, per requirement, whether it is **deterministic** or **judgement**.

Changing what the couple want must mean editing config, never code.

## Scope

- A single config file (e.g. `config.yaml` / `config.toml`) holding: work anchor + commute budget, the must-haves (price, bedrooms, bathrooms, EPC, fibre, commute-boundary, safety threshold), and shortlist weights. No requirement values live in code.
- A loader that validates the file on read and **fails loudly** with a precise message on a malformed or incomplete config — never silently defaults a missing must-have into "satisfied".
- A compiler that produces, per must-have, a rule object carrying: the field it reads, the comparison/operator, the bound value, the **kind** (`deterministic` | `judgement`), and the name of the data source that supplies its field.
- The kind classification is an explicit, recorded, inspectable property of each rule — not inferred at evaluation time.
- A rule evaluates a single property's populated fields to `pass` / `fail`, and **fails closed**: a field that is missing or unrecoverable evaluates to `fail`, never `pass`.
- In v1 the compiler accepts judgement-kind requirements only insofar as the spec defines their v1 handling (area questions → area-data threshold in Ticket 06; condition → out of scope, left to the couple). No model judgement is compiled.

## Out of scope

- Actually fetching any data (that's the source tickets).
- Model judgement evaluation (v2).

## How to validate

- Loading a valid config yields one rule per declared must-have, each tagged with the correct kind and bound source name.
- A config missing a required field, or with an unparseable value, aborts with a clear, specific error and produces no rules.
- Unit tests: a rule whose input field is present and satisfies the bound → `pass`; satisfies-not → `fail`; **field absent → `fail`** (fail-closed proven).
- Changing a bound (e.g. max price) in config and reloading changes the rule's verdict with zero code edits.
- Each compiled rule's `kind` is queryable and matches the spec's classification for that requirement.
