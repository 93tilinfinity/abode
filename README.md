# Abode

A personal tool for finding a first home in the UK. See [`docs/SPEC.md`](docs/SPEC.md)
for the full system spec and [`docs/tickets/`](docs/tickets/) for the build plan.

## Status

**Ticket 0 (infrastructure) is built.** The overnight loop runs end to end as a
*heartbeat*: a scheduled GitHub Actions workflow runs the Go batch nightly, which
writes a static page and publishes it to GitHub Pages. The page currently just
confirms the run happened. Feature tickets 1–5 fill this skeleton with the real
product (config-driven gates, the data pipeline, the listings page, the shortlist
scorer) without changing this shape.

## Run it locally

```sh
go run ./cmd/abode-daily      # writes ./public/index.html
open public/index.html        # (or just open the file)
make check                    # gofmt + vet + build + test
```

## Layout

```
cmd/abode-daily/       # the daily batch entrypoint (Ticket 4)
cmd/abode-shortlist/   # the shortlist tool (Ticket 5, stub)
internal/page/         # static page rendering (Ticket 4)
docs/                  # spec, tickets, plans, decisions
```

(See [`docs/tickets/00-project-setup-go.md`](docs/tickets/00-project-setup-go.md)
for the full intended package layout the feature tickets grow into.)

## Continuous integration

- **`ci`** runs on every push/PR: `gofmt` check, `go vet`, `go build`, `go test`.
- **`daily`** runs on a cron schedule (and on demand via *Run workflow*): runs the
  batch and deploys the page to Pages. A failed run exits non-zero, skips the
  deploy (so yesterday's page stays up), and triggers GitHub's failure email.

## One-time setup (required for the nightly page to publish)

1. In the repo: **Settings → Pages → Build and deployment → Source: GitHub Actions.**
2. Trigger the `daily` workflow once via **Actions → daily → Run workflow** to
   confirm the page publishes.

The repo is **public**, so GitHub Pages is available on the Free plan; the page
only ever shows already-public listing data.

Future feature tickets will add API credentials as **Actions secrets**
(`PROPERTYDATA_KEY`, `GOOGLE_ROUTES_KEY`, `EPC_API_KEY`); none are needed for the
heartbeat.
