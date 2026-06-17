package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"back-retro-log/internal/i18n"
	"back-retro-log/internal/providers"
	"back-retro-log/ui"
)

type SearchHandler struct {
	Provider providers.GameProvider
}

func (h *SearchHandler) Page(w http.ResponseWriter, r *http.Request) {
	ui.Layout(true, ui.SearchPage()).Render(r.Context(), w)
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

	games, totalResults, err := h.Provider.Search(r.Context(), query, page)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %s", i18n.T(r.Context(), "error_failed_search"), err.Error()), http.StatusBadGateway)
		return
	}

	totalPages := (totalResults + providers.PageSize - 1) / providers.PageSize

	ui.SearchResults(games, query, page, totalResults, totalPages).Render(r.Context(), w)
}
