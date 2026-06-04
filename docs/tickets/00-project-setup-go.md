# Ticket 0 — Go project setup (foundation)

**Depends on:** nothing. Stands up the skeleton Tickets 1–4 fill in.

> Code/workflows land in [PR #3](https://github.com/93tilinfinity/abode/pull/3); this doc is the plan.

## Goal

One Go module with a conventional layout, a pinned toolchain, fmt/vet/lint + tests
wired into CI on every push, **and the whole overnight loop stood up end-to-end as
a heartbeat** — so Tickets 1–4 are each "fill in a package and its tests", not "set
up plumbing first". Done when `go build ./...` and `go test ./...` pass, CI is
green, and the scheduled job publishes a page nightly (it just says "alive" until
Ticket 4 makes it meaningful).

## Why Go

Single static binary (wake-work-exit, nothing to install); errors are values you
must handle (fails loudly); stdlib `context`/goroutines cover the per-property
lookups (Ticket 3) without extra dependencies.

## Layout

```
abode/
├── go.mod / go.sum            # module github.com/93tilinfinity/abode; pinned go + toolchain
├── Makefile                   # build / test / fmt / vet / lint / check
├── .golangci.yml
├── abode.config.yaml          # single source of truth (Ticket 1)
├── cmd/
│   ├── abode-daily/           # daily batch entrypoint (Ticket 4)
│   └── abode-shortlist/       # on-demand shortlist (Ticket 5, v2)
└── internal/
    ├── config/ requirement/                                              # Ticket 1
    ├── rightmove/ broadband/ greenspace/ crime/ journey/ epc/ pipeline/  # Ticket 3
    ├── web/                                                              # Ticket 4
    └── shortlist/                                                        # Ticket 5 (v2)
```

Keep `cmd/` thin (parse flags, wire dependencies, call `internal/`); logic lives in
packages so it's testable. No `pkg/`. Each `internal/` data package is plain
functions the `pipeline` calls in order — no plugin framework (Decision 0001).

## Toolchain & dependencies

- Pin both lines: `go 1.24` + `toolchain go1.24.x` (latest stable), so CI and every
  machine use the same compiler.
- Stdlib-first — `net/http`, `encoding/json`, `log/slog`, `context`. The only
  third-party dependency is `gopkg.in/yaml.v3` (added in Ticket 1). Run
  `go mod tidy` after import changes; commit `go.mod`/`go.sum`.

## Idioms (the principles, enforced by CI)

- **`gofmt` is law** — CI rejects unformatted code.
- **Handle every error**, wrapping with `%w` so it propagates to the entrypoint and
  exits non-zero (this is how Abode "fails loudly").
- **Thin `main`, logic in packages** — packages are unit-testable, `main` isn't.
- **Accept interfaces, return structs** — fake an HTTP client in tests.
- **`context.Context` first** on anything doing I/O (timeouts/cancellation).
- **Doc comments** start with the name; **table-driven tests** live beside the code
  (`*_test.go`).

## Tooling & CI

- `gofmt`, `go vet ./...`, `golangci-lint` (govet, staticcheck, errcheck,
  ineffassign, gofmt).
- `Makefile` targets `build`/`test`/`fmt`/`vet`/`lint`, plus `check` = fmt + vet +
  lint + test (run before every commit).
- `.github/workflows/ci.yml` on push/PR: build, test, vet, lint, and fail if
  `gofmt -l .` reports any file.

## Overnight loop (heartbeat first)

Stand up the unattended path before any product exists, proven by a heartbeat, so
Tickets 3–4 only have to make the page meaningful:

- **`cmd/abode-daily`** = wake-work-exit. Its v0 "work" writes a tiny static page
  ("Daily run OK. Matches today: 0." + timestamp) with structured start/work/done
  logs (`log/slog`). Ticket 3 fills the pipeline; Ticket 4 fills the page.
- **`.github/workflows/daily.yml`** runs it on cron (+ manual `workflow_dispatch`)
  and publishes to **GitHub Pages**.
- **Fail loudly from day one:** non-zero exit fails the job → deploy is skipped
  (yesterday's page survives) → GitHub's failure email fires. **Zero matches is not
  a failure** — it publishes the honest page and exits 0.
- **One-time manual setup** (not codeable): Settings → Pages → Source: GitHub
  Actions, then run `daily` once. Repo is public (Pages on the Free plan; only
  already-public listing data is shown).

## Out of scope

Any Abode logic (Tickets 1–4); provider choices (Decision 0001); the real page
contents (Ticket 4).

## Validate

1. Module path + pinned `go`/`toolchain`; the layout exists, every package compiles.
2. `go build ./...`, `go test ./...`, `go vet ./...`, `golangci-lint run` all exit
   0; `gofmt -l .` prints nothing; `go mod tidy` is a no-op diff.
3. At least one real table-driven test exists.
4. `go run ./cmd/abode-daily` exits 0 and writes the heartbeat page.
5. CI is green on push; the `daily` workflow (run once) deploys to a live Pages URL.
6. A forced non-zero exit fails the workflow, skips the deploy (previous page
   intact), and fires the failure email.

## Hygiene gate

`make check` green (gofmt/vet/build/test) with tests for new behaviour before every
commit — established here and restated by Tickets 1–4.
