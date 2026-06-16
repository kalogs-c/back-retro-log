package handlers

import (
	"net/http"

	"back-retro-log/internal/providers"
	"back-retro-log/ui"
)

type SearchHandler struct {
	Provider providers.GameProvider
}

func (h *SearchHandler) Page(w http.ResponseWriter, r *http.Request) {
	ui.Layout("Search", true, ui.SearchPage()).Render(r.Context(), w)
}

func (h *SearchHandler) Results(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	games, err := h.Provider.Search(r.Context(), query)
	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	ui.SearchResults(games, query).Render(r.Context(), w)
}
