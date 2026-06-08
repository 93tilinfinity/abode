// Package web renders Abode's static HTML page.
//
// Today it renders only an infrastructure heartbeat; Ticket 4 extends it to
// render the sortable table of matching properties.
package web

import (
	"fmt"
	"html"
	"time"
)

// Heartbeat is the data shown on the placeholder page.
type Heartbeat struct {
	GeneratedAt time.Time
	MatchCount  int
	Note        string
}

// RenderHeartbeat returns a complete, self-contained HTML document confirming the
// daily run executed. It is deliberately tiny — proof of life, not product.
func RenderHeartbeat(h Heartbeat) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Abode</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 3rem auto; max-width: 40rem; padding: 0 1rem; color: #222; }
  .stamp { color: #666; font-variant-numeric: tabular-nums; }
</style>
</head>
<body>
<h1>Abode</h1>
<p>Daily run OK. Matches today: <strong>%d</strong>.</p>
<p class="stamp">Generated %s</p>
<p>%s</p>
</body>
</html>
`,
		h.MatchCount,
		html.EscapeString(h.GeneratedAt.Format(time.RFC3339)),
		html.EscapeString(h.Note),
	)
}
