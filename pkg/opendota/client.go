package opendota

import (
	"context"
	"doproj/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiBaseURL      = "https://api.opendota.com"
	ImageCDNBaseURL = "https://cdn.cloudflare.steamstatic.com"
)

type OpenDotaClient struct {
	httpClient *http.Client
}

func NewOpenDotaClient() *OpenDotaClient {
	return &OpenDotaClient{
		httpClient: &http.Client{},
	}
}

func (c *OpenDotaClient) FetchHeroes(ctx context.Context) ([]models.Hero, error) {
	// Making Request
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/api/constants/heroes", nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	// Executing Request
	response, err := c.httpClient.Do(request)
	// Closing handlers
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	// Reading Response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	// Parsing Response
	var resp FetchHeroesResponse = make(FetchHeroesResponse)
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	// Mapping to internal model
	mappedHeroes := make([]models.Hero, 0, len(resp))
	for _, hero := range resp {
		hero := models.Hero{
			ID:       hero.ID,
			Name:     hero.Name,
			ImageURL: ImageCDNBaseURL + hero.ImageURL,
			Type:     hero.Type,
		}
		mappedHeroes = append(mappedHeroes, hero)
	}
	return mappedHeroes, nil
}

func (c *OpenDotaClient) FetchItems(ctx context.Context) ([]models.Item, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/api/constants/items", nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	var resp FetchItemsResponse = make(FetchItemsResponse)
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	mappedItems := make([]models.Item, 0, len(resp))
	for _, item := range resp {
		item := models.Item{
			ID:       item.ID,
			Name:     item.Name,
			ImageURL: ImageCDNBaseURL + item.ImageURL,
		}
		mappedItems = append(mappedItems, item)
	}
	return mappedItems, nil
}

func (c *OpenDotaClient) FetchPublicMatches(ctx context.Context) ([]models.PublicMatch, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/api/publicMatches", nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	// Decoding directly to struct
	var matches []models.PublicMatch
	if err := json.NewDecoder(response.Body).Decode(&matches); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}
	return matches, nil
}

func (c *OpenDotaClient) FetchMatchDetales(ctx context.Context, matchID uint64) (*models.MatchDetailsResponse, error) {
	url := fmt.Sprintf("%s/api/matches/%d", apiBaseURL, matchID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	var matchDetails models.MatchDetailsResponse
	if err := json.NewDecoder(response.Body).Decode(&matchDetails); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}
	return &matchDetails, nil
}
