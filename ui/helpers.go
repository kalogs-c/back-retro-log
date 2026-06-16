package ui

import "database/sql"

var StatusLabels = map[string]string{
	"one_day_i_play":    "One day I play",
	"finished":          "Finished",
	"didnt_liked":       "Didn't liked",
	"maybe_i_come_back": "Maybe I come back",
	"library":           "Library",
}

func AllStatuses() []string {
	return []string{
		"one_day_i_play",
		"finished",
		"didnt_liked",
		"maybe_i_come_back",
		"library",
	}
}

func StrVal(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
