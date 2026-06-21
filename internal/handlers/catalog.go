package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"back-retro-log/internal/db"
	"back-retro-log/internal/i18n"
	"back-retro-log/ui"
)

type CatalogHandler struct {
	Queries *db.Queries
}

func (h *CatalogHandler) List(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())
	userID := r.Context().Value(CtxUserID).(int64)
	status := r.URL.Query().Get("status")

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := int64(20)
	offset := int64((page - 1) * int(limit))

	var entries []db.ListCatalogEntriesRow
	var total int64
	var err error

	if status != "" {
		statusRows, err := h.Queries.ListCatalogEntriesByStatus(r.Context(), db.ListCatalogEntriesByStatusParams{
			UserID: userID,
			Status: status,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			http.Error(w, loc.T("error_failed_load"), http.StatusInternalServerError)
			return
		}
		entries = make([]db.ListCatalogEntriesRow, len(statusRows))
		for i, e := range statusRows {
			entries[i] = db.ListCatalogEntriesRow(e)
		}
		cnt, err := h.Queries.CountCatalogEntriesByStatus(r.Context(), db.CountCatalogEntriesByStatusParams{
			UserID: userID,
			Status: status,
		})
		if err != nil {
			http.Error(w, loc.T("error_failed_load"), http.StatusInternalServerError)
			return
		}
		total = cnt
	} else {
		entries, err = h.Queries.ListCatalogEntries(r.Context(), db.ListCatalogEntriesParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			http.Error(w, loc.T("error_failed_load"), http.StatusInternalServerError)
			return
		}
		cnt, err := h.Queries.CountCatalogEntries(r.Context(), userID)
		if err != nil {
			http.Error(w, loc.T("error_failed_load"), http.StatusInternalServerError)
			return
		}
		total = cnt
	}

	totalPages := (int(total) + int(limit) - 1) / int(limit)

	isHTMX := r.Header.Get("HX-Request") == "true"
	if isHTMX {
		ui.CatalogCardList(entries, page, totalPages, status).Render(r.Context(), w)
	} else {
		ui.Layout(true, ui.CatalogPage(entries, status, page, totalPages)).Render(r.Context(), w)
	}
}

func (h *CatalogHandler) Add(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		http.Error(w, loc.T("error_bad_request"), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(CtxUserID).(int64)

	rawgIDStr := r.FormValue("rawg_id")
	title := r.FormValue("title")
	coverURL := r.FormValue("cover_url")
	releaseDate := r.FormValue("release_date")

	var game db.Game
	var err error

	if rawgIDStr != "" {
		rawgID, parseErr := strconv.ParseInt(rawgIDStr, 10, 64)
		if parseErr == nil {
			game, err = h.Queries.GetGameByRawgID(r.Context(), sql.NullInt64{Int64: rawgID, Valid: true})
			if err == sql.ErrNoRows {
				game, err = h.Queries.CreateGame(r.Context(), db.CreateGameParams{
					RawgID:      sql.NullInt64{Int64: rawgID, Valid: true},
					Title:       title,
					CoverUrl:    sql.NullString{String: coverURL, Valid: coverURL != ""},
					Description: sql.NullString{Valid: false},
					ReleaseDate: sql.NullString{String: releaseDate, Valid: releaseDate != ""},
				})
			}
		} else {
			http.Error(w, loc.T("error_invalid_rawg"), http.StatusBadRequest)
			return
		}
	} else {
		game, err = h.Queries.CreateGame(r.Context(), db.CreateGameParams{
			RawgID:      sql.NullInt64{Valid: false},
			Title:       title,
			CoverUrl:    sql.NullString{String: coverURL, Valid: coverURL != ""},
			Description: sql.NullString{Valid: false},
			ReleaseDate: sql.NullString{String: releaseDate, Valid: releaseDate != ""},
		})
	}
	if err != nil {
		http.Error(w, loc.T("error_failed_save"), http.StatusInternalServerError)
		return
	}

	_, err = h.Queries.GetCatalogEntryByUserAndGame(r.Context(), db.GetCatalogEntryByUserAndGameParams{
		UserID: userID,
		GameID: game.ID,
	})
	if err == nil {
		http.Error(w, loc.T("toast_duplicate"), http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, loc.T("error_internal"), http.StatusInternalServerError)
		return
	}

	_, err = h.Queries.CreateCatalogEntry(r.Context(), db.CreateCatalogEntryParams{
		UserID: userID,
		GameID: game.ID,
		Status: "library",
	})
	if err != nil {
		http.Error(w, loc.T("error_failed_add"), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/catalog")
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
	}
}

func (h *CatalogHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())
	userID := r.Context().Value(CtxUserID).(int64)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, loc.T("error_invalid_id"), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, loc.T("error_bad_request"), http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")

	err = h.Queries.UpdateCatalogEntryStatus(r.Context(), db.UpdateCatalogEntryStatusParams{
		ID:     id,
		UserID: userID,
		Status: status,
	})
	if err != nil {
		http.Error(w, loc.T("error_failed_update"), http.StatusInternalServerError)
		return
	}

	entry, err := h.Queries.GetCatalogEntry(r.Context(), db.GetCatalogEntryParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		http.Error(w, loc.T("error_not_found"), http.StatusNotFound)
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"message":"%s","type":"success"}}`, loc.T("toast_status_updated")))

	row := db.ListCatalogEntriesRow{
		ID:          entry.ID,
		UserID:      entry.UserID,
		GameID:      entry.GameID,
		Status:      entry.Status,
		CreatedAt:   entry.CreatedAt,
		UpdatedAt:   entry.UpdatedAt,
		Title:       entry.Title,
		CoverUrl:    entry.CoverUrl,
		Description: entry.Description,
		ReleaseDate: entry.ReleaseDate,
	}
	ui.CatalogCard(row).Render(r.Context(), w)
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())
	userID := r.Context().Value(CtxUserID).(int64)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	results, err := h.Queries.SearchCatalogEntries(r.Context(), db.SearchCatalogEntriesParams{
		UserID: userID,
		Title:  "%" + q + "%",
	})
	if err != nil {
		http.Error(w, loc.T("error_failed_search"), http.StatusInternalServerError)
		return
	}

	entries := make([]db.ListCatalogEntriesRow, 0, len(results))
	for _, e := range results {
		if status != "" && e.Status != status {
			continue
		}
		entries = append(entries, db.ListCatalogEntriesRow{
			ID:          e.ID,
			UserID:      e.UserID,
			GameID:      e.GameID,
			Status:      e.Status,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
			Title:       e.Title,
			CoverUrl:    e.CoverUrl,
			Description: e.Description,
			ReleaseDate: e.ReleaseDate,
		})
	}

	ui.CatalogCardList(entries, 1, 1, status).Render(r.Context(), w)
}

func (h *CatalogHandler) Delete(w http.ResponseWriter, r *http.Request) {
	loc := i18n.FromContext(r.Context())
	userID := r.Context().Value(CtxUserID).(int64)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, loc.T("error_invalid_id"), http.StatusBadRequest)
		return
	}

	err = h.Queries.DeleteCatalogEntry(r.Context(), db.DeleteCatalogEntryParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		http.Error(w, loc.T("error_failed_delete"), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"message":"%s","type":"success"}}`, loc.T("toast_game_removed")))
	w.WriteHeader(http.StatusOK)
}
