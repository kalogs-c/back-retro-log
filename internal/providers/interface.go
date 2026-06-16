package providers

import "context"

type Game struct {
	RawgID      int
	Title       string
	CoverURL    string
	Description string
	ReleaseDate string
}

type GameProvider interface {
	Search(ctx context.Context, query string) ([]Game, error)
	GetByID(ctx context.Context, id int) (*Game, error)
}
