# Ticket 1 — Config & requirement compiler

**Depends on:** [Ticket 0](00-project-setup-go.md). First feature ticket; every
other one reads its rules.

> Plan: [`plans/ticket-01-config-and-compiler.md`](../plans/ticket-01-config-and-compiler.md) · sample: [`examples/abode-config.example.yaml`](../examples/abode-config.example.yaml).

## Goal

Make config the single source of truth for the couple's must-haves and settings,
and compile each plain-language must-have into a pass/fail rule bound to a data
source. Changing what they want means editing config, never code.

Deliver:

- **One editable config file** (YAML) holding the must-haves and all settings: work
  anchor, commute target, price/beds/baths bounds, EPC floor, fibre requirement,
  safety threshold, shortlist weights.
- **A requirement model.** Each must-have declares: field, comparison
  (`>=`/`<=`/`within`/`is true`), threshold, kind (`deterministic`|`judgement`),
  and the source supplying the field. Kind is a recorded, visible decision.
- **A compiler** producing, per must-have, a rule that takes a property record and
  returns `pass` / `fail` / `fail (no data)` — never a score or a maybe.
- **Fail-closed semantics.** An absent-and-unrecoverable field returns `fail` with
  the reason recorded. Nothing is satisfied by assumption.
- **v1 guard.** Every must-have must compile `deterministic`; a `judgement` one is
  accepted into config but refused loudly at build (a v2 capability), never
  silently passed or failed.

Owns the *rules*, not the *data* — where a field comes from is Ticket 3.

## Out of scope

Fetching data (Ticket 3); model judgement (v2); scoring/weighting (Ticket 5 — same
config, separate path).

## Validate

1. **Round-trip.** Every sample must-have compiles to a rule with the declared
   field, comparison, threshold, kind, and bound source.
2. **Pass/fail boundaries.** Hand-built records return the expected verdict above,
   at, and below each threshold.
3. **Fail-closed.** A missing required field returns `fail (no data)` with the
   reason, never `pass`.
4. **Config is the only knob.** Changing a threshold (or adding a must-have) changes
   the compiled rules with no code edit.
5. **v1 judgement guard.** A `judgement` must-have is rejected loudly in v1.
6. **Visible kind.** Each rule exposes its `deterministic`/`judgement` kind for
   downstream stages and the page.

## Hygiene gate

`make check` green with tests for new behaviour before every commit — see
[Ticket 0](00-project-setup-go.md).
