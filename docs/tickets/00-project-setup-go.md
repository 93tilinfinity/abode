# Ticket 0 — Go project setup (foundation)

**Depends on:** nothing. This precedes Ticket 1 — it creates the skeleton every
other ticket fills in.

> This ticket is written to teach. If you've never used Go, read it top to
> bottom: it explains *why* each choice is idiomatic, not just *what* to create.
> Nothing here is exotic — it's the boring, conventional Go setup, which is
> exactly what you want for a tool that must "run unattended and fail loudly".

## Goal

Stand up a single Go **module** with a conventional layout, a pinned toolchain,
formatting/vetting/linting wired in, a test harness, and CI that runs all of it
on every push — **and stand up the whole overnight pipeline end to end as a
heartbeat** — so that Tickets 1–5 are each "fill in a package and its tests"
rather than "set up plumbing first". When this ticket is done, `go build ./...`
and `go test ./...` succeed, CI is green, and the scheduled job actually runs
nightly and publishes a page — it just says "alive" instead of showing listings.

> **Implemented in [PR #3](https://github.com/93tilinfinity/abode/pull/3).** This
> ticket's doc is the plan; the code/workflows live in that separate PR so this
> planning PR stays docs-only.

## Why Go here (the short version)

Go fits Abode's principles well:

- **"Wake-work-exit" batch.** Go compiles to a single static binary with no
  runtime to install. The daily job is one file the scheduler executes — copy
  it, run it, done.
- **"Fails loudly."** Errors in Go are ordinary values you must handle, not
  exceptions you can forget. This nudges you toward surfacing failures instead
  of swallowing them.
- **Small, composable packages.** Each data lookup is an ordinary package the
  pipeline calls in a fixed order (Ticket 3) — simple to read and to test.
- **Concurrency** (for the per-property lookups in Ticket 3) is built into the
  language (`goroutines`, `context`), without extra libraries.

## Key Go vocabulary (so the rest of this makes sense)

- **Module** — the unit of dependency management, declared by a `go.mod` file at
  the repo root. Its *module path* is a globally-unique name, by convention the
  repo URL: `github.com/93tilinfinity/abode`. Import paths are built from it.
- **Package** — every `.go` file declares `package <name>`. A directory is one
  package; the package name is normally the directory name. Packages are the
  unit of reuse and of testing.
- **Exported vs unexported** — an identifier starting with a **Capital** letter
  is visible to other packages (`PublicThing`); lowercase is private to its own
  package (`internalHelper`). This is the *only* visibility rule — there's no
  `public`/`private` keyword.
- **`internal/`** — a magic directory name: packages under `internal/` can only
  be imported by code rooted at `internal/`'s parent. We put almost everything
  here so the project's internals can't accidentally become a public API.
- **`cmd/`** — by convention, each subdirectory of `cmd/` is one executable
  (`package main` with a `func main()`). Keep these *thin*: parse flags, wire
  dependencies, call into `internal/` packages. Logic lives in packages so it's
  testable.

## Proposed layout (maps directly onto Tickets 1–5)

```
abode/
├── go.mod                      # module path + pinned Go/toolchain versions
├── go.sum                      # checksums of dependencies (committed)
├── Makefile                    # short names for the common commands
├── .golangci.yml               # linter config
├── abode.config.yaml           # the single source of truth (Ticket 1)
├── cmd/
│   ├── abode-daily/            # the daily batch entrypoint (Ticket 4)
│   │   └── main.go
│   └── abode-shortlist/        # the on-demand shortlist tool (Ticket 5)
│       └── main.go
└── internal/
    ├── config/        # load + validate abode.config.yaml          (Ticket 1)
    ├── requirement/   # compile must-haves → pass/fail rules        (Ticket 1)
    ├── rightmove/     # scrape the search API (paginate + tile)     (Ticket 3)
    ├── broadband/     # fibre / Mbps lookup (Ofcom)                 (Ticket 3)
    ├── greenspace/    # nearest-park distance                       (Ticket 3)
    ├── crime/         # area crime data (police.uk)                 (Ticket 3)
    ├── journey/       # per-property commute time + modes (Routes)  (Ticket 3)
    ├── epc/           # floor area + EPC rating (page column only)  (Ticket 3)
    ├── pipeline/      # runs the gates in order over the set        (Ticket 3)
    ├── web/           # render the static sortable page             (Ticket 4)
    └── shortlist/     # weighted explainable scoring          (Ticket 5, v2)
```

Each `internal/` data package is just ordinary functions the `pipeline` calls in
sequence — there is no source-framework or plugin registry (see Decision 0001).

A note on conventions you'll see referenced online: avoid `pkg/` (an older
fashion that adds a directory for no benefit on a project like this), and don't
create packages "just in case" — add a directory only when a ticket needs it.

## The toolchain (pin it, don't float it)

`go.mod` carries two version lines. Best practice is to **pin** both so every
machine and CI run uses the same compiler:

