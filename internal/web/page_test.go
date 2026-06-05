package web

import (
	"strings"
	"testing"
	"time"
)

func TestRenderHeartbeat(t *testing.T) {
	at := time.Date(2026, 6, 4, 5, 0, 0, 0, time.UTC)
	cases := []struct {
		name     string
		in       Heartbeat
		contains []string
	}{
		{
			name:     "zero matches",
			in:       Heartbeat{GeneratedAt: at, MatchCount: 0, Note: "hello"},
			contains: []string{"<!doctype html>", "Matches today: <strong>0</strong>", "2026-06-04T05:00:00Z", "hello"},
		},
		{
			name:     "some matches",
			in:       Heartbeat{GeneratedAt: at, MatchCount: 7, Note: "n"},
			contains: []string{"Matches today: <strong>7</strong>"},
		},
		{
			name:     "note is html-escaped",
			in:       Heartbeat{GeneratedAt: at, Note: "<script>x</script>"},
			contains: []string{"&lt;script&gt;x&lt;/script&gt;"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RenderHeartbeat(c.in)
			for _, want := range c.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n--- got ---\n%s", want, got)
				}
			}
		})
	}
}
