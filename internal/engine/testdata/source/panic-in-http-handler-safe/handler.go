package handlers

import (
	"log/slog"
	"net/http"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("recovered panic", "panic", rec)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}()

	panic("unexpected nil user")
}
