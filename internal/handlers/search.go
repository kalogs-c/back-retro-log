package handlers

import (
	"net/http"
	"strconv"

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

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	games, total, err := h.Provider.Search(r.Context(), query, page)
	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	totalPages := (total + providers.PageSize - 1) / providers.PageSize

	ui.SearchResults(games, query, page, totalPages).Render(r.Context(), w)
}
