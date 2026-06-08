// Command abode-shortlist is the on-demand shortlist comparison tool.
//
// It is a stub today so the project layout matches the plan; the weighted,
// explainable scoring is implemented in Ticket 5.
package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.New(slog.NewJSONHandler(os.Stdout, nil)).
		Info("abode-shortlist is not implemented yet (Ticket 5)")
}
