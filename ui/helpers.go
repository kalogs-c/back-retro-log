package ui

import (
	"context"
	"database/sql"

	"back-retro-log/internal/i18n"
)

func AllStatuses() []string {
	return []string{
		"one_day_i_play",
		"finished",
		"didnt_liked",
		"maybe_i_come_back",
		"library",
	}
}

func StatusLabel(ctx context.Context, status string) string {
	return i18n.T(ctx, "status_"+status)
}

func StrVal(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
