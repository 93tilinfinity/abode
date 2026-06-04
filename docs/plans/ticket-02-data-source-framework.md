# Plan — Ticket 2: Pluggable data-source framework

Implementation plan for [Ticket 2](../tickets/02-data-source-framework.md). This
is the plumbing every real source (Ticket 3) plugs into: a uniform interface, an
ordering engine that derives execution order from each source's own
declarations, the shrinking working set, in-run caching, provenance, and the
failure policy. It consumes the compiled rules from [Ticket 1](../plans/ticket-01-config-and-compiler.md).

## 1. What this ticket owns (and doesn't)

Ticket 2 owns the **framework and the ordering**, not the data and not the
rules. It defines what a "source" *is*, computes the order sources run in, walks
the property set through them, evaluates each Ticket 1 rule the moment its inputs
exist (dropping failures early), and records where every value came from. The
concrete sources (catchment, property feed, broadband, …) arrive in Ticket 3;
the rules arrive from Ticket 1; the page that renders the result is Ticket 4.

## 2. The Source interface (the heart of "uniform parts")

In Go, the contract is one small interface. Every source — present and future —
satisfies it, which is what makes "adding a source is a one-file change" true.

```go
// Source is a single pluggable data source. Adding a new one means writing one
// file that implements this interface and registering it; the framework orders
// and runs it automatically.
type Source interface {
    // Name identifies the source (used in provenance, logs, ordering).
    Name() string

    // Needs lists the field names this source requires as input (its
    // dependencies). The ordering engine guarantees these are populated first.
    Needs() []Field

    // Produces lists the field names this source populates.
    Produces() []Field

    // Cost classifies how expensive this source is, used to order independent
    // sources cheapest-first.
    Cost() CostClass

    // Fetch populates this source's Produces() fields onto the working set.
    // It receives a context for timeouts/cancellation. Returning an error
    // aborts the run (see §6).
    Fetch(ctx context.Context, set *WorkingSet) error
}
```

Two Go idioms doing real work here:

- **Small interface, many implementations** — "accept interfaces, return
  structs". The engine only ever knows `Source`; each real source is a concrete
  struct in its own package under `internal/source/...`.
- **`context.Context` first** — every `Fetch` can be timed out and cancelled, so
  a hung API can't wedge the daily run.

## 3. The ordering engine (order is derived, never hand-written)

Each source declares `Needs`, `Produces`, and `Cost`. The engine builds a
dependency graph (an edge from a source to every source that produces a field it
needs) and computes an execution order by **topological sort**, breaking ties
between independent, ready-to-run sources by **cost, cheapest first**.

The spec's algorithm order *emerges* from declarations rather than being coded:

- the property search `Needs` the catchment boundary → catchment runs first;
- the broadband/greenspace/journey lookups `Need` the candidates' coordinates
  (produced by the property search) → they run after it;
- crime-by-area `Needs` the survivors' areas → it runs last.

If a future source changes its `Needs` or `Cost`, the order recomputes with no
edit to the engine. A cycle in the graph, or a needed field no source produces,
is a loud startup error — the run never begins half-wired.

## 4. The working set and "shrink early"

The engine carries a `WorkingSet` (the list of candidate properties) through the
ordered sources. The key move for "heavy on the fewest, light on the most":

> **A compiled rule is evaluated the instant all its required fields are
> populated — and failing properties are dropped immediately.**

So after the property feed populates price/beds/toilets/garden, those gates fire
and shrink the set *before* the per-property broadband/journey lookups run on
it; after those lookups, their gates fire; after crime-by-area, the safety gate
fires. Each later, costlier source sees a smaller set. This is enforced by the
engine, not left to discipline.

Rules whose fields are still missing simply don't evaluate yet. A field that no
remaining source can supply is **marked unrecoverable**, and its rule then fails
closed (Ticket 1 semantics).

## 5. Caching — in-run only (your decision: no cross-run persistence)

There are two distinct caches; you chose to keep one and drop the other:

- **In-run memoization — KEPT.** Within a single run, a source is asked for a
  given input at most once. This is what makes "pull crime *by area* and apply
  to every property in it" cheap — the area is fetched once and reused for all
  properties in it. This is core to the cost model, so it stays.
- **Cross-run persistence — DROPPED.** Nothing survives between daily runs. Each
  run rebuilds everything from scratch, **including the catchment isochrone**.
  Simplest possible: zero cache infrastructure, no external store, nothing to
  invalidate.

The framework still defines a tiny `Cache` interface with an in-memory
implementation, so a persistent backend could be added later as a one-file
change if cost ever becomes a concern — but **v1 ships in-memory only** and the
catchment is recomputed every run.

> **Trade-off to be aware of:** recomputing the catchment isochrone daily is the
> most expensive recurring piece of "recompute all". If your journey/isochrone
> provider charges per isochrone or rate-limits, this is the line item to watch
> — and the easiest later optimisation is to flip the catchment to a persistent
> cache. Flagged, not blocking.

