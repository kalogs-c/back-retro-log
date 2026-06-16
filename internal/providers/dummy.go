package providers

import "context"

type dummyProvider struct{}

func NewDummy() GameProvider {
	return &dummyProvider{}
}

func (d *dummyProvider) Search(_ context.Context, query string) ([]Game, error) {
	return []Game{
		{
			RawgID:      1,
			Title:       "The Legend of Zelda: Breath of the Wild",
			CoverURL:    "https://media.rawg.io/media/games/cc5/cc5826d0e6ba3a25a2d2b3888ceb6c25.jpg",
			Description: "An open-world action-adventure game.",
			ReleaseDate: "2017-03-03",
		},
		{
			RawgID:      2,
			Title:       "Elden Ring",
			CoverURL:    "https://media.rawg.io/media/games/5ec/5ecac5aeb5a2851e3bca0f946a3c7f41.jpg",
			Description: "An action RPG set in a vast fantasy world.",
			ReleaseDate: "2022-02-25",
		},
	}, nil
}

func (d *dummyProvider) GetByID(_ context.Context, id int) (*Game, error) {
	return &Game{
		RawgID:      id,
		Title:       "Test Game",
		CoverURL:    "",
		Description: "A dummy game for testing.",
		ReleaseDate: "2024-01-01",
	}, nil
}
