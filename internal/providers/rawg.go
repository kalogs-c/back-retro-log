package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type rawgProvider struct {
	apiKey string
	client *http.Client
}

func NewRAWG(apiKey string) GameProvider {
	return &rawgProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type rawgGame struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Rating          float64 `json:"rating"`
	BackgroundImage *string `json:"background_image"`
	Released        *string `json:"released"`
	DescriptionRaw  string  `json:"description_raw"`
}

type rawgSearchResponse struct {
	Results []rawgGame `json:"results"`
	Count   int        `json:"count"`
}

type rawgDetailResponse struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	DescriptionRaw  string  `json:"description_raw"`
	BackgroundImage *string `json:"background_image"`
	Released        *string `json:"released"`
}

func (p *rawgProvider) Search(ctx context.Context, query string, page int) ([]Game, int, error) {
	u := fmt.Sprintf("https://api.rawg.io/api/games?key=%s&search=%s&page_size=%d&page=%d",
		p.apiKey, url.QueryEscape(query), PageSize, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to call RAWG: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("RAWG returned status %d", resp.StatusCode)
	}

	var searchResp rawgSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	games := make([]Game, 0, len(searchResp.Results))
	for _, g := range searchResp.Results {
		game := Game{
			RawgID:      g.ID,
			Title:       g.Name,
			Description: g.DescriptionRaw,
		}
		if g.BackgroundImage != nil {
			game.CoverURL = *g.BackgroundImage
		}
		if g.Released != nil {
			game.ReleaseDate = *g.Released
		}
		games = append(games, game)
	}
	return games, searchResp.Count, nil
}

func (p *rawgProvider) GetByID(ctx context.Context, id int) (*Game, error) {
	u := fmt.Sprintf("https://api.rawg.io/api/games/%d?key=%s", id, p.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call RAWG: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RAWG returned status %d", resp.StatusCode)
	}

	var detail rawgDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	game := &Game{
		RawgID:      detail.ID,
		Title:       detail.Name,
		Description: detail.DescriptionRaw,
	}
	if detail.BackgroundImage != nil {
		game.CoverURL = *detail.BackgroundImage
	}
	if detail.Released != nil {
		game.ReleaseDate = *detail.Released
	}
	return game, nil
}