## 6. Failure policy — always fail the run loudly (your decision)

Any source error aborts the **entire** run:

1. On a `Fetch` error, the source is retried a few times with exponential
   backoff (transient network blips shouldn't abort the day).
2. If it still fails — **whatever the source**, gating or secondary — the run
   stops, exits non-zero, and surfaces the error.
3. Ticket 4 leaves **yesterday's good page intact** rather than overwriting it.

This is the strictest reading of "fails loudly": the page is never silently
shrunk by a degraded source. The cost is that one flaky secondary API blocks the
whole day's run until it recovers (or you intervene) — which you accepted in
exchange for never publishing quietly-wrong results.

This is kept distinct from the legitimate "ran fine, zero matches today" case,
which is **not** a failure (Ticket 4 renders an honest empty page, exit 0).

## 7. Paid-API spend — log, don't cap (your decision)

The architecture still confines paid breadth to the single property-search
source (the framework never fans the paid API out per-property — that's a design
invariant, not a budget knob). On top of that, the engine records **per-source
call counts and estimated cost** and logs them via `log/slog` at the end of each
run, so you can watch spend. No hard ceiling is enforced. A `BudgetGuard` hook is
left in the interface so a cap can be switched on later without reworking
anything.

## 8. Provenance — subtle on the page (your decision)

Every populated field is wrapped so it carries its origin, not just its value:

```go
type Value[T any] struct {
    V         T
    Source    string // which source produced it
    Recovered bool   // true if filled by a secondary source, not the primary feed
    Heuristic bool   // true if it's a labelled heuristic rather than hard data
}
```

A value recovered from a secondary source is **as authoritative** as a feed
value for gating, but is flagged so Ticket 4 can render a **subtle marker /
tooltip** on recovered or heuristic values — visible if you look, not cluttering
the table. Recovered values are **sanity-checked** on the way in (a value outside
a plausible range is treated as not recovered), per the Ticket 1/3 plans.

## 9. Adding a source is one file

The registration pattern: a new source is a struct implementing `Source`, added
to the registry in its own file. The engine discovers it, slots it into the
derived order, and wires its `Produces` fields to any rules that need them — no
other file changes.

## 10. Decisions captured during the interview

| Decision | Choice |
|----------|--------|
| Source failure | **Always fail the run loudly** (after bounded retries); any source, gating or secondary, aborts the run and leaves yesterday's page up. |
| Cross-run caching | **None.** Recompute everything each run, including the catchment. In-run memoization kept. |
| Paid-API budget | **Log spend** (per-source calls + estimated cost); paid breadth stays confined to one broad search; no hard cap. |
| Provenance | **Subtle on the page** — markers/tooltips on recovered or heuristic values. |

## 11. Hand-offs this plan creates

- **Ticket 1** provides the compiled rules the engine evaluates field-by-field.
- **Ticket 3** implements concrete `Source`s (catchment, property feed,
  broadband, greenspace, journey, crime), each declaring `Needs`/`Produces`/
  `Cost`.
- **Ticket 4** consumes the failure semantics (don't clobber yesterday's page on
  abort), the spend log, and the provenance flags (subtle markers).

## 12. Validation (acceptance for Ticket 2)

Extends the ticket's generic checklist with cases from these decisions:

1. **Derived ordering.** Register stub sources with a dependency chain; assert
   the engine runs producers before consumers, and orders independent sources
   cheapest-cost first — with no hand-written sequence. Changing a stub's
   `Needs`/`Cost` changes the order with no engine edit.
2. **Bad wiring fails at startup.** A dependency cycle, or a needed field no
   source produces, aborts before any fetch with a clear error.
3. **Shrink early.** A gate satisfied by an early source's fields drops failing
   properties *before* a later source runs; assert the later source's `Fetch`
   receives only survivors.
4. **In-run memoization kept.** An area-level source asked about the same area
   for many properties fetches it once per unique area within a run.
5. **No cross-run persistence.** Assert nothing is written to disk/external
   store between runs and that a second run refetches from scratch (including the
   catchment).
6. **Fail loudly + retry.** A source returning a transient error is retried with
   backoff; a source that keeps failing aborts the whole run (non-zero exit) and
   the working set is **not** published — for both a gating and a secondary
   source.
7. **Spend logged.** After a run, per-source call counts and estimated cost are
   emitted in the structured log; assert the paid source is invoked once (one
   broad call), never per-property.
8. **Provenance flagged.** Each populated field carries its source; a value from
   a secondary source is marked `Recovered`; an out-of-range recovered value is
   rejected; the flags are exposed for the page to render subtly.
9. **One-file add.** Adding a new stub source in a single registered file slots
   it into the order automatically with no other change.