```
module github.com/93tilinfinity/abode

go 1.24          // the minimum language version this code requires
toolchain go1.24.4   // the exact toolchain `go` will download/use if needed
```

Use the latest stable release available to you (Go 1.24 or newer). The `go`
directive sets the language baseline; the `toolchain` directive lets the `go`
command auto-fetch that exact version, so "works on my machine" stops being a
problem.

## Dependencies: as few as possible

Go's standard library is large and we lean on it. For this ticket we need at
most one third-party module:

- **`gopkg.in/yaml.v3`** — to parse `abode.config.yaml` in Ticket 1. (Add it
  with `go get gopkg.in/yaml.v3` when Ticket 1 starts; this ticket can stand up
  with zero dependencies.)

For things you might expect to need a library for, prefer the stdlib:

- **HTTP calls** (the property/broadband/journey APIs) → `net/http`.
- **JSON** → `encoding/json`.
- **Logging** → `log/slog` (structured logging, in the stdlib since Go 1.21).
- **Concurrency + timeouts** → `context` + goroutines.

Run `go mod tidy` after any import change; it adds what you use and removes what
you don't, keeping `go.mod`/`go.sum` honest. Commit both files.

## Idioms to follow from day one (these are the "best practices")

1. **`gofmt` is law.** Formatting isn't a style debate in Go — the tool decides.
   Configure your editor to run `gofmt` (or `gofumpt`, a stricter superset) on
   save. CI rejects unformatted code.
2. **Handle every error explicitly.** The idiom is:
   ```go
   v, err := doThing()
   if err != nil {
       return fmt.Errorf("doing thing: %w", err)  // wrap with %w to keep the chain
   }
   ```
   `%w` lets callers use `errors.Is`/`errors.As`. This is how Abode "fails
   loudly" — errors propagate up to the batch entrypoint, which exits non-zero.
3. **Keep `main` thin; put logic in packages.** `func main()` wires things and
   calls `internal/...`. Packages are unit-testable; `main` is not.
4. **Accept interfaces, return structs.** Functions take small interfaces (easy
   to fake in tests — e.g. an HTTP client) and return concrete types, so the
   pipeline's lookups stay testable without hitting the network.
5. **Pass `context.Context` as the first argument** to anything doing I/O, so
   network calls can time out and cancel: `func Fetch(ctx context.Context, ...)`.
6. **Doc comments** start with the name being documented and are full sentences:
   ```go
   // Compile turns a Requirement into a callable Rule. It returns an error if
   // the requirement is judgement-kind but configured to gate in v1.
   func Compile(r Requirement) (Rule, error) { ... }
   ```
7. **Table-driven tests** are the standard pattern. One test function, many
   cases in a slice — perfect for the boundary checks the feature tickets need:
   ```go
   func TestBedroomsGate(t *testing.T) {
       cases := []struct {
           name string
           beds int
           want Verdict
       }{
           {"below", 1, Fail},
           {"at",    2, Pass},
           {"above", 3, Pass},
       }
       for _, c := range cases {
           t.Run(c.name, func(t *testing.T) {
               if got := bedroomsRule(c.beds); got != c.want {
                   t.Errorf("beds=%d: got %v, want %v", c.beds, got, c.want)
               }
           })
       }
   }
   ```
   Run with `go test ./...`. Test files are named `*_test.go` and live beside
   the code they test.

## Tooling to install / wire up

- **`gofmt`** — ships with Go. (`gofumpt` optional, stricter.)
- **`go vet ./...`** — ships with Go; catches suspicious constructs.
- **`golangci-lint`** — the standard meta-linter; bundles `staticcheck` and
  others. A minimal `.golangci.yml` enabling `govet`, `staticcheck`,
  `errcheck`, `ineffassign`, and `gofmt`/`gofumpt` is plenty to start.
- **`gopls`** — the Go language server, used by the VS Code Go extension and
  JetBrains GoLand. Gives you autocomplete, jump-to-definition, inline errors.

A `Makefile` so you don't memorise flags:

```make
.PHONY: build test fmt vet lint check
build:  ; go build ./...
test:   ; go test ./...
fmt:    ; gofmt -l -w .
vet:    ; go vet ./...
lint:   ; golangci-lint run
check:  fmt vet lint test    # run before every commit
```

## CI (so "fails loudly" is enforced, not hoped for)

A GitHub Actions workflow at `.github/workflows/ci.yml` that, on push/PR:

1. checks out the code,
2. sets up Go using the version in `go.mod`,
3. runs `go build ./...`, `go test ./...`, `go vet ./...`, and
   `golangci-lint run`,
4. fails the check if `gofmt -l .` reports any file (i.e. something is
   unformatted).

This is the project's safety net: the daily tool can't silently rot if every
change must pass build + test + lint first.

## The overnight infrastructure (heartbeat first)

