package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// quietLogger discards log output so test runs stay clean.
func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestRunWritesPage is the happy path: run wakes, does the placeholder work, and
// writes a heartbeat page to ABODE_OUTPUT_DIR, returning no error.
func TestRunWritesPage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ABODE_OUTPUT_DIR", dir)

	if err := run(context.Background(), quietLogger()); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	out := filepath.Join(dir, "index.html")
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("expected a page at %s: %v", out, err)
	}
	for _, want := range []string{"<!doctype html>", "Matches today: <strong>0</strong>"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("page missing %q\n--- got ---\n%s", want, data)
		}
	}
}

// TestRunFailsLoudlyWhenOutputDirUnwritable proves the fail-loudly path: when the
// output directory can't be created, run returns an error (so the entrypoint
// exits non-zero and the scheduled deploy is skipped) rather than swallowing it.
func TestRunFailsLoudlyWhenOutputDirUnwritable(t *testing.T) {
	// Put a file where a parent directory would need to be, so MkdirAll fails.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ABODE_OUTPUT_DIR", filepath.Join(blocker, "sub"))

	if err := run(context.Background(), quietLogger()); err == nil {
		t.Fatal("expected an error when the output dir cannot be created, got nil")
	}
}

// TestRunHonoursCancelledContext proves the context is actually wired through: a
// cancelled context stops the run before it writes anything.
func TestRunHonoursCancelledContext(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ABODE_OUTPUT_DIR", dir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before run starts

	if err := run(ctx, quietLogger()); err == nil {
		t.Fatal("expected run to fail on a cancelled context, got nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "index.html")); !os.IsNotExist(err) {
		t.Errorf("page should not be written when the context is cancelled (stat err: %v)", err)
	}
}
