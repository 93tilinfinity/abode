# Ticket 2 — Pluggable data-source framework

**Depends on:** Ticket 1 (rules name the fields and sources this framework supplies).

> **Worked plan:** [`plans/ticket-02-data-source-framework.md`](../plans/ticket-02-data-source-framework.md)
> — interface design, derived ordering, and the captured decisions (fail loudly,
> no cross-run cache, log spend, subtle provenance).

## Goal

Make every data source the same kind of pluggable component, and build the
orchestrator that runs them in cost order — cheapest work across the most
properties, most expensive across the fewest — automatically, from each
component's own declarations rather than a hand-maintained sequence.

Concretely, deliver:

- **A uniform source interface.** Every source declares: what it **needs**
  (input fields / dependencies), what it **produces** (output fields), and its
  **cost class** (e.g. free-local, paid-broad, light-lookup). Adding a source is
  a one-file change.
- **An ordering engine.** From the declared needs/produces/cost across all
  registered sources, the orchestrator computes a valid execution order — a
  source runs only after the sources it depends on, and cheaper-cost work is
  scheduled ahead of more expensive work. Order is derived, never hand-written.
- **The shrinking-set contract.** The orchestrator carries a working set of
  properties through the ordered sources, and each stage can drop properties
  that fail a gate so later, costlier sources see a smaller set. "Heavy on the
  fewest, light on the most" is enforced by the engine, not by hoping.
- **Caching per source.** Each source's results are cached so reruns and shared
  lookups (e.g. area-level data) don't re-pay. Cache keys are the source's
  declared inputs.
- **Source/provenance flagging.** Every field value records which source
  produced it; a value recovered from a secondary source is as authoritative as
  a feed value but is flagged by origin (used by Tickets 3 and the page).
- **Missing-and-unrecoverable marking.** When no registered source can supply a
  required field, the framework marks it unrecoverable so the gate fails closed.

This ticket owns the *plumbing and ordering*. The actual sources (catchment,
paid API, fibre, area data) are implemented in Ticket 3 against this interface.

## Out of scope

- Any specific real data source (Ticket 3).
- The rules themselves (Ticket 1) — this framework feeds them fields.

## How to validate

1. **Uniform contract.** Register two or three stub sources with different
   needs/produces/cost and assert each conforms to the interface and self-reports
   its declarations.
2. **Derived ordering.** Give stubs a dependency chain (B needs a field A
   produces) and assert the engine schedules A before B with no manual order.
   Assert that, among independent sources, cheaper cost classes run first.
3. **Ordering changes with declarations.** Change a stub's cost class or
   dependency and assert the computed order changes accordingly — no code edit to
   the engine.
4. **Shrinking set.** Have an early stub drop half the set and assert a later
   stub is invoked on only the survivors (prove the costliest touches the
   fewest).
5. **Caching.** Run a source twice with identical declared inputs and assert the
   second run is served from cache (no second fetch).
6. **Provenance & fail-closed.** Assert a value carries its producing source, and
   that a field no source can supply is marked unrecoverable so its gate fails
   closed.
7. **One-file add.** Add a new stub source in a single file, register it, and
   assert it slots into the order automatically with no other change.

## Hygiene gate (before committing)

Before every commit on this ticket, the project hygiene gate must be green:

- `make check` passes — `gofmt -l .` clean, and `go vet ./...`, `go build ./...`,
  `go test ./...` all succeed;
- the new/changed behaviour is covered by tests;
- nothing is committed red (CI re-runs the same checks on push).

(The gate and the `make check` target are established in
[Ticket 0](00-project-setup-go.md).)