The riskiest thing in a "runs unattended" tool is the *unattended* part — the
schedule, the publish, the fail-loudly path. So this ticket stands that whole
loop up **before any product exists**, proven by a heartbeat, and the feature
tickets then fill it in without touching its shape:

- **`cmd/abode-daily`** is the wake-work-exit batch. In this ticket its "work" is
  a placeholder that writes a tiny static page reading *"Daily run OK. Matches
  today: 0."* with a timestamp, logging structured start/work/done lines
  (`log/slog`). Ticket 3 replaces the placeholder with the real finding pipeline;
  Ticket 4 replaces the page with the sortable listings table.
- **`.github/workflows/daily.yml`** runs it on a **cron schedule** (plus manual
  `workflow_dispatch`) and publishes the page to **GitHub Pages**.
- **Fail loudly is wired from day one:** a non-zero exit fails the job, which
  **skips the deploy** (so yesterday's page survives) and triggers GitHub's
  failure email. A zero-match run is *not* a failure — it publishes the honest
  page and exits 0.

This means Tickets 3 and 4 inherit a working, scheduled, self-publishing loop and
only have to make the page meaningful.

**One-time manual setup** (can't be done from code): enable **Settings → Pages →
Source: GitHub Actions**, then run the `daily` workflow once. The repo is
**public**, so GitHub Pages works on the Free plan; the page only ever shows
already-public listing data.

## Out of scope

- Any Abode logic — that's Tickets 1–5. This ticket proves the skeleton compiles,
  tests, lints, CI passes, and the scheduled loop publishes a heartbeat.
- Choosing the property API / journey provider (Ticket 3).
- The *contents* of the page — the real sortable listings table is Ticket 4; this
  ticket only establishes the publish loop.

## How to validate

1. **Module & layout.** `go.mod` exists with the module path
   `github.com/93tilinfinity/abode` and pinned `go`/`toolchain` lines; the
   directory skeleton above exists, each package compiling (even if it only
   holds a doc comment + package clause).
2. **Builds.** `go build ./...` exits 0 with the two `cmd/` entrypoints present
   (each can just print a version string for now).
3. **Tests run.** `go test ./...` exits 0, with at least one real (even if
   trivial) table-driven test demonstrating the pattern.
4. **Formatting clean.** `gofmt -l .` prints nothing (no unformatted files).
5. **Vet & lint clean.** `go vet ./...` and `golangci-lint run` both exit 0.
6. **Tidy module.** `go mod tidy` produces no diff in `go.mod`/`go.sum` (imports
   and dependencies agree).
7. **CI green.** Pushing the branch triggers the workflow and all steps pass.
8. **Runnable binary.** `go run ./cmd/abode-daily` runs, exits 0, and writes the
   heartbeat page — proving the wake-work-exit shape end-to-end.
9. **Overnight loop publishes.** The `daily` workflow (run manually once) builds
   the page and deploys it to a live GitHub Pages URL.
10. **Fail loudly.** A forced non-zero exit fails the workflow, the deploy step is
    skipped (previous page intact), and the failure email fires.

## Future fit: OCR and model calls in Go (your question)

Short answer: **yes, both the v2 OCR and the v2 model-judgement steps are fine
in Go** — and the cleanest way keeps them as ordinary pipeline lookups (Ticket 3)
that make HTTP calls, with no special language support needed.

- **OCR (v2 — recover bathroom count / square footage from floor plans).** Two
  routes:
  - *Hosted document-AI / OCR API over HTTP* (e.g. a cloud Vision/Document AI
    service): pure Go using `net/http` + `encoding/json`. **Recommended** — it
    slots in as just another source, needs no special build setup, and these
    services read floor-plan text far more accurately than self-hosted OCR.
  - *Tesseract locally* via the `gosseract` binding: works, but it uses **CGo**
    (Go calling C), which means you need the Tesseract C library installed and
    you lose Go's effortless static-binary/cross-compile story. Avoid unless you
    specifically want offline OCR.
- **Model judgement (v2 — condition from photos, area questions).** Calling a
  model is just an HTTPS request; there's an official Anthropic Go SDK, or you
  can use `net/http` directly. No friction.

So the v2 roadmap doesn't constrain the language choice: prefer **calling hosted
APIs over HTTP** for both OCR and model work, which keeps them consistent with
the rest of the pipeline's lookups and avoids CGo. The only time Go gets
fiddly is if you insist on running OCR *in-process* via Tesseract/CGo — and you
don't have to.

## Hygiene gate (before committing) — established here, repeated by every ticket

This ticket defines the gate every later ticket must pass **before each commit**:

- `make check` is green — `gofmt -l .` reports nothing, and `go vet ./...`,
  `go build ./...`, `go test ./...` all pass;
- new or changed behaviour carries tests (table-driven where it fits);
- nothing is committed red. CI re-runs the same checks on push, and the `daily`
  workflow publishes only on a successful run.

Every subsequent ticket (1–5) restates this gate in its own **Hygiene gate**
section so it is never skipped.
