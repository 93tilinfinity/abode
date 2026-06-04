# Ticket 02 — Data-source component framework & pipeline orchestrator

## Goal

Make every data source the same kind of pluggable component, and have the system order and run them automatically so the lightest work touches the most properties and the heaviest touches the fewest. Adding a source must be a one-file change.

## Scope

- A uniform source interface. Every source declares: what it **needs** (input fields / preconditions), what it **produces** (output fields), and its **cost** (cheap local → light lookup → paid breadth → heavy). It exposes one operation that, given the working set, populates its produced fields.
- An orchestrator that, from the set of registered sources plus the compiled rules (Ticket 01), computes an execution order automatically from declared needs/produces (dependencies) and cost — **ordering is enforced, not hoped for**. The cheap, broad steps run first and shrink the set; expensive steps run only on survivors.
- After each source runs, the orchestrator applies any rules whose bound fields are now available, dropping failing properties immediately so later, costlier sources see the smallest possible set.
- Per-source **caching** hook so repeated/identical lookups are not re-paid.
- Fail-closed integration: properties missing a field a downstream gate needs are marked and dropped at that gate.
- Registration is declarative: dropping in a new source file with its needs/produces/cost makes it participate with no edits to the orchestrator.

## Out of scope

- The concrete sources themselves (Tickets 03–06) — this ticket provides the contract and the engine, with at least one trivial fake source for tests.

## How to validate

- Register two fake sources where B needs a field A produces; orchestrator runs A before B without any hand-ordered list.
- Register sources with mixed costs; assert the realised order is cost-ascending subject to dependencies, and that a costly source receives a set already reduced by cheaper gates (instrument the set size handed to each).
- Adding a third fake source as a single new file (no orchestrator edits) makes it run in the correct position.
- Cache test: identical lookup invoked twice hits the source once.
- A property missing a downstream-required field is dropped at the relevant gate and never reaches the next source.
