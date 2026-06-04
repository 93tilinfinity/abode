# Ticket 1 — Config & requirement compiler

**Depends on:** [Ticket 0](00-project-setup-go.md) (the Go project skeleton).
This is the first *feature* ticket — the foundation every other feature ticket
reads from.

> **Worked plan:** [`plans/ticket-01-config-and-compiler.md`](../plans/ticket-01-config-and-compiler.md)
> — built against the couple's real must-haves, with a sample config at
> [`examples/abode-config.example.yaml`](../examples/abode-config.example.yaml).

## Goal

Make the config file the single source of truth for the couple's must-haves and
settings, and build the compiler that turns each plain-language must-have into a
checkable pass/fail rule bound to a data source. Changing what the couple want
must mean editing config, never code.

Concretely, deliver:

- **One editable config file** holding the must-haves and all settings (work
  anchor, commute target, price/beds/baths bounds, EPC floor, fibre requirement,
  safety threshold, shortlist weights, etc.). Plain, human-editable format.
- **A requirement model.** Each must-have declares: the field it checks, the
  comparison (e.g. `>=`, `<=`, `within`, `is true`), the threshold, the kind
  (`deterministic` or `judgement`), and which data source supplies the field.
  The kind is a recorded, visible decision on the requirement itself.
- **A compiler** that reads the config and produces, for each must-have, a
  callable rule that takes a property record and returns `pass`, `fail`, or
  `fail (no data)` — never a score or a maybe.
- **Fail-closed semantics built in.** A rule whose field is absent and
  unrecoverable returns `fail`, with the reason recorded ("checkable by no
  data"). Nothing is ever satisfied by assumption.
- **v1 guard:** every must-have must compile to `deterministic`. A must-have
  declared `judgement` is accepted into config but the compiler refuses to build
  it in v1 (it is a v2 capability) — and says so loudly rather than silently
  passing or failing it.

This ticket owns the *rules*, not the *data*: it produces rules that ask a data
source for a field. Where that field comes from is the pipeline's concern (Ticket 3).

## Out of scope

- Fetching any real data (Ticket 3).
- Any model judgement (v2).
- Scoring or weighting — that is the shortlist stage (Ticket 5), which reads the
  same config but is a separate path.

## How to validate

1. **Round-trip a config.** Load the sample config and assert every must-have
   compiles to a rule with the declared field, comparison, threshold, kind, and
   bound source.
2. **Pass/fail correctness.** Feed hand-built property records through the
   compiled rules and assert each gate returns the expected `pass`/`fail` for
   values above, at, and below the threshold (boundary cases included).
3. **Fail-closed.** Feed a record with a required field missing; assert the rule
   returns `fail (no data)` with the recorded reason — never `pass`.
4. **Config is the only knob.** Change a threshold in the config, recompile, and
   assert the rule's behaviour changes with no code edit. Add a new must-have in
   config and assert it appears as a compiled rule.
5. **v1 judgement guard.** Put a `judgement` must-have in config and assert the
   compiler rejects it loudly in v1 mode rather than treating it as met or unmet.
6. **Visible kind.** Assert each compiled rule exposes its `deterministic`/
   `judgement` kind so downstream stages and the page can show it.

## Hygiene gate (before committing)

Before every commit on this ticket, the project hygiene gate must be green:

- `make check` passes — `gofmt -l .` clean, and `go vet ./...`, `go build ./...`,
  `go test ./...` all succeed;
- the new/changed behaviour is covered by tests;
- nothing is committed red (CI re-runs the same checks on push).

(The gate and the `make check` target are established in
[Ticket 0](00-project-setup-go.md).)
