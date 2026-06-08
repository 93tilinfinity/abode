// Command abode-daily is the once-a-day batch entrypoint for Abode.
//
// Today it is a heartbeat: it wakes, does a placeholder unit of work, writes the
// static page, and exits. The scheduled GitHub Actions workflow runs it nightly
// and publishes the page to GitHub Pages, which proves the whole unattended loop
// works end to end. Feature tickets (1-4) replace the placeholder work and the
// page contents with the real product without changing this wake-work-exit shape
// or the workflow around it.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/93tilinfinity/abode/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// The batch responds to cancellation so it can be stopped cleanly when run
	// unattended. Feature tickets thread this context into their I/O lookups
	// (the listings/commute/crime/... sources) for per-call timeouts.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		// Fail loudly: a non-zero exit fails the scheduled workflow, which leaves
		// yesterday's published page intact and triggers the failure email.
		logger.Error("daily run failed", "err", err)
		os.Exit(1)
	}
}

// run is wake -> work -> write -> exit. It takes a context so the work can be
// timed out or cancelled, and returns an error rather than calling os.Exit so it
// stays unit-testable.
func run(ctx context.Context, logger *slog.Logger) error {
	start := time.Now().UTC()
	logger.Info("abode daily run starting", "at", start.Format(time.RFC3339))

	// --- work (placeholder) --------------------------------------------------
	// Ticket 3 replaces this with the real finding pipeline; for now it is a
	// heartbeat so we can assert the infrastructure runs unattended.
	matches, err := runPlaceholderWork(ctx, logger)
	if err != nil {
		return fmt.Errorf("running work: %w", err)
	}

	// --- write the page ------------------------------------------------------
	outDir := envOr("ABODE_OUTPUT_DIR", "public")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir %q: %w", outDir, err)
	}
	html := web.RenderHeartbeat(web.Heartbeat{
		GeneratedAt: start,
		MatchCount:  matches,
		Note:        "Infrastructure heartbeat — feature tickets will replace this with the real listings page.",
	})
	outFile := filepath.Join(outDir, "index.html")
	if err := os.WriteFile(outFile, []byte(html), 0o644); err != nil {
		return fmt.Errorf("writing %q: %w", outFile, err)
	}

	logger.Info("abode daily run complete",
		"output", outFile,
		"matches", matches,
		"duration", time.Since(start).String(),
	)
	return nil
}

// runPlaceholderWork stands in for the finding pipeline until Ticket 3. It takes
// a context and returns (matches, error) to model the shape the real pipeline
// will have — its data sources do network I/O that can fail or be cancelled.
func runPlaceholderWork(ctx context.Context, logger *slog.Logger) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	logger.Info("placeholder work: no real data sources wired yet")
	return 0, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
