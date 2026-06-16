package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"back-retro-log/internal/db"
	"back-retro-log/ui"
)

type CatalogHandler struct {
	Queries *db.Queries
}

func (h *CatalogHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(CtxUserID).(int64)
	status := r.URL.Query().Get("status")

	entries, err := h.Queries.ListCatalogEntries(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to load catalog", http.StatusInternalServerError)
		return
	}

	if status != "" {
		filtered := make([]db.ListCatalogEntriesRow, 0, len(entries))
		for _, e := range entries {
			if e.Status == status {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	isHTMX := r.Header.Get("HX-Request") == "true"
	if isHTMX {
		ui.CatalogCardList(entries).Render(r.Context(), w)
	} else {
		ui.Layout("Catalog", true, ui.CatalogPage(entries, status)).Render(r.Context(), w)
	}
}

func (h *CatalogHandler) Add(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
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
			http.Error(w, "Invalid rawg_id", http.StatusBadRequest)
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
		http.Error(w, "Failed to save game", http.StatusInternalServerError)
		return
	}

	_, err = h.Queries.GetCatalogEntryByUserAndGame(r.Context(), db.GetCatalogEntryByUserAndGameParams{
		UserID: userID,
		GameID: game.ID,
	})
	if err == nil {
		http.Error(w, "Game already in your catalog", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_, err = h.Queries.CreateCatalogEntry(r.Context(), db.CreateCatalogEntryParams{
		UserID: userID,
		GameID: game.ID,
		Status: "library",
	})
	if err != nil {
		http.Error(w, "Failed to add game", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

func (h *CatalogHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(CtxUserID).(int64)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")

	err = h.Queries.UpdateCatalogEntryStatus(r.Context(), db.UpdateCatalogEntryStatusParams{
		ID:     id,
		UserID: userID,
		Status: status,
	})
	if err != nil {
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	entry, err := h.Queries.GetCatalogEntry(r.Context(), db.GetCatalogEntryParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

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

func (h *CatalogHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(CtxUserID).(int64)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	err = h.Queries.DeleteCatalogEntry(r.Context(), db.DeleteCatalogEntryParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		http.Error(w, "Failed to delete entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
